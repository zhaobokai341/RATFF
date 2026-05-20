__author__ = "赵博凯"
__license__ = "MIT"

import json
import os
import load_lang_pack
import requests
import rich.console
import rich.table
import rich.text
import shlex

from sys import exit  # pylint: disable=redefined-builtin
from requests.exceptions import HTTPError

# 服务器配置
LANGUAGE = "zh"  # 语言
API_SITE = "http://127.0.0.1:5000"  # 服务器地址
API_PATH = "fuck"  # API路径
APT_PASSWORD = "fuck"  # 访问密码
REQUEST_TIMEOUT = 10  # 请求超时时间

# 初始化语言包
lp = load_lang_pack.LanguagePack("server.json", LANGUAGE)
lp.load()


# 服务器操作类
class Server:
    def __init__(self, device_id: str, cookie: dict):
        self.id = device_id
        self.cookie = cookie

        def request_post(data: dict, files: dict = None) -> requests.Response:
            response = requests.post(
                f"{API_SITE}/{API_PATH}/function",
                data=data,
                cookies=self.cookie,
                timeout=REQUEST_TIMEOUT,
                files=files,
            )
            if not response.ok:
                raise HTTPError(
                    f"{lp.g('request_failed')}: {response.status_code} {response.json()}"
                )
            return response

        self.request_post = request_post

    def device_list(self) -> list[dict] | str:
        """获取设备列表"""
        response = self.request_post({"func_name": "device_list"})
        return response.json()

    def select_device(self) -> str:
        """选择要控制的设备"""
        response = self.request_post({"func_name": "device_list"})
        result = response.json()
        if not isinstance(result, list):
            raise TypeError(result)
        for device in result:
            if (
                self.id in device["id"]
                or self.id in device["ip"]
                or self.id in device["systeminfo"]
            ):
                return device["id"]
        raise ValueError(f"{lp.g('device_not_found')}: {self.id}")

    def delete_device(self):
        """删除指定设备"""
        self.request_post({"func_name": "delete", "id": self.id})

    def systeminfo(self) -> str:
        """获取设备系统信息"""
        response = self.request_post({"func_name": "systeminfo", "id": self.id})
        return response.json()["message"]

    def ls(self) -> str:
        """列出设备当前目录文件"""
        response = self.request_post({"func_name": "get_list_file", "id": self.id})
        result = response.json()
        return result["message"]

    def pwd(self) -> str:
        """获取设备当前工作目录"""
        response = self.request_post({"func_name": "get_pwd", "id": self.id})
        result = response.json()
        return result["message"]

    def delete(self, file_path: str) -> str:
        """删除设备上的文件"""
        response = self.request_post(
            {"func_name": "delete_file", "id": self.id, "path": file_path}
        )
        return response.json()

    def move(self, old_path: str, new_path: str) -> str:
        """移动该设备上的文件"""
        response = self.request_post(
            {
                "func_name": "move_file",
                "id": self.id,
                "old_path": old_path,
                "new_path": new_path,
            }
        )
        return response.json()

    def copy_file(self, source_path: str, target_path: str) -> str:
        """复制文件"""
        response = self.request_post(
            {
                "func_name": "copy_file",
                "id": self.id,
                "old_path": source_path,
                "new_path": target_path,
            }
        )
        return response.json()

    def compress(self, source_path: str, target_path: str) -> str:
        """压缩文件"""
        response = self.request_post(
            {
                "func_name": "compress",
                "id": self.id,
                "source_path": source_path,
                "target_path": target_path,
            }
        )
        return response.json()

    def extract(self, source_path: str, target_path: str) -> str:
        """解压文件"""
        response = self.request_post(
            {
                "func_name": "extract",
                "id": self.id,
                "source_path": source_path,
                "target_path": target_path,
            }
        )
        return response.json()

    def command(self, cmd: str) -> str:
        """执行设备命令"""
        response = self.request_post(
            {"func_name": "command", "id": self.id, "command": cmd}
        )
        return response.json()["message"]

    def background(self, command: str) -> str:
        """在设备上后台执行命令"""
        response = self.request_post(
            {"func_name": "background", "id": self.id, "command": command}
        )
        return response.json()

    def mkdir(self, directory: str) -> str:
        """在设备上创建目录"""
        response = self.request_post(
            {"func_name": "create_directory", "id": self.id, "path": directory}
        )
        return response.json()

    def cd(self, directory: str) -> str:
        """切换设备工作目录"""
        response = self.request_post(
            {"func_name": "change_directory", "id": self.id, "directory": directory}
        )
        return response.json()

    def upload(self, local_file_path: str, target_file_path: str) -> str:
        """上传文件到设备"""
        try:
            with open(local_file_path, "rb") as f:
                response = self.request_post(
                    data={
                        "func_name": "upload",
                        "id": self.id,
                        "path": target_file_path,
                    },
                    files={"file": f},
                )
        except Exception as e:
            raise IOError(f"{lp.g('file_upload_failed')}: {e}") from e

        return response.json()

    def download(self, target_file_path: str, local_file_path: str) -> str:
        """下载设备上的文件"""
        response = self.request_post(
            {"func_name": "download", "id": self.id, "path": target_file_path}
        )
        result = response.json()
        if not response.ok:
            raise HTTPError(
                f"{lp.g('request_failed')}: {response.status_code} {result['error']}"
            )
        result = result["message"]
        if len(result) != 64 or not result.isalpha():
            raise IOError(f"{lp.g('file_download_failed')}: {result}")
        response = requests.get(
            f"{API_SITE}/{API_PATH}/download/{result}",
            cookies=self.cookie,
            timeout=REQUEST_TIMEOUT,
        )
        if not response.ok:
            raise HTTPError(f"{lp.g('request_failed')}: {response.status_code}")
        try:
            with open(local_file_path, "wb") as f:
                f.write(response.content)
        except Exception as e:
            raise HTTPError(f"{lp.g('file_download_failed')}: {e}") from e


# 用户输入类
class CommandInput:
    def __init__(self, cookie):
        self.select_device = None
        self.cookie = cookie

        def command_parse(command: str) -> tuple[str, list[str]]:
            command = shlex.split(command)
            return command[0].lower(), command[1:]

        self.command_parse = command_parse

    def command_input(self):
        """命令行交互主循环"""
        while True:
            try:
                server = Server(self.select_device, self.cookie)
                # 设备控制模式
                if self.select_device is not None:
                    server = Server(self.select_device, self.cookie)
                    while True:
                        command = input(f"(console)<{self.select_device}>>")
                        command = command.strip()
                        command, args = self.command_parse(command)
                        if self.console(command, args, server) == "break":
                            break

                # 服务器控制模式
                server = Server("", self.cookie)
                while True:
                    command = input("(server)>")
                    command = command.strip()
                    command, args = self.command_parse(command)
                    if self.server_(command, args) == "break":
                        break
            except Exception as e:
                output(
                    f"{lp.g('error_occurred')}: {type(e).__name__}: {e}", type_="error"
                )

    def server_(self, command: str, args: tuple[str, ...]) -> str | None:
        """服务端模式
        服务端模式指的是未选择设备时使用的模式，主要用于选择，查看和管理设备
        """
        server = Server("", self.cookie)
        match command:
            case "":
                return
            case "exit":
                exit(0)
            case "help":
                help_list = lp.g("server_help_info")
                for help_text in help_list:
                    output(help_text, type_="info")
            case "clear":
                if os.name == "nt":
                    os.system("cls")
                else:
                    os.system("clear")
            case "about":
                output(lp.g("about_info"), type_="info")
            case "list":
                result = server.device_list()
                if not isinstance(result, list):
                    output(result, type_="info")
                else:
                    Table = rich.table.Table(title=lp.g("device_list_title"))
                    Table.add_column(lp.g("device_id"), justify="center", style="cyan")
                    Table.add_column(
                        lp.g("device_ip"), justify="center", style="magenta"
                    )
                    Table.add_column(
                        lp.g("device_info"), justify="center", style="green"
                    )
                    for i in result:
                        Table.add_row(
                            rich.text.Text(i["id"], overflow="fold"),
                            rich.text.Text(i["ip"], overflow="fold"),
                            rich.text.Text(i["systeminfo"], overflow="fold"),
                        )
                    output(Table, type_="info")
            case "select":
                if not args:
                    output(lp.g("no_device_id_specified"), type_="error")
                    return
                ready_select_id = args[0]
                server = Server(ready_select_id, self.cookie)
                self.select_device = server.select_device()
                output(
                    f"{lp.g('selected_device')}: {self.select_device}", type_="success"
                )
                return "break"
            case "delete":
                if not args:
                    output(lp.g("no_device_id_specified"), type_="error")
                    return
                ready_delete_id = args[0]
                server = Server(ready_delete_id, self.cookie)
                server.delete_device()
                output(f"{lp.g('deleted_device')}: {ready_delete_id}", type_="success")
            case _:
                output(f"{lp.g('unknown_command')}: {command}", type_="error")

    def console(
        self, command: str, args: tuple[str, ...], server: Server
    ) -> str | None:
        """控制台模式
        控制台模式指的是已选择设备时使用的模式，主要用于控制设备
        """
        match command:
            case "":
                return
            case "help":
                help_list = lp.g("console_help_info")
                for help_text in help_list:
                    output(help_text, type_="info")
            case "back":
                self.select_device = None
                return "break"
            case "clear":
                if os.name == "nt":
                    os.system("cls")
                else:
                    os.system("clear")
            case "list":
                result = server.device_list()
                if not isinstance(result, list):
                    output(result, type_="info")
                else:
                    Table = rich.table.Table(title=lp.g("device_list_title"))
                    Table.add_column(lp.g("device_id"), justify="center", style="cyan")
                    Table.add_column(
                        lp.g("device_ip"), justify="center", style="magenta"
                    )
                    Table.add_column(
                        lp.g("device_info"), justify="center", style="green"
                    )
                    for i in result:
                        Table.add_row(
                            rich.text.Text(i["id"], overflow="fold"),
                            rich.text.Text(i["ip"], overflow="fold"),
                            rich.text.Text(i["systeminfo"], overflow="fold"),
                        )
                    output(Table, type_="info")
            case "systeminfo":
                system_info = server.systeminfo()
                with open("systeminfo.json", "w", encoding="utf-8") as f:
                    json.dump(system_info, f, indent=4, ensure_ascii=False)
                rich.print_json(data=system_info)
                output(lp.g("system_info_saved"), type_="success")
            case "ls":
                result = server.ls()
                rich.print_json(data=result)
            case "pwd":
                result = server.pwd()
                output(result, type_="info")
            case "select":
                if not args:
                    output(lp.g("no_device_id_specified"), type_="error")
                    return
                ready_select_id = args[0]
                server = Server(ready_select_id, self.cookie)
                self.select_device = server.select_device()
                output(
                    f"{lp.g('selected_device')}: {self.select_device}", type_="success"
                )
            case "rm":
                if not args:
                    output(lp.g("no_file_name_specified"), type_="error")
                    return
                file_name = args[0]
                result = server.delete(file_name)
                if result["message"] == "ok":
                    output(f"{lp.g('deleted_file')}: {file_name}", type_="success")
                else:
                    output(
                        f"{lp.g('delete_file_failed')}: {result['message']}",
                        type_="error",
                    )
            case "command":
                while True:
                    cmd = input(f"(command)<{self.select_device}>>")
                    if cmd.strip().lower() == "exit":
                        break
                    if cmd.strip().lower() == "":
                        continue
                    result = server.command(cmd)
                    for i in result.items():
                        output(f"{i[0]}: {i[1]}", type_="info")
            case "bg":
                if not args:
                    output(lp.g("no_command_specified"), type_="error")
                    return
                bg_command = args[0]
                result = server.background(bg_command)
                if "ok" in result["message"]:
                    output(
                        f"{lp.g('command_executed_in_background')}: {bg_command}",
                        type_="success",
                    )
                else:
                    output(
                        f"{lp.g('command_execution_failed')}: {result['message']}",
                        type_="error",
                    )
            case "cd":
                if not args:
                    output(lp.g("no_directory_specified"), type_="error")
                    return
                directory = args[0]
                result = server.cd(directory)
                if result["message"] == "ok":
                    output(
                        f"{lp.g('directory_changed')}: {bg_command}", type_="success"
                    )
                else:
                    output(
                        f"{lp.g('directory_change_failed')}: {result['message']}",
                        type_="error",
                    )
            case "mkdir":
                if not args:
                    output(lp.g("no_directory_specified"), type_="error")
                    return
                directory = args[0]
                result = server.mkdir(directory)
                if result["message"] == "ok":
                    output(f"{lp.g('directory_created')}: {directory}", type_="success")
                else:
                    output(
                        f"{lp.g('directory_creation_failed')}: {result['message']}",
                        type_="error",
                    )
            case "compress":
                if len(args) < 2:
                    output(f"{lp.g('invalid_file_info')}", type_="error")
                    return
                source_path = args[0]
                target_path = args[1]
                result = server.compress(source_path, target_path)
                if result["message"] == "ok":
                    output(
                        f"{lp.g('compressed_file')}: {source_path} -> {target_path}",
                        type_="success",
                    )
                else:
                    output(
                        f"{lp.g('compress_file_failed')}: {result['message']}",
                        type_="error",
                    )
            case "extract":
                if len(args) < 2:
                    output(f"{lp.g('invalid_file_info')}", type_="error")
                    return
                source_path = args[0]
                target_path = args[1]
                result = server.extract(source_path, target_path)
                if result["message"] == "ok":
                    output(
                        f"{lp.g('extracted_file')}: {source_path} -> {target_path}",
                        type_="success",
                    )
                else:
                    output(
                        f"{lp.g('extract_file_failed')}: {result['message']}",
                        type_="error",
                    )
            case "mv":
                if len(args) < 2:
                    output(f"{lp.g('invalid_file_info')}", type_="error")
                    return
                old_name = args[0]
                new_name = args[1]
                result = server.move(old_name, new_name)
                if result["message"] == "ok":
                    output(
                        f"{lp.g('moved_file')}: {old_name} -> {new_name}",
                        type_="success",
                    )
                else:
                    output(
                        f"{lp.g('move_file_failed')}: {result['message']}",
                        type_="error",
                    )
            case "cp":
                if len(args) < 2:
                    output(f"{lp.g('invalid_file_info')}", type_="error")
                    return
                source_path = args[0]
                target_path = args[1]
                result = server.copy_file(source_path, target_path)
                if result["message"] == "ok":
                    output(
                        f"{lp.g('copied_file')}: {source_path} -> {target_path}",
                        type_="success",
                    )
                else:
                    output(
                        f"{lp.g('copy_file_failed')}: {result['message']}",
                        type_="error",
                    )
            case "upload":
                if len(args) < 2:
                    output(f"{lp.g('invalid_file_info')}", type_="error")
                    return
                local_file_path = args[0]
                target_file_path = args[1]
                output(lp.g("uploading_file"), type_="info")
                try:
                    result = server.upload(local_file_path, target_file_path)
                    if "ok" in result["message"]:
                        output(
                            f"{lp.g('file_uploaded')}: {local_file_path} -> {target_file_path}",
                            type_="success",
                        )
                    else:
                        output(
                            f"{lp.g('file_upload_failed')}: {result['message']}",
                            type_="error",
                        )
                except Exception as e:
                    output(str(e), type_="error")
            case "download":
                if len(args) < 2:
                    output(f"{lp.g('invalid_file_info')}", type_="error")
                    return
                target_file_path = args[0]
                local_file_path = args[1]
                output(lp.g("downloading_file"), type_="info")
                try:
                    server.download(target_file_path, local_file_path)
                    output(
                        f"{lp.g('file_downloaded')}: {target_file_path} -> {local_file_path}",
                        type_="success",
                    )
                except Exception as e:
                    output(str(e), type_="error")
            case _:
                output(f"{lp.g('unknown_command')}: {command}", type_="error")


# 日志美化函数
def output(*args, type_: str = ""):
    console = rich.console.Console()
    if type_.strip() == "":
        print(*args)
    else:
        if type_ == "info":
            console.log("[white on blue][*][/white on blue]", *args, style="white")
        elif type_ == "warning":
            console.log("[white on blue][!][/white on blue]", *args, style="white")
        elif type_ == "error":
            console.log("[white on red][-][/white on red]", *args, style="bold red")
        elif type_ == "success":
            console.log("[white on green][+][/white on green]", *args, style="green")
        elif type_ == "debug":
            console.log("[grey50][|][/grey50]", *args, style="grey50")
        else:
            raise ValueError(f"Invalid type: {type_}")


if __name__ == "__main__":
    # 程序入口
    output(lp.g("copyright"), type_="info")
    output(lp.g("program_starting"), type_="info")
    output(lp.g("verifying_password"), type_="info")
    try:
        # 验证服务器密码
        response = requests.post(
            f"{API_SITE}/{API_PATH}/verify",
            json={"password": APT_PASSWORD},
            timeout=REQUEST_TIMEOUT,
        )
        if response.status_code == 200:
            cookie = response.json()
            output(f"{lp.g('verification_successful')}: {cookie}", type_="success")
        else:
            output(
                f"{lp.g('verification_failed')}: {response.status_code} {response.json()}",
                type_="error",
            )
            exit(1)
    except Exception as e:
        output(f"{lp.g('verification_failed')}: {type(e).__name__}: {e}", type_="error")
        exit(1)

    try:
        # 启动命令行交互
        CommandInput(cookie).command_input()
    except Exception as e:
        output(f"{lp.g('error_occurred')}: {type(e).__name__}: {e}", type_="error")
    except KeyboardInterrupt:
        output(lp.g("user_interrupted"), type_="warning")
        exit(1)
