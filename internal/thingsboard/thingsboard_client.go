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

// Client manages the MQTT connection and data publishing to ThingsBoard.
type Client struct {
	mqttClient        mqtt.Client
	logger            *log.Logger
	config            config.ThingsBoardConfig
	stopChan          chan struct{}
	dataChan          <-chan map[string]interface{}        // Receives data from SensorManager
	sharedAttributes  map[string]interface{}               // Speichert empfangene Shared Attributes
	attributesMutex   sync.RWMutex                         // Mutex für Thread-safe Zugriff auf Attributes
	attributeCallback func(map[string]interface{})         // Callback für neue Attribute
	rpcCallback       func(string, map[string]interface{}) // Callback für RPC Anfragen
}

// NewClient creates a new ThingsBoard MQTT client.
func NewClient(cfg config.ThingsBoardConfig, dataChan <-chan map[string]interface{}) *Client {
	logger := log.New(os.Stdout, "[ThingsBoardClient] ", log.LstdFlags)
	return &Client{
		logger:           logger,
		config:           cfg,
		stopChan:         make(chan struct{}),
		dataChan:         dataChan,
		sharedAttributes: make(map[string]interface{}),
	}
}

// Connect establishes the MQTT connection to ThingsBoard.
func (c *Client) Connect() error {
	opts := mqtt.NewClientOptions()
	broker := fmt.Sprintf("tcp://%s:%d", c.config.Host, c.config.Port) // Changed c.config.Server to c.config.Host
	opts.AddBroker(broker)
	opts.SetClientID(fmt.Sprintf("go-rs485-reader-%d", time.Now().UnixNano()))
	opts.SetUsername(c.config.AccessToken) // ThingsBoard uses Access Token as username
	// opts.SetPassword("") // No password typically needed when using Access Token as username
	opts.SetDefaultPublishHandler(func(client mqtt.Client, msg mqtt.Message) {
		c.logger.Printf("Received unexpected message: %s from topic: %s\n", msg.Payload(), msg.Topic())
	})
	opts.OnConnect = func(client mqtt.Client) {
		c.logger.Println("Connected to ThingsBoard MQTT broker")

		// Subscribe to shared attributes updates
		if token := client.Subscribe("v1/devices/me/attributes", 1, c.handleAttributeUpdate); token.Wait() && token.Error() != nil {
			c.logger.Printf("Failed to subscribe to attribute updates: %v", token.Error())
		} else {
			c.logger.Println("Subscribed to attribute updates")
		}

		// Subscribe to shared attributes responses
		if token := client.Subscribe("v1/devices/me/attributes/response/+", 1, c.handleAttributeResponse); token.Wait() && token.Error() != nil {
			c.logger.Printf("Failed to subscribe to attribute responses: %v", token.Error())
		} else {
			c.logger.Println("Subscribed to attribute responses")
		}

		// Subscribe to RPC requests
		if token := client.Subscribe("v1/devices/me/rpc/request/+", 1, c.handleRPCRequest); token.Wait() && token.Error() != nil {
			c.logger.Printf("Failed to subscribe to RPC requests: %v", token.Error())
		} else {
			c.logger.Println("Subscribed to RPC requests")
		}

		// Request current shared attributes after connection
		c.requestSharedAttributes()
	}
	opts.OnConnectionLost = func(client mqtt.Client, err error) {
		c.logger.Printf("Connection lost to ThingsBoard MQTT broker: %v", err)
		// Implement reconnection logic if desired
	}
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(10 * time.Second)

	c.mqttClient = mqtt.NewClient(opts)
	if token := c.mqttClient.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to connect to ThingsBoard: %w", token.Error())
	}
	return nil
}

// handleAttributeUpdate handles incoming attribute updates
func (c *Client) handleAttributeUpdate(client mqtt.Client, msg mqtt.Message) {
	c.logger.Printf("Received attribute update: %s", string(msg.Payload()))

	var attributeData map[string]interface{}
	if err := json.Unmarshal(msg.Payload(), &attributeData); err != nil {
		c.logger.Printf("Error unmarshalling attribute data: %v", err)
		return
	}

	// Update local cache of attributes
	c.attributesMutex.Lock()
	for k, v := range attributeData {
		c.sharedAttributes[k] = v
	}
	c.attributesMutex.Unlock()

	// Call attribute callback if set
	if c.attributeCallback != nil {
		c.attributeCallback(attributeData)
	}
}

// handleAttributeResponse handles responses to attribute requests
func (c *Client) handleAttributeResponse(client mqtt.Client, msg mqtt.Message) {
	c.logger.Printf("Received attribute response: %s", string(msg.Payload()))

	var attributeData map[string]interface{}
	if err := json.Unmarshal(msg.Payload(), &attributeData); err != nil {
		c.logger.Printf("Error unmarshalling attribute response: %v", err)
		return
	}

	// Update local cache of attributes
	c.attributesMutex.Lock()
	for k, v := range attributeData {
		c.sharedAttributes[k] = v
	}
	c.attributesMutex.Unlock()

	// Call attribute callback if set
	if c.attributeCallback != nil {
		c.attributeCallback(attributeData)
	}
}

// handleRPCRequest handles incoming RPC requests
func (c *Client) handleRPCRequest(client mqtt.Client, msg mqtt.Message) {
	c.logger.Printf("Received RPC request: %s from topic: %s", string(msg.Payload()), msg.Topic())

	// Extract request ID from topic (format: v1/devices/me/rpc/request/{requestId})
	topicParts := []rune(msg.Topic())
	requestID := string(topicParts[len("v1/devices/me/rpc/request/"):])

	var rpcData map[string]interface{}
	if err := json.Unmarshal(msg.Payload(), &rpcData); err != nil {
		c.logger.Printf("Error unmarshalling RPC data: %v", err)
		return
	}

	// Extract method
	method, ok := rpcData["method"].(string)
	if !ok {
		c.logger.Printf("RPC request missing method field: %v", rpcData)
		return
	}

	// Extract params
	params, _ := rpcData["params"].(map[string]interface{})

	// Call RPC callback if set
	if c.rpcCallback != nil {
		c.rpcCallback(method, params)
	}

	// Respond to the RPC request
	c.respondToRPCRequest(requestID, map[string]interface{}{
		"success": true,
		"result":  fmt.Sprintf("Method %s processed", method),
	})
}

// respondToRPCRequest sends a response to an RPC request
func (c *Client) respondToRPCRequest(requestID string, response interface{}) {
	responseTopic := fmt.Sprintf("v1/devices/me/rpc/response/%s", requestID)

	responseData, err := json.Marshal(response)
	if err != nil {
		c.logger.Printf("Error marshalling RPC response: %v", err)
		return
	}

	token := c.mqttClient.Publish(responseTopic, 1, false, responseData)
	go func() {
		if token.Wait() && token.Error() != nil {
			c.logger.Printf("Error publishing RPC response: %v", token.Error())
		}
	}()
}

// requestSharedAttributes requests current values of shared attributes
func (c *Client) requestSharedAttributes() {
	requestID := fmt.Sprintf("%d", time.Now().UnixNano())
	requestTopic := fmt.Sprintf("v1/devices/me/attributes/request/%s", requestID)

	// Request all shared attributes
	requestData := map[string]interface{}{
		"sharedKeys": "",
	}

	payload, err := json.Marshal(requestData)
	if err != nil {
		c.logger.Printf("Error marshalling attribute request: %v", err)
		return
	}

	token := c.mqttClient.Publish(requestTopic, 1, false, payload)
	go func() {
		if token.Wait() && token.Error() != nil {
			c.logger.Printf("Error publishing attribute request: %v", token.Error())
		}
	}()
}

// SetAttributeCallback sets a callback function that will be called when shared attributes are updated
func (c *Client) SetAttributeCallback(callback func(map[string]interface{})) {
	c.attributeCallback = callback
}

// SetRPCCallback sets a callback function that will be called when RPC requests are received
func (c *Client) SetRPCCallback(callback func(string, map[string]interface{})) {
	c.rpcCallback = callback
}

// GetAttribute gets the value of a shared attribute
func (c *Client) GetAttribute(key string) (interface{}, bool) {
	c.attributesMutex.RLock()
	defer c.attributesMutex.RUnlock()
	value, exists := c.sharedAttributes[key]
	return value, exists
}

// GetAllAttributes returns a copy of all shared attributes
func (c *Client) GetAllAttributes() map[string]interface{} {
	c.attributesMutex.RLock()
	defer c.attributesMutex.RUnlock()

	attributes := make(map[string]interface{}, len(c.sharedAttributes))
	for k, v := range c.sharedAttributes {
		attributes[k] = v
	}

	return attributes
}

// Start begins listening for data from the SensorManager and publishing it.
func (c *Client) Start() {
	c.logger.Println("Starting ThingsBoard client data publishing loop...")
	go func() {
		for {
			select {
			case dataPayload := <-c.dataChan:
				if dataPayload == nil {
					c.logger.Println("Received nil data payload, skipping.")
					continue
				}
				c.publishData(dataPayload)
			case <-c.stopChan:
				c.logger.Println("ThingsBoard client data publishing loop stopping.")
				return
			}
		}
	}()
}

func (c *Client) publishData(dataPayload map[string]interface{}) {
	if !c.mqttClient.IsConnected() {
		c.logger.Println("MQTT client not connected, attempting to send later or reconnect.")
		// Potentially queue data or rely on auto-reconnect
		return
	}

	// The Python project sends telemetry for 'simple' and 'json' formats separately.
	// The SensorManager now prepares a map: {"simple": {...}, "json": {...}}

	if simpleData, ok := dataPayload["simple"].(map[string]interface{}); ok && len(simpleData) > 0 {
		topic := "v1/devices/me/telemetry"
		jsonData, err := json.Marshal(simpleData)
		if err != nil {
			c.logger.Printf("Error marshalling simple telemetry data: %v", err)
			return
		}
		token := c.mqttClient.Publish(topic, 1, false, jsonData)
		// token.Wait() // Can wait for confirmation, or publish asynchronously
		go func(t mqtt.Token, d []byte) {
			_ = t.WaitTimeout(5 * time.Second) // Wait for a reasonable time
			if t.Error() != nil {
				c.logger.Printf("Failed to publish simple telemetry: %v. Data: %s", t.Error(), string(d))
			} else {
				c.logger.Printf("Successfully published simple telemetry: %s", string(d))
			}
		}(token, jsonData)
	}

	if jsonDataMap, ok := dataPayload["json"].(map[string]interface{}); ok && len(jsonDataMap) > 0 {
		// In the Python project, the JSON data is sent with the sensor_id as the key
		// e.g., {"sensor_id_data": { "info": ..., "measurements": ...}}
		// The SensorManager already formats it this way.
		topic := "v1/devices/me/telemetry"
		jsonData, err := json.Marshal(jsonDataMap) // Marshal the whole map under "json"
		if err != nil {
			c.logger.Printf("Error marshalling JSON telemetry data: %v", err)
			return
		}
		token := c.mqttClient.Publish(topic, 1, false, jsonData)
		go func(t mqtt.Token, d []byte) {
			_ = t.WaitTimeout(5 * time.Second)
			if t.Error() != nil {
				c.logger.Printf("Failed to publish JSON telemetry: %v. Data: %s", t.Error(), string(d))
			} else {
				c.logger.Printf("Successfully published JSON telemetry: %s", string(d))
			}
		}(token, jsonData)
	}
}

// Stop disconnects the MQTT client.
func (c *Client) Stop() {
	c.logger.Println("Stopping ThingsBoard client...")
	close(c.stopChan)
	if c.mqttClient != nil && c.mqttClient.IsConnected() {
		c.mqttClient.Disconnect(250) // 250ms timeout for disconnection
	}
	c.logger.Println("ThingsBoard client stopped.")
}
