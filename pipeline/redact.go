package pipeline

import (
	"regexp"
	"strings"
)

// RedactRule defines a single field redaction rule.
type RedactRule struct {
	Field   string
	Pattern string
	Replace string
	re      *regexp.Regexp
}

// Redactor applies redaction rules to log entry fields.
type Redactor struct {
	rules []RedactRule
}

// NewRedactor creates a Redactor from a list of rules.
// Rules with invalid patterns are skipped.
func NewRedactor(rules []RedactRule) *Redactor {
	compiled := make([]RedactRule, 0, len(rules))
	for _, r := range rules {
		if r.Replace == "" {
			r.Replace = "[REDACTED]"
		}
		if r.Pattern != "" {
			re, err := regexp.Compile(r.Pattern)
			if err != nil {
				continue
			}
			r.re = re
		}
		compiled = append(compiled, r)
	}
	return &Redactor{rules: compiled}
}

// Redact applies all redaction rules to the given log entry, returning a modified copy.
func (r *Redactor) Redact(entry LogEntry) LogEntry {
	if len(r.rules) == 0 {
		return entry
	}

	// Copy fields map to avoid mutating the original.
	newFields := make(map[string]interface{}, len(entry.Fields))
	for k, v := range entry.Fields {
		newFields[k] = v
	}
	entry.Fields = newFields

	for _, rule := range r.rules {
		if rule.Field == "message" {
			entry.Message = redactString(entry.Message, rule)
			continue
		}
		if val, ok := entry.Fields[rule.Field]; ok {
			if s, ok := val.(string); ok {
				entry.Fields[rule.Field] = redactString(s, rule)
			}
		}
	}
	return entry
}

func redactString(s string, rule RedactRule) string {
	if rule.re != nil {
		return rule.re.ReplaceAllString(s, rule.Replace)
	}
	// Full field replacement when no pattern is given.
	if strings.TrimSpace(s) != "" {
		return rule.Replace
	}
	return s
}
