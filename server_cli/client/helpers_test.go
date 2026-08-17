package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"RATFF/server_cli/api"
	"RATFF/shared"

	"github.com/stretchr/testify/assert"
)

func TestNewManager(t *testing.T) {
	apiClient := api.NewClient("http://localhost:8080", "test-token")
	tFunc := func(key string) string { return key }
	tfFunc := func(key string, args ...interface{}) string { return key }
	printFuncs := PrintFuncs{
		Success: func(s string) {},
		Error:   func(s string) {},
		Info:    func(s string) {},
		Warn:    func(s string) {},
	}

	mgr := NewManager(apiClient, tFunc, tfFunc, printFuncs)

	assert.NotNil(t, mgr)
	assert.Equal(t, apiClient, mgr.apiClient)
	assert.NotNil(t, mgr.pendingCmd)
	assert.NotNil(t, mgr.T)
	assert.NotNil(t, mgr.Tf)
}

func TestHandleMessageRoutesToPending(t *testing.T) {
	mgr := &Manager{
		pendingCmd: make(map[string]*PendingCommand),
		T:          func(key string) string { return key },
		Tf:         func(key string, args ...interface{}) string { return key },
		Print: PrintFuncs{
			Error: func(s string) {},
		},
	}

	ch := make(chan shared.Message, 1)
	mgr.pendingCmd["client-1"] = &PendingCommand{ch: ch}

	msg := shared.NewMessage(shared.MsgResponse, "", "client-1", map[string]interface{}{"result": "success"})
	mgr.HandleMessage(msg)

	select {
	case received := <-ch:
		assert.Equal(t, "client-1", received.ClientID)
	default:
		t.Fatal("Message was not routed to pending command")
	}

	assert.Empty(t, mgr.pendingCmd)
}

func TestHandleMessageIgnoresUnknownClient(t *testing.T) {
	mgr := &Manager{
		pendingCmd: make(map[string]*PendingCommand),
	}

	msg := shared.NewMessage(shared.MsgResponse, "", "unknown-client", nil)
	mgr.HandleMessage(msg)

	assert.Empty(t, mgr.pendingCmd)
}

func TestHandleMessageDoesNotBlockOnFullChannel(t *testing.T) {
	mgr := &Manager{
		pendingCmd: make(map[string]*PendingCommand),
	}

	ch := make(chan shared.Message, 1)
	ch <- shared.NewMessage(shared.MsgResponse, "", "client-1", nil)
	mgr.pendingCmd["client-1"] = &PendingCommand{ch: ch}

	msg := shared.NewMessage(shared.MsgResponse, "", "client-1", nil)
	mgr.HandleMessage(msg)

	assert.Empty(t, mgr.pendingCmd)
}

func TestWaitForResponseSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
		})
	}))
	defer server.Close()

	apiClient := api.NewClient(server.URL, "test-token")
	mgr := NewManager(apiClient,
		func(key string) string { return key },
		func(key string, args ...interface{}) string { return key },
		PrintFuncs{
			Success: func(s string) {},
			Error:   func(s string) {},
			Info:    func(s string) {},
			Warn:    func(s string) {},
		},
	)

	ch := make(chan shared.Message, 1)
	mgr.pendingCmd["client-1"] = &PendingCommand{ch: ch}

	go func() {
		time.Sleep(50 * time.Millisecond)
		mgr.HandleMessage(shared.NewMessage(shared.MsgResponse, "", "client-1", nil))
	}()

	result := mgr.WaitForResponse("client-1", shared.CmdShellExec, nil, 2*time.Second)
	assert.True(t, result)
}

func TestWaitForResponseWithMsgSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
		})
	}))
	defer server.Close()

	apiClient := api.NewClient(server.URL, "test-token")
	mgr := NewManager(apiClient,
		func(key string) string { return key },
		func(key string, args ...interface{}) string { return key },
		PrintFuncs{
			Success: func(s string) {},
			Error:   func(s string) {},
			Info:    func(s string) {},
			Warn:    func(s string) {},
		},
	)

	ch := make(chan shared.Message, 1)
	mgr.pendingCmd["client-1"] = &PendingCommand{ch: ch}

	go func() {
		time.Sleep(50 * time.Millisecond)
		mgr.HandleMessage(shared.NewMessage(shared.MsgResponse, "", "client-1", map[string]interface{}{"result": "ok"}))
	}()

	msg := mgr.WaitForResponseWithMsg("client-1", shared.CmdShellExec, nil, 2*time.Second)
	assert.NotNil(t, msg)
	assert.Equal(t, "ok", msg.Payload["result"])
}

func TestWaitForResponseWithMsgError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
		})
	}))
	defer server.Close()

	apiClient := api.NewClient(server.URL, "test-token")
	var errorMsg string
	mgr := NewManager(apiClient,
		func(key string) string { return key },
		func(key string, args ...interface{}) string { return key },
		PrintFuncs{
			Success: func(s string) {},
			Error:   func(s string) { errorMsg = s },
			Info:    func(s string) {},
			Warn:    func(s string) {},
		},
	)

	ch := make(chan shared.Message, 1)
	mgr.pendingCmd["client-1"] = &PendingCommand{ch: ch}

	go func() {
		time.Sleep(50 * time.Millisecond)
		mgr.HandleMessage(shared.NewMessage(shared.MsgResponse, "", "client-1", map[string]interface{}{"error": "test error"}))
	}()

	msg := mgr.WaitForResponseWithMsg("client-1", shared.CmdShellExec, nil, 2*time.Second)
	assert.Nil(t, msg)
	assert.Equal(t, "operation_failed", errorMsg)
}

func TestWaitForResponseTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
		})
	}))
	defer server.Close()

	apiClient := api.NewClient(server.URL, "test-token")
	var errorMsg string
	mgr := NewManager(apiClient,
		func(key string) string { return key },
		func(key string, args ...interface{}) string { return key },
		PrintFuncs{
			Success: func(s string) {},
			Error:   func(s string) { errorMsg = s },
			Info:    func(s string) {},
			Warn:    func(s string) {},
		},
	)

	result := mgr.WaitForResponse("client-1", shared.CmdShellExec, nil, 100*time.Millisecond)
	assert.False(t, result)
	assert.Contains(t, errorMsg, "timeout")
}

func TestWaitForResponseRaw(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
		})
	}))
	defer server.Close()

	apiClient := api.NewClient(server.URL, "test-token")
	mgr := NewManager(apiClient,
		func(key string) string { return key },
		func(key string, args ...interface{}) string { return key },
		PrintFuncs{
			Success: func(s string) {},
			Error:   func(s string) {},
			Info:    func(s string) {},
			Warn:    func(s string) {},
		},
	)

	ch := make(chan shared.Message, 1)
	mgr.pendingCmd["client-1"] = &PendingCommand{ch: ch}

	go func() {
		time.Sleep(50 * time.Millisecond)
		mgr.HandleMessage(shared.NewMessage(shared.MsgResponse, "", "client-1", map[string]interface{}{"error": "should not be checked"}))
	}()

	msg := mgr.WaitForResponseRaw("client-1", shared.CmdShellExec, nil, 2*time.Second)
	assert.NotNil(t, msg)
	assert.Equal(t, "should not be checked", msg.Payload["error"])
}
