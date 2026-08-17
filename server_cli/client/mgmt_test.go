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

func TestListClientsSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"clients": []map[string]interface{}{
				{"id": "client-1", "ip": "127.0.0.1", "hostname": "host1", "os_info": "linux"},
				{"id": "client-2", "ip": "127.0.0.2", "hostname": "host2", "os_info": "windows"},
			},
		})
	}))
	defer server.Close()

	apiClient := api.NewClient(server.URL, "test-token")
	var printedClients []shared.ClientInfo
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

	mgr.ListClients(func(clients []shared.ClientInfo, t func(string) string, tf func(string, ...interface{}) string) {
		printedClients = clients
	})

	assert.Len(t, printedClients, 2)
}

func TestListClientsFetchFailed(t *testing.T) {
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

	mgr.ListClients(func(clients []shared.ClientInfo, t func(string) string, tf func(string, ...interface{}) string) {})
	assert.Contains(t, errorMsg, "failed")
}

func TestSelectClientFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"clients": []map[string]interface{}{
				{"id": "client-1", "ip": "127.0.0.1", "hostname": "host1", "os_info": "linux"},
				{"id": "client-2", "ip": "127.0.0.2", "hostname": "host2", "os_info": "windows"},
			},
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

	result := mgr.SelectClient("client-1")
	assert.True(t, result)
}

func TestSelectClientNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"clients": []map[string]interface{}{
				{"id": "client-1", "ip": "127.0.0.1", "hostname": "host1", "os_info": "linux"},
			},
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

	result := mgr.SelectClient("nonexistent")
	assert.False(t, result)
}

func TestSelectClientFetchFailed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
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

	result := mgr.SelectClient("client-1")
	assert.False(t, result)
}

func TestDeleteClientSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/clients" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"clients": []map[string]interface{}{
					{"id": "client-1", "ip": "127.0.0.1", "hostname": "host1", "os_info": "linux"},
				},
			})
		} else {
			w.WriteHeader(http.StatusOK)
		}
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

	mgr.DeleteClient("client-1")
	assert.Contains(t, successMsg, "delete")
}

func TestDeleteClientNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"clients": []map[string]interface{}{
				{"id": "client-1", "ip": "127.0.0.1", "hostname": "host1", "os_info": "linux"},
			},
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

	mgr.DeleteClient("nonexistent")
	assert.Contains(t, errorMsg, "not")
}

func TestDeleteClientFetchFailed(t *testing.T) {
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

	mgr.DeleteClient("client-1")
	assert.Contains(t, errorMsg, "failed")
}
