package pipeline

import (
	"strconv"
	"testing"
)

func TestSequencer_DefaultConfig(t *testing.T) {
	s := NewSequencer(SequenceConfig{})
	if s.cfg.StartAt != 1 {
		t.Fatalf("expected StartAt=1, got %d", s.cfg.StartAt)
	}
	if s.cfg.Field != "seq" {
		t.Fatalf("expected Field=seq, got %s", s.cfg.Field)
	}
}

func TestSequencer_StampIncrements(t *testing.T) {
	s := NewSequencer(SequenceConfig{})
	entry := LogEntry{Message: "hello", Fields: map[string]string{}}

	for i := int64(1); i <= 5; i++ {
		stamped := s.Stamp(entry)
		got := stamped.Fields["seq"]
		want := strconv.FormatInt(i, 10)
		if got != want {
			t.Fatalf("step %d: expected seq=%s, got %s", i, want, got)
		}
	}
}

func TestSequencer_CustomStartAndField(t *testing.T) {
	s := NewSequencer(SequenceConfig{StartAt: 100, Field: "n"})
	entry := LogEntry{Message: "msg", Fields: map[string]string{}}

	stamped := s.Stamp(entry)
	if stamped.Fields["n"] != "100" {
		t.Fatalf("expected n=100, got %s", stamped.Fields["n"])
	}
	if s.Current() != 101 {
		t.Fatalf("expected Current()=101, got %d", s.Current())
	}
}

func TestSequencer_DoesNotMutateOriginal(t *testing.T) {
	s := NewSequencer(SequenceConfig{})
	original := LogEntry{Message: "x", Fields: map[string]string{"a": "b"}}

	s.Stamp(original)

	if _, ok := original.Fields["seq"]; ok {
		t.Fatal("original entry was mutated")
	}
}

func TestSequencer_Reset(t *testing.T) {
	s := NewSequencer(SequenceConfig{StartAt: 10})
	entry := LogEntry{Message: "r", Fields: map[string]string{}}

	s.Stamp(entry)
	s.Stamp(entry)
	if s.Current() != 12 {
		t.Fatalf("expected 12 before reset, got %d", s.Current())
	}

	s.Reset()
	if s.Current() != 10 {
		t.Fatalf("expected 10 after reset, got %d", s.Current())
	}

	stamped := s.Stamp(entry)
	if stamped.Fields["seq"] != "10" {
		t.Fatalf("expected seq=10 after reset, got %s", stamped.Fields["seq"])
	}
}

func TestSequencer_NilFieldsMap(t *testing.T) {
	s := NewSequencer(SequenceConfig{})
	entry := LogEntry{Message: "no fields"}

	stamped := s.Stamp(entry)
	if stamped.Fields["seq"] != "1" {
		t.Fatalf("expected seq=1 on nil fields, got %s", stamped.Fields["seq"])
	}
}
