package pipeline

import (
	"sync"
	"time"
)

// WindowConfig holds configuration for a sliding window aggregator.
type WindowConfig struct {
	Size     time.Duration
	MaxItems int
}

// WindowEntry holds a log entry with its arrival timestamp.
type WindowEntry struct {
	Entry     LogEntry
	ReceivedAt time.Time
}

// SlidingWindow aggregates log entries within a time window.
type SlidingWindow struct {
	mu      sync.Mutex
	size    time.Duration
	maxItems int
	items   []WindowEntry
}

// NewSlidingWindow creates a SlidingWindow with the given config.
// Defaults: Size=60s, MaxItems=1000.
func NewSlidingWindow(cfg WindowConfig) *SlidingWindow {
	if cfg.Size <= 0 {
		cfg.Size = 60 * time.Second
	}
	if cfg.MaxItems <= 0 {
		cfg.MaxItems = 1000
	}
	return &SlidingWindow{
		size:     cfg.Size,
		maxItems: cfg.MaxItems,
	}
}

// Add inserts a log entry into the window, evicting expired entries first.
func (w *SlidingWindow) Add(entry LogEntry) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.evict()
	if len(w.items) >= w.maxItems {
		w.items = w.items[1:]
	}
	w.items = append(w.items, WindowEntry{Entry: entry, ReceivedAt: time.Now()})
}

// Snapshot returns a copy of all non-expired entries in the window.
func (w *SlidingWindow) Snapshot() []WindowEntry {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.evict()
	result := make([]WindowEntry, len(w.items))
	copy(result, w.items)
	return result
}

// Len returns the number of non-expired entries currently in the window.
func (w *SlidingWindow) Len() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.evict()
	return len(w.items)
}

// evict removes entries older than the window size. Must be called with mu held.
func (w *SlidingWindow) evict() {
	cutoff := time.Now().Add(-w.size)
	i := 0
	for i < len(w.items) && w.items[i].ReceivedAt.Before(cutoff) {
		i++
	}
	w.items = w.items[i:]
}
