package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"RATFF/shared"
)

// setupTestServer creates a mock HTTP server for testing.
func setupTestServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

// configureTestServer updates cfg to point to the test server.
func configureTestServer(server *httptest.Server) func() {
	originalHost := cfg.Host
	originalPort := cfg.Port
	originalPathPwd := cfg.PathPassword

	// Extract host and port from test server URL
	// server.URL format: http://127.0.0.1:12345
	cfg.Host = server.URL[7:] // Remove "http://"
	cfg.Port = ""             // Clear port since Host already contains port
	cfg.PathPassword = ""

	return func() {
		cfg.Host = originalHost
		cfg.Port = originalPort
		cfg.PathPassword = originalPathPwd
	}
}

// TestFetchClientsSuccess tests successful client list retrieval.
func TestFetchClientsSuccess(t *testing.T) {
	server := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/clients" {
			http.NotFound(w, r)
			return
		}

		resp := map[string]interface{}{
			"clients": []shared.ClientInfo{
				{ID: "test-001", IP: "192.168.1.1", Hostname: "host1", OSInfo: "Linux"},
				{ID: "test-002", IP: "192.168.1.2", Hostname: "host2", OSInfo: "Windows"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, "Encode error", http.StatusInternalServerError)
			return
		}
	})
	defer server.Close()

	cleanup := configureTestServer(server)
	defer cleanup()

	clients, err := fetchClients()
	if err != nil {
		t.Fatalf("fetchClients() error = %v", err)
	}

	if len(clients) != 2 {
		t.Errorf("expected 2 clients, got %d", len(clients))
	}

	if clients[0].ID != "test-001" {
		t.Errorf("expected client ID test-001, got %s", clients[0].ID)
	}
}

// TestFetchClientsEmpty tests fetching empty client list.
func TestFetchClientsEmpty(t *testing.T) {
	server := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"clients": []shared.ClientInfo{},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, "Encode error", http.StatusInternalServerError)
			return
		}
	})
	defer server.Close()

	cleanup := configureTestServer(server)
	defer cleanup()

	clients, err := fetchClients()
	if err != nil {
		t.Fatalf("fetchClients() error = %v", err)
	}

	if len(clients) != 0 {
		t.Errorf("expected 0 clients, got %d", len(clients))
	}
}

// TestFetchClientsServerError tests server error handling.
func TestFetchClientsServerError(t *testing.T) {
	server := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	})
	defer server.Close()

	cleanup := configureTestServer(server)
	defer cleanup()

	_, err := fetchClients()
	if err == nil {
		t.Fatal("expected error for server error, got nil")
	}
}

// TestFetchClientsInvalidJSON tests invalid JSON response handling.
func TestFetchClientsInvalidJSON(t *testing.T) {
	server := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("invalid json")); err != nil {
			t.Logf("write error: %v", err)
		}
	})
	defer server.Close()

	cleanup := configureTestServer(server)
	defer cleanup()

	_, err := fetchClients()
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

// TestPostCommandSuccess tests successful command posting.
func TestPostCommandSuccess(t *testing.T) {
	server := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if r.URL.Path != "/api/command" {
			http.NotFound(w, r)
			return
		}

		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	cleanup := configureTestServer(server)
	defer cleanup()

	payload := map[string]interface{}{
		"client_id": "test-001",
		"command":   "shell_exec",
	}

	err := postCommand(payload)
	if err != nil {
		t.Fatalf("postCommand() error = %v", err)
	}
}

// TestPostCommandServerError tests command posting with server error.
func TestPostCommandServerError(t *testing.T) {
	server := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	})
	defer server.Close()

	cleanup := configureTestServer(server)
	defer cleanup()

	payload := map[string]interface{}{
		"client_id": "test-001",
		"command":   "shell_exec",
	}

	err := postCommand(payload)
	if err == nil {
		t.Fatal("expected error for server error, got nil")
	}
}

// TestLoginToAPISuccess tests successful login.
func TestLoginToAPISuccess(t *testing.T) {
	server := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/verify" {
			http.NotFound(w, r)
			return
		}

		resp := map[string]string{"token": "test-jwt-token"}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, "Encode error", http.StatusInternalServerError)
			return
		}
	})
	defer server.Close()

	cleanup := configureTestServer(server)
	defer cleanup()

	token, err := loginToAPI("test-password")
	if err != nil {
		t.Fatalf("loginToAPI() error = %v", err)
	}

	if token != "test-jwt-token" {
		t.Errorf("expected token test-jwt-token, got %s", token)
	}
}

// TestLoginToAPIPathNotFound tests login with wrong path password.
func TestLoginToAPIPathNotFound(t *testing.T) {
	server := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	defer server.Close()

	cleanup := configureTestServer(server)
	defer cleanup()

	_, err := loginToAPI("wrong-password")
	if err == nil {
		t.Fatal("expected error for wrong path password, got nil")
	}
}

// TestLoginToAPIUnauthorized tests login with wrong login password.
func TestLoginToAPIUnauthorized(t *testing.T) {
	server := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})
	defer server.Close()

	cleanup := configureTestServer(server)
	defer cleanup()

	_, err := loginToAPI("wrong-password")
	if err == nil {
		t.Fatal("expected error for wrong login password, got nil")
	}
}

// TestLoginToAPIInvalidResponse tests login with invalid JSON response.
func TestLoginToAPIInvalidResponse(t *testing.T) {
	server := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("invalid json")); err != nil {
			t.Logf("write error: %v", err)
		}
	})
	defer server.Close()

	cleanup := configureTestServer(server)
	defer cleanup()

	_, err := loginToAPI("test-password")
	if err == nil {
		t.Fatal("expected error for invalid response, got nil")
	}
}

// TestAPIGetWithAuth tests GET request with JWT token.
func TestAPIGetWithAuth(t *testing.T) {
	server := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	cleanup := configureTestServer(server)
	defer cleanup()

	originalToken := jwtToken
	jwtToken = "test-token"
	defer func() { jwtToken = originalToken }()

	resp, err := apiGet("/test")
	if err != nil {
		t.Fatalf("apiGet() error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

// TestAPIPostWithAuth tests POST request with JWT token.
func TestAPIPostWithAuth(t *testing.T) {
	server := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Invalid content type", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	cleanup := configureTestServer(server)
	defer cleanup()

	originalToken := jwtToken
	jwtToken = "test-token"
	defer func() { jwtToken = originalToken }()

	err := apiPost("/test", map[string]string{"key": "value"})
	if err != nil {
		t.Fatalf("apiPost() error = %v", err)
	}
}

// TestAPIPostServerError tests POST request with server error.
func TestAPIPostServerError(t *testing.T) {
	server := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	})
	defer server.Close()

	cleanup := configureTestServer(server)
	defer cleanup()

	err := apiPost("/test", map[string]string{"key": "value"})
	if err == nil {
		t.Fatal("expected error for server error, got nil")
	}
}

// TestSelectClientSuccess tests successful client selection.
func TestSelectClientSuccess(t *testing.T) {
	server := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"clients": []shared.ClientInfo{
				{ID: "test-001", IP: "192.168.1.1", Hostname: "host1", OSInfo: "Linux"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, "Encode error", http.StatusInternalServerError)
			return
		}
	})
	defer server.Close()

	cleanup := configureTestServer(server)
	defer cleanup()

	result := selectClient("test-001")
	if !result {
		t.Error("expected selectClient to return true")
	}
}

// TestSelectClientNotFound tests selecting non-existent client.
func TestSelectClientNotFound(t *testing.T) {
	server := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"clients": []shared.ClientInfo{
				{ID: "test-001", IP: "192.168.1.1", Hostname: "host1", OSInfo: "Linux"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, "Encode error", http.StatusInternalServerError)
			return
		}
	})
	defer server.Close()

	cleanup := configureTestServer(server)
	defer cleanup()

	result := selectClient("non-existent")
	if result {
		t.Error("expected selectClient to return false for non-existent client")
	}
}

// TestSelectClientFetchError tests client selection with fetch error.
func TestSelectClientFetchError(t *testing.T) {
	server := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	})
	defer server.Close()

	cleanup := configureTestServer(server)
	defer cleanup()

	result := selectClient("test-001")
	if result {
		t.Error("expected selectClient to return false on fetch error")
	}
}

// TestDeleteClientSuccess tests successful client deletion.
func TestDeleteClientSuccess(t *testing.T) {
	callCount := 0
	server := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		switch r.URL.Path {
		case "/api/clients":
			resp := map[string]interface{}{
				"clients": []shared.ClientInfo{
					{ID: "test-001", IP: "192.168.1.1", Hostname: "host1", OSInfo: "Linux"},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(resp); err != nil {
				http.Error(w, "Encode error", http.StatusInternalServerError)
				return
			}
		case "/api/command":
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	})
	defer server.Close()

	cleanup := configureTestServer(server)
	defer cleanup()

	deleteClient("test-001")

	if callCount != 2 {
		t.Errorf("expected 2 API calls, got %d", callCount)
	}
}

// TestDeleteClientNotFound tests deleting non-existent client.
func TestDeleteClientNotFound(t *testing.T) {
	callCount := 0
	server := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		resp := map[string]interface{}{
			"clients": []shared.ClientInfo{
				{ID: "test-001", IP: "192.168.1.1", Hostname: "host1", OSInfo: "Linux"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, "Encode error", http.StatusInternalServerError)
			return
		}
	})
	defer server.Close()

	cleanup := configureTestServer(server)
	defer cleanup()

	deleteClient("non-existent")

	if callCount != 1 {
		t.Errorf("expected 1 API call (only fetch), got %d", callCount)
	}
}

// TestDeleteClientFetchError tests deletion with fetch error.
func TestDeleteClientFetchError(t *testing.T) {
	callCount := 0
	server := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	})
	defer server.Close()

	cleanup := configureTestServer(server)
	defer cleanup()

	deleteClient("test-001")

	if callCount != 1 {
		t.Errorf("expected 1 API call, got %d", callCount)
	}
}

// TestListClientsSuccess tests successful client list display.
func TestListClientsSuccess(t *testing.T) {
	server := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"clients": []shared.ClientInfo{
				{ID: "test-001", IP: "192.168.1.1", Hostname: "host1", OSInfo: "Linux"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, "Encode error", http.StatusInternalServerError)
			return
		}
	})
	defer server.Close()

	cleanup := configureTestServer(server)
	defer cleanup()

	listClients()
}

// TestListClientsError tests client list display with error.
func TestListClientsError(t *testing.T) {
	server := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	})
	defer server.Close()

	cleanup := configureTestServer(server)
	defer cleanup()

	listClients()
}
