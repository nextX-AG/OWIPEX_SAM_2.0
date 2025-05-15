# owipexRS485GO - Task List

## In Bearbeitung
- [x] Minimalen ThingsBoard-Client für Tests implementieren
- [x] Robuste MQTT-Verbindung zu ThingsBoard aufbauen
- [x] Shared Attributes von ThingsBoard empfangen

## Hohe Priorität
- [ ] RPC-Befehle von ThingsBoard empfangen und verarbeiten
- [ ] Robustes Error-Handling für Modbus-Verbindungen implementieren
- [ ] Reconnect-Mechanismus für verlorene RS485-Verbindungen
- [ ] Umgang mit Timeouts bei Modbus-Kommunikation
- [ ] Automatische Wiederherstellung der ThingsBoard-Verbindung

## Sensor-Integration
- [ ] Konfiguration für Sensoren erweitern (enabled/disabled-Flag)
- [ ] Radar-Sensor vollständig integrieren
- [ ] PH-Sensor-Integration abschließen
- [ ] Flow-Sensor-Integration optimieren
- [ ] Turbidity-Sensor-Integration verbessern
- [ ] Kalibrierungsfunktionen für Sensoren implementieren

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
- [ ] Systemarchitektur dokumentieren (ARCHITECTURE.md)
- [ ] Installationsanleitung vervollständigen
- [ ] Fehlerbehebungshandbuch erstellen
- [ ] Konfigurationsoptionen dokumentieren
- [ ] ThingsBoard-Setup-Anleitung schreiben

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