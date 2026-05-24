package pipeline

import "testing"

// TestLabel_WithFilter verifies that a Labeler only enriches entries that pass
// a filter, simulating a conditional labelling step in a pipeline.
func TestLabel_WithFilter(t *testing.T) {
	labeler := NewLabeler(LabelConfig{
		Static: map[string]string{"tier": "critical"},
	})

	entries := []LogEntry{
		{Level: "error", Message: "disk full", Fields: map[string]string{}},
		{Level: "info", Message: "started", Fields: map[string]string{}},
		{Level: "error", Message: "oom", Fields: map[string]string{}},
	}

	filterCfg := FilterConfig{Level: "error"}
	var labelled []LogEntry
	for _, e := range entries {
		if matchesFilter(e, filterCfg) {
			labelled = append(labelled, labeler.Apply(e))
		}
	}

	if len(labelled) != 2 {
		t.Fatalf("expected 2 labelled entries, got %d", len(labelled))
	}
	for _, e := range labelled {
		if e.Fields["tier"] != "critical" {
			t.Errorf("expected tier=critical on error entry")
		}
	}
}

// TestLabel_WithTransform verifies that labels applied before a Transform are
// visible to rename/remove rules in the Transform step.
func TestLabel_WithTransform(t *testing.T) {
	labeler := NewLabeler(LabelConfig{
		Static: map[string]string{"source": "logpipe"},
		Prefix: "lp",
	})
	transformCfg := TransformConfig{
		RenameFields: map[string]string{"lp_source": "origin"},
	}

	entry := LogEntry{Level: "info", Message: "hello", Fields: map[string]string{}}
	labelled := labeler.Apply(entry)
	transformed := Transform(labelled, transformCfg)

	if transformed.Fields["origin"] != "logpipe" {
		t.Errorf("expected origin=logpipe after rename, got %q", transformed.Fields["origin"])
	}
	if _, ok := transformed.Fields["lp_source"]; ok {
		t.Error("lp_source should have been renamed to origin")
	}
}
