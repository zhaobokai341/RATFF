__author__ = "赵博凯"
__license__ = "GPL v3"

from quart import Quart, request, jsonify, websocket
from sys import exit

import load_lang_pack

import bcrypt
import asyncio
import websockets
import ssl
import rich
import logging

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
SECURITY_PATH = 'fuck'
SECURITY_PASSWORD_HASH = b'$2b$04$T8NZ.WUIuO05WyVpLrQYiOdgqc2zbx7E9ysF03696dYvwGohCFzwC' 

# 全局变量
app = Quart(__name__)
control_list = {}
lp = load_lang_pack.LanguagePack("server_api.json", "en")
lp.load()

# 服务器核心类
class Server:
    def __init__(self):
        logging.info(lp.g("server_init"))

    def about(self):
        logging.info(lp.g("about"))
        return lp.g("about")
    
    def client_list(self):
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
            logging.info(f"{lp.g('device_list')}：{devices}")
            return devices
    
    async def delete(self, id):
        logging.info(f"{lp.g('deleting_device')}：{id}")
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
                logging.error(f"{lp.g('failed_to_delete_device')} {id}：{str(e)}")
                return f"{lp.g('exception_occurred_while_disconnecting')}{id}：{e}"
        else:
            logging.error(f"{lp.g('device_does_not_exist')} {id}")
            return f"{lp.g('device_with_id_does_not_exist')}{id}。"

# 客户端控制类
class ControlClient:
    def __init__(self, id):
        logging.info(f"{lp.g('initializing_client_controller')}：{id}")
        self.id = id
        self.websocket = control_list[id]['websocket']
    
    async def system_info(self):
        logging.info(lp.g("getting_device_system_information"))
        try:
            await self.websocket.send("systeminfo")
            result = await self.websocket.recv()
            logging.info(f"{lp.g('system_information')}：{result}")
            return result
        except Exception as e:
            logging.error(f"{lp.g('failed_to_get_system_information')}：{str(e)}")
            raise
    
    async def execute_command(self, command):
        logging.info(f"{lp.g('executing_command')}：{command}")
        try:
            await self.websocket.send(f"command:{command}")
            result = await self.websocket.recv()
            logging.info(f"{lp.g('command_result')}：{result}")
            return result
        except Exception as e:
            logging.error(f"{lp.g('command_execution_failed')}：{str(e)}")
            raise

    async def background(self, command):
        logging.info(f"{lp.g('executing_command_in_background')}：{command}")
        try:
            await self.websocket.send(f"background:{command}")
            await self.websocket.recv()
            logging.info(lp.g("background_command_sent_successfully"))
            return lp.g("command_sent")
        except Exception as e:
            logging.error(f"{lp.g('background_command_execution_failed')}：{str(e)}")
            raise

    async def change_directory(self, directory):
        logging.info(f"{lp.g('changing_directory')}：{directory}")
        try:
            await self.websocket.send(f"change_directory:{directory}")
            result = await self.websocket.recv()
            logging.info(f"{lp.g('directory_change_result')}：{result}")
            return result
        except Exception as e:
            logging.error(f"{lp.g('directory_change_failed')}：{str(e)}")
            raise

# 安全验证函数
def check():
    logging.info(lp.g("performing_security_verification"))
    logging.debug(f"{lp.g('request_cookies')}：{request.cookies}")
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
    return '''User-Agent: *
    Disallow: /
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
        logging.error(f"{lp.g('verification_process_error')}：{str(e)}")
        return jsonify({'error': lp.g('server_error')}), 500

# API功能路由
@app.route(f"/{SECURITY_PATH}/function", methods=['POST'])
async def function():
    logging.info(lp.g("received_function_request"))
    try:
        if not check():
            logging.warning(lp.g("unauthorized_request"))
            return jsonify({'error': lp.g('unauthorized_request')}), 401

        json_data = await request.get_json()
        logging.debug(f"{lp.g('request_data')}：{json_data}")
        
        if json_data is None or "func_name" not in json_data:
            logging.warning(lp.g("function_name_not_provided"))
            return jsonify({'error': lp.g('function_name_not_provided')}), 400

        func_name = json_data["func_name"]
        logging.info(f"{lp.g('requested_function')}：{func_name}")

        valid_functions = ["device_list", "delete", "systeminfo", "command", "background", "change_directory"]
        if func_name not in valid_functions:
            logging.warning(f"{lp.g('invalid_function')}：{func_name}")
            return jsonify({'error': lp.g('invalid_function_name_provided')}), 400

        server = Server()

        if func_name == "device_list":
            return jsonify(server.client_list())

        if "id" not in json_data:
            logging.warning(lp.g("device_id_not_provided"))
            return jsonify({'error': lp.g('device_id_not_provided')}), 400
        
        device_id = json_data["id"]
        logging.info(f"{lp.g('target_device')}：{device_id}")
        
        client_list = server.client_list()
        if "没有" in client_list or "No" in client_list:
            logging.warning(f"{lp.g('device_does_not_exist')}：{device_id}")
            return jsonify({'error': f"{lp.g('device_with_id_does_not_exist')}{device_id}"}), 400
        if not any(device_id in device.values() for device in client_list):
            logging.warning(f"{lp.g('device_does_not_exist')}：{device_id}")
            return jsonify({'error': f"{lp.g('device_with_id_does_not_exist')}{device_id}"}), 400
        
        if func_name == "delete":
            return jsonify({"message": await server.delete(device_id)})
        
        control_client = ControlClient(device_id)
        
        if func_name == "systeminfo":
            return jsonify({"message": await control_client.system_info()})

        if func_name in ["command", "background"]:
            if "command" not in json_data:
                logging.warning(lp.g("command_not_provided"))
                return jsonify({'error': lp.g('command_not_provided')}), 400
            
            command = json_data["command"]
            if func_name == "command":
                return jsonify({"message": await control_client.execute_command(command)})
            else:
                return jsonify({"message": await control_client.background(command)})
        
        if func_name == "change_directory":
            if "directory" not in json_data:
                logging.warning(lp.g("directory_not_provided"))
                return jsonify({'error': lp.g('directory_not_provided')}), 400
            
            directory = json_data["directory"]
            return jsonify({"message": await control_client.change_directory(directory)})
    
    except Exception as e:
        logging.error(f"{lp.g('request_processing_error')}：{str(e)}")
        return jsonify({'error': str(e)}), 500

# WebSocket客户端处理
async def handle_client(websocket):
    ip = websocket.remote_address[0] + ":" + str(websocket.remote_address[1])
    logging.info(f"{lp.g('new_client_connected')}：{ip}")
    
    try:
        systeminfo = await websocket.recv()
        logging.info(f"{lp.g('client_system_information')}：{systeminfo}")
    except Exception as e:
        logging.error(f"{lp.g('failed_to_get_system_info')}：{str(e)}")
        systeminfo = "ERROR"
    
    control_list[str(websocket.id)] = {
        "ip": ip,
        "websocket": websocket,
        "systeminfo": systeminfo
    }
    
    logging.info(f"{lp.g('client_connected_successfully')} {ip}，ID：{websocket.id}")
    try:
        await websocket.wait_closed()
    except Exception as e:
        logging.error(f"{lp.g('connection_exception')}：{str(e)}")
    finally:
        if str(websocket.id) in control_list:
            del control_list[str(websocket.id)]
            logging.info(f"{lp.g('client_disconnected')} {ip}")

# 服务器主循环
async def server_loop():
    logging.info(lp.g("initializing_server"))
    logging.info(f"{lp.g('certificate_path')}：{SSL_CERT}，{lp.g('key_path')}：{SSL_KEY}")
    
    try:
        ssl_context = ssl.SSLContext(ssl.PROTOCOL_TLS_SERVER)
        ssl_context.load_cert_chain(SSL_CERT, SSL_KEY)
        logging.info(lp.g("ssl_certificate_loaded_successfully"))
    except FileNotFoundError:
        logging.error(lp.g("certificate_file_not_found"))
        exit(1)
    except Exception as e:
        logging.error(f"{lp.g('ssl_loading_failed')}：{str(e)}")
        exit(1)

    logging.info(f"{lp.g('starting_server')}：{HOST}:{PORT}")
    try:
        async with websockets.serve(handle_client, HOST, PORT, ssl=ssl_context):
            logging.info(lp.g("server_started_successfully"))
            try:
                quit_event = asyncio.Event()
                await quit_event.wait()
            except KeyboardInterrupt:
                logging.warning(lp.g("server_interrupted_by_user"))
    except Exception as e:
        logging.error(f"{lp.g('server_startup_failed')}：{str(e)}")
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
        logging.error(f"{lp.g('program_error')}：{str(e)}")
        exit(1)

if __name__ == '__main__':
    try:
        print("\033[H\033[J")
        logging.info(lp.g("copyright"))
        asyncio.run(main())
    except KeyboardInterrupt:
        logging.warning(lp.g("program_interrupted_by_user"))
        exit(0)
    except Exception as e:
        logging.critical(f"{lp.g('fatal_error')}：{str(e)}")
        logging.error(lp.g("please_report_to_issues"))
        exit(1)
