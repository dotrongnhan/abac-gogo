package pep

import (
	"sync"
	"time"
)

// RateLimiter implements token bucket rate limiting
type RateLimiter struct {
	rate       float64   // tokens per second
	capacity   int       // bucket capacity
	tokens     float64   // current tokens
	lastUpdate time.Time // last token update time
	mu         sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerSecond int, burstSize int) *RateLimiter {
	return &RateLimiter{
		rate:       float64(requestsPerSecond),
		capacity:   burstSize,
		tokens:     float64(burstSize), // Start with full bucket
		lastUpdate: time.Now(),
	}
}

// Allow checks if a request is allowed under the rate limit
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastUpdate).Seconds()

	// Add tokens based on elapsed time
	rl.tokens += elapsed * rl.rate
	if rl.tokens > float64(rl.capacity) {
		rl.tokens = float64(rl.capacity)
	}

	rl.lastUpdate = now

	// Check if we have tokens available
	if rl.tokens >= 1.0 {
		rl.tokens--
		return true
	}

	return false
}

// GetTokens returns the current number of available tokens
func (rl *RateLimiter) GetTokens() float64 {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastUpdate).Seconds()

	// Update tokens without consuming
	rl.tokens += elapsed * rl.rate
	if rl.tokens > float64(rl.capacity) {
		rl.tokens = float64(rl.capacity)
	}

	rl.lastUpdate = now
	return rl.tokens
}

// Reset resets the rate limiter to full capacity
func (rl *RateLimiter) Reset() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.tokens = float64(rl.capacity)
	rl.lastUpdate = time.Now()
}
