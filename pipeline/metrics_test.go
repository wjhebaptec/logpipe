package pipeline

import (
	"sync"
	"testing"
)

func TestMetrics_InitialValuesAreZero(t *testing.T) {
	m := NewMetrics()
	if m.Received != 0 || m.Filtered != 0 || m.Forwarded != 0 {
		t.Error("expected all initial metric values to be zero")
	}
	if m.ParseErrors != 0 || m.OutputErrors != 0 {
		t.Error("expected error counters to be zero")
	}
}

func TestMetrics_Increments(t *testing.T) {
	m := NewMetrics()
	m.IncReceived()
	m.IncReceived()
	m.IncFiltered()
	m.IncForwarded()
	m.IncParseErrors()
	m.IncOutputErrors()

	if m.Received != 2 {
		t.Errorf("expected Received=2, got %d", m.Received)
	}
	if m.Filtered != 1 {
		t.Errorf("expected Filtered=1, got %d", m.Filtered)
	}
	if m.Forwarded != 1 {
		t.Errorf("expected Forwarded=1, got %d", m.Forwarded)
	}
	if m.ParseErrors != 1 {
		t.Errorf("expected ParseErrors=1, got %d", m.ParseErrors)
	}
	if m.OutputErrors != 1 {
		t.Errorf("expected OutputErrors=1, got %d", m.OutputErrors)
	}
}

func TestMetrics_LevelCounts(t *testing.T) {
	m := NewMetrics()
	m.IncLevel("info")
	m.IncLevel("info")
	m.IncLevel("error")

	counts := m.LevelCounts()
	if counts["info"] != 2 {
		t.Errorf("expected info=2, got %d", counts["info"])
	}
	if counts["error"] != 1 {
		t.Errorf("expected error=1, got %d", counts["error"])
	}
}

func TestMetrics_Snapshot(t *testing.T) {
	m := NewMetrics()
	m.IncReceived()
	m.IncFiltered()
	m.IncLevel("warn")

	snap := m.Snapshot()
	if snap.Received != 1 {
		t.Errorf("snapshot Received expected 1, got %d", snap.Received)
	}
	if snap.Filtered != 1 {
		t.Errorf("snapshot Filtered expected 1, got %d", snap.Filtered)
	}
	if snap.LevelCounts()["warn"] != 1 {
		t.Errorf("snapshot warn level expected 1")
	}
}

func TestMetrics_ConcurrentIncrements(t *testing.T) {
	m := NewMetrics()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m.IncReceived()
			m.IncLevel("debug")
		}()
	}
	wg.Wait()
	if m.Received != 100 {
		t.Errorf("expected Received=100, got %d", m.Received)
	}
	if m.LevelCounts()["debug"] != 100 {
		t.Errorf("expected debug level count=100")
	}
}
