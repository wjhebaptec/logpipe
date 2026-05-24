package pipeline

import (
	"fmt"
	"time"
)

// Envelope wraps a log entry with routing metadata such as source, hop count,
// and a forwarding deadline. It allows pipeline stages to track provenance
// and enforce TTL-based expiry during multi-hop forwarding.
type Envelope struct {
	Entry     LogEntry
	Source    string
	Hops      int
	MaxHops   int
	Deadline  time.Time
	ForwardTo []string
}

// NewEnvelope creates an Envelope with the given source and TTL duration.
// MaxHops defaults to 10 if zero is provided.
func NewEnvelope(entry LogEntry, source string, ttl time.Duration, maxHops int) *Envelope {
	if maxHops <= 0 {
		maxHops = 10
	}
	return &Envelope{
		Entry:    entry,
		Source:   source,
		Hops:     0,
		MaxHops:  maxHops,
		Deadline: time.Now().Add(ttl),
	}
}

// Advance increments the hop counter and returns an error if the envelope
// has exceeded MaxHops or its deadline.
func (e *Envelope) Advance() error {
	if time.Now().After(e.Deadline) {
		return fmt.Errorf("envelope expired: deadline %s exceeded", e.Deadline.Format(time.RFC3339))
	}
	e.Hops++
	if e.Hops > e.MaxHops {
		return fmt.Errorf("envelope exceeded max hops (%d)", e.MaxHops)
	}
	return nil
}

// Expired returns true when the envelope's deadline has passed.
func (e *Envelope) Expired() bool {
	return time.Now().After(e.Deadline)
}

// AddForwardTarget appends a named output target to the forwarding list.
func (e *Envelope) AddForwardTarget(target string) {
	e.ForwardTo = append(e.ForwardTo, target)
}
