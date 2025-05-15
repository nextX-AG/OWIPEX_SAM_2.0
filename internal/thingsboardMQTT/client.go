package thingsboardMQTT

import (
	"fmt"
	"log"
	"os"
	"time"

	"owipex_reader/internal/config"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// NewClient erstellt eine neue Instanz des ThingsBoard-Clients.
func NewClient(cfg config.ThingsBoardConfig, dataChan <-chan map[string]interface{}, options ...ClientOption) *Client {
	client := &Client{
		Logger:           log.New(os.Stdout, "[ThingsBoard] ", log.LstdFlags),
		Config:           cfg,
		stopChan:         make(chan struct{}),
		dataChan:         dataChan,
		sharedAttributes: make(map[string]interface{}),
		clientAttributes: make(map[string]interface{}),
		deviceInfo:       make(map[string]interface{}),
		pendingRequests:  make(map[string]chan interface{}),
		nextRequestID:    time.Now().UnixNano(),
		threadSafety:     &threadSafety{},
	}

	// Optionen anwenden
	for _, option := range options {
		option(client)
	}

	return client
}

// WithLogger setzt einen benutzerdefinierten Logger.
func WithLogger(logger *log.Logger) ClientOption {
	return func(c *Client) {
		c.Logger = logger
	}
}

// WithClientID setzt eine benutzerdefinierte Client-ID.
func WithClientID(clientID string) ClientOption {
	return func(c *Client) {
		// Die Client-ID wird später in Connect() verwendet
		c.deviceInfo["clientID"] = clientID
	}
}

// Connect stellt eine Verbindung zum ThingsBoard-Server her.
func (c *Client) Connect() error {
	opts := mqtt.NewClientOptions()
	broker := fmt.Sprintf("tcp://%s:%d", c.Config.Host, c.Config.Port)
	opts.AddBroker(broker)

	// Client-ID bestimmen (entweder aus deviceInfo oder generieren)
	var clientID string
	if id, ok := c.deviceInfo["clientID"].(string); ok && id != "" {
		clientID = id
	} else {
		clientID = fmt.Sprintf("go-owipex-client-%d", time.Now().UnixNano())
	}
	opts.SetClientID(clientID)
	opts.SetUsername(c.Config.AccessToken) // ThingsBoard verwendet AccessToken als Benutzernamen

	// Verbindungsoptionen für maximale Stabilität
	opts.SetCleanSession(true)
	opts.SetOrderMatters(false)
	opts.SetAutoReconnect(true)
	opts.SetConnectTimeout(30 * time.Second)
	opts.SetMaxReconnectInterval(5 * time.Second)
	opts.SetKeepAlive(60 * time.Second)
	opts.SetResumeSubs(true)
	opts.SetWriteTimeout(10 * time.Second)
	opts.SetPingTimeout(5 * time.Second)

	// Default-Handler für unerwartete Nachrichten
	opts.SetDefaultPublishHandler(func(client mqtt.Client, msg mqtt.Message) {
		c.Logger.Printf("Unerwartete Nachricht: %s auf Topic: %s", string(msg.Payload()), msg.Topic())
	})

	// OnConnect Handler
	opts.OnConnect = func(client mqtt.Client) {
		c.Logger.Printf("Verbunden mit ThingsBoard MQTT (Client-ID: %s)", clientID)

		// Mit Verzögerung abonnieren
		go func() {
			time.Sleep(2 * time.Second)
			c.setupSubscriptions()
		}()
	}

	// OnConnectionLost Handler
	opts.OnConnectionLost = func(client mqtt.Client, err error) {
		c.Logger.Printf("Verbindung zu ThingsBoard verloren: %v", err)
	}

	// OnReconnecting Handler
	opts.OnReconnecting = func(client mqtt.Client, opts *mqtt.ClientOptions) {
		c.Logger.Println("Verbindung zu ThingsBoard wird wiederhergestellt...")
	}

	c.mqttClient = mqtt.NewClient(opts)

	c.Logger.Printf("Verbinde zu MQTT-Broker: %s", broker)
	if token := c.getMQTTClient().Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("Verbindungsfehler: %w", token.Error())
	}

	c.Logger.Println("MQTT-Verbindung erfolgreich hergestellt")
	return nil
}

// IsConnected prüft, ob der Client verbunden ist.
func (c *Client) IsConnected() bool {
	return c.mqttClient != nil && c.getMQTTClient().IsConnected()
}

// Start beginnt das Senden von Daten aus dem Datenkanal.
func (c *Client) Start() {
	c.Logger.Println("ThingsBoard-Client-Datenverarbeitung wird gestartet...")
	go c.dataProcessingLoop()
}

// dataProcessingLoop verarbeitet eingehende Daten aus dem Datenkanal.
func (c *Client) dataProcessingLoop() {
	for {
		select {
		case data := <-c.dataChan:
			if data == nil {
				continue
			}
			c.processIncomingData(data)
		case <-c.stopChan:
			c.Logger.Println("Datenverarbeitung wird beendet.")
			return
		}
	}
}

// processIncomingData verarbeitet eingehende Daten und sendet sie an ThingsBoard.
func (c *Client) processIncomingData(data map[string]interface{}) {
	if !c.IsConnected() {
		c.Logger.Println("Kann Daten nicht senden - Client nicht verbunden")
		return
	}

	// Verarbeite unterschiedliche Datenformate
	if simpleData, ok := data["simple"].(map[string]interface{}); ok && len(simpleData) > 0 {
		if err := c.SendTelemetry(simpleData); err != nil {
			c.Logger.Printf("Fehler beim Senden einfacher Telemetrie: %v", err)
		} else {
			c.Logger.Printf("Einfache Telemetrie erfolgreich gesendet: %v", simpleData)
		}
	}

	if jsonData, ok := data["json"].(map[string]interface{}); ok && len(jsonData) > 0 {
		if err := c.SendTelemetry(jsonData); err != nil {
			c.Logger.Printf("Fehler beim Senden von JSON-Telemetrie: %v", err)
		} else {
			c.Logger.Printf("JSON-Telemetrie erfolgreich gesendet: %v", jsonData)
		}
	}
}

// setupSubscriptions konfiguriert alle notwendigen MQTT-Abonnements.
func (c *Client) setupSubscriptions() {
	if !c.IsConnected() {
		c.Logger.Println("Kann Abonnements nicht einrichten - Client nicht verbunden")
		return
	}

	c.Logger.Println("Richte MQTT-Abonnements ein...")

	// Attribute-Updates abonnieren
	topics := []struct {
		topic   string
		handler mqtt.MessageHandler
		qos     byte
		desc    string
	}{
		{"v1/devices/me/attributes", c.handleAttributeUpdate, 1, "Attribute-Updates"},
		{"v1/devices/me/attributes/response/+", c.handleAttributeResponse, 1, "Attribute-Antworten"},
		{"v1/devices/me/rpc/request/+", c.handleRPCRequest, 1, "RPC-Anfragen"},
	}

	for _, t := range topics {
		token := c.getMQTTClient().Subscribe(t.topic, t.qos, t.handler)
		if token.Wait() && token.Error() != nil {
			c.Logger.Printf("Fehler beim Abonnieren von %s: %v", t.desc, token.Error())
		} else {
			c.Logger.Printf("Erfolgreich abonniert: %s", t.desc)
		}
	}

	// Nach einer kurzen Wartezeit die aktuellen Shared Attributes abfragen
	go func() {
		time.Sleep(1 * time.Second)
		c.RequestSharedAttributes(nil) // nil = alle Attribute abfragen
	}()
}

// Stop beendet den Client und gibt Ressourcen frei.
func (c *Client) Stop() {
	c.Logger.Println("ThingsBoard-Client wird beendet...")
	close(c.stopChan)
	if c.mqttClient != nil && c.getMQTTClient().IsConnected() {
		c.getMQTTClient().Disconnect(250)
	}
	c.Logger.Println("ThingsBoard-Client beendet.")
}

// SetAttributeCallback setzt den Callback für Attribut-Updates.
func (c *Client) SetAttributeCallback(callback AttrUpdateCallback) {
	c.attributeCallback = callback
}

// SetRPCCallback setzt den Callback für RPC-Anfragen.
func (c *Client) SetRPCCallback(callback RPCCallback) {
	c.rpcCallback = callback
}

// SetFirmwareUpdateCallback setzt den Callback für Firmware-Updates.
func (c *Client) SetFirmwareUpdateCallback(callback FirmwareUpdateCallback) {
	c.firmwareCallback = callback
}

// getMQTTClient gibt den MQTT-Client zurück (mit Type Assertion).
func (c *Client) getMQTTClient() mqtt.Client {
	return c.mqttClient.(mqtt.Client)
}
