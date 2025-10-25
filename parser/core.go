package parser

import (
	"fmt"
	"strings"

	"github.com/deicod/gojinja/lexer"
	"github.com/deicod/gojinja/nodes"
)

// ParseStatement parses a single statement
func (p *Parser) ParseStatement() (nodes.Node, error) {
	token := p.stream.Peek()
	if token.Type != lexer.TokenName {
		return nil, p.Fail("tag name expected", token.Line, &TemplateSyntaxError{})
	}

	p.tagStack = append(p.tagStack, token.Value)
	defer func() {
		if len(p.tagStack) > 0 {
			p.tagStack = p.tagStack[:len(p.tagStack)-1]
		}
	}()

	// Check for built-in statement keywords
	if statementKeywords[token.Value] {
		switch token.Value {
		case "break":
			return p.ParseBreak()
		case "continue":
			return p.ParseContinue()
		case "do":
			return p.ParseDo()
		case "spaceless":
			return p.ParseSpaceless()
		case "for":
			return p.ParseFor()
		case "if":
			return p.ParseIf()
		case "block":
			return p.ParseBlock()
		case "extends":
			return p.ParseExtends()
		case "macro":
			return p.ParseMacro()
		case "include":
			return p.ParseInclude()
		case "from":
			return p.ParseFrom()
		case "import":
			return p.ParseImport()
		case "set":
			return p.ParseSet()
		case "with":
			return p.ParseWith()
		case "namespace":
			return p.ParseNamespace()
		case "export":
			return p.ParseExport()
		case "trans":
			return p.ParseTrans(false)
		case "blocktrans":
			return p.ParseTrans(true)
		case "autoescape":
			return p.ParseAutoescape()
		case "print":
			return p.ParsePrint()
		default:
			// This shouldn't happen due to the map check above
			return nil, p.Fail(fmt.Sprintf("unhandled statement keyword: %s", token.Value), token.Line, &TemplateSyntaxError{})
		}
	}

	// Check for special constructs
	if token.Value == "call" {
		return p.ParseCallBlock()
	}
	if token.Value == "filter" {
		return p.ParseFilterBlock()
	}

	// Check for extensions
	if ext, ok := p.extensions[token.Value]; ok {
		return ext.Parse(p)
	}

	// Unknown tag
	return nil, p.FailUnknownTag(token.Value, token.Line)
}

// ParseStatements parses multiple statements until one of the end tokens is reached
func (p *Parser) ParseStatements(endTokens []string, dropNeedle bool) ([]nodes.Node, error) {
	// Skip optional colon (Python compatibility)
	p.SkipIf(lexer.TokenColon)

	// Expect block end
	if _, err := p.Expect(lexer.TokenBlockEnd); err != nil {
		return nil, err
	}

	result, err := p.Subparse(endTokens)
	if err != nil {
		return nil, err
	}

	// Check if we reached EOF too early
	if p.stream.Peek().Type == lexer.TokenEOF {
		return nil, p.FailEOF(endTokens, 0)
	}

	if dropNeedle {
		p.stream.Next()
	}

	return result, nil
}

// Subparse parses until one of the end tokens is reached
func (p *Parser) Subparse(endTokens []string) ([]nodes.Node, error) {
	var body []nodes.Node
	var dataBuffer []nodes.Node

	if endTokens != nil {
		p.endTokenStack = append(p.endTokenStack, endTokens)
		defer func() {
			if len(p.endTokenStack) > 0 {
				p.endTokenStack = p.endTokenStack[:len(p.endTokenStack)-1]
			}
		}()
	}

	flushData := func() {
		if len(dataBuffer) > 0 {
			lineno := dataBuffer[0].GetPosition().Line
			output := &nodes.Output{Nodes: make([]nodes.Expr, len(dataBuffer))}
			for i, node := range dataBuffer {
				if expr, ok := node.(nodes.Expr); ok {
					output.Nodes[i] = expr
				}
			}
			output.SetPosition(nodes.NewPosition(lineno, 0))
			body = append(body, output)
			dataBuffer = dataBuffer[:0]
		}
	}

	for !p.stream.Eof() {
		token := p.stream.Peek()

		if token.Type == lexer.TokenText {
			if token.Value != "" {
				templateData := &nodes.TemplateData{Data: token.Value}
				templateData.SetPosition(nodes.NewPosition(token.Line, token.Column))
				dataBuffer = append(dataBuffer, templateData)
			}
			p.stream.Next()
		} else if token.Type == lexer.TokenVariableStart {
			p.stream.Next()
			expr, err := p.ParseTuple()
			if err != nil {
				return nil, err
			}
			if _, err := p.Expect(lexer.TokenVariableEnd); err != nil {
				return nil, err
			}
			dataBuffer = append(dataBuffer, expr)
		} else if token.Type == lexer.TokenBlockStart {
			flushData()
			p.stream.Next()

			// Check if we've reached an end token
			if endTokens != nil && p.testEndTokens(endTokens) {
				return body, nil
			}

			stmt, err := p.ParseStatement()
			if err != nil {
				return nil, err
			}

			body = append(body, stmt)

			if _, err := p.Expect(lexer.TokenBlockEnd); err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("internal parsing error: unexpected token type %s", token.Type)
		}
	}

	flushData()
	return body, nil
}

// testEndTokens checks if the current token matches any of the end tokens
func (p *Parser) testEndTokens(endTokens []string) bool {
	token := p.stream.Peek()
	for _, endToken := range endTokens {
		if p.tokenMatchesRule(token, endToken) {
			return true
		}
	}
	return false
}

// Parse parses the whole template into a Template node
func (p *Parser) Parse() (*nodes.Template, error) {
	body, err := p.Subparse(nil)
	if err != nil {
		return nil, err
	}

	template := &nodes.Template{Body: body}
	template.SetPosition(nodes.NewPosition(1, 0))

	return template, nil
}

// ParseSet parses an assign statement
func (p *Parser) ParseSet() (nodes.Node, error) {
	lineno := p.stream.Next().Line // consume 'set'

	target, err := p.ParseAssignTarget()
	if err != nil {
		return nil, err
	}

	if p.SkipIf(lexer.TokenAssign) {
		expr, err := p.ParseTuple()
		if err != nil {
			return nil, err
		}
		assign := &nodes.Assign{Target: target, Node: expr}
		assign.SetPosition(nodes.NewPosition(lineno, 0))
		return assign, nil
	}

	filterNode, err := p.ParseFilter(nil)
	if err != nil {
		return nil, err
	}

	body, err := p.ParseStatements([]string{"name:endset"}, true)
	if err != nil {
		return nil, err
	}

	// Convert filterNode to proper Filter type if needed
	var filter *nodes.Filter
	if filterNode != nil {
		if f, ok := filterNode.(*nodes.Filter); ok {
			filter = f
		} else {
			// Create a wrapper filter if needed
			filter = &nodes.Filter{}
			// This is a temporary fix - the proper solution would be to ensure ParseFilter returns *nodes.Filter
		}
	}

	assignBlock := &nodes.AssignBlock{Target: target, Filter: filter, Body: body}
	assignBlock.SetPosition(nodes.NewPosition(lineno, 0))
	return assignBlock, nil
}

// ParseFor parses a for loop
func (p *Parser) ParseFor() (nodes.Node, error) {
	token, err := p.stream.Expect(lexer.TokenName)
	if err != nil {
		return nil, err
	}
	lineno := token.Line // consume 'for'

	target, err := p.ParseAssignTargetForLoop(true, false, []string{"name:in"}, false)
	if err != nil {
		return nil, err
	}

	if _, err := p.ExpectByName("in"); err != nil {
		return nil, err
	}

	iter, err := p.ParseTupleWithExtraRules(false, []string{"name:recursive"})
	if err != nil {
		return nil, err
	}

	var test nodes.Expr
	if p.SkipIfByName("if") {
		test, err = p.ParseExpression()
		if err != nil {
			return nil, err
		}
	}

	recursive := p.SkipIfByName("recursive")

	body, err := p.ParseStatements([]string{"name:endfor", "name:else"}, false)
	if err != nil {
		return nil, err
	}

	var elseBody []nodes.Node
	if p.stream.Peek().Value == "else" {
		p.stream.Next() // consume 'else'
		elseBody, err = p.ParseStatements([]string{"name:endfor"}, true)
		if err != nil {
			return nil, err
		}
	} else {
		// consume 'endfor'
		p.stream.Next()
	}

	forNode := &nodes.For{
		Target:    target,
		Iter:      iter,
		Body:      body,
		Else:      elseBody,
		Test:      test,
		Recursive: recursive,
	}
	forNode.SetPosition(nodes.NewPosition(lineno, 0))
	return forNode, nil
}

// ParseIf parses an if construct
func (p *Parser) ParseIf() (nodes.Node, error) {
	token, err := p.stream.Expect(lexer.TokenName)
	if err != nil {
		return nil, err
	}
	lineno := token.Line // consume 'if'

	root := &nodes.If{}
	root.SetPosition(nodes.NewPosition(lineno, 0))
	current := root

	for {
		// Parse test expression
		test, err := p.ParseExpression()
		if err != nil {
			return nil, err
		}
		current.Test = test

		// Parse body
		body, err := p.ParseStatements([]string{"name:elif", "name:else", "name:endif"}, false)
		if err != nil {
			return nil, err
		}
		current.Body = body

		// Check what comes next
		token := p.stream.Next()
		if token.Value == "elif" {
			// Create new If node for elif
			elifNode := &nodes.If{}
			elifNode.SetPosition(nodes.NewPosition(token.Line, 0))
			root.Elif = append(root.Elif, elifNode)
			current = elifNode
			continue
		} else if token.Value == "else" {
			// Parse else body
			elseBody, err := p.ParseStatements([]string{"name:endif"}, true)
			if err != nil {
				return nil, err
			}
			current.Else = elseBody
			break
		} else if token.Value == "endif" {
			break
		} else {
			return nil, p.Fail(fmt.Sprintf("expected elif, else, or endif, got %s", token.Value), token.Line, &TemplateSyntaxError{})
		}
	}

	return root, nil
}

// ParseBlock parses a block
func (p *Parser) ParseBlock() (nodes.Node, error) {
	lineno := p.stream.Next().Line // consume 'block'

	// Parse block name
	nameToken, err := p.Expect(lexer.TokenName)
	if err != nil {
		return nil, err
	}

	// Check for hyphens in block names (common Django -> Jinja migration issue)
	if p.stream.Peek().Type == lexer.TokenSub {
		return nil, p.Fail("Block names in Jinja have to be valid Python identifiers and may not contain hyphens, use an underscore instead.", nameToken.Line, &TemplateSyntaxError{})
	}

	block := &nodes.Block{
		Name:     nameToken.Value,
		Scoped:   p.SkipIfByName("scoped"),
		Required: p.SkipIfByName("required"),
	}

	// Parse block body
	body, err := p.ParseStatements([]string{"name:endblock"}, true)
	if err != nil {
		return nil, err
	}

	// If required, ensure body only contains whitespace or comments
	if block.Required {
		for _, bodyNode := range body {
			if output, ok := bodyNode.(*nodes.Output); ok {
				for _, outputNode := range output.Nodes {
					if templateData, ok := outputNode.(*nodes.TemplateData); ok {
						if strings.TrimSpace(templateData.Data) != "" {
							return nil, p.Fail("Required blocks can only contain comments or whitespace", nameToken.Line, &TemplateSyntaxError{})
						}
					} else {
						return nil, p.Fail("Required blocks can only contain comments or whitespace", nameToken.Line, &TemplateSyntaxError{})
					}
				}
			} else {
				return nil, p.Fail("Required blocks can only contain comments or whitespace", nameToken.Line, &TemplateSyntaxError{})
			}
		}
	}

	// Skip optional block name after endblock
	if p.stream.Peek().Type == lexer.TokenName && p.stream.Peek().Value == block.Name {
		p.stream.Next()
	}

	block.Body = body
	block.SetPosition(nodes.NewPosition(lineno, 0))
	return block, nil
}

// ParseExtends parses an extends statement
func (p *Parser) ParseExtends() (nodes.Node, error) {
	lineno := p.stream.Next().Line // consume 'extends'

	template, err := p.ParseExpression()
	if err != nil {
		return nil, err
	}

	extends := &nodes.Extends{Template: template}
	extends.SetPosition(nodes.NewPosition(lineno, 0))
	return extends, nil
}

// ParseWith parses a with statement
func (p *Parser) ParseWith() (nodes.Node, error) {
	lineno := p.stream.Next().Line // consume 'with'

	var targets []nodes.Expr
	var values []nodes.Expr

	for p.stream.Peek().Type != lexer.TokenBlockEnd {
		if len(targets) > 0 {
			if _, err := p.Expect(lexer.TokenComma); err != nil {
				return nil, err
			}
		}

		target, err := p.ParseAssignTargetWithExtraRules(false, false, nil, false)
		if err != nil {
			return nil, err
		}
		nodes.SetCtx(target, nodes.CtxParam)
		targets = append(targets, target)

		if _, err := p.Expect(lexer.TokenAssign); err != nil {
			return nil, err
		}

		value, err := p.ParseExpression()
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}

	body, err := p.ParseStatements([]string{"name:endwith"}, true)
	if err != nil {
		return nil, err
	}

	with := &nodes.With{
		Targets: targets,
		Values:  values,
		Body:    body,
	}
	with.SetPosition(nodes.NewPosition(lineno, 0))
	return with, nil
}

// ParseAutoescape parses an autoescape statement
func (p *Parser) ParseAutoescape() (nodes.Node, error) {
	lineno := p.stream.Next().Line // consume 'autoescape'

	expr, err := p.ParseExpression()
	if err != nil {
		return nil, err
	}

	body, err := p.ParseStatements([]string{"name:endautoescape"}, true)
	if err != nil {
		return nil, err
	}

	// Create keyword node for autoescape option
	keyword := &nodes.Keyword{
		Key:   "autoescape",
		Value: expr,
	}
	keyword.SetPosition(nodes.NewPosition(lineno, 0))

	// Create scoped eval context modifier
	modifier := &nodes.ScopedEvalContextModifier{
		EvalContextModifier: nodes.EvalContextModifier{
			Options: []*nodes.Keyword{keyword},
		},
		Body: body,
	}
	modifier.SetPosition(nodes.NewPosition(lineno, 0))

	// Wrap in scope node
	scope := &nodes.Scope{Body: []nodes.Node{modifier}}
	scope.SetPosition(nodes.NewPosition(lineno, 0))

	return scope, nil
}

// ParsePrint parses a print statement
func (p *Parser) ParsePrint() (nodes.Node, error) {
	lineno := p.stream.Next().Line // consume 'print'

	var exprs []nodes.Expr

	for p.stream.Peek().Type != lexer.TokenBlockEnd {
		if len(exprs) > 0 {
			if _, err := p.Expect(lexer.TokenComma); err != nil {
				return nil, err
			}
		}

		expr, err := p.ParseExpression()
		if err != nil {
			return nil, err
		}
		exprs = append(exprs, expr)
	}

	output := &nodes.Output{Nodes: exprs}
	output.SetPosition(nodes.NewPosition(lineno, 0))
	return output, nil
}

// Helper methods for consuming specific tokens by name
func (p *Parser) ExpectByName(name string) (lexer.Token, error) {
	token := p.stream.Peek()

	// Special handling for "in" keyword which can be TokenName or TokenComparison
	if name == "in" && (token.Type == lexer.TokenName || token.Type == lexer.TokenComparison) && token.Value == name {
		return p.stream.Next(), nil
	}

	if token.Type == lexer.TokenName && token.Value == name {
		return p.stream.Next(), nil
	}
	return token, p.Fail(fmt.Sprintf("expected name %q, got %s", name, p.describeCurrentToken()), token.Line, &TemplateSyntaxError{})
}

func (p *Parser) SkipIfByName(name string) bool {
	token := p.stream.Peek()

	// Special handling for "in" keyword which can be TokenName or TokenComparison
	if name == "in" && (token.Type == lexer.TokenName || token.Type == lexer.TokenComparison) && token.Value == name {
		p.stream.Next()
		return true
	}

	if token.Type == lexer.TokenName && token.Value == name {
		p.stream.Next()
		return true
	}
	return false
}
