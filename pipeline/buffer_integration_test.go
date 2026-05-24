package pipeline

import (
	"testing"
)

// TestRingBuffer_WithFilter verifies that entries pushed into the ring buffer
// can be drained and filtered using the existing Filter logic.
func TestRingBuffer_WithFilter(t *testing.T) {
	rb := NewRingBuffer(10)

	rb.Push(LogEntry{Level: "error", Message: "disk full", Fields: map[string]string{}})
	rb.Push(LogEntry{Level: "info", Message: "started", Fields: map[string]string{}})
	rb.Push(LogEntry{Level: "error", Message: "connection refused", Fields: map[string]string{}})
	rb.Push(LogEntry{Level: "debug", Message: "heartbeat", Fields: map[string]string{}})

	all := rb.Drain()
	if len(all) != 4 {
		t.Fatalf("expected 4 entries, got %d", len(all))
	}

	var errors []LogEntry
	for _, e := range all {
		if normalizeLevel(e.Level) == "error" {
			errors = append(errors, e)
		}
	}
	if len(errors) != 2 {
		t.Errorf("expected 2 error entries, got %d", len(errors))
	}
}

// TestRingBuffer_OverwritePreservesLatest ensures that after overflow the most
// recent entries are retained, which is the desired behaviour for a tail buffer.
func TestRingBuffer_OverwritePreservesLatest(t *testing.T) {
	rb := NewRingBuffer(3)
	msgs := []string{"old1", "old2", "keep1", "keep2", "keep3"}
	for _, m := range msgs {
		rb.Push(LogEntry{Level: "info", Message: m})
	}

	entries := rb.Drain()
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries after overflow, got %d", len(entries))
	}

	expected := []string{"keep1", "keep2", "keep3"}
	for i, e := range entries {
		if e.Message != expected[i] {
			t.Errorf("pos %d: want %q got %q", i, expected[i], e.Message)
		}
	}
}
