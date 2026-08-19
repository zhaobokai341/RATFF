package client

// UpdateClient sends the service_update command to the remote client.
func (m *Manager) UpdateClient(id, tempPath string) {
	payload := map[string]interface{}{
		"client_id": id,
		"command":   "service_update",
		"payload": map[string]interface{}{
			"temp_path": tempPath,
		},
	}

	if err := m.postCommand(payload); err != nil {
		m.Print.Error(m.Tf("update_failed", err))
		return
	}

	m.Print.Success(m.Tf("update_success", id))
}
