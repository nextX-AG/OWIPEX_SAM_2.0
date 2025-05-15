package creator

import (
	"owipex_reader/internal/device/sensor/flow"
	"owipex_reader/internal/device/sensor/ph"
	"owipex_reader/internal/device/sensor/radar"
	"owipex_reader/internal/device/sensor/turbidity"
)

// RegisterAllSensorTypes registriert alle verfügbaren Sensortypen in der Registry
func RegisterAllSensorTypes(registry *SensorRegistry) {
	// pH-Sensor registrieren
	registry.RegisterSensor("ph_sensor", ph.CreatePHSensor)

	// Durchflusssensor registrieren
	registry.RegisterSensor("flow_sensor", flow.CreateFlowSensor)

	// Radar-Sensor registrieren
	registry.RegisterSensor("radar_sensor", radar.CreateRadarSensor)

	// Trübungssensor registrieren
	registry.RegisterSensor("turbidity_sensor", turbidity.CreateTurbiditySensor)
}
