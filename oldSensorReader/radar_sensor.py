from .sensor_base import SensorBase
from device_config.radar_sensor_config import RadarSensorConfig
from calculations.radar_calculations import RadarCalculations

class RadarSensor(SensorBase):
    def __init__(self, device_id, device_manager):
        super().__init__(device_id, device_manager)
        # Lade Konfiguration
        sensor_number = int(str(device_id))
        config = RadarSensorConfig(sensor_number)
        # Initialisiere Berechnungsmodul
        self.calculations = RadarCalculations(config.config)
        
    def read_data(self):
        """Read radar sensor data and calculate derived values"""
        try:
            measured_air_distance = self.device.read_radar_sensor(register_address=0x0001)
            
            if measured_air_distance is None:
                self.logger.error(f"Fehler beim Lesen des Füllstands von Gerät {self.device_id}")
                return None
                
            # Berechne alle abgeleiteten Werte mit dem Berechnungsmodul
            actual_water_level = self.calculations.calculate_water_level(measured_air_distance)
            actual_volume = self.calculations.calculate_volume(actual_water_level)
            volume_percentage = self.calculations.calculate_volume_percentage(actual_water_level)
            level_above_normal = self.calculations.calculate_level_above_normal(actual_water_level)
            water_level_alarm = self.calculations.check_water_level_alarm(actual_water_level)
            
            self.logger.debug(f"""
                Radar Sensor {self.device_id}:
                Gemessene Luftstrecke: {measured_air_distance} mm
                Aktueller Wasserstand: {actual_water_level} mm
                Aktuelles Volumen: {actual_volume:.2f} m³
                Füllstand: {volume_percentage:.1f}%
                Abweichung von Normal: {level_above_normal} mm
                Alarm: {water_level_alarm}
            """)
            
            return {
                'measured_air_distance': measured_air_distance,
                'actual_water_level': actual_water_level,
                'actual_volume': round(actual_volume, 3),
                'volume_percentage': round(volume_percentage, 1),
                'level_above_normal': level_above_normal,
                'water_level_alarm': water_level_alarm
            }
        except Exception as e:
            self.logger.error(f"Fehler beim Lesen des Radarsensors (ID: {self.device_id}): {str(e)}")
            return None