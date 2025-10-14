package runtime

import (
	"strings"
	"testing"
)

func TestCallBlockBasic(t *testing.T) {
	env := NewEnvironment()
	templates := map[string]string{
		"base.html": `{% macro wrapper() %}<div>{{ caller() }}</div>{% endmacro %}{% call wrapper() %}Hello{% endcall %}`,
	}
	env.SetLoader(NewMapLoader(templates))

	tmpl, err := env.ParseFile("base.html")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	result, err := tmpl.ExecuteToString(nil)
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}

	expected := "<div>Hello</div>"
	if strings.TrimSpace(result) != expected {
		t.Fatalf("expected %q, got %q", expected, strings.TrimSpace(result))
	}
}

func TestCallBlockWithContext(t *testing.T) {
	env := NewEnvironment()
	templates := map[string]string{
		"page.html": `{% macro wrap() %}<div class="card">{{ caller() }}</div>{% endmacro %}{% call wrap() %}<p>Hello {{ name|capitalize }}</p>{% endcall %}`,
	}
	env.SetLoader(NewMapLoader(templates))

	tmpl, err := env.ParseFile("page.html")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	result, err := tmpl.ExecuteToString(map[string]interface{}{"name": "alice"})
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}

	expected := `<div class="card"><p>Hello Alice</p></div>`
	if strings.TrimSpace(result) != expected {
		t.Fatalf("expected %q, got %q", expected, strings.TrimSpace(result))
	}
}

func TestCallBlockNested(t *testing.T) {
	env := NewEnvironment()
	templates := map[string]string{
		"nested.html": `{% macro outer() %}<section>{{ caller() }}</section>{% endmacro %}{% macro inner() %}{% call outer() %}<article>{{ caller() }}</article>{% endcall %}{% endmacro %}{% call inner() %}Content{% endcall %}`,
	}
	env.SetLoader(NewMapLoader(templates))

	tmpl, err := env.ParseFile("nested.html")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	result, err := tmpl.ExecuteToString(nil)
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}

	expected := "<section><article>Content</article></section>"
	if strings.TrimSpace(result) != expected {
		t.Fatalf("expected %q, got %q", expected, strings.TrimSpace(result))
	}
}
