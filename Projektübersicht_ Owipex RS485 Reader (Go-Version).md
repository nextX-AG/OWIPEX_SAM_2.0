# Projektübersicht: Owipex RS485 Reader (Go-Version)

## Einführung

Dieses Dokument beschreibt die nach Go portierte Version des Owipex RS485 Reader Projekts. Das Hauptziel der Portierung war die Verbesserung der Systemstabilität und der Zuverlässigkeit der RS485/Modbus-Kommunikation im Vergleich zur ursprünglichen Python-Version.

Die Go-Anwendung liest Daten von verschiedenen Sensoren (Trübung, Radar, Durchfluss) über eine serielle RS485-Schnittstelle mittels Modbus-Protokoll aus und sendet die aufbereiteten Daten an einen ThingsBoard MQTT-Broker.

## Projektstruktur

Das Go-Projekt ist modular aufgebaut und befindet sich im Verzeichnis `/home/ubuntu/owipex_project/go_owipex_reader`. Die Struktur ist wie folgt:

-   `/cmd/main.go`: Hauptanwendung, die die verschiedenen Module initialisiert und startet.
-   `/internal/config`: Verantwortlich für das Laden der Konfiguration aus JSON-Dateien und Umgebungsvariablen (`.env`).
-   `/internal/modbus`: Stellt den Modbus-Client für die Kommunikation mit den Sensoren bereit. Verwendet die Bibliothek `github.com/simonvetter/modbus`.
-   `/internal/sensor`: Enthält die Logik für die einzelnen Sensortypen (ph, turbidity, flow, radar). Definiert ein gemeinsames Sensor-Interface und spezifische Implementierungen.
-   `/internal/manager`: Der `SensorManager` koordiniert das Auslesen der Sensoren in den definierten Intervallen und leitet die Daten an den ThingsBoard-Client weiter.
-   `/internal/thingsboard`: Der `ThingsBoardClient` stellt die Verbindung zum MQTT-Broker her und publiziert die Sensordaten im JSON-Format.
-   `/config/sensors.json`: Konfigurationsdatei für die Definition der anzusprechenden Sensoren (Beispieldatei, die tatsächliche Konfiguration wird zur Laufzeit geladen).
-   `go.mod`, `go.sum`: Go-Moduldateien zur Verwaltung der Abhängigkeiten.

## Konfiguration

Die Anwendung verwendet zwei Hauptkonfigurationsquellen:

1.  **Sensor-Konfigurationsdatei**: Eine JSON-Datei (standardmäßig `config/sensors.json` relativ zum Ausführungsverzeichnis oder ein übergebener Pfad), die die Details jedes Sensors definiert (ID, Typ, Modbus-Adresse, etc.). Eine Beispieldatei (`/home/ubuntu/owipex_project/owipex_RS485_reader-development/sensors.json`) aus dem ursprünglichen Projekt kann als Vorlage dienen und muss ggf. in das `config`-Verzeichnis des Go-Projekts kopiert werden.
2.  **Umgebungsvariablen-Datei (`.env`)**: Eine Datei unter `/etc/owipex/go_reader.env` (oder ein über `GO_READER_ENV_PATH` definierter Pfad) enthält die Parameter für die RS485-Schnittstelle und die ThingsBoard-Verbindung. Diese Datei überschreibt Standardwerte oder Werte aus einer optionalen globalen JSON-Konfigurationsdatei.

    Beispielinhalt für `/etc/owipex/go_reader.env`:
    ```env
    RS485_PORT=rtu:///dev/ttyS0
    RS485_BAUDRATE=9600
    RS485_DATABITS=8
    RS485_PARITY=N
    RS485_STOPBITS=1
    RS485_TIMEOUT_MS=1000
    RS485_THINGSBOARD_SERVER=146.4.67.141
    RS485_THINGSBOARD_PORT=1883
    RS485_ACCESS_TOKEN=5Ohlb6ZKO4uNw9O2DHwk
    # Optional: LOG_FILE_PATH=/pfad/zur/logdatei.log
    ```

## Ausführung

1.  **Voraussetzungen**:
    *   Go-Compiler ist installiert (wurde im Rahmen des Projekts installiert).
    *   Die serielle Schnittstelle (z.B. `/dev/ttyS0`) ist vorhanden und für den Benutzer, der die Anwendung ausführt, zugänglich (ggf. `sudo` verwenden oder Benutzer zur `dialout`-Gruppe hinzufügen).
    *   Die Konfigurationsdatei `sensors.json` ist im Verzeichnis `config` (relativ zum Ausführungsverzeichnis `go_owipex_reader`) vorhanden.
    *   Die Datei `/etc/owipex/go_reader.env` mit den korrekten Zugangsdaten und RS485-Parametern existiert.

2.  **Starten der Anwendung**:
    Navigieren Sie in das Projektverzeichnis:
    ```bash
    cd /home/ubuntu/owipex_project/go_owipex_reader
    ```
    Führen Sie die Anwendung aus (ggf. mit `sudo` für den Zugriff auf serielle Ports):
    ```bash
    sudo go run ./cmd/main.go
    ```
    Die Anwendung gibt Log-Meldungen auf der Konsole aus.

## Wichtige Erkenntnisse und Status des Vergleichstests

*   **Go-Version**: Die portierte Go-Anwendung läuft stabil. Sie initialisiert den Modbus-Client korrekt, liest simulierte Sensordaten aus und sendet diese erfolgreich an den konfigurierten ThingsBoard MQTT-Broker. Die Daten werden sowohl im einfachen Format als auch im strukturierten JSON-Format publiziert.
*   **Python-Version**: Die ursprüngliche Python-Version zeigte während der Testphase persistente Probleme. Zunächst fehlten diverse Abhängigkeiten, die schrittweise nachinstalliert wurden. Anschließend traten Probleme mit den Zugriffsrechten auf die serielle Schnittstelle auf, die durch Ausführung mit `sudo` umgangen werden konnten. Jedoch blieben auch nach korrekter Konfiguration der Zugangsdaten (`.envRS485`) Fehler bei der MQTT-Authentifizierung ("Not authorized") und beim Auslesen der Modbus-Geräte (keine oder unvollständige Header-Antworten) bestehen. Ein direkter, stabiler Vergleich der Laufzeitstabilität und Kommunikationszuverlässigkeit war daher nicht abschließend möglich.
*   **Vorteile der Go-Version (erwartet)**: Basierend auf den Zielen der Portierung und den Eigenschaften von Go ist zu erwarten, dass die Go-Version im Langzeitbetrieb eine höhere Stabilität und eine robustere Fehlerbehandlung bei der seriellen Kommunikation aufweist. Das strikte Typsystem und das Concurrency-Modell von Go tragen ebenfalls zur Robustheit bei.

## Nächste Schritte und Empfehlungen

*   **Hardwaretests**: Die aktuelle Go-Version wurde mit simulierten Sensordaten und einer simulierten seriellen Schnittstelle getestet. Umfassende Tests mit der realen Hardware-Umgebung sind unerlässlich, um die Funktionalität unter realen Bedingungen zu validieren.
*   **Fehlerbehandlung und Logging**: Das Logging kann weiter verfeinert werden, um im Fehlerfall detailliertere Diagnosen zu ermöglichen. Die Fehlerbehandlung, insbesondere bei Modbus-Timeouts oder -Fehlern, kann weiter optimiert werden (z.B. durch implementierte Retry-Mechanismen mit Backoff).
*   **Python-Version Debugging**: Falls die Python-Version weiterhin relevant ist, müssten die MQTT-Authorisierungsprobleme und die Modbus-Kommunikationsfehler genauer untersucht werden. Dies könnte eine Überprüfung der ThingsBoard-Gerätekonfiguration, der Token-Gültigkeit oder der spezifischen Modbus-Implementierung in Python erfordern.

## Enthaltene Dateien im ZIP-Archiv

Das bereitgestellte ZIP-Archiv (`go_owipex_reader_final.zip`) enthält den vollständigen Quellcode des Go-Projekts unter `/home/ubuntu/owipex_project/go_owipex_reader`.

