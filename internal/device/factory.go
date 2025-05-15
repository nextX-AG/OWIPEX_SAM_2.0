// Package device implementiert eine Factory für die Erstellung von Geräten.
package device

import (
	"fmt"
	"sync"

	"owipex_reader/internal/types"
)

// DeviceCreator ist eine Funktion, die ein Gerät anhand einer Konfiguration erstellt
type DeviceCreator func(config types.DeviceConfig) (types.Device, error)

// Factory erstellt Geräte basierend auf ihrer Konfiguration
type Factory struct {
	creators map[string]DeviceCreator
	mutex    sync.RWMutex
}

// NewFactory erstellt eine neue Gerätefabrik
func NewFactory() *Factory {
	return &Factory{
		creators: make(map[string]DeviceCreator),
	}
}

// RegisterCreator registriert einen neuen Geräte-Creator für einen bestimmten Typ
func (f *Factory) RegisterCreator(deviceType string, creator DeviceCreator) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.creators[deviceType] = creator
}

// CreateDevice erstellt ein Gerät basierend auf der Konfiguration
func (f *Factory) CreateDevice(config types.DeviceConfig) (types.Device, error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	deviceType := config.Type
	creator, exists := f.creators[deviceType]
	if !exists {
		return nil, fmt.Errorf("kein Creator für Gerätetyp '%s' registriert", deviceType)
	}

	return creator(config)
}

// CreateDevices erstellt mehrere Geräte aus einem Array von Konfigurationen
func (f *Factory) CreateDevices(configs []types.DeviceConfig) ([]types.Device, []error) {
	var devices []types.Device
	var errors []error

	for _, config := range configs {
		device, err := f.CreateDevice(config)
		if err != nil {
			errors = append(errors, fmt.Errorf("fehler beim Erstellen von Gerät '%s': %w", config.ID, err))
			continue
		}

		devices = append(devices, device)
	}

	return devices, errors
}

// CreateAndRegisterDevices erstellt Geräte und registriert sie in der Registry
func (f *Factory) CreateAndRegisterDevices(configs []types.DeviceConfig, registry *Registry) ([]types.Device, []error) {
	var devices []types.Device
	var errors []error

	for _, config := range configs {
		device, err := f.CreateDevice(config)
		if err != nil {
			errors = append(errors, fmt.Errorf("fehler beim Erstellen von Gerät '%s': %w", config.ID, err))
			continue
		}

		if err := registry.AddDevice(device); err != nil {
			errors = append(errors, fmt.Errorf("fehler beim Registrieren von Gerät '%s': %w", config.ID, err))
			continue
		}

		devices = append(devices, device)
	}

	return devices, errors
}
