package sensor

import (
	"fmt"
	"log"
	"math"
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
// Implementation based on the Python flow_sensor.py
func (s *FlowSensor) ReadData(client *modbus.Client) (map[string]interface{}, error) {
	// From Python implementation:
	// - Register 0x000A: Total Flow Low
	// - Register 0x0011: Total Flow High
	// - Register 0x1438: Flow Unit
	// - Register 0x1439: Flow Decimal Point
	// - Register 0x0001: Flow Rate

	// Read total flow low register
	time.Sleep(300 * time.Millisecond) // Längere initiale Pause vor jeder Kommunikation

	flowIntLowRegs, err := client.ReadHoldingRegisters(s.DeviceID, 0x000A, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to read total flow low register for sensor %s: %w", s.ID, err)
	}
	flowIntLow := flowIntLowRegs[0]

	time.Sleep(500 * time.Millisecond) // Wichtige Pause zwischen den Register-Abfragen

	// Read total flow high register
	flowIntHighRegs, err := client.ReadHoldingRegisters(s.DeviceID, 0x0011, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to read total flow high register for sensor %s: %w", s.ID, err)
	}
	flowIntHigh := flowIntHighRegs[0]

	time.Sleep(500 * time.Millisecond) // Längere Pause vor den Konfigurationsregistern

	// Read flow unit register
	flowUnitRegs, err := client.ReadHoldingRegisters(s.DeviceID, 0x1438, 1)
	if err != nil {
		log.Printf("Warning: failed to read flow unit register for sensor %s: %v, using default (0)", s.ID, err)
		flowUnitRegs = []uint16{0} // Default: m³
	}
	flowUnit := flowUnitRegs[0]

	time.Sleep(300 * time.Millisecond)

	// Read flow decimal point register
	flowDecimalPointRegs, err := client.ReadHoldingRegisters(s.DeviceID, 0x1439, 1)
	if err != nil {
		log.Printf("Warning: failed to read flow decimal point register for sensor %s: %v, using default (3)", s.ID, err)
		flowDecimalPointRegs = []uint16{3} // Standard-Dezimalpunkt
	}
	flowDecimalPoint := flowDecimalPointRegs[0]

	time.Sleep(1000 * time.Millisecond) // Lange Pause vor dem Lesen der Flow Rate

	// Read flow rate
	flowRateRegs, err := client.ReadHoldingRegisters(s.DeviceID, 0x0001, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to read flow rate register for sensor %s: %w", s.ID, err)
	}
	flowRate := float64(flowRateRegs[0])

	// Validate values and use defaults if needed
	if flowUnit == 0xFFFF {
		flowUnit = 0 // Default: m³
	}

	if flowDecimalPoint == 0xFFFF {
		flowDecimalPoint = 3 // Default decimal point
	}

	// Calculate total flow
	flowInteger := (uint32(flowIntHigh) << 16) | uint32(flowIntLow)
	multiplier := math.Pow(10, float64(flowDecimalPoint)-3)
	totalFlow := float64(flowInteger) * multiplier

	// Map flow unit to string
	flowUnitMap := map[uint16]string{
		0: "m³",
		1: "L",
		2: "GAL",
		3: "CF",
		5: "ft³",
	}
	flowUnitStr, ok := flowUnitMap[flowUnit]
	if !ok {
		flowUnitStr = "m³" // Default
	}

	data := map[string]interface{}{
		"flow_rate":       fmt.Sprintf("%.2f", flowRate),
		"total_flow":      fmt.Sprintf("%.2f", totalFlow),
		"total_flow_unit": flowUnitStr,
	}

	return data, nil
}
