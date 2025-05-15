#!/bin/bash
# ThingsBoard Test Starter Shell Script
# Dieses Skript startet das Python-Skript zum Testen der ThingsBoard-Verbindung

# Verzeichnis des Skripts ermitteln
SCRIPT_DIR=$(dirname "$(readlink -f "$0")")
cd "$SCRIPT_DIR" || exit 1

# Terminal-Farben definieren
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # Keine Farbe

echo -e "${GREEN}[ThingsBoard Test]${NC} Starte ThingsBoard-Variablentest..."

# Prüfen, ob Python installiert ist
if ! command -v python3 &> /dev/null; then
    echo -e "${RED}[FEHLER]${NC} Python 3 ist nicht installiert. Bitte installieren Sie Python 3."
    exit 1
fi

# Umgebungsvariablen konfigurieren (falls nötig)
# Zum Überschreiben hier Werte angeben:
# export TB_HOST="192.168.1.100"
# export TB_PORT="1883"
# export TB_TOKEN="meinToken123"

# Führe das Python-Skript aus
if [ -f "$SCRIPT_DIR/start_thingsboard_test.py" ]; then
    python3 "$SCRIPT_DIR/start_thingsboard_test.py"
    EXIT_CODE=$?
    
    if [ $EXIT_CODE -ne 0 ]; then
        echo -e "${RED}[FEHLER]${NC} Das ThingsBoard-Test-Programm wurde mit Fehlercode $EXIT_CODE beendet."
        exit $EXIT_CODE
    fi
else
    echo -e "${RED}[FEHLER]${NC} Die Datei start_thingsboard_test.py wurde nicht gefunden."
    exit 1
fi 