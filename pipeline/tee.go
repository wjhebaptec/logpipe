package pipeline

import "sync"

// TeeConfig holds configuration for the Tee processor.
type TeeConfig struct {
	// BufferSize is the channel buffer size for each branch. Defaults to 64.
	BufferSize int
}

// Tee fans out a single stream of LogEntry values to multiple independent
// consumer channels. Each consumer receives every entry; slow consumers do
// not block other consumers because each branch has its own buffered channel.
type Tee struct {
	mu       sync.Mutex
	branches []chan LogEntry
	bufSize  int
}

// NewTee creates a Tee with the given configuration.
func NewTee(cfg TeeConfig) *Tee {
	bs := cfg.BufferSize
	if bs <= 0 {
		bs = 64
	}
	return &Tee{bufSize: bs}
}

// Branch creates a new output channel and registers it with the Tee.
// The caller is responsible for draining the returned channel.
func (t *Tee) Branch() <-chan LogEntry {
	ch := make(chan LogEntry, t.bufSize)
	t.mu.Lock()
	t.branches = append(t.branches, ch)
	t.mu.Unlock()
	return ch
}

// Send forwards entry to all registered branches.
// Entries are dropped on a branch whose buffer is full to avoid blocking.
func (t *Tee) Send(entry LogEntry) {
	t.mu.Lock()
	bs := make([]chan LogEntry, len(t.branches))
	copy(bs, t.branches)
	t.mu.Unlock()

	for _, ch := range bs {
		select {
		case ch <- entry:
		default:
			// branch buffer full — drop to avoid blocking
		}
	}
}

// Close closes all registered branch channels.
func (t *Tee) Close() {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, ch := range t.branches {
		close(ch)
	}
	t.branches = nil
}
