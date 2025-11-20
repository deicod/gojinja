package runtime

import "testing"

func TestUndefinedFactoryDebugDefault(t *testing.T) {
	env := NewEnvironment()
	val := env.newUndefined("foo")
	if _, ok := val.(DebugUndefined); !ok {
		t.Fatalf("expected DebugUndefined, got %T", val)
	}
}

func TestUndefinedFactoryStrict(t *testing.T) {
	env := NewEnvironment()
	env.SetUndefinedFactory(func(name string) undefinedType {
		return StrictUndefined{}
	})

	val := env.newUndefined("foo")
	if _, ok := val.(StrictUndefined); !ok {
		t.Fatalf("expected StrictUndefined, got %T", val)
	}
}

func TestContextResolveCreatesUndefined(t *testing.T) {
	env := NewEnvironment()
	ctx := NewContextWithEnvironment(env, nil)

	value, err := ctx.Resolve("missing")
	if err != nil {
		t.Fatalf("Resolve error: %v", err)
	}

	if !isUndefinedValue(value) {
		t.Fatalf("expected undefined value, got %#v", value)
	}
}

func TestStrictUndefinedRaisesError(t *testing.T) {
	env := NewEnvironment()
	env.SetLoader(NewMapLoader(map[string]string{
		"main.html": `{{ missing }}`,
	}))
	env.SetUndefinedFactory(func(name string) undefinedType {
		return StrictUndefined{}
	})

	tmpl, err := env.ParseFile("main.html")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if _, err := tmpl.ExecuteToString(nil); err == nil {
		t.Fatalf("expected strict undefined error")
	}
}

func TestMissingAttributeResolvesToUndefined(t *testing.T) {
	env := NewEnvironment()
	tmpl, err := env.ParseString(`{{ user.name|default('anon') }}`, "attr_missing")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	output, err := tmpl.ExecuteToString(map[string]interface{}{})
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}

	if output != "anon" {
		t.Fatalf("expected default fallback, got %q", output)
	}
}

func TestChainableUndefinedPropagatesThroughLookups(t *testing.T) {
	env := NewEnvironment()
	env.SetUndefinedFactory(func(name string) undefinedType {
		return ChainableUndefined{name: name}
	})

	tmpl, err := env.ParseString(`{{ user.missing.attr|default('fallback') }}`, "chainable")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	output, err := tmpl.ExecuteToString(nil)
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}

	if output != "fallback" {
		t.Fatalf("expected chainable undefined to allow default, got %q", output)
	}
}

func TestMissingIndexReturnsUndefined(t *testing.T) {
	env := NewEnvironment()
	tmpl, err := env.ParseString(`{{ data['missing']|default('none') }}`, "index_missing")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	output, err := tmpl.ExecuteToString(map[string]interface{}{"data": map[string]interface{}{"present": "value"}})
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}

	if output != "none" {
		t.Fatalf("expected default for missing index, got %q", output)
	}
}
