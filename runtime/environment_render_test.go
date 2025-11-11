package runtime

import (
	"bytes"
	"strings"
	"testing"
)

func TestEnvironmentRenderHelpers(t *testing.T) {
	env := NewEnvironment()
	env.SetLoader(NewMapLoader(map[string]string{
		"greet.html": "Hello {{ name }}!",
	}))

	rendered, err := env.RenderTemplate("greet.html", map[string]interface{}{"name": "Parity"})
	if err != nil {
		t.Fatalf("RenderTemplate error: %v", err)
	}
	if strings.TrimSpace(rendered) != "Hello Parity!" {
		t.Fatalf("unexpected RenderTemplate output: %q", rendered)
	}

	var buf bytes.Buffer
	if err := env.RenderTemplateToWriter("greet.html", map[string]interface{}{"name": "Writer"}, &buf); err != nil {
		t.Fatalf("RenderTemplateToWriter error: %v", err)
	}
	if strings.TrimSpace(buf.String()) != "Hello Writer!" {
		t.Fatalf("unexpected RenderTemplateToWriter output: %q", buf.String())
	}
}

func TestEnvironmentGenerateHelper(t *testing.T) {
	env := NewEnvironment()
	env.SetKeepTrailingNewline(true)
	env.SetLoader(NewMapLoader(map[string]string{
		"stream.txt": "Value: {{ value }}\n",
	}))

	stream, err := env.Generate("stream.txt", map[string]interface{}{"value": "42"})
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	collected, err := stream.Collect()
	if err != nil {
		t.Fatalf("Collect error: %v", err)
	}
	if collected != "Value: 42\n" {
		t.Fatalf("unexpected stream output: %q", collected)
	}
}
