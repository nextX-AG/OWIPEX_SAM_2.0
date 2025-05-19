<!-- Version: 0.1.6 | Last Updated: 2025-05-19 14:54:14 UTC -->


# Deployment und CI/CD-Prozess

## Übersicht

Dieses Dokument beschreibt den Entwicklungs-, Build-, Test- und Deployment-Prozess für das OWIPEX_SAM_2.0-System. Es definiert die Branching-Strategie, CI/CD-Pipeline und die Best Practices für Releases und Rollbacks.

## Branching-Strategie und Git-Workflow

### Branching-Modell

Wir verwenden eine angepasste Version des GitFlow-Workflows:

```
main ─────────────────────●───────●───────●───────────────────●───────
                          │       │       │                   │
                          │       │       │                   │
development ●────●────●───●───●───●───●───●───●───●───●───●───●───────
               │       │      │       │       │           │
               │       │      │       │       │           │
feature/xyz ───●───●───┘      │       │       │           │
                              │       │       │           │
                              │       │       │           │
feature/abc ──────────────────●───●───┘       │           │
                                              │           │
                                              │           │
bugfix/xyz ──────────────────────────────────●───●───────┘
```

### Branch-Typen

- **main**: Produktive Version, immer stabil und einsatzbereit
- **development**: Integrationsbranche für neue Features, Bugfixes und Verbesserungen
- **feature/\***: Entwicklung neuer Funktionalitäten
- **bugfix/\***: Behebung von Fehlern
- **hotfix/\***: Kritische Fehlerbehebungen direkt für den main-Branch
- **release/\***: Vorbereitungszweig für neue Releases

### Commit-Konventionen

Wir folgen dem konventionellen Commit-Format:

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

Beispiele:
- `feat(sensor): Unterstützung für neue pH-Sensoren hinzugefügt`
- `fix(modbus): Timeout-Handling bei fehlerhafter Verbindung korrigiert`
- `docs(api): ThingsBoard-API-Dokumentation aktualisiert`

### Pull Request-Prozess

1. Feature-/Bugfix-Branches werden von `development` abgezweigt
2. Nach Fertigstellung wird ein Pull Request erstellt
3. Code-Review durch mindestens einen anderen Entwickler
4. Automatisierte Tests müssen erfolgreich sein
5. Nach Genehmigung erfolgt der Merge in `development`

## CI/CD-Pipeline

### Kontinuierliche Integration

Für jeden Push und Pull Request wird automatisch folgende Pipeline ausgeführt:

```
┌───────────┐     ┌─────────┐     ┌────────────┐     ┌─────────────┐
│ Code-Push │────>│ Linting │────>│ Unit-Tests │────>│ Build & Tag │
└───────────┘     └─────────┘     └────────────┘     └─────────────┘
```

#### Build-Prozess

```
┌──────────────┐     ┌────────────────┐     ┌──────────────────┐
│ Go-Abhängig- │────>│ Kompilierung   │────>│ Statische Analyse│
│ keiten laden │     │ (Cross-Compile)│     │ (go vet, etc.)   │
└──────────────┘     └────────────────┘     └──────────────────┘
        │                                             │
        └─────────────────────┬─────────────────────┘
                              ▼
                     ┌─────────────────┐
                     │  Artefakte      │
                     │  (Binaries,     │
                     │   Konfiguration)│
                     └─────────────────┘
```

### Kontinuierliches Deployment

```
┌───────────┐     ┌────────────┐     ┌────────────┐     ┌──────────┐     ┌──────────┐
│ Entwickler│────>│ Development│────>│   Staging  │────>│ Produktion│────>│ Monitoring│
│ Umgebung  │     │ Umgebung   │     │  Umgebung  │     │ Umgebung  │     │          │
└───────────┘     └────────────┘     └────────────┘     └──────────┘     └──────────┘
   Kontinuierlich   Nach jedem PR    Mit Release-Tag   Nach Freigabe     Kontinuierlich
```

## Deployment-Umgebungen

### Entwicklungsumgebung

- **Zweck**: Lokale Entwicklung und Tests
- **Aktualisierung**: Kontinuierlich (manuelle Builds)
- **Konfiguration**: Entwicklungsspezifisch, mit Mocks für externe Dienste

### Development-Umgebung

- **Zweck**: Integration und Tests neuer Features
- **Aktualisierung**: Automatisch nach jedem Merge in den development-Branch
- **Konfiguration**: Testumgebung mit echten Sensoren oder Simulatoren

### Staging-Umgebung

- **Zweck**: Vorproduktive Tests unter realistischen Bedingungen
- **Aktualisierung**: Bei Release-Kandidaten (Tagged Commits)
- **Konfiguration**: Produktionsähnlich mit Testdaten

### Produktionsumgebung

- **Zweck**: Produktivsystem
- **Aktualisierung**: Nur nach manueller Freigabe eines Release-Kandidaten
- **Konfiguration**: Produktionskonfiguration

## Deployment-Prozess

### Kompilierung und Paketierung

1. **Kompilierung**:
   - Cross-Kompilierung für verschiedene Zielplattformen
   - Statische Verlinkung zur Minimierung von Abhängigkeiten
   - Build-Informationen in Binärdatei eingebettet

2. **Paketierung**:
   - Erstellung von Deployment-Paketen (Debian-Pakete, tarballs)
   - Erstellung Docker-Images (optional)
   - Signierung der Pakete

3. **Artefakt-Speicherung**:
   - Ablage in einem Artefakt-Repository
   - Versionierung basierend auf Git-Tags und Commit-Hash

### Deployment-Schritte

1. **Pre-Deployment Checks**:
   - Überprüfung der Zielsystemkompatibilität
   - Validierung von Konfigurationsparametern
   - Sicherung der aktuellen Konfiguration

2. **Deployment-Prozess**:
   - Stoppen des laufenden Dienstes
   - Einspielen des Updates
   - Anpassung der Konfiguration
   - Starten des Dienstes
   - Verifizierung der Funktionalität

3. **Post-Deployment Validierung**:
   - Funktionelle Tests
   - Performance-Monitoring
   - Logging-Überprüfung

## Rollback-Strategie

### Automatische Rollbacks

- Bei fehlgeschlagener Bereitstellung wird automatisch zum letzten stabilen Zustand zurückgekehrt
- Fehlschlagkriterien:
  - Dienst startet nicht innerhalb definierter Zeitspanne
  - Healthchecks schlagen fehl
  - Kritische Fehler in den Logs

### Manuelle Rollbacks

Prozess für manuelle Rollbacks:

1. Initiierung des Rollback-Befehls
2. Zurücksetzen auf vorherige Version
3. Wiederherstellung der entsprechenden Konfiguration
4. Validierung der wiederhergestellten Funktionalität

### Wiederherstellungspunkte

- Es werden die letzten drei stabilen Versionen vorgehalten
- Konfigurationen werden versioniert und mit den Releases verknüpft
- Datenbankschemata unterstützen Abwärtskompatibilität

## Monitoring nach Deployment

### Laufzeitüberwachung

- **Systemmetriken**: CPU, Speicher, Disk, Netzwerk
- **Anwendungsmetriken**: Anfragen pro Sekunde, Latenz, Fehlerrate
- **Geschäftsmetriken**: Erfolgsrate von Sensormessungen, Datenqualität

### Alert-System

- Schwellenwertbasierte Alarme für kritische Metriken
- Eskalationsstufen basierend auf Schweregrad
- Benachrichtigungskanäle: E-Mail, SMS, Chat-Integration

### Dashboards

- Echtzeit-Monitoring-Dashboards
- Deployment-Historie und -Performance
- Anomalieerkennung

## Release-Management

### Versionsschema

Wir verwenden [Semantische Versionierung](https://semver.org/):

- **Major Version**: Inkompatible API-Änderungen
- **Minor Version**: Abwärtskompatible Funktionserweiterungen
- **Patch Version**: Abwärtskompatible Bugfixes

### Release-Zyklus

- Regelmäßiges Release-Fenster alle zwei Wochen
- Hotfixes nach Bedarf (außerhalb des regulären Zyklus)
- Feature-Freezes eine Woche vor größeren Releases

### Release-Dokumentation

- Release Notes mit allen Änderungen
- Upgrade-Anleitung für Breaking Changes
- Bekannte Probleme und Workarounds

## Systemd Service und Autostart

### Systemd-Integration

Für Linux-basierte Deployments verwenden wir systemd-Services:

```ini
[Unit]
Description=OWIPEX_SAM_2.0 - Modbus/MQTT Bridge for Water Treatment
After=network.target redis.service
Wants=redis.service

[Service]
Type=simple
User=owipex
Group=owipex
WorkingDirectory=/opt/owipex/rs485go
ExecStart=/opt/owipex/rs485go/bin/OWIPEX_SAM_2.0 --config /etc/owipex/config.json
Restart=on-failure
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=owipexrs485go
Environment=PATH=/usr/bin:/bin:/usr/local/bin

# Security hardening
PrivateTmp=true
ProtectSystem=full
ReadWritePaths=/var/lib/owipex /var/log/owipex
NoNewPrivileges=true
ProtectHome=true
ProtectControlGroups=true
ProtectKernelModules=true
ProtectKernelTunables=true

[Install]
WantedBy=multi-user.target
```

### Start beim Systemstart

Aktivierung des Systemd-Services:

```bash
sudo systemctl enable owipexrs485go.service
```

### Logging-Integration

- Logging in systemd-Journal
- Optional: Weiterleitung an externes Logging-System
- Log-Rotation für langfristige Protokollierung

## Entwicklungsworkflow

### Lokale Entwicklung

1. Repository klonen:
   ```bash
   git clone https://github.com/KARIM-Technologies/OWIPEX_SAM_2.0.git
   cd OWIPEX_SAM_2.0
   ```

2. Branch für neue Funktion erstellen:
   ```bash
   git checkout development
   git pull
   git checkout -b feature/new-sensor-type
   ```

3. Änderungen entwickeln, testen und committen

4. Push und Pull Request erstellen:
   ```bash
   git push -u origin feature/new-sensor-type
   ```

5. Nach Code-Review und erfolgreichen Tests: Merge in `development`

### Tipps für Entwickler

- Lokalen Build-Prozess vor Pull Requests ausführen
- Tests für neue Funktionen schreiben
- Linting-Tools zur Qualitätssicherung verwenden
- Sorgfältige Dokumentation neuer Features 