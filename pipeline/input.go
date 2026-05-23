package pipeline

import (
	"bufio"
	"context"
	"io"
	"log"
)

// InputReader reads log lines from an io.Reader and sends parsed entries to a channel.
type InputReader struct {
	reader  io.Reader
	format  string
	output  chan<- LogEntry
}

// NewInputReader creates a new InputReader.
func NewInputReader(r io.Reader, format string, out chan<- LogEntry) *InputReader {
	return &InputReader{
		reader: r,
		format: format,
		output: out,
	}
}

// Run starts reading lines from the reader until EOF or context cancellation.
func (ir *InputReader) Run(ctx context.Context) error {
	scanner := bufio.NewScanner(ir.reader)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return err
			}
			return io.EOF
		}

		line := scanner.Text()
		entry, err := ParseEntry(line, ir.format)
		if err != nil {
			log.Printf("[input] failed to parse line: %v", err)
			continue
		}
		if entry == nil {
			continue
		}

		select {
		case ir.output <- *entry:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
