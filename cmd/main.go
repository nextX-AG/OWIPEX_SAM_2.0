package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"owipex_reader/internal/config"
	"owipex_reader/internal/manager"
	"owipex_reader/internal/thingsboard"
)

// defaultConfigPath is relative to the binary's execution directory.
const defaultConfigPath = "config/sensors.json"

func main() {
	logger := log.New(os.Stdout, "[MainApp] ", log.LstdFlags)
	logger.Println("Starting Owipex RS485 Reader (Go Version)...")

	// Use the config path relative to the go_owipex_reader directory
	// If running `go run cmd/main.go` from `go_owipex_reader` directory, path is `config/sensors.json`
	// If binary is in `go_owipex_reader/bin/main`, path would be `../config/sensors.json`
	// For `go run ./cmd/main.go` from `go_owipex_reader` it's `config/sensors.json`
	actualConfigPath := defaultConfigPath // Using the const defined above

	// Check if the config file exists at the expected path
	if _, err := os.Stat(actualConfigPath); os.IsNotExist(err) {
		logger.Printf("Config file not found at %s. Trying alternative path for development: /home/ubuntu/owipex_project/go_owipex_reader/config/sensors.json", actualConfigPath)
		actualConfigPath = "/home/ubuntu/owipex_project/go_owipex_reader/config/sensors.json"
		if _, err := os.Stat(actualConfigPath); os.IsNotExist(err) {
			logger.Fatalf("Config file also not found at /home/ubuntu/owipex_project/go_owipex_reader/config/sensors.json. Please ensure it exists. Error: %v", err)
		}
	}

	appCfg, err := config.LoadAppConfig(actualConfigPath)
	if err != nil {
		logger.Fatalf("Failed to load application configuration from %s: %v", actualConfigPath, err)
	}

	logger.Printf("Configuration loaded. RS485 Port: %s, ThingsBoard Host: %s:%d, AccessToken: %s", appCfg.RS485.Port, appCfg.ThingsBoard.Host, appCfg.ThingsBoard.Port, appCfg.ThingsBoard.AccessToken) // Changed Server to Host

	// Create a channel for SensorManager to send data to ThingsBoardClient
	dataToThingsBoardChan := make(chan map[string]interface{}, 100) // Buffered channel

	sensorMgr, err := manager.NewSensorManager(actualConfigPath, dataToThingsBoardChan)
	if err != nil {
		logger.Fatalf("Failed to initialize SensorManager: %v", err)
	}

	tbClient := thingsboard.NewClient(appCfg.ThingsBoard, dataToThingsBoardChan)

	// Setup attribute callback to handle shared attributes updates
	tbClient.SetAttributeCallback(func(attributes map[string]interface{}) {
		logger.Printf("Received shared attributes update: %v", attributes)

		// Handle button status if present
		if buttonStatus, ok := attributes["buttonStatus"]; ok {
			logger.Printf("Button status changed: %v", buttonStatus)
			// Hier kann die Logik für Aktionen basierend auf dem Button-Status implementiert werden
			// Zum Beispiel, bestimmte Modi aktivieren/deaktivieren
		}

		// Logging-Level anpassen, wenn vorhanden
		if loggingLevel, ok := attributes["loggingLevel"]; ok {
			if level, ok := loggingLevel.(string); ok {
				logger.Printf("Changing logging level to: %s", level)
				// Hier kann die Logik zur Anpassung des Logging-Levels implementiert werden
			}
		}
	})

	// Setup RPC callback
	tbClient.SetRPCCallback(func(method string, params map[string]interface{}) {
		logger.Printf("Received RPC call: method=%s, params=%v", method, params)

		// Beispiel für die Verarbeitung verschiedener RPC-Methoden
		switch method {
		case "setSamplingInterval":
			if interval, ok := params["interval"].(float64); ok {
				logger.Printf("Setting sampling interval to: %.0f seconds", interval)
				// Hier kann die Logik zur Anpassung des Sampling-Intervalls implementiert werden
			}
		case "restartDevice":
			logger.Printf("Received restart command")
			// Hier kann die Logik für einen Neustart implementiert werden
		case "forceReconnect":
			logger.Printf("Forcing reconnect to sensors")
			// Hier kann die Logik für einen Neuaufbau der Verbindung implementiert werden
		default:
			logger.Printf("Unknown RPC method: %s", method)
		}
	})

	if err := tbClient.Connect(); err != nil {
		logger.Printf("Warning: Failed to connect to ThingsBoard: %v. Check Access Token and server details in /etc/owipex/go_reader.env. Proceeding with application start, MQTT will attempt to reconnect.", err)
		// Allow application to start even if MQTT connection fails initially, as it has auto-reconnect
	}

	// Start services
	sensorMgr.Start()
	tbClient.Start()

	logger.Println("Application started. Press Ctrl+C to exit.")

	// Wait for termination signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Println("Shutdown signal received. Stopping services...")

	tbClient.Stop()
	sensorMgr.Stop()

	// Allow some time for graceful shutdown
	time.Sleep(2 * time.Second)
	logger.Println("Application shut down gracefully.")
}
