package pipeline

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestRateLimiter_AllowsUpToBurst(t *testing.T) {
	rl := NewRateLimiter(5)
	allowed := 0
	for i := 0; i < 10; i++ {
		if rl.Allow() {
			allowed++
		}
	}
	if allowed != 5 {
		t.Errorf("expected 5 allowed (burst), got %d", allowed)
	}
}

func TestRateLimiter_RefillsOverTime(t *testing.T) {
	now := time.Now()
	rl := NewRateLimiter(10)
	rl.clock = func() time.Time { return now }

	// Drain all tokens
	for i := 0; i < 10; i++ {
		rl.Allow()
	}
	if rl.Allow() {
		t.Fatal("expected rate limiter to be exhausted")
	}

	// Advance clock by 0.5s — should refill 5 tokens
	now = now.Add(500 * time.Millisecond)
	allowed := 0
	for i := 0; i < 10; i++ {
		if rl.Allow() {
			allowed++
		}
	}
	if allowed != 5 {
		t.Errorf("expected 5 tokens after 0.5s refill at rate 10/s, got %d", allowed)
	}
}

func TestRateLimiter_ZeroRateDefaultsToOne(t *testing.T) {
	rl := NewRateLimiter(0)
	if !rl.Allow() {
		t.Error("expected at least one token for zero-rate limiter")
	}
	if rl.Allow() {
		t.Error("expected second allow to fail for rate=1")
	}
}

func TestRateLimiter_Remaining(t *testing.T) {
	rl := NewRateLimiter(4)
	rl.Allow()
	rl.Allow()
	rem := rl.Remaining()
	if rem != 2.0 {
		t.Errorf("expected 2 remaining tokens, got %f", rem)
	}
}

func TestRateLimiter_Concurrent(t *testing.T) {
	rl := NewRateLimiter(100)
	var allowed atomic.Int64
	var wg sync.WaitGroup
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if rl.Allow() {
				allowed.Add(1)
			}
		}()
	}
	wg.Wait()
	if allowed.Load() > 100 {
		t.Errorf("expected at most 100 allowed, got %d", allowed.Load())
	}
}
