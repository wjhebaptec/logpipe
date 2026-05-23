package pipeline

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler_StatusOK(t *testing.T) {
	m := NewMetrics()
	m.IncProcessed()
	m.IncProcessed()
	m.IncDropped()

	h := NewHealthServer(m, []string{"stdout", "file"})
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	h.Handler()(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var status HealthStatus
	if err := json.NewDecoder(rec.Body).Decode(&status); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if status.Status != "ok" {
		t.Errorf("expected status 'ok', got %q", status.Status)
	}
	if status.Processed != 2 {
		t.Errorf("expected processed=2, got %d", status.Processed)
	}
	if status.Dropped != 1 {
		t.Errorf("expected dropped=1, got %d", status.Dropped)
	}
}

func TestHealthHandler_OutputsReported(t *testing.T) {
	m := NewMetrics()
	outputs := []string{"stdout", "errors.log"}
	h := NewHealthServer(m, outputs)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	h.Handler()(rec, req)

	var status HealthStatus
	_ = json.NewDecoder(rec.Body).Decode(&status)

	for _, name := range outputs {
		if v, ok := status.Outputs[name]; !ok || v != "ok" {
			t.Errorf("expected output %q to be 'ok', got %q", name, v)
		}
	}
}

func TestHealthHandler_ContentTypeJSON(t *testing.T) {
	m := NewMetrics()
	h := NewHealthServer(m, nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	h.Handler()(rec, req)

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}
}

func TestHealthHandler_UptimeNonEmpty(t *testing.T) {
	m := NewMetrics()
	h := NewHealthServer(m, nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	h.Handler()(rec, req)

	var status HealthStatus
	_ = json.NewDecoder(rec.Body).Decode(&status)

	if status.Uptime == "" {
		t.Error("expected non-empty uptime")
	}
}
