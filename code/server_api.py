__author__ = "赵博凯"
__license__ = "GPL v3"

from quart import Quart, request, jsonify, send_file
from sys import exit

import load_lang_pack

import asyncio
import base64
import bcrypt
import logging
import os
import random
import rich
import ssl
import websockets

import rich.traceback 
import rich.logging

# 配置Rich的回溯追踪
rich.traceback.install(show_locals=True)

# 日志配置
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s', handlers=[rich.logging.RichHandler()])

# 服务器配置
HOST = '0.0.0.0' 
PORT = 8765
WEB_HOST = '0.0.0.0'
WEB_PORT = 5000
SSL_CERT = 'cert.pem' 
SSL_KEY = 'key.pem'
LANGUAGE = 'zh'
SECURITY_PATH = 'fuck'
SECURITY_PASSWORD_HASH = b'$2b$04$T8NZ.WUIuO05WyVpLrQYiOdgqc2zbx7E9ysF03696dYvwGohCFzwC'

# 全局变量
app = Quart(__name__)
app.config['MAX_CONTENT_LENGTH'] = None
app.config['BODY_TIMEOUT'] = None 
control_list = {}
delete_file = ""
download_directory_name = "download"
lp = load_lang_pack.LanguagePack("server_api.json", LANGUAGE)
lp.load()

# 服务器核心类
class Server:
    def __init__(self):
        logging.info(lp.g("server_init"))

    def about(self):
        logging.info(lp.g("about"))
        return lp.g("about")
    
    async def client_list(self):
        logging.info(lp.g("getting_client_list"))
        if len(control_list) == 0:
            logging.info(lp.g("no_connected_devices"))
            return lp.g("no_devices_connected")
        else:
            devices = []
            for device in control_list.items():
                devices.append({
                    'id': device[0],
                    'ip': device[1]['ip'],
                    'systeminfo': device[1]['systeminfo']
                })
            logging.info(f"{lp.g('device_list')}: {devices}")
            return devices
    
    async def delete(self, id):
        logging.info(f"{lp.g('deleting_device')}: {id}")
        global control_list
        if id in control_list:
            websocket_ = control_list[id]['websocket']
            try:
                await websocket_.send("exit")
                control_list.pop(id)
                logging.info(f"{lp.g('device_deleted_successfully')} {id}")
                return f"{lp.g('successfully_deleted_device')}{id}。"
            except Exception as e:
                control_list.pop(id)
                logging.error(f"{lp.g('failed_to_delete_device')} {id}: {str(e)}")
                return f"{lp.g('exception_occurred_while_disconnecting')}{id}: {e}"
        else:
            logging.error(f"{lp.g('device_does_not_exist')} {id}")
            return f"{lp.g('device_with_id_does_not_exist')}{id}。"

# 客户端控制类
class ControlClient:
    def __init__(self, id):
        logging.info(f"{lp.g('initializing_client_controller')}: {id}")
        self.id = id
        self.websocket = control_list[id]['websocket']
    
    async def system_info(self):
        logging.info(lp.g("getting_device_system_information"))
        try:
            await self.websocket.send("systeminfo")
            result = await self.websocket.recv()
            logging.info(f"{lp.g('system_information')}: {result}")
            return result
        except Exception as e:
            logging.error(f"{lp.g('failed_to_get_system_information')}: {str(e)}")
            raise
    
    async def delete_file(self, file):
        logging.info(f"{lp.g('deleting_file')}: {file}")
        try:
            await self.websocket.send(f"delete:{file}")
            result = await self.websocket.recv()
            if result != "ok":
                raise Exception(result)
            logging.info(f"{lp.g('file_deletion_result')}: {result}")
            return result
        except Exception as e:
            logging.error(f"{lp.g('failed_to_delete_file')}: {str(e)}")
            raise
    
    async def move_file(self, old_file, new_file):
        logging.info(f"{lp.g('moving_file')}: {old_file} -> {new_file}")
        try:
            await self.websocket.send(f"mv:{old_file}(*.*){new_file}")
            result = await self.websocket.recv()
            if result != "ok":
                raise Exception(result)
            logging.info(f"{lp.g('file_move_result')}: {result}")
            return result
        except Exception as e:
            logging.error(f"{lp.g('failed_to_move_file')}: {str(e)}")
            raise

    async def compress_file(self, file, target_file):
        logging.info(f"{lp.g('compressing_file')}: {file} -> {target_file}")
        try:
            await self.websocket.send(f"compress:{file}(*.*){target_file}")
            result = await self.websocket.recv()
            if result != "ok":
                raise Exception(result)
            logging.info(f"{lp.g('file_compression_result')}: {result}")
            return result
        except Exception as e:
            logging.error(f"{lp.g('failed_to_compress_file')}: {str(e)}")
            raise

    async def extract_file(self, file, target_directory):
        logging.info(f"{lp.g('extracting_file')}: {file} -> {target_directory}")
        try:
            await self.websocket.send(f"extract:{file}(*.*){target_directory}")
            result = await self.websocket.recv()
            if result != "ok":
                raise Exception(result)
            logging.info(f"{lp.g('file_extraction_result')}: {result}")
            return result
        except Exception as e:
            logging.error(f"{lp.g('failed_to_extract_file')}: {str(e)}")
            raise

    async def copy_file(self, old_file, new_file):
        logging.info(f"{lp.g('copying_file')}: {old_file} -> {new_file}")
        try:
            await self.websocket.send(f"cp:{old_file}(*.*){new_file}")
            result = await self.websocket.recv()
            if result != "ok":
                raise Exception(result)
            logging.info(f"{lp.g('file_copy_result')}: {result}")
            return result
        except Exception as e:
            logging.error(f"{lp.g('failed_to_copy_file')}: {str(e)}")
            raise

    async def execute_command(self, command):
        logging.info(f"{lp.g('executing_command')}: {command}")
        try:
            await self.websocket.send(f"command:{command}")
            result = await self.websocket.recv()
            logging.info(f"{lp.g('command_result')}: {result}")
            return result
        except Exception as e:
            logging.error(f"{lp.g('command_execution_failed')}: {str(e)}")
            raise

    async def background(self, command):
        logging.info(f"{lp.g('executing_command_in_background')}: {command}")
        try:
            await self.websocket.send(f"background:{command}")
            await self.websocket.recv()
            logging.info(lp.g("background_command_sent_successfully"))
            return "ok"
        except Exception as e:
            logging.error(f"{lp.g('background_command_execution_failed')}: {str(e)}")
            raise

    async def get_file_list(self):
        logging.info(lp.g("getting_file_list"))
        try:
            await self.websocket.send("ls")
            result = await self.websocket.recv()
            logging.info(f"{lp.g('file_list')}: {result}")
            return result
        except Exception as e:
            logging.error(f"{lp.g('failed_to_get_file_list')}: {str(e)}")
            raise

    async def get_pwd(self):
        logging.info(lp.g("getting_current_directory"))
        try:
            await self.websocket.send("pwd")
            result = await self.websocket.recv()
            logging.info(f"{lp.g('current_directory')}: {result}")
            return result
        except Exception as e:
            logging.error(f"{lp.g('failed_to_get_current_directory')}: {str(e)}")
            raise

    async def create_directory(self, directory):
        logging.info(f"{lp.g('creating_directory')}: {directory}")
        try:
            await self.websocket.send(f"mkdir:{directory}")
            result = await self.websocket.recv()
            logging.info(f"{lp.g('directory_creation_result')}: {result}")
            return result
        except Exception as e:
            logging.error(f"{lp.g('directory_creation_failed')}: {str(e)}")
            raise

    async def change_directory(self, directory):
        logging.info(f"{lp.g('changing_directory')}: {directory}")
        try:
            await self.websocket.send(f"cd:{directory}")
            result = await self.websocket.recv()
            logging.info(f"{lp.g('directory_change_result')}: {result}")
            return result
        except Exception as e:
            logging.error(f"{lp.g('directory_change_failed')}: {str(e)}")
            raise
    
    async def upload_file(self, file_path, file):
        logging.info(f"{lp.g('uploading_file')}: {file_path}")
        try:
            logging.info(f"{lp.g('rewriting_file')}")
            await self.websocket.send(f"upload:{file_path}(*.*)")
            result = await self.websocket.recv()
            if result != "ok":
                raise Exception(result)
            
            # Read file in chunks directly
            chunk_size = 4096
            while True:
                chunk = file.read(chunk_size)
                if not chunk:
                    break
                # logging.info(f"{lp.g('sending_chunk')}: {len(chunk)} bytes")
                # Encode the chunk as base64 to safely send over WebSocket
                encoded_chunk = base64.b64encode(chunk).decode('utf-8')
                await self.websocket.send(f"upload:{file_path}(*.*){encoded_chunk}")
                result = await self.websocket.recv()
                #logging.info(f"{lp.g('upload_result')}: {result}")
                if result != "ok":
                    raise Exception(result)
            
            logging.info(lp.g("file_uploaded_successfully"))
            return "ok"
        except Exception as e:
            logging.error(f"{lp.g('file_upload_failed')}: {str(e)}")
            raise

    async def download_file(self, file_path):
        logging.info(f"{lp.g('downloading_file')}: {file_path}")
        random_file_name = [chr(random.randint(97, 122)) for _ in range(64)]
        random_file_name = ''.join(random_file_name)
        logging.info(f"{lp.g('generated_random_file_name')}: {random_file_name}")
        try:
            await self.websocket.send(f"download:{file_path}")
            while True:
                result = await self.websocket.recv()
                #logging.info(f"{lp.g('received_chunk_length')}: {len(result)}")
                if result.startswith("error:zhaobokai"):
                    logging.error(f"{lp.g('file_download_failed')}: {result}")
                    raise Exception(result)
                if result == "finish:zhaobokai":
                    logging.info(lp.g("file_downloaded_successfully"))
                    return random_file_name
                with open(f"{download_directory_name}/{random_file_name}", "ab") as f:
                    f.write(base64.b64decode(result))
        except Exception as e:
            logging.error(f"{lp.g('file_download_failed')}: {str(e)}")
            raise

# 安全验证函数
def check():
    logging.info(lp.g("performing_security_verification"))
    logging.debug(f"{lp.g('request_cookies')}: {request.cookies}")
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
    return '''<pre>User-Agent: *
Disallow: /</pre>
    '''

# 密码验证路由
@app.route(f"/{SECURITY_PATH}/verify", methods=['POST'])
async def verify():
    logging.info(lp.g("received_password_verification_request"))
    try:
        json_data = await request.get_json()
        if "password" not in json_data:
            logging.warning(lp.g("password_not_provided"))
            return jsonify({'error': lp.g('password_not_provided')}), 400
        
        password = json_data["password"]
        if bcrypt.checkpw(password.encode(), SECURITY_PASSWORD_HASH):
            logging.info(lp.g("password_verification_successful"))
            salt = bcrypt.gensalt(rounds=4)
            cookie = bcrypt.hashpw(SECURITY_PASSWORD_HASH, salt).decode()
            return jsonify({'Cookie': cookie})
        else:
            logging.warning(lp.g("incorrect_password"))
            return jsonify({'error': lp.g('incorrect_password')}), 401
    except Exception as e:
        logging.error(f"{lp.g('verification_process_error')}: {str(e)}")
        return jsonify({'error': lp.g('server_error')}), 500

# 下载文件路由
@app.route(f"/{SECURITY_PATH}/download/<filename>", methods=['GET'])
async def download(filename: str):
    logging.info(f"{lp.g('received_file_download_request')}: {filename}")
    try:
        if not check():
            logging.warning(lp.g("unauthorized_request"))
            return jsonify({'error': lp.g('unauthorized_request')}), 401
        return await send_file(f"{download_directory_name}/{filename}")
    except Exception as e:
        logging.error(f"{lp.g('file_download_failed')}: {str(e)}")
        return jsonify({'error': lp.g('server_error')}), 500

# API功能路由
@app.route(f"/{SECURITY_PATH}/function", methods=['POST'])
async def function():
    logging.info(lp.g("received_function_request"))
    try:
        if not check():
            logging.warning(lp.g("unauthorized_request"))
            return jsonify({'error': lp.g('unauthorized_request')}), 401

        json_data = await request.form
        logging.debug(f"{lp.g('request_data')}: {json_data}")
        
        if json_data is None or "func_name" not in json_data:
            logging.warning(lp.g("function_name_not_provided"))
            return jsonify({'error': lp.g('function_name_not_provided')}), 400

        func_name = json_data["func_name"]
        logging.info(f"{lp.g('requested_function')}: {func_name}")

        if func_name in ["device_list",
                        "delete",
                        "verify",
                        "systeminfo",
                        "command",
                        "background",
                        "delete_file",
                        "extract",
                        "compress",
                        "copy_file",
                        "move_file",
                        "create_directory",
                        "change_directory", 
                        "upload", 
                        "download", 
                        "get_list_file", 
                        "get_pwd"]:
            server = Server()
            if func_name == "verify":
                return jsonify({"type": "ok"})
            elif func_name == "device_list":
                return jsonify(await server.client_list())
            else:
                if "id" not in json_data:
                    logging.warning(lp.g("device_id_not_provided"))
                    return jsonify({'error': lp.g('device_id_not_provided')}), 400
                device_id = json_data["id"]
                logging.info(f"{lp.g('target_device')}: {device_id}")
                client_list = await server.client_list()
                if isinstance(client_list, str):
                    logging.warning(f"{lp.g('device_does_not_exist')}: {device_id}")
                    return jsonify({'error': f"{lp.g('device_with_id_does_not_exist')}{device_id}"}), 400
                if not any(device_id in device.values() for device in client_list):
                    logging.warning(f"{lp.g('device_does_not_exist')}: {device_id}")
                    return jsonify({'error': f"{lp.g('device_with_id_does_not_exist')}{device_id}"}), 400
                
                if func_name == "delete":
                    return jsonify({"message": await server.delete(device_id)})
                else:
                    control_client = ControlClient(device_id)
                    if func_name == "systeminfo":
                        return jsonify({"message": await control_client.system_info()})
                    else:
                        if func_name in ["command", "background"]:
                            if "command" not in json_data:
                                logging.warning(lp.g("command_not_provided"))
                                return jsonify({'error': lp.g('command_not_provided')}), 400
                            command = json_data["command"]
                            if func_name == "command":
                                return jsonify({"message": await control_client.execute_command(command)})
                            else:
                                return jsonify({"message": await control_client.background(command)})
                        elif func_name in ["delete_file", "create_directory", "download"]:
                            if "path" not in json_data:
                                logging.warning(lp.g("path_not_provided"))
                                return jsonify({'error': lp.g('path_not_provided')}), 400
                            target_path = json_data["path"]
                            if func_name == "delete_file":
                                return jsonify({"message": await control_client.delete_file(target_path)})
                            elif func_name == "create_directory":
                                return jsonify({"message": await control_client.create_directory(target_path)})
                            elif func_name == "download":
                                return jsonify({"message": await control_client.download_file(target_path)})
                        elif func_name in ["copy_file", "move_file"]:
                            if "old_path" not in json_data or "new_path" not in json_data:
                                logging.warning(lp.g("file_name_not_provided"))
                                return jsonify({'error': lp.g('file_name_not_provided')}), 400
                            old_name = json_data["old_path"]
                            new_name = json_data["new_path"]
                            if func_name == "copy_file":
                                return jsonify({"message": await control_client.copy_file(old_name, new_name)})
                            elif func_name == "move_file":
                                return jsonify({"message": await control_client.move_file(old_name, new_name)})
                        elif func_name in ["compress", "extract"]:
                            if "source_path" not in json_data or "target_path" not in json_data:
                                logging.warning(lp.g("path_not_provided"))
                                return jsonify({'error': lp.g('path_not_provided')}), 400
                            source_path = json_data["source_path"]
                            target_path = json_data["target_path"]
                            if func_name == "compress":
                                return jsonify({"message": await control_client.compress_file(source_path, target_path)})
                            elif func_name == "extract":
                                return jsonify({"message": await control_client.extract_file(source_path, target_path)})
                        elif func_name == "get_list_file":
                            return jsonify({"message": await control_client.get_file_list()})
                        elif func_name == "get_pwd":
                            return jsonify({"message": await control_client.get_pwd()})
                        elif func_name == "change_directory":
                            if "directory" not in json_data:
                                logging.warning(lp.g("directory_not_provided"))
                                return jsonify({'error': lp.g('directory_not_provided')}), 400
                            
                            directory = json_data["directory"]
                            return jsonify({"message": await control_client.change_directory(directory)})
                        else:
                            if func_name == "upload":
                                file = await request.files
                                if "file" not in file:
                                    logging.warning(lp.g("file_not_provided"))
                                    return jsonify({'error': lp.g('file_not_provided')}), 400
                                if "path" not in json_data:
                                    logging.warning(lp.g("path_not_provided"))
                                    return jsonify({'error': lp.g('path_not_provided')}), 400
                                
                                target_path = json_data['path']
                                file = file["file"]
                                return jsonify({"message": await control_client.upload_file(target_path, file)})
        else:
            logging.warning(f"{lp.g('invalid_function')}: {func_name}")
            return jsonify({'error': lp.g('invalid_function_name_provided')}), 400
    
    except Exception as e:
        logging.error(f"{lp.g('request_processing_error')}: {str(e)}")
        return jsonify({'error': str(e)}), 500

# WebSocket客户端处理
async def handle_client(websocket):
    ip = websocket.remote_address[0] + ":" + str(websocket.remote_address[1])
    logging.info(f"{lp.g('new_client_connected')}: {ip}")
    
    try:
        systeminfo = await websocket.recv()
        logging.info(f"{lp.g('client_system_information')}: {systeminfo}")
    except Exception as e:
        logging.error(f"{lp.g('failed_to_get_system_info')}: {str(e)}")
        systeminfo = "ERROR"
    
    control_list[str(websocket.id)] = {
        "ip": ip,
        "websocket": websocket,
        "systeminfo": systeminfo
    }
    
    logging.info(f"{lp.g('client_connected_successfully')} {ip}，ID: {websocket.id}")
    try:
        await websocket.wait_closed()
    except Exception as e:
        logging.error(f"{lp.g('connection_exception')}: {str(e)}")
    finally:
        if str(websocket.id) in control_list:
            del control_list[str(websocket.id)]
            logging.info(f"{lp.g('client_disconnected')} {ip}")

# 服务器主循环
async def server_loop():
    logging.info(lp.g("initializing_server"))
    logging.info(f"{lp.g('certificate_path')}: {SSL_CERT}，{lp.g('key_path')}: {SSL_KEY}")
    
    try:
        ssl_context = ssl.SSLContext(ssl.PROTOCOL_TLS_SERVER)
        ssl_context.load_cert_chain(SSL_CERT, SSL_KEY)
        logging.info(lp.g("ssl_certificate_loaded_successfully"))
    except FileNotFoundError:
        logging.error(lp.g("certificate_file_not_found"))
        exit(1)
    except Exception as e:
        logging.error(f"{lp.g('ssl_loading_failed')}: {str(e)}")
        exit(1)

    logging.info(f"{lp.g('starting_server')}: {HOST}:{PORT}")
    try:
        async with websockets.serve(handle_client, HOST, PORT, ssl=ssl_context):
            logging.info(lp.g("server_started_successfully"))
            try:
                quit_event = asyncio.Event()
                await quit_event.wait()
            except KeyboardInterrupt:
                logging.warning(lp.g("server_interrupted_by_user"))
    except Exception as e:
        logging.error(f"{lp.g('server_startup_failed')}: {str(e)}")
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
            [server_task, web_task],
            return_when=asyncio.FIRST_COMPLETED
        )
        for task in tasks:
            task.cancel()
    except Exception as e:
        logging.error(f"{lp.g('program_error')}: {str(e)}")
        exit(1)

if __name__ == '__main__':
    try:
        if not os.path.exists(download_directory_name):
            os.mkdir(download_directory_name)
        logging.info(lp.g("download_directory_created"))
    except Exception as e:
        logging.error(f"{lp.g('failed_to_create_download_directory')}: {str(e)}")
        exit(1)
    try:
        print("\033[H\033[J")
        logging.info(lp.g("copyright"))
        asyncio.run(main())
    except KeyboardInterrupt:
        logging.warning(lp.g("program_interrupted_by_user"))
        exit(0)
    except Exception as e:
        logging.critical(f"{lp.g('fatal_error')}: {str(e)}")
        logging.error(lp.g("please_report_to_issues"))
        exit(1)
