package client

import (
	"time"

	"RATFF/shared"
)

// CdClient changes the working directory on the remote client.
func (m *Manager) CdClient(id string, dir string) {
	payload := map[string]interface{}{"dir": dir}

	msg := m.WaitForResponseWithMsg(id, shared.CmdCd, payload, 10*time.Second)
	if msg == nil {
		return
	}

	if msg.Payload == nil {
		m.Print.Error(m.T("cd_failed"))
		return
	}

	if errMsg, ok := msg.Payload["error"].(string); ok {
		m.Print.Error(m.Tf("cd_failed", errMsg))
		return
	}

	if currentDir, ok := msg.Payload["current_dir"].(string); ok {
		m.Print.Success(m.Tf("cd_success", dir, currentDir))
	}
}

// ListFiles lists files in a remote directory.
func (m *Manager) ListFiles(id, path string, printFileTable func(string, []interface{}, func(string) string, func(string, ...interface{}) string)) {
	payload := map[string]interface{}{"path": path}

	msg := m.WaitForResponseWithMsg(id, shared.CmdFileList, payload, 10*time.Second)
	if msg == nil {
		return
	}

	if msg.Payload == nil {
		m.Print.Error(m.T("file_list_empty"))
		return
	}

	if errMsg, ok := msg.Payload["error"].(string); ok {
		m.Print.Error(m.Tf("file_list_failed", errMsg))
		return
	}

	currentPath, _ := msg.Payload["path"].(string)
	filesInterface, ok := msg.Payload["files"].([]interface{})
	if !ok {
		m.Print.Error(m.T("file_list_parse_failed"))
		return
	}

	printFileTable(currentPath, filesInterface, m.T, m.Tf)
}

// MoveFile moves a file on the remote client.
func (m *Manager) MoveFile(id, originPath, newPath string) {
	payload := map[string]interface{}{
		"origin_path": originPath,
		"new_path":    newPath,
	}

	msg := m.WaitForResponseWithMsg(id, shared.CmdFileMove, payload, 10*time.Second)
	if msg == nil {
		return
	}

	if msg.Payload == nil {
		m.Print.Error(m.T("file_move_failed"))
		return
	}

	if errMsg, ok := msg.Payload["error"].(string); ok {
		m.Print.Error(m.Tf("file_move_failed_detail", errMsg))
		return
	}

	m.Print.Success(m.Tf("file_move_success", originPath, newPath))
}

// DeleteFile deletes a file on the remote client.
func (m *Manager) DeleteFile(id, path string) {
	payload := map[string]interface{}{"path": path}

	msg := m.WaitForResponseWithMsg(id, shared.CmdFileDelete, payload, 10*time.Second)
	if msg == nil {
		return
	}

	if msg.Payload == nil {
		m.Print.Error(m.T("file_delete_failed"))
		return
	}

	if errMsg, ok := msg.Payload["error"].(string); ok {
		m.Print.Error(m.Tf("file_delete_failed_detail", errMsg))
		return
	}

	m.Print.Success(m.Tf("file_delete_success", path))
}

// CopyRemoteFile copies a file on the remote client.
func (m *Manager) CopyRemoteFile(id, originPath, newPath string) {
	payload := map[string]interface{}{
		"origin_path": originPath,
		"new_path":    newPath,
	}

	msg := m.WaitForResponseWithMsg(id, shared.CmdFileCopy, payload, 10*time.Second)
	if msg == nil {
		return
	}

	if msg.Payload == nil {
		m.Print.Error(m.T("file_copy_failed"))
		return
	}

	if errMsg, ok := msg.Payload["error"].(string); ok {
		m.Print.Error(m.Tf("file_copy_failed_detail", errMsg))
		return
	}

	m.Print.Success(m.Tf("file_copy_success", originPath, newPath))
}

// PwdClient gets the current working directory on the remote client.
func (m *Manager) PwdClient(id string) {
	payload := map[string]interface{}{}

	msg := m.WaitForResponseWithMsg(id, shared.CmdPwd, payload, 10*time.Second)
	if msg == nil {
		return
	}

	if msg.Payload == nil {
		m.Print.Error(m.T("pwd_failed"))
		return
	}

	if errMsg, ok := msg.Payload["error"].(string); ok {
		m.Print.Error(m.Tf("pwd_failed_detail", errMsg))
		return
	}

	if currentDir, ok := msg.Payload["current_dir"].(string); ok {
		m.Print.Info(currentDir)
	}
}
