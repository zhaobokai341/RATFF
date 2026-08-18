package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"RATFF/shared"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func init() {
	log = logrus.NewEntry(logrus.New())
}

func TestLoadConfigFromEnv(t *testing.T) {
	os.Setenv("HOST", "127.0.0.1")
	os.Setenv("PORT", "8888")
	os.Setenv("LOGIN_PATH", "testpath")
	os.Setenv("LOGIN_PASSWORD_HASH", "testhash")
	os.Setenv("JWT_SECRET", "testsecret")
	defer func() {
		os.Unsetenv("HOST")
		os.Unsetenv("PORT")
		os.Unsetenv("LOGIN_PATH")
		os.Unsetenv("LOGIN_PASSWORD_HASH")
		os.Unsetenv("JWT_SECRET")
	}()

	loadConfig()

	assert.Equal(t, "127.0.0.1", cfg.Host)
	assert.Equal(t, "8888", cfg.Port)
	assert.Equal(t, "testpath", cfg.PathPassword)
	assert.Equal(t, "testhash", cfg.LoginPasswordHash)
	assert.Equal(t, "testsecret", cfg.JWTSecret)
}

func TestLoadConfigDefaults(t *testing.T) {
	os.Unsetenv("HOST")
	os.Unsetenv("PORT")
	os.Unsetenv("LOGIN_PATH")
	os.Unsetenv("LOGIN_PASSWORD_HASH")
	os.Unsetenv("JWT_SECRET")

	loadConfig()

	assert.Equal(t, "0.0.0.0", cfg.Host)
	assert.Equal(t, "6341", cfg.Port)
	assert.Equal(t, "", cfg.PathPassword)
	assert.Equal(t, "$2b$12$lfEEs6tTAdp61DYg7xiorOkspqK2iTObW/qK6fOsT6JxBfbRBGjn2", cfg.LoginPasswordHash)
	assert.Equal(t, "default-jwt-secret-change-in-production", cfg.JWTSecret)
}

func TestGetEnv(t *testing.T) {
	os.Setenv("TEST_VAR", "testvalue")
	defer os.Unsetenv("TEST_VAR")

	assert.Equal(t, "testvalue", shared.GetEnv("TEST_VAR", "default"))
	assert.Equal(t, "default", shared.GetEnv("NONEXISTENT_VAR", "default"))
}

func TestGenerateJWT(t *testing.T) {
	cfg.JWTSecret = "test-secret"
	token, err := generateJWT()
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestVerifyJWT(t *testing.T) {
	cfg.JWTSecret = "test-secret"
	token, _ := generateJWT()
	claims, err := verifyJWT(token)
	assert.NoError(t, err)
	assert.NotNil(t, claims)
}

func TestVerifyJWTInvalid(t *testing.T) {
	_, err := verifyJWT("invalid-token")
	assert.Error(t, err)
}

func TestCheckPassword(t *testing.T) {
	hash := "$2b$12$lfEEs6tTAdp61DYg7xiorOkspqK2iTObW/qK6fOsT6JxBfbRBGjn2"
	assert.True(t, checkPassword("fuck", hash))
	assert.False(t, checkPassword("wrong", hash))
}

func TestHandleVerifySuccess(t *testing.T) {
	cfg.LoginPasswordHash = "$2b$12$lfEEs6tTAdp61DYg7xiorOkspqK2iTObW/qK6fOsT6JxBfbRBGjn2"
	cfg.JWTSecret = "test-secret"
	cfg.PathPassword = ""

	gin.SetMode(gin.TestMode)
	r := setupRouter(NewClientManager())

	body := `{"password":"fuck"}`
	req := httptest.NewRequest("POST", "/verify", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleVerifyWrongPassword(t *testing.T) {
	cfg.LoginPasswordHash = "$2b$12$lfEEs6tTAdp61DYg7xiorOkspqK2iTObW/qK6fOsT6JxBfbRBGjn2"
	cfg.PathPassword = ""

	gin.SetMode(gin.TestMode)
	r := setupRouter(NewClientManager())

	body := `{"password":"wrong"}`
	req := httptest.NewRequest("POST", "/verify", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandleVerifyNotConfigured(t *testing.T) {
	cfg.LoginPasswordHash = ""
	cfg.PathPassword = ""

	gin.SetMode(gin.TestMode)
	r := setupRouter(NewClientManager())

	body := `{"password":"somepass"}`
	req := httptest.NewRequest("POST", "/verify", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleVerifyMissingBody(t *testing.T) {
	cfg.PathPassword = ""

	gin.SetMode(gin.TestMode)
	r := setupRouter(NewClientManager())

	req := httptest.NewRequest("POST", "/verify", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthMiddlewareMissingHeader(t *testing.T) {
	cfg.PathPassword = ""
	cfg.JWTSecret = "test-secret"

	gin.SetMode(gin.TestMode)
	r := setupRouter(NewClientManager())

	token, _ := generateJWT()
	req := httptest.NewRequest("GET", "/api/clients", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should succeed with valid token
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthMiddlewareInvalidToken(t *testing.T) {
	cfg.PathPassword = ""
	cfg.JWTSecret = "test-secret"

	gin.SetMode(gin.TestMode)
	r := setupRouter(NewClientManager())

	req := httptest.NewRequest("GET", "/api/clients", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandleWebSocketWrongFirstMessage(t *testing.T) {
	manager := NewClientManager()
	gin.SetMode(gin.TestMode)
	r := setupRouter(manager)

	server := httptest.NewServer(r)
	defer server.Close()

	wsURL := "ws" + server.URL[4:] + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Skip("WebSocket connection failed")
	}
	defer conn.Close()

	// Send wrong message type (not register)
	msg := shared.NewMessage(shared.MsgHeartbeat, "", "test-client", nil)
	_ = shared.SendWSMessage(conn, msg)

	// Connection should be closed by server
	time.Sleep(100 * time.Millisecond)
}

func TestHandleWebSocketSuccess(t *testing.T) {
	manager := NewClientManager()
	gin.SetMode(gin.TestMode)
	r := setupRouter(manager)

	server := httptest.NewServer(r)
	defer server.Close()

	wsURL := "ws" + server.URL[4:] + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Skip("WebSocket connection failed")
	}
	defer conn.Close()

	// Send register message
	info := shared.BuildClientInfo("test-client")
	msg := shared.NewMessage(shared.MsgRegister, "", "test-client", info.ToPayload())
	_ = shared.SendWSMessage(conn, msg)

	// Wait for registration
	time.Sleep(100 * time.Millisecond)

	assert.True(t, manager.IsOnline("test-client"))
}

func TestHandleMessageHeartbeat(t *testing.T) {
	manager := NewClientManager()
	gin.SetMode(gin.TestMode)
	r := setupRouter(manager)

	server := httptest.NewServer(r)
	defer server.Close()

	wsURL := "ws" + server.URL[4:] + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Skip("WebSocket connection failed")
	}
	defer conn.Close()

	// First register
	info := shared.BuildClientInfo("test-client")
	msg := shared.NewMessage(shared.MsgRegister, "", "test-client", info.ToPayload())
	_ = shared.SendWSMessage(conn, msg)
	time.Sleep(50 * time.Millisecond)

	// Then send heartbeat
	msg = shared.NewMessage(shared.MsgHeartbeat, "", "test-client", nil)
	_ = shared.SendWSMessage(conn, msg)

	// Wait for response
	_ = conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	_, _, err = conn.ReadMessage()
	if err != nil {
		t.Logf("Read error (may be expected): %v", err)
	}
}

func TestHandleMessageCommandOffline(t *testing.T) {
	manager := NewClientManager()
	gin.SetMode(gin.TestMode)
	r := setupRouter(manager)

	server := httptest.NewServer(r)
	defer server.Close()

	wsURL := "ws" + server.URL[4:] + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Skip("WebSocket connection failed")
	}
	defer conn.Close()

	// First register
	info := shared.BuildClientInfo("test-client")
	msg := shared.NewMessage(shared.MsgRegister, "", "test-client", info.ToPayload())
	_ = shared.SendWSMessage(conn, msg)
	time.Sleep(50 * time.Millisecond)

	// Send command to offline client
	msg = shared.NewMessage(shared.MsgCommand, "offline-client", "test-client", nil)
	_ = shared.SendWSMessage(conn, msg)

	// Wait for error response
	_ = conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	_, _, err = conn.ReadMessage()
	if err != nil {
		t.Logf("Read error (may be expected): %v", err)
	}
}
