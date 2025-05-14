package config

import (
	"os"
	"path/filepath"
	"testing"
	// "time" // Removed unused time import
)

// TestLoadAppConfig tests the LoadAppConfig function.
func TestLoadAppConfig(t *testing.T) {
	// Create a temporary directory for test config files
	tempDir, err := os.MkdirTemp("", "config_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a dummy sensors.json file
	dummySensorsJSON := `{
		"rs485_settings": {
			"port": "/dev/ttyUSB0",
			"baudrate": 9600,
			"parity": "N",
			"stopbits": 1,
			"bytesize": 8,
			"timeout_ms": 1000
		},
		"sensors": [
			{
				"id": "ph_sensor_1",
				"name": "pH Probe Lab",
				"type": "ph",
				"device_id": 1,
				"location": "Lab Bench 1",
				"read_interval_seconds": 15,
				"transmission": {
					"interval": 15,
					"formats": ["simple", "json"]
				},
				"metadata": {"version": "1.0"}
			}
		]
	}`
	configFilePath := filepath.Join(tempDir, "sensors.json")
	if err := os.WriteFile(configFilePath, []byte(dummySensorsJSON), 0644); err != nil {
		t.Fatalf("Failed to write dummy sensors.json: %v", err)
	}

	// Create a dummy .env file (content is not directly used by LoadAppConfig in this test setup due to hardcoded path)
	// dummyEnvContent := `
// RS485_THINGSBOARD_SERVER=localhost
// RS485_THINGSBOARD_PORT=1883
// RS485_ACCESS_TOKEN=test_token
// RS485_READ_INTERVAL=30
// `
	// originalEnvPath := "/etc/owipex/.envRS485" // Path used in config.go, not directly used in test logic here

	// Backup original env vars if they exist and defer restoration
	originalTBServer := os.Getenv("RS485_THINGSBOARD_SERVER")
	originalTBPort := os.Getenv("RS485_THINGSBOARD_PORT")
	originalTBAccessToken := os.Getenv("RS485_ACCESS_TOKEN")
	originalReadInterval := os.Getenv("RS485_READ_INTERVAL")

	defer func() {
		os.Setenv("RS485_THINGSBOARD_SERVER", originalTBServer)
		os.Setenv("RS485_THINGSBOARD_PORT", originalTBPort)
		os.Setenv("RS485_ACCESS_TOKEN", originalTBAccessToken)
		os.Setenv("RS485_READ_INTERVAL", originalReadInterval)
	}()

	os.Setenv("RS485_THINGSBOARD_SERVER", "test.thingsboard.local")
	os.Setenv("RS485_THINGSBOARD_PORT", "11883")
	os.Setenv("RS485_ACCESS_TOKEN", "env_test_token")
	os.Setenv("RS485_READ_INTERVAL", "45")

	cfg, err := LoadAppConfig(configFilePath)
	if err != nil {
		t.Fatalf("LoadAppConfig failed: %v", err)
	}

	if cfg.RS485.Port != "/dev/ttyUSB0" {
		t.Errorf("Expected RS485 port /dev/ttyUSB0, got %s", cfg.RS485.Port)
	}
	if cfg.RS485.Baudrate != 9600 {
		t.Errorf("Expected RS485 baudrate 9600, got %d", cfg.RS485.Baudrate)
	}
	if len(cfg.Sensors) != 1 {
		t.Fatalf("Expected 1 sensor, got %d", len(cfg.Sensors))
	}
	if cfg.Sensors[0].ID != "ph_sensor_1" {
		t.Errorf("Expected sensor ID ph_sensor_1, got %s", cfg.Sensors[0].ID)
	}
	if cfg.Sensors[0].Type != "ph" {
		t.Errorf("Expected sensor type ph, got %s", cfg.Sensors[0].Type)
	}

	// Check values loaded from (mocked) environment variables
	if cfg.ThingsBoard.Server != "test.thingsboard.local" {
		t.Errorf("Expected ThingsBoard server test.thingsboard.local, got %s", cfg.ThingsBoard.Server)
	}
	if cfg.ThingsBoard.Port != 11883 {
		t.Errorf("Expected ThingsBoard port 11883, got %d", cfg.ThingsBoard.Port)
	}
	if cfg.ThingsBoard.AccessToken != "env_test_token" {
		t.Errorf("Expected ThingsBoard access token env_test_token, got %s", cfg.ThingsBoard.AccessToken)
	}
	if cfg.ReadInterval != 45 {
		t.Errorf("Expected general read interval 45, got %d", cfg.ReadInterval)
	}

	t.Log("TestLoadAppConfig passed.")
}

