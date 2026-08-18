package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"RATFF/shared"

	"github.com/gin-gonic/gin"
)

func TestNewClientManager(t *testing.T) {
	m := NewClientManager()
	if m == nil {
		t.Fatal("manager should not be nil")
	}
	if m.clients == nil {
		t.Error("clients map should be initializing")
	}
}

func TestRegisterAndUnregister(t *testing.T) {
	m := NewClientManager()
	info := shared.ClientInfo{ID: "test-1", IP: "127.0.0.1", Hostname: "test", OSInfo: "linux/amd64"}
	m.Register("test-1", nil, info)
	if !m.IsOnline("test-1") {
		t.Error("client should be online after register")
	}
	m.Unregister("test-1")
	if m.IsOnline("test-1") {
		t.Error("client should be offline after unregister")
	}
}

func TestListClients(t *testing.T) {
	m := NewClientManager()
	info1 := shared.ClientInfo{ID: "client-1", IP: "10.0.0.1", Hostname: "host1", OSInfo: "linux/amd64"}
	info2 := shared.ClientInfo{ID: "client-2", IP: "10.0.0.2", Hostname: "host2", OSInfo: "darwin/arm64"}
	m.Register("client-1", nil, info1)
	m.Register("client-2", nil, info2)
	if len(m.ListClients()) != 2 {
		t.Fatalf("expected 2 clients, got %d", len(m.ListClients()))
	}
}

func TestListClientsEmpty(t *testing.T) {
	m := NewClientManager()
	if len(m.ListClients()) != 0 {
		t.Errorf("expected 0 clients, got %d", len(m.ListClients()))
	}
}

func TestSendOfflineClient(t *testing.T) {
	m := NewClientManager()
	if err := m.Send("nonexistent", shared.Message{}); err == nil {
		t.Error("expected error for offline client")
	}
}

func TestHandleListClients(t *testing.T) {
	m := NewClientManager()
	m.Register("test-1", nil, shared.ClientInfo{ID: "test-1", IP: "127.0.0.1", Hostname: "test", OSInfo: "linux/amd64"})
	gin.SetMode(gin.TestMode)
	r := setupRouter(m)
	token, _ := generateJWT()
	req := httptest.NewRequest("GET", "/api/clients", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestHandleSendCommandInvalidJSON(t *testing.T) {
	m := NewClientManager()
	gin.SetMode(gin.TestMode)
	r := setupRouter(m)
	token, _ := generateJWT()
	body := `{invalid}`
	req := httptest.NewRequest("POST", "/api/command", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleSendCommandOfflineClient(t *testing.T) {
	m := NewClientManager()
	gin.SetMode(gin.TestMode)
	r := setupRouter(m)
	token, _ := generateJWT()
	body := `{"client_id":"offline","command":"shell_exec","payload":{"cmd":"ls"}}`
	req := httptest.NewRequest("POST", "/api/command", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(shared.GlobalRateLimitMiddleware())
	r.GET("/test", func(c *gin.Context) { c.String(200, "ok") })
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestSetupRouter(t *testing.T) {
	m := NewClientManager()
	if router := setupRouter(m); router == nil {
		t.Error("router should not be nil")
	}
}
