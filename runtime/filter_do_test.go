package runtime

import (
	"strings"
	"testing"
)

func TestDoFilterSuppressesOutput(t *testing.T) {
	env := NewEnvironment()
	env.AddGlobal("record", func(ctx *Context, args ...interface{}) (interface{}, error) {
		target := args[1].(map[interface{}]interface{})
		target["value"] = target["value"].(string) + args[0].(string)
		return "SHOULD_NOT_PRINT", nil
	})

	tmpl, err := env.ParseString(`{% set ns = {'value': ''} %}{{ record('A', ns) | do }}{{ ns.value }}`, "main")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	result, err := tmpl.ExecuteToString(nil)
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}

	if strings.TrimSpace(result) != "A" {
		t.Fatalf("expected 'A', got %q", strings.TrimSpace(result))
	}
}

func TestDoFilterIgnoresPlainValues(t *testing.T) {
	env := NewEnvironment()
	tmpl, err := env.ParseString("{{ 'Hello' | do }}World", "inline")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	result, err := tmpl.ExecuteToString(nil)
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}

	if strings.TrimSpace(result) != "World" {
		t.Fatalf("expected 'World', got %q", strings.TrimSpace(result))
	}
}
