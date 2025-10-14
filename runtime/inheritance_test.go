package runtime

import (
	"strings"
	"testing"
	"time"
)

func TestBasicInheritance(t *testing.T) {
	env := NewEnvironment()

	templates := map[string]string{
		"base.html": `<html>
<head><title>{% block title %}Default Title{% endblock %}</title></head>
<body>
{% block content %}{% endblock %}
</body>
</html>`,

		"child.html": `{% extends "base.html" %}
{% block title %}Child Title{% endblock %}
{% block content %}<p>Child content</p>{% endblock %}`,
	}

	loader := NewMapLoader(templates)
	env.SetLoader(loader)

	tmpl, err := env.ParseFile("child.html")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	result, err := tmpl.ExecuteToString(map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	expected := `<html>
<head><title>Child Title</title></head>
<body>
<p>Child content</p>
</body>
</html>`

	if strings.TrimSpace(result) != strings.TrimSpace(expected) {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestSuperFunction(t *testing.T) {
	env := NewEnvironment()

	templates := map[string]string{
		"base.html": `<html>
{% block content %}<p>Base content</p>{% endblock %}
</html>`,

		"child.html": `{% extends "base.html" %}
{% block content %}
{{ super() }}
<p>Child content</p>
{% endblock %}`,
	}

	loader := NewMapLoader(templates)
	env.SetLoader(loader)

	tmpl, err := env.ParseFile("child.html")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	result, err := tmpl.ExecuteToString(map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	expected := `<html>
<p>Base content</p>
<p>Child content</p>
</html>`

	if strings.TrimSpace(result) != strings.TrimSpace(expected) {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestNestedInheritance(t *testing.T) {
	env := NewEnvironment()

	templates := map[string]string{
		"base.html": `<html>
{% block header %}<h1>Base Header</h1>{% endblock %}
{% block content %}{% endblock %}
</html>`,

		"middle.html": `{% extends "base.html" %}
{% block content %}
<p>Middle content</p>
{% block subcontent %}{% endblock %}
{% endblock %}`,

		"child.html": `{% extends "middle.html" %}
{% block header %}<h1>Child Header</h1>{% endblock %}
{% block subcontent %}<p>Child subcontent</p>{% endblock %}`,
	}

	loader := NewMapLoader(templates)
	env.SetLoader(loader)

	tmpl, err := env.ParseFile("child.html")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	result, err := tmpl.ExecuteToString(map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	expected := `<html>
<h1>Child Header</h1>
<p>Middle content</p>
<p>Child subcontent</p>
</html>`

	if strings.TrimSpace(result) != strings.TrimSpace(expected) {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestCircularDependencyDetection(t *testing.T) {
	env := NewEnvironment()

	templates := map[string]string{
		"a.html": `{% extends "b.html" %}`,
		"b.html": `{% extends "a.html" %}`,
	}

	loader := NewMapLoader(templates)
	env.SetLoader(loader)

	_, err := env.ParseFile("a.html")
	if err == nil {
		t.Fatal("Expected error for circular dependency, got nil")
	}

	expectedMsg := "circular template inheritance detected"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error message to contain %q, got %q", expectedMsg, err.Error())
	}
}

func TestMissingParentTemplate(t *testing.T) {
	env := NewEnvironment()

	templates := map[string]string{
		"child.html": `{% extends "missing.html" %}
{% block content %}<p>Child content</p>{% endblock %}`,
	}

	loader := NewMapLoader(templates)
	env.SetLoader(loader)

	_, err := env.ParseFile("child.html")
	if err == nil {
		t.Fatal("Expected error for missing parent template, got nil")
	}

	expectedMsg := "template missing.html not found"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error message to contain %q, got %q", expectedMsg, err.Error())
	}
}

func TestTemplateCaching(t *testing.T) {
	env := NewEnvironment()

	templates := map[string]string{
		"test.html": `<p>{{ message }}</p>`,
	}

	loader := NewMapLoader(templates)
	env.SetLoader(loader)

	// Parse template multiple times
	for i := 0; i < 3; i++ {
		tmpl, err := env.ParseFile("test.html")
		if err != nil {
			t.Fatalf("Failed to parse template: %v", err)
		}

		result, err := tmpl.ExecuteToString(map[string]interface{}{
			"message": "Hello World",
		})
		if err != nil {
			t.Fatalf("Failed to execute template: %v", err)
		}

		expected := "<p>Hello World</p>"
		if strings.TrimSpace(result) != strings.TrimSpace(expected) {
			t.Errorf("Expected %q, got %q", expected, result)
		}
	}

	// Check cache size
	cacheSize := env.CacheSize()
	if cacheSize == 0 {
		t.Error("Expected cache size > 0, got 0")
	}

	// Clear cache
	env.ClearCache()
	if env.CacheSize() != 0 {
		t.Errorf("Expected cache size 0 after clear, got %d", env.CacheSize())
	}
}

func TestCacheTTL(t *testing.T) {
	env := NewEnvironment()
	env.SetCacheTTL(100 * time.Millisecond) // 100ms TTL

	templates := map[string]string{
		"test.html": `<p>{{ message }}</p>`,
	}

	loader := NewMapLoader(templates)
	env.SetLoader(loader)

	// Parse template
	tmpl, err := env.ParseFile("test.html")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	// Should be cached
	if env.CacheSize() != 1 {
		t.Errorf("Expected cache size 1, got %d", env.CacheSize())
	}

	// Execute template
	result, err := tmpl.ExecuteToString(map[string]interface{}{
		"message": "Hello World",
	})
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	expected := "<p>Hello World</p>"
	if strings.TrimSpace(result) != strings.TrimSpace(expected) {
		t.Errorf("Expected %q, got %q", expected, result)
	}

	// Wait for cache to expire
	time.Sleep(150 * time.Millisecond)

	// Cache should be cleaned on next access
	tmpl2, err := env.ParseFile("test.html")
	if err != nil {
		t.Fatalf("Failed to parse template after TTL: %v", err)
	}

	// Should still work (re-parsed)
	result2, err := tmpl2.ExecuteToString(map[string]interface{}{
		"message": "Hello World",
	})
	if err != nil {
		t.Fatalf("Failed to execute template after TTL: %v", err)
	}

	if strings.TrimSpace(result2) != strings.TrimSpace(expected) {
		t.Errorf("Expected %q, got %q", expected, result2)
	}
}

func TestBlockScoping(t *testing.T) {
	env := NewEnvironment()

	templates := map[string]string{
		"base.html": `<html>
{% block content %}{% endblock %}
{% block footer %}<p>Base footer: {{ footer_var }}</p>{% endblock %}
</html>`,

		"child.html": `{% extends "base.html" %}
{% block content %}
{% set content_var = "Child Content" %}
<p>{{ content_var }}</p>
{% endblock %}

{% block footer scoped %}
{% set footer_var = "Child Footer" %}
{{ super() }}
{% endblock %}`,
	}

	loader := NewMapLoader(templates)
	env.SetLoader(loader)

	tmpl, err := env.ParseFile("child.html")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	result, err := tmpl.ExecuteToString(map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	// The scoped block should not affect the parent block's variables
	if !strings.Contains(result, "Base footer:") {
		t.Errorf("Expected result to contain 'Base footer:', got %q", result)
	}
}

func TestFileSystemLoader(t *testing.T) {
	// This test would require actual files, so we'll just test the loader creation
	env := NewEnvironment()

	loader := NewFileSystemLoader("./templates")
	env.SetLoader(loader)

	if loader == nil {
		t.Fatal("Failed to create FileSystemLoader")
	}
}

func TestMultipleExtendsError(t *testing.T) {
	env := NewEnvironment()

	templates := map[string]string{
		"base.html": `<html>{% block content %}{% endblock %}</html>`,
		"child.html": `{% extends "base.html" %}
{% extends "base.html" %}
{% block content %}<p>Child content</p>{% endblock %}`,
	}

	loader := NewMapLoader(templates)
	env.SetLoader(loader)

	_, err := env.ParseFile("child.html")
	if err == nil {
		t.Fatal("Expected error for multiple extends statements, got nil")
	}

	expectedMsg := "multiple extends statements not allowed"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error message to contain %q, got %q", expectedMsg, err.Error())
	}
}

func TestSuperWithArgument(t *testing.T) {
	env := NewEnvironment()

	templates := map[string]string{
		"base.html": `<html>
{% block header %}<h1>Base Header</h1>{% endblock %}
{% block content %}<p>Base Content</p>{% endblock %}
{% block footer %}<p>Base Footer</p>{% endblock %}
</html>`,

		"child.html": `{% extends "base.html" %}
{% block header %}<h1>Child Header</h1>{% endblock %}
{% block content %}
{{ super('header') }}
<p>Child Content</p>
{% endblock %}`,
	}

	loader := NewMapLoader(templates)
	env.SetLoader(loader)

	tmpl, err := env.ParseFile("child.html")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	result, err := tmpl.ExecuteToString(map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	// Should include the header block content
	if !strings.Contains(result, "Base Header") {
		t.Errorf("Expected result to contain 'Base Header', got %q", result)
	}

	if !strings.Contains(result, "Child Content") {
		t.Errorf("Expected result to contain 'Child Content', got %q", result)
	}
}