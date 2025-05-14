#!/bin/bash

# Funktion zum Ausführen von Befehlen mit Sperre-Erkennung
run_with_lock_detection() {
    local cmd="$1"
    local max_tries=5
    local try=1
    local success=false
    
    while [ $try -le $max_tries ]; do
        echo "Führe aus: $cmd"
        if eval "$cmd"; then
            success=true
            break
        else
            ret=$?
            if echo "$cmd" | grep -qE "apt-get|dpkg"; then
                # Prüfe auf dpkg-Sperre
                if [ $ret -eq 100 ] || grep -q "Sperre" <<< "$(LANG=C $cmd 2>&1)" || grep -q "lock" <<< "$(LANG=C $cmd 2>&1)"; then
                    pid=$(ps aux | grep -E "apt-get|dpkg|unattended" | grep -v grep | awk 'NR==1{print $2}')
                    if [ -n "$pid" ]; then
                        process=$(ps -p $pid -o comm=)
                        echo "Paket-Manager ist blockiert durch Prozess $pid ($process). Versuch $try/$max_tries"
                    else
                        echo "Paket-Manager ist blockiert. Versuch $try/$max_tries"
                    fi
                    echo "Warte 30 Sekunden und versuche es erneut..."
                    sleep 30
                    try=$((try + 1))
                else
                    echo "Befehl fehlgeschlagen mit Exit-Code $ret"
                    return $ret
                fi
            else
                echo "Befehl fehlgeschlagen mit Exit-Code $ret"
                return $ret
            fi
        fi
    done
    
    if [ "$success" = true ]; then
        return 0
    else
        echo "Maximale Anzahl an Versuchen erreicht ($max_tries)"
        return 1
    fi
}

# Funktion zum Prüfen, ob Go installiert ist
check_go_installed() {
    if command -v go &> /dev/null; then
        echo "Go ist bereits installiert: $(go version)"
        return 0
    else
        return 1
    fi
}

# Funktion zur manuellen Installation von Go
install_go_manually() {
    echo "Installiere Go manuell..."
    
    # Go-Version
    GO_VERSION="1.19.13"
    GO_FILE="go${GO_VERSION}.linux-amd64.tar.gz"
    GO_URL="https://golang.org/dl/${GO_FILE}"
    
    # Arbeitsverzeichnis für Downloads
    TEMP_DIR="/tmp"
    DOWNLOAD_PATH="${TEMP_DIR}/${GO_FILE}"
    
    # Herunterladen der Go-Binärdatei
    echo "Lade Go ${GO_VERSION} herunter..."
    if ! wget -O "$DOWNLOAD_PATH" "$GO_URL"; then
        echo "Fehler: Konnte Go nicht herunterladen."
        return 1
    fi
    
    if [ ! -f "$DOWNLOAD_PATH" ]; then
        echo "Fehler: Konnte Go nicht herunterladen. Datei $DOWNLOAD_PATH existiert nicht."
        return 1
    fi
    
    # Entpacken nach /usr/local
    echo "Entpacke Go nach /usr/local..."
    
    # Prüfe, ob wir Root-Rechte haben, wenn nicht, verwende sudo
    if [ $(id -u) -ne 0 ]; then
        SUDO="sudo"
    else
        SUDO=""
    fi
    
    # Entferne altes Go-Verzeichnis und entpacke das neue
    $SUDO rm -rf /usr/local/go && $SUDO tar -C /usr/local -xzf "$DOWNLOAD_PATH"
    if [ $? -ne 0 ]; then
        echo "Fehler: Konnte Go nicht entpacken."
        return 1
    fi
    
    # PATH setzen
    echo "Setze PATH-Variable..."
    GO_BIN_PATH="/usr/local/go/bin"
    
    # PATH für aktuelle Sitzung setzen
    export PATH="$GO_BIN_PATH:$PATH"
    
    # PATH permanent setzen
    PROFILE_PATH="$HOME/.profile"
    
    # Prüfe, ob der Pfad bereits in .profile ist
    if [ -f "$PROFILE_PATH" ] && grep -q "/usr/local/go/bin" "$PROFILE_PATH"; then
        echo "PATH bereits in .profile konfiguriert."
    else
        echo 'export PATH=$PATH:/usr/local/go/bin' >> "$PROFILE_PATH"
        echo "PATH zu .profile hinzugefügt."
    fi
    
    # Lade .profile neu
    source "$PROFILE_PATH" || true
    
    # Prüfe Go-Installation
    if check_go_installed; then
        return 0
    else
        echo "Fehler: Go wurde installiert, ist aber nicht im PATH verfügbar."
        echo "Bitte führen Sie manuell aus:"
        echo "  export PATH=$PATH:/usr/local/go/bin"
        echo "  source ~/.profile"
        return 1
    fi
}

# Erkennung des Betriebssystems
OS=$(uname -s)
echo "Erkanntes Betriebssystem: $OS"

# Installiere Abhängigkeiten auf Ubuntu/Linux
install_dependencies() {
    echo "Installiere Abhängigkeiten für Ubuntu/Linux..."

    # Prüfe, ob Go bereits installiert ist
    GO_INSTALLED=false
    if check_go_installed; then
        GO_INSTALLED=true
    else
        echo "Go ist nicht installiert."
        
        # Verwende die manuelle Installation direkt
        if install_go_manually; then
            GO_INSTALLED=true
        else
            echo "Go konnte nicht installiert werden. Das Programm benötigt Go zum Ausführen."
            return 1
        fi
    fi

    # Prüfe, ob wir Root-Rechte haben, wenn nicht, verwende sudo
    if [ $(id -u) -ne 0 ]; then
        SUDO="sudo"
    else
        SUDO=""
    fi

    # Aktualisiere Paketlisten
    if ! run_with_lock_detection "$SUDO apt-get update"; then
        echo "Warnung: Konnte Paketlisten nicht aktualisieren, fahre fort..."
    fi

    # Installiere weitere benötigte Pakete
    if ! run_with_lock_detection "$SUDO apt-get install -y build-essential git"; then
        echo "Warnung: Konnte build-essential und git nicht installieren"
    fi

    # Füge den Benutzer zur Gruppe dialout hinzu (für seriellen Port-Zugriff)
    if $SUDO usermod -a -G dialout $USER; then
        echo "WICHTIG: Sie müssen sich möglicherweise neu anmelden, damit die Gruppenmitgliedschaft 'dialout' wirksam wird."
    fi

    # Installiere Go-Abhängigkeiten
    if [ "$GO_INSTALLED" = true ]; then
        echo "Installiere Go-Abhängigkeiten..."
        go mod tidy || echo "Konnte Go-Module nicht installieren"
    fi

    return 0
}

# Funktion zum Suchen von seriellen Ports
find_serial_ports() {
    echo "Suche nach verfügbaren seriellen Ports..."
    local found_ports=$(ls -l /dev/ttyS* /dev/ttyUSB* /dev/ttyACM* 2>/dev/null)
    if [ -z "$found_ports" ]; then
        echo "Keine seriellen Ports gefunden."
        return 1
    fi
    
    echo "Gefundene serielle Ports:"
    echo "$found_ports"
    
    # Extrahiere den ersten Port (bevorzuge USB-Ports)
    local usb_port=$(echo "$found_ports" | grep "USB" | head -1 | awk '{print $NF}')
    if [ -n "$usb_port" ]; then
        echo "Verwende USB-Port: $usb_port"
        echo "rtu://$usb_port"
        return 0
    fi
    
    # Wenn kein USB-Port gefunden wurde, nimm den ersten verfügbaren Port
    local first_port=$(echo "$found_ports" | head -1 | awk '{print $NF}')
    if [ -n "$first_port" ]; then
        echo "Verwende seriellen Port: $first_port"
        echo "rtu://$first_port"
        return 0
    fi
    
    return 1
}

# Betriebssystemspezifische Konfiguration
DEPENDENCIES_INSTALLED=true
if [ "$OS" == "Linux" ]; then
    # Führe Installationsroutine aus
    install_dependencies || DEPENDENCIES_INSTALLED=false

    # Finde verfügbare serielle Ports
    PORT=$(find_serial_ports)
    if [ $? -eq 0 ]; then
        SERIAL_PORT=$PORT
    else
        # Setze seriellen Port für Linux-Systeme
        SERIAL_PORT="rtu:///dev/ttyS0"
        
        # Prüfe, ob der Port existiert
        if [ ! -e "${SERIAL_PORT#rtu://}" ]; then
            echo "WARNUNG: Der konfigurierte serielle Port ${SERIAL_PORT#rtu://} existiert nicht!"
            echo "Bitte schließen Sie ein RS485-Gerät an oder konfigurieren Sie den korrekten Port in diesem Skript."
        fi
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

if [ "$DEPENDENCIES_INSTALLED" = false ]; then
    echo "Abhängigkeiten nicht vollständig installiert. Das Programm kann möglicherweise nicht korrekt starten."
    read -p "Möchten Sie dennoch fortfahren? (j/n): " ANSWER
    if [[ ! "$ANSWER" =~ ^[jJyY] ]]; then
        echo "Abbruch."
        exit 1
    fi
fi

echo "Starte Modbus-Reader..."

# Starte den Go-Modbus-Reader
GO_READER_ENV_PATH="$(pwd)/${CONFIG_FILE}" go run cmd/main.go 