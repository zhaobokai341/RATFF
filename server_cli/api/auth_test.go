package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoginToAPISuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/verify", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var req struct {
			Password string `json:"password"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "correct", req.Password)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"token": "test-token-123",
		})
	}))
	defer server.Close()

	tFunc := func(key string) string { return key }

	token, err := LoginToAPI(server.URL, "correct", tFunc)
	assert.NoError(t, err)
	assert.Equal(t, "test-token-123", token)
}

func TestLoginToAPIPathError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	tFunc := func(key string) string { return key }

	token, err := LoginToAPI(server.URL, "password", tFunc)
	assert.Error(t, err)
	assert.Empty(t, token)
}

func TestLoginToAPIWrongPassword(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "invalid password",
		})
	}))
	defer server.Close()

	tFunc := func(key string) string { return key }

	token, err := LoginToAPI(server.URL, "wrong", tFunc)
	assert.Error(t, err)
	assert.Empty(t, token)
}

func TestLoginToAPIInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{invalid json`))
	}))
	defer server.Close()

	tFunc := func(key string) string { return key }

	token, err := LoginToAPI(server.URL, "password", tFunc)
	assert.Error(t, err)
	assert.Empty(t, token)
}

func TestLoginToAPINoToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{})
	}))
	defer server.Close()

	tFunc := func(key string) string { return key }

	token, err := LoginToAPI(server.URL, "password", tFunc)
	assert.NoError(t, err)
	assert.Empty(t, token)
}

func TestLoginToAPIConnectionFailed(t *testing.T) {
	tFunc := func(key string) string { return key }

	token, err := LoginToAPI("http://localhost:99999", "password", tFunc)
	assert.Error(t, err)
	assert.Empty(t, token)
}
