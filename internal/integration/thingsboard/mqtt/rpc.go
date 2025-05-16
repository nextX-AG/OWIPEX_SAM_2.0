package thingsboardMQTT

import (
	"encoding/json"
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

//
// RPC-Funktionen
//

// handleRPCRequest verarbeitet eingehende RPC-Anfragen.
func (c *Client) handleRPCRequest(client mqtt.Client, msg mqtt.Message) {
	topicStr := msg.Topic()
	payload := string(msg.Payload())
	c.Logger.Printf("RPC-Anfrage empfangen: %s von Topic: %s", payload, topicStr)

	// RequestID aus dem Topic extrahieren
	requestID := topicStr[len("v1/devices/me/rpc/request/"):]

	var rpcData map[string]interface{}
	if err := json.Unmarshal(msg.Payload(), &rpcData); err != nil {
		c.Logger.Printf("Fehler beim Unmarshalling der RPC-Daten: %v", err)
		return
	}

	// Methode extrahieren
	method, ok := rpcData["method"].(string)
	if !ok {
		c.Logger.Printf("RPC-Anfrage ohne method-Feld: %v", rpcData)
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
		c.Logger.Println("Kann RPC-Antwort nicht senden - Client nicht verbunden")
		return
	}

	responseTopic := fmt.Sprintf("v1/devices/me/rpc/response/%s", requestID)

	payload, err := json.Marshal(response)
	if err != nil {
		c.Logger.Printf("Fehler beim Marshalling der RPC-Antwort: %v", err)
		return
	}

	token := c.getMQTTClient().Publish(responseTopic, 1, false, payload)
	if token.Wait() && token.Error() != nil {
		c.Logger.Printf("Fehler beim Senden der RPC-Antwort: %v", token.Error())
		return
	}

	c.Logger.Printf("RPC-Antwort gesendet: %s", string(payload))
}

// SendRPCRequest sendet eine Client-seitige RPC-Anfrage an den Server.
func (c *Client) SendRPCRequest(method string, params map[string]interface{}) (interface{}, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("nicht verbunden")
	}

	// RequestID generieren
	c.threadSafety.RequestIDMutex.Lock()
	requestID := fmt.Sprintf("%d", c.nextRequestID)
	c.nextRequestID++
	c.threadSafety.RequestIDMutex.Unlock()

	requestTopic := fmt.Sprintf("v1/devices/me/rpc/request/%s", requestID)
	responseTopic := fmt.Sprintf("v1/devices/me/rpc/response/%s", requestID)

	// Anfragekanal erstellen
	responseChan := make(chan interface{}, 1)
	c.threadSafety.RequestIDMutex.Lock()
	c.pendingRequests[requestID] = responseChan
	c.threadSafety.RequestIDMutex.Unlock()

	// Aufräumen bei Beendigung
	defer func() {
		c.threadSafety.RequestIDMutex.Lock()
		delete(c.pendingRequests, requestID)
		c.threadSafety.RequestIDMutex.Unlock()
	}()

	// Response-Handler für diesen speziellen Request einrichten
	responseHandler := func(client mqtt.Client, msg mqtt.Message) {
		var response interface{}
		if err := json.Unmarshal(msg.Payload(), &response); err != nil {
			c.Logger.Printf("Fehler beim Unmarshalling der RPC-Antwort: %v", err)
			return
		}

		// Antwort in den Kanal schreiben
		responseChan <- response

		// Abonnement beenden
		c.getMQTTClient().Unsubscribe(responseTopic)
	}

	// Response-Topic abonnieren
	if token := c.getMQTTClient().Subscribe(responseTopic, 1, responseHandler); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("fehler beim Abonnieren des Response-Topics: %w", token.Error())
	}

	// Anfragedaten erstellen
	request := map[string]interface{}{
		"method": method,
		"params": params,
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("JSON-Marshalling-Fehler für RPC-Anfrage: %w", err)
	}

	token := c.getMQTTClient().Publish(requestTopic, 1, false, payload)
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
