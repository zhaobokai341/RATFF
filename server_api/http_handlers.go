package main

import (
	"time"

	"RATFF/shared"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

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
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"status": "sent", "message_id": msg.ID})
	}
}

// rateLimitMiddleware creates a middleware that limits requests to 50 per second.
func rateLimitMiddleware() gin.HandlerFunc {
	limiter := rate.NewLimiter(rate.Every(time.Second), 50)

	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.JSON(429, gin.H{"error": "too many requests"})
			c.Abort()
			return
		}
		c.Next()
	}
}
