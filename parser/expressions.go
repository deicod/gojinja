package parser

import (
	"fmt"
	"strconv"

	"github.com/deicod/gojinja/lexer"
	"github.com/deicod/gojinja/nodes"
)

// ParseExpression parses an expression
func (p *Parser) ParseExpression() (nodes.Expr, error) {
	return p.ParseConditionalExpr()
}

// ParseConditionalExpr parses conditional expressions (ternary operator)
func (p *Parser) ParseConditionalExpr() (nodes.Expr, error) {
	lineno := p.Current().Line

	expr1, err := p.ParseOr()
	if err != nil {
		return nil, err
	}

	for p.SkipIfByName("if") {
		expr2, err := p.ParseOr()
		if err != nil {
			return nil, err
		}

		var expr3 nodes.Expr
		if p.SkipIfByName("else") {
			expr3, err = p.ParseConditionalExpr()
			if err != nil {
				return nil, err
			}
		}

		condExpr := &nodes.CondExpr{
			Test:  expr2,
			Expr1: expr1,
			Expr2: expr3,
		}
		condExpr.SetPosition(nodes.NewPosition(lineno, 0))
		expr1 = condExpr
		lineno = p.Current().Line
	}

	return expr1, nil
}

// ParseOr parses logical OR expressions
func (p *Parser) ParseOr() (nodes.Expr, error) {
	lineno := p.Current().Line

	left, err := p.ParseAnd()
	if err != nil {
		return nil, err
	}

	for {
		token := p.stream.Peek()
		// Check for 'or' as either TokenName with value "or" or TokenOr
		if (token.Type == lexer.TokenName && token.Value == "or") || token.Type == lexer.TokenOr {
			p.stream.Next()
			right, err := p.ParseAnd()
			if err != nil {
				return nil, err
			}

			left = nodes.NewOr(left, right)
			left.SetPosition(nodes.NewPosition(lineno, 0))
			lineno = p.Current().Line
		} else {
			break
		}
	}

	return left, nil
}

// ParseAnd parses logical AND expressions
func (p *Parser) ParseAnd() (nodes.Expr, error) {
	lineno := p.Current().Line

	left, err := p.ParseNot()
	if err != nil {
		return nil, err
	}

	for {
		token := p.stream.Peek()
		// Check for 'and' as either TokenName with value "and" or TokenAnd
		if (token.Type == lexer.TokenName && token.Value == "and") || token.Type == lexer.TokenAnd {
			p.stream.Next()
			right, err := p.ParseNot()
			if err != nil {
				return nil, err
			}

			left = nodes.NewAnd(left, right)
			left.SetPosition(nodes.NewPosition(lineno, 0))
			lineno = p.Current().Line
		} else {
			break
		}
	}

	return left, nil
}

// ParseNot parses logical NOT expressions
func (p *Parser) ParseNot() (nodes.Expr, error) {
	token := p.stream.Peek()
	// Check for 'not' as either TokenName with value "not" or TokenNot
	if (token.Type == lexer.TokenName && token.Value == "not") || token.Type == lexer.TokenNot {
		lineno := p.stream.Next().Line
		expr, err := p.ParseNot()
		if err != nil {
			return nil, err
		}

		notNode := nodes.NewNot(expr)
		notNode.SetPosition(nodes.NewPosition(lineno, 0))
		return notNode, nil
	}

	return p.ParseCompare()
}

// testNotIn checks if we're at a "not in" sequence
func (p *Parser) testNotIn() bool {
	token := p.stream.Peek()
	nextToken := p.stream.PeekN(1)

	// Check for "not" followed by "in"
	isNot := (token.Type == lexer.TokenName && token.Value == "not") || token.Type == lexer.TokenNot
	isIn := (nextToken.Type == lexer.TokenName && nextToken.Value == "in") || nextToken.Type == lexer.TokenComparison && nextToken.Value == "in"

	return isNot && isIn
}

// ParseCompare parses comparison expressions
func (p *Parser) ParseCompare() (nodes.Expr, error) {
	lineno := p.Current().Line

	expr, err := p.ParseMath1()
	if err != nil {
		return nil, err
	}

	var ops []*nodes.Operand

	for {
		token := p.stream.Peek()
		// Check for comparison operators - can be TokenComparison type or specific value in compareOperators map
		if token.Type == lexer.TokenComparison || compareOperators[token.Value] {
			p.stream.Next()
			right, err := p.ParseMath1()
			if err != nil {
				return nil, err
			}

			ops = append(ops, &nodes.Operand{
				Op:   token.Value,
				Expr: right,
			})
		} else if p.SkipIfByName("in") {
			right, err := p.ParseMath1()
			if err != nil {
				return nil, err
			}

			ops = append(ops, &nodes.Operand{
				Op:   "in",
				Expr: right,
			})
		} else if p.testNotIn() {
			// Handle "not in" operator
			p.stream.Next() // consume 'not' (TokenName or TokenNot)
			p.stream.Next() // consume 'in' (TokenName or TokenComparison)
			right, err := p.ParseMath1()
			if err != nil {
				return nil, err
			}

			ops = append(ops, &nodes.Operand{
				Op:   "notin",
				Expr: right,
			})
		} else {
			break
		}
		lineno = p.Current().Line
	}

	if len(ops) == 0 {
		return expr, nil
	}

	compare := &nodes.Compare{
		Expr: expr,
		Ops:  ops,
	}
	compare.SetPosition(nodes.NewPosition(lineno, 0))
	return compare, nil
}

// ParseMath1 parses addition and subtraction
func (p *Parser) ParseMath1() (nodes.Expr, error) {
	lineno := p.Current().Line

	left, err := p.ParseConcat()
	if err != nil {
		return nil, err
	}

	for p.stream.Peek().Type == lexer.TokenAdd || p.stream.Peek().Type == lexer.TokenSub {
		opToken := p.stream.Next()
		right, err := p.ParseConcat()
		if err != nil {
			return nil, err
		}

		var expr nodes.Expr
		if opToken.Type == lexer.TokenAdd {
			expr = nodes.NewAdd(left, right)
		} else {
			expr = nodes.NewSub(left, right)
		}
		expr.SetPosition(nodes.NewPosition(lineno, 0))
		left = expr
		lineno = p.Current().Line
	}

	return left, nil
}

// ParseConcat parses string concatenation
func (p *Parser) ParseConcat() (nodes.Expr, error) {
	lineno := p.Current().Line

	// Parse first argument
	expr, err := p.ParseMath2()
	if err != nil {
		return nil, err
	}

	args := []nodes.Expr{expr}

	for p.stream.Peek().Type == lexer.TokenAdd && p.stream.Peek().Value == "~" {
		p.stream.Next() // consume '~'
		expr, err := p.ParseMath2()
		if err != nil {
			return nil, err
		}
		args = append(args, expr)
	}

	if len(args) == 1 {
		return args[0], nil
	}

	concat := &nodes.Concat{Nodes: args}
	concat.SetPosition(nodes.NewPosition(lineno, 0))
	return concat, nil
}

// ParseMath2 parses multiplication, division, and modulo
func (p *Parser) ParseMath2() (nodes.Expr, error) {
	lineno := p.Current().Line

	left, err := p.ParsePow()
	if err != nil {
		return nil, err
	}

	token := p.stream.Peek()
	for token.Type == lexer.TokenMul || token.Type == lexer.TokenDiv || token.Type == lexer.TokenFloorDiv || token.Type == lexer.TokenMod {
		opToken := p.stream.Next()
		right, err := p.ParsePow()
		if err != nil {
			return nil, err
		}

		var expr nodes.Expr
		switch opToken.Type {
		case lexer.TokenMul:
			expr = nodes.NewMul(left, right)
		case lexer.TokenDiv:
			expr = nodes.NewDiv(left, right)
		case lexer.TokenFloorDiv:
			expr = nodes.NewFloorDiv(left, right)
		case lexer.TokenMod:
			expr = nodes.NewMod(left, right)
		}
		expr.SetPosition(nodes.NewPosition(lineno, 0))
		left = expr
		token = p.stream.Peek()
		lineno = token.Line
	}

	return left, nil
}

// ParsePow parses exponentiation
func (p *Parser) ParsePow() (nodes.Expr, error) {
	lineno := p.Current().Line

	left, err := p.ParseUnary()
	if err != nil {
		return nil, err
	}

	for p.stream.Peek().Type == lexer.TokenPow {
		p.stream.Next() // consume '**'
		right, err := p.ParseUnary()
		if err != nil {
			return nil, err
		}

		pow := nodes.NewPow(left, right)
		pow.SetPosition(nodes.NewPosition(lineno, 0))
		left = pow
		lineno = p.Current().Line
	}

	return left, nil
}

// ParseUnary parses unary expressions (+, -)
func (p *Parser) ParseUnary() (nodes.Expr, error) {
	return p.parseUnary(true)
}

func (p *Parser) parseUnary(withFilter bool) (nodes.Expr, error) {
	token := p.stream.Peek()
	lineno := token.Line

	var node nodes.Expr

	if token.Type == lexer.TokenName && token.Value == "await" {
		if p.environment == nil || !p.environment.EnableAsync {
			return nil, p.Fail("encountered 'await' but async support is disabled; enable 'enable_async' to use await expressions", token.Line, &TemplateSyntaxError{})
		}
		p.stream.Next()
		expr, err := p.parseUnary(false)
		if err != nil {
			return nil, err
		}

		awaitNode := &nodes.Await{Node: expr}
		awaitNode.SetPosition(nodes.NewPosition(lineno, 0))
		node = awaitNode
	} else if token.Type == lexer.TokenSub {
		p.stream.Next()
		expr, err := p.parseUnary(false)
		if err != nil {
			return nil, err
		}

		node = nodes.NewNeg(expr)
		node.SetPosition(nodes.NewPosition(lineno, 0))
	} else if token.Type == lexer.TokenAdd {
		p.stream.Next()
		expr, err := p.parseUnary(false)
		if err != nil {
			return nil, err
		}

		node = nodes.NewPos(expr)
		node.SetPosition(nodes.NewPosition(lineno, 0))
	} else {
		expr, err := p.ParsePrimary()
		if err != nil {
			return nil, err
		}
		node = expr
	}

	node, err := p.parsePostfix(node)
	if err != nil {
		return nil, err
	}

	if withFilter {
		node, err = p.parseFilterExpr(node)
		if err != nil {
			return nil, err
		}
	}

	return node, nil
}

// ParsePrimary parses primary expressions (names, literals, parentheses)
func (p *Parser) ParsePrimary() (nodes.Expr, error) {
	return p.parsePrimary(false)
}

func (p *Parser) parsePrimary(withNamespace bool) (nodes.Expr, error) {
	token := p.stream.Peek()
	lineno := token.Line

	switch token.Type {
	case lexer.TokenName:
		p.stream.Next()
		if token.Value == "true" || token.Value == "false" || token.Value == "True" || token.Value == "False" {
			value := token.Value == "true" || token.Value == "True"
			constNode := &nodes.Const{Value: value}
			constNode.SetPosition(nodes.NewPosition(lineno, 0))
			return constNode, nil
		} else if token.Value == "none" || token.Value == "None" {
			constNode := &nodes.Const{Value: nil}
			constNode.SetPosition(nodes.NewPosition(lineno, 0))
			return constNode, nil
		} else if withNamespace && p.stream.Peek().Type == lexer.TokenDot {
			// Namespace reference
			p.stream.Next() // consume '.'
			attrToken, err := p.Expect(lexer.TokenName)
			if err != nil {
				return nil, err
			}

			nsRef := &nodes.NSRef{
				Name: token.Value,
				Attr: attrToken.Value,
			}
			nsRef.SetPosition(nodes.NewPosition(lineno, 0))
			return nsRef, nil
		} else {
			name := &nodes.Name{
				Name: token.Value,
				Ctx:  nodes.CtxLoad,
			}
			name.SetPosition(nodes.NewPosition(lineno, 0))
			return name, nil
		}

	case lexer.TokenString:
		p.stream.Next()
		// Concatenate adjacent strings
		value := token.Value
		for p.stream.Peek().Type == lexer.TokenString {
			value += p.stream.Next().Value
		}

		constNode := &nodes.Const{Value: value}
		constNode.SetPosition(nodes.NewPosition(lineno, 0))
		return constNode, nil

	case lexer.TokenNumber:
		p.stream.Next()
		var value interface{}
		if p.isInteger(token.Value) {
			if intVal, err := strconv.ParseInt(token.Value, 10, 64); err == nil {
				value = intVal
			} else {
				return nil, fmt.Errorf("invalid integer: %s", token.Value)
			}
		} else {
			if floatVal, err := strconv.ParseFloat(token.Value, 64); err == nil {
				value = floatVal
			} else {
				return nil, fmt.Errorf("invalid float: %s", token.Value)
			}
		}

		constNode := &nodes.Const{Value: value}
		constNode.SetPosition(nodes.NewPosition(lineno, 0))
		return constNode, nil

	case lexer.TokenLeftParen:
		p.stream.Next()
		expr, err := p.parseTuple(true, nil, false, false)
		if err != nil {
			return nil, err
		}
		if _, err := p.Expect(lexer.TokenRightParen); err != nil {
			return nil, err
		}
		return expr, nil

	case lexer.TokenLeftBracket:
		return p.parseList()

	case lexer.TokenLeftCurly:
		return p.parseDict()

	default:
		return nil, p.Fail(fmt.Sprintf("unexpected %s", p.describeCurrentToken()), token.Line, &TemplateSyntaxError{})
	}
}

// ParseTuple parses tuple expressions
func (p *Parser) ParseTuple() (nodes.Expr, error) {
	return p.ParseTupleWithCondExpr(true)
}

func (p *Parser) ParseTupleWithCondExpr(withCondExpr bool) (nodes.Expr, error) {
	return p.parseTuple(withCondExpr, nil, false, false)
}

func (p *Parser) ParseTupleSimplified(withCondExpr bool, extraEndRules []string) (nodes.Expr, error) {
	return p.parseTuple(withCondExpr, extraEndRules, true, false)
}

func (p *Parser) parseTuple(withCondExpr bool, extraEndRules []string, simplified bool, explicitParentheses bool) (nodes.Expr, error) {
	lineno := p.Current().Line

	var parse func() (nodes.Expr, error)
	if simplified {
		parse = func() (nodes.Expr, error) {
			return p.parsePrimary(false)
		}
	} else {
		parse = func() (nodes.Expr, error) {
			if withCondExpr {
				return p.ParseExpression()
			}
			return p.ParseOr()
		}
	}

	var args []nodes.Expr
	isTuple := false

	for {
		if len(args) > 0 {
			if _, err := p.Expect(lexer.TokenComma); err != nil {
				return nil, err
			}
		}

		if p.IsTupleEnd(extraEndRules) {
			break
		}

		expr, err := parse()
		if err != nil {
			return nil, err
		}
		args = append(args, expr)

		if p.stream.Peek().Type == lexer.TokenComma {
			isTuple = true
		} else {
			break
		}
		lineno = p.Current().Line
	}

	if !isTuple {
		if len(args) > 0 {
			return args[0], nil
		}

		// Empty tuple only valid with explicit parentheses
		if !explicitParentheses {
			return nil, p.Fail(fmt.Sprintf("Expected an expression, got %s", p.describeCurrentToken()), p.Current().Line, &TemplateSyntaxError{})
		}
	}

	tuple := &nodes.Tuple{
		Items: args,
		Ctx:   nodes.CtxLoad,
	}
	tuple.SetPosition(nodes.NewPosition(lineno, 0))
	return tuple, nil
}

// parseList parses list literals
func (p *Parser) parseList() (nodes.Expr, error) {
	token, err := p.Expect(lexer.TokenLeftBracket)
	if err != nil {
		return nil, err
	}
	lineno := token.Line

	var items []nodes.Expr

	for p.stream.Peek().Type != lexer.TokenRightBracket {
		if len(items) > 0 {
			if _, err := p.Expect(lexer.TokenComma); err != nil {
				return nil, err
			}
		}

		if p.stream.Peek().Type == lexer.TokenRightBracket {
			break
		}

		expr, err := p.ParseExpression()
		if err != nil {
			return nil, err
		}
		items = append(items, expr)
	}

	if _, err := p.Expect(lexer.TokenRightBracket); err != nil {
		return nil, err
	}

	list := &nodes.List{Items: items}
	list.SetPosition(nodes.NewPosition(lineno, 0))
	return list, nil
}

// parseDict parses dictionary literals
func (p *Parser) parseDict() (nodes.Expr, error) {
	token, err := p.Expect(lexer.TokenLeftCurly)
	if err != nil {
		return nil, err
	}
	lineno := token.Line

	var items []*nodes.Pair

	for p.stream.Peek().Type != lexer.TokenRightCurly {
		if len(items) > 0 {
			if _, err := p.Expect(lexer.TokenComma); err != nil {
				return nil, err
			}
		}

		if p.stream.Peek().Type == lexer.TokenRightCurly {
			break
		}

		key, err := p.ParseExpression()
		if err != nil {
			return nil, err
		}

		if _, err := p.Expect(lexer.TokenColon); err != nil {
			return nil, err
		}

		value, err := p.ParseExpression()
		if err != nil {
			return nil, err
		}

		pair := &nodes.Pair{Key: key, Value: value}
		pair.SetPosition(nodes.NewPosition(key.GetPosition().Line, 0))
		items = append(items, pair)
	}

	if _, err := p.Expect(lexer.TokenRightCurly); err != nil {
		return nil, err
	}

	dict := &nodes.Dict{Items: items}
	dict.SetPosition(nodes.NewPosition(lineno, 0))
	return dict, nil
}

// parsePostfix parses postfix expressions (attribute access, indexing, calls)
func (p *Parser) parsePostfix(node nodes.Expr) (nodes.Expr, error) {
	for {
		token := p.stream.Peek()
		if token.Type == lexer.TokenDot || token.Type == lexer.TokenLeftBracket {
			var err error
			node, err = p.parseSubscript(node)
			if err != nil {
				return nil, err
			}
		} else if token.Type == lexer.TokenLeftParen {
			var err error
			node, err = p.parseCall(node)
			if err != nil {
				return nil, err
			}
		} else {
			break
		}
	}
	return node, nil
}

// parseSubscript parses attribute access and indexing
func (p *Parser) parseSubscript(node nodes.Expr) (nodes.Expr, error) {
	token := p.stream.Next()
	lineno := token.Line

	if token.Type == lexer.TokenDot {
		attrToken, err := p.Expect(lexer.TokenName)
		if err != nil {
			return nil, err
		}

		getattr := &nodes.Getattr{
			Node: node,
			Attr: attrToken.Value,
			Ctx:  nodes.CtxLoad,
		}
		getattr.SetPosition(nodes.NewPosition(lineno, 0))
		return getattr, nil
	}

	if token.Type == lexer.TokenLeftBracket {
		var args []nodes.Expr

		for p.stream.Peek().Type != lexer.TokenRightBracket {
			if len(args) > 0 {
				if _, err := p.Expect(lexer.TokenComma); err != nil {
					return nil, err
				}
			}
			arg, err := p.parseSubscribed()
			if err != nil {
				return nil, err
			}
			args = append(args, arg)
		}

		if _, err := p.Expect(lexer.TokenRightBracket); err != nil {
			return nil, err
		}

		var arg nodes.Expr
		if len(args) == 1 {
			arg = args[0]
		} else {
			tuple := &nodes.Tuple{
				Items: args,
				Ctx:   nodes.CtxLoad,
			}
			tuple.SetPosition(nodes.NewPosition(lineno, 0))
			arg = tuple
		}

		getitem := &nodes.Getitem{
			Node: node,
			Arg:  arg,
			Ctx:  nodes.CtxLoad,
		}
		getitem.SetPosition(nodes.NewPosition(lineno, 0))
		return getitem, nil
	}

	return nil, p.Fail("expected subscript expression", token.Line, &TemplateSyntaxError{})
}

// parseSubscribed parses slice expressions
func (p *Parser) parseSubscribed() (nodes.Expr, error) {
	lineno := p.Current().Line

	var args []nodes.Expr

	if p.stream.Peek().Type == lexer.TokenColon {
		p.stream.Next()
		args = append(args, nil)
	} else {
		expr, err := p.ParseExpression()
		if err != nil {
			return nil, err
		}
		args = append(args, expr)

		if p.stream.Peek().Type != lexer.TokenColon {
			return args[0], nil
		}
		p.stream.Next()
	}

	if p.stream.Peek().Type == lexer.TokenColon {
		args = append(args, nil)
	} else if !p.TestAny("rbracket", "comma") {
		expr, err := p.ParseExpression()
		if err != nil {
			return nil, err
		}
		args = append(args, expr)
	} else {
		args = append(args, nil)
	}

	if p.stream.Peek().Type == lexer.TokenColon {
		p.stream.Next()
		if !p.TestAny("rbracket", "comma") {
			expr, err := p.ParseExpression()
			if err != nil {
				return nil, err
			}
			args = append(args, expr)
		} else {
			args = append(args, nil)
		}
	} else {
		args = append(args, nil)
	}

	// Pad args to 3 elements if needed
	for len(args) < 3 {
		args = append(args, nil)
	}

	slice := &nodes.Slice{
		Start: args[0],
		Stop:  args[1],
		Step:  args[2],
	}
	slice.SetPosition(nodes.NewPosition(lineno, 0))
	return slice, nil
}

// parseCall parses function calls
func (p *Parser) parseCall(node nodes.Expr) (nodes.Expr, error) {
	lineno := p.Current().Line

	args, kwargs, dynArgs, dynKwargs, err := p.parseCallArgs()
	if err != nil {
		return nil, err
	}

	call := &nodes.Call{
		Node:      node,
		Args:      args,
		Kwargs:    kwargs,
		DynArgs:   dynArgs,
		DynKwargs: dynKwargs,
	}
	call.SetPosition(nodes.NewPosition(lineno, 0))
	return call, nil
}

// parseCallArgs parses function call arguments
func (p *Parser) parseCallArgs() ([]nodes.Expr, []*nodes.Keyword, nodes.Expr, nodes.Expr, error) {
	if _, err := p.Expect(lexer.TokenLeftParen); err != nil {
		return nil, nil, nil, nil, err
	}

	var args []nodes.Expr
	var kwargs []*nodes.Keyword
	var dynArgs, dynKwargs nodes.Expr
	requireComma := false

	for p.stream.Peek().Type != lexer.TokenRightParen {
		if requireComma {
			if _, err := p.Expect(lexer.TokenComma); err != nil {
				return nil, nil, nil, nil, err
			}

			// Support trailing comma
			if p.stream.Peek().Type == lexer.TokenRightParen {
				break
			}
		}

		token := p.stream.Peek()
		if token.Type == lexer.TokenMul {
			if dynArgs != nil || dynKwargs != nil {
				return nil, nil, nil, nil, p.Fail("invalid syntax for function call expression", token.Line, &TemplateSyntaxError{})
			}
			p.stream.Next()
			expr, err := p.ParseExpression()
			if err != nil {
				return nil, nil, nil, nil, err
			}
			dynArgs = expr
		} else if token.Type == lexer.TokenPow {
			if dynKwargs != nil {
				return nil, nil, nil, nil, p.Fail("invalid syntax for function call expression", token.Line, &TemplateSyntaxError{})
			}
			p.stream.Next()
			expr, err := p.ParseExpression()
			if err != nil {
				return nil, nil, nil, nil, err
			}
			dynKwargs = expr
		} else if token.Type == lexer.TokenName && p.Look().Type == lexer.TokenAssign {
			// Keyword argument
			if dynKwargs != nil {
				return nil, nil, nil, nil, p.Fail("invalid syntax for function call expression", token.Line, &TemplateSyntaxError{})
			}
			key := token.Value
			p.stream.Next() // consume name
			p.stream.Next() // consume '='
			value, err := p.ParseExpression()
			if err != nil {
				return nil, nil, nil, nil, err
			}

			keyword := &nodes.Keyword{
				Key:   key,
				Value: value,
			}
			keyword.SetPosition(nodes.NewPosition(token.Line, 0))
			kwargs = append(kwargs, keyword)
		} else {
			// Positional argument
			if dynArgs != nil || dynKwargs != nil || len(kwargs) > 0 {
				return nil, nil, nil, nil, p.Fail("invalid syntax for function call expression", token.Line, &TemplateSyntaxError{})
			}
			expr, err := p.ParseExpression()
			if err != nil {
				return nil, nil, nil, nil, err
			}
			args = append(args, expr)
		}

		requireComma = true
	}

	if _, err := p.Expect(lexer.TokenRightParen); err != nil {
		return nil, nil, nil, nil, err
	}

	return args, kwargs, dynArgs, dynKwargs, nil
}

// parseFilterExpr parses filter and test expressions
func (p *Parser) parseFilterExpr(node nodes.Expr) (nodes.Expr, error) {
	for {
		token := p.stream.Peek()
		if token.Type == lexer.TokenPipe {
			var err error
			node, err = p.parseFilter(node)
			if err != nil {
				return nil, err
			}
		} else if token.Type == lexer.TokenName && token.Value == "is" {
			var err error
			node, err = p.parseTest(node)
			if err != nil {
				return nil, err
			}
		} else if token.Type == lexer.TokenLeftParen {
			var err error
			node, err = p.parseCall(node)
			if err != nil {
				return nil, err
			}
		} else {
			break
		}
	}
	return node, nil
}

// ParseFilter parses filter expressions (public method)
func (p *Parser) ParseFilter(node nodes.Expr) (nodes.Expr, error) {
	return p.parseFilter(node)
}

// parseFilter parses filter expressions
func (p *Parser) parseFilter(node nodes.Expr) (nodes.Expr, error) {
	return p.parseFilterInternal(node, false)
}

func (p *Parser) parseFilterInternal(node nodes.Expr, startInline bool) (nodes.Expr, error) {
	for p.stream.Peek().Type == lexer.TokenPipe || startInline {
		if !startInline {
			p.stream.Next() // consume '|'
		}

		token, err := p.Expect(lexer.TokenName)
		if err != nil {
			return nil, err
		}
		name := token.Value

		// Handle dotted filter names
		for p.stream.Peek().Type == lexer.TokenDot {
			p.stream.Next() // consume '.'
			attrToken, err := p.Expect(lexer.TokenName)
			if err != nil {
				return nil, err
			}
			name += "." + attrToken.Value
		}

		var args []nodes.Expr
		var kwargs []*nodes.Keyword
		var dynArgs, dynKwargs nodes.Expr

		if p.stream.Peek().Type == lexer.TokenLeftParen {
			args, kwargs, dynArgs, dynKwargs, err = p.parseCallArgs()
			if err != nil {
				return nil, err
			}
		}

		// Convert kwargs to pairs for Filter node
		var kwargPairs []*nodes.Pair
		for _, kw := range kwargs {
			pair := &nodes.Pair{
				Key:   &nodes.Const{Value: kw.Key},
				Value: kw.Value,
			}
			pair.SetPosition(kw.GetPosition())
			kwargPairs = append(kwargPairs, pair)
		}

		filter := &nodes.Filter{
			FilterTestCommon: nodes.FilterTestCommon{
				Node:      node,
				Name:      name,
				Args:      args,
				Kwargs:    kwargPairs,
				DynArgs:   dynArgs,
				DynKwargs: dynKwargs,
				IsFilter:  true,
			},
		}
		filter.SetPosition(nodes.NewPosition(token.Line, 0))
		node = filter
		startInline = false
	}

	return node, nil
}

// parseTest parses test expressions
func (p *Parser) parseTest(node nodes.Expr) (nodes.Expr, error) {
	token := p.stream.Next() // consume 'is'
	lineno := token.Line

	negated := p.SkipIfByName("not")

	nameToken, err := p.Expect(lexer.TokenName)
	if err != nil {
		return nil, err
	}
	name := nameToken.Value

	// Handle dotted test names
	for p.stream.Peek().Type == lexer.TokenDot {
		p.stream.Next() // consume '.'
		attrToken, err := p.Expect(lexer.TokenName)
		if err != nil {
			return nil, err
		}
		name += "." + attrToken.Value
	}

	var args []nodes.Expr
	var kwargs []*nodes.Keyword
	var dynArgs, dynKwargs nodes.Expr

	if p.stream.Peek().Type == lexer.TokenLeftParen {
		args, kwargs, dynArgs, dynKwargs, err = p.parseCallArgs()
		if err != nil {
			return nil, err
		}
	} else if p.TestAny("name", "string", "integer", "float", "lparen", "lbracket", "lbrace") &&
		!p.TestAny("name:else", "name:or", "name:and") {
		if p.stream.Peek().Value == "is" {
			return nil, p.Fail("You cannot chain multiple tests with is", p.stream.Peek().Line, &TemplateSyntaxError{})
		}

		argNode, err := p.parsePrimary(false)
		if err != nil {
			return nil, err
		}
		argNode, err = p.parsePostfix(argNode)
		if err != nil {
			return nil, err
		}
		args = []nodes.Expr{argNode}
	}

	// Convert kwargs to pairs for Test node
	var kwargPairs []*nodes.Pair
	for _, kw := range kwargs {
		pair := &nodes.Pair{
			Key:   &nodes.Const{Value: kw.Key},
			Value: kw.Value,
		}
		pair.SetPosition(kw.GetPosition())
		kwargPairs = append(kwargPairs, pair)
	}

	test := &nodes.Test{
		FilterTestCommon: nodes.FilterTestCommon{
			Node:      node,
			Name:      name,
			Args:      args,
			Kwargs:    kwargPairs,
			DynArgs:   dynArgs,
			DynKwargs: dynKwargs,
			IsFilter:  false,
		},
	}
	test.SetPosition(nodes.NewPosition(lineno, 0))

	if negated {
		notNode := nodes.NewNot(test)
		notNode.SetPosition(nodes.NewPosition(lineno, 0))
		return notNode, nil
	}

	return test, nil
}

// ParseAssignTargetWithExtraRules parses assignment targets with extra end rules (public method)
func (p *Parser) ParseAssignTargetWithExtraRules(withTuple bool, nameOnly bool, extraEndRules []string, withNamespace bool) (nodes.Expr, error) {
	return p.parseAssignTarget(withTuple, nameOnly, extraEndRules, withNamespace)
}

// ParseTupleWithExtraRules parses tuple with extra end rules (public method)
func (p *Parser) ParseTupleWithExtraRules(withCondExpr bool, extraEndRules []string) (nodes.Expr, error) {
	return p.parseTuple(withCondExpr, extraEndRules, false, false)
}

// ParseAssignTargetForLoop parses assignment targets for for loops with extra rules
func (p *Parser) ParseAssignTargetForLoop(withTuple bool, nameOnly bool, extraEndRules []string, withNamespace bool) (nodes.Expr, error) {
	return p.parseAssignTarget(withTuple, nameOnly, extraEndRules, withNamespace)
}

// parseAssignTarget is the private method for parsing assignment targets
func (p *Parser) parseAssignTarget(withTuple bool, nameOnly bool, extraEndRules []string, withNamespace bool) (nodes.Expr, error) {
	var target nodes.Expr
	var err error

	if withNamespace {
		target, err = p.parsePrimary(true)
	} else {
		target, err = p.parsePrimary(false)
	}
	if err != nil {
		return nil, err
	}

	target, err = p.parsePostfix(target)
	if err != nil {
		return nil, err
	}

	// Handle tuple unpacking
	if withTuple && p.stream.Peek().Type == lexer.TokenComma {
		var items []nodes.Expr
		items = append(items, target)

		for p.stream.Peek().Type == lexer.TokenComma {
			p.stream.Next() // consume ','

			if p.IsTupleEnd(extraEndRules) {
				break
			}

			item, err := p.parsePrimary(false)
			if err != nil {
				return nil, err
			}
			items = append(items, item)
		}

		tuple := &nodes.Tuple{
			Items: items,
			Ctx:   nodes.CtxStore,
		}
		tuple.SetPosition(target.GetPosition())
		return tuple, nil
	}

	// Set context for assignment
	nodes.SetCtx(target, nodes.CtxStore)
	return target, nil
}
