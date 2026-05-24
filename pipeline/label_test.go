package pipeline

import (
	"testing"
)

func labelEntry(fields map[string]string) LogEntry {
	return LogEntry{
		Level:   "info",
		Message: "test",
		Fields:  fields,
	}
}

func TestLabeler_StaticLabels(t *testing.T) {
	l := NewLabeler(LabelConfig{
		Static: map[string]string{"env": "prod", "region": "us-east-1"},
	})
	out := l.Apply(labelEntry(map[string]string{"svc": "api"}))
	if out.Fields["env"] != "prod" {
		t.Errorf("expected env=prod, got %q", out.Fields["env"])
	}
	if out.Fields["region"] != "us-east-1" {
		t.Errorf("expected region=us-east-1")
	}
	if out.Fields["svc"] != "api" {
		t.Errorf("original field svc should be preserved")
	}
}

func TestLabeler_CopyFrom(t *testing.T) {
	l := NewLabeler(LabelConfig{
		CopyFrom: map[string]string{"service_label": "svc"},
	})
	out := l.Apply(labelEntry(map[string]string{"svc": "gateway"}))
	if out.Fields["service_label"] != "gateway" {
		t.Errorf("expected service_label=gateway, got %q", out.Fields["service_label"])
	}
}

func TestLabeler_CopyFrom_MissingSource(t *testing.T) {
	l := NewLabeler(LabelConfig{
		CopyFrom: map[string]string{"dest": "missing_field"},
	})
	out := l.Apply(labelEntry(map[string]string{}))
	if _, ok := out.Fields["dest"]; ok {
		t.Error("dest should not be set when source field is missing")
	}
}

func TestLabeler_Prefix(t *testing.T) {
	l := NewLabeler(LabelConfig{
		Prefix: "lp",
		Static: map[string]string{"env": "staging"},
	})
	out := l.Apply(labelEntry(map[string]string{}))
	if out.Fields["lp_env"] != "staging" {
		t.Errorf("expected lp_env=staging, got %q", out.Fields["lp_env"])
	}
}

func TestLabeler_StaticDoesNotOverwriteExisting(t *testing.T) {
	l := NewLabeler(LabelConfig{
		Static: map[string]string{"env": "prod"},
	})
	out := l.Apply(labelEntry(map[string]string{"env": "dev"}))
	// static DOES overwrite — verify the documented behaviour
	if out.Fields["env"] != "prod" {
		t.Errorf("static label should overwrite existing field, got %q", out.Fields["env"])
	}
}

func TestLabeler_DoesNotMutateOriginal(t *testing.T) {
	original := labelEntry(map[string]string{"k": "v"})
	l := NewLabeler(LabelConfig{Static: map[string]string{"added": "yes"}})
	l.Apply(original)
	if _, ok := original.Fields["added"]; ok {
		t.Error("Apply must not mutate the original entry")
	}
}

func TestLabeler_EmptyConfig(t *testing.T) {
	l := NewLabeler(LabelConfig{})
	entry := labelEntry(map[string]string{"x": "1"})
	out := l.Apply(entry)
	if len(out.Fields) != 1 || out.Fields["x"] != "1" {
		t.Error("empty config should leave fields unchanged")
	}
}
