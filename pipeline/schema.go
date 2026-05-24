package pipeline

import (
	"fmt"
	"regexp"
)

// SchemaRule defines a validation rule for a log entry field.
type SchemaRule struct {
	Field    string
	Required bool
	Pattern  string // optional regex pattern
	compiled *regexp.Regexp
}

// SchemaValidator validates log entries against a set of schema rules.
type SchemaValidator struct {
	rules []SchemaRule
}

// NewSchemaValidator creates a SchemaValidator from the provided rules.
// Returns an error if any rule contains an invalid regex pattern.
func NewSchemaValidator(rules []SchemaRule) (*SchemaValidator, error) {
	compiled := make([]SchemaRule, len(rules))
	for i, r := range rules {
		compiled[i] = r
		if r.Pattern != "" {
			re, err := regexp.Compile(r.Pattern)
			if err != nil {
				return nil, fmt.Errorf("schema rule %q: invalid pattern %q: %w", r.Field, r.Pattern, err)
			}
			compiled[i].compiled = re
		}
	}
	return &SchemaValidator{rules: compiled}, nil
}

// Validate checks a LogEntry against the schema rules.
// Returns a slice of validation error messages (empty if valid).
func (sv *SchemaValidator) Validate(entry LogEntry) []string {
	var errs []string
	for _, rule := range sv.rules {
		val, ok := entry.Fields[rule.Field]
		if !ok || val == "" {
			if rule.Required {
				errs = append(errs, fmt.Sprintf("missing required field %q", rule.Field))
			}
			continue
		}
		if rule.compiled != nil && !rule.compiled.MatchString(val) {
			errs = append(errs, fmt.Sprintf("field %q value %q does not match pattern %q", rule.Field, val, rule.Pattern))
		}
	}
	return errs
}

// IsValid returns true if the entry passes all schema rules.
func (sv *SchemaValidator) IsValid(entry LogEntry) bool {
	return len(sv.Validate(entry)) == 0
}
