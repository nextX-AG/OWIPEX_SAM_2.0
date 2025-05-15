package radar

import (
	"fmt"

	"owipex_reader/internal/protocol/factory"
	"owipex_reader/internal/types"
)

// CreateRadarSensor erstellt einen Radar-Sensor aus einer Konfiguration
func CreateRadarSensor(config types.DeviceConfig) (types.Sensor, error) {
	// Neuen Radar-Sensor erstellen
	sensor := NewRadarSensor(config.ID, config.Name)

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

	// Falls Container-Konfiguration vorhanden ist, sollte diese
	// durch die Metadaten im Sensor direkt verfügbar sein

	return sensor, nil
}

// createProtocolHandler erstellt einen Protokoll-Handler basierend auf der Konfiguration
func createProtocolHandler(config map[string]interface{}) (types.ProtocolHandler, error) {
	// TODO: Implementierung für verschiedene Protokolltypen
	// Diese Funktion sollte in einem gemeinsamen Paket implementiert werden

	return nil, fmt.Errorf("protokoll-Handler-Erstellung noch nicht implementiert")
}
