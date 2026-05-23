package pipeline

import (
	"strings"
	"testing"
	"time"
)

func TestParseEntry_JSON(t *testing.T) {
	line := `{"level":"error","message":"something failed","timestamp":"2024-01-15T10:00:00Z","service":"api"}`
	entry, err := ParseEntry(line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.Level != "ERROR" {
		t.Errorf("expected ERROR, got %s", entry.Level)
	}
	if entry.Message != "something failed" {
		t.Errorf("unexpected message: %s", entry.Message)
	}
	if entry.Fields["service"] != "api" {
		t.Errorf("expected field service=api, got %s", entry.Fields["service"])
	}
	if entry.Timestamp.Year() != 2024 {
		t.Errorf("unexpected timestamp year: %d", entry.Timestamp.Year())
	}
}

func TestParseEntry_JSON_MsgAlias(t *testing.T) {
	line := `{"level":"info","msg":"startup complete"}`
	entry, err := ParseEntry(line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.Message != "startup complete" {
		t.Errorf("unexpected message: %s", entry.Message)
	}
}

func TestParseEntry_PlainText(t *testing.T) {
	line := "2024-01-15 ERROR failed to connect to database"
	entry, err := ParseEntry(line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.Level != "ERROR" {
		t.Errorf("expected ERROR, got %s", entry.Level)
	}
	if entry.Message != line {
		t.Errorf("expected raw line as message")
	}
	if entry.Timestamp.IsZero() {
		t.Error("timestamp should not be zero")
	}
}

func TestParseEntry_PlainText_DefaultLevel(t *testing.T) {
	line := "server started on port 8080"
	entry, err := ParseEntry(line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.Level != "INFO" {
		t.Errorf("expected default INFO, got %s", entry.Level)
	}
}

func TestParseEntry_EmptyLine(t *testing.T) {
	_, err := ParseEntry("   ")
	if err == nil {
		t.Error("expected error for empty line")
	}
}

func TestParseEntry_InvalidJSON(t *testing.T) {
	_, err := ParseEntry("{invalid json")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "json parse error") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestParseEntry_TimestampFallback(t *testing.T) {
	before := time.Now()
	line := `{"level":"warn","message":"disk space low"}`
	entry, err := ParseEntry(line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.Timestamp.Before(before) {
		t.Error("timestamp should be set to approximately now when missing")
	}
}
