package runtime

import (
	"strings"
	"testing"
)

func TestBlockCompilation(t *testing.T) {
	env := NewEnvironment()

	// Create a simple template with blocks
	templates := map[string]string{
		"simple.html": `<html>
{% block header %}<header>Default Header</header>{% endblock %}
{% block content %}<main>Default Content</main>{% endblock %}
{% block footer %}<footer>Default Footer</footer>{% endblock %}
</html>`,
	}

	loader := NewMapLoader(templates)
	env.SetLoader(loader)

	tmpl, err := env.ParseFile("simple.html")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	// Test that blocks are collected correctly
	if !tmpl.HasBlock("header") {
		t.Error("Expected template to have 'header' block")
	}

	if !tmpl.HasBlock("content") {
		t.Error("Expected template to have 'content' block")
	}

	if !tmpl.HasBlock("footer") {
		t.Error("Expected template to have 'footer' block")
	}

	// Test block names
	blockNames := tmpl.BlockNames()
	expectedBlocks := []string{"header", "content", "footer"}

	if len(blockNames) != len(expectedBlocks) {
		t.Errorf("Expected %d blocks, got %d", len(expectedBlocks), len(blockNames))
	}

	for _, expected := range expectedBlocks {
		found := false
		for _, name := range blockNames {
			if name == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected block name %q not found in %v", expected, blockNames)
		}
	}

	// Test rendering individual blocks
	result, err := tmpl.RenderBlockToString("header", map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to render header block: %v", err)
	}

	expected := "<header>Default Header</header>"
	if strings.TrimSpace(result) != strings.TrimSpace(expected) {
		t.Errorf("Expected header block to render %q, got %q", expected, result)
	}
}

func TestBlockInheritanceCompilation(t *testing.T) {
	env := NewEnvironment()

	templates := map[string]string{
		"base.html": `<html>
{% block header %}<header>Base Header</header>{% endblock %}
{% block content %}<main>Base Content</main>{% endblock %}
</html>`,

		"child.html": `{% extends "base.html" %}
{% block header %}<header>Child Header</header>{% endblock %}
{% block content %}<main>Child Content</main>{% endblock %}`,
	}

	loader := NewMapLoader(templates)
	env.SetLoader(loader)

	tmpl, err := env.ParseFile("child.html")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	// The child template should have access to all blocks
	if !tmpl.HasBlock("header") {
		t.Error("Expected template to have 'header' block")
	}

	if !tmpl.HasBlock("content") {
		t.Error("Expected template to have 'content' block")
	}

	// Test rendering - should use child blocks
	result, err := tmpl.ExecuteToString(map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	if !strings.Contains(result, "Child Header") {
		t.Errorf("Expected result to contain 'Child Header', got %q", result)
	}

	if !strings.Contains(result, "Child Content") {
		t.Errorf("Expected result to contain 'Child Content', got %q", result)
	}

	if strings.Contains(result, "Base Header") {
		t.Errorf("Expected result not to contain 'Base Header', got %q", result)
	}

	if strings.Contains(result, "Base Content") {
		t.Errorf("Expected result not to contain 'Base Content', got %q", result)
	}
}

func TestBlockWithVariables(t *testing.T) {
	env := NewEnvironment()

	templates := map[string]string{
		"vars.html": `<html>
{% block header %}<h1>{{ title }}</h1>{% endblock %}
{% block content %}<p>{{ content }}</p>{% endblock %}
</html>`,
	}

	loader := NewMapLoader(templates)
	env.SetLoader(loader)

	tmpl, err := env.ParseFile("vars.html")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	context := map[string]interface{}{
		"title":   "Test Title",
		"content": "Test Content",
	}

	result, err := tmpl.ExecuteToString(context)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	if !strings.Contains(result, "Test Title") {
		t.Errorf("Expected result to contain 'Test Title', got %q", result)
	}

	if !strings.Contains(result, "Test Content") {
		t.Errorf("Expected result to contain 'Test Content', got %q", result)
	}
}

func TestNestedBlocks(t *testing.T) {
	env := NewEnvironment()

	templates := map[string]string{
		"nested.html": `<html>
{% block outer %}
<div class="outer">
    {% block inner %}
    <div class="inner">Inner Content</div>
    {% endblock %}
</div>
{% endblock %}
</html>`,
	}

	loader := NewMapLoader(templates)
	env.SetLoader(loader)

	tmpl, err := env.ParseFile("nested.html")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	result, err := tmpl.ExecuteToString(map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	if !strings.Contains(result, "div class=\"outer\"") {
		t.Errorf("Expected result to contain outer div, got %q", result)
	}

	if !strings.Contains(result, "div class=\"inner\"") {
		t.Errorf("Expected result to contain inner div, got %q", result)
	}

	if !strings.Contains(result, "Inner Content") {
		t.Errorf("Expected result to contain 'Inner Content', got %q", result)
	}
}

func TestBlockOverridingInheritance(t *testing.T) {
	env := NewEnvironment()

	templates := map[string]string{
		"base.html": `<html>
{% block header %}<header>Base Header</header>{% endblock %}
{% block content %}<main>Base Content</main>{% endblock %}
</html>`,

		"child.html": `{% extends "base.html" %}
{% block header %}<header>Child Header</header>{% endblock %}
{% block content %}<main>Child Content</main>{% endblock %}`,

		"grandchild.html": `{% extends "child.html" %}
{% block content %}<main>Grandchild Content</main>{% endblock %}`,
	}

	loader := NewMapLoader(templates)
	env.SetLoader(loader)

	tmpl, err := env.ParseFile("grandchild.html")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	result, err := tmpl.ExecuteToString(map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	// Should have child header (not overridden by grandchild)
	if !strings.Contains(result, "Child Header") {
		t.Errorf("Expected result to contain 'Child Header', got %q", result)
	}

	// Should have grandchild content (overridden by grandchild)
	if !strings.Contains(result, "Grandchild Content") {
		t.Errorf("Expected result to contain 'Grandchild Content', got %q", result)
	}

	// Should not have base or child versions that were overridden
	if strings.Contains(result, "Base Header") {
		t.Errorf("Expected result not to contain 'Base Header', got %q", result)
	}

	if strings.Contains(result, "Base Content") {
		t.Errorf("Expected result not to contain 'Base Content', got %q", result)
	}

	if strings.Contains(result, "Child Content") {
		t.Errorf("Expected result not to contain 'Child Content', got %q", result)
	}
}

func TestBlockRenderingWithDifferentContexts(t *testing.T) {
	env := NewEnvironment()

	templates := map[string]string{
		"context.html": `<p>{{ message }}</p>
{% block content %}<p>{{ block_message }}</p>{% endblock %}`,
	}

	loader := NewMapLoader(templates)
	env.SetLoader(loader)

	tmpl, err := env.ParseFile("context.html")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	// Test full template rendering
	fullResult, err := tmpl.ExecuteToString(map[string]interface{}{
		"message":       "Full Template",
		"block_message": "Block Content",
	})
	if err != nil {
		t.Fatalf("Failed to execute full template: %v", err)
	}

	if !strings.Contains(fullResult, "Full Template") {
		t.Errorf("Expected full result to contain 'Full Template', got %q", fullResult)
	}

	if !strings.Contains(fullResult, "Block Content") {
		t.Errorf("Expected full result to contain 'Block Content', got %q", fullResult)
	}

	// Test block rendering with different context
	blockResult, err := tmpl.RenderBlockToString("content", map[string]interface{}{
		"message":       "Block Template",
		"block_message": "Different Block Content",
	})
	if err != nil {
		t.Fatalf("Failed to render block: %v", err)
	}

	if strings.Contains(blockResult, "Full Template") {
		t.Errorf("Expected block result not to contain 'Full Template', got %q", blockResult)
	}

	if !strings.Contains(blockResult, "Different Block Content") {
		t.Errorf("Expected block result to contain 'Different Block Content', got %q", blockResult)
	}
}