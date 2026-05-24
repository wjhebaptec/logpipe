package pipeline

import (
	"testing"
	"time"
)

// TestDeadLetter_WithRetryer verifies that exhausted retries land in the DLQ.
func TestDeadLetter_WithRetryer(t *testing.T) {
	dlq := NewDeadLetterQueue(10, "")

	attempts := 0
	r := NewRetryer(RetryConfig{MaxAttempts: 3, InitialDelay: time.Millisecond})

	err := r.Run(func() error {
		attempts++
		return &testRetryError{}
	})

	if err == nil {
		t.Fatal("expected error after exhausted retries")
	}

	entry := dlEntry("failed message", "error")
	dlq.Push(entry, err.Error(), attempts)

	if dlq.Len() != 1 {
		t.Fatalf("expected 1 dead letter entry, got %d", dlq.Len())
	}

	out := dlq.Drain()
	if out[0].Attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", out[0].Attempts)
	}
}

// TestDeadLetter_WithFilter verifies entries can be filtered before DLQ storage.
func TestDeadLetter_WithFilter(t *testing.T) {
	dlq := NewDeadLetterQueue(10, "")

	filterCfg := FilterConfig{Level: "error"}

	entries := []LogEntry{
		{Timestamp: time.Now(), Level: "info", Message: "ok", Fields: map[string]interface{}{}},
		{Timestamp: time.Now(), Level: "error", Message: "fail", Fields: map[string]interface{}{}},
		{Timestamp: time.Now(), Level: "warn", Message: "skip", Fields: map[string]interface{}{}},
	}

	for _, e := range entries {
		if Filter(e, filterCfg) {
			dlq.Push(e, "filtered error", 1)
		}
	}

	if dlq.Len() != 1 {
		t.Fatalf("expected 1 entry in DLQ, got %d", dlq.Len())
	}

	out := dlq.Drain()
	if out[0].Original.Message != "fail" {
		t.Errorf("expected message 'fail', got %q", out[0].Original.Message)
	}
}
