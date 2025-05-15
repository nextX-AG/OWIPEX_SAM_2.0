package thingsboardMQTT

import (
	"encoding/json"
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

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

	token := c.getMQTTClient().Publish(topic, 1, false, payload)
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("Claim-Anfrage fehlgeschlagen: %w", token.Error())
	}

	return nil
}

// ProvisionDevice stellt ein neues Gerät bereit.
func (c *Client) ProvisionDevice(deviceName, provisionKey, provisionSecret string) error {
	// Temporären MQTT-Client für die Provision-Anfrage erstellen
	opts := mqtt.NewClientOptions()
	broker := fmt.Sprintf("tcp://%s:%d", c.Config.Host, c.Config.Port)
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
