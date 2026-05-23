package pipeline

import (
	"context"
	"log"
	"sync"

	"github.com/user/logpipe/config"
)

// Pipeline wires together input readers, a filter, and a router.
type Pipeline struct {
	cfg    *config.Config
	entries chan LogEntry
}

// New creates a Pipeline from the given config.
func New(cfg *config.Config) *Pipeline {
	return &Pipeline{
		cfg:    cfg,
		entries: make(chan LogEntry, 256),
	}
}

// Run starts all inputs and routes entries until ctx is cancelled.
func (p *Pipeline) Run(ctx context.Context) error {
	outputs, err := buildOutputs(p.cfg)
	if err != nil {
		return err
	}

	router := NewRouter(p.cfg.Outputs, outputs)

	var wg sync.WaitGroup
	for _, inp := range p.cfg.Inputs {
		inp := inp
		wg.Add(1)
		go func() {
			defer wg.Done()
			log.Printf("[pipeline] input %q starting (type=%s)", inp.Name, inp.Type)
			ir := NewInputReader(inp.Reader(), inp.Format, p.entries)
			if err := ir.Run(ctx); err != nil {
				log.Printf("[pipeline] input %q stopped: %v", inp.Name, err)
			}
		}()
	}

	go func() {
		wg.Wait()
		close(p.entries)
	}()

	for entry := range p.entries {
		if !Filter(entry, p.cfg.Filters) {
			continue
		}
		if err := router.Route(entry); err != nil {
			log.Printf("[pipeline] route error: %v", err)
		}
	}
	return nil
}

func buildOutputs(cfg *config.Config) (map[string]*OutputWriter, error) {
	writers := make(map[string]*OutputWriter, len(cfg.Outputs))
	for _, o := range cfg.Outputs {
		ow, err := NewFileOutput(o.Name, o.Path, o.Format)
		if err != nil {
			return nil, err
		}
		writers[o.Name] = ow
	}
	return writers, nil
}
