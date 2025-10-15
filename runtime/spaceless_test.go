package runtime

import "testing"

func TestSpacelessRemovesWhitespaceBetweenTags(t *testing.T) {
	template := `{% spaceless %}
<div>
    <span>Hello</span>
</div>
{% endspaceless %}`

	out, err := ExecuteToString(template, nil)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}

	expected := `<div><span>Hello</span></div>`
	if out != expected {
		t.Fatalf("expected %q, got %q", expected, out)
	}
}

func TestSpacelessPreservesInnerText(t *testing.T) {
	template := `{% spaceless %}<p> hello   world </p>{% endspaceless %}`

	out, err := ExecuteToString(template, nil)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}

	expected := `<p> hello   world </p>`
	if out != expected {
		t.Fatalf("expected %q, got %q", expected, out)
	}
}

func TestSpacelessWithInterpolatedContent(t *testing.T) {
	template := `{% spaceless %}<div>{{ value }}</div>{% endspaceless %}`
	ctx := map[string]interface{}{"value": "Hi"}

	out, err := ExecuteToString(template, ctx)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}

	expected := `<div>Hi</div>`
	if out != expected {
		t.Fatalf("expected %q, got %q", expected, out)
	}
}
