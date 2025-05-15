# -----------------------------------------------------------------------------
# Company: KARIM Technologies
# Author: Sayed Amir Karim
# Copyright: 2023 KARIM Technologies
#
# License: All Rights Reserved
#
# Module: Modbus Lib V0.5
# Description: Modbus Communication Module
# -----------------------------------------------------------------------------

import struct
import serial
import crcmod.predefined
from threading import Thread
from time import sleep


class ModbusClient:
    def __init__(self, device_manager, device_id):
        self.device_manager = device_manager
        self.device_id = device_id
        self.auto_read_enabled = False

    def read_register(self, start_address, register_count, data_format='>f'):
        return self.device_manager.read_register(self.device_id, start_address, register_count, data_format)

    def read_radar_sensor(self, register_address):
        return self.device_manager.read_radar_sensor(self.device_id, register_address)

    def auto_read_registers(self, start_address, register_count, data_format='>f', interval=1):
        self.auto_read_enabled = True
        def read_loop():
            while self.auto_read_enabled:
                value = self.read_register(start_address, register_count, data_format)
                print(f'Auto Read: {value}')
                sleep(interval)

        Thread(target=read_loop).start()

    def stop_auto_read(self):
        self.auto_read_enabled = False


class DeviceManager:
    def __init__(self, port, baudrate, parity, stopbits, bytesize, timeout):
        self.ser = serial.Serial(
            port=port,
            baudrate=baudrate,
            parity=serial.PARITY_NONE if parity == 'N' else serial.PARITY_EVEN if parity == 'E' else serial.PARITY_ODD,
            stopbits=serial.STOPBITS_ONE if stopbits == 1 else serial.STOPBITS_TWO,
            bytesize=serial.EIGHTBITS if bytesize == 8 else serial.SEVENBITS,
            timeout=timeout
        )
        self.devices = {}
        self.last_read_values = {}  # Dictionary to store last read values for each device and register


    def add_device(self, device_id):
        self.devices[device_id] = ModbusClient(self, device_id)

    def get_device(self, device_id):
        return self.devices.get(device_id)

    
    def read_register(self, device_id, start_address, register_count, data_format):
        function_code = 0x03

        message = struct.pack('>B B H H', device_id, function_code, start_address, register_count)

        crc16 = crcmod.predefined.mkPredefinedCrcFun('modbus')(message)
        message += struct.pack('<H', crc16)

        self.ser.write(message)

        response = self.ser.read(100)
        
        # Check if the response is at least 2 bytes long
        if len(response) < 2:
            print('Received response is shorter than expected')
            return self.last_read_values.get((device_id, start_address), None)

        received_crc = struct.unpack('<H', response[-2:])[0]
        calculated_crc = crcmod.predefined.mkPredefinedCrcFun('modbus')(response[:-2])
        if received_crc != calculated_crc:
            print('CRC error in response')
            return self.last_read_values.get((device_id, start_address), None)

        data = response[3:-2]
        swapped_data = data[2:4] + data[0:2]
        try:
            floating_point = struct.unpack(data_format, swapped_data)[0]
        except struct.error:
            print(f'Error decoding data from device {device_id}')
            return self.last_read_values.get((device_id, start_address), None)

        if floating_point is None:
            print(f'Error reading register from device {device_id}')
            return self.last_read_values.get((device_id, start_address), None)

        # Store the read value in the last_read_values dictionary
        self.last_read_values[(device_id, start_address)] = floating_point

        return floating_point



    def read_radar_sensor(self, device_id, register_address):
        return self.read_register(device_id, register_address, 1, data_format='>H')

