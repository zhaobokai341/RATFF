package client

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

const (
	chunkSize    = 64 * 1024
	chunkTimeout = 30 * time.Second
)

// UploadFile uploads a file or directory to the remote client.
func (m *Manager) UploadFile(id, localPath, remotePath string, newProgressBar ProgressBarCreator) {
	stat, err := os.Stat(localPath)
	if err != nil {
		m.Print.Error(m.Tf("file_not_exist", localPath))
		return
	}

	if stat.IsDir() {
		m.uploadDirectory(id, localPath, remotePath, newProgressBar)
		return
	}

	m.uploadSingleFile(id, localPath, remotePath, newProgressBar)
}

func (m *Manager) uploadSingleFile(id, localPath, remotePath string, newProgressBar ProgressBarCreator) {
	file, err := os.Open(localPath)
	if err != nil {
		m.Print.Error(m.Tf("file_open_failed", localPath, err))
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		m.Print.Error(m.Tf("file_open_failed", localPath, err))
		return
	}

	fileSize := stat.Size()
	totalChunks := (fileSize + chunkSize - 1) / chunkSize
	fileID := shared.GenerateID()

	filename := filepath.Base(localPath)

	if remotePath == "" || remotePath == "." {
		remotePath = filename
	}

	m.Print.Info(m.Tf("upload_starting", filename, m.Print.FormatBytes(fileSize)))

	payload := map[string]interface{}{
		"file_id":      fileID,
		"remote_path":  remotePath,
		"file_size":    fileSize,
		"chunk_size":   chunkSize,
		"total_chunks": totalChunks,
	}

	if !m.WaitForResponse(id, shared.CmdFileUploadStart, payload, 10*time.Second) {
		return
	}

	progressBar := newProgressBar(fileSize, filename)

	md5hash := md5.New()

	for i := 0; i < int(totalChunks); i++ {
		chunk := make([]byte, chunkSize)
		n, err := file.Read(chunk)
		if err != nil && err != io.EOF {
			progressBar.MarkDone()
			progressBar.Display()
			m.Print.Error(m.Tf("file_read_failed", err))
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

		if !m.WaitForResponse(id, shared.CmdFileUploadChunk, chunkPayload, chunkTimeout) {
			progressBar.MarkDone()
			progressBar.Display()
			m.Print.Error(m.Tf("upload_chunk_failed", i))
			return
		}

		progressBar.Add(int64(n))
		progressBar.Display()
	}

	completePayload := map[string]interface{}{
		"file_id": fileID,
	}

	msg := m.WaitForResponseWithMsg(id, shared.CmdFileUploadComplete, completePayload, 10*time.Second)
	if msg == nil {
		progressBar.MarkDone()
		progressBar.Display()
		m.Print.Error(m.T("upload_complete_failed"))
		return
	}

	progressBar.MarkDone()
	progressBar.Display()

	if msg.Payload != nil {
		if remoteMD5, ok := msg.Payload["md5"].(string); ok {
			localMD5 := hex.EncodeToString(md5hash.Sum(nil))
			if remoteMD5 != localMD5 {
				m.Print.Warn(m.Tf("upload_md5_mismatch", localMD5, remoteMD5))
			} else {
				m.Print.Success(m.Tf("upload_success", filename, remotePath))
			}
		} else {
			m.Print.Success(m.Tf("upload_success", filename, remotePath))
		}
	}
}

func (m *Manager) uploadDirectory(id, localDir, remoteDir string, newProgressBar ProgressBarCreator) {
	dirName := filepath.Base(localDir)
	if remoteDir == "" || remoteDir == "." {
		remoteDir = dirName
	} else {
		remoteDir = filepath.Join(remoteDir, dirName)
	}

	var files []string
	err := filepath.Walk(localDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			m.Print.Warn(m.Tf("upload_dir_walk_error", path, err))
			return nil
		}
		if info.IsDir() {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		m.Print.Error(m.Tf("upload_dir_walk_failed", localDir, err))
		return
	}

	if len(files) == 0 {
		m.Print.Info(m.Tf("upload_dir_empty", dirName))
		return
	}

	m.Print.Info(m.Tf("upload_dir_starting", dirName, len(files)))

	successCount := 0
	for i, filePath := range files {
		relPath, err := filepath.Rel(localDir, filePath)
		if err != nil {
			m.Print.Warn(m.Tf("upload_dir_relpath_failed", filePath, err))
			continue
		}

		remoteFilePath := filepath.Join(remoteDir, relPath)
		remoteFilePath = filepath.ToSlash(remoteFilePath)

		m.Print.Info(m.Tf("upload_dir_file", i+1, len(files), relPath))

		m.uploadSingleFile(id, filePath, remoteFilePath, newProgressBar)
		successCount++
	}

	m.Print.Success(m.Tf("upload_dir_success", dirName, remoteDir, successCount, len(files)))
}
