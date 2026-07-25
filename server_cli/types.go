package main

import (
	"sync"

	"RATFF/shared"
)

var (
	jwtToken string
)

type pendingCommand struct {
	ch chan shared.Message
}

var (
	pendingMu  sync.Mutex
	pendingCmd = make(map[string]*pendingCommand)
)
