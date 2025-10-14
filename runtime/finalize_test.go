package runtime

import (
	"errors"
	"strings"
	"testing"
)

func TestFinalizeAppliedToOutput(t *testing.T) {
	env := NewEnvironment()
	env.SetLoader(NewMapLoader(map[string]string{
		"main.html": `Hi {{ name }}!`,
	}))

	env.SetFinalize(func(value interface{}) (interface{}, error) {
		if str, ok := value.(string); ok {
			return strings.ToUpper(str), nil
		}
		return value, nil
	})

	tmpl, err := env.ParseFile("main.html")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	result, err := tmpl.ExecuteToString(map[string]interface{}{"name": "world"})
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}

	if result != "Hi WORLD!" {
		t.Fatalf("expected 'Hi WORLD!', got %q", result)
	}
}

func TestFinalizeNotAppliedToTemplateData(t *testing.T) {
	env := NewEnvironment()
	env.SetLoader(NewMapLoader(map[string]string{
		"main.html": `Static text`,
	}))

	env.SetFinalize(func(value interface{}) (interface{}, error) {
		if str, ok := value.(string); ok {
			return "##" + str + "##", nil
		}
		return value, nil
	})

	tmpl, err := env.ParseFile("main.html")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	result, err := tmpl.ExecuteToString(nil)
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}

	if result != "Static text" {
		t.Fatalf("expected template data untouched, got %q", result)
	}
}

func TestFinalizeErrorPropagates(t *testing.T) {
	env := NewEnvironment()
	env.SetLoader(NewMapLoader(map[string]string{
		"main.html": `{{ value }}`,
	}))

	env.SetFinalize(func(value interface{}) (interface{}, error) {
		return nil, errors.New("finalize boom")
	})

	tmpl, err := env.ParseFile("main.html")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if _, err := tmpl.ExecuteToString(map[string]interface{}{"value": "data"}); err == nil {
		t.Fatalf("expected finalize error, got nil")
	}
}
