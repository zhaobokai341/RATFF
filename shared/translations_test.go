package shared

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadLanguagePacksSuccess(t *testing.T) {
	tmpDir := t.TempDir()

	enContent := `{"hello": "Hello", "goodbye": "Goodbye"}`
	zhContent := `{"hello": "你好", "goodbye": "再见"}`

	if err := os.WriteFile(filepath.Join(tmpDir, "en.json"), []byte(enContent), 0644); err != nil {
		t.Fatalf("Failed to create en.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "zh.json"), []byte(zhContent), 0644); err != nil {
		t.Fatalf("Failed to create zh.json: %v", err)
	}

	if err := LoadLanguagePacks(tmpDir); err != nil {
		t.Fatalf("LoadLanguagePacks failed: %v", err)
	}

	if len(langPacks) != 2 {
		t.Errorf("Expected 2 language packs, got %d", len(langPacks))
	}
}

func TestLoadLanguagePacksDirectoryNotFound(t *testing.T) {
	err := LoadLanguagePacks("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Error("Expected error for nonexistent directory")
	}
}

func TestLoadLanguagePacksInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(tmpDir, "en.json"), []byte(`{invalid json`), 0644); err != nil {
		t.Fatalf("Failed to create en.json: %v", err)
	}

	err := LoadLanguagePacks(tmpDir)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestLoadLanguagePacksIgnoresNonJSON(t *testing.T) {
	tmpDir := t.TempDir()

	enContent := `{"hello": "Hello"}`
	if err := os.WriteFile(filepath.Join(tmpDir, "en.json"), []byte(enContent), 0644); err != nil {
		t.Fatalf("Failed to create en.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("readme"), 0644); err != nil {
		t.Fatalf("Failed to create readme.txt: %v", err)
	}
	if err := os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	if err := LoadLanguagePacks(tmpDir); err != nil {
		t.Fatalf("LoadLanguagePacks failed: %v", err)
	}

	if len(langPacks) != 1 {
		t.Errorf("Expected 1 language pack, got %d", len(langPacks))
	}
}

func TestTSuccess(t *testing.T) {
	langPacks = map[string]map[string]string{
		"en": {"hello": "Hello", "key1": "Value1"},
		"zh": {"hello": "你好", "key2": "值2"},
	}

	if got := T("en", "hello"); got != "Hello" {
		t.Errorf("Expected 'Hello', got '%s'", got)
	}

	if got := T("zh", "hello"); got != "你好" {
		t.Errorf("Expected '你好', got '%s'", got)
	}
}

func TestTFallbackToOtherLanguage(t *testing.T) {
	langPacks = map[string]map[string]string{
		"en": {"hello": "Hello", "only_en": "English only"},
		"zh": {"hello": "你好"},
	}

	if got := T("zh", "only_en"); got != "English only" {
		t.Errorf("Expected fallback 'English only', got '%s'", got)
	}
}

func TestTReturnsKeyWhenNotFound(t *testing.T) {
	langPacks = map[string]map[string]string{
		"en": {"hello": "Hello"},
	}

	if got := T("en", "nonexistent_key"); got != "nonexistent_key" {
		t.Errorf("Expected key itself, got '%s'", got)
	}
}

func TestTEmptyLangCode(t *testing.T) {
	langPacks = map[string]map[string]string{
		"en": {"hello": "Hello"},
	}

	if got := T("", "hello"); got != "Hello" {
		t.Errorf("Expected 'Hello' with empty lang code, got '%s'", got)
	}
}

func TestTNoLangPacks(t *testing.T) {
	langPacks = nil

	if got := T("en", "hello"); got != "hello" {
		t.Errorf("Expected key itself when no packs loaded, got '%s'", got)
	}
}

func TestTfSuccess(t *testing.T) {
	langPacks = map[string]map[string]string{
		"en": {"greeting": "Hello, %s!", "count": "You have %d messages"},
	}

	if got := Tf("en", "greeting", "World"); got != "Hello, World!" {
		t.Errorf("Expected 'Hello, World!', got '%s'", got)
	}

	if got := Tf("en", "count", 5); got != "You have 5 messages" {
		t.Errorf("Expected 'You have 5 messages', got '%s'", got)
	}
}

func TestTfNoArgs(t *testing.T) {
	langPacks = map[string]map[string]string{
		"en": {"hello": "Hello"},
	}

	if got := Tf("en", "hello"); got != "Hello" {
		t.Errorf("Expected 'Hello', got '%s'", got)
	}
}
