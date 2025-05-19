<!-- Version: 0.1.6 | Last Updated: 2025-05-19 14:54:14 UTC -->


# OWIPEX_SAM_2.0 - Task List

## In Bearbeitung
- [x] Minimalen ThingsBoard-Client für Tests implementieren
- [x] Robuste MQTT-Verbindung zu ThingsBoard aufbauen
- [x] Shared Attributes von ThingsBoard empfangen
- [x] Modulare Sensorarchitektur implementieren
- [x] Migration der bestehenden Sensoren in die neue Architektur
- [x] Integration der neuen Architektur in die Hauptanwendung

## Projektstruktur
- [x] Reorganisation des Projekts in eine klare, modulare Struktur
- [x] Wechsel zu einer hierarchischen Ordnerstruktur mit logischen Gruppierungen
- [x] Verschieben des ThingsBoard MQTT Clients in die Integration-Schicht
- [x] Dokumentation der neuen Struktur in ARCHITECTURE.md
- [x] Anpassung und Verschiebung aller Module in die neue Struktur
- [x] Verschieben der Hilfsskripte in den scripts-Ordner
- [ ] Aktualisierung der Import-Pfade in allen Dateien
- [x] Entfernen von Altlasten aus der vorherigen Projektstruktur
- [ ] Refactoring der ThingsBoard-Integration in die neue Struktur 

## Hohe Priorität
- [ ] RPC-Befehle von ThingsBoard empfangen und verarbeiten
- [x] Robustes Error-Handling für Modbus-Verbindungen implementieren
- [x] Reconnect-Mechanismus für verlorene RS485-Verbindungen
- [x] Umgang mit Timeouts bei Modbus-Kommunikation
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
- [x] Implementierung der weiteren Sensortypen (Flow, Radar, Turbidity)
- [ ] Implementierung der spezifischen Aktortypen (Ventil, Pumpe, etc.)
- [x] Protokoll-Abstraktion für verschiedene Kommunikationsarten
- [x] Einheitliches Messwert- und Befehlssystem
- [x] Konverter für Rohdaten zu physikalischen Werten

## Sensor-Migration
- [x] Basis-Sensor-Implementierung erstellen (BaseSensor)
- [x] PH-Sensor in neue Struktur migrieren
- [x] Flow-Sensor in neue Struktur migrieren
- [x] Radar-Sensor in neue Struktur migrieren
- [x] Turbidity-Sensor in neue Struktur migrieren
- [x] Konfigurationsbeispiele für alle Sensoren erstellen
- [ ] Alte Sensorimplementierungen entfernen

## System-Integration
- [ ] Integration der neuen Gerätemodelle in die Hauptanwendung
- [ ] Ersetzung der alten Modbus-Implementierung durch die neue
- [ ] Anpassung der ThingsBoard-Integration an die neue Gerätestruktur
- [ ] End-to-End-Tests mit realen Geräten
- [ ] Überwachung der Ressourcennutzung

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

## GPIO-Implementierung
- [ ] Abstrakte GPIO-Schnittstelle definieren
- [ ] Plattformspezifische Implementierungen:
  - [ ] Linux-GPIO (sysfs) implementieren
  - [ ] Linux-GPIO (gpiod) implementieren
  - [ ] Mock-GPIO für Tests implementieren
- [ ] Input-Handling mit Debouncing für Buttons
- [ ] Output-Handling für LEDs und Relais
- [ ] Event-basierte GPIO-Überwachung
- [ ] Konfiguration von GPIO-Pins über JSON
- [ ] Integration der GPIO-Funktionalität mit Aktoren
- [ ] Watchdog-Funktionalität aus der alten Implementierung portieren

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
- [x] Automatisches Versionierungssystem für Dokumentation implementieren:
  - [x] VERSION.txt für zentrale Dokumentationsversion erstellen
  - [x] Git pre-commit Hook für automatische Versionierung einrichten
  - [x] CI/CD-Workflow für automatisches Changelog konfigurieren
  - [x] VERSIONING.md mit Beschreibung des Versionierungssystems erstellen
  - [x] Integration des Versionssystems in bestehende Dokumentationsdateien
- [ ] Management-Summary für Entscheidungsträger erstellen:
  - [ ] EXECUTIVE_SUMMARY.md erstellen mit Business-fokussierter Übersicht
  - [ ] Infografik zur Gesamtarchitektur für Nicht-Techniker
  - [ ] Erstellung einer Nutzen- und ROI-Übersicht
  - [ ] Timeline und wichtigste Meilensteine visualisieren
- [ ] Sicherheitskonzept dokumentieren:
  - [ ] Erweitertes Security-Kapitel in ARCHITECTURE.md hinzufügen
  - [ ] Authentifizierungskonzept für alle Schnittstellen definieren
  - [ ] Verschlüsselungsstrategien für Daten beschreiben
  - [ ] Zertifikatsmanagement für TLS/MQTT/Redis dokumentieren
  - [ ] Absicherungsmaßnahmen gegen IoT-Angriffsvektoren beschreiben
- [ ] CI/CD und Deployment-Prozess dokumentieren:
  - [ ] Branching-Strategie und Git-Workflow definieren
  - [ ] Testphasen und -stufen dokumentieren
  - [ ] Deployment-Pipeline mit Stages beschreiben
  - [ ] Rollback-Strategien und Notfallmaßnahmen definieren
  - [ ] Monitoring nach Deployment dokumentieren

## Testing
- [ ] Umfassende Tests für Modbus-Kommunikation
- [ ] Tests für MQTT-Verbindung zu ThingsBoard
- [ ] End-to-End-Tests mit simulierten Sensoren
- [ ] Stresstests für Langzeitstabilität
- [ ] Offline-Modus mit Datenpufferung testen

## Bereinigung nach vollständiger Migration
- [x] Entfernen der alten Modbus-Implementierung:
  - [x] internal/modbus/modbus_client.go
  - [x] internal/modbus/modbus_client_test.go
- [x] Entfernen der alten Sensor-Implementierungen:
  - [x] internal/sensor/ph_sensor.go
  - [x] internal/sensor/flow_sensor.go
  - [x] internal/sensor/radar_sensor.go
  - [x] internal/sensor/turbidity_sensor.go
  - [x] internal/sensor/sensor.go
- [x] Anpassen des SensorManager:
  - [x] internal/manager/sensor_manager.go
- [ ] Anpassen der ThingsBoard-Integration:
  - [ ] Anpassungen in internal/thingsboard/thingsboard_client.go

## Zukünftige Funktionen
- [ ] Datenpufferung bei ThingsBoard-Verbindungsverlust
- [ ] Web-Interface für lokale Konfiguration
- [ ] Verschlüsselte MQTT-Verbindung zu ThingsBoard
- [ ] Unterstützung für weitere Sensortypen
- [ ] Remote-Update-Mechanismus implementieren 