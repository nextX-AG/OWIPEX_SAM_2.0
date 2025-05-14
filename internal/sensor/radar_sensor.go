package sensor

import (
	"fmt"
	"time"

	"owipex_reader/internal/modbus"
)

// RadarSensor represents a radar sensor.
type RadarSensor struct {
	BaseSensor
	// Container configuration for calculations
	ContainerConfig struct {
		WidthMM             float64 `json:"width_mm"`
		LengthMM            float64 `json:"length_mm"`
		MaxVolumeMm3        float64 `json:"max_volume_m3"`
		AirDistanceMaxLevel float64 `json:"air_distance_max_level_mm"`
		MaxWaterLevel       float64 `json:"max_water_level_mm"`
		NormalWaterLevel    float64 `json:"normal_water_level_mm"`
	}
}

// NewRadarSensor creates a new RadarSensor instance.
func NewRadarSensor(id string, deviceID uint8, client *modbus.Client, config map[string]interface{}) *RadarSensor {
	readIntervalSeconds, ok := config["read_interval_seconds"].(float64) // JSON numbers are float64
	if !ok {
		readIntervalSeconds = 20 // Default interval for radar sensor
	}

	// Create the radar sensor
	rs := &RadarSensor{
		BaseSensor: BaseSensor{
			ID:           id,
			Type:         "radar",
			DeviceID:     deviceID,
			ModbusClient: client,
			ReadInterval: time.Duration(readIntervalSeconds) * time.Second,
			Config:       config,
		},
	}

	// Try to get container config from the metadata
	if metadata, ok := config["metadata"].(map[string]interface{}); ok {
		if containerConfig, ok := metadata["container_config"].(map[string]interface{}); ok {
			if width, ok := containerConfig["width_mm"].(float64); ok {
				rs.ContainerConfig.WidthMM = width
			}
			if length, ok := containerConfig["length_mm"].(float64); ok {
				rs.ContainerConfig.LengthMM = length
			}
			if maxVolume, ok := containerConfig["max_volume_m3"].(float64); ok {
				rs.ContainerConfig.MaxVolumeMm3 = maxVolume
			}
			if airDistanceMax, ok := containerConfig["air_distance_max_level_mm"].(float64); ok {
				rs.ContainerConfig.AirDistanceMaxLevel = airDistanceMax
			}
			if maxWaterLevel, ok := containerConfig["max_water_level_mm"].(float64); ok {
				rs.ContainerConfig.MaxWaterLevel = maxWaterLevel
			}
			if normalWaterLevel, ok := containerConfig["normal_water_level_mm"].(float64); ok {
				rs.ContainerConfig.NormalWaterLevel = normalWaterLevel
			}
		}
	}

	return rs
}

// ReadData reads data from the radar sensor.
// Implementation based on the Python radar_sensor.py
func (s *RadarSensor) ReadData(client *modbus.Client) (map[string]interface{}, error) {
	// Read measured air distance from Register 0x0001
	airDistanceRegs, err := client.ReadHoldingRegisters(s.DeviceID, 0x0001, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to read air distance for sensor %s: %w", s.ID, err)
	}

	measuredAirDistance := float64(airDistanceRegs[0])

	// Calculate derived values
	// Simplified implementation without the full calculation module
	// In a real implementation, we should create a proper calculator like in Python

	// Basic calculations
	actualWaterLevel := calculateWaterLevel(measuredAirDistance, s.ContainerConfig.AirDistanceMaxLevel)
	actualVolume := calculateVolume(actualWaterLevel, s.ContainerConfig.WidthMM, s.ContainerConfig.LengthMM, s.ContainerConfig.MaxWaterLevel)
	volumePercentage := calculateVolumePercentage(actualWaterLevel, s.ContainerConfig.MaxWaterLevel)
	levelAboveNormal := calculateLevelAboveNormal(actualWaterLevel, s.ContainerConfig.NormalWaterLevel)
	waterLevelAlarm := checkWaterLevelAlarm(actualWaterLevel, s.ContainerConfig.MaxWaterLevel)

	data := map[string]interface{}{
		"distance":              fmt.Sprintf("%.2f", measuredAirDistance/1000), // Convert to meters for display
		"measured_air_distance": fmt.Sprintf("%d", int(measuredAirDistance)),
		"actual_water_level":    fmt.Sprintf("%d", int(actualWaterLevel)),
		"actual_volume":         fmt.Sprintf("%.3f", actualVolume),
		"volume_percentage":     fmt.Sprintf("%.1f", volumePercentage),
		"level_above_normal":    fmt.Sprintf("%d", int(levelAboveNormal)),
		"water_level_alarm":     fmt.Sprintf("%t", waterLevelAlarm),
	}

	return data, nil
}

// Helper calculation functions (simplified versions of the Python calculations)

func calculateWaterLevel(measuredAirDistance, airDistanceMaxLevel float64) float64 {
	if airDistanceMaxLevel == 0 {
		airDistanceMaxLevel = 5500 // Default value
	}

	waterLevel := airDistanceMaxLevel - measuredAirDistance
	if waterLevel < 0 {
		waterLevel = 0
	}

	return waterLevel
}

func calculateVolume(waterLevel, width, length, maxWaterLevel float64) float64 {
	// Default values if not provided
	if width == 0 {
		width = 2500 // Default width in mm
	}
	if length == 0 {
		length = 4000 // Default length in mm
	}
	if maxWaterLevel == 0 {
		maxWaterLevel = 1500 // Default max water level in mm
	}

	// Calculate volume in cubic meters
	volumeM3 := (waterLevel * width * length) / 1000000000 // mm³ to m³

	if volumeM3 < 0 {
		volumeM3 = 0
	}

	return volumeM3
}

func calculateVolumePercentage(waterLevel, maxWaterLevel float64) float64 {
	if maxWaterLevel == 0 {
		maxWaterLevel = 1500 // Default
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
		normalWaterLevel = 800 // Default
	}

	return waterLevel - normalWaterLevel
}

func checkWaterLevelAlarm(waterLevel, maxWaterLevel float64) bool {
	if maxWaterLevel == 0 {
		maxWaterLevel = 1500 // Default
	}

	// Alarm at 90% of max level
	alarmThreshold := maxWaterLevel * 0.9

	return waterLevel >= alarmThreshold
}
