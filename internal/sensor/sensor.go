package sensor

import (
	"time"

	"owipex_reader/internal/modbus"
)

// Sensor is an interface for different types of sensors.
type Sensor interface {
	GetID() string
	GetType() string
	GetDeviceID() uint8 // Modbus Slave ID
	ReadData(client *modbus.Client) (map[string]interface{}, error)
	GetReadInterval() time.Duration
	GetLastReadTime() time.Time
	SetLastReadTime(t time.Time)
}

// BaseSensor provides common fields and methods for sensors.
type BaseSensor struct {
	ID            string
	Type          string
	DeviceID      uint8 // Modbus Slave ID
	ModbusClient  *modbus.Client
	ReadInterval  time.Duration
	LastReadTime  time.Time
	Config        map[string]interface{} // Sensor-specific configuration
}

// GetID returns the sensor's unique identifier.
func (s *BaseSensor) GetID() string {
	return s.ID
}

// GetType returns the sensor's type.
func (s *BaseSensor) GetType() string {
	return s.Type
}

// GetDeviceID returns the Modbus slave ID of the sensor.
func (s *BaseSensor) GetDeviceID() uint8 {
	return s.DeviceID
}

// GetReadInterval returns the configured read interval for the sensor.
func (s *BaseSensor) GetReadInterval() time.Duration {
	return s.ReadInterval
}

// GetLastReadTime returns the time the sensor was last read.
func (s *BaseSensor) GetLastReadTime() time.Time {
	return s.LastReadTime
}

// SetLastReadTime updates the time the sensor was last read.
func (s *BaseSensor) SetLastReadTime(t time.Time) {
	s.LastReadTime = t
}

// Placeholder for specific sensor implementations (e.g., PHSensor, TurbiditySensor)
// These will embed BaseSensor and implement the ReadData method.

