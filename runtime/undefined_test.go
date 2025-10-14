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
