package sensor

import (
	"fmt"
	"math/rand"
	"time"

	"owipex_reader/internal/modbus"
)

// FlowSensor represents a flow sensor.
type FlowSensor struct {
	BaseSensor
}

// NewFlowSensor creates a new FlowSensor instance.
func NewFlowSensor(id string, deviceID uint8, client *modbus.Client, config map[string]interface{}) *FlowSensor {
	readIntervalSeconds, ok := config["read_interval_seconds"].(float64) // JSON numbers are float64
	if !ok {
		readIntervalSeconds = 10 // Default interval for flow sensor
	}
	return &FlowSensor{
		BaseSensor: BaseSensor{
			ID:           id,
			Type:         "flow",
			DeviceID:     deviceID,
			ModbusClient: client,
			ReadInterval: time.Duration(readIntervalSeconds) * time.Second,
			Config:       config,
		},
	}
}

// ReadData reads data from the flow sensor.
// This is a placeholder implementation. Actual Modbus register reading logic will be added.
func (s *FlowSensor) ReadData(client *modbus.Client) (map[string]interface{}, error) {
	// Placeholder: Simulate reading flow rate and total volume from Modbus registers
	// In a real scenario, you would use client.ReadHoldingRegisters or similar methods
	// e.g., registers, err := client.ReadHoldingRegisters(s.DeviceID, 20, 4) // Read 4 registers (two 32-bit floats) from address 20
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to read registers for flow sensor %s: %w", s.ID, err)
	// }
	// simulatedFlowRate := math.Float32frombits(uint32(registers[0])<<16 | uint32(registers[1])) // Example conversion for 32-bit float
	// simulatedTotalVolume := math.Float32frombits(uint32(registers[2])<<16 | uint32(registers[3])) // Example conversion

	// Simulate data for now
	simulatedFlowRate := 10.0 + rand.Float32()*5    // Random flow rate between 10 and 15 L/min
	simulatedTotalVolume := 1000.0 + rand.Float32()*100 // Random total volume between 1000 and 1100 L

	data := map[string]interface{}{
		"flow_rate":    fmt.Sprintf("%.2f", simulatedFlowRate),
		"total_volume": fmt.Sprintf("%.2f", simulatedTotalVolume),
	}
	return data, nil
}

