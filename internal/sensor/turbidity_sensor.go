package sensor

import (
	"fmt"
	"math/rand"
	"time"

	"owipex_reader/internal/modbus"
)

// TurbiditySensor represents a turbidity sensor.
type TurbiditySensor struct {
	BaseSensor
}

// NewTurbiditySensor creates a new TurbiditySensor instance.
func NewTurbiditySensor(id string, deviceID uint8, client *modbus.Client, config map[string]interface{}) *TurbiditySensor {
	readIntervalSeconds, ok := config["read_interval_seconds"].(float64) // JSON numbers are float64
	if !ok {
		readIntervalSeconds = 15 // Default interval
	}
	return &TurbiditySensor{
		BaseSensor: BaseSensor{
			ID:           id,
			Type:         "turbidity",
			DeviceID:     deviceID,
			ModbusClient: client,
			ReadInterval: time.Duration(readIntervalSeconds) * time.Second,
			Config:       config,
		},
	}
}

// ReadData reads data from the turbidity sensor.
// This is a placeholder implementation. Actual Modbus register reading logic will be added.
func (s *TurbiditySensor) ReadData(client *modbus.Client) (map[string]interface{}, error) {
	// Placeholder: Simulate reading a turbidity value from Modbus registers
	// In a real scenario, you would use client.ReadHoldingRegisters or similar methods
	// e.g., registers, err := client.ReadHoldingRegisters(s.DeviceID, 10, 1) // Read 1 register from address 10
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to read registers for turbidity sensor %s: %w", s.ID, err)
	// }
	// simulatedTurbidityValue := float32(registers[0]) / 10.0 // Example conversion

	// Simulate data for now
	simulatedTurbidityValue := 50.0 + rand.Float32()*100 // Random turbidity between 50 and 150 NTU

	data := map[string]interface{}{
		"turbidity": fmt.Sprintf("%.1f", simulatedTurbidityValue),
	}
	return data, nil
}

