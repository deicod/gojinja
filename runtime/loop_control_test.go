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

func TestLoopCycle(t *testing.T) {
	env := NewEnvironment()
	templates := map[string]string{
		"loop.html": `{% for x in items %}{{ loop.cycle('odd', 'even') }} {% endfor %}`,
	}
	env.SetLoader(NewMapLoader(templates))

	tmpl, err := env.ParseFile("loop.html")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	result, err := tmpl.ExecuteToString(map[string]interface{}{"items": []int{1, 2, 3, 4, 5}})
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}

	normalized := strings.Join(strings.Fields(result), " ")
	expected := "odd even odd even odd"
	if normalized != expected {
		t.Fatalf("expected %q, got %q", expected, normalized)
	}
}

func TestLoopCycleRequiresArguments(t *testing.T) {
	env := NewEnvironment()
	templates := map[string]string{
		"loop.html": `{% for x in items %}{{ loop.cycle() }}{% endfor %}`,
	}
	env.SetLoader(NewMapLoader(templates))

	tmpl, err := env.ParseFile("loop.html")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	_, err = tmpl.ExecuteToString(map[string]interface{}{"items": []int{1}})
	if err == nil {
		t.Fatal("expected execute error when calling loop.cycle without arguments")
	}
	if !strings.Contains(err.Error(), "no items for cycling given") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestLoopChanged(t *testing.T) {
	env := NewEnvironment()
	templates := map[string]string{
		"loop.html": `{% for item in items %}{% if loop.changed(item.category) %}[{{ item.category }}]{% endif %}{{ item.value }} {% endfor %}`,
	}
	env.SetLoader(NewMapLoader(templates))

	tmpl, err := env.ParseFile("loop.html")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	items := []map[string]interface{}{
		{"category": "A", "value": "one"},
		{"category": "A", "value": "two"},
		{"category": "B", "value": "three"},
		{"category": "B", "value": "four"},
		{"category": "A", "value": "five"},
	}

	result, err := tmpl.ExecuteToString(map[string]interface{}{"items": items})
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}

	expected := "[A]one two [B]three four [A]five"
	if strings.TrimSpace(result) != expected {
		t.Fatalf("expected %q, got %q", expected, strings.TrimSpace(result))
	}
}

func TestLoopDepthInNestedLoops(t *testing.T) {
	env := NewEnvironment()
	templates := map[string]string{
		"loop.html": `{% for row in rows %}O{{ loop.depth }}:{{ loop.depth0 }}|{% for col in row %}I{{ loop.depth }}:{{ loop.depth0 }} {% endfor %}{% endfor %}`,
	}
	env.SetLoader(NewMapLoader(templates))

	tmpl, err := env.ParseFile("loop.html")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	rows := [][]int{{1, 2}, {3}}
	result, err := tmpl.ExecuteToString(map[string]interface{}{"rows": rows})
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}

	expected := "O1:0|I2:1 I2:1 O1:0|I2:1"
	if strings.TrimSpace(result) != expected {
		t.Fatalf("expected %q, got %q", expected, strings.TrimSpace(result))
	}
}
