package pipeline

import (
	"testing"
)

func TestRedact_NoRules(t *testing.T) {
	r := NewRedactor(nil)
	entry := LogEntry{Message: "hello world", Fields: map[string]interface{}{"token": "secret"}}
	out := r.Redact(entry)
	if out.Message != "hello world" {
		t.Errorf("expected message unchanged, got %q", out.Message)
	}
	if out.Fields["token"] != "secret" {
		t.Errorf("expected field unchanged")
	}
}

func TestRedact_FullFieldReplacement(t *testing.T) {
	r := NewRedactor([]RedactRule{
		{Field: "password"},
	})
	entry := LogEntry{
		Message: "login attempt",
		Fields:  map[string]interface{}{"password": "hunter2"},
	}
	out := r.Redact(entry)
	if out.Fields["password"] != "[REDACTED]" {
		t.Errorf("expected password redacted, got %v", out.Fields["password"])
	}
}

func TestRedact_PatternReplacement(t *testing.T) {
	r := NewRedactor([]RedactRule{
		{Field: "message", Pattern: `\b\d{4}-\d{4}-\d{4}-\d{4}\b`, Replace: "[CARD]"},
	})
	entry := LogEntry{Message: "charged card 1234-5678-9012-3456 ok", Fields: map[string]interface{}{}}
	out := r.Redact(entry)
	expected := "charged card [CARD] ok"
	if out.Message != expected {
		t.Errorf("expected %q, got %q", expected, out.Message)
	}
}

func TestRedact_DoesNotMutateOriginal(t *testing.T) {
	r := NewRedactor([]RedactRule{
		{Field: "token"},
	})
	original := LogEntry{
		Message: "test",
		Fields:  map[string]interface{}{"token": "abc123"},
	}
	_ = r.Redact(original)
	if original.Fields["token"] != "abc123" {
		t.Errorf("original entry was mutated")
	}
}

func TestRedact_InvalidPatternSkipped(t *testing.T) {
	r := NewRedactor([]RedactRule{
		{Field: "message", Pattern: `[invalid`},
	})
	if len(r.rules) != 0 {
		t.Errorf("expected invalid rule to be skipped")
	}
}

func TestRedact_CustomReplacement(t *testing.T) {
	r := NewRedactor([]RedactRule{
		{Field: "email", Pattern: `[^@]+@[^@]+`, Replace: "***@***.***"},
	})
	entry := LogEntry{
		Message: "user login",
		Fields:  map[string]interface{}{"email": "user@example.com"},
	}
	out := r.Redact(entry)
	if out.Fields["email"] != "***@***.***" {
		t.Errorf("expected custom replacement, got %v", out.Fields["email"])
	}
}
