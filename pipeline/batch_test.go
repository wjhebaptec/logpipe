package pipeline

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestBatch_FlushOnFullBatch(t *testing.T) {
	var mu sync.Mutex
	var received []LogEntry

	bp := NewBatchProcessor(3, 10*time.Second, func(entries []LogEntry) {
		mu.Lock()
		defer mu.Unlock()
		received = append(received, entries...)
	})

	bp.Add(LogEntry{Message: "a"})
	bp.Add(LogEntry{Message: "b"})
	bp.Add(LogEntry{Message: "c"}) // should trigger flush

	time.Sleep(50 * time.Millisecond)
	mu.Lock()
	defer mu.Unlock()
	if len(received) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(received))
	}
}

func TestBatch_FlushOnInterval(t *testing.T) {
	var mu sync.Mutex
	var received []LogEntry

	bp := NewBatchProcessor(100, 50*time.Millisecond, func(entries []LogEntry) {
		mu.Lock()
		defer mu.Unlock()
		received = append(received, entries...)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	go bp.Run(ctx)

	bp.Add(LogEntry{Message: "x"})
	bp.Add(LogEntry{Message: "y"})

	<-ctx.Done()
	time.Sleep(30 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(received) < 2 {
		t.Fatalf("expected at least 2 entries after interval flush, got %d", len(received))
	}
}

func TestBatch_FlushOnContextCancel(t *testing.T) {
	var mu sync.Mutex
	var received []LogEntry

	bp := NewBatchProcessor(100, 10*time.Second, func(entries []LogEntry) {
		mu.Lock()
		defer mu.Unlock()
		received = append(received, entries...)
	})

	ctx, cancel := context.WithCancel(context.Background())
	go bp.Run(ctx)

	bp.Add(LogEntry{Message: "flush-me"})
	cancel()
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(received) != 1 {
		t.Fatalf("expected 1 entry on cancel flush, got %d", len(received))
	}
}

func TestBatch_DefaultsForInvalidConfig(t *testing.T) {
	bp := NewBatchProcessor(0, 0, func([]LogEntry) {})
	if bp.size != 1 {
		t.Errorf("expected size=1, got %d", bp.size)
	}
	if bp.interval != time.Second {
		t.Errorf("expected interval=1s, got %v", bp.interval)
	}
}
