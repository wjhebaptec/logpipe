package pipeline

import (
	"fmt"
	"io"
	"strings"

	"github.com/user/logpipe/config"
)

// Router forwards log entries to configured outputs based on matching rules.
type Router struct {
	rules   []config.Output
	writers map[string]io.Writer
}

// NewRouter creates a Router with the given output rules and writer map.
// The writers map keys should match output names defined in config.
func NewRouter(outputs []config.Output, writers map[string]io.Writer) *Router {
	return &Router{
		rules:   outputs,
		writers: writers,
	}
}

// Route sends the log entry to all matching outputs.
// Returns an error if any write fails.
func (r *Router) Route(entry map[string]interface{}) error {
	for _, output := range r.rules {
		if !matchesOutput(entry, output) {
			continue
		}
		w, ok := r.writers[output.Name]
		if !ok {
			return fmt.Errorf("router: no writer registered for output %q", output.Name)
		}
		line := formatEntry(entry)
		if _, err := fmt.Fprintln(w, line); err != nil {
			return fmt.Errorf("router: write to %q failed: %w", output.Name, err)
		}
	}
	return nil
}

// matchesOutput returns true if the entry satisfies the output's level filter.
func matchesOutput(entry map[string]interface{}, output config.Output) bool {
	if output.Level == "" {
		return true
	}
	level, ok := entry["level"].(string)
	if !ok {
		return false
	}
	return strings.EqualFold(level, output.Level)
}

// formatEntry serialises a log entry as a simple key=value string.
func formatEntry(entry map[string]interface{}) string {
	var parts []string
	for k, v := range entry {
		parts = append(parts, fmt.Sprintf("%s=%v", k, v))
	}
	return strings.Join(parts, " ")
}
