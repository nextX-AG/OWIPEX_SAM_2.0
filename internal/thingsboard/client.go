// Package thingsboard implementiert einen umfassenden Client für die ThingsBoard IoT-Plattform.
// Diese Implementierung bietet Zugriff auf alle MQTT-APIs, die ThingsBoard bereitstellt.
package thingsboard

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"owipex_reader/internal/config"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// Client ist die Hauptschnittstelle für die ThingsBoard-Kommunikation.
// Er stellt alle Methoden bereit, um mit der ThingsBoard-Plattform zu interagieren.
type Client struct {
	mqttClient mqtt.Client
	logger     *log.Logger
	config     config.ThingsBoardConfig

	// Kanäle und Steuerung
	stopChan chan struct{}
	dataChan <-chan map[string]interface{}

	// Thread-sicherer Zugriff auf lokale Caches
	attributesMutex sync.RWMutex
	deviceInfoMutex sync.RWMutex
	requestIDMutex  sync.Mutex

	// Lokale Caches
	sharedAttributes map[string]interface{}
	clientAttributes map[string]interface{}
	deviceInfo       map[string]interface{}

	// Verwaltung für asynchrone Anfragen
	pendingRequests map[string]chan interface{}
	nextRequestID   int64

	// Callback-Funktionen
	attributeCallback AttrUpdateCallback
	rpcCallback       RPCCallback
	firmwareCallback  FirmwareUpdateCallback
}

// Callback-Typen
type AttrUpdateCallback func(map[string]interface{})
type RPCCallback func(string, map[string]interface{}) (interface{}, error)
type FirmwareUpdateCallback func(title, version, checksum, algorithm string)

// ClientOption ist ein Funktionstyp, der verwendet wird, um Client-Optionen zu konfigurieren.
type ClientOption func(*Client)

// NewClient erstellt eine neue Instanz des ThingsBoard-Clients.
func NewClient(cfg config.ThingsBoardConfig, dataChan <-chan map[string]interface{}, options ...ClientOption) *Client {
	client := &Client{
		logger:           log.New(os.Stdout, "[ThingsBoard] ", log.LstdFlags),
		config:           cfg,
		stopChan:         make(chan struct{}),
		dataChan:         dataChan,
		sharedAttributes: make(map[string]interface{}),
		clientAttributes: make(map[string]interface{}),
		deviceInfo:       make(map[string]interface{}),
		pendingRequests:  make(map[string]chan interface{}),
		nextRequestID:    time.Now().UnixNano(),
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
		c.logger = logger
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
	broker := fmt.Sprintf("tcp://%s:%d", c.config.Host, c.config.Port)
	opts.AddBroker(broker)

	// Client-ID bestimmen (entweder aus deviceInfo oder generieren)
	var clientID string
	if id, ok := c.deviceInfo["clientID"].(string); ok && id != "" {
		clientID = id
	} else {
		clientID = fmt.Sprintf("go-owipex-client-%d", time.Now().UnixNano())
	}
	opts.SetClientID(clientID)
	opts.SetUsername(c.config.AccessToken) // ThingsBoard verwendet AccessToken als Benutzernamen

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
		c.logger.Printf("Unerwartete Nachricht: %s auf Topic: %s", string(msg.Payload()), msg.Topic())
	})

	// OnConnect Handler
	opts.OnConnect = func(client mqtt.Client) {
		c.logger.Printf("Verbunden mit ThingsBoard MQTT (Client-ID: %s)", clientID)

		// Mit Verzögerung abonnieren
		go func() {
			time.Sleep(2 * time.Second)
			c.setupSubscriptions()
		}()
	}

	// OnConnectionLost Handler
	opts.OnConnectionLost = func(client mqtt.Client, err error) {
		c.logger.Printf("Verbindung zu ThingsBoard verloren: %v", err)
	}

	// OnReconnecting Handler
	opts.OnReconnecting = func(client mqtt.Client, opts *mqtt.ClientOptions) {
		c.logger.Println("Verbindung zu ThingsBoard wird wiederhergestellt...")
	}

	c.mqttClient = mqtt.NewClient(opts)

	c.logger.Printf("Verbinde zu MQTT-Broker: %s", broker)
	if token := c.mqttClient.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("Verbindungsfehler: %w", token.Error())
	}

	c.logger.Println("MQTT-Verbindung erfolgreich hergestellt")
	return nil
}

// IsConnected prüft, ob der Client verbunden ist.
func (c *Client) IsConnected() bool {
	return c.mqttClient != nil && c.mqttClient.IsConnected()
}

// setupSubscriptions konfiguriert alle notwendigen MQTT-Abonnements.
func (c *Client) setupSubscriptions() {
	if !c.IsConnected() {
		c.logger.Println("Kann Abonnements nicht einrichten - Client nicht verbunden")
		return
	}

	c.logger.Println("Richte MQTT-Abonnements ein...")

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
		token := c.mqttClient.Subscribe(t.topic, t.qos, t.handler)
		if token.Wait() && token.Error() != nil {
			c.logger.Printf("Fehler beim Abonnieren von %s: %v", t.desc, token.Error())
		} else {
			c.logger.Printf("Erfolgreich abonniert: %s", t.desc)
		}
	}

	// Nach einer kurzen Wartezeit die aktuellen Shared Attributes abfragen
	go func() {
		time.Sleep(1 * time.Second)
		c.RequestSharedAttributes(nil) // nil = alle Attribute abfragen
	}()
}

//
// Telemetrie-Funktionen
//

// SendTelemetry sendet Telemetriedaten an ThingsBoard.
func (c *Client) SendTelemetry(data map[string]interface{}) error {
	if !c.IsConnected() {
		return fmt.Errorf("nicht verbunden")
	}

	topic := "v1/devices/me/telemetry"
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("JSON-Marshalling-Fehler: %w", err)
	}

	token := c.mqttClient.Publish(topic, 1, false, payload)
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("Telemetrie-Senden fehlgeschlagen: %w", token.Error())
	}

	return nil
}

// SendTelemetryWithTs sendet Telemetriedaten mit einem Zeitstempel.
func (c *Client) SendTelemetryWithTs(data map[string]interface{}, ts int64) error {
	// Kopiere Daten mit Zeitstempel
	dataWithTs := make(map[string]interface{})
	for k, v := range data {
		dataWithTs[k] = v
	}
	dataWithTs["ts"] = ts

	return c.SendTelemetry(dataWithTs)
}

// BatchSendTelemetry sendet mehrere Telemetriedatensätze in einem Batch.
func (c *Client) BatchSendTelemetry(dataArray []map[string]interface{}) error {
	if !c.IsConnected() {
		return fmt.Errorf("nicht verbunden")
	}

	topic := "v1/devices/me/telemetry"
	payload, err := json.Marshal(dataArray)
	if err != nil {
		return fmt.Errorf("JSON-Marshalling-Fehler für Batch: %w", err)
	}

	token := c.mqttClient.Publish(topic, 1, false, payload)
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("Batch-Telemetrie-Senden fehlgeschlagen: %w", token.Error())
	}

	return nil
}

//
// Attribute-Funktionen
//

// PublishAttributes veröffentlicht Client-Attribute an ThingsBoard.
func (c *Client) PublishAttributes(attributes map[string]interface{}) error {
	if !c.IsConnected() {
		return fmt.Errorf("nicht verbunden")
	}

	topic := "v1/devices/me/attributes"
	payload, err := json.Marshal(attributes)
	if err != nil {
		return fmt.Errorf("JSON-Marshalling-Fehler für Attribute: %w", err)
	}

	token := c.mqttClient.Publish(topic, 1, false, payload)
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("Attribute-Veröffentlichung fehlgeschlagen: %w", token.Error())
	}

	// Update lokaler Cache
	c.attributesMutex.Lock()
	for k, v := range attributes {
		c.clientAttributes[k] = v
	}
	c.attributesMutex.Unlock()

	return nil
}

// RequestClientAttributes fordert Client-Attribute vom Server an.
func (c *Client) RequestClientAttributes(keys []string) error {
	return c.requestAttributesByType("client", keys)
}

// RequestSharedAttributes fordert Shared-Attribute vom Server an.
func (c *Client) RequestSharedAttributes(keys []string) error {
	return c.requestAttributesByType("shared", keys)
}

// requestAttributesByType fordert Attribute eines bestimmten Typs an.
func (c *Client) requestAttributesByType(attrType string, keys []string) error {
	if !c.IsConnected() {
		return fmt.Errorf("nicht verbunden")
	}

	// RequestID generieren
	c.requestIDMutex.Lock()
	requestID := fmt.Sprintf("%d", c.nextRequestID)
	c.nextRequestID++
	c.requestIDMutex.Unlock()

	requestTopic := fmt.Sprintf("v1/devices/me/attributes/request/%s", requestID)

	// Anfragedaten erstellen
	requestData := make(map[string]interface{})

	// Je nach Attributtyp den richtigen Schlüssel setzen
	var keyField string
	switch attrType {
	case "client":
		keyField = "clientKeys"
	case "shared":
		keyField = "sharedKeys"
	default:
		return fmt.Errorf("ungültiger Attributtyp: %s", attrType)
	}

	// Schlüssel als kommagetrennten String oder leer für alle Attribute
	if keys != nil && len(keys) > 0 {
		requestData[keyField] = stringJoin(keys, ",")
	} else {
		requestData[keyField] = ""
	}

	payload, err := json.Marshal(requestData)
	if err != nil {
		return fmt.Errorf("JSON-Marshalling-Fehler für Attributanfrage: %w", err)
	}

	c.logger.Printf("Fordere %s-Attribute an: %s", attrType, string(payload))

	token := c.mqttClient.Publish(requestTopic, 1, false, payload)
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("Attributanfrage fehlgeschlagen: %w", token.Error())
	}

	return nil
}

// handleAttributeUpdate verarbeitet eingehende Attribute-Updates.
func (c *Client) handleAttributeUpdate(client mqtt.Client, msg mqtt.Message) {
	c.logger.Printf("Attribut-Update empfangen: %s", string(msg.Payload()))

	var attributeData map[string]interface{}
	if err := json.Unmarshal(msg.Payload(), &attributeData); err != nil {
		c.logger.Printf("Fehler beim Unmarshalling der Attributdaten: %v", err)
		return
	}

	// Shared Attributes im lokalen Cache aktualisieren
	c.attributesMutex.Lock()
	for k, v := range attributeData {
		c.sharedAttributes[k] = v
	}
	c.attributesMutex.Unlock()

	// Auf Firmware-Updates prüfen
	c.checkForFirmwareUpdate(attributeData)

	// Callback aufrufen, wenn gesetzt
	if c.attributeCallback != nil {
		c.attributeCallback(attributeData)
	}
}

// checkForFirmwareUpdate prüft Attribut-Updates auf Firmware-Updates
func (c *Client) checkForFirmwareUpdate(attributes map[string]interface{}) {
	// Firmware-Update-Felder prüfen
	fwTitle, hasTitle := attributes["fw_title"].(string)
	fwVersion, hasVersion := attributes["fw_version"].(string)
	fwChecksum, hasChecksum := attributes["fw_checksum"].(string)
	fwAlgorithm, hasAlgorithm := attributes["fw_checksum_algorithm"].(string)

	// Wenn alle Firmware-Felder vorhanden sind und ein Callback registriert ist
	if hasTitle && hasVersion && hasChecksum && hasAlgorithm && c.firmwareCallback != nil {
		c.firmwareCallback(fwTitle, fwVersion, fwChecksum, fwAlgorithm)
	}
}

// handleAttributeResponse verarbeitet Antworten auf Attributanfragen.
func (c *Client) handleAttributeResponse(client mqtt.Client, msg mqtt.Message) {
	c.logger.Printf("Attribut-Antwort empfangen: %s von Topic: %s", string(msg.Payload()), msg.Topic())

	// RequestID aus dem Topic extrahieren
	topic := msg.Topic()
	requestID := topic[len("v1/devices/me/attributes/response/"):]

	var responseData map[string]interface{}
	if err := json.Unmarshal(msg.Payload(), &responseData); err != nil {
		c.logger.Printf("Fehler beim Unmarshalling der Attributantwort: %v", err)
		return
	}

	// Shared Attributes aktualisieren, wenn vorhanden
	if shared, ok := responseData["shared"].(map[string]interface{}); ok {
		c.attributesMutex.Lock()
		for k, v := range shared {
			c.sharedAttributes[k] = v
		}
		c.attributesMutex.Unlock()
	}

	// Client Attributes aktualisieren, wenn vorhanden
	if client, ok := responseData["client"].(map[string]interface{}); ok {
		c.attributesMutex.Lock()
		for k, v := range client {
			c.clientAttributes[k] = v
		}
		c.attributesMutex.Unlock()
	}

	// Callback aufrufen, wenn gesetzt
	if c.attributeCallback != nil {
		c.attributeCallback(responseData)
	}
}

//
// RPC-Funktionen
//

// handleRPCRequest verarbeitet eingehende RPC-Anfragen.
func (c *Client) handleRPCRequest(client mqtt.Client, msg mqtt.Message) {
	topicStr := msg.Topic()
	payload := string(msg.Payload())
	c.logger.Printf("RPC-Anfrage empfangen: %s von Topic: %s", payload, topicStr)

	// RequestID aus dem Topic extrahieren
	requestID := topicStr[len("v1/devices/me/rpc/request/"):]

	var rpcData map[string]interface{}
	if err := json.Unmarshal(msg.Payload(), &rpcData); err != nil {
		c.logger.Printf("Fehler beim Unmarshalling der RPC-Daten: %v", err)
		return
	}

	// Methode extrahieren
	method, ok := rpcData["method"].(string)
	if !ok {
		c.logger.Printf("RPC-Anfrage ohne method-Feld: %v", rpcData)
		return
	}

	// Parameter extrahieren
	params, ok := rpcData["params"].(map[string]interface{})
	if !ok {
		// Leere Parameter, wenn keine vorhanden
		params = make(map[string]interface{})
	}

	// Spezielle Methode getSessionLimits
	if method == "getSessionLimits" {
		c.handleGetSessionLimits(requestID)
		return
	}

	// RPC-Callback aufrufen, wenn gesetzt
	var response interface{}
	var err error

	if c.rpcCallback != nil {
		response, err = c.rpcCallback(method, params)
		if err != nil {
			response = map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			}
		}
	} else {
		// Standard-Antwort, wenn kein Callback gesetzt ist
		response = map[string]interface{}{
			"success": true,
			"result":  fmt.Sprintf("Methode %s wurde empfangen, aber nicht verarbeitet", method),
		}
	}

	// Antwort senden
	c.respondToRPC(requestID, response)
}

// handleGetSessionLimits behandelt die spezielle RPC-Methode getSessionLimits
func (c *Client) handleGetSessionLimits(requestID string) {
	// Standard-Limits für ThingsBoard
	limits := map[string]interface{}{
		"maxPayloadSize":      65536,
		"maxInflightMessages": 100,
		"rateLimits": map[string]interface{}{
			"messages":            "200:1,6000:60,14000:3600",
			"telemetryMessages":   "100:1,3000:60,7000:3600",
			"telemetryDataPoints": "200:1,6000:60,14000:3600",
		},
	}

	c.respondToRPC(requestID, limits)
}

// respondToRPC sendet eine Antwort auf eine RPC-Anfrage.
func (c *Client) respondToRPC(requestID string, response interface{}) {
	if !c.IsConnected() {
		c.logger.Println("Kann RPC-Antwort nicht senden - Client nicht verbunden")
		return
	}

	responseTopic := fmt.Sprintf("v1/devices/me/rpc/response/%s", requestID)

	payload, err := json.Marshal(response)
	if err != nil {
		c.logger.Printf("Fehler beim Marshalling der RPC-Antwort: %v", err)
		return
	}

	token := c.mqttClient.Publish(responseTopic, 1, false, payload)
	if token.Wait() && token.Error() != nil {
		c.logger.Printf("Fehler beim Senden der RPC-Antwort: %v", token.Error())
		return
	}

	c.logger.Printf("RPC-Antwort gesendet: %s", string(payload))
}

// SendRPCRequest sendet eine Client-seitige RPC-Anfrage an den Server.
func (c *Client) SendRPCRequest(method string, params map[string]interface{}) (interface{}, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("nicht verbunden")
	}

	// RequestID generieren
	c.requestIDMutex.Lock()
	requestID := fmt.Sprintf("%d", c.nextRequestID)
	c.nextRequestID++
	c.requestIDMutex.Unlock()

	requestTopic := fmt.Sprintf("v1/devices/me/rpc/request/%s", requestID)

	// Anfragekanal erstellen
	responseChan := make(chan interface{}, 1)
	c.requestIDMutex.Lock()
	c.pendingRequests[requestID] = responseChan
	c.requestIDMutex.Unlock()

	// Aufräumen bei Beendigung
	defer func() {
		c.requestIDMutex.Lock()
		delete(c.pendingRequests, requestID)
		c.requestIDMutex.Unlock()
	}()

	// Anfragedaten erstellen
	request := map[string]interface{}{
		"method": method,
		"params": params,
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("JSON-Marshalling-Fehler für RPC-Anfrage: %w", err)
	}

	token := c.mqttClient.Publish(requestTopic, 1, false, payload)
	if token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("RPC-Anfrage-Senden fehlgeschlagen: %w", token.Error())
	}

	// Auf Antwort warten (mit Timeout)
	select {
	case response := <-responseChan:
		return response, nil
	case <-time.After(10 * time.Second):
		return nil, fmt.Errorf("timeout beim Warten auf RPC-Antwort")
	}
}

//
// Firmware-Funktionen
//

// SetFirmwareUpdateCallback setzt den Callback für Firmware-Updates.
func (c *Client) SetFirmwareUpdateCallback(callback FirmwareUpdateCallback) {
	c.firmwareCallback = callback
}

// RequestFirmwareChunk fordert einen Chunk der Firmware an.
func (c *Client) RequestFirmwareChunk(requestId int, chunkIndex int, chunkSize int) ([]byte, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("nicht verbunden")
	}

	// Topic zusammenbauen
	requestTopic := fmt.Sprintf("v2/fw/request/%d/chunk/%d", requestId, chunkIndex)
	responseTopic := fmt.Sprintf("v2/fw/response/%d/chunk/%d", requestId, chunkIndex)

	// Kanal für die Antwort erstellen
	responseChan := make(chan []byte, 1)

	// Abonnieren der Antwort
	token := c.mqttClient.Subscribe(responseTopic, 1, func(client mqtt.Client, msg mqtt.Message) {
		responseChan <- msg.Payload()

		// Abonnement nach Erhalt der Antwort beenden
		c.mqttClient.Unsubscribe(responseTopic)
	})

	if token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("fehler beim Abonnieren des Firmware-Response-Topics: %w", token.Error())
	}

	// Anfrage senden (Payload ist die gewünschte Chunk-Größe)
	chunkSizeBytes, _ := json.Marshal(chunkSize)
	token = c.mqttClient.Publish(requestTopic, 1, false, chunkSizeBytes)
	if token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("fehler beim Senden der Firmware-Chunk-Anfrage: %w", token.Error())
	}

	// Auf Antwort warten (mit Timeout)
	select {
	case response := <-responseChan:
		return response, nil
	case <-time.After(30 * time.Second):
		return nil, fmt.Errorf("timeout beim Warten auf Firmware-Chunk")
	}
}

//
// Device-Provisioning-Funktionen
//

// ClaimDevice beansprucht ein Gerät.
func (c *Client) ClaimDevice(secretKey string, durationMs int64) error {
	if !c.IsConnected() {
		return fmt.Errorf("nicht verbunden")
	}

	topic := "v1/devices/me/claim"

	// Anfragedaten erstellen
	request := make(map[string]interface{})
	if secretKey != "" {
		request["secretKey"] = secretKey
	}
	if durationMs > 0 {
		request["durationMs"] = durationMs
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("JSON-Marshalling-Fehler für Claim-Anfrage: %w", err)
	}

	token := c.mqttClient.Publish(topic, 1, false, payload)
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("Claim-Anfrage fehlgeschlagen: %w", token.Error())
	}

	return nil
}

// ProvisionDevice stellt ein neues Gerät bereit.
func (c *Client) ProvisionDevice(deviceName, provisionKey, provisionSecret string) error {
	// Temporären MQTT-Client für die Provision-Anfrage erstellen
	opts := mqtt.NewClientOptions()
	broker := fmt.Sprintf("tcp://%s:%d", c.config.Host, c.config.Port)
	opts.AddBroker(broker)

	// Bei Provisioning muss der Benutzername "provision" sein
	opts.SetUsername("provision")
	opts.SetClientID("provision-" + deviceName)

	provClient := mqtt.NewClient(opts)
	if token := provClient.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("verbindung für Provisioning fehlgeschlagen: %w", token.Error())
	}

	defer provClient.Disconnect(250)

	// Anfragedaten erstellen
	request := map[string]interface{}{
		"deviceName":            deviceName,
		"provisionDeviceKey":    provisionKey,
		"provisionDeviceSecret": provisionSecret,
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("JSON-Marshalling-Fehler für Provision-Anfrage: %w", err)
	}

	// Anfrage an das spezielle Provisioning-Topic senden
	token := provClient.Publish("/provision", 1, false, payload)
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("provision-Anfrage fehlgeschlagen: %w", token.Error())
	}

	return nil
}

//
// Hilfsfunktionen
//

// Start beginnt das Senden von Daten aus dem Datenkanal.
func (c *Client) Start() {
	c.logger.Println("ThingsBoard-Client-Datenverarbeitung wird gestartet...")
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
			c.logger.Println("Datenverarbeitung wird beendet.")
			return
		}
	}
}

// processIncomingData verarbeitet eingehende Daten und sendet sie an ThingsBoard.
func (c *Client) processIncomingData(data map[string]interface{}) {
	if !c.IsConnected() {
		c.logger.Println("Kann Daten nicht senden - Client nicht verbunden")
		return
	}

	// Verarbeite unterschiedliche Datenformate
	if simpleData, ok := data["simple"].(map[string]interface{}); ok && len(simpleData) > 0 {
		if err := c.SendTelemetry(simpleData); err != nil {
			c.logger.Printf("Fehler beim Senden einfacher Telemetrie: %v", err)
		} else {
			c.logger.Printf("Einfache Telemetrie erfolgreich gesendet: %v", simpleData)
		}
	}

	if jsonData, ok := data["json"].(map[string]interface{}); ok && len(jsonData) > 0 {
		if err := c.SendTelemetry(jsonData); err != nil {
			c.logger.Printf("Fehler beim Senden von JSON-Telemetrie: %v", err)
		} else {
			c.logger.Printf("JSON-Telemetrie erfolgreich gesendet: %v", jsonData)
		}
	}
}

// Stop beendet den Client und gibt Ressourcen frei.
func (c *Client) Stop() {
	c.logger.Println("ThingsBoard-Client wird beendet...")
	close(c.stopChan)
	if c.mqttClient != nil && c.mqttClient.IsConnected() {
		c.mqttClient.Disconnect(250)
	}
	c.logger.Println("ThingsBoard-Client beendet.")
}

// SetAttributeCallback setzt den Callback für Attribut-Updates.
func (c *Client) SetAttributeCallback(callback AttrUpdateCallback) {
	c.attributeCallback = callback
}

// SetRPCCallback setzt den Callback für RPC-Anfragen.
func (c *Client) SetRPCCallback(callback RPCCallback) {
	c.rpcCallback = callback
}

// GetAttribute gibt ein Attribut aus dem lokalen Cache zurück.
func (c *Client) GetAttribute(key string, attrType string) (interface{}, bool) {
	c.attributesMutex.RLock()
	defer c.attributesMutex.RUnlock()

	var value interface{}
	var exists bool

	switch attrType {
	case "shared":
		value, exists = c.sharedAttributes[key]
	case "client":
		value, exists = c.clientAttributes[key]
	default:
		// Suche in beiden Caches
		value, exists = c.sharedAttributes[key]
		if !exists {
			value, exists = c.clientAttributes[key]
		}
	}

	return value, exists
}

// GetAllAttributes gibt alle Attribute eines bestimmten Typs zurück.
func (c *Client) GetAllAttributes(attrType string) map[string]interface{} {
	c.attributesMutex.RLock()
	defer c.attributesMutex.RUnlock()

	result := make(map[string]interface{})

	switch attrType {
	case "shared":
		for k, v := range c.sharedAttributes {
			result[k] = v
		}
	case "client":
		for k, v := range c.clientAttributes {
			result[k] = v
		}
	default:
		// Beide Typen zusammenführen
		for k, v := range c.sharedAttributes {
			result[k] = v
		}
		for k, v := range c.clientAttributes {
			result[k] = v
		}
	}

	return result
}

// stringJoin verbindet Strings mit einem Trennzeichen.
func stringJoin(arr []string, sep string) string {
	if len(arr) == 0 {
		return ""
	}

	result := arr[0]
	for i := 1; i < len(arr); i++ {
		result += sep + arr[i]
	}

	return result
}
