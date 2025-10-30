package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/deicod/gojinja/lexer"
	"github.com/deicod/gojinja/nodes"
)

// TemplateSyntaxError represents a syntax error in a template
type TemplateSyntaxError struct {
	Message  string
	Line     int
	Column   int
	Name     string
	Filename string
}

func (e *TemplateSyntaxError) Error() string {
	if e.Filename != "" {
		return fmt.Sprintf("%s at line %d, column %d in %s", e.Message, e.Line, e.Column, e.Filename)
	}
	if e.Name != "" {
		return fmt.Sprintf("%s at line %d, column %d in %s", e.Message, e.Line, e.Column, e.Name)
	}
	return fmt.Sprintf("%s at line %d, column %d", e.Message, e.Line, e.Column)
}

// TemplateAssertionError represents an assertion error in a template
type TemplateAssertionError struct {
	TemplateSyntaxError
}

func (e *TemplateAssertionError) Error() string {
	return e.TemplateSyntaxError.Error()
}

// Extension represents a parser extension
type Extension interface {
	Tags() []string
	Parse(parser *Parser) (nodes.Node, error)
}

// Environment represents the template environment (placeholder for now)
type Environment struct {
	Extensions          []Extension
	TrimBlocks          bool
	LstripBlocks        bool
	KeepTrailingNewline bool
	LineStatementPrefix string
	LineCommentPrefix   string
	EnableAsync         bool
}

// Parser represents the central parsing class Jinja uses
type Parser struct {
	environment    *Environment
	stream         *lexer.TokenStream
	name           string
	filename       string
	closed         bool
	extensions     map[string]Extension
	lastIdentifier int
	tagStack       []string
	endTokenStack  [][]string
}

// NewParser creates a new parser instance
func NewParser(env *Environment, source, name, filename string, state string) (*Parser, error) {
	lexerConfig := lexer.DefaultLexerConfig()
	if env != nil {
		lexerConfig.TrimBlocks = env.TrimBlocks
		lexerConfig.LstripBlocks = env.LstripBlocks
		lexerConfig.KeepTrailingNewline = env.KeepTrailingNewline
		lexerConfig.Delimiters.LineStatement = env.LineStatementPrefix
		lexerConfig.Delimiters.LineComment = env.LineCommentPrefix
	}
	l := lexer.NewLexer(lexerConfig)

	stream, err := l.Tokenize(source, name, filename, lexer.LexerState(state))
	if err != nil {
		return nil, err
	}

	parser := &Parser{
		environment:   env,
		stream:        stream,
		name:          name,
		filename:      filename,
		extensions:    make(map[string]Extension),
		tagStack:      make([]string, 0),
		endTokenStack: make([][]string, 0),
	}

	// Register extensions
	if env != nil {
		for _, ext := range env.Extensions {
			for _, tag := range ext.Tags() {
				parser.extensions[tag] = ext
			}
		}
	}

	return parser, nil
}

// Fail creates a syntax error with position information
func (p *Parser) Fail(msg string, lineno int, errType error) error {
	if lineno == 0 {
		token := p.stream.Peek()
		if token.Type != lexer.TokenEOF {
			lineno = token.Line
		} else {
			lineno = 1
		}
	}

	if errType == nil {
		errType = &TemplateSyntaxError{}
	}

	switch errType.(type) {
	case *TemplateSyntaxError:
		return &TemplateSyntaxError{
			Message:  msg,
			Line:     lineno,
			Column:   0, // We don't track column in current implementation
			Name:     p.name,
			Filename: p.filename,
		}
	case *TemplateAssertionError:
		return &TemplateAssertionError{
			TemplateSyntaxError: TemplateSyntaxError{
				Message:  msg,
				Line:     lineno,
				Column:   0,
				Name:     p.name,
				Filename: p.filename,
			},
		}
	default:
		return fmt.Errorf("%s at line %d", msg, lineno)
	}
}

// FailUnknownTag is called when the parser encounters an unknown tag
func (p *Parser) FailUnknownTag(name string, lineno int) error {
	return p.failUntilEOF(name, p.endTokenStack, lineno)
}

// FailEOF is called when EOF is encountered unexpectedly
func (p *Parser) FailEOF(endTokens []string, lineno int) error {
	stack := make([][]string, len(p.endTokenStack))
	copy(stack, p.endTokenStack)
	if endTokens != nil {
		stack = append(stack, endTokens)
	}
	return p.failUntilEOF("", stack, lineno)
}

// failUntilEOF creates an appropriate error message for EOF or unknown tag situations
func (p *Parser) failUntilEOF(name string, endTokenStack [][]string, lineno int) error {
	expected := make(map[string]bool)
	for _, exprs := range endTokenStack {
		for _, expr := range exprs {
			expected[expr] = true
		}
	}

	var currentlyLooking string
	if len(endTokenStack) > 0 {
		lastTokens := endTokenStack[len(endTokenStack)-1]
		if len(lastTokens) > 0 {
			currentlyLooking = strings.Join(lastTokens, " or ")
		}
	}

	var message strings.Builder
	if name == "" {
		message.WriteString("Unexpected end of template.")
	} else {
		message.WriteString(fmt.Sprintf("Encountered unknown tag %q.", name))
	}

	if currentlyLooking != "" {
		if name != "" && expected[name] {
			message.WriteString(fmt.Sprintf(" You probably made a nesting mistake. Jinja is expecting this tag, but currently looking for %s.", currentlyLooking))
		} else {
			message.WriteString(fmt.Sprintf(" Jinja was looking for the following tags: %s.", currentlyLooking))
		}
	}

	if len(p.tagStack) > 0 {
		message.WriteString(fmt.Sprintf(" The innermost block that needs to be closed is %q.", p.tagStack[len(p.tagStack)-1]))
	}

	return p.Fail(message.String(), lineno, &TemplateSyntaxError{})
}

// IsTupleEnd checks if we're at the end of a tuple
func (p *Parser) IsTupleEnd(extraEndRules []string) bool {
	token := p.stream.Peek()

	// These tokens always end a tuple
	if token.Type == lexer.TokenVariableEnd || token.Type == lexer.TokenBlockEnd || token.Type == lexer.TokenRightParen {
		return true
	}

	// Check extra end rules
	if extraEndRules != nil {
		for _, rule := range extraEndRules {
			if p.tokenMatchesRule(token, rule) {
				return true
			}
		}
	}

	return false
}

// tokenMatchesRule checks if a token matches a specific rule
func (p *Parser) tokenMatchesRule(token lexer.Token, rule string) bool {
	// Handle name:value rules like "name:for"
	if strings.Contains(rule, ":") {
		parts := strings.SplitN(rule, ":", 2)
		if parts[0] == "name" && token.Type == lexer.TokenName && token.Value == parts[1] {
			return true
		}
	}

	// Direct token type match
	switch rule {
	case "name":
		return token.Type == lexer.TokenName
	case "string":
		return token.Type == lexer.TokenString
	case "integer":
		return token.Type == lexer.TokenNumber && p.isInteger(token.Value)
	case "float":
		return token.Type == lexer.TokenNumber && !p.isInteger(token.Value)
	case "lparen":
		return token.Type == lexer.TokenLeftParen
	case "rparen":
		return token.Type == lexer.TokenRightParen
	case "lbracket":
		return token.Type == lexer.TokenLeftBracket
	case "rbracket":
		return token.Type == lexer.TokenRightBracket
	case "lbrace":
		return token.Type == lexer.TokenLeftCurly
	case "rbrace":
		return token.Type == lexer.TokenRightCurly
	case "comma":
		return token.Type == lexer.TokenComma
	case "colon":
		return token.Type == lexer.TokenColon
	case "assign":
		return token.Type == lexer.TokenAssign
	case "pipe":
		return token.Type == lexer.TokenPipe
	case "dot":
		return token.Type == lexer.TokenDot
	case "add":
		return token.Type == lexer.TokenAdd
	case "sub":
		return token.Type == lexer.TokenSub
	case "mul":
		return token.Type == lexer.TokenMul
	case "div":
		return token.Type == lexer.TokenDiv
	case "mod":
		return token.Type == lexer.TokenMod
	case "pow":
		return token.Type == lexer.TokenPow
	case "not":
		return token.Type == lexer.TokenNot
	case "and":
		return token.Type == lexer.TokenAnd
	case "or":
		return token.Type == lexer.TokenOr
	}

	return false
}

// isInteger checks if a string represents an integer value
func (p *Parser) isInteger(value string) bool {
	_, err := strconv.ParseInt(value, 10, 64)
	return err == nil
}

// FreeIdentifier returns a new free identifier as InternalName
func (p *Parser) FreeIdentifier(lineno int) *nodes.InternalName {
	p.lastIdentifier++
	node := &nodes.InternalName{Name: fmt.Sprintf("fi%d", p.lastIdentifier)}
	if lineno != 0 {
		node.SetPosition(nodes.NewPosition(lineno, 0))
	}
	return node
}

// Current returns the current token without consuming it
func (p *Parser) Current() lexer.Token {
	return p.stream.Peek()
}

// SkipIf skips a token if it matches the expected type
func (p *Parser) SkipIf(expectedType lexer.TokenType) bool {
	token := p.stream.Peek()
	if token.Type == expectedType {
		p.stream.Next()
		return true
	}
	return false
}

// Expect consumes and returns a token, failing if it doesn't match the expected type
func (p *Parser) Expect(expectedType lexer.TokenType) (lexer.Token, error) {
	token := p.stream.Peek()
	if token.Type == expectedType {
		return p.stream.Next(), nil
	}

	// Try to get a better description of what was expected
	expected := p.describeToken(expectedType)
	got := p.describeCurrentToken()
	return token, p.Fail(fmt.Sprintf("expected %s, got %s", expected, got), token.Line, &TemplateSyntaxError{})
}

// describeToken provides a human-readable description of a token type
func (p *Parser) describeToken(tokenType lexer.TokenType) string {
	switch tokenType {
	case lexer.TokenEOF:
		return "end of template"
	case lexer.TokenText:
		return "text"
	case lexer.TokenVariableStart:
		return "variable start ('{{')"
	case lexer.TokenVariableEnd:
		return "variable end ('}}')"
	case lexer.TokenBlockStart:
		return "block start ('{%')"
	case lexer.TokenBlockEnd:
		return "block end ('%}')"
	case lexer.TokenCommentStart:
		return "comment start ('{#')"
	case lexer.TokenCommentEnd:
		return "comment end ('#}')"
	case lexer.TokenName:
		return "name"
	case lexer.TokenString:
		return "string"
	case lexer.TokenNumber:
		return "number"
	case lexer.TokenAssign:
		return "assignment operator ('=')"
	case lexer.TokenComma:
		return "comma (',')"
	case lexer.TokenColon:
		return "colon (':')"
	case lexer.TokenLeftParen:
		return "left parenthesis ('(')"
	case lexer.TokenRightParen:
		return "right parenthesis (')')"
	case lexer.TokenLeftBracket:
		return "left bracket ('[')"
	case lexer.TokenRightBracket:
		return "right bracket (']')"
	case lexer.TokenLeftCurly:
		return "left curly brace ('{')"
	case lexer.TokenRightCurly:
		return "right curly brace ('}')"
	case lexer.TokenDot:
		return "dot ('.')"
	case lexer.TokenPipe:
		return "pipe ('|')"
	default:
		return tokenType.String()
	}
}

// describeCurrentToken provides a description of the current token
func (p *Parser) describeCurrentToken() string {
	token := p.stream.Peek()
	if token.Type == lexer.TokenName {
		return fmt.Sprintf("name %q", token.Value)
	}
	if token.Type == lexer.TokenString {
		return fmt.Sprintf("string %q", token.Value)
	}
	return p.describeToken(token.Type)
}

// TestAny checks if the current token matches any of the given rules
func (p *Parser) TestAny(rules ...string) bool {
	token := p.stream.Peek()
	for _, rule := range rules {
		if p.tokenMatchesRule(token, rule) {
			return true
		}
	}
	return false
}

// Skip consumes and discards the next n tokens
func (p *Parser) Skip(n int) {
	for i := 0; i < n; i++ {
		if p.stream.Peek().Type != lexer.TokenEOF {
			p.stream.Next()
		}
	}
}

// Look returns the token after the current one without consuming
func (p *Parser) Look() lexer.Token {
	return p.stream.PeekN(1)
}

// statement keywords that define statement types
var statementKeywords = map[string]bool{
	"for":        true,
	"if":         true,
	"block":      true,
	"extends":    true,
	"print":      true,
	"macro":      true,
	"include":    true,
	"from":       true,
	"import":     true,
	"set":        true,
	"with":       true,
	"namespace":  true,
	"export":     true,
	"trans":      true,
	"blocktrans": true,
	"autoescape": true,
	"break":      true,
	"continue":   true,
	"do":         true,
	"spaceless":  true,
}

// compare operators for parsing
var compareOperators = map[string]bool{
	"eq":   true,
	"ne":   true,
	"lt":   true,
	"lteq": true,
	"gt":   true,
	"gteq": true,
}

// math operators mapping
var mathNodes = map[string]func(left, right nodes.Expr) nodes.Expr{
	"add": func(left, right nodes.Expr) nodes.Expr {
		return nodes.NewAdd(left, right)
	},
	"sub": func(left, right nodes.Expr) nodes.Expr {
		return nodes.NewSub(left, right)
	},
	"mul": func(left, right nodes.Expr) nodes.Expr {
		return nodes.NewMul(left, right)
	},
	"div": func(left, right nodes.Expr) nodes.Expr {
		return nodes.NewDiv(left, right)
	},
	"floordiv": func(left, right nodes.Expr) nodes.Expr {
		return nodes.NewFloorDiv(left, right)
	},
	"mod": func(left, right nodes.Expr) nodes.Expr {
		return nodes.NewMod(left, right)
	},
}
