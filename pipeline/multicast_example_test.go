package pipeline_test

import (
	"fmt"
	"sync"

	"logpipe/pipeline"
)

// ExampleMulticast demonstrates broadcasting a log entry to multiple
// independent subscribers.
func ExampleMulticast() {
	mc := pipeline.NewMulticast(pipeline.MulticastConfig{BufferSize: 8})

	var wg sync.WaitGroup
	results := make([]string, 2)

	for i := 0; i < 2; i++ {
		sub := mc.Subscribe()
		idx := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			e := <-sub
			results[idx] = fmt.Sprintf("sub%d: %s", idx+1, e.Message)
		}()
	}

	mc.Publish(pipeline.LogEntry{Message: "hello world", Level: "info"})
	wg.Wait()
	mc.Close()

	for _, r := range results {
		fmt.Println(r)
	}
	// Unordered output:
	// sub1: hello world
	// sub2: hello world
}
