package sensor

import (
	"fmt"
	"math/rand"
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
// This is a placeholder implementation. Actual Modbus register reading logic will be added.
func (s *PHSensor) ReadData(client *modbus.Client) (map[string]interface{}, error) {
	// Placeholder: Simulate reading a pH value and temperature from Modbus registers
	// In a real scenario, you would use client.ReadHoldingRegisters or similar methods
	// e.g., registers, err := client.ReadHoldingRegisters(s.DeviceID, 0, 2) // Read 2 registers from address 0
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to read registers for pH sensor %s: %w", s.ID, err)
	// }
	// simulatedPHValue := float32(registers[0]) / 100.0 // Example conversion
	// simulatedTempValue := float32(registers[1]) / 10.0  // Example conversion

	// Simulate data for now
	simulatedPHValue := 6.5 + rand.Float32() // Random pH between 6.5 and 7.5
	simulatedTempValue := 20.0 + rand.Float32()*5 // Random temp between 20 and 25

	data := map[string]interface{}{
		"ph":          fmt.Sprintf("%.2f", simulatedPHValue),
		"temperature": fmt.Sprintf("%.1f", simulatedTempValue),
	}
	return data, nil
}

