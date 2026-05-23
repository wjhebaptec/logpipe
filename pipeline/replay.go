package pipeline

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"time"
)

// ReplayOptions controls how a replay session behaves.
type ReplayOptions struct {
	// RateLimit is the maximum number of log entries to emit per second.
	// Zero means no rate limiting.
	RateLimit int
	// Filter, if non-nil, is applied before forwarding each entry.
	Filter *FilterConfig
}

// FilterConfig holds the filter criteria used during replay.
type FilterConfig struct {
	Level    string
	Contains string
}

// ReplayFile reads a previously recorded log file and re-emits entries to the
// provided writer, optionally rate-limited and filtered.
func ReplayFile(ctx context.Context, path string, w io.Writer, opts ReplayOptions) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, fmt.Errorf("replay: open %q: %w", path, err)
	}
	defer f.Close()
	return replayReader(ctx, f, w, opts)
}

func replayReader(ctx context.Context, r io.Reader, w io.Writer, opts ReplayOptions) (int, error) {
	var (
		scanner  = bufio.NewScanner(r)
		count    int
		interval time.Duration
	)
	if opts.RateLimit > 0 {
		interval = time.Second / time.Duration(opts.RateLimit)
	}

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return count, ctx.Err()
		default:
		}

		line := scanner.Text()
		if line == "" {
			continue
		}

		entry, err := ParseEntry(line)
		if err != nil {
			continue
		}

		if opts.Filter != nil && !replayMatchesFilter(entry, opts.Filter) {
			continue
		}

		fmt.Fprintln(w, formatEntry(entry))
		count++

		if interval > 0 {
			select {
			case <-ctx.Done():
				return count, ctx.Err()
			case <-time.After(interval):
			}
		}
	}
	return count, scanner.Err()
}

func replayMatchesFilter(e LogEntry, f *FilterConfig) bool {
	if f.Level != "" && e.Level != f.Level {
		return false
	}
	if f.Contains != "" && !containsString(e.Message, f.Contains) {
		return false
	}
	return true
}
