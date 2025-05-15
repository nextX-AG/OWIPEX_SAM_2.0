// Package flow implementiert einen Durchflusssensor.
package flow

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"owipex_reader/internal/device/sensor"
	"owipex_reader/internal/types"
)

// Konstanten für Flow-Sensoren
const (
	// Standard-Register-Namen
	RegisterFlowRate         = "flow_rate"
	RegisterTotalFlowLow     = "total_flow_low"
	RegisterTotalFlowHigh    = "total_flow_high"
	RegisterFlowUnit         = "flow_unit"
	RegisterFlowDecimalPoint = "flow_decimal_point"

	// Kalibrierungsparameter
	CalibrationOffset = "offset"
	CalibrationScale  = "scale"
)

// Konstanten für Register-Adressen (basierend auf der alten Implementierung)
const (
	DefaultRegisterFlowRate         = uint16(0x0001)
	DefaultRegisterTotalFlowLow     = uint16(0x000A)
	DefaultRegisterTotalFlowHigh    = uint16(0x0011)
	DefaultRegisterFlowUnit         = uint16(0x1438)
	DefaultRegisterFlowDecimalPoint = uint16(0x1439)
)

// FlowSensor implementiert einen Durchflusssensor
type FlowSensor struct {
	*sensor.BaseSensor
}

// NewFlowSensor erstellt einen neuen Durchflusssensor
func NewFlowSensor(id, name string) *FlowSensor {
	base := sensor.NewBaseSensor(id, name, types.ReadingTypeFlow, types.ReadingTypeCustom)

	return &FlowSensor{
		BaseSensor: base,
	}
}

// Read liest die Durchflussdaten vom Sensor
func (s *FlowSensor) Read(ctx context.Context) (types.Reading, error) {
	protocol := s.GetProtocol()
	if protocol == nil {
		return types.Reading{}, fmt.Errorf("kein Protokoll-Handler konfiguriert")
	}

	// Konfiguration für Flow-Rate-Register
	flowRateCfg := protocol.GetRegisterConfig(RegisterFlowRate)
	if flowRateCfg.Address == 0 {
		// Fallback auf Standard-Adresse
		flowRateCfg.Address = DefaultRegisterFlowRate
		flowRateCfg.Length = 1
	}

	// Gesamtfluss Low-Register
	totalFlowLowCfg := protocol.GetRegisterConfig(RegisterTotalFlowLow)
	if totalFlowLowCfg.Address == 0 {
		// Fallback auf Standard-Adresse
		totalFlowLowCfg.Address = DefaultRegisterTotalFlowLow
		totalFlowLowCfg.Length = 1
	}

	// Gesamtfluss High-Register
	totalFlowHighCfg := protocol.GetRegisterConfig(RegisterTotalFlowHigh)
	if totalFlowHighCfg.Address == 0 {
		// Fallback auf Standard-Adresse
		totalFlowHighCfg.Address = DefaultRegisterTotalFlowHigh
		totalFlowHighCfg.Length = 1
	}

	// Flow-Unit-Register
	flowUnitCfg := protocol.GetRegisterConfig(RegisterFlowUnit)
	if flowUnitCfg.Address == 0 {
		// Fallback auf Standard-Adresse
		flowUnitCfg.Address = DefaultRegisterFlowUnit
		flowUnitCfg.Length = 1
	}

	// Flow-Decimal-Point-Register
	flowDecimalPointCfg := protocol.GetRegisterConfig(RegisterFlowDecimalPoint)
	if flowDecimalPointCfg.Address == 0 {
		// Fallback auf Standard-Adresse
		flowDecimalPointCfg.Address = DefaultRegisterFlowDecimalPoint
		flowDecimalPointCfg.Length = 1
	}

	// 1. Flow Rate lesen
	flowRateData, err := protocol.ReadRegister(ctx, flowRateCfg.Address, flowRateCfg.Length)
	if err != nil {
		return types.Reading{}, fmt.Errorf("fehler beim Lesen der Flow-Rate: %w", err)
	}
	var flowRate float64
	if flowRateCfg.Length == 1 && len(flowRateData) >= 2 {
		flowRate = float64(uint16(flowRateData[0])<<8 | uint16(flowRateData[1]))
	} else {
		// Komplexere Konvertierung wenn nötig
		var tmpValue interface{}
		tmpValue, err = convertRawToValue(flowRateData, flowRateCfg.DataType, flowRateCfg.ByteOrder)
		if err != nil {
			return types.Reading{}, fmt.Errorf("fehler bei der Konvertierung der Flow-Rate: %w", err)
		}
		// Typ-Umwandlung nach float64
		flowRate, _ = tmpValue.(float64)
	}

	// 2. Total Flow Low lesen
	totalFlowLowData, err := protocol.ReadRegister(ctx, totalFlowLowCfg.Address, totalFlowLowCfg.Length)
	if err != nil {
		return types.Reading{}, fmt.Errorf("fehler beim Lesen des Total-Flow-Low: %w", err)
	}
	var totalFlowLow uint16
	if totalFlowLowCfg.Length == 1 && len(totalFlowLowData) >= 2 {
		totalFlowLow = uint16(totalFlowLowData[0])<<8 | uint16(totalFlowLowData[1])
	} else {
		log.Printf("Warnung: Unerwartetes Datenformat für Total-Flow-Low")
		totalFlowLow = 0
	}

	// Pause zwischen den Lesevorgängen
	time.Sleep(100 * time.Millisecond)

	// 3. Total Flow High lesen
	totalFlowHighData, err := protocol.ReadRegister(ctx, totalFlowHighCfg.Address, totalFlowHighCfg.Length)
	if err != nil {
		return types.Reading{}, fmt.Errorf("fehler beim Lesen des Total-Flow-High: %w", err)
	}
	var totalFlowHigh uint16
	if totalFlowHighCfg.Length == 1 && len(totalFlowHighData) >= 2 {
		totalFlowHigh = uint16(totalFlowHighData[0])<<8 | uint16(totalFlowHighData[1])
	} else {
		log.Printf("Warnung: Unerwartetes Datenformat für Total-Flow-High")
		totalFlowHigh = 0
	}

	// 4. Flow Unit lesen
	flowUnitData, err := protocol.ReadRegister(ctx, flowUnitCfg.Address, flowUnitCfg.Length)
	if err != nil {
		log.Printf("Warnung: Fehler beim Lesen der Flow-Unit: %v, verwende Standard", err)
		// Default-Wert verwenden
		flowUnitData = []byte{0, 0}
	}
	var flowUnit uint16
	if flowUnitCfg.Length == 1 && len(flowUnitData) >= 2 {
		flowUnit = uint16(flowUnitData[0])<<8 | uint16(flowUnitData[1])
	} else {
		flowUnit = 0 // Standard: m³
	}

	// 5. Flow Decimal Point lesen
	flowDecimalPointData, err := protocol.ReadRegister(ctx, flowDecimalPointCfg.Address, flowDecimalPointCfg.Length)
	if err != nil {
		log.Printf("Warnung: Fehler beim Lesen des Flow-Decimal-Point: %v, verwende Standard", err)
		// Default-Wert verwenden
		flowDecimalPointData = []byte{0, 3}
	}
	var flowDecimalPoint uint16
	if flowDecimalPointCfg.Length == 1 && len(flowDecimalPointData) >= 2 {
		flowDecimalPoint = uint16(flowDecimalPointData[0])<<8 | uint16(flowDecimalPointData[1])
	} else {
		flowDecimalPoint = 3 // Standard-Dezimalpunkt
	}

	// Werte validieren und Standardwerte verwenden, falls nötig
	if flowUnit == 0xFFFF {
		flowUnit = 0 // Default: m³
	}

	if flowDecimalPoint == 0xFFFF {
		flowDecimalPoint = 3 // Default decimal point
	}

	// Gesamtfluss berechnen
	flowInteger := (uint32(totalFlowHigh) << 16) | uint32(totalFlowLow)
	multiplier := math.Pow(10, float64(flowDecimalPoint)-3)
	totalFlow := float64(flowInteger) * multiplier

	// Flow Unit zu String mappen
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

	// Kalibrierungsdaten abrufen
	calibration := s.GetCalibration()
	offset, _ := getFloatFromMap(calibration, CalibrationOffset, 0.0)
	scale, _ := getFloatFromMap(calibration, CalibrationScale, 1.0)

	// Kalibrierung anwenden
	flowRate = flowRate*scale + offset

	// Reading-Objekt erstellen
	reading := types.NewReading(types.ReadingTypeFlow, flowRate, flowUnitStr, nil)

	// Zusätzliche Metadaten hinzufügen
	reading.Metadata["total_flow"] = totalFlow
	reading.Metadata["total_flow_low"] = totalFlowLow
	reading.Metadata["total_flow_high"] = totalFlowHigh
	reading.Metadata["flow_decimal_point"] = flowDecimalPoint

	return reading, nil
}

// ReadRaw liest die Rohdaten vom Durchflusssensor
func (s *FlowSensor) ReadRaw(ctx context.Context) ([]byte, error) {
	protocol := s.GetProtocol()
	if protocol == nil {
		return nil, fmt.Errorf("kein Protokoll-Handler konfiguriert")
	}

	// Konfiguration für das Flow-Rate-Register abrufen
	registerConfig := protocol.GetRegisterConfig(RegisterFlowRate)
	if registerConfig.Address == 0 {
		registerConfig.Address = DefaultRegisterFlowRate
		registerConfig.Length = 1
	}

	// Rohdaten vom Register lesen
	return protocol.ReadRegister(ctx, registerConfig.Address, registerConfig.Length)
}

// SetCalibration setzt neue Kalibrierungsparameter für den Durchflusssensor
func (s *FlowSensor) SetCalibration(calibration map[string]interface{}) error {
	// Überprüfen, ob erforderliche Kalibrierungsparameter vorhanden sind
	if _, ok := calibration[CalibrationOffset]; !ok {
		calibration[CalibrationOffset] = 0.0
	}

	if _, ok := calibration[CalibrationScale]; !ok {
		calibration[CalibrationScale] = 1.0
	}

	// Kalibrierung auf den BaseSensor anwenden
	return s.BaseSensor.SetCalibration(calibration)
}

// Hilfsfunktion zum Abrufen eines Float-Werts aus einer Map
func getFloatFromMap(m map[string]interface{}, key string, defaultValue float64) (float64, bool) {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case float64:
			return v, true
		case float32:
			return float64(v), true
		case int:
			return float64(v), true
		}
	}
	return defaultValue, false
}

// convertRawToValue konvertiert Rohdaten in einen Wert basierend auf dem Datentyp
func convertRawToValue(rawData []byte, dataType, byteOrder string) (interface{}, error) {
	// Vereinfachte Version, könnte durch eine allgemeinere Implementierung ersetzt werden
	if len(rawData) < 2 {
		return 0, fmt.Errorf("nicht genügend Daten zur Konvertierung")
	}

	// Für den einfachen Fall:
	return float64(uint16(rawData[0])<<8 | uint16(rawData[1])), nil
}
