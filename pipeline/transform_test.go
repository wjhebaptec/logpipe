package pipeline

import (
	"testing"
	"strings"
)

func baseEntry() *LogEntry {
	return &LogEntry{
		Level:   "info",
		Message: "hello world",
		Fields:  map[string]interface{}{"app": "logpipe"},
	}
}

func TestTransform_AddFields(t *testing.T) {
	entry := baseEntry()
	Transform(entry, TransformConfig{
		AddFields: map[string]string{"env": "production"},
	})
	if entry.Fields["env"] != "production" {
		t.Errorf("expected env=production, got %v", entry.Fields["env"])
	}
}

func TestTransform_RemoveFields(t *testing.T) {
	entry := baseEntry()
	Transform(entry, TransformConfig{
		RemoveFields: []string{"app"},
	})
	if _, ok := entry.Fields["app"]; ok {
		t.Error("expected 'app' field to be removed")
	}
}

func TestTransform_RenameFields(t *testing.T) {
	entry := baseEntry()
	Transform(entry, TransformConfig{
		RenameFields: map[string]string{"app": "service"},
	})
	if _, ok := entry.Fields["app"]; ok {
		t.Error("expected 'app' to be removed after rename")
	}
	if entry.Fields["service"] != "logpipe" {
		t.Errorf("expected service=logpipe, got %v", entry.Fields["service"])
	}
}

func TestTransform_UpperLevel(t *testing.T) {
	entry := baseEntry()
	Transform(entry, TransformConfig{UpperLevel: true})
	if entry.Level != strings.ToUpper("info") {
		t.Errorf("expected level INFO, got %s", entry.Level)
	}
}

func TestTransform_AddTimestamp(t *testing.T) {
	entry := baseEntry()
	Transform(entry, TransformConfig{AddTimestamp: true})
	if _, ok := entry.Fields["transform_ts"]; !ok {
		t.Error("expected transform_ts field to be added")
	}
}

func TestTransform_NilEntry(t *testing.T) {
	// Should not panic
	Transform(nil, TransformConfig{AddFields: map[string]string{"k": "v"}})
}

func TestTransform_NilFields_AddFields(t *testing.T) {
	entry := &LogEntry{Level: "warn", Message: "test", Fields: nil}
	Transform(entry, TransformConfig{
		AddFields: map[string]string{"region": "us-east-1"},
	})
	if entry.Fields["region"] != "us-east-1" {
		t.Errorf("expected region=us-east-1, got %v", entry.Fields["region"])
	}
}
