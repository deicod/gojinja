package runtime

import (
	"math"
	"strings"
	"testing"

	"github.com/deicod/gojinja/nodes"
)

func TestEqTest(t *testing.T) {
	result, err := ExecuteToString("{% if value is eq(42) %}yes{% else %}no{% endif %}", map[string]interface{}{"value": 42})
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if result != "yes" {
		t.Fatalf("expected 'yes', got %q", result)
	}
}

func TestNeTest(t *testing.T) {
	result, err := ExecuteToString("{% if value is ne(42) %}yes{% else %}no{% endif %}", map[string]interface{}{"value": 41})
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if result != "yes" {
		t.Fatalf("expected 'yes', got %q", result)
	}
}

func TestLtGtTests(t *testing.T) {
	tpl := "{% if value is lt(10) and value is gt(1) %}in range{% else %}out{% endif %}"
	result, err := ExecuteToString(tpl, map[string]interface{}{"value": 5})
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if result != "in range" {
		t.Fatalf("expected 'in range', got %q", result)
	}
}

func TestGeLeTests(t *testing.T) {
	tpl := "{% if value is ge(5) and value is le(5) %}equal{% else %}no{% endif %}"
	result, err := ExecuteToString(tpl, map[string]interface{}{"value": 5})
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if result != "equal" {
		t.Fatalf("expected 'equal', got %q", result)
	}
}

func TestEqualtoAlias(t *testing.T) {
	result, err := ExecuteToString("{% if value is equalto(42) %}yes{% else %}no{% endif %}", map[string]interface{}{"value": 42})
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if result != "yes" {
		t.Fatalf("expected 'yes', got %q", result)
	}
}

func TestIntegerFloatTests(t *testing.T) {
	intResult, err := ExecuteToString("{% if value is integer %}yes{% else %}no{% endif %}", map[string]interface{}{"value": 7})
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if intResult != "yes" {
		t.Fatalf("expected 'yes', got %q", intResult)
	}

	floatResult, err := ExecuteToString("{% if value is float %}yes{% else %}no{% endif %}", map[string]interface{}{"value": 7.5})
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if floatResult != "yes" {
		t.Fatalf("expected 'yes', got %q", floatResult)
	}

	boolResult, err := ExecuteToString("{% if value is integer %}yes{% else %}no{% endif %}", map[string]interface{}{"value": true})
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if boolResult != "no" {
		t.Fatalf("expected 'no' for boolean, got %q", boolResult)
	}
}

func TestTrueFalseTests(t *testing.T) {
	result, err := ExecuteToString("{% if value is true %}yes{% else %}no{% endif %}", map[string]interface{}{"value": true})
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if result != "yes" {
		t.Fatalf("expected 'yes', got %q", result)
	}

	result, err = ExecuteToString("{% if value is false %}yes{% else %}no{% endif %}", map[string]interface{}{"value": false})
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if result != "yes" {
		t.Fatalf("expected 'yes', got %q", result)
	}
}

func TestFilterAndTestPredicates(t *testing.T) {
	result, err := ExecuteToString("{% if 'upper' is filter and 'even' is test %}yes{% else %}no{% endif %}", nil)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if result != "yes" {
		t.Fatalf("expected 'yes', got %q", result)
	}

	result, err = ExecuteToString("{% if 'missing' is filter %}yes{% else %}no{% endif %}", nil)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if result != "no" {
		t.Fatalf("expected 'no', got %q", result)
	}
}

func TestCallableAttribute(t *testing.T) {
	ctx := map[string]interface{}{
		"fn":   func() {},
		"obj":  map[string]interface{}{"run": func() {}},
		"text": "hello",
	}

	result, err := ExecuteToString("{% if fn is callable %}yes{% else %}no{% endif %}", ctx)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if result != "yes" {
		t.Fatalf("expected 'yes', got %q", result)
	}

	result, err = ExecuteToString("{% if obj is callable(attribute='run') %}yes{% else %}no{% endif %}", ctx)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if result != "yes" {
		t.Fatalf("expected 'yes', got %q", result)
	}

	result, err = ExecuteToString("{% if obj is callable(attribute='missing') %}yes{% else %}no{% endif %}", ctx)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if result != "no" {
		t.Fatalf("expected 'no', got %q", result)
	}

	result, err = ExecuteToString("{% if text is callable(attribute='upper') %}yes{% else %}no{% endif %}", ctx)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if result != "yes" {
		t.Fatalf("expected 'yes', got %q", result)
	}
}

func TestEscapedTest(t *testing.T) {
	value := Markup("<strong>safe</strong>")
	result, err := ExecuteToString("{% if value is escaped %}safe{% else %}unsafe{% endif %}", map[string]interface{}{"value": value})
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if result != "safe" {
		t.Fatalf("expected 'safe', got %q", result)
	}
}

func TestModuleTest(t *testing.T) {
	namespace := NewMacroNamespace("helpers", nil)
	result, err := ExecuteToString("{% if ns is module %}yes{% else %}no{% endif %}", map[string]interface{}{"ns": namespace})
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if result != "yes" {
		t.Fatalf("expected 'yes', got %q", result)
	}

	result, err = ExecuteToString("{% if value is module %}yes{% else %}no{% endif %}", map[string]interface{}{"value": map[string]interface{}{"a": 1}})
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if result != "no" {
		t.Fatalf("expected 'no', got %q", result)
	}
}

func TestNamespaceGlobal(t *testing.T) {
	env := NewEnvironment()
	tpl, err := env.ParseString(`{% set ns = namespace(counter=1) %}{% do ns.set('counter', ns.counter + 4) %}{{ ns.counter }}`, "namespace")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	result, err := tpl.ExecuteToString(nil)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if strings.TrimSpace(result) != "5" {
		t.Fatalf("expected '5', got %q", result)
	}

	tpl, err = env.ParseString(`{{ namespace(foo='bar').foo }}`, "namespace_kw")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	result, err = tpl.ExecuteToString(nil)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if result != "bar" {
		t.Fatalf("expected 'bar', got %q", result)
	}
}

func TestNamespaceStatementCreatesNamespace(t *testing.T) {
	tpl := `{% namespace ns %}{% set ns.value = 42 %}{% endnamespace %}{{ ns.value }}`
	res, err := ExecuteToString(tpl, nil)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if strings.TrimSpace(res) != "42" {
		t.Fatalf("expected namespace to expose updated value, got %q", res)
	}
}

func TestNamespaceStatementInitializer(t *testing.T) {
	tpl := `{% set seed = namespace(counter=2) %}{% namespace ns = seed %}{% set ns.counter = ns.counter + 3 %}{% endnamespace %}{{ ns.counter }}`
	res, err := ExecuteToString(tpl, nil)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if strings.TrimSpace(res) != "5" {
		t.Fatalf("expected namespace initializer to be reused, got %q", res)
	}
}

func TestNamespaceStatementRejectsInvalidInitializer(t *testing.T) {
	_, err := ExecuteToString(`{% namespace ns = 42 %}{% endnamespace %}`, nil)
	if err == nil {
		t.Fatal("expected error for invalid namespace initializer")
	}
	if !strings.Contains(err.Error(), "expects a namespace or mapping") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNamespaceStatementIsolatesAssignments(t *testing.T) {
	tpl := `{% namespace ns %}{% set foo = 42 %}{% endnamespace %}{{ ns.foo }}`
	res, err := ExecuteToString(tpl, nil)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if strings.TrimSpace(res) != "42" {
		t.Fatalf("expected namespace attribute to be set, got %q", res)
	}

	res, err = ExecuteToString(`{% namespace ns %}{% set foo = 42 %}{% endnamespace %}{% if foo is defined %}leak{% else %}isolated{% endif %}`, nil)
	if err != nil {
		t.Fatalf("execution error when checking isolation: %v", err)
	}
	if strings.TrimSpace(res) != "isolated" {
		t.Fatalf("expected assignment to remain isolated, got %q", res)
	}
}

func TestNamespaceStatementIsolatesMacros(t *testing.T) {
	tpl := `{% namespace ns %}{% macro greet(name) %}hi {{ name }}{% endmacro %}{% endnamespace %}{{ ns.greet('World') }}`
	res, err := ExecuteToString(tpl, nil)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if strings.TrimSpace(res) != "hi World" {
		t.Fatalf("expected namespace macro to be callable via namespace, got %q", res)
	}

	res, err = ExecuteToString(`{% namespace ns %}{% macro greet(name) %}hi {{ name }}{% endmacro %}{% endnamespace %}{% if greet is defined %}leak{% else %}isolated{% endif %}`, nil)
	if err != nil {
		t.Fatalf("execution error when checking macro isolation: %v", err)
	}
	if strings.TrimSpace(res) != "isolated" {
		t.Fatalf("expected macro to remain isolated, got %q", res)
	}
}

func TestDictGlobalKeywords(t *testing.T) {
	res, err := ExecuteToString("{{ dict(foo='bar', baz=2).foo }}", nil)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if res != "bar" {
		t.Fatalf("expected 'bar', got %q", res)
	}
}

func TestTranslationGlobals(t *testing.T) {
	res, err := ExecuteToString("{{ _('Hello %(name)s', {'name': 'World'}) }}", nil)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if res != "Hello World" {
		t.Fatalf("expected 'Hello World', got %q", res)
	}

	res, err = ExecuteToString("{{ gettext('Plain text') }}", nil)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if res != "Plain text" {
		t.Fatalf("expected 'Plain text', got %q", res)
	}

	res, err = ExecuteToString("{{ ngettext('%(count)d apple', '%(count)d apples', 1, {'count': 1}) }}", nil)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if res != "1 apple" {
		t.Fatalf("expected '1 apple', got %q", res)
	}

	res, err = ExecuteToString("{{ ngettext('%(count)d apple', '%(count)d apples', 3, {'count': 3}) }}", nil)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if res != "3 apples" {
		t.Fatalf("expected '3 apples', got %q", res)
	}
}

func TestDebugGlobal(t *testing.T) {
	res, err := ExecuteToString("{{ debug() }}", map[string]interface{}{"value": 42})
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if !strings.Contains(res, "value") || !strings.Contains(res, "42") {
		t.Fatalf("expected debug output to include value, got %q", res)
	}
}

func TestClassGlobal(t *testing.T) {
	res, err := ExecuteToString(`{% set User = class('User', {'role': 'guest'}) %}{{ User.role }}`, nil)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if res != "guest" {
		t.Fatalf("expected 'guest', got %q", res)
	}

	res, err = ExecuteToString(`{% set Base = namespace(role='guest') %}{% set Admin = class('Admin', Base, {'role': 'admin'}) %}{{ Admin.role }}`, nil)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if res != "admin" {
		t.Fatalf("expected 'admin', got %q", res)
	}

	res, err = ExecuteToString(`{{ class('Thing', foo='bar').foo }}`, nil)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if res != "bar" {
		t.Fatalf("expected 'bar', got %q", res)
	}
}

func TestSelfAndEnvironmentGlobals(t *testing.T) {
	env := NewEnvironment()
	tpl, err := env.ParseString("{{ self }}", "self_test")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	res, err := tpl.ExecuteToString(nil)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if strings.TrimSpace(res) == "" {
		t.Fatalf("expected non-empty self output, got %q", res)
	}

	tpl, err = env.ParseString("{{ environment }}", "env_test")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	res, err = tpl.ExecuteToString(nil)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if strings.TrimSpace(res) == "" {
		t.Fatalf("expected non-empty environment output, got %q", res)
	}
}

func TestKeepTrailingNewlineOption(t *testing.T) {
	env := NewEnvironment()
	tpl, err := env.ParseString("Hello\n", "trim_test")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	res, err := tpl.ExecuteToString(nil)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if strings.HasSuffix(res, "\n") {
		t.Fatalf("expected trailing newline to be trimmed, got %q", res)
	}

	env2 := NewEnvironment()
	env2.SetKeepTrailingNewline(true)
	tpl2, err := env2.ParseString("Hello\n", "keep_test")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if !tpl2.Environment().ShouldKeepTrailingNewline() {
		t.Fatalf("expected environment to keep trailing newline")
	}
	if len(tpl2.AST().Body) > 0 {
		if out, ok := tpl2.AST().Body[0].(*nodes.Output); ok {
			if len(out.Nodes) > 0 {
				if data, ok := out.Nodes[0].(*nodes.TemplateData); ok {
					if !strings.HasSuffix(data.Data, "\n") {
						t.Fatalf("expected AST data to keep trailing newline")
					}
				}
			}
		}
	}
	res, err = tpl2.ExecuteToString(nil)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if !strings.HasSuffix(res, "\n") {
		t.Fatalf("expected trailing newline to be preserved, got %q", res)
	}
}

func TestLineStatementRendering(t *testing.T) {
	env := NewEnvironment()
	env.SetLineStatementPrefix("#")

	tpl, err := env.ParseString(`# for v in values:
- {{ v }}
# endfor
`, "line_stmt")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	result, err := tpl.ExecuteToString(map[string]interface{}{"values": []int{0, 1, 2}})
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}

	expected := "- 0\n- 1\n- 2"
	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestLineStatementRenderingAfterText(t *testing.T) {
	env := NewEnvironment()
	env.SetLineStatementPrefix("#")

	tpl, err := env.ParseString(`Title
# for v in values:
- {{ v }}
# endfor
`, "line_stmt_after_text")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	result, err := tpl.ExecuteToString(map[string]interface{}{"values": []int{0, 1, 2}})
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}

	expected := "Title\n- 0\n- 1\n- 2"
	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestLineCommentRendering(t *testing.T) {
	env := NewEnvironment()
	env.SetLineStatementPrefix("#")
	env.SetLineCommentPrefix("//")

	tpl, err := env.ParseString(`// ignore me
# set value = 5
value: {{ value }}
`, "line_comment")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	result, err := tpl.ExecuteToString(nil)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}

	if result != "value: 5" {
		t.Fatalf("expected 'value: 5', got %q", result)
	}
}

func TestListTupleDictTests(t *testing.T) {
	listResult, err := ExecuteToString("{% if value is list %}yes{% else %}no{% endif %}", map[string]interface{}{"value": []int{1, 2, 3}})
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if listResult != "yes" {
		t.Fatalf("expected 'yes', got %q", listResult)
	}

	tupleResult, err := ExecuteToString("{% if value is tuple %}yes{% else %}no{% endif %}", map[string]interface{}{"value": [2]int{1, 2}})
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if tupleResult != "yes" {
		t.Fatalf("expected 'yes', got %q", tupleResult)
	}

	dictResult, err := ExecuteToString("{% if value is dict %}yes{% else %}no{% endif %}", map[string]interface{}{"value": map[string]int{"a": 1}})
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if dictResult != "yes" {
		t.Fatalf("expected 'yes', got %q", dictResult)
	}
}

func TestPatternTests(t *testing.T) {
	value := map[string]interface{}{"value": "foobarbaz"}

	matchResult, err := ExecuteToString("{% if value is matching('^foo') %}yes{% else %}no{% endif %}", value)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if matchResult != "yes" {
		t.Fatalf("expected 'yes', got %q", matchResult)
	}

	searchResult, err := ExecuteToString("{% if value is search('bar') %}yes{% else %}no{% endif %}", value)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if searchResult != "yes" {
		t.Fatalf("expected 'yes', got %q", searchResult)
	}

	startResult, err := ExecuteToString("{% if value is startingwith('foo') %}yes{% else %}no{% endif %}", value)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if startResult != "yes" {
		t.Fatalf("expected 'yes', got %q", startResult)
	}

	endResult, err := ExecuteToString("{% if value is endingwith('baz') %}yes{% else %}no{% endif %}", value)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if endResult != "yes" {
		t.Fatalf("expected 'yes', got %q", endResult)
	}

	containsResult, err := ExecuteToString("{% if value is containing('oba') %}yes{% else %}no{% endif %}", value)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if containsResult != "yes" {
		t.Fatalf("expected 'yes', got %q", containsResult)
	}
}

func TestNumericStateTests(t *testing.T) {
	result, err := ExecuteToString("{% if value is infinite %}yes{% else %}no{% endif %}", map[string]interface{}{"value": math.Inf(1)})
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if result != "yes" {
		t.Fatalf("expected 'yes', got %q", result)
	}

	result, err = ExecuteToString("{% if value is nan %}yes{% else %}no{% endif %}", map[string]interface{}{"value": math.NaN()})
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if result != "yes" {
		t.Fatalf("expected 'yes', got %q", result)
	}

	result, err = ExecuteToString("{% if value is finite %}yes{% else %}no{% endif %}", map[string]interface{}{"value": 3.14})
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if result != "yes" {
		t.Fatalf("expected 'yes', got %q", result)
	}

	result, err = ExecuteToString("{% if value is finite %}yes{% else %}no{% endif %}", map[string]interface{}{"value": math.Inf(-1)})
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if result != "no" {
		t.Fatalf("expected 'no', got %q", result)
	}
}

func TestComparisonAliases(t *testing.T) {
	ctx := map[string]interface{}{"value": 7}
	result, err := ExecuteToString("{% if value is greaterthan(10) %}yes{% else %}no{% endif %}", ctx)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if result != "no" {
		t.Fatalf("expected 'no', got %q", result)
	}

	result, err = ExecuteToString("{% if value is ge(5) %}yes{% else %}no{% endif %}", ctx)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if result != "yes" {
		t.Fatalf("expected 'yes', got %q", result)
	}

	result, err = ExecuteToString("{% if value is lessthan(5) %}yes{% else %}no{% endif %}", ctx)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if result != "no" {
		t.Fatalf("expected 'no', got %q", result)
	}
}
