package pipeline

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestAlerter_NoMatchOnDifferentLevel(t *testing.T) {
	var fired int32
	rule := AlertRuleFromConfig("test", "error", "", 1, time.Minute, func(r string, c int, e LogEntry) {
		atomic.AddInt32(&fired, 1)
	})
	a := NewAlerter([]AlertRule{rule})
	a.Evaluate(LogEntry{Level: "info", Message: "hello"})
	if atomic.LoadInt32(&fired) != 0 {
		t.Fatal("expected no alert for non-matching level")
	}
}

func TestAlerter_FiresOnThreshold(t *testing.T) {
	var fired int32
	rule := AlertRuleFromConfig("errors", "error", "", 3, time.Minute, func(r string, c int, e LogEntry) {
		atomic.AddInt32(&fired, 1)
	})
	a := NewAlerter([]AlertRule{rule})
	entry := LogEntry{Level: "error", Message: "boom"}
	a.Evaluate(entry)
	a.Evaluate(entry)
	if atomic.LoadInt32(&fired) != 0 {
		t.Fatal("should not fire before threshold")
	}
	a.Evaluate(entry)
	if atomic.LoadInt32(&fired) != 1 {
		t.Fatal("expected alert after threshold")
	}
}

func TestAlerter_ContainsFilter(t *testing.T) {
	var fired int32
	rule := AlertRuleFromConfig("oom", "", "out of memory", 1, time.Minute, func(r string, c int, e LogEntry) {
		atomic.AddInt32(&fired, 1)
	})
	a := NewAlerter([]AlertRule{rule})
	a.Evaluate(LogEntry{Level: "error", Message: "disk full"})
	if atomic.LoadInt32(&fired) != 0 {
		t.Fatal("should not fire for non-matching message")
	}
	a.Evaluate(LogEntry{Level: "error", Message: "out of memory error"})
	if atomic.LoadInt32(&fired) != 1 {
		t.Fatal("expected alert on contains match")
	}
}

func TestAlerter_ResetsAfterFiring(t *testing.T) {
	var count int32
	rule := AlertRuleFromConfig("reset", "warn", "", 2, time.Minute, func(r string, c int, e LogEntry) {
		atomic.AddInt32(&count, 1)
	})
	a := NewAlerter([]AlertRule{rule})
	entry := LogEntry{Level: "warn", Message: "watch out"}
	a.Evaluate(entry)
	a.Evaluate(entry) // fires, resets
	a.Evaluate(entry) // count=1 again, no fire
	if atomic.LoadInt32(&count) != 1 {
		t.Fatalf("expected 1 alert, got %d", count)
	}
}

func TestAlerter_WindowExpiry(t *testing.T) {
	var fired int32
	rule := AlertRuleFromConfig("window", "error", "", 2, 50*time.Millisecond, func(r string, c int, e LogEntry) {
		atomic.AddInt32(&fired, 1)
	})
	a := NewAlerter([]AlertRule{rule})
	entry := LogEntry{Level: "error", Message: "err"}
	a.Evaluate(entry)
	time.Sleep(80 * time.Millisecond)
	a.Evaluate(entry) // first event expired, count=1, no fire
	if atomic.LoadInt32(&fired) != 0 {
		t.Fatal("expected no alert after window expiry")
	}
}

func TestAlertRuleFromConfig_Defaults(t *testing.T) {
	rule := AlertRuleFromConfig("d", "", "", 0, 0, nil)
	if rule.Threshold != 1 {
		t.Errorf("expected threshold default 1, got %d", rule.Threshold)
	}
	if rule.Window != time.Minute {
		t.Errorf("expected window default 1m, got %v", rule.Window)
	}
}
