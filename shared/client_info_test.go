package shared

import (
	"testing"
	"time"
)

func TestBuildClientInfo(t *testing.T) {
	info := BuildClientInfo("test-id")

	if info.ID != "test-id" {
		t.Errorf("expected ID test-id, got %s", info.ID)
	}
	if info.Hostname == "" {
		t.Error("hostname should not be empty")
	}
	if info.OSInfo == "" {
		t.Error("os_info should not be empty")
	}
}

func TestClientInfoToPayload(t *testing.T) {
	info := ClientInfo{
		ID:       "id-1",
		IP:       "127.0.0.1",
		Hostname: "test-host",
		OSInfo:   "linux test-host go1.21.0 amd64",
	}

	payload := info.ToPayload()

	if payload["id"] != "id-1" {
		t.Errorf("expected id-1, got %v", payload["id"])
	}
	if payload["ip"] != "127.0.0.1" {
		t.Errorf("expected 127.0.0.1, got %v", payload["ip"])
	}
}

func TestClientInfoFromPayload(t *testing.T) {
	payload := map[string]interface{}{
		"id":       "id-2",
		"ip":       "10.0.0.1",
		"hostname": "my-host",
		"os_info":  "linux my-host go1.21.0 amd64",
	}

	info := ClientInfoFromPayload(payload)

	if info.ID != "id-2" {
		t.Errorf("expected id-2, got %s", info.ID)
	}
	if info.IP != "10.0.0.1" {
		t.Errorf("expected 10.0.0.1, got %s", info.IP)
	}
	if info.Hostname != "my-host" {
		t.Errorf("expected my-host, got %s", info.Hostname)
	}
	if info.OSInfo == "" {
		t.Error("os_info should not be empty")
	}
}

func TestClientInfoFromPayloadEmpty(t *testing.T) {
	payload := map[string]interface{}{}
	info := ClientInfoFromPayload(payload)

	if info.ID != "" {
		t.Errorf("expected empty ID, got %s", info.ID)
	}
}

func TestGenerateClientID(t *testing.T) {
	id := GenerateClientID()
	if len(id) == 0 {
		t.Error("client ID should not be empty")
	}
}

func TestCalculateBackoff(t *testing.T) {
	tests := []struct {
		attempt int
		minDur  time.Duration
		maxDur  time.Duration
	}{
		{0, time.Second, time.Second},
		{1, 2 * time.Second, 2 * time.Second},
		{2, 4 * time.Second, 4 * time.Second},
		{10, 30 * time.Second, 30 * time.Second},
	}

	for _, tt := range tests {
		d := CalculateBackoff(tt.attempt)
		if d < tt.minDur || d > tt.maxDur {
			t.Errorf("attempt %d: expected %v-%v, got %v", tt.attempt, tt.minDur, tt.maxDur, d)
		}
	}
}
