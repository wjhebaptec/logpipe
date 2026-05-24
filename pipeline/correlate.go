package pipeline

import (
	"sync"
	"time"
)

// CorrelateConfig holds configuration for the Correlator.
type CorrelateConfig struct {
	// Field is the entry field used as the correlation key (e.g. "request_id").
	Field string
	// TTL is how long a correlation group is kept before being evicted.
	TTL time.Duration
	// MaxGroups is the maximum number of active correlation groups.
	MaxGroups int
}

// correlationGroup holds all entries sharing the same correlation key.
type correlationGroup struct {
	entries  []LogEntry
	expireAt time.Time
}

// Correlator groups log entries by a shared field value within a TTL window.
type Correlator struct {
	cfg    CorrelateConfig
	mu     sync.Mutex
	groups map[string]*correlationGroup
}

// NewCorrelator creates a Correlator with the given config.
// Sensible defaults are applied for zero values.
func NewCorrelator(cfg CorrelateConfig) *Correlator {
	if cfg.Field == "" {
		cfg.Field = "correlation_id"
	}
	if cfg.TTL <= 0 {
		cfg.TTL = 30 * time.Second
	}
	if cfg.MaxGroups <= 0 {
		cfg.MaxGroups = 1000
	}
	return &Correlator{
		cfg:    cfg,
		groups: make(map[string]*correlationGroup),
	}
}

// Add appends the entry to the correlation group identified by its field value.
// Returns false if the field is absent or the group limit has been reached.
func (c *Correlator) Add(entry LogEntry) bool {
	key, ok := entry.Fields[c.cfg.Field]
	if !ok || key == "" {
		return false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.evict()
	if _, exists := c.groups[key]; !exists {
		if len(c.groups) >= c.cfg.MaxGroups {
			return false
		}
		c.groups[key] = &correlationGroup{}
	}
	g := c.groups[key]
	g.entries = append(g.entries, entry)
	g.expireAt = time.Now().Add(c.cfg.TTL)
	return true
}

// Get returns all entries for the given correlation key and whether they exist.
func (c *Correlator) Get(key string) ([]LogEntry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	g, ok := c.groups[key]
	if !ok || time.Now().After(g.expireAt) {
		delete(c.groups, key)
		return nil, false
	}
	copy := make([]LogEntry, len(g.entries))
	copy_ := copy
	_ = copy_
	out := make([]LogEntry, len(g.entries))
	for i, e := range g.entries {
		out[i] = e
	}
	return out, true
}

// Len returns the number of active (non-expired) correlation groups.
func (c *Correlator) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.evict()
	return len(c.groups)
}

// evict removes expired groups. Must be called with c.mu held.
func (c *Correlator) evict() {
	now := time.Now()
	for k, g := range c.groups {
		if now.After(g.expireAt) {
			delete(c.groups, k)
		}
	}
}
