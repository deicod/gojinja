package runtime

import (
	"fmt"
	"strings"

	"github.com/deicod/gojinja/nodes"
)

// InheritanceContext tracks the current inheritance chain and super() calls
type InheritanceContext struct {
	Template     *Template
	ParentBlocks map[string]*nodes.Block
	CurrentBlock string
	SuperStack   []string
}

// NewInheritanceContext creates a new inheritance context
func NewInheritanceContext(tmpl *Template) *InheritanceContext {
	return &InheritanceContext{
		Template:     tmpl,
		ParentBlocks: make(map[string]*nodes.Block),
		SuperStack:   make([]string, 0),
	}
}

// PushBlock enters a new block context
func (ic *InheritanceContext) PushBlock(blockName string) {
	ic.SuperStack = append(ic.SuperStack, ic.CurrentBlock)
	ic.CurrentBlock = blockName
}

// PopBlock exits the current block context
func (ic *InheritanceContext) PopBlock() {
	if len(ic.SuperStack) > 0 {
		ic.CurrentBlock = ic.SuperStack[len(ic.SuperStack)-1]
		ic.SuperStack = ic.SuperStack[:len(ic.SuperStack)-1]
	} else {
		ic.CurrentBlock = ""
	}
}

// GetParentBlock returns the parent block for the current block
func (ic *InheritanceContext) GetParentBlock() *nodes.Block {
	if ic.CurrentBlock == "" {
		return nil
	}
	return ic.ParentBlocks[ic.CurrentBlock]
}

// SetParentBlock sets the parent block for a given block name
func (ic *InheritanceContext) SetParentBlock(blockName string, block *nodes.Block) {
	ic.ParentBlocks[blockName] = block
}

// CanSuper returns true if super() can be called in the current context
func (ic *InheritanceContext) CanSuper() bool {
	if ic.CurrentBlock == "" {
		return false
	}
	_, exists := ic.ParentBlocks[ic.CurrentBlock]
	return exists
}

// InheritanceResolver handles template inheritance resolution
type InheritanceResolver struct {
	environment *Environment
}

// NewInheritanceResolver creates a new inheritance resolver
func NewInheritanceResolver(env *Environment) *InheritanceResolver {
	return &InheritanceResolver{
		environment: env,
	}
}

// ResolveInheritance resolves the complete inheritance chain for a template
func (ir *InheritanceResolver) ResolveInheritance(tmpl *Template) (*InheritanceContext, error) {
	context := NewInheritanceContext(tmpl)

	// Walk through the inheritance chain to understand the structure
	err := ir.resolveInheritanceChain(tmpl.AST(), context, make(map[string]bool))
	if err != nil {
		return nil, err
	}

	return context, nil
}

// resolveInheritanceChain recursively resolves the inheritance chain
func (ir *InheritanceResolver) resolveInheritanceChain(ast *nodes.Template, context *InheritanceContext, visited map[string]bool) error {
	// Find extends statement
	var extendsNode *nodes.Extends

	for _, node := range ast.Body {
		if ext, ok := node.(*nodes.Extends); ok {
			extendsNode = ext
			break
		}
	}

	// If no extends, we're done - this is just for structure detection
	if extendsNode == nil {
		return nil
	}

	// Evaluate parent template name
	parentNameValue := ir.environment.evaluateExpression(extendsNode.Template)
	if err, ok := parentNameValue.(error); ok {
		return err
	}

	parentName, ok := parentNameValue.(string)
	if !ok {
		return NewError(ErrorTypeTemplate, "extends template name must be a string", extendsNode.GetPosition(), extendsNode)
	}

	// Check for circular dependencies
	if visited[parentName] {
		return NewError(ErrorTypeTemplate, fmt.Sprintf("circular template inheritance detected: %s", parentName), nodes.Position{}, nil)
	}
	visited[parentName] = true

	// Load parent template
	parent, err := ir.environment.LoadTemplate(parentName)
	if err != nil {
		return err
	}

	// Recursively resolve parent inheritance
	err = ir.resolveInheritanceChain(parent.AST(), context, visited)
	if err != nil {
		return err
	}

	return nil
}

// CreateSuperFunction creates a super() function for the given inheritance context
func CreateSuperFunction(ctx *Context, inheritanceCtx *InheritanceContext) GlobalFunc {
	return func(ctx *Context, args ...interface{}) (interface{}, error) {
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
}

// ExtendTemplateWithInheritance extends a template to support inheritance
func ExtendTemplateWithInheritance(tmpl *Template) error {
	// Add super() function to the environment
	if tmpl.environment == nil {
		return fmt.Errorf("template has no environment")
	}

	// Create inheritance resolver
	resolver := NewInheritanceResolver(tmpl.environment)
	inheritanceCtx, err := resolver.ResolveInheritance(tmpl)
	if err != nil {
		return err
	}

	// Add super function as a global
	tmpl.environment.AddGlobal("super", CreateSuperFunction(nil, inheritanceCtx))

	return nil
}