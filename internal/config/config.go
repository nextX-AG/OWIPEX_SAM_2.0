package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	//	"time" // Removed unused import

	"github.com/joho/godotenv"
)

// SensorConfig defines the structure for individual sensor configurations
type SensorConfig struct {
	ID                  string                 `json:"id"`
	Name                string                 `json:"name"`
	Type                string                 `json:"type"`
	DeviceID            int                    `json:"device_id"` // Modbus slave ID
	Location            string                 `json:"location"`
	Enabled             bool                   `json:"enabled"` // Determines if the sensor is active
	Metadata            map[string]interface{} `json:"metadata"`
	ReadIntervalSeconds int                    `json:"read_interval_seconds"`
	Transmission        struct {
		Formats  []string `json:"formats"`
		Interval int      `json:"interval"` // This seems redundant with ReadIntervalSeconds
	} `json:"transmission"`
}

// RS485Config defines the Modbus RTU (RS485) connection parameters
type RS485Config struct {
	Port      string `json:"port"`
	Baudrate  int    `json:"baudrate"`
	DataBits  int    `json:"databits"`
	Parity    string `json:"parity"`
	StopBits  int    `json:"stopbits"`
	TimeoutMS int    `json:"timeout_ms"`
}

// ThingsBoardConfig defines the ThingsBoard MQTT connection parameters
type ThingsBoardConfig struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	AccessToken string `json:"access_token"`
}

// AppConfig is the top-level configuration structure
type AppConfig struct {
	RS485       RS485Config       `json:"rs485_settings"`
	ThingsBoard ThingsBoardConfig `json:"thingsboard_settings"`
	Sensors     []SensorConfig    `json:"sensors"`
	LogFilePath string            `json:"log_file_path"`
}

// LoadAppConfig loads configuration from a JSON file and overrides with .env values
func LoadAppConfig(configFilePath string) (*AppConfig, error) {
	logger := log.New(os.Stdout, "[ConfigLoader] ", log.LstdFlags)

	// Default AppConfig
	appConfig := &AppConfig{
		RS485: RS485Config{
			Port:      "/dev/ttyS0", // Default, will be overridden by env if present
			Baudrate:  9600,
			DataBits:  8,
			Parity:    "N",
			StopBits:  1,
			TimeoutMS: 1000,
		},
		ThingsBoard: ThingsBoardConfig{
			Host: "localhost", // Default, will be overridden by env if present
			Port: 1883,
		},
		LogFilePath: "/var/log/owipex/go_reader.log",
	}

	// Load from JSON config file if provided and exists
	if configFilePath != "" {
		data, err := ioutil.ReadFile(configFilePath)
		if err != nil {
			logger.Printf("Warning: Could not read JSON config file %s: %v. Using defaults and .env values.", configFilePath, err)
		} else {
			err = json.Unmarshal(data, appConfig)
			if err != nil {
				return nil, fmt.Errorf("error unmarshalling JSON config file %s: %w", configFilePath, err)
			}
			logger.Printf("Loaded configuration from JSON file: %s", configFilePath)
		}
	}

	// Override with .env file values
	envPath := "/etc/owipex/go_reader.env" // Default .env path
	if os.Getenv("GO_READER_ENV_PATH") != "" {
		envPath = os.Getenv("GO_READER_ENV_PATH")
	}

	err := godotenv.Load(envPath)
	if err != nil {
		logger.Printf("Warning: Could not load .env file from %s: %v. Using JSON or default values.", envPath, err)
	} else {
		logger.Printf("Successfully loaded .env file from %s", envPath)
	}

	if val := os.Getenv("RS485_PORT"); val != "" {
		appConfig.RS485.Port = val
		logger.Printf("ENV Override: RS485_PORT=%s", val)
	}

	if val := os.Getenv("RS485_BAUDRATE"); val != "" {
		baud, err := strconv.Atoi(val)
		if err != nil {
			logger.Printf("Warning: Could not parse RS485_BAUDRATE from env ('%s'): %v. Using value %d", val, err, appConfig.RS485.Baudrate)
		} else {
			appConfig.RS485.Baudrate = baud
			logger.Printf("ENV Override: RS485_BAUDRATE=%d", baud)
		}
	}
	if val := os.Getenv("RS485_DATABITS"); val != "" {
		dbits, err := strconv.Atoi(val)
		if err != nil {
			logger.Printf("Warning: Could not parse RS485_DATABITS from env ('%s'): %v. Using value %d", val, err, appConfig.RS485.DataBits)
		} else {
			appConfig.RS485.DataBits = dbits
			logger.Printf("ENV Override: RS485_DATABITS=%d", dbits)
		}
	}
	if val := os.Getenv("RS485_PARITY"); val != "" {
		appConfig.RS485.Parity = val
		logger.Printf("ENV Override: RS485_PARITY=%s", val)
	}
	if val := os.Getenv("RS485_STOPBITS"); val != "" {
		sbits, err := strconv.Atoi(val)
		if err != nil {
			logger.Printf("Warning: Could not parse RS485_STOPBITS from env ('%s'): %v. Using value %d", val, err, appConfig.RS485.StopBits)
		} else {
			appConfig.RS485.StopBits = sbits
			logger.Printf("ENV Override: RS485_STOPBITS=%d", sbits)
		}
	}
	if val := os.Getenv("RS485_TIMEOUT_MS"); val != "" {
		tout, err := strconv.Atoi(val)
		if err != nil {
			logger.Printf("Warning: Could not parse RS485_TIMEOUT_MS from env ('%s'): %v. Using value %d", val, err, appConfig.RS485.TimeoutMS)
		} else {
			appConfig.RS485.TimeoutMS = tout
			logger.Printf("ENV Override: RS485_TIMEOUT_MS=%d", tout)
		}
	}

	if val := os.Getenv("RS485_THINGSBOARD_SERVER"); val != "" {
		appConfig.ThingsBoard.Host = val
		logger.Printf("ENV Override: RS485_THINGSBOARD_SERVER=%s", val)
	}
	if val := os.Getenv("RS485_THINGSBOARD_PORT"); val != "" {
		port, err := strconv.Atoi(val)
		if err != nil {
			logger.Printf("Warning: Could not parse RS485_THINGSBOARD_PORT from env ('%s'): %v. Using value %d", val, err, appConfig.ThingsBoard.Port)
		} else {
			appConfig.ThingsBoard.Port = port
			logger.Printf("ENV Override: RS485_THINGSBOARD_PORT=%d", port)
		}
	}
	if val := os.Getenv("RS485_ACCESS_TOKEN"); val != "" {
		appConfig.ThingsBoard.AccessToken = val
		logger.Printf("ENV Override: RS485_ACCESS_TOKEN=%s", val)
	}

	// Load sensor configurations from the specified JSON file (appConfig.Sensors might have been populated by a main config JSON already)
	// If appConfig.Sensors is empty and a specific sensors.json path is given (or default), load it.
	// For this project, sensors.json is typically in ./config/sensors.json relative to executable or a path from AppConfig itself.
	// We assume sensors are part of the main config JSON or a fixed path for now.
	// If sensors are in a separate file, that logic needs to be here.
	// For now, we assume sensors are loaded from the main config file path if provided.

	logger.Printf("Final loaded AppConfig: %+v", appConfig)
	logger.Printf("Final RS485 Config: %+v", appConfig.RS485)
	logger.Printf("Final ThingsBoard Config: %+v", appConfig.ThingsBoard)

	return appConfig, nil
}

// LoadSensorDefinitions loads sensor definitions from a separate JSON file.
func LoadSensorDefinitions(filePath string) ([]SensorConfig, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not read sensor definitions file %s: %w", filePath, err)
	}

	var sensors []SensorConfig
	err = json.Unmarshal(data, &sensors)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling sensor definitions file %s: %w", filePath, err)
	}
	return sensors, nil
}
