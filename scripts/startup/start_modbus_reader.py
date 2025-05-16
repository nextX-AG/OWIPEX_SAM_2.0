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
def run_command(command, check=True, retry_on_lock=True, max_retries=5, shell=False):
    """Führt einen Shell-Befehl aus und gibt die Ausgabe zurück"""
    print(f"Führe aus: {' '.join(command) if not shell else command}")
    
    # Bei apt-get Befehlen prüfen, ob dpkg gesperrt ist
    retry_count = 0
    while retry_count < max_retries:
        try:
            if shell:
                result = subprocess.run(command, check=check, text=True, capture_output=True, shell=True)
            else:
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
            if retry_on_lock and not shell and ("dpkg" in " ".join(command) or "apt-get" in " ".join(command)):
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
                print("  sudo apt-get install -y golang-go build-essential git")
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

def install_go_manually():
    """Installiert Go manuell durch Herunterladen und Entpacken des Binärpakets"""
    print("Installiere Go manuell...")
    
    # Go-Version
    go_version = "1.19.13"
    go_file = f"go{go_version}.linux-amd64.tar.gz"
    go_url = f"https://golang.org/dl/{go_file}"
    
    # Arbeitsverzeichnis für Downloads
    temp_dir = "/tmp"
    download_path = os.path.join(temp_dir, go_file)
    
    try:
        # Herunterladen der Go-Binärdatei
        print(f"Lade Go {go_version} herunter...")
        run_command(["wget", "-O", download_path, go_url], check=False)
        
        if not os.path.exists(download_path):
            print(f"Fehler: Konnte Go nicht herunterladen. Datei {download_path} existiert nicht.")
            return False
        
        # Entpacken nach /usr/local
        print("Entpacke Go nach /usr/local...")
        is_root = os.geteuid() == 0
        sudo_cmd = [] if is_root else ["sudo"]
        
        # Entferne altes Go-Verzeichnis und entpacke das neue
        run_command(sudo_cmd + ["rm", "-rf", "/usr/local/go"], check=False)
        extract_cmd = sudo_cmd + ["tar", "-C", "/usr/local", "-xzf", download_path]
        if run_command(extract_cmd, check=False).returncode != 0:
            print("Fehler: Konnte Go nicht entpacken.")
            return False
        
        # PATH setzen
        print("Setze PATH-Variable...")
        go_bin_path = "/usr/local/go/bin"
        
        # PATH für aktuelle Sitzung setzen
        os.environ["PATH"] = f"{go_bin_path}:{os.environ.get('PATH', '')}"
        
        # PATH permanent setzen
        profile_path = os.path.expanduser("~/.profile")
        
        # Prüfe, ob der Pfad bereits in .profile ist
        add_path = True
        if os.path.exists(profile_path):
            with open(profile_path, "r") as f:
                if f"/usr/local/go/bin" in f.read():
                    add_path = False
        
        if add_path:
            with open(profile_path, "a") as f:
                f.write('\nexport PATH=$PATH:/usr/local/go/bin\n')
            
            # Lade .profile neu
            run_command(["source", os.path.expanduser("~/.profile")], check=False, shell=True)
        
        # Prüfe Go-Installation
        if check_go_installed():
            return True
        else:
            print("Fehler: Go wurde installiert, ist aber nicht im PATH verfügbar.")
            print("Bitte führen Sie manuell aus:")
            print("  export PATH=$PATH:/usr/local/go/bin")
            print("  source ~/.profile")
            return False
            
    except Exception as e:
        print(f"Fehler bei der manuellen Installation von Go: {e}")
        return False

def install_linux_dependencies():
    """Installiert benötigte Abhängigkeiten auf Ubuntu/Linux"""
    print("Installiere Abhängigkeiten für Ubuntu/Linux...")

    # Prüfe, ob Go bereits installiert ist
    go_installed = check_go_installed()
    if not go_installed:
        print("Go ist nicht installiert.")
        
        # Verwende die manuelle Installation direkt
        go_installed = install_go_manually()
        if not go_installed:
            print("Go konnte nicht installiert werden. Das Programm benötigt Go zum Ausführen.")
            return False
    
    # Prüfen, ob wir Root-Rechte haben
    is_root = os.geteuid() == 0
    sudo_cmd = [] if is_root else ["sudo"]
    
    # Aktualisiere Paketlisten, überspringen wenn dpkg gesperrt ist
    update_result = run_command(sudo_cmd + ["apt-get", "update"], check=False, retry_on_lock=True)
    
    # Wenn Pakete nicht aktualisiert werden können, trotzdem fortfahren
    if update_result and update_result.returncode != 0:
        print("Warnung: Konnte Paketlisten nicht aktualisieren, fahre fort...")
    
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
    subprocess.run(["go", "run", "cmd/reader/main.go"], env=env, check=True)
except KeyboardInterrupt:
    print("\nModbus-Reader wurde durch Benutzer beendet")
except subprocess.CalledProcessError as e:
    print(f"Fehler beim Ausführen des Go-Programms: {e}")
    sys.exit(1)
except Exception as e:
    print(f"Unerwarteter Fehler: {e}")
    sys.exit(1) 