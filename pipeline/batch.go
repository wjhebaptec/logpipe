package pipeline

import (
	"context"
	"sync"
	"time"
)

// BatchProcessor accumulates log entries and flushes them in batches
// either when the batch is full or a flush interval elapses.
type BatchProcessor struct {
	mu       sync.Mutex
	batch    []LogEntry
	size     int
	interval time.Duration
	flushFn  func([]LogEntry)
}

// LogEntry is a minimal representation used by the batch processor.
type LogEntry struct {
	Level   string
	Message string
	Fields  map[string]string
}

// NewBatchProcessor creates a BatchProcessor that calls flushFn when
// the batch reaches size entries or the interval elapses.
func NewBatchProcessor(size int, interval time.Duration, flushFn func([]LogEntry)) *BatchProcessor {
	if size <= 0 {
		size = 1
	}
	if interval <= 0 {
		interval = time.Second
	}
	return &BatchProcessor{
		size:     size,
		interval: interval,
		flushFn:  flushFn,
	}
}

// Add appends an entry to the current batch and flushes if full.
func (b *BatchProcessor) Add(entry LogEntry) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.batch = append(b.batch, entry)
	if len(b.batch) >= b.size {
		b.flush()
	}
}

// flush sends the current batch to flushFn and resets the buffer.
// Caller must hold b.mu.
func (b *BatchProcessor) flush() {
	if len(b.batch) == 0 {
		return
	}
	copy := make([]LogEntry, len(b.batch))
	_ = copy[:copy(copy, b.batch)]
	b.batch = b.batch[:0]
	go b.flushFn(copy)
}

// Run starts the background ticker that flushes on the interval.
// It blocks until ctx is cancelled.
func (b *BatchProcessor) Run(ctx context.Context) {
	ticker := time.NewTicker(b.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			b.mu.Lock()
			b.flush()
			b.mu.Unlock()
		case <-ctx.Done():
			b.mu.Lock()
			b.flush()
			b.mu.Unlock()
			return
		}
	}
}
