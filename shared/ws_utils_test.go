package shared

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

var testUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func TestSetupHeartbeat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := testUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		SetupHeartbeat(conn)

		// Wait for a ping to arrive
		time.Sleep(100 * time.Millisecond)
		conn.Close()
	}))
	defer server.Close()

	// Convert http:// to ws://
	wsURL := "ws" + server.URL[4:] + "/ws"

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Set pong handler
	conn.SetPongHandler(func(string) error {
		return nil
	})

	// Read messages (should receive ping)
	_ = conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func TestSendWSMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := testUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		_, data, err := conn.ReadMessage()
		if err != nil {
			t.Errorf("Failed to read message: %v", err)
			return
		}

		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			t.Errorf("Failed to unmarshal: %v", err)
			return
		}

		if msg.Type != MsgHeartbeat {
			t.Errorf("Expected heartbeat, got %s", msg.Type)
		}
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[4:] + "/ws"

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	msg := NewMessage(MsgHeartbeat, "", "", nil)
	if err := SendWSMessage(conn, msg); err != nil {
		t.Errorf("Failed to send message: %v", err)
	}
}

func TestReadWSMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := testUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		msg := NewMessage(MsgRegister, "", "test-client", nil)
		if err := SendWSMessage(conn, msg); err != nil {
			t.Errorf("Failed to send: %v", err)
			return
		}

		time.Sleep(50 * time.Millisecond)
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[4:] + "/ws"

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	var msg Message
	if err := ReadWSMessage(conn, &msg); err != nil {
		t.Fatalf("Failed to read message: %v", err)
	}

	if msg.Type != MsgRegister {
		t.Errorf("Expected register, got %s", msg.Type)
	}
	if msg.ClientID != "test-client" {
		t.Errorf("Expected test-client, got %s", msg.ClientID)
	}
}

func TestReadWSMessageConnectionClosed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := testUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		conn.Close()
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[4:] + "/ws"

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	var msg Message
	err = ReadWSMessage(conn, &msg)
	if err == nil {
		t.Error("Expected error when connection closed")
	}
}
