// Dieses Paket dient zum Testen der Modbus-Implementierung.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"owipex_reader/internal/protocol/modbus/test"
	"owipex_reader/internal/service"
)

func main() {
	// Kommandozeilenparameter definieren
	configDir := flag.String("config", "config/devices", "Pfad zum Konfigurationsverzeichnis")
	testProtocol := flag.Bool("test-protocol", false, "Testet die Modbus-Protokollimplementierung")
	testService := flag.Bool("test-service", false, "Testet den Geräte-Service")
	flag.Parse()

	// Absoluten Pfad zum Konfigurationsverzeichnis ermitteln
	absConfigDir, err := filepath.Abs(*configDir)
	if err != nil {
		fmt.Printf("Fehler beim Ermitteln des absoluten Pfads: %v\n", err)
		os.Exit(1)
	}

	// Je nach Kommandozeilenparameter den entsprechenden Test ausführen
	if *testProtocol {
		fmt.Println("=== Teste Modbus-Protokollimplementierung ===")
		test.RunTest()
	}

	if *testService {
		fmt.Println("=== Teste Geräte-Service ===")
		testDeviceService(absConfigDir)
	}

	// Wenn kein Test ausgewählt wurde, Hilfe anzeigen
	if !*testProtocol && !*testService {
		fmt.Println("Bitte wähle einen Test aus:")
		fmt.Println("  -test-protocol: Testet die Modbus-Protokollimplementierung")
		fmt.Println("  -test-service: Testet den Geräte-Service")
	}
}

// testDeviceService testet den Geräte-Service
func testDeviceService(configDir string) {
	// Geräte-Service erstellen
	deviceService := service.NewDeviceService(configDir)

	// Service initialisieren
	if err := deviceService.Initialize(); err != nil {
		fmt.Printf("Fehler beim Initialisieren des Services: %v\n", err)
		return
	}

	// Sensoren aus den Konfigurationsdateien laden
	sensors, err := deviceService.LoadSensorsFromConfig()
	if err != nil {
		fmt.Printf("Fehler beim Laden der Sensoren: %v\n", err)
		return
	}

	// Informationen zu den geladenen Sensoren ausgeben
	fmt.Printf("Erfolgreich %d Sensoren geladen:\n", len(sensors))
	for i, sensor := range sensors {
		fmt.Printf("  %d. %s (ID: %s, Typ: %s)\n", i+1, sensor.Name(), sensor.ID(), sensor.Type())
	}
}
