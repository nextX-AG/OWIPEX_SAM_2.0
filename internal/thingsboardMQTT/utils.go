package thingsboardMQTT

// stringJoin verbindet Strings mit einem Trennzeichen.
func stringJoin(arr []string, sep string) string {
	if len(arr) == 0 {
		return ""
	}

	result := arr[0]
	for i := 1; i < len(arr); i++ {
		result += sep + arr[i]
	}

	return result
}

// checkForFirmwareUpdate prüft Attribut-Updates auf Firmware-Updates
func (c *Client) checkForFirmwareUpdate(attributes map[string]interface{}) {
	// Firmware-Update-Felder prüfen
	fwTitle, hasTitle := attributes["fw_title"].(string)
	fwVersion, hasVersion := attributes["fw_version"].(string)
	fwChecksum, hasChecksum := attributes["fw_checksum"].(string)
	fwAlgorithm, hasAlgorithm := attributes["fw_checksum_algorithm"].(string)

	// Wenn alle Firmware-Felder vorhanden sind und ein Callback registriert ist
	if hasTitle && hasVersion && hasChecksum && hasAlgorithm && c.firmwareCallback != nil {
		c.firmwareCallback(fwTitle, fwVersion, fwChecksum, fwAlgorithm)
	}
}
