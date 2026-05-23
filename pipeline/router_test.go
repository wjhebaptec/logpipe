package pipeline

import (
	"bytes"
	"strings"
	"testing"

	"github.com/user/logpipe/config"
)

func makeOutputs(names ...string) []config.Output {
	var outs []config.Output
	for _, n := range names {
		outs = append(outs, config.Output{Name: n})
	}
	return outs
}

func TestRouter_RouteToAllOutputs(t *testing.T) {
	var buf1, buf2 bytes.Buffer
	outputs := makeOutputs("sink1", "sink2")
	writers := map[string]io.Writer{"sink1": &buf1, "sink2": &buf2}
	r := NewRouter(outputs, writers)

	entry := map[string]interface{}{"level": "info", "msg": "hello"}
	if err := r.Route(entry); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf1.Len() == 0 || buf2.Len() == 0 {
		t.Error("expected both sinks to receive data")
	}
}

func TestRouter_LevelFilter_Match(t *testing.T) {
	var buf bytes.Buffer
	outputs := []config.Output{{Name: "errors", Level: "error"}}
	writers := map[string]io.Writer{"errors": &buf}
	r := NewRouter(outputs, writers)

	entry := map[string]interface{}{"level": "error", "msg": "boom"}
	if err := r.Route(entry); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected entry to be routed to errors sink")
	}
}

func TestRouter_LevelFilter_NoMatch(t *testing.T) {
	var buf bytes.Buffer
	outputs := []config.Output{{Name: "errors", Level: "error"}}
	writers := map[string]io.Writer{"errors": &buf}
	r := NewRouter(outputs, writers)

	entry := map[string]interface{}{"level": "info", "msg": "all good"}
	if err := r.Route(entry); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() != 0 {
		t.Error("expected entry NOT to be routed to errors sink")
	}
}

func TestRouter_MissingWriter(t *testing.T) {
	outputs := []config.Output{{Name: "ghost"}}
	r := NewRouter(outputs, map[string]io.Writer{})

	entry := map[string]interface{}{"level": "info", "msg": "test"}
	err := r.Route(entry)
	if err == nil {
		t.Fatal("expected error for missing writer")
	}
	if !strings.Contains(err.Error(), "ghost") {
		t.Errorf("error should mention missing output name, got: %v", err)
	}
}
