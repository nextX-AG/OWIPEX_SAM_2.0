# owipexRS485GO - Systemarchitektur

## Übersicht

Das owipexRS485GO-System ist eine in Go implementierte Kommunikationsbrücke zwischen Wasseraufbereitungssensoren (über RS485/Modbus) und der ThingsBoard IoT-Plattform. Die Anwendung wurde von einem Python-Projekt portiert, um die Stabilität, Leistung und Zuverlässigkeit zu verbessern.

```
+----------------+        +-------------------+        +-----------------+
| Sensoren       |        | owipexRS485GO     |        | ThingsBoard     |
| (RS485/Modbus) | <----> | (Go Application)  | <----> | (IoT Platform)  |
+----------------+        +-------------------+        +-----------------+
                                    ^
                                    |
                          +---------+---------+
                          | Konfiguration     |
                          | (JSON/Umgebung)   |
                          +-------------------+
```

## Hauptkomponenten

### 1. Command Layer (`cmd/`)
- **main.go** - Anwendungseinstiegspunkt und übergeordnete Orchestrierung
- Initialisiert und koordiniert alle Komponenten
- Implementiert Hauptschleife für periodisches Polling
- Verwaltet Ressourcen und Shutdown
- Behandelt Signale (SIGTERM, SIGINT)

### 2. Konfiguration (`config/`, `internal/config/`)
- **sensors.json** - Sensordefinitionen und Konfiguration
- **config.go** - Laden und Verwalten von Konfigurationen
- **config_test.go** - Tests für Konfigurationsfunktionen
- Lädt Konfigurationen aus JSON-Dateien und Umgebungsvariablen
- Validiert Konfigurationsparameter
- Stellt Konfigurationen für andere Komponenten bereit
- Unterstützt Neuladen der Konfiguration zur Laufzeit

### 3. Modbus-Kommunikation (`internal/modbus/`)
- **modbus_client.go** - Handhabt RS485 Modbus-Kommunikation
- **modbus_client_test.go** - Tests für Modbus-Client
- Zuständig für die Kommunikation mit den Sensoren über RS485/Modbus-Protokoll
- Implementiert Lese- und Schreiboperationen für Register
- Behandelt Timeouts und Verbindungsfehler
- Unterstützt verschiedene Modbus-Funktionscodes

### 4. Sensor-Management (`internal/sensor/`, `internal/manager/`)
- **sensor.go** - Basis-Sensor-Interface
- **sensor_manager.go** - Verwaltet alle Sensoroperationen
- Spezifische Sensorimplementierungen:
  - **flow_sensor.go** - Durchflusssensor
  - **ph_sensor.go** - pH-Wert-Sensor
  - **radar_sensor.go** - Radarsensor für Füllstandmessung
  - **turbidity_sensor.go** - Trübungssensor
- Abstrahiert die verschiedenen Sensortypen
- Konvertiert Rohwerte aus Modbus-Registern in physikalische Messwerte
- Implementiert sensorspezifische Kalibrierungen und Berechnungen
- Verwaltet den Zustand der Sensoren (aktiv/inaktiv)

### 5. ThingsBoard-Integration (`internal/thingsboard/`)
- **thingsboard_client.go** - Handhabt Kommunikation mit ThingsBoard-Plattform
- MQTT-basierte Kommunikation mit der ThingsBoard-Plattform
- Sendet Telemetriedaten von den Sensoren
- Empfängt Shared Attributes für Konfigurationsänderungen
- Verarbeitet RPC-Befehle für Fernsteuerung
- Behandelt Verbindungsabbrüche und Wiederverbindung

## Datenfluss

1. **Sensordatenerfassung:**
   - Die Anwendung pollt regelmäßig die konfigurierten Sensoren über Modbus
   - Rohwerte werden in physikalische Messgrößen umgewandelt
   - Messwerte werden für die Übertragung an ThingsBoard aufbereitet

2. **Datenübertragung an ThingsBoard:**
   - Messwerte werden als Telemetriedaten über MQTT an ThingsBoard gesendet
   - Die Daten werden mit Zeitstempeln versehen
   - Bei Verbindungsverlust werden die Daten gepuffert (zukünftige Funktion)

3. **Empfang von Konfigurationsänderungen:**
   - ThingsBoard sendet Konfigurationsänderungen als Shared Attributes
   - Der ThingsBoard-Client empfängt diese Änderungen und leitet sie weiter
   - Der Konfigurations-Manager aktualisiert die Konfiguration entsprechend
   - Die Änderungen werden an die betroffenen Komponenten weitergegeben

4. **RPC-Befehle für Fernsteuerung:**
   - ThingsBoard sendet RPC-Befehle für Aktionen (z.B. Kalibrierung)
   - Der ThingsBoard-Client empfängt und verarbeitet diese Befehle
   - Aktionen werden ausgeführt und Ergebnisse zurück an ThingsBoard gemeldet

## Fehlerbehandlung

Fehlerbehandlung erfolgt auf jeder Ebene mit angemessener Protokollierung und Wiederherstellungsmechanismen:

- **Verbindungsabbrüche:**
  - Automatische Wiederverbindung zu ThingsBoard bei Verbindungsverlust
  - Wiederholungsversuche für fehlgeschlagene Modbus-Kommunikation

- **Timeouts:**
  - Timeout-Handling für alle externen Kommunikationsvorgänge
  - Kontrollierte Wiederholung fehlgeschlagener Operationen

- **Logging:**
  - Umfassende Protokollierung für Diagnose und Fehlerbehebung
  - Verschiedene Log-Level für Entwicklung und Produktion

## Konfigurationsstruktur

Die Anwendung verwendet mehrere Konfigurationsebenen:

1. **Hauptkonfigurationsdatei:**
   - Allgemeine Einstellungen (Logging, Polling-Intervall, etc.)
   - ThingsBoard-Verbindungsparameter (Host, Port, Token)
   - Modbus-Parameter (Port, Baudrate, Parity, etc.)

2. **Sensorkonfiguration:**
   - Liste der verfügbaren Sensoren
   - Pro Sensor: Typ, Modbus-ID, Register-Adressen, Enabled-Flag
   - Kalibrierungsparameter für jeden Sensor

3. **Umgebungsvariablen:**
   - Überschreiben von Konfigurationsparametern
   - Anpassung an verschiedene Umgebungen (Entwicklung, Produktion)

## Deployment

Die Anwendung unterstützt verschiedene Deployment-Szenarien:

1. **Als Systemd-Service:**
   - Automatischer Start beim Systemstart
   - Neustart bei Absturz
   - Log-Integration in systemd-Journaling

2. **Standalone-Ausführung:**
   - Start über Shell-Skripte
   - Unterstützung für verschiedene Plattformen

3. **Automatisierte Installation:**
   - Skripte für die Installation aller Abhängigkeiten
   - Erstellung notwendiger Konfigurationsdateien

## Sicherheit

- Unterstützung für verschlüsselte MQTT-Verbindungen (zukünftig)
- Trennung von Konfiguration und Code
- Sichere Speicherung von Zugangsdaten

## Erweiterbarkeit

Das System ist modular aufgebaut und kann einfach erweitert werden:

- Hinzufügen neuer Sensortypen ohne Änderung des Kerncodex
- Unterstützung für alternative IoT-Plattformen neben ThingsBoard
- Erweiterung um zusätzliche Funktionen durch neue Module

## Zukünftige Verbesserungen

- Dashboard für Sensorstatus hinzufügen
- Caching für Sensorwerte implementieren
- Unterstützung für weitere Sensortypen hinzufügen
- Datenpufferung bei ThingsBoard-Verbindungsverlust
- Web-Interface für lokale Konfiguration 