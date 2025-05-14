from .sensor_base import SensorBase

class PHSensor(SensorBase):
    def __init__(self, device_id, device_manager):
        super().__init__(device_id, device_manager)
        
    def read_data(self):
        """Read pH sensor data"""
        try:
            # Lese PH-Wert (Register 0x0001, 2 Register)
            ph_value = self.device.read_register(start_address=0x0001, register_count=2)
            if ph_value is None:
                self.logger.error(f"Fehler beim Lesen des PH-Werts von Gerät {self.device_id}")
                return None
                
            # Lese Temperatur (Register 0x0003, 2 Register)
            temperature = self.device.read_register(start_address=0x0003, register_count=2)
            if temperature is None:
                self.logger.error(f"Fehler beim Lesen der Temperatur von Gerät {self.device_id}")
                return None
                
            self.logger.debug(f"PH-Wert: {ph_value}, Temperatur: {temperature}")
            return {
                'ph_value': ph_value,
                'temperature': temperature
            }
        except Exception as e:
            self.logger.error(f"Fehler beim Lesen des PH-Sensors (ID: {self.device_id}): {str(e)}")
            return None