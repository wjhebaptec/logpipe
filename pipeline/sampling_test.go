package pipeline

import (
	"testing"
)

func TestSampler_RateOne_KeepsAll(t *testing.T) {
	s := NewSampler(1.0)
	for i := 0; i < 100; i++ {
		if !s.ShouldKeep() {
			t.Fatal("rate=1.0 should keep every entry")
		}
	}
}

func TestSampler_RateZero_DropsAll(t *testing.T) {
	s := NewSampler(0.0)
	for i := 0; i < 100; i++ {
		if s.ShouldKeep() {
			t.Fatal("rate=0.0 should drop every entry")
		}
	}
}

func TestSampler_Clamp_Above(t *testing.T) {
	s := NewSampler(5.0)
	if s.Rate() != 1.0 {
		t.Fatalf("expected rate clamped to 1.0, got %f", s.Rate())
	}
}

func TestSampler_Clamp_Below(t *testing.T) {
	s := NewSampler(-0.5)
	if s.Rate() != 0.0 {
		t.Fatalf("expected rate clamped to 0.0, got %f", s.Rate())
	}
}

func TestSampler_HalfRate_Approximate(t *testing.T) {
	s := NewSampler(0.5)
	kept := 0
	total := 10000
	for i := 0; i < total; i++ {
		if s.ShouldKeep() {
			kept++
		}
	}
	ratio := float64(kept) / float64(total)
	if ratio < 0.45 || ratio > 0.55 {
		t.Fatalf("expected ~50%% kept, got %.2f%%", ratio*100)
	}
}

func TestSampler_SampleEntries_RateOne(t *testing.T) {
	s := NewSampler(1.0)
	entries := []LogEntry{
		{Message: "a"},
		{Message: "b"},
		{Message: "c"},
	}
	out := s.SampleEntries(entries)
	if len(out) != len(entries) {
		t.Fatalf("expected %d entries, got %d", len(entries), len(out))
	}
}

func TestSampler_SampleEntries_RateZero(t *testing.T) {
	s := NewSampler(0.0)
	entries := []LogEntry{
		{Message: "a"},
		{Message: "b"},
	}
	out := s.SampleEntries(entries)
	if len(out) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(out))
	}
}

func TestSampler_SampleEntries_Empty(t *testing.T) {
	s := NewSampler(0.5)
	out := s.SampleEntries([]LogEntry{})
	if len(out) != 0 {
		t.Fatalf("expected 0 entries from empty input, got %d", len(out))
	}
}
