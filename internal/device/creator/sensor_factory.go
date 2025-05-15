// Package creator enthält Factory-Funktionen für die Erstellung von Geräten.
package creator

import (
	"fmt"

	"owipex_reader/internal/types"
)

// SensorCreator ist eine Funktion, die einen Sensor aus einer Konfiguration erstellt
type SensorCreator func(config types.DeviceConfig) (types.Sensor, error)

// SensorRegistry verwaltet die Registrierung und Erstellung von Sensoren
type SensorRegistry struct {
	creators map[string]SensorCreator
}

// NewSensorRegistry erstellt eine neue Sensor-Registry
func NewSensorRegistry() *SensorRegistry {
	return &SensorRegistry{
		creators: make(map[string]SensorCreator),
	}
}

// RegisterSensor registriert einen Creator für einen bestimmten Sensortyp
func (r *SensorRegistry) RegisterSensor(sensorType string, creator SensorCreator) {
	r.creators[sensorType] = creator
}

// CreateSensor erstellt einen Sensor basierend auf der Konfiguration
func (r *SensorRegistry) CreateSensor(config types.DeviceConfig) (types.Sensor, error) {
	creator, ok := r.creators[config.Type]
	if !ok {
		return nil, fmt.Errorf("kein Creator für Sensortyp '%s' registriert", config.Type)
	}

	return creator(config)
}

// CreateSensors erstellt mehrere Sensoren aus einem Array von Konfigurationen
func (r *SensorRegistry) CreateSensors(configs []types.DeviceConfig) ([]types.Sensor, []error) {
	var sensors []types.Sensor
	var errors []error

	for _, config := range configs {
		sensor, err := r.CreateSensor(config)
		if err != nil {
			errors = append(errors, fmt.Errorf("fehler beim Erstellen von Sensor '%s': %w", config.ID, err))
			continue
		}

		sensors = append(sensors, sensor)
	}

	return sensors, errors
}
