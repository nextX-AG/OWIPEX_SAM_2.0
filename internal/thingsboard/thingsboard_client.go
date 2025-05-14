package thingsboard

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"owipex_reader/internal/config"
)

// Client manages the MQTT connection and data publishing to ThingsBoard.
type Client struct {
	mqttClient mqtt.Client
	logger     *log.Logger
	config     config.ThingsBoardConfig
	stopChan   chan struct{}
	dataChan   <-chan map[string]interface{} // Receives data from SensorManager
}

// NewClient creates a new ThingsBoard MQTT client.
func NewClient(cfg config.ThingsBoardConfig, dataChan <-chan map[string]interface{}) *Client {
	logger := log.New(os.Stdout, "[ThingsBoardClient] ", log.LstdFlags)
	return &Client{
		logger:   logger,
		config:   cfg,
		stopChan: make(chan struct{}),
		dataChan: dataChan,
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

