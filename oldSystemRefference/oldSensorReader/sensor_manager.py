import os
import time
import json
import logging
from dotenv import load_dotenv
from tb_gateway_mqtt import TBDeviceMqttClient
from modbus_manager import DeviceManager
from .ph_sensor import PHSensor
from .turbidity_sensor import TurbiditySensor
from .flow_sensor import FlowSensor
from .radar_sensor import RadarSensor
from queue import Queue
from threading import Lock

class SensorManager:
    def __init__(self, config_path='config/sensors.json'):
        # Load environment variables
        load_dotenv(dotenv_path='/etc/owipex/.envRS485')
        
        # Initialize logging
        logging.basicConfig(
            level=logging.INFO,
            format='%(asctime)s - %(levelname)s - [%(name)s] - %(message)s'
        )
        self.logger = logging.getLogger('SensorManager')
        
        self.logger.info("Initialisiere SensorManager...")
        
        # Load configuration
        with open(config_path, 'r') as f:
            self.config = json.load(f)
        
        # Initialize Modbus connection with settings from config
        rs485_settings = self.config.get('rs485_settings', {})
        self.logger.info("Stelle Modbus-Verbindung her...")
        self.dev_manager = DeviceManager(
            port=rs485_settings.get('port', '/dev/ttyS0'),
            baudrate=rs485_settings.get('baudrate', 9600),
            parity=rs485_settings.get('parity', 'N'),
            stopbits=rs485_settings.get('stopbits', 1),
            bytesize=rs485_settings.get('bytesize', 8),
            timeout=rs485_settings.get('timeout', 1)
        )
        
        # RS485 Bus Management
        self.bus_lock = Lock()
        self.last_communication_time = 0
        self.DEBOUNCE_TIME = 0.5  # 500ms Mindestabstand zwischen Kommunikationen
        
        # Sensor Reading Queue
        self.read_queue = Queue()
        
        # Load sensor configuration
        self.sensors = self.load_sensors(self.config.get('sensors', []))
        
        # Initialize ThingsBoard connection
        self.client = None
        self.running = False
        self.last_read_times = {}
        self.READ_INTERVAL = int(os.environ.get('RS485_READ_INTERVAL', 15))
        self.logger.info(f"Read Interval: {self.READ_INTERVAL} Sekunden")

    def load_sensors(self, sensor_configs):
        """Load sensor configuration from config"""
        sensors = {}
        sensor_classes = {
            'ph': PHSensor,
            'turbidity': TurbiditySensor,
            'flow': FlowSensor,
            'radar': RadarSensor
        }
        
        for sensor_config in sensor_configs:
            sensor_type = sensor_config['type']
            sensor_id = sensor_config['id']
            device_id = sensor_config['device_id']
            
            self.logger.info(f"Konfiguriere Sensor: {sensor_id} (Typ: {sensor_type}, Device ID: {device_id})")
            
            if sensor_type in sensor_classes:
                sensor_class = sensor_classes[sensor_type]
                sensor = sensor_class(
                    device_id=device_id,
                    device_manager=self.dev_manager
                )
                sensors[sensor_id] = {
                    'sensor': sensor,
                    'config': sensor_config,
                    'last_read': 0
                }
                self.logger.info(f"Sensor {sensor_id} erfolgreich initialisiert")
            else:
                self.logger.warning(f"Unbekannter Sensor-Typ: {sensor_type}")
        
        self.logger.info(f"Insgesamt {len(sensors)} Sensoren geladen")
        return sensors

    def connect_to_server(self):
        """Connect to ThingsBoard server"""
        access_token = os.environ.get('RS485_ACCESS_TOKEN')
        if not access_token:
            self.logger.error("RS485_ACCESS_TOKEN nicht gefunden in .envRS485")
            raise ValueError("RS485_ACCESS_TOKEN not found in .envRS485")
            
        server = os.environ.get('RS485_THINGSBOARD_SERVER', 'localhost')
        port = int(os.environ.get('RS485_THINGSBOARD_PORT', 1883))
        
        self.logger.info(f"Verbinde mit ThingsBoard Server: {server}:{port}")
        self.client = TBDeviceMqttClient(server, port, access_token)
        self.client.connect()
        self.logger.info("Erfolgreich mit ThingsBoard verbunden")

    def format_sensor_data(self, sensor_id, sensor_info, sensor_data):
        """Format sensor data according to configuration"""
        config = sensor_info['config']
        formats = config['transmission']['formats']
        formatted_data = {}

        if 'simple' in formats:
            # Einfaches Format (key-value Paare)
            formatted_data['simple'] = {
                f"{sensor_id}_{k}": v for k, v in sensor_data.items()
            }

        if 'json' in formats:
            # JSON Format mit Metadaten
            formatted_data['json'] = {
                f"{sensor_id}_data": {
                    "info": {
                        "name": config['name'],
                        "location": config['location'],
                        "type": config['type'],
                        "device_id": config['device_id']
                    },
                    "metadata": config['metadata'],
                    "measurements": sensor_data,
                    "timestamp": int(time.time() * 1000),
                    "status": "active"
                }
            }

        return formatted_data

    def should_read_sensor(self, sensor_info):
        """Check if sensor should be read based on its interval"""
        current_time = time.time()
        interval = sensor_info['config']['transmission']['interval']
        last_read = sensor_info['last_read']
        
        return (current_time - last_read) >= interval

    def wait_for_bus(self):
        """Wartet bis der RS485-Bus verfügbar ist"""
        with self.bus_lock:
            current_time = time.time()
            time_since_last = current_time - self.last_communication_time
            
            if time_since_last < self.DEBOUNCE_TIME:
                wait_time = self.DEBOUNCE_TIME - time_since_last
                time.sleep(wait_time)
            
            self.last_communication_time = time.time()

    def read_sensor_data(self, sensor_id, sensor_info):
        """Liest Daten von einem Sensor mit Bus-Management"""
        try:
            self.wait_for_bus()  # Warte auf Bus-Verfügbarkeit
            
            sensor = sensor_info['sensor']
            sensor_data = sensor.read_data()
            
            if sensor_data:
                self.logger.debug(f"Sensor {sensor_id} erfolgreich gelesen: {sensor_data}")
                return sensor_data
            else:
                self.logger.error(f"Keine Daten von Sensor {sensor_id} erhalten")
                return None
                
        except Exception as e:
            self.logger.error(f"Fehler beim Lesen von Sensor {sensor_id}: {e}")
            return None

    def run(self):
        """Main run loop"""
        self.logger.info("Starte SensorManager...")
        self.running = True
        error_counts = {}  # Zähler für Fehler pro Sensor
        
        while self.running:
            current_time = time.time()
            
            # Sammle alle Sensoren die gelesen werden müssen
            sensors_to_read = [
                (sensor_id, sensor_info)
                for sensor_id, sensor_info in self.sensors.items()
                if self.should_read_sensor(sensor_info)
            ]
            
            # Verarbeite jeden Sensor mit Fehlerbehandlung
            for sensor_id, sensor_info in sensors_to_read:
                try:
                    # Prüfe ob Sensor zu oft fehlgeschlagen ist
                    if error_counts.get(sensor_id, 0) >= 5:
                        if current_time - sensor_info.get('last_error_log', 0) > 300:
                            self.logger.warning(f"Sensor {sensor_id} temporär deaktiviert wegen zu vieler Fehler")
                            sensor_info['last_error_log'] = current_time
                        continue
                    
                    # Bestimme Wartezeit basierend auf Sensor-Typ und ID
                    device_id = sensor_info['config'].get('device_id', 0)
                    sensor_type = sensor_info['config'].get('type', '')
                    
                    # Längere Pause für Sensoren mit bekannt langen Kabeln
                    is_long_cable = (
                        (sensor_type == 'turbidity' and device_id in [2, 22]) or  # Trübungssensoren
                        (sensor_type == 'flow' and device_id >= 42)               # Flow-Sensoren 3 und 4
                    )
                    
                    if is_long_cable:
                        self.logger.debug(f"Längere Pause für Sensor {sensor_id} mit langem Kabel")
                        time.sleep(1.0)  # 1 Sekunde Pause
                    else:
                        time.sleep(0.5)  # Standard-Pause
                    
                    self.logger.debug(f"Lese Sensor {sensor_id}...")
                    sensor_data = self.read_sensor_data(sensor_id, sensor_info)
                    
                    if sensor_data:
                        # Erfolgreicher Read - Reset Error Counter
                        error_counts[sensor_id] = 0
                        
                        # Format and send data
                        formatted_data = self.format_sensor_data(sensor_id, sensor_info, sensor_data)
                        self.send_telemetry(formatted_data)
                        sensor_info['last_read'] = current_time
                        
                        # Warte nach erfolgreicher Übertragung
                        time.sleep(0.2)
                    else:
                        # Erhöhe Fehlerzähler bei None-Rückgabe
                        error_counts[sensor_id] = error_counts.get(sensor_id, 0) + 1
                        self.logger.warning(f"Keine Daten von Sensor {sensor_id} erhalten (Fehler: {error_counts[sensor_id]})")
                    
                except Exception as e:
                    # Fehlerbehandlung für einzelne Sensoren
                    error_counts[sensor_id] = error_counts.get(sensor_id, 0) + 1
                    self.logger.error(f"Fehler beim Lesen von Sensor {sensor_id} (Fehler: {error_counts[sensor_id]}): {e}")
                    
                    # Sende Fehlerstatus an ThingsBoard wenn möglich
                    try:
                        error_telemetry = {
                            "simple": {
                                f"{sensor_id}_error": str(e),
                                f"{sensor_id}_error_count": error_counts[sensor_id]
                            }
                        }
                        self.send_telemetry(error_telemetry)
                    except:
                        pass  # Ignoriere Fehler beim Senden des Fehlerstatus
                    
                    continue
                
                # Prüfe ob Sensor sich erholt hat
                if error_counts.get(sensor_id, 0) > 0 and sensor_data:
                    self.logger.info(f"Sensor {sensor_id} hat sich erholt nach {error_counts[sensor_id]} Fehlern")
                    error_counts[sensor_id] = 0
            
            # Automatische Reaktivierung von Sensoren nach Fehler
            for sensor_id in list(error_counts.keys()):
                if error_counts[sensor_id] >= 5:
                    sensor_info = self.sensors.get(sensor_id)
                    if not sensor_info:
                        continue
                    
                    # Prüfe ob es ein Sensor mit langem Kabel ist
                    device_id = sensor_info['config'].get('device_id', 0)
                    sensor_type = sensor_info['config'].get('type', '')
                    
                    is_long_cable = (
                        (sensor_type == 'turbidity' and device_id in [2, 22]) or  # Trübungssensoren
                        (sensor_type == 'flow' and device_id >= 42)               # Flow-Sensoren 3 und 4
                    )
                    
                    if is_long_cable:
                        # Kürzere Reaktivierungszeit für Sensoren mit langen Kabeln
                        retry_time = 120  # 2 Minuten
                        if (current_time - sensor_info.get('last_read', 0)) > retry_time:
                            self.logger.info(f"Versuche Sensor {sensor_id} nach 2 Minuten zu reaktivieren (Sensor mit langem Kabel)")
                            error_counts[sensor_id] = 0
                    else:
                        # Progressives Reaktivierungsschema für andere Sensoren
                        # Wenn nicht vorhanden, initialiere Reaktivierungszeiten (in Sekunden)
                        if 'retry_times' not in sensor_info:
                            sensor_info['retry_times'] = [300, 600, 1800, 3600]  # 5min, 10min, 30min, 1h
                            sensor_info['retry_index'] = 0
                        
                        # Hole aktuelle Reaktivierungszeit
                        retry_time = sensor_info['retry_times'][sensor_info['retry_index']]
                        
                        if (current_time - sensor_info.get('last_read', 0)) > retry_time:
                            self.logger.info(f"Versuche Sensor {sensor_id} nach {retry_time/60} Minuten zu reaktivieren")
                            error_counts[sensor_id] = 0
                            
                            # Erhöhe den Index für die nächste Reaktivierungszeit (bleibe aber im gültigen Bereich)
                            sensor_info['retry_index'] = min(sensor_info['retry_index'] + 1, len(sensor_info['retry_times']) - 1)
            
            # Längere Pause am Ende eines Durchlaufs
            time.sleep(1.0)

    def send_telemetry(self, data):
        """Send telemetry data to ThingsBoard with retry"""
        if not self.client or not data:
            return

        max_retries = 3
        retry_delay = 1.0  # Sekunden
        
        for attempt in range(max_retries):
            try:
                # Sende Simple-Format Daten
                if data.get("simple"):
                    self.client.send_telemetry(data["simple"])
                    self.logger.debug("Simple format Telemetrie erfolgreich gesendet")
                    
                # Sende JSON-Format Daten - jetzt einzeln pro Sensor
                if data.get("json"):
                    for sensor_data_key, sensor_data in data["json"].items():
                        self.client.send_telemetry({
                            sensor_data_key: sensor_data
                        })
                        self.logger.debug(f"JSON format Telemetrie für {sensor_data_key} erfolgreich gesendet")
                
                return  # Erfolgreich gesendet, verlasse die Funktion
                
            except Exception as e:
                if attempt < max_retries - 1:
                    self.logger.warning(f"Fehler beim Senden der Telemetrie (Versuch {attempt + 1}): {e}")
                    time.sleep(retry_delay)
                else:
                    self.logger.error(f"Fehler beim Senden der Telemetrie nach {max_retries} Versuchen: {e}")
                
    def stop(self):
        """Stop the sensor manager"""
        self.logger.info("Stoppe SensorManager...")
        self.running = False
        if self.client:
            self.client.disconnect()
        self.logger.info("SensorManager gestoppt")