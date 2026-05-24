package pipeline

import (
	"testing"
	"time"
)

func TestWindow_DefaultConfig(t *testing.T) {
	w := NewSlidingWindow(WindowConfig{})
	if w.size != 60*time.Second {
		t.Errorf("expected default size 60s, got %v", w.size)
	}
	if w.maxItems != 1000 {
		t.Errorf("expected default maxItems 1000, got %d", w.maxItems)
	}
}

func TestWindow_AddAndLen(t *testing.T) {
	w := NewSlidingWindow(WindowConfig{Size: time.Minute, MaxItems: 10})
	w.Add(LogEntry{Message: "a"})
	w.Add(LogEntry{Message: "b"})
	if w.Len() != 2 {
		t.Errorf("expected 2, got %d", w.Len())
	}
}

func TestWindow_SnapshotContents(t *testing.T) {
	w := NewSlidingWindow(WindowConfig{Size: time.Minute, MaxItems: 10})
	w.Add(LogEntry{Message: "hello"})
	snap := w.Snapshot()
	if len(snap) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(snap))
	}
	if snap[0].Entry.Message != "hello" {
		t.Errorf("unexpected message: %s", snap[0].Entry.Message)
	}
}

func TestWindow_MaxItemsDropsOldest(t *testing.T) {
	w := NewSlidingWindow(WindowConfig{Size: time.Minute, MaxItems: 3})
	for i := 0; i < 5; i++ {
		w.Add(LogEntry{Message: "x"})
	}
	if w.Len() != 3 {
		t.Errorf("expected 3 after overflow, got %d", w.Len())
	}
}

func TestWindow_ExpiredEntriesEvicted(t *testing.T) {
	w := NewSlidingWindow(WindowConfig{Size: 50 * time.Millisecond, MaxItems: 100})
	w.Add(LogEntry{Message: "old"})
	time.Sleep(80 * time.Millisecond)
	w.Add(LogEntry{Message: "new"})
	if w.Len() != 1 {
		t.Errorf("expected 1 after expiry, got %d", w.Len())
	}
	snap := w.Snapshot()
	if snap[0].Entry.Message != "new" {
		t.Errorf("expected 'new', got %s", snap[0].Entry.Message)
	}
}

func TestWindow_EmptySnapshot(t *testing.T) {
	w := NewSlidingWindow(WindowConfig{Size: time.Minute, MaxItems: 10})
	snap := w.Snapshot()
	if len(snap) != 0 {
		t.Errorf("expected empty snapshot, got %d entries", len(snap))
	}
}
