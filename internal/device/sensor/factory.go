// Package sensor enthält die Implementierung verschiedener Sensortypen.
package sensor

import (
	"fmt"

	"owipex_reader/internal/device/sensor/flow"
	"owipex_reader/internal/device/sensor/ph"
	"owipex_reader/internal/device/sensor/radar"
	"owipex_reader/internal/device/sensor/turbidity"
	"owipex_reader/internal/types"
)

// SensorCreator ist eine Funktion, die einen Sensor aus einer Konfiguration erstellt
type SensorCreator func(config types.DeviceConfig) (types.Sensor, error)

// SensorFactory verwaltet die Erstellung verschiedener Sensortypen
type SensorFactory struct {
	creators map[string]SensorCreator
}

// NewSensorFactory erstellt eine neue Sensor-Factory
func NewSensorFactory() *SensorFactory {
	factory := &SensorFactory{
		creators: make(map[string]SensorCreator),
	}

	// Standard-Sensortypen registrieren
	RegisterStandardSensorTypes(factory)

	return factory
}

// RegisterCreator registriert einen Creator für einen bestimmten Sensortyp
func (f *SensorFactory) RegisterCreator(sensorType string, creator SensorCreator) {
	f.creators[sensorType] = creator
}

// CreateSensor erstellt einen Sensor basierend auf der Konfiguration
func (f *SensorFactory) CreateSensor(config types.DeviceConfig) (types.Sensor, error) {
	creator, ok := f.creators[config.Type]
	if !ok {
		return nil, fmt.Errorf("kein Creator für Sensortyp '%s' registriert", config.Type)
	}

	return creator(config)
}

// RegisterStandardSensorTypes registriert alle Standard-Sensortypen in der Factory
func RegisterStandardSensorTypes(factory *SensorFactory) {
	// pH-Sensor registrieren
	factory.RegisterCreator("ph_sensor", ph.CreatePHSensor)

	// Durchflusssensor registrieren
	factory.RegisterCreator("flow_sensor", flow.CreateFlowSensor)

	// Radar-Sensor registrieren
	factory.RegisterCreator("radar_sensor", radar.CreateRadarSensor)

	// Trübungssensor registrieren
	factory.RegisterCreator("turbidity_sensor", turbidity.CreateTurbiditySensor)
}
