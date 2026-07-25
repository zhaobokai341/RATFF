package main

import (
	"os"
	"testing"

	"RATFF/shared"
)

func TestMain(m *testing.M) {
	log = shared.InitLogger("info", "text")
	os.Exit(m.Run())
}

func TestBuildClientInfo(t *testing.T) {
	info := shared.BuildClientInfo("test-123")

	if info.ID != "test-123" {
		t.Errorf("expected ID test-123, got %s", info.ID)
	}
	if info.Hostname == "" {
		t.Error("hostname should not be empty")
	}
	if info.OSInfo == "" {
		t.Error("os_info should not be empty")
	}
}

func TestGetEnv(t *testing.T) {
	t.Setenv("TEST_KEY", "test_value")

	if got := getEnv("TEST_KEY", "default"); got != "test_value" {
		t.Errorf("expected test_value, got %s", got)
	}

	if got := getEnv("NONEXISTENT", "default"); got != "default" {
		t.Errorf("expected default, got %s", got)
	}
}

func TestExecuteCommandExit(t *testing.T) {
	msg := shared.NewMessage(shared.MsgCommand, shared.CmdExit, "test-client", nil)

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected os.Exit to be called")
		}
	}()

	_ = executeCommand(msg)
}

func TestExecuteCommandUnknown(t *testing.T) {
	msg := shared.NewMessage(shared.MsgCommand, "unknown_cmd", "test-client", nil)

	resp := executeCommand(msg)

	if resp.Type != shared.MsgError {
		t.Errorf("expected error type, got %s", resp.Type)
	}
}

func TestHandleShellExecEmpty(t *testing.T) {
	msg := shared.NewMessage(shared.MsgCommand, shared.CmdShellExec, "test-client",
		map[string]interface{}{"cmd": ""})

	resp := handleShellExec(msg)

	if resp.Type != shared.MsgError {
		t.Errorf("expected error for empty command, got %s", resp.Type)
	}
}

func TestHandleShellExecEcho(t *testing.T) {
	msg := shared.NewMessage(shared.MsgCommand, shared.CmdShellExec, "test-client",
		map[string]interface{}{"cmd": "echo hello"})

	resp := handleShellExec(msg)

	if resp.Type != shared.MsgResponse {
		t.Fatalf("expected response, got %s", resp.Type)
	}
	if resp.Payload["output"] == nil {
		t.Error("expected output in payload")
	}
}

func TestHandleSystemInfo(t *testing.T) {
	msg := shared.NewMessage(shared.MsgCommand, shared.CmdSystemInfo, "test-client", nil)

	resp := handleSystemInfo(msg)

	if resp.Type != shared.MsgResponse {
		t.Fatalf("expected response, got %s", resp.Type)
	}
	if resp.Payload["id"] != "test-client" {
		t.Errorf("expected client id in payload, got %v", resp.Payload["id"])
	}
}
