// Package evasion implements rate limiting, IP rotation, and user-agent rotation
// to help scanning nodes avoid detection.
package evasion

import (
	"sync"
	"time"
)

// RateLimiter implements a token-bucket rate limiter.
type RateLimiter struct {
	rate       float64 // tokens per second
	burst      int     // maximum burst size
	tokens     float64
	lastUpdate time.Time
	mu         sync.Mutex
}

// NewRateLimiter creates a new RateLimiter with the given rate (tokens/sec) and burst.
func NewRateLimiter(rate int, burst int) *RateLimiter {
	if burst <= 0 {
		burst = rate
	}
	return &RateLimiter{
		rate:       float64(rate),
		burst:      burst,
		tokens:     float64(burst),
		lastUpdate: time.Now(),
	}
}

// Allow checks if a single token is available and consumes it if so.
// Returns true if the action is allowed, false otherwise.
func (rl *RateLimiter) Allow() bool {
	return rl.AllowN(1)
}

// AllowN checks if n tokens are available and consumes them if so.
func (rl *RateLimiter) AllowN(n int) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastUpdate).Seconds()
	rl.tokens += elapsed * rl.rate
	if rl.tokens > float64(rl.burst) {
		rl.tokens = float64(rl.burst)
	}
	rl.lastUpdate = now

	if rl.tokens >= float64(n) {
		rl.tokens -= float64(n)
		return true
	}
	return false
}

// Wait blocks until n tokens are available, then consumes them.
// Returns if the context is cancelled.
func (rl *RateLimiter) Wait(n int) {
	for !rl.AllowN(n) {
		time.Sleep(time.Duration(float64(time.Second) / rl.rate))
	}
}

// SetRate updates the token generation rate.
func (rl *RateLimiter) SetRate(rate int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.rate = float64(rate)
}
