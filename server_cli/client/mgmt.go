package client

import (
	"RATFF/shared"
)

// ListClients retrieves and prints the list of connected clients.
func (m *Manager) ListClients(printClientTable func([]shared.ClientInfo, func(string) string, func(string, ...interface{}) string)) {
	clients, err := m.apiClient.FetchClients()
	if err != nil {
		m.Print.Error(m.Tf("fetch_clients_failed", err))
		return
	}

	printClientTable(clients, m.T, m.Tf)
}

// SelectClient checks if a client with the given ID exists.
func (m *Manager) SelectClient(id string) bool {
	clients, err := m.apiClient.FetchClients()
	if err != nil {
		m.Print.Error(m.Tf("fetch_clients_failed", err))
		return false
	}

	for _, c := range clients {
		if c.ID == id {
			m.Print.Success(m.Tf("selected_client", id))
			return true
		}
	}

	m.Print.Error(m.T("client_not_exists"))
	return false
}

// DeleteClient sends an exit command to the specified client.
func (m *Manager) DeleteClient(id string) {
	clients, err := m.apiClient.FetchClients()
	if err != nil {
		m.Print.Error(m.Tf("fetch_clients_failed", err))
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
		m.Print.Error(m.T("client_not_exists"))
		return
	}

	payload := map[string]interface{}{
		"client_id": id,
		"command":   "exit",
	}
	if err := m.apiClient.PostCommand(payload); err != nil {
		m.Print.Error(m.Tf("send_exit_failed", err))
		return
	}

	m.Print.Success(m.T("delete_success"))
}
