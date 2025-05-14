# owipexRS485GO - Architecture Documentation

## Overview
This document describes the architecture of the owipexRS485GO application, which serves as a communication bridge between various sensors using the Modbus RS485 protocol and the ThingsBoard IoT platform.

## Components

### Command Layer (`cmd/`)
- **main.go** - Application entry point and high-level orchestration

### Configuration (`config/`, `internal/config/`)
- **sensors.json** - Sensor definitions and configuration
- **config.go** - Configuration loading and management
- **config_test.go** - Tests for configuration functions

### Modbus Communication (`internal/modbus/`)
- **modbus_client.go** - Handles RS485 Modbus communication
- **modbus_client_test.go** - Tests for Modbus client

### Sensor Management (`internal/sensor/`, `internal/manager/`)
- **sensor.go** - Base sensor interface
- **sensor_manager.go** - Manages all sensor operations
- Specific sensor implementations:
  - **flow_sensor.go**
  - **ph_sensor.go**
  - **radar_sensor.go**
  - **turbidity_sensor.go**

### ThingsBoard Integration (`internal/thingsboard/`)
- **thingsboard_client.go** - Handles communication with ThingsBoard platform

## Data Flow
1. Main application initializes components
2. Sensor Manager reads configuration
3. Modbus Client connects to sensors
4. Sensor data is collected at configured intervals
5. Data is processed and sent to ThingsBoard
6. System logs and handles errors throughout the process

## Error Handling
Error handling occurs at each layer with appropriate logging and recovery mechanisms.

## Future Improvements
- Add monitoring dashboard for sensor statuses
- Implement caching for sensor readings
- Add support for more sensor types 