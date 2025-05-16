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

## Neue Projektstruktur

Die Projekt-Ordnerstruktur wurde vollständig überarbeitet, um eine modulare und wartbare Codebasis zu schaffen:

```
owipexRS485GO/
├── cmd/                      # Ausführbare Anwendungen
│   ├── reader/               # Hauptanwendung
│   └── tools/                # Hilfswerkzeuge, Tests, etc.
│
├── internal/                 # Nicht-öffentlicher Code
│   ├── config/               # Konfigurationslogik
│   ├── controller/           # Controller-Logik (Steuerungsalgorithmen)
│   │   ├── flow/             # Durchfluss-Controller
│   │   ├── ph/               # pH-Wert-Controller
│   │   └── system/           # Systemsteuerung
│   │
│   ├── device/               # Geräteabstraktionen
│   │   ├── actuator/         # Aktoren (Relais, Ventile, etc.)
│   │   │   ├── relay/        # Relais-spezifischer Code
│   │   │   └── valve/        # Ventil-spezifischer Code
│   │   │
│   │   └── sensor/           # Sensoren
│   │       ├── flow/         # Durchflusssensoren
│   │       ├── ph/           # pH-Sensoren
│   │       ├── radar/        # Radarsensoren
│   │       └── turbidity/    # Trübungssensoren
│   │
│   ├── hardware/             # Hardware-Abstraktionen
│   │   ├── gpio/             # GPIO-Schnittstelle
│   │   └── uart/             # UART-Schnittstelle
│   │
│   ├── integration/          # Integration mit externen Plattformen
│   │   ├── thingsboard/      # ThingsBoard-spezifischer Code
│   │   │   ├── mqtt/         # MQTT-Kommunikation mit ThingsBoard
│   │   │   └── rest/         # REST-API-Kommunikation mit ThingsBoard
│   │   └── other_platform/   # Andere Plattformen (zukünftig)
│   │
│   ├── protocol/             # Kommunikationsprotokolle
│   │   ├── factory/          # Factory für Protokoll-Handler
│   │   └── modbus/           # Modbus-Implementierung
│   │
│   ├── service/              # Anwendungsdienste
│   │   ├── monitoring/       # Überwachungsdienste
│   │   └── scheduler/        # Zeitplanungsdienste
│   │
│   ├── storage/              # Datenspeicherung
│   │
│   └── types/                # Gemeinsame Typen und Interfaces
│
├── pkg/                      # Potenziell wiederverwendbare Pakete
│
└── scripts/                  # Hilfsskripte für Entwicklung, Deployment, etc.
```

### Vorteile der neuen Struktur

- **Klare Trennung der Verantwortlichkeiten**: Jeder Ordner hat eine definierte Aufgabe
- **Erweiterbarkeit**: Neue Module können einfach hinzugefügt werden
- **Konsistente Begrifflichkeiten**: Controller, Devices, Hardware, Protokolle, etc. sind klar voneinander getrennt
- **Zukunftssicherheit**: Platz für zukünftige Komponenten wie GPIO, weitere Controller, etc.
- **Modulare Integration**: Klare Trennung zwischen Plattformen und Kommunikationsprotokollen

## Hauptkomponenten

### 1. Command Layer (`cmd/`)
- **reader/main.go** - Anwendungseinstiegspunkt und übergeordnete Orchestrierung
- Initialisiert und koordiniert alle Komponenten
- Implementiert Hauptschleife für periodisches Polling
- Verwaltet Ressourcen und Shutdown
- Behandelt Signale (SIGTERM, SIGINT)

### 2. Konfiguration (`internal/config/`)
- **config.go** - Laden und Verwalten von Konfigurationen
- **config_test.go** - Tests für Konfigurationsfunktionen
- Lädt Konfigurationen aus JSON-Dateien und Umgebungsvariablen
- Validiert Konfigurationsparameter
- Stellt Konfigurationen für andere Komponenten bereit
- Unterstützt Neuladen der Konfiguration zur Laufzeit

### 3. Modbus-Kommunikation (`internal/protocol/modbus/`)
- **client.go** - Implementiert das `types.ProtocolHandler`-Interface
- **test/test_client.go** - Test-Client für die Modbus-Implementierung
- Vollständig konfigurierbar über JSON-Dateien
- Unterstützt verschiedene Register-Typen (Holding, Input, Coil, Discrete)
- Protokoll-agnostisches Interface für Sensorimplementierungen
- Thread-sicher durch Mutex-Schutz für alle Operationen

### 4. Protokoll-Factory (`internal/protocol/factory/`)
- **protocol_factory.go** - Factory für die Erstellung von Protokoll-Handlern
- Erstellt Protokoll-Handler basierend auf Konfigurationen
- Unterstützt verschiedene Protokolltypen (derzeit Modbus)
- Extrahiert Konfigurationsparameter aus JSON-Strukturen

### 5. Geräte-Architektur (`internal/device/`)

Die Geräte-Architektur verwendet eine mehrstufige Abstraktion, die Sensortypen, Kommunikationsprotokolle und herstellerspezifische Konfigurationen trennt:

#### 5.1 Basisinterfaces und -strukturen (`internal/device/` und `internal/types/`)
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

#### 5.2 Sensortypen (`internal/device/sensor/`)
Jeder Sensortyp hat seine eigene Implementierung mit einer gemeinsamen Basisklasse:
- **sensor/base.go** - Gemeinsame Basisfunktionalität für alle Sensoren
- **sensor/ph/ph_sensor.go** - pH-Sensor-Implementierung
- **sensor/flow/** - Implementierungen für Durchflusssensoren
- **sensor/radar/** - Implementierungen für Radar-Füllstandsensoren
- **sensor/turbidity/** - Implementierungen für Trübungssensoren

#### 5.3 Aktortypen (`internal/device/actuator/`)
Jeder Aktortyp hat seine eigene Implementierung:
- **actuator/relay/** - Implementierungen für Relais
- **actuator/valve/** - Implementierungen für Ventile

### 6. Controller (`internal/controller/`)
- Enthält die Steuerungslogik für verschiedene Teilsysteme
- **flow/** - Steuerung der Durchflussregelung
- **ph/** - Steuerung der pH-Wert-Regelung
- **system/** - Übergeordnete Systemsteuerung

### 7. Hardware-Abstraktion (`internal/hardware/`)
- **gpio/** - Abstraktion für GPIO-Zugriff
  - **gpio.go** - Definiert die zentrale GPIO-Schnittstelle und Typen
  - **factory.go** - Factory für verschiedene GPIO-Implementierungen basierend auf Plattform
  - **linux_sysfs/** - Linux sysfs-basierte GPIO-Implementierung
  - **linux_gpiod/** - Linux libgpiod-basierte GPIO-Implementierung (moderne Methode)
  - **mock/** - Mock-Implementierung für Tests
  - **config.go** - GPIO-Konfigurationsstrukturen
  - **manager.go** - Zentraler GPIO-Manager zur Verwaltung mehrerer Pins
  - **event.go** - Event-basiertes System für Pin-Statusänderungen
- **uart/** - Abstraktion für UART-Kommunikation

### 8. Plattform-Integration (`internal/integration/`)
- **thingsboard/mqtt/** - MQTT-Kommunikation mit ThingsBoard
- **thingsboard/rest/** - REST-API-Kommunikation mit ThingsBoard (zukünftig)
- Die Integration-Schicht bietet eine Abstraktion für die Kommunikation mit externen Plattformen

#### 8.1 ThingsBoard MQTT Integration (`internal/integration/thingsboard/mqtt/`)
Vollständig modulare Implementierung aller ThingsBoard MQTT-APIs:
- **types.go** - Gemeinsame Typen und Strukturdefinitionen
- **client.go** - Basis-Client mit Kern-Funktionalität (Verbindung, Start/Stop)
- **telemetry.go** - Senden von Telemetriedaten
- **attributes.go** - Verwaltung von Client- und Shared-Attributen
- **rpc.go** - Serverseitige und Clientseitige RPC-Funktionen
- **firmware.go** - API für Firmware-Updates
- **provisioning.go** - Funktionen für Device Provisioning und Claiming
- **utils.go** - Hilfsfunktionen

### 9. Service-Schicht (`internal/service/`)
- **device_service.go** - Service zur Verbindung aller Komponenten
- **monitoring/** - Dienste zur Systemüberwachung
- **scheduler/** - Zeitplaner für wiederkehrende Aufgaben

Die Service-Schicht dient als Bindeglied zwischen der Konfiguration, den Factories und den tatsächlichen Geräten. Sie ermöglicht es, die verschiedenen Komponenten des Systems lose zu koppeln und vermeidet so zirkuläre Abhängigkeiten.

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

## GPIO-Implementierung

Die GPIO-Implementierung bietet eine abstrakte Schnittstelle für den Zugriff auf Hardware-Pins und unterstützt verschiedene Plattformen und Nutzungsszenarien:

```
+-------------------+            +----------------------+
| Controller-Logik  |            | Hardware-Abstraction |
| (Aktoren, LEDs)   +----------->| (GPIO Interface)     |
+-------------------+            +----------+-----------+
                                            |
                                 +----------v-----------+
                                 |    GPIO Factory       |
                                 +----------+-----------+
                                            |
            +------------------------+------+-------+------------------------+
            |                        |              |                        |
+-----------v-----------+ +----------v-----------+  |           +-----------v-----------+
| Linux sysfs GPIO      | | Linux gpiod GPIO     |  |           | Mock GPIO (Tests)     |
| (Legacy-Unterstützung)| | (Moderne Linux)      |  |           | (Für Testzwecke)      |
+-----------------------+ +----------------------+  |           +-----------------------+
                                                    |
                                          +---------v----------+
                                          | Andere Plattformen |
                                          | (erweiterbar)      |
                                          +--------------------+
```

### GPIO-Interface-Design

Das Kerninterface für GPIO-Operationen definiert die folgenden grundlegenden Operationen:

```go
// PinMode definiert die möglichen Modi eines GPIO-Pins
type PinMode int

const (
    Input PinMode = iota
    Output
    InputPullUp
    InputPullDown
)

// Edge definiert Trigger-Events für GPIO-Pins
type Edge int

const (
    None Edge = iota
    Rising
    Falling
    Both
)

// Pin repräsentiert einen einzelnen GPIO-Pin
type Pin interface {
    // Grundlegende Operationen
    Read() (bool, error)              // Liest den aktuellen Zustand des Pins
    Write(value bool) error           // Setzt den Pin auf high (true) oder low (false)
    SetMode(mode PinMode) error       // Ändert den Modus des Pins
    SetEdge(edge Edge) error          // Ändert die Trigger-Bedingung für Events
    Close() error                     // Gibt Ressourcen frei
    
    // Eigenschaften
    Number() int                      // Gibt die Pin-Nummer zurück
    Name() string                     // Gibt den Namen des Pins zurück
    Mode() PinMode                    // Gibt den aktuellen Modus zurück
    
    // Event-basierte Funktionen
    RegisterCallback(func(bool)) error  // Registriert Callback für Statusänderungen
    EnableInterrupt() error           // Aktiviert Interrupts
    DisableInterrupt() error          // Deaktiviert Interrupts
}

// Manager ist verantwortlich für die Verwaltung mehrerer GPIO-Pins
type Manager interface {
    // Pin-Verwaltung
    OpenPin(pinNumber int, mode PinMode) (Pin, error)
    ClosePin(pinNumber int) error
    CloseAll() error
    
    // Pin-Lookup
    GetPin(pinNumber int) (Pin, bool)
    GetPinByName(name string) (Pin, bool)
    
    // Konfiguration
    LoadConfig(configFile string) error
    AddPinAlias(number int, name string) error
}
```

### Plattformspezifische Implementierungen

#### Linux sysfs GPIO

Diese Implementierung nutzt das traditionelle sysfs-Interface (`/sys/class/gpio/`) für den Zugriff auf GPIO-Pins. Sie ist kompatibel mit älteren Linux-Systemen, wird aber in neueren Kerneln als veraltet angesehen.

Features:
- Kompatibilität mit älteren Linux-Versionen
- Einfaches Polling von Pin-Zuständen
- Interrupt-Unterstützung über epoll

#### Linux gpiod GPIO

Diese Implementierung nutzt die moderne libgpiod-Bibliothek für den Zugriff auf GPIO-Pins. Sie ist die empfohlene Methode für neuere Linux-Systeme (Kernel 4.8+) und bietet verbesserte Funktionalität und Leistung.

Features:
- Thread-safe Zugriff auf GPIO-Pins
- Effiziente Event-Handling-Mechanismen
- Konsumententracking und Pin-Eigentümerschaft
- Support für Bulk-Operationen
- Verbesserte Leistung gegenüber sysfs

#### Mock GPIO

Eine simulierte GPIO-Implementierung für Testzwecke:
- Erlaubt Tests ohne physische Hardware
- Simuliert Pin-Zustände und Ereignisse
- Beinhaltet Werkzeuge zum Testen von Timing und Edge-Erkennung

### GPIO-Konfiguration

Die GPIO-Pins werden über eine JSON-Konfigurationsdatei definiert:

```json
{
  "pins": [
    {
      "number": 17,
      "name": "pump_relay",
      "mode": "output",
      "initial_state": false,
      "description": "Relais für die Hauptpumpe"
    },
    {
      "number": 18,
      "name": "co2_relay",
      "mode": "output",
      "initial_state": false,
      "description": "Relais für das CO2-Ventil"
    },
    {
      "number": 27,
      "name": "status_led_red",
      "mode": "output",
      "initial_state": false,
      "description": "Rote Status-LED"
    },
    {
      "number": 9,
      "name": "power_button",
      "mode": "input_pullup",
      "edge": "falling",
      "debounce_ms": 50,
      "description": "Power/Mode-Knopf"
    }
  ]
}
```

### Integration mit Geräten und Controllern

Die GPIO-Implementierung bildet die Grundlage für:

1. **Relais-Steuerung**:
   - Wasseraufbereitungssystem-Steuerung
   - CO2-Ventilsteuerung
   - Heizungssteuerung

2. **Statusanzeige**:
   - RGB-LED zur Systemstatusanzeige
   - Fehlerzustände und Betriebsmodi

3. **Benutzereingabe**:
   - Knöpfe für manuelle Steuerung
   - Modus-Auswahl und Reset-Funktionen

4. **Watchdog-Funktionalität**:
   - Überwachung des Systemzustands
   - Automatischer Neustart bei Hängern

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

## Stand der Migration

Die Migration von der alten zur neuen Architektur ist in folgenden Bereichen abgeschlossen:

1. **Zentrale Typen und Interfaces:**
   - Vollständige Implementierung der Basisinterfaces in `internal/types/`
   - Interface-Definitionen für Geräte, Sensoren und Protokolle

2. **Basiskomponenten:**
   - Geräteregistry und Factory in `internal/device/`
   - Sensor-Basis in `internal/device/sensor/base.go`
   - Protokoll-Handler-Interface und Modbus-Implementierung

3. **Sensor-Typen:**
   - Migration aller Sensor-Typen abgeschlossen (pH, Flow, Radar, Turbidity)
   - Spezifische Implementierungen in eigenen Paketen
   - Factory-Funktionen für alle Sensortypen

4. **Konfiguration:**
   - Ladeprozess für Konfigurationen in der Service-Schicht

5. **Modbus-Protokoll:**
   - Neue implementierung `internal/protocol/modbus/` ersetzt die alte vollständig
   - Vollständig konfigurierbar über JSON
   - Unterstützt alle benötigten Register-Typen und Datenformate

6. **Hauptanwendung:**
   - Integration der neuen Architektur in die Hauptanwendung abgeschlossen
   - Die neue Architektur mit dem SensorAdapter ist in der Hauptanwendung (`cmd/main.go`) implementiert
   - Die alten Implementierungen wurden entfernt

7. **Projektstruktur:**
   - Die Projektstruktur wurde gemäß dem neuen modularen Design reorganisiert
   - Klare Trennung zwischen Geräten, Protokollen, Controllern und Integrationen
   - ThingsBoard MQTT Client in `internal/integration/thingsboard/mqtt/` platziert

Folgende Schritte sind noch offen:

1. **ThingsBoard-Integration:**
   - Weitere Anpassungen der ThingsBoard-Integration an die neue Architektur
   - Optimierung der Zuordnung von Messwerten zu Telemetriedaten

2. **Aktoren-Integration:**
   - Implementierung und Integration von Aktorgeräten 