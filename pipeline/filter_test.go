package pipeline

import (
	"testing"

	"github.com/yourorg/logpipe/config"
)

func TestFilter_NoFilters(t *testing.T) {
	entry := LogEntry{Level: "info", Message: "hello"}
	if !Filter(entry, nil) {
		t.Error("expected entry to pass with no filters")
	}
}

func TestFilter_LevelMatch(t *testing.T) {
	entry := LogEntry{Level: "error", Message: "something failed"}
	filters := []config.Filter{{Level: "error"}}
	if !Filter(entry, filters) {
		t.Error("expected error-level entry to match")
	}
}

func TestFilter_LevelNoMatch(t *testing.T) {
	entry := LogEntry{Level: "debug", Message: "verbose output"}
	filters := []config.Filter{{Level: "error"}}
	if Filter(entry, filters) {
		t.Error("expected debug entry not to match error filter")
	}
}

func TestFilter_ContainsMatch(t *testing.T) {
	entry := LogEntry{Level: "info", Message: "user login successful"}
	filters := []config.Filter{{Contains: "login"}}
	if !Filter(entry, filters) {
		t.Error("expected entry to match contains filter")
	}
}

func TestFilter_FieldsMatch(t *testing.T) {
	entry := LogEntry{
		Level:   "warn",
		Message: "disk usage high",
		Fields:  map[string]string{"host": "web-01"},
	}
	filters := []config.Filter{{Fields: map[string]string{"host": "web-01"}}}
	if !Filter(entry, filters) {
		t.Error("expected entry to match field filter")
	}
}

func TestFilter_FieldsNoMatch(t *testing.T) {
	entry := LogEntry{
		Level:   "warn",
		Message: "disk usage high",
		Fields:  map[string]string{"host": "web-02"},
	}
	filters := []config.Filter{{Fields: map[string]string{"host": "web-01"}}}
	if Filter(entry, filters) {
		t.Error("expected entry not to match field filter with wrong host")
	}
}

func TestFilter_MultipleFiltersAnyMatch(t *testing.T) {
	entry := LogEntry{Level: "info", Message: "startup complete"}
	filters := []config.Filter{
		{Level: "error"},
		{Contains: "startup"},
	}
	if !Filter(entry, filters) {
		t.Error("expected entry to match at least one filter")
	}
}
