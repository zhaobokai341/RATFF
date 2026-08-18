package shared

import (
	"bytes"
	"testing"
	"time"
)

func TestAsyncWriterWrite(t *testing.T) {
	var buf bytes.Buffer
	aw := NewAsyncWriter(&buf, 10)

	_, err := aw.Write([]byte("hello"))
	if err != nil {
		t.Errorf("Write failed: %v", err)
	}

	aw.Close()

	if buf.String() != "hello" {
		t.Errorf("Expected 'hello', got '%s'", buf.String())
	}
}

func TestAsyncWriterMultipleWrites(t *testing.T) {
	var buf bytes.Buffer
	aw := NewAsyncWriter(&buf, 10)

	_, _ = aw.Write([]byte("hello"))
	_, _ = aw.Write([]byte(" "))
	_, _ = aw.Write([]byte("world"))

	aw.Close()

	if buf.String() != "hello world" {
		t.Errorf("Expected 'hello world', got '%s'", buf.String())
	}
}

func TestAsyncWriterClose(t *testing.T) {
	var buf bytes.Buffer
	aw := NewAsyncWriter(&buf, 10)

	_, _ = aw.Write([]byte("test"))
	aw.Close()

	if buf.String() != "test" {
		t.Errorf("Expected 'test', got '%s'", buf.String())
	}
}

func TestAsyncWriterFullChannel(t *testing.T) {
	var buf bytes.Buffer
	aw := NewAsyncWriter(&buf, 2)

	_, _ = aw.Write([]byte("msg1"))
	_, _ = aw.Write([]byte("msg2"))
	_, _ = aw.Write([]byte("msg3"))

	aw.Close()

	if buf.Len() == 0 {
		t.Error("Expected some output")
	}
}

func TestCalculateBackoffZero(t *testing.T) {
	d := CalculateBackoff(0)
	if d != time.Second {
		t.Errorf("Expected 1s, got %v", d)
	}
}

func TestCalculateBackoffSmall(t *testing.T) {
	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{0, 1 * time.Second},
		{1, 2 * time.Second},
		{2, 4 * time.Second},
		{3, 8 * time.Second},
		{4, 16 * time.Second},
	}

	for _, tt := range tests {
		d := CalculateBackoff(tt.attempt)
		if d != tt.want {
			t.Errorf("attempt %d: expected %v, got %v", tt.attempt, tt.want, d)
		}
	}
}

func TestCalculateBackoffCapped(t *testing.T) {
	d := CalculateBackoff(10)
	if d != 30*time.Second {
		t.Errorf("Expected 30s (capped), got %v", d)
	}

	d2 := CalculateBackoff(100)
	if d2 != 30*time.Second {
		t.Errorf("Expected 30s (capped), got %v", d2)
	}
}

func TestCalculateBackoffNegative(t *testing.T) {
	d := CalculateBackoff(-1)
	if d < 0 {
		t.Errorf("Negative attempt should not produce negative duration, got %v", d)
	}
}

func TestEncryptAESInvalidKeySize(t *testing.T) {
	plaintext := []byte("test data")

	invalidKeySizes := []int{0, 1, 8, 15, 20, 30, 33, 64}

	for _, size := range invalidKeySizes {
		key := make([]byte, size)
		_, err := EncryptAES(plaintext, key)
		if err == nil {
			t.Errorf("EncryptAES with key size %d should fail, but got no error", size)
		}
	}
}

func TestDecryptAESInvalidKeySize(t *testing.T) {
	invalidKeySizes := []int{0, 1, 8, 15, 20, 30, 33, 64}

	for _, size := range invalidKeySizes {
		key := make([]byte, size)
		_, err := DecryptAES([]byte("test"), key)
		if err == nil {
			t.Errorf("DecryptAES with key size %d should fail, but got no error", size)
		}
	}
}

func TestEncryptDecryptAESValidKeySizes(t *testing.T) {
	validKeySizes := []int{16, 24, 32}
	plaintext := []byte("test data for valid key sizes")

	for _, size := range validKeySizes {
		key := make([]byte, size)
		for i := range key {
			key[i] = byte(i)
		}

		ciphertext, err := EncryptAES(plaintext, key)
		if err != nil {
			t.Errorf("EncryptAES with key size %d failed: %v", size, err)
			continue
		}

		decrypted, err := DecryptAES(ciphertext, key)
		if err != nil {
			t.Errorf("DecryptAES with key size %d failed: %v", size, err)
			continue
		}

		if string(decrypted) != string(plaintext) {
			t.Errorf("Key size %d: expected %s, got %s", size, plaintext, decrypted)
		}
	}
}
