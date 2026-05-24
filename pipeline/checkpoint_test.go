package pipeline

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func tempCheckpointPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "checkpoint.json")
}

func TestCheckpoint_NewFileNotExist(t *testing.T) {
	cp, err := NewCheckpoint(tempCheckpointPath(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, ok := cp.Get("source-a")
	if ok {
		t.Error("expected no record for unknown source")
	}
}

func TestCheckpoint_SaveAndGet(t *testing.T) {
	cp, _ := NewCheckpoint(tempCheckpointPath(t))
	if err := cp.Save("source-a", 1024, 42); err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	rec, ok := cp.Get("source-a")
	if !ok {
		t.Fatal("expected record to exist")
	}
	if rec.Offset != 1024 || rec.LineCount != 42 {
		t.Errorf("unexpected record: %+v", rec)
	}
	if rec.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be set")
	}
}

func TestCheckpoint_PersistsAcrossReload(t *testing.T) {
	path := tempCheckpointPath(t)
	cp, _ := NewCheckpoint(path)
	_ = cp.Save("source-b", 512, 10)

	cp2, err := NewCheckpoint(path)
	if err != nil {
		t.Fatalf("reload failed: %v", err)
	}
	rec, ok := cp2.Get("source-b")
	if !ok {
		t.Fatal("record not found after reload")
	}
	if rec.Offset != 512 {
		t.Errorf("expected offset 512, got %d", rec.Offset)
	}
}

func TestCheckpoint_Delete(t *testing.T) {
	cp, _ := NewCheckpoint(tempCheckpointPath(t))
	_ = cp.Save("source-c", 100, 5)
	if err := cp.Delete("source-c"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	_, ok := cp.Get("source-c")
	if ok {
		t.Error("expected record to be deleted")
	}
}

func TestCheckpoint_UpdatedAtIsRecent(t *testing.T) {
	cp, _ := NewCheckpoint(tempCheckpointPath(t))
	before := time.Now().UTC()
	_ = cp.Save("source-d", 0, 0)
	after := time.Now().UTC()

	rec, _ := cp.Get("source-d")
	if rec.UpdatedAt.Before(before) || rec.UpdatedAt.After(after) {
		t.Errorf("UpdatedAt %v not in expected range [%v, %v]", rec.UpdatedAt, before, after)
	}
}

func TestCheckpoint_InvalidFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.json")
	_ = os.WriteFile(path, []byte("not-json{"), 0o644)
	_, err := NewCheckpoint(path)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}
