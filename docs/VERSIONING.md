<!-- Version: 0.1.6 | Last Updated: 2025-05-19 14:54:14 UTC -->


# Dokumentations-Versionierungssystem

## Übersicht

Dieses Dokument beschreibt das automatische Versionierungssystem für die Dokumentation des OWIPEX_SAM_2.0-Projekts. Das System ist darauf ausgelegt, ohne manuelle Eingriffe zu funktionieren und eine konsistente Versionierung aller Dokumentationsdateien sicherzustellen.

## Versionierungskonzept

### Versionsnummerierung

Wir verwenden [Semantische Versionierung](https://semver.org/lang/de/) für unsere Dokumentation:

- **Major.Minor.Patch** (z.B. 1.2.3)
  - **Major**: Strukturelle oder konzeptionelle Änderungen der Dokumentation
  - **Minor**: Neue Abschnitte oder signifikante Erweiterungen
  - **Patch**: Kleine Änderungen, Korrekturen, Verbesserungen

Die Version wird automatisch erhöht, wobei:
- Bei normalen Änderungen wird die Patch-Version automatisch erhöht
- Minor- und Major-Versionen werden manuell bei signifikanten Änderungen angepasst

### Zentrale Versionsverwaltung

Die aktuelle Dokumentationsversion wird zentral in der Datei `docs/VERSION.txt` verwaltet. Diese Datei dient als Single Source of Truth für die Dokumentationsversion.

## Automatische Versionierung

### Git Pre-Commit Hook

Ein Git Pre-Commit Hook übernimmt die automatische Versionierung:

1. Erkennt Änderungen an Markdown-Dateien im `docs/`-Verzeichnis
2. Liest die aktuelle Version aus `docs/VERSION.txt`
3. Erhöht die Patch-Version automatisch
4. Aktualisiert die Versionsinfo in den geänderten Dokumenten
5. Aktualisiert `docs/VERSION.txt`

Hier ist der Code des Pre-Commit Hooks:

```bash
#!/bin/bash

# Finde geänderte Markdown-Dateien im docs/-Verzeichnis
docs_changed=$(git diff --cached --name-only --diff-filter=ACM | grep '^docs/.*\.md$')

if [ -n "$docs_changed" ]; then
  # Extrahiere letzte Version aus VERSION.txt oder initialisiere
  if [ -f docs/VERSION.txt ]; then
    version=$(cat docs/VERSION.txt)
  else
    version="0.1.0"
  fi
  
  # Erhöhe Patch-Version
  patch_version=$(echo $version | awk -F. '{print $3}')
  new_patch=$((patch_version + 1))
  new_version=$(echo $version | awk -F. -v p=$new_patch '{print $1"."$2"."p}')
  
  # Aktualisiere VERSION.txt
  echo $new_version > docs/VERSION.txt
  git add docs/VERSION.txt
  
  # Füge Versionsinfo in geänderte Dokumente ein
  for file in $docs_changed; do
    timestamp=$(date -u +"%Y-%m-%d %H:%M:%S UTC")
    # Entferne vorhandenen Versions-Header (falls vorhanden)
    sed -i '/^<!-- Version: .* | Last Updated: .* -->/d' $file
    # Füge neuen Versions-Header hinzu
    sed -i "1s/^/<!-- Version: $new_version | Last Updated: $timestamp -->\n\n/" $file
    git add $file
  done
fi
```

> **Hinweis:** Für macOS muss der `sed`-Befehl angepasst werden. MacOS erfordert ein leeres Argument nach dem `-i`-Flag. Die korrekte Version für macOS sieht so aus:
> ```bash
> sed -i '' '/^<!-- Version: .* | Last Updated: .* -->/d' $file
> sed -i '' "1s/^/<!-- Version: $new_version | Last Updated: $timestamp -->\n\n/" $file
> ```
> Der im Repository eingerichtete pre-commit Hook verwendet bereits die macOS-kompatible Version.

### CI/CD-Integration

Ein CI/CD-Workflow (GitHub Actions) generiert automatisch ein Changelog der Dokumentationsänderungen:

1. Wird bei Änderungen an Dateien im `docs/`-Verzeichnis ausgelöst
2. Liest die aktuelle Version aus `docs/VERSION.txt`
3. Generiert einen Eintrag im Changelog für die neue Version
4. Listet alle geänderten Dokumentationsdateien und zugehörige Commits auf

Hier ist ein Beispiel für den GitHub Actions Workflow:

```yaml
# In .github/workflows/docs-version.yml
name: Update Docs Version

on:
  push:
    paths:
      - 'docs/**'
    branches:
      - main

jobs:
  update-changelog:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
        with:
          fetch-depth: 0
      
      - name: Generate Docs Changelog
        run: |
          # Hole aktuelle Version
          version=$(cat docs/VERSION.txt)
          
          # Generiere Changelog seit letztem Tag
          echo "# Dokumentations-Changelog" > docs/CHANGELOG.md
          echo "## Version $version - $(date +"%Y-%m-%d")" >> docs/CHANGELOG.md
          echo "" >> docs/CHANGELOG.md
          
          # Füge geänderte Dateien mit Commits hinzu
          git log --name-status --pretty=format:"%h %s" HEAD~10..HEAD -- docs/ | grep -v "^$" | grep -v "Update VERSION" >> docs/CHANGELOG.md
          
          # Committe Änderungen
          git config --local user.email "action@github.com"
          git config --local user.name "GitHub Action"
          git add docs/CHANGELOG.md
          git commit -m "Update docs changelog [skip ci]"
          git push
```

## Versionsinformationen in Dokumenten

Jedes Markdown-Dokument im `docs/`-Verzeichnis erhält einen Versions-Header im folgenden Format:

```markdown

# Dokumenttitel
...
```

Diese Header werden automatisch vom Pre-Commit Hook eingefügt oder aktualisiert.

## Manuelle Versionsverwaltung (bei Bedarf)

In Ausnahmefällen kann die Version manuell angepasst werden:

1. **Erhöhung der Minor-Version:**
   ```bash
   # Aktualisiere VERSION.txt
   version=$(cat docs/VERSION.txt)
   major=$(echo $version | cut -d. -f1)
   minor=$(echo $version | cut -d. -f2)
   new_minor=$((minor + 1))
   echo "$major.$new_minor.0" > docs/VERSION.txt
   git add docs/VERSION.txt
   git commit -m "Increment docs minor version to $major.$new_minor.0"
   ```

2. **Erhöhung der Major-Version:**
   ```bash
   # Aktualisiere VERSION.txt
   version=$(cat docs/VERSION.txt)
   major=$(echo $version | cut -d. -f1)
   new_major=$((major + 1))
   echo "$new_major.0.0" > docs/VERSION.txt
   git add docs/VERSION.txt
   git commit -m "Increment docs major version to $new_major.0.0"
   ```

## Installation und Wartung

### Installation

1. Erstellen der initialen Versionsdatei:
   ```bash
   mkdir -p docs
   echo "0.1.0" > docs/VERSION.txt
   git add docs/VERSION.txt
   git commit -m "Initialize documentation versioning"
   ```

2. Einrichten des Git-Hooks:
   ```bash
   mkdir -p .git/hooks
   # Kopiere den obigen Pre-Commit Hook-Code in die Datei
   vi .git/hooks/pre-commit
   chmod +x .git/hooks/pre-commit
   ```

3. Einrichten des CI/CD-Workflows:
   ```bash
   mkdir -p .github/workflows
   # Kopiere den obigen YAML-Code in die Datei
   vi .github/workflows/docs-version.yml
   git add .github/workflows/docs-version.yml
   git commit -m "Add docs versioning workflow"
   ```

### Wartung

- Der Pre-Commit Hook sollte mit Änderungen an der Repository-Struktur aktualisiert werden
- Bei Problemen mit der automatischen Versionierung überprüfen:
  - Berechtigungen des Pre-Commit Hooks
  - Zugriffsrechte für GitHub Actions
  - Format der VERSION.txt-Datei

## Vorteile

Dieses automatische Versionierungssystem bietet mehrere Vorteile:

1. **Einfachheit**: Keine manuelle Eingabe von Versionsnummern nötig
2. **Konsistenz**: Einheitliche Versionierung aller Dokumentationsdateien
3. **Nachverfolgbarkeit**: Klare Zuordnung von Änderungen zu Versionen
4. **Transparenz**: Sichtbarkeit der Versionshistorie direkt in den Dokumenten
5. **Automatisierung**: Vollständige Integration in den Git-Workflow 