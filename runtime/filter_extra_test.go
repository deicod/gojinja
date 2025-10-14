package runtime

import (
	"strings"
	"testing"
)

func TestFilesizeformatDecimal(t *testing.T) {
	output, err := ExecuteToString("{{ 3000|filesizeformat }}", nil)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if output != "2.9 KB" {
		t.Fatalf("expected '2.9 KB', got %q", output)
	}
}

func TestFilesizeformatBinary(t *testing.T) {
	env := NewEnvironment()
	template, err := env.ParseString("{{ size|filesizeformat(true) }}", "test")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	result, err := template.ExecuteToString(map[string]interface{}{"size": 2048})
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if result != "2.0 KiB" {
		t.Fatalf("expected '2.0 KiB', got %q", result)
	}
}

func TestEscapeJS(t *testing.T) {
	out, err := ExecuteToString("{{ value|escapejs }}", map[string]interface{}{"value": "<script>\n"})
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	expected := "\\u003cscript\\u003e\\n"
	if out != expected {
		t.Fatalf("expected %q, got %q", expected, out)
	}
}

func TestForceEscape(t *testing.T) {
	res, err := ExecuteToString("{{ '<b>'|forceescape }}", nil)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if res != "&lt;b&gt;" {
		t.Fatalf("expected '&lt;b&gt;', got %q", res)
	}
}

func TestToJSONFilter(t *testing.T) {
	result, err := ExecuteToString("{{ data|tojson }}", map[string]interface{}{"data": []string{"go", "jinja"}})
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if result != "[\"go\",\"jinja\"]" {
		t.Fatalf("unexpected json output: %q", result)
	}
}

func TestFromJSONFilter(t *testing.T) {
	env := NewEnvironment()
	tmpl, err := env.ParseString("{{ (data|fromjson).name }}", "test")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	res, err := tmpl.ExecuteToString(map[string]interface{}{"data": `{"name":"world"}`})
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if res != "world" {
		t.Fatalf("expected 'world', got %q", res)
	}
}

func TestRandomFilterWithSeed(t *testing.T) {
	env := NewEnvironment()
	tmpl, err := env.ParseString("{{ items|random(1) }}", "test")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	res, err := tmpl.ExecuteToString(map[string]interface{}{"items": []interface{}{"a", "b", "c"}})
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if res != "c" {
		t.Fatalf("expected 'c', got %q", res)
	}
}

func TestFloatformat(t *testing.T) {
	tpl := "{{ value|floatformat(2) }}"
	res, err := ExecuteToString(tpl, map[string]interface{}{"value": 3.14159})
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if res != "3.14" {
		t.Fatalf("expected '3.14', got %q", res)
	}

	res, err = ExecuteToString("{{ value|floatformat('-2') }}", map[string]interface{}{"value": 3.100})
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if res != "3.10" {
		t.Fatalf("expected '3.10', got %q", res)
	}
}

func TestPprint(t *testing.T) {
	res, err := ExecuteToString("{{ data|pprint }}", map[string]interface{}{"data": map[string]interface{}{"a": 1}})
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if !strings.Contains(res, "\"a\": 1") {
		t.Fatalf("expected pretty json output, got %q", res)
	}
}

func TestFormatFilter(t *testing.T) {
	res, err := ExecuteToString("{{ \"%s - %d\"|format('item', 3) }}", nil)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if res != "item - 3" {
		t.Fatalf("expected 'item - 3', got %q", res)
	}

	res, err = ExecuteToString("{{ '%(name)s scored %(score)d'|format({'name':'Go','score':42}) }}", nil)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if res != "Go scored 42" {
		t.Fatalf("expected 'Go scored 42', got %q", res)
	}
}

func TestUrlize(t *testing.T) {
	res, err := ExecuteToString("{{ 'Visit http://example.com'|urlize(0, true, '_blank', 'noopener') }}", nil)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if !strings.Contains(res, "href=\"http://example.com\"") || !strings.Contains(res, "rel=\"nofollow noopener\"") || !strings.Contains(res, "target=\"_blank\"") {
		t.Fatalf("unexpected urlize output: %q", res)
	}
}

func TestUrlizeBareDomain(t *testing.T) {
	res, err := ExecuteToString("{{ 'See example.org for docs'|urlize }}", nil)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if !strings.Contains(res, "href=\"https://example.org\"") {
		t.Fatalf("expected bare domain to be linked, got %q", res)
	}
}

func TestUrlizeExtraSchemes(t *testing.T) {
	res, err := ExecuteToString("{{ 'Call tel:123-456'|urlize(0, false, '', '', ['tel:']) }}", nil)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if !strings.Contains(res, "href=\"tel:123-456\"") {
		t.Fatalf("expected tel scheme to be linked, got %q", res)
	}
}

func TestUrlizeEmail(t *testing.T) {
	res, err := ExecuteToString("{{ 'Contact admin@example.com'|urlize }}", nil)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if !strings.Contains(res, "href=\"mailto:admin@example.com\"") {
		t.Fatalf("expected email address to be linked, got %q", res)
	}
}

func TestXMLAttr(t *testing.T) {
	res, err := ExecuteToString("<tag{{ attrs|xmlattr }} />", map[string]interface{}{"attrs": map[string]interface{}{"id": "main", "class": []string{"btn", "primary"}}})
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if res != "<tag class=\"btn primary\" id=\"main\" />" && res != "<tag id=\"main\" class=\"btn primary\" />" {
		t.Fatalf("unexpected xmlattr output: %q", res)
	}
}

func TestXMLAttrNoAutospace(t *testing.T) {
	res, err := ExecuteToString("<tag{{ attrs|xmlattr(false) }} />", map[string]interface{}{"attrs": map[string]interface{}{"class": "btn", "id": "main"}})
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if res != "<tagclass=\"btn\" id=\"main\" />" {
		t.Fatalf("unexpected xmlattr output without autospace: %q", res)
	}
}

func TestXMLAttrInvalidKey(t *testing.T) {
	_, err := ExecuteToString("<tag{{ attrs|xmlattr }} />", map[string]interface{}{"attrs": map[string]interface{}{"invalid key": "value"}})
	if err == nil {
		t.Fatal("expected error for invalid attribute key")
	}
}

func TestXMLAttrInvalidInput(t *testing.T) {
	_, err := ExecuteToString("<tag{{ value|xmlattr }} />", map[string]interface{}{"value": "not-a-map"})
	if err == nil {
		t.Fatal("expected error for non-mapping input")
	}
}

func TestShuffleFilter(t *testing.T) {
	values, err := filterShuffle(nil, []interface{}{1, 2, 3, 4}, 42)
	if err != nil {
		t.Fatalf("shuffle error: %v", err)
	}
	res, ok := values.([]interface{})
	if !ok {
		t.Fatalf("expected []interface{}, got %T", values)
	}
	if len(res) != 4 {
		t.Fatalf("expected length 4, got %d", len(res))
	}
	seen := map[interface{}]bool{}
	for _, v := range res {
		seen[v] = true
	}
	for i := 1; i <= 4; i++ {
		if !seen[i] {
			t.Fatalf("missing element %d in shuffled result", i)
		}
	}
}

func TestBatchFilter(t *testing.T) {
	res, err := ExecuteToString("{{ items|batch(2, 'X')|tojson }}", map[string]interface{}{
		"items": []interface{}{1, 2, 3},
	})
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if res != "[[1,2],[3,\"X\"]]" {
		t.Fatalf("unexpected batch output: %q", res)
	}
	if res != "[[1,2],[3,\"X\"]]" {
		t.Fatalf("unexpected batch output: %q", res)
	}
}

func TestSliceFilterColumns(t *testing.T) {
	items := []interface{}{"a", "b", "c", "d", "e", "f"}
	res, err := filterSlice(nil, items, 3)
	if err != nil {
		t.Fatalf("filter error: %v", err)
	}
	columns, ok := res.([][]interface{})
	if !ok {
		t.Fatalf("expected [][]interface{}, got %T", res)
	}
	if len(columns) != 3 {
		t.Fatalf("expected 3 columns, got %d", len(columns))
	}
	if len(columns[0]) == 0 || columns[0][0] != "a" {
		t.Fatalf("unexpected first column: %#v", columns[0])
	}
}

func TestMapFilterAttributeKeyword(t *testing.T) {
	res, err := ExecuteToString("{{ users|map(attribute='name')|tojson }}", map[string]interface{}{
		"users": []map[string]interface{}{{"name": "Alice"}, {"name": "Bob"}},
	})
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if res != "[\"Alice\",\"Bob\"]" {
		t.Fatalf("unexpected map output: %q", res)
	}
}

func TestSelectAttrKeyword(t *testing.T) {
	res, err := ExecuteToString("{{ users|selectattr(attribute='active')|tojson }}", map[string]interface{}{
		"users": []map[string]interface{}{{"name": "Alice", "active": true}, {"name": "Bob", "active": false}},
	})
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if res != "[{\"active\":true,\"name\":\"Alice\"}]" {
		t.Fatalf("unexpected selectattr output: %q", res)
	}
}
