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
	logger.Println("Starting ThingsBoard Variables Test Tool")

	// Konfiguration aus Umgebungsvariablen oder Standardwerten
	config := ThingsBoardConfig{
		Host:        getEnv("TB_HOST", "146.4.67.141"),
		Port:        getIntEnv("TB_PORT", 1883),
		AccessToken: getEnv("TB_TOKEN", "5Ohlb6ZKO4uNw9O2DHwk"),
	}

	logger.Printf("Connecting to ThingsBoard at %s:%d with token: %s",
		config.Host, config.Port, config.AccessToken)

	// Verbesserte MQTT-Optionen
	opts := mqtt.NewClientOptions()
	broker := fmt.Sprintf("tcp://%s:%d", config.Host, config.Port)
	opts.AddBroker(broker)

	// Eindeutige Client-ID mit Hostname und Prozess-ID für bessere Identifikation
	hostname, _ := os.Hostname()
	clientID := fmt.Sprintf("tb-test-%s-%d", hostname, os.Getpid())
	opts.SetClientID(clientID)
	opts.SetUsername(config.AccessToken)

	// Verbesserte Verbindungsoptionen
	opts.SetCleanSession(true)
	opts.SetOrderMatters(false)
	opts.SetAutoReconnect(true)
	opts.SetConnectTimeout(20 * time.Second)
	opts.SetMaxReconnectInterval(10 * time.Second)
	opts.SetKeepAlive(30 * time.Second) // Reduzierte KeepAlive für schnellere Erkennung von Verbindungsproblemen
	opts.SetResumeSubs(false)           // Ändere zu false, um Probleme mit abgebrochenen Abonnements zu vermeiden
	opts.SetWriteTimeout(10 * time.Second)
	opts.SetPingTimeout(5 * time.Second)

	// Kritische Konfiguration für Verbindungsstabilität
	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		logger.Printf("Connection lost: %v. Auto-reconnect is enabled.", err)
		// Wichtig: Verzögere weitere Aktionen, um dem Client Zeit zur Wiederverbindung zu geben
		time.Sleep(2 * time.Second)
	})

	// Verbesserte OnConnect-Handler
	connectHandler := func(client mqtt.Client) {
		logger.Printf("Connected to ThingsBoard MQTT broker (client ID: %s)", clientID)

		// Wichtig: Gib dem Broker Zeit, die Verbindung vollständig aufzubauen
		time.Sleep(1 * time.Second)

		// Mit Verzögerung abonnieren und Fehler besser behandeln
		go func() {
			// Mehr Zeit zwischen Verbindung und Subscription
			time.Sleep(2 * time.Second)

			// Reihenfolge: Erst RPC, dann Attribute
			subscribeToRPC(client)
			time.Sleep(500 * time.Millisecond)
			subscribeToAttributes(client)
			time.Sleep(500 * time.Millisecond)

			// Erst nach erfolgreichen Subscriptions Attribute anfragen
			requestAttributes(client)
		}()
	}
	opts.OnConnect = connectHandler

	// Default handler für unerwartete Nachrichten
	opts.SetDefaultPublishHandler(func(client mqtt.Client, msg mqtt.Message) {
		logger.Printf("Received unexpected message from topic %s: %s", msg.Topic(), string(msg.Payload()))
	})

	// Client erstellen und verbinden mit Wiederholungslogik
	client := mqtt.NewClient(opts)

	// Timeout-Handler für Verbindung
	connectSuccess := make(chan bool, 1)
	go func() {
		retryCount := 0
		maxRetries := 5

		for retryCount < maxRetries {
			logger.Printf("Connecting to ThingsBoard (attempt %d/%d)...", retryCount+1, maxRetries)
			token := client.Connect()
			connected := token.WaitTimeout(15 * time.Second)

			if connected && token.Error() == nil {
				connectSuccess <- true
				return
			}

			if !connected {
				logger.Printf("Connection attempt timed out")
			} else {
				logger.Printf("Connection error: %v", token.Error())
			}

			retryCount++
			time.Sleep(2 * time.Second)
		}

		connectSuccess <- false
	}()

	// Auf Verbindung warten
	if success := <-connectSuccess; !success {
		logger.Fatalf("Failed to connect after multiple attempts")
	}

	// Regelmäßiger Versand von Telemetriedaten (als Test)
	go sendPeriodicTelemetry(client)

	// Warten auf Beendigung
	logger.Println("Test client running. Press Ctrl+C to exit.")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Println("Shutting down...")
	client.Disconnect(250)
	logger.Println("Test client stopped.")
}

// Verbesserte Subscription-Funktionen
func subscribeToRPC(client mqtt.Client) {
	token := client.Subscribe("v1/devices/me/rpc/request/+", 1, handleRPCRequest)
	if token.WaitTimeout(10*time.Second) && token.Error() != nil {
		logger.Printf("Failed to subscribe to RPC requests: %v", token.Error())
	} else {
		logger.Println("✓ Subscribed to RPC requests")
	}
}

func subscribeToAttributes(client mqtt.Client) {
	// 1. Attributes-Änderungen abonnieren
	token := client.Subscribe("v1/devices/me/attributes", 1, handleAttributeUpdate)
	if token.WaitTimeout(10*time.Second) && token.Error() != nil {
		logger.Printf("Failed to subscribe to attribute updates: %v", token.Error())
	} else {
		logger.Println("✓ Subscribed to attribute updates")
	}

	// 2. Responses auf Attribute-Anfragen abonnieren
	token = client.Subscribe("v1/devices/me/attributes/response/+", 1, handleAttributeResponse)
	if token.WaitTimeout(10*time.Second) && token.Error() != nil {
		logger.Printf("Failed to subscribe to attribute responses: %v", token.Error())
	} else {
		logger.Println("✓ Subscribed to attribute responses")
	}
}

// handleAttributeUpdate verarbeitet eingehende Attribute-Updates
func handleAttributeUpdate(client mqtt.Client, msg mqtt.Message) {
	logger.Printf("📢 ATTRIBUTE UPDATE received:")
	prettyPrintJSON(msg.Payload())
}

// handleAttributeResponse verarbeitet Antworten auf Attribute-Anfragen
func handleAttributeResponse(client mqtt.Client, msg mqtt.Message) {
	topic := msg.Topic()
	requestID := topic[len("v1/devices/me/attributes/response/"):]

	logger.Printf("📋 ATTRIBUTE RESPONSE received for requestID: %s", requestID)

	// Prüfen, ob die Antwort Daten enthält
	if len(msg.Payload()) == 0 {
		logger.Println("Warning: Empty attribute response received")
		return
	}

	// Versuchen, die Attributdaten zu parsen
	var responseData map[string]interface{}
	if err := json.Unmarshal(msg.Payload(), &responseData); err != nil {
		logger.Printf("Error parsing attribute response: %v", err)
		logger.Printf("Raw payload: %s", string(msg.Payload()))
		return
	}

	// Spezifische Attribute extrahieren, wenn vorhanden
	shared, hasShared := responseData["shared"]
	if hasShared {
		logger.Println("Shared attributes:")
		prettyPrintJSON(msg.Payload())

		// Hier können wir spezifische Attribute für unsere Applikation verarbeiten
		sharedMap, ok := shared.(map[string]interface{})
		if ok {
			for key, value := range sharedMap {
				logger.Printf("  • %s: %v", key, value)
			}
		}
	} else {
		logger.Println("No shared attributes found in response")
		prettyPrintJSON(msg.Payload())
	}
}

// handleRPCRequest verarbeitet eingehende RPC-Aufrufe
func handleRPCRequest(client mqtt.Client, msg mqtt.Message) {
	logger.Printf("🔄 RPC REQUEST received on topic %s:", msg.Topic())
	prettyPrintJSON(msg.Payload())

	// Extrahiere request ID aus Topic
	topic := msg.Topic()
	requestID := topic[len("v1/devices/me/rpc/request/"):]

	// Parse RPC-Anfrage
	var rpcData map[string]interface{}
	if err := json.Unmarshal(msg.Payload(), &rpcData); err != nil {
		logger.Printf("Error parsing RPC payload: %v", err)
		return
	}

	// Extrahiere Methode und Parameter
	method, ok := rpcData["method"].(string)
	if !ok {
		logger.Printf("RPC request missing method field")
		return
	}

	params, _ := rpcData["params"].(map[string]interface{})
	logger.Printf("Method: %s, Params: %v", method, params)

	// Sende Antwort
	response := map[string]interface{}{
		"success": true,
		"result":  fmt.Sprintf("Method '%s' processed by test client", method),
	}

	responseTopic := fmt.Sprintf("v1/devices/me/rpc/response/%s", requestID)
	respPayload, _ := json.Marshal(response)

	token := client.Publish(responseTopic, 1, false, respPayload)
	if token.Wait() && token.Error() != nil {
		logger.Printf("Error sending RPC response: %v", token.Error())
	} else {
		logger.Printf("✓ RPC response sent")
	}
}

// Verbesserte Attributanfrage mit Timeout und spezifischen Attributen
func requestAttributes(client mqtt.Client) {
	logger.Println("Requesting specific shared attributes...")

	// Prüfe zuerst, ob der Client verbunden ist
	if !client.IsConnected() {
		logger.Println("Cannot request attributes - client not connected")
		return
	}

	// Dynamische Request-ID mit Zeitstempel
	requestID := fmt.Sprintf("%d", time.Now().UnixNano())
	requestTopic := fmt.Sprintf("v1/devices/me/attributes/request/%s", requestID)

	// Nur tatsächlich existierende Attribute anfragen
	// WICHTIG: Diese müssen in ThingsBoard definiert sein!
	requestData := map[string]interface{}{
		"sharedKeys": "testValue,testString,testBoolean",
	}

	payload, err := json.Marshal(requestData)
	if err != nil {
		logger.Printf("Error marshalling attribute request: %v", err)
		return
	}

	// Mit QoS=1 senden, damit die Nachricht garantiert ankommt
	token := client.Publish(requestTopic, 1, false, payload)

	if token.WaitTimeout(10*time.Second) && token.Error() != nil {
		logger.Printf("Error requesting attributes: %v", token.Error())
	} else {
		logger.Printf("✓ Attribute request sent with requestID: %s", requestID)
		logger.Printf("  Requested keys: %s", requestData["sharedKeys"])
	}
}

// sendPeriodicTelemetry sendet regelmäßig Testdaten
func sendPeriodicTelemetry(client mqtt.Client) {
	ticker := time.NewTicker(60 * time.Second) // Sende alle 60 Sekunden
	counter := 0

	for {
		select {
		case <-ticker.C:
			if !client.IsConnected() {
				logger.Println("Cannot send telemetry - client not connected")
				continue
			}

			counter++
			data := map[string]interface{}{
				"testValue":  counter,
				"timestamp":  time.Now().UnixNano() / int64(time.Millisecond),
				"clientInfo": fmt.Sprintf("TestClient_%d", os.Getpid()),
			}

			payload, _ := json.Marshal(data)
			token := client.Publish("v1/devices/me/telemetry", 1, false, payload)

			if token.Wait() && token.Error() != nil {
				logger.Printf("Error sending telemetry: %v", token.Error())
			} else {
				logger.Printf("✓ Test telemetry sent: %v", data)
			}
		}
	}
}

// prettyPrintJSON gibt JSON-Daten formatiert aus
func prettyPrintJSON(data []byte) {
	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		logger.Printf("Raw data (not valid JSON): %s", string(data))
		return
	}

	prettyJSON, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		logger.Printf("Error formatting JSON: %v", err)
		logger.Printf("Raw data: %s", string(data))
		return
	}

	fmt.Println(string(prettyJSON))
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
