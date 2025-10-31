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

func TestTemplateModuleExportsImports(t *testing.T) {
	env := NewEnvironment()
	env.SetLoader(NewMapLoader(map[string]string{
		"helpers.html": `{% macro greet(name) %}Hello {{ name }}!{% endmacro %}`,
		"main.html": `
{% import "helpers.html" as helpers %}
{% from "helpers.html" import greet as salute %}
{% from "helpers.html" import greet %}
`,
	}))

	tmpl, err := env.ParseFile("main.html")
	if err != nil {
		t.Fatalf("failed to parse main template: %v", err)
	}

	module, err := tmpl.MakeModule(nil)
	if err != nil {
		t.Fatalf("failed to create module with imports: %v", err)
	}

	helpersValue, ok := module.Resolve("helpers")
	if !ok {
		t.Fatalf("expected namespace 'helpers' to be exported")
	}

	helpersNS, ok := helpersValue.(*MacroNamespace)
	if !ok {
		t.Fatalf("expected 'helpers' export to be a MacroNamespace, got %T", helpersValue)
	}

	if _, err := helpersNS.GetMacro("greet"); err != nil {
		t.Fatalf("expected helpers namespace to provide greet macro: %v", err)
	}

	saluteValue, ok := module.Resolve("salute")
	if !ok {
		t.Fatalf("expected imported macro alias 'salute' to be exported")
	}

	saluteMacro, ok := saluteValue.(*Macro)
	if !ok {
		t.Fatalf("expected 'salute' export to be a Macro, got %T", saluteValue)
	}

	ctx := NewContextWithEnvironment(env, nil)
	rendered, err := saluteMacro.Call(ctx, "Go")
	if err != nil {
		t.Fatalf("failed to execute imported macro: %v", err)
	}

	if strings.TrimSpace(rendered.(string)) != "Hello Go!" {
		t.Fatalf("unexpected macro output: %v", rendered)
	}

	greetValue, ok := module.Resolve("greet")
	if !ok {
		t.Fatalf("expected direct import 'greet' to be exported")
	}

	if _, ok := greetValue.(*Macro); !ok {
		t.Fatalf("expected 'greet' export to be a Macro, got %T", greetValue)
	}
}
