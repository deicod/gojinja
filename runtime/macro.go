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
	positionalArgs := make([]*MacroArgument, len(macroNode.Args))
	positionalCount := len(macroNode.Args)
	defaultsCount := len(macroNode.Defaults)

	for i, arg := range macroNode.Args {
		defaultIndex := i - (positionalCount - defaultsCount)
		hasDefault := defaultIndex >= 0 && defaultIndex < defaultsCount && macroNode.Defaults[defaultIndex] != nil

		positionalArgs[i] = &MacroArgument{
			Name:       arg.Name,
			HasDefault: hasDefault,
		}

		if hasDefault {
			positionalArgs[i].Default = macroNode.Defaults[defaultIndex]
		}
	}

	args := make([]*MacroArgument, 0, len(positionalArgs)+2)
	args = append(args, positionalArgs...)

	if macroNode.VarArg != nil {
		args = append(args, &MacroArgument{
			Name:     macroNode.VarArg.Name,
			Variadic: true,
		})
	}

	if macroNode.KwArg != nil {
		args = append(args, &MacroArgument{
			Name:    macroNode.KwArg.Name,
			Keyword: true,
		})
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
	if kwargs == nil {
		kwargs = map[string]interface{}{}
	}

	argValues := make(map[string]interface{})
	var positionalArgs []*MacroArgument
	var variadicArg *MacroArgument
	var keywordCollector *MacroArgument

	for _, arg := range m.Arguments {
		switch {
		case arg.Variadic:
			variadicArg = arg
		case arg.Keyword:
			keywordCollector = arg
		default:
			positionalArgs = append(positionalArgs, arg)
		}
	}

	// Bind positional arguments to positional parameters first
	for idx, param := range positionalArgs {
		if idx < len(args) {
			argValues[param.Name] = args[idx]
		}
	}

	// Handle extra positional arguments
	if len(args) > len(positionalArgs) {
		if variadicArg == nil {
			return NewMacroError(m.Name,
				fmt.Sprintf("too many positional arguments (got %d, expected at most %d)", len(args), len(positionalArgs)),
				m.Position, nil)
		}
		extras := make([]interface{}, len(args)-len(positionalArgs))
		copy(extras, args[len(positionalArgs):])
		argValues[variadicArg.Name] = extras
	} else if variadicArg != nil {
		argValues[variadicArg.Name] = []interface{}{}
	}

	// Prepare keyword collector map if present
	extraKeywords := make(map[string]interface{})

	// Bind keyword arguments
	for key, value := range kwargs {
		if key == "__caller" {
			continue
		}

		if param := m.argumentByName(key); param != nil && !param.Variadic && !param.Keyword {
			if _, exists := argValues[key]; exists {
				return NewMacroError(m.Name,
					fmt.Sprintf("multiple values for argument '%s'", key),
					m.Position, nil)
			}
			argValues[key] = value
			continue
		}

		if keywordCollector != nil {
			extraKeywords[key] = value
			continue
		}

		return NewMacroError(m.Name,
			fmt.Sprintf("unexpected keyword argument '%s'", key),
			m.Position, nil)
	}

	if keywordCollector != nil {
		argValues[keywordCollector.Name] = extraKeywords
	}

	// Apply defaults for missing positional parameters
	for _, param := range positionalArgs {
		if _, exists := argValues[param.Name]; exists {
			continue
		}

		if param.HasDefault && param.Default != nil {
			evaluator := NewEvaluator(ctx)
			defaultValue := evaluator.Evaluate(param.Default.(nodes.Expr))
			if err, ok := defaultValue.(error); ok {
				return err
			}
			argValues[param.Name] = defaultValue
			continue
		}

		return NewMacroError(m.Name,
			fmt.Sprintf("missing required argument '%s'", param.Name),
			m.Position, nil)
	}

	for name, value := range argValues {
		ctx.Set(name, value)
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

// argumentByName returns the macro argument with the given name, if present
func (m *Macro) argumentByName(name string) *MacroArgument {
	for _, arg := range m.Arguments {
		if arg.Name == name {
			return arg
		}
	}
	return nil
}

// ValidateCall validates a macro call with the given arguments
func (m *Macro) ValidateCall(args []interface{}, kwargs map[string]interface{}) error {
	if kwargs == nil {
		kwargs = map[string]interface{}{}
	}

	var positionalArgs []*MacroArgument
	var variadicArg *MacroArgument
	var keywordCollector *MacroArgument

	for _, arg := range m.Arguments {
		switch {
		case arg.Variadic:
			variadicArg = arg
		case arg.Keyword:
			keywordCollector = arg
		default:
			positionalArgs = append(positionalArgs, arg)
		}
	}

	if len(args) > len(positionalArgs) && variadicArg == nil {
		return NewMacroError(m.Name,
			fmt.Sprintf("too many positional arguments (got %d, expected at most %d)", len(args), len(positionalArgs)),
			m.Position, nil)
	}

	provided := make(map[string]bool)

	for idx, param := range positionalArgs {
		if idx < len(args) {
			provided[param.Name] = true
		}
	}

	for key := range kwargs {
		if key == "__caller" {
			continue
		}

		if param := m.argumentByName(key); param != nil && !param.Variadic && !param.Keyword {
			if provided[key] {
				return NewMacroError(m.Name,
					fmt.Sprintf("multiple values for argument '%s'", key),
					m.Position, nil)
			}
			provided[key] = true
			continue
		}

		if keywordCollector != nil {
			continue
		}

		return NewMacroError(m.Name,
			fmt.Sprintf("unexpected keyword argument '%s'", key),
			m.Position, nil)
	}

	for _, param := range positionalArgs {
		if param.HasDefault {
			continue
		}
		if !provided[param.Name] {
			return NewMacroError(m.Name,
				fmt.Sprintf("missing required argument '%s'", param.Name),
				m.Position, nil)
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
