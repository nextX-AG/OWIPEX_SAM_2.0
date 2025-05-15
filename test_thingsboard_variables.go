package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var logger = log.New(os.Stdout, "[ThingsBoardTest] ", log.LstdFlags)

// ThingsBoardConfig definiert die Verbindungsparameter
type ThingsBoardConfig struct {
	Host        string
	Port        int
	AccessToken string
}

func main() {
	logger.Println("Starting MINIMAL ThingsBoard MQTT Client")

	// Konfiguration aus Umgebungsvariablen oder Standardwerten
	config := ThingsBoardConfig{
		Host:        getEnv("TB_HOST", "146.4.67.141"),
		Port:        getIntEnv("TB_PORT", 1883),
		AccessToken: getEnv("TB_TOKEN", "5Ohlb6ZKO4uNw9O2DHwk"),
	}

	logger.Printf("Connecting to ThingsBoard at %s:%d with token: %s",
		config.Host, config.Port, config.AccessToken)

	// Extremes Minimal-Setup für MQTT
	opts := mqtt.NewClientOptions()
	broker := fmt.Sprintf("tcp://%s:%d", config.Host, config.Port)
	opts.AddBroker(broker)

	// Client-ID (kurz und identifizierbar)
	hostname, _ := os.Hostname()
	shortID := fmt.Sprintf("%d", time.Now().Unix()%1000) // Kurze ID
	clientID := fmt.Sprintf("tb-minimal-%s-%s", hostname, shortID)
	opts.SetClientID(clientID)
	opts.SetUsername(config.AccessToken)

	// Reduzierte Verbindungsoptionen
	opts.SetCleanSession(true)
	opts.SetConnectTimeout(10 * time.Second)
	opts.SetKeepAlive(30 * time.Second)
	opts.SetWriteTimeout(5 * time.Second)
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(5 * time.Second)

	// WICHTIG: Deaktiviere Nachrichtenwiederherstellung
	opts.SetResumeSubs(false)
	opts.SetOrderMatters(false)

	// Einfache Connection-Handler
	opts.OnConnect = func(client mqtt.Client) {
		logger.Printf("Connected to ThingsBoard (ID: %s)", clientID)

		// ⚠️ SEHR WICHTIG: Auf keinen Fall sofort subscriben oder andere Aktionen ausführen
		// Warte mind. 2 Sekunden zur Stabilisierung der Verbindung
		time.Sleep(2 * time.Second)

		// Nur ein einzelnes Subscription - das einfachste
		if token := client.Subscribe("v1/devices/me/attributes", 0, handleMessage); token.Wait() && token.Error() != nil {
			logger.Printf("Failed to subscribe: %v", token.Error())
		} else {
			logger.Println("✓ Subscribed to attribute updates")
		}
	}

	opts.OnConnectionLost = func(client mqtt.Client, err error) {
		logger.Printf("Connection lost: %v", err)
	}

	// Einfacher Default-Handler für alle Nachrichten
	opts.SetDefaultPublishHandler(handleMessage)

	// Client erstellen und verbinden
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		logger.Fatalf("Could not connect to ThingsBoard: %v", token.Error())
	}

	// Telemetry-Routine (nach erfolgreicher Verbindung)
	go func() {
		// Warte, bis die Verbindung stabil ist
		time.Sleep(5 * time.Second)
		ticker := time.NewTicker(30 * time.Second)
		for {
			select {
			case <-ticker.C:
				sendSimpleTelemetry(client)
			}
		}
	}()

	// Signal-Handling für sauberes Beenden
	logger.Println("Client running. Press Ctrl+C to exit.")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Println("Shutting down...")
	client.Disconnect(250)
	logger.Println("Client stopped.")
}

// Universeller Message-Handler für alle Nachrichten
func handleMessage(client mqtt.Client, msg mqtt.Message) {
	logger.Printf("📨 Message received on topic: %s", msg.Topic())

	// Prüfe, ob die Nachricht leer ist
	if len(msg.Payload()) == 0 {
		logger.Println("Empty message received")
		return
	}

	// Versuche, JSON zu parsen
	var data map[string]interface{}
	if err := json.Unmarshal(msg.Payload(), &data); err != nil {
		logger.Printf("Raw data (not valid JSON): %s", string(msg.Payload()))
		return
	}

	// Zeige formatierte Daten
	prettyJSON, _ := json.MarshalIndent(data, "", "  ")
	fmt.Println(string(prettyJSON))
}

// Einfache Telemetrie senden
func sendSimpleTelemetry(client mqtt.Client) {
	if !client.IsConnected() {
		logger.Println("Cannot send telemetry - not connected")
		return
	}

	data := map[string]interface{}{
		"heartbeat": time.Now().Unix(),
		"status":    "running",
	}

	payload, _ := json.Marshal(data)
	if token := client.Publish("v1/devices/me/telemetry", 0, false, payload); token.Wait() && token.Error() != nil {
		logger.Printf("Failed to send telemetry: %v", token.Error())
	} else {
		logger.Println("✓ Telemetry sent")
	}
}

// getEnv liest eine Umgebungsvariable oder gibt den Standardwert zurück
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists && strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	return defaultValue
}

// getIntEnv liest eine Umgebungsvariable als Integer
func getIntEnv(key string, defaultValue int) int {
	if valueStr, exists := os.LookupEnv(key); exists {
		var value int
		if _, err := fmt.Sscanf(valueStr, "%d", &value); err == nil {
			return value
		}
	}
	return defaultValue
}
