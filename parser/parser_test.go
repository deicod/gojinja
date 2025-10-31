package parser

import (
	"strings"
	"testing"

	"github.com/deicod/gojinja/nodes"
)

func TestParser_BasicExpressions(t *testing.T) {
	env := &Environment{}

	tests := []struct {
		name     string
		template string
		validate func(*testing.T, *nodes.Template)
	}{
		{
			name:     "SimpleVariable",
			template: "{{ name }}",
			validate: func(t *testing.T, tmpl *nodes.Template) {
				if len(tmpl.Body) != 1 {
					t.Fatalf("expected 1 node in body, got %d", len(tmpl.Body))
				}
				output, ok := tmpl.Body[0].(*nodes.Output)
				if !ok {
					t.Fatalf("expected Output node, got %T", tmpl.Body[0])
				}
				if len(output.Nodes) != 1 {
					t.Fatalf("expected 1 expression in output, got %d", len(output.Nodes))
				}
				name, ok := output.Nodes[0].(*nodes.Name)
				if !ok {
					t.Fatalf("expected Name node, got %T", output.Nodes[0])
				}
				if name.Name != "name" {
					t.Errorf("expected name 'name', got '%s'", name.Name)
				}
			},
		},
		{
			name:     "StringLiteral",
			template: `{{ "hello" }}`,
			validate: func(t *testing.T, tmpl *nodes.Template) {
				if len(tmpl.Body) != 1 {
					t.Fatalf("expected 1 node in body, got %d", len(tmpl.Body))
				}
				output, ok := tmpl.Body[0].(*nodes.Output)
				if !ok {
					t.Fatalf("expected Output node, got %T", tmpl.Body[0])
				}
				if len(output.Nodes) != 1 {
					t.Fatalf("expected 1 expression in output, got %d", len(output.Nodes))
				}
				constNode, ok := output.Nodes[0].(*nodes.Const)
				if !ok {
					t.Fatalf("expected Const node, got %T", output.Nodes[0])
				}
				if constNode.Value != "hello" {
					t.Errorf("expected value 'hello', got '%v'", constNode.Value)
				}
			},
		},
		{
			name:     "IntegerLiteral",
			template: "{{ 42 }}",
			validate: func(t *testing.T, tmpl *nodes.Template) {
				if len(tmpl.Body) != 1 {
					t.Fatalf("expected 1 node in body, got %d", len(tmpl.Body))
				}
				output, ok := tmpl.Body[0].(*nodes.Output)
				if !ok {
					t.Fatalf("expected Output node, got %T", tmpl.Body[0])
				}
				if len(output.Nodes) != 1 {
					t.Fatalf("expected 1 expression in output, got %d", len(output.Nodes))
				}
				constNode, ok := output.Nodes[0].(*nodes.Const)
				if !ok {
					t.Fatalf("expected Const node, got %T", output.Nodes[0])
				}
				if constNode.Value != int64(42) {
					t.Errorf("expected value 42, got '%v'", constNode.Value)
				}
			},
		},
		{
			name:     "BinaryAdd",
			template: "{{ a + b }}",
			validate: func(t *testing.T, tmpl *nodes.Template) {
				if len(tmpl.Body) != 1 {
					t.Fatalf("expected 1 node in body, got %d", len(tmpl.Body))
				}
				output, ok := tmpl.Body[0].(*nodes.Output)
				if !ok {
					t.Fatalf("expected Output node, got %T", tmpl.Body[0])
				}
				if len(output.Nodes) != 1 {
					t.Fatalf("expected 1 expression in output, got %d", len(output.Nodes))
				}
				add, ok := output.Nodes[0].(*nodes.Add)
				if !ok {
					t.Fatalf("expected Add node, got %T", output.Nodes[0])
				}
				if add.Left == nil || add.Right == nil {
					t.Fatal("expected left and right operands")
				}
			},
		},
		{
			name:     "FilterExpression",
			template: "{{ name | upper }}",
			validate: func(t *testing.T, tmpl *nodes.Template) {
				if len(tmpl.Body) != 1 {
					t.Fatalf("expected 1 node in body, got %d", len(tmpl.Body))
				}
				output, ok := tmpl.Body[0].(*nodes.Output)
				if !ok {
					t.Fatalf("expected Output node, got %T", tmpl.Body[0])
				}
				if len(output.Nodes) != 1 {
					t.Fatalf("expected 1 expression in output, got %d", len(output.Nodes))
				}
				filter, ok := output.Nodes[0].(*nodes.Filter)
				if !ok {
					t.Fatalf("expected Filter node, got %T", output.Nodes[0])
				}
				if filter.Name != "upper" {
					t.Errorf("expected filter name 'upper', got '%s'", filter.Name)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := NewParser(env, tt.template, "test", "test.html", "")
			if err != nil {
				t.Fatalf("failed to create parser: %v", err)
			}

			tmpl, err := parser.Parse()
			if err != nil {
				t.Fatalf("failed to parse template: %v", err)
			}

			tt.validate(t, tmpl)
		})
	}
}

func TestParser_BasicStatements(t *testing.T) {
	env := &Environment{}

	tests := []struct {
		name     string
		template string
		validate func(*testing.T, *nodes.Template)
	}{
		{
			name:     "IfStatement",
			template: `{% if condition %}hello{% endif %}`,
			validate: func(t *testing.T, tmpl *nodes.Template) {
				if len(tmpl.Body) != 1 {
					t.Fatalf("expected 1 node in body, got %d", len(tmpl.Body))
				}
				ifStmt, ok := tmpl.Body[0].(*nodes.If)
				if !ok {
					t.Fatalf("expected If node, got %T", tmpl.Body[0])
				}
				if ifStmt.Test == nil {
					t.Fatal("expected test expression")
				}
				if len(ifStmt.Body) != 1 {
					t.Fatalf("expected 1 node in if body, got %d", len(ifStmt.Body))
				}
				if len(ifStmt.Elif) != 0 {
					t.Fatalf("expected 0 elif nodes, got %d", len(ifStmt.Elif))
				}
				if len(ifStmt.Else) != 0 {
					t.Fatalf("expected 0 else nodes, got %d", len(ifStmt.Else))
				}
			},
		},
		{
			name:     "ForStatement",
			template: `{% for item in items %}{{ item }}{% endfor %}`,
			validate: func(t *testing.T, tmpl *nodes.Template) {
				if len(tmpl.Body) != 1 {
					t.Fatalf("expected 1 node in body, got %d", len(tmpl.Body))
				}
				forStmt, ok := tmpl.Body[0].(*nodes.For)
				if !ok {
					t.Fatalf("expected For node, got %T", tmpl.Body[0])
				}
				if forStmt.Target == nil {
					t.Fatal("expected target")
				}
				if forStmt.Iter == nil {
					t.Fatal("expected iterator")
				}
				if len(forStmt.Body) != 1 {
					t.Fatalf("expected 1 node in for body, got %d", len(forStmt.Body))
				}
				if len(forStmt.Else) != 0 {
					t.Fatalf("expected 0 else nodes, got %d", len(forStmt.Else))
				}
			},
		},
		{
			name:     "SetStatement",
			template: `{% set variable = value %}`,
			validate: func(t *testing.T, tmpl *nodes.Template) {
				if len(tmpl.Body) != 1 {
					t.Fatalf("expected 1 node in body, got %d", len(tmpl.Body))
				}
				assign, ok := tmpl.Body[0].(*nodes.Assign)
				if !ok {
					t.Fatalf("expected Assign node, got %T", tmpl.Body[0])
				}
				if assign.Target == nil {
					t.Fatal("expected target")
				}
				if assign.Node == nil {
					t.Fatal("expected value")
				}
			},
		},
		{
			name:     "BlockStatement",
			template: `{% block content %}hello{% endblock %}`,
			validate: func(t *testing.T, tmpl *nodes.Template) {
				if len(tmpl.Body) != 1 {
					t.Fatalf("expected 1 node in body, got %d", len(tmpl.Body))
				}
				block, ok := tmpl.Body[0].(*nodes.Block)
				if !ok {
					t.Fatalf("expected Block node, got %T", tmpl.Body[0])
				}
				if block.Name != "content" {
					t.Errorf("expected block name 'content', got '%s'", block.Name)
				}
				if len(block.Body) != 1 {
					t.Fatalf("expected 1 node in block body, got %d", len(block.Body))
				}
			},
		},
		{
			name:     "DoStatement",
			template: `{% do log('message') %}`,
			validate: func(t *testing.T, tmpl *nodes.Template) {
				if len(tmpl.Body) != 1 {
					t.Fatalf("expected 1 node in body, got %d", len(tmpl.Body))
				}
				doNode, ok := tmpl.Body[0].(*nodes.Do)
				if !ok {
					t.Fatalf("expected Do node, got %T", tmpl.Body[0])
				}
				call, ok := doNode.Expr.(*nodes.Call)
				if !ok {
					t.Fatalf("expected Call expression, got %T", doNode.Expr)
				}
				name, ok := call.Node.(*nodes.Name)
				if !ok {
					t.Fatalf("expected Name node, got %T", call.Node)
				}
				if name.Name != "log" {
					t.Errorf("expected function name 'log', got '%s'", name.Name)
				}
				if len(call.Args) != 1 {
					t.Fatalf("expected one argument, got %d", len(call.Args))
				}
				arg, ok := call.Args[0].(*nodes.Const)
				if !ok {
					t.Fatalf("expected Const argument, got %T", call.Args[0])
				}
				if arg.Value != "message" {
					t.Errorf("expected argument value 'message', got %v", arg.Value)
				}
			},
		},
		{
			name:     "SpacelessBlock",
			template: `{% spaceless %}<div>  <span>{{ value }}</span>  </div>{% endspaceless %}`,
			validate: func(t *testing.T, tmpl *nodes.Template) {
				if len(tmpl.Body) != 1 {
					t.Fatalf("expected 1 node in body, got %d", len(tmpl.Body))
				}
				block, ok := tmpl.Body[0].(*nodes.Spaceless)
				if !ok {
					t.Fatalf("expected Spaceless node, got %T", tmpl.Body[0])
				}
				if len(block.Body) != 1 {
					t.Fatalf("expected spaceless body to contain 1 node, got %d", len(block.Body))
				}
				if _, ok := block.Body[0].(*nodes.Output); !ok {
					t.Fatalf("expected spaceless body to render Output node, got %T", block.Body[0])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := NewParser(env, tt.template, "test", "test.html", "")
			if err != nil {
				t.Fatalf("failed to create parser: %v", err)
			}

			tmpl, err := parser.Parse()
			if err != nil {
				t.Fatalf("failed to parse template: %v", err)
			}

			tt.validate(t, tmpl)
		})
	}
}

func TestParser_AsyncStatements(t *testing.T) {
	template := `{% async for item in items %}{{ item }}{% endfor %}`
	if _, err := ParseTemplateWithEnv(&Environment{}, template, "test", "test.html"); err == nil {
		t.Fatalf("expected error when parsing async statements without enable_async")
	}

	asyncEnv := &Environment{EnableAsync: true}
	tmpl, err := ParseTemplateWithEnv(asyncEnv, template, "test", "test.html")
	if err != nil {
		t.Fatalf("unexpected error parsing async for: %v", err)
	}

	if len(tmpl.Body) != 1 {
		t.Fatalf("expected single body node, got %d", len(tmpl.Body))
	}
	forStmt, ok := tmpl.Body[0].(*nodes.For)
	if !ok {
		t.Fatalf("expected For node, got %T", tmpl.Body[0])
	}
	if !forStmt.Async {
		t.Fatalf("expected async flag to be set on async for")
	}

	withTemplate := `{% async with value = 5 %}{{ value }}{% endwith %}`
	tmpl, err = ParseTemplateWithEnv(asyncEnv, withTemplate, "test", "test.html")
	if err != nil {
		t.Fatalf("unexpected error parsing async with: %v", err)
	}
	if len(tmpl.Body) != 1 {
		t.Fatalf("expected single body node, got %d", len(tmpl.Body))
	}
	withStmt, ok := tmpl.Body[0].(*nodes.With)
	if !ok {
		t.Fatalf("expected With node, got %T", tmpl.Body[0])
	}
	if !withStmt.Async {
		t.Fatalf("expected async flag to be set on async with")
	}

	// Verify error message mentions async when disabled
	if _, err := ParseTemplateWithEnv(&Environment{}, `{% async with value = 5 %}{% endwith %}`, "test", "test.html"); err != nil {
		if !strings.Contains(err.Error(), "async") {
			t.Fatalf("expected error message to mention async, got %v", err)
		}
	}
}

func TestParser_AwaitExpressions(t *testing.T) {
	template := `{{ await fetch() }}`
	if _, err := ParseTemplateWithEnv(&Environment{}, template, "test", "test.html"); err == nil {
		t.Fatalf("expected error when parsing await expression without enable_async")
	}

	asyncEnv := &Environment{EnableAsync: true}
	tmpl, err := ParseTemplateWithEnv(asyncEnv, template, "test", "test.html")
	if err != nil {
		t.Fatalf("unexpected error parsing await expression: %v", err)
	}

	if len(tmpl.Body) != 1 {
		t.Fatalf("expected single body node, got %d", len(tmpl.Body))
	}

	output, ok := tmpl.Body[0].(*nodes.Output)
	if !ok {
		t.Fatalf("expected Output node, got %T", tmpl.Body[0])
	}

	if len(output.Nodes) != 1 {
		t.Fatalf("expected single expression in output, got %d", len(output.Nodes))
	}

	awaitExpr, ok := output.Nodes[0].(*nodes.Await)
	if !ok {
		t.Fatalf("expected Await node, got %T", output.Nodes[0])
	}

	if _, ok := awaitExpr.Node.(*nodes.Call); !ok {
		t.Fatalf("expected await to wrap a Call expression, got %T", awaitExpr.Node)
	}
}

func TestParser_MacroVarArgs(t *testing.T) {
	env := &Environment{}

	parser, err := NewParser(env, "{% macro test(a, b=1, *args, **kwargs) %}{% endmacro %}", "test", "test.html", "")
	if err != nil {
		t.Fatalf("failed to create parser: %v", err)
	}

	tmpl, err := parser.Parse()
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	if len(tmpl.Body) != 1 {
		t.Fatalf("expected single node in template body, got %d", len(tmpl.Body))
	}

	macro, ok := tmpl.Body[0].(*nodes.Macro)
	if !ok {
		t.Fatalf("expected Macro node, got %T", tmpl.Body[0])
	}

	if macro.VarArg == nil || macro.VarArg.Name != "args" {
		t.Fatalf("expected vararg name 'args', got %+v", macro.VarArg)
	}

	if macro.KwArg == nil || macro.KwArg.Name != "kwargs" {
		t.Fatalf("expected kwarg name 'kwargs', got %+v", macro.KwArg)
	}

	if len(macro.Args) != 2 {
		t.Fatalf("expected two positional args, got %d", len(macro.Args))
	}

	if len(macro.Defaults) != 1 {
		t.Fatalf("expected one default, got %d", len(macro.Defaults))
	}
}

func TestParser_MacroKeywordOnly(t *testing.T) {
	env := &Environment{}

	parser, err := NewParser(env, "{% macro test(a, *, b, c=42, **kwargs) %}{% endmacro %}", "test", "test.html", "")
	if err != nil {
		t.Fatalf("failed to create parser: %v", err)
	}

	tmpl, err := parser.Parse()
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	if len(tmpl.Body) != 1 {
		t.Fatalf("expected single node in template body, got %d", len(tmpl.Body))
	}

	macro, ok := tmpl.Body[0].(*nodes.Macro)
	if !ok {
		t.Fatalf("expected Macro node, got %T", tmpl.Body[0])
	}

	if len(macro.Args) != 1 {
		t.Fatalf("expected one positional arg, got %d", len(macro.Args))
	}

	if len(macro.KwonlyArgs) != 2 {
		t.Fatalf("expected two keyword-only args, got %d", len(macro.KwonlyArgs))
	}

	if macro.KwonlyArgs[0].Name != "b" || macro.KwonlyArgs[1].Name != "c" {
		t.Fatalf("unexpected keyword-only args: %+v", macro.KwonlyArgs)
	}

	if macro.KwDefaults == nil {
		t.Fatalf("expected keyword-only defaults map")
	}

	if _, ok := macro.KwDefaults["c"]; !ok {
		t.Fatalf("expected default for keyword-only arg 'c'")
	}

	if macro.KwArg == nil || macro.KwArg.Name != "kwargs" {
		t.Fatalf("expected kwarg collector, got %+v", macro.KwArg)
	}
}

func TestParserLineStatements(t *testing.T) {
	env := &Environment{LineStatementPrefix: "#"}

	template := "# for item in [1, 2]:\n- {{ item }}\n# endfor\n"

	parser, err := NewParser(env, template, "test", "test.html", "")
	if err != nil {
		t.Fatalf("failed to create parser: %v", err)
	}

	tmpl, err := parser.Parse()
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	if len(tmpl.Body) != 1 {
		t.Fatalf("expected 1 top-level node, got %d", len(tmpl.Body))
	}

	forStmt, ok := tmpl.Body[0].(*nodes.For)
	if !ok {
		t.Fatalf("expected first node to be For, got %T", tmpl.Body[0])
	}

	if len(forStmt.Body) == 0 {
		t.Fatalf("expected loop body to contain nodes")
	}

	if _, ok := forStmt.Body[0].(*nodes.Output); !ok {
		t.Fatalf("expected loop body to start with Output node, got %T", forStmt.Body[0])
	}
}

func TestParserLineStatementsAfterText(t *testing.T) {
	env := &Environment{LineStatementPrefix: "#"}

	template := "Title\n# for item in [1, 2]:\n- {{ item }}\n# endfor\n"

	parser, err := NewParser(env, template, "test", "test.html", "")
	if err != nil {
		t.Fatalf("failed to create parser: %v", err)
	}

	tmpl, err := parser.Parse()
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	if len(tmpl.Body) != 2 {
		t.Fatalf("expected 2 top-level nodes, got %d", len(tmpl.Body))
	}

	if _, ok := tmpl.Body[0].(*nodes.Output); !ok {
		t.Fatalf("expected first node to be Output, got %T", tmpl.Body[0])
	}

	if _, ok := tmpl.Body[1].(*nodes.For); !ok {
		t.Fatalf("expected second node to be For, got %T", tmpl.Body[1])
	}
}

func TestParserLineComments(t *testing.T) {
	env := &Environment{
		LineStatementPrefix: "#",
		LineCommentPrefix:   "//",
	}

	parser, err := NewParser(env, "// comment line\n{{ 42 }}\n", "test", "test.html", "")
	if err != nil {
		t.Fatalf("failed to create parser: %v", err)
	}

	tmpl, err := parser.Parse()
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	if len(tmpl.Body) != 1 {
		t.Fatalf("expected 1 node after stripping line comment, got %d", len(tmpl.Body))
	}

	output, ok := tmpl.Body[0].(*nodes.Output)
	if !ok {
		t.Fatalf("expected Output node, got %T", tmpl.Body[0])
	}

	if len(output.Nodes) != 1 {
		t.Fatalf("expected single expression in output, got %d", len(output.Nodes))
	}

	if constNode, ok := output.Nodes[0].(*nodes.Const); !ok || constNode.Value != int64(42) {
		t.Fatalf("expected constant 42, got %#v", output.Nodes[0])
	}
}

func TestParser_RawBlock(t *testing.T) {
	env := &Environment{}
	template := `{% raw %}{{ value }}{% endraw %}`

	parser, err := NewParser(env, template, "test", "test.html", "")
	if err != nil {
		t.Fatalf("failed to create parser: %v", err)
	}

	tmpl, err := parser.Parse()
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	if len(tmpl.Body) != 1 {
		t.Fatalf("expected 1 node in body, got %d", len(tmpl.Body))
	}

	output, ok := tmpl.Body[0].(*nodes.Output)
	if !ok {
		t.Fatalf("expected Output node, got %T", tmpl.Body[0])
	}

	if len(output.Nodes) != 1 {
		t.Fatalf("expected 1 node in output, got %d", len(output.Nodes))
	}

	data, ok := output.Nodes[0].(*nodes.TemplateData)
	if !ok {
		t.Fatalf("expected TemplateData node, got %T", output.Nodes[0])
	}

	if data.Data != "{{ value }}" {
		t.Fatalf("expected raw content '{{ value }}', got %q", data.Data)
	}
}

func TestParser_VerbatimBlock(t *testing.T) {
	env := &Environment{}
	template := `{% verbatim %}{{ value }}{% endverbatim %}`

	parser, err := NewParser(env, template, "test", "test.html", "")
	if err != nil {
		t.Fatalf("failed to create parser: %v", err)
	}

	tmpl, err := parser.Parse()
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	if len(tmpl.Body) != 1 {
		t.Fatalf("expected 1 node in body, got %d", len(tmpl.Body))
	}

	output, ok := tmpl.Body[0].(*nodes.Output)
	if !ok {
		t.Fatalf("expected Output node, got %T", tmpl.Body[0])
	}

	if len(output.Nodes) != 1 {
		t.Fatalf("expected 1 node in output, got %d", len(output.Nodes))
	}

	data, ok := output.Nodes[0].(*nodes.TemplateData)
	if !ok {
		t.Fatalf("expected TemplateData node, got %T", output.Nodes[0])
	}

	if data.Data != "{{ value }}" {
		t.Fatalf("expected verbatim content '{{ value }}', got %q", data.Data)
	}
}

func TestParser_RawBlockWhitespaceControl(t *testing.T) {
	env := &Environment{}
	template := `{%- raw -%}{{ value }}{%- endraw -%}`

	parser, err := NewParser(env, template, "test", "test.html", "")
	if err != nil {
		t.Fatalf("failed to create parser: %v", err)
	}

	tmpl, err := parser.Parse()
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	if len(tmpl.Body) != 1 {
		t.Fatalf("expected 1 node in body, got %d", len(tmpl.Body))
	}

	output, ok := tmpl.Body[0].(*nodes.Output)
	if !ok {
		t.Fatalf("expected Output node, got %T", tmpl.Body[0])
	}

	if len(output.Nodes) != 1 {
		t.Fatalf("expected 1 node in output, got %d", len(output.Nodes))
	}

	data, ok := output.Nodes[0].(*nodes.TemplateData)
	if !ok {
		t.Fatalf("expected TemplateData node, got %T", output.Nodes[0])
	}

	if data.Data != "{{ value }}" {
		t.Fatalf("expected raw content '{{ value }}', got %q", data.Data)
	}
}

func TestParser_ComplexExpressions(t *testing.T) {
	env := &Environment{}

	tests := []struct {
		name     string
		template string
		validate func(*testing.T, *nodes.Template)
	}{
		{
			name:     "NestedFilters",
			template: "{{ name | upper | truncate(10) }}",
			validate: func(t *testing.T, tmpl *nodes.Template) {
				output := tmpl.Body[0].(*nodes.Output)
				filter := output.Nodes[0].(*nodes.Filter)
				if filter.Name != "truncate" {
					t.Errorf("expected outer filter 'truncate', got '%s'", filter.Name)
				}
				// The inner filter should be the node of the outer filter
				innerFilter, ok := filter.Node.(*nodes.Filter)
				if !ok {
					t.Fatalf("expected inner Filter node, got %T", filter.Node)
				}
				if innerFilter.Name != "upper" {
					t.Errorf("expected inner filter 'upper', got '%s'", innerFilter.Name)
				}
			},
		},
		{
			name:     "AttributeAccess",
			template: "{{ user.name }}",
			validate: func(t *testing.T, tmpl *nodes.Template) {
				output := tmpl.Body[0].(*nodes.Output)
				getattr := output.Nodes[0].(*nodes.Getattr)
				if getattr.Attr != "name" {
					t.Errorf("expected attribute 'name', got '%s'", getattr.Attr)
				}
			},
		},
		{
			name:     "MethodCall",
			template: "{{ user.get_name() }}",
			validate: func(t *testing.T, tmpl *nodes.Template) {
				output := tmpl.Body[0].(*nodes.Output)
				call := output.Nodes[0].(*nodes.Call)
				getattr := call.Node.(*nodes.Getattr)
				if getattr.Attr != "get_name" {
					t.Errorf("expected method 'get_name', got '%s'", getattr.Attr)
				}
			},
		},
		{
			name:     "ListLiteral",
			template: "{{ [1, 2, 3] }}",
			validate: func(t *testing.T, tmpl *nodes.Template) {
				output := tmpl.Body[0].(*nodes.Output)
				list := output.Nodes[0].(*nodes.List)
				if len(list.Items) != 3 {
					t.Fatalf("expected 3 items in list, got %d", len(list.Items))
				}
			},
		},
		{
			name:     "DictLiteral",
			template: "{{ {'key': 'value', 'num': 42} }}",
			validate: func(t *testing.T, tmpl *nodes.Template) {
				output := tmpl.Body[0].(*nodes.Output)
				dict := output.Nodes[0].(*nodes.Dict)
				if len(dict.Items) != 2 {
					t.Fatalf("expected 2 items in dict, got %d", len(dict.Items))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := NewParser(env, tt.template, "test", "test.html", "")
			if err != nil {
				t.Fatalf("failed to create parser: %v", err)
			}

			tmpl, err := parser.Parse()
			if err != nil {
				t.Fatalf("failed to parse template: %v", err)
			}

			tt.validate(t, tmpl)
		})
	}
}

func TestParser_ErrorHandling(t *testing.T) {
	env := &Environment{}

	tests := []struct {
		name     string
		template string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "UnclosedVariable",
			template: "{{ name",
			wantErr:  true,
			errMsg:   "missing }}",
		},
		{
			name:     "UnclosedBlock",
			template: "{% if condition %}",
			wantErr:  true,
			errMsg:   "Jinja was looking for",
		},
		{
			name:     "UnknownTag",
			template: "{% unknown %}",
			wantErr:  true,
			errMsg:   "unknown tag",
		},
		{
			name:     "InvalidSyntax",
			template: "{{ + }}",
			wantErr:  true,
			errMsg:   "expected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := NewParser(env, tt.template, "test", "test.html", "")
			if err != nil {
				// Check if this error matches our expectations
				if !tt.wantErr {
					t.Fatalf("failed to create parser: %v", err)
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("error message '%s' should contain '%s'", err.Error(), tt.errMsg)
				}
				return
			}

			_, err = parser.Parse()
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errMsg != "" {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("error message '%s' should contain '%s'", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestParser_OperatorPrecedence(t *testing.T) {
	env := &Environment{}

	// Test that multiplication has higher precedence than addition
	template := "{{ a + b * c }}"
	parser, err := NewParser(env, template, "test", "test.html", "")
	if err != nil {
		t.Fatalf("failed to create parser: %v", err)
	}

	tmpl, err := parser.Parse()
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	output := tmpl.Body[0].(*nodes.Output)
	add := output.Nodes[0].(*nodes.Add)

	// The right side of addition should be multiplication
	if _, ok := add.Right.(*nodes.Mul); !ok {
		t.Errorf("expected multiplication on right side of addition, got %T", add.Right)
	}
}

func TestParser_ConditionalExpression(t *testing.T) {
	env := &Environment{}

	template := "{{ a if condition else b }}"
	parser, err := NewParser(env, template, "test", "test.html", "")
	if err != nil {
		t.Fatalf("failed to create parser: %v", err)
	}

	tmpl, err := parser.Parse()
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	output := tmpl.Body[0].(*nodes.Output)
	condExpr := output.Nodes[0].(*nodes.CondExpr)

	if condExpr.Test == nil {
		t.Fatal("expected test expression")
	}
	if condExpr.Expr1 == nil {
		t.Fatal("expected expr1")
	}
	if condExpr.Expr2 == nil {
		t.Fatal("expected expr2")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			func() bool {
				for i := 1; i <= len(s)-len(substr); i++ {
					if s[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			}())))
}

// BenchmarkParser_ParseSimple benchmarks parsing a simple template
func BenchmarkParser_ParseSimple(b *testing.B) {
	env := &Environment{}
	template := "Hello {{ name }}!"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser, err := NewParser(env, template, "bench", "bench.html", "")
		if err != nil {
			b.Fatalf("failed to create parser: %v", err)
		}
		_, err = parser.Parse()
		if err != nil {
			b.Fatalf("failed to parse template: %v", err)
		}
	}
}

// BenchmarkParser_ParseComplex benchmarks parsing a more complex template
func BenchmarkParser_ParseComplex(b *testing.B) {
	env := &Environment{}
	template := `
	<html>
		<head><title>{{ title | default("Default Title") }}</title></head>
		<body>
			{% for user in users %}
				<div class="user">
					<h2>{{ user.name | title }}</h2>
					{% if user.email %}
						<p>{{ user.email }}</p>
					{% endif %}
					<ul>
						{% for item in user.items %}
							<li>{{ item.name }}: ${{ item.price | round(2) }}</li>
						{% endfor %}
					</ul>
				</div>
			{% endfor %}
		</body>
	</html>
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser, err := NewParser(env, template, "bench", "bench.html", "")
		if err != nil {
			b.Fatalf("failed to create parser: %v", err)
		}
		_, err = parser.Parse()
		if err != nil {
			b.Fatalf("failed to parse template: %v", err)
		}
	}
}
