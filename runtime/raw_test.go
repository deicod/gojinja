package runtime

import "testing"

func TestRawBlockRendering(t *testing.T) {
	env := NewEnvironment()
	env.SetLoader(NewMapLoader(map[string]string{
		"raw.html": `{% raw %}{{ name }}{% endraw %} {{ name }}`,
	}))

	tmpl, err := env.ParseFile("raw.html")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	result, err := tmpl.ExecuteToString(map[string]interface{}{
		"name": "World",
	})
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}

	expected := "{{ name }} World"
	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestVerbatimBlockRendering(t *testing.T) {
	env := NewEnvironment()
	env.SetLoader(NewMapLoader(map[string]string{
		"verbatim.html": `{% verbatim %}{{ name }}{% endverbatim %} {{ name }}`,
	}))

	tmpl, err := env.ParseFile("verbatim.html")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	result, err := tmpl.ExecuteToString(map[string]interface{}{
		"name": "World",
	})
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}

	expected := "{{ name }} World"
	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestRawBlockWhitespaceControlRendering(t *testing.T) {
	env := NewEnvironment()
	env.SetLoader(NewMapLoader(map[string]string{
		"raw-trim.html": `{%- raw -%}{{ name }}{%- endraw -%}{{ name }}`,
	}))

	tmpl, err := env.ParseFile("raw-trim.html")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	result, err := tmpl.ExecuteToString(map[string]interface{}{
		"name": "World",
	})
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}

	expected := "{{ name }}World"
	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}
