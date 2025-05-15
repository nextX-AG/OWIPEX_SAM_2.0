package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"owipex_reader/internal/adapter"
	"owipex_reader/internal/config"
	"owipex_reader/internal/thingsboard"
)

// Dateipfade
const (
	defaultConfigPath  = "config/sensors.json"
	defaultDevicesPath = "config/devices"
)

func main() {
	logger := log.New(os.Stdout, "[MainApp] ", log.LstdFlags)
	logger.Println("Starte Owipex RS485 Reader (Go Version mit neuer Architektur)...")

	// Konfigurationspfad bestimmen
	actualConfigPath := defaultConfigPath
	if _, err := os.Stat(actualConfigPath); os.IsNotExist(err) {
		logger.Printf("Konfigurationsdatei nicht gefunden unter %s. Versuche alternativen Pfad für Entwicklung: /home/ubuntu/owipex_project/go_owipex_reader/config/sensors.json", actualConfigPath)
		actualConfigPath = "/home/ubuntu/owipex_project/go_owipex_reader/config/sensors.json"
		if _, err := os.Stat(actualConfigPath); os.IsNotExist(err) {
			logger.Fatalf("Konfigurationsdatei auch nicht gefunden unter /home/ubuntu/owipex_project/go_owipex_reader/config/sensors.json. Bitte stelle sicher, dass sie existiert. Fehler: %v", err)
		}
	}

	// Konfiguration laden
	appCfg, err := config.LoadAppConfig(actualConfigPath)
	if err != nil {
		logger.Fatalf("Fehler beim Laden der Anwendungskonfiguration aus %s: %v", actualConfigPath, err)
	}

	logger.Printf("Konfiguration geladen. RS485 Port: %s, ThingsBoard Host: %s:%d, AccessToken: %s",
		appCfg.RS485.Port, appCfg.ThingsBoard.Host, appCfg.ThingsBoard.Port, appCfg.ThingsBoard.AccessToken)

	// Kanal für die Kommunikation zwischen SensorAdapter und ThingsBoard-Client erstellen
	dataToThingsBoardChan := make(chan map[string]interface{}, 100) // Gepufferter Kanal

	// Sensor-Adapter erstellen
	sensorAdapter, err := adapter.NewSensorAdapter(actualConfigPath, defaultDevicesPath, dataToThingsBoardChan)
	if err != nil {
		logger.Fatalf("Fehler beim Initialisieren des SensorAdapter: %v", err)
	}

	// ThingsBoard-Client erstellen
	tbClient := thingsboard.NewClient(appCfg.ThingsBoard, dataToThingsBoardChan)

	// Callback für Attribute-Updates setzen
	tbClient.SetAttributeCallback(func(attributes map[string]interface{}) {
		logger.Printf("Shared Attributes Update empfangen: %v", attributes)

		// Button-Status-Änderung verarbeiten
		if buttonStatus, ok := attributes["buttonStatus"]; ok {
			logger.Printf("Button-Status geändert: %v", buttonStatus)
			// Hier kann die Logik für Aktionen basierend auf dem Button-Status implementiert werden
		}

		// Logging-Level anpassen
		if loggingLevel, ok := attributes["loggingLevel"].(string); ok {
			logger.Printf("Ändere Logging-Level auf: %s", loggingLevel)
			// Hier kann die Logik zur Anpassung des Logging-Levels implementiert werden
		}
	})

	// Callback für RPC-Aufrufe setzen
	tbClient.SetRPCCallback(func(method string, params map[string]interface{}) {
		logger.Printf("RPC-Aufruf empfangen: method=%s, params=%v", method, params)

		// Beispiel für die Verarbeitung verschiedener RPC-Methoden
		switch method {
		case "setSamplingInterval":
			if interval, ok := params["interval"].(float64); ok {
				logger.Printf("Setze Sampling-Intervall auf: %.0f Sekunden", interval)
				// Hier kann die Logik zur Anpassung des Sampling-Intervalls implementiert werden
			}
		case "restartDevice":
			logger.Printf("Neustart-Befehl empfangen")
			// Hier kann die Logik für einen Neustart implementiert werden
		case "forceReconnect":
			logger.Printf("Erzwinge Neuverbindung zu den Sensoren")
			// Hier kann die Logik für einen Neuaufbau der Verbindung implementiert werden
		default:
			logger.Printf("Unbekannte RPC-Methode: %s", method)
		}
	})

	// Verbindung zu ThingsBoard herstellen
	if err := tbClient.Connect(); err != nil {
		logger.Printf("Warnung: Fehler beim Verbinden mit ThingsBoard: %v. Überprüfe Access Token und Server-Details. Anwendung wird trotzdem gestartet, MQTT wird versuchen, sich erneut zu verbinden.", err)
		// Anwendung kann auch bei einem anfänglichen MQTT-Verbindungsfehler starten, da sie automatisch wieder verbindet
	}

	// Dienste starten
	sensorAdapter.Start()
	tbClient.Start()

	logger.Println("Anwendung gestartet. Drücke Strg+C zum Beenden.")

	// Auf Beendigungssignal warten
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Println("Shutdown-Signal empfangen. Stoppe Dienste...")

	tbClient.Stop()
	sensorAdapter.Stop()

	// Etwas Zeit für graceful Shutdown lassen
	time.Sleep(2 * time.Second)
	logger.Println("Anwendung wurde ordnungsgemäß beendet.")
}
