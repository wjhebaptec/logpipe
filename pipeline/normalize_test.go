package pipeline

import (
	"testing"
	"time"
)

func normEntry(level, msg string, fields map[string]interface{}) LogEntry {
	return LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   msg,
		Fields:    fields,
	}
}

func TestNormalizer_DefaultConfig(t *testing.T) {
	n := NewNormalizer(NormalizeConfig{})
	if n.cfg.TimestampField != "time" {
		t.Errorf("expected default timestamp field 'time', got %q", n.cfg.TimestampField)
	}
	if n.cfg.LevelField != "level" {
		t.Errorf("expected default level field 'level', got %q", n.cfg.LevelField)
	}
	if len(n.cfg.TimestampFormats) == 0 {
		t.Error("expected default timestamp formats to be set")
	}
}

func TestNormalizer_UppercasesLevel(t *testing.T) {
	n := NewNormalizer(NormalizeConfig{})
	out := n.Normalize(normEntry("warn", "something", nil))
	if out.Level != "WARN" {
		t.Errorf("expected level WARN, got %q", out.Level)
	}
}

func TestNormalizer_TrimsLevelWhitespace(t *testing.T) {
	n := NewNormalizer(NormalizeConfig{})
	out := n.Normalize(normEntry("  error  ", "msg", nil))
	if out.Level != "ERROR" {
		t.Errorf("expected ERROR, got %q", out.Level)
	}
}

func TestNormalizer_NormalizesTimestampField(t *testing.T) {
	n := NewNormalizer(NormalizeConfig{})
	fields := map[string]interface{}{
		"time": "2024-03-15 09:00:00",
	}
	out := n.Normalize(normEntry("info", "msg", fields))
	got, ok := out.Fields["time"].(string)
	if !ok {
		t.Fatal("expected time field to be a string")
	}
	if got != "2024-03-15T09:00:00Z" {
		t.Errorf("unexpected normalized time: %q", got)
	}
}

func TestNormalizer_UnrecognizedTimestampLeftAlone(t *testing.T) {
	n := NewNormalizer(NormalizeConfig{})
	fields := map[string]interface{}{
		"time": "not-a-date",
	}
	out := n.Normalize(normEntry("info", "msg", fields))
	if out.Fields["time"] != "not-a-date" {
		t.Errorf("expected unrecognized timestamp to be left unchanged")
	}
}

func TestNormalizer_TrimFields(t *testing.T) {
	n := NewNormalizer(NormalizeConfig{
		TrimFields: []string{"service"},
	})
	fields := map[string]interface{}{
		"service": "  auth-service  ",
	}
	out := n.Normalize(normEntry("info", "msg", fields))
	if out.Fields["service"] != "auth-service" {
		t.Errorf("expected trimmed service field, got %q", out.Fields["service"])
	}
}

func TestNormalizer_DoesNotMutateOriginal(t *testing.T) {
	n := NewNormalizer(NormalizeConfig{})
	fields := map[string]interface{}{"service": "  svc  "}
	entry := normEntry("debug", "msg", fields)
	n.Normalize(entry)
	if entry.Fields["service"] != "  svc  " {
		t.Error("original entry was mutated")
	}
}
