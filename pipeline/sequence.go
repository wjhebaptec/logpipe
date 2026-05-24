package pipeline

import "fmt"

// SequenceConfig holds configuration for a log entry sequencer.
type SequenceConfig struct {
	// StartAt sets the initial sequence number (default 1).
	StartAt int64
	// Field is the field name to inject into each entry (default "seq").
	Field string
}

// Sequencer assigns a monotonically increasing sequence number to each log entry.
type Sequencer struct {
	cfg     SequenceConfig
	current int64
}

// NewSequencer creates a new Sequencer with the given config.
// Zero values are replaced with sensible defaults.
func NewSequencer(cfg SequenceConfig) *Sequencer {
	if cfg.StartAt <= 0 {
		cfg.StartAt = 1
	}
	if cfg.Field == "" {
		cfg.Field = "seq"
	}
	return &Sequencer{
		cfg:     cfg,
		current: cfg.StartAt,
	}
}

// Stamp injects the next sequence number into the entry's Fields map
// and returns the annotated entry. The original entry is not mutated.
func (s *Sequencer) Stamp(entry LogEntry) LogEntry {
	fields := make(map[string]string, len(entry.Fields)+1)
	for k, v := range entry.Fields {
		fields[k] = v
	}
	fields[s.cfg.Field] = fmt.Sprintf("%d", s.current)
	s.current++
	entry.Fields = fields
	return entry
}

// Current returns the next sequence number that will be assigned
// without advancing the counter.
func (s *Sequencer) Current() int64 {
	return s.current
}

// Reset resets the counter back to the configured StartAt value.
func (s *Sequencer) Reset() {
	s.current = s.cfg.StartAt
}
