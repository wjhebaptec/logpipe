package pipeline

import (
	"sync"
)

// RingBuffer is a fixed-capacity circular buffer for log entries.
// When full, the oldest entry is overwritten.
type RingBuffer struct {
	mu       sync.Mutex
	entries  []LogEntry
	capacity int
	head     int
	count    int
}

// LogEntry represents a single structured log entry stored in the buffer.
type LogEntry struct {
	Level   string
	Message string
	Fields  map[string]string
}

// NewRingBuffer creates a RingBuffer with the given capacity.
// If capacity <= 0, it defaults to 100.
func NewRingBuffer(capacity int) *RingBuffer {
	if capacity <= 0 {
		capacity = 100
	}
	return &RingBuffer{
		entries:  make([]LogEntry, capacity),
		capacity: capacity,
	}
}

// Push adds an entry to the buffer. If the buffer is full, the oldest entry
// is overwritten.
func (rb *RingBuffer) Push(entry LogEntry) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	index := (rb.head + rb.count) % rb.capacity
	rb.entries[index] = entry

	if rb.count < rb.capacity {
		rb.count++
	} else {
		// Overwrite oldest: advance head
		rb.head = (rb.head + 1) % rb.capacity
	}
}

// Drain returns all buffered entries in insertion order and clears the buffer.
func (rb *RingBuffer) Drain() []LogEntry {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	result := make([]LogEntry, rb.count)
	for i := 0; i < rb.count; i++ {
		result[i] = rb.entries[(rb.head+i)%rb.capacity]
	}
	rb.head = 0
	rb.count = 0
	return result
}

// Len returns the number of entries currently in the buffer.
func (rb *RingBuffer) Len() int {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	return rb.count
}

// Cap returns the maximum capacity of the buffer.
func (rb *RingBuffer) Cap() int {
	return rb.capacity
}
