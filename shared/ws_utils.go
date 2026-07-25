package shared

import (
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
)

// WSPingInterval is the interval for sending ping messages.
const WSPingInterval = 30 * time.Second

// WSPongTimeout is the timeout for waiting for a pong response.
const WSPongTimeout = 60 * time.Second

// SetupHeartbeat configures ping/pong for the WebSocket connection.
// It sets a read deadline and pong handler, then starts a goroutine
// that sends ping messages at regular intervals.
func SetupHeartbeat(conn *websocket.Conn) {
	_ = conn.SetReadDeadline(time.Now().Add(WSPongTimeout))
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(WSPongTimeout))
		return nil
	})

	go func() {
		ticker := time.NewTicker(WSPingInterval)
		defer ticker.Stop()
		for range ticker.C {
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}()
}

// SendWSMessage marshals and sends a message over WebSocket.
func SendWSMessage(conn *websocket.Conn, msg Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return conn.WriteMessage(websocket.TextMessage, data)
}

// ReadWSMessage reads and unmarshals a WebSocket message.
func ReadWSMessage(conn *websocket.Conn, msg *Message) error {
	_, data, err := conn.ReadMessage()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, msg)
}
