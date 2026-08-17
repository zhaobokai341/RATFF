package main

import (
	"encoding/json"
	"sync"
	"time"

	"RATFF/shared"

	"github.com/gorilla/websocket"
)

type pendingCommand struct {
	ch chan shared.Message
}

// pendingMu protects the pendingCmd map for concurrent access.
var pendingMu sync.Mutex

// pendingCmd stores pending command responses keyed by client ID.
// Each client can have multiple pending commands (queue).
var pendingCmd = make(map[string][]*pendingCommand)

// cleanupPending removes the first pending command for a client (FIFO).
func cleanupPending(clientID string) {
	pendingMu.Lock()
	defer pendingMu.Unlock()
	queue := pendingCmd[clientID]
	if len(queue) > 0 {
		pendingCmd[clientID] = queue[1:]
		if len(pendingCmd[clientID]) == 0 {
			delete(pendingCmd, clientID)
		}
	}
}

// connectWS establishes a WebSocket connection to the server for receiving responses.
func connectWS(pathPassword string) (*websocket.Conn, error) {
	wsURL := buildWSURL(pathPassword)
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
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

	return conn, nil
}

func setResponseConn(conn *websocket.Conn) {
	wsConnMu.Lock()
	wsConn = conn
	wsConnMu.Unlock()
}

func ensureResponseConn(pathPassword string) (*websocket.Conn, error) {
	wsConnMu.Lock()
	if wsConn != nil {
		conn := wsConn
		wsConnMu.Unlock()
		return conn, nil
	}
	wsConnMu.Unlock()

	conn, err := connectWS(pathPassword)
	if err != nil {
		return nil, err
	}
	setResponseConn(conn)
	go startResponseListener(pathPassword, conn)
	return conn, nil
}

// startResponseListener keeps the response websocket alive by reconnecting when the server drops it.
func startResponseListener(pathPassword string, initialConn *websocket.Conn) {
	conn := initialConn
	for {
		if conn == nil {
			var err error
			conn, err = connectWS(pathPassword)
			if err != nil {
				time.Sleep(2 * time.Second)
				continue
			}
			setResponseConn(conn)
		}

		err := listenResponses(conn)
		conn.Close()
		setResponseConn(nil)
		conn = nil
		if err == nil {
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// listenResponses reads command responses from the WebSocket and routes them to pending commands.
func listenResponses(conn *websocket.Conn) error {
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
			pendingMu.Lock()
			if queue, ok := pendingCmd[msg.ClientID]; ok && len(queue) > 0 {
				pc := queue[0]
				pendingCmd[msg.ClientID] = queue[1:]
				if len(pendingCmd[msg.ClientID]) == 0 {
					delete(pendingCmd, msg.ClientID)
				}
				select {
				case pc.ch <- msg:
				default:
				}
			}
			pendingMu.Unlock()
		}
	}
}
