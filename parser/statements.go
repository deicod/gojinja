package parser

import (
	"fmt"

	"github.com/deicod/gojinja/lexer"
	"github.com/deicod/gojinja/nodes"
)

// ParseMacro parses a macro definition
func (p *Parser) ParseMacro() (nodes.Node, error) {
	lineno := p.stream.Next().Line // consume 'macro'

	// Parse macro name
	nameToken, err := p.Expect(lexer.TokenName)
	if err != nil {
		return nil, err
	}

	macro := &nodesMacro{
		Name: nameToken.Value,
	}

	// Parse signature
	if err := p.parseSignature(macro); err != nil {
		return nil, err
	}

	// Parse macro body
	body, err := p.ParseStatements([]string{"name:endmacro"}, true)
	if err != nil {
		return nil, err
	}

	macro.Body = body
	macro.SetPosition(nodes.NewPosition(lineno, 0))
	return macro, nil
}

// ParseCallBlock parses a call block
func (p *Parser) ParseCallBlock() (nodes.Node, error) {
	lineno := p.stream.Next().Line // consume 'call'

	callBlock := &nodes.CallBlock{}

	// Parse optional signature
	if p.stream.Peek().Type == lexer.TokenLeftParen {
		if err := p.parseSignature(callBlock); err != nil {
			return nil, err
		}
	}

	// Parse the call expression
	callExpr, err := p.ParseExpression()
	if err != nil {
		return nil, err
	}

	call, ok := callExpr.(*nodes.Call)
	if !ok {
		return nil, p.Fail("expected call", lineno, &TemplateSyntaxError{})
	}
	callBlock.Call = call

	// Parse call block body
	body, err := p.ParseStatements([]string{"name:endcall"}, true)
	if err != nil {
		return nil, err
	}

	callBlock.Body = body
	callBlock.SetPosition(nodes.NewPosition(lineno, 0))
	return callBlock, nil
}

// ParseFilterBlock parses a filter block
func (p *Parser) ParseFilterBlock() (nodes.Node, error) {
	lineno := p.stream.Next().Line // consume 'filter'

	filter, err := p.parseFilterInternal(nil, true)
	if err != nil {
		return nil, err
	}

	// Parse filter block body
	body, err := p.ParseStatements([]string{"name:endfilter"}, true)
	if err != nil {
		return nil, err
	}

	filterBlock := &nodes.FilterBlock{
		Body:   body,
		Filter: filter.(*nodes.Filter),
	}
	filterBlock.SetPosition(nodes.NewPosition(lineno, 0))
	return filterBlock, nil
}

// ParseBreak parses a break statement
func (p *Parser) ParseBreak() (nodes.Node, error) {
	lineno := p.stream.Next().Line // consume 'break'
	brk := &nodes.Break{}
	brk.SetPosition(nodes.NewPosition(lineno, 0))
	return brk, nil
}

// ParseContinue parses a continue statement
func (p *Parser) ParseContinue() (nodes.Node, error) {
	lineno := p.stream.Next().Line // consume 'continue'
	cont := &nodes.Continue{}
	cont.SetPosition(nodes.NewPosition(lineno, 0))
	return cont, nil
}

// ParseDo parses a do statement
func (p *Parser) ParseDo() (nodes.Node, error) {
	lineno := p.stream.Next().Line // consume 'do'

	expr, err := p.ParseTuple()
	if err != nil {
		return nil, err
	}

	doNode := &nodes.Do{Expr: expr}
	doNode.SetPosition(nodes.NewPosition(lineno, 0))
	return doNode, nil
}

// ParseInclude parses an include statement
func (p *Parser) ParseInclude() (nodes.Node, error) {
	lineno := p.stream.Next().Line // consume 'include'

	template, err := p.ParseExpression()
	if err != nil {
		return nil, err
	}

	include := &nodes.Include{
		Template: template,
	}

	// Handle "ignore missing"
	if p.stream.Peek().Value == "ignore" && p.Look().Value == "missing" {
		include.IgnoreMissing = true
		p.Skip(2) // consume 'ignore' and 'missing'
	}

	// Handle context options
	if err := p.parseImportContext(include, true); err != nil {
		return nil, err
	}

	include.SetPosition(nodes.NewPosition(lineno, 0))
	return include, nil
}

// ParseImport parses an import statement
func (p *Parser) ParseImport() (nodes.Node, error) {
	lineno := p.stream.Next().Line // consume 'import'

	template, err := p.ParseExpression()
	if err != nil {
		return nil, err
	}

	if _, err := p.ExpectByName("as"); err != nil {
		return nil, err
	}

	targetToken, err := p.Expect(lexer.TokenName)
	if err != nil {
		return nil, err
	}

	importStmt := &nodes.Import{
		Template: template,
		Target:   targetToken.Value,
	}

	// Handle context options
	if err := p.parseImportContext(importStmt, false); err != nil {
		return nil, err
	}

	importStmt.SetPosition(nodes.NewPosition(lineno, 0))
	return importStmt, nil
}

// ParseFrom parses a from import statement
func (p *Parser) ParseFrom() (nodes.Node, error) {
	lineno := p.stream.Next().Line // consume 'from'

	template, err := p.ParseExpression()
	if err != nil {
		return nil, err
	}

	if _, err := p.ExpectByName("import"); err != nil {
		return nil, err
	}

	fromImport := &nodes.FromImport{
		Template: template,
	}

	// Parse import names
	for {
		if len(fromImport.Names) > 0 {
			if p.stream.Peek().Type != lexer.TokenComma {
				break
			}
			p.stream.Next() // consume comma
		}

		// Check for context options before parsing name
		if p.stream.Peek().Type == lexer.TokenName {
			if p.stream.Peek().Value == "with" || p.stream.Peek().Value == "without" {
				if p.Look().Value == "context" {
					fromImport.WithContext = p.stream.Peek().Value == "with"
					p.Skip(2) // consume 'with'/'without' and 'context'
					break
				}
			}
		}

		// Parse name
		if p.stream.Peek().Type != lexer.TokenName {
			break
		}

		target, err := p.ParseAssignTargetWithExtraRules(false, false, nil, false)
		if err != nil {
			return nil, err
		}

		name, ok := target.(*nodes.Name)
		if !ok {
			return nil, p.Fail("expected name in import list", target.GetPosition().Line, &TemplateSyntaxError{})
		}

		// Check for names starting with underscore
		if len(name.Name) > 0 && name.Name[0] == '_' {
			return nil, p.Fail("names starting with an underline can not be imported", name.GetPosition().Line, &TemplateAssertionError{})
		}

		if p.SkipIfByName("as") {
			aliasTarget, err := p.ParseAssignTargetWithExtraRules(false, false, nil, false)
			if err != nil {
				return nil, err
			}
			alias, ok := aliasTarget.(*nodes.Name)
			if !ok {
				return nil, p.Fail("expected name after 'as'", aliasTarget.GetPosition().Line, &TemplateSyntaxError{})
			}
			fromImport.Names = append(fromImport.Names, nodes.ImportName{
				Name:  name.Name,
				Alias: alias.Name,
			})
		} else {
			fromImport.Names = append(fromImport.Names, nodes.ImportName{Name: name.Name})
		}
	}

	fromImport.SetPosition(nodes.NewPosition(lineno, 0))
	return fromImport, nil
}

// parseImportContext parses context options for import/include statements
func (p *Parser) parseImportContext(node interface{}, defaultValue bool) error {
	var withContext *bool

	// Handle "with context" / "without context"
	if p.stream.Peek().Value == "with" || p.stream.Peek().Value == "without" {
		if p.Look().Value == "context" {
			wc := p.stream.Peek().Value == "with"
			withContext = &wc
			p.Skip(2) // consume 'with'/'without' and 'context'
		}
	}

	// Set the context value
	switch n := node.(type) {
	case *nodes.Include:
		if withContext != nil {
			n.WithContext = *withContext
		} else {
			n.WithContext = defaultValue
		}
	case *nodes.Import:
		if withContext != nil {
			n.WithContext = *withContext
		} else {
			n.WithContext = defaultValue
		}
	default:
		return fmt.Errorf("unsupported node type for import context")
	}

	return nil
}

// parseSignature parses function/macro signatures
func (p *Parser) parseSignature(node interface{}) error {
	if _, err := p.Expect(lexer.TokenLeftParen); err != nil {
		return err
	}

	var args []*nodes.Name
	var defaults []nodes.Expr

	for p.stream.Peek().Type != lexer.TokenRightParen {
		if len(args) > 0 {
			if _, err := p.Expect(lexer.TokenComma); err != nil {
				return err
			}
			// Check for trailing comma
			if p.stream.Peek().Type == lexer.TokenRightParen {
				break
			}
		}

		arg, err := p.ParseAssignTargetWithExtraRules(false, false, nil, false)
		if err != nil {
			return err
		}

		if name, ok := arg.(*nodes.Name); ok {
			name.Ctx = nodes.CtxParam
			args = append(args, name)

			if p.SkipIf(lexer.TokenAssign) {
				defaultExpr, err := p.ParseExpression()
				if err != nil {
					return err
				}
				defaults = append(defaults, defaultExpr)
			} else if len(defaults) > 0 {
				return p.Fail("non-default argument follows default argument", name.GetPosition().Line, &TemplateSyntaxError{})
			}
		} else {
			return p.Fail("expected name in argument list", arg.GetPosition().Line, &TemplateSyntaxError{})
		}
	}

	if _, err := p.Expect(lexer.TokenRightParen); err != nil {
		return err
	}

	// Set the signature on the node
	switch n := node.(type) {
	case *nodes.Macro:
		n.Args = args
		n.Defaults = defaults
	case *nodes.CallBlock:
		n.Args = args
		n.Defaults = defaults
	default:
		return fmt.Errorf("unsupported node type for signature")
	}

	return nil
}

// ParseAssignTarget parses assignment targets
func (p *Parser) ParseAssignTarget() (nodes.Expr, error) {
	return p.parseAssignTarget(true, false, nil, false)
}

func (p *Parser) ParseAssignTargetWithTuple(withTuple bool, nameOnly bool) (nodes.Expr, error) {
	return p.parseAssignTarget(withTuple, nameOnly, nil, false)
}

func (p *Parser) ParseAssignTargetWithNamespace(withNamespace bool) (nodes.Expr, error) {
	return p.parseAssignTarget(true, false, nil, withNamespace)
}

// Type aliases for compatibility with existing nodes
type nodesMacro = nodes.Macro
type nodesCallBlock = nodes.CallBlock
type nodesFilterBlock = nodes.FilterBlock
type nodesInclude = nodes.Include
type nodesImport = nodes.Import
type nodesFromImport = nodes.FromImport
