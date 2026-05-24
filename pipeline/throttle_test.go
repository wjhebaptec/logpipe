package pipeline

import (
	"testing"
	"time"
)

func TestThrottler_AllowsUpToMax(t *testing.T) {
	th := NewThrottler(ThrottleConfig{MaxPerWindow: 3, Window: time.Second})
	entry := LogEntry{Level: "info", Message: "msg"}

	for i := 0; i < 3; i++ {
		if !th.Allow(entry) {
			t.Fatalf("expected entry %d to be allowed", i)
		}
	}
	if th.Allow(entry) {
		t.Fatal("expected 4th entry to be throttled")
	}
}

func TestThrottler_ResetsAfterWindow(t *testing.T) {
	th := NewThrottler(ThrottleConfig{MaxPerWindow: 2, Window: 50 * time.Millisecond})
	entry := LogEntry{Level: "info", Message: "msg"}

	th.Allow(entry)
	th.Allow(entry)
	if th.Allow(entry) {
		t.Fatal("expected throttle before window reset")
	}

	time.Sleep(60 * time.Millisecond)
	if !th.Allow(entry) {
		t.Fatal("expected allow after window reset")
	}
}

func TestThrottler_LevelFilter_OnlyThrottlesMatchingLevel(t *testing.T) {
	th := NewThrottler(ThrottleConfig{MaxPerWindow: 1, Window: time.Second, MatchLevel: "error"})

	info := LogEntry{Level: "info", Message: "info msg"}
	err := LogEntry{Level: "error", Message: "err msg"}

	// info entries should always pass through
	for i := 0; i < 5; i++ {
		if !th.Allow(info) {
			t.Fatalf("info entry %d should not be throttled", i)
		}
	}

	if !th.Allow(err) {
		t.Fatal("first error entry should be allowed")
	}
	if th.Allow(err) {
		t.Fatal("second error entry should be throttled")
	}
}

func TestThrottler_Remaining(t *testing.T) {
	th := NewThrottler(ThrottleConfig{MaxPerWindow: 5, Window: time.Second})
	entry := LogEntry{Level: "warn", Message: "msg"}

	if r := th.Remaining(); r != 5 {
		t.Fatalf("expected 5 remaining, got %d", r)
	}
	th.Allow(entry)
	th.Allow(entry)
	if r := th.Remaining(); r != 3 {
		t.Fatalf("expected 3 remaining, got %d", r)
	}
}

func TestThrottler_Defaults(t *testing.T) {
	th := NewThrottler(ThrottleConfig{})
	if th.cfg.MaxPerWindow != 100 {
		t.Fatalf("expected default MaxPerWindow=100, got %d", th.cfg.MaxPerWindow)
	}
	if th.cfg.Window != time.Second {
		t.Fatalf("expected default Window=1s, got %v", th.cfg.Window)
	}
}

func TestThrottler_ZeroRemaining(t *testing.T) {
	th := NewThrottler(ThrottleConfig{MaxPerWindow: 2, Window: time.Second})
	entry := LogEntry{Level: "debug", Message: "msg"}
	th.Allow(entry)
	th.Allow(entry)
	th.Allow(entry) // exceeds max
	if r := th.Remaining(); r != 0 {
		t.Fatalf("expected 0 remaining, got %d", r)
	}
}
