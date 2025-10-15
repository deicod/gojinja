package runtime

import "testing"

func TestTransSimple(t *testing.T) {
	tpl := "{% trans %}Hello {{ name }}{% endtrans %}"
	result, err := ExecuteToString(tpl, map[string]interface{}{"name": "World"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "Hello World" {
		t.Fatalf("expected 'Hello World', got %q", result)
	}
}

func TestTransAssignments(t *testing.T) {
	tpl := "{% trans user_name=user.name %}Hi {{ user_name }}{% endtrans %}"
	ctx := map[string]interface{}{
		"user": map[string]interface{}{"name": "Alice"},
	}
	result, err := ExecuteToString(tpl, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "Hi Alice" {
		t.Fatalf("expected 'Hi Alice', got %q", result)
	}
}

func TestTransPluralize(t *testing.T) {
	tpl := "{% trans count=items|length %}{{ count }} apple{% pluralize %}{{ count }} apples{% endtrans %}"

	one, err := ExecuteToString(tpl, map[string]interface{}{"items": []int{1}})
	if err != nil {
		t.Fatalf("unexpected error rendering singular: %v", err)
	}
	if one != "1 apple" {
		t.Fatalf("expected '1 apple', got %q", one)
	}

	many, err := ExecuteToString(tpl, map[string]interface{}{"items": []int{1, 2, 3}})
	if err != nil {
		t.Fatalf("unexpected error rendering plural: %v", err)
	}
	if many != "3 apples" {
		t.Fatalf("expected '3 apples', got %q", many)
	}
}

func TestTransPluralizeWithoutCountFails(t *testing.T) {
	tpl := "{% trans %}one{% pluralize %}many{% endtrans %}"
	if _, err := ExecuteToString(tpl, nil); err == nil {
		t.Fatalf("expected error for pluralize without count")
	}
}
