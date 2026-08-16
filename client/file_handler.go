package main

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"sync"

	"RATFF/shared"
)

type UploadSession struct {
	FileID         string
	RemotePath     string
	File           *os.File
	TotalChunks    int
	ReceivedChunks int
	ExpectedSize   int64
	ReceivedSize   int64
	MD5Hash        hash.Hash
}

type DownloadSession struct {
	FileID      string
	LocalPath   string
	File        *os.File
	TotalChunks int
	SentChunks  int
	TotalSize   int64
	SentSize    int64
	MD5Hash     hash.Hash
}

var (
	uploadSessions   = make(map[string]*UploadSession)
	uploadMu         sync.Mutex
	downloadSessions = make(map[string]*DownloadSession)
	downloadMu       sync.Mutex
)

func handleFileUploadStart(msg shared.Message) shared.Message {
	fileID, _ := msg.Payload["file_id"].(string)
	remotePath, _ := msg.Payload["remote_path"].(string)
	fileSize, _ := msg.Payload["file_size"].(float64)
	totalChunks, _ := msg.Payload["total_chunks"].(float64)

	if fileID == "" || remotePath == "" {
		return shared.NewMessage(shared.MsgError, shared.CmdFileUploadStart, msg.ClientID,
			map[string]interface{}{"error": "missing file_id or remote_path"})
	}

	dir := filepath.Dir(remotePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return shared.NewMessage(shared.MsgError, shared.CmdFileUploadStart, msg.ClientID,
			map[string]interface{}{"error": fmt.Sprintf("create directory failed: %v", err)})
	}

	file, err := os.Create(remotePath)
	if err != nil {
		return shared.NewMessage(shared.MsgError, shared.CmdFileUploadStart, msg.ClientID,
			map[string]interface{}{"error": fmt.Sprintf("create file failed: %v", err)})
	}

	uploadMu.Lock()
	if old, exists := uploadSessions[fileID]; exists {
		old.File.Close()
	}
	uploadSessions[fileID] = &UploadSession{
		FileID:       fileID,
		RemotePath:   remotePath,
		File:         file,
		TotalChunks:  int(totalChunks),
		ExpectedSize: int64(fileSize),
		MD5Hash:      md5.New(),
	}
	uploadMu.Unlock()

	return shared.NewMessage(shared.MsgResponse, shared.CmdFileUploadStart, msg.ClientID,
		map[string]interface{}{"status": "ready", "file_id": fileID})
}

func handleFileUploadChunk(msg shared.Message) shared.Message {
	fileID, _ := msg.Payload["file_id"].(string)
	chunkDataB64, _ := msg.Payload["chunk_data"].(string)
	chunkIndex, _ := msg.Payload["chunk_index"].(float64)

	chunkData, err := base64.StdEncoding.DecodeString(chunkDataB64)
	if err != nil {
		return shared.NewMessage(shared.MsgError, shared.CmdFileUploadChunk, msg.ClientID,
			map[string]interface{}{"error": fmt.Sprintf("decode chunk failed: %v", err)})
	}

	uploadMu.Lock()
	session, exists := uploadSessions[fileID]
	if !exists {
		uploadMu.Unlock()
		return shared.NewMessage(shared.MsgError, shared.CmdFileUploadChunk, msg.ClientID,
			map[string]interface{}{"error": "upload session not found"})
	}

	n, err := session.File.Write(chunkData)
	if err != nil {
		uploadMu.Unlock()
		return shared.NewMessage(shared.MsgError, shared.CmdFileUploadChunk, msg.ClientID,
			map[string]interface{}{"error": fmt.Sprintf("write chunk failed: %v", err)})
	}

	session.MD5Hash.Write(chunkData[:n])
	session.ReceivedSize += int64(n)
	session.ReceivedChunks++
	uploadMu.Unlock()

	return shared.NewMessage(shared.MsgResponse, shared.CmdFileUploadChunk, msg.ClientID,
		map[string]interface{}{
			"status":      "ok",
			"file_id":     fileID,
			"chunk_index": int(chunkIndex),
		})
}

func handleFileUploadComplete(msg shared.Message) shared.Message {
	fileID, _ := msg.Payload["file_id"].(string)

	uploadMu.Lock()
	session, exists := uploadSessions[fileID]
	delete(uploadSessions, fileID)
	uploadMu.Unlock()

	if !exists {
		return shared.NewMessage(shared.MsgError, shared.CmdFileUploadComplete, msg.ClientID,
			map[string]interface{}{"error": "upload session not found"})
	}

	if err := session.File.Close(); err != nil {
		return shared.NewMessage(shared.MsgError, shared.CmdFileUploadComplete, msg.ClientID,
			map[string]interface{}{"error": fmt.Sprintf("close file failed: %v", err)})
	}

	uploadMu.Lock()
	receivedSize := session.ReceivedSize
	md5sum := hex.EncodeToString(session.MD5Hash.Sum(nil))
	uploadMu.Unlock()

	return shared.NewMessage(shared.MsgResponse, shared.CmdFileUploadComplete, msg.ClientID,
		map[string]interface{}{
			"status":    "complete",
			"file_id":   fileID,
			"file_size": receivedSize,
			"md5":       md5sum,
		})
}

func handleFileDownloadStart(msg shared.Message) shared.Message {
	fileID, _ := msg.Payload["file_id"].(string)
	localPath, _ := msg.Payload["local_path"].(string)

	if fileID == "" || localPath == "" {
		return shared.NewMessage(shared.MsgError, shared.CmdFileDownloadStart, msg.ClientID,
			map[string]interface{}{"error": "missing file_id or local_path"})
	}

	file, err := os.Open(localPath)
	if err != nil {
		return shared.NewMessage(shared.MsgError, shared.CmdFileDownloadStart, msg.ClientID,
			map[string]interface{}{"error": fmt.Sprintf("open file failed: %v", err)})
	}

	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return shared.NewMessage(shared.MsgError, shared.CmdFileDownloadStart, msg.ClientID,
			map[string]interface{}{"error": fmt.Sprintf("stat file failed: %v", err)})
	}

	chunkSize := int64(64 * 1024)
	totalChunks := (stat.Size() + chunkSize - 1) / chunkSize

	downloadMu.Lock()
	if old, exists := downloadSessions[fileID]; exists {
		old.File.Close()
	}
	downloadSessions[fileID] = &DownloadSession{
		FileID:      fileID,
		LocalPath:   localPath,
		File:        file,
		TotalChunks: int(totalChunks),
		TotalSize:   stat.Size(),
		MD5Hash:     md5.New(),
	}
	downloadMu.Unlock()

	return shared.NewMessage(shared.MsgResponse, shared.CmdFileDownloadStart, msg.ClientID,
		map[string]interface{}{
			"status":       "ready",
			"file_id":      fileID,
			"file_size":    stat.Size(),
			"chunk_size":   chunkSize,
			"total_chunks": totalChunks,
		})
}

func handleFileDownloadChunk(msg shared.Message) shared.Message {
	fileID, _ := msg.Payload["file_id"].(string)
	chunkIndex, _ := msg.Payload["chunk_index"].(float64)

	downloadMu.Lock()
	session, exists := downloadSessions[fileID]
	if !exists {
		downloadMu.Unlock()
		return shared.NewMessage(shared.MsgError, shared.CmdFileDownloadChunk, msg.ClientID,
			map[string]interface{}{"error": "download session not found"})
	}

	chunkSize := int64(64 * 1024)
	offset := int64(chunkIndex) * chunkSize

	_, err := session.File.Seek(offset, io.SeekStart)
	if err != nil {
		downloadMu.Unlock()
		return shared.NewMessage(shared.MsgError, shared.CmdFileDownloadChunk, msg.ClientID,
			map[string]interface{}{"error": fmt.Sprintf("seek file failed: %v", err)})
	}

	chunk := make([]byte, chunkSize)
	n, err := session.File.Read(chunk)
	if err != nil && err != io.EOF {
		downloadMu.Unlock()
		return shared.NewMessage(shared.MsgError, shared.CmdFileDownloadChunk, msg.ClientID,
			map[string]interface{}{"error": fmt.Sprintf("read chunk failed: %v", err)})
	}

	chunkData := chunk[:n]
	session.MD5Hash.Write(chunkData)
	session.SentSize += int64(n)
	session.SentChunks++
	downloadMu.Unlock()

	chunkB64 := base64.StdEncoding.EncodeToString(chunkData)

	return shared.NewMessage(shared.MsgResponse, shared.CmdFileDownloadChunk, msg.ClientID,
		map[string]interface{}{
			"status":      "ok",
			"file_id":     fileID,
			"chunk_index": int(chunkIndex),
			"chunk_data":  chunkB64,
			"chunk_size":  n,
		})
}

func handleFileDownloadComplete(msg shared.Message) shared.Message {
	fileID, _ := msg.Payload["file_id"].(string)

	downloadMu.Lock()
	session, exists := downloadSessions[fileID]
	delete(downloadSessions, fileID)
	downloadMu.Unlock()

	if !exists {
		return shared.NewMessage(shared.MsgError, shared.CmdFileDownloadComplete, msg.ClientID,
			map[string]interface{}{"error": "download session not found"})
	}

	if err := session.File.Close(); err != nil {
		return shared.NewMessage(shared.MsgError, shared.CmdFileDownloadComplete, msg.ClientID,
			map[string]interface{}{"error": fmt.Sprintf("close file failed: %v", err)})
	}

	downloadMu.Lock()
	sentSize := session.SentSize
	md5sum := hex.EncodeToString(session.MD5Hash.Sum(nil))
	downloadMu.Unlock()

	return shared.NewMessage(shared.MsgResponse, shared.CmdFileDownloadComplete, msg.ClientID,
		map[string]interface{}{
			"status":    "complete",
			"file_id":   fileID,
			"file_size": sentSize,
			"md5":       md5sum,
		})
}
