package pipeline

import "strings"

// containsString is a thin wrapper around strings.Contains used across the
// pipeline package so callers don't need to import "strings" directly.
func containsString(s, substr string) bool {
	return strings.Contains(s, substr)
}

// normalizeLevel lowercases a level string and trims surrounding whitespace.
func normalizeLevel(level string) string {
	return strings.ToLower(strings.TrimSpace(level))
}
