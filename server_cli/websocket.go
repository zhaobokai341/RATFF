package main

import (
	"encoding/json"
	"sync"
	"time"

	"RATFF/shared"

	"github.com/gorilla/websocket"
)

var (
	responseConnMu sync.Mutex
	responseConn   *websocket.Conn
)

func setResponseConn(conn *websocket.Conn) {
	responseConnMu.Lock()
	responseConn = conn
	responseConnMu.Unlock()
}

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

// startResponseListener keeps the response websocket alive by reconnecting when the server drops it.
func startResponseListener(wsURL string, initialConn *websocket.Conn) {
	conn := initialConn
	for {
		if conn == nil {
			var err error
			conn, err = connectWS(wsURL)
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

// listenResponses reads command responses and routes them to pending commands.
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
			if pc, ok := pendingCmd[msg.ClientID]; ok {
				select {
				case pc.ch <- msg:
				default:
				}
				delete(pendingCmd, msg.ClientID)
			}
			pendingMu.Unlock()
		}
	}
}
