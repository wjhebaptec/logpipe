package pipeline

import (
	"testing"
	"time"
)

func TestWindowAggregator_EmptyStats(t *testing.T) {
	w := NewSlidingWindow(WindowConfig{Size: time.Minute, MaxItems: 100})
	a := NewWindowAggregator(w)
	stats := a.Stats()
	if stats.Total != 0 {
		t.Errorf("expected 0 total, got %d", stats.Total)
	}
	if len(stats.ByLevel) != 0 {
		t.Errorf("expected empty ByLevel")
	}
}

func TestWindowAggregator_TotalCount(t *testing.T) {
	w := NewSlidingWindow(WindowConfig{Size: time.Minute, MaxItems: 100})
	a := NewWindowAggregator(w)
	w.Add(LogEntry{Message: "a", Level: "info"})
	w.Add(LogEntry{Message: "b", Level: "error"})
	w.Add(LogEntry{Message: "c", Level: "info"})
	stats := a.Stats()
	if stats.Total != 3 {
		t.Errorf("expected total 3, got %d", stats.Total)
	}
}

func TestWindowAggregator_ByLevel(t *testing.T) {
	w := NewSlidingWindow(WindowConfig{Size: time.Minute, MaxItems: 100})
	a := NewWindowAggregator(w)
	w.Add(LogEntry{Level: "INFO"})
	w.Add(LogEntry{Level: "info"})
	w.Add(LogEntry{Level: "ERROR"})
	stats := a.Stats()
	if stats.ByLevel["info"] != 2 {
		t.Errorf("expected info=2, got %d", stats.ByLevel["info"])
	}
	if stats.ByLevel["error"] != 1 {
		t.Errorf("expected error=1, got %d", stats.ByLevel["error"])
	}
}

func TestWindowAggregator_OldestNewest(t *testing.T) {
	w := NewSlidingWindow(WindowConfig{Size: time.Minute, MaxItems: 100})
	a := NewWindowAggregator(w)
	w.Add(LogEntry{Message: "first"})
	time.Sleep(10 * time.Millisecond)
	w.Add(LogEntry{Message: "last"})
	stats := a.Stats()
	if !stats.Oldest.Before(stats.Newest) {
		t.Errorf("expected oldest before newest")
	}
}

func TestWindowAggregator_TopMessages(t *testing.T) {
	w := NewSlidingWindow(WindowConfig{Size: time.Minute, MaxItems: 100})
	a := NewWindowAggregator(w)
	for _, msg := range []string{"a", "b", "c", "d", "e"} {
		w.Add(LogEntry{Message: msg})
	}
	top := a.TopMessages(3)
	if len(top) != 3 {
		t.Fatalf("expected 3, got %d", len(top))
	}
	if top[0] != "c" || top[1] != "d" || top[2] != "e" {
		t.Errorf("unexpected top messages: %v", top)
	}
}

func TestWindowAggregator_TopMessages_FewerThanN(t *testing.T) {
	w := NewSlidingWindow(WindowConfig{Size: time.Minute, MaxItems: 100})
	a := NewWindowAggregator(w)
	w.Add(LogEntry{Message: "only"})
	top := a.TopMessages(5)
	if len(top) != 1 {
		t.Errorf("expected 1, got %d", len(top))
	}
}
