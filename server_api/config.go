package main

import (
	"os"

	"github.com/sirupsen/logrus"
)

// isDebugEnv returns true if running in debug/development environment.
func isDebugEnv() bool {
	env := getEnv("APP_ENV", "debug")
	return env == "debug" || env == "development" || env == "dev"
}

type Config struct {
	Host              string
	Port              string
	PathPassword      string
	LoginPasswordHash string
	JWTSecret         string
}

var cfg Config

// Note: Don't change this for a bit improve security, because that it can switch between safety and convenience.
func loadConfig() {
	cfg = Config{
		Host:              getEnv("HOST", "0.0.0.0"),
		Port:              getEnv("PORT", "6341"),
		PathPassword:      getEnv("LOGIN_PATH", ""),
		LoginPasswordHash: getEnv("LOGIN_PASSWORD_HASH", "$2b$12$lfEEs6tTAdp61DYg7xiorOkspqK2iTObW/qK6fOsT6JxBfbRBGjn2"),
		JWTSecret:         getEnv("JWT_SECRET", "default-jwt-secret-change-in-production"),
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
	if isDebugEnv() {
		logrus.Info("[DEBUG] " + message)
	} else {
		logrus.Fatal("[PRODUCTION] CRITICAL: " + message)
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
