package main

import (
	"encoding/base64"
	"testing"

	"RATFF/shared"

	"github.com/kbinani/screenshot"
	"github.com/stretchr/testify/assert"
)

func TestHandleScreenCaptureDefault(t *testing.T) {
	log = shared.InitLogger("error", "text")

	if screenshot.NumActiveDisplays() == 0 {
		t.Skip("No displays available for testing")
	}

	msg := shared.NewMessage(shared.MsgCommand, shared.CmdScreenCapture, "test-client", nil)
	resp := handleScreenCapture(msg)

	assert.Equal(t, shared.MsgResponse, resp.Type)
	assert.NotNil(t, resp.Payload)

	imageData, ok := resp.Payload["image_data"].(string)
	assert.True(t, ok)
	assert.NotEmpty(t, imageData)

	_, err := base64.StdEncoding.DecodeString(imageData)
	assert.NoError(t, err)

	assert.Equal(t, "png", resp.Payload["format"])
	assert.NotNil(t, resp.Payload["width"])
	assert.NotNil(t, resp.Payload["height"])
	assert.Equal(t, 0, resp.Payload["display_index"])
}

func TestHandleScreenCaptureJPEG(t *testing.T) {
	log = shared.InitLogger("error", "text")

	if screenshot.NumActiveDisplays() == 0 {
		t.Skip("No displays available for testing")
	}

	msg := shared.NewMessage(shared.MsgCommand, shared.CmdScreenCapture, "test-client",
		map[string]interface{}{
			"format":  "jpeg",
			"quality": 85.0,
		})
	resp := handleScreenCapture(msg)

	assert.Equal(t, shared.MsgResponse, resp.Type)
	assert.Equal(t, "jpeg", resp.Payload["format"])

	imageData := resp.Payload["image_data"].(string)
	decoded, err := base64.StdEncoding.DecodeString(imageData)
	assert.NoError(t, err)
	assert.NotEmpty(t, decoded)
}

func TestHandleScreenCaptureInvalidDisplay(t *testing.T) {
	log = shared.InitLogger("error", "text")

	displayCount := screenshot.NumActiveDisplays()

	msg := shared.NewMessage(shared.MsgCommand, shared.CmdScreenCapture, "test-client",
		map[string]interface{}{
			"display_index": float64(displayCount + 10),
		})
	resp := handleScreenCapture(msg)

	assert.Equal(t, shared.MsgError, resp.Type)
	assert.Contains(t, resp.Payload["error"], "invalid display index")
}

func TestHandleScreenCaptureNegativeDisplay(t *testing.T) {
	log = shared.InitLogger("error", "text")

	msg := shared.NewMessage(shared.MsgCommand, shared.CmdScreenCapture, "test-client",
		map[string]interface{}{
			"display_index": -1.0,
		})
	resp := handleScreenCapture(msg)

	assert.Equal(t, shared.MsgError, resp.Type)
	assert.Contains(t, resp.Payload["error"], "invalid display index")
}

func TestHandleScreenCaptureInvalidQuality(t *testing.T) {
	log = shared.InitLogger("error", "text")

	if screenshot.NumActiveDisplays() == 0 {
		t.Skip("No displays available for testing")
	}

	msg := shared.NewMessage(shared.MsgCommand, shared.CmdScreenCapture, "test-client",
		map[string]interface{}{
			"format":  "jpeg",
			"quality": 150.0,
		})
	resp := handleScreenCapture(msg)

	assert.Equal(t, shared.MsgResponse, resp.Type)
	assert.Equal(t, "jpeg", resp.Payload["format"])
}

func TestHandleScreenCaptureInvalidFormat(t *testing.T) {
	log = shared.InitLogger("error", "text")

	if screenshot.NumActiveDisplays() == 0 {
		t.Skip("No displays available for testing")
	}

	msg := shared.NewMessage(shared.MsgCommand, shared.CmdScreenCapture, "test-client",
		map[string]interface{}{
			"format": "bmp",
		})
	resp := handleScreenCapture(msg)

	assert.Equal(t, shared.MsgResponse, resp.Type)
	assert.Equal(t, "png", resp.Payload["format"])
}
