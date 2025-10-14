package runtime

import (
	"strings"
	"testing"
)

func TestIncludeBasic(t *testing.T) {
	env := NewEnvironment()
	templates := map[string]string{
		"partial.html": `<p>{{ message }}</p>`,
		"main.html":    `{% include "partial.html" %}`,
	}
	env.SetLoader(NewMapLoader(templates))

	tmpl, err := env.ParseFile("main.html")
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	result, err := tmpl.ExecuteToString(map[string]interface{}{
		"message": "Hello",
	})
	if err != nil {
		t.Fatalf("failed to execute template: %v", err)
	}

	if strings.TrimSpace(result) != "<p>Hello</p>" {
		t.Fatalf("unexpected include output: %q", result)
	}
}

func TestIncludeWithoutContext(t *testing.T) {
	env := NewEnvironment()
	templates := map[string]string{
		"partial.html": `<p>{{ message|default('missing', true) }}</p>`,
		"main.html":    `{% include "partial.html" without context %}`,
	}
	env.SetLoader(NewMapLoader(templates))

	tmpl, err := env.ParseFile("main.html")
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	result, err := tmpl.ExecuteToString(map[string]interface{}{
		"message": "Hello",
	})
	if err != nil {
		t.Fatalf("failed to execute template: %v", err)
	}

	if strings.TrimSpace(result) != "<p>missing</p>" {
		t.Fatalf("expected missing message, got %q", result)
	}
}

func TestIncludeIgnoreMissing(t *testing.T) {
	env := NewEnvironment()
	templates := map[string]string{
		"main.html": `{% include "missing.html" ignore missing %}`,
	}
	env.SetLoader(NewMapLoader(templates))

	tmpl, err := env.ParseFile("main.html")
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	result, err := tmpl.ExecuteToString(nil)
	if err != nil {
		t.Fatalf("include with ignore missing should not error: %v", err)
	}

	if strings.TrimSpace(result) != "" {
		t.Fatalf("expected empty output for missing include, got %q", result)
	}
}

func TestIncludeTemplateListFallback(t *testing.T) {
	env := NewEnvironment()
	templates := map[string]string{
		"partial.html": `<p>fallback</p>`,
		"main.html":    `{% include ["missing.html", "partial.html"] %}`,
	}
	env.SetLoader(NewMapLoader(templates))

	tmpl, err := env.ParseFile("main.html")
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	result, err := tmpl.ExecuteToString(nil)
	if err != nil {
		t.Fatalf("failed to execute template: %v", err)
	}

	if strings.TrimSpace(result) != "<p>fallback</p>" {
		t.Fatalf("expected fallback include, got %q", result)
	}
}
