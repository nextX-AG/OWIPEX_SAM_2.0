#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import os
import subprocess
import sys
import platform
import shutil

# Funktionen für die Installation von Abhängigkeiten
def run_command(command, check=True):
    """Führt einen Shell-Befehl aus und gibt die Ausgabe zurück"""
    print(f"Führe aus: {' '.join(command)}")
    try:
        result = subprocess.run(command, check=check, text=True, capture_output=True)
        if result.stdout:
            print(result.stdout)
        return result
    except subprocess.CalledProcessError as e:
        print(f"Fehler beim Ausführen des Befehls: {e}")
        print(e.stdout)
        print(e.stderr, file=sys.stderr)
        if check:
            sys.exit(1)
        return e

def install_linux_dependencies():
    """Installiert benötigte Abhängigkeiten auf Ubuntu/Linux"""
    print("Installiere Abhängigkeiten für Ubuntu/Linux...")

    # Prüfen, ob wir Root-Rechte haben
    is_root = os.geteuid() == 0
    sudo_cmd = [] if is_root else ["sudo"]

    # Aktualisiere Paketlisten
    run_command(sudo_cmd + ["apt-get", "update"])

    # Installiere Go, falls nicht vorhanden
    if not shutil.which("go"):
        print("Go ist nicht installiert. Installiere Go...")
        run_command(sudo_cmd + ["apt-get", "install", "-y", "golang"])

    # Installiere weitere benötigte Pakete
    run_command(sudo_cmd + ["apt-get", "install", "-y", "build-essential", "git"])

    # Füge den Benutzer zur Gruppe dialout hinzu (für seriellen Port-Zugriff)
    run_command(sudo_cmd + ["usermod", "-a", "-G", "dialout", os.environ.get("USER", os.environ.get("USERNAME", ""))])
    print("WICHTIG: Sie müssen sich möglicherweise neu anmelden, damit die Gruppenmitgliedschaft 'dialout' wirksam wird.")

    # Prüfe Go-Version
    run_command(["go", "version"])

    # Installiere Go-Abhängigkeiten
    print("Installiere Go-Abhängigkeiten...")
    run_command(["go", "mod", "tidy"])

# Konfiguration für den Modbus-Reader
# -----------------------------------

# Aktuelles Betriebssystem erkennen
current_os = platform.system()
print(f"Erkanntes Betriebssystem: {current_os}")

# OS-spezifische Ports
serial_ports = {
    'Darwin': 'rtu:///dev/tty.usbserial-B001QSI5',  # macOS
    'Linux': 'rtu:///dev/ttyS0',                    # Linux
    'Windows': 'rtu:///COM1'                        # Windows
}

# Installation von Abhängigkeiten auf Linux
if current_os == "Linux":
    try:
        install_linux_dependencies()
    except Exception as e:
        print(f"Fehler bei der Installation der Abhängigkeiten: {e}")
        print("Installieren Sie die benötigten Pakete manuell: golang, build-essential, git")
        print("Und fügen Sie Ihren Benutzer zur Gruppe 'dialout' hinzu: sudo usermod -a -G dialout $USER")

    # Prüfe, ob der serielle Port existiert
    port_path = serial_ports['Linux'].replace('rtu://', '')
    if not os.path.exists(port_path):
        print(f"WARNUNG: Der konfigurierte serielle Port {port_path} existiert nicht!")
        print("Verfügbare serielle Ports:")
        for pattern in ['/dev/ttyS*', '/dev/ttyUSB*', '/dev/ttyACM*']:
            try:
                ports = subprocess.check_output(f"ls -l {pattern} 2>/dev/null || echo ''", shell=True).decode().strip()
                if ports:
                    print(ports)
            except:
                pass
        print("\nBitte konfigurieren Sie den korrekten Port in diesem Skript.")

# Standardmäßig den Port für das aktuelle Betriebssystem verwenden
serial_port = serial_ports.get(current_os, serial_ports['Darwin'])

# Modbus-Einstellungen
baudrate = 9600
databits = 8
parity = "N"
stopbits = 1
timeout_ms = 3000

# ThingsBoard-Einstellungen
thingsboard_server = "146.4.67.141"
thingsboard_port = 1883
access_token = "5Ohlb6ZKO4uNw9O2DHwk"

# Arbeitsverzeichnis und Konfigurationsdatei
work_dir = os.getcwd()
config_dir = os.path.join(work_dir, "config", "env")
config_file = os.path.join(config_dir, "go_reader.env")

# Verzeichnis erstellen, falls es nicht existiert
os.makedirs(config_dir, exist_ok=True)

# Konfigurationsdatei erstellen
config_content = f"""RS485_PORT={serial_port}
RS485_BAUDRATE={baudrate}
RS485_DATABITS={databits}
RS485_PARITY={parity}
RS485_STOPBITS={stopbits}
RS485_TIMEOUT_MS={timeout_ms}
RS485_THINGSBOARD_SERVER={thingsboard_server}
RS485_THINGSBOARD_PORT={thingsboard_port}
RS485_ACCESS_TOKEN={access_token}
"""

with open(config_file, 'w') as f:
    f.write(config_content)

print(f"Konfigurationsdatei erstellt unter: {config_file}")
print("Starte Modbus-Reader...")

# Umgebungsvariable setzen und den Go-Modbus-Reader starten
env = os.environ.copy()
env["GO_READER_ENV_PATH"] = config_file

try:
    # Go-Programm starten
    subprocess.run(["go", "run", "cmd/main.go"], env=env, check=True)
except KeyboardInterrupt:
    print("\nModbus-Reader wurde durch Benutzer beendet")
except subprocess.CalledProcessError as e:
    print(f"Fehler beim Ausführen des Go-Programms: {e}")
    sys.exit(1)
except Exception as e:
    print(f"Unerwarteter Fehler: {e}")
    sys.exit(1) 