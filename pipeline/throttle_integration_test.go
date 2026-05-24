package pipeline

import (
	"testing"
	"time"
)

// TestThrottle_WithFilter verifies that Throttler and Filter work together:
// only entries that pass the filter are subject to throttling.
func TestThrottle_WithFilter(t *testing.T) {
	th := NewThrottler(ThrottleConfig{MaxPerWindow: 2, Window: time.Second})

	cfg := FilterConfig{
		Level:    "error",
		Contains: "",
	}

	entries := []LogEntry{
		{Level: "info", Message: "info 1"},
		{Level: "error", Message: "err 1"},
		{Level: "error", Message: "err 2"},
		{Level: "error", Message: "err 3"},
		{Level: "info", Message: "info 2"},
	}

	allowed := 0
	for _, e := range entries {
		if !Filter(e, cfg) {
			continue
		}
		if th.Allow(e) {
			allowed++
		}
	}

	// Only 2 error entries should be allowed through throttle (max=2)
	if allowed != 2 {
		t.Fatalf("expected 2 allowed error entries, got %d", allowed)
	}
}

// TestThrottle_WithSampler verifies Throttler and Sampler compose correctly.
func TestThrottle_WithSampler(t *testing.T) {
	th := NewThrottler(ThrottleConfig{MaxPerWindow: 50, Window: time.Second})
	sampler := NewSampler(1.0) // keep all

	passed := 0
	for i := 0; i < 100; i++ {
		e := LogEntry{Level: "info", Message: "msg"}
		if sampler.Sample(e) && th.Allow(e) {
			passed++
		}
	}

	if passed != 50 {
		t.Fatalf("expected 50 entries to pass throttle, got %d", passed)
	}
}
