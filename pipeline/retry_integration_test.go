package pipeline

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestRetryer_WithCircuitBreaker validates that a Retryer and CircuitBreaker
// interact correctly: once the breaker opens, retries stop propagating errors.
func TestRetryer_WithCircuitBreaker(t *testing.T) {
	cb := NewCircuitBreaker(2, 500*time.Millisecond)
	r := NewRetryer(RetryConfig{
		MaxAttempts:  5,
		InitialDelay: time.Millisecond,
		Multiplier:   1,
	})

	writeErr := errors.New("write failure")
	calls := 0

	// Wrap circuit-breaker guarded call inside retryer.
	err := r.Do(context.Background(), func(ctx context.Context) error {
		return cb.Do(func() error {
			calls++
			return writeErr
		})
	})

	// After 2 failures the breaker opens; subsequent retry calls will get
	// ErrCircuitOpen rather than writeErr.
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// Circuit should have opened at some point during retries.
	if cb.State() != StateOpen {
		t.Errorf("expected circuit to be open, got %v", cb.State())
	}
	// Calls should be capped at threshold, not all 5 attempts.
	if calls > 3 {
		t.Errorf("expected at most 3 calls before circuit opened, got %d", calls)
	}
}

// TestRetryer_ImmediateContextExpiry ensures no attempts are made when the
// context is already cancelled before Do is called.
func TestRetryer_ImmediateContextExpiry(t *testing.T) {
	r := NewRetryer(RetryConfig{MaxAttempts: 3, InitialDelay: time.Millisecond})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	calls := 0
	err := r.Do(ctx, func(_ context.Context) error {
		calls++
		return nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if calls != 0 {
		t.Fatalf("expected 0 calls, got %d", calls)
	}
}
