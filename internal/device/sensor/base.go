// Package sensor implementiert verschiedene Sensortypen.
package sensor

import (
	"context"
	"fmt"
	"sync"

	"owipex_reader/internal/types"
)

// BaseSensor ist eine grundlegende Implementierung eines Sensors,
// die von spezifischen Sensortypen erweitert werden kann.
type BaseSensor struct {
	id                string
	name              string
	enabled           bool
	metadata          map[string]interface{}
	availableReadings []types.ReadingType
	calibration       map[string]interface{}
	mutex             sync.RWMutex
	protocol          types.ProtocolHandler
}

// NewBaseSensor erstellt einen neuen BaseSensor
func NewBaseSensor(id, name string, readings ...types.ReadingType) *BaseSensor {
	return &BaseSensor{
		id:                id,
		name:              name,
		enabled:           true,
		metadata:          make(map[string]interface{}),
		availableReadings: readings,
		calibration:       make(map[string]interface{}),
	}
}

// ID gibt die eindeutige Kennung des Sensors zurück
func (s *BaseSensor) ID() string {
	return s.id
}

// Name gibt den Anzeigenamen des Sensors zurück
func (s *BaseSensor) Name() string {
	return s.name
}

// Type gibt den Typ des Geräts zurück
func (s *BaseSensor) Type() types.DeviceType {
	return types.TypeSensor
}

// Metadata gibt sensorspezifische Metadaten zurück
func (s *BaseSensor) Metadata() map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Kopie der Metadaten erstellen
	metadata := make(map[string]interface{}, len(s.metadata))
	for k, v := range s.metadata {
		metadata[k] = v
	}

	return metadata
}

// IsEnabled prüft, ob der Sensor aktiviert ist
func (s *BaseSensor) IsEnabled() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.enabled
}

// Enable aktiviert oder deaktiviert den Sensor
func (s *BaseSensor) Enable(enabled bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.enabled = enabled
}

// SetMetadata setzt einen Metadaten-Wert
func (s *BaseSensor) SetMetadata(key string, value interface{}) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.metadata[key] = value
}

// GetMetadata gibt einen Metadaten-Wert zurück
func (s *BaseSensor) GetMetadata(key string) (interface{}, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	value, exists := s.metadata[key]
	return value, exists
}

// SetProtocol setzt den Protocol-Handler für die Kommunikation
func (s *BaseSensor) SetProtocol(protocol types.ProtocolHandler) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.protocol = protocol
}

// GetProtocol gibt den aktuellen Protocol-Handler zurück
func (s *BaseSensor) GetProtocol() types.ProtocolHandler {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.protocol
}

// ReadRaw liest die Rohdaten vom Sensor
func (s *BaseSensor) ReadRaw(ctx context.Context) ([]byte, error) {
	s.mutex.RLock()
	protocol := s.protocol
	enabled := s.enabled
	s.mutex.RUnlock()

	if !enabled {
		return nil, fmt.Errorf("sensor %s ist deaktiviert", s.id)
	}

	if protocol == nil {
		return nil, fmt.Errorf("kein Protokoll-Handler für Sensor %s konfiguriert", s.id)
	}

	// Diese Basisimplementierung muss von konkreten Sensoren überschrieben werden
	return nil, fmt.Errorf("ReadRaw muss von spezifischen Sensorimplementierungen überschrieben werden")
}

// Read liest einen Messwert und gibt ihn zurück
func (s *BaseSensor) Read(ctx context.Context) (types.Reading, error) {
	// Diese Basisimplementierung muss von konkreten Sensoren überschrieben werden
	return types.Reading{}, fmt.Errorf("Read muss von spezifischen Sensorimplementierungen überschrieben werden")
}

// AvailableReadings gibt die verfügbaren Messwerttypen zurück
func (s *BaseSensor) AvailableReadings() []types.ReadingType {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Kopie erstellen, um Änderungen zu verhindern
	readings := make([]types.ReadingType, len(s.availableReadings))
	copy(readings, s.availableReadings)

	return readings
}

// GetCalibration gibt die aktuelle Kalibrierung des Sensors zurück
func (s *BaseSensor) GetCalibration() map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Kopie der Kalibrierung erstellen
	calibration := make(map[string]interface{}, len(s.calibration))
	for k, v := range s.calibration {
		calibration[k] = v
	}

	return calibration
}

// SetCalibration setzt neue Kalibrierungsparameter
func (s *BaseSensor) SetCalibration(calibration map[string]interface{}) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Kalibrierungsdaten validieren (kann von speziellen Sensortypen überschrieben werden)

	// Neue Kalibrierung setzen
	s.calibration = make(map[string]interface{})
	for k, v := range calibration {
		s.calibration[k] = v
	}

	return nil
}

// Close gibt Ressourcen frei und beendet die Sensorkommunikation
func (s *BaseSensor) Close() error {
	s.mutex.Lock()
	protocol := s.protocol
	s.mutex.Unlock()

	if protocol != nil {
		return protocol.Close()
	}
	return nil
}
