package runtime

import (
	"strings"
	"testing"
)

func TestTemplateMakeModule(t *testing.T) {
	env := NewEnvironment()
	tmpl, err := env.NewTemplate(`
{% macro greet(name) %}
Hello {{ name }}!
{% endmacro %}
{% set answer = 42 %}
{% export answer %}
`)
	if err != nil {
		t.Fatalf("failed to create template: %v", err)
	}

	module, err := tmpl.MakeModule(nil)
	if err != nil {
		t.Fatalf("failed to create module: %v", err)
	}

	macro, err := module.GetMacro("greet")
	if err != nil {
		t.Fatalf("expected macro 'greet' to be available: %v", err)
	}

	ctx := NewContextWithEnvironment(env, nil)
	value, err := macro.Call(ctx, "World")
	if err != nil {
		t.Fatalf("failed to call exported macro: %v", err)
	}
	if strings.TrimSpace(value.(string)) != "Hello World!" {
		t.Errorf("expected macro output 'Hello World!', got %q", value)
	}

	exported, ok := module.Resolve("answer")
	if !ok {
		t.Fatalf("expected exported value 'answer' to be present")
	}
	switch v := exported.(type) {
	case int:
		if v != 42 {
			t.Errorf("expected exported answer to equal 42, got %v", v)
		}
	case int64:
		if v != 42 {
			t.Errorf("expected exported answer to equal 42, got %v", v)
		}
	default:
		t.Errorf("expected exported answer to be numeric, got %T", exported)
	}

	exports := module.GetExportNames()
	found := false
	for _, name := range exports {
		if name == "answer" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected exported names to include 'answer', got %v", exports)
	}
}
