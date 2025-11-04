package runtime

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestEnvironmentGetTemplate(t *testing.T) {
	env := NewEnvironment()
	env.SetLoader(NewMapLoader(map[string]string{
		"base.html": "Hello {{ name }}",
	}))

	tmpl, err := env.GetTemplate("base.html")
	if err != nil {
		t.Fatalf("expected template to load, got error: %v", err)
	}

	out, err := tmpl.ExecuteToString(map[string]interface{}{"name": "Go"})
	if err != nil {
		t.Fatalf("render error: %v", err)
	}

	if strings.TrimSpace(out) != "Hello Go" {
		t.Fatalf("unexpected render output: %q", out)
	}
}

func TestEnvironmentSelectTemplate(t *testing.T) {
	env := NewEnvironment()
	env.SetLoader(NewMapLoader(map[string]string{
		"fallback.html": "Fallback",
	}))

	tmpl, err := env.SelectTemplate([]string{"missing.html", "fallback.html"})
	if err != nil {
		t.Fatalf("expected select_template to find fallback, got error: %v", err)
	}

	out, err := tmpl.ExecuteToString(nil)
	if err != nil {
		t.Fatalf("render error: %v", err)
	}

	if strings.TrimSpace(out) != "Fallback" {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestEnvironmentSelectTemplateNotFound(t *testing.T) {
	env := NewEnvironment()
	env.SetLoader(NewMapLoader(map[string]string{}))

	_, err := env.SelectTemplate([]string{"missing1.html", "missing2.html"})
	if err == nil {
		t.Fatalf("expected error when no templates can be selected")
	}

	var multi *TemplatesNotFoundError
	if !errors.As(err, &multi) {
		t.Fatalf("expected TemplatesNotFoundError, got %T", err)
	}

	if len(multi.Names) != 2 {
		t.Fatalf("expected two missing template names, got %v", multi.Names)
	}
}

func TestEnvironmentGetOrSelectTemplate(t *testing.T) {
	env := NewEnvironment()
	env.SetLoader(NewMapLoader(map[string]string{
		"one.html": "One",
		"two.html": "Two",
	}))

	// Single string path
	tmpl, err := env.GetOrSelectTemplate("one.html")
	if err != nil {
		t.Fatalf("unexpected error resolving string template: %v", err)
	}
	if result, _ := tmpl.ExecuteToString(nil); strings.TrimSpace(result) != "One" {
		t.Fatalf("unexpected render output for string template: %q", result)
	}

	// Slice of strings path
	tmpl, err = env.GetOrSelectTemplate([]string{"missing.html", "two.html"})
	if err != nil {
		t.Fatalf("unexpected error resolving slice template: %v", err)
	}
	if result, _ := tmpl.ExecuteToString(nil); strings.TrimSpace(result) != "Two" {
		t.Fatalf("unexpected render output for slice template: %q", result)
	}

	// Slice of interfaces
	tmpl, err = env.GetOrSelectTemplate([]interface{}{"missing.html", "one.html"})
	if err != nil {
		t.Fatalf("unexpected error resolving interface slice template: %v", err)
	}
	if result, _ := tmpl.ExecuteToString(nil); strings.TrimSpace(result) != "One" {
		t.Fatalf("unexpected render output for interface slice template: %q", result)
	}

	// Unsupported type
	if _, err := env.GetOrSelectTemplate(123); err == nil {
		t.Fatalf("expected error for unsupported template identifier type")
	}
}

func TestEnvironmentGetOrSelectTemplateWithTemplateObject(t *testing.T) {
	env := NewEnvironment()

	tmpl, err := env.NewTemplateWithName("Hello", "inline")
	if err != nil {
		t.Fatalf("failed to create inline template: %v", err)
	}

	resolved, err := env.GetOrSelectTemplate(tmpl)
	if err != nil {
		t.Fatalf("unexpected error resolving template object: %v", err)
	}

	if resolved != tmpl {
		t.Fatalf("expected returned template to match provided object")
	}
}

func TestEnvironmentGetOrSelectTemplateMixedList(t *testing.T) {
	env := NewEnvironment()
	env.SetLoader(NewMapLoader(map[string]string{
		"fallback.html": "Fallback",
	}))

	inline, err := env.NewTemplateWithName("Inline", "inline")
	if err != nil {
		t.Fatalf("failed to create inline template: %v", err)
	}

	// Prefer loader result when available
	resolved, err := env.GetOrSelectTemplate([]interface{}{"fallback.html", inline})
	if err != nil {
		t.Fatalf("unexpected error resolving list with loader template: %v", err)
	}

	if resolved.Name() != "fallback.html" {
		t.Fatalf("expected loader template to be selected, got %q", resolved.Name())
	}

	// Fall back to provided template object when names cannot be loaded
	resolved, err = env.GetOrSelectTemplate([]interface{}{"missing.html", inline})
	if err != nil {
		t.Fatalf("unexpected error falling back to template object: %v", err)
	}

	if resolved != inline {
		t.Fatalf("expected inline template to be returned when names are missing")
	}
}

func TestEnvironmentFromString(t *testing.T) {
	env := NewEnvironment()
	tmpl, err := env.FromString("Hello {{ name }}!")
	if err != nil {
		t.Fatalf("from_string error: %v", err)
	}

	out, err := tmpl.ExecuteToString(map[string]interface{}{"name": "Parity"})
	if err != nil {
		t.Fatalf("render error: %v", err)
	}

	if strings.TrimSpace(out) != "Hello Parity!" {
		t.Fatalf("unexpected output from from_string: %q", out)
	}
}

func TestEnvironmentJoinPath(t *testing.T) {
	env := NewEnvironment()

	joined, err := env.JoinPath("partials/header.html", "layouts/base.html")
	if err != nil {
		t.Fatalf("join_path error: %v", err)
	}
	if joined != "layouts/partials/header.html" {
		t.Fatalf("unexpected joined path: %q", joined)
	}

	joined, err = env.JoinPath("child.html", "base.html")
	if err != nil {
		t.Fatalf("join_path error for root parent: %v", err)
	}
	if joined != "child.html" {
		t.Fatalf("expected child.html for root parent, got %q", joined)
	}

	joined, err = env.JoinPath("/absolute/path.html", "layouts/base.html")
	if err != nil {
		t.Fatalf("join_path error for absolute path: %v", err)
	}
	if joined != "/absolute/path.html" {
		t.Fatalf("unexpected absolute join result: %q", joined)
	}
}

func TestEnvironmentJoinPathDelegatesToLoader(t *testing.T) {
	env := NewEnvironment()
	loader := &trackingJoinLoader{}
	env.SetLoader(loader)

	joined, err := env.JoinPath("component.html", "layouts/base.html")
	if err != nil {
		t.Fatalf("join_path delegation error: %v", err)
	}
	if joined != "loader:layouts/base.html:component.html" {
		t.Fatalf("expected loader join result, got %q", joined)
	}
	if loader.calls != 1 {
		t.Fatalf("expected loader JoinPath to be invoked once, got %d", loader.calls)
	}

	empty := ""
	loader.nextReturn = &empty
	joined, err = env.JoinPath("component.html", "layouts/base.html")
	if err != nil {
		t.Fatalf("join_path fallback error: %v", err)
	}
	if joined != "layouts/component.html" {
		t.Fatalf("expected fallback join result, got %q", joined)
	}
	if loader.calls != 2 {
		t.Fatalf("expected loader JoinPath to be invoked twice, got %d", loader.calls)
	}

	loader.nextErr = errors.New("boom")
	if _, err := env.JoinPath("component.html", "layouts/base.html"); err == nil {
		t.Fatalf("expected loader join error to propagate")
	}
}

type trackingJoinLoader struct {
	nextReturn *string
	nextErr    error
	calls      int
}

func (l *trackingJoinLoader) Load(name string) (string, error) {
	return "", os.ErrNotExist
}

func (l *trackingJoinLoader) JoinPath(template, parent string) (string, error) {
	l.calls++
	if l.nextErr != nil {
		err := l.nextErr
		l.nextErr = nil
		return "", err
	}
	if l.nextReturn != nil {
		res := *l.nextReturn
		l.nextReturn = nil
		return res, nil
	}
	return "loader:" + parent + ":" + template, nil
}
