package main

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

func downloadFile(id, remotePath, localPath string) {
	if isRemoteDirectory(id, remotePath) {
		downloadDirectory(id, remotePath, localPath)
		return
	}

	downloadSingleFile(id, remotePath, localPath)
}

func isRemoteDirectory(id, remotePath string) bool {
	msg := waitForCommandResponseRaw(id, shared.CmdFileList, map[string]interface{}{"path": remotePath}, 10*time.Second)
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

func downloadSingleFile(id, remotePath, localPath string) {
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

type remoteEntry struct {
	Name  string
	IsDir bool
}

func listRemoteDir(id, path string) ([]remoteEntry, error) {
	msg := waitForCommandResponseRaw(id, shared.CmdFileList, map[string]interface{}{"path": path}, 10*time.Second)
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

func downloadDirectory(id, remoteDir, localDir string) {
	dirName := filepath.Base(remoteDir)
	if localDir == "" || localDir == "." {
		localDir = dirName
	}

	if err := os.MkdirAll(localDir, 0755); err != nil {
		PrintError(Tf("create_dir_failed", localDir, err))
		return
	}

	PrintInfo(Tf("download_dir_starting", dirName, remoteDir))

	var allFiles []struct {
		RemotePath string
		LocalPath  string
	}

	err := walkRemoteDir(id, remoteDir, localDir, &allFiles)
	if err != nil {
		PrintError(Tf("download_dir_walk_failed", remoteDir, err))
		return
	}

	if len(allFiles) == 0 {
		PrintInfo(Tf("download_dir_empty", dirName))
		return
	}

	PrintInfo(Tf("download_dir_file_count", len(allFiles)))

	successCount := 0
	for i, f := range allFiles {
		relPath, _ := filepath.Rel(localDir, f.LocalPath)
		PrintInfo(Tf("download_dir_file", i+1, len(allFiles), relPath))

		downloadSingleFile(id, f.RemotePath, f.LocalPath)
		successCount++
	}

	PrintSuccess(Tf("download_dir_success", dirName, localDir, successCount, len(allFiles)))
}

func walkRemoteDir(id, remoteDir, localDir string, allFiles *[]struct {
	RemotePath string
	LocalPath  string
}) error {
	entries, err := listRemoteDir(id, remoteDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		remotePath := filepath.Join(remoteDir, entry.Name)
		remotePath = filepath.ToSlash(remotePath)
		localPath := filepath.Join(localDir, entry.Name)

		if entry.IsDir {
			if err := os.MkdirAll(localPath, 0755); err != nil {
				PrintWarn(Tf("create_dir_failed", localPath, err))
				continue
			}
			if err := walkRemoteDir(id, remotePath, localPath, allFiles); err != nil {
				PrintWarn(Tf("download_dir_walk_failed", remotePath, err))
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
