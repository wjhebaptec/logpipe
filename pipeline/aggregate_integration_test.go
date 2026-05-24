package pipeline

import (
	"sync"
	"testing"
	"time"
)

func TestAggregate_WithFilter(t *testing.T) {
	a := NewAggregator(AggregateConfig{GroupBy: "service"})
	filterCfg := FilterConfig{Level: "error"}

	entries := []LogEntry{
		{Timestamp: time.Now(), Level: "error", Message: "boom", Fields: map[string]any{"service": "auth"}},
		{Timestamp: time.Now(), Level: "info", Message: "ok", Fields: map[string]any{"service": "auth"}},
		{Timestamp: time.Now(), Level: "error", Message: "fail", Fields: map[string]any{"service": "db"}},
	}

	for _, e := range entries {
		if Filter(e, filterCfg) {
			a.Add(e)
		}
	}

	results := a.Flush()
	if len(results) != 2 {
		t.Fatalf("expected 2 groups after filter, got %d", len(results))
	}
	for _, r := range results {
		if r.Levels["ERROR"] == 0 {
			t.Errorf("expected ERROR count > 0 for group %s", r.Key)
		}
	}
}

func TestAggregate_ConcurrentAdds(t *testing.T) {
	a := NewAggregator(AggregateConfig{GroupBy: "source", MaxGroups: 50})
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				a.Add(LogEntry{
					Timestamp: time.Now(),
					Level:     "info",
					Message:   "msg",
					Fields:    map[string]any{"source": "worker"},
				})
			}
		}(i)
	}
	wg.Wait()

	results := a.Flush()
	if len(results) != 1 {
		t.Fatalf("expected 1 group, got %d", len(results))
	}
	if results[0].Count != 200 {
		t.Errorf("expected count=200, got %d", results[0].Count)
	}
}

func TestAggregate_FirstAndLastTimestamps(t *testing.T) {
	a := NewAggregator(AggregateConfig{GroupBy: "source"})

	t1 := time.Now()
	t2 := t1.Add(5 * time.Second)

	a.Add(LogEntry{Timestamp: t1, Level: "info", Message: "first", Fields: map[string]any{"source": "svc"}})
	a.Add(LogEntry{Timestamp: t2, Level: "info", Message: "last", Fields: map[string]any{"source": "svc"}})

	results := a.Flush()
	if len(results) != 1 {
		t.Fatalf("expected 1 result")
	}
	if !results[0].First.Equal(t1) {
		t.Errorf("expected First=%v, got %v", t1, results[0].First)
	}
	if !results[0].Last.Equal(t2) {
		t.Errorf("expected Last=%v, got %v", t2, results[0].Last)
	}
}
