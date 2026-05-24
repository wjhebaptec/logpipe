package pipeline

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// Checkpoint tracks the last successfully processed position/offset
// for a named input source, enabling resume-on-restart behaviour.
type Checkpoint struct {
	mu      sync.Mutex
	path    string
	records map[string]CheckpointRecord
}

// CheckpointRecord holds the persisted state for a single source.
type CheckpointRecord struct {
	Source    string    `json:"source"`
	Offset    int64     `json:"offset"`
	LineCount int64     `json:"line_count"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewCheckpoint loads an existing checkpoint file or starts fresh.
func NewCheckpoint(path string) (*Checkpoint, error) {
	cp := &Checkpoint{
		path:    path,
		records: make(map[string]CheckpointRecord),
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cp, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(data, &cp.records); err != nil {
		return nil, err
	}
	return cp, nil
}

// Get returns the checkpoint record for a source, and whether it exists.
func (c *Checkpoint) Get(source string) (CheckpointRecord, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	rec, ok := c.records[source]
	return rec, ok
}

// Save updates the checkpoint for a source and flushes to disk.
func (c *Checkpoint) Save(source string, offset, lineCount int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.records[source] = CheckpointRecord{
		Source:    source,
		Offset:    offset,
		LineCount: lineCount,
		UpdatedAt: time.Now().UTC(),
	}
	return c.flush()
}

// Delete removes a checkpoint record and flushes to disk.
func (c *Checkpoint) Delete(source string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.records, source)
	return c.flush()
}

func (c *Checkpoint) flush() error {
	data, err := json.MarshalIndent(c.records, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(c.path, data, 0o644)
}
