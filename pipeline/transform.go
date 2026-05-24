package pipeline

import (
	"strings"
	"time"
)

// TransformConfig defines optional transformations to apply to log entries.
type TransformConfig struct {
	AddFields    map[string]string `yaml:"add_fields"`
	RemoveFields []string          `yaml:"remove_fields"`
	RenameFields map[string]string `yaml:"rename_fields"`
	UpperLevel   bool              `yaml:"upper_level"`
	AddTimestamp bool              `yaml:"add_timestamp"`
}

// LogEntry represents a parsed log entry (imported from parser.go context).
// Transform applies the configured transformations to a log entry in-place.
func Transform(entry *LogEntry, cfg TransformConfig) {
	if entry == nil {
		return
	}

	// Add static fields
	for k, v := range cfg.AddFields {
		ensureFields(entry)
		entry.Fields[k] = v
	}

	// Remove fields
	for _, key := range cfg.RemoveFields {
		delete(entry.Fields, key)
	}

	// Rename fields
	for oldKey, newKey := range cfg.RenameFields {
		if val, ok := entry.Fields[oldKey]; ok {
			ensureFields(entry)
			entry.Fields[newKey] = val
			delete(entry.Fields, oldKey)
		}
	}

	// Normalize level to uppercase
	if cfg.UpperLevel {
		entry.Level = strings.ToUpper(entry.Level)
	}

	// Inject timestamp if missing or forced
	if cfg.AddTimestamp {
		ensureFields(entry)
		if _, exists := entry.Fields["transform_ts"]; !exists {
			entry.Fields["transform_ts"] = time.Now().UTC().Format(time.RFC3339)
		}
	}
}

// ensureFields initialises the Fields map on entry if it is nil.
func ensureFields(entry *LogEntry) {
	if entry.Fields == nil {
		entry.Fields = make(map[string]interface{})
	}
}
