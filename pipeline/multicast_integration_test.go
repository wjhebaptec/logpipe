package pipeline

import (
	"sync"
	"testing"
	"time"
)

// TestMulticast_WithFilter verifies that subscribers can independently filter
// entries they receive without affecting other subscribers.
func TestMulticast_WithFilter(t *testing.T) {
	mc := NewMulticast(MulticastConfig{BufferSize: 16})

	allSub := mc.Subscribe()
	errorSub := mc.Subscribe()

	entries := []LogEntry{
		{Message: "ok", Level: "info"},
		{Message: "boom", Level: "error"},
		{Message: "warn msg", Level: "warn"},
	}
	for _, e := range entries {
		mc.Publish(e)
	}

	// allSub should receive all 3
	count := 0
	for i := 0; i < 3; i++ {
		select {
		case <-allSub:
			count++
		case <-time.After(100 * time.Millisecond):
			t.Fatalf("allSub timed out on entry %d", i)
		}
	}
	if count != 3 {
		t.Fatalf("allSub expected 3 entries, got %d", count)
	}

	// errorSub: manually filter only errors
	errorCount := 0
	for {
		select {
		case e := <-errorSub:
			if e.Level == "error" {
				errorCount++
			}
		default:
			goto done
		}
	}
done:
	if errorCount != 1 {
		t.Fatalf("expected 1 error entry, got %d", errorCount)
	}
}

// TestMulticast_ConcurrentPublish verifies thread-safety under concurrent
// publishers and a single subscriber.
func TestMulticast_ConcurrentPublish(t *testing.T) {
	mc := NewMulticast(MulticastConfig{BufferSize: 256})
	sub := mc.Subscribe()

	const goroutines = 10
	const perGoroutine = 5

	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < perGoroutine; j++ {
				mc.Publish(mcEntry("concurrent"))
			}
		}()
	}
	wg.Wait()

	received := 0
	for {
		select {
		case <-sub:
			received++
		default:
			goto check
		}
	}
check:
	expected := goroutines * perGoroutine
	if received != expected {
		t.Fatalf("expected %d received, got %d", expected, received)
	}
}
