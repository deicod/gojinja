package runtime

import (
	"strings"
	"testing"
)

func TestDoStatementWithSideEffect(t *testing.T) {
	env := NewEnvironment()
	env.AddGlobal("append", func(ctx *Context, args ...interface{}) (interface{}, error) {
		target := args[1].(map[interface{}]interface{})
		target["value"] = target["value"].(string) + args[0].(string)
		return nil, nil
	})

	templates := map[string]string{
		"main.html": `{% set ns = {'value': ''} %}{% do append('A', ns) %}{{ ns.value }}`,
	}
	env.SetLoader(NewMapLoader(templates))

	tmpl, err := env.ParseFile("main.html")
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

func TestDoStatementDoesNotOutput(t *testing.T) {
	env := NewEnvironment()
	templates := map[string]string{
		"test.html": `{% do noop() %}Hello`,
	}

	env.SetLoader(NewMapLoader(templates))
	env.AddGlobal("noop", func(ctx *Context, args ...interface{}) (interface{}, error) {
		return nil, nil
	})

	tmpl, err := env.ParseFile("test.html")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	result, err := tmpl.ExecuteToString(nil)
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}

	if strings.TrimSpace(result) != "Hello" {
		t.Fatalf("expected 'Hello', got %q", strings.TrimSpace(result))
	}
}
