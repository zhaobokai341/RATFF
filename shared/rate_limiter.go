package shared

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// rateLimiterEntry holds a rate limiter with its last access time for cleanup.
type rateLimiterEntry struct {
	limiter  *rate.Limiter
	lastUsed time.Time
}

// perClientRateLimiter manages per-client rate limiters with automatic cleanup.
type PerClientRateLimiter struct {
	limiters map[string]*rateLimiterEntry
	mu       sync.Mutex
}

// CleanupInterval is how often to run the cleanup goroutine.
const CleanupInterval = 10 * time.Minute

// StaleThreshold is the time after which an unused limiter is removed.
const StaleThreshold = 30 * time.Minute

// NewPerClientRateLimiter creates a new PerClientRateLimiter and starts background cleanup.
func NewPerClientRateLimiter() *PerClientRateLimiter {
	p := &PerClientRateLimiter{
		limiters: make(map[string]*rateLimiterEntry),
	}
	go p.startCleanup()
	return p
}

// startCleanup runs the background cleanup goroutine for stale rate limiters.
func (p *PerClientRateLimiter) startCleanup() {
	ticker := time.NewTicker(CleanupInterval)
	defer ticker.Stop()
	for range ticker.C {
		p.cleanup()
	}
}

// cleanup removes rate limiters that have not been used for longer than StaleThreshold.
func (p *PerClientRateLimiter) cleanup() {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	for id, entry := range p.limiters {
		if now.Sub(entry.lastUsed) > StaleThreshold {
			delete(p.limiters, id)
		}
	}
}

// GetLimiter returns or creates a rate limiter for a client.
func (p *PerClientRateLimiter) GetLimiter(clientID string) *rate.Limiter {
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

// GlobalRateLimitMiddleware creates a middleware that limits requests to 50 per second.
func GlobalRateLimitMiddleware() gin.HandlerFunc {
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

// PerClientRateLimitMiddleware creates a middleware that applies per-client rate limiting.
// The tokenExtractor function extracts the client identifier from the request.
func PerClientRateLimitMiddleware(limiter *PerClientRateLimiter, tokenExtractor func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := tokenExtractor(c)
		if token == "" {
			token = c.ClientIP()
		}

		rateLimiter := limiter.GetLimiter(token)
		if !rateLimiter.Allow() {
			c.JSON(429, gin.H{"error": "too many requests"})
			c.Abort()
			return
		}
		c.Next()
	}
}
