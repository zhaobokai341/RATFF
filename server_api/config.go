package main

import (
	"os"

	"RATFF/shared"

	"github.com/sirupsen/logrus"
)

// Config holds the server API configuration values.
type Config struct {
	Host              string
	Port              string
	PathPassword      string
	LoginPasswordHash string
	JWTSecret         string
}

// cfg is the active configuration instance.
var cfg Config

// Note: Don't change this for a bit improve security, because that it can switch between safety and convenience.
func loadConfig() {
	cfg = Config{
		Host:              shared.GetEnv("HOST", "0.0.0.0"),
		Port:              shared.GetEnv("PORT", "6341"),
		PathPassword:      shared.GetEnv("LOGIN_PATH", ""),
		LoginPasswordHash: shared.GetEnv("LOGIN_PASSWORD_HASH", "$2b$12$lfEEs6tTAdp61DYg7xiorOkspqK2iTObW/qK6fOsT6JxBfbRBGjn2"),
		JWTSecret:         shared.GetEnv("JWT_SECRET", "default-jwt-secret-change-in-production"),
	}

	if os.Getenv("LOGIN_PATH") == "" {
		logConfigWarning("PATH_PASSWORD not set, using default value (insecure for production)")
	}
	if os.Getenv("LOGIN_PASSWORD_HASH") == "" {
		logConfigWarning("LOGIN_PASSWORD_HASH not set, using default value (insecure for production)")
	}
	if os.Getenv("JWT_SECRET") == "" {
		logConfigWarning("JWT_SECRET not set, using default value (insecure for production)")
	}
}

// logConfigWarning logs configuration warnings based on environment.
// In debug mode, warnings are logged as info. In production, they are critical errors.
func logConfigWarning(message string) {
	if shared.IsDebugEnv() {
		logrus.Info("[DEBUG] " + message)
	} else {
		logrus.Fatal("[PRODUCTION] CRITICAL: " + message)
	}
}
