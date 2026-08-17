package client

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"RATFF/server_cli/api"
	"RATFF/shared"

	"github.com/stretchr/testify/assert"
)

func TestScreenCaptureSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
		})
	}))
	defer server.Close()

	apiClient := api.NewClient(server.URL, "test-token")
	var capturedImage string
	var capturedWidth, capturedHeight int
	var capturedFormat string
	var capturedDispIdx, capturedDispCount int

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
			"image_data":    "dGVzdF9pbWFnZV9kYXRh",
			"width":         1920.0,
			"height":        1080.0,
			"format":        "png",
			"display_index": 0.0,
			"display_count": 2.0,
		}))
	}()

	mgr.ScreenCapture("client-1", "png", 90, 0, func(imageData string, width, height int, format string, displayIndex, displayCount int) {
		capturedImage = imageData
		capturedWidth = width
		capturedHeight = height
		capturedFormat = format
		capturedDispIdx = displayIndex
		capturedDispCount = displayCount
	})

	assert.Equal(t, "dGVzdF9pbWFnZV9kYXRh", capturedImage)
	assert.Equal(t, 1920, capturedWidth)
	assert.Equal(t, 1080, capturedHeight)
	assert.Equal(t, "png", capturedFormat)
	assert.Equal(t, 0, capturedDispIdx)
	assert.Equal(t, 2, capturedDispCount)
}

func TestScreenCaptureError(t *testing.T) {
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
			"error": "no displays found",
		}))
	}()

	mgr.ScreenCapture("client-1", "png", 90, 0, func(imageData string, width, height int, format string, displayIndex, displayCount int) {
	})

	assert.Contains(t, errorMsg, "screen_capture_failed")
}

func TestScreenCaptureTimeout(t *testing.T) {
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

	mgr.ScreenCapture("client-1", "png", 90, 0, func(imageData string, width, height int, format string, displayIndex, displayCount int) {
	})

	assert.Contains(t, errorMsg, "timeout")
}

func TestSaveScreenCapturePNG(t *testing.T) {
	imageData := base64.StdEncoding.EncodeToString([]byte("fake_png_data"))
	tmpFile, err := os.CreateTemp("", "screenshot_*.png")
	assert.NoError(t, err)
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	err = SaveScreenCapture(imageData, "png", tmpPath)
	assert.NoError(t, err)

	data, err := os.ReadFile(tmpPath)
	assert.NoError(t, err)
	assert.Equal(t, []byte("fake_png_data"), data)
}

func TestSaveScreenCaptureJPEG(t *testing.T) {
	imageData := base64.StdEncoding.EncodeToString([]byte("fake_jpeg_data"))
	tmpFile, err := os.CreateTemp("", "screenshot_*.jpg")
	assert.NoError(t, err)
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	err = SaveScreenCapture(imageData, "jpeg", tmpPath)
	assert.NoError(t, err)

	data, err := os.ReadFile(tmpPath)
	assert.NoError(t, err)
	assert.Equal(t, []byte("fake_jpeg_data"), data)
}

func TestSaveScreenCaptureDefaultPath(t *testing.T) {
	imageData := base64.StdEncoding.EncodeToString([]byte("test_data"))
	tmpDir, err := os.MkdirTemp("", "screencap_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	oldDir, err := os.Getwd()
	assert.NoError(t, err)
	err = os.Chdir(tmpDir)
	assert.NoError(t, err)
	defer os.Chdir(oldDir)

	err = SaveScreenCapture(imageData, "png", "")
	assert.NoError(t, err)

	_, err = os.Stat("screenshot.png")
	assert.NoError(t, err)
}

func TestSaveScreenCaptureInvalidBase64(t *testing.T) {
	err := SaveScreenCapture("invalid_base64!!!", "png", "test.png")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "decode base64 failed")
}
