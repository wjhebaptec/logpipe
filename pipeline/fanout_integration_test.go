package pipeline

import (
	"context"
	"sync/atomic"
	"testing"
)

// TestFanout_WithFilter verifies that Fanout dispatches only entries that pass
// through a Filter stage placed upstream.
func TestFanout_WithFilter(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var dispatched int64
	f := NewFanout(ctx, FanoutConfig{Workers: 2, BufferSize: 16}, func(e LogEntry) {
		atomic.AddInt64(&dispatched, 1)
	})

	entries := []LogEntry{
		{Level: "error", Message: "disk full"},
		{Level: "info", Message: "started"},
		{Level: "error", Message: "oom"},
		{Level: "debug", Message: "verbose"},
	}

	cfg := FilterConfig{Level: "error"}
	for _, e := range entries {
		if out, ok := Filter(e, cfg); ok {
			f.Send(out)
		}
	}

	cancel()
	f.Wait()

	if got := atomic.LoadInt64(&dispatched); got != 2 {
		t.Fatalf("expected 2 dispatched entries after filter, got %d", got)
	}
}

// TestFanout_WithSampler verifies Fanout works correctly when entries are
// pre-sampled before being sent.
func TestFanout_WithSampler(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var dispatched int64
	f := NewFanout(ctx, FanoutConfig{Workers: 2, BufferSize: 64}, func(e LogEntry) {
		atomic.AddInt64(&dispatched, 1)
	})

	// rate=1.0 keeps everything
	s := NewSampler(SamplerConfig{Rate: 1.0})

	for i := 0; i < 30; i++ {
		e := LogEntry{Level: "info", Message: "ping"}
		if out, ok := s.Sample(e); ok {
			f.Send(out)
		}
	}

	cancel()
	f.Wait()

	if got := atomic.LoadInt64(&dispatched); got != 30 {
		t.Fatalf("expected 30 dispatched entries, got %d", got)
	}
}
