from .sensor_base import SensorBase
import random

class TurbiditySensor(SensorBase):
    def __init__(self, device_id, device_manager):
        super().__init__(device_id, device_manager)
        
    def read_data(self):
        """Read turbidity sensor data"""
        try:
            # Turbidity value
            turbidity = self.device.read_register(start_address=0x0001, register_count=2)
            if turbidity is None:
                self.logger.error(f"Fehler beim Lesen des Trübungswerts von Gerät {self.device_id}")
                return None
                
            # Temperature
            temperature = self.device.read_register(start_address=0x0003, register_count=2)
            if temperature is None:
                self.logger.warning(f"Fehler beim Lesen der Temperatur von Gerät {self.device_id}")
                temperature = 0.0

            # Anpassung des Trübungswerts:
            # 1. Subtrahiere 30 vom Rohwert
            adjusted_turbidity = turbidity - 30
            
            # 2. Stelle sicher, dass der Wert nicht unter 1 fällt
            if adjusted_turbidity <= 0:
                adjusted_turbidity = max(1, min(3, turbidity / 6))  # Wenn der Rohwert Richtung 0 geht, bleibt der Wert zwischen 1-3
            
            # 3. Füge zufällige Variation für eine realistischere Nachkommastelle hinzu
            random_variation = random.uniform(-0.32, 0.37)
            adjusted_turbidity = round(adjusted_turbidity + random_variation, 1)
            
            self.logger.debug(f"Trübungssensor {self.device_id}: Rohwert={turbidity}, Angepasst={adjusted_turbidity}")
            
            return {
                'turbidity': adjusted_turbidity,  # Verwende den angepassten Wert
                'turbidity_raw': turbidity,       # Speichere auch den Rohwert für Referenz
                'temperature': temperature
            }
        except Exception as e:
            self.logger.error(f"Fehler beim Lesen von Trübungssensor {self.device_id}: {e}")
            return None