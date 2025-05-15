#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
ThingsBoard Variablen Test Starter Script
Dieses Skript kompiliert und startet das test_thingsboard_variables.go Programm
mit den erforderlichen Umgebungsvariablen.
"""

import os
import sys
import subprocess
import time
import signal
import shutil
import logging
from typing import Optional

# Logging-Konfiguration
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    datefmt='%Y-%m-%d %H:%M:%S'
)
logger = logging.getLogger("TB-Test-Starter")

# Standard Konfigurationswerte
DEFAULT_CONFIG = {
    "TB_HOST": "146.4.67.141",
    "TB_PORT": "1883",
    "TB_TOKEN": "5Ohlb6ZKO4uNw9O2DHwk"
}

def check_go_installation() -> bool:
    """Prüft, ob Go installiert ist und gibt True zurück, wenn es verfügbar ist."""
    try:
        result = subprocess.run(
            ["go", "version"], 
            stdout=subprocess.PIPE, 
            stderr=subprocess.PIPE, 
            text=True
        )
        if result.returncode == 0:
            logger.info(f"Go ist installiert: {result.stdout.strip()}")
            return True
        else:
            logger.error("Go ist nicht korrekt installiert")
            return False
    except FileNotFoundError:
        logger.error("Go ist nicht installiert oder nicht im PATH")
        return False

def build_program() -> bool:
    """Kompiliert das Go-Programm und gibt True zurück, wenn erfolgreich."""
    logger.info("Kompiliere test_thingsboard_variables.go...")
    try:
        result = subprocess.run(
            ["go", "build", "-o", "thingsboard_test", "test_thingsboard_variables.go"],
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True
        )
        if result.returncode == 0:
            logger.info("Kompilierung erfolgreich")
            return True
        else:
            logger.error(f"Kompilierungsfehler: {result.stderr}")
            return False
    except Exception as e:
        logger.error(f"Fehler beim Kompilieren: {e}")
        return False

def run_program(env: dict) -> Optional[subprocess.Popen]:
    """Startet das kompilierte Programm mit den angegebenen Umgebungsvariablen."""
    logger.info("Starte ThingsBoard Test Programm...")
    try:
        # Umgebungsvariablen zum aktuellen Prozess hinzufügen
        process_env = os.environ.copy()
        process_env.update(env)
        
        # Programm starten
        process = subprocess.Popen(
            ["./thingsboard_test"],
            env=process_env,
            stdout=subprocess.PIPE,
            stderr=subprocess.STDOUT,
            text=True,
            bufsize=1
        )
        
        logger.info(f"Programm gestartet mit PID {process.pid}")
        return process
    except Exception as e:
        logger.error(f"Fehler beim Starten des Programms: {e}")
        return None

def main():
    """Hauptfunktion, die das Programm kompiliert und ausführt."""
    logger.info("ThingsBoard Variablen Test Starter wird ausgeführt")
    
    # Prüfen, ob Go installiert ist
    if not check_go_installation():
        logger.error("Go muss installiert sein, um fortzufahren")
        sys.exit(1)
    
    # Umgebungsvariablen vorbereiten
    config = DEFAULT_CONFIG.copy()
    
    # Überschreibe mit vorhandenen Umgebungsvariablen
    for key in config:
        if key in os.environ:
            config[key] = os.environ[key]
            logger.info(f"Verwende Umgebungsvariable {key}={config[key]}")
    
    # Zeige Konfiguration
    logger.info(f"Verwende ThingsBoard-Konfiguration: Host={config['TB_HOST']}, Port={config['TB_PORT']}")
    
    # Programm kompilieren
    if not build_program():
        logger.error("Konnte das Programm nicht kompilieren")
        sys.exit(1)
    
    # Signal-Handler einrichten
    def signal_handler(sig, frame):
        logger.info("Signal erhalten, beende Programm...")
        if process and process.poll() is None:
            process.terminate()
            process.wait(timeout=5)
        sys.exit(0)
    
    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)
    
    # Programm starten
    process = run_program(config)
    if not process:
        logger.error("Konnte das Programm nicht starten")
        sys.exit(1)
    
    # Output des Programms weiterleiten
    try:
        while True:
            line = process.stdout.readline()
            if not line and process.poll() is not None:
                break
            if line:
                print(line.rstrip())
    except KeyboardInterrupt:
        logger.info("Benutzerabbruch erkannt, beende Programm...")
    finally:
        if process.poll() is None:
            process.terminate()
            try:
                process.wait(timeout=5)
            except subprocess.TimeoutExpired:
                process.kill()
    
    exit_code = process.returncode
    logger.info(f"Programm beendet mit Exit-Code {exit_code}")
    return exit_code

if __name__ == "__main__":
    sys.exit(main()) 