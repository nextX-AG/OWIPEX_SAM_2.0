package thingsboardMQTT

import (
	"encoding/json"
	"fmt"
)

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

	token := c.getMQTTClient().Publish(topic, 1, false, payload)
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

	token := c.getMQTTClient().Publish(topic, 1, false, payload)
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("Batch-Telemetrie-Senden fehlgeschlagen: %w", token.Error())
	}

	return nil
}
