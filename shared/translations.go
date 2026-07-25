package shared

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// defaultLanguage is the fallback language code when the requested lang is not found.
const defaultLanguage = "en"

// langPacks stores all loaded language packs, keyed by language code.
var langPacks map[string]map[string]string

// LoadLanguagePacks scans the specified directory and loads all .json language files.
// Each file must be named <lang_code>.json (e.g., zh.json, en.json).
func LoadLanguagePacks(langDir string) error {
	langPacks = make(map[string]map[string]string)

	entries, err := os.ReadDir(langDir)
	if err != nil {
		return fmt.Errorf("failed to read lang directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := filepath.Ext(entry.Name())
		if ext != ".json" {
			continue
		}

		langCode := entry.Name()[:len(entry.Name())-len(ext)]
		filePath := filepath.Join(langDir, entry.Name())

		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read language file %s: %w", filePath, err)
		}

		var translations map[string]string
		if err := json.Unmarshal(data, &translations); err != nil {
			return fmt.Errorf("failed to parse language file %s: %w", filePath, err)
		}

		langPacks[langCode] = translations
	}

	return nil
}

// T looks up a translation key in the specified language pack.
// If the key is not found in the requested language, it searches other loaded
// language packs as fallback. If still not found, returns the key itself.
func T(langCode, key string) string {
	if langCode == "" {
		langCode = defaultLanguage
	}

	if pack, ok := langPacks[langCode]; ok {
		if val, exists := pack[key]; exists {
			return val
		}
	}

	for code, pack := range langPacks {
		if code == langCode {
			continue
		}
		if val, exists := pack[key]; exists {
			return val
		}
	}

	return key
}

// Tf looks up a translation key and formats it with the provided arguments.
func Tf(langCode, key string, args ...interface{}) string {
	template := T(langCode, key)
	return fmt.Sprintf(template, args...)
}
