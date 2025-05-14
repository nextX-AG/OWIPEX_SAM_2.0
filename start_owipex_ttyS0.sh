#!/bin/bash

# Dieses Skript startet den Modbus-Reader speziell mit dem seriellen Port ttyS0

# Verzeichnis für Konfigurationsdateien
CONFIG_DIR="config/env"
CONFIG_FILE="${CONFIG_DIR}/go_reader.env"

# Stellt sicher, dass das Verzeichnis existiert
mkdir -p "$CONFIG_DIR"

# Konfigurationsdatei mit ttyS0 als Port
cat > "$CONFIG_FILE" << EOF
RS485_PORT=rtu:///dev/ttyS0
RS485_BAUDRATE=9600
RS485_DATABITS=8
RS485_PARITY=N
RS485_STOPBITS=1
RS485_TIMEOUT_MS=3000
RS485_THINGSBOARD_SERVER=146.4.67.141
RS485_THINGSBOARD_PORT=1883
RS485_ACCESS_TOKEN=5Ohlb6ZKO4uNw9O2DHwk
EOF

echo "Konfigurationsdatei mit ttyS0 erstellt unter: $CONFIG_FILE"
echo "Starte Modbus-Reader mit /dev/ttyS0..."

# Go-Pfad hinzufügen, falls nötig
export PATH=$PATH:/usr/local/go/bin

# Start des Go-Programms
GO_READER_ENV_PATH="$(pwd)/${CONFIG_FILE}" go run cmd/main.go 