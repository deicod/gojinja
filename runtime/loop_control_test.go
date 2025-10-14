package runtime

import (
	"strings"
	"testing"
)

func TestForLoopContinue(t *testing.T) {
	env := NewEnvironment()
	templates := map[string]string{
		"loop.html": `{% for x in items %}{% if x == 2 %}{% continue %}{% endif %}{{ x }}{% endfor %}`,
	}
	env.SetLoader(NewMapLoader(templates))

	tmpl, err := env.ParseFile("loop.html")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	result, err := tmpl.ExecuteToString(map[string]interface{}{"items": []interface{}{0, 1, 2, 3, 4}})
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}

	expected := "0134"
	if strings.TrimSpace(result) != expected {
		t.Fatalf("expected %q, got %q", expected, strings.TrimSpace(result))
	}
}

func TestForLoopBreak(t *testing.T) {
	env := NewEnvironment()
	templates := map[string]string{
		"loop.html": `{% for x in items %}{% if x == 3 %}{% break %}{% endif %}{{ x }}{% endfor %}`,
	}
	env.SetLoader(NewMapLoader(templates))

	tmpl, err := env.ParseFile("loop.html")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	result, err := tmpl.ExecuteToString(map[string]interface{}{"items": []interface{}{0, 1, 2, 3, 4}})
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}

	expected := "012"
	if strings.TrimSpace(result) != expected {
		t.Fatalf("expected %q, got %q", expected, strings.TrimSpace(result))
	}
}

func TestForLoopElseEmpty(t *testing.T) {
	env := NewEnvironment()
	templates := map[string]string{
		"loop.html": `{% for x in items %}{{ x }}{% else %}empty{% endfor %}`,
	}
	env.SetLoader(NewMapLoader(templates))

	tmpl, err := env.ParseFile("loop.html")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	result, err := tmpl.ExecuteToString(map[string]interface{}{"items": []interface{}{}})
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}

	expected := "empty"
	if strings.TrimSpace(result) != expected {
		t.Fatalf("expected %q, got %q", expected, strings.TrimSpace(result))
	}
}

func TestForLoopElseAfterBreak(t *testing.T) {
	env := NewEnvironment()
	templates := map[string]string{
		"loop.html": `{% for x in items %}{% if x == 2 %}{% break %}{% endif %}{{ x }}{% else %}done{% endfor %}`,
	}
	env.SetLoader(NewMapLoader(templates))

	tmpl, err := env.ParseFile("loop.html")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	result, err := tmpl.ExecuteToString(map[string]interface{}{"items": []interface{}{0, 1, 2, 3}})
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}

	expected := "01"
	if strings.TrimSpace(result) != expected {
		t.Fatalf("expected %q, got %q", expected, strings.TrimSpace(result))
	}
}

func TestForLoopContinueInsideFilterBlock(t *testing.T) {
	env := NewEnvironment()
	templates := map[string]string{
		"loop.html": `{% for x in items %}{% filter lower %}{% if x == 'B' %}{% continue %}{% endif %}{{ x }}{% endfilter %}{% endfor %}`,
	}
	env.SetLoader(NewMapLoader(templates))

	tmpl, err := env.ParseFile("loop.html")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	result, err := tmpl.ExecuteToString(map[string]interface{}{"items": []interface{}{"A", "B", "C"}})
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}

	expected := "ac"
	if strings.TrimSpace(result) != expected {
		t.Fatalf("expected %q, got %q", expected, strings.TrimSpace(result))
	}
}
