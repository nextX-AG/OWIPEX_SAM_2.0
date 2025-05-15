// Package test enthält Tests für das Modbus-Protokoll.
package test

import (
	"context"
	"fmt"
	"time"

	"owipex_reader/internal/protocol/modbus"
	"owipex_reader/internal/types"
)

// TestModbusConnection testet die Verbindung zu einem Modbus-Gerät
func TestModbusConnection() error {
	// Konfiguration für den Modbus-Client erstellen
	config := modbus.ModbusConfig{
		SlaveID:  1,
		Port:     "/dev/ttyUSB0",
		BaudRate: 9600,
		DataBits: 8,
		StopBits: 1,
		Parity:   "N",
		Timeout:  5 * time.Second,
		RegisterMaps: map[string]types.RegisterMap{
			"test_register": {
				Name:      "test_register",
				Type:      types.RegisterTypeHolding,
				Address:   0x0001,
				Length:    1,
				DataType:  "uint16",
				ByteOrder: "big_endian",
			},
		},
	}

	// Modbus-Client erstellen
	client, err := modbus.NewModbusClient(config)
	if err != nil {
		return fmt.Errorf("fehler beim Erstellen des Modbus-Clients: %w", err)
	}
	defer client.Close()

	// Register lesen
	ctx := context.Background()
	data, err := client.ReadRegister(ctx, 0x0001, 1)
	if err != nil {
		return fmt.Errorf("fehler beim Lesen des Registers: %w", err)
	}

	// Ergebnis ausgeben
	fmt.Printf("Daten aus Register 0x0001: %v\n", data)

	return nil
}

// RunTest führt den Test für die Modbus-Verbindung aus
func RunTest() {
	fmt.Println("Starte Test für Modbus-Protokoll...")
	if err := TestModbusConnection(); err != nil {
		fmt.Printf("Test fehlgeschlagen: %v\n", err)
	} else {
		fmt.Println("Test erfolgreich abgeschlossen.")
	}
}
