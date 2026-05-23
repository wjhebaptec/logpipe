package pipeline

import (
	"sync"
	"time"
)

// RateLimiter limits the number of log entries forwarded per second
// using a token bucket algorithm.
type RateLimiter struct {
	mu       sync.Mutex
	tokens   float64
	max      float64
	rate     float64 // tokens per second
	lastTick time.Time
	clock    func() time.Time
}

// NewRateLimiter creates a RateLimiter that allows up to ratePerSec entries
// per second with a burst capacity equal to ratePerSec.
func NewRateLimiter(ratePerSec float64) *RateLimiter {
	if ratePerSec <= 0 {
		ratePerSec = 1
	}
	return &RateLimiter{
		tokens:   ratePerSec,
		max:      ratePerSec,
		rate:     ratePerSec,
		lastTick: time.Now(),
		clock:    time.Now,
	}
}

// Allow returns true if the entry should be forwarded, false if it should be
// dropped due to rate limiting.
func (r *RateLimiter) Allow() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := r.clock()
	elapsed := now.Sub(r.lastTick).Seconds()
	r.lastTick = now

	r.tokens += elapsed * r.rate
	if r.tokens > r.max {
		r.tokens = r.max
	}

	if r.tokens >= 1.0 {
		r.tokens -= 1.0
		return true
	}
	return false
}

// Remaining returns the current number of available tokens (approximate).
func (r *RateLimiter) Remaining() float64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.tokens
}
