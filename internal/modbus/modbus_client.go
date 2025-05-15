package modbus

import (
	"fmt"
	"time"

	"github.com/simonvetter/modbus"
)

// Client wraps the modbus client and provides methods for communication.
type Client struct {
	mbClient *modbus.ModbusClient
	Config   ClientConfig
}

// ClientConfig holds the configuration for the Modbus client.
type ClientConfig struct {
	URL      string        // e.g., "rtu:///dev/ttyUSB0"
	BaudRate uint          // e.g., 19200. Matched to library's `Speed uint`
	DataBits uint          // e.g., 8. Matched to library's `DataBits uint`
	Parity   string        // e.g., "N", "E", "O"
	StopBits uint          // e.g., 1, 2. Matched to library's `StopBits uint`
	Timeout  time.Duration // e.g., 300 * time.Millisecond
}

// NewClient creates and configures a new Modbus client.
// It also attempts to open the connection.
func NewClient(config ClientConfig) (*Client, error) {
	var parityValue uint
	switch config.Parity {
	case "E":
		parityValue = modbus.PARITY_EVEN
	case "O":
		parityValue = modbus.PARITY_ODD
	default:
		parityValue = modbus.PARITY_NONE
	}

	mbClient, err := modbus.NewClient(&modbus.ClientConfiguration{
		URL:      config.URL,
		Speed:    config.BaudRate,
		DataBits: config.DataBits,
		Parity:   parityValue,
		StopBits: config.StopBits,
		Timeout:  config.Timeout,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create modbus client core: %w", err)
	}

	c := &Client{mbClient: mbClient, Config: config}
	err = c.Open() // Attempt to open the connection using the new Open method.
	if err != nil {
		// Return the client instance along with the error from Open().
		// This allows the caller to have the client instance for potential retries if appropriate.
		return c, fmt.Errorf("failed to open modbus connection: %w", err)
	}

	return c, nil
}

// Open attempts to open the underlying modbus connection.
func (c *Client) Open() error {
	if c.mbClient == nil {
		return fmt.Errorf("underlying modbus client (mbClient) is nil, cannot open")
	}
	fmt.Printf("DEBUG: Attempting to open Modbus connection to %s (BaudRate: %d, DataBits: %d, Parity: %s, StopBits: %d, Timeout: %v)\n",
		c.Config.URL, c.Config.BaudRate, c.Config.DataBits, c.Config.Parity, c.Config.StopBits, c.Config.Timeout)
	err := c.mbClient.Open()
	if err != nil {
		fmt.Printf("DEBUG: Failed to open Modbus connection: %v\n", err)
	} else {
		fmt.Printf("DEBUG: Successfully opened Modbus connection\n")
	}
	return err
}

// ReadHoldingRegisters reads a number of holding registers starting from a given address.
func (c *Client) ReadHoldingRegisters(slaveID uint8, address uint16, quantity uint16) ([]uint16, error) {
	if c.mbClient == nil {
		return nil, fmt.Errorf("modbus client not initialized")
	}
	fmt.Printf("DEBUG: Reading %d registers from address 0x%04X from slave ID %d\n", quantity, address, slaveID)
	c.mbClient.SetUnitId(slaveID) // Set slave ID for this transaction
	registers, err := c.mbClient.ReadRegisters(address, quantity, modbus.HOLDING_REGISTER)
	if err != nil {
		fmt.Printf("DEBUG: Error reading registers: %v\n", err)
		return nil, err
	}
	fmt.Printf("DEBUG: Successfully read registers: %v\n", registers)
	return registers, nil
}

// Close closes the Modbus connection.
func (c *Client) Close() error {
	if c.mbClient != nil {
		return c.mbClient.Close()
	}
	return nil
}

// IsConnected checks if the client is connected.
// NOTE: The simonvetter/modbus library does not provide a direct IsConnected method.
// Connection status is typically inferred from the success or failure of operations or Open().
// This implementation is a placeholder and may not accurately reflect connection state without an operation.
func (c *Client) IsConnected() bool {
    if c.mbClient == nil {
        return false
    }
    // This is a basic check. A robust way would be to see if the underlying client's `conn` is nil,
    // but that's an internal field. For now, if Open() succeeded without error (or last op was ok),
    // we might assume it's connected. This is not reliable.
    // The library expects users to handle errors on operations to detect disconnection.
    // A pragmatic approach for the SensorManager might be to attempt operations and handle errors,
    // potentially retrying Open() if an operation suggests disconnection.
    // For now, if NewClient (which calls Open) didn't return an error for Open, assume connected initially.
    return true // This is a simplified placeholder.
}
