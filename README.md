# 远程访问木马 (Remote Access Trojan)

语言：[中文](README.md) [English](README_en.md)

这个仓库包含一个功能完整的远程访问木马程序实现。

**重要声明**：本软件仅供教育和研究目的使用。用户使用该软件造成的任何损失和法律责任，均由用户自行承担。

## 主要特性：
- 支持多种平台的可执行文件生成 (exe, elf, apk, app) ✅
- 完整的中英文双语支持 ✅
- 基于WSS (WebSocket + SSL) 安全协议通信 ✅
- 提供命令行和网页两种控制界面 ✅
- 支持同时控制多个设备 ✅
- 丰富的远程控制功能：
  - 执行系统命令 ✅
  - 获取详细的系统信息（CPU、内存、磁盘、网络等）✅
  - 文件上传/下载 ✅
  - 文件管理（复制、移动、删除、压缩、解压）✅
  - 目录浏览和切换 ✅
  - 后台命令执行 ✅
  - 实时文件列表查看 ✅

## 环境准备：
请参考 [环境准备](wiki/zh/准备环境.md)

## 使用方法：
### 下载项目
方法1：克隆仓库并进入目录
```bash
git clone https://github.com/zhaobokai341/remote_access_trojan.git
cd remote_access_trojan
```
方法2：通过浏览器下载压缩包并解压，然后进入目录
方法3：访问 [Releases](https://github.com/zhaobokai341/RATFF/releases) 页面，选择合适版本下载 code.zip 文件并解压，然后进入目录

### 项目配置
请参考 [配置项目](wiki/zh/配置项目.md)

### 运行程序
请参考 [运行](wiki/zh/运行.md)
