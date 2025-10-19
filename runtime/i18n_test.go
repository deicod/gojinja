package runtime

import (
	"strconv"
	"strings"
	"testing"
)

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

func TestTransPluralizeAlias(t *testing.T) {
	tpl := "{% trans user_count=users|length %}{{ user_count }} user{% pluralize user_count %}{{ user_count }} users{% endtrans %}"

	one, err := ExecuteToString(tpl, map[string]interface{}{"users": []int{1}})
	if err != nil {
		t.Fatalf("unexpected error rendering singular: %v", err)
	}
	if one != "1 user" {
		t.Fatalf("expected '1 user', got %q", one)
	}

	many, err := ExecuteToString(tpl, map[string]interface{}{"users": []int{1, 2, 3}})
	if err != nil {
		t.Fatalf("unexpected error rendering plural: %v", err)
	}
	if many != "3 users" {
		t.Fatalf("expected '3 users', got %q", many)
	}
}

func TestTransTrimmedOption(t *testing.T) {
	tpl := `{% trans trimmed %}
  Hello   {{ name }}
{% endtrans %}`

	result, err := ExecuteToString(tpl, map[string]interface{}{"name": "World"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "Hello World" {
		t.Fatalf("expected trimmed output 'Hello World', got %q", result)
	}
}

func TestTransTrimmedPolicyAndOverride(t *testing.T) {
	env := NewEnvironment()
	env.SetI18nTrimmed(true)

	tpl, err := env.ParseString(`{% trans %}
  Hello   {{ name }}
{% endtrans %}`, "trimmed_policy")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	result, err := tpl.ExecuteToString(map[string]interface{}{"name": "World"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "Hello World" {
		t.Fatalf("expected policy trimmed output 'Hello World', got %q", result)
	}

	override, err := env.ParseString(`{% trans notrimmed %}
  Hello   {{ name }}
{% endtrans %}`, "notrimmed_override")
	if err != nil {
		t.Fatalf("parse error for override: %v", err)
	}

	overrideResult, err := override.ExecuteToString(map[string]interface{}{"name": "World"})
	if err != nil {
		t.Fatalf("unexpected error rendering override: %v", err)
	}
	if !strings.Contains(overrideResult, "\n") {
		t.Fatalf("expected override output to retain newlines, got %q", overrideResult)
	}
}

func TestTransContextUsesPGettext(t *testing.T) {
	env := NewEnvironment()
	env.AddGlobal("pgettext", GlobalFunc(func(ctx *Context, args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			t.Fatalf("expected context and message")
		}
		return args[0].(string) + ":" + args[1].(string), nil
	}))

	tpl, err := env.ParseString(`{% trans 'email' %}Reset{% endtrans %}`, "contextual")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	result, err := tpl.ExecuteToString(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "email:Reset" {
		t.Fatalf("expected context-aware translation, got %q", result)
	}
}

func TestTransContextPluralUsesNPGettext(t *testing.T) {
	env := NewEnvironment()
	env.AddGlobal("npgettext", GlobalFunc(func(ctx *Context, args ...interface{}) (interface{}, error) {
		if len(args) < 4 {
			t.Fatalf("expected context, singular, plural, count")
		}
		context := args[0].(string)
		singular := args[1].(string)
		plural := args[2].(string)

		count := 0
		switch v := args[3].(type) {
		case int:
			count = v
		case int8:
			count = int(v)
		case int16:
			count = int(v)
		case int32:
			count = int(v)
		case int64:
			count = int(v)
		case uint:
			count = int(v)
		case uint8:
			count = int(v)
		case uint16:
			count = int(v)
		case uint32:
			count = int(v)
		case uint64:
			count = int(v)
		default:
			t.Fatalf("unexpected count type %T", args[3])
		}

		message := singular
		if count != 1 {
			message = plural
		}
		message = strings.ReplaceAll(message, "%(count)s", strconv.Itoa(count))
		return context + ":" + message, nil
	}))

	tpl, err := env.ParseString(`{% trans 'status' count=items|length %}{{ count }} item{% pluralize %}{{ count }} items{% endtrans %}`, "context_plural")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	single, err := tpl.ExecuteToString(map[string]interface{}{"items": []int{1}})
	if err != nil {
		t.Fatalf("unexpected error rendering singular: %v", err)
	}
	if single != "status:1 item" {
		t.Fatalf("expected context-aware singular, got %q", single)
	}

	many, err := tpl.ExecuteToString(map[string]interface{}{"items": []int{1, 2}})
	if err != nil {
		t.Fatalf("unexpected error rendering plural: %v", err)
	}
	if many != "status:2 items" {
		t.Fatalf("expected context-aware plural, got %q", many)
	}
}
