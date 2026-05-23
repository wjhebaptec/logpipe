package pipeline

import (
	"testing"
	"time"
)

func TestCircuitBreaker_InitiallyClosed(t *testing.T) {
	cb := NewCircuitBreaker(3, time.Second)
	if !cb.Allow() {
		t.Fatal("expected Allow() true when closed")
	}
	if cb.CurrentState() != StateClosed {
		t.Fatal("expected StateClosed initially")
	}
}

func TestCircuitBreaker_OpensAfterThreshold(t *testing.T) {
	cb := NewCircuitBreaker(3, 10*time.Second)
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}
	if cb.CurrentState() != StateOpen {
		t.Fatal("expected StateOpen after threshold failures")
	}
	if cb.Allow() {
		t.Fatal("expected Allow() false when open")
	}
}

func TestCircuitBreaker_SuccessResetFailures(t *testing.T) {
	cb := NewCircuitBreaker(3, 10*time.Second)
	cb.RecordFailure()
	cb.RecordFailure()
	cb.RecordSuccess()
	if cb.CurrentState() != StateClosed {
		t.Fatal("expected StateClosed after success resets failures")
	}
	// Should not open with only 2 failures after reset
	cb.RecordFailure()
	cb.RecordFailure()
	if cb.CurrentState() != StateClosed {
		t.Fatal("expected still closed after 2 failures post reset")
	}
}

func TestCircuitBreaker_HalfOpenAfterTimeout(t *testing.T) {
	cb := NewCircuitBreaker(2, 20*time.Millisecond)
	cb.RecordFailure()
	cb.RecordFailure()
	if cb.CurrentState() != StateOpen {
		t.Fatal("expected StateOpen")
	}
	time.Sleep(30 * time.Millisecond)
	if !cb.Allow() {
		t.Fatal("expected Allow() true after reset timeout")
	}
	if cb.CurrentState() != StateHalfOpen {
		t.Fatal("expected StateHalfOpen after timeout")
	}
}

func TestCircuitBreaker_HalfOpen_SuccessCloses(t *testing.T) {
	cb := NewCircuitBreaker(2, 10*time.Millisecond)
	cb.RecordFailure()
	cb.RecordFailure()
	time.Sleep(15 * time.Millisecond)
	cb.Allow() // transitions to half-open
	cb.RecordSuccess()
	if cb.CurrentState() != StateClosed {
		t.Fatal("expected StateClosed after probe success")
	}
}

func TestCircuitBreaker_HalfOpen_FailureReopens(t *testing.T) {
	cb := NewCircuitBreaker(2, 10*time.Millisecond)
	cb.RecordFailure()
	cb.RecordFailure()
	time.Sleep(15 * time.Millisecond)
	cb.Allow() // transitions to half-open
	cb.RecordFailure()
	if cb.CurrentState() != StateOpen {
		t.Fatal("expected StateOpen after probe failure")
	}
}

func TestCircuitBreaker_DefaultsForInvalidConfig(t *testing.T) {
	cb := NewCircuitBreaker(0, 0)
	if cb.threshold != 3 {
		t.Fatalf("expected default threshold 3, got %d", cb.threshold)
	}
	if cb.resetTimeout != 10*time.Second {
		t.Fatalf("expected default resetTimeout 10s, got %v", cb.resetTimeout)
	}
}
