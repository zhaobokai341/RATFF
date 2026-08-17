package shared

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WSPingInterval is the interval for sending ping messages.
const WSPingInterval = 30 * time.Second

// WSPongTimeout is the timeout for waiting for a pong response.
const WSPongTimeout = 60 * time.Second

// WSConn wraps a WebSocket connection with a mutex to prevent concurrent writes.
type WSConn struct {
	*websocket.Conn
	writeMu sync.Mutex
}

// WriteMessage safely writes a message to the WebSocket connection.
func (wc *WSConn) WriteMessage(messageType int, data []byte) error {
	wc.writeMu.Lock()
	defer wc.writeMu.Unlock()
	return wc.Conn.WriteMessage(messageType, data)
}

// NewWSConn creates a new thread-safe WebSocket connection wrapper.
func NewWSConn(conn *websocket.Conn) *WSConn {
	return &WSConn{Conn: conn}
}

// SetupHeartbeat configures ping/pong for the WebSocket connection.
// It sets a read deadline and pong handler, then starts a goroutine
// that sends ping messages at regular intervals. The goroutine exits
// automatically when the connection is closed (WriteMessage fails).
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

// SetupSafeHeartbeat configures ping/pong for a WSConn with thread-safe writes.
func SetupSafeHeartbeat(wc *WSConn) {
	_ = wc.Conn.SetReadDeadline(time.Now().Add(WSPongTimeout))
	wc.Conn.SetPongHandler(func(string) error {
		_ = wc.Conn.SetReadDeadline(time.Now().Add(WSPongTimeout))
		return nil
	})

	go func() {
		ticker := time.NewTicker(WSPingInterval)
		defer ticker.Stop()
		for range ticker.C {
			if err := wc.WriteMessage(websocket.PingMessage, nil); err != nil {
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

// SendSafeWSMessage marshals and sends a message over WebSocket with thread-safe writes.
func SendSafeWSMessage(wc *WSConn, msg Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return wc.WriteMessage(websocket.TextMessage, data)
}

// ReadWSMessage reads and unmarshals a WebSocket message.
func ReadWSMessage(conn *websocket.Conn, msg *Message) error {
	_, data, err := conn.ReadMessage()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, msg)
}
