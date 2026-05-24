package pipeline

import (
	"context"
	"sync"
)

// FanoutConfig controls fanout behaviour.
type FanoutConfig struct {
	// Workers is the number of parallel goroutines used to dispatch entries.
	// Defaults to 2 if zero or negative.
	Workers int
	// BufferSize is the capacity of each worker's input channel.
	// Defaults to 64 if zero or negative.
	BufferSize int
}

// FanoutHandler is called for each log entry dispatched by the Fanout.
type FanoutHandler func(entry LogEntry)

// Fanout distributes incoming LogEntry values across a pool of worker
// goroutines, providing parallel dispatch to multiple downstream handlers.
type Fanout struct {
	workers int
	ch      chan LogEntry
	wg      sync.WaitGroup
	handler FanoutHandler
}

// NewFanout creates a Fanout and starts its worker pool. It runs until ctx is
// cancelled, after which all workers drain and exit cleanly.
func NewFanout(ctx context.Context, cfg FanoutConfig, handler FanoutHandler) *Fanout {
	if cfg.Workers <= 0 {
		cfg.Workers = 2
	}
	if cfg.BufferSize <= 0 {
		cfg.BufferSize = 64
	}

	f := &Fanout{
		workers: cfg.Workers,
		ch:      make(chan LogEntry, cfg.BufferSize),
		handler: handler,
	}

	for i := 0; i < cfg.Workers; i++ {
		f.wg.Add(1)
		go func() {
			defer f.wg.Done()
			for entry := range f.ch {
				f.handler(entry)
			}
		}()
	}

	go func() {
		<-ctx.Done()
		close(f.ch)
	}()

	return f
}

// Send enqueues an entry for dispatch. It returns false if the channel is full
// or already closed (context cancelled).
func (f *Fanout) Send(entry LogEntry) bool {
	select {
	case f.ch <- entry:
		return true
	default:
		return false
	}
}

// Wait blocks until all workers have finished processing.
func (f *Fanout) Wait() {
	f.wg.Wait()
}
