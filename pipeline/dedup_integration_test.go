package pipeline

import (
	"strings"
	"testing"
	"time"
)

// TestDedup_WithFilter verifies deduplication works alongside the Filter function.
func TestDedup_WithFilter(t *testing.T) {
	dedup := NewDeduplicator(5*time.Second, 100)

	filterCfg := FilterConfig{
		Level: "warn",
	}

	entries := []LogEntry{
		{Level: "warn", Message: "disk full", Fields: map[string]interface{}{}},
		{Level: "info", Message: "disk full", Fields: map[string]interface{}{}},
		{Level: "warn", Message: "disk full", Fields: map[string]interface{}{}}, // duplicate
	}

	var passed []LogEntry
	for _, e := range entries {
		if !Filter(e, filterCfg) {
			continue
		}
		if dedup.IsDuplicate(e) {
			continue
		}
		passed = append(passed, e)
	}

	if len(passed) != 1 {
		t.Errorf("expected 1 entry after filter+dedup, got %d", len(passed))
	}
	if !strings.EqualFold(passed[0].Level, "warn") {
		t.Errorf("expected warn level, got %s", passed[0].Level)
	}
}

// TestDedup_EntryKeyConsistency ensures the same entry always produces the same key.
func TestDedup_EntryKeyConsistency(t *testing.T) {
	e := LogEntry{
		Level:   "error",
		Message: "connection refused",
		Fields:  map[string]interface{}{"host": "localhost"},
	}
	k1 := entryKey(e)
	k2 := entryKey(e)
	if k1 != k2 {
		t.Errorf("entryKey not deterministic: %s vs %s", k1, k2)
	}
}
