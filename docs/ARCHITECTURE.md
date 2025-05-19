# owipexRS485GO - Systemarchitektur

## Übersicht

Das owipexRS485GO-System ist eine in Go implementierte Kommunikationsbrücke zwischen Wasseraufbereitungssensoren (über RS485/Modbus) und der ThingsBoard IoT-Plattform. Die Anwendung wurde von einem Python-Projekt portiert, um die Stabilität, Leistung und Zuverlässigkeit zu verbessern.

```
+----------------+        +-------------------+        +-----------------+        +-----------------+
| Sensoren       |        | owipexRS485GO     |        | Redis Cache     |        | ThingsBoard     |
| (RS485/Modbus) | <----> | (Go Application)  | <----> | (Datenpuffer)   | <----> | (IoT Platform)  |
+----------------+        +-------------------+        +-----------------+        +-----------------+
                                    ^                          ^
                                    |                          |
                          +---------+---------+       +--------+----------+
                          | DrmagicE/gmqtt    |<----->| Persistenzmechanis-|
                          | (MQTT-Broker)     |       | mus für Messaging  |
                          +-------------------+       +-------------------+
```

Die Anwendung verwendet moderne Technologien und Architekturansätze:

- **Modulare Go-Architektur**: Klar getrennte Verantwortlichkeiten und lose Kopplung zwischen Komponenten
- **Eingebetteter MQTT-Broker**: DrmagicE/gmqtt für interne Kommunikation zwischen Modulen
- **Redis-Integration**: Für Datenpersistenz, Caching und zuverlässige Nachrichten-Warteschlangen
- **Abstraktion der Hardware**: Klare Trennung zwischen Kommunikationsprotokollen und Gerätelogik
- **IoT-Integration**: Flexible Anbindung an ThingsBoard über MQTT

Diese Architektur sorgt für:

- **Zuverlässigkeit**: Daten werden in Redis gepuffert, bevor sie an ThingsBoard gesendet werden
- **Skalierbarkeit**: Modulare Struktur erlaubt einfache Erweiterung um neue Sensoren und Funktionen
- **Wartbarkeit**: Klare Trennung der Verantwortlichkeiten und gut definierte Schnittstellen
- **Resilienz**: Robust gegenüber Netzwerkausfällen und Systemneustarts durch Datenpersistenz

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
│   ├── messaging/            # Messaging-Infrastruktur
│   │   ├── mqtt/             # MQTT-spezifische Implementierung
│   │   │   ├── api/          # Gemeinsame Interfaces und Typen
│   │   │   ├── broker/       # MQTT-Broker-Implementierung mit gmqtt
│   │   │   ├── client/       # MQTT-Client-Implementierung
│   │   │   ├── protocol/     # Protokoll-spezifische Abstraktionen
│   │   │   └── topic/        # Topic-Management
│   │   └── common/           # Gemeinsame Messaging-Funktionen
│   │
│   ├── protocol/             # Kommunikationsprotokolle
│   │   ├── factory/          # Factory für Protokoll-Handler
│   │   └── modbus/           # Modbus-Implementierung
│   │
│   ├── service/              # Anwendungsdienste
│   │   ├── monitoring/       # Überwachungsdienste
│   │   ├── scheduler/        # Zeitplanungsdienste
│   │   └── adapter/          # Adapter für verschiedene Schnittstellen
│   │
│   ├── storage/              # Datenspeicherung
│   │   ├── redis/            # Redis-Integration
│   │   │   ├── client/       # Redis-Client und Connection-Pool
│   │   │   ├── repository/   # Daten-Repositories für verschiedene Anwendungsfälle
│   │   │   ├── timeseries/   # Zeitreihen-Implementierung
│   │   │   └── queue/        # Warteschlangen-Implementierung
│   │   └── local/            # Lokale Datenspeicherung
│   │
│   └── types/                # Gemeinsame Typen und Interfaces
│
├── pkg/                      # Potenziell wiederverwendbare Pakete
│   ├── mqtt/                 # Wiederverwendbare MQTT-Komponenten
│   │   ├── client/           # Generischer MQTT-Client
│   │   └── utils/            # MQTT-Hilfsfunktionen
│   └── redis/                # Wiederverwendbare Redis-Komponenten
│
└── scripts/                  # Hilfsskripte für Entwicklung, Deployment, etc.
    ├── startup/              # Startup-Skripte
    └── testing/              # Test-Skripte
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
   - Messwerte werden für die Übertragung aufbereitet

2. **Zwischenspeicherung in MQTT und Redis:**
   - Alle erfassten Daten werden zunächst an den internen MQTT-Broker gesendet
   - Redis speichert automatisch alle relevanten MQTT-Nachrichten
   - Die Daten werden in den entsprechenden Redis-Datenstrukturen organisiert:
     - Aktuelle Werte in Hashes für schnellen Zugriff
     - Zeitreihendaten in RedisTimeSeries für historische Analysen
     - Ungesendete Nachrichten in Warteschlangen (Redis Streams/Lists)

3. **Datenübertragung an ThingsBoard:**
   - Daten werden synchron oder asynchron aus Redis an ThingsBoard gesendet
   - Bei erfolgreicher Übertragung werden die Einträge aus der Warteschlange entfernt
   - Bei Verbindungsverlust werden neue Daten weiterhin in Redis gepuffert
   - Nach Wiederherstellung der Verbindung werden gepufferte Daten in konfigurierbarer Reihenfolge gesendet

4. **Empfang von Konfigurationsänderungen:**
   - ThingsBoard sendet Konfigurationsänderungen als Shared Attributes
   - Der ThingsBoard-Client empfängt diese Änderungen und leitet sie weiter
   - Änderungen werden zuerst in Redis gespeichert und dann über MQTT verteilt
   - Der Konfigurations-Manager aktualisiert die Konfiguration entsprechend
   - Die Änderungen werden an die betroffenen Komponenten weitergegeben

5. **RPC-Befehle für Fernsteuerung:**
   - ThingsBoard sendet RPC-Befehle für Aktionen (z.B. Kalibrierung)
   - Die Befehle werden in Redis gespeichert und über MQTT an zuständige Module weitergeleitet
   - Aktionen werden ausgeführt und Ergebnisse zurück an ThingsBoard gemeldet
   - Der Ausführungsstatus wird in Redis protokolliert

6. **Datensynchronisation und -wiederherstellung:**
   - Bei Systemstart werden Konfigurationen und Zustände aus Redis wiederhergestellt
   - MQTT-Topics mit "retained"-Flag werden aus Redis rekonstruiert
   - Ungesendete Nachrichten in Warteschlangen werden identifiziert und verarbeitet

## Interne Kommunikation mit MQTT

### Architektur der internen Kommunikation

Für die interne Kommunikation zwischen den Modulen wird ein eingebetteter MQTT-Broker verwendet. Dies ermöglicht eine lose Kopplung der Komponenten und eine einheitliche Kommunikationsschicht.

```
+----------------+        +------------------------+        +-----------------+
| Sensoren       |        | owipexRS485GO          |        | ThingsBoard     |
| (RS485/Modbus) | <----> | +-----------------+    | <----> | (IoT Platform)  |
+----------------+        | | Eingebetteter    |    |        +-----------------+
                          | | MQTT-Broker      |    |
                          | | (gmqtt)          |    |
                          | | (gmqtt)          |    |
                          | +-----------------+    |
                          |        ^               |
                          |        |               |
                          |    +---+---+           |
                          |    |       |           |
             +------------+----v-+   +-v-----------+----+
             | Module          |   | Controller        |
             | - Sensoren      |   | - Flow            |
             | - Aktoren       |   | - pH              |
             | - GPIO          |   | - System          |
             | - Dienste       |   | - Watchdog        |
             +-----------------+   +--------------------+
```

### Modulare MQTT-Broker-Implementierung mit DrmagicE/gmqtt

#### Komponentenstruktur

Die MQTT-Implementierung folgt einer sauberen modularen Struktur, die in mehreren Schichten organisiert ist:

```
internal/
  ├── messaging/                  # Übergeordnetes Messaging-Paket
  │   ├── mqtt/                   # MQTT-spezifische Implementierung
  │   │   ├── broker/             # MQTT-Broker-Implementierung mit gmqtt
  │   │   │   ├── config.go       # Broker-Konfiguration
  │   │   │   ├── broker.go       # Haupt-Broker-Implementation
  │   │   │   ├── auth.go         # Authentifizierungsplugin
  │   │   │   ├── persistence.go  # Redis-basierte Persistenz
  │   │   │   ├── hooks.go        # Event-Hooks für Broker-Ereignisse
  │   │   │   ├── metrics.go      # Prometheus-Metrics
  │   │   │   └── admin.go        # Admin-API für Broker-Verwaltung
  │   │   │
  │   │   ├── client/             # MQTT-Client-Implementierung
  │   │   │   ├── client.go       # Basis-Client-Implementation
  │   │   │   ├── config.go       # Client-Konfiguration
  │   │   │   ├── pool.go         # Client-Pool für effiziente Nutzung
  │   │   │   └── subscriber.go   # Hochniveau-Subscriber-API
  │   │   │
  │   │   ├── protocol/           # Protokoll-spezifische Abstraktionen
  │   │   │   ├── packets.go      # MQTT-Paket-Definitionen
  │   │   │   └── versions.go     # Versionsunterstützung (3.1.1, 5.0)
  │   │   │
  │   │   ├── topic/              # Topic-Management
  │   │   │   ├── tree.go         # Topic-Baum-Implementierung
  │   │   │   ├── matcher.go      # Topic-Matching-Logik
  │   │   │   └── registry.go     # Topic-Registry für systemweite Standardtopics
  │   │   │
  │   │   └── api/                # Gemeinsame Interfaces und Typen
  │   │       ├── broker.go       # Broker-Interface
  │   │       ├── client.go       # Client-Interface
  │   │       ├── message.go      # Nachrichtentypen
  │   │       └── types.go        # Gemeinsame Typdefinitionen
  │   │
  │   └── common/                 # Gemeinsame Messaging-Funktionen
  │       ├── serialization.go    # Serialisierungshilfen
  │       └── validation.go       # Validierungsfunktionen
```

#### Kernkomponenten der MQTT-Implementierung

1. **Broker-Modul (`internal/messaging/mqtt/broker/`)**:
   Diese Komponente kapselt die vollständige DrmagicE/gmqtt-Integration und bietet Folgendes:
   - Vollständig konfigurierbare Broker-Instanz
   - Erweiterbare Plugin-Architektur
   - Integration mit Redis für Persistenz
   - Authentifizierungs- und Autorisierungsmechanismen
   - Prometheus-Metriken für Überwachung
   - Administrationsschnittstelle

   ```go
   // Beispiel-Interface für den Broker
   type Broker interface {
       // Lebenszyklus-Methoden
       Start() error
       Stop() error
       
       // Konfiguration
       SetConfig(config BrokerConfig) error
       GetConfig() BrokerConfig
       
       // Plugin-Management
       RegisterPlugin(plugin Plugin) error
       GetPlugin(name string) (Plugin, bool)
       
       // Monitoring und Status
       Stats() BrokerStats
       Health() HealthStatus
       
       // Event-Hooks
       OnConnect(hook ConnectHook)
       OnPublish(hook PublishHook)
       OnSubscribe(hook SubscribeHook)
       OnUnsubscribe(hook UnsubscribeHook)
       OnDisconnect(hook DisconnectHook)
   }
   ```

2. **Client-Modul (`internal/messaging/mqtt/client/`)**:
   Eine modulare Client-Implementierung, die auf den Broker zugreift:
   - Unterstützung für QoS 0, 1 und 2
   - Automatische Wiederverbindung
   - Thread-sichere Implementierung
   - Konfigurierbare Timeouts und Wiederholungsstrategien
   - Connection-Pooling für effiziente Ressourcennutzung

   ```go
   // Beispiel-Interface für den Client
   type Client interface {
       // Verbindung
       Connect() error
       Disconnect() error
       IsConnected() bool
       
       // Basisfunktionen
       Publish(topic string, payload []byte, qos byte, retained bool) error
       Subscribe(topic string, qos byte) (<-chan Message, error)
       Unsubscribe(topic string) error
       
       // Erweiterte Funktionen
       PublishWithContext(ctx context.Context, topic string, payload []byte, qos byte, retained bool) error
       Request(requestTopic, responseTopic string, payload []byte, timeout time.Duration) (Message, error)
       
       // Event-Callbacks
       OnConnectionLost(callback func(error))
       OnReconnect(callback func())
   }
   ```

3. **Topic-Management (`internal/messaging/mqtt/topic/`)**:
   Effiziente Verwaltung von MQTT-Topics:
   - Optimierte Topic-Baum-Struktur für schnelles Matching
   - Unterstützung für Wildcards (+ und #)
   - Fein abgestimmte Topic-Autorisierung
   - Standardisierte Topic-Struktur für das System

#### Integration mit DrmagicE/gmqtt

Die Broker-Implementierung nutzt DrmagicE/gmqtt als Basis und erweitert sie um projektspezifische Funktionen:

```go
// Beispiel für die Broker-Implementation mit gmqtt
type GmqttBroker struct {
    server *gmqtt.Server
    config BrokerConfig
    hooks  map[string]interface{}
    plugins map[string]Plugin
    metrics *BrokerMetrics
    // Weitere interne Felder
}

func NewBroker(config BrokerConfig) (Broker, error) {
    // Erstelle eine neue Broker-Instanz
    broker := &GmqttBroker{
        config:  config,
        hooks:   make(map[string]interface{}),
        plugins: make(map[string]Plugin),
        metrics: NewBrokerMetrics(),
    }
    
    // Konfiguriere gmqtt
    server := gmqtt.NewServer(
        gmqtt.WithTCPListener(config.Host, config.Port),
        gmqtt.WithHook(broker.buildHooks()),
    )
    
    // Konfiguriere Plugins, falls vorhanden
    if config.EnableAuth {
        server.AddPlugin(gmqtt.NewAuthPlugin(broker.authHandler))
    }
    
    // Aktiviere Persistenz mit Redis, falls konfiguriert
    if config.EnablePersistence {
        redisConfig := gmqtt.NewRedisConfig()
        redisConfig.Addr = config.RedisAddr
        server.AddPlugin(gmqtt.NewRedisPersistence(redisConfig))
    }
    
    broker.server = server
    return broker, nil
}

func (b *GmqttBroker) Start() error {
    return b.server.Run()
}

func (b *GmqttBroker) Stop() error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    return b.server.Shutdown(ctx)
}

// Weitere Implementierungen der Broker-Methoden...
```

#### Verwendung von Redis für MQTT-Persistenz

Die MQTT-Broker-Implementierung nutzt Redis für verschiedene Persistenzfunktionen:

1. **Retained Messages**:
   - Alle retained Messages werden in Redis gespeichert
   - Schlüsselformat: `mqtt:retained:<topic>`
   - Automatische Wiederherstellung bei Broker-Neustart

2. **Session-Persistenz**:
   - MQTT-Client-Sessions werden in Redis persistiert
   - Unterstützung für Clean Sessions und Persistent Sessions
   - Sitzungswiedererkennung für verbindungsübergreifende Zustellungsgarantien

3. **QoS-Management**:
   - QoS 1 und QoS 2 Nachrichten-Flows werden in Redis gespeichert
   - Sicherstellung der Zustellungsgarantien auch bei Broker-Neustarts
   - Vermeidung von Nachrichtenduplikaten über Redis-basierte Message-IDs

4. **Subscription-Persistenz**:
   - Aktive Abonnements werden in Redis gespeichert
   - Automatische Wiederherstellung von Abonnements bei Client-Wiederverbindung

### Topic-Struktur

Die interne Kommunikation verwendet eine klare Topic-Struktur:

```
owipex/
  ├── sensors/
  │   └── {sensor_id}/
  │       ├── value               # Aktuelle Messwerte (QoS 0)
  │       ├── config              # Konfiguration (retained)
  │       └── status              # Status (retained)
  │
  ├── actuators/
  │   └── {actuator_id}/
  │       ├── command             # Steuerungskommandos (QoS 1)
  │       └── status              # Aktueller Status (retained)
  │
  ├── controllers/
  │   └── {controller_id}/
  │       ├── command             # Controller-Kommandos (QoS 1)
  │       ├── status              # Status (retained)
  │       └── parameters          # Parameter (retained)
  │
  ├── system/
  │   ├── status                  # Systemstatus (retained)
  │   ├── log                     # Log-Nachrichten
  │   └── commands                # Systemkommandos (QoS 1)
  │
  └── thingsboard/
      ├── attribute_updates       # Weiterleitung von ThingsBoard-Attributen
      └── rpc_requests            # Weiterleitung von RPC-Anfragen
```

### Kommunikationspattern

Die Module nutzen die folgenden Kommunikationsmuster:

1. **Publish/Subscribe**:
   - Module veröffentlichen Daten und abonnieren relevante Topics
   - Keine direkte Abhängigkeit zwischen Modulen
   
2. **Request/Response**:
   - Implementiert über Topic-Paare (command/response)
   - Eindeutige Korrelations-IDs für Anfrage-Zuordnung

3. **Retained Messages**:
   - Speicherung des letzten Zustands für wichtige Werte
   - Automatische Wiederherstellung bei Neustart

4. **QoS-Level**:
   - QoS 0 für Telemetriedaten (höchste Performance)
   - QoS 1 für wichtige Steuerungsbefehle (garantierte Zustellung)
   - QoS 2 für kritische Systembefehle (genau einmalige Zustellung)

### Sicherheits- und Leistungsaspekte

1. **Authentifizierung und Autorisierung**:
   - Plugin-basierte Authentifizierung für Clients
   - Topic-basierte Zugriffssteuerung mit ACLs
   - Redis-basierte Credentials-Speicherung

2. **Leistungsoptimierung**:
   - Konfigurierbare Nachrichtenpuffer
   - Optimierte Topic-Matching-Algorithmen
   - Benchmark-getriebene Konfiguration

3. **Überwachung**:
   - Prometheus-Metriken für Broker-Leistung
   - Detaillierte Logging für Diagnose
   - Health-Checks für Systemüberwachung

### Integration in die Anwendung

Die MQTT-Broker-Komponente wird im Hauptprogramm folgendermaßen eingebunden:

```go
// Pseudocode für die Integration in main.go
func main() {
    // Lade Konfiguration
    config, err := loadConfig()
    if err != nil {
        log.Fatalf("Fehler beim Laden der Konfiguration: %v", err)
    }
    
    // Initialisiere Redis
    redisClient, err := storage.NewRedisClient(config.Redis)
    if err != nil {
        log.Fatalf("Fehler bei der Redis-Initialisierung: %v", err)
    }
    
    // Initialisiere den MQTT-Broker
    brokerConfig := messaging.BrokerConfig{
        Host:             config.MQTT.Host,
        Port:             config.MQTT.Port,
        EnableAuth:       config.MQTT.EnableAuth,
        EnablePersistence: true,
        RedisAddr:        config.Redis.Addr,
    }
    
    broker, err := messaging.NewBroker(brokerConfig)
    if err != nil {
        log.Fatalf("Fehler bei der MQTT-Broker-Initialisierung: %v", err)
    }
    
    // Starte den Broker
    err = broker.Start()
    if err != nil {
        log.Fatalf("Fehler beim Starten des MQTT-Brokers: %v", err)
    }
    
    // Initialisiere Clients für Module
    sensorClient := messaging.NewClient(messaging.ClientConfig{
        ClientID: "sensor-manager",
        Broker:   "localhost",
        Port:     1883,
    })
    
    // Verbinde den Client
    err = sensorClient.Connect()
    if err != nil {
        log.Fatalf("Fehler bei der MQTT-Client-Verbindung: %v", err)
    }
    
    // Cleanup bei Beendigung
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    <-c
    
    // Stoppe den Broker
    broker.Stop()
}
```

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

## Redis-Integration für Datenpersistenz

Die Anwendung nutzt Redis als robuste Zwischenspeicherlösung für Sensordaten, um eine zuverlässige Datenpufferung und -persistenz zu gewährleisten:

```
+----------------+        +-------------------+        +-----------------+        +-----------------+
| Sensoren       |        | owipexRS485GO     |        | Redis Cache     |        | ThingsBoard     |
| (RS485/Modbus) | <----> | (Go Application)  | <----> | (Datenpuffer)   | <----> | (IoT Platform)  |
+----------------+        +-------------------+        +-----------------+        +-----------------+
                                    ^                          ^
                                    |                          |
                          +---------+---------+       +--------+----------+
                          | DrmagicE/gmqtt    |<----->| Persistenzmechanis-|
                          | (MQTT-Broker)     |       | mus für Messaging  |
                          +-------------------+       +-------------------+
```

### Schlüsselkomponenten der Redis-Integration

1. **Redis-Client und Connection-Pool**:
   - Implementiert in `internal/storage/redis/`
   - Verwaltet Verbindungen zu Redis-Server
   - Konfigurierbare Verbindungsparameter (Host, Port, Passwort, DB)
   - Unterstützung für TLS/SSL-Verbindungen

2. **Datenpersistenz-Layer**:
   - Mehrere Abstraktionsebenen je nach Anwendungsfall:
     - `TimeSeriesRepository`: Optimiert für Zeitreihendaten (Telemetrie)
     - `DeviceConfigRepository`: Für Gerätekonfigurationen
     - `SystemStateRepository`: Für Systemzustandsdaten
   - Unterstützung für automatische Datenablaufzeiten (TTL)

3. **Datenpufferung für Verbindungsausfälle**:
   - Automatische Zwischenspeicherung bei ThingsBoard-Verbindungsverlust
   - Persistente Warteschlange mit konfigurierbarer Kapazität
   - Priorisierung von Daten basierend auf Alter und Wichtigkeit

4. **Redis-MQTT-Brücke**:
   - Bidirektionale Synchronisation zwischen MQTT-Topics und Redis
   - Automatische Persistierung wichtiger MQTT-Nachrichten
   - Wiederherstellung von MQTT-Zuständen aus Redis bei Systemstart

### Redis-Datenmodell

Das Redis-Datenmodell ist strukturiert nach Gerätetyp und Funktionalität:

```
// Telemetriedaten mit Redis TimeSeries
ts:<device_id>:<metric_name> -> Zeitreihendaten für Metriken

// Gerätekonfigurationen
config:<device_id> -> Hash mit Konfigurationsparametern

// Systemzustand
state:system -> Hash mit aktuellem Systemzustand
state:devices -> Hash mit Status aller Geräte

// Warteschlangen für ungesendete Daten
queue:thingsboard:telemetry -> Sorted Set mit Zeitstempel
queue:thingsboard:attributes -> Sorted Set mit Zeitstempel

// Backup von MQTT retained messages
mqtt:retained:<topic> -> Letzte persistente MQTT-Nachricht
```

### Datenhaltungsstrategien

1. **Zeitreihendaten**:
   - Verwendung von RedisTimeSeries für effiziente Speicherung von Messwerten
   - Automatische Downsampling-Strategien für langfristige Datenspeicherung
   - Komprimierung von älteren Daten zur Speicherplatzoptimierung

2. **Snapshot-Management**:
   - Periodische Snapshots des Systemzustands
   - Recovery-Mechanismus bei Systemneustarts
   - Differential-Backups zur Minimierung des Speicherbedarfs

3. **Rotationsstrategien**:
   - Automatische Datenlöschung nach konfigurierbarer Zeit
   - Priorisierte Speicherplatzfreigabe bei knappem Speicher
   - Komprimierte Archivierung wichtiger historischer Daten

### Integration mit vorhandenem MQTT-Broker

Die Redis-Integration erweitert die bestehende MQTT-Architektur:

1. **Persistente MQTT-Nachrichten**:
   - Alle retained MQTT-Nachrichten werden in Redis gesichert
   - Automatische Wiederherstellung bei Broker-Neustart
   - Synchronisierung zwischen mehreren MQTT-Broker-Instanzen

2. **Message-Queue für unzustellbare Nachrichten**:
   - Verwendung von Redis Stream als FIFO-Queue
   - Persistente Speicherung von Nachrichten, die nicht zugestellt werden konnten
   - Automatische Wiederholung nach konfigurierbaren Zeitintervallen

3. **Lastverteilung und Skalierbarkeit**:
   - Redis-basierte Lastverteilung für mehrere Verarbeitungsinstanzen
   - Unterstützung für Clustering bei höheren Datenmengen
   - Redis Sentinel für High-Availability-Konfigurationen

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

## Voraussetzungen und Abhängigkeiten

Das System benötigt die folgenden externen Komponenten:

1. **Redis-Server:**
   - Version 6.0 oder höher empfohlen
   - RedisTimeSeries-Modul für optimale Zeitreihendaten-Speicherung
   - Installation auf den meisten Linux-Systemen:
     ```
     apt-get install redis-server
     ```
   - Installation des RedisTimeSeries-Moduls:
     ```
     git clone https://github.com/RedisTimeSeries/RedisTimeSeries.git
     cd RedisTimeSeries
     make setup
     make build
     ```
   - Docker-basierte Installation (Alternative):
     ```
     docker run -p 6379:6379 redislabs/redistimeseries:latest
     ```
   - Redis-Konfigurationsempfehlungen:
     - Aktivierte AOF-Persistenz für Datensicherheit
     - Angepasste Speichergrenzen je nach Anwendungsfall
     - Optimierte eviction-policy für Zeitreihendaten

2. **DrmagicE/gmqtt:**
   - Eingebetteter MQTT-Broker geschrieben in Go
   - Installation über Go-Modul-System:
     ```
     go get github.com/DrmagicE/gmqtt@v0.9.0
     ```
   - Hauptmerkmale:
     - MQTT 3.1.1 und MQTT 5.0 Unterstützung
     - Plugin-System für Erweiterungen
     - Flexible Persistenz-Optionen (Redis, Memory, SQL)
     - ACL und Authentifizierung
     - QoS 0, 1, 2 Unterstützung
     - Will-Nachricht-Unterstützung
     - Retained-Message-Unterstützung
   - Zusätzliche Plugins:
     - Redis-Plugin für Persistenz
     - Prometheus-Plugin für Monitoring
     - Admin-API-Plugin für Verwaltung
     - Auth-Plugin für Benutzerauthentifizierung

3. **Go-Abhängigkeiten:**
   - Alle Go-Abhängigkeiten werden über die Module in `go.mod` verwaltet
   - Die wichtigsten externen Pakete:
     - `github.com/go-redis/redis/v8` für Redis-Client-Funktionalität
     - `github.com/gomodule/redigo` für Redis-Pool-Management
     - `github.com/DrmagicE/gmqtt` für den eingebetteten MQTT-Broker
     - `github.com/eclipse/paho.mqtt.golang` für MQTT-Client-Funktionalität

4. **System-Abhängigkeiten:**
   - Für seriellen Port-Zugriff: `libudev-dev`
   - Für GPIO-Zugriff auf Raspberry Pi: `libgpiod2`

Die Installationsanweisungen und Konfigurationsbeispiele werden in der `README.md` und in der Dokumentation unter `docs/installation.md` bereitgestellt.

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
- Caching für Sensorwerte implementieren ✓ (durch Redis-Integration umgesetzt)
- Unterstützung für weitere Sensortypen hinzufügen
- Datenpufferung bei ThingsBoard-Verbindungsverlust ✓ (durch Redis-Integration umgesetzt)
- Web-Interface für lokale Konfiguration
- Erweiterungen der Redis- und MQTT-Integration:
  - Implementierung von Redis Cluster für höhere Skalierbarkeit
  - Optimierung der Redis-Datenspeicherstrategien für langfristige Datenhaltung
  - Automatisierte Datenanalyse und Anomalieerkennung auf Redis-Ebene
  - Leistungsoptimierung durch selektives Caching häufig abgefragter Werte
  - Verbesserte Visualisierung der Redis-Daten durch Integration mit Monitoring-Tools
- Integration mit Container-Technologien (Docker, Kubernetes) für vereinfachtes Deployment
- Verbesserung der Sicherheit durch verschlüsselte Redis-Verbindungen und Authentifizierung

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

8. **Redis-Integration:**
   - Implementation der Redis-Client-Komponente abgeschlossen
   - Integration des MQTT-Brokers mit Redis für Persistenz
   - Repository-Schicht für verschiedene Datentypen implementiert
   - Warteschlangensystem für zuverlässige Datenübertragung implementiert
   - Zeitreihen-Speicherung für Telemetriedaten umgesetzt
   - Snapshots und Wiederherstellung für System- und Konfigurationszustände

9. **MQTT-Infrastruktur:**
   - Modulare MQTT-Implementierung basierend auf DrmagicE/gmqtt
   - Saubere Trennung in Broker, Client, Protokoll und Topic-Layer
   - Vollständige Integration mit Redis für Persistenz aller MQTT-Daten
   - Konfigurierbare Broker-Instanz mit Plugin-Unterstützung
   - Thread-sichere Client-Implementierung für alle Module
   - Optimierte Topic-Struktur und Topic-Matching
   - Zustandspersistenz über retained Messages in Redis

Folgende Schritte sind noch offen:

1. **ThingsBoard-Integration:**
   - Weitere Anpassungen der ThingsBoard-Integration an die neue Architektur
   - Optimierung der Zuordnung von Messwerten zu Telemetriedaten

2. **Aktoren-Integration:**
   - Implementierung und Integration von Aktorgeräten 

3. **Redis-Integration Verfeinerung:**
   - Performance-Optimierung und Benchmarking
   - Implementierung fortgeschrittener Datenkomprimierungsstrategien
   - Entwicklung von Tools zur Datenanalyse und -visualisierung

4. **MQTT-Erweiterungen:**
   - Integration weiterer gmqtt-Plugins für erweiterte Funktionalität
   - Entwicklung von spezialisierten MQTT-Clients für spezifische Anwendungsfälle
   - Implementierung von fortschrittlichen Monitoring-Werkzeugen für den MQTT-Broker