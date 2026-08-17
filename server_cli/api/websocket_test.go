package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"RATFF/shared"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

var testUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func TestWebSocketManagerConnectSuccess(t *testing.T) {
	var mu sync.Mutex
	var receivedMsg string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := testUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		_, data, _ := conn.ReadMessage()
		mu.Lock()
		receivedMsg = string(data)
		mu.Unlock()

		time.Sleep(100 * time.Millisecond)
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[4:] + "/ws"

	wsManager := NewWebSocketManager(wsURL, nil)

	conn, err := wsManager.Connect()
	assert.NoError(t, err)
	assert.NotNil(t, conn)

	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	assert.True(t, strings.Contains(receivedMsg, "register"))
	mu.Unlock()
}

func TestWebSocketManagerConnectInvalidURL(t *testing.T) {
	wsManager := NewWebSocketManager("ws://invalid-url-that-does-not-exist-12345:9999/ws", nil)

	conn, err := wsManager.Connect()
	assert.Error(t, err)
	assert.Nil(t, conn)
}

func TestWebSocketManagerListenResponses(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := testUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		msg := shared.NewMessage(shared.MsgResponse, "", "client-1", map[string]interface{}{
			"result": "success",
		})
		data, _ := json.Marshal(msg)
		conn.WriteMessage(websocket.TextMessage, data)

		errMsg := shared.NewMessage(shared.MsgError, "", "client-1", map[string]interface{}{
			"error": "test error",
		})
		data, _ = json.Marshal(errMsg)
		conn.WriteMessage(websocket.TextMessage, data)

		time.Sleep(100 * time.Millisecond)
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[4:] + "/ws"

	var receivedMessages []shared.Message
	var mu sync.Mutex

	handler := func(msg shared.Message) {
		mu.Lock()
		defer mu.Unlock()
		receivedMessages = append(receivedMessages, msg)
	}

	wsManager := NewWebSocketManager(wsURL, handler)

	conn, err := wsManager.Connect()
	assert.NoError(t, err)

	go func() {
		wsManager.listenResponses(conn)
	}()

	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	assert.Len(t, receivedMessages, 2)
	mu.Unlock()
}

func TestWebSocketManagerListenResponsesInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := testUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		conn.WriteMessage(websocket.TextMessage, []byte(`{invalid json`))

		msg := shared.NewMessage(shared.MsgResponse, "", "client-1", nil)
		data, _ := json.Marshal(msg)
		conn.WriteMessage(websocket.TextMessage, data)

		time.Sleep(100 * time.Millisecond)
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[4:] + "/ws"

	var receivedCount int
	var mu sync.Mutex

	handler := func(msg shared.Message) {
		mu.Lock()
		defer mu.Unlock()
		receivedCount++
	}

	wsManager := NewWebSocketManager(wsURL, handler)

	conn, err := wsManager.Connect()
	assert.NoError(t, err)

	go func() {
		wsManager.listenResponses(conn)
	}()

	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	assert.Equal(t, 1, receivedCount)
	mu.Unlock()
}

func TestWebSocketManagerListenResponsesIgnoresNonResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := testUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		msg := shared.NewMessage(shared.MsgHeartbeat, "", "client-1", nil)
		data, _ := json.Marshal(msg)
		conn.WriteMessage(websocket.TextMessage, data)

		time.Sleep(100 * time.Millisecond)
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[4:] + "/ws"

	var receivedCount int
	var mu sync.Mutex

	handler := func(msg shared.Message) {
		mu.Lock()
		defer mu.Unlock()
		receivedCount++
	}

	wsManager := NewWebSocketManager(wsURL, handler)

	conn, err := wsManager.Connect()
	assert.NoError(t, err)

	go func() {
		wsManager.listenResponses(conn)
	}()

	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	assert.Equal(t, 0, receivedCount)
	mu.Unlock()
}

func TestWebSocketManagerStartResponseListener(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := testUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		msg := shared.NewMessage(shared.MsgResponse, "", "client-1", nil)
		data, _ := json.Marshal(msg)
		conn.WriteMessage(websocket.TextMessage, data)

		time.Sleep(100 * time.Millisecond)
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[4:] + "/ws"

	var receivedCount int
	var mu sync.Mutex

	handler := func(msg shared.Message) {
		mu.Lock()
		defer mu.Unlock()
		receivedCount++
	}

	wsManager := NewWebSocketManager(wsURL, handler)

	conn, err := wsManager.Connect()
	assert.NoError(t, err)

	go wsManager.StartResponseListener(conn)

	time.Sleep(200 * time.Millisecond)
}

func TestWebSocketManagerSendSuccess(t *testing.T) {
	var messages []string
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := testUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		for i := 0; i < 2; i++ {
			_, data, err := conn.ReadMessage()
			if err != nil {
				break
			}
			mu.Lock()
			messages = append(messages, string(data))
			mu.Unlock()
		}

		time.Sleep(100 * time.Millisecond)
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[4:] + "/ws"

	wsManager := NewWebSocketManager(wsURL, nil)

	conn, err := wsManager.Connect()
	assert.NoError(t, err)

	msg := shared.NewMessage(shared.MsgCommand, shared.CmdShellExec, "client-1", nil)
	data, _ := json.Marshal(msg)
	err = conn.WriteMessage(websocket.TextMessage, data)
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	assert.Len(t, messages, 2)
	assert.True(t, strings.Contains(messages[0], "register"))
	assert.True(t, strings.Contains(messages[1], "shell_exec"))
	mu.Unlock()
}
