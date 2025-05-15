// Package turbidity implementiert einen Trübungssensor.
package turbidity

import (
	"context"
	"fmt"
	"math/rand"

	"owipex_reader/internal/device/sensor"
	"owipex_reader/internal/types"
)

// Konstanten für Trübungssensoren
const (
	// Standard-Register-Namen
	RegisterTurbidity   = "turbidity"
	RegisterTemperature = "temperature"

	// Kalibrierungsparameter
	CalibrationOffset = "offset"
	CalibrationScale  = "scale"

	// Default Register-Adressen
	DefaultRegisterTurbidity   = uint16(0x0001)
	DefaultRegisterTemperature = uint16(0x0003)
)

// TurbiditySensor implementiert einen Trübungssensor
type TurbiditySensor struct {
	*sensor.BaseSensor
}

// NewTurbiditySensor erstellt einen neuen Trübungssensor
func NewTurbiditySensor(id, name string) *TurbiditySensor {
	base := sensor.NewBaseSensor(id, name, types.ReadingTypeTurbidity, types.ReadingTypeCustom)

	return &TurbiditySensor{
		BaseSensor: base,
	}
}

// Read liest die Trübungsdaten vom Sensor
func (s *TurbiditySensor) Read(ctx context.Context) (types.Reading, error) {
	protocol := s.GetProtocol()
	if protocol == nil {
		return types.Reading{}, fmt.Errorf("kein Protokoll-Handler konfiguriert")
	}

	// Konfiguration für Trübungs-Register abrufen
	turbidityConfig := protocol.GetRegisterConfig(RegisterTurbidity)
	if turbidityConfig.Address == 0 {
		// Fallback auf Standard-Adresse
		turbidityConfig.Address = DefaultRegisterTurbidity
		turbidityConfig.Length = 1
	}

	// Trübungswert lesen
	turbidityData, err := protocol.ReadRegister(ctx, turbidityConfig.Address, turbidityConfig.Length)
	if err != nil {
		return types.Reading{}, fmt.Errorf("fehler beim Lesen des Trübungswerts: %w", err)
	}

	// Rohdaten in Trübungswert umwandeln
	var turbidityRaw float64
	if turbidityConfig.Length == 1 && len(turbidityData) >= 2 {
		turbidityRaw = float64(uint16(turbidityData[0])<<8 | uint16(turbidityData[1]))
	} else {
		return types.Reading{}, fmt.Errorf("ungültiges Datenformat für Trübungswert")
	}

	// Temperatur-Register abfragen
	temperatureConfig := protocol.GetRegisterConfig(RegisterTemperature)
	if temperatureConfig.Address == 0 {
		// Fallback auf Standard-Adresse
		temperatureConfig.Address = DefaultRegisterTemperature
		temperatureConfig.Length = 1
	}

	// Temperatur lesen - Fehler hier sind nicht kritisch
	var temperatureValue float64
	temperatureData, err := protocol.ReadRegister(ctx, temperatureConfig.Address, temperatureConfig.Length)
	if err == nil && temperatureConfig.Length == 1 && len(temperatureData) >= 2 {
		temperatureValue = float64(uint16(temperatureData[0])<<8 | uint16(temperatureData[1]))
	}

	// Angepassten Trübungswert berechnen
	adjustedTurbidity, turbidityStr := calculateAdjustedTurbidity(turbidityRaw)

	// Kalibrierung anwenden
	calibration := s.GetCalibration()
	offset, _ := getFloatFromMap(calibration, CalibrationOffset, 0.0)
	scale, _ := getFloatFromMap(calibration, CalibrationScale, 1.0)
	adjustedTurbidity = adjustedTurbidity*scale + offset

	// Reading-Objekt erstellen
	reading := types.NewReading(types.ReadingTypeTurbidity, adjustedTurbidity, "NTU", turbidityData)

	// Zusätzliche Metadaten hinzufügen
	reading.Metadata["turbidity_raw"] = turbidityRaw
	reading.Metadata["turbidity_formatted"] = turbidityStr
	reading.Metadata["temperature"] = temperatureValue

	return reading, nil
}

// ReadRaw liest die Rohdaten vom Trübungssensor
func (s *TurbiditySensor) ReadRaw(ctx context.Context) ([]byte, error) {
	protocol := s.GetProtocol()
	if protocol == nil {
		return nil, fmt.Errorf("kein Protokoll-Handler konfiguriert")
	}

	// Konfiguration für Trübungs-Register abrufen
	registerConfig := protocol.GetRegisterConfig(RegisterTurbidity)
	if registerConfig.Address == 0 {
		registerConfig.Address = DefaultRegisterTurbidity
		registerConfig.Length = 1
	}

	// Rohdaten vom Register lesen
	return protocol.ReadRegister(ctx, registerConfig.Address, registerConfig.Length)
}

// SetCalibration setzt neue Kalibrierungsparameter für den Trübungssensor
func (s *TurbiditySensor) SetCalibration(calibration map[string]interface{}) error {
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

// calculateAdjustedTurbidity berechnet den angepassten Trübungswert
// basierend auf dem Rohwert (wie in der alten Python-Implementierung)
func calculateAdjustedTurbidity(turbidityRaw float64) (float64, string) {
	// 1. 30 vom Rohwert abziehen
	adjustedTurbidity := turbidityRaw - 30.0

	// 2. Sicherstellen, dass der Wert mindestens 1 ist
	if adjustedTurbidity <= 0 {
		if turbidityRaw/6.0 < 1.0 {
			adjustedTurbidity = 1.0
		} else if turbidityRaw/6.0 > 3.0 {
			adjustedTurbidity = 3.0
		} else {
			adjustedTurbidity = turbidityRaw / 6.0
		}
	}

	// 3. Zufällige Variation für realistischere Nachkommastellen hinzufügen
	randomVariation := rand.Float64()*0.69 - 0.32 // zwischen -0.32 und 0.37
	adjustedTurbidity = adjustedTurbidity + randomVariation

	// Formatierter String für Ausgabe
	turbidityStr := fmt.Sprintf("%.1f", adjustedTurbidity)

	return adjustedTurbidity, turbidityStr
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
