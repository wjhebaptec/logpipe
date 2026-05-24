package pipeline

import (
	"strconv"
	"testing"
)

// TestSequencer_WithFilter verifies that only entries passing the filter
// receive a sequence stamp.
func TestSequencer_WithFilter(t *testing.T) {
	s := NewSequencer(SequenceConfig{})
	cfg := FilterConfig{Level: "error"}

	entries := []LogEntry{
		{Level: "info", Message: "info msg", Fields: map[string]string{}},
		{Level: "error", Message: "err msg", Fields: map[string]string{}},
		{Level: "warn", Message: "warn msg", Fields: map[string]string{}},
		{Level: "error", Message: "err2", Fields: map[string]string{}},
	}

	var stamped []LogEntry
	for _, e := range entries {
		if Filter(e, cfg) {
			stamped = append(stamped, s.Stamp(e))
		}
	}

	if len(stamped) != 2 {
		t.Fatalf("expected 2 stamped entries, got %d", len(stamped))
	}
	for i, e := range stamped {
		want := strconv.Itoa(i + 1)
		if e.Fields["seq"] != want {
			t.Errorf("entry %d: expected seq=%s, got %s", i, want, e.Fields["seq"])
		}
	}
}

// TestSequencer_WithTransform verifies sequencing then transformation
// preserves the sequence field.
func TestSequencer_WithTransform(t *testing.T) {
	s := NewSequencer(SequenceConfig{Field: "id"})
	tcfg := TransformConfig{
		AddFields: map[string]string{"env": "prod"},
	}

	entry := LogEntry{Level: "info", Message: "hello", Fields: map[string]string{}}

	sequenced := s.Stamp(entry)
	result := Transform(sequenced, tcfg)

	if result.Fields["id"] != "1" {
		t.Errorf("expected id=1, got %s", result.Fields["id"])
	}
	if result.Fields["env"] != "prod" {
		t.Errorf("expected env=prod, got %s", result.Fields["env"])
	}
}
