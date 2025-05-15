// Package thingsboardMQTT implementiert einen umfassenden Client für die ThingsBoard IoT-Plattform.
// Diese Implementierung bietet Zugriff auf alle MQTT-APIs, die ThingsBoard bereitstellt.
package thingsboardMQTT

import (
	"log"
	"sync"

	"owipex_reader/internal/config"
)

// Client ist die Hauptschnittstelle für die ThingsBoard-Kommunikation.
// Er stellt alle Methoden bereit, um mit der ThingsBoard-Plattform zu interagieren.
type Client struct {
	Logger *log.Logger
	Config config.ThingsBoardConfig

	// Interne Felder, die von verschiedenen Modulen genutzt werden
	mqttClient interface{} // Tatsächlicher Typ: mqtt.Client
	stopChan   chan struct{}
	dataChan   <-chan map[string]interface{}

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

	// Mutex für Thread-Sicherheit, als Embed-Struct
	*threadSafety
}

// threadSafety kapselt alle Mutex, die der Client verwendet
type threadSafety struct {
	AttributesMutex sync.RWMutex
	DeviceInfoMutex sync.RWMutex
	RequestIDMutex  sync.Mutex
}

// Callback-Typen für verschiedene Events
type AttrUpdateCallback func(map[string]interface{})
type RPCCallback func(string, map[string]interface{}) (interface{}, error)
type FirmwareUpdateCallback func(title, version, checksum, algorithm string)

// ClientOption ist ein Funktionstyp für das Options-Pattern des Clients
type ClientOption func(*Client)
