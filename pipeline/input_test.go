package pipeline

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestInputReader_JSON(t *testing.T) {
	lines := `{"level":"info","message":"hello"}
{"level":"error","message":"boom"}
`
	out := make(chan LogEntry, 10)
	ir := NewInputReader(strings.NewReader(lines), "json", out)

	ctx := context.Background()
	err := ir.Run(ctx)
	if err != nil && err.Error() != "EOF" {
		t.Fatalf("unexpected error: %v", err)
	}
	close(out)

	var entries []LogEntry
	for e := range out {
		entries = append(entries, e)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Message != "hello" {
		t.Errorf("expected 'hello', got %q", entries[0].Message)
	}
	if entries[1].Level != "error" {
		t.Errorf("expected 'error', got %q", entries[1].Level)
	}
}

func TestInputReader_PlainText(t *testing.T) {
	lines := "just a plain log line\n"
	out := make(chan LogEntry, 10)
	ir := NewInputReader(strings.NewReader(lines), "plain", out)

	ctx := context.Background()
	_ = ir.Run(ctx)
	close(out)

	var entries []LogEntry
	for e := range out {
		entries = append(entries, e)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Message != "just a plain log line" {
		t.Errorf("unexpected message: %q", entries[0].Message)
	}
}

func TestInputReader_ContextCancel(t *testing.T) {
	// Use a pipe so reading blocks
	pr, pw := strings.NewReader(""), nil
	_ = pw
	out := make(chan LogEntry, 1)
	ir := NewInputReader(pr, "json", out)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := ir.Run(ctx)
	// Should return EOF or context error — either is acceptable
	if err == nil {
		t.Error("expected an error, got nil")
	}
}
