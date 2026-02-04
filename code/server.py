__author__ = "赵博凯"
__license__ = "GPL v3"

import os
import requests
import json

import rich.console
import rich.table
import rich.text

from sys import exit
import load_lang_pack

# 服务器配置
API_SITE = "http://localhost:5000"  # 服务器地址
API_PATH = "fuck"  # API路径
APT_PASSWORD = "fuck"  # 访问密码

# 初始化语言包
lp = load_lang_pack.LanguagePack("server.json", "en")
lp.load()

# 输出函数
def output(*args, type=""):
    console = rich.console.Console()
    if type.strip() == "":
        print(*args)
    else:
        if type == "info":
            console.log(f"[white on blue][*][/white on blue]", *args, style="white")
        elif type == "warning":
            console.log(f"[white on blue][!][/white on blue]", *args, style="white")
        elif type == "error":
            console.log(f"[white on red][-][/white on red]", *args, style="bold red")
        elif type == "success":
            console.log(f"[white on green][+][/white on green]", *args, style="green")
        elif type == "debug":
            console.log(f"[grey50][|][/grey50]", *args, style="grey50")
        else:
            raise ValueError(f"Invalid type: {type}")

# 服务器操作类
class Server:
    @staticmethod
    def device_list():
        """获取并显示设备列表"""
        response = requests.post(f"{API_SITE}/{API_PATH}/function", 
                                data={"func_name": "device_list"}, 
                                cookies=cookie)
        if not response.ok: 
            raise Exception(f"{lp.g('request_failed')}: {response.status_code} {response.json()}")
        result = response.json()
        if type(result) != list:
            output(result, type="info")
            return
        Table = rich.table.Table(title=lp.g("device_list_title"))
        Table.add_column(lp.g("device_id"), justify="center", style="cyan")
        Table.add_column(lp.g("device_ip"), justify="center", style="magenta")
        Table.add_column(lp.g("device_info"), justify="center", style="green")
        for i in result:
            Table.add_row(rich.text.Text(i["id"], overflow="fold"), 
                            rich.text.Text(i["ip"], overflow="fold"), 
                            rich.text.Text(i["systeminfo"], overflow="fold"))
        output(Table, type="info")

    @staticmethod
    def select_device(id):
        """选择要控制的设备"""
        response = requests.post(f"{API_SITE}/{API_PATH}/function", 
                                    data={"func_name": "command", "id": id, "command": "echo hello world"}, 
                                    cookies=cookie)
        if not response.ok: 
            raise Exception(f"{lp.g('request_failed')}: {response.status_code} {response.json()}")
        output(f"{lp.g('selected_device')}: {id}", type="success")
        return id

    @staticmethod   
    def delete_device(id):
        """删除指定设备"""
        response = requests.post(f"{API_SITE}/{API_PATH}/function", 
                                    data={"func_name": "delete", "id": id}, 
                                    cookies=cookie)
        if not response.ok: 
            raise Exception(f"{lp.g('request_failed')}: {response.status_code} {response.json()}")
        output(f"{lp.g('deleted_device')}: {id}", type="success")
        
    @staticmethod
    def systeminfo(id):
        """获取设备系统信息"""
        response = requests.post(f"{API_SITE}/{API_PATH}/function", 
                                data={"func_name": "systeminfo", "id": id}, 
                                cookies=cookie)
        if not response.ok: 
            raise Exception(f"{lp.g('request_failed')}: {response.status_code} {response.json()}")
        system_info = json.loads(response.json()["message"])
        with open("systeminfo.json", "w") as f:
            json.dump(system_info, f, indent=4, ensure_ascii=False)
        rich.print_json(data=system_info)
        output(lp.g("system_info_saved"), type="success")

    @staticmethod
    def command(id):
        """进入设备命令模式"""
        while True:
            command = input(f"(command)<{id}>>")
            if command.strip().lower() == "exit": 
                break
            response = requests.post(f"{API_SITE}/{API_PATH}/function", 
                                     data={"func_name": "command", "id": id, "command": command},
                                     cookies=cookie)
            result = response.json()
            if not response.ok: 
                raise Exception(f"{lp.g('request_failed')}: {response.status_code} {result["error"]}")
            result = json.loads(result["message"])
            for i in result.items():
                output(f"{i[0]}: {i[1]}", type="info")
    
    @staticmethod
    def background(id, command):
        """在设备上后台执行命令"""
        response = requests.post(f"{API_SITE}/{API_PATH}/function", 
                                data={"func_name": "background", "id": id, "command": command}, 
                                cookies=cookie)
        result = response.json()
        if not response.ok: 
            raise Exception(f"{lp.g('request_failed')}: {response.status_code} {result["error"]}")
        if "已发送" in result["message"] or "sent" in result["message"]:
            output(f"{lp.g('command_executed_in_background')}: {command}", type="success")
        else:
            output(f"{lp.g('command_execution_failed')}: {result["message"]}", type="error")
    
    @staticmethod
    def cd(id, directory):
        """切换设备工作目录"""
        response = requests.post(f"{API_SITE}/{API_PATH}/function", 
                                data={"func_name": "change_directory", "id": id, "directory": directory}, 
                                cookies=cookie)
        result = response.json()
        if not response.ok: 
            raise Exception(f"{lp.g('request_failed')}: {response.status_code} {result["error"]}")
        if "successfully" in result["message"]:
            output(f"{lp.g('directory_changed')}: {directory}", type="success")
        else:
            output(f"{lp.g('directory_change_failed')}: {result["message"]}", type="error")
    
    @staticmethod
    def upload(id, local_file_path, target_file_path):
        """上传文件到设备"""
        output(lp.g("uploading_file"), type="info")
        try:
            with open(local_file_path, "rb") as f:
                response = requests.post(f"{API_SITE}/{API_PATH}/function", 
                        data={"func_name": "upload" ,"id": id, "path": target_file_path}, 
                        files={"file": f},
                        cookies=cookie)
        except Exception as e:
            output(f"{lp.g('file_upload_failed')}: {e}", type="error")
            return
        
        result = response.json()
        if not response.ok:
            raise Exception(f"{lp.g('request_failed')}: {response.status_code} {result['error']}")
        if "ok" in result["message"]:
            output(f"{lp.g('file_uploaded')}: {local_file_path} -> {target_file_path}", type="success")
        else:
            output(f"{lp.g('file_upload_failed')}: {result['message']}", type="error")

    @staticmethod
    def download(id, target_file_path, local_file_path):
        ''"下载设备上的文件"""
        output(lp.g("downloading_file"), type="info")
        response = requests.post(f"{API_SITE}/{API_PATH}/function", 
                                data={"func_name": "download", "id": id, "path": target_file_path}, 
                                cookies=cookie)
        result = response.json()
        if not response.ok: 
            raise Exception(f"{lp.g('request_failed')}: {response.status_code} {result["error"]}")
        result = result["message"]
        if len(result) != 64 or not result.isalpha():
            output(f"{lp.g('file_download_failed')}: {result}", type="error")
            return
        response = requests.get(f"{API_SITE}/{API_PATH}/download/{result}", cookies=cookie)
        if not response.ok: 
            raise Exception(f"{lp.g('request_failed')}: {response.status_code}")
        try:
            with open(local_file_path, "wb") as f:
                f.write(response.content)
        except Exception as e:
            output(f"{lp.g('file_download_failed')}: {e}", type="error")
            return
        output(f"{lp.g('file_downloaded')}: {target_file_path} -> {local_file_path}", type="success")


def command_input():
    """命令行交互主循环"""
    select_device = None
    while True:
        try:
            # 设备控制模式
            if not select_device is None:
                command = input(f"(console)<{select_device}>>")
                command = command.strip()
                match command:
                    case "": pass
                    case "help": 
                        output(lp.g("console_help_info"), type="info")
                    case "back": 
                        select_device = None
                    case "clear": 
                        os.system("cls") if os.name == "nt" else os.system("clear")
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
                    case command if command.startswith("upload "):
                        file_info = command.split(" ", 1)[1]
                        file_info = file_info.split("(*.*)")
                        if len(file_info) != 2:
                            output(f"{lp.g('invalid_file_info')}", type="error")
                            continue
                        local_file_path = file_info[0]
                        target_file_path = file_info[1]
                        Server.upload(select_device, local_file_path, target_file_path)
                    case command if command.startswith("download "):
                        file_info = command.split(" ", 1)[1]
                        file_info = file_info.split("(*.*)")
                        if len(file_info) != 2:
                            output(f"{lp.g('invalid_file_info')}", type="error")
                            continue
                        target_file_path = file_info[0]
                        local_file_path = file_info[1]
                        Server.download(select_device, target_file_path, local_file_path)
                    case _: 
                        output(f"{lp.g('unknown_command')}: {command}", type="error")
                continue

            # 服务器控制模式
            command = input("(server)>")
            command = command.strip().lower()
            match command:
                case "": pass
                case "exit": 
                    exit(0)
                case "help":
                    output(lp.g("server_help_info"), type="info")
                case "clear": 
                    os.system("cls") if os.name == "nt" else os.system("clear")
                case "about": 
                    output(lp.g("about_info"), type="info")
                case "list": 
                    Server.device_list()
                case command if command.startswith("select "): 
                    select_device = Server.select_device(command.split(" ", 1)[1])
                case command if command.startswith("delete "): 
                    Server.delete_device(command.split(" ", 1)[1])
                case _: 
                    output(f"{lp.g('unknown_command')}: {command}", type="error")
        except Exception as e:
            output(f"{lp.g('error_occurred')}: {type(e).__name__}: {e}", type="error")

if __name__ == "__main__":
    # 程序入口
    output(lp.g("copyright"), type="info")
    output(lp.g("program_starting"), type="info")
    output(lp.g("verifying_password"), type="info")
    try:
        # 验证服务器密码
        response = requests.post(f"{API_SITE}/{API_PATH}/verify", json={"password": APT_PASSWORD})
        if response.status_code == 200:
            cookie = response.json()
            output(f"{lp.g('verification_successful')}: {cookie}", type="success")
        else:
            output(f"{lp.g('verification_failed')}: {response.status_code} {response.json()}", type="error")
            exit(1)
    except Exception as e:
        output(f"{lp.g('verification_failed')}: {type(e).__name__}: {e}", type="error")
        exit(1)

    try:
        # 启动命令行交互
        command_input()
    except Exception as e:
        output(f"{lp.g('error_occurred')}: {type(e).__name__}: {e}", type="error")
    except KeyboardInterrupt:
        output(lp.g("user_interrupted"), type="warning")
        exit(1)
