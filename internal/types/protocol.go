// Package types enthält zentrale Typen und Interfaces, die von verschiedenen Paketen verwendet werden.
package types

import "context"

// RegisterConfig enthält die Konfiguration für ein Register
type RegisterConfig struct {
	Name      string `json:"name"`
	Address   uint16 `json:"address"`
	Length    uint16 `json:"length"`
	DataType  string `json:"data_type"`
	ByteOrder string `json:"byte_order"`
}

// ProtocolHandler definiert die Schnittstelle für die Kommunikation mit Geräten
type ProtocolHandler interface {
	// ReadRegister liest Daten aus einem Register
	ReadRegister(ctx context.Context, address uint16, length uint16) ([]byte, error)

	// WriteRegister schreibt Daten in ein Register
	WriteRegister(ctx context.Context, address uint16, data []byte) error

	// GetRegisterConfig gibt die Konfiguration für ein Register zurück
	GetRegisterConfig(name string) RegisterConfig

	// Close schließt die Verbindung
	Close() error
}

// ModbusRegisterType definiert den Typ des Modbus-Registers
type ModbusRegisterType string

const (
	// Modbus-Registertypen
	RegisterTypeHolding  ModbusRegisterType = "HOLDING"
	RegisterTypeInput    ModbusRegisterType = "INPUT"
	RegisterTypeCoil     ModbusRegisterType = "COIL"
	RegisterTypeDiscrete ModbusRegisterType = "DISCRETE"
)

// RegisterMap definiert die Zuordnung von Namen zu Registern für Modbus
type RegisterMap struct {
	Name       string             `json:"name"`
	Type       ModbusRegisterType `json:"type"`
	Address    uint16             `json:"address"`
	Length     uint16             `json:"length"`
	DataType   string             `json:"data_type"`
	ByteOrder  string             `json:"byte_order"`
	Multiplier float64            `json:"multiplier"`
	Offset     float64            `json:"offset"`
}
