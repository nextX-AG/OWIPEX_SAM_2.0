// Package adapter enthält Adapter, die zwischen unterschiedlichen Schichten der Anwendung vermitteln.
package adapter

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"owipex_reader/internal/config"
	"owipex_reader/internal/service"
	"owipex_reader/internal/types"
)

// SensorAdapter verbindet die neue Sensorarchitektur mit der ThingsBoard-Integration.
type SensorAdapter struct {
	deviceService   *service.DeviceService
	sensors         []types.Sensor
	logger          *log.Logger
	stopChan        chan struct{}
	wg              sync.WaitGroup
	thingsboardChan chan map[string]interface{}
	readIntervals   map[string]time.Duration
	lastReadTimes   map[string]time.Time
	appConfig       *config.AppConfig
}

// NewSensorAdapter erstellt einen neuen SensorAdapter.
func NewSensorAdapter(configPath string, deviceConfigPath string, tbChan chan map[string]interface{}) (*SensorAdapter, error) {
	logger := log.New(os.Stdout, "[SensorAdapter] ", log.LstdFlags)

	// Anwendungskonfiguration laden
	appCfg, err := config.LoadAppConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("Fehler beim Laden der Anwendungskonfiguration: %w", err)
	}

	// DeviceService erstellen
	deviceService := service.NewDeviceService(deviceConfigPath)
	if err := deviceService.Initialize(); err != nil {
		return nil, fmt.Errorf("Fehler beim Initialisieren des DeviceService: %w", err)
	}

	// Sensoren laden
	sensors, err := deviceService.LoadSensorsFromConfig()
	if err != nil {
		return nil, fmt.Errorf("Fehler beim Laden der Sensoren: %w", err)
	}

	logger.Printf("Erfolgreich %d Sensoren geladen", len(sensors))

	// Read-Intervalle aus der Konfiguration extrahieren
	readIntervals := make(map[string]time.Duration)
	for _, sensorCfg := range appCfg.Sensors {
		if sensorCfg.Enabled {
			readIntervals[sensorCfg.ID] = time.Duration(sensorCfg.ReadIntervalSeconds) * time.Second
		}
	}

	return &SensorAdapter{
		deviceService:   deviceService,
		sensors:         sensors,
		logger:          logger,
		stopChan:        make(chan struct{}),
		thingsboardChan: tbChan,
		readIntervals:   readIntervals,
		lastReadTimes:   make(map[string]time.Time),
		appConfig:       appCfg,
	}, nil
}

// Start startet den SensorAdapter.
func (a *SensorAdapter) Start() {
	a.logger.Println("Starte SensorAdapter...")
	a.wg.Add(1)
	go a.run()
}

// Stop stoppt den SensorAdapter.
func (a *SensorAdapter) Stop() {
	a.logger.Println("Stoppe SensorAdapter...")
	close(a.stopChan)
	a.wg.Wait()
	a.logger.Println("SensorAdapter gestoppt.")
}

// run führt den Hauptloop des SensorAdapters aus.
func (a *SensorAdapter) run() {
	defer a.wg.Done()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-a.stopChan:
			a.logger.Println("SensorAdapter-Loop wird gestoppt.")
			return
		case <-ticker.C:
			// Über alle Sensoren iterieren
			for _, sensor := range a.sensors {
				sensorID := sensor.ID()

				// Prüfen, ob es Zeit ist, den Sensor zu lesen
				lastRead, exists := a.lastReadTimes[sensorID]
				readInterval, intervalExists := a.readIntervals[sensorID]

				if !intervalExists {
					// Standardintervall verwenden, wenn keines konfiguriert ist
					readInterval = 15 * time.Second
				}

				if !exists || time.Since(lastRead) >= readInterval {
					a.logger.Printf("Lese Sensor: %s", sensorID)
					a.wg.Add(1)

					go func(s types.Sensor) {
						defer a.wg.Done()

						// Sensor lesen
						ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
						defer cancel()

						reading, err := s.Read(ctx)
						a.lastReadTimes[s.ID()] = time.Now()

						if err != nil {
							a.logger.Printf("Fehler beim Lesen des Sensors %s: %v", s.ID(), err)
							a.thingsboardChan <- map[string]interface{}{
								fmt.Sprintf("%s_error", s.ID()): err.Error(),
							}
							return
						}

						a.logger.Printf("Sensor %s erfolgreich gelesen: %v", s.ID(), reading.Value)

						// Daten für ThingsBoard formatieren
						formattedData := a.formatReadingForThingsboard(s, reading)
						a.thingsboardChan <- formattedData

					}(sensor)
				}
			}
		}
	}
}

// formatReadingForThingsboard formatiert die Sensordaten für ThingsBoard.
func (a *SensorAdapter) formatReadingForThingsboard(s types.Sensor, reading types.Reading) map[string]interface{} {
	// Konfiguration für diesen Sensor finden
	var sensorCfg config.SensorConfig
	for _, cfg := range a.appConfig.Sensors {
		if cfg.ID == s.ID() {
			sensorCfg = cfg
			break
		}
	}

	// Einfaches Format für ThingsBoard
	simplePayload := make(map[string]interface{})

	// Hauptwert hinzufügen
	valueName := fmt.Sprintf("%s_%s", s.ID(), reading.Type)
	simplePayload[valueName] = reading.Value

	// Weitere Metadaten hinzufügen
	for key, value := range reading.Metadata {
		simplePayload[fmt.Sprintf("%s_%s", s.ID(), key)] = value
	}

	// JSON-Format für ThingsBoard (strukturierte Daten)
	measurements := map[string]interface{}{
		string(reading.Type): reading.Value,
	}

	// Metadaten zu den Messungen hinzufügen
	for key, value := range reading.Metadata {
		measurements[key] = value
	}

	jsonPayload := map[string]interface{}{
		fmt.Sprintf("%s_data", s.ID()): map[string]interface{}{
			"info": map[string]interface{}{
				"name":     sensorCfg.Name,
				"location": sensorCfg.Location,
				"type":     string(s.Type()),
				"id":       s.ID(),
			},
			"metadata":     sensorCfg.Metadata,
			"measurements": measurements,
			"timestamp":    reading.Timestamp,
			"status":       "active",
			"unit":         reading.Unit,
		},
	}

	return map[string]interface{}{
		"simple": simplePayload,
		"json":   jsonPayload,
	}
}
