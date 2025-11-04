package gojinja2

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
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

func TestParseStringWithName(t *testing.T) {
	tmpl, err := ParseStringWithName("{{ greeting }}", "custom")
	if err != nil {
		t.Fatalf("ParseStringWithName error: %v", err)
	}

	if tmpl.Name() != "custom" {
		t.Fatalf("expected template name 'custom', got %q", tmpl.Name())
	}

	out, err := tmpl.ExecuteToString(map[string]interface{}{"greeting": "Hi"})
	if err != nil {
		t.Fatalf("ExecuteToString error: %v", err)
	}

	if out != "Hi" {
		t.Fatalf("expected 'Hi', got %q", out)
	}
}

func TestExecuteConvenienceFunctions(t *testing.T) {
	result, err := ExecuteToString("Hello {{ name }}", map[string]interface{}{"name": "Gopher"})
	if err != nil {
		t.Fatalf("ExecuteToString error: %v", err)
	}
	if result != "Hello Gopher" {
		t.Fatalf("expected 'Hello Gopher', got %q", result)
	}

	var buf bytes.Buffer
	if err := Execute("{{ value }}!", map[string]interface{}{"value": 42}, &buf); err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	if buf.String() != "42!" {
		t.Fatalf("expected '42!', got %q", buf.String())
	}
}

func TestTemplateChainAndBatchRenderer(t *testing.T) {
	env := NewEnvironment()

	chain := NewTemplateChain(env)
	if err := chain.AddFromString("{{ greeting }}", "welcome"); err != nil {
		t.Fatalf("AddFromString error: %v", err)
	}

	tmpl, ok := chain.Get("welcome")
	if !ok {
		t.Fatalf("expected template 'welcome' to be present in chain")
	}

	out, err := tmpl.ExecuteToString(map[string]interface{}{"greeting": "Howdy"})
	if err != nil {
		t.Fatalf("ExecuteToString error: %v", err)
	}
	if out != "Howdy" {
		t.Fatalf("expected 'Howdy', got %q", out)
	}

	renderer := NewBatchRenderer(env)
	if err := renderer.AddTemplate("farewell", "Bye {{ name }}"); err != nil {
		t.Fatalf("AddTemplate error: %v", err)
	}

	rendered, err := renderer.Render("farewell", map[string]interface{}{"name": "Go"})
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if rendered != "Bye Go" {
		t.Fatalf("expected 'Bye Go', got %q", rendered)
	}

	buf := bytes.Buffer{}
	if err := renderer.RenderToWriter("farewell", map[string]interface{}{"name": "Go"}, &buf); err != nil {
		t.Fatalf("RenderToWriter error: %v", err)
	}
	if buf.String() != "Bye Go" {
		t.Fatalf("expected 'Bye Go', got %q", buf.String())
	}
}

func TestMakeModuleExports(t *testing.T) {
	tmpl, err := ParseString(`
{% macro greet(name) %}
Hello {{ name }}!
{% endmacro %}
{% set answer = 42 %}
{% export answer %}
`)
	if err != nil {
		t.Fatalf("ParseString error: %v", err)
	}

	module, err := tmpl.MakeModule(nil)
	if err != nil {
		t.Fatalf("MakeModule error: %v", err)
	}

	macro, err := module.GetMacro("greet")
	if err != nil {
		t.Fatalf("expected greet macro to be exported: %v", err)
	}

	ctx := NewContextWithEnvironment(tmpl.Environment(), nil)
	value, err := macro.Call(ctx, "Parity")
	if err != nil {
		t.Fatalf("macro call failed: %v", err)
	}

	if result := strings.TrimSpace(value.(string)); result != "Hello Parity!" {
		t.Fatalf("unexpected macro output: %q", result)
	}

	exported, ok := module.Resolve("answer")
	if !ok {
		t.Fatalf("expected exported value 'answer' to be resolvable")
	}
	switch v := exported.(type) {
	case int:
		if v != 42 {
			t.Fatalf("expected exported answer to be 42, got %d", v)
		}
	case int64:
		if v != 42 {
			t.Fatalf("expected exported answer to be 42, got %d", v)
		}
	default:
		t.Fatalf("expected exported answer to be numeric, got %T", exported)
	}

	names := module.GetExportNames()
	found := false
	for _, name := range names {
		if name == "answer" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected export names to include 'answer', got %v", names)
	}
}
