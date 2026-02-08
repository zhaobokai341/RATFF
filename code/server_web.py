__author__ = "赵博凯"
__license__ = "GPL v3"

from quart import Quart, redirect, url_for, request, make_response, render_template, jsonify, websocket
import asyncio
import rich
from sys import exit
import logging
import requests

import rich.traceback
import load_lang_pack 

# 配置Rich的回溯追踪功能
rich.traceback.install(show_locals=True)

# 基础配置
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
app = Quart(__name__)
LANGUAGE = "zh"
WEB_HOST = "0.0.0.0"
WEB_PORT = 8000
API_SITE = 'http://localhost:5000'
SECURITY_PATH = 'fuck'  # 安全路径
SECURITY_PASSWORD_HASH = '6ac3c336e4094835293a3fed8a4b5fedde1b5e2626d9838fed50693bba00af0e'  # 密码哈希值

# 全局变量
url_root = f'{API_SITE}/{SECURITY_PATH}'  # API根URL
lp = load_lang_pack.LanguagePack("server_web.json", LANGUAGE)  # 语言包
lp.load()
lp_login = load_lang_pack.LanguagePack("templates/login.json", LANGUAGE)  # 登录页面语言包
lp_login.load()
lp_index = load_lang_pack.LanguagePack("templates/index.json", LANGUAGE)  # 主页语言包
lp_index.load()
lp_device = load_lang_pack.LanguagePack("templates/device.json", LANGUAGE)  # 设备页面语言包
lp_device.load()
lp_file_manager = load_lang_pack.LanguagePack("templates/file_manager.json", LANGUAGE)  # 文件管理器页面语言包
lp_file_manager.load()

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
    return await render_template('index.html', lp=lp_index, language=LANGUAGE)

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
        return await render_template('login.html', error=lp.g('password_error'), lp=lp_login, language=LANGUAGE)
    return await render_template('login.html', lp=lp_login, language=LANGUAGE)

# 登出路由
@app.route(f'/{SECURITY_PATH}/logout')
async def logout():
    """处理用户登出"""
    resp = await make_response(redirect(url_for('index')))
    resp.delete_cookie('Cookie')
    return resp

# 设备管理页面路由
@app.route(f'/{SECURITY_PATH}/device/<id>')
async def device(id):
    """显示设备控制页面"""
    if not check(request.cookies): 
        return redirect(url_for('login'))
    return await render_template('device.html', id=id, url_root=url_root, lp=lp_device, language=LANGUAGE)

# 设备文件管理器页面路由
@app.route(f'/{SECURITY_PATH}/file_manager/<id>')
async def file_manager(id):
    """显示文3件管理器页面"""
    if not check(request.cookies): 
        return redirect(url_for('login'))
    return await render_template('file_manager.html', id=id, url_root=url_root, lp=lp_file_manager, language=LANGUAGE)

# API请求转发路由
@app.route(f'/{SECURITY_PATH}/requests_to_function', methods=['POST'])
async def requests_to_function():
    """转发API请求到后端服务"""
    if not check(request.cookies): 
        return redirect(url_for('login'))
    json = await request.json
    response = requests.post(f"{url_root}/function", data=json, verify=False, cookies=request.cookies)
    return jsonify(response.json())

# 主程序入口
async def main():
    """启动Web服务"""
    logging.info(lp.g("starting_program"))
    await asyncio.gather(
        app.run_task(host=WEB_HOST, port=WEB_PORT)
    )

if __name__ == '__main__':
    try:
        print("\033[H\033[J")
        logging.info(lp.g("copyright"))
        asyncio.run(main())
    except KeyboardInterrupt:
        logging.warning(lp.g("user_manually_interrupted"))
        exit()
    except Exception as e:
        logging.error(f"{lp.g('error_report')}: {e}")
