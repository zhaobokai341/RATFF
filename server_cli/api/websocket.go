package api

import (
	"encoding/json"
	"time"

	"RATFF/shared"

	"github.com/gorilla/websocket"
)

// MessageHandler is a callback function for handling incoming messages.
type MessageHandler func(msg shared.Message)

// WebSocketManager manages WebSocket connection and message routing.
type WebSocketManager struct {
	conn    *websocket.Conn
	wsURL   string
	handler MessageHandler
}

// NewWebSocketManager creates a new WebSocket manager.
func NewWebSocketManager(wsURL string, handler MessageHandler) *WebSocketManager {
	return &WebSocketManager{wsURL: wsURL, handler: handler}
}

// Connect establishes a WebSocket connection to the server.
func (m *WebSocketManager) Connect() (*websocket.Conn, error) {
	conn, _, err := websocket.DefaultDialer.Dial(m.wsURL, nil)
	if err != nil {
		return nil, err
	}

	clientID := "__cli__" + shared.GenerateID()[:8]
	registerMsg := shared.NewMessage(shared.MsgRegister, "", clientID, nil)
	data, err := json.Marshal(registerMsg)
	if err != nil {
		return nil, err
	}

	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return nil, err
	}

	m.conn = conn
	return conn, nil
}

// StartResponseListener keeps the response websocket alive by reconnecting when the server drops it.
func (m *WebSocketManager) StartResponseListener(initialConn *websocket.Conn) {
	conn := initialConn
	for {
		if conn == nil {
			var err error
			conn, err = m.Connect()
			if err != nil {
				time.Sleep(2 * time.Second)
				continue
			}
		}

		err := m.listenResponses(conn)
		conn.Close()
		conn = nil
		if err == nil {
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// listenResponses reads command responses and routes them to the handler.
func (m *WebSocketManager) listenResponses(conn *websocket.Conn) error {
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			return err
		}

		var msg shared.Message
		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}

		if msg.Type == shared.MsgResponse || msg.Type == shared.MsgError {
			if m.handler != nil {
				m.handler(msg)
			}
		}
	}
}
