package runtime

import (
	"testing"

	"github.com/deicod/gojinja/lexer"
	"github.com/deicod/gojinja/nodes"
	"github.com/deicod/gojinja/parser"
)

type testSayExtension struct{}

func (e *testSayExtension) Tags() []string {
	return []string{"say"}
}

func (e *testSayExtension) Parse(p *parser.Parser) (nodes.Node, error) {
	token, err := p.Expect(lexer.TokenName)
	if err != nil {
		return nil, err
	}

	var expr nodes.Expr
	next := p.Current()
	if next.Type != lexer.TokenBlockEnd {
		parsed, err := p.ParseExpression()
		if err != nil {
			return nil, err
		}
		expr = parsed
	} else {
		data := &nodes.TemplateData{Data: "from extension"}
		data.SetPosition(nodes.NewPosition(token.Line, token.Column))
		expr = data
	}

	output := &nodes.Output{Nodes: []nodes.Expr{expr}}
	output.SetPosition(nodes.NewPosition(token.Line, token.Column))
	return output, nil
}

func TestEnvironmentCustomExtension(t *testing.T) {
	env := NewEnvironment()
	ext := &testSayExtension{}
	env.AddExtension(ext)

	tmpl, err := env.ParseString("{% say 'Go' %}", "ext")
	if err != nil {
		t.Fatalf("ParseString with extension failed: %v", err)
	}

	result, err := tmpl.ExecuteToString(nil)
	if err != nil {
		t.Fatalf("ExecuteToString failed: %v", err)
	}
	if result != "Go" {
		t.Fatalf("expected extension output 'Go', got %q", result)
	}

	defaultTmpl, err := env.ParseString("{% say %}", "ext-default")
	if err != nil {
		t.Fatalf("ParseString without args failed: %v", err)
	}
	defaultResult, err := defaultTmpl.ExecuteToString(nil)
	if err != nil {
		t.Fatalf("ExecuteToString default failed: %v", err)
	}
	if defaultResult != "from extension" {
		t.Fatalf("expected default extension output, got %q", defaultResult)
	}

	snapshot := env.Extensions()
	if len(snapshot) != 1 {
		t.Fatalf("expected 1 registered extension, got %d", len(snapshot))
	}
	snapshot = append(snapshot, &testSayExtension{})
	if len(env.Extensions()) != 1 {
		t.Fatalf("snapshot mutation should not affect environment extensions")
	}

	if !env.RemoveExtension(ext) {
		t.Fatalf("expected extension removal to succeed")
	}
	if env.RemoveExtension(ext) {
		t.Fatalf("expected removing the same extension twice to fail")
	}
	if _, err := env.ParseString("{% say %}", "missing-extension"); err == nil {
		t.Fatalf("expected parsing without registered extension to fail")
	}

	env.AddExtension(ext)
	env.ClearExtensions()
	if _, err := env.ParseString("{% say %}", "cleared-extension"); err == nil {
		t.Fatalf("expected parsing after clearing extensions to fail")
	}
}
