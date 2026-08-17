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

func TestCdClientSuccess(t *testing.T) {
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

	ch := make(chan shared.Message, 1)
	mgr.pendingCmd["client-1"] = &PendingCommand{ch: ch}

	go func() {
		mgr.HandleMessage(shared.NewMessage(shared.MsgResponse, "", "client-1", map[string]interface{}{
			"current_dir": "/tmp",
		}))
	}()

	mgr.CdClient("client-1", "/tmp")
	assert.Equal(t, "cd_success", successMsg)
}

func TestCdClientError(t *testing.T) {
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
		mgr.HandleMessage(shared.NewMessage(shared.MsgResponse, "", "client-1", map[string]interface{}{
			"error": "directory not found",
		}))
	}()

	mgr.CdClient("client-1", "/nonexistent")
	assert.Equal(t, "operation_failed", errorMsg)
}

func TestListFilesSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
		})
	}))
	defer server.Close()

	apiClient := api.NewClient(server.URL, "test-token")
	var printedPath string
	printFileTable := func(path string, files []interface{}, t func(string) string, tf func(string, ...interface{}) string) {
		printedPath = path
	}
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
			"path": "/home/user",
			"files": []interface{}{
				map[string]interface{}{"name": "file1.txt", "type": "file", "size": 100.0},
			},
		}))
	}()

	mgr.ListFiles("client-1", "/home/user", printFileTable)
	assert.Equal(t, "/home/user", printedPath)
}

func TestListFilesError(t *testing.T) {
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
		mgr.HandleMessage(shared.NewMessage(shared.MsgResponse, "", "client-1", map[string]interface{}{
			"error": "permission denied",
		}))
	}()

	mgr.ListFiles("client-1", "/root", func(path string, files []interface{}, t func(string) string, tf func(string, ...interface{}) string) {
	})
	assert.Equal(t, "operation_failed", errorMsg)
}

func TestMoveFileSuccess(t *testing.T) {
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

	ch := make(chan shared.Message, 1)
	mgr.pendingCmd["client-1"] = &PendingCommand{ch: ch}

	go func() {
		mgr.HandleMessage(shared.NewMessage(shared.MsgResponse, "", "client-1", map[string]interface{}{
			"moved": true,
		}))
	}()

	mgr.MoveFile("client-1", "/tmp/file1.txt", "/tmp/file2.txt")
	assert.Equal(t, "file_move_success", successMsg)
}

func TestMoveFileError(t *testing.T) {
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
		mgr.HandleMessage(shared.NewMessage(shared.MsgResponse, "", "client-1", map[string]interface{}{
			"error": "file not found",
		}))
	}()

	mgr.MoveFile("client-1", "/tmp/nonexistent.txt", "/tmp/new.txt")
	assert.Equal(t, "operation_failed", errorMsg)
}

func TestDeleteFileSuccess(t *testing.T) {
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

	ch := make(chan shared.Message, 1)
	mgr.pendingCmd["client-1"] = &PendingCommand{ch: ch}

	go func() {
		mgr.HandleMessage(shared.NewMessage(shared.MsgResponse, "", "client-1", map[string]interface{}{
			"deleted": true,
		}))
	}()

	mgr.DeleteFile("client-1", "/tmp/file.txt")
	assert.Equal(t, "file_delete_success", successMsg)
}

func TestDeleteFileError(t *testing.T) {
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
		mgr.HandleMessage(shared.NewMessage(shared.MsgResponse, "", "client-1", map[string]interface{}{
			"error": "permission denied",
		}))
	}()

	mgr.DeleteFile("client-1", "/root/secret.txt")
	assert.Equal(t, "operation_failed", errorMsg)
}

func TestCopyRemoteFileSuccess(t *testing.T) {
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

	ch := make(chan shared.Message, 1)
	mgr.pendingCmd["client-1"] = &PendingCommand{ch: ch}

	go func() {
		mgr.HandleMessage(shared.NewMessage(shared.MsgResponse, "", "client-1", map[string]interface{}{
			"copied": true,
		}))
	}()

	mgr.CopyRemoteFile("client-1", "/tmp/file1.txt", "/tmp/file2.txt")
	assert.Equal(t, "file_copy_success", successMsg)
}

func TestCopyRemoteFileError(t *testing.T) {
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
		mgr.HandleMessage(shared.NewMessage(shared.MsgResponse, "", "client-1", map[string]interface{}{
			"error": "source file not found",
		}))
	}()

	mgr.CopyRemoteFile("client-1", "/tmp/nonexistent.txt", "/tmp/copy.txt")
	assert.Equal(t, "operation_failed", errorMsg)
}

func TestPwdClientSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
		})
	}))
	defer server.Close()

	apiClient := api.NewClient(server.URL, "test-token")
	var infoMsg string
	mgr := NewManager(apiClient,
		func(key string) string { return key },
		func(key string, args ...interface{}) string { return key },
		PrintFuncs{
			Success: func(s string) {},
			Error:   func(s string) {},
			Info:    func(s string) { infoMsg = s },
			Warn:    func(s string) {},
		},
	)

	ch := make(chan shared.Message, 1)
	mgr.pendingCmd["client-1"] = &PendingCommand{ch: ch}

	go func() {
		mgr.HandleMessage(shared.NewMessage(shared.MsgResponse, "", "client-1", map[string]interface{}{
			"current_dir": "/home/user",
		}))
	}()

	mgr.PwdClient("client-1")
	assert.Equal(t, "/home/user", infoMsg)
}

func TestPwdClientError(t *testing.T) {
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
		mgr.HandleMessage(shared.NewMessage(shared.MsgResponse, "", "client-1", map[string]interface{}{
			"error": "client not ready",
		}))
	}()

	mgr.PwdClient("client-1")
	assert.Equal(t, "operation_failed", errorMsg)
}
