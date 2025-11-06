package runtime

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestTemplateGenerateStream(t *testing.T) {
	env := NewEnvironment()
	tmpl, err := env.ParseString("Hello {{ name }}!", "stream_basic")
	if err != nil {
		t.Fatalf("ParseString error: %v", err)
	}

	stream, err := tmpl.Generate(map[string]interface{}{"name": "Go"})
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	var parts []string
	for {
		chunk, err := stream.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatalf("stream next error: %v", err)
		}
		parts = append(parts, chunk)
	}

	if joined := strings.Join(parts, ""); joined != "Hello Go!" {
		t.Fatalf("unexpected stream output: %q", joined)
	}
}

func TestTemplateStreamWriteToTrimsTrailingNewline(t *testing.T) {
	env := NewEnvironment()
	tmpl, err := env.ParseString("value\n", "stream_trim")
	if err != nil {
		t.Fatalf("ParseString error: %v", err)
	}

	stream, err := tmpl.Generate(nil)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	var buf bytes.Buffer
	if written, err := stream.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo error: %v", err)
	} else if written != int64(len("value")) {
		t.Fatalf("expected to write %d bytes, wrote %d", len("value"), written)
	}

	if buf.String() != "value" {
		t.Fatalf("expected trailing newline to be trimmed, got %q", buf.String())
	}
}

func TestTemplateStreamKeepTrailingNewline(t *testing.T) {
	env := NewEnvironment()
	env.SetKeepTrailingNewline(true)

	tmpl, err := env.ParseString("value\n", "stream_keep")
	if err != nil {
		t.Fatalf("ParseString error: %v", err)
	}

	stream, err := tmpl.Generate(nil)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	var buf bytes.Buffer
	if written, err := stream.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo error: %v", err)
	} else if written != int64(len("value\n")) {
		t.Fatalf("expected to write %d bytes, wrote %d", len("value\n"), written)
	}

	if buf.String() != "value\n" {
		t.Fatalf("expected trailing newline to be preserved, got %q", buf.String())
	}
}

func TestTemplateStreamCollect(t *testing.T) {
	env := NewEnvironment()
	tmpl, err := env.ParseString("{{ 'a' }}{{ 'b' }}{{ 'c' }}\n", "stream_collect")
	if err != nil {
		t.Fatalf("ParseString error: %v", err)
	}

	stream, err := tmpl.Generate(nil)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	result, err := stream.Collect()
	if err != nil {
		t.Fatalf("Collect error: %v", err)
	}

	if result != "abc" {
		t.Fatalf("unexpected collected output: %q", result)
	}
}

func TestTemplateStreamPropagatesErrors(t *testing.T) {
	env := NewEnvironment()
	tmpl, err := env.ParseString("{{ 1 // 0 }}", "stream_error")
	if err != nil {
		t.Fatalf("ParseString error: %v", err)
	}

	stream, err := tmpl.Generate(nil)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	if _, err := stream.Next(); err == nil {
		t.Fatalf("expected stream to report rendering error")
	} else {
		var tplErr *Error
		if !errors.As(err, &tplErr) {
			t.Fatalf("expected template error, got %T", err)
		}
	}
}
