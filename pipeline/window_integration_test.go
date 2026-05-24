package pipeline

import (
	"testing"
	"time"
)

func TestWindow_WithFilter(t *testing.T) {
	w := NewSlidingWindow(WindowConfig{Size: time.Minute, MaxItems: 100})
	entries := []LogEntry{
		{Message: "disk full", Level: "error"},
		{Message: "started", Level: "info"},
		{Message: "timeout", Level: "error"},
	}
	for _, e := range entries {
		w.Add(e)
	}
	snap := w.Snapshot()
	var errors []WindowEntry
	for _, we := range snap {
		if normalizeLevel(we.Entry.Level) == "error" {
			errors = append(errors, we)
		}
	}
	if len(errors) != 2 {
		t.Errorf("expected 2 error entries, got %d", len(errors))
	}
}

func TestWindow_AggregatorAfterExpiry(t *testing.T) {
	w := NewSlidingWindow(WindowConfig{Size: 40 * time.Millisecond, MaxItems: 100})
	a := NewWindowAggregator(w)
	w.Add(LogEntry{Level: "warn", Message: "expiring"})
	w.Add(LogEntry{Level: "warn", Message: "expiring2"})
	time.Sleep(60 * time.Millisecond)
	w.Add(LogEntry{Level: "info", Message: "fresh"})
	stats := a.Stats()
	if stats.Total != 1 {
		t.Errorf("expected 1 after expiry, got %d", stats.Total)
	}
	if stats.ByLevel["info"] != 1 {
		t.Errorf("expected info=1, got %d", stats.ByLevel["info"])
	}
	if _, ok := stats.ByLevel["warn"]; ok {
		t.Errorf("expected warn entries to be evicted")
	}
}

func TestWindow_ConcurrentAdds(t *testing.T) {
	w := NewSlidingWindow(WindowConfig{Size: time.Minute, MaxItems: 500})
	done := make(chan struct{})
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 20; j++ {
				w.Add(LogEntry{Message: "concurrent", Level: "info"})
			}
			done <- struct{}{}
		}()
	}
	for i := 0; i < 10; i++ {
		<-done
	}
	if w.Len() > 200 {
		t.Errorf("expected at most 200 entries, got %d", w.Len())
	}
}
