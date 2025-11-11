package gojinja2

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "greeting.html")
	if err := os.WriteFile(path, []byte("Hello {{ name }}!"), 0o644); err != nil {
		t.Fatalf("failed to write template: %v", err)
	}

	tmpl, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile error: %v", err)
	}

	output, err := tmpl.ExecuteToString(map[string]interface{}{"name": "Go"})
	if err != nil {
		t.Fatalf("ExecuteToString error: %v", err)
	}

	if output != "Hello Go!" {
		t.Fatalf("expected 'Hello Go!', got %q", output)
	}
}

func TestFloorDivisionOperator(t *testing.T) {
	tmpl, err := ParseString("{{ 7 // 2 }}")
	if err != nil {
		t.Fatalf("ParseString error: %v", err)
	}

	output, err := tmpl.ExecuteToString(nil)
	if err != nil {
		t.Fatalf("ExecuteToString error: %v", err)
	}

	if output != "3" {
		t.Fatalf("expected '3', got %q", output)
	}
}

func TestParseStringWithName(t *testing.T) {
	tmpl, err := ParseStringWithName("{{ greeting }}", "custom")
	if err != nil {
		t.Fatalf("ParseStringWithName error: %v", err)
	}

	if tmpl.Name() != "custom" {
		t.Fatalf("expected template name 'custom', got %q", tmpl.Name())
	}

	out, err := tmpl.ExecuteToString(map[string]interface{}{"greeting": "Hi"})
	if err != nil {
		t.Fatalf("ExecuteToString error: %v", err)
	}

	if out != "Hi" {
		t.Fatalf("expected 'Hi', got %q", out)
	}
}

func TestExecuteConvenienceFunctions(t *testing.T) {
	result, err := ExecuteToString("Hello {{ name }}", map[string]interface{}{"name": "Gopher"})
	if err != nil {
		t.Fatalf("ExecuteToString error: %v", err)
	}
	if result != "Hello Gopher" {
		t.Fatalf("expected 'Hello Gopher', got %q", result)
	}

	var buf bytes.Buffer
	if err := Execute("{{ value }}!", map[string]interface{}{"value": 42}, &buf); err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	if buf.String() != "42!" {
		t.Fatalf("expected '42!', got %q", buf.String())
	}
}

type testStreamAwaitable struct {
	value string
}

func (a testStreamAwaitable) Await(ctx *Context) (interface{}, error) {
	if ctx == nil {
		return nil, fmt.Errorf("missing context")
	}
	return a.value, nil
}

var _ Awaitable = (*testStreamAwaitable)(nil)

func TestStreamingConvenienceFunctions(t *testing.T) {
	stream, err := Generate("Hello {{ name }}", map[string]interface{}{"name": "Stream"})
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	var builder strings.Builder
	for {
		chunk, err := stream.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatalf("stream iteration error: %v", err)
		}
		builder.WriteString(chunk)
	}

	if result := builder.String(); result != "Hello Stream" {
		t.Fatalf("unexpected stream output: %q", result)
	}

	var directWriter bytes.Buffer
	if written, err := GenerateToWriter("Hello {{ who }}", map[string]interface{}{"who": "Writer"}, &directWriter); err != nil {
		t.Fatalf("GenerateToWriter error: %v", err)
	} else if written != int64(len("Hello Writer")) {
		t.Fatalf("unexpected number of bytes written: %d", written)
	}
	if directWriter.String() != "Hello Writer" {
		t.Fatalf("GenerateToWriter produced %q", directWriter.String())
	}

	env := NewEnvironment()
	env.SetKeepTrailingNewline(true)

	streamWithEnv, err := GenerateWithEnvironment(env, "value\n", nil)
	if err != nil {
		t.Fatalf("GenerateWithEnvironment error: %v", err)
	}

	var buf bytes.Buffer
	written, err := streamWithEnv.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo error: %v", err)
	}
	if written != int64(len("value\n")) {
		t.Fatalf("expected to write %d bytes, wrote %d", len("value\n"), written)
	}
	if buf.String() != "value\n" {
		t.Fatalf("expected trailing newline to be preserved, got %q", buf.String())
	}

	buf.Reset()
	if written, err := GenerateToWriterWithEnvironment(env, "value\n", nil, &buf); err != nil {
		t.Fatalf("GenerateToWriterWithEnvironment error: %v", err)
	} else if written != int64(len("value\n")) {
		t.Fatalf("expected to write %d bytes, wrote %d", len("value\n"), written)
	}
	if buf.String() != "value\n" {
		t.Fatalf("GenerateToWriterWithEnvironment preserved unexpected output: %q", buf.String())
	}

	asyncEnv := NewEnvironment()
	asyncEnv.SetEnableAsync(true)

	asyncStream, err := GenerateWithEnvironment(asyncEnv, "{{ await value }}", map[string]interface{}{
		"value": &testStreamAwaitable{value: "done"},
	})
	if err != nil {
		t.Fatalf("GenerateWithEnvironment async error: %v", err)
	}

	awaited, err := asyncStream.Collect()
	if err != nil {
		t.Fatalf("Collect error: %v", err)
	}
	if strings.TrimSpace(awaited) != "done" {
		t.Fatalf("expected awaited result 'done', got %q", strings.TrimSpace(awaited))
	}
}

func TestTemplateChainAndBatchRenderer(t *testing.T) {
	env := NewEnvironment()

	chain := NewTemplateChain(env)
	if err := chain.AddFromString("{{ greeting }}", "welcome"); err != nil {
		t.Fatalf("AddFromString error: %v", err)
	}

	tmpl, ok := chain.Get("welcome")
	if !ok {
		t.Fatalf("expected template 'welcome' to be present in chain")
	}

	out, err := tmpl.ExecuteToString(map[string]interface{}{"greeting": "Howdy"})
	if err != nil {
		t.Fatalf("ExecuteToString error: %v", err)
	}
	if out != "Howdy" {
		t.Fatalf("expected 'Howdy', got %q", out)
	}

	renderer := NewBatchRenderer(env)
	if err := renderer.AddTemplate("farewell", "Bye {{ name }}"); err != nil {
		t.Fatalf("AddTemplate error: %v", err)
	}

	rendered, err := renderer.Render("farewell", map[string]interface{}{"name": "Go"})
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if rendered != "Bye Go" {
		t.Fatalf("expected 'Bye Go', got %q", rendered)
	}

	buf := bytes.Buffer{}
	if err := renderer.RenderToWriter("farewell", map[string]interface{}{"name": "Go"}, &buf); err != nil {
		t.Fatalf("RenderToWriter error: %v", err)
	}
	if buf.String() != "Bye Go" {
		t.Fatalf("expected 'Bye Go', got %q", buf.String())
	}
}

func TestMakeModuleExports(t *testing.T) {
	tmpl, err := ParseString(`
{% macro greet(name) %}
Hello {{ name }}!
{% endmacro %}
{% set answer = 42 %}
{% export answer %}
`)
	if err != nil {
		t.Fatalf("ParseString error: %v", err)
	}

	module, err := tmpl.MakeModule(nil)
	if err != nil {
		t.Fatalf("MakeModule error: %v", err)
	}

	macro, err := module.GetMacro("greet")
	if err != nil {
		t.Fatalf("expected greet macro to be exported: %v", err)
	}

	ctx := NewContextWithEnvironment(tmpl.Environment(), nil)
	value, err := macro.Call(ctx, "Parity")
	if err != nil {
		t.Fatalf("macro call failed: %v", err)
	}

	if result := strings.TrimSpace(value.(string)); result != "Hello Parity!" {
		t.Fatalf("unexpected macro output: %q", result)
	}

	exported, ok := module.Resolve("answer")
	if !ok {
		t.Fatalf("expected exported value 'answer' to be resolvable")
	}
	switch v := exported.(type) {
	case int:
		if v != 42 {
			t.Fatalf("expected exported answer to be 42, got %d", v)
		}
	case int64:
		if v != 42 {
			t.Fatalf("expected exported answer to be 42, got %d", v)
		}
	default:
		t.Fatalf("expected exported answer to be numeric, got %T", exported)
	}

	names := module.GetExportNames()
	found := false
	for _, name := range names {
		if name == "answer" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected export names to include 'answer', got %v", names)
	}
}

func TestEnvironmentHelperFunctions(t *testing.T) {
	env := NewEnvironment()
	env.SetLoader(NewMapLoader(map[string]string{
		"one.html": "One",
		"two.html": "Two",
	}))

	tmpl, err := GetTemplate(env, "one.html")
	if err != nil {
		t.Fatalf("GetTemplate error: %v", err)
	}

	if out, err := tmpl.ExecuteToString(nil); err != nil || strings.TrimSpace(out) != "One" {
		if err != nil {
			t.Fatalf("template execution failed: %v", err)
		}
		t.Fatalf("unexpected GetTemplate render output: %q", out)
	}

	selected, err := SelectTemplate(env, []string{"missing.html", "two.html"})
	if err != nil {
		t.Fatalf("SelectTemplate error: %v", err)
	}

	if out, err := selected.ExecuteToString(nil); err != nil || strings.TrimSpace(out) != "Two" {
		if err != nil {
			t.Fatalf("selected template execution failed: %v", err)
		}
		t.Fatalf("unexpected SelectTemplate output: %q", out)
	}

	resolved, err := GetOrSelectTemplate(env, []interface{}{"missing.html", tmpl})
	if err != nil {
		t.Fatalf("GetOrSelectTemplate error: %v", err)
	}

	if resolved != tmpl {
		t.Fatalf("expected GetOrSelectTemplate to return provided template instance")
	}

	joined, err := JoinPath(env, "partials/header.html", "layouts/base.html")
	if err != nil {
		t.Fatalf("JoinPath error: %v", err)
	}

	if joined != "layouts/partials/header.html" {
		t.Fatalf("unexpected JoinPath result: %q", joined)
	}

	if _, err := GetTemplate(nil, "missing.html"); err == nil {
		t.Fatalf("expected GetTemplate to error with nil environment")
	}

	if _, err := JoinPath(nil, "child.html", "base.html"); err == nil {
		t.Fatalf("expected JoinPath to error with nil environment")
	}
}

func TestEnvironmentRenderingHelpers(t *testing.T) {
	env := NewEnvironment()
	env.SetKeepTrailingNewline(true)
	env.SetLoader(NewMapLoader(map[string]string{
		"hello.html": "Hello {{ name }}!\n",
	}))

	rendered, err := env.RenderTemplate("hello.html", map[string]interface{}{"name": "Parity"})
	if err != nil {
		t.Fatalf("RenderTemplate error: %v", err)
	}
	if strings.TrimSpace(rendered) != "Hello Parity!" {
		t.Fatalf("unexpected RenderTemplate output: %q", rendered)
	}

	var buf bytes.Buffer
	if err := env.RenderTemplateToWriter("hello.html", map[string]interface{}{"name": "Writer"}, &buf); err != nil {
		t.Fatalf("RenderTemplateToWriter error: %v", err)
	}
	if strings.TrimSpace(buf.String()) != "Hello Writer!" {
		t.Fatalf("unexpected RenderTemplateToWriter output: %q", buf.String())
	}

	stream, err := env.Generate("hello.html", map[string]interface{}{"name": "Stream"})
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	collected, err := stream.Collect()
	if err != nil {
		t.Fatalf("Collect error: %v", err)
	}
	if collected != "Hello Stream!\n" {
		t.Fatalf("unexpected collected output: %q", collected)
	}
}

func TestMakeModuleWithContext(t *testing.T) {
	env := NewEnvironment()
	tmpl, err := env.ParseString(`
{% set combined = prefix ~ ' ' ~ suffix %}
{% export combined %}
`, "module_shared")
	if err != nil {
		t.Fatalf("ParseString error: %v", err)
	}

	ctx := NewContextWithEnvironment(env, map[string]interface{}{"prefix": "Ms."})

	module, err := tmpl.MakeModuleWithContext(ctx, map[string]interface{}{
		"prefix": "Dr.",
		"suffix": "Parity",
	})
	if err != nil {
		t.Fatalf("MakeModuleWithContext error: %v", err)
	}

	exported, ok := module.Resolve("combined")
	if !ok {
		t.Fatalf("expected exported value 'combined' to be resolvable")
	}

	if result := strings.TrimSpace(exported.(string)); result != "Dr. Parity" {
		t.Fatalf("unexpected exported value: %q", result)
	}

	if value, ok := ctx.Get("prefix"); !ok || value.(string) != "Ms." {
		t.Fatalf("expected original context prefix to remain 'Ms.', got %v", value)
	}

	if _, ok := ctx.Get("suffix"); ok {
		t.Fatalf("expected temporary variables to be cleared from the shared context")
	}

	if exports := ctx.Exports(); len(exports) != 0 {
		t.Fatalf("expected no exports on original context, got %v", exports)
	}
}
