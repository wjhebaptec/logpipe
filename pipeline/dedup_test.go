package pipeline

import (
	"testing"
	"time"
)

func makeEntry(level, msg string) LogEntry {
	return LogEntry{Level: level, Message: msg, Fields: map[string]interface{}{}}
}

func TestDedup_FirstEntryNotDuplicate(t *testing.T) {
	d := NewDeduplicator(5*time.Second, 100)
	entry := makeEntry("info", "hello world")
	if d.IsDuplicate(entry) {
		t.Error("first occurrence should not be a duplicate")
	}
}

func TestDedup_SecondEntryIsDuplicate(t *testing.T) {
	d := NewDeduplicator(5*time.Second, 100)
	entry := makeEntry("info", "hello world")
	d.IsDuplicate(entry)
	if !d.IsDuplicate(entry) {
		t.Error("second occurrence within window should be a duplicate")
	}
}

func TestDedup_DifferentEntriesNotDuplicate(t *testing.T) {
	d := NewDeduplicator(5*time.Second, 100)
	a := makeEntry("info", "message A")
	b := makeEntry("info", "message B")
	d.IsDuplicate(a)
	if d.IsDuplicate(b) {
		t.Error("different message should not be a duplicate")
	}
}

func TestDedup_ExpiredWindowNotDuplicate(t *testing.T) {
	d := NewDeduplicator(50*time.Millisecond, 100)
	entry := makeEntry("warn", "transient error")
	d.IsDuplicate(entry)
	time.Sleep(80 * time.Millisecond)
	if d.IsDuplicate(entry) {
		t.Error("entry after window expiry should not be a duplicate")
	}
}

func TestDedup_CapacityEviction(t *testing.T) {
	d := NewDeduplicator(10*time.Second, 5)
	for i := 0; i < 5; i++ {
		d.IsDuplicate(makeEntry("info", time.Now().String()))
		time.Sleep(1 * time.Millisecond)
	}
	// Adding one more should trigger eviction without panic.
	d.IsDuplicate(makeEntry("info", "overflow entry"))
	if len(d.seen) > 6 {
		t.Errorf("expected seen map to stay near capacity, got %d", len(d.seen))
	}
}

func TestDedup_DefaultsApplied(t *testing.T) {
	d := NewDeduplicator(0, 0)
	if d.window != 5*time.Second {
		t.Errorf("expected default window 5s, got %v", d.window)
	}
	if d.capacity != 1000 {
		t.Errorf("expected default capacity 1000, got %d", d.capacity)
	}
}
