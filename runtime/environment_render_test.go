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

func TestEnvironmentGenerateAPIHelpers(t *testing.T) {
	env := NewEnvironment()
	env.SetKeepTrailingNewline(true)
	env.SetLoader(NewMapLoader(map[string]string{
		"api.txt": "API {{ value }}\n",
	}))

	stream, err := GenerateTemplateWithEnvironment(env, "api.txt", map[string]interface{}{"value": "stream"})
	if err != nil {
		t.Fatalf("GenerateTemplateWithEnvironment error: %v", err)
	}

	collected, err := stream.Collect()
	if err != nil {
		t.Fatalf("Collect error: %v", err)
	}
	if collected != "API stream\n" {
		t.Fatalf("unexpected collected output: %q", collected)
	}

	var buf bytes.Buffer
	written, err := GenerateTemplateToWriterWithEnvironment(env, "api.txt", map[string]interface{}{"value": "writer"}, &buf)
	if err != nil {
		t.Fatalf("GenerateTemplateToWriterWithEnvironment error: %v", err)
	}
	if written != int64(len("API writer\n")) {
		t.Fatalf("expected to write %d bytes, wrote %d", len("API writer\n"), written)
	}
	if buf.String() != "API writer\n" {
		t.Fatalf("unexpected writer output: %q", buf.String())
	}
}

func TestEnvironmentGenerateToWriter(t *testing.T) {
	env := NewEnvironment()
	env.SetKeepTrailingNewline(true)
	env.SetLoader(NewMapLoader(map[string]string{
		"stream.txt": "Chunk {{ value }}\n",
	}))

	var buf bytes.Buffer
	written, err := env.GenerateToWriter("stream.txt", map[string]interface{}{"value": "42"}, &buf)
	if err != nil {
		t.Fatalf("GenerateToWriter error: %v", err)
	}

	if written != int64(len("Chunk 42\n")) {
		t.Fatalf("expected to write %d bytes, wrote %d", len("Chunk 42\n"), written)
	}
	if buf.String() != "Chunk 42\n" {
		t.Fatalf("unexpected GenerateToWriter output: %q", buf.String())
	}

	if _, err := env.GenerateToWriter("stream.txt", nil, nil); err == nil {
		t.Fatalf("expected GenerateToWriter to error with nil writer")
	}
}
