package pipeline

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func corrEntry(corrID, msg, level string) LogEntry {
	return LogEntry{
		Message: msg,
		Level:   level,
		Fields:  map[string]string{"correlation_id": corrID},
	}
}

func TestCorrelator_DefaultConfig(t *testing.T) {
	c := NewCorrelator(CorrelateConfig{})
	if c.cfg.Field != "correlation_id" {
		t.Errorf("expected default field 'correlation_id', got %q", c.cfg.Field)
	}
	if c.cfg.TTL != 30*time.Second {
		t.Errorf("expected default TTL 30s, got %v", c.cfg.TTL)
	}
	if c.cfg.MaxGroups != 1000 {
		t.Errorf("expected default MaxGroups 1000, got %d", c.cfg.MaxGroups)
	}
}

func TestCorrelator_AddAndGet(t *testing.T) {
	c := NewCorrelator(CorrelateConfig{TTL: time.Minute})
	e1 := corrEntry("req-1", "start", "info")
	e2 := corrEntry("req-1", "end", "info")
	c.Add(e1)
	c.Add(e2)
	entries, ok := c.Get("req-1")
	if !ok {
		t.Fatal("expected group to exist")
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}
}

func TestCorrelator_MissingField_ReturnsFalse(t *testing.T) {
	c := NewCorrelator(CorrelateConfig{TTL: time.Minute})
	e := LogEntry{Message: "no id", Level: "info", Fields: map[string]string{}}
	if c.Add(e) {
		t.Error("expected Add to return false for missing field")
	}
}

func TestCorrelator_ExpiredGroup_NotReturned(t *testing.T) {
	c := NewCorrelator(CorrelateConfig{TTL: 10 * time.Millisecond})
	c.Add(corrEntry("req-2", "msg", "info"))
	time.Sleep(20 * time.Millisecond)
	_, ok := c.Get("req-2")
	if ok {
		t.Error("expected expired group to not be returned")
	}
}

func TestCorrelator_MaxGroupsEnforced(t *testing.T) {
	c := NewCorrelator(CorrelateConfig{TTL: time.Minute, MaxGroups: 3})
	for i := 0; i < 3; i++ {
		c.Add(corrEntry(fmt.Sprintf("req-%d", i), "msg", "info"))
	}
	if c.Len() != 3 {
		t.Errorf("expected 3 groups, got %d", c.Len())
	}
	added := c.Add(corrEntry("req-overflow", "msg", "info"))
	if added {
		t.Error("expected Add to fail when MaxGroups reached")
	}
}

func TestCorrelator_Len_EvictsExpired(t *testing.T) {
	c := NewCorrelator(CorrelateConfig{TTL: 10 * time.Millisecond})
	c.Add(corrEntry("req-3", "msg", "info"))
	time.Sleep(20 * time.Millisecond)
	if c.Len() != 0 {
		t.Errorf("expected 0 after expiry, got %d", c.Len())
	}
}

func TestCorrelator_Concurrent(t *testing.T) {
	c := NewCorrelator(CorrelateConfig{TTL: time.Minute, MaxGroups: 500})
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("req-%d", i%10)
			c.Add(corrEntry(key, "msg", "info"))
			c.Get(key)
		}(i)
	}
	wg.Wait()
}
