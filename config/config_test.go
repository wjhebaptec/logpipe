package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yourorg/logpipe/config"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("writeTemp: %v", err)
	}
	return p
}

func TestLoad_Valid(t *testing.T) {
	yaml := `
inputs:
  - name: app-logs
    type: file
    path: /var/log/app.log
filters:
  - name: errors-only
    field: level
    match: error
outputs:
  - name: stdout-out
    type: stdout
    format: json
`
	cfg, err := config.Load(writeTemp(t, yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Inputs) != 1 || cfg.Inputs[0].Name != "app-logs" {
		t.Errorf("unexpected inputs: %+v", cfg.Inputs)
	}
	if len(cfg.Filters) != 1 || cfg.Filters[0].Field != "level" {
		t.Errorf("unexpected filters: %+v", cfg.Filters)
	}
	if len(cfg.Outputs) != 1 || cfg.Outputs[0].Format != "json" {
		t.Errorf("unexpected outputs: %+v", cfg.Outputs)
	}
}

func TestLoad_MissingInputs(t *testing.T) {
	yaml := `
outputs:
  - name: out
    type: stdout
`
	_, err := config.Load(writeTemp(t, yaml))
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestLoad_MissingOutputs(t *testing.T) {
	yaml := `
inputs:
  - name: in
    type: stdin
`
	_, err := config.Load(writeTemp(t, yaml))
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := config.Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	_, err := config.Load(writeTemp(t, ":::invalid yaml:::"))
	if err == nil {
		t.Fatal("expected YAML parse error, got nil")
	}
}
