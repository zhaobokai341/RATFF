package main

import (
	"archive/zip"
	"encoding/base64"
	"errors"
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

// 客户端类
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

// 删除文件或文件夹
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

// 复制文件或文件夹
func (e *ExecuteCommand) copy_file(source, target string) string {
	file_info, err := os.Stat(source)
	fcpy := FileCopy{source, target}
	if err != nil {
		return err.Error()
	}
	if file_info.IsDir() {
		err = fcpy.copy_dir(source, target)
		if err != nil {
			return err.Error()
		}
	} else {
		err = fcpy.copy_file(source, target)
		if err != nil {
			return err.Error()
		}
	}
	return "ok"
}

// 重命名文件或文件夹
func (e *ExecuteCommand) move_file(source, target string) string {
	err := os.Rename(source, target)
	if err != nil {
		return err.Error()
	}
	return "ok"
}

// 压缩文件或文件夹
func (e *ExecuteCommand) compress(source, target string) string {
	file_info, err := os.Stat(source)
	if err != nil {
		return err.Error()
	}
	fcmps := FileCompress{source, target}

	if file_info.IsDir() {
		err = fcmps.compress_dir(source, target)
		if err != nil {
			return err.Error()
		}
	} else {
		err = fcmps.compress_file(source, target)
		if err != nil {
			return err.Error()
		}
	}
	return "ok"
}

// 解压文件或文件夹
func (e *ExecuteCommand) extract(source, target string) (string, error) {
	err_info := ""
	if err := os.MkdirAll(target, 0755); err != nil {
		return "", err
	}
	reader, err := zip.OpenReader(source)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	abs_dest_dir, err := filepath.Abs(target)
	if err != nil {
		return "", err
	}

	for _, file := range reader.File {
		source_path := filepath.Join(target, filepath.Clean(file.Name))
		abs_source_path, err := filepath.Abs(source_path)
		// 防止Zip Slip攻击
		if !strings.HasPrefix(abs_source_path, abs_dest_dir+string(os.PathSeparator)) ||
			strings.Contains(abs_source_path, "..") {
			continue
		}

		if file.FileInfo().IsDir() {
			err = os.MkdirAll(source_path, file.Mode())
			if err != nil {
				err_info += err.Error() + "\n"
				continue
			}
		} else {
			// 确保父目录存在
			if err := os.MkdirAll(filepath.Dir(source_path), 0755); err != nil {
				err_info += err.Error() + "\n"
				continue
			}

			sourceFile, err := os.Create(source_path)
			if err != nil {
				err_info += err.Error() + "\n"
				continue
			}
			targetFile, err := file.Open()
			if err != nil {
				err_info += err.Error() + "\n"
				sourceFile.Close()
				continue
			}
			_, err = io.Copy(sourceFile, targetFile)
			if err != nil {
				err_info += err.Error() + "\n"
				sourceFile.Close()
				targetFile.Close()
				continue
			}
			sourceFile.Close()
			targetFile.Close()
		}
	}
	if err_info != "" {
		return "", errors.New(err_info)
	}
	return "ok", nil
}

// 执行命令并返回结果
func (e *ExecuteCommand) execute_command(command string) map[string]interface{} {
	host_info, _ := host.Info()
	var shell, flag string

	if host_info.OS == "windows" {
		shell, flag = "cmd", "/C"
		// 强制设置 UTF-8 输出
		command = fmt.Sprintf("chcp %d >nul & ", 65001) + command
	} else {
		shell, flag = "sh", "-c"
	}

	cmd := exec.Command(shell, flag, command)
	output, err := cmd.CombinedOutput()

	// 将字节转换为字符串后，处理可能存在的无效 UTF-8 序列
	output_str := strings.ToValidUTF8(string(output), "")

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
		file, err := os.Create(file_path)
		file.Write([]byte(""))
		if err != nil {
			message = err.Error()
		} else {
			file.Close()
		}
	} else {
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
			// 解码base64编码的内容
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
