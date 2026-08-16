package main

import (
	"crypto/md5"
	"encoding/base64"
	"hash"
	"os"
	"path/filepath"
	"testing"

	"RATFF/shared"
)

func TestHandleFileUploadStart(t *testing.T) {
	log = shared.InitLogger("error", "text")

	tmpDir := t.TempDir()
	remotePath := filepath.Join(tmpDir, "test.txt")

	msg := shared.NewMessage(shared.MsgCommand, shared.CmdFileUploadStart, "test-client", map[string]interface{}{
		"file_id":      "test-file-1",
		"remote_path":  remotePath,
		"file_size":    100.0,
		"chunk_size":   64.0,
		"total_chunks": 1.0,
	})

	resp := executeCommand(msg)

	if resp.Type != shared.MsgResponse {
		t.Errorf("Expected response, got %s", resp.Type)
	}

	if resp.Payload["status"] != "ready" {
		t.Errorf("Expected status ready, got %v", resp.Payload["status"])
	}

	uploadMu.Lock()
	_, exists := uploadSessions["test-file-1"]
	uploadMu.Unlock()

	if !exists {
		t.Error("Expected upload session to exist")
	}

	uploadSessions["test-file-1"].File.Close()
	delete(uploadSessions, "test-file-1")
}

func TestHandleFileUploadStartMissingFields(t *testing.T) {
	log = shared.InitLogger("error", "text")

	msg := shared.NewMessage(shared.MsgCommand, shared.CmdFileUploadStart, "test-client", map[string]interface{}{
		"file_id": "",
	})

	resp := executeCommand(msg)

	if resp.Type != shared.MsgError {
		t.Errorf("Expected error, got %s", resp.Type)
	}
}

func TestHandleFileUploadChunk(t *testing.T) {
	log = shared.InitLogger("error", "text")

	tmpDir := t.TempDir()
	remotePath := filepath.Join(tmpDir, "test.txt")

	file, _ := os.Create(remotePath)
	uploadMu.Lock()
	uploadSessions["test-file-2"] = &UploadSession{
		FileID:      "test-file-2",
		RemotePath:  remotePath,
		File:        file,
		TotalChunks: 1,
		MD5Hash:     newMD5Hash(),
	}
	uploadMu.Unlock()

	chunkData := []byte("hello world")
	chunkB64 := base64.StdEncoding.EncodeToString(chunkData)

	msg := shared.NewMessage(shared.MsgCommand, shared.CmdFileUploadChunk, "test-client", map[string]interface{}{
		"file_id":     "test-file-2",
		"chunk_index": 0.0,
		"chunk_data":  chunkB64,
	})

	resp := executeCommand(msg)

	if resp.Type != shared.MsgResponse {
		t.Errorf("Expected response, got %s", resp.Type)
	}

	uploadMu.Lock()
	session := uploadSessions["test-file-2"]
	uploadMu.Unlock()

	if session.ReceivedChunks != 1 {
		t.Errorf("Expected 1 received chunk, got %d", session.ReceivedChunks)
	}

	session.File.Close()
	delete(uploadSessions, "test-file-2")

	content, _ := os.ReadFile(remotePath)
	if string(content) != "hello world" {
		t.Errorf("Expected 'hello world', got %s", string(content))
	}
}

func TestHandleFileUploadChunkNotFound(t *testing.T) {
	log = shared.InitLogger("error", "text")

	msg := shared.NewMessage(shared.MsgCommand, shared.CmdFileUploadChunk, "test-client", map[string]interface{}{
		"file_id":     "nonexistent",
		"chunk_index": 0.0,
		"chunk_data":  "dGVzdA==",
	})

	resp := executeCommand(msg)

	if resp.Type != shared.MsgError {
		t.Errorf("Expected error, got %s", resp.Type)
	}
}

func TestHandleFileUploadComplete(t *testing.T) {
	log = shared.InitLogger("error", "text")

	tmpDir := t.TempDir()
	remotePath := filepath.Join(tmpDir, "test.txt")

	file, _ := os.Create(remotePath)
	uploadMu.Lock()
	uploadSessions["test-file-3"] = &UploadSession{
		FileID:       "test-file-3",
		RemotePath:   remotePath,
		File:         file,
		TotalChunks:  1,
		ReceivedSize: 100,
		MD5Hash:      newMD5Hash(),
	}
	uploadMu.Unlock()

	msg := shared.NewMessage(shared.MsgCommand, shared.CmdFileUploadComplete, "test-client", map[string]interface{}{
		"file_id": "test-file-3",
	})

	resp := executeCommand(msg)

	if resp.Type != shared.MsgResponse {
		t.Errorf("Expected response, got %s", resp.Type)
	}

	if resp.Payload["status"] != "complete" {
		t.Errorf("Expected status complete, got %v", resp.Payload["status"])
	}

	uploadMu.Lock()
	_, exists := uploadSessions["test-file-3"]
	uploadMu.Unlock()

	if exists {
		t.Error("Expected upload session to be deleted")
	}
}

func TestHandleFileDownloadStart(t *testing.T) {
	log = shared.InitLogger("error", "text")

	tmpDir := t.TempDir()
	localPath := filepath.Join(tmpDir, "download.txt")
	if err := os.WriteFile(localPath, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	msg := shared.NewMessage(shared.MsgCommand, shared.CmdFileDownloadStart, "test-client", map[string]interface{}{
		"file_id":    "test-file-4",
		"local_path": localPath,
	})

	resp := executeCommand(msg)

	if resp.Type != shared.MsgResponse {
		t.Errorf("Expected response, got %s", resp.Type)
	}

	if resp.Payload["status"] != "ready" {
		t.Errorf("Expected status ready, got %v", resp.Payload["status"])
	}

	downloadMu.Lock()
	_, exists := downloadSessions["test-file-4"]
	downloadMu.Unlock()

	if !exists {
		t.Error("Expected download session to exist")
	}

	downloadSessions["test-file-4"].File.Close()
	delete(downloadSessions, "test-file-4")
}

func TestHandleFileDownloadStartFileNotFound(t *testing.T) {
	log = shared.InitLogger("error", "text")

	msg := shared.NewMessage(shared.MsgCommand, shared.CmdFileDownloadStart, "test-client", map[string]interface{}{
		"file_id":    "test-file-5",
		"local_path": "/nonexistent/file.txt",
	})

	resp := executeCommand(msg)

	if resp.Type != shared.MsgError {
		t.Errorf("Expected error, got %s", resp.Type)
	}
}

func TestHandleFileDownloadChunk(t *testing.T) {
	log = shared.InitLogger("error", "text")

	tmpDir := t.TempDir()
	localPath := filepath.Join(tmpDir, "download.txt")
	if err := os.WriteFile(localPath, []byte("hello world test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	file, _ := os.Open(localPath)
	downloadMu.Lock()
	downloadSessions["test-file-6"] = &DownloadSession{
		FileID:      "test-file-6",
		LocalPath:   localPath,
		File:        file,
		TotalChunks: 1,
		TotalSize:   24,
		MD5Hash:     newMD5Hash(),
	}
	downloadMu.Unlock()

	msg := shared.NewMessage(shared.MsgCommand, shared.CmdFileDownloadChunk, "test-client", map[string]interface{}{
		"file_id":     "test-file-6",
		"chunk_index": 0.0,
	})

	resp := executeCommand(msg)

	if resp.Type != shared.MsgResponse {
		t.Errorf("Expected response, got %s", resp.Type)
	}

	chunkDataB64, _ := resp.Payload["chunk_data"].(string)
	chunkData, _ := base64.StdEncoding.DecodeString(chunkDataB64)

	if string(chunkData) != "hello world test content" {
		t.Errorf("Expected 'hello world test content', got %s", string(chunkData))
	}

	downloadSessions["test-file-6"].File.Close()
	delete(downloadSessions, "test-file-6")
}

func TestHandleFileDownloadComplete(t *testing.T) {
	log = shared.InitLogger("error", "text")

	tmpDir := t.TempDir()
	localPath := filepath.Join(tmpDir, "download.txt")
	if err := os.WriteFile(localPath, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	file, _ := os.Open(localPath)
	downloadMu.Lock()
	downloadSessions["test-file-7"] = &DownloadSession{
		FileID:    "test-file-7",
		LocalPath: localPath,
		File:      file,
		TotalSize: 4,
		SentSize:  4,
		MD5Hash:   newMD5Hash(),
	}
	downloadMu.Unlock()

	msg := shared.NewMessage(shared.MsgCommand, shared.CmdFileDownloadComplete, "test-client", map[string]interface{}{
		"file_id": "test-file-7",
	})

	resp := executeCommand(msg)

	if resp.Type != shared.MsgResponse {
		t.Errorf("Expected response, got %s", resp.Type)
	}

	if resp.Payload["status"] != "complete" {
		t.Errorf("Expected status complete, got %v", resp.Payload["status"])
	}

	downloadMu.Lock()
	_, exists := downloadSessions["test-file-7"]
	downloadMu.Unlock()

	if exists {
		t.Error("Expected download session to be deleted")
	}
}

func newMD5Hash() hash.Hash {
	return md5.New()
}

func TestHandleFileCopyFile(t *testing.T) {
	log = shared.InitLogger("error", "text")

	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "source.txt")
	if err := os.WriteFile(srcPath, []byte("copy test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	dstPath := filepath.Join(tmpDir, "dest.txt")

	msg := shared.NewMessage(shared.MsgCommand, shared.CmdFileCopy, "test-client", map[string]interface{}{
		"origin_path": srcPath,
		"new_path":    dstPath,
	})

	resp := executeCommand(msg)

	if resp.Type != shared.MsgResponse {
		t.Errorf("Expected response, got %s", resp.Type)
	}

	if resp.Payload["origin_path"] != srcPath {
		t.Errorf("Expected origin_path %s, got %v", srcPath, resp.Payload["origin_path"])
	}

	content, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}

	if string(content) != "copy test content" {
		t.Errorf("Expected 'copy test content', got %s", string(content))
	}

	srcContent, err := os.ReadFile(srcPath)
	if err != nil {
		t.Fatalf("Source file should still exist: %v", err)
	}

	if string(srcContent) != "copy test content" {
		t.Errorf("Source file should not be modified")
	}
}

func TestHandleFileCopyToDir(t *testing.T) {
	log = shared.InitLogger("error", "text")

	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "source.txt")
	if err := os.WriteFile(srcPath, []byte("copy to dir"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	destDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(destDir, 0755); err != nil {
		t.Fatalf("Failed to create dest dir: %v", err)
	}

	msg := shared.NewMessage(shared.MsgCommand, shared.CmdFileCopy, "test-client", map[string]interface{}{
		"origin_path": srcPath,
		"new_path":    destDir,
	})

	resp := executeCommand(msg)

	if resp.Type != shared.MsgResponse {
		t.Errorf("Expected response, got %s", resp.Type)
	}

	expectedDst := filepath.Join(destDir, "source.txt")
	content, err := os.ReadFile(expectedDst)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}

	if string(content) != "copy to dir" {
		t.Errorf("Expected 'copy to dir', got %s", string(content))
	}
}

func TestHandleFileCopyDir(t *testing.T) {
	log = shared.InitLogger("error", "text")

	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "srcdir")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatalf("Failed to create src dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(srcDir, "a.txt"), []byte("file a"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	if err := os.MkdirAll(filepath.Join(srcDir, "sub"), 0755); err != nil {
		t.Fatalf("Failed to create sub dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(srcDir, "sub", "b.txt"), []byte("file b"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	dstDir := filepath.Join(tmpDir, "dstdir")

	msg := shared.NewMessage(shared.MsgCommand, shared.CmdFileCopy, "test-client", map[string]interface{}{
		"origin_path": srcDir,
		"new_path":    dstDir,
	})

	resp := executeCommand(msg)

	if resp.Type != shared.MsgResponse {
		t.Errorf("Expected response, got %s", resp.Type)
	}

	contentA, err := os.ReadFile(filepath.Join(dstDir, "a.txt"))
	if err != nil {
		t.Fatalf("Failed to read a.txt: %v", err)
	}

	if string(contentA) != "file a" {
		t.Errorf("Expected 'file a', got %s", string(contentA))
	}

	contentB, err := os.ReadFile(filepath.Join(dstDir, "sub", "b.txt"))
	if err != nil {
		t.Fatalf("Failed to read b.txt: %v", err)
	}

	if string(contentB) != "file b" {
		t.Errorf("Expected 'file b', got %s", string(contentB))
	}

	if _, err := os.Stat(srcDir); err != nil {
		t.Errorf("Source directory should still exist: %v", err)
	}
}

func TestHandleFileCopyNotFound(t *testing.T) {
	log = shared.InitLogger("error", "text")

	msg := shared.NewMessage(shared.MsgCommand, shared.CmdFileCopy, "test-client", map[string]interface{}{
		"origin_path": "/nonexistent/path/file.txt",
		"new_path":    "/tmp/dest.txt",
	})

	resp := executeCommand(msg)

	if resp.Type != shared.MsgError {
		t.Errorf("Expected error, got %s", resp.Type)
	}
}

func TestHandleFileCopyMissingFields(t *testing.T) {
	log = shared.InitLogger("error", "text")

	msg := shared.NewMessage(shared.MsgCommand, shared.CmdFileCopy, "test-client", map[string]interface{}{
		"origin_path": "",
		"new_path":    "",
	})

	resp := executeCommand(msg)

	if resp.Type != shared.MsgError {
		t.Errorf("Expected error, got %s", resp.Type)
	}
}
