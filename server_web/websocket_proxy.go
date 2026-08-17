package main

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// upgrader upgrades HTTP connections to WebSocket.
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

// handlePathWebSocket handles WebSocket at /<path>/ws.
func handlePathWebSocket(c *gin.Context) {
	pathPassword := c.Param("pathPassword")
	c.SetCookie("path_prefix", pathPassword, 3600, "/", "", cfg.CookieSecure, true)
	handleWebSocketWithPath(c, pathPassword)
}

// handlePathAPIClients handles API clients endpoint at /<path>/api/clients.
func handlePathAPIClients(c *gin.Context) {
	pathPassword := c.Param("pathPassword")
	c.SetCookie("path_prefix", pathPassword, 3600, "/", "", cfg.CookieSecure, true)
	handleAPIProxyWithPath(c, pathPassword, "/api/clients")
}

// handleWebSocketWithPath upgrades HTTP to WebSocket and proxies traffic with path password.
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

	done := make(chan struct{})
	go func() {
		for {
			_, data, err := clientConn.ReadMessage()
			if err != nil {
				close(done)
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				close(done)
				return
			}
		}
	}()

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			return
		}
		select {
		case <-done:
			return
		default:
		}
		if err := clientConn.WriteMessage(websocket.TextMessage, data); err != nil {
			return
		}
	}
}

// handleAPIProxyWithPath proxies API requests to server_api with path password.
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
