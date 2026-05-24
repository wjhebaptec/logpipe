package pipeline

import (
	"testing"
	"time"
)

func TestCorrelate_WithFilter(t *testing.T) {
	cfg := CorrelateConfig{Field: "trace_id", TTL: time.Minute}
	c := NewCorrelator(cfg)

	filterCfg := FilterConfig{Level: "error"}

	entries := []LogEntry{
		{Message: "db timeout", Level: "error", Fields: map[string]string{"trace_id": "t1"}},
		{Message: "retry ok", Level: "info", Fields: map[string]string{"trace_id": "t1"}},
		{Message: "another error", Level: "error", Fields: map[string]string{"trace_id": "t1"}},
	}

	for _, e := range entries {
		if Filter(e, filterCfg) {
			c.Add(e)
		}
	}

	result, ok := c.Get("t1")
	if !ok {
		t.Fatal("expected group t1 to exist")
	}
	if len(result) != 2 {
		t.Errorf("expected 2 error entries, got %d", len(result))
	}
	for _, e := range result {
		if e.Level != "error" {
			t.Errorf("unexpected level %q in correlated group", e.Level)
		}
	}
}

func TestCorrelate_MultipleKeys(t *testing.T) {
	c := NewCorrelator(CorrelateConfig{TTL: time.Minute})

	for _, id := range []string{"a", "b", "a", "c", "b"} {
		c.Add(corrEntry(id, "msg", "info"))
	}

	for _, tc := range []struct {
		key      string
		expected int
	}{
		{"a", 2},
		{"b", 2},
		{"c", 1},
	} {
		entries, ok := c.Get(tc.key)
		if !ok {
			t.Errorf("key %q not found", tc.key)
			continue
		}
		if len(entries) != tc.expected {
			t.Errorf("key %q: expected %d entries, got %d", tc.key, tc.expected, len(entries))
		}
	}
}
