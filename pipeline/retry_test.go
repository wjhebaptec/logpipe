package pipeline

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

var errTemp = errors.New("temporary error")

func TestRetryer_SuccessOnFirstAttempt(t *testing.T) {
	r := NewRetryer(RetryConfig{MaxAttempts: 3})
	calls := 0
	err := r.Do(context.Background(), func(_ context.Context) error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestRetryer_RetriesAndSucceeds(t *testing.T) {
	r := NewRetryer(RetryConfig{MaxAttempts: 3, InitialDelay: time.Millisecond, Multiplier: 2})
	var calls int32
	err := r.Do(context.Background(), func(_ context.Context) error {
		if atomic.AddInt32(&calls, 1) < 3 {
			return errTemp
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestRetryer_ExhaustsRetries(t *testing.T) {
	r := NewRetryer(RetryConfig{MaxAttempts: 3, InitialDelay: time.Millisecond})
	var calls int32
	err := r.Do(context.Background(), func(_ context.Context) error {
		atomic.AddInt32(&calls, 1)
		return errTemp
	})
	if !errors.Is(err, ErrMaxRetriesExceeded) {
		t.Fatalf("expected ErrMaxRetriesExceeded, got %v", err)
	}
	if !errors.Is(err, errTemp) {
		t.Fatalf("expected wrapped errTemp, got %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestRetryer_ContextCancelled(t *testing.T) {
	r := NewRetryer(RetryConfig{MaxAttempts: 5, InitialDelay: 50 * time.Millisecond})
	ctx, cancel := context.WithCancel(context.Background())
	var calls int32
	err := r.Do(ctx, func(_ context.Context) error {
		if atomic.AddInt32(&calls, 1) == 1 {
			cancel()
		}
		return errTemp
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestRetryer_Defaults(t *testing.T) {
	r := NewRetryer(RetryConfig{})
	if r.cfg.MaxAttempts != 3 {
		t.Errorf("expected default MaxAttempts=3, got %d", r.cfg.MaxAttempts)
	}
	if r.cfg.Multiplier != 2.0 {
		t.Errorf("expected default Multiplier=2.0, got %f", r.cfg.Multiplier)
	}
}

func TestRetryer_MaxDelayCap(t *testing.T) {
	r := NewRetryer(RetryConfig{
		MaxAttempts:  4,
		InitialDelay: time.Millisecond,
		MaxDelay:     2 * time.Millisecond,
		Multiplier:   100,
	})
	var calls int32
	_ = r.Do(context.Background(), func(_ context.Context) error {
		atomic.AddInt32(&calls, 1)
		return errTemp
	})
	if calls != 4 {
		t.Fatalf("expected 4 calls, got %d", calls)
	}
}
