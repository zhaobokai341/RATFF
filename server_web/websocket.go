package main

import (
	"encoding/json"
	"sync"

	"RATFF/shared"

	"github.com/gorilla/websocket"
)

type pendingCommand struct {
	ch chan shared.Message
}

// pendingMu protects the pendingCmd map for concurrent access.
var pendingMu sync.Mutex

// pendingCmd stores pending command responses keyed by client ID.
var pendingCmd = make(map[string]*pendingCommand)

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

// listenResponses reads command responses from the WebSocket and routes them to pending commands.
func listenResponses(conn *websocket.Conn) {
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			return
		}

		var msg shared.Message
		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}

		if msg.Type == shared.MsgResponse {
			pendingMu.Lock()
			if pc, ok := pendingCmd[msg.ClientID]; ok {
				pc.ch <- msg
				delete(pendingCmd, msg.ClientID)
			}
			pendingMu.Unlock()
		}
	}
}
