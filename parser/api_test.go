package parser

import (
	"testing"
	"github.com/deicod/gojinja/nodes"
)

func TestAPI_ParseTemplate(t *testing.T) {
	// Test the complex template mentioned in the requirements
	template := `Hello {{ user.name }}, balance: {{ balance + 42.5 }}`

	ast, err := ParseTemplate(template)
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	if len(ast.Body) != 1 {
		t.Fatalf("expected 1 node in body, got %d", len(ast.Body))
	}

	output, ok := ast.Body[0].(*nodes.Output)
	if !ok {
		t.Fatalf("expected Output node, got %T", ast.Body[0])
	}

	// Should have 4 expression parts: "Hello ", user.name, ", balance: ", balance + 42.5
	if len(output.Nodes) != 4 {
		t.Fatalf("expected 4 expressions in output, got %d", len(output.Nodes))
	}

	// Check first part is template data
	if _, ok := output.Nodes[0].(*nodes.TemplateData); !ok {
		t.Errorf("expected first node to be TemplateData, got %T", output.Nodes[0])
	}

	// Check second part is attribute access
	if getattr, ok := output.Nodes[1].(*nodes.Getattr); !ok {
		t.Errorf("expected second node to be Getattr, got %T", output.Nodes[1])
	} else if getattr.Attr != "name" {
		t.Errorf("expected attribute 'name', got '%s'", getattr.Attr)
	}

	// Check fourth part is addition
	if add, ok := output.Nodes[3].(*nodes.Add); !ok {
		t.Errorf("expected fourth node to be Add, got %T", output.Nodes[3])
	} else {
		// Check left side is a name
		if _, ok := add.Left.(*nodes.Name); !ok {
			t.Errorf("expected left side of Add to be Name, got %T", add.Left)
		}
		// Check right side is a constant
		if constNode, ok := add.Right.(*nodes.Const); !ok {
			t.Errorf("expected right side of Add to be Const, got %T", add.Right)
		} else if constNode.Value != 42.5 {
			t.Errorf("expected constant value 42.5, got %v", constNode.Value)
		}
	}
}

func TestAPI_ParseTemplateErrorHandling(t *testing.T) {
	tests := []struct {
		name     string
		template string
		wantErr  bool
	}{
		{
			name:     "ValidTemplate",
			template: "{{ name }}",
			wantErr:  false,
		},
		{
			name:     "UnclosedVariable",
			template: "{{ name",
			wantErr:  true,
		},
		{
			name:     "UnclosedBlock",
			template: "{% if condition %}",
			wantErr:  true,
		},
		{
			name:     "InvalidSyntax",
			template: "{{ + }}",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseTemplate(tt.template)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTemplate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAPI_ParseComplexFeatures(t *testing.T) {
	// Test dictionary literal parsing
	template := `{{ {'key': 'value', 'num': 42} }}`
	ast, err := ParseTemplate(template)
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	output := ast.Body[0].(*nodes.Output)
	dict, ok := output.Nodes[0].(*nodes.Dict)
	if !ok {
		t.Fatalf("expected Dict node, got %T", output.Nodes[0])
	}

	if len(dict.Items) != 2 {
		t.Fatalf("expected 2 items in dict, got %d", len(dict.Items))
	}

	// Test filter chaining
	template = `{{ name | upper | truncate(10) }}`
	ast, err = ParseTemplate(template)
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	output = ast.Body[0].(*nodes.Output)
	filter, ok := output.Nodes[0].(*nodes.Filter)
	if !ok {
		t.Fatalf("expected Filter node, got %T", output.Nodes[0])
	}

	if filter.Name != "truncate" {
		t.Errorf("expected outer filter 'truncate', got '%s'", filter.Name)
	}

	// Test conditional expression
	template = `{{ a if condition else b }}`
	ast, err = ParseTemplate(template)
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	output = ast.Body[0].(*nodes.Output)
	condExpr, ok := output.Nodes[0].(*nodes.CondExpr)
	if !ok {
		t.Fatalf("expected CondExpr node, got %T", output.Nodes[0])
	}

	if condExpr.Test == nil || condExpr.Expr1 == nil || condExpr.Expr2 == nil {
		t.Fatal("conditional expression missing required components")
	}
}