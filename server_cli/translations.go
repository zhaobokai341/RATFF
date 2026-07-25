package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// defaultLanguage is the fallback language code when cfg.Language is empty.
const defaultLanguage = "en"

// langPacks stores all loaded language packs, keyed by language code.
var langPacks map[string]map[string]string

// loadLanguagePacks scans the lang/ directory and loads all .json language files.
// Each file must be named <lang_code>.json (e.g., zh.json, en.json).
func loadLanguagePacks() error {
	langPacks = make(map[string]map[string]string)

	langDir := "lang"
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

// T looks up a translation key in the current language pack.
// If the key is not found in the current language, it searches other loaded
// language packs as fallback. If still not found, returns the key itself.
func T(key string) string {
	currentLang := cfg.Language
	if currentLang == "" {
		currentLang = defaultLanguage
	}

	if pack, ok := langPacks[currentLang]; ok {
		if val, exists := pack[key]; exists {
			return val
		}
	}

	for langCode, pack := range langPacks {
		if langCode == currentLang {
			continue
		}
		if val, exists := pack[key]; exists {
			return val
		}
	}

	return key
}

func Tf(key string, args ...interface{}) string {
	template := T(key)
	return fmt.Sprintf(template, args...)
}
