__author__ = "赵博凯"
__license__ = "GPL v3"

# TODO: upload, download不要定义特殊字符串

from collections import defaultdict
from quart import Quart, request, jsonify, send_file
from sys import exit  # pylint: disable=redefined-builtin

import load_lang_pack

import asyncio
import base64
import bcrypt
import logging
import os
import random
import rich
import json
import ssl
import typing
import websockets

import rich.traceback
import rich.logging

# 配置Rich的回溯追踪
rich.traceback.install(show_locals=True)

# 日志配置
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(levelname)s - %(message)s",
    handlers=[rich.logging.RichHandler()],
)

# 服务器配置
HOST = "0.0.0.0" # 服务器地址
PORT = 8765 # 服务器端口
WEB_HOST = "0.0.0.0" # Web服务器监听地址
WEB_PORT = 5000 # Web服务器监听端口
SSL_CERT = "cert.pem" # SSL证书
SSL_KEY = "key.pem" # SSL密钥
LANGUAGE = "zh" # 语言
SECURITY_PATH = "fuck" # 安全路径
SECURITY_PASSWORD_HASH = b"$2b$04$T8NZ.WUIuO05WyVpLrQYiOdgqc2zbx7E9ysF03696dYvwGohCFzwC" # 安全密码哈希

# 全局变量
app = Quart(__name__)
app.config["MAX_CONTENT_LENGTH"] = None
app.config["BODY_TIMEOUT"] = None
control_list = {}
device_locks = defaultdict(asyncio.Lock)
download_directory_name = "download"
lp = load_lang_pack.LanguagePack("server_api.json", LANGUAGE)
lp.load()


# 服务器核心类
class Server:
    def __init__(self):
        logging.info(lp.g("server_init"))

    def about(self):
        '''获取关于信息'''
        logging.info(lp.g("about"))
        return lp.g("about")

    async def client_list(self) -> tuple[str, list | str]:
        '''获取客户端列表'''
        logging.info(lp.g("getting_client_list"))
        if len(control_list) == 0:
            logging.info(lp.g("no_connected_devices"))
            return ("error", lp.g("no_devices_connected"))
        else:
            devices = []
            for device in control_list.items():
                devices.append(
                    {
                        "id": device[0],
                        "ip": device[1]["ip"],
                        "systeminfo": device[1]["systeminfo"],
                    }
                )
            logging.info("%s: %s", lp.g("device_list"), devices)
            return ("success", devices)

    async def delete(self, device_id) -> str:
        '''删除设备'''
        logging.info("%s: %s", lp.g("deleting_device"), device_id)
        if device_id in control_list:
            websocket_ = control_list[device_id]["websocket"]
            try:
                # 使客户端也能退出
                await websocket_.send("exit")
                control_list.pop(device_id)
                logging.info("%s %s", lp.g("device_deleted_successfully"), device_id)
                return f"{lp.g('successfully_deleted_device')}{device_id}。"
            except Exception as e:
                control_list.pop(device_id)
                logging.error(
                    "%s %s: %s", lp.g("failed_to_delete_device"), device_id, str(e)
                )
                return (
                    f"{lp.g('exception_occurred_while_disconnecting')}{device_id}: {e}"
                )
        else:
            logging.error("%s %s", lp.g("device_does_not_exist"), device_id)
            return f"{lp.g('device_with_id_does_not_exist')}{device_id}。"


# 客户端控制类
class ControlClient:
    def __init__(self, device_id: str):
        logging.info("%s: %s", lp.g("initializing_client_controller"), device_id)

        async def ws_send_and_receive(command: str, load_json: bool = False) -> str | dict:
            await self.websocket.send(command)
            result = await self.websocket.recv()
            if load_json:
                try:
                    result = json.loads(result)
                except json.JSONDecodeError:
                    logging.warning("%s: %s", lp.g("json_decode_error"), result)
            return result

        self.id = device_id
        self.websocket = control_list[device_id]["websocket"]
        self.ws_send_and_receive = ws_send_and_receive

    async def system_info(self) -> dict:
        '''获取设备系统信息'''
        logging.info(lp.g("getting_device_system_information"))
        try:
            result = await self.ws_send_and_receive("systeminfo", load_json=True)
            logging.info("%s: %s", lp.g("system_information"), json.dumps(result))
            return result
        except Exception as e:
            logging.error("%s: %s", lp.g("failed_to_get_system_information"), str(e))
            raise

    async def delete_file(self, file: str) -> str:
        '''删除文件'''
        logging.info("%s: %s", lp.g("deleting_file"), file)
        try:
            result = await self.ws_send_and_receive(f"delete:{file}")
            if result != "ok":
                raise RuntimeError(result)
            logging.info("%s: %s", lp.g("file_deletion_result"), result)
            return result
        except Exception as e:
            logging.error("%s: %s", lp.g("failed_to_delete_file"), str(e))
            raise

    async def move_file(self, source: str, target: str) -> str:
        '''移动文件'''
        logging.info("%s: %s -> %s", lp.g("moving_file"), source, target)
        try:
            result = await self.ws_send_and_receive(f"mv:{source}(*.*){target}")
            if result != "ok":
                raise RuntimeError(result)
            logging.info("%s: %s", lp.g("file_move_result"), result)
            return result
        except Exception as e:
            logging.error("%s: %s", lp.g("failed_to_move_file"), str(e))
            raise

    async def compress_file(self, file: str, target_file: str) -> str:
        """压缩文件"""
        logging.info("%s: %s -> %s", lp.g("compressing_file"), file, target_file)
        try:
            result = await self.ws_send_and_receive(
                f"compress:{file}(*.*){target_file}"
            )
            if result != "ok":
                raise RuntimeError(result)
            logging.info("%s: %s", lp.g("file_compression_result"), result)
            return result
        except Exception as e:
            logging.error("%s: %s", lp.g("failed_to_compress_file"), str(e))
            raise

    async def extract_file(self, file: str, target_directory: str) -> str:
        '''解压文件'''
        logging.info("%s: %s -> %s", lp.g("extracting_file"), file, target_directory)
        try:
            result = await self.ws_send_and_receive(
                f"extract:{file}(*.*){target_directory}"
            )
            if result != "ok":
                raise RuntimeError(result)
            logging.info("%s: %s", lp.g("file_extraction_result"), result)
            return result
        except Exception as e:
            logging.error("%s: %s", lp.g("failed_to_extract_file"), str(e))
            raise

    async def copy_file(self, source: str, target: str) -> str:
        '''复制文件'''
        logging.info("%s: %s -> %s", lp.g("copying_file"), source, target)
        try:
            result = await self.ws_send_and_receive(f"cp:{source}(*.*){target}")
            if result != "ok":
                raise RuntimeError(result)
            logging.info("%s: %s", lp.g("file_copy_result"), result)
            return result
        except Exception as e:
            logging.error("%s: %s", lp.g("failed_to_copy_file"), str(e))
            raise

    async def execute_command(self, command: str) -> dict:
        '''执行命令'''
        logging.info("%s: %s", lp.g("executing_command"), command)
        try:
            result = await self.ws_send_and_receive(f"cmd:{command}", load_json=True)
            logging.info("%s: %s", lp.g("command_result"), result)
            return result
        except Exception as e:
            logging.error("%s: %s", lp.g("command_execution_failed"), str(e))
            raise

    async def background(self, command: str) -> str:
        '''执行后台命令'''
        logging.info("%s: %s", lp.g("executing_command_in_background"), command)
        try:
            await self.websocket.send(f"bg:{command}")
            await self.websocket.recv()
            logging.info(lp.g("background_command_sent_successfully"))
            return "ok"
        except Exception as e:
            logging.error("%s: %s", lp.g("background_command_execution_failed"), str(e))
            raise

    async def get_file_list(self) -> dict:
        '''获取文件列表'''
        logging.info(lp.g("getting_file_list"))
        try:
            result = await self.ws_send_and_receive("ls", load_json=True)
            logging.info("%s: %s", lp.g("file_list"), result)
            return result
        except Exception as e:
            logging.error("%s: %s", lp.g("failed_to_get_file_list"), str(e))
            raise

    async def get_pwd(self) -> str:
        '''获取当前目录'''
        logging.info(lp.g("getting_current_directory"))
        try:
            result = await self.ws_send_and_receive("pwd")
            logging.info("%s: %s", lp.g("current_directory"), result)
            return result
        except Exception as e:
            logging.error("%s: %s", lp.g("failed_to_get_current_directory"), str(e))
            raise

    async def create_directory(self, directory: str) -> str:
        '''新建目录'''
        logging.info("%s: %s", lp.g("creating_directory"), directory)
        try:
            result = await self.ws_send_and_receive(f"mkdir:{directory}")
            logging.info("%s: %s", lp.g("directory_creation_result"), result)
            return result
        except Exception as e:
            logging.error("%s: %s", lp.g("directory_creation_failed"), str(e))
            raise

    async def change_directory(self, directory: str) -> str:
        '''切换目录'''
        logging.info("%s: %s", lp.g("changing_directory"), directory)
        try:
            result = await self.ws_send_and_receive(f"cd:{directory}")
            logging.info("%s: %s", lp.g("directory_change_result"), result)
            return result
        except Exception as e:
            logging.error("%s: %s", lp.g("directory_change_failed"), str(e))
            raise

    async def upload_file(self, file_path: str, file: typing.IO[bytes]) -> str:
        '''上传文件'''
        logging.info("%s: %s", lp.g("uploading_file"), file_path)
        try:
            logging.info(lp.g("rewriting_file"))
            result = await self.ws_send_and_receive(f"upload:{file_path}(*.*)")
            if result != "ok":
                raise RuntimeError(result)

            # 分段传输文件
            chunk_size = 4096
            while True:
                chunk = file.read(chunk_size)
                if not chunk:
                    break
                # logging.info("%s: %s bytes", lp.g('sending_chunk'), len(chunk))
                # 为了能传输任何形式的内容，使用base64编码
                encoded_chunk = base64.b64encode(chunk).decode("utf-8")
                result = await self.ws_send_and_receive(
                    f"upload:{file_path}(*.*){encoded_chunk}"
                )
                # logging.info("%s: %s", lp.g('upload_result'), result)
                if result != "ok":
                    raise RuntimeError(result)

            logging.info(lp.g("file_uploaded_successfully"))
            return "ok"
        except Exception as e:
            logging.error("%s: %s", lp.g("file_upload_failed"), str(e))
            raise

    async def download_file(self, file_path: str) -> str:
        '''下载文件'''
        logging.info("%s: %s", lp.g("downloading_file"), file_path)
        random_file_name = [chr(random.randint(97, 122)) for _ in range(64)]
        random_file_name = "".join(random_file_name)
        logging.info("%s: %s", lp.g("generated_random_file_name"), random_file_name)
        try:
            await self.websocket.send(f"download:{file_path}")
            while True:
                result = await self.websocket.recv()
                # logging.info("%s: %s", lp.g('received_chunk_length'), len(result))
                if result.startswith("error:zhaobokai"):
                    logging.error("%s: %s", lp.g("file_download_failed"), result)
                    raise RuntimeError(result)
                if result == "finish:zhaobokai":
                    logging.info(lp.g("file_downloaded_successfully"))
                    return random_file_name
                # base64解码写入内容
                with open(f"{download_directory_name}/{random_file_name}", "ab") as f:
                    f.write(base64.b64decode(result))
        except Exception as e:
            logging.error("%s: %s", lp.g("file_download_failed"), str(e))
            raise


# 安全验证函数
def check() -> bool:
    logging.info(lp.g("performing_security_verification"))
    logging.debug("%s: %s", lp.g("request_cookies"), request.cookies)
    if "Cookie" not in request.cookies:
        logging.warning(lp.g("cookie_not_found"))
        return False
    if bcrypt.checkpw(SECURITY_PASSWORD_HASH, request.cookies["Cookie"].encode()):
        logging.info(lp.g("verification_passed"))
        return True
    else:
        logging.warning(lp.g("verification_failed"))
        return False


# 禁止所有爬虫爬取所有页面
@app.route("/robots.txt")
async def robots():
    return """<pre>User-Agent: *
Disallow: /</pre>
    """


# 密码验证路由
@app.route(f"/{SECURITY_PATH}/verify", methods=["POST"])
async def verify():
    logging.info(lp.g("received_password_verification_request"))
    try:
        json_data = await request.get_json()
        if "password" not in json_data:
            logging.warning(lp.g("password_not_provided"))
            return jsonify({"error": lp.g("password_not_provided")}), 400

        password = json_data["password"]
        if bcrypt.checkpw(password.encode(), SECURITY_PASSWORD_HASH):
            logging.info(lp.g("password_verification_successful"))
            salt = bcrypt.gensalt(rounds=4)
            cookie = bcrypt.hashpw(SECURITY_PASSWORD_HASH, salt).decode()
            return jsonify({"Cookie": cookie})
        else:
            logging.warning(lp.g("incorrect_password"))
            return jsonify({"error": lp.g("incorrect_password")}), 401
    except Exception as e:
        logging.error("%s: %s", lp.g("verification_process_error"), str(e))
        return jsonify({"error": lp.g("server_error")}), 500


# 下载文件路由
@app.route(f"/{SECURITY_PATH}/download/<filename>", methods=["GET"])
async def download(filename: str):
    logging.info("%s: %s", lp.g("received_file_download_request"), filename)
    try:
        if not check():
            logging.warning(lp.g("unauthorized_request"))
            return jsonify({"error": lp.g("unauthorized_request")}), 401
        return await send_file(f"{download_directory_name}/{filename}")
    except Exception as e:
        logging.error("%s: %s", lp.g("file_download_failed"), str(e))
        return jsonify({"error": lp.g("server_error")}), 500


# API功能路由
@app.route(f"/{SECURITY_PATH}/function", methods=["POST"])
async def function():
    logging.info(lp.g("received_function_request"))
    try:
        if not check():
            logging.warning(lp.g("unauthorized_request"))
            return jsonify({"error": lp.g("unauthorized_request")}), 401

        json_data = await request.form

        # Use get with default None for func_name
        func_name = json_data.get("func_name")
        if not func_name:
            logging.warning(lp.g("function_name_not_provided"))
            return jsonify({"error": lp.g("function_name_not_provided")}), 400

        logging.info("%s: %s", lp.g("requested_function"), func_name)

        server = Server()
        client_list = await server.client_list()

        if func_name == "verify":
            return jsonify({"type": "ok"})
        elif func_name == "device_list":
            return jsonify(client_list[1])

        # Use get with default None for id
        device_id = json_data.get("id")
        if not device_id:
            logging.warning(lp.g("device_id_not_provided"))
            return jsonify({"error": lp.g("device_id_not_provided")}), 400

        logging.info("%s: %s", lp.g("target_device"), device_id)
        if client_list[0] == "error" and not any(
            device_id in device.values() for device in client_list[1]
        ):
            logging.warning("%s: %s", lp.g("device_does_not_exist"), device_id)
            return jsonify(
                {"error": f"{lp.g('device_with_id_does_not_exist')}{device_id}"}
            ), 400

        if func_name == "delete":
            return jsonify({"message": await server.delete(device_id)})

        device_lock = device_locks[device_id]
        try:
            await asyncio.wait_for(device_lock.acquire(), timeout=30.0)
        except asyncio.TimeoutError:
            logging.warning("%s: %s", lp.g("device_operation_timeout"), device_id)
            return jsonify(
                {"error": f"{lp.g('device_operation_timeout')}{device_id}"}
            ), 408

        control_client = ControlClient(device_id)
        try:
            if func_name == "systeminfo":
                return jsonify({"message": await control_client.system_info()})

            if func_name in ["command", "background"]:
                command = json_data.get("command")
                if not command:
                    logging.warning(lp.g("command_not_provided"))
                    return jsonify({"error": lp.g("command_not_provided")}), 400

                if func_name == "command":
                    return jsonify(
                        {"message": await control_client.execute_command(command)}
                    )
                else:
                    return jsonify(
                        {"message": await control_client.background(command)}
                    )

            if func_name in ["delete_file", "create_directory", "download"]:
                target_path = json_data.get("path")
                if not target_path:
                    logging.warning(lp.g("path_not_provided"))
                    return jsonify({"error": lp.g("path_not_provided")}), 400

                if func_name == "delete_file":
                    return jsonify(
                        {"message": await control_client.delete_file(target_path)}
                    )

                if func_name == "create_directory":
                    return jsonify(
                        {"message": await control_client.create_directory(target_path)}
                    )
                else:
                    return jsonify(
                        {"message": await control_client.download_file(target_path)}
                    )

            if func_name in ["copy_file", "move_file"]:
                old_name = json_data.get("old_path")
                new_name = json_data.get("new_path")

                if not old_name or not new_name:
                    logging.warning(lp.g("file_name_not_provided"))
                    return jsonify({"error": lp.g("file_name_not_provided")}), 400

                if func_name == "copy_file":
                    return jsonify(
                        {"message": await control_client.copy_file(old_name, new_name)}
                    )
                else:
                    return jsonify(
                        {"message": await control_client.move_file(old_name, new_name)}
                    )

            if func_name in ["compress", "extract"]:
                source_path = json_data.get("source_path")
                target_path = json_data.get("target_path")

                if not source_path or not target_path:
                    logging.warning(lp.g("path_not_provided"))
                    return jsonify({"error": lp.g("path_not_provided")}), 400

                if func_name == "compress":
                    return jsonify(
                        {
                            "message": await control_client.compress_file(
                                source_path, target_path
                            )
                        }
                    )
                else:
                    return jsonify(
                        {
                            "message": await control_client.extract_file(
                                source_path, target_path
                            )
                        }
                    )

            if func_name == "get_list_file":
                return jsonify({"message": await control_client.get_file_list()})

            if func_name == "get_pwd":
                return jsonify({"message": await control_client.get_pwd()})

            if func_name == "change_directory":
                directory = json_data.get("directory")
                if not directory:
                    logging.warning(lp.g("directory_not_provided"))
                    return jsonify({"error": lp.g("directory_not_provided")}), 400

                return jsonify(
                    {"message": await control_client.change_directory(directory)}
                )

            if func_name == "upload":
                file_dict = await request.files
                file_obj = file_dict.get("file")
                target_path = json_data.get("path")

                if not file_obj:
                    logging.warning(lp.g("file_not_provided"))
                    return jsonify({"error": lp.g("file_not_provided")}), 400

                if not target_path:
                    logging.warning(lp.g("path_not_provided"))
                    return jsonify({"error": lp.g("path_not_provided")}), 400

                return jsonify(
                    {"message": await control_client.upload_file(target_path, file_obj)}
                )
            else:
                logging.warning("%s: %s", lp.g("invalid_function"), func_name)
                return jsonify({"error": lp.g("invalid_function_name_provided")}), 400
        finally:
            device_lock.release()

    except Exception as e:
        logging.error("%s: %s", lp.g("request_processing_error"), str(e))
        return jsonify({"error": str(e)}), 500


# WebSocket客户端处理
async def handle_client(websocket):
    ip = f"{websocket.remote_address[0]}:{websocket.remote_address[1]}"
    logging.info("%s: %s", lp.g("new_client_connected"), ip)

    try:
        systeminfo = await websocket.recv()
        logging.info("%s: %s", lp.g("client_system_information"), systeminfo)
    except Exception as e:
        logging.error("%s: %s", lp.g("failed_to_get_system_info"), str(e))
        systeminfo = "ERROR"

    control_list[str(websocket.id)] = {
        "ip": ip,
        "websocket": websocket,
        "systeminfo": systeminfo,
    }

    logging.info(
        "%s %s, ID: %s", lp.g("client_connected_successfully"), ip, websocket.id
    )
    try:
        await websocket.wait_closed()
    except Exception as e:
        logging.error("%s: %s", lp.g("connection_exception"), str(e))
    finally:
        if str(websocket.id) in control_list:
            del control_list[str(websocket.id)]
            logging.info("%s %s", lp.g("client_disconnected"), ip)


# 服务器主循环
async def server_loop():
    logging.info(lp.g("initializing_server"))
    logging.info(
        "%s: %s, %s: %s", lp.g("certificate_path"), SSL_CERT, lp.g("key_path"), SSL_KEY
    )

    try:
        ssl_context = ssl.SSLContext(ssl.PROTOCOL_TLS_SERVER)
        ssl_context.load_cert_chain(SSL_CERT, SSL_KEY)
        logging.info(lp.g("ssl_certificate_loaded_successfully"))
    except FileNotFoundError:
        logging.error(lp.g("certificate_file_not_found"))
        exit(1)
    except Exception as e:
        logging.error("%s: %s", lp.g("ssl_loading_failed"), str(e))
        exit(1)

    logging.info("%s: %s:%s", lp.g("starting_server"), HOST, PORT)
    try:
        async with websockets.serve(handle_client, HOST, PORT, ssl=ssl_context):
            logging.info(lp.g("server_started_successfully"))
            try:
                quit_event = asyncio.Event()
                await quit_event.wait()
            except KeyboardInterrupt:
                logging.warning(lp.g("server_interrupted_by_user"))
    except Exception as e:
        logging.error("%s: %s", lp.g("server_startup_failed"), str(e))
        exit(1)


# 主程序入口
async def main():
    logging.info(lp.g("program_starting"))
    try:
        # 创建任务
        server_task = asyncio.create_task(server_loop())
        web_task = asyncio.create_task(app.run_task(host=WEB_HOST, port=WEB_PORT))

        # 等待任意一个任务完成或出错
        _, tasks = await asyncio.wait(
            [server_task, web_task], return_when=asyncio.FIRST_COMPLETED
        )
        for task in tasks:
            task.cancel()
    except Exception as e:
        logging.error("%s: %s", lp.g("program_error"), str(e))
        exit(1)


if __name__ == "__main__":
    try:
        if not os.path.exists(download_directory_name):
            os.mkdir(download_directory_name)
        logging.info(lp.g("download_directory_created"))
    except Exception as e:
        logging.error("%s: %s", lp.g("failed_to_create_download_directory"), str(e))
        exit(1)
    try:
        print("\033[H\033[J")
        logging.info(lp.g("copyright"))
        asyncio.run(main())
    except KeyboardInterrupt:
        logging.warning(lp.g("program_interrupted_by_user"))
        exit(0)
    except Exception as e:
        logging.critical("%s: %s", lp.g("fatal_error"), str(e))
        logging.error(lp.g("please_report_to_issues"))
        exit(1)
