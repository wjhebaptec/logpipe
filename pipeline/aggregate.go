package pipeline

import (
	"sync"
	"time"
)

// AggregateConfig controls how log entries are grouped and summarized.
type AggregateConfig struct {
	// GroupBy is the field name used to group entries.
	GroupBy string
	// Window is how long to accumulate entries before emitting a summary.
	Window time.Duration
	// MaxGroups caps the number of distinct groups tracked at once.
	MaxGroups int
}

// AggregateResult holds a summary for one group key.
type AggregateResult struct {
	Key    string
	Count  int
	Levels map[string]int
	First  time.Time
	Last   time.Time
}

// Aggregator groups log entries by a field and periodically emits summaries.
type Aggregator struct {
	cfg    AggregateConfig
	mu     sync.Mutex
	groups map[string]*AggregateResult
}

// NewAggregator creates an Aggregator with the given config.
func NewAggregator(cfg AggregateConfig) *Aggregator {
	if cfg.GroupBy == "" {
		cfg.GroupBy = "source"
	}
	if cfg.Window <= 0 {
		cfg.Window = 30 * time.Second
	}
	if cfg.MaxGroups <= 0 {
		cfg.MaxGroups = 100
	}
	return &Aggregator{
		cfg:    cfg,
		groups: make(map[string]*AggregateResult),
	}
}

// Add records a log entry into the appropriate group.
// Returns false if the group limit has been reached and the key is new.
func (a *Aggregator) Add(entry LogEntry) bool {
	key, ok := entry.Fields[a.cfg.GroupBy]
	if !ok {
		key = "(unknown)"
	}
	ks, _ := key.(string)
	if ks == "" {
		ks = "(unknown)"
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if _, exists := a.groups[ks]; !exists {
		if len(a.groups) >= a.cfg.MaxGroups {
			return false
		}
		a.groups[ks] = &AggregateResult{
			Key:    ks,
			Levels: make(map[string]int),
			First:  entry.Timestamp,
		}
	}

	g := a.groups[ks]
	g.Count++
	g.Levels[normalizeLevel(entry.Level)]++
	g.Last = entry.Timestamp
	return true
}

// Flush returns all current group summaries and resets the internal state.
func (a *Aggregator) Flush() []AggregateResult {
	a.mu.Lock()
	defer a.mu.Unlock()

	results := make([]AggregateResult, 0, len(a.groups))
	for _, g := range a.groups {
		results = append(results, *g)
	}
	a.groups = make(map[string]*AggregateResult)
	return results
}

// Len returns the current number of tracked groups.
func (a *Aggregator) Len() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return len(a.groups)
}
