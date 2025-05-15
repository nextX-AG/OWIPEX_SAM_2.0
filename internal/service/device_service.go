// Package service enthält Dienste, die die verschiedenen Komponenten des Systems verbinden.
package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"owipex_reader/internal/device/creator"
	"owipex_reader/internal/types"
)

// DeviceService verwaltet die Erstellung und Verwaltung von Geräten
type DeviceService struct {
	sensorRegistry *creator.SensorRegistry
	configPath     string
}

// NewDeviceService erstellt einen neuen DeviceService
func NewDeviceService(configPath string) *DeviceService {
	registry := creator.NewSensorRegistry()

	return &DeviceService{
		sensorRegistry: registry,
		configPath:     configPath,
	}
}

// Initialize initialisiert den Service und registriert Sensor-Typen
func (s *DeviceService) Initialize() error {
	// Sensor-Typen registrieren
	creator.RegisterAllSensorTypes(s.sensorRegistry)

	return nil
}

// LoadSensorsFromConfig lädt Sensoren aus Konfigurationsdateien
func (s *DeviceService) LoadSensorsFromConfig() ([]types.Sensor, error) {
	var sensorConfigs []types.DeviceConfig
	var loadingErrors []error

	// Verzeichnisse für verschiedene Sensortypen
	sensorDirs := []string{
		filepath.Join(s.configPath, "sensors", "ph"),
		filepath.Join(s.configPath, "sensors", "flow"),
		filepath.Join(s.configPath, "sensors", "radar"),
		filepath.Join(s.configPath, "sensors", "turbidity"),
	}

	// Konfigurationsdateien aus allen Verzeichnissen laden
	for _, dir := range sensorDirs {
		configs, err := loadConfigFilesFromDir(dir)
		if err != nil {
			loadingErrors = append(loadingErrors, fmt.Errorf("fehler beim Laden aus %s: %w", dir, err))
			continue
		}
		sensorConfigs = append(sensorConfigs, configs...)
	}

	// Sensoren aus Konfigurationen erstellen
	sensors, errs := s.sensorRegistry.CreateSensors(sensorConfigs)
	for _, err := range errs {
		loadingErrors = append(loadingErrors, err)
	}

	// Wenn Fehler aufgetreten sind, diese zusammenfassen
	if len(loadingErrors) > 0 {
		fmt.Println("Fehler beim Laden der Sensor-Konfigurationen:")
		for _, err := range loadingErrors {
			fmt.Printf("  - %v\n", err)
		}
	}

	return sensors, nil
}

// loadConfigFilesFromDir lädt alle JSON-Konfigurationsdateien aus einem Verzeichnis
func loadConfigFilesFromDir(dirPath string) ([]types.DeviceConfig, error) {
	var configs []types.DeviceConfig

	// Verzeichnis lesen
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		// Wenn das Verzeichnis nicht existiert, ist das kein kritischer Fehler
		if os.IsNotExist(err) {
			return configs, nil
		}
		return nil, fmt.Errorf("fehler beim Lesen des Verzeichnisses %s: %w", dirPath, err)
	}

	// Alle JSON-Dateien verarbeiten
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		filePath := filepath.Join(dirPath, entry.Name())
		config, err := loadConfigFile(filePath)
		if err != nil {
			fmt.Printf("Warnung: Fehler beim Laden von %s: %v\n", filePath, err)
			continue
		}

		configs = append(configs, *config)
	}

	return configs, nil
}

// loadConfigFile lädt eine einzelne Konfigurationsdatei
func loadConfigFile(filePath string) (*types.DeviceConfig, error) {
	// Datei öffnen
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("fehler beim Öffnen der Datei: %w", err)
	}
	defer file.Close()

	// Konfiguration aus der Datei lesen
	var config types.DeviceConfig
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("fehler beim Dekodieren der Konfiguration: %w", err)
	}

	// Konfiguration validieren
	if config.ID == "" {
		return nil, fmt.Errorf("keine ID in der Konfiguration angegeben")
	}
	if config.Type == "" {
		return nil, fmt.Errorf("kein Typ in der Konfiguration angegeben")
	}

	return &config, nil
}
