package pipeline

import (
	"strings"

	"github.com/yourorg/logpipe/config"
)

// LogEntry represents a single structured log record.
type LogEntry struct {
	Level   string            `json:"level"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

// Filter evaluates whether a LogEntry matches the given filter rules.
// An empty filter list means all entries pass.
func Filter(entry LogEntry, filters []config.Filter) bool {
	if len(filters) == 0 {
		return true
	}
	for _, f := range filters {
		if matchesFilter(entry, f) {
			return true
		}
	}
	return false
}

func matchesFilter(entry LogEntry, f config.Filter) bool {
	if f.Level != "" && !strings.EqualFold(entry.Level, f.Level) {
		return false
	}
	if f.Contains != "" && !strings.Contains(entry.Message, f.Contains) {
		return false
	}
	for k, v := range f.Fields {
		if entry.Fields[k] != v {
			return false
		}
	}
	return true
}
