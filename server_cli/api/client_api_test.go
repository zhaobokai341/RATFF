package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClientGetWithToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, "/api/clients", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"clients": []map[string]interface{}{
				{"id": "client-1", "ip": "127.0.0.1"},
			},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")

	resp, err := client.Get("/clients")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestClientGetWithoutToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Empty(t, r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, "")

	resp, err := client.Get("/clients")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	defer resp.Body.Close()
}

func TestClientPostWithToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, "POST", r.Method)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")

	err := client.Post("/command", nil)
	assert.NoError(t, err)
}

func TestClientPostError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")

	err := client.Post("/command", nil)
	assert.Error(t, err)
}

func TestClientFetchClientsSuccess(t *testing.T) {
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

	client := NewClient(server.URL, "test-token")

	clients, err := client.FetchClients()
	assert.NoError(t, err)
	assert.Len(t, clients, 2)
	assert.Equal(t, "client-1", clients[0].ID)
	assert.Equal(t, "client-2", clients[1].ID)
}

func TestClientFetchClientsEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"clients": []map[string]interface{}{},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")

	clients, err := client.FetchClients()
	assert.NoError(t, err)
	assert.Empty(t, clients)
}

func TestClientFetchClientsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")

	clients, err := client.FetchClients()
	assert.Error(t, err)
	assert.Nil(t, clients)
}

func TestClientFetchClientsInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{invalid json`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")

	clients, err := client.FetchClients()
	assert.Error(t, err)
	assert.Nil(t, clients)
}

func TestClientPostCommandSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "sent",
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")

	err := client.PostCommand(map[string]interface{}{
		"client_id": "client-1",
		"command":   "shell_exec",
		"payload":   map[string]interface{}{"cmd": "ls"},
	})
	assert.NoError(t, err)
}

func TestClientPostCommandError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")

	err := client.PostCommand(nil)
	assert.Error(t, err)
}
