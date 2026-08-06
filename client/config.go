package main

import "os"

// Config holds the client configuration values.
type Config struct {
	ServerHost   string
	ServerPort   string
	PathPassword string
	ClientID     string
}

// cfg is the active configuration instance.
var cfg Config

func loadConfig() {
	cfg = Config{
		ServerHost:   getEnv("SERVER_HOST", "localhost"),
		ServerPort:   getEnv("SERVER_PORT", "6341"),
		PathPassword: getEnv("PATH_PASSWORD", ""),
		ClientID:     "",
	}
}

func getServerURL() string {
	baseURL := "ws://" + cfg.ServerHost + ":" + cfg.ServerPort
	if cfg.PathPassword != "" {
		return baseURL + "/" + cfg.PathPassword + "/ws"
	}
	return baseURL + "/ws"
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
