package thingsboardMQTT

import (
	"encoding/json"
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

//
// Firmware-Update-Funktionen
//

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
	token := c.getMQTTClient().Subscribe(responseTopic, 1, func(client mqtt.Client, msg mqtt.Message) {
		responseChan <- msg.Payload()

		// Abonnement nach Erhalt der Antwort beenden
		c.getMQTTClient().Unsubscribe(responseTopic)
	})

	if token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("fehler beim Abonnieren des Firmware-Response-Topics: %w", token.Error())
	}

	// Anfrage senden (Payload ist die gewünschte Chunk-Größe)
	chunkSizeBytes, _ := json.Marshal(chunkSize)
	token = c.getMQTTClient().Publish(requestTopic, 1, false, chunkSizeBytes)
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

// StartFirmwareUpdate initiiert einen Firmware-Update-Prozess.
// Diese Methode erleichtert das Herunterladen einer vollständigen Firmware in Chunks.
func (c *Client) StartFirmwareUpdate(requestId int, chunkSize int, callback func([]byte, int, error)) {
	var totalBytes int = 0
	var chunkIndex int = 0

	for {
		chunk, err := c.RequestFirmwareChunk(requestId, chunkIndex, chunkSize)

		// Auch bei Fehler den Callback aufrufen, damit er entsprechend reagieren kann
		callback(chunk, totalBytes, err)

		// Bei Fehler oder leerem Chunk (Ende der Firmware) abbrechen
		if err != nil || len(chunk) == 0 {
			break
		}

		totalBytes += len(chunk)
		chunkIndex++
	}

	c.Logger.Printf("Firmware-Update abgeschlossen. Insgesamt %d Bytes in %d Chunks heruntergeladen.",
		totalBytes, chunkIndex)
}
