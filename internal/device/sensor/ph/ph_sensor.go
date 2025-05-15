// Package ph implementiert einen pH-Sensor.
package ph

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"

	"owipex_reader/internal/device/sensor"
	"owipex_reader/internal/types"
)

// Konstanten für pH-Sensoren
const (
	// Standard-Register-Namen
	RegisterPHValue     = "ph_value"
	RegisterTemperature = "temperature"
	RegisterCalibration = "calibration"

	// Kalibrierungsparameter
	CalibrationOffset = "offset"
	CalibrationScale  = "scale"
)

// PHSensor implementiert einen pH-Wert-Sensor
type PHSensor struct {
	*sensor.BaseSensor
}

// NewPHSensor erstellt einen neuen pH-Sensor
func NewPHSensor(id, name string) *PHSensor {
	base := sensor.NewBaseSensor(id, name, types.ReadingTypePH, types.ReadingTypeCustom)

	return &PHSensor{
		BaseSensor: base,
	}
}

// Read liest den pH-Wert vom Sensor
func (s *PHSensor) Read(ctx context.Context) (types.Reading, error) {
	protocol := s.GetProtocol()
	if protocol == nil {
		return types.Reading{}, fmt.Errorf("kein Protokoll-Handler konfiguriert")
	}

	// Konfiguration für das pH-Wert-Register holen
	registerConfig := protocol.GetRegisterConfig(RegisterPHValue)
	if registerConfig.Address == 0 && registerConfig.Length == 0 {
		return types.Reading{}, fmt.Errorf("keine Konfiguration für pH-Wert-Register gefunden")
	}

	// Rohdaten vom Register lesen
	rawData, err := protocol.ReadRegister(ctx, registerConfig.Address, registerConfig.Length)
	if err != nil {
		return types.Reading{}, fmt.Errorf("fehler beim Lesen des pH-Wert-Registers: %w", err)
	}

	// Kalibrierungsdaten abrufen
	calibration := s.GetCalibration()
	offset, _ := getFloatFromMap(calibration, CalibrationOffset, 0.0)
	scale, _ := getFloatFromMap(calibration, CalibrationScale, 1.0)

	// Rohdaten in pH-Wert konvertieren
	phValue, err := convertRawToPH(rawData, registerConfig.DataType, registerConfig.ByteOrder, offset, scale)
	if err != nil {
		return types.Reading{}, fmt.Errorf("fehler bei der Konvertierung der pH-Wert-Daten: %w", err)
	}

	// Reading-Objekt erstellen
	reading := types.NewReading(types.ReadingTypePH, phValue, "pH", rawData)

	// Optional: Temperatur abrufen, wenn verfügbar
	if tempConfig := protocol.GetRegisterConfig(RegisterTemperature); tempConfig.Address != 0 {
		tempData, err := protocol.ReadRegister(ctx, tempConfig.Address, tempConfig.Length)
		if err == nil {
			tempValue, err := convertRawToFloat(tempData, tempConfig.DataType, tempConfig.ByteOrder)
			if err == nil {
				reading.Metadata["temperature"] = tempValue
			}
		}
	}

	return reading, nil
}

// ReadRaw liest die Rohdaten vom pH-Sensor
func (s *PHSensor) ReadRaw(ctx context.Context) ([]byte, error) {
	protocol := s.GetProtocol()
	if protocol == nil {
		return nil, fmt.Errorf("kein Protokoll-Handler konfiguriert")
	}

	// Konfiguration für das pH-Wert-Register abrufen
	registerConfig := protocol.GetRegisterConfig(RegisterPHValue)
	if registerConfig.Address == 0 && registerConfig.Length == 0 {
		return nil, fmt.Errorf("keine Konfiguration für pH-Wert-Register gefunden")
	}

	// Rohdaten vom Register lesen
	return protocol.ReadRegister(ctx, registerConfig.Address, registerConfig.Length)
}

// SetCalibration setzt neue Kalibrierungsparameter für den pH-Sensor
func (s *PHSensor) SetCalibration(calibration map[string]interface{}) error {
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

// convertRawToPH konvertiert Rohdaten in einen pH-Wert
func convertRawToPH(rawData []byte, dataType, byteOrder string, offset, scale float64) (float64, error) {
	// Rohdaten in Float konvertieren
	value, err := convertRawToFloat(rawData, dataType, byteOrder)
	if err != nil {
		return 0, err
	}

	// Kalibrierung anwenden
	phValue := value*scale + offset

	// pH-Werte auf den gültigen Bereich (0-14) begrenzen
	if phValue < 0 {
		phValue = 0
	} else if phValue > 14 {
		phValue = 14
	}

	return phValue, nil
}

// convertRawToFloat konvertiert Rohdaten in einen Float-Wert
func convertRawToFloat(rawData []byte, dataType, byteOrder string) (float64, error) {
	if len(rawData) == 0 {
		return 0, fmt.Errorf("keine Daten zum Konvertieren")
	}

	var value float64

	switch dataType {
	case "float32":
		if len(rawData) < 4 {
			return 0, fmt.Errorf("nicht genügend Daten für float32: %d Bytes", len(rawData))
		}
		var bits uint32
		if byteOrder == "big_endian" {
			bits = binary.BigEndian.Uint32(rawData)
		} else {
			bits = binary.LittleEndian.Uint32(rawData)
		}
		value = float64(math.Float32frombits(bits))

	case "float64":
		if len(rawData) < 8 {
			return 0, fmt.Errorf("nicht genügend Daten für float64: %d Bytes", len(rawData))
		}
		var bits uint64
		if byteOrder == "big_endian" {
			bits = binary.BigEndian.Uint64(rawData)
		} else {
			bits = binary.LittleEndian.Uint64(rawData)
		}
		value = math.Float64frombits(bits)

	case "int16":
		if len(rawData) < 2 {
			return 0, fmt.Errorf("nicht genügend Daten für int16: %d Bytes", len(rawData))
		}
		var bits uint16
		if byteOrder == "big_endian" {
			bits = binary.BigEndian.Uint16(rawData)
		} else {
			bits = binary.LittleEndian.Uint16(rawData)
		}
		value = float64(int16(bits))

	case "uint16":
		if len(rawData) < 2 {
			return 0, fmt.Errorf("nicht genügend Daten für uint16: %d Bytes", len(rawData))
		}
		var bits uint16
		if byteOrder == "big_endian" {
			bits = binary.BigEndian.Uint16(rawData)
		} else {
			bits = binary.LittleEndian.Uint16(rawData)
		}
		value = float64(bits)

	case "int32":
		if len(rawData) < 4 {
			return 0, fmt.Errorf("nicht genügend Daten für int32: %d Bytes", len(rawData))
		}
		var bits uint32
		if byteOrder == "big_endian" {
			bits = binary.BigEndian.Uint32(rawData)
		} else {
			bits = binary.LittleEndian.Uint32(rawData)
		}
		value = float64(int32(bits))

	case "uint32":
		if len(rawData) < 4 {
			return 0, fmt.Errorf("nicht genügend Daten für uint32: %d Bytes", len(rawData))
		}
		var bits uint32
		if byteOrder == "big_endian" {
			bits = binary.BigEndian.Uint32(rawData)
		} else {
			bits = binary.LittleEndian.Uint32(rawData)
		}
		value = float64(bits)

	default:
		return 0, fmt.Errorf("unbekannter Datentyp: %s", dataType)
	}

	return value, nil
}
