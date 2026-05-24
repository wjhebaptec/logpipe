package pipeline

import (
	"testing"
)

// TestCheckpoint_MultipleSourcesIndependent verifies that saving one source
// does not affect another source's checkpoint.
func TestCheckpoint_MultipleSourcesIndependent(t *testing.T) {
	cp, _ := NewCheckpoint(tempCheckpointPath(t))

	_ = cp.Save("source-x", 100, 10)
	_ = cp.Save("source-y", 200, 20)

	rx, _ := cp.Get("source-x")
	ry, _ := cp.Get("source-y")

	if rx.Offset != 100 {
		t.Errorf("source-x offset: want 100, got %d", rx.Offset)
	}
	if ry.Offset != 200 {
		t.Errorf("source-y offset: want 200, got %d", ry.Offset)
	}
}

// TestCheckpoint_OverwriteUpdatesRecord verifies that saving the same source
// twice overwrites the previous record.
func TestCheckpoint_OverwriteUpdatesRecord(t *testing.T) {
	cp, _ := NewCheckpoint(tempCheckpointPath(t))

	_ = cp.Save("source-z", 50, 5)
	_ = cp.Save("source-z", 150, 15)

	rec, ok := cp.Get("source-z")
	if !ok {
		t.Fatal("expected record")
	}
	if rec.Offset != 150 || rec.LineCount != 15 {
		t.Errorf("expected offset=150 lineCount=15, got %+v", rec)
	}
}

// TestCheckpoint_ReloadAfterDelete verifies that after deleting a source and
// reloading, the record is absent.
func TestCheckpoint_ReloadAfterDelete(t *testing.T) {
	path := tempCheckpointPath(t)
	cp, _ := NewCheckpoint(path)
	_ = cp.Save("source-del", 99, 9)
	_ = cp.Delete("source-del")

	cp2, err := NewCheckpoint(path)
	if err != nil {
		t.Fatalf("reload error: %v", err)
	}
	_, ok := cp2.Get("source-del")
	if ok {
		t.Error("expected deleted record to be absent after reload")
	}
}
