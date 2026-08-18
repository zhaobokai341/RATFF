package main

import (
	"RATFF/shared"

	"github.com/gin-gonic/gin"
)

var globalPerClientLimiter = shared.NewPerClientRateLimiter()

// handleListClients returns a handler that lists connected clients.
func handleListClients(manager *ClientManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		clients := manager.ListClients()
		c.JSON(200, gin.H{"clients": clients, "count": len(clients)})
	}
}

// handleSendCommand returns a handler that sends commands to clients.
func handleSendCommand(manager *ClientManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ClientID string                 `json:"client_id" binding:"required"`
			Command  string                 `json:"command" binding:"required"`
			Payload  map[string]interface{} `json:"payload"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		msg := shared.NewMessage(shared.MsgCommand, shared.CommandType(req.Command), req.ClientID, req.Payload)

		if err := manager.Send(req.ClientID, msg); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"status": "sent", "message_id": msg.ID})
	}
}
