package pipeline

import (
	"testing"
)

func schemaEntry(fields map[string]string) LogEntry {
	return LogEntry{Fields: fields}
}

func TestSchema_NoRules_AlwaysValid(t *testing.T) {
	sv, err := NewSchemaValidator(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	entry := schemaEntry(map[string]string{})
	if !sv.IsValid(entry) {
		t.Error("expected valid with no rules")
	}
}

func TestSchema_RequiredField_Present(t *testing.T) {
	sv, _ := NewSchemaValidator([]SchemaRule{
		{Field: "service", Required: true},
	})
	entry := schemaEntry(map[string]string{"service": "auth"})
	if !sv.IsValid(entry) {
		t.Error("expected valid when required field is present")
	}
}

func TestSchema_RequiredField_Missing(t *testing.T) {
	sv, _ := NewSchemaValidator([]SchemaRule{
		{Field: "service", Required: true},
	})
	entry := schemaEntry(map[string]string{})
	errs := sv.Validate(entry)
	if len(errs) == 0 {
		t.Error("expected validation error for missing required field")
	}
}

func TestSchema_PatternMatch(t *testing.T) {
	sv, err := NewSchemaValidator([]SchemaRule{
		{Field: "request_id", Pattern: `^[a-f0-9\-]{36}$`},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	entry := schemaEntry(map[string]string{"request_id": "550e8400-e29b-41d4-a716-446655440000"})
	if !sv.IsValid(entry) {
		t.Error("expected valid UUID to match pattern")
	}
}

func TestSchema_PatternNoMatch(t *testing.T) {
	sv, _ := NewSchemaValidator([]SchemaRule{
		{Field: "request_id", Pattern: `^[a-f0-9\-]{36}$`},
	})
	entry := schemaEntry(map[string]string{"request_id": "not-a-uuid"})
	errs := sv.Validate(entry)
	if len(errs) == 0 {
		t.Error("expected validation error for pattern mismatch")
	}
}

func TestSchema_InvalidPattern_ReturnsError(t *testing.T) {
	_, err := NewSchemaValidator([]SchemaRule{
		{Field: "foo", Pattern: `[invalid`},
	})
	if err == nil {
		t.Error("expected error for invalid regex pattern")
	}
}

func TestSchema_OptionalField_AbsentNoError(t *testing.T) {
	sv, _ := NewSchemaValidator([]SchemaRule{
		{Field: "trace_id", Required: false, Pattern: `^[0-9a-f]+$`},
	})
	entry := schemaEntry(map[string]string{})
	if !sv.IsValid(entry) {
		t.Error("optional absent field should not cause error")
	}
}

func TestSchema_MultipleErrors(t *testing.T) {
	sv, _ := NewSchemaValidator([]SchemaRule{
		{Field: "service", Required: true},
		{Field: "env", Required: true},
	})
	entry := schemaEntry(map[string]string{})
	errs := sv.Validate(entry)
	if len(errs) != 2 {
		t.Errorf("expected 2 errors, got %d", len(errs))
	}
}
