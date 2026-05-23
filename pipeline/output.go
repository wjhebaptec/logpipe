package pipeline

import (
	"fmt"
	"io"
	"os"
)

// OutputWriter wraps an io.Writer with a name and optional format.
type OutputWriter struct {
	Name   string
	Format string
	Writer io.Writer
}

// NewFileOutput creates an OutputWriter that writes to a file path.
// Pass "-" to write to stdout.
func NewFileOutput(name, path, format string) (*OutputWriter, error) {
	var w io.Writer
	if path == "-" {
		w = os.Stdout
	} else {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("output %q: open file %q: %w", name, path, err)
		}
		w = f
	}
	return &OutputWriter{Name: name, Format: format, Writer: w}, nil
}

// Write formats and writes a LogEntry to the underlying writer.
func (ow *OutputWriter) Write(entry LogEntry) error {
	line := formatEntry(entry, ow.Format)
	_, err := fmt.Fprintln(ow.Writer, line)
	return err
}
