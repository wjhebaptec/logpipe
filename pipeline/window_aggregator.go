package pipeline

import "time"

// WindowStats holds aggregate statistics computed over a window snapshot.
type WindowStats struct {
	Total      int
	ByLevel    map[string]int
	WindowSize time.Duration
	Oldest     time.Time
	Newest     time.Time
}

// WindowAggregator computes statistics over a SlidingWindow.
type WindowAggregator struct {
	window *SlidingWindow
}

// NewWindowAggregator creates an aggregator backed by the given SlidingWindow.
func NewWindowAggregator(w *SlidingWindow) *WindowAggregator {
	return &WindowAggregator{window: w}
}

// Stats returns aggregate statistics for the current window contents.
func (a *WindowAggregator) Stats() WindowStats {
	snap := a.window.Snapshot()
	stats := WindowStats{
		ByLevel:    make(map[string]int),
		WindowSize: a.window.size,
	}
	if len(snap) == 0 {
		return stats
	}
	stats.Total = len(snap)
	stats.Oldest = snap[0].ReceivedAt
	stats.Newest = snap[len(snap)-1].ReceivedAt
	for _, we := range snap {
		lvl := normalizeLevel(we.Entry.Level)
		if lvl == "" {
			lvl = "unknown"
		}
		stats.ByLevel[lvl]++
	}
	return stats
}

// TopMessages returns up to n most recently added messages from the window.
func (a *WindowAggregator) TopMessages(n int) []string {
	snap := a.window.Snapshot()
	start := len(snap) - n
	if start < 0 {
		start = 0
	}
	result := make([]string, 0, len(snap)-start)
	for _, we := range snap[start:] {
		result = append(result, we.Entry.Message)
	}
	return result
}
