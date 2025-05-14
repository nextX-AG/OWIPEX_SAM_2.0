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
// Implementation based on the Python turbidity_sensor.py
func (s *TurbiditySensor) ReadData(client *modbus.Client) (map[string]interface{}, error) {
	// Read turbidity value (Register 0x0001, 2 registers)
	turbidityRegs, err := client.ReadHoldingRegisters(s.DeviceID, 0x0001, 2)
	if err != nil {
		return nil, fmt.Errorf("failed to read turbidity value for sensor %s: %w", s.ID, err)
	}

	// Combine the two registers if needed based on modbus implementation
	// In this implementation, using just the first register as in the Python code
	turbidityRaw := float64(turbidityRegs[0])

	// Read temperature (Register 0x0003, 2 registers)
	temperatureRegs, err := client.ReadHoldingRegisters(s.DeviceID, 0x0003, 2)
	if err != nil {
		// Just log warning and continue without temperature
		temperatureValue := 0.0
		data := map[string]interface{}{
			"turbidity":     calculateAdjustedTurbidity(turbidityRaw),
			"turbidity_raw": fmt.Sprintf("%.1f", turbidityRaw),
			"temperature":   fmt.Sprintf("%.1f", temperatureValue),
		}
		return data, nil
	}

	// Combine the two registers if needed
	// Using just the first register as in the Python code
	temperatureValue := float64(temperatureRegs[0])

	data := map[string]interface{}{
		"turbidity":     calculateAdjustedTurbidity(turbidityRaw),
		"turbidity_raw": fmt.Sprintf("%.1f", turbidityRaw),
		"temperature":   fmt.Sprintf("%.1f", temperatureValue),
	}

	return data, nil
}

// calculateAdjustedTurbidity applies the same adjustment as in the Python code
func calculateAdjustedTurbidity(turbidityRaw float64) string {
	// 1. Subtract 30 from raw value
	adjustedTurbidity := turbidityRaw - 30.0

	// 2. Make sure value is at least 1
	if adjustedTurbidity <= 0 {
		if turbidityRaw/6.0 < 1.0 {
			adjustedTurbidity = 1.0
		} else if turbidityRaw/6.0 > 3.0 {
			adjustedTurbidity = 3.0
		} else {
			adjustedTurbidity = turbidityRaw / 6.0
		}
	}

	// 3. Add random variation for more realistic decimal places
	randomVariation := rand.Float64()*0.69 - 0.32 // between -0.32 and 0.37
	adjustedTurbidity = adjustedTurbidity + randomVariation

	return fmt.Sprintf("%.1f", adjustedTurbidity)
}
