package pipeline

import (
	"strings"
	"time"
)

// NormalizeConfig holds configuration for the Normalizer.
type NormalizeConfig struct {
	// TimestampField is the field to normalize as a timestamp (default: "time").
	TimestampField string
	// TimestampFormats are the input formats to try when parsing timestamps.
	TimestampFormats []string
	// LevelField is the field to normalize to uppercase (default: "level").
	LevelField string
	// TrimFields lists field names whose string values should be whitespace-trimmed.
	TrimFields []string
}

// Normalizer standardizes common fields in a log entry.
type Normalizer struct {
	cfg NormalizeConfig
}

// NewNormalizer returns a Normalizer with the given config.
// Defaults are applied for empty fields.
func NewNormalizer(cfg NormalizeConfig) *Normalizer {
	if cfg.TimestampField == "" {
		cfg.TimestampField = "time"
	}
	if cfg.LevelField == "" {
		cfg.LevelField = "level"
	}
	if len(cfg.TimestampFormats) == 0 {
		cfg.TimestampFormats = []string{
			time.RFC3339,
			"2006-01-02T15:04:05",
			"2006-01-02 15:04:05",
			"02/Jan/2006:15:04:05 -0700",
		}
	}
	return &Normalizer{cfg: cfg}
}

// Normalize returns a new LogEntry with normalized fields.
// It does not mutate the original entry.
func (n *Normalizer) Normalize(entry LogEntry) LogEntry {
	fields := make(map[string]interface{}, len(entry.Fields))
	for k, v := range entry.Fields {
		fields[k] = v
	}

	// Normalize level to uppercase.
	if entry.Level != "" {
		fields[n.cfg.LevelField] = strings.ToUpper(strings.TrimSpace(entry.Level))
	}

	// Normalize timestamp field if present as a string.
	if raw, ok := fields[n.cfg.TimestampField]; ok {
		if s, ok := raw.(string); ok {
			for _, fmt := range n.cfg.TimestampFormats {
				if t, err := time.Parse(fmt, strings.TrimSpace(s)); err == nil {
					fields[n.cfg.TimestampField] = t.UTC().Format(time.RFC3339)
					break
				}
			}
		}
	}

	// Trim configured fields.
	for _, f := range n.cfg.TrimFields {
		if v, ok := fields[f]; ok {
			if s, ok := v.(string); ok {
				fields[f] = strings.TrimSpace(s)
			}
		}
	}

	return LogEntry{
		Timestamp: entry.Timestamp,
		Level:     strings.ToUpper(strings.TrimSpace(entry.Level)),
		Message:   entry.Message,
		Fields:    fields,
	}
}
