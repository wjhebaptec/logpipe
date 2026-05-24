package pipeline

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// DeadLetterEntry wraps a failed log entry with metadata about the failure.
type DeadLetterEntry struct {
	Original  LogEntry  `json:"original"`
	Reason    string    `json:"reason"`
	FailedAt  time.Time `json:"failed_at"`
	Attempts  int       `json:"attempts"`
}

// DeadLetterQueue stores entries that could not be processed.
type DeadLetterQueue struct {
	mu      sync.Mutex
	entries []DeadLetterEntry
	maxSize int
	path    string
}

// NewDeadLetterQueue creates a DeadLetterQueue that persists to path.
// If path is empty, entries are kept in memory only.
func NewDeadLetterQueue(maxSize int, path string) *DeadLetterQueue {
	if maxSize <= 0 {
		maxSize = 100
	}
	return &DeadLetterQueue{
		maxSize: maxSize,
		path:    path,
	}
}

// Push adds a failed entry to the dead letter queue.
func (d *DeadLetterQueue) Push(entry LogEntry, reason string, attempts int) {
	d.mu.Lock()
	defer d.mu.Unlock()

	dle := DeadLetterEntry{
		Original: entry,
		Reason:   reason,
		FailedAt: time.Now(),
		Attempts: attempts,
	}

	if len(d.entries) >= d.maxSize {
		// Drop oldest entry to make room
		d.entries = d.entries[1:]
	}
	d.entries = append(d.entries, dle)

	if d.path != "" {
		_ = d.appendToFile(dle)
	}
}

// Drain returns all queued entries and clears the queue.
func (d *DeadLetterQueue) Drain() []DeadLetterEntry {
	d.mu.Lock()
	defer d.mu.Unlock()
	out := make([]DeadLetterEntry, len(d.entries))
	copy(out, d.entries)
	d.entries = d.entries[:0]
	return out
}

// Len returns the current number of entries in the queue.
func (d *DeadLetterQueue) Len() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.entries)
}

func (d *DeadLetterQueue) appendToFile(dle DeadLetterEntry) error {
	f, err := os.OpenFile(d.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("dead letter queue: open file: %w", err)
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(dle)
}
