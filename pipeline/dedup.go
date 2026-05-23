package pipeline

import (
	"crypto/md5"
	"fmt"
	"sync"
	"time"
)

// DedupConfig holds configuration for deduplication.
type DedupConfig struct {
	Enabled  bool          `yaml:"enabled"`
	Window   time.Duration `yaml:"window"`
	Capacity int           `yaml:"capacity"`
}

// Deduplicator tracks recently seen log entries and suppresses duplicates.
type Deduplicator struct {
	mu       sync.Mutex
	seen     map[string]time.Time
	window   time.Duration
	capacity int
}

// NewDeduplicator creates a new Deduplicator with the given window and capacity.
func NewDeduplicator(window time.Duration, capacity int) *Deduplicator {
	if window == 0 {
		window = 5 * time.Second
	}
	if capacity <= 0 {
		capacity = 1000
	}
	return &Deduplicator{
		seen:     make(map[string]time.Time, capacity),
		window:   window,
		capacity: capacity,
	}
}

// IsDuplicate returns true if the entry was seen within the dedup window.
// It also records the entry if it is not a duplicate.
func (d *Deduplicator) IsDuplicate(entry LogEntry) bool {
	key := entryKey(entry)
	now := time.Now()

	d.mu.Lock()
	defer d.mu.Unlock()

	if t, ok := d.seen[key]; ok && now.Sub(t) < d.window {
		return true
	}

	// Evict oldest entries if at capacity.
	if len(d.seen) >= d.capacity {
		d.evictOldest(now)
	}

	d.seen[key] = now
	return false
}

// evictOldest removes entries outside the current window. Must be called with mu held.
func (d *Deduplicator) evictOldest(now time.Time) {
	for k, t := range d.seen {
		if now.Sub(t) >= d.window {
			delete(d.seen, k)
		}
	}
}

// entryKey produces a stable hash key for a log entry.
func entryKey(entry LogEntry) string {
	raw := fmt.Sprintf("%s|%s|%v", entry.Level, entry.Message, entry.Fields)
	sum := md5.Sum([]byte(raw))
	return fmt.Sprintf("%x", sum)
}
