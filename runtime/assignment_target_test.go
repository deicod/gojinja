package runtime

import (
	"strings"
	"testing"
)

func TestSetMapIndex(t *testing.T) {
	env := NewEnvironment()
	templates := map[string]string{
		"tmpl.html": `{% set data = {'a': 1} %}{% set data['b'] = 5 %}{{ data['b'] }}{{ data['a'] }}`,
	}
	env.SetLoader(NewMapLoader(templates))

	tmpl, err := env.ParseFile("tmpl.html")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	result, err := tmpl.ExecuteToString(nil)
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}

	expected := "51"
	if strings.TrimSpace(result) != expected {
		t.Fatalf("expected %q, got %q", expected, strings.TrimSpace(result))
	}
}

func TestSetMapAttribute(t *testing.T) {
	env := NewEnvironment()
	templates := map[string]string{
		"tmpl.html": `{% set data = {'x': 1} %}{% set data.y = 42 %}{{ data.y }}`,
	}
	env.SetLoader(NewMapLoader(templates))

	tmpl, err := env.ParseFile("tmpl.html")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	result, err := tmpl.ExecuteToString(nil)
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}

	if strings.TrimSpace(result) != "42" {
		t.Fatalf("expected 42, got %q", strings.TrimSpace(result))
	}
}

func TestSetListIndex(t *testing.T) {
	env := NewEnvironment()
	templates := map[string]string{
		"tmpl.html": `{% set items = [1, 2, 3] %}{% set items[1] = 9 %}{{ items[1] }}`,
	}
	env.SetLoader(NewMapLoader(templates))

	tmpl, err := env.ParseFile("tmpl.html")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	result, err := tmpl.ExecuteToString(nil)
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}

	if strings.TrimSpace(result) != "9" {
		t.Fatalf("expected 9, got %q", strings.TrimSpace(result))
	}
}

func TestSetMapIndexTyped(t *testing.T) {
	env := NewEnvironment()
	templates := map[string]string{
		"tmpl.html": `{% set data = data %}{% set data['b'] = 8 %}{{ data['b'] }}`,
	}
	env.SetLoader(NewMapLoader(templates))

	ctx := map[string]interface{}{
		"data": map[string]int{"a": 1},
	}

	tmpl, err := env.ParseFile("tmpl.html")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	result, err := tmpl.ExecuteToString(ctx)
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}

	if strings.TrimSpace(result) != "8" {
		t.Fatalf("expected 8, got %q", strings.TrimSpace(result))
	}
}

func TestSetListNegativeIndex(t *testing.T) {
	env := NewEnvironment()
	templates := map[string]string{
		"tmpl.html": `{% set items = ['a', 'b', 'c'] %}{% set items[-1] = 'z' %}{{ items[2] }}`,
	}
	env.SetLoader(NewMapLoader(templates))

	tmpl, err := env.ParseFile("tmpl.html")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	result, err := tmpl.ExecuteToString(nil)
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}

	if strings.TrimSpace(result) != "z" {
		t.Fatalf("expected z, got %q", strings.TrimSpace(result))
	}
}

func TestSetIndexOutOfRange(t *testing.T) {
	env := NewEnvironment()
	templates := map[string]string{
		"tmpl.html": `{% set items = [0, 1] %}{% set items[5] = 9 %}`,
	}
	env.SetLoader(NewMapLoader(templates))

	tmpl, err := env.ParseFile("tmpl.html")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	_, err = tmpl.ExecuteToString(nil)
	if err == nil || !strings.Contains(err.Error(), "index") {
		t.Fatalf("expected index error, got %v", err)
	}
}

func TestSetNestedAttribute(t *testing.T) {
	env := NewEnvironment()
	templates := map[string]string{
		"tmpl.html": `{% set data = {'user': {'name': 'Initial'}} %}{% set data.user.name = 'Updated' %}{{ data.user.name }}`,
	}
	env.SetLoader(NewMapLoader(templates))

	tmpl, err := env.ParseFile("tmpl.html")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	result, err := tmpl.ExecuteToString(nil)
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}

	if strings.TrimSpace(result) != "Updated" {
		t.Fatalf("expected nested attribute assignment to update value, got %q", strings.TrimSpace(result))
	}
}

func TestSetNestedIndexAndAttribute(t *testing.T) {
	env := NewEnvironment()
	templates := map[string]string{
		"tmpl.html": `{% set data = {'users': [{'name': 'Ada'}, {'name': 'Lin'}]} %}{% set data.users[0].name = 'Grace' %}{{ data.users[0].name }}`,
	}
	env.SetLoader(NewMapLoader(templates))

	tmpl, err := env.ParseFile("tmpl.html")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	result, err := tmpl.ExecuteToString(nil)
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}

	if strings.TrimSpace(result) != "Grace" {
		t.Fatalf("expected nested index assignment to update value, got %q", strings.TrimSpace(result))
	}
}
