package pipeline

import (
	"testing"
	"time"
)

// TestEnvelope_WithFilter verifies that an envelope's entry passes through
// the Filter stage correctly and that hop tracking still works afterward.
func TestEnvelope_WithFilter(t *testing.T) {
	entry := LogEntry{Level: "error", Message: "disk full", Timestamp: time.Now()}
	env := NewEnvelope(entry, "storage-svc", 10*time.Second, 5)

	cfg := FilterConfig{Level: "error"}
	if !matchesFilter(env.Entry, cfg) {
		t.Fatal("expected entry to match error filter")
	}

	if err := env.Advance(); err != nil {
		t.Fatalf("unexpected advance error: %v", err)
	}
	if env.Hops != 1 {
		t.Errorf("expected hops=1 after filter stage, got %d", env.Hops)
	}
}

// TestEnvelope_TTLExpiredBeforeRouting ensures that an expired envelope is
// detected before it reaches the router, preventing stale log delivery.
func TestEnvelope_TTLExpiredBeforeRouting(t *testing.T) {
	entry := LogEntry{Level: "warn", Message: "slow query", Timestamp: time.Now()}
	env := NewEnvelope(entry, "db-svc", -500*time.Millisecond, 5)

	if !env.Expired() {
		t.Fatal("envelope should be expired before routing")
	}

	err := env.Advance()
	if err == nil {
		t.Error("expected routing to fail for expired envelope")
	}
}

// TestEnvelope_MultiHopPipeline simulates an envelope passing through
// multiple named pipeline stages and verifies hop accounting.
func TestEnvelope_MultiHopPipeline(t *testing.T) {
	entry := LogEntry{Level: "info", Message: "user login", Timestamp: time.Now()}
	env := NewEnvelope(entry, "auth-svc", 30*time.Second, 3)

	stages := []string{"filter", "transform", "router"}
	for _, stage := range stages {
		env.AddForwardTarget(stage)
		if err := env.Advance(); err != nil {
			t.Fatalf("unexpected error at stage %s: %v", stage, err)
		}
	}

	if env.Hops != 3 {
		t.Errorf("expected 3 hops, got %d", env.Hops)
	}
	if len(env.ForwardTo) != 3 {
		t.Errorf("expected 3 forward targets, got %d", len(env.ForwardTo))
	}
}
