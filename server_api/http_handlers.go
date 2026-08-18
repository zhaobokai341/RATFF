package main

import (
	"sync"
	"time"

	"RATFF/shared"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// rateLimiterEntry holds a rate limiter with its last access time for cleanup.
type rateLimiterEntry struct {
	limiter  *rate.Limiter
	lastUsed time.Time
}

// perClientRateLimiter manages per-client rate limiters with automatic cleanup.
type perClientRateLimiter struct {
	limiters map[string]*rateLimiterEntry
	mu       sync.Mutex
}

// cleanupInterval is how often to run the cleanup goroutine.
const cleanupInterval = 10 * time.Minute

// staleThreshold is the time after which an unused limiter is removed.
const staleThreshold = 30 * time.Minute

var globalPerClientLimiter = &perClientRateLimiter{
	limiters: make(map[string]*rateLimiterEntry),
}

// init starts the background cleanup goroutine for stale rate limiters.
func init() {
	go func() {
		ticker := time.NewTicker(cleanupInterval)
		defer ticker.Stop()
		for range ticker.C {
			globalPerClientLimiter.cleanup()
		}
	}()
}

// cleanup removes rate limiters that have not been used for longer than staleThreshold.
func (p *perClientRateLimiter) cleanup() {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	for id, entry := range p.limiters {
		if now.Sub(entry.lastUsed) > staleThreshold {
			delete(p.limiters, id)
		}
	}
}

// getLimiter returns or creates a rate limiter for a client.
func (p *perClientRateLimiter) getLimiter(clientID string) *rate.Limiter {
	p.mu.Lock()
	defer p.mu.Unlock()

	entry, exists := p.limiters[clientID]
	if !exists {
		// 10000 requests per second per client, burst 2000
		// Supports ~640MB/s file transfer with 64KB chunks
		limiter := rate.NewLimiter(rate.Every(time.Millisecond/10), 2000)
		p.limiters[clientID] = &rateLimiterEntry{
			limiter:  limiter,
			lastUsed: time.Now(),
		}
		return limiter
	}

	entry.lastUsed = time.Now()
	return entry.limiter
}

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

// rateLimitMiddleware creates a global rate limiter for non-API endpoints.
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

// apiRateLimitMiddleware applies per-client rate limiting for API endpoints.
func apiRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			token = c.ClientIP()
		}

		limiter := globalPerClientLimiter.getLimiter(token)
		if !limiter.Allow() {
			c.JSON(429, gin.H{"error": "too many requests"})
			c.Abort()
			return
		}
		c.Next()
	}
}
