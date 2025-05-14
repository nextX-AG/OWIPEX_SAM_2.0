package sensor

import (
	"fmt"
	"time"

	"owipex_reader/internal/modbus"
)

// PHSensor represents a pH sensor.
type PHSensor struct {
	BaseSensor
}

// NewPHSensor creates a new PHSensor instance.
func NewPHSensor(id string, deviceID uint8, client *modbus.Client, config map[string]interface{}) *PHSensor {
	readIntervalSeconds, ok := config["read_interval_seconds"].(float64) // JSON numbers are float64
	if !ok {
		readIntervalSeconds = 15 // Default interval
	}
	return &PHSensor{
		BaseSensor: BaseSensor{
			ID:           id,
			Type:         "ph",
			DeviceID:     deviceID,
			ModbusClient: client,
			ReadInterval: time.Duration(readIntervalSeconds) * time.Second,
			Config:       config,
		},
	}
}

// ReadData reads data from the pH sensor.
// Implementation based on the Python ph_sensor.py
func (s *PHSensor) ReadData(client *modbus.Client) (map[string]interface{}, error) {
	// Read PH value (Register 0x0001, 2 registers)
	phRegs, err := client.ReadHoldingRegisters(s.DeviceID, 0x0001, 2)
	if err != nil {
		return nil, fmt.Errorf("failed to read pH value for sensor %s: %w", s.ID, err)
	}

	// Use the first register as the pH value (same as in Python)
	phValue := float64(phRegs[0])

	// Read temperature (Register 0x0003, 2 registers)
	tempRegs, err := client.ReadHoldingRegisters(s.DeviceID, 0x0003, 2)
	if err != nil {
		return nil, fmt.Errorf("failed to read temperature for sensor %s: %w", s.ID, err)
	}

	// Use the first register as the temperature value (same as in Python)
	temperature := float64(tempRegs[0])

	data := map[string]interface{}{
		"ph":          fmt.Sprintf("%.2f", phValue),
		"temperature": fmt.Sprintf("%.1f", temperature),
	}

	return data, nil
}
