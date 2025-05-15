// Package modbus implementiert einen Modbus-Client für die Kommunikation mit Sensoren und Aktoren.
package modbus

import (
	"context"
	"fmt"
	"sync"
	"time"

	"owipex_reader/internal/types"

	"github.com/goburrow/modbus"
)

// RegisterType definiert den Typ des Modbus-Registers
type RegisterType string

const (
	// Modbus-Registertypen
	RegisterTypeHolding  RegisterType = "HOLDING"
	RegisterTypeInput    RegisterType = "INPUT"
	RegisterTypeCoil     RegisterType = "COIL"
	RegisterTypeDiscrete RegisterType = "DISCRETE"
)

// ModbusConfig enthält die Konfiguration für die Modbus-Verbindung
type ModbusConfig struct {
	SlaveID      byte                         `json:"slave_id"`
	Port         string                       `json:"port"`
	BaudRate     int                          `json:"baud_rate"`
	DataBits     int                          `json:"data_bits"`
	StopBits     int                          `json:"stop_bits"`
	Parity       string                       `json:"parity"`
	Timeout      time.Duration                `json:"timeout"`
	RegisterMaps map[string]types.RegisterMap `json:"register_maps"`
}

// RegisterMap definiert die Zuordnung von Namen zu Registern
type RegisterMap struct {
	Name       string       `json:"name"`
	Type       RegisterType `json:"type"`
	Address    uint16       `json:"address"`
	Length     uint16       `json:"length"`
	DataType   string       `json:"data_type"`
	ByteOrder  string       `json:"byte_order"`
	Multiplier float64      `json:"multiplier"`
	Offset     float64      `json:"offset"`
}

// ModbusClient implementiert einen Modbus-Client für die Kommunikation
type ModbusClient struct {
	config       ModbusConfig
	client       modbus.Client
	handler      *modbus.RTUClientHandler
	registerMaps map[string]types.RegisterMap
	mutex        sync.RWMutex
}

// NewModbusClient erstellt einen neuen Modbus-Client
func NewModbusClient(config ModbusConfig) (*ModbusClient, error) {
	handler := modbus.NewRTUClientHandler(config.Port)
	handler.BaudRate = config.BaudRate
	handler.DataBits = config.DataBits
	handler.StopBits = config.StopBits
	handler.SlaveId = config.SlaveID
	handler.Timeout = config.Timeout

	// Parität setzen
	switch config.Parity {
	case "N", "n", "none", "NONE":
		handler.Parity = "N"
	case "E", "e", "even", "EVEN":
		handler.Parity = "E"
	case "O", "o", "odd", "ODD":
		handler.Parity = "O"
	default:
		handler.Parity = "N"
	}

	// Verbindung öffnen
	if err := handler.Connect(); err != nil {
		return nil, fmt.Errorf("fehler beim Verbinden mit Modbus: %w", err)
	}

	client := modbus.NewClient(handler)

	// Standardwerte setzen, falls nicht konfiguriert
	if config.Timeout <= 0 {
		config.Timeout = 5 * time.Second
	}

	return &ModbusClient{
		config:       config,
		client:       client,
		handler:      handler,
		registerMaps: config.RegisterMaps,
	}, nil
}

// ReadRegister liest Daten aus einem Register
func (c *ModbusClient) ReadRegister(ctx context.Context, address uint16, length uint16) ([]byte, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Standard-Lesefunktion für Holding-Register verwenden
	result, err := c.client.ReadHoldingRegisters(address, length)
	if err != nil {
		return nil, fmt.Errorf("fehler beim Lesen des Registers %d: %w", address, err)
	}

	return result, nil
}

// WriteRegister schreibt Daten in ein Register
func (c *ModbusClient) WriteRegister(ctx context.Context, address uint16, data []byte) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if len(data) == 0 {
		return fmt.Errorf("keine Daten zum Schreiben")
	}

	// Für einzelnes Register
	if len(data) == 2 {
		value := uint16(data[0])<<8 | uint16(data[1])
		_, err := c.client.WriteSingleRegister(address, value)
		if err != nil {
			return fmt.Errorf("fehler beim Schreiben in Register %d: %w", address, err)
		}
		return nil
	}

	// Für mehrere Register
	_, err := c.client.WriteMultipleRegisters(address, uint16(len(data)/2), data)
	if err != nil {
		return fmt.Errorf("fehler beim Schreiben in Register %d: %w", address, err)
	}

	return nil
}

// ReadRegisterByName liest ein Register anhand seines Namens
func (c *ModbusClient) ReadRegisterByName(ctx context.Context, name string) ([]byte, error) {
	c.mutex.RLock()
	registerMap, exists := c.registerMaps[name]
	c.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("register-Name %s nicht gefunden", name)
	}

	var result []byte
	var err error

	switch registerMap.Type {
	case types.RegisterTypeHolding:
		c.mutex.Lock()
		result, err = c.client.ReadHoldingRegisters(registerMap.Address, registerMap.Length)
		c.mutex.Unlock()
	case types.RegisterTypeInput:
		c.mutex.Lock()
		result, err = c.client.ReadInputRegisters(registerMap.Address, registerMap.Length)
		c.mutex.Unlock()
	case types.RegisterTypeCoil:
		c.mutex.Lock()
		result, err = c.client.ReadCoils(registerMap.Address, registerMap.Length)
		c.mutex.Unlock()
	case types.RegisterTypeDiscrete:
		c.mutex.Lock()
		result, err = c.client.ReadDiscreteInputs(registerMap.Address, registerMap.Length)
		c.mutex.Unlock()
	default:
		return nil, fmt.Errorf("unbekannter Register-Typ: %s", registerMap.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("fehler beim Lesen des Registers %s: %w", name, err)
	}

	return result, nil
}

// GetRegisterConfig gibt die Konfiguration für ein Register zurück
func (c *ModbusClient) GetRegisterConfig(name string) types.RegisterConfig {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	registerMap, exists := c.registerMaps[name]
	if !exists {
		return types.RegisterConfig{}
	}

	return types.RegisterConfig{
		Name:      registerMap.Name,
		Address:   registerMap.Address,
		Length:    registerMap.Length,
		DataType:  registerMap.DataType,
		ByteOrder: registerMap.ByteOrder,
	}
}

// Close schließt die Modbus-Verbindung
func (c *ModbusClient) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.handler != nil {
		return c.handler.Close()
	}
	return nil
}
