// Package radar implementiert einen Radar-Sensor zur Abstandsmessung.
package radar

import (
	"context"
	"fmt"

	"owipex_reader/internal/device/sensor"
	"owipex_reader/internal/types"
)

// Konstanten für Radar-Sensoren
const (
	// Register-Namen
	RegisterAirDistance = "air_distance"

	// Container-Konfiguration
	ConfigWidthMM             = "width_mm"
	ConfigLengthMM            = "length_mm"
	ConfigMaxVolumeM3         = "max_volume_m3"
	ConfigAirDistanceMaxLevel = "air_distance_max_level_mm"
	ConfigMaxWaterLevel       = "max_water_level_mm"
	ConfigNormalWaterLevel    = "normal_water_level_mm"

	// Default Register-Adresse
	DefaultRegisterAirDistance = uint16(0x0001)
)

// ContainerConfig enthält die Konfigurationsdaten des Behälters
type ContainerConfig struct {
	WidthMM             float64
	LengthMM            float64
	MaxVolumeM3         float64
	AirDistanceMaxLevel float64
	MaxWaterLevel       float64
	NormalWaterLevel    float64
}

// RadarSensor implementiert einen Radar-Füllstandsensor
type RadarSensor struct {
	*sensor.BaseSensor
	containerConfig ContainerConfig
}

// NewRadarSensor erstellt einen neuen Radar-Sensor
func NewRadarSensor(id, name string) *RadarSensor {
	base := sensor.NewBaseSensor(id, name, types.ReadingTypeLevel, types.ReadingTypeCustom)

	// Standardwerte für Container-Konfiguration
	container := ContainerConfig{
		WidthMM:             2500, // 2,5 Meter
		LengthMM:            4000, // 4 Meter
		MaxVolumeM3:         15,   // 15 Kubikmeter
		AirDistanceMaxLevel: 5500, // 5,5 Meter
		MaxWaterLevel:       1500, // 1,5 Meter
		NormalWaterLevel:    800,  // 0,8 Meter
	}

	return &RadarSensor{
		BaseSensor:      base,
		containerConfig: container,
	}
}

// Read liest Daten vom Radar-Sensor
func (s *RadarSensor) Read(ctx context.Context) (types.Reading, error) {
	protocol := s.GetProtocol()
	if protocol == nil {
		return types.Reading{}, fmt.Errorf("kein Protokoll-Handler konfiguriert")
	}

	// Metadaten aus der Sensorkonfiguration lesen
	s.updateContainerConfigFromMetadata()

	// Konfiguration für Luftabstands-Register abrufen
	registerConfig := protocol.GetRegisterConfig(RegisterAirDistance)
	if registerConfig.Address == 0 {
		// Fallback auf Standard-Adresse
		registerConfig.Address = DefaultRegisterAirDistance
		registerConfig.Length = 1
	}

	// Luftabstand vom Register lesen
	rawData, err := protocol.ReadRegister(ctx, registerConfig.Address, registerConfig.Length)
	if err != nil {
		return types.Reading{}, fmt.Errorf("fehler beim Lesen des Luftabstands: %w", err)
	}

	// Rohdaten in Luftabstand umwandeln
	var measuredAirDistance float64
	if len(rawData) >= 2 {
		measuredAirDistance = float64(uint16(rawData[0])<<8 | uint16(rawData[1]))
	} else {
		return types.Reading{}, fmt.Errorf("ungültiges Datenformat für Luftabstand")
	}

	// Berechnete Werte
	actualWaterLevel := calculateWaterLevel(measuredAirDistance, s.containerConfig.AirDistanceMaxLevel)
	actualVolume := calculateVolume(actualWaterLevel, s.containerConfig.WidthMM, s.containerConfig.LengthMM, s.containerConfig.MaxWaterLevel)
	volumePercentage := calculateVolumePercentage(actualWaterLevel, s.containerConfig.MaxWaterLevel)
	levelAboveNormal := calculateLevelAboveNormal(actualWaterLevel, s.containerConfig.NormalWaterLevel)
	waterLevelAlarm := checkWaterLevelAlarm(actualWaterLevel, s.containerConfig.MaxWaterLevel)

	// Reading-Objekt erstellen mit Wasserstand als Hauptwert
	reading := types.NewReading(types.ReadingTypeLevel, actualWaterLevel, "mm", rawData)

	// Zusätzliche Metadaten hinzufügen
	reading.Metadata["measured_air_distance"] = measuredAirDistance
	reading.Metadata["actual_volume"] = actualVolume
	reading.Metadata["volume_percentage"] = volumePercentage
	reading.Metadata["level_above_normal"] = levelAboveNormal
	reading.Metadata["water_level_alarm"] = waterLevelAlarm
	reading.Metadata["distance_m"] = measuredAirDistance / 1000 // in Metern

	return reading, nil
}

// ReadRaw liest die Rohdaten vom Radar-Sensor
func (s *RadarSensor) ReadRaw(ctx context.Context) ([]byte, error) {
	protocol := s.GetProtocol()
	if protocol == nil {
		return nil, fmt.Errorf("kein Protokoll-Handler konfiguriert")
	}

	// Konfiguration für Luftabstands-Register abrufen
	registerConfig := protocol.GetRegisterConfig(RegisterAirDistance)
	if registerConfig.Address == 0 {
		registerConfig.Address = DefaultRegisterAirDistance
		registerConfig.Length = 1
	}

	// Rohdaten vom Register lesen
	return protocol.ReadRegister(ctx, registerConfig.Address, registerConfig.Length)
}

// SetCalibration setzt neue Kalibrierungsparameter für den Radar-Sensor
func (s *RadarSensor) SetCalibration(calibration map[string]interface{}) error {
	// Bei Radar-Sensoren ist die "Kalibrierung" eigentlich die Container-Konfiguration
	return s.BaseSensor.SetCalibration(calibration)
}

// updateContainerConfigFromMetadata aktualisiert die Container-Konfiguration aus den Metadaten
func (s *RadarSensor) updateContainerConfigFromMetadata() {
	metadata := s.Metadata()

	// Container-Konfiguration aus den Metadaten extrahieren
	if containerConfig, ok := metadata["container_config"].(map[string]interface{}); ok {
		if width, ok := containerConfig[ConfigWidthMM].(float64); ok {
			s.containerConfig.WidthMM = width
		}
		if length, ok := containerConfig[ConfigLengthMM].(float64); ok {
			s.containerConfig.LengthMM = length
		}
		if maxVolume, ok := containerConfig[ConfigMaxVolumeM3].(float64); ok {
			s.containerConfig.MaxVolumeM3 = maxVolume
		}
		if airDistanceMax, ok := containerConfig[ConfigAirDistanceMaxLevel].(float64); ok {
			s.containerConfig.AirDistanceMaxLevel = airDistanceMax
		}
		if maxWaterLevel, ok := containerConfig[ConfigMaxWaterLevel].(float64); ok {
			s.containerConfig.MaxWaterLevel = maxWaterLevel
		}
		if normalWaterLevel, ok := containerConfig[ConfigNormalWaterLevel].(float64); ok {
			s.containerConfig.NormalWaterLevel = normalWaterLevel
		}
	}
}

// Hilfsfunktionen für Berechnungen (basierend auf der alten Implementierung)

func calculateWaterLevel(measuredAirDistance, airDistanceMaxLevel float64) float64 {
	if airDistanceMaxLevel == 0 {
		airDistanceMaxLevel = 5500 // Standardwert
	}

	waterLevel := airDistanceMaxLevel - measuredAirDistance
	if waterLevel < 0 {
		waterLevel = 0
	}

	return waterLevel
}

func calculateVolume(waterLevel, width, length, maxWaterLevel float64) float64 {
	// Standardwerte falls nicht angegeben
	if width == 0 {
		width = 2500 // Standardbreite in mm
	}
	if length == 0 {
		length = 4000 // Standardlänge in mm
	}
	if maxWaterLevel == 0 {
		maxWaterLevel = 1500 // Standard max. Wasserstand in mm
	}

	// Volumen in Kubikmetern berechnen
	volumeM3 := (waterLevel * width * length) / 1000000000 // mm³ zu m³

	if volumeM3 < 0 {
		volumeM3 = 0
	}

	return volumeM3
}

func calculateVolumePercentage(waterLevel, maxWaterLevel float64) float64 {
	if maxWaterLevel == 0 {
		maxWaterLevel = 1500 // Standard
	}

	percentage := (waterLevel / maxWaterLevel) * 100

	if percentage < 0 {
		percentage = 0
	} else if percentage > 100 {
		percentage = 100
	}

	return percentage
}

func calculateLevelAboveNormal(waterLevel, normalWaterLevel float64) float64 {
	if normalWaterLevel == 0 {
		normalWaterLevel = 800 // Standard
	}

	return waterLevel - normalWaterLevel
}

func checkWaterLevelAlarm(waterLevel, maxWaterLevel float64) bool {
	if maxWaterLevel == 0 {
		maxWaterLevel = 1500 // Standard
	}

	// Alarm bei 90% des max. Pegels
	alarmThreshold := maxWaterLevel * 0.9

	return waterLevel >= alarmThreshold
}
