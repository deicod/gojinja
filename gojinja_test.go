package gojinja2

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "greeting.html")
	if err := os.WriteFile(path, []byte("Hello {{ name }}!"), 0o644); err != nil {
		t.Fatalf("failed to write template: %v", err)
	}

	tmpl, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile error: %v", err)
	}

	output, err := tmpl.ExecuteToString(map[string]interface{}{"name": "Go"})
	if err != nil {
		t.Fatalf("ExecuteToString error: %v", err)
	}

	if output != "Hello Go!" {
		t.Fatalf("expected 'Hello Go!', got %q", output)
	}
}

func TestFloorDivisionOperator(t *testing.T) {
	tmpl, err := ParseString("{{ 7 // 2 }}")
	if err != nil {
		t.Fatalf("ParseString error: %v", err)
	}

	output, err := tmpl.ExecuteToString(nil)
	if err != nil {
		t.Fatalf("ExecuteToString error: %v", err)
	}

	if output != "3" {
		t.Fatalf("expected '3', got %q", output)
	}
}
