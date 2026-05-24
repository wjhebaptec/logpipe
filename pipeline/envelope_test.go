package pipeline

import (
	"testing"
	"time"
)

func envelopeEntry() LogEntry {
	return LogEntry{Level: "info", Message: "test", Timestamp: time.Now()}
}

func TestEnvelope_NewDefaults(t *testing.T) {
	e := NewEnvelope(envelopeEntry(), "svc-a", 5*time.Second, 0)
	if e.MaxHops != 10 {
		t.Errorf("expected default MaxHops 10, got %d", e.MaxHops)
	}
	if e.Hops != 0 {
		t.Errorf("expected initial hops 0, got %d", e.Hops)
	}
	if e.Source != "svc-a" {
		t.Errorf("unexpected source: %s", e.Source)
	}
}

func TestEnvelope_AdvanceIncrementsHops(t *testing.T) {
	e := NewEnvelope(envelopeEntry(), "src", 10*time.Second, 5)
	if err := e.Advance(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.Hops != 1 {
		t.Errorf("expected hops=1, got %d", e.Hops)
	}
}

func TestEnvelope_AdvanceExceedsMaxHops(t *testing.T) {
	e := NewEnvelope(envelopeEntry(), "src", 10*time.Second, 2)
	_ = e.Advance()
	_ = e.Advance()
	err := e.Advance()
	if err == nil {
		t.Fatal("expected error when max hops exceeded")
	}
}

func TestEnvelope_ExpiredDeadline(t *testing.T) {
	e := NewEnvelope(envelopeEntry(), "src", -1*time.Second, 5)
	if !e.Expired() {
		t.Error("expected envelope to be expired")
	}
	err := e.Advance()
	if err == nil {
		t.Error("expected error advancing expired envelope")
	}
}

func TestEnvelope_NotExpired(t *testing.T) {
	e := NewEnvelope(envelopeEntry(), "src", 30*time.Second, 5)
	if e.Expired() {
		t.Error("envelope should not be expired")
	}
}

func TestEnvelope_AddForwardTarget(t *testing.T) {
	e := NewEnvelope(envelopeEntry(), "src", 10*time.Second, 5)
	e.AddForwardTarget("output-a")
	e.AddForwardTarget("output-b")
	if len(e.ForwardTo) != 2 {
		t.Errorf("expected 2 targets, got %d", len(e.ForwardTo))
	}
	if e.ForwardTo[0] != "output-a" || e.ForwardTo[1] != "output-b" {
		t.Errorf("unexpected targets: %v", e.ForwardTo)
	}
}
