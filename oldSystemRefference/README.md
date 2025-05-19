# Alte System-Referenzimplementierung

Dieses Verzeichnis enthält die ursprüngliche Python-Implementierung des Owipex-Systems, die als Referenz für die Go-Migration dient.

## Bedeutung

Die hier enthaltenen Dateien sind wichtig für:
- Das Verständnis der ursprünglichen Funktionalität
- Die Überprüfung der korrekten Implementierung in Go
- Die Rückverfolgung von unerwarteten Verhaltensweisen zur Quelle
- Den Zugriff auf Algorithmen und Geschäftslogik im Original

## Wichtige Komponenten

### oldSensorReader/oldH2O

Enthält die ursprüngliche Python-Implementierung mit:
- `h2o.py` - Hauptskript für die Wasseraufbereitungssteuerung
- `modbus_lib.py` - Kommunikation mit Modbus-Sensoren
- `powerWatchdog.py` - Überwachung und Neustart des Systems

## Verwendung

Diese Dateien sollten nicht produktiv eingesetzt, sondern nur als Referenz genutzt werden. Die Go-Implementierung ersetzt diese Python-Version vollständig. 