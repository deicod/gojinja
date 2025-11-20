package runtime

import (
	"fmt"
	"strings"
	"testing"

	"github.com/deicod/gojinja/parser"
)

type testAwaitable struct {
	result interface{}
	err    error
}

func (a testAwaitable) Await(ctx *Context) (interface{}, error) {
	return a.result, a.err
}

func TestBasicRendering(t *testing.T) {
	tests := []struct {
		name     string
		template string
		ctx      map[string]interface{}
		expected string
	}{
		{
			name:     "simple text",
			template: "Hello World",
			ctx:      nil,
			expected: "Hello World",
		},
		{
			name:     "simple variable",
			template: "Hello {{ name }}!",
			ctx:      map[string]interface{}{"name": "World"},
			expected: "Hello World!",
		},
		{
			name:     "multiple variables",
			template: "{{ greeting }} {{ name }}!",
			ctx:      map[string]interface{}{"greeting": "Hello", "name": "World"},
			expected: "Hello World!",
		},
		{
			name:     "variable with filter",
			template: "Hello {{ name|upper }}!",
			ctx:      map[string]interface{}{"name": "world"},
			expected: "Hello WORLD!",
		},
		{
			name:     "multiple filters",
			template: "{{ name|trim|upper }}",
			ctx:      map[string]interface{}{"name": "  world  "},
			expected: "WORLD",
		},
		{
			name:     "for loop",
			template: "{% for item in items %}{{ item }} {% endfor %}",
			ctx:      map[string]interface{}{"items": []interface{}{"a", "b", "c"}},
			expected: "a b c ",
		},
		{
			name:     "for loop with empty",
			template: "{% for item in items %}{{ item }}{% else %}empty{% endfor %}",
			ctx:      map[string]interface{}{"items": []interface{}{}},
			expected: "empty",
		},
		{
			name:     "if statement",
			template: "{% if show %}visible{% else %}hidden{% endif %}",
			ctx:      map[string]interface{}{"show": true},
			expected: "visible",
		},
		{
			name:     "if statement false",
			template: "{% if show %}visible{% else %}hidden{% endif %}",
			ctx:      map[string]interface{}{"show": false},
			expected: "hidden",
		},
		{
			name:     "if elif else",
			template: "{% if value > 10 %}high{% elif value > 5 %}medium{% else %}low{% endif %}",
			ctx:      map[string]interface{}{"value": 7},
			expected: "medium",
		},
		{
			name:     "set statement",
			template: "{% set x = 42 %}{{ x }}",
			ctx:      nil,
			expected: "42",
		},
		{
			name:     "arithmetic",
			template: "{{ 2 + 3 }} {{ 10 - 4 }} {{ 3 * 4 }} {{ 15 / 3 }}",
			ctx:      nil,
			expected: "5 6 12 5",
		},
		{
			name:     "floor division semantics",
			template: "{{ -3 // 2 }} {{ 5 // -2 }} {{ -3.5 // 2 }}",
			ctx:      nil,
			expected: "-2 -3 -2",
		},
		{
			name:     "modulo semantics",
			template: "{{ -3 % 2 }} {{ 5 % -2 }} {{ -3 % -2 }} {{ -3.0 % 2 }}",
			ctx:      nil,
			expected: "1 -1 -1 1",
		},
		{
			name:     "numeric coercion",
			template: "{{ a + b }} {{ c * d }} {{ e / 2 }} {{ f ** g }}",
			ctx: map[string]interface{}{
				"a": int8(5),
				"b": uint16(7),
				"c": uint8(3),
				"d": int32(4),
				"e": uint32(5),
				"f": uint8(3),
				"g": uint8(2),
			},
			expected: "12 12 2.5 9",
		},
		{
			name:     "string repetition with unsigned",
			template: "{{ 'ha' * repeat }}",
			ctx:      map[string]interface{}{"repeat": uint8(2)},
			expected: "haha",
		},
		{
			name:     "negative exponent produces float",
			template: "{{ 2 ** -1 }}",
			ctx:      nil,
			expected: "0.5",
		},
		{
			name:     "string concatenation",
			template: "{{ 'Hello' + ' ' + 'World' }}",
			ctx:      nil,
			expected: "Hello World",
		},
		{
			name:     "list literal",
			template: "{{ [1, 2, 3]|length }}",
			ctx:      nil,
			expected: "3",
		},
		{
			name:     "dict literal",
			template: "{{ {'a': 1, 'b': 2}.a }}",
			ctx:      nil,
			expected: "1",
		},
		{
			name:     "comparison",
			template: "{% if 5 > 3 %}yes{% endif %}",
			ctx:      nil,
			expected: "yes",
		},
		{
			name:     "logical and",
			template: "{% if true and false %}yes{% else %}no{% endif %}",
			ctx:      nil,
			expected: "no",
		},
		{
			name:     "logical or",
			template: "{% if true or false %}yes{% else %}no{% endif %}",
			ctx:      nil,
			expected: "yes",
		},
		{
			name:     "not operator",
			template: "{% if not false %}yes{% endif %}",
			ctx:      nil,
			expected: "yes",
		},
		{
			name:     "range function",
			template: "{% for i in range(3) %}{{ i }}{% endfor %}",
			ctx:      nil,
			expected: "012",
		},
		{
			name:     "loop variable",
			template: "{% for item in ['a', 'b'] %}{{ loop.index }}:{{ item }} {% endfor %}",
			ctx:      nil,
			expected: "1:a 2:b ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExecuteToString(tt.template, tt.ctx)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestFilters(t *testing.T) {
	tests := []struct {
		name     string
		template string
		ctx      map[string]interface{}
		expected string
	}{
		{
			name:     "upper filter",
			template: "{{ 'hello'|upper }}",
			ctx:      nil,
			expected: "HELLO",
		},
		{
			name:     "lower filter",
			template: "{{ 'HELLO'|lower }}",
			ctx:      nil,
			expected: "hello",
		},
		{
			name:     "capitalize filter",
			template: "{{ 'hello world'|capitalize }}",
			ctx:      nil,
			expected: "Hello world",
		},
		{
			name:     "title filter",
			template: "{{ 'hello world'|title }}",
			ctx:      nil,
			expected: "Hello World",
		},
		{
			name:     "trim filter",
			template: "{{ '  hello  '|trim }}",
			ctx:      nil,
			expected: "hello",
		},
		{
			name:     "length filter on string",
			template: "{{ 'hello'|length }}",
			ctx:      nil,
			expected: "5",
		},
		{
			name:     "length filter on list",
			template: "{{ [1, 2, 3, 4]|length }}",
			ctx:      nil,
			expected: "4",
		},
		{
			name:     "first filter",
			template: "{{ ['a', 'b', 'c']|first }}",
			ctx:      nil,
			expected: "a",
		},
		{
			name:     "last filter",
			template: "{{ ['a', 'b', 'c']|last }}",
			ctx:      nil,
			expected: "c",
		},
		{
			name:     "join filter",
			template: "{{ ['a', 'b', 'c']|join(', ') }}",
			ctx:      nil,
			expected: "a, b, c",
		},
		{
			name:     "default filter",
			template: "{{ name|default('Anonymous') }}",
			ctx:      map[string]interface{}{},
			expected: "Anonymous",
		},
		{
			name:     "default filter with value",
			template: "{{ name|default('Anonymous') }}",
			ctx:      map[string]interface{}{"name": "John"},
			expected: "John",
		},
		{
			name:     "round filter",
			template: "{{ 3.14159|round(2) }}",
			ctx:      nil,
			expected: "3.14",
		},
		{
			name:     "abs filter",
			template: "{{ -5|abs }}",
			ctx:      nil,
			expected: "5",
		},
		{
			name:     "sort filter",
			template: "{{ [3, 1, 4, 1, 5]|sort|join(',') }}",
			ctx:      nil,
			expected: "1,1,3,4,5",
		},
		{
			name:     "unique filter",
			template: "{{ [1, 2, 2, 3, 1]|unique|sort|join(',') }}",
			ctx:      nil,
			expected: "1,2,3",
		},
		{
			name:     "reverse filter",
			template: "{{ [1, 2, 3]|reverse|join(',') }}",
			ctx:      nil,
			expected: "3,2,1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExecuteToString(tt.template, tt.ctx)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestExpressions(t *testing.T) {
	tests := []struct {
		name     string
		template string
		ctx      map[string]interface{}
		expected string
	}{
		{
			name:     "attribute access",
			template: "{{ user.name }}",
			ctx: map[string]interface{}{
				"user": map[string]interface{}{"name": "John"},
			},
			expected: "John",
		},
		{
			name:     "nested attribute access",
			template: "{{ user.profile.name }}",
			ctx: map[string]interface{}{
				"user": map[string]interface{}{
					"profile": map[string]interface{}{"name": "John"},
				},
			},
			expected: "John",
		},
		{
			name:     "index access",
			template: "{{ items[1] }}",
			ctx: map[string]interface{}{
				"items": []interface{}{"a", "b", "c"},
			},
			expected: "b",
		},
		{
			name:     "string index access",
			template: "{{ text[1] }}",
			ctx: map[string]interface{}{
				"text": "hello",
			},
			expected: "e",
		},
		{
			name:     "negative index",
			template: "{{ items[-1] }}",
			ctx: map[string]interface{}{
				"items": []interface{}{"a", "b", "c"},
			},
			expected: "c",
		},
		{
			name:     "map key access",
			template: "{{ data['key'] }}",
			ctx: map[string]interface{}{
				"data": map[string]interface{}{"key": "value"},
			},
			expected: "value",
		},
		{
			name:     "method call",
			template: "{{ text.upper() }}",
			ctx: map[string]interface{}{
				"text": strings.ToLower("HELLO"),
			},
			expected: "HELLO",
		},
		{
			name:     "function call with args",
			template: "{{ range(3) }}",
			ctx:      nil,
			expected: "[0 1 2]",
		},
		{
			name:     "dict function",
			template: "{{ dict('a', 1, 'b', 2).a }}",
			ctx:      nil,
			expected: "1",
		},
		{
			name:     "conditional expression",
			template: "{{ 'yes' if true else 'no' }}",
			ctx:      nil,
			expected: "yes",
		},
		{
			name:     "in operator",
			template: "{% if 3 in [1, 2, 3] %}yes{% endif %}",
			ctx:      nil,
			expected: "yes",
		},
		{
			name:     "not in operator",
			template: "{% if 4 not in [1, 2, 3] %}yes{% endif %}",
			ctx:      nil,
			expected: "yes",
		},
		{
			name:     "chain comparisons",
			template: "{% if 1 < 2 < 3 %}yes{% endif %}",
			ctx:      nil,
			expected: "yes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExecuteToString(tt.template, tt.ctx)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		name     string
		template string
		ctx      map[string]interface{}
		contains string
		strict   bool
	}{
		{
			name:     "undefined variable",
			template: "{{ undefined_var }}",
			ctx:      nil,
			contains: "undefined",
			strict:   true,
		},
		{
			name:     "division by zero",
			template: "{{ 1 / 0 }}",
			ctx:      nil,
			contains: "division by zero",
		},
		{
			name:     "unknown filter",
			template: "{{ 'hello'|unknown_filter }}",
			ctx:      nil,
			contains: "unknown filter",
		},
		{
			name:     "unknown test",
			template: "{% if 'hello' is unknown_test %}yes{% endif %}",
			ctx:      nil,
			contains: "unknown test",
		},
		{
			name:     "invalid attribute",
			template: "{{ 'hello'.invalid_attr }}",
			ctx:      nil,
			contains: "undefined",
		},
		{
			name:     "index out of bounds",
			template: "{{ [1, 2, 3][10] }}",
			ctx:      nil,
			contains: "out of range",
		},
		{
			name:     "invalid operation",
			template: "{{ 'hello' + 5 }}",
			ctx:      nil,
			contains: "unsupported operand",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if tt.strict {
				env := NewEnvironment()
				env.SetUndefinedFactory(func(name string) undefinedType { return StrictUndefined{name: name} })
				template, parseErr := env.ParseString(tt.template, "test")
				if parseErr != nil {
					t.Fatalf("parse error: %v", parseErr)
				}
				_, err = template.ExecuteToString(tt.ctx)
			} else {
				_, err = ExecuteToString(tt.template, tt.ctx)
			}
			if err == nil {
				t.Errorf("Expected error containing %q, but got no error", tt.contains)
				return
			}
			if !strings.Contains(err.Error(), tt.contains) {
				t.Errorf("Expected error containing %q, got: %v", tt.contains, err)
			}
		})
	}
}

func TestEnvironmentFeatures(t *testing.T) {
	t.Run("custom filter", func(t *testing.T) {
		env := NewEnvironment()
		env.AddFilter("reverse", func(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
			if str, ok := value.(string); ok {
				runes := []rune(str)
				for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
					runes[i], runes[j] = runes[j], runes[i]
				}
				return string(runes), nil
			}
			return value, nil
		})

		template, err := env.ParseString("{{ 'hello'|reverse }}", "test")
		if err != nil {
			t.Fatalf("Failed to parse template: %v", err)
		}

		result, err := template.ExecuteToString(nil)
		if err != nil {
			t.Fatalf("Failed to execute template: %v", err)
		}

		expected := "olleh"
		if result != expected {
			t.Errorf("Expected %q, got %q", expected, result)
		}
	})

	t.Run("custom global function", func(t *testing.T) {
		env := NewEnvironment()
		env.AddGlobal("greet", func(ctx *Context, args ...interface{}) (interface{}, error) {
			if len(args) > 0 {
				if name, ok := args[0].(string); ok {
					return "Hello, " + name + "!", nil
				}
			}
			return "Hello, World!", nil
		})

		template, err := env.ParseString("{{ greet('Alice') }}", "test")
		if err != nil {
			t.Fatalf("Failed to parse template: %v", err)
		}

		result, err := template.ExecuteToString(nil)
		if err != nil {
			t.Fatalf("Failed to execute template: %v", err)
		}

		expected := "Hello, Alice!"
		if result != expected {
			t.Errorf("Expected %q, got %q", expected, result)
		}
	})

	t.Run("autoescape enabled", func(t *testing.T) {
		env := NewEnvironment()
		env.SetAutoescape(true)

		template, err := env.ParseString("{{ '<script>alert(\"xss\")</script>' }}", "test.html")
		if err != nil {
			t.Fatalf("Failed to parse template: %v", err)
		}

		result, err := template.ExecuteToString(nil)
		if err != nil {
			t.Fatalf("Failed to execute template: %v", err)
		}

		expected := "&lt;script&gt;alert(&#34;xss&#34;)&lt;/script&gt;"
		if result != expected {
			t.Errorf("Expected %q, got %q", expected, result)
		}
	})

	t.Run("autoescape disabled", func(t *testing.T) {
		env := NewEnvironment()
		env.SetAutoescape(false)

		template, err := env.ParseString("{{ '<script>alert(\"xss\")</script>' }}", "test.txt")
		if err != nil {
			t.Fatalf("Failed to parse template: %v", err)
		}

		result, err := template.ExecuteToString(nil)
		if err != nil {
			t.Fatalf("Failed to execute template: %v", err)
		}

		expected := "<script>alert(\"xss\")</script>"
		if result != expected {
			t.Errorf("Expected %q, got %q", expected, result)
		}
	})
}

func TestAsyncStatements(t *testing.T) {
	env := NewEnvironment()

	if _, err := env.ParseString(`{% async for item in items %}{% endfor %}`, "test"); err == nil {
		t.Fatalf("expected parse error for async statements when enable_async is disabled")
	}

	env.SetEnableAsync(true)

	tmpl, err := env.ParseString(`{% async for item in items %}{{ item }}{% endfor %}`, "async_for")
	if err != nil {
		t.Fatalf("failed to parse async for template: %v", err)
	}

	result, err := tmpl.ExecuteToString(map[string]interface{}{
		"items": []int{1, 2, 3},
	})
	if err != nil {
		t.Fatalf("failed to execute async for template: %v", err)
	}
	if strings.TrimSpace(result) != "123" {
		t.Fatalf("expected concatenated output '123', got %q", strings.TrimSpace(result))
	}

	withTemplate, err := env.ParseString(`{% async with value = 5 %}{{ value }}{% endwith %}`, "async_with")
	if err != nil {
		t.Fatalf("failed to parse async with template: %v", err)
	}
	withResult, err := withTemplate.ExecuteToString(nil)
	if err != nil {
		t.Fatalf("failed to execute async with template: %v", err)
	}
	if strings.TrimSpace(withResult) != "5" {
		t.Fatalf("expected output '5', got %q", strings.TrimSpace(withResult))
	}
}

func TestAwaitExpressions(t *testing.T) {
	env := NewEnvironment()

	if _, err := env.ParseString(`{{ await value }}`, "await_disabled"); err == nil {
		t.Fatalf("expected parse error for await when enable_async is disabled")
	}

	env.SetEnableAsync(true)

	tmpl, err := env.ParseString(`{{ await value }}`, "await_value")
	if err != nil {
		t.Fatalf("failed to parse await template: %v", err)
	}

	awaited := testAwaitable{result: "done"}
	output, err := tmpl.ExecuteToString(map[string]interface{}{"value": awaited})
	if err != nil {
		t.Fatalf("failed to execute await template: %v", err)
	}
	if strings.TrimSpace(output) != "done" {
		t.Fatalf("expected awaited value 'done', got %q", strings.TrimSpace(output))
	}

	tmplFunc, err := env.ParseString(`{{ await loader }}`, "await_func")
	if err != nil {
		t.Fatalf("failed to parse await function template: %v", err)
	}

	funcResult, err := tmplFunc.ExecuteToString(map[string]interface{}{"loader": func() (interface{}, error) {
		return "loaded", nil
	}})
	if err != nil {
		t.Fatalf("failed to execute await function template: %v", err)
	}
	if strings.TrimSpace(funcResult) != "loaded" {
		t.Fatalf("expected awaited function result 'loaded', got %q", strings.TrimSpace(funcResult))
	}

	ctxFuncResult, err := tmplFunc.ExecuteToString(map[string]interface{}{"loader": func(ctx *Context) (interface{}, error) {
		if ctx == nil {
			return nil, fmt.Errorf("missing context")
		}
		return "ctx", nil
	}})
	if err != nil {
		t.Fatalf("failed to execute await context function template: %v", err)
	}
	if strings.TrimSpace(ctxFuncResult) != "ctx" {
		t.Fatalf("expected awaited context function result 'ctx', got %q", strings.TrimSpace(ctxFuncResult))
	}

	typedFuncResult, err := tmplFunc.ExecuteToString(map[string]interface{}{"loader": func() (string, error) {
		return "typed", nil
	}})
	if err != nil {
		t.Fatalf("failed to execute await typed function template: %v", err)
	}
	if strings.TrimSpace(typedFuncResult) != "typed" {
		t.Fatalf("expected awaited typed function result 'typed', got %q", strings.TrimSpace(typedFuncResult))
	}

	typedCtxResult, err := tmplFunc.ExecuteToString(map[string]interface{}{"loader": func(ctx *Context) (string, error) {
		if ctx == nil {
			return "", fmt.Errorf("missing context")
		}
		return "typed-ctx", nil
	}})
	if err != nil {
		t.Fatalf("failed to execute await typed context function template: %v", err)
	}
	if strings.TrimSpace(typedCtxResult) != "typed-ctx" {
		t.Fatalf("expected awaited typed context function result 'typed-ctx', got %q", strings.TrimSpace(typedCtxResult))
	}

	valueOnlyResult, err := tmplFunc.ExecuteToString(map[string]interface{}{"loader": func() string {
		return "only-value"
	}})
	if err != nil {
		t.Fatalf("failed to execute await value-only function template: %v", err)
	}
	if strings.TrimSpace(valueOnlyResult) != "only-value" {
		t.Fatalf("expected awaited value-only function result 'only-value', got %q", strings.TrimSpace(valueOnlyResult))
	}

	tmplError, err := env.ParseString(`{{ await value }}`, "await_error")
	if err != nil {
		t.Fatalf("failed to parse await error template: %v", err)
	}

	_, execErr := tmplError.ExecuteToString(map[string]interface{}{"value": testAwaitable{err: fmt.Errorf("boom")}})
	if execErr == nil {
		t.Fatalf("expected error when awaitable returns error")
	}
	if !strings.Contains(execErr.Error(), "boom") {
		t.Fatalf("expected error to contain original message, got %v", execErr)
	}

	if _, err := tmplFunc.ExecuteToString(map[string]interface{}{"loader": func() (string, error) {
		return "", fmt.Errorf("typed boom")
	}}); err == nil {
		t.Fatalf("expected error when typed awaitable returns error")
	} else if !strings.Contains(err.Error(), "typed boom") {
		t.Fatalf("expected typed awaitable error to contain original message, got %v", err)
	}

	tmplInvalid, err := env.ParseString(`{{ await value }}`, "await_invalid")
	if err != nil {
		t.Fatalf("failed to parse await invalid template: %v", err)
	}

	if _, err := tmplInvalid.ExecuteToString(map[string]interface{}{"value": 42}); err == nil {
		t.Fatalf("expected error when awaiting non-awaitable value")
	}

	tmplNone, err := env.ParseString(`{{ await none }}`, "await_none")
	if err != nil {
		t.Fatalf("failed to parse await none template: %v", err)
	}

	if _, err := tmplNone.ExecuteToString(nil); err == nil {
		t.Fatalf("expected error when awaiting literal none")
	} else if !strings.Contains(err.Error(), "not awaitable") {
		t.Fatalf("expected await error to mention non-awaitable value, got %v", err)
	}
}

func TestAutoAwaitLookups(t *testing.T) {
	env := NewEnvironment()
	env.SetEnableAsync(true)

	tmpl, err := env.ParseString(`{{ value }}|{{ obj.field }}|{{ mapping['key'] }}`, "auto_await")
	if err != nil {
		t.Fatalf("failed to parse auto-await template: %v", err)
	}

	ctx := map[string]interface{}{
		"value":   testAwaitable{result: "alpha"},
		"obj":     map[string]interface{}{"field": testAwaitable{result: "beta"}},
		"mapping": map[string]interface{}{"key": testAwaitable{result: "gamma"}},
	}

	output, err := tmpl.ExecuteToString(ctx)
	if err != nil {
		t.Fatalf("failed to execute auto-await template: %v", err)
	}

	if strings.TrimSpace(output) != "alpha|beta|gamma" {
		t.Fatalf("expected awaited lookups, got %q", strings.TrimSpace(output))
	}
}

func TestAsyncFiltersTestsAndGlobals(t *testing.T) {
	env := NewEnvironment()
	env.SetEnableAsync(true)

	env.AddFilter("defer", func(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
		return testAwaitable{result: fmt.Sprintf("%v!", value)}, nil
	})

	env.AddTest("async_truthy", func(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
		return testAwaitable{result: value == "go"}, nil
	})

	env.AddGlobal("async_value", func(ctx *Context, args ...interface{}) (interface{}, error) {
		return testAwaitable{result: "global"}, nil
	})

	tmplFilter, err := env.ParseString("{{ 'hello'|defer }}", "async_filter")
	if err != nil {
		t.Fatalf("failed to parse async filter template: %v", err)
	}

	filterOutput, err := tmplFilter.ExecuteToString(nil)
	if err != nil {
		t.Fatalf("failed to execute async filter template: %v", err)
	}
	if strings.TrimSpace(filterOutput) != "hello!" {
		t.Fatalf("expected awaited filter result 'hello!', got %q", strings.TrimSpace(filterOutput))
	}

	tmplTest, err := env.ParseString("{% if subject is async_truthy %}yes{% else %}no{% endif %}", "async_test")
	if err != nil {
		t.Fatalf("failed to parse async test template: %v", err)
	}

	testOutput, err := tmplTest.ExecuteToString(map[string]interface{}{"subject": "go"})
	if err != nil {
		t.Fatalf("failed to execute async test template: %v", err)
	}
	if strings.TrimSpace(testOutput) != "yes" {
		t.Fatalf("expected awaited test result 'yes', got %q", strings.TrimSpace(testOutput))
	}

	tmplGlobal, err := env.ParseString("{{ async_value() }}", "async_global")
	if err != nil {
		t.Fatalf("failed to parse async global template: %v", err)
	}

	globalOutput, err := tmplGlobal.ExecuteToString(nil)
	if err != nil {
		t.Fatalf("failed to execute async global template: %v", err)
	}
	if strings.TrimSpace(globalOutput) != "global" {
		t.Fatalf("expected awaited global result 'global', got %q", strings.TrimSpace(globalOutput))
	}
}

func TestContextScoping(t *testing.T) {
	t.Run("variable shadowing", func(t *testing.T) {
		template := "{% set x = 'outer' %}{{ x }}{% set x = 'inner' %}{{ x }}"
		result, err := ExecuteToString(template, nil)
		if err != nil {
			t.Fatalf("Failed to execute template: %v", err)
		}

		expected := "outerinner"
		if result != expected {
			t.Errorf("Expected %q, got %q", expected, result)
		}
	})

	t.Run("for loop scoping", func(t *testing.T) {
		template := "{% set x = 'global' %}{{ x }}{% for x in [1, 2] %}{{ x }}{% endfor %}{{ x }}"
		result, err := ExecuteToString(template, nil)
		if err != nil {
			t.Fatalf("Failed to execute template: %v", err)
		}

		expected := "global12global"
		if result != expected {
			t.Errorf("Expected %q, got %q", expected, result)
		}
	})

	t.Run("nested for loops", func(t *testing.T) {
		template := "{% for i in range(2) %}{% for j in range(2) %}{{ i }}{{ j }} {% endfor %}{% endfor %}"
		result, err := ExecuteToString(template, nil)
		if err != nil {
			t.Fatalf("Failed to execute template: %v", err)
		}

		expected := "00 01 10 11 "
		if result != expected {
			t.Errorf("Expected %q, got %q", expected, result)
		}
	})
}

func TestTemplateAPI(t *testing.T) {
	t.Run("template from string", func(t *testing.T) {
		template, err := ParseString("Hello {{ name }}!")
		if err != nil {
			t.Fatalf("Failed to create template: %v", err)
		}

		if template.Name() != "template" {
			t.Errorf("Expected template name 'template', got %q", template.Name())
		}

		result, err := template.ExecuteToString(map[string]interface{}{"name": "World"})
		if err != nil {
			t.Fatalf("Failed to execute template: %v", err)
		}

		expected := "Hello World!"
		if result != expected {
			t.Errorf("Expected %q, got %q", expected, result)
		}
	})

	t.Run("template from AST", func(t *testing.T) {
		ast, err := parser.ParseTemplate("Hello {{ name }}!")
		if err != nil {
			t.Fatalf("Failed to parse template: %v", err)
		}

		template, err := ParseASTWithName(ast, "greeting")
		if err != nil {
			t.Fatalf("Failed to create template: %v", err)
		}

		if template.Name() != "greeting" {
			t.Errorf("Expected template name 'greeting', got %q", template.Name())
		}

		result, err := template.ExecuteToString(map[string]interface{}{"name": "World"})
		if err != nil {
			t.Fatalf("Failed to execute template: %v", err)
		}

		expected := "Hello World!"
		if result != expected {
			t.Errorf("Expected %q, got %q", expected, result)
		}
	})

	t.Run("batch renderer", func(t *testing.T) {
		renderer := NewBatchRenderer(NewEnvironment())

		err := renderer.AddTemplate("greeting", "Hello {{ name }}!")
		if err != nil {
			t.Fatalf("Failed to add template: %v", err)
		}

		err = renderer.AddTemplate("farewell", "Goodbye {{ name }}!")
		if err != nil {
			t.Fatalf("Failed to add template: %v", err)
		}

		if renderer.Size() != 2 {
			t.Errorf("Expected 2 templates, got %d", renderer.Size())
		}

		result1, err := renderer.Render("greeting", map[string]interface{}{"name": "World"})
		if err != nil {
			t.Fatalf("Failed to render greeting: %v", err)
		}

		result2, err := renderer.Render("farewell", map[string]interface{}{"name": "World"})
		if err != nil {
			t.Fatalf("Failed to render farewell: %v", err)
		}

		expected1 := "Hello World!"
		expected2 := "Goodbye World!"

		if result1 != expected1 {
			t.Errorf("Expected %q, got %q", expected1, result1)
		}

		if result2 != expected2 {
			t.Errorf("Expected %q, got %q", expected2, result2)
		}
	})
}

func TestComplexTemplate(t *testing.T) {
	template := `
<!DOCTYPE html>
<html>
<head>
    <title>{{ title|title }}</title>
</head>
<body>
    <h1>{{ heading|upper }}</h1>

    {% if user %}
    <p>Welcome, {{ user.name|capitalize }}!</p>
    {% endif %}

    {% if items %}
    <ul>
    {% for item in items %}
        <li>{{ loop.index }}. {{ item.title }} ({{ item.count }} item{% if item.count != 1 %}s{% endif %})</li>
    {% endfor %}
    </ul>
    {% else %}
    <p>No items found.</p>
    {% endif %}

    <p>Total: {{ items|length }} items</p>
</body>
</html>
`

	ctx := map[string]interface{}{
		"title":   "welcome to my site",
		"heading": "dashboard",
		"user": map[string]interface{}{
			"name": "john doe",
		},
		"items": []interface{}{
			map[string]interface{}{"title": "First Item", "count": 1},
			map[string]interface{}{"title": "Second Item", "count": 3},
			map[string]interface{}{"title": "Third Item", "count": 0},
		},
	}

	result, err := ExecuteToString(template, ctx)
	if err != nil {
		t.Fatalf("Failed to execute complex template: %v", err)
	}

	// Check for key content
	expectedParts := []string{
		"<title>Welcome To My Site</title>",
		"<h1>DASHBOARD</h1>",
		"Welcome, John doe!",
		"<li>1. First Item (1 item)</li>",
		"<li>2. Second Item (3 items)</li>",
		"<li>3. Third Item (0 items)</li>",
		"Total: 3 items",
	}

	for _, part := range expectedParts {
		if !strings.Contains(result, part) {
			t.Errorf("Expected result to contain %q, but it didn't. Result:\n%s", part, result)
		}
	}
}

// Benchmark tests
func BenchmarkSimpleTemplate(b *testing.B) {
	template := "Hello {{ name }}!"
	ctx := map[string]interface{}{"name": "World"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ExecuteToString(template, ctx)
		if err != nil {
			b.Fatalf("Failed to execute template: %v", err)
		}
	}
}

func BenchmarkForLoop(b *testing.B) {
	template := "{% for i in range(10) %}{{ i }} {% endfor %}"
	ctx := map[string]interface{}{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ExecuteToString(template, ctx)
		if err != nil {
			b.Fatalf("Failed to execute template: %v", err)
		}
	}
}

func BenchmarkFilters(b *testing.B) {
	template := "{{ text|upper|trim|replace('WORLD', 'Go') }}"
	ctx := map[string]interface{}{"text": "  hello world  "}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ExecuteToString(template, ctx)
		if err != nil {
			b.Fatalf("Failed to execute template: %v", err)
		}
	}
}
