// Package device implementiert das Laden von Gerätekonfigurationen.
package device

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"owipex_reader/internal/types"
)

// LoadDeviceConfig lädt eine Gerätekonfiguration aus einer JSON-Datei
func LoadDeviceConfig(filePath string) (*types.DeviceConfig, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("fehler beim Lesen der Konfigurationsdatei: %w", err)
	}

	var config types.DeviceConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("fehler beim Dekodieren der Konfiguration: %w", err)
	}

	return &config, nil
}

// LoadDeviceConfigs lädt alle Gerätekonfigurationen aus einem Verzeichnis
func LoadDeviceConfigs(dirPath string) ([]types.DeviceConfig, error) {
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("fehler beim Lesen des Verzeichnisses: %w", err)
	}

	var configs []types.DeviceConfig
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		fullPath := filepath.Join(dirPath, file.Name())
		config, err := LoadDeviceConfig(fullPath)
		if err != nil {
			return nil, fmt.Errorf("fehler beim Laden der Konfiguration %s: %w", file.Name(), err)
		}

		configs = append(configs, *config)
	}

	return configs, nil
}

// SaveDeviceConfig speichert eine Gerätekonfiguration in einer JSON-Datei
func SaveDeviceConfig(config *types.DeviceConfig, filePath string) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("fehler beim Kodieren der Konfiguration: %w", err)
	}

	err = ioutil.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("fehler beim Schreiben der Konfigurationsdatei: %w", err)
	}

	return nil
}
