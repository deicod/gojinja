package runtime

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/deicod/gojinja/nodes"
)

// Macro represents a compiled Jinja2 macro
type Macro struct {
	// Basic macro information
	Name      string
	Arguments []*MacroArgument
	Defaults  []nodes.Expr
	Body      []nodes.Node
	Template  *Template
	Position  nodes.Position

	// Execution context
	Caller      *MacroCaller
	CallContext *Context
	callerFunc  GlobalFunc

	// Thread safety
	mu sync.RWMutex
}

// MacroArgument represents a macro argument with optional default
type MacroArgument struct {
	Name       string
	Default    interface{}
	HasDefault bool
	Variadic   bool
	Keyword    bool
}

// MacroCaller represents the caller context for a macro
type MacroCaller struct {
	Name      string
	Args      []interface{}
	Kwargs    map[string]interface{}
	Variables map[string]interface{}
}

// MacroNamespace represents a namespace for imported macros
type MacroNamespace struct {
	Name     string
	Macros   map[string]*Macro
	Template *Template
	Context  *Context

	// Thread safety
	mu sync.RWMutex
}

// MacroRegistry manages macro storage and resolution
type MacroRegistry struct {
	// Global macros
	globals map[string]*Macro

	// Template-level macros
	templates map[string]map[string]*Macro

	// Imported namespaces
	namespaces map[string]*MacroNamespace

	// Thread safety
	mu sync.RWMutex
}

// NewMacroRegistry creates a new macro registry
func NewMacroRegistry() *MacroRegistry {
	return &MacroRegistry{
		globals:    make(map[string]*Macro),
		templates:  make(map[string]map[string]*Macro),
		namespaces: make(map[string]*MacroNamespace),
	}
}

// NewMacro creates a new macro from an AST node
func NewMacro(macroNode *nodes.Macro, template *Template) *Macro {
	args := make([]*MacroArgument, len(macroNode.Args))
	for i, arg := range macroNode.Args {
		// Check if this argument has a default (right-to-left mapping)
		defaultIndex := len(macroNode.Args) - 1 - i
		hasDefault := defaultIndex >= 0 && defaultIndex < len(macroNode.Defaults) && macroNode.Defaults[defaultIndex] != nil
		args[i] = &MacroArgument{
			Name:       arg.Name,
			HasDefault: hasDefault,
		}
	}

	// Set defaults for arguments that have them (right-to-left mapping)
	for i, arg := range args {
		if arg.HasDefault {
			defaultIndex := len(args) - 1 - i
			if defaultIndex >= 0 && defaultIndex < len(macroNode.Defaults) {
				arg.Default = macroNode.Defaults[defaultIndex]
			}
		}
	}

	macro := &Macro{
		Name:      macroNode.Name,
		Arguments: args,
		Defaults:  macroNode.Defaults,
		Body:      macroNode.Body,
		Template:  template,
		Position:  macroNode.GetPosition(),
	}

	return macro
}

// NewMacroNamespace creates a new macro namespace
func NewMacroNamespace(name string, template *Template) *MacroNamespace {
	return &MacroNamespace{
		Name:     name,
		Macros:   make(map[string]*Macro),
		Template: template,
	}
}

// RegisterGlobal registers a global macro
func (r *MacroRegistry) RegisterGlobal(name string, macro *Macro) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.globals[name] = macro
}

// RegisterTemplate registers a template-level macro
func (r *MacroRegistry) RegisterTemplate(templateName, macroName string, macro *Macro) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.templates[templateName] == nil {
		r.templates[templateName] = make(map[string]*Macro)
	}
	r.templates[templateName][macroName] = macro
}

// RegisterNamespace registers a macro namespace
func (r *MacroRegistry) RegisterNamespace(name string, namespace *MacroNamespace) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.namespaces[name] = namespace
}

// FindMacro finds a macro by name, searching in order: local scope, template, globals
func (r *MacroRegistry) FindMacro(ctx *Context, name string) (*Macro, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// First check if it's in the current context/scope
	if ctx != nil {
		if value, ok := ctx.Get(name); ok {
			if macro, ok := value.(*Macro); ok {
				return macro, nil
			}
			// Check if it's a callable macro function
			if _, ok := value.(func(...interface{}) (interface{}, error)); ok {
				// Create a wrapper macro for the function
				return &Macro{
					Name: name,
					Caller: &MacroCaller{
						Name:      name,
						Variables: make(map[string]interface{}),
					},
				}, nil
			}
		}
	}

	// Check template-level macros
	if ctx != nil && ctx.current != nil {
		if templateMacros, exists := r.templates[ctx.current.name]; exists {
			if macro, exists := templateMacros[name]; exists {
				return macro, nil
			}
		}
	}

	// Check global macros
	if macro, exists := r.globals[name]; exists {
		return macro, nil
	}

	return nil, NewMacroError(name, "macro not found", nodes.Position{}, nil)
}

// ResolveMacroPath resolves a macro path (e.g., "namespace.macro")
func (r *MacroRegistry) ResolveMacroPath(ctx *Context, path string) (*Macro, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Split the path to check for namespace notation
	parts := strings.Split(path, ".")
	if len(parts) == 1 {
		// Simple macro name, use existing FindMacro
		return r.FindMacro(ctx, parts[0])
	}

	// Namespace resolution
	namespaceName := strings.Join(parts[:len(parts)-1], ".")
	macroName := parts[len(parts)-1]

	// Find namespace
	namespace, exists := r.namespaces[namespaceName]
	if !exists {
		return nil, NewMacroError(path, fmt.Sprintf("namespace '%s' not found", namespaceName), nodes.Position{}, nil)
	}

	// Find macro in namespace
	macro, exists := namespace.Macros[macroName]
	if !exists {
		return nil, NewMacroError(path, fmt.Sprintf("macro '%s' not found in namespace '%s'", macroName, namespaceName), nodes.Position{}, nil)
	}

	return macro, nil
}

// FindNamespaceMacro finds a macro in a specific namespace
func (r *MacroRegistry) FindNamespaceMacro(namespaceName, macroName string) (*Macro, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	namespace, exists := r.namespaces[namespaceName]
	if !exists {
		return nil, NewMacroError(fmt.Sprintf("%s.%s", namespaceName, macroName),
			fmt.Sprintf("namespace '%s' not found", namespaceName), nodes.Position{}, nil)
	}

	macro, exists := namespace.Macros[macroName]
	if !exists {
		return nil, NewMacroError(fmt.Sprintf("%s.%s", namespaceName, macroName),
			fmt.Sprintf("macro '%s' not found in namespace '%s'", macroName, namespaceName),
			nodes.Position{}, nil)
	}

	return macro, nil
}

// GetNamespace returns a macro namespace
func (r *MacroRegistry) GetNamespace(name string) (*MacroNamespace, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	namespace, exists := r.namespaces[name]
	if !exists {
		return nil, NewMacroError(name, fmt.Sprintf("namespace '%s' not found", name), nodes.Position{}, nil)
	}

	return namespace, nil
}

// Execute executes a macro with the given arguments
func (m *Macro) Execute(ctx *Context, args []interface{}, kwargs map[string]interface{}) (interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ctx.PushMacro(m)
	defer ctx.PopMacro()

	// Create a new scope for the macro
	ctx.PushScope()
	defer ctx.PopScope()

	// Bind arguments to macro parameters
	if err := m.bindArguments(ctx, args, kwargs); err != nil {
		return nil, err
	}

	// Execute macro body
	var result strings.Builder
	oldWriter := ctx.writer
	ctx.writer = &result
	defer func() { ctx.writer = oldWriter }()

	// Use the template's evaluator if available
	var evaluator *Evaluator
	if m.Template != nil {
		evaluator = NewEvaluator(ctx)
	} else {
		evaluator = NewEvaluator(ctx)
	}

	for _, node := range m.Body {
		value := evaluator.Evaluate(node)
		if err, ok := value.(error); ok {
			return nil, err
		}
		if signal, ok := isControlSignal(value); ok {
			return signal, nil
		}
	}

	return Markup(result.String()), nil
}

// bindArguments binds function arguments to macro parameters
func (m *Macro) bindArguments(ctx *Context, args []interface{}, kwargs map[string]interface{}) error {
	// Convert arguments to a more manageable form
	argMap := make(map[string]interface{})
	var positionalArgs []interface{}
	var keywordArgs []string

	// First, process positional arguments to identify which are positional
	for i, arg := range m.Arguments {
		if !arg.Keyword && !arg.Variadic {
			if i < len(args) {
				argMap[arg.Name] = args[i]
				positionalArgs = append(positionalArgs, arg.Name)
			}
		}
	}

	// Process keyword arguments
	for key, value := range kwargs {
		if key == "__caller" {
			continue
		}
		argMap[key] = value
		keywordArgs = append(keywordArgs, key)
	}

	// Handle variadic arguments (if supported)
	var variadicArgs []interface{}

	// Bind remaining positional arguments to variadic parameter if exists
	for i, arg := range m.Arguments {
		if arg.Variadic && i < len(args) {
			// Collect remaining positional args
			for j := i; j < len(args); j++ {
				variadicArgs = append(variadicArgs, args[j])
			}
			argMap[arg.Name] = variadicArgs
			break
		}
	}

	// Apply defaults for missing arguments
	for _, arg := range m.Arguments {
		if _, exists := argMap[arg.Name]; !exists {
			if arg.HasDefault && arg.Default != nil {
				// Evaluate default expression
				evaluator := NewEvaluator(ctx)
				defaultValue := evaluator.Evaluate(arg.Default.(nodes.Expr))
				if err, ok := defaultValue.(error); ok {
					return err
				}
				argMap[arg.Name] = defaultValue
			} else if !arg.Variadic && !arg.Keyword {
				return NewMacroError(m.Name,
					fmt.Sprintf("missing required argument '%s'", arg.Name),
					m.Position, nil)
			}
		}
	}

	// Set all arguments in the context
	for name, value := range argMap {
		ctx.Set(name, value)
	}

	// Check for unexpected arguments
	expectedArgs := make(map[string]bool)
	for _, arg := range m.Arguments {
		expectedArgs[arg.Name] = true
	}

	// Check for unexpected keyword arguments
	for key := range kwargs {
		if key == "__caller" {
			continue
		}
		if !expectedArgs[key] {
			return NewMacroError(m.Name,
				fmt.Sprintf("unexpected keyword argument '%s'", key),
				m.Position, nil)
		}
	}

	if m.callerFunc != nil {
		ctx.Set("caller", m.callerFunc)
	}

	return nil
}

// Call executes a macro call from within a template
func (m *Macro) Call(ctx *Context, args ...interface{}) (interface{}, error) {
	result, err := m.Execute(ctx, args, nil)
	if err != nil {
		return nil, err
	}
	if markup, ok := result.(Markup); ok {
		return string(markup), nil
	}
	return result, nil
}

// CallKwargs executes a macro call with keyword arguments
func (m *Macro) CallKwargs(ctx *Context, args []interface{}, kwargs map[string]interface{}) (interface{}, error) {
	result, err := m.Execute(ctx, args, kwargs)
	if err != nil {
		return nil, err
	}
	if markup, ok := result.(Markup); ok {
		return string(markup), nil
	}
	return result, nil
}

// GetPosition returns the position information for this macro
func (m *Macro) GetPosition() nodes.Position {
	return m.Position
}

// SetPosition sets the position information for this macro
func (m *Macro) SetPosition(pos nodes.Position) {
	m.Position = pos
}

// GetChildren returns all child nodes of this macro
func (m *Macro) GetChildren() []nodes.Node {
	var children []nodes.Node

	// Add arguments as children (they are Name nodes)
	for _, arg := range m.Arguments {
		// Convert runtime arg to nodes.Node if possible
		if nameNode := convertArgToNode(arg); nameNode != nil {
			children = append(children, nameNode)
		}
	}

	// Add defaults as children
	for _, def := range m.Defaults {
		if def != nil {
			children = append(children, def)
		}
	}

	// Add body nodes
	for _, node := range m.Body {
		if node != nil {
			children = append(children, node)
		}
	}

	return children
}

// Accept implements the visitor pattern
func (m *Macro) Accept(visitor nodes.Visitor) interface{} {
	return visitor.Visit(m)
}

// Type returns the node type
func (m *Macro) Type() string {
	return "RuntimeMacro"
}

// String returns a string representation of the macro
func (m *Macro) String() string {
	argNames := make([]string, len(m.Arguments))
	for i, arg := range m.Arguments {
		suffix := ""
		if arg.HasDefault {
			suffix = "=..."
		}
		if arg.Variadic {
			suffix = "*"
		}
		if arg.Keyword {
			suffix = "**"
		}
		argNames[i] = arg.Name + suffix
	}

	return fmt.Sprintf("Macro(%s(%s))", m.Name, strings.Join(argNames, ", "))
}

// GetArgumentNames returns the names of all arguments
func (m *Macro) GetArgumentNames() []string {
	names := make([]string, len(m.Arguments))
	for i, arg := range m.Arguments {
		names[i] = arg.Name
	}
	return names
}

// GetRequiredArgumentNames returns the names of required arguments (no defaults)
func (m *Macro) GetRequiredArgumentNames() []string {
	var names []string
	for _, arg := range m.Arguments {
		if !arg.HasDefault && !arg.Variadic && !arg.Keyword {
			names = append(names, arg.Name)
		}
	}
	return names
}

// HasArgument checks if the macro has an argument with the given name
func (m *Macro) HasArgument(name string) bool {
	for _, arg := range m.Arguments {
		if arg.Name == name {
			return true
		}
	}
	return false
}

// ValidateCall validates a macro call with the given arguments
func (m *Macro) ValidateCall(args []interface{}, kwargs map[string]interface{}) error {
	// Check for too many positional arguments
	positionalCount := 0
	for _, arg := range m.Arguments {
		if !arg.Keyword && !arg.Variadic {
			positionalCount++
		}
	}

	if len(args) > positionalCount {
		// Check if there's a variadic argument
		hasVariadic := false
		for _, arg := range m.Arguments {
			if arg.Variadic {
				hasVariadic = true
				break
			}
		}
		if !hasVariadic {
			return NewMacroError(m.Name,
				fmt.Sprintf("too many positional arguments (got %d, expected at most %d)",
					len(args), positionalCount), m.Position, nil)
		}
	}

	// Check for unexpected keyword arguments
	expectedArgs := make(map[string]bool)
	for _, arg := range m.Arguments {
		expectedArgs[arg.Name] = true
	}

	for key := range kwargs {
		if !expectedArgs[key] {
			return NewMacroError(m.Name,
				fmt.Sprintf("unexpected keyword argument '%s'", key),
				m.Position, nil)
		}
	}

	// Check for missing required arguments
	// Build a map of provided arguments
	providedArgs := make(map[string]bool)

	// Add positional arguments
	for i := 0; i < len(args) && i < len(m.Arguments); i++ {
		if !m.Arguments[i].Keyword && !m.Arguments[i].Variadic {
			providedArgs[m.Arguments[i].Name] = true
		}
	}

	// Add keyword arguments
	for key := range kwargs {
		providedArgs[key] = true
	}

	// Check each argument to see if required ones are provided
	for _, arg := range m.Arguments {
		if !arg.HasDefault && !arg.Variadic && !arg.Keyword {
			if !providedArgs[arg.Name] {
				return NewMacroError(m.Name,
					fmt.Sprintf("missing required argument '%s'", arg.Name),
					m.Position, nil)
			}
		}
	}

	return nil
}

// Namespace methods

// AddMacro adds a macro to the namespace
func (ns *MacroNamespace) AddMacro(name string, macro *Macro) {
	ns.mu.Lock()
	defer ns.mu.Unlock()
	ns.Macros[name] = macro
}

// GetMacro gets a macro from the namespace
func (ns *MacroNamespace) GetMacro(name string) (*Macro, error) {
	ns.mu.RLock()
	defer ns.mu.RUnlock()

	macro, exists := ns.Macros[name]
	if !exists {
		return nil, NewMacroError(fmt.Sprintf("%s.%s", ns.Name, name),
			fmt.Sprintf("macro '%s' not found in namespace '%s'", name, ns.Name),
			nodes.Position{}, nil)
	}

	return macro, nil
}

// HasMacro checks if the namespace contains a macro
func (ns *MacroNamespace) HasMacro(name string) bool {
	ns.mu.RLock()
	defer ns.mu.RUnlock()
	_, exists := ns.Macros[name]
	return exists
}

// GetMacroNames returns all macro names in the namespace
func (ns *MacroNamespace) GetMacroNames() []string {
	ns.mu.RLock()
	defer ns.mu.RUnlock()

	names := make([]string, 0, len(ns.Macros))
	for name := range ns.Macros {
		names = append(names, name)
	}
	return names
}

// String returns a string representation of the namespace
func (ns *MacroNamespace) String() string {
	ns.mu.RLock()
	defer ns.mu.RUnlock()

	macroNames := make([]string, 0, len(ns.Macros))
	for name := range ns.Macros {
		macroNames = append(macroNames, name)
	}

	return fmt.Sprintf("MacroNamespace(%s: [%s])", ns.Name, strings.Join(macroNames, ", "))
}

// Registry methods

// GetGlobalMacros returns all global macros
func (r *MacroRegistry) GetGlobalMacros() map[string]*Macro {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]*Macro)
	for name, macro := range r.globals {
		result[name] = macro
	}
	return result
}

// GetTemplateMacros returns all macros for a specific template
func (r *MacroRegistry) GetTemplateMacros(templateName string) map[string]*Macro {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if templateMacros, exists := r.templates[templateName]; exists {
		result := make(map[string]*Macro)
		for name, macro := range templateMacros {
			result[name] = macro
		}
		return result
	}
	return make(map[string]*Macro)
}

// Clear clears the registry
func (r *MacroRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.globals = make(map[string]*Macro)
	r.templates = make(map[string]map[string]*Macro)
	r.namespaces = make(map[string]*MacroNamespace)
}

// Stats returns registry statistics
func (r *MacroRegistry) Stats() map[string]int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats := map[string]int{
		"globals":    len(r.globals),
		"templates":  len(r.templates),
		"namespaces": len(r.namespaces),
	}

	totalTemplateMacros := 0
	for _, templateMacros := range r.templates {
		totalTemplateMacros += len(templateMacros)
	}
	stats["template_macros"] = totalTemplateMacros

	totalNamespaceMacros := 0
	for _, namespace := range r.namespaces {
		totalNamespaceMacros += len(namespace.Macros)
	}
	stats["namespace_macros"] = totalNamespaceMacros

	return stats
}

// Helper functions

// convertArgToNode converts a runtime MacroArgument to a nodes.Node
func convertArgToNode(arg *MacroArgument) nodes.Node {
	if arg == nil {
		return nil
	}

	// Create a Name node for the argument
	nameNode := &nodes.Name{
		Name: arg.Name,
		Ctx:  nodes.CtxParam, // Arguments are parameters
	}
	nameNode.SetPosition(nodes.Position{}) // Default position

	return nameNode
}

// Utility functions

// isMacroCallable checks if a value is a macro or macro-like callable
func isMacroCallable(value interface{}) bool {
	if value == nil {
		return false
	}

	switch value.(type) {
	case *Macro:
		return true
	case func(...interface{}) (interface{}, error):
		return true
	case func(*Context, ...interface{}) (interface{}, error):
		return true
	case func(...interface{}) interface{}:
		return true
	default:
		// Check if it's a function with the right signature
		fnType := reflect.TypeOf(value)
		if fnType.Kind() == reflect.Func {
			// Check if it can be called with variable arguments
			if fnType.NumOut() > 0 {
				return true
			}
		}
		return false
	}
}

// callMacroCallable calls a macro-like value with the given arguments
func callMacroCallable(ctx *Context, value interface{}, args []interface{}, kwargs map[string]interface{}) (interface{}, error) {
	if fn, ok := value.(func(...interface{}) (interface{}, error)); ok {
		return fn(args...)
	}

	if fn, ok := value.(func(*Context, ...interface{}) (interface{}, error)); ok {
		return fn(ctx, args...)
	}

	if fn, ok := value.(func(...interface{}) interface{}); ok {
		return fn(args...), nil
	}

	if macro, ok := value.(*Macro); ok {
		return macro.Execute(ctx, args, kwargs)
	}

	// Try to call using reflection
	fnValue := reflect.ValueOf(value)
	if fnValue.Kind() == reflect.Func {
		fnType := fnValue.Type()

		// Convert arguments to reflect values
		reflectArgs := make([]reflect.Value, fnType.NumIn())

		// Simple conversion - this is a basic implementation
		for i := 0; i < len(args) && i < len(reflectArgs); i++ {
			reflectArgs[i] = reflect.ValueOf(args[i])
		}

		results := fnValue.Call(reflectArgs)
		if len(results) > 0 {
			return results[0].Interface(), nil
		}
		return nil, nil
	}

	return nil, NewMacroError("", "value is not callable", nodes.Position{}, nil)
}
