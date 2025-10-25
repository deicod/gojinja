package runtime

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/deicod/gojinja/nodes"
)

// Template represents a compiled template ready for rendering
type Template struct {
	name           string
	environment    *Environment
	ast            *nodes.Template
	autoescape     bool
	blocks         map[string]*nodes.Block
	macros         map[string]*nodes.Macro
	imports        map[string]*Template
	inheritanceCtx *InheritanceContext
	macroRegistry  *MacroRegistry
}

// NewTemplate creates a new template from an AST
func NewTemplate(env *Environment, ast *nodes.Template, name string) (*Template, error) {
	if env == nil {
		return nil, NewError(ErrorTypeTemplate, "environment cannot be nil", nodes.Position{}, nil)
	}
	if ast == nil {
		return nil, NewError(ErrorTypeTemplate, "AST cannot be nil", nodes.Position{}, nil)
	}

	template := &Template{
		name:          name,
		environment:   env,
		ast:           ast,
		autoescape:    env.shouldAutoescape(name),
		blocks:        make(map[string]*nodes.Block),
		macros:        make(map[string]*nodes.Macro),
		imports:       make(map[string]*Template),
		macroRegistry: env.macroRegistry,
	}

	// Pre-process the template to collect blocks and macros
	if err := template.preprocess(); err != nil {
		return nil, fmt.Errorf("failed to preprocess template: %w", err)
	}

	// Set up inheritance if this template extends another
	if err := template.setupInheritance(); err != nil {
		return nil, fmt.Errorf("failed to setup inheritance: %w", err)
	}

	return template, nil
}

// preprocess analyzes the AST to collect blocks and macros
func (t *Template) preprocess() error {
	visitor := nodes.NodeVisitorFunc(func(node nodes.Node) interface{} {
		switch n := node.(type) {
		case *nodes.Block:
			t.blocks[n.Name] = n
		case *nodes.Macro:
			t.macros[n.Name] = n
		}
		return nil
	})

	nodes.Walk(visitor, t.ast)
	return nil
}

// setupInheritance sets up template inheritance context
func (t *Template) setupInheritance() error {
	// Create inheritance resolver
	resolver := NewInheritanceResolver(t.environment)
	inheritanceCtx, err := resolver.ResolveInheritance(t)
	if err != nil {
		return err
	}

	t.inheritanceCtx = inheritanceCtx

	// Create a super function bound to this template
	var superFunc GlobalFunc = func(ctx *Context, args ...interface{}) (interface{}, error) {

		if !inheritanceCtx.CanSuper() {
			return "", NewError(ErrorTypeTemplate, "super() can only be called within a block that has a parent block", nodes.Position{}, nil)
		}

		if len(args) > 1 {
			return "", NewError(ErrorTypeTemplate, "super() takes at most one argument (block name)", nodes.Position{}, nil)
		}

		var blockName string
		if len(args) == 1 {
			// Get block name from argument
			if name, ok := args[0].(string); ok {
				blockName = name
			} else {
				return "", NewError(ErrorTypeTemplate, "super() argument must be a string", nodes.Position{}, nil)
			}
		} else {
			// Use current block
			blockName = inheritanceCtx.CurrentBlock
		}

		// Get parent block
		parentBlock := inheritanceCtx.ParentBlocks[blockName]
		if parentBlock == nil {
			return "", NewError(ErrorTypeTemplate, fmt.Sprintf("no parent block found for '%s'", blockName), nodes.Position{}, nil)
		}

		// Execute parent block without autoescaping
		// Temporarily disable autoescaping for super() execution
		oldAutoescape := ctx.ShouldAutoescape()
		ctx.SetAutoescape(false)
		defer func() { ctx.SetAutoescape(oldAutoescape) }()

		var buf strings.Builder
		oldWriter := ctx.writer
		ctx.writer = &buf
		defer func() { ctx.writer = oldWriter }()

		// Save current context
		oldCurrent := ctx.current
		defer func() { ctx.current = oldCurrent }()

		// Create evaluator and execute block
		evaluator := NewEvaluator(ctx)
		result := evaluator.Evaluate(parentBlock)
		if err, ok := result.(error); ok {
			return "", err
		}

		return Markup(buf.String()), nil
	}

	// Add super function to environment
	t.environment.AddGlobal("super", superFunc)

	return nil
}

// Execute renders the template to the given writer with the provided context
func (t *Template) Execute(vars map[string]interface{}, writer io.Writer) error {
	if writer == nil {
		return NewError(ErrorTypeTemplate, "writer cannot be nil", nodes.Position{}, nil)
	}

	useTrim := !t.environment.ShouldKeepTrailingNewline()
	var buffer bytes.Buffer
	outWriter := &buffer

	// Create context
	ctx := NewContextWithEnvironment(t.environment, vars)
	ctx.SetAutoescape(t.autoescape)
	ctx.current = t
	ctx.writer = outWriter

	if err := t.ExecuteWithContext(ctx); err != nil {
		return err
	}

	output := buffer.String()
	if useTrim {
		switch {
		case strings.HasSuffix(output, "\r\n"):
			output = output[:len(output)-2]
		case strings.HasSuffix(output, "\n"):
			output = output[:len(output)-1]
		}
	}
	_, err := writer.Write([]byte(output))
	return err
}

// ExecuteWithContext renders the template using an existing context
func (t *Template) ExecuteWithContext(ctx *Context) error {
	// Create evaluator - use secure evaluator if environment is sandboxed
	var evaluator *Evaluator
	if t.environment.IsSandboxed() {
		// Create security context for sandboxed execution
		secCtx, err := t.environment.GetSecurityManager().CreateSecurityContext("default", t.name)
		if err != nil {
			return fmt.Errorf("failed to create security context: %w", err)
		}
		defer t.environment.GetSecurityManager().CleanupSecurityContext(fmt.Sprintf("%s_%d", t.name, time.Now().UnixNano()))

		evaluator = NewSecureEvaluator(ctx, secCtx)
	} else {
		evaluator = NewEvaluator(ctx)
	}

	// Ensure current template is set
	if ctx.current == nil {
		ctx.current = t
	}

	// Evaluate the template
	result := evaluator.Evaluate(t.ast)
	if err, ok := result.(error); ok {
		return err
	}

	// Check for any errors that occurred during rendering
	if ctx.HasErrors() {
		return ctx.GetErrors()[0] // Return the first error
	}

	return nil
}

// ExecuteToString renders the template to a string
func (t *Template) ExecuteToString(vars map[string]interface{}) (string, error) {
	var buf bytes.Buffer
	err := t.Execute(vars, &buf)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// newModuleContext prepares a context suitable for module execution.
func (t *Template) newModuleContext(vars map[string]interface{}) *Context {
	ctx := NewContextWithEnvironment(t.environment, vars)
	ctx.SetAutoescape(t.autoescape)
	ctx.current = t

	var buf strings.Builder
	ctx.writer = &buf

	return ctx
}

// makeModuleFromContext executes the template with the provided context and produces a module namespace.
func (t *Template) makeModuleFromContext(ctx *Context) (*MacroNamespace, error) {
	if err := t.ExecuteWithContext(ctx); err != nil {
		return nil, err
	}

	module := NewMacroNamespace(t.name, t)
	module.Context = ctx

	if t.environment != nil {
		if registry := t.environment.GetMacroRegistry(); registry != nil {
			for name, macro := range registry.GetTemplateMacros(t.name) {
				module.AddMacro(name, macro)
			}
		}
	}

	for name, value := range ctx.Exports() {
		module.AddExport(name, value)
	}

	return module, nil
}

// MakeModule executes the template in module mode and returns a namespace with exported members.
func (t *Template) MakeModule(vars map[string]interface{}) (*MacroNamespace, error) {
	ctx := t.newModuleContext(vars)
	return t.makeModuleFromContext(ctx)
}

// Name returns the template name
func (t *Template) Name() string {
	return t.name
}

// Environment returns the template's environment
func (t *Template) Environment() *Environment {
	return t.environment
}

// AST returns the template's AST
func (t *Template) AST() *nodes.Template {
	return t.ast
}

// Autoescape returns whether autoescaping is enabled
func (t *Template) Autoescape() bool {
	return t.autoescape
}

// GetBlock returns a block by name
func (t *Template) GetBlock(name string) (*nodes.Block, bool) {
	block, ok := t.blocks[name]
	return block, ok
}

// GetMacro returns a macro by name
func (t *Template) GetMacro(name string) (*nodes.Macro, bool) {
	macro, ok := t.macros[name]
	return macro, ok
}

// HasBlock checks if a block exists
func (t *Template) HasBlock(name string) bool {
	_, ok := t.blocks[name]
	return ok
}

// HasMacro checks if a macro exists
func (t *Template) HasMacro(name string) bool {
	_, ok := t.macros[name]
	return ok
}

// BlockNames returns all block names
func (t *Template) BlockNames() []string {
	names := make([]string, 0, len(t.blocks))
	for name := range t.blocks {
		names = append(names, name)
	}
	return names
}

// MacroNames returns all macro names
func (t *Template) MacroNames() []string {
	names := make([]string, 0, len(t.macros))
	for name := range t.macros {
		names = append(names, name)
	}
	return names
}

// GetBlocks returns all blocks
func (t *Template) GetBlocks() map[string]*nodes.Block {
	blocks := make(map[string]*nodes.Block)
	for name, block := range t.blocks {
		blocks[name] = block
	}
	return blocks
}

// GetMacros returns all macros
func (t *Template) GetMacros() map[string]*nodes.Macro {
	macros := make(map[string]*nodes.Macro)
	for name, macro := range t.macros {
		macros[name] = macro
	}
	return macros
}

// RenderBlock renders a specific block
func (t *Template) RenderBlock(blockName string, vars map[string]interface{}, writer io.Writer) error {
	block, ok := t.blocks[blockName]
	if !ok {
		return NewError(ErrorTypeTemplate, fmt.Sprintf("block '%s' not found", blockName), nodes.Position{}, nil)
	}

	ctx := NewContextWithEnvironment(t.environment, vars)
	ctx.SetAutoescape(t.autoescape)
	ctx.current = t
	ctx.writer = writer

	evaluator := NewEvaluator(ctx)
	result := evaluator.Evaluate(block)
	if err, ok := result.(error); ok {
		return err
	}

	return nil
}

// RenderBlockToString renders a specific block to a string
func (t *Template) RenderBlockToString(blockName string, vars map[string]interface{}) (string, error) {
	var buf bytes.Buffer
	err := t.RenderBlock(blockName, vars, &buf)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// String returns a string representation of the template
func (t *Template) String() string {
	return fmt.Sprintf("Template(name=%s, autoescape=%t)", t.name, t.autoescape)
}

// InheritanceContext returns the template's inheritance context
func (t *Template) InheritanceContext() interface{} {
	if t.inheritanceCtx == nil {
		return nil
	}

	parentBlockNames := make([]string, 0, len(t.inheritanceCtx.ParentBlocks))
	for name := range t.inheritanceCtx.ParentBlocks {
		parentBlockNames = append(parentBlockNames, name)
	}

	return map[string]interface{}{
		"CurrentBlock": t.inheritanceCtx.CurrentBlock,
		"ParentBlocks": parentBlockNames,
		"CanSuper":     t.inheritanceCtx.CanSuper(),
	}
}

// Dump returns a debug representation of the template's AST
func (t *Template) Dump() string {
	return nodes.Dump(t.ast)
}

// NewTemplateFromString is a convenience function to create a template from a string
func NewTemplateFromString(templateString string) (*Template, error) {
	env := NewEnvironment()
	return env.NewTemplate(templateString)
}

// NewTemplateFromAST is a convenience function to create a template from an AST
func NewTemplateFromAST(ast *nodes.Template, name string) (*Template, error) {
	env := NewEnvironment()
	return env.NewTemplateFromAST(ast, name)
}

// TemplateList represents a collection of templates
type TemplateList struct {
	templates   map[string]*Template
	environment *Environment
}

// NewTemplateList creates a new template list
func NewTemplateList(env *Environment) *TemplateList {
	return &TemplateList{
		templates:   make(map[string]*Template),
		environment: env,
	}
}

// Add adds a template to the list
func (tl *TemplateList) Add(template *Template) {
	tl.templates[template.Name()] = template
}

// Get gets a template by name
func (tl *TemplateList) Get(name string) (*Template, bool) {
	template, ok := tl.templates[name]
	return template, ok
}

// Has checks if a template exists
func (tl *TemplateList) Has(name string) bool {
	_, ok := tl.templates[name]
	return ok
}

// Remove removes a template by name
func (tl *TemplateList) Remove(name string) {
	delete(tl.templates, name)
}

// Clear removes all templates
func (tl *TemplateList) Clear() {
	tl.templates = make(map[string]*Template)
}

// Names returns all template names
func (tl *TemplateList) Names() []string {
	names := make([]string, 0, len(tl.templates))
	for name := range tl.templates {
		names = append(names, name)
	}
	return names
}

// Size returns the number of templates
func (tl *TemplateList) Size() int {
	return len(tl.templates)
}

// All returns all templates
func (tl *TemplateList) All() map[string]*Template {
	templates := make(map[string]*Template)
	for name, template := range tl.templates {
		templates[name] = template
	}
	return templates
}

// Environment returns the environment
func (tl *TemplateList) Environment() *Environment {
	return tl.environment
}

// String returns a string representation of the template list
func (tl *TemplateList) String() string {
	var names []string
	for name := range tl.templates {
		names = append(names, name)
	}
	return fmt.Sprintf("TemplateList(templates=[%s])", strings.Join(names, ", "))
}
