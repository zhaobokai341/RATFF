package main

import (
	"net"
	"net/http"
	"os"
	"time"
)

// Config holds the web server configuration values.
type Config struct {
	Host         string
	Port         string
	APIBaseURL   string
	WsURL        string
	CookieSecure bool
}

// cfg is the active configuration instance.
var cfg Config

// httpClient is a shared HTTP client with configured timeouts to prevent goroutine leaks.
var httpClient = &http.Client{
	Timeout: 30 * time.Second,
	Transport: &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 20 * time.Second,
	},
}

func loadConfig() {
	cfg = Config{
		Host:         getEnv("HOST", "0.0.0.0"),
		Port:         getEnv("PORT", "7993"),
		APIBaseURL:   getEnv("API_URL", "http://127.0.0.1:6341"),
		WsURL:        getEnv("WS_URL", "ws://127.0.0.1:6341"),
		CookieSecure: !isDebugEnv(),
	}
}

// isDebugEnv returns true if running in debug/development environment.
func isDebugEnv() bool {
	env := getEnv("APP_ENV", "debug")
	return env == "debug" || env == "development" || env == "dev"
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
