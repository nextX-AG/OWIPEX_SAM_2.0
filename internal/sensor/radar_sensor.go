package sensor

import (
	"fmt"
	"math/rand"
	"time"

	"owipex_reader/internal/modbus"
)

// RadarSensor represents a radar sensor.
type RadarSensor struct {
	BaseSensor
}

// NewRadarSensor creates a new RadarSensor instance.
func NewRadarSensor(id string, deviceID uint8, client *modbus.Client, config map[string]interface{}) *RadarSensor {
	readIntervalSeconds, ok := config["read_interval_seconds"].(float64) // JSON numbers are float64
	if !ok {
		readIntervalSeconds = 20 // Default interval for radar sensor
	}
	return &RadarSensor{
		BaseSensor: BaseSensor{
			ID:           id,
			Type:         "radar",
			DeviceID:     deviceID,
			ModbusClient: client,
			ReadInterval: time.Duration(readIntervalSeconds) * time.Second,
			Config:       config,
		},
	}
}

// ReadData reads data from the radar sensor.
// This is a placeholder implementation. Actual Modbus register reading logic will be added.
func (s *RadarSensor) ReadData(client *modbus.Client) (map[string]interface{}, error) {
	// Placeholder: Simulate reading distance from Modbus registers
	// In a real scenario, you would use client.ReadHoldingRegisters or similar methods
	// e.g., registers, err := client.ReadHoldingRegisters(s.DeviceID, 30, 1) // Read 1 register from address 30
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to read registers for radar sensor %s: %w", s.ID, err)
	// }
	// simulatedDistance := float32(registers[0]) / 100.0 // Example conversion (e.g., cm to m)

	// Simulate data for now
	simulatedDistance := 1.5 + rand.Float32()*2 // Random distance between 1.5 and 3.5 meters

	data := map[string]interface{}{
		"distance": fmt.Sprintf("%.2f", simulatedDistance),
	}
	return data, nil
}

