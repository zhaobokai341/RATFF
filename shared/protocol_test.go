package shared

import (
	"testing"
)

func TestNewMessage(t *testing.T) {
	msg := NewMessage(MsgRegister, "", "test-client", nil)

	if msg.ID == "" {
		t.Error("message ID should not be empty")
	}
	if msg.Type != MsgRegister {
		t.Errorf("expected type %s, got %s", MsgRegister, msg.Type)
	}
	if msg.ClientID != "test-client" {
		t.Errorf("expected client_id test-client, got %s", msg.ClientID)
	}
	if msg.Timestamp == 0 {
		t.Error("timestamp should not be zero")
	}
}

func TestNewMessageWithPayload(t *testing.T) {
	payload := map[string]interface{}{"key": "value"}
	msg := NewMessage(MsgCommand, CmdShellExec, "client-1", payload)

	if msg.Command != CmdShellExec {
		t.Errorf("expected command %s, got %s", CmdShellExec, msg.Command)
	}
	if msg.Payload["key"] != "value" {
		t.Errorf("expected payload key=value, got %v", msg.Payload)
	}
}

func TestGenerateID(t *testing.T) {
	id1 := GenerateID()
	id2 := GenerateID()

	if id1 == id2 {
		t.Error("generated IDs should be unique")
	}
	if len(id1) == 0 {
		t.Error("generated ID should not be empty")
	}
}

func TestEncryptDecryptAES(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	plaintext := []byte("hello world")

	ciphertext, err := EncryptAES(plaintext, key)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	if string(ciphertext) == string(plaintext) {
		t.Error("ciphertext should differ from plaintext")
	}

	decrypted, err := DecryptAES(ciphertext, key)
	if err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Errorf("expected %s, got %s", plaintext, decrypted)
	}
}

func TestDecryptAESWrongKey(t *testing.T) {
	key1 := make([]byte, 32)
	key2 := make([]byte, 32)
	key2[0] = 1

	plaintext := []byte("secret")
	ciphertext, _ := EncryptAES(plaintext, key1)

	_, err := DecryptAES(ciphertext, key2)
	if err == nil {
		t.Error("decrypt with wrong key should fail")
	}
}

func TestDecryptAESShortCiphertext(t *testing.T) {
	key := make([]byte, 32)
	_, err := DecryptAES([]byte("short"), key)
	if err != nil {
		t.Errorf("short ciphertext should return nil, got error: %v", err)
	}
}

func TestInitLogger(t *testing.T) {
	logger := InitLogger("info", "text")
	if logger == nil {
		t.Error("logger should not be nil")
	}

	loggerJSON := InitLogger("debug", "json")
	if loggerJSON == nil {
		t.Error("json logger should not be nil")
	}

	loggerInvalid := InitLogger("invalid", "text")
	if loggerInvalid == nil {
		t.Error("logger with invalid level should default to info, not nil")
	}
}

func TestInitLoggerAsync(t *testing.T) {
	logger := InitLoggerAsync("info", "text", true)
	if logger == nil {
		t.Error("async logger should not be nil")
	}

	loggerSync := InitLoggerAsync("info", "text", false)
	if loggerSync == nil {
		t.Error("sync logger should not be nil")
	}
}

func TestInitLoggerWithWriter(t *testing.T) {
	logger, writer := InitLoggerWithWriter("info", "text", true)
	if logger == nil {
		t.Error("logger should not be nil")
	}
	if writer == nil {
		t.Error("async writer should not be nil when async is true")
	}

	loggerNoAsync, writerNoAsync := InitLoggerWithWriter("info", "text", false)
	if loggerNoAsync == nil {
		t.Error("logger should not be nil")
	}
	if writerNoAsync != nil {
		t.Error("async writer should be nil when async is false")
	}
}
