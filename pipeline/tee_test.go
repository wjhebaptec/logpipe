package pipeline

import (
	"sync"
	"testing"
	"time"
)

func teeEntry(msg string) LogEntry {
	return LogEntry{Message: msg, Level: "info"}
}

func TestTee_NoBranches(t *testing.T) {
	tee := NewTee(TeeConfig{})
	// Should not panic when no branches are registered.
	tee.Send(teeEntry("hello"))
	tee.Close()
}

func TestTee_SingleBranch(t *testing.T) {
	tee := NewTee(TeeConfig{BufferSize: 8})
	ch := tee.Branch()

	tee.Send(teeEntry("msg1"))
	tee.Close()

	entry := <-ch
	if entry.Message != "msg1" {
		t.Fatalf("expected msg1, got %q", entry.Message)
	}
}

func TestTee_MultipleBranchesReceiveAll(t *testing.T) {
	tee := NewTee(TeeConfig{BufferSize: 16})
	ch1 := tee.Branch()
	ch2 := tee.Branch()
	ch3 := tee.Branch()

	msgs := []string{"a", "b", "c"}
	for _, m := range msgs {
		tee.Send(teeEntry(m))
	}
	tee.Close()

	for _, ch := range []<-chan LogEntry{ch1, ch2, ch3} {
		var got []string
		for e := range ch {
			got = append(got, e.Message)
		}
		if len(got) != len(msgs) {
			t.Fatalf("expected %d entries, got %d", len(msgs), len(got))
		}
		for i, m := range msgs {
			if got[i] != m {
				t.Errorf("branch entry[%d]: expected %q, got %q", i, m, got[i])
			}
		}
	}
}

func TestTee_FullBranchDropsEntry(t *testing.T) {
	// Buffer of 1 so the second send overflows.
	tee := NewTee(TeeConfig{BufferSize: 1})
	ch := tee.Branch()

	tee.Send(teeEntry("first"))
	tee.Send(teeEntry("dropped")) // should not block
	tee.Close()

	got := <-ch
	if got.Message != "first" {
		t.Fatalf("expected first, got %q", got.Message)
	}
}

func TestTee_ConcurrentSend(t *testing.T) {
	tee := NewTee(TeeConfig{BufferSize: 128})
	ch := tee.Branch()

	const n = 50
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tee.Send(teeEntry("concurrent"))
		}()
	}
	wg.Wait()
	tee.Close()

	count := 0
	timeout := time.After(time.Second)
	for {
		select {
		case _, ok := <-ch:
			if !ok {
				goto done
			}
			count++
		case <-timeout:
			t.Fatal("timed out draining branch")
		}
	}
done:
	if count > n {
		t.Errorf("received more entries (%d) than sent (%d)", count, n)
	}
}
