package main

import (
	"net"
	"net/http"
	"time"

	"RATFF/shared"
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
		Host:         shared.GetEnv("HOST", "0.0.0.0"),
		Port:         shared.GetEnv("PORT", "7993"),
		APIBaseURL:   shared.GetEnv("API_URL", "http://127.0.0.1:6341"),
		WsURL:        shared.GetEnv("WS_URL", "ws://127.0.0.1:6341"),
		CookieSecure: !shared.IsDebugEnv(),
	}
}
