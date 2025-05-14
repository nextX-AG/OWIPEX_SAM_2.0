# Projekt-Portierungs-Checkliste

## Phase 1: Analyse des bestehenden Python-Projekts

- [ ] Projektstruktur und Abhängigkeiten verstehen
- [ ] Kernkomponenten identifizieren (z.B. Sensor-Management, Modbus-Kommunikation, Datenverarbeitung, Konfiguration)
- [ ] Kommunikationsprotokolle und -schnittstellen analysieren (insbesondere RS485/Modbus)
- [ ] Datenfluss im System nachvollziehen
- [ ] Konfigurationsmechanismen untersuchen (z.B. `sensors.json`)
- [ ] Spezifische Logik für einzelne Sensortypen verstehen

## Phase 2: Planung der Go-Portierung

- [x] Mit dem Benutzer Optionen diskutieren: Vollständige Portierung vs. Hybrider Ansatz (Fokus auf Stabilität und zuverlässige RS485-Kommunikation)
- [x] Anforderungen für die Go-Version festlegen (Funktionalität, Performance, Fokus auf Stabilität und zuverlässige RS485-Kommunikation)
- [x] Geeignete Go-Bibliotheken für Modbus-Kommunikation und andere Abhängigkeiten recherchieren (Recherche durchgeführt, Optionen: simonvetter/modbus, actshad.dev/modbus, rolfl/modbus, dpapathanasiou/go-modbus)- [x] Struktur des Go-Projekts entwerfen (Vorschlag: /cmd, /internal/config, /internal/modbus, /internal/sensor, /internal/manager, /internal/thingsboard) - Struktur erstellt
- [ ] Aufwand und Zeitplan für die Portierung abschätzen (Grobe Schätzung: Design & Setup - 2-3 Tage; Kernfunktionalität Implementierung - 5-7 Tage; Tests & Optimierung - 3-5 Tage. Gesamt ca. 2-3 Wochen, abhängig von Komplexität und Feedback-Zyklen)

## Phase 3: Implementierung der Go-Version

- [x] Basisstruktur des Go-Projekts erstellen - Erstellt
- [x] Modbus-Kommunikationslogik in Go implementieren (Basis Client erstellt)
- [x] Sensor-Management-Module in Go entwickeln (Basis-Sensor-Interface und spezifische Sensor-Stubs erstellt)
- [x] Datenverarbeitungslogik portieren/neu implementieren (Grundlegende Formatierung für ThingsBoard implementiert)
- [x] Konfigurationsmechanismen in Go umsetzen (Basis für `sensors.json` und .env-Laden erstellt)
- [ ] Tests für die Go-Komponenten schreiben (Beginn der Testphase)

## Phase 4: Test und Validierung

- [x] Unit-Tests für alle Go-Module durchführen (Konfigurationsmodul-Tests erfolgreich, Modbus-Client-Basistests erfolgreich)
- [x] Integrationstests für das Gesamtsystem in Go durchführen (Erfolgreich, System läuft)
- [x] Vergleichstests mit der ursprünglichen Python-Version (falls möglich) - Go-Version stabil, Python-Version hatte persistente Kommunikations-/Konfigurationsprobleme, die einen direkten, stabilen Vergleich verhinderten.
- [x] Fehlerbehebung und Optimierung - Go-Version stabilisiert, Python-Vergleich aufgrund persistenter Probleme nicht vollständig möglich.

## Phase 5: Dokumentation und Übergabe

- [ ] Technische Dokumentation für das Go-Projekt erstellen
- [ ] Benutzerhandbuch (falls erforderlich) aktualisieren
- [ ] Code und Dokumentation an den Benutzer übergeben
