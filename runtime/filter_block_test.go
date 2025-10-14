package runtime

import (
	"strings"
	"testing"
)

func TestFilterBlockUpper(t *testing.T) {
	env := NewEnvironment()
	templates := map[string]string{
		"filter.html": `{% filter upper %}hello world{% endfilter %}`,
	}
	env.SetLoader(NewMapLoader(templates))

	tmpl, err := env.ParseFile("filter.html")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	result, err := tmpl.ExecuteToString(nil)
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}

	expected := "HELLO WORLD"
	if strings.TrimSpace(result) != expected {
		t.Fatalf("expected %q, got %q", expected, strings.TrimSpace(result))
	}
}

func TestFilterBlockChained(t *testing.T) {
	env := NewEnvironment()
	templates := map[string]string{
		"chain.html": `{% filter replace('l', 'L')|reverse %}hello{% endfilter %}`,
	}
	env.SetLoader(NewMapLoader(templates))

	tmpl, err := env.ParseFile("chain.html")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	result, err := tmpl.ExecuteToString(nil)
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}

	expected := "oLLeh"
	if strings.TrimSpace(result) != expected {
		t.Fatalf("expected %q, got %q", expected, strings.TrimSpace(result))
	}
}

func TestFilterBlockContextVariables(t *testing.T) {
	env := NewEnvironment()
	templates := map[string]string{
		"context.html": `{% filter upper %}Hello {{ name }}{% endfilter %}`,
	}
	env.SetLoader(NewMapLoader(templates))

	tmpl, err := env.ParseFile("context.html")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	result, err := tmpl.ExecuteToString(map[string]interface{}{"name": "world"})
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}

	expected := "HELLO WORLD"
	if strings.TrimSpace(result) != expected {
		t.Fatalf("expected %q, got %q", expected, strings.TrimSpace(result))
	}
}
