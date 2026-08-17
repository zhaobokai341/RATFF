package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"RATFF/server_cli/api"
	"RATFF/shared"

	"github.com/stretchr/testify/assert"
)

type mockProgressBar struct {
	addCalls   []int64
	displayed  bool
	markedDone bool
}

func (m *mockProgressBar) Add(n int64) {
	m.addCalls = append(m.addCalls, n)
}

func (m *mockProgressBar) Display() {
	m.displayed = true
}

func (m *mockProgressBar) MarkDone() {
	m.markedDone = true
}

func TestUploadFileLocalPathNotExist(t *testing.T) {
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

	mgr.UploadFile("client-1", "/nonexistent/path", "/remote/path", nil)
	assert.Contains(t, errorMsg, "file_not_exist")
}

func TestUploadSingleFileOpenFailed(t *testing.T) {
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

	tempDir := t.TempDir()
	nonExistentFile := filepath.Join(tempDir, "nonexistent.txt")

	mgr.uploadSingleFile("client-1", nonExistentFile, "/remote/path", nil)
	assert.Contains(t, errorMsg, "file_open_failed")
}

func TestIsRemoteDirectorySuccess(t *testing.T) {
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
		mgr.HandleMessage(shared.NewMessage(shared.MsgResponse, "", "client-1", map[string]interface{}{
			"files": []interface{}{},
		}))
	}()

	result := mgr.isRemoteDirectory("client-1", "/remote/path")
	assert.True(t, result)
}

func TestIsRemoteDirectoryWithError(t *testing.T) {
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
		mgr.HandleMessage(shared.NewMessage(shared.MsgResponse, "", "client-1", map[string]interface{}{
			"error": "not found",
		}))
	}()

	result := mgr.isRemoteDirectory("client-1", "/remote/path")
	assert.False(t, result)
}

func TestIsRemoteDirectoryWithNilPayload(t *testing.T) {
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
		mgr.HandleMessage(shared.NewMessage(shared.MsgResponse, "", "client-1", nil))
	}()

	result := mgr.isRemoteDirectory("client-1", "/remote/path")
	assert.False(t, result)
}

func TestDownloadSingleFileCreateDirFailed(t *testing.T) {
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

	// Create a file to use as a directory path (will fail to create dir)
	tempDir := t.TempDir()
	existingFile := filepath.Join(tempDir, "existing.txt")
	os.WriteFile(existingFile, []byte("test"), 0644)

	// Try to create a directory with the same name as a file
	invalidDir := filepath.Join(existingFile, "subdir", "file.txt")
	mgr.downloadSingleFile("client-1", "/remote/file.txt", invalidDir, nil)
	assert.Contains(t, errorMsg, "create_dir_failed")
}

func TestDownloadSingleFileStartFailed(t *testing.T) {
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
		mgr.HandleMessage(shared.NewMessage(shared.MsgResponse, "", "client-1", map[string]interface{}{
			"error": "file not found",
		}))
	}()

	tempDir := t.TempDir()
	localPath := filepath.Join(tempDir, "downloaded.txt")

	mgr.downloadSingleFile("client-1", "/remote/file.txt", localPath, nil)
}

func TestDownloadSingleFileWithNilPayload(t *testing.T) {
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

	tempDir := t.TempDir()
	localPath := filepath.Join(tempDir, "downloaded.txt")

	mgr.downloadSingleFile("client-1", "/remote/file.txt", localPath, nil)
}
