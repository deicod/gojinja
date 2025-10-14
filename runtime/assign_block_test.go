package runtime

import (
	"strings"
	"testing"
)

func TestAssignBlockBasic(t *testing.T) {
	env := NewEnvironment()
	templates := map[string]string{
		"assign.html": `{% set message %}Hello{% endset %}Result: {{ message }}`,
	}
	env.SetLoader(NewMapLoader(templates))

	tmpl, err := env.ParseFile("assign.html")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	result, err := tmpl.ExecuteToString(nil)
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}

	expected := "Result: Hello"
	if strings.TrimSpace(result) != expected {
		t.Fatalf("expected %q, got %q", expected, strings.TrimSpace(result))
	}
}

func TestAssignBlockWithFilter(t *testing.T) {
	env := NewEnvironment()
	templates := map[string]string{
		"assign_filter.html": `{% set shout|upper %}hello{% endset %}{{ shout }}`,
	}
	env.SetLoader(NewMapLoader(templates))

	tmpl, err := env.ParseFile("assign_filter.html")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	result, err := tmpl.ExecuteToString(nil)
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}

	expected := "HELLO"
	if strings.TrimSpace(result) != expected {
		t.Fatalf("expected %q, got %q", expected, strings.TrimSpace(result))
	}
}

func TestAssignBlockWithContext(t *testing.T) {
	env := NewEnvironment()
	templates := map[string]string{
		"assign_ctx.html": `{% set greeting %}Hello {{ name }}{% endset %}{{ greeting|trim }}`,
	}
	env.SetLoader(NewMapLoader(templates))

	tmpl, err := env.ParseFile("assign_ctx.html")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	result, err := tmpl.ExecuteToString(map[string]interface{}{"name": "World"})
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}

	expected := "Hello World"
	if strings.TrimSpace(result) != expected {
		t.Fatalf("expected %q, got %q", expected, strings.TrimSpace(result))
	}
}
