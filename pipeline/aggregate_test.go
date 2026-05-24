package pipeline

import (
	"testing"
	"time"
)

func aggEntry(level, source string) LogEntry {
	return LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   "test",
		Fields:    map[string]any{"source": source},
	}
}

func TestAggregator_DefaultConfig(t *testing.T) {
	a := NewAggregator(AggregateConfig{})
	if a.cfg.GroupBy != "source" {
		t.Errorf("expected default GroupBy=source, got %s", a.cfg.GroupBy)
	}
	if a.cfg.Window != 30*time.Second {
		t.Errorf("expected default Window=30s, got %v", a.cfg.Window)
	}
	if a.cfg.MaxGroups != 100 {
		t.Errorf("expected default MaxGroups=100, got %d", a.cfg.MaxGroups)
	}
}

func TestAggregator_AddAndLen(t *testing.T) {
	a := NewAggregator(AggregateConfig{GroupBy: "source"})
	a.Add(aggEntry("info", "app1"))
	a.Add(aggEntry("error", "app1"))
	a.Add(aggEntry("warn", "app2"))

	if a.Len() != 2 {
		t.Errorf("expected 2 groups, got %d", a.Len())
	}
}

func TestAggregator_FlushReturnsAndResets(t *testing.T) {
	a := NewAggregator(AggregateConfig{GroupBy: "source"})
	a.Add(aggEntry("info", "svc"))
	a.Add(aggEntry("error", "svc"))

	results := a.Flush()
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Count != 2 {
		t.Errorf("expected count=2, got %d", results[0].Count)
	}
	if a.Len() != 0 {
		t.Errorf("expected Len=0 after flush, got %d", a.Len())
	}
}

func TestAggregator_LevelCounts(t *testing.T) {
	a := NewAggregator(AggregateConfig{GroupBy: "source"})
	a.Add(aggEntry("INFO", "x"))
	a.Add(aggEntry("info", "x"))
	a.Add(aggEntry("ERROR", "x"))

	results := a.Flush()
	if results[0].Levels["INFO"] != 2 {
		t.Errorf("expected INFO=2, got %d", results[0].Levels["INFO"])
	}
	if results[0].Levels["ERROR"] != 1 {
		t.Errorf("expected ERROR=1, got %d", results[0].Levels["ERROR"])
	}
}

func TestAggregator_MaxGroups(t *testing.T) {
	a := NewAggregator(AggregateConfig{GroupBy: "source", MaxGroups: 2})
	a.Add(aggEntry("info", "a"))
	a.Add(aggEntry("info", "b"))
	ok := a.Add(aggEntry("info", "c"))

	if ok {
		t.Error("expected Add to return false when MaxGroups exceeded")
	}
	if a.Len() != 2 {
		t.Errorf("expected Len=2, got %d", a.Len())
	}
}

func TestAggregator_UnknownGroupKey(t *testing.T) {
	a := NewAggregator(AggregateConfig{GroupBy: "source"})
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   "no source field",
		Fields:    map[string]any{},
	}
	a.Add(entry)
	results := a.Flush()
	if len(results) != 1 || results[0].Key != "(unknown)" {
		t.Errorf("expected key=(unknown), got %+v", results)
	}
}
