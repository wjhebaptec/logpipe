package pipeline

import "strings"

// LabelConfig defines rules for attaching static or dynamic labels to log entries.
type LabelConfig struct {
	// Static labels always added to every entry.
	Static map[string]string `yaml:"static"`
	// CopyFrom copies an existing field value into a new label key.
	// Key is the destination label, value is the source field name.
	CopyFrom map[string]string `yaml:"copy_from"`
	// Prefix is prepended to all label keys.
	Prefix string `yaml:"prefix"`
}

// Labeler attaches labels to log entries.
type Labeler struct {
	cfg LabelConfig
}

// NewLabeler creates a Labeler from the given config.
func NewLabeler(cfg LabelConfig) *Labeler {
	return &Labeler{cfg: cfg}
}

// Apply returns a new LogEntry with labels merged into its Fields.
// Existing fields are not overwritten by static labels.
func (l *Labeler) Apply(entry LogEntry) LogEntry {
	out := entry
	out.Fields = make(map[string]string, len(entry.Fields))
	for k, v := range entry.Fields {
		out.Fields[k] = v
	}

	prefix := l.cfg.Prefix

	// Apply copy_from first so static can override if needed.
	for dest, src := range l.cfg.CopyFrom {
		if val, ok := entry.Fields[src]; ok {
			key := labelKey(prefix, dest)
			if _, exists := out.Fields[key]; !exists {
				out.Fields[key] = val
			}
		}
	}

	for k, v := range l.cfg.Static {
		key := labelKey(prefix, k)
		out.Fields[key] = v
	}

	return out
}

func labelKey(prefix, key string) string {
	if prefix == "" {
		return key
	}
	return strings.TrimRight(prefix, "_") + "_" + key
}
