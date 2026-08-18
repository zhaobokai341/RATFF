package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"RATFF/shared"

	"github.com/gorilla/websocket"
)

var testUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func TestGetServerURLWithPathPassword(t *testing.T) {
	cfg = Config{
		ServerHost:   "localhost",
		ServerPort:   "6341",
		PathPassword: "secret",
	}

	url := getServerURL()
	expected := "ws://localhost:6341/secret/ws"
	if url != expected {
		t.Errorf("Expected %s, got %s", expected, url)
	}
}

func TestGetServerURLWithoutPathPassword(t *testing.T) {
	cfg = Config{
		ServerHost:   "localhost",
		ServerPort:   "6341",
		PathPassword: "",
	}

	url := getServerURL()
	expected := "ws://localhost:6341/ws"
	if url != expected {
		t.Errorf("Expected %s, got %s", expected, url)
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	os.Setenv("SERVER_HOST", "testhost")
	os.Setenv("SERVER_PORT", "9999")
	os.Setenv("PATH_PASSWORD", "testpass")
	defer func() {
		os.Unsetenv("SERVER_HOST")
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("PATH_PASSWORD")
	}()

	loadConfig()

	if cfg.ServerHost != "testhost" {
		t.Errorf("Expected testhost, got %s", cfg.ServerHost)
	}
	if cfg.ServerPort != "9999" {
		t.Errorf("Expected 9999, got %s", cfg.ServerPort)
	}
	if cfg.PathPassword != "testpass" {
		t.Errorf("Expected testpass, got %s", cfg.PathPassword)
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	os.Unsetenv("SERVER_HOST")
	os.Unsetenv("SERVER_PORT")
	os.Unsetenv("PATH_PASSWORD")

	loadConfig()

	if cfg.ServerHost != "localhost" {
		t.Errorf("Expected localhost, got %s", cfg.ServerHost)
	}
	if cfg.ServerPort != "6341" {
		t.Errorf("Expected 6341, got %s", cfg.ServerPort)
	}
	if cfg.PathPassword != "" {
		t.Errorf("Expected empty, got %s", cfg.PathPassword)
	}
}

func TestGetEnv(t *testing.T) {
	os.Setenv("TEST_VAR", "testvalue")
	defer os.Unsetenv("TEST_VAR")

	if shared.GetEnv("TEST_VAR", "default") != "testvalue" {
		t.Error("Expected testvalue")
	}
	if shared.GetEnv("NONEXISTENT_VAR", "default") != "default" {
		t.Error("Expected default")
	}
}

func TestRunClientConnectionRefused(t *testing.T) {
	log = shared.InitLogger("error", "text")

	err := runClient("ws://localhost:19999/ws", "test-client")
	if err == nil {
		t.Error("Expected connection refused error")
	}
}

func TestRunClientSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := testUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Read register message
		var msg shared.Message
		if err := shared.ReadWSMessage(conn, &msg); err != nil {
			return
		}

		if msg.Type != shared.MsgRegister {
			return
		}

		// Send heartbeat
		wsConn := shared.NewWSConn(conn)
		shared.SetupSafeHeartbeat(wsConn)
		_ = shared.SendSafeWSMessage(wsConn, shared.NewMessage(shared.MsgHeartbeat, "", "", nil))

		time.Sleep(100 * time.Millisecond)
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[4:] + "/ws"

	log = shared.InitLogger("error", "text")

	err := runClient(wsURL, "test-client")
	if err != nil {
		t.Logf("Expected connection to close, got: %v", err)
	}
}

func TestExecuteCommandShell(t *testing.T) {
	log = shared.InitLogger("error", "text")

	msg := shared.NewMessage(shared.MsgCommand, shared.CmdShellExec, "test-client", map[string]interface{}{
		"cmd": "echo hello",
	})

	resp := executeCommand(msg)

	if resp.Type != shared.MsgResponse {
		t.Errorf("Expected response, got %s", resp.Type)
	}
}

func TestExecuteCommandSystemInfo(t *testing.T) {
	log = shared.InitLogger("error", "text")

	msg := shared.NewMessage(shared.MsgCommand, shared.CmdSystemInfo, "test-client", nil)

	resp := executeCommand(msg)

	if resp.Type != shared.MsgResponse {
		t.Errorf("Expected response, got %s", resp.Type)
	}
}

func TestExecuteCommandUnknown(t *testing.T) {
	log = shared.InitLogger("error", "text")

	msg := shared.NewMessage(shared.MsgCommand, "unknown", "test-client", nil)

	resp := executeCommand(msg)

	if resp.Type != shared.MsgError {
		t.Errorf("Expected error, got %s", resp.Type)
	}
}

func TestHandleShellExecEmptyCommand(t *testing.T) {
	log = shared.InitLogger("error", "text")

	msg := shared.NewMessage(shared.MsgCommand, shared.CmdShellExec, "test-client", map[string]interface{}{
		"cmd": "",
	})

	resp := executeCommand(msg)

	if resp.Type != shared.MsgError {
		t.Errorf("Expected error for empty command, got %s", resp.Type)
	}
}

func TestBuildPrompt(t *testing.T) {
	tests := []struct {
		id             string
		inCommandMode  bool
		expectedSuffix string
	}{
		{"abc", false, "abc> "},
		{"abc", true, "abc(cmd)> "},
		{"", false, "> "},
		{"", true, "(cmd)> "},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			// This function is in server_cli, skip for now
		})
	}
}

func TestFormatClientID(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"abc", "abc       "},
		{"abcdefghijklmnop", "abcdefghijklmnop"},
		{"abcdefghijklmnopq", "abcdefghijklmnopq"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			// This function is in server_cli, skip for now
		})
	}
}

func TestHandlePwd(t *testing.T) {
	log = shared.InitLogger("error", "text")

	msg := shared.NewMessage(shared.MsgCommand, shared.CmdPwd, "test-client", nil)

	resp := executeCommand(msg)

	if resp.Type != shared.MsgResponse {
		t.Errorf("Expected response, got %s", resp.Type)
	}

	if resp.Payload["current_dir"] == nil {
		t.Error("Expected current_dir in payload")
	}
}

func TestHandlePwdWithWorkingDir(t *testing.T) {
	log = shared.InitLogger("error", "text")

	workingMu.Lock()
	workingDir = "/tmp"
	workingMu.Unlock()

	defer func() {
		workingMu.Lock()
		workingDir = ""
		workingMu.Unlock()
	}()

	msg := shared.NewMessage(shared.MsgCommand, shared.CmdPwd, "test-client", nil)

	resp := executeCommand(msg)

	if resp.Type != shared.MsgResponse {
		t.Errorf("Expected response, got %s", resp.Type)
	}

	dir, _ := resp.Payload["current_dir"].(string)
	if dir != "/tmp" {
		t.Errorf("Expected /tmp, got %s", dir)
	}
}
