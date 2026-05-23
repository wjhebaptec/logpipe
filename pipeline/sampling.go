package pipeline

import (
	"math/rand"
	"sync"
)

// Sampler drops log entries probabilistically based on a configured rate.
// A rate of 1.0 keeps all entries; 0.1 keeps roughly 10%.
type Sampler struct {
	mu   sync.Mutex
	rate float64
	rng  *rand.Rand
}

// NewSampler creates a Sampler with the given keep rate (0.0–1.0).
// Values outside that range are clamped.
func NewSampler(rate float64) *Sampler {
	if rate < 0 {
		rate = 0
	}
	if rate > 1 {
		rate = 1
	}
	return &Sampler{
		rate: rate,
		rng:  rand.New(rand.NewSource(rand.Int63())),
	}
}

// ShouldKeep returns true if the entry should be forwarded downstream.
func (s *Sampler) ShouldKeep() bool {
	if s.rate >= 1.0 {
		return true
	}
	if s.rate <= 0.0 {
		return false
	}
	s.mu.Lock()
	v := s.rng.Float64()
	s.mu.Unlock()
	return v < s.rate
}

// Rate returns the configured sampling rate.
func (s *Sampler) Rate() float64 {
	return s.rate
}

// SampleEntries filters a slice of LogEntries according to the sampling rate.
func (s *Sampler) SampleEntries(entries []LogEntry) []LogEntry {
	if s.rate >= 1.0 {
		return entries
	}
	out := make([]LogEntry, 0, len(entries))
	for _, e := range entries {
		if s.ShouldKeep() {
			out = append(out, e)
		}
	}
	return out
}
