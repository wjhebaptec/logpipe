package pipeline

import (
	"sync"
	"testing"
)

func makeLogEntry(msg, level string) LogEntry {
	return LogEntry{Level: level, Message: msg, Fields: map[string]string{}}
}

func TestRingBuffer_DefaultCapacity(t *testing.T) {
	rb := NewRingBuffer(0)
	if rb.Cap() != 100 {
		t.Fatalf("expected default capacity 100, got %d", rb.Cap())
	}
}

func TestRingBuffer_PushAndLen(t *testing.T) {
	rb := NewRingBuffer(5)
	rb.Push(makeLogEntry("a", "info"))
	rb.Push(makeLogEntry("b", "warn"))
	if rb.Len() != 2 {
		t.Fatalf("expected len 2, got %d", rb.Len())
	}
}

func TestRingBuffer_DrainOrder(t *testing.T) {
	rb := NewRingBuffer(5)
	messages := []string{"first", "second", "third"}
	for _, m := range messages {
		rb.Push(makeLogEntry(m, "info"))
	}
	entries := rb.Drain()
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	for i, e := range entries {
		if e.Message != messages[i] {
			t.Errorf("entry %d: expected %q, got %q", i, messages[i], e.Message)
		}
	}
}

func TestRingBuffer_DrainClearsBuffer(t *testing.T) {
	rb := NewRingBuffer(5)
	rb.Push(makeLogEntry("x", "error"))
	rb.Drain()
	if rb.Len() != 0 {
		t.Fatalf("expected empty buffer after drain, got %d", rb.Len())
	}
}

func TestRingBuffer_Overwrites_WhenFull(t *testing.T) {
	rb := NewRingBuffer(3)
	rb.Push(makeLogEntry("a", "info"))
	rb.Push(makeLogEntry("b", "info"))
	rb.Push(makeLogEntry("c", "info"))
	rb.Push(makeLogEntry("d", "info")) // overwrites "a"

	if rb.Len() != 3 {
		t.Fatalf("expected len 3, got %d", rb.Len())
	}
	entries := rb.Drain()
	expected := []string{"b", "c", "d"}
	for i, e := range entries {
		if e.Message != expected[i] {
			t.Errorf("entry %d: expected %q, got %q", i, expected[i], e.Message)
		}
	}
}

func TestRingBuffer_Concurrent(t *testing.T) {
	rb := NewRingBuffer(50)
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			rb.Push(makeLogEntry("msg", "info"))
		}(i)
	}
	wg.Wait()
	if rb.Len() > 50 {
		t.Errorf("buffer exceeded capacity: len=%d", rb.Len())
	}
}
