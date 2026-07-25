package main

import (
	"os"

	"github.com/sirupsen/logrus"
)

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
		PathPassword:      getEnv("LOGIN_PATH", "fuck"),
		LoginPasswordHash: getEnv("LOGIN_PASSWORD_HASH", "$2b$12$lfEEs6tTAdp61DYg7xiorOkspqK2iTObW/qK6fOsT6JxBfbRBGjn2"),
		JWTSecret:         getEnv("JWT_SECRET", "default-jwt-secret-change-in-production"),
	}

	if os.Getenv("LOGIN_PATH") == "" {
		logrus.Warn("PATH_PASSWORD not set, using default value (insecure for production)")
	}
	if os.Getenv("LOGIN_PASSWORD_HASH") == "" {
		logrus.Warn("LOGIN_PASSWORD_HASH not set, using default value (insecure for production)")
	}
	if os.Getenv("JWT_SECRET") == "" {
		logrus.Warn("JWT_SECRET not set, using default value (insecure for production)")
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
