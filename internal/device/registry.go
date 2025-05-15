// Package device implementiert die Geräteregistry und -verwaltung.
package device

import (
	"fmt"
	"sync"

	"owipex_reader/internal/types"
)

// Registry ist ein Thread-sicheres Register für alle im System verfügbaren Geräte.
// Es ermöglicht das Hinzufügen, Entfernen und Abrufen von Geräten.
type Registry struct {
	devices  map[string]types.Device
	mutex    sync.RWMutex
	handlers map[string]DeviceEventHandler
}

// DeviceEventHandler ist ein Callback-Typ für Geräteereignisse
type DeviceEventHandler func(event types.DeviceEvent)

// DeviceEventType definiert die Art des Geräteereignisses
type DeviceEventType string

const (
	// EventAdded wird ausgelöst, wenn ein Gerät hinzugefügt wird
	EventAdded DeviceEventType = "ADDED"
	// EventRemoved wird ausgelöst, wenn ein Gerät entfernt wird
	EventRemoved DeviceEventType = "REMOVED"
	// EventUpdated wird ausgelöst, wenn ein Gerät aktualisiert wird
	EventUpdated DeviceEventType = "UPDATED"
	// EventEnabled wird ausgelöst, wenn ein Gerät aktiviert oder deaktiviert wird
	EventEnabled DeviceEventType = "ENABLED"
)

// DeviceEvent repräsentiert ein Ereignis in Bezug auf ein Gerät
type DeviceEvent struct {
	Type     DeviceEventType
	DeviceID string
	Device   types.Device
	Data     map[string]interface{}
}

// NewRegistry erstellt eine neue Geräteregistry
func NewRegistry() *Registry {
	return &Registry{
		devices:  make(map[string]types.Device),
		handlers: make(map[string]DeviceEventHandler),
	}
}

// AddDevice fügt ein Gerät zur Registry hinzu
func (r *Registry) AddDevice(device types.Device) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	id := device.ID()
	if _, exists := r.devices[id]; exists {
		return fmt.Errorf("gerät mit ID %s ist bereits registriert", id)
	}

	r.devices[id] = device
	r.notifyHandlers(types.DeviceEvent{
		Type:     types.EventAdded,
		DeviceID: id,
		Device:   device,
	})

	return nil
}

// RemoveDevice entfernt ein Gerät aus der Registry
func (r *Registry) RemoveDevice(id string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	device, exists := r.devices[id]
	if !exists {
		return fmt.Errorf("gerät mit ID %s nicht gefunden", id)
	}

	delete(r.devices, id)
	r.notifyHandlers(types.DeviceEvent{
		Type:     types.EventRemoved,
		DeviceID: id,
		Device:   device,
	})

	return nil
}

// GetDevice gibt ein Gerät aus der Registry zurück
func (r *Registry) GetDevice(id string) (types.Device, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	device, exists := r.devices[id]
	if !exists {
		return nil, fmt.Errorf("gerät mit ID %s nicht gefunden", id)
	}

	return device, nil
}

// GetDevices gibt alle Geräte eines bestimmten Typs zurück
func (r *Registry) GetDevices(deviceType types.DeviceType) []types.Device {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var devices []types.Device
	for _, device := range r.devices {
		if deviceType == "" || device.Type() == deviceType {
			devices = append(devices, device)
		}
	}

	return devices
}

// GetAllDevices gibt alle Geräte zurück
func (r *Registry) GetAllDevices() []types.Device {
	return r.GetDevices("")
}

// GetSensors gibt alle Sensoren zurück
func (r *Registry) GetSensors() []types.Sensor {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var sensors []types.Sensor
	for _, device := range r.devices {
		if device.Type() == types.TypeSensor {
			if sensor, ok := device.(types.Sensor); ok {
				sensors = append(sensors, sensor)
			}
		}
	}

	return sensors
}

// GetActors gibt alle Aktoren zurück
func (r *Registry) GetActors() []types.Actor {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var actors []types.Actor
	for _, device := range r.devices {
		if device.Type() == types.TypeActor {
			if actor, ok := device.(types.Actor); ok {
				actors = append(actors, actor)
			}
		}
	}

	return actors
}

// RegisterHandler registriert einen Event-Handler mit einer eindeutigen ID
func (r *Registry) RegisterHandler(id string, handler DeviceEventHandler) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.handlers[id] = handler
}

// UnregisterHandler entfernt einen Event-Handler
func (r *Registry) UnregisterHandler(id string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	delete(r.handlers, id)
}

// notifyHandlers informiert alle registrierten Handler über ein Ereignis
func (r *Registry) notifyHandlers(event types.DeviceEvent) {
	for _, handler := range r.handlers {
		go handler(event)
	}
}

// Close schließt alle Geräte und gibt Ressourcen frei
func (r *Registry) Close() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for id, device := range r.devices {
		if err := device.Close(); err != nil {
			fmt.Printf("Fehler beim Schließen des Geräts %s: %v\n", id, err)
		}
	}

	r.devices = make(map[string]types.Device)
	r.handlers = make(map[string]DeviceEventHandler)
}
