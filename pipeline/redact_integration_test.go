package pipeline

import (
	"testing"
)

// TestRedact_WithFilter verifies that Redactor and Filter compose correctly:
// only entries passing the filter have redaction applied.
func TestRedact_WithFilter(t *testing.T) {
	redactor := NewRedactor([]RedactRule{
		{Field: "secret"},
	})

	entries := []LogEntry{
		{Level: "error", Message: "bad thing", Fields: map[string]interface{}{"secret": "topsecret"}},
		{Level: "info", Message: "ok thing", Fields: map[string]interface{}{"secret": "topsecret"}},
	}

	cfgFilter := FilterConfig{Level: "error"}

	for _, entry := range entries {
		if Filter(entry, cfgFilter) {
			out := redactor.Redact(entry)
			if out.Fields["secret"] != "[REDACTED]" {
				t.Errorf("expected secret redacted for level=%s", entry.Level)
			}
		}
	}
}

// TestRedact_WithTransform verifies that Redact and Transform can be chained
// without interfering with each other's field operations.
func TestRedact_WithTransform(t *testing.T) {
	redactor := NewRedactor([]RedactRule{
		{Field: "api_key"},
	})

	transformCfg := TransformConfig{
		AddFields: map[string]string{"env": "prod"},
	}

	entry := LogEntry{
		Level:   "info",
		Message: "request",
		Fields:  map[string]interface{}{"api_key": "sk-abc123"},
	}

	// First transform, then redact.
	transformed := Transform(entry, transformCfg)
	redacted := redactor.Redact(transformed)

	if redacted.Fields["api_key"] != "[REDACTED]" {
		t.Errorf("expected api_key redacted, got %v", redacted.Fields["api_key"])
	}
	if redacted.Fields["env"] != "prod" {
		t.Errorf("expected env=prod from transform, got %v", redacted.Fields["env"])
	}
}
