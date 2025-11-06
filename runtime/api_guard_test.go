package runtime

import (
	"bytes"
	"testing"

	"github.com/deicod/gojinja/parser"
)

func TestAPIEnvironmentGuards(t *testing.T) {
	ast, err := parser.ParseTemplate("Hello")
	if err != nil {
		t.Fatalf("failed to parse template for AST: %v", err)
	}

	if _, err := ParseASTWithEnvironment(nil, ast, "guarded"); err == nil {
		t.Fatalf("expected error when environment is nil for ParseASTWithEnvironment")
	}

	if _, err := FromASTWithEnvironment(nil, ast); err == nil {
		t.Fatalf("expected error when environment is nil for FromASTWithEnvironment")
	}

	if _, err := RenderTemplateWithEnvironment(nil, "{{ value }}", nil); err == nil {
		t.Fatalf("expected error when environment is nil for RenderTemplateWithEnvironment")
	}

	if err := RenderTemplateToWriterWithEnvironment(nil, "{{ value }}", nil, &bytes.Buffer{}); err == nil {
		t.Fatalf("expected error when environment is nil for RenderTemplateToWriterWithEnvironment")
	}
}
