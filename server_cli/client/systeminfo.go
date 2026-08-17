package client

import (
	"time"

	"RATFF/shared"
)

// PrintSystemInfoDetail is a function type for printing system info details.
type PrintSystemInfoDetail func(map[string]interface{}, []string)

// SystemInfo retrieves system information from a client.
func (m *Manager) SystemInfo(id string, fields []string, printDetail PrintSystemInfoDetail) {
	payload := map[string]interface{}{
		"client_id": id,
		"command":   string(shared.CmdSystemInfo),
		"payload":   map[string]interface{}{"fields": fields},
	}

	msg := m.WaitForResponseWithMsg(id, shared.CmdSystemInfo, payload, 15*time.Second)
	if msg != nil && msg.Payload != nil {
		printDetail(msg.Payload, fields)
	}
}
