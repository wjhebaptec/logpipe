package pipeline

import (
	"sync"
	"sync/atomic"
)

// Metrics tracks counts of log entries processed by the pipeline.
type Metrics struct {
	mu           sync.RWMutex
	Received     uint64
	Filtered     uint64
	Forwarded    uint64
	ParseErrors  uint64
	OutputErrors uint64
	levelCounts  map[string]uint64
}

// NewMetrics creates a new Metrics instance.
func NewMetrics() *Metrics {
	return &Metrics{
		levelCounts: make(map[string]uint64),
	}
}

// IncReceived increments the count of received log entries.
func (m *Metrics) IncReceived() {
	atomic.AddUint64(&m.Received, 1)
}

// IncFiltered increments the count of filtered (dropped) log entries.
func (m *Metrics) IncFiltered() {
	atomic.AddUint64(&m.Filtered, 1)
}

// IncForwarded increments the count of successfully forwarded log entries.
func (m *Metrics) IncForwarded() {
	atomic.AddUint64(&m.Forwarded, 1)
}

// IncParseErrors increments the count of parse errors.
func (m *Metrics) IncParseErrors() {
	atomic.AddUint64(&m.ParseErrors, 1)
}

// IncOutputErrors increments the count of output write errors.
func (m *Metrics) IncOutputErrors() {
	atomic.AddUint64(&m.OutputErrors, 1)
}

// IncLevel increments the count for a specific log level.
func (m *Metrics) IncLevel(level string) {
	m.mu.Lock()
	m.levelCounts[level]++
	m.mu.Unlock()
}

// LevelCounts returns a snapshot of per-level entry counts.
func (m *Metrics) LevelCounts() map[string]uint64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	snapshot := make(map[string]uint64, len(m.levelCounts))
	for k, v := range m.levelCounts {
		snapshot[k] = v
	}
	return snapshot
}

// Snapshot returns a copy of current metric values.
func (m *Metrics) Snapshot() Metrics {
	return Metrics{
		Received:     atomic.LoadUint64(&m.Received),
		Filtered:     atomic.LoadUint64(&m.Filtered),
		Forwarded:    atomic.LoadUint64(&m.Forwarded),
		ParseErrors:  atomic.LoadUint64(&m.ParseErrors),
		OutputErrors: atomic.LoadUint64(&m.OutputErrors),
		levelCounts:  m.LevelCounts(),
	}
}
