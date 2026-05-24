package pipeline

import (
	"sync"
)

// MulticastConfig controls how entries are broadcast to subscribers.
type MulticastConfig struct {
	// BufferSize is the per-subscriber channel capacity. Defaults to 64.
	BufferSize int
}

// Multicast broadcasts each LogEntry to all registered subscriber channels.
type Multicast struct {
	mu          sync.RWMutex
	subscribers []chan LogEntry
	bufferSize  int
}

// NewMulticast creates a Multicast broadcaster with the given config.
func NewMulticast(cfg MulticastConfig) *Multicast {
	bs := cfg.BufferSize
	if bs <= 0 {
		bs = 64
	}
	return &Multicast{bufferSize: bs}
}

// Subscribe registers a new subscriber and returns its receive channel.
// The caller is responsible for draining the channel.
func (m *Multicast) Subscribe() <-chan LogEntry {
	ch := make(chan LogEntry, m.bufferSize)
	m.mu.Lock()
	m.subscribers = append(m.subscribers, ch)
	m.mu.Unlock()
	return ch
}

// Unsubscribe removes a previously subscribed channel and closes it.
func (m *Multicast) Unsubscribe(sub <-chan LogEntry) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, ch := range m.subscribers {
		if ch == sub {
			m.subscribers = append(m.subscribers[:i], m.subscribers[i+1:]...)
			close(ch)
			return
		}
	}
}

// Publish sends entry to every subscriber. Slow subscribers that have full
// buffers are skipped (non-blocking) so they do not stall the pipeline.
func (m *Multicast) Publish(entry LogEntry) (delivered, dropped int) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, ch := range m.subscribers {
		select {
		case ch <- entry:
			delivered++
		default:
			dropped++
		}
	}
	return
}

// Len returns the current number of active subscribers.
func (m *Multicast) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.subscribers)
}

// Close closes all subscriber channels and removes them.
func (m *Multicast) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, ch := range m.subscribers {
		close(ch)
	}
	m.subscribers = nil
}
