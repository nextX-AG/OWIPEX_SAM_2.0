package manager

import (
	"fmt" // Added fmt import
	"log"
	"os"
	"sync"
	"time"

	"owipex_reader/internal/config"
	"owipex_reader/internal/modbus"
	"owipex_reader/internal/sensor"
)

// SensorManager manages all sensors and their readings.
type SensorManager struct {
	configPath      string
	appConfig       *config.AppConfig
	sensors         map[string]sensor.Sensor
	modbusClient    *modbus.Client
	logger          *log.Logger
	stopChan        chan struct{}
	wg              sync.WaitGroup
	busLock         sync.Mutex
	lastCommTime    time.Time
	debounceTime    time.Duration
	thingsboardChan chan map[string]interface{} // Channel to send data to ThingsBoard
}

// NewSensorManager creates a new SensorManager.
func NewSensorManager(configPath string, tbChan chan map[string]interface{}) (*SensorManager, error) {
	logger := log.New(os.Stdout, "[SensorManager] ", log.LstdFlags)

	appCfg, err := config.LoadAppConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load application config: %w", err)
	}

	mbClient, err := modbus.NewClient(modbus.ClientConfig{
		URL:      appCfg.RS485.Port,
		BaudRate: uint(appCfg.RS485.Baudrate), // Cast to uint
		DataBits: uint(appCfg.RS485.DataBits), // Cast to uint
		Parity:   appCfg.RS485.Parity,
		StopBits: uint(appCfg.RS485.StopBits), // Cast to uint
		Timeout:  time.Duration(appCfg.RS485.TimeoutMS) * time.Millisecond,
	})
	if err != nil {
		logger.Printf("Failed to initialize or open Modbus client, will retry connection attempts during run: %v", err)
	}

	sm := &SensorManager{
		configPath:      configPath,
		appConfig:       appCfg,
		sensors:         make(map[string]sensor.Sensor),
		modbusClient:    mbClient,
		logger:          logger,
		stopChan:        make(chan struct{}),
		debounceTime:    500 * time.Millisecond,
		thingsboardChan: tbChan,
	}

	errLoadSensors := sm.loadSensorsFromConfig(appCfg.Sensors)
	if errLoadSensors != nil {
		return nil, fmt.Errorf("failed to load sensors: %w", errLoadSensors)
	}

	return sm, nil
}

func (sm *SensorManager) loadSensorsFromConfig(sensorConfigs []config.SensorConfig) error {
	for _, sc := range sensorConfigs {
		// Skip sensors that are not enabled
		if !sc.Enabled {
			sm.logger.Printf("Skipping disabled sensor: %s (Type: %s, DeviceID: %d)", sc.ID, sc.Type, sc.DeviceID)
			continue
		}

		var s sensor.Sensor

		sensorSpecificConfig := make(map[string]interface{})
		sensorSpecificConfig["read_interval_seconds"] = sc.ReadIntervalSeconds
		sensorSpecificConfig["name"] = sc.Name
		sensorSpecificConfig["location"] = sc.Location
		sensorSpecificConfig["metadata"] = sc.Metadata

		switch sc.Type {
		case "ph":
			s = sensor.NewPHSensor(sc.ID, uint8(sc.DeviceID), sm.modbusClient, sensorSpecificConfig)
		case "turbidity":
			s = sensor.NewTurbiditySensor(sc.ID, uint8(sc.DeviceID), sm.modbusClient, sensorSpecificConfig)
		case "flow":
			s = sensor.NewFlowSensor(sc.ID, uint8(sc.DeviceID), sm.modbusClient, sensorSpecificConfig)
		case "radar":
			s = sensor.NewRadarSensor(sc.ID, uint8(sc.DeviceID), sm.modbusClient, sensorSpecificConfig)
		default:
			sm.logger.Printf("Unknown sensor type: %s for sensor ID: %s", sc.Type, sc.ID)
			continue
		}

		sm.sensors[sc.ID] = s
		sm.logger.Printf("Loaded sensor: %s (Type: %s, DeviceID: %d)", sc.ID, sc.Type, sc.DeviceID)
	}
	sm.logger.Printf("Total sensors loaded: %d", len(sm.sensors))
	return nil
}

func (sm *SensorManager) Start() {
	sm.logger.Println("Starting SensorManager...")
	sm.wg.Add(1)
	go sm.run()
}

func (sm *SensorManager) Stop() {
	sm.logger.Println("Stopping SensorManager...")
	close(sm.stopChan)
	sm.wg.Wait()
	if sm.modbusClient != nil {
		err := sm.modbusClient.Close()
		if err != nil {
			sm.logger.Printf("Error closing modbus client: %v", err)
		}
	}
	sm.logger.Println("SensorManager stopped.")
}

func (sm *SensorManager) waitForBus() {
	sm.busLock.Lock()
	defer sm.busLock.Unlock()

	currentTime := time.Now()
	timeSinceLast := currentTime.Sub(sm.lastCommTime)

	if timeSinceLast < sm.debounceTime {
		waitTime := sm.debounceTime - timeSinceLast
		time.Sleep(waitTime)
	}
	sm.lastCommTime = time.Now()
}

func (sm *SensorManager) run() {
	defer sm.wg.Done()

	if sm.modbusClient != nil && !sm.modbusClient.IsConnected() {
		sm.logger.Printf("Modbus client not connected at start of run loop. Attempting to open...")
		err := sm.modbusClient.Open()
		if err != nil {
			sm.logger.Printf("Failed to open Modbus client on start, will retry on demand: %v", err)
		} else {
			sm.logger.Printf("Modbus client opened successfully at start of run loop.")
		}
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sm.stopChan:
			sm.logger.Println("SensorManager run loop stopping.")
			return
		case <-ticker.C:
			for id, s := range sm.sensors {
				if time.Since(s.GetLastReadTime()) >= s.GetReadInterval() {
					sm.logger.Printf("Scheduled to read sensor: %s", id)
					sm.wg.Add(1)
					go func(currentSensor sensor.Sensor) {
						defer sm.wg.Done()
						sm.waitForBus()

						if sm.modbusClient == nil {
							sm.logger.Printf("Modbus client is nil, cannot read sensor %s.", currentSensor.GetID())
							return
						}

						if !sm.modbusClient.IsConnected() {
							sm.logger.Printf("Modbus client not connected. Attempting to (re)open for sensor %s...", currentSensor.GetID())
							err := sm.modbusClient.Open()
							if err != nil {
								sm.logger.Printf("Failed to (re)open Modbus client for sensor %s: %v", currentSensor.GetID(), err)
								return
							}
							sm.logger.Printf("Modbus client (re)opened successfully for sensor %s.", currentSensor.GetID())
						}

						data, err := currentSensor.ReadData(sm.modbusClient)
						currentSensor.SetLastReadTime(time.Now())
						if err != nil {
							sm.logger.Printf("Error reading sensor %s: %v", currentSensor.GetID(), err)
							sm.thingsboardChan <- map[string]interface{}{
								fmt.Sprintf("%s_error", currentSensor.GetID()): err.Error(),
							}
							return
						}
						sm.logger.Printf("Successfully read sensor %s: %v", currentSensor.GetID(), data)

						formattedData := sm.formatSensorDataForThingsboard(currentSensor, data)
						sm.thingsboardChan <- formattedData

					}(s)
				}
			}
		}
	}
}

func (sm *SensorManager) formatSensorDataForThingsboard(s sensor.Sensor, data map[string]interface{}) map[string]interface{} {
	var sensorCfg config.SensorConfig
	for _, cfg := range sm.appConfig.Sensors {
		if cfg.ID == s.GetID() {
			sensorCfg = cfg
			break
		}
	}

	simplePayload := make(map[string]interface{})
	for key, value := range data {
		simplePayload[fmt.Sprintf("%s_%s", s.GetID(), key)] = value
	}

	jsonPayload := map[string]interface{}{
		fmt.Sprintf("%s_data", s.GetID()): map[string]interface{}{
			"info": map[string]interface{}{
				"name":      sensorCfg.Name,
				"location":  sensorCfg.Location,
				"type":      s.GetType(),
				"device_id": s.GetDeviceID(),
			},
			"metadata":     sensorCfg.Metadata,
			"measurements": data,
			"timestamp":    time.Now().UnixNano() / int64(time.Millisecond),
			"status":       "active",
		},
	}

	return map[string]interface{}{
		"simple": simplePayload,
		"json":   jsonPayload,
	}
}
