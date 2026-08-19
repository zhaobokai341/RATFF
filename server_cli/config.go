package main

import "RATFF/shared"

// Config holds the CLI client configuration values.
type Config struct {
	Host          string
	Port          string
	PathPassword  string
	LoginPassword string
	Language      string
}

// cfg is the active configuration instance.
var cfg Config

func loadConfig() {
	cfg = Config{
		Host:     shared.GetEnv("HOST", "0.0.0.0"),
		Port:     shared.GetEnv("PORT", "6341"),
		Language: shared.GetEnv("LANGUAGE", "en"),
	}
}

func getAPIBaseURL() string {
	base := "http://" + cfg.Host
	if cfg.Port != "" {
		base += ":" + cfg.Port
	}
	if cfg.PathPassword != "" {
		base += "/" + cfg.PathPassword
	}
	return base
}

func getWSURL() string {
	base := "ws://" + cfg.Host
	if cfg.Port != "" {
		base += ":" + cfg.Port
	}
	if cfg.PathPassword != "" {
		base += "/" + cfg.PathPassword
	}
	return base + "/ws"
}
