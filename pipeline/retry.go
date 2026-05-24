package pipeline

import (
	"context"
	"errors"
	"time"
)

// RetryConfig holds configuration for retry behaviour.
type RetryConfig struct {
	MaxAttempts int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
}

// RetryFunc is the function signature that Retry will attempt.
type RetryFunc func(ctx context.Context) error

// ErrMaxRetriesExceeded is returned when all retry attempts are exhausted.
var ErrMaxRetriesExceeded = errors.New("max retries exceeded")

// Retryer executes a function with exponential backoff retry logic.
type Retryer struct {
	cfg RetryConfig
}

// NewRetryer creates a Retryer with the given config, applying safe defaults.
func NewRetryer(cfg RetryConfig) *Retryer {
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 3
	}
	if cfg.InitialDelay <= 0 {
		cfg.InitialDelay = 100 * time.Millisecond
	}
	if cfg.MaxDelay <= 0 {
		cfg.MaxDelay = 5 * time.Second
	}
	if cfg.Multiplier <= 1 {
		cfg.Multiplier = 2.0
	}
	return &Retryer{cfg: cfg}
}

// Do runs fn up to MaxAttempts times, backing off exponentially between attempts.
// It returns nil on first success, or ErrMaxRetriesExceeded wrapping the last error.
func (r *Retryer) Do(ctx context.Context, fn RetryFunc) error {
	delay := r.cfg.InitialDelay
	var lastErr error
	for attempt := 0; attempt < r.cfg.MaxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		lastErr = fn(ctx)
		if lastErr == nil {
			return nil
		}
		if attempt < r.cfg.MaxAttempts-1 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
			delay = time.Duration(float64(delay) * r.cfg.Multiplier)
			if delay > r.cfg.MaxDelay {
				delay = r.cfg.MaxDelay
			}
		}
	}
	return errors.Join(ErrMaxRetriesExceeded, lastErr)
}
