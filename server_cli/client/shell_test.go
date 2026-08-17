package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"RATFF/server_cli/api"
	"RATFF/shared"

	"github.com/stretchr/testify/assert"
)

func TestShellCommandSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
		})
	}))
	defer server.Close()

	apiClient := api.NewClient(server.URL, "test-token")
	var stdout, stderr string
	var exitCode int
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
		mgr.HandleMessage(shared.NewMessage(shared.MsgResponse, "", "client-1", map[string]interface{}{
			"stdout":    "hello world",
			"stderr":    "",
			"exit_code": 0.0,
		}))
	}()

	mgr.ShellCommand("client-1", "echo hello", func(s, e string, c int) {
		stdout = s
		stderr = e
		exitCode = c
	})

	assert.Equal(t, "hello world", stdout)
	assert.Empty(t, stderr)
	assert.Equal(t, 0, exitCode)
}

func TestShellCommandWithError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
		})
	}))
	defer server.Close()

	apiClient := api.NewClient(server.URL, "test-token")
	var stdout, stderr string
	var exitCode int
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
		mgr.HandleMessage(shared.NewMessage(shared.MsgResponse, "", "client-1", map[string]interface{}{
			"stdout":    "",
			"stderr":    "command not found",
			"exit_code": 127.0,
		}))
	}()

	mgr.ShellCommand("client-1", "invalidcmd", func(s, e string, c int) {
		stdout = s
		stderr = e
		exitCode = c
	})

	assert.Empty(t, stdout)
	assert.Equal(t, "command not found", stderr)
	assert.Equal(t, 127, exitCode)
}

func TestShellCommandTimeout(t *testing.T) {
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

	mgr.ShellCommand("client-1", "sleep 100", func(s, e string, c int) {})
	assert.Contains(t, errorMsg, "timeout")
}

func TestShellCommandExitCodeInt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
		})
	}))
	defer server.Close()

	apiClient := api.NewClient(server.URL, "test-token")
	var exitCode int
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
		mgr.HandleMessage(shared.NewMessage(shared.MsgResponse, "", "client-1", map[string]interface{}{
			"stdout":    "",
			"stderr":    "",
			"exit_code": 1,
		}))
	}()

	mgr.ShellCommand("client-1", "false", func(s, e string, c int) {
		exitCode = c
	})

	assert.Equal(t, 1, exitCode)
}

func TestBgCommandWithoutOutputFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
		})
	}))
	defer server.Close()

	apiClient := api.NewClient(server.URL, "test-token")
	var successMsg string
	mgr := NewManager(apiClient,
		func(key string) string { return key },
		func(key string, args ...interface{}) string { return key },
		PrintFuncs{
			Success: func(s string) { successMsg = s },
			Error:   func(s string) {},
			Info:    func(s string) {},
			Warn:    func(s string) {},
		},
	)

	mgr.BgCommand("client-1", "sleep 10", "")
	assert.Equal(t, "bg_command_started", successMsg)
}

func TestBgCommandWithOutputFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
		})
	}))
	defer server.Close()

	apiClient := api.NewClient(server.URL, "test-token")
	var successMsg string
	mgr := NewManager(apiClient,
		func(key string) string { return key },
		func(key string, args ...interface{}) string { return key },
		PrintFuncs{
			Success: func(s string) { successMsg = s },
			Error:   func(s string) {},
			Info:    func(s string) {},
			Warn:    func(s string) {},
		},
	)

	mgr.BgCommand("client-1", "sleep 10", "/tmp/output.txt")
	assert.Equal(t, "bg_command_started_with_output", successMsg)
}

func TestBgCommandFailed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
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

	mgr.BgCommand("client-1", "sleep 10", "")
	assert.Contains(t, errorMsg, "failed")
}
