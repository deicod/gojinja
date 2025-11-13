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

	if _, err := ParseFileWithEnvironment(nil, "template.html"); err == nil {
		t.Fatalf("expected error when environment is nil for ParseFileWithEnvironment")
	}

	if _, err := RenderTemplateWithEnvironment(nil, "{{ value }}", nil); err == nil {
		t.Fatalf("expected error when environment is nil for RenderTemplateWithEnvironment")
	}

	if err := RenderTemplateToWriterWithEnvironment(nil, "{{ value }}", nil, &bytes.Buffer{}); err == nil {
		t.Fatalf("expected error when environment is nil for RenderTemplateToWriterWithEnvironment")
	}
}

func TestParseFileWithEnvironment(t *testing.T) {
	env := NewEnvironment()
	env.SetLoader(NewMapLoader(map[string]string{
		"greeting.html": "Hello {{ name }}!",
	}))

	tmpl, err := ParseFileWithEnvironment(env, "greeting.html")
	if err != nil {
		t.Fatalf("ParseFileWithEnvironment error: %v", err)
	}

	output, err := tmpl.ExecuteToString(map[string]interface{}{"name": "Parity"})
	if err != nil {
		t.Fatalf("template execution error: %v", err)
	}

	if output != "Hello Parity!" {
		t.Fatalf("unexpected template output: %q", output)
	}

	if _, err := ParseFileWithEnvironment(env, ""); err == nil {
		t.Fatalf("expected error when template name is empty")
	}
}
