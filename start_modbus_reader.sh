#!/bin/bash

# Erkennung des Betriebssystems
OS=$(uname -s)
echo "Erkanntes Betriebssystem: $OS"

# Installiere Abhängigkeiten auf Ubuntu/Linux
install_dependencies() {
    echo "Installiere Abhängigkeiten für Ubuntu/Linux..."

    # Prüfe, ob wir Root-Rechte haben, wenn nicht, verwende sudo
    if [ $(id -u) -ne 0 ]; then
        SUDO="sudo"
    else
        SUDO=""
    fi

    # Aktualisiere Paketlisten
    $SUDO apt-get update

    # Installiere Go, falls nicht vorhanden
    if ! command -v go &> /dev/null; then
        echo "Go ist nicht installiert. Installiere Go..."
        $SUDO apt-get install -y golang
    fi

    # Installiere weitere benötigte Pakete
    $SUDO apt-get install -y build-essential git

    # Füge den Benutzer zur Gruppe dialout hinzu (für seriellen Port-Zugriff)
    $SUDO usermod -a -G dialout $USER
    echo "WICHTIG: Sie müssen sich möglicherweise neu anmelden, damit die Gruppenmitgliedschaft 'dialout' wirksam wird."

    # Prüfe Go-Version
    go version

    # Installiere Go-Abhängigkeiten
    echo "Installiere Go-Abhängigkeiten..."
    go mod tidy
}

# Konfiguration für den Modbus-Reader
# -----------------------------------

# Betriebssystemspezifische Konfiguration
if [ "$OS" == "Linux" ]; then
    # Führe Installationsroutine aus
    install_dependencies

    # Setze seriellen Port für Linux-Systeme
    SERIAL_PORT="rtu:///dev/ttyS0"
    
    # Prüfe, ob der Port existiert
    if [ ! -e "${SERIAL_PORT#rtu://}" ]; then
        echo "WARNUNG: Der konfigurierte serielle Port ${SERIAL_PORT#rtu://} existiert nicht!"
        echo "Verfügbare serielle Ports:"
        ls -l /dev/ttyS* /dev/ttyUSB* /dev/ttyACM* 2>/dev/null || echo "Keine seriellen Ports gefunden"
        echo
        echo "Bitte konfigurieren Sie den korrekten Port in diesem Skript."
    fi
else
    # Serieller Port für macOS
    SERIAL_PORT="rtu:///dev/tty.usbserial-B001QSI5"
fi

# Modbus-Einstellungen
BAUDRATE=9600
DATABITS=8
PARITY="N"
STOPBITS=1
TIMEOUT_MS=3000

# ThingsBoard-Einstellungen
THINGSBOARD_SERVER="146.4.67.141"
THINGSBOARD_PORT=1883
ACCESS_TOKEN="5Ohlb6ZKO4uNw9O2DHwk"

# Konfigurationsdatei erstellen
CONFIG_DIR="config/env"
CONFIG_FILE="${CONFIG_DIR}/go_reader.env"

# Stellt sicher, dass das Verzeichnis existiert
mkdir -p "$CONFIG_DIR"

# Konfigurationsdatei schreiben
cat > "$CONFIG_FILE" << EOF
RS485_PORT=${SERIAL_PORT}
RS485_BAUDRATE=${BAUDRATE}
RS485_DATABITS=${DATABITS}
RS485_PARITY=${PARITY}
RS485_STOPBITS=${STOPBITS}
RS485_TIMEOUT_MS=${TIMEOUT_MS}
RS485_THINGSBOARD_SERVER=${THINGSBOARD_SERVER}
RS485_THINGSBOARD_PORT=${THINGSBOARD_PORT}
RS485_ACCESS_TOKEN=${ACCESS_TOKEN}
EOF

echo "Konfigurationsdatei erstellt unter: $CONFIG_FILE"
echo "Starte Modbus-Reader..."

# Starte den Go-Modbus-Reader
GO_READER_ENV_PATH="$(pwd)/${CONFIG_FILE}" go run cmd/main.go 