package thingsboardMQTT

import (
	"encoding/json"
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

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

	token := c.getMQTTClient().Publish(topic, 1, false, payload)
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("Attribute-Veröffentlichung fehlgeschlagen: %w", token.Error())
	}

	// Update lokaler Cache
	c.threadSafety.AttributesMutex.Lock()
	for k, v := range attributes {
		c.clientAttributes[k] = v
	}
	c.threadSafety.AttributesMutex.Unlock()

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
	c.threadSafety.RequestIDMutex.Lock()
	requestID := fmt.Sprintf("%d", c.nextRequestID)
	c.nextRequestID++
	c.threadSafety.RequestIDMutex.Unlock()

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

	c.Logger.Printf("Fordere %s-Attribute an: %s", attrType, string(payload))

	token := c.getMQTTClient().Publish(requestTopic, 1, false, payload)
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("Attributanfrage fehlgeschlagen: %w", token.Error())
	}

	return nil
}

// handleAttributeUpdate verarbeitet eingehende Attribute-Updates.
func (c *Client) handleAttributeUpdate(client mqtt.Client, msg mqtt.Message) {
	c.Logger.Printf("Attribut-Update empfangen: %s", string(msg.Payload()))

	var attributeData map[string]interface{}
	if err := json.Unmarshal(msg.Payload(), &attributeData); err != nil {
		c.Logger.Printf("Fehler beim Unmarshalling der Attributdaten: %v", err)
		return
	}

	// Shared Attributes im lokalen Cache aktualisieren
	c.threadSafety.AttributesMutex.Lock()
	for k, v := range attributeData {
		c.sharedAttributes[k] = v
	}
	c.threadSafety.AttributesMutex.Unlock()

	// Auf Firmware-Updates prüfen
	c.checkForFirmwareUpdate(attributeData)

	// Callback aufrufen, wenn gesetzt
	if c.attributeCallback != nil {
		c.attributeCallback(attributeData)
	}
}

// handleAttributeResponse verarbeitet Antworten auf Attributanfragen.
func (c *Client) handleAttributeResponse(client mqtt.Client, msg mqtt.Message) {
	c.Logger.Printf("Attribut-Antwort empfangen: %s von Topic: %s", string(msg.Payload()), msg.Topic())

	// RequestID aus dem Topic extrahieren - wird aktuell nicht verwendet, könnte aber später nützlich sein
	topic := msg.Topic()
	_ = topic[len("v1/devices/me/attributes/response/"):]

	var responseData map[string]interface{}
	if err := json.Unmarshal(msg.Payload(), &responseData); err != nil {
		c.Logger.Printf("Fehler beim Unmarshalling der Attributantwort: %v", err)
		return
	}

	// Shared Attributes aktualisieren, wenn vorhanden
	if shared, ok := responseData["shared"].(map[string]interface{}); ok {
		c.threadSafety.AttributesMutex.Lock()
		for k, v := range shared {
			c.sharedAttributes[k] = v
		}
		c.threadSafety.AttributesMutex.Unlock()
	}

	// Client Attributes aktualisieren, wenn vorhanden
	if client, ok := responseData["client"].(map[string]interface{}); ok {
		c.threadSafety.AttributesMutex.Lock()
		for k, v := range client {
			c.clientAttributes[k] = v
		}
		c.threadSafety.AttributesMutex.Unlock()
	}

	// Callback aufrufen, wenn gesetzt
	if c.attributeCallback != nil {
		c.attributeCallback(responseData)
	}
}

// GetAttribute gibt ein Attribut aus dem lokalen Cache zurück.
func (c *Client) GetAttribute(key string, attrType string) (interface{}, bool) {
	c.threadSafety.AttributesMutex.RLock()
	defer c.threadSafety.AttributesMutex.RUnlock()

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
	c.threadSafety.AttributesMutex.RLock()
	defer c.threadSafety.AttributesMutex.RUnlock()

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
