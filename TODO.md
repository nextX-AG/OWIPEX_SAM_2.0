# owipexRS485GO - Task List

## In Bearbeitung
- [x] Minimalen ThingsBoard-Client für Tests implementieren
- [x] Robuste MQTT-Verbindung zu ThingsBoard aufbauen
- [x] Shared Attributes von ThingsBoard empfangen
- [x] Modulare Sensorarchitektur implementieren
- [ ] Migration der bestehenden Sensoren in die neue Architektur

## Hohe Priorität
- [ ] RPC-Befehle von ThingsBoard empfangen und verarbeiten
- [ ] Robustes Error-Handling für Modbus-Verbindungen implementieren
- [ ] Reconnect-Mechanismus für verlorene RS485-Verbindungen
- [ ] Umgang mit Timeouts bei Modbus-Kommunikation
- [ ] Automatische Wiederherstellung der ThingsBoard-Verbindung

## ThingsBoard-Kommunikationsstruktur
- [x] Modulare Basisklasse für ThingsBoard-Client implementieren
- [x] Telemetrie-Modul (SendTelemetry, SendTelemetryWithTs, BatchSendTelemetry)
- [x] Attribute-Modul (PublishAttributes, RequestAttributes, RequestSharedAttributes)
- [x] RPC-Modul (SubscribeToRPC, SendRPCRequest, GetSessionLimits)
- [x] Device-Provisioning-Modul (ClaimDevice, ProvisionDevice)
- [x] Firmware-Update-Modul (RequestFirmwareChunk)
- [x] Hilfsfunktionen für JSON, Topic-Parsing und Request-IDs
- [x] Umfassende Konfigurationsstruktur für alle ThingsBoard-Parameter

## Neue Sensor-/Aktor-Architektur
- [x] Implementierung von Basisinterfaces für Geräte (Device, Sensor, Actor, etc.)
- [x] Factory-Pattern für Geräteerstellung basierend auf Konfiguration
- [x] Trennung von Sensortyp, Kommunikationsprotokoll und Herstellerspezifika
- [x] Konfigurationssystem für herstellerspezifische Sensordetails
- [x] JSON-Schemata für Gerätekonfigurationen
- [x] Implementierung der spezifischen Sensortypen (pH)
- [ ] Implementierung der weiteren Sensortypen (Flow, Radar, Turbidity)
- [ ] Implementierung der spezifischen Aktortypen (Ventil, Pumpe, etc.)
- [x] Protokoll-Abstraktion für verschiedene Kommunikationsarten
- [x] Einheitliches Messwert- und Befehlssystem
- [x] Konverter für Rohdaten zu physikalischen Werten

## Sensor-Migration
- [x] Basis-Sensor-Implementierung erstellen (BaseSensor)
- [x] PH-Sensor in neue Struktur migrieren
- [ ] Flow-Sensor in neue Struktur migrieren
- [ ] Radar-Sensor in neue Struktur migrieren
- [ ] Turbidity-Sensor in neue Struktur migrieren
- [ ] Konfigurationsbeispiele für alle Sensoren erstellen
- [ ] Alte Sensorimplementierungen entfernen

## Sensor-Integration (bisheriger Ansatz)
- [ ] Konfiguration für Sensoren erweitern (enabled/disabled-Flag)
- [ ] Radar-Sensor vollständig integrieren
- [ ] PH-Sensor-Integration abschließen
- [ ] Flow-Sensor-Integration optimieren
- [ ] Turbidity-Sensor-Integration verbessern
- [ ] Kalibrierungsfunktionen für Sensoren implementieren

## Aktor-Integration
- [ ] Schnittstelle für RS485-Aktoren definieren
- [ ] Aktor-Steuerung über Shared Attributes implementieren
- [ ] Aktor-Status-Rückmeldung an ThingsBoard

## ThingsBoard-Integration
- [ ] Verbesserter Umgang mit Shared Attributes
- [ ] RPC-Kommandos für alle Sensoren implementieren
- [ ] Dashboard-Integration mit Echtzeit-Updates
- [ ] Alarmfunktionen in ThingsBoard konfigurieren
- [ ] Statusüberwachung der Geräte implementieren

## Deployment & Stabilität
- [ ] Systemd-Service-Datei erstellen
- [ ] Auto-Start beim Boot konfigurieren
- [ ] Automatische Abhängigkeitsinstallation verbessern
- [ ] Logging-System mit Rotation implementieren
- [ ] Watchdog für Neustarts bei Problemen

## Dokumentation
- [x] Systemarchitektur dokumentieren (ARCHITECTURE.md)
- [ ] Installationsanleitung vervollständigen
- [ ] Fehlerbehebungshandbuch erstellen
- [ ] Konfigurationsoptionen dokumentieren
- [ ] ThingsBoard-Setup-Anleitung schreiben
- [ ] PROZESS_WORKFLOW.md für Systemabläufe erstellen

## Testing
- [ ] Umfassende Tests für Modbus-Kommunikation
- [ ] Tests für MQTT-Verbindung zu ThingsBoard
- [ ] End-to-End-Tests mit simulierten Sensoren
- [ ] Stresstests für Langzeitstabilität
- [ ] Offline-Modus mit Datenpufferung testen

## Zukünftige Funktionen
- [ ] Datenpufferung bei ThingsBoard-Verbindungsverlust
- [ ] Web-Interface für lokale Konfiguration
- [ ] Verschlüsselte MQTT-Verbindung zu ThingsBoard
- [ ] Unterstützung für weitere Sensortypen
- [ ] Remote-Update-Mechanismus implementieren 