package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
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

	// Setup MQTT options
	opts := mqtt.NewClientOptions()
	broker := fmt.Sprintf("tcp://%s:%d", config.Host, config.Port)
	opts.AddBroker(broker)

	// Eindeutige Client-ID
	clientID := fmt.Sprintf("tb-test-client-%d", time.Now().UnixNano())
	opts.SetClientID(clientID)
	opts.SetUsername(config.AccessToken)

	// Verbindungsoptionen
	opts.SetCleanSession(true)
	opts.SetOrderMatters(false)
	opts.SetAutoReconnect(true)
	opts.SetConnectTimeout(30 * time.Second)
	opts.SetMaxReconnectInterval(5 * time.Second)
	opts.SetKeepAlive(60 * time.Second)
	opts.SetResumeSubs(true)
	opts.SetWriteTimeout(10 * time.Second)
	opts.SetPingTimeout(5 * time.Second)

	// Default handler für unerwartete Nachrichten
	opts.SetDefaultPublishHandler(func(client mqtt.Client, msg mqtt.Message) {
		logger.Printf("Received unexpected message from topic %s: %s", msg.Topic(), string(msg.Payload()))
	})

	// Callback bei Verbindung
	opts.OnConnect = func(client mqtt.Client) {
		logger.Printf("Connected to ThingsBoard MQTT broker (client ID: %s)", clientID)

		// Mit Verzögerung abonnieren
		go func() {
			time.Sleep(2 * time.Second)
			setupSubscriptions(client)
		}()
	}

	// Handler für Verbindungsverlust
	opts.OnConnectionLost = func(client mqtt.Client, err error) {
		logger.Printf("Connection lost: %v. Auto-reconnect is enabled.", err)
	}

	// Client erstellen und verbinden
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		logger.Fatalf("Failed to connect: %v", token.Error())
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

// setupSubscriptions konfiguriert alle MQTT-Abonnements
func setupSubscriptions(client mqtt.Client) {
	// 1. Attributes-Änderungen abonnieren
	token := client.Subscribe("v1/devices/me/attributes", 1, handleAttributeUpdate)
	if token.Wait() && token.Error() != nil {
		logger.Printf("Failed to subscribe to attribute updates: %v", token.Error())
	} else {
		logger.Println("✓ Subscribed to attribute updates")
	}

	// 2. Responses auf Attribute-Anfragen abonnieren
	token = client.Subscribe("v1/devices/me/attributes/response/+", 1, handleAttributeResponse)
	if token.Wait() && token.Error() != nil {
		logger.Printf("Failed to subscribe to attribute responses: %v", token.Error())
	} else {
		logger.Println("✓ Subscribed to attribute responses")
	}

	// 3. RPC-Anfragen abonnieren
	token = client.Subscribe("v1/devices/me/rpc/request/+", 1, handleRPCRequest)
	if token.Wait() && token.Error() != nil {
		logger.Printf("Failed to subscribe to RPC requests: %v", token.Error())
	} else {
		logger.Println("✓ Subscribed to RPC requests")
	}

	// Nach aktuellen Attributen fragen
	requestAttributes(client)
}

// handleAttributeUpdate verarbeitet eingehende Attribute-Updates
func handleAttributeUpdate(client mqtt.Client, msg mqtt.Message) {
	logger.Printf("📢 ATTRIBUTE UPDATE received:")
	prettyPrintJSON(msg.Payload())
}

// handleAttributeResponse verarbeitet Antworten auf Attribute-Anfragen
func handleAttributeResponse(client mqtt.Client, msg mqtt.Message) {
	logger.Printf("📋 ATTRIBUTE RESPONSE received for topic %s:", msg.Topic())
	prettyPrintJSON(msg.Payload())
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

// requestAttributes fragt aktuelle Attribute ab
func requestAttributes(client mqtt.Client) {
	logger.Println("Requesting current shared attributes...")
	requestID := fmt.Sprintf("%d", time.Now().UnixNano())
	requestTopic := fmt.Sprintf("v1/devices/me/attributes/request/%s", requestID)

	// Frage nach allen Attributen
	requestData := map[string]interface{}{
		"sharedKeys": "",
	}

	payload, _ := json.Marshal(requestData)
	token := client.Publish(requestTopic, 1, false, payload)

	if token.Wait() && token.Error() != nil {
		logger.Printf("Error requesting attributes: %v", token.Error())
	} else {
		logger.Println("✓ Attribute request sent")
	}
}

// sendPeriodicTelemetry sendet regelmäßig Testdaten
func sendPeriodicTelemetry(client mqtt.Client) {
	ticker := time.NewTicker(60 * time.Second) // Sende alle 60 Sekunden
	counter := 0

	for {
		select {
		case <-ticker.C:
			counter++
			data := map[string]interface{}{
				"testValue": counter,
				"timestamp": time.Now().UnixNano() / int64(time.Millisecond),
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
	if value, exists := os.LookupEnv(key); exists {
		return value
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
