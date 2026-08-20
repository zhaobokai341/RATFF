package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"RATFF/shared"

	"github.com/gin-gonic/gin"
)

func init() {
	log = shared.InitLogger("info", "text")
	cfg = Config{
		APIBaseURL: "http://localhost:9090",
	}
}

func TestRootRedirectsToLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := setupRouter()

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("expected 302 redirect, got %d", w.Code)
	}
}

func TestLoginRedirectsToLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := setupRouter()

	req := httptest.NewRequest("GET", "/api/clients", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("expected 302 redirect, got %d", w.Code)
	}
}

func TestPathIndexRedirectsToLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := setupRouter()

	req := httptest.NewRequest("GET", "/mypath/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("expected 302 redirect, got %d", w.Code)
	}
}

func TestPathAPIClientsRedirectsToLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := setupRouter()

	req := httptest.NewRequest("GET", "/mypath/api/clients", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("expected 302 redirect, got %d", w.Code)
	}
}

func TestAPIClientsError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg.APIBaseURL = "http://localhost:19999"

	r := gin.New()
	r.GET("/api/clients", handleAPIClientsRoot)

	req := httptest.NewRequest("GET", "/api/clients", nil)
	req.AddCookie(&http.Cookie{Name: "auth_token", Value: "test"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestAPICommandError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg.APIBaseURL = "http://localhost:19999"

	r := gin.New()
	r.POST("/api/command", handleExecCommand)

	payload := `{"client_id":"test","command":"shell_exec"}`
	req := httptest.NewRequest("POST", "/api/command", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "auth_token", Value: "test"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestFileListMissingClientID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.POST("/api/file/list", handleFileList)

	payload := `{"path":"/tmp"}`
	req := httptest.NewRequest("POST", "/api/file/list", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "auth_token", Value: "test"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestFileMoveMissingRequiredFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.POST("/api/file/move", handleFileMove)

	payload := `{"client_id":"test"}`
	req := httptest.NewRequest("POST", "/api/file/move", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "auth_token", Value: "test"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestFileDeleteMissingClientID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.POST("/api/file/delete", handleFileDelete)

	payload := `{"path":"/tmp/test.txt"}`
	req := httptest.NewRequest("POST", "/api/file/delete", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "auth_token", Value: "test"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestFileListErrorNoWebsocket(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg.APIBaseURL = "http://localhost:19999"

	r := gin.New()
	r.POST("/api/file/list", handleFileList)

	payload := `{"client_id":"test","path":"/tmp"}`
	req := httptest.NewRequest("POST", "/api/file/list", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "auth_token", Value: "test"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestFileMoveErrorNoWebsocket(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg.APIBaseURL = "http://localhost:19999"

	r := gin.New()
	r.POST("/api/file/move", handleFileMove)

	payload := `{"client_id":"test","origin_path":"/tmp/a.txt","new_path":"/tmp/b.txt"}`
	req := httptest.NewRequest("POST", "/api/file/move", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "auth_token", Value: "test"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestFileDeleteErrorNoWebsocket(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg.APIBaseURL = "http://localhost:19999"

	r := gin.New()
	r.POST("/api/file/delete", handleFileDelete)

	payload := `{"client_id":"test","path":"/tmp/test.txt"}`
	req := httptest.NewRequest("POST", "/api/file/delete", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "auth_token", Value: "test"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}
