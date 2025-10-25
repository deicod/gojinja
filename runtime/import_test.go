package runtime

import (
	"strings"
	"testing"
)

func TestImportManagerDetectsCircularImports(t *testing.T) {
	env := NewEnvironment()
	loader := NewMapLoader(map[string]string{
		"a.html": "{% import 'b.html' as b %}",
		"b.html": "{% import 'a.html' as a %}",
	})
	env.SetLoader(loader)

	importManager := NewImportManager(env)
	ctx := NewContextWithEnvironment(env, nil)

	if _, err := importManager.ImportTemplate(ctx, "a.html", false); err == nil {
		t.Fatal("expected error for circular import, got nil")
	} else if !strings.Contains(err.Error(), "circular import detected") {
		t.Fatalf("expected circular import error, got %v", err)
	}
}

func TestFromImportAll(t *testing.T) {
	env := NewEnvironment()
	loader := NewMapLoader(map[string]string{
		"helpers.html": `
{% macro greet(name) %}Hello {{ name }}!{% endmacro %}
{% macro _hidden_macro() %}secret{% endmacro %}
{% set answer = 42 %}
{% export answer %}
{% set not_exported = 'nope' %}
`,
		"main.html": `
{% from "helpers.html" import * %}
{{ greet("World") }}|{{ answer }}|
{% if _hidden_macro is defined %}hidden{% else %}nohidden{% endif %}|
{% if not_exported is defined %}exported{% else %}noexport{% endif %}
`,
	})
	env.SetLoader(loader)

	tmpl, err := env.LoadTemplate("main.html")
	if err != nil {
		t.Fatalf("failed to load main template: %v", err)
	}

	output, err := tmpl.ExecuteToString(nil)
	if err != nil {
		t.Fatalf("failed to render template: %v", err)
	}

	cleaned := strings.ReplaceAll(strings.TrimSpace(output), "\n", "")
	expected := "Hello World!|42|nohidden|noexport"
	if cleaned != expected {
		t.Fatalf("expected %q, got %q", expected, cleaned)
	}
}
