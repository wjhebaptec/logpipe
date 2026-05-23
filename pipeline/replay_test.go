package pipeline

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"
)

const replaySample = `{"level":"info","message":"server started"}
{"level":"error","message":"disk full"}
{"level":"info","message":"request handled"}
{"level":"debug","message":"verbose detail"}
`

func TestReplay_NoFilter(t *testing.T) {
	var buf bytes.Buffer
	n, err := replayReader(context.Background(), strings.NewReader(replaySample), &buf, ReplayOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 4 {
		t.Fatalf("expected 4 entries, got %d", n)
	}
}

func TestReplay_LevelFilter(t *testing.T) {
	var buf bytes.Buffer
	n, err := replayReader(context.Background(), strings.NewReader(replaySample), &buf, ReplayOptions{
		Filter: &FilterConfig{Level: "error"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected 1 entry, got %d", n)
	}
	if !strings.Contains(buf.String(), "disk full") {
		t.Errorf("expected 'disk full' in output, got: %s", buf.String())
	}
}

func TestReplay_ContainsFilter(t *testing.T) {
	var buf bytes.Buffer
	n, err := replayReader(context.Background(), strings.NewReader(replaySample), &buf, ReplayOptions{
		Filter: &FilterConfig{Contains: "request"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected 1 entry, got %d", n)
	}
}

func TestReplay_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately
	var buf bytes.Buffer
	_, err := replayReader(ctx, strings.NewReader(replaySample), &buf, ReplayOptions{})
	if err == nil {
		t.Fatal("expected context cancellation error")
	}
}

func TestReplay_RateLimit(t *testing.T) {
	// 10 entries/sec → ~100ms each; with 2 entries expect ≥100ms total
	input := "{\"level\":\"info\",\"message\":\"a\"}\n{\"level\":\"info\",\"message\":\"b\"}\n"
	var buf bytes.Buffer
	start := time.Now()
	n, err := replayReader(context.Background(), strings.NewReader(input), &buf, ReplayOptions{
		RateLimit: 10,
	})
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 2 {
		t.Fatalf("expected 2 entries, got %d", n)
	}
	if elapsed < 100*time.Millisecond {
		t.Errorf("rate limit not respected: elapsed %v", elapsed)
	}
}
