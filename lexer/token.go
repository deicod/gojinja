package lexer

import (
	"fmt"
)

// TokenType represents the type of a token
type TokenType int

const (
	TokenEOF TokenType = iota
	TokenText
	TokenVariableStart
	TokenVariableEnd
	TokenBlockStart
	TokenBlockEnd
	TokenCommentStart
	TokenCommentEnd
	TokenOperator
	TokenName
	TokenString
	TokenNumber
	TokenAssign
	TokenComma
	TokenColon
	TokenSemicolon
	TokenPipe
	TokenLeftParen
	TokenRightParen
	TokenLeftBracket
	TokenRightBracket
	TokenLeftCurly
	TokenRightCurly
	TokenDot
	TokenComparison
	TokenAdd
	TokenSub
	TokenMul
	TokenDiv
	TokenFloorDiv
	TokenMod
	TokenPow
	TokenNot
	TokenAnd
	TokenOr
	TokenTernary
	TokenTernaryElse
	TokenWhitespace
	TokenLinebreak
)

var tokenNames = map[TokenType]string{
	TokenEOF:           "EOF",
	TokenText:          "TEXT",
	TokenVariableStart: "VAR_START",
	TokenVariableEnd:   "VAR_END",
	TokenBlockStart:    "BLOCK_START",
	TokenBlockEnd:      "BLOCK_END",
	TokenCommentStart:  "COMMENT_START",
	TokenCommentEnd:    "COMMENT_END",
	TokenOperator:      "OPERATOR",
	TokenName:          "NAME",
	TokenString:        "STRING",
	TokenNumber:        "NUMBER",
	TokenAssign:        "ASSIGN",
	TokenComma:         "COMMA",
	TokenColon:         "COLON",
	TokenSemicolon:     "SEMICOLON",
	TokenPipe:          "PIPE",
	TokenLeftParen:     "LPAREN",
	TokenRightParen:    "RPAREN",
	TokenLeftBracket:   "LBRACKET",
	TokenRightBracket:  "RBRACKET",
	TokenLeftCurly:     "LCURLY",
	TokenRightCurly:    "RCURLY",
	TokenDot:           "DOT",
	TokenComparison:    "COMPARISON",
	TokenAdd:           "ADD",
	TokenSub:           "SUB",
	TokenMul:           "MUL",
	TokenDiv:           "DIV",
	TokenFloorDiv:      "FLOORDIV",
	TokenMod:           "MOD",
	TokenPow:           "POW",
	TokenNot:           "NOT",
	TokenAnd:           "AND",
	TokenOr:            "OR",
	TokenTernary:       "TERNARY",
	TokenTernaryElse:   "TERNARY_ELSE",
	TokenWhitespace:    "WHITESPACE",
	TokenLinebreak:     "LINEBREAK",
}

func (tt TokenType) String() string {
	if name, ok := tokenNames[tt]; ok {
		return name
	}
	return fmt.Sprintf("Token(%d)", tt)
}

// Token represents a single token in the template
type Token struct {
	Type     TokenType
	Value    string
	Line     int
	Column   int
	Position int
}

func (t Token) String() string {
	return fmt.Sprintf("%s('%s') at %d:%d", t.Type, t.Value, t.Line, t.Column)
}

// TokenStream represents a stream of tokens
type TokenStream struct {
	tokens []Token
	pos    int
}

func NewTokenStream(tokens []Token) *TokenStream {
	return &TokenStream{
		tokens: tokens,
		pos:    0,
	}
}

func (ts *TokenStream) Next() Token {
	if ts.pos >= len(ts.tokens) {
		return Token{Type: TokenEOF}
	}
	token := ts.tokens[ts.pos]
	ts.pos++
	return token
}

func (ts *TokenStream) Peek() Token {
	if ts.pos >= len(ts.tokens) {
		return Token{Type: TokenEOF}
	}
	return ts.tokens[ts.pos]
}

func (ts *TokenStream) PeekN(n int) Token {
	if ts.pos+n >= len(ts.tokens) {
		return Token{Type: TokenEOF}
	}
	return ts.tokens[ts.pos+n]
}

func (ts *TokenStream) Consume(expected TokenType) (Token, error) {
	token := ts.Next()
	if token.Type != expected {
		return token, fmt.Errorf("expected %s, got %s at %d:%d",
			expected, token.Type, token.Line, token.Column)
	}
	return token, nil
}

// Expect consumes and returns a token, failing if it doesn't match the expected type
func (ts *TokenStream) Expect(expectedType TokenType) (Token, error) {
	return ts.Consume(expectedType)
}

func (ts *TokenStream) Eof() bool {
	return ts.Peek().Type == TokenEOF
}

// ExpectNamed consumes and returns a token, failing if it doesn't match the expected type and value
func (ts *TokenStream) ExpectNamed(expectedType TokenType, expectedValue string) (Token, error) {
	token := ts.Peek()
	if token.Type == expectedType && token.Value == expectedValue {
		return ts.Next(), nil
	}
	return token, fmt.Errorf("expected %s %q, got %s at %d:%d",
		expectedType, expectedValue, token.Type, token.Line, token.Column)
}

// ExpectNameValue consumes and returns a name token with specific value
func (ts *TokenStream) ExpectNameValue(expectedValue string) (Token, error) {
	return ts.ExpectNamed(TokenName, expectedValue)
}

// PeekNamed returns the current token without consuming it (alias for Peek)
func (ts *TokenStream) PeekNamed() Token {
	return ts.Peek()
}
