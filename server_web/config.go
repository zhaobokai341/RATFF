package main

import "os"

type Config struct {
	Host         string
	Port         string
	APIBaseURL   string
	WsURL        string
	CookieSecure bool
}

var cfg Config

func loadConfig() {
	cfg = Config{
		Host:         getEnv("HOST", "0.0.0.0"),
		Port:         getEnv("PORT", "7993"),
		APIBaseURL:   getEnv("API_URL", "http://localhost:6341"),
		WsURL:        getEnv("WS_URL", "ws://localhost:6341"),
		CookieSecure: !isDebugEnv(),
	}
}

// isDebugEnv returns true if running in debug/development environment.
func isDebugEnv() bool {
	env := os.Getenv("APP_ENV")
	return env == "debug" || env == "development" || env == "dev"
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
