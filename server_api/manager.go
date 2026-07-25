package main

import (
	"encoding/json"
	"errors"
	"strings"
	"sync"

	"RATFF/shared"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// ClientEntry holds a client's WebSocket connection and info.
type ClientEntry struct {
	Conn *websocket.Conn
	Info shared.ClientInfo
}

// ClientManager manages connected clients and message routing.
type ClientManager struct {
	mu      sync.RWMutex
	clients map[string]*ClientEntry
	log     *logrus.Entry
}

// NewClientManager creates a new ClientManager instance.
func NewClientManager() *ClientManager {
	return &ClientManager{
		clients: make(map[string]*ClientEntry),
		log:     shared.InitLogger("info", "text"),
	}
}

// Register adds a client with its connection and info.
func (m *ClientManager) Register(clientID string, conn *websocket.Conn, info shared.ClientInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.clients[clientID] = &ClientEntry{Conn: conn, Info: info}
	m.log.WithField("client_id", clientID).Info("Client registered")
}

// Unregister removes a client from the manager.
func (m *ClientManager) Unregister(clientID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.clients, clientID)
	m.log.WithField("client_id", clientID).Info("Client unregistered")
}

// Send sends a message to a specific client.
func (m *ClientManager) Send(clientID string, msg shared.Message) error {
	m.mu.RLock()
	entry, ok := m.clients[clientID]
	m.mu.RUnlock()

	if !ok {
		return errors.New("client offline")
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return entry.Conn.WriteMessage(websocket.TextMessage, data)
}

// Broadcast sends a message to all clients except the specified one.
func (m *ClientManager) Broadcast(msg shared.Message, excludeID string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, _ := json.Marshal(msg)
	for id, entry := range m.clients {
		if id == excludeID {
			continue
		}
		if err := entry.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
			m.log.WithField("client_id", id).Error("Broadcast failed")
		}
	}
}

// IsOnline checks if a client is currently connected.
func (m *ClientManager) IsOnline(clientID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.clients[clientID]
	return ok
}

// ListClients returns info for all connected clients, excluding CLI connections.
func (m *ClientManager) ListClients() []shared.ClientInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	infos := make([]shared.ClientInfo, 0, len(m.clients))
	for id, entry := range m.clients {
		if strings.HasPrefix(id, "__cli__") {
			continue
		}
		infos = append(infos, entry.Info)
	}
	return infos
}
