package turbidity

import (
	"fmt"

	"owipex_reader/internal/protocol/factory"
	"owipex_reader/internal/types"
)

// CreateTurbiditySensor erstellt einen Trübungssensor aus einer Konfiguration
func CreateTurbiditySensor(config types.DeviceConfig) (types.Sensor, error) {
	// Neuen Trübungssensor erstellen
	sensor := NewTurbiditySensor(config.ID, config.Name)

	// Protokoll-Handler konfigurieren
	if config.Protocol == "modbus" {
		if modbusConfig, ok := config.Metadata["modbus"].(map[string]interface{}); ok {
			protocol, err := factory.CreateProtocolHandler("modbus", modbusConfig)
			if err != nil {
				return nil, fmt.Errorf("fehler beim Erstellen des Protokoll-Handlers: %w", err)
			}
			sensor.BaseSensor.SetProtocol(protocol)
		}
	}

	// Kalibrierung setzen, falls vorhanden
	if calibration, ok := config.Metadata["calibration"].(map[string]interface{}); ok {
		if err := sensor.SetCalibration(calibration); err != nil {
			return nil, fmt.Errorf("fehler beim Setzen der Kalibrierung: %w", err)
		}
	}

	return sensor, nil
}

// createProtocolHandler erstellt einen Protokoll-Handler basierend auf der Konfiguration
func createProtocolHandler(config map[string]interface{}) (types.ProtocolHandler, error) {
	// TODO: Implementierung für verschiedene Protokolltypen
	// Diese Funktion sollte in einem gemeinsamen Paket implementiert werden

	return nil, fmt.Errorf("protokoll-Handler-Erstellung noch nicht implementiert")
}
