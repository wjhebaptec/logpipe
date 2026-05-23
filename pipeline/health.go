package pipeline

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
	"time"
)

// HealthStatus represents the current health of the pipeline.
type HealthStatus struct {
	Status    string            `json:"status"`
	Uptime    string            `json:"uptime"`
	Processed uint64            `json:"processed"`
	Dropped   uint64            `json:"dropped"`
	Outputs   map[string]string `json:"outputs"`
}

// HealthServer exposes an HTTP endpoint for pipeline health checks.
type HealthServer struct {
	metrics   *Metrics
	start     time.Time
	outputs   []string
	running   atomic.Bool
}

// NewHealthServer creates a HealthServer backed by the given Metrics.
func NewHealthServer(m *Metrics, outputNames []string) *HealthServer {
	return &HealthServer{
		metrics: m,
		start:   time.Now(),
		outputs: outputNames,
	}
}

// Handler returns an http.HandlerFunc that serves health status as JSON.
func (h *HealthServer) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snap := h.metrics.Snapshot()

		outputStatuses := make(map[string]string, len(h.outputs))
		for _, name := range h.outputs {
			outputStatuses[name] = "ok"
		}

		status := HealthStatus{
			Status:    "ok",
			Uptime:    time.Since(h.start).Round(time.Second).String(),
			Processed: snap.Processed,
			Dropped:   snap.Dropped,
			Outputs:   outputStatuses,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(status)
	}
}

// ListenAndServe starts the health HTTP server on the given address.
func (h *HealthServer) ListenAndServe(addr string) error {
	h.running.Store(true)
	mux := http.NewServeMux()
	mux.HandleFunc("/health", h.Handler())
	return http.ListenAndServe(addr, mux)
}
