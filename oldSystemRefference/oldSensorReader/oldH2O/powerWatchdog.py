import requests
import subprocess
import time
import signal
from periphery import GPIO
import threading
import json
from tb_gateway_mqtt import TBDeviceMqttClient
import sys
import os
import re

# Configuration
SERVER_URL = 'http://localhost:8080'
MAIN_SCRIPT_PATH = '/home/owipex_adm/owipex-sps/h2o.py'
BUTTON_PIN = 9  # GPIO-Pin für den Button
LED_PINS = {'R': 5, 'G': 6, 'B': 26}  # Angenommene Pins für die LEDs
CHECK_INTERVAL = 10
THINGSBOARD_SERVER = 'localhost'
THINGSBOARD_PORT = 1883  # Ensure the port is an integer
ACCESS_TOKEN = '3VuMh3c5TfbpagAO4Ndr'
DATA_SEND_INTERVAL = 10  # Data send interval in seconds

# Initialisierung
try:
    button_gpio = GPIO(BUTTON_PIN, "in")
    leds = {color: GPIO(pin, "out") for color, pin in LED_PINS.items()}
except Exception as e:
    print(f"Error initializing GPIOs: {e}")
    sys.exit(1)

main_process = None

# Initialize ThingsBoard MQTT client
tb_client = TBDeviceMqttClient(THINGSBOARD_SERVER, username=ACCESS_TOKEN)

# Status variables
is_main_script_running = False  # Statusvariable für den Hauptskript-Zustand
manually_stopped = False  # Gibt an, ob das Skript manuell gestoppt wurde
script_status = {'ScriptRunning': True}

cleanup_done = False  # Flag to ensure cleanup is only done once
stop_event = threading.Event()  # Event to signal threads to stop

last_send_time = time.time() - DATA_SEND_INTERVAL

def set_led_color(color):
    try:
        for led in leds.values():
            if led._fd is not None:
                led.write(True)  # Alle LEDs ausschalten
        if color in leds and leds[color]._fd is not None:
            leds[color].write(False)  # Gewünschte LED einschalten
    except Exception as e:
        print(f"Error setting LED color: {e}")

def check_server_availability():
    try:
        response = requests.get(SERVER_URL)
        return response.status_code == 200
    except requests.exceptions.RequestException:
        return False

def start_main_script():
    global main_process, is_main_script_running, script_status, manually_stopped
    if not is_main_script_running and check_server_availability():
        main_process = subprocess.Popen(['python3', MAIN_SCRIPT_PATH])
        if main_process.poll() is None:
            is_main_script_running = True
            manually_stopped = False
            script_status['ScriptRunning'] = True
            set_led_color('G')
            tb_client.send_attributes(script_status)

def stop_main_script():
    global main_process, is_main_script_running, manually_stopped, script_status
    if is_main_script_running and main_process:
        main_process.send_signal(signal.SIGINT)
        main_process.wait()
        main_process = None
        is_main_script_running = False
        manually_stopped = True
        script_status['ScriptRunning'] = False
        set_led_color('B')
        tb_client.send_attributes(script_status)

def rpc_callback(id, request_body):
    global is_main_script_running, script_status, tb_client
    command = request_body.get('method')
    if command == 'setScriptRunning':
        params = request_body.get('params')
        if params:
            start_main_script()
        else:
            stop_main_script()

def button_press_handler():
    global manually_stopped
    press_time = None
    while not stop_event.is_set():
        try:
            if button_gpio.read() == False:
                press_time = time.time() if press_time is None else press_time
            else:
                if press_time is not None:
                    elapsed_time = time.time() - press_time
                    if 2 < elapsed_time <= 5:
                        start_main_script() if not is_main_script_running else None
                    elif elapsed_time > 5:
                        stop_main_script() if is_main_script_running else None
                    press_time = None
        except Exception as e:
            print(f"Error in button_press_handler: {e}")
        time.sleep(0.1)

def monitor_system():
    global manually_stopped, script_status
    while not stop_event.is_set():
        try:
            if check_server_availability():
                start_main_script() if not is_main_script_running and not manually_stopped else None
            else:
                set_led_color('R')
        except Exception as e:
            print(f"Error in monitor_system: {e}")
        time.sleep(CHECK_INTERVAL)

def cleanup():
    global cleanup_done
    if cleanup_done:
        return
    cleanup_done = True
    try:
        for gpio in leds.values():
            if gpio._fd is not None:
                gpio.write(True)  # LEDs ausschalten
                gpio.close()
        if button_gpio._fd is not None:
            button_gpio.close()
    except Exception as e:
        print(f"Error during cleanup: {e}")

def signal_handler(sig, frame):
    stop_event.set()  # Signal all threads to stop
    cleanup()
    tb_client.disconnect()
    if main_process:
        main_process.terminate()
    sys.exit(0)

def read_cpu_temperature():
    try:
        path = '/sys/class/thermal/thermal_zone0/temp'
        with open(path, 'r') as file:
            temp = int(file.read().strip()) / 1000.0
            return temp
    except FileNotFoundError:
        return None  # No temperature found

def get_mobile_signal():
    try:
        result = subprocess.check_output(['mmcli', '-m', '0', '--signal-get'], text=True)
        
        rssi_match = re.search(r'rssi: (-?\d+.\d+) dBm', result)
        rsrq_match = re.search(r'rsrq: (-?\d+.\d+) dB', result)
        rsrp_match = re.search(r'rsrp: (-?\d+.\d+) dBm', result)
        snr_match = re.search(r's/n: (\d+.\d+) dB', result)

        rssi = float(rssi_match.group(1)) if rssi_match else -999.0
        rsrq = float(rsrq_match.group(1)) if rsrq_match else -999.0
        rsrp = float(rsrp_match.group(1)) if rsrp_match else -999.0
        snr = float(snr_match.group(1)) if snr_match else -999.0

        return rssi, rsrq, rsrp, snr
    except subprocess.CalledProcessError as e:
        print("Error retrieving signal information: ", e)
        return -999.0, -999.0, -999.0, -999.0

def get_data():
    cpu_usage = round(float(os.popen("grep 'cpu ' /proc/stat | awk '{usage=($2+$4)*100/($2+$4+$5)} END {print usage }'").readline().strip().replace(',', '.')), 2)
    ip_address = os.popen('hostname -I').readline().strip()
    mac_address = os.popen('cat /sys/class/net/*/address').readline().strip()
    processes_count = os.popen('ps -Al | grep -c bash').readline().strip()
    
    swap_mem = os.popen("free -m | grep Swap | awk '{print ($3/$2)*100}'").readline().strip().replace(',', '.')
    swap_memory_usage = round(float(swap_mem) if swap_mem else 0, 2)
    
    ram_mem = os.popen("free -m | grep -E 'Mem|Speicher' | awk '{print ($3/$2) * 100}'").readline().strip().replace(',', '.')
    ram_usage = round(float(ram_mem) if ram_mem else 0, 2)
    
    st = os.statvfs('/')
    used = (st.f_blocks - st.f_bfree) * st.f_frsize
    boot_time = os.popen('uptime -p').read().strip()
    avg_load = round((cpu_usage + ram_usage) / 2, 2)
    cpu_temperature = read_cpu_temperature()

    rssi, rsrq, rsrp, snr = get_mobile_signal()

    attributes = {
        'ip_address': ip_address,
        'mac_address': mac_address,
        'cpu_usage': cpu_usage,
        'processes_count': processes_count,
        'disk_usage': used,
        'RAM_usage': ram_usage,
        'swap_memory_usage': swap_memory_usage,
        'boot_time': boot_time,
        'avg_load': avg_load,
        'cpu_temperature': cpu_temperature,
        'rssi': rssi,
        'rsrq': rsrq,
        'rsrp': rsrp,
        'snr': snr
    }
    return attributes

# Set up signal handlers
signal.signal(signal.SIGINT, signal_handler)
signal.signal(signal.SIGTERM, signal_handler)

# Set up ThingsBoard MQTT client
tb_client.set_server_side_rpc_request_handler(rpc_callback)
tb_client.connect()

# Run threads
try:
    monitor_system_thread = threading.Thread(target=monitor_system)
    monitor_system_thread.start()
    button_press_thread = threading.Thread(target=button_press_handler)
    button_press_thread.start()
    
    while not stop_event.is_set():
        current_time = time.time()
        if current_time - last_send_time >= DATA_SEND_INTERVAL:
            data = get_data()
            tb_client.send_attributes(data)
            last_send_time = current_time
        time.sleep(0.1)
finally:
    cleanup()
    tb_client.disconnect()
