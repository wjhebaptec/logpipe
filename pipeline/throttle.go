package pipeline

import (
	"sync"
	"time"
)

// ThrottleConfig holds configuration for the Throttler.
type ThrottleConfig struct {
	// MaxPerWindow is the maximum number of entries allowed per window.
	MaxPerWindow int
	// Window is the duration of each throttle window.
	Window time.Duration
	// MatchLevel, if set, only throttles entries with this log level.
	MatchLevel string
}

// Throttler limits log entries to a maximum count per time window.
type Throttler struct {
	mu       sync.Mutex
	cfg      ThrottleConfig
	count    int
	windowAt time.Time
}

// NewThrottler creates a new Throttler with the given config.
// Defaults: MaxPerWindow=100, Window=1s.
func NewThrottler(cfg ThrottleConfig) *Throttler {
	if cfg.MaxPerWindow <= 0 {
		cfg.MaxPerWindow = 100
	}
	if cfg.Window <= 0 {
		cfg.Window = time.Second
	}
	return &Throttler{
		cfg:      cfg,
		windowAt: time.Now(),
	}
}

// Allow returns true if the entry should be forwarded, false if throttled.
func (t *Throttler) Allow(entry LogEntry) bool {
	if t.cfg.MatchLevel != "" && normalizeLevel(entry.Level) != normalizeLevel(t.cfg.MatchLevel) {
		return true
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	if now.Sub(t.windowAt) >= t.cfg.Window {
		t.count = 0
		t.windowAt = now
	}

	if t.count >= t.cfg.MaxPerWindow {
		return false
	}
	t.count++
	return true
}

// Remaining returns the number of entries still allowed in the current window.
func (t *Throttler) Remaining() int {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	if now.Sub(t.windowAt) >= t.cfg.Window {
		return t.cfg.MaxPerWindow
	}
	remaining := t.cfg.MaxPerWindow - t.count
	if remaining < 0 {
		return 0
	}
	return remaining
}
