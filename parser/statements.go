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

// ParseSpaceless parses a spaceless block
func (p *Parser) ParseSpaceless() (nodes.Node, error) {
	lineno := p.stream.Next().Line // consume 'spaceless'

	body, err := p.ParseStatements([]string{"name:endspaceless"}, true)
	if err != nil {
		return nil, err
	}

	node := &nodes.Spaceless{Body: body}
	node.SetPosition(nodes.NewPosition(lineno, 0))
	return node, nil
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

// ParseNamespace parses a namespace declaration block
func (p *Parser) ParseNamespace() (nodes.Node, error) {
	lineno := p.stream.Next().Line // consume 'namespace'

	nameToken, err := p.Expect(lexer.TokenName)
	if err != nil {
		return nil, err
	}

	namespace := &nodes.Namespace{
		Name: nameToken.Value,
	}

	if p.SkipIf(lexer.TokenAssign) {
		valueExpr, err := p.ParseExpression()
		if err != nil {
			return nil, err
		}
		namespace.Value = valueExpr
	}

	body, err := p.ParseStatements([]string{"name:endnamespace"}, true)
	if err != nil {
		return nil, err
	}

	namespace.Body = body
	namespace.SetPosition(nodes.NewPosition(lineno, 0))
	return namespace, nil
}

// ParseExport parses an export statement
func (p *Parser) ParseExport() (nodes.Node, error) {
	lineno := p.stream.Next().Line // consume 'export'

	export := &nodes.Export{}

	// At least one name is required
	if p.stream.Peek().Type != lexer.TokenName {
		return nil, p.Fail("expected at least one name to export", lineno, &TemplateSyntaxError{})
	}

	for {
		nameToken, err := p.Expect(lexer.TokenName)
		if err != nil {
			return nil, err
		}

		name := &nodes.Name{Name: nameToken.Value, Ctx: nodes.CtxStore}
		name.SetPosition(nodes.NewPosition(nameToken.Line, nameToken.Column))
		export.Names = append(export.Names, name)

		if !p.SkipIf(lexer.TokenComma) {
			break
		}

		if p.stream.Peek().Type != lexer.TokenName {
			return nil, p.Fail("expected name after ',' in export statement", p.stream.Peek().Line, &TemplateSyntaxError{})
		}
	}

	export.SetPosition(nodes.NewPosition(lineno, 0))
	return export, nil
}

// ParseTrans parses both trans and blocktrans statements.
func (p *Parser) ParseTrans(_ bool) (nodes.Node, error) {
	token := p.stream.Next()
	lineno := token.Line

	trans := &nodes.Trans{
		Variables: make(map[string]nodes.Expr),
	}

	// Optional context string: {% trans "context" %}
	if p.stream.Peek().Type == lexer.TokenString {
		ctxToken := p.stream.Next()
		trans.Context = ctxToken.Value
		trans.HasContext = true
	}

	// Parse optional variable assignments (including count expressions)
	for {
		current := p.stream.Peek()
		if current.Type == lexer.TokenBlockEnd {
			break
		}

		if current.Type == lexer.TokenColon {
			p.stream.Next()
			break
		}

		if current.Type == lexer.TokenComma {
			p.stream.Next()
			continue
		}

		if current.Type != lexer.TokenName {
			return nil, p.Fail("expected name in trans tag", current.Line, &TemplateSyntaxError{})
		}

		nameToken := p.stream.Next()
		name := nameToken.Value

		switch name {
		case "trimmed", "notrimmed":
			if trans.TrimmedSet {
				return nil, p.Fail("trimmed or notrimmed specified multiple times", nameToken.Line, &TemplateSyntaxError{})
			}
			trans.TrimmedSet = true
			trans.Trimmed = name == "trimmed"
			continue
		case "count":
			if trans.CountExpr != nil {
				return nil, p.Fail("count expression already defined", nameToken.Line, &TemplateSyntaxError{})
			}

			alias := "count"
			var expr nodes.Expr

			if p.SkipIf(lexer.TokenAssign) {
				parsed, err := p.ParseExpression()
				if err != nil {
					return nil, err
				}
				expr = parsed
			} else {
				parsed, err := p.ParseExpression()
				if err != nil {
					return nil, err
				}
				expr = parsed

				if p.SkipIfByName("as") {
					aliasToken, err := p.Expect(lexer.TokenName)
					if err != nil {
						return nil, err
					}
					alias = aliasToken.Value
				}
			}

			trans.CountExpr = expr
			trans.CountName = alias

			if alias != "" && alias != "count" {
				if _, exists := trans.Variables[alias]; !exists {
					trans.Variables[alias] = expr
				}
			}
			continue
		}

		var valueExpr nodes.Expr
		if p.SkipIf(lexer.TokenAssign) {
			parsed, err := p.ParseExpression()
			if err != nil {
				return nil, err
			}
			valueExpr = parsed
		} else {
			nameExpr := &nodes.Name{Name: name, Ctx: nodes.CtxLoad}
			nameExpr.SetPosition(nodes.NewPosition(nameToken.Line, 0))
			valueExpr = nameExpr
		}

		if _, exists := trans.Variables[name]; exists {
			return nil, p.Fail(fmt.Sprintf("trans variable %q defined multiple times", name), nameToken.Line, &TemplateAssertionError{})
		}

		trans.Variables[name] = valueExpr
	}

	if trans.CountName == "" && trans.CountExpr != nil {
		trans.CountName = "count"
	}

	singular, err := p.ParseStatements([]string{"name:pluralize", "name:endtrans"}, false)
	if err != nil {
		return nil, err
	}
	trans.Singular = singular

	next := p.stream.Peek()
	if next.Type != lexer.TokenName {
		return nil, p.Fail("expected 'pluralize' or 'endtrans'", next.Line, &TemplateSyntaxError{})
	}

	if next.Value == "pluralize" {
		p.stream.Next()

		if p.stream.Peek().Type == lexer.TokenName {
			aliasToken, err := p.Expect(lexer.TokenName)
			if err != nil {
				return nil, err
			}
			alias := aliasToken.Value

			if expr, ok := trans.Variables[alias]; ok {
				trans.CountExpr = expr
				trans.CountName = alias
			} else if trans.CountExpr != nil && (alias == trans.CountName || (alias == "count" && trans.CountName == "")) {
				trans.CountName = alias
			} else {
				nameExpr := &nodes.Name{Name: alias, Ctx: nodes.CtxLoad}
				nameExpr.SetPosition(nodes.NewPosition(aliasToken.Line, 0))
				trans.CountExpr = nameExpr
				trans.CountName = alias
			}
		}

		if trans.CountExpr == nil {
			return nil, p.Fail("pluralize is only allowed if a count is provided", next.Line, &TemplateSyntaxError{})
		}

		plural, err := p.ParseStatements([]string{"name:endtrans"}, false)
		if err != nil {
			return nil, err
		}
		trans.Plural = plural
		next = p.stream.Peek()
	}

	if next.Type != lexer.TokenName || next.Value != "endtrans" {
		return nil, p.Fail("expected 'endtrans'", next.Line, &TemplateSyntaxError{})
	}

	p.stream.Next()

	trans.SetPosition(nodes.NewPosition(lineno, 0))
	return trans, nil
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

	var (
		args        []*nodes.Name
		defaults    []nodes.Expr
		kwonlyArgs  []*nodes.Name
		kwDefaults  map[string]nodes.Expr
		varArg      *nodes.Name
		kwArg       *nodes.Name
		seenDefault bool
		expectComma bool
		seenStar    bool
		allowKwOnly bool
	)

	for p.stream.Peek().Type != lexer.TokenRightParen {
		if expectComma {
			if _, err := p.Expect(lexer.TokenComma); err != nil {
				return err
			}
			if p.stream.Peek().Type == lexer.TokenRightParen {
				break
			}
		}

		switch p.stream.Peek().Type {
		case lexer.TokenMul:
			p.stream.Next()

			if p.stream.Peek().Type == lexer.TokenMul {
				// Handle ``**kwargs`` written as ``*`` ``*`` for robustness, though lexer should emit TokenPow
				if kwArg != nil {
					return p.Fail("duplicate keyword argument collector", p.stream.Peek().Line, &TemplateSyntaxError{})
				}
				p.stream.Next()
				nameToken, err := p.Expect(lexer.TokenName)
				if err != nil {
					return err
				}
				kwArg = &nodes.Name{Name: nameToken.Value, Ctx: nodes.CtxParam}
				expectComma = true
				continue
			}

			if seenStar {
				return p.Fail("duplicate '*' marker in signature", p.stream.Peek().Line, &TemplateSyntaxError{})
			}
			seenStar = true
			allowKwOnly = true

			if varArg != nil {
				return p.Fail("duplicate varargs parameter", p.stream.Peek().Line, &TemplateSyntaxError{})
			}

			if p.stream.Peek().Type == lexer.TokenComma || p.stream.Peek().Type == lexer.TokenRightParen {
				// Bare ``*`` indicates keyword-only arguments follow
				expectComma = true
				continue
			}

			nameToken, err := p.Expect(lexer.TokenName)
			if err != nil {
				return err
			}
			if nameToken.Value == "" {
				return p.Fail("expected name after '*' in signature", nameToken.Line, &TemplateSyntaxError{})
			}

			name := &nodes.Name{Name: nameToken.Value, Ctx: nodes.CtxParam}
			varArg = name
			expectComma = true

		case lexer.TokenPow:
			p.stream.Next()
			if kwArg != nil {
				return p.Fail("duplicate keyword argument collector", p.stream.Peek().Line, &TemplateSyntaxError{})
			}

			nameToken, err := p.Expect(lexer.TokenName)
			if err != nil {
				return err
			}
			kwArg = &nodes.Name{Name: nameToken.Value, Ctx: nodes.CtxParam}
			expectComma = true

		default:
			arg, err := p.ParseAssignTargetWithExtraRules(false, false, nil, false)
			if err != nil {
				return err
			}

			name, ok := arg.(*nodes.Name)
			if !ok {
				return p.Fail("expected name in argument list", arg.GetPosition().Line, &TemplateSyntaxError{})
			}

			name.Ctx = nodes.CtxParam

			if allowKwOnly {
				kwonlyArgs = append(kwonlyArgs, name)
				if p.SkipIf(lexer.TokenAssign) {
					defaultExpr, err := p.ParseExpression()
					if err != nil {
						return err
					}
					if kwDefaults == nil {
						kwDefaults = make(map[string]nodes.Expr)
					}
					kwDefaults[name.Name] = defaultExpr
				}
				expectComma = true
				continue
			}

			args = append(args, name)

			if p.SkipIf(lexer.TokenAssign) {
				defaultExpr, err := p.ParseExpression()
				if err != nil {
					return err
				}
				defaults = append(defaults, defaultExpr)
				seenDefault = true
			} else {
				if seenDefault {
					return p.Fail("non-default argument follows default argument", name.GetPosition().Line, &TemplateSyntaxError{})
				}
			}

			expectComma = true
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
		n.KwonlyArgs = kwonlyArgs
		n.KwDefaults = kwDefaults
		n.VarArg = varArg
		n.KwArg = kwArg
	case *nodes.CallBlock:
		n.Args = args
		n.Defaults = defaults
		n.KwonlyArgs = kwonlyArgs
		n.KwDefaults = kwDefaults
		n.VarArg = varArg
		n.KwArg = kwArg
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
