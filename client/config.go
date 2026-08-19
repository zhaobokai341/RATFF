package main

import "RATFF/shared"

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
		ServerHost:   shared.GetEnv("HOST", "localhost"),
		ServerPort:   shared.GetEnv("PORT", "6341"),
		PathPassword: shared.GetEnv("PATH_PASSWORD", ""),
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
