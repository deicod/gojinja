package runtime

import (
	"strings"
	"testing"

	"github.com/deicod/gojinja/nodes"
)

func TestMacroRegistry(t *testing.T) {
	registry := NewMacroRegistry()

	// Test registration and retrieval
	macro := &Macro{
		Name: "test_macro",
		Arguments: []*MacroArgument{
			{Name: "arg1"},
			{Name: "arg2", HasDefault: true},
		},
		Body: []nodes.Node{},
	}

	registry.RegisterGlobal("test_macro", macro)

	found, err := registry.FindMacro(nil, "test_macro")
	if err != nil {
		t.Fatalf("Failed to find macro: %v", err)
	}

	if found.Name != "test_macro" {
		t.Errorf("Expected macro name 'test_macro', got '%s'", found.Name)
	}

	// Test non-existent macro
	_, err = registry.FindMacro(nil, "non_existent")
	if err == nil {
		t.Error("Expected error when finding non-existent macro")
	}
}

func TestMacroExecution(t *testing.T) {
	// Create a simple macro
	macro := &Macro{
		Name: "greet",
		Arguments: []*MacroArgument{
			{Name: "name"},
			{Name: "greeting", HasDefault: true, Default: &nodes.Const{Value: "Hello"}},
		},
		Defaults: []nodes.Expr{&nodes.Const{Value: "Hello"}},
		Body: []nodes.Node{
			&nodes.Output{
				Nodes: []nodes.Expr{
					&nodes.Concat{
						Nodes: []nodes.Expr{
							&nodes.Name{Name: "greeting", Ctx: nodes.CtxLoad},
							&nodes.TemplateData{Data: ", "},
							&nodes.Name{Name: "name", Ctx: nodes.CtxLoad},
							&nodes.TemplateData{Data: "!"},
						},
					},
				},
			},
		},
	}

	// Create context and environment
	env := NewEnvironment()
	ctx := NewContextWithEnvironment(env, map[string]interface{}{})

	// Test execution with all arguments
	result, err := macro.CallKwargs(ctx, []interface{}{"World"}, nil)
	if err != nil {
		t.Fatalf("Failed to execute macro: %v", err)
	}

	expected := "Hello, World!"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}

	// Test execution with keyword arguments
	result, err = macro.CallKwargs(ctx, []interface{}{}, map[string]interface{}{
		"name":     "Alice",
		"greeting": "Hi",
	})
	if err != nil {
		t.Fatalf("Failed to execute macro with kwargs: %v", err)
	}

	expected = "Hi, Alice!"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}

	// Test missing required argument
	_, err = macro.CallKwargs(ctx, []interface{}{}, nil)
	if err == nil {
		t.Error("Expected error for missing required argument")
	}
}

func TestMacroNamespace(t *testing.T) {
	namespace := NewMacroNamespace("helpers", nil)

	// Add macros to namespace
	macro1 := &Macro{
		Name: "input",
		Arguments: []*MacroArgument{
			{Name: "name"},
			{Name: "type", HasDefault: true, Default: &nodes.Const{Value: "text"}},
		},
		Defaults: []nodes.Expr{&nodes.Const{Value: "text"}},
		Body: []nodes.Node{
			&nodes.Output{
				Nodes: []nodes.Expr{
					&nodes.Concat{
						Nodes: []nodes.Expr{
							&nodes.TemplateData{Data: "<input type=\""},
							&nodes.Name{Name: "type", Ctx: nodes.CtxLoad},
							&nodes.TemplateData{Data: "\" name=\""},
							&nodes.Name{Name: "name", Ctx: nodes.CtxLoad},
							&nodes.TemplateData{Data: "\">"},
						},
					},
				},
			},
		},
	}

	namespace.AddMacro("input", macro1)

	// Test retrieval
	found, err := namespace.GetMacro("input")
	if err != nil {
		t.Fatalf("Failed to get macro: %v", err)
	}

	if found.Name != "input" {
		t.Errorf("Expected macro name 'input', got '%s'", found.Name)
	}

	// Test non-existent macro
	_, err = namespace.GetMacro("non_existent")
	if err == nil {
		t.Error("Expected error when getting non-existent macro")
	}

	// Test HasMacro
	if !namespace.HasMacro("input") {
		t.Error("Expected namespace to have 'input' macro")
	}

	if namespace.HasMacro("non_existent") {
		t.Error("Expected namespace to not have 'non_existent' macro")
	}

	// Test GetMacroNames
	names := namespace.GetMacroNames()
	if len(names) != 1 || names[0] != "input" {
		t.Errorf("Expected ['input'], got %v", names)
	}
}

func TestMacroValidation(t *testing.T) {
	macro := &Macro{
		Name: "test",
		Arguments: []*MacroArgument{
			{Name: "required"},
			{Name: "optional", HasDefault: true, Default: &nodes.Const{Value: "default"}},
		},
		Defaults: []nodes.Expr{&nodes.Const{Value: "default"}},
		Body:     []nodes.Node{},
	}

	// Test valid call
	err := macro.ValidateCall([]interface{}{"value"}, nil)
	if err != nil {
		t.Errorf("Expected no error for valid call: %v", err)
	}

	// Test missing required argument
	err = macro.ValidateCall([]interface{}{}, nil)
	if err == nil {
		t.Error("Expected error for missing required argument")
	}

	// Test too many arguments (if no variadic)
	err = macro.ValidateCall([]interface{}{"value1", "value2", "value3"}, nil)
	if err == nil {
		t.Error("Expected error for too many arguments")
	}

	// Test unexpected keyword argument
	err = macro.ValidateCall([]interface{}{"value"}, map[string]interface{}{
		"unexpected": "value",
	})
	if err == nil {
		t.Error("Expected error for unexpected keyword argument")
	}
}

func TestImportManager(t *testing.T) {
	env := NewEnvironment()
	importManager := NewImportManager(env)

	// Create a simple template with macros
	templateContent := `
{% macro helper_macro(name) %}
Hello, {{ name }}!
{% endmacro %}

{% macro complex_macro(a, b=2) %}
{{ a }} + {{ b }} = {{ a + b }}
{% endmacro %}
`

	// Create a mock loader
	loader := NewMapLoader(map[string]string{
		"helpers.html": templateContent,
	})
	env.SetLoader(loader)

	ctx := NewContextWithEnvironment(env, map[string]interface{}{})

	// Test template import
	namespace, err := importManager.ImportTemplate(ctx, "helpers.html", false)
	if err != nil {
		t.Fatalf("Failed to import template: %v", err)
	}

	if namespace.Name != "helpers.html" {
		t.Errorf("Expected namespace name 'helpers.html', got '%s'", namespace.Name)
	}

	// Test macro retrieval from namespace
	macro, err := namespace.GetMacro("helper_macro")
	if err != nil {
		t.Fatalf("Failed to get macro from namespace: %v", err)
	}

	if macro.Name != "helper_macro" {
		t.Errorf("Expected macro name 'helper_macro', got '%s'", macro.Name)
	}

	// Test macro execution
	result, err := macro.Call(ctx, "World")
	if err != nil {
		t.Fatalf("Failed to execute imported macro: %v", err)
	}

	expected := "Hello, World!"
	if strings.TrimSpace(result.(string)) != expected {
		t.Errorf("Expected '%s', got '%s'", expected, strings.TrimSpace(result.(string)))
	}

	// Test from import
	macros, err := importManager.ImportMacros(ctx, "helpers.html", []string{"helper_macro", "complex_macro"}, false)
	if err != nil {
		t.Fatalf("Failed to import macros: %v", err)
	}

	if len(macros) != 2 {
		t.Errorf("Expected 2 macros, got %d", len(macros))
	}

	if _, exists := macros["helper_macro"]; !exists {
		t.Error("Expected 'helper_macro' to be imported")
	}

	if _, exists := macros["complex_macro"]; !exists {
		t.Error("Expected 'complex_macro' to be imported")
	}
}

func TestMacroArgumentBinding(t *testing.T) {
	macro := &Macro{
		Name: "test_args",
		Arguments: []*MacroArgument{
			{Name: "pos1"},
			{Name: "pos2"},
			{Name: "kw1", Keyword: true},
			{Name: "default1", HasDefault: true, Default: &nodes.Const{Value: "default_value"}},
		},
		Defaults: []nodes.Expr{&nodes.Const{Value: "default_value"}},
		Body: []nodes.Node{
			&nodes.Output{
				Nodes: []nodes.Expr{
					&nodes.Concat{
						Nodes: []nodes.Expr{
							&nodes.Name{Name: "pos1", Ctx: nodes.CtxLoad},
							&nodes.TemplateData{Data: "|"},
							&nodes.Name{Name: "pos2", Ctx: nodes.CtxLoad},
							&nodes.TemplateData{Data: "|"},
							&nodes.Name{Name: "kw1", Ctx: nodes.CtxLoad},
							&nodes.TemplateData{Data: "|"},
							&nodes.Name{Name: "default1", Ctx: nodes.CtxLoad},
						},
					},
				},
			},
		},
	}

	ctx := NewContextWithEnvironment(NewEnvironment(), map[string]interface{}{})

	// Test positional arguments
	result, err := macro.CallKwargs(ctx, []interface{}{"val1", "val2"}, map[string]interface{}{
		"kw1": "val3",
	})
	if err != nil {
		t.Fatalf("Failed to execute macro: %v", err)
	}

	expected := "val1|val2|val3|default_value"
	if strings.TrimSpace(result.(string)) != expected {
		t.Errorf("Expected '%s', got '%s'", expected, strings.TrimSpace(result.(string)))
	}

	// Test with default override
	result, err = macro.CallKwargs(ctx, []interface{}{"val1", "val2"}, map[string]interface{}{
		"kw1":      "val3",
		"default1": "custom_value",
	})
	if err != nil {
		t.Fatalf("Failed to execute macro with custom default: %v", err)
	}

	expected = "val1|val2|val3|custom_value"
	if strings.TrimSpace(result.(string)) != expected {
		t.Errorf("Expected '%s', got '%s'", expected, strings.TrimSpace(result.(string)))
	}
}

func TestMacroContext(t *testing.T) {
	env := NewEnvironment()
	ctx := NewContextWithEnvironment(env, map[string]interface{}{
		"global_var": "global_value",
	})

	// Test macro stack
	macro1 := &Macro{Name: "macro1"}
	macro2 := &Macro{Name: "macro2"}

	ctx.PushMacro(macro1)
	if ctx.CurrentMacro() != macro1 {
		t.Error("Expected current macro to be macro1")
	}

	if !ctx.InMacro() {
		t.Error("Expected to be in macro context")
	}

	if ctx.MacroDepth() != 1 {
		t.Error("Expected macro depth to be 1")
	}

	ctx.PushMacro(macro2)
	if ctx.CurrentMacro() != macro2 {
		t.Error("Expected current macro to be macro2")
	}

	if ctx.MacroDepth() != 2 {
		t.Error("Expected macro depth to be 2")
	}

	ctx.PopMacro()
	if ctx.CurrentMacro() != macro1 {
		t.Error("Expected current macro to be macro1 after pop")
	}

	ctx.PopMacro()
	if ctx.CurrentMacro() != nil {
		t.Error("Expected current macro to be nil after popping all")
	}

	if ctx.InMacro() {
		t.Error("Expected to not be in macro context after popping all")
	}

	// Test caller stack
	caller := ctx.CreateMacroCaller("caller_macro", []interface{}{"arg1"}, map[string]interface{}{
		"kwarg1": "value1",
	})

	ctx.PushCaller(caller)
	if !ctx.HasCaller() {
		t.Error("Expected to have caller context")
	}

	if ctx.CurrentCaller() != caller {
		t.Error("Expected current caller to be the pushed caller")
	}

	ctx.PopCaller()
	if ctx.HasCaller() {
		t.Error("Expected to not have caller context after pop")
	}

	// Test macro variable resolution
	ctx.PushMacro(macro1)
	ctx.PushCaller(caller)

	// Set variables in caller context
	ctx.SetMacroVariable("caller_var", "caller_value")

	// Test retrieval
	value, ok := ctx.GetMacroVariable("caller_var")
	if !ok {
		t.Error("Expected to find caller_var")
	}

	if value != "caller_value" {
		t.Errorf("Expected 'caller_value', got '%v'", value)
	}

	// Test global variable access
	value, ok = ctx.GetMacroVariable("global_var")
	if !ok {
		t.Error("Expected to find global_var")
	}

	if value != "global_value" {
		t.Errorf("Expected 'global_value', got '%v'", value)
	}
}

func TestEnvironmentMacroIntegration(t *testing.T) {
	env := NewEnvironment()

	// Create a global macro
	globalMacro := &Macro{
		Name: "global_helper",
		Arguments: []*MacroArgument{
			{Name: "text"},
		},
		Body: []nodes.Node{
			&nodes.Output{
				Nodes: []nodes.Expr{
					&nodes.Concat{
						Nodes: []nodes.Expr{
							&nodes.TemplateData{Data: "[GLOBAL: "},
							&nodes.Name{Name: "text", Ctx: nodes.CtxLoad},
							&nodes.TemplateData{Data: "]"},
						},
					},
				},
			},
		},
	}

	env.AddGlobalMacro("global_helper", globalMacro)

	// Test retrieval
	found, err := env.GetGlobalMacro("global_helper")
	if err != nil {
		t.Fatalf("Failed to get global macro: %v", err)
	}

	if found.Name != "global_helper" {
		t.Errorf("Expected macro name 'global_helper', got '%s'", found.Name)
	}

	// Test stats
	stats := env.GetMacroStats()
	if stats["globals"] != 1 {
		t.Errorf("Expected 1 global macro, got %d", stats["globals"])
	}

	// Clear and test
	env.ClearMacroRegistry()
	stats = env.GetMacroStats()
	if stats["globals"] != 0 {
		t.Errorf("Expected 0 global macros after clear, got %d", stats["globals"])
	}
}

func TestMacroErrors(t *testing.T) {
	macro := &Macro{
		Name: "error_macro",
		Arguments: []*MacroArgument{
			{Name: "required_arg"},
		},
		Body: []nodes.Node{},
		Position: nodes.Position{Line: 10, Column: 5},
	}

	ctx := NewContextWithEnvironment(NewEnvironment(), map[string]interface{}{})

	// Test missing required argument
	_, err := macro.Call(ctx)
	if err == nil {
		t.Error("Expected error for missing required argument")
	}

	if !IsMacroError(err) {
		t.Error("Expected MacroError type")
	}

	// Test invalid macro name in registry
	registry := NewMacroRegistry()
	_, err = registry.FindMacro(ctx, "non_existent_macro")
	if err == nil {
		t.Error("Expected error for non-existent macro")
	}

	if !IsMacroError(err) {
		t.Error("Expected MacroError type")
	}
}

func TestMacroComplexScenario(t *testing.T) {
	// Create a complex scenario with nested macros and imports
	env := NewEnvironment()

	// Create templates
	macrosTemplate := `
{% macro form_input(name, type="text", value="") %}
<input type="{{ type }}" name="{{ name }}" value="{{ value }}">
{% endmacro %}

{% macro user_list(users) %}
<ul>
{% for user in users %}
  <li>{{ user.name|escape }}</li>
{% endfor %}
</ul>
{% endmacro %}
`

	mainTemplate := `
{% from 'macros.html' import form_input, user_list %}

<form>
{{ form_input("username", value=user.name) }}
{{ form_input("email", type="email") }}
</form>

{{ user_list(users) }}
`

	loader := NewMapLoader(map[string]string{
		"macros.html": macrosTemplate,
		"main.html":  mainTemplate,
	})
	env.SetLoader(loader)

	// Parse and render main template
	template, err := env.LoadTemplate("main.html")
	if err != nil {
		t.Fatalf("Failed to load template: %v", err)
	}

	ctx := NewContextWithEnvironment(env, map[string]interface{}{
		"user": map[string]interface{}{
			"name": "john_doe",
		},
		"users": []map[string]interface{}{
			{"name": "Alice"},
			{"name": "Bob"},
			{"name": "Charlie"},
		},
	})

	result, err := template.ExecuteToString(ctx.scope.All())
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	// Basic checks
	if !strings.Contains(result, "<form>") {
		t.Error("Expected output to contain <form> tag")
	}

	if !strings.Contains(result, "<input") {
		t.Error("Expected output to contain input elements")
	}

	if !strings.Contains(result, "<ul>") {
		t.Error("Expected output to contain user list")
	}

	if !strings.Contains(result, "Alice") {
		t.Error("Expected output to contain Alice")
	}
}