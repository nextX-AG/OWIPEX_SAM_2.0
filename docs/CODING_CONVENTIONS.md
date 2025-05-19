<!-- Version: 0.1.6 | Last Updated: 2025-05-19 14:54:14 UTC -->


# Coding-Konventionen für OWIPEX_SAM_2.0

Dieses Dokument definiert die einheitlichen Coding-Konventionen für das OWIPEX_SAM_2.0-Projekt. Durch die Einhaltung dieser Standards wird die Codebasis konsistenter, lesbarer und wartbarer.

## Namenskonventionen

### Dateinamen

#### Allgemeine Regeln
- Vermeidung generischer Namen in verschiedenen Paketen
- Spezifische, eindeutige Namen, die den Inhalt klar beschreiben
- Kleinbuchstaben mit Unterstrichen zwischen Wörtern (snake_case)
- Gruppierung zusammengehöriger Dateien durch gemeinsame Präfixe

#### Beispiele für gute Dateinamen

| ❌ Schlecht          | ✅ Gut                    | Erklärung                               |
|---------------------|--------------------------|----------------------------------------|
| `factory.go`        | `flow_sensor_factory.go` | Spezifiziert den Typ der Factory       |
| `client.go`         | `modbus_client.go`       | Gibt an, welcher Client implementiert ist |
| `registry.go`       | `device_registry.go`     | Verdeutlicht, was registriert wird     |
| `types.go`          | `protocol_types.go`      | Spezifiziert den Bereich der Typen     |
| `config.go`         | `modbus_config.go`       | Gibt den Konfigurationsbereich an      |

### Typnamen und Strukturen

#### Allgemeine Regeln
- CamelCase für alle Typen, Interfaces und Strukturen
- Präfixe zur Verdeutlichung der Zugehörigkeit
- Konsistente Benennungsmuster in verwandten Typen
- Vermeidung von Abkürzungen (außer allgemein verständliche wie HTTP, JSON)

#### Beispiele für gute Typnamen

| ❌ Schlecht          | ✅ Gut                    | Erklärung                               |
|---------------------|--------------------------|----------------------------------------|
| `PHSensor`          | `PhSensor`               | Konsistentes CamelCase                 |
| `Client`            | `ModbusClient`           | Spezifiziert den Client-Typ            |
| `Config`            | `SensorConfig`           | Verdeutlicht den Konfigurationsbereich |
| `MQTTHandler`       | `MqttHandler`            | Konsistentes CamelCase                 |
| `IOTDevice`         | `IoTDevice`              | Konsistentes CamelCase                 |

### Funktionsnamen

#### Allgemeine Regeln
- CamelCase mit führendem Kleinbuchstaben für Methodennamen
- Verben zur Beschreibung von Aktionen
- Präzise Namen, die den Zweck klar beschreiben
- Konsistente Benennung für ähnliche Operationen

#### Beispiele für gute Funktionsnamen

| ❌ Schlecht          | ✅ Gut                    | Erklärung                               |
|---------------------|--------------------------|----------------------------------------|
| `Register`          | `RegisterSensorType`     | Spezifiziert genau, was registriert wird |
| `Process`           | `ProcessModbusResponse`  | Verdeutlicht, was verarbeitet wird     |
| `Handle`            | `HandleMqttMessage`      | Spezifiziert, was behandelt wird       |
| `Init`              | `InitializeSensorConfig` | Detaillierter und selbsterklärender    |

### Konstanten und Variablen

#### Allgemeine Regeln
- Konstanten: ALL_CAPS mit Unterstrichen
- Globale Variablen: CamelCase mit führendem Großbuchstaben
- Lokale Variablen: camelCase mit führendem Kleinbuchstaben
- Beschreibende Namen, die Zweck und Inhalt verdeutlichen

#### Beispiele für gute Konstanten und Variablen

| ❌ Schlecht          | ✅ Gut                    | Erklärung                               |
|---------------------|--------------------------|----------------------------------------|
| `timeout`           | `DEFAULT_TIMEOUT_MS`     | Klare Konstante mit Einheit            |
| `list`              | `sensorConfigList`       | Spezifiziert Inhalt der Liste          |
| `mgr`               | `deviceManager`          | Vollständiger, klarer Name             |
| `s`                 | `sensor`                 | Aussagekräftiger Name                  |

## Vermeidung von Duplizierung

### Zentrale Typdefinitionen

Gemeinsam verwendete Typen müssen zentral in `internal/types/` definiert werden:

```go
// In internal/types/protocol.go
type ModbusRegisterType string

const (
    // Modbus-Registertypen
    RegisterTypeHolding  ModbusRegisterType = "HOLDING"
    RegisterTypeInput    ModbusRegisterType = "INPUT"
    RegisterTypeCoil     ModbusRegisterType = "COIL"
    RegisterTypeDiscrete ModbusRegisterType = "DISCRETE"
)

// Gemeinsame Strukturen
type ModbusRegisterMap map[string]ModbusRegisterConfig

type ModbusRegisterConfig struct {
    Address      uint16
    RegisterType ModbusRegisterType
    DataType     string
    Scaling      float64
    Unit         string
    Description  string
}
```

### DTO-Muster (Data Transfer Objects)

Datenübertragungsobjekte sollten konsistent benannt werden:

- Suffix `Dto` an Klassenname anhängen
- In Datei mit Suffix `_dto.go` definieren
- Gruppierung nach Funktionsbereich

Beispiel:
```go
// modbus_dto.go
type ModbusCommandDto struct {
    Address     uint16
    RegisterType string
    Value       interface{}
}

type ModbusResponseDto struct {
    Value       interface{}
    Timestamp   time.Time
    Error       error
}
```

## Paketstruktur und -organisation

### Paketpfade und -namen

- Kurze, eindeutige Paketnamen
- Hierarchische Organisation nach Funktionalität
- Vermeidung von zu tiefen Verschachtelungen

### Importstruktur

- Gruppierung von Imports:
  1. Standardbibliotheken
  2. Externe Abhängigkeiten
  3. Interne Pakete
- Keine Punkt-Imports (`.`)
- Keine ungenutzten Importe

Beispiel:
```go
import (
    "context"
    "time"
    
    "github.com/go-redis/redis/v8"
    "github.com/DrmagicE/gmqtt"
    
    "github.com/owipex/OWIPEX_SAM_2.0/internal/types"
    "github.com/owipex/OWIPEX_SAM_2.0/internal/config"
)
```

## Kommentare und Dokumentation

### Paketkommentare

Jede Datei sollte mit einem Paketkommentar beginnen:

```go
// Package modbus implements the Modbus protocol handler for communicating with sensors.
// It supports RTU and TCP modes with configurable parameters and register mapping.
package modbus
```

### Funktions- und Typkommentare

Öffentliche Funktionen und Typen müssen dokumentiert werden:

```go
// NewModbusClient creates a new Modbus client with the given configuration.
// It establishes a connection to the specified device and validates the connection.
// Returns an error if the connection cannot be established.
func NewModbusClient(config ModbusConfig) (*ModbusClient, error) {
    // ...
}
```

### Inline-Kommentare

- Komplexe Logik erklären
- Nicht offensichtliche Entscheidungen dokumentieren
- Workarounds und technische Schulden kennzeichnen

## Fehlerbehandlung

### Fehlerkonventionen

- Fehler immer zurückgeben, nicht loggen und ignorieren
- Fehler mit aussagekräftigen Informationen anreichern
- Kontextuelle Informationen hinzufügen

Beispiel:
```go
if err != nil {
    return fmt.Errorf("failed to read modbus register %d: %w", registerAddress, err)
}
```

### Fehlertypen

Spezifische Fehlertypen für unterschiedliche Fehlerklassen definieren:

```go
type ModbusConnectionError struct {
    Device string
    Cause  error
}

func (e *ModbusConnectionError) Error() string {
    return fmt.Sprintf("failed to connect to modbus device %s: %v", e.Device, e.Cause)
}
```

## Best Practices

### Testkonventionen

- Testdateien mit Suffix `_test.go`
- Testfunktionsnamen, die beschreiben, was getestet wird: `TestModbusClient_ReadHoldingRegister_Success`
- Table-driven Tests für mehrere Testfälle

### Interfacedefinitionen

- Interfaces sollten dort definiert werden, wo sie verwendet werden
- Klein und fokussiert halten
- Dokumentieren, welche Verantwortung sie kapseln

### Logging

- Konsistente Log-Level verwenden
- Strukturiertes Logging mit Kontextinformationen
- Keine sensiblen Daten loggen

## Weiterführende Links

- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Go Style Guide bei Uber](https://github.com/uber-go/guide/blob/master/style.md) 