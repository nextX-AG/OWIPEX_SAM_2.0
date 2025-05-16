# OWIPEX RS485 Go

Eine Go-Implementierung für die Kommunikation zwischen Wasseraufbereitungssensoren (über RS485/Modbus) und der ThingsBoard IoT-Plattform.

## Übersicht

Das owipexRS485GO-System dient als Kommunikationsbrücke zwischen:
- Sensoren (pH, Durchfluss, Radar, Trübung) über RS485/Modbus
- ThingsBoard IoT-Plattform für Datenvisualisierung und -steuerung

Das Projekt wurde von einer Python-Implementierung zu Go migriert, um Stabilität, Leistung und Zuverlässigkeit zu verbessern.

## Dokumentation

Ausführliche Dokumentation finden Sie im [docs](docs/) Verzeichnis:

- [Systemarchitektur](docs/ARCHITECTURE.md) - Detaillierte Beschreibung der Systemkomponenten
- [Offene Aufgaben](docs/TODO.md) - Aktuelle und geplante Entwicklungen

## Ausführen des Projekts

### Schnellstart

Verwenden Sie eines der Startskripte im `scripts/startup/` Verzeichnis:

```bash
# Standard-Starter (erkennt Betriebssystem und wählt den passenden seriellen Port)
./scripts/startup/start_modbus_reader.sh

# Spezifisch für Linux ttyS0 Port
./scripts/startup/start_owipex_ttyS0.sh

# Python-Version des Starters (mit Abhängigkeitsinstallation)
python3 ./scripts/startup/start_modbus_reader.py
```

## Projektstruktur

Das Projekt folgt einer modularen Struktur mit klarer Trennung der Zuständigkeiten:

- `cmd/` - Ausführbare Anwendungen
- `internal/` - Nicht-öffentlicher Code
- `scripts/` - Hilfsskripte
- `docs/` - Projektdokumentation

## Entwicklung

Zur Weiterentwicklung empfehlen wir die Verwendung von Go 1.19 oder höher.