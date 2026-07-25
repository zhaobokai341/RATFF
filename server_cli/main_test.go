package main

import (
	"os"
	"testing"

	"RATFF/shared"
)

func TestMain(m *testing.M) {
	_ = shared.LoadLanguagePacks("lang")
	os.Exit(m.Run())
}

// TestBuildPrompt tests prompt generation in different modes.
func TestBuildPrompt(t *testing.T) {
	tests := []struct {
		name          string
		id            string
		inCommandMode bool
	}{
		{"server_mode", "", false},
		{"console_mode", "dev-01", false},
		{"command_mode", "dev-01", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildPrompt(tt.id, tt.inCommandMode)
			if got == "" {
				t.Errorf("BuildPrompt(%q, %v) returned empty string", tt.id, tt.inCommandMode)
			}
		})
	}
}

// TestHandleConsoleMode tests all console mode commands.
func TestHandleConsoleModeHelp(t *testing.T) {
	result := handleConsoleMode("help")
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestHandleConsoleModeCommand(t *testing.T) {
	result := handleConsoleMode("command")
	if result != "enter_command" {
		t.Errorf("expected enter_command, got %q", result)
	}
}

func TestHandleConsoleModeBack(t *testing.T) {
	result := handleConsoleMode("back")
	if result != "back" {
		t.Errorf("expected back, got %q", result)
	}
}

func TestHandleConsoleModeExit(t *testing.T) {
	result := handleConsoleMode("exit")
	if result != "exit" {
		t.Errorf("expected exit, got %q", result)
	}
}

func TestHandleConsoleModeInvalid(t *testing.T) {
	result := handleConsoleMode("invalid")
	if result != "" {
		t.Errorf("expected empty string for invalid command, got %q", result)
	}
}

// TestHandleServerMode tests server mode command handling.
func TestHandleServerModeHelp(t *testing.T) {
	selectedID := handleServerMode("help", "")
	if selectedID != "" {
		t.Errorf("expected empty selectedID, got %q", selectedID)
	}
}

func TestHandleServerModeInvalid(t *testing.T) {
	selectedID := handleServerMode("invalid", "")
	if selectedID != "" {
		t.Errorf("expected empty selectedID for invalid command, got %q", selectedID)
	}
}

func TestHandleServerModeSelectNoID(t *testing.T) {
	selectedID := "test-id"
	result := handleServerMode("select", selectedID)
	if result != selectedID {
		t.Errorf("expected %q, got %q", selectedID, result)
	}
}

func TestHandleServerModeDeleteNoID(t *testing.T) {
	selectedID := "test-id"
	result := handleServerMode("delete", selectedID)
	if result != selectedID {
		t.Errorf("expected %q, got %q", selectedID, result)
	}
}

func TestClearScreen(t *testing.T) {
	clearScreen()
}

// TestOutputFunctions tests all output styling functions.
func TestPrintSuccess(t *testing.T) {
	PrintSuccess("Test success message")
}

func TestPrintError(t *testing.T) {
	PrintError("Test error message")
}

func TestPrintInfo(t *testing.T) {
	PrintInfo("Test info message")
}

func TestPrintDebug(t *testing.T) {
	PrintDebug("Test debug message")
}

func TestPrintWarn(t *testing.T) {
	PrintWarn("Test warn message")
}

func TestStyleCommandOutput(t *testing.T) {
	output := StyleCommandOutput("test output")
	if output == "" {
		t.Error("StyleCommandOutput returned empty string")
	}
}

func TestPrintHelp(t *testing.T) {
	PrintHelp([]HelpCommand{
		{"test", "Test command"},
	})
}

// TestClientTable tests client table rendering.
func TestPrintClientTableEmpty(t *testing.T) {
	PrintClientTable(nil)
}

// TestFormatID tests ID formatting for table display.
func TestFormatID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxWidth int
	}{
		{"short_id", "abc123", 20},
		{"exact_width", "12345678901234567890", 20},
		{"long_id", "123456789012345678901234567890", 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatID(tt.input)
			if result == "" {
				t.Error("formatID returned empty string")
			}
		})
	}
}

// TestTranslations tests translation functions.
func TestTSimpleKey(t *testing.T) {
	result := T("login_success")
	if result == "" {
		t.Error("T() returned empty string for valid key")
	}
}

func TestTInvalidKey(t *testing.T) {
	result := T("invalid_key_that_does_not_exist")
	if result != "invalid_key_that_does_not_exist" {
		t.Errorf("expected key itself, got %q", result)
	}
}

func TestTfWithArgs(t *testing.T) {
	result := Tf("login_failed", "test error")
	if result == "" {
		t.Error("Tf() returned empty string")
	}
}

func TestTfNoArgs(t *testing.T) {
	result := Tf("login_success")
	if result == "" {
		t.Error("Tf() returned empty string")
	}
}

// TestConfigURLGeneration tests URL building functions.
func TestGetAPIBaseURLWithPathPassword(t *testing.T) {
	originalPathPwd := cfg.PathPassword
	originalPort := cfg.Port
	cfg.PathPassword = "secret"
	cfg.Port = "6341"
	defer func() {
		cfg.PathPassword = originalPathPwd
		cfg.Port = originalPort
	}()

	result := getAPIBaseURL()
	expected := "http://localhost:6341/secret"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestGetAPIBaseURLWithoutPathPassword(t *testing.T) {
	originalPathPwd := cfg.PathPassword
	originalPort := cfg.Port
	cfg.PathPassword = ""
	cfg.Port = "6341"
	defer func() {
		cfg.PathPassword = originalPathPwd
		cfg.Port = originalPort
	}()

	result := getAPIBaseURL()
	expected := "http://localhost:6341"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestGetWSURLWithPathPassword(t *testing.T) {
	originalPathPwd := cfg.PathPassword
	originalPort := cfg.Port
	cfg.PathPassword = "secret"
	cfg.Port = "6341"
	defer func() {
		cfg.PathPassword = originalPathPwd
		cfg.Port = originalPort
	}()

	result := getWSURL()
	expected := "ws://localhost:6341/secret/ws"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestGetWSURLWithoutPathPassword(t *testing.T) {
	originalPathPwd := cfg.PathPassword
	originalPort := cfg.Port
	cfg.PathPassword = ""
	cfg.Port = "6341"
	defer func() {
		cfg.PathPassword = originalPathPwd
		cfg.Port = originalPort
	}()

	result := getWSURL()
	expected := "ws://localhost:6341/ws"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// TestHandleServerModeList tests list command in server mode.
func TestHandleServerModeList(t *testing.T) {
	selectedID := handleServerMode("list", "")
	if selectedID != "" {
		t.Errorf("expected empty selectedID, got %q", selectedID)
	}
}

// TestHandleServerModeClear tests clear command in server mode.
func TestHandleServerModeClear(t *testing.T) {
	selectedID := handleServerMode("clear", "")
	if selectedID != "" {
		t.Errorf("expected empty selectedID, got %q", selectedID)
	}
}
