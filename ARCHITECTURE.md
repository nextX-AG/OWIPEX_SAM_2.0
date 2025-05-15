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

### 4. Neue modulare Geräte-Architektur (`internal/device/`)

Die Geräte-Architektur verwendet eine mehrstufige Abstraktion, die Sensortypen, Kommunikationsprotokolle und herstellerspezifische Konfigurationen trennt:

#### 4.1 Basisinterfaces und -strukturen (`internal/device/` und `internal/types/`)
- **types/device.go** - Definiert grundlegende Interfaces für alle Geräte:
  - `Device` - Basisinterface für alle Geräte
  - `ReadableDevice` - Interface für lesende Geräte
  - `WritableDevice` - Interface für schreibende Geräte
  - `Sensor` - Spezialisiertes Interface für Sensoren
  - `Actor` - Spezialisiertes Interface für Aktoren
  - `HybridDevice` - Interface für Geräte, die lesen und schreiben können
- **device/registry.go** - Zentrales Register für alle verfügbaren Geräte
- **device/factory.go** - Factory-Pattern für die Geräteerstellung
- **device/loader.go** - Funktionen zum Laden von Gerätekonfigurationen

Ein zentraler Aspekt der Architektur ist die Vermeidung zirkulärer Abhängigkeiten durch das `types`-Paket, das alle gemeinsamen Interfaces enthält:

```go
// Beispiel aus internal/types/device.go
type Device interface {
    ID() string
    Name() string
    Type() DeviceType
    Metadata() map[string]interface{}
    IsEnabled() bool
    Enable(enabled bool)
    Close() error
}

type ReadableDevice interface {
    Device
    Read(ctx context.Context) (Reading, error)
    ReadRaw(ctx context.Context) ([]byte, error)
    AvailableReadings() []ReadingType
}
```

#### 4.2 Sensortypen (`internal/device/sensor/`)
Jeder Sensortyp hat seine eigene Implementierung mit einer gemeinsamen Basisklasse:
- **sensor/base.go** - Gemeinsame Basisfunktionalität für alle Sensoren
- **sensor/ph/ph_sensor.go** - pH-Sensor-Implementierung
- **sensor/flow/** - Implementierungen für Durchflusssensoren (geplant)
- **sensor/radar/** - Implementierungen für Radar-Füllstandsensoren (geplant)
- **sensor/turbidity/** - Implementierungen für Trübungssensoren (geplant)

Diese Implementierungen enthalten die Logik zur Verarbeitung und Interpretation der spezifischen Messwerte, unabhängig vom verwendeten Kommunikationsprotokoll oder Hersteller.

#### 4.3 Aktortypen (`internal/device/actor/`)
Jeder Aktortyp hat seine eigene Implementierung:
- **valve/** - Implementierungen für Ventile
- **pump/** - Implementierungen für Pumpen

#### 4.4 Protokolle (`internal/protocol/`)
Kommunikationsprotokolle werden vollständig abstrahiert:
- **types/protocol.go** - Gemeinsame Protokoll-Interfaces
- **protocol/modbus/client.go** - Modbus-Protokollimplementierung

#### 4.5 Gerätekonfiguration (`config/devices/`)
Herstellerspezifische Details werden in Konfigurationsdateien ausgelagert:
```
config/
  devices/
    sensors/
      ph/
        hersteller_a.json  // Konfiguration für Hersteller A
        hersteller_b.json  // Konfiguration für Hersteller B
      flow/
        hersteller_c.json
      ...
    actors/
      valve/
        hersteller_d.json
      ...
```

Beispielkonfiguration:
```json
{
  "manufacturer": "HerstellerA",
  "model": "PH-2000",
  "type": "ph_sensor",
  "protocol": "modbus",
  "modbus": {
    "slave_id": 1,
    "registers": {
      "ph_value": {
        "address": 100,
        "length": 2,
        "data_type": "float32",
        "byte_order": "big_endian"
      }
    }
  },
  "calibration": {
    "ph": {
      "offset": 0.0,
      "scale_factor": 1.0
    }
  }
}
```

#### 4.6 Vorteile der neuen Architektur
- **Modularer Aufbau:** Klare Trennung zwischen Sensortyp, Kommunikation und Herstellerspezifika
- **Erweiterbarkeit:** Neue Sensoren durch einfaches Hinzufügen von Konfigurationsdateien
- **Austauschbarkeit:** Einfacher Austausch von Sensoren durch Ändern der Konfiguration
- **Konfigurierbarkeit:** Alle herstellerspezifischen Details als Konfiguration, nicht im Code
- **Typsicherheit:** Strukturierte Datentypen für Messwerte und Befehle
- **Einfache Wartung:** Bei Änderungen muss meist nur die Konfiguration angepasst werden
- **Vermeidung zirkulärer Abhängigkeiten:** Durch zentrales Typen-Paket
- **Testbarkeit:** Klare Interfaces für einfaches Mocking und Unit-Testing

## Migration von der alten zur neuen Architektur

Die Migration der bestehenden Sensoren in die neue Architektur erfolgt schrittweise:

1. **Implementierung der Basisstrukturen:**
   - Zentrale Typen und Interfaces (`internal/types/`)
   - Geräteregistry und Factory-Muster (`internal/device/`)
   - Protokollabstraktion (`internal/protocol/`)
   
2. **Migration der Sensortypen:**
   - Jeder Sensortyp (pH, Flow, Radar, Turbidity) wird einzeln migriert
   - Für jeden Typ wird ein eigenes Unterpaket mit Typ-spezifischer Logik erstellt
   - Die Konfiguration wird in JSON-Dateien ausgelagert

3. **Konfigurationsumstellung:**
   - Migration von hartkodierten Parametern zu konfigurierbaren Werten
   - Erstellung von Standardkonfigurationen für gängige Sensormodelle
   - Validierung und Dokumentation der Konfigurationsoptionen

4. **Anpassung der ThingsBoard-Integration:**
   - Umstellung auf die neue Gerätestruktur
   - Zuordnung von Messwerten zu Telemetriedaten
   - Erweiterung der RPC-Funktionen für die neuen Gerätetypen

### 5. ThingsBoard-Integration
#### 5.1 Bestehende Implementierung (`internal/thingsboard/`)
- **thingsboard_client.go** - Handhabt Kommunikation mit ThingsBoard-Plattform
- MQTT-basierte Kommunikation mit der ThingsBoard-Plattform
- Sendet Telemetriedaten von den Sensoren
- Empfängt Shared Attributes für Konfigurationsänderungen
- Verarbeitet RPC-Befehle für Fernsteuerung
- Behandelt Verbindungsabbrüche und Wiederverbindung

#### 5.2 Neue modulare Implementierung (`internal/thingsboardMQTT/`)
Vollständig modulare Implementierung aller ThingsBoard MQTT-APIs:
- **types.go** - Gemeinsame Typen und Strukturdefinitionen
- **client.go** - Basis-Client mit Kern-Funktionalität (Verbindung, Start/Stop)
- **telemetry.go** - Senden von Telemetriedaten
- **attributes.go** - Verwaltung von Client- und Shared-Attributen
- **rpc.go** - Serverseitige und Clientseitige RPC-Funktionen
- **firmware.go** - API für Firmware-Updates
- **provisioning.go** - Funktionen für Device Provisioning und Claiming
- **utils.go** - Hilfsfunktionen

Vorteile der neuen Implementierung:
- Klare Trennung der Funktionalitäten in separate Module
- Bessere Wartbarkeit und Erweiterbarkeit
- Umfassende Fehlerbehandlung und Thread-Sicherheit
- Vollständige Abdeckung aller ThingsBoard MQTT-APIs
- Flexibles Konfigurations-System über Options-Pattern

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

2. **Gerätekonfiguration:**
   - Liste der verfügbaren Sensoren und Aktoren
   - Pro Gerät: Typ, Modbus-ID, Register-Adressen, Konfigurationsparameter
   - Kalibrierungsparameter für jeden Sensor
   - Herstellerspezifische Details für verschiedene Gerätemodelle

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
- Einfaches Hinzufügen neuer Gerätemodelle durch Konfigurationsdateien

## Zukünftige Verbesserungen

- Dashboard für Sensorstatus hinzufügen
- Caching für Sensorwerte implementieren
- Unterstützung für weitere Sensortypen hinzufügen
- Datenpufferung bei ThingsBoard-Verbindungsverlust
- Web-Interface für lokale Konfiguration 