// Package factory enthält Funktionen zur Erstellung von Protokoll-Handlern.
package factory

import (
	"fmt"
	"time"

	"owipex_reader/internal/protocol/modbus"
	"owipex_reader/internal/types"
)

// CreateProtocolHandler erstellt einen Protokoll-Handler basierend auf der Konfiguration
func CreateProtocolHandler(protocolType string, config map[string]interface{}) (types.ProtocolHandler, error) {
	switch protocolType {
	case "modbus":
		return createModbusHandler(config)
	default:
		return nil, fmt.Errorf("unbekannter Protokolltyp: %s", protocolType)
	}
}

// createModbusHandler erstellt einen Modbus-Protokoll-Handler
func createModbusHandler(config map[string]interface{}) (types.ProtocolHandler, error) {
	// Standardwerte setzen
	modbusConfig := modbus.ModbusConfig{
		Port:         "/dev/ttyUSB0",  // Standard-Port
		BaudRate:     9600,            // Standard-Baudrate
		DataBits:     8,               // Standard-Datenbits
		StopBits:     1,               // Standard-Stopbits
		Parity:       "N",             // Standard-Parität (None)
		Timeout:      5 * time.Second, // Standard-Timeout
		RegisterMaps: make(map[string]types.RegisterMap),
	}

	// SlaveID aus der Konfiguration extrahieren
	if slaveID, ok := config["slave_id"].(float64); ok {
		modbusConfig.SlaveID = byte(slaveID)
	}

	// Port aus der Konfiguration extrahieren
	if port, ok := config["port"].(string); ok {
		modbusConfig.Port = port
	}

	// BaudRate aus der Konfiguration extrahieren
	if baudRate, ok := config["baud_rate"].(float64); ok {
		modbusConfig.BaudRate = int(baudRate)
	}

	// DataBits aus der Konfiguration extrahieren
	if dataBits, ok := config["data_bits"].(float64); ok {
		modbusConfig.DataBits = int(dataBits)
	}

	// StopBits aus der Konfiguration extrahieren
	if stopBits, ok := config["stop_bits"].(float64); ok {
		modbusConfig.StopBits = int(stopBits)
	}

	// Parity aus der Konfiguration extrahieren
	if parity, ok := config["parity"].(string); ok {
		modbusConfig.Parity = parity
	}

	// Timeout aus der Konfiguration extrahieren
	if timeout, ok := config["timeout"].(float64); ok {
		modbusConfig.Timeout = time.Duration(timeout) * time.Millisecond
	}

	// Register-Maps aus der Konfiguration extrahieren
	if registerMaps, ok := config["register_maps"].(map[string]interface{}); ok {
		for name, regMapInterface := range registerMaps {
			if regMap, ok := regMapInterface.(map[string]interface{}); ok {
				registerMap := types.RegisterMap{
					Name: name,
				}

				// Register-Typ extrahieren
				if regType, ok := regMap["type"].(string); ok {
					// Umwandlung in ModbusRegisterType
					switch regType {
					case "HOLDING", "holding":
						registerMap.Type = types.RegisterTypeHolding
					case "INPUT", "input":
						registerMap.Type = types.RegisterTypeInput
					case "COIL", "coil":
						registerMap.Type = types.RegisterTypeCoil
					case "DISCRETE", "discrete":
						registerMap.Type = types.RegisterTypeDiscrete
					default:
						// Standardmäßig Holding-Register verwenden
						registerMap.Type = types.RegisterTypeHolding
					}
				}

				// Register-Adresse extrahieren
				if address, ok := regMap["address"].(float64); ok {
					registerMap.Address = uint16(address)
				}

				// Register-Länge extrahieren
				if length, ok := regMap["length"].(float64); ok {
					registerMap.Length = uint16(length)
				}

				// Datentyp extrahieren
				if dataType, ok := regMap["data_type"].(string); ok {
					registerMap.DataType = dataType
				}

				// Byte-Order extrahieren
				if byteOrder, ok := regMap["byte_order"].(string); ok {
					registerMap.ByteOrder = byteOrder
				}

				// Multiplikator extrahieren
				if multiplier, ok := regMap["multiplier"].(float64); ok {
					registerMap.Multiplier = multiplier
				}

				// Offset extrahieren
				if offset, ok := regMap["offset"].(float64); ok {
					registerMap.Offset = offset
				}

				// Register-Map zur Konfiguration hinzufügen
				modbusConfig.RegisterMaps[name] = registerMap
			}
		}
	}

	// Modbus-Client erstellen
	return modbus.NewModbusClient(modbusConfig)
}
