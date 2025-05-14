from .sensor_base import SensorBase
import time
import struct
import logging

class FlowSensor(SensorBase):
    def __init__(self, device_id, device_manager):
        super().__init__(device_id, device_manager)
        self.logger = logging.getLogger(f'Sensor_FlowSensor_{device_id}')
        self.logger.info(f"FlowSensor {device_id} initialisiert")
        
    def read_data(self):
        """Read flow sensor data"""
        try:
            # Längere initiale Pause vor jeder Kommunikation
            time.sleep(0.3)
            
            # ZUERST: Total Flow lesen (vor Flow Rate)
            # Lese Total Flow mit den korrekten Registern (0x000A/0x0011)
            self.logger.debug(f"Lese Total Flow Low Register (0x000A) von Sensor {self.device_id}")
            flow_int_low = self.device.read_register(0x000A)
            
            # Wichtige Pause zwischen den Register-Abfragen
            time.sleep(0.5)
            
            self.logger.debug(f"Lese Total Flow High Register (0x0011) von Sensor {self.device_id}")
            flow_int_high = self.device.read_register(0x0011)
            
            # Längere Pause vor den Konfigurationsregistern
            time.sleep(0.5)
            
            # Lese Konfigurationsregister
            self.logger.debug(f"Lese Flow Unit (0x1438) von Sensor {self.device_id}")
            flow_unit = self.device.read_register(0x1438)
            
            time.sleep(0.3)
            
            self.logger.debug(f"Lese Flow Decimal Point (0x1439) von Sensor {self.device_id}")
            flow_decimal_point = self.device.read_register(0x1439)
            
            # Lange Pause vor dem Lesen der Flow Rate
            time.sleep(1.0)
            
            # ZULETZT: Read flow rate
            self.logger.debug(f"Lese Flow Rate (0x0001) von Sensor {self.device_id}")
            flow_rate = self.device.read_flow_sensor(0x0001)
            if flow_rate is None:
                self.logger.error(f"Konnte flow_rate nicht lesen von Sensor {self.device_id}")
                return None
                
            # Diese Werte werden berechnet, aber nicht in der Telemetrie gesendet
            velocity = self.device.read_flow_sensor(0x0005) or 0.0
            temp_supply = self.device.read_flow_sensor(0x0033) or 0.0
            temp_return = self.device.read_flow_sensor(0x0035) or 0.0
            
            # Prüfe auf ungültige Werte und verwende Standardwerte
            if flow_unit is None or flow_unit == 0xFFFF:
                flow_unit = 0  # Standard: m³
                
            if flow_decimal_point is None or flow_decimal_point == 0xFFFF:
                flow_decimal_point = 3  # Standard-Dezimalpunkt
                
            # Berechne Gesamtdurchfluss
            if flow_int_low is not None and flow_int_high is not None:
                flow_integer = (flow_int_high << 16) | flow_int_low
                multiplier = 10 ** (flow_decimal_point - 3)
                total_flow = flow_integer * multiplier
                
                # Log der Rohwerte für Debugging
                self.logger.debug(f"Flow integer: {flow_integer} (Low: 0x{flow_int_low:04X}, High: 0x{flow_int_high:04X})")
                self.logger.debug(f"Decimal point: {flow_decimal_point}, Multiplier: {multiplier}")
            else:
                flow_integer = 0
                multiplier = 0.001
                total_flow = 0.0
                
            # Map flow unit zu String
            flow_unit_map = {0: 'm³', 1: 'L', 2: 'GAL', 3: 'CF', 5: 'ft³'}
            flow_unit_str = flow_unit_map.get(flow_unit, 'm³')
            
            self.logger.info(f"Erfolgreich gelesen - Flow Sensor {self.device_id}: flow_rate={flow_rate}, total_flow={total_flow}")
            
            # Gib nur die benötigten Werte zurück
            return {
                'flow_rate': flow_rate,
                'total_flow': total_flow,
                'total_flow_unit': flow_unit_str
            }
            
        except Exception as e:
            self.logger.error(f"Fehler beim Lesen von Flow Sensor {self.device_id}: {e}")
            return None