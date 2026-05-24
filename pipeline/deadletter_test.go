package pipeline

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func dlEntry(msg, level string) LogEntry {
	return LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   msg,
		Fields:    map[string]interface{}{},
	}
}

func TestDeadLetter_PushAndLen(t *testing.T) {
	q := NewDeadLetterQueue(10, "")
	q.Push(dlEntry("fail", "error"), "timeout", 3)
	if q.Len() != 1 {
		t.Fatalf("expected 1, got %d", q.Len())
	}
}

func TestDeadLetter_DrainClearsQueue(t *testing.T) {
	q := NewDeadLetterQueue(10, "")
	q.Push(dlEntry("a", "error"), "network", 1)
	q.Push(dlEntry("b", "error"), "timeout", 2)

	out := q.Drain()
	if len(out) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(out))
	}
	if q.Len() != 0 {
		t.Fatalf("expected queue empty after drain")
	}
}

func TestDeadLetter_MaxSizeDropsOldest(t *testing.T) {
	q := NewDeadLetterQueue(3, "")
	for i := 0; i < 5; i++ {
		q.Push(dlEntry("msg", "warn"), "err", i)
	}
	if q.Len() != 3 {
		t.Fatalf("expected 3 entries (max), got %d", q.Len())
	}
	// Oldest should have been dropped; newest attempts should be 2,3,4
	out := q.Drain()
	if out[0].Attempts != 2 {
		t.Errorf("expected oldest remaining attempts=2, got %d", out[0].Attempts)
	}
}

func TestDeadLetter_DefaultMaxSize(t *testing.T) {
	q := NewDeadLetterQueue(0, "")
	if q.maxSize != 100 {
		t.Errorf("expected default maxSize 100, got %d", q.maxSize)
	}
}

func TestDeadLetter_PersistsToFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dlq.jsonl")

	q := NewDeadLetterQueue(10, path)
	q.Push(dlEntry("persisted", "error"), "disk full", 1)

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected file to exist: %v", err)
	}

	var dle DeadLetterEntry
	if err := json.Unmarshal(data, &dle); err != nil {
		t.Fatalf("failed to parse dead letter entry: %v", err)
	}
	if dle.Original.Message != "persisted" {
		t.Errorf("unexpected message: %s", dle.Original.Message)
	}
	if dle.Reason != "disk full" {
		t.Errorf("unexpected reason: %s", dle.Reason)
	}
}

func TestDeadLetter_DrainPreservesOrder(t *testing.T) {
	q := NewDeadLetterQueue(10, "")
	msgs := []string{"first", "second", "third"}
	for _, m := range msgs {
		q.Push(dlEntry(m, "error"), "reason", 1)
	}
	out := q.Drain()
	for i, m := range msgs {
		if out[i].Original.Message != m {
			t.Errorf("index %d: expected %q got %q", i, m, out[i].Original.Message)
		}
	}
}
