package runtime

import (
	"strings"
	"testing"
)

func TestImportManagerDetectsCircularImports(t *testing.T) {
	env := NewEnvironment()
	loader := NewMapLoader(map[string]string{
		"a.html": "{% import 'b.html' as b %}",
		"b.html": "{% import 'a.html' as a %}",
	})
	env.SetLoader(loader)

	importManager := NewImportManager(env)
	ctx := NewContextWithEnvironment(env, nil)

	if _, err := importManager.ImportTemplate(ctx, "a.html", false); err == nil {
		t.Fatal("expected error for circular import, got nil")
	} else if !strings.Contains(err.Error(), "circular import detected") {
		t.Fatalf("expected circular import error, got %v", err)
	}
}
