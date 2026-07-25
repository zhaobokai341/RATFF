package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"RATFF/shared"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// handleWebSocket upgrades HTTP to WebSocket and manages the client connection.
func handleWebSocket(manager *ClientManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Error("WebSocket upgrade failed: ", err)
			return
		}
		defer conn.Close()

		setupHeartbeat(conn)

		var msg shared.Message
		if err := readMessage(conn, &msg); err != nil {
			log.Error("Read register message failed: ", err)
			return
		}

		if msg.Type != shared.MsgRegister {
			log.Error("First message must be register")
			return
		}

		clientID := msg.ClientID
		info := shared.ClientInfoFromPayload(msg.Payload)
		info.ID = clientID

		remoteAddr := c.Request.RemoteAddr
		if idx := strings.LastIndex(remoteAddr, ":"); idx != -1 {
			info.IP = remoteAddr[:idx]
		} else {
			info.IP = remoteAddr
		}

		manager.Register(clientID, conn, info)
		defer manager.Unregister(clientID)

		for {
			if err := handleMessage(conn, manager, clientID); err != nil {
				return
			}
		}
	}
}

// setupHeartbeat configures ping/pong for the WebSocket connection.
func setupHeartbeat(conn *websocket.Conn) {
	_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}()
}

// readMessage reads and unmarshals a WebSocket message.
func readMessage(conn *websocket.Conn, msg *shared.Message) error {
	_, data, err := conn.ReadMessage()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, msg)
}

// sendMessage marshals and sends a message over WebSocket.
func sendMessage(conn *websocket.Conn, msg shared.Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return conn.WriteMessage(websocket.TextMessage, data)
}

// handleMessage processes incoming messages from a client connection.
func handleMessage(conn *websocket.Conn, manager *ClientManager, clientID string) error {
	var msg shared.Message
	if err := readMessage(conn, &msg); err != nil {
		return err
	}

	log.WithFields(logrus.Fields{
		"client_id": clientID,
		"type":      msg.Type,
		"command":   msg.Command,
	}).Info("Received message")

	switch msg.Type {
	case shared.MsgHeartbeat:
		return sendMessage(conn, shared.NewMessage(shared.MsgHeartbeat, "", "", nil))

	case shared.MsgCommand:
		targetID := msg.ClientID
		if !manager.IsOnline(targetID) {
			return sendMessage(conn, shared.NewMessage(shared.MsgError, "", "",
				map[string]interface{}{"error": "client offline"}))
		}
		return manager.Send(targetID, msg)

	case shared.MsgResponse:
		manager.Broadcast(msg, clientID)
	}

	return nil
}
