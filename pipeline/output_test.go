package pipeline

import (
	"bytes"
	"strings"
	"testing"
)

func TestOutputWriter_Write_JSON(t *testing.T) {
	var buf bytes.Buffer
	ow := &OutputWriter{Name: "test", Format: "json", Writer: &buf}

	entry := LogEntry{Level: "info", Message: "hello world", Fields: map[string]interface{}{}}
	if err := ow.Write(entry); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "hello world") {
		t.Errorf("expected message in output, got: %q", out)
	}
	if !strings.Contains(out, "info") {
		t.Errorf("expected level in output, got: %q", out)
	}
}

func TestOutputWriter_Write_Plain(t *testing.T) {
	var buf bytes.Buffer
	ow := &OutputWriter{Name: "test", Format: "plain", Writer: &buf}

	entry := LogEntry{Level: "warn", Message: "disk full", Fields: map[string]interface{}{}}
	if err := ow.Write(entry); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "disk full") {
		t.Errorf("expected message in output, got: %q", out)
	}
}

func TestOutputWriter_Write_WithFields(t *testing.T) {
	var buf bytes.Buffer
	ow := &OutputWriter{Name: "test", Format: "json", Writer: &buf}

	entry := LogEntry{
		Level:   "error",
		Message: "connection refused",
		Fields:  map[string]interface{}{"host": "localhost", "port": 5432},
	}
	if err := ow.Write(entry); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "localhost") {
		t.Errorf("expected field value 'localhost' in output, got: %q", out)
	}
}

func TestNewFileOutput_Stdout(t *testing.T) {
	ow, err := NewFileOutput("stdout-out", "-", "plain")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ow.Writer == nil {
		t.Error("expected non-nil writer")
	}
}

func TestNewFileOutput_InvalidPath(t *testing.T) {
	_, err := NewFileOutput("bad", "/nonexistent/dir/file.log", "json")
	if err == nil {
		t.Error("expected error for invalid path")
	}
}
