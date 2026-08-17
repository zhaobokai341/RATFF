package client

import (
	"time"

	"RATFF/shared"
)

// GetPublicIP retrieves public IP information from a client.
func (m *Manager) GetPublicIP(id string, printDetail PrintPublicIPDetail) {
	payload := map[string]interface{}{
		"client_id": id,
		"command":   string(shared.CmdPublicIP),
	}

	msg := m.WaitForResponseWithMsg(id, shared.CmdPublicIP, payload, 15*time.Second)
	if msg != nil && msg.Payload != nil {
		printDetail(msg.Payload)
	}
}

// PrintPublicIPDetail is a function type for printing public IP details.
type PrintPublicIPDetail func(map[string]interface{})
