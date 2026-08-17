package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildAPIURLWithPathPrefix(t *testing.T) {
	cfg.APIBaseURL = "http://localhost:6341"
	result := buildAPIURL("secret", "/api/command")
	assert.Equal(t, "http://localhost:6341/secret/api/command", result)
}

func TestBuildAPIURLWithoutPathPrefix(t *testing.T) {
	cfg.APIBaseURL = "http://localhost:6341"
	result := buildAPIURL("", "/api/command")
	assert.Equal(t, "http://localhost:6341/api/command", result)
}

func TestBuildWSURLWithPathPrefix(t *testing.T) {
	cfg.WsURL = "ws://localhost:6341"
	result := buildWSURL("secret")
	assert.Equal(t, "ws://localhost:6341/secret/ws", result)
}

func TestBuildWSURLWithoutPathPrefix(t *testing.T) {
	cfg.WsURL = "ws://localhost:6341"
	result := buildWSURL("")
	assert.Equal(t, "ws://localhost:6341/ws", result)
}

func TestBuildWSURLWithDefault(t *testing.T) {
	cfg.WsURL = ""
	result := buildWSURL("")
	assert.Equal(t, "ws://127.0.0.1:6341/ws", result)
}
