package pipeline

import "testing"

func TestContainsString(t *testing.T) {
	tests := []struct {
		s, sub string
		want   bool
	}{
		{"hello world", "world", true},
		{"hello world", "missing", false},
		{"", "", true},
		{"abc", "", true},
	}
	for _, tt := range tests {
		got := containsString(tt.s, tt.sub)
		if got != tt.want {
			t.Errorf("containsString(%q, %q) = %v, want %v", tt.s, tt.sub, got, tt.want)
		}
	}
}

func TestNormalizeLevel(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"INFO", "info"},
		{" Error ", "error"},
		{"DEBUG", "debug"},
		{"", ""},
		{"WARN", "warn"},
	}
	for _, tt := range tests {
		got := normalizeLevel(tt.input)
		if got != tt.want {
			t.Errorf("normalizeLevel(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
