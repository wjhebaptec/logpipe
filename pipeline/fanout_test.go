package pipeline

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestFanout_DispatchesAllEntries(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var count int64
	f := NewFanout(ctx, FanoutConfig{Workers: 4, BufferSize: 32}, func(e LogEntry) {
		atomic.AddInt64(&count, 1)
	})

	for i := 0; i < 20; i++ {
		f.Send(LogEntry{Message: "hello"})
	}

	cancel()
	f.Wait()

	if got := atomic.LoadInt64(&count); got != 20 {
		t.Fatalf("expected 20 dispatched entries, got %d", got)
	}
}

func TestFanout_DefaultConfig(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var called int64
	f := NewFanout(ctx, FanoutConfig{}, func(e LogEntry) {
		atomic.AddInt64(&called, 1)
	})

	f.Send(LogEntry{Message: "test"})
	cancel()
	f.Wait()

	if atomic.LoadInt64(&called) != 1 {
		t.Fatal("expected handler to be called once")
	}
}

func TestFanout_SendReturnsFalseWhenFull(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// block handler so buffer fills up
	block := make(chan struct{})
	f := NewFanout(ctx, FanoutConfig{Workers: 1, BufferSize: 2}, func(e LogEntry) {
		<-block
	})

	// fill the buffer
	f.Send(LogEntry{Message: "a"})
	f.Send(LogEntry{Message: "b"})
	f.Send(LogEntry{Message: "c"})

	// next send should fail
	if f.Send(LogEntry{Message: "overflow"}) {
		t.Error("expected Send to return false on full buffer")
	}

	close(block)
	cancel()
	f.Wait()
}

func TestFanout_ConcurrentSenders(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var mu sync.Mutex
	received := make([]LogEntry, 0, 50)

	f := NewFanout(ctx, FanoutConfig{Workers: 4, BufferSize: 128}, func(e LogEntry) {
		mu.Lock()
		received = append(received, e)
		mu.Unlock()
	})

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				f.Send(LogEntry{Message: "msg"})
			}
		}()
	}
	wg.Wait()

	cancel()
	f.Wait()

	mu.Lock()
	defer mu.Unlock()
	if len(received) != 50 {
		t.Fatalf("expected 50 entries, got %d", len(received))
	}
}

func TestFanout_WaitBlocksUntilDone(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	var done int64
	f := NewFanout(ctx, FanoutConfig{Workers: 2, BufferSize: 8}, func(e LogEntry) {
		time.Sleep(10 * time.Millisecond)
		atomic.AddInt64(&done, 1)
	})

	f.Send(LogEntry{Message: "slow"})
	cancel()
	f.Wait()

	if atomic.LoadInt64(&done) != 1 {
		t.Fatal("Wait returned before handler completed")
	}
}
