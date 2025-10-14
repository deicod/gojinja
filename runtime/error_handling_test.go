package runtime

import (
	"errors"
	"testing"
)

func TestMapLoaderTemplateNotFound(t *testing.T) {
	env := NewEnvironment()
	env.SetLoader(NewMapLoader(map[string]string{}))

	_, err := env.LoadTemplate("missing.html")
	if err == nil {
		t.Fatalf("expected error for missing template")
	}

	var notFound *TemplateNotFoundError
	if !errors.As(err, &notFound) {
		t.Fatalf("expected TemplateNotFoundError, got %T: %v", err, err)
	}

	if notFound.Name != "missing.html" {
		t.Fatalf("expected missing.html, got %s", notFound.Name)
	}

	if len(notFound.Tried) != 1 || notFound.Tried[0] != "missing.html" {
		t.Fatalf("unexpected tried list: %#v", notFound.Tried)
	}
}

func TestIncludeTemplatesNotFoundAggregates(t *testing.T) {
	env := NewEnvironment()
	env.SetLoader(NewMapLoader(map[string]string{}))

	tpl, err := env.ParseString("{% include ['missing1.html', 'missing2.html'] %}", "test")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	_, err = tpl.ExecuteToString(nil)
	if err == nil {
		t.Fatalf("expected error when including missing templates")
	}

	var multi *TemplatesNotFoundError
	if !errors.As(err, &multi) {
		t.Fatalf("expected TemplatesNotFoundError, got %T: %v", err, err)
	}

	if len(multi.Names) != 2 {
		t.Fatalf("expected two names, got %v", multi.Names)
	}

	if multi.Names[0] != "missing1.html" || multi.Names[1] != "missing2.html" {
		t.Fatalf("unexpected template names: %v", multi.Names)
	}

	if len(multi.Tried) != 2 {
		t.Fatalf("expected tried list to include attempted templates, got %v", multi.Tried)
	}

	if multi.Unwrap() == nil {
		t.Fatalf("expected aggregated error to retain underlying cause")
	}
}
