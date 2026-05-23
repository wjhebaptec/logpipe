package pipeline

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// LogEntry represents a structured log entry parsed from input.
type LogEntry struct {
	Timestamp time.Time         `json:"timestamp"`
	Level     string            `json:"level"`
	Message   string            `json:"message"`
	Fields    map[string]string `json:"fields,omitempty"`
	Raw       string            `json:"-"`
}

// ParseEntry attempts to parse a raw log line into a LogEntry.
// It first tries JSON parsing, then falls back to a plain-text heuristic.
func ParseEntry(line string) (*LogEntry, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, fmt.Errorf("empty log line")
	}

	if strings.HasPrefix(line, "{") {
		return parseJSON(line)
	}

	return parsePlain(line)
}

func parseJSON(line string) (*LogEntry, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(line), &raw); err != nil {
		return nil, fmt.Errorf("json parse error: %w", err)
	}

	entry := &LogEntry{
		Raw:    line,
		Fields: make(map[string]string),
	}

	if v, ok := raw["message"].(string); ok {
		entry.Message = v
	} else if v, ok := raw["msg"].(string); ok {
		entry.Message = v
	}

	if v, ok := raw["level"].(string); ok {
		entry.Level = strings.ToUpper(v)
	}

	if v, ok := raw["timestamp"].(string); ok {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			entry.Timestamp = t
		}
	}

	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	for k, v := range raw {
		if k == "message" || k == "msg" || k == "level" || k == "timestamp" {
			continue
		}
		entry.Fields[k] = fmt.Sprintf("%v", v)
	}

	return entry, nil
}

func parsePlain(line string) (*LogEntry, error) {
	entry := &LogEntry{
		Raw:       line,
		Message:   line,
		Timestamp: time.Now(),
		Level:     "INFO",
		Fields:    make(map[string]string),
	}

	upper := strings.ToUpper(line)
	for _, lvl := range []string{"ERROR", "WARN", "WARNING", "INFO", "DEBUG", "TRACE", "FATAL"} {
		if strings.Contains(upper, lvl) {
			entry.Level = strings.Replace(lvl, "WARNING", "WARN", 1)
			break
		}
	}

	return entry, nil
}
