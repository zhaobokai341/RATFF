# 远程访问木马 / Remote Access Trojan

语言 / Language: 中文 | English

---

## 中文 / Chinese

这个仓库包含一个功能完整的远程访问木马程序实现。当前版本：**v3.0-beta.1**

**重要声明**：本软件仅供教育和研究目的使用。用户使用该软件造成的任何损失和法律责任，均由用户自行承担。

### 主要特性：
- 支持多种平台的可执行文件生成 (exe, elf, apk, app) ✅
- 完整的中英文双语支持 ✅
- 基于WSS (WebSocket + SSL) 安全协议通信 ✅
- 提供命令行和网页两种控制界面 ✅
- 支持同时控制多个设备 ✅
- 丰富的远程控制功能：
  - 执行系统命令（前台/后台）✅
  - 获取详细的系统信息（CPU、内存、磁盘、网络、进程等）✅
  - 文件上传/下载 ✅
  - 文件管理（复制、移动、删除、压缩、解压）✅
  - 目录浏览和切换 ✅
  - 实时文件列表查看 ✅
  - Windows平台无窗口命令执行 ✅

### 环境准备：
请参考 [环境准备](wiki/zh/准备环境.md)

### 使用方法：

#### 下载项目
方法1：克隆仓库并进入目录
```bash
git clone https://github.com/zhaobokai341/remote_access_trojan.git
cd remote_access_trojan
```
方法2：通过浏览器下载压缩包并解压，然后进入目录
方法3：访问 [Releases](https://github.com/zhaobokai341/RATFF/releases) 页面，选择合适版本下载 RATFF.zip 文件并解压，然后进入目录

#### 项目配置
请参考 [配置项目](wiki/zh/配置项目.md)

#### 跨平台编译
请参考 [跨平台编译](wiki/zh/跨平台编译.md)

#### 运行程序
请参考 [运行](wiki/zh/运行.md)

#### 常见问题
请参考 [常见问题](wiki/zh/常见问题.md)

---

## English / 英文

This repository contains a complete implementation of a remote access trojan program. Current version: **v3.0-beta.1**

**Important Notice**: This software is for educational and research purposes only. Users bear full responsibility for any losses or legal liabilities caused by using this software.

### Main Features:
- Supports executable file generation for multiple platforms (exe, elf, apk, app) ✅
- Complete bilingual support for Chinese and English ✅
- Based on WSS (WebSocket + SSL) secure communication protocol ✅
- Provides both command-line and web control interfaces ✅
- Supports simultaneous control of multiple devices ✅
- Rich remote control functions:
  - Execute system commands (foreground/background) ✅
  - Obtain detailed system information (CPU, memory, disk, network, processes, etc.) ✅
  - File upload/download ✅
  - File management (copy, move, delete, compress, extract) ✅
  - Directory browsing and navigation ✅
  - Real-time file listing view ✅
  - Windowless command execution on Windows ✅

### Environment Setup:
Refer to [Environment Setup](wiki/en/Environment-Setup.md)

### How to Use:

#### Download Project
Method 1: Clone the repository and enter the directory
```bash
git clone https://github.com/zhaobokai341/remote_access_trojan.git
cd remote_access_trojan
```
Method 2: Download the compressed package through browser and extract, then enter the directory
Method 3: Visit [Releases](https://github.com/zhaobokai341/RATFF/releases) page, select appropriate version to download RATFF.zip file and extract, then enter the directory

#### Project Configuration
Refer to [Project Configuration](wiki/en/Project-Configuration.md)

#### Cross-Platform Build
Refer to [Cross-Platform Build](wiki/en/Cross-Platform-Build.md)

#### Running the Program
Refer to [Running](wiki/en/Running.md)

#### Frequently Asked Questions
Refer to [FAQ](wiki/en/Frequently-Asked-Questions.md)
