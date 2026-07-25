package main

import "os"

type Config struct {
	Host       string
	Port       string
	APIBaseURL string
	WsURL      string
}

var cfg Config

func loadConfig() {
	cfg = Config{
		Host:       getEnv("HOST", "0.0.0.0"),
		Port:       getEnv("PORT", "7993"),
		APIBaseURL: getEnv("API_URL", "http://localhost:6341"),
		WsURL:      getEnv("WS_URL", "ws://localhost:6341"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
