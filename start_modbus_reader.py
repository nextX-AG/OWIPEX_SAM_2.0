#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import os
import subprocess
import sys
import platform
import shutil
import time
import re

# Funktionen für die Installation von Abhängigkeiten
def run_command(command, check=True, retry_on_lock=True, max_retries=5):
    """Führt einen Shell-Befehl aus und gibt die Ausgabe zurück"""
    print(f"Führe aus: {' '.join(command)}")
    
    # Bei apt-get Befehlen prüfen, ob dpkg gesperrt ist
    retry_count = 0
    while retry_count < max_retries:
        try:
            result = subprocess.run(command, check=check, text=True, capture_output=True)
            if result.stdout:
                print(result.stdout)
            return result
        except subprocess.CalledProcessError as e:
            print(f"Fehler beim Ausführen des Befehls: {e}")
            if result.stdout:
                print(result.stdout)
            if result.stderr:
                print(result.stderr, file=sys.stderr)
            
            # Prüfe auf dpkg lock Fehler
            if retry_on_lock and ("dpkg" in " ".join(command) or "apt-get" in " ".join(command)):
                if "Sperre" in e.stderr or "lock" in e.stderr or "E: Konnte Sperre nicht bekommen" in e.stderr:
                    # Versuche herauszufinden, welcher Prozess die Sperre hält
                    retry_count += 1
                    print(f"Paket-Manager ist blockiert. Versuch {retry_count}/{max_retries}")
                    
                    # Finde PID des blockierenden Prozesses
                    match = re.search(r"Prozess (\d+)", e.stderr)
                    if match:
                        pid = match.group(1)
                        print(f"Prozess {pid} blockiert den Paket-Manager.")
                        
                        # Zeige Informationen zum blockierenden Prozess
                        try:
                            ps_result = subprocess.run(["ps", "-p", pid, "-o", "comm="], 
                                                    text=True, capture_output=True, check=False)
                            if ps_result.stdout:
                                print(f"Blockierender Prozess: {ps_result.stdout.strip()}")
                        except:
                            pass
                    
                    print("Warte 30 Sekunden und versuche es erneut...")
                    time.sleep(30)
                    continue
            
            if check:
                print("Abhängigkeiten müssen manuell installiert werden:")
                print("  sudo apt-get update")
                print("  sudo apt-get install -y golang build-essential git")
                print("  sudo usermod -a -G dialout $USER")
                if retry_on_lock:
                    sys.exit(1)
            return e

    print(f"Maximale Anzahl an Versuchen ({max_retries}) erreicht. Fahre ohne Installation fort.")
    return None

def check_go_installed():
    """Prüft, ob Go installiert ist, ohne apt-get zu verwenden"""
    if shutil.which("go"):
        try:
            # Prüfe Go-Version
            result = subprocess.run(["go", "version"], text=True, capture_output=True, check=False)
            if result.returncode == 0:
                print(f"Go ist bereits installiert: {result.stdout.strip()}")
                return True
        except:
            pass
    return False

def install_linux_dependencies():
    """Installiert benötigte Abhängigkeiten auf Ubuntu/Linux"""
    print("Installiere Abhängigkeiten für Ubuntu/Linux...")

    # Prüfen, ob wir Root-Rechte haben
    is_root = os.geteuid() == 0
    sudo_cmd = [] if is_root else ["sudo"]
    
    go_installed = check_go_installed()

    # Aktualisiere Paketlisten, überspringen wenn dpkg gesperrt ist
    update_result = run_command(sudo_cmd + ["apt-get", "update"], check=False, retry_on_lock=True)
    
    # Wenn Pakete nicht aktualisiert werden können, trotzdem fortfahren
    if update_result and update_result.returncode != 0:
        print("Warnung: Konnte Paketlisten nicht aktualisieren, fahre fort...")
    
    # Installiere Go, falls nicht vorhanden
    if not go_installed:
        print("Go ist nicht installiert. Installiere Go...")
        # Zuerst versuchen mit golang-go (korrekt für Ubuntu)
        install_result = run_command(sudo_cmd + ["apt-get", "install", "-y", "golang-go"], 
                                  check=False, retry_on_lock=True)
        if install_result and install_result.returncode != 0:
            # Wenn der erste Versuch fehlschlägt, versuche mit golang (für andere Distributionen)
            print("Versuche alternative Installation mit 'golang' Paket...")
            install_result = run_command(sudo_cmd + ["apt-get", "install", "-y", "golang"], 
                                          check=False, retry_on_lock=True)
            if install_result and install_result.returncode != 0:
                # Wenn beide fehlschlagen, versuche snap
                print("Versuche Installation mit snap...")
                install_result = run_command(sudo_cmd + ["snap", "install", "go", "--classic"], 
                                      check=False, retry_on_lock=True)
    
    # Installiere weitere benötigte Pakete, aber nur wenn apt nicht gesperrt ist
    pkg_result = run_command(sudo_cmd + ["apt-get", "install", "-y", "build-essential", "git"], 
                         check=False, retry_on_lock=True)
    
    # Füge den Benutzer zur Gruppe dialout hinzu (für seriellen Port-Zugriff)
    # Dies sollte auch funktionieren, wenn apt-get noch gesperrt ist
    user = os.environ.get("USER", os.environ.get("USERNAME", ""))
    if user:
        user_result = run_command(sudo_cmd + ["usermod", "-a", "-G", "dialout", user], check=False)
        if user_result and user_result.returncode == 0:
            print("WICHTIG: Sie müssen sich möglicherweise neu anmelden, damit die Gruppenmitgliedschaft 'dialout' wirksam wird.")
    
    # Installiere Go-Abhängigkeiten
    if go_installed:
        print("Installiere Go-Abhängigkeiten...")
        run_command(["go", "mod", "tidy"], check=False)
    
    return go_installed

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
dependencies_installed = True
if current_os == "Linux":
    try:
        dependencies_installed = install_linux_dependencies()
    except Exception as e:
        print(f"Fehler bei der Installation der Abhängigkeiten: {e}")
        print("Installieren Sie die benötigten Pakete manuell: golang, build-essential, git")
        print("Und fügen Sie Ihren Benutzer zur Gruppe 'dialout' hinzu: sudo usermod -a -G dialout $USER")
        dependencies_installed = False

    # Prüfe serielle Ports auf Linux-Systemen
    # Suche nach verfügbaren seriellen Ports
    available_ports = []
    port_patterns = ['/dev/ttyS*', '/dev/ttyUSB*', '/dev/ttyACM*']
    for pattern in port_patterns:
        try:
            found_ports = subprocess.check_output(f"ls -l {pattern} 2>/dev/null || true", shell=True).decode().strip()
            if found_ports:
                for line in found_ports.split('\n'):
                    if 'tty' in line:
                        port = line.split()[-1]
                        available_ports.append(port)
        except:
            pass

    # Wenn verfügbare Ports gefunden wurden, nehme den ersten
    if available_ports:
        print(f"Gefundene serielle Ports: {', '.join(available_ports)}")
        # Wähle den ersten USB-Port, wenn verfügbar, sonst den ersten in der Liste
        usb_ports = [p for p in available_ports if 'USB' in p]
        if usb_ports:
            port_path = usb_ports[0]
        else:
            port_path = available_ports[0]
        serial_ports['Linux'] = f"rtu://{port_path}"
        print(f"Verwende seriellen Port: {port_path}")
    else:
        # Prüfe, ob der konfigurierte Port existiert
        port_path = serial_ports['Linux'].replace('rtu://', '')
        if not os.path.exists(port_path):
            print(f"WARNUNG: Der konfigurierte serielle Port {port_path} existiert nicht!")
            print("Keine seriellen Ports gefunden. Bitte schließen Sie ein RS485-Gerät an.")

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

if not dependencies_installed:
    print("Abhängigkeiten nicht vollständig installiert. Das Programm kann möglicherweise nicht korrekt starten.")
    user_input = input("Möchten Sie dennoch fortfahren? (j/n): ")
    if user_input.lower() not in ['j', 'ja', 'y', 'yes']:
        print("Abbruch.")
        sys.exit(1)

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