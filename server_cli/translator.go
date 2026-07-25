package main

import (
	"fmt"

	"RATFF/shared"
)

// Translator wraps shared translation functions with a bound language code.
type Translator struct {
	lang string
}

// NewTranslator creates a Translator instance using the current config language.
func NewTranslator() *Translator {
	lang := cfg.Language
	if lang == "" {
		lang = "en"
	}
	return &Translator{lang: lang}
}

// T looks up a translation key in the bound language.
func (t *Translator) T(key string) string {
	return shared.T(t.lang, key)
}

// Tf looks up a translation key and formats it with the provided arguments.
func (t *Translator) Tf(key string, args ...interface{}) string {
	return shared.Tf(t.lang, key, args...)
}

// TranslateFunc returns a closure for convenient inline translation.
func (t *Translator) TranslateFunc() func(string) string {
	return t.T
}

// TranslateFormatFunc returns a closure for convenient inline formatted translation.
func (t *Translator) TranslateFormatFunc() func(string, ...interface{}) string {
	return t.Tf
}

// MustT is like T but panics if the translator is nil.
func MustT(t *Translator) func(string) string {
	if t == nil {
		panic("translator is nil")
	}
	return t.T
}

// MustTf is like Tf but panics if the translator is nil.
func MustTf(t *Translator) func(string, ...interface{}) string {
	if t == nil {
		panic("translator is nil")
	}
	return t.Tf
}

// Helper functions for backward compatibility during migration.
func T(key string) string {
	return shared.T(cfg.Language, key)
}

func Tf(key string, args ...interface{}) string {
	return fmt.Sprintf(shared.T(cfg.Language, key), args...)
}
