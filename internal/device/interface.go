// Package device definiert die grundlegenden Interfaces und Typen
// für alle Geräte (Sensoren und Aktoren) im System.
package device

import (
	"context"
	"time"
)

// DeviceType kategorisiert die Art des Geräts
type DeviceType string

const (
	// TypeSensor kennzeichnet Geräte, die hauptsächlich Daten messen
	TypeSensor DeviceType = "SENSOR"
	// TypeActor kennzeichnet Geräte, die hauptsächlich Aktionen ausführen
	TypeActor DeviceType = "ACTOR"
	// TypeHybrid kennzeichnet Geräte, die sowohl messen als auch Aktionen ausführen können
	TypeHybrid DeviceType = "HYBRID"
)

// ReadingType definiert die Art des Messwerts
type ReadingType string

const (
	ReadingTypePH        ReadingType = "PH"
	ReadingTypeFlow      ReadingType = "FLOW"
	ReadingTypeTurbidity ReadingType = "TURBIDITY"
	ReadingTypeLevel     ReadingType = "LEVEL"
	ReadingTypePosition  ReadingType = "POSITION"
	ReadingTypeState     ReadingType = "STATE"
	ReadingTypeCustom    ReadingType = "CUSTOM"
)

// ReadingQuality gibt die Qualität eines Messwerts an
type ReadingQuality string

const (
	QualityGood      ReadingQuality = "GOOD"
	QualityUncertain ReadingQuality = "UNCERTAIN"
	QualityBad       ReadingQuality = "BAD"
)

// CommandType definiert die Art des Steuerbefehls
type CommandType string

const (
	CommandTypeSetPosition CommandType = "SET_POSITION"
	CommandTypeSetState    CommandType = "SET_STATE"
	CommandTypeCalibrate   CommandType = "CALIBRATE"
	CommandTypeReset       CommandType = "RESET"
	CommandTypeCustom      CommandType = "CUSTOM"
)

// Device ist das Basisinterface für alle Geräte im System
type Device interface {
	// ID gibt die eindeutige Kennung des Geräts zurück
	ID() string

	// Name gibt den Anzeigenamen des Geräts zurück
	Name() string

	// Type gibt den Typ des Geräts zurück
	Type() DeviceType

	// Metadata gibt gerätespezifische Metadaten zurück
	Metadata() map[string]interface{}

	// IsEnabled prüft, ob das Gerät aktiviert ist
	IsEnabled() bool

	// Enable aktiviert oder deaktiviert das Gerät
	Enable(enabled bool)

	// Close gibt Ressourcen frei und beendet die Gerätekommunikation
	Close() error
}

// Reading repräsentiert einen Messwert mit Metadaten
type Reading struct {
	// Type gibt die Art des Messwerts an
	Type ReadingType

	// Value enthält den eigentlichen Messwert (typischerweise float64, int oder bool)
	Value interface{}

	// Unit ist die Einheit des Messwerts (z.B. "pH", "l/min", "cm")
	Unit string

	// Timestamp ist der Zeitpunkt der Messung (Unix-Millisekunden)
	Timestamp int64

	// RawValue enthält die unverarbeiteten Binärdaten, wie sie vom Gerät empfangen wurden
	RawValue []byte

	// Quality gibt die Qualität des Messwerts an
	Quality ReadingQuality

	// Metadata enthält zusätzliche messwertbezogene Informationen
	Metadata map[string]interface{}
}

// NewReading erstellt ein neues Reading mit aktuellem Zeitstempel
func NewReading(readingType ReadingType, value interface{}, unit string, rawValue []byte) Reading {
	return Reading{
		Type:      readingType,
		Value:     value,
		Unit:      unit,
		Timestamp: time.Now().UnixNano() / int64(time.Millisecond),
		RawValue:  rawValue,
		Quality:   QualityGood,
		Metadata:  make(map[string]interface{}),
	}
}

// Command repräsentiert einen Steuerbefehl für ein Gerät
type Command struct {
	// Type gibt die Art des Befehls an
	Type CommandType

	// Value enthält den Hauptwert des Befehls (z.B. Position, Zustand)
	Value interface{}

	// Parameters enthält zusätzliche Parameter für den Befehl
	Parameters map[string]interface{}
}

// NewCommand erstellt einen neuen Befehl
func NewCommand(commandType CommandType, value interface{}) Command {
	return Command{
		Type:       commandType,
		Value:      value,
		Parameters: make(map[string]interface{}),
	}
}

// ReadableDevice ist ein Gerät, das Messwerte liefern kann
type ReadableDevice interface {
	Device

	// Read liest einen Messwert und gibt ihn zurück
	Read(ctx context.Context) (Reading, error)

	// ReadRaw liest die Rohdaten vom Gerät
	ReadRaw(ctx context.Context) ([]byte, error)

	// AvailableReadings gibt die verfügbaren Messwerttypen zurück
	AvailableReadings() []ReadingType
}

// WritableDevice ist ein Gerät, das gesteuert werden kann
type WritableDevice interface {
	Device

	// Write sendet einen Steuerbefehl an das Gerät
	Write(ctx context.Context, command Command) error

	// WriteRaw sendet Rohdaten an das Gerät
	WriteRaw(ctx context.Context, data []byte) error

	// AvailableCommands gibt die verfügbaren Befehlstypen zurück
	AvailableCommands() []CommandType
}

// Sensor ist ein spezialisiertes Interface für Messgeräte
type Sensor interface {
	ReadableDevice

	// GetCalibration gibt die aktuelle Kalibrierung des Sensors zurück
	GetCalibration() map[string]interface{}

	// SetCalibration setzt neue Kalibrierungsparameter
	SetCalibration(calibration map[string]interface{}) error
}

// Actor ist ein spezialisiertes Interface für Steuergeräte
type Actor interface {
	WritableDevice

	// GetState gibt den aktuellen Zustand des Aktors zurück
	GetState() (interface{}, error)
}

// HybridDevice ist ein Gerät, das sowohl messen als auch steuern kann
type HybridDevice interface {
	ReadableDevice
	WritableDevice
}

// DeviceConfig enthält die grundlegende Konfiguration für ein Gerät
type DeviceConfig struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	Manufacturer string                 `json:"manufacturer"`
	Model        string                 `json:"model"`
	Protocol     string                 `json:"protocol"`
	Enabled      bool                   `json:"enabled"`
	Metadata     map[string]interface{} `json:"metadata"`
}
