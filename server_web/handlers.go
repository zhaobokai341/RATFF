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

	wsConnMu.Lock()
	if wsConn == nil {
		newConn, err := connectWS(pathPrefix)
		if err != nil {
			wsConnMu.Unlock()
			pendingMu.Lock()
			delete(pendingCmd, req.ClientID)
			pendingMu.Unlock()
			c.JSON(500, gin.H{"error": "websocket not connected: " + err.Error()})
			return
		}
		wsConn = newConn
		go listenResponses(wsConn)
		log.Info("Connected to WebSocket server")
	}
	wsConnMu.Unlock()

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
