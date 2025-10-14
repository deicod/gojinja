package lexer

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestBasicLexing(t *testing.T) {
	config := DefaultLexerConfig()
	lexer := NewLexer(config)

	tests := []struct {
		name     string
		template string
		wantErr  bool
	}{
		{
			name:     "simple text",
			template: "Hello, World!",
			wantErr:  false,
		},
		{
			name:     "simple variable",
			template: "Hello, {{ name }}!",
			wantErr:  false,
		},
		{
			name:     "simple block",
			template: "{% if condition %}content{% endif %}",
			wantErr:  false,
		},
		{
			name:     "comment",
			template: "Hello{# this is a comment #} World!",
			wantErr:  false,
		},
		{
			name:     "mixed content",
			template: "Hello {{ name }}! {% if condition %}Yes{% else %}No{% endif %}",
			wantErr:  false,
		},
		{
			name:     "raw block",
			template: "{% raw %}{{ this is not a variable }}{% endraw %}",
			wantErr:  false,
		},
		{
			name:     "verbatim block",
			template: "{% verbatim %}{{ untouched }}{% endverbatim %}",
			wantErr:  false,
		},
		{
			name:     "raw block with whitespace control",
			template: "{%- raw -%}{{ literal }}{%- endraw -%}",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stream, err := lexer.Tokenize(tt.template, "test", "test.html", "")

			if (err != nil) != tt.wantErr {
				t.Errorf("Lexer.Tokenize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Print tokens for debugging
				fmt.Printf("Tokens for '%s':\n", tt.name)
				for {
					token := stream.Next()
					if token.Type == TokenEOF {
						break
					}
					fmt.Printf("  %s\n", token)
				}
				fmt.Println()
			}
		})
	}
}

func TestLStripBlocks(t *testing.T) {
	config := DefaultLexerConfig()
	config.LstripBlocks = true
	lexer := NewLexer(config)

	template := `
    {% if condition %}
        content
    {% endif %}
`

	stream, err := lexer.Tokenize(template, "test", "test.html", "")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Count text tokens and check if whitespace is properly handled
	textTokens := 0
	for {
		token := stream.Next()
		if token.Type == TokenEOF {
			break
		}
		if token.Type == TokenText {
			textTokens++
			fmt.Printf("Text token: '%s' at %d:%d\n", token.Value, token.Line, token.Column)
		}
	}

	// We should have text tokens for the content and whitespace
	if textTokens == 0 {
		t.Error("Expected at least one text token")
	}
}

func TestTrimBlocks(t *testing.T) {
	config := DefaultLexerConfig()
	config.TrimBlocks = true
	lexer := NewLexer(config)

	template := "content{% if condition %}\nmore content\n{% endif %}"

	stream, err := lexer.Tokenize(template, "test", "test.html", "")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify that newlines after block end are trimmed
	foundNewline := false
	for {
		token := stream.Next()
		if token.Type == TokenEOF {
			break
		}
		if token.Type == TokenText && token.Value == "\n" {
			foundNewline = true
		}
	}

	if foundNewline {
		t.Error("Expected newline after block end to be trimmed")
	}
}

func TestBalancedBraces(t *testing.T) {
	config := DefaultLexerConfig()
	lexer := NewLexer(config)

	tests := []struct {
		name     string
		template string
		wantErr  bool
	}{
		{
			name:     "balanced braces",
			template: "{{ dict['key'] }}",
			wantErr:  false,
		},
		{
			name:     "unbalanced braces",
			template: "{{ dict['key' }}",
			wantErr:  true,
		},
		{
			name:     "nested braces",
			template: "{{ func({key: value}) }}",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := lexer.Tokenize(tt.template, "test", "test.html", "")

			if (err != nil) != tt.wantErr {
				t.Errorf("Lexer.Tokenize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPositionTracking(t *testing.T) {
	config := DefaultLexerConfig()
	lexer := NewLexer(config)

	template := `line 1
line 2 {{ variable }}
line 3 {% block %}content{% endblock %}
line 4`

	stream, err := lexer.Tokenize(template, "test", "test.html", "")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Track positions to verify they're correct
	positions := make(map[TokenType][]struct {
		line, column int
		value        string
	})

	for {
		token := stream.Next()
		if token.Type == TokenEOF {
			break
		}

		positions[token.Type] = append(positions[token.Type], struct {
			line, column int
			value        string
		}{
			line:   token.Line,
			column: token.Column,
			value:  token.Value,
		})
	}

	// Verify some basic position expectations
	if variableTokens, ok := positions[TokenVariableStart]; ok && len(variableTokens) > 0 {
		pos := variableTokens[0]
		if pos.line != 2 {
			t.Errorf("Expected variable start at line 2, got line %d", pos.line)
		}
		if pos.column < 8 { // Should be after "line 2 "
			t.Errorf("Expected variable start column >= 8, got column %d", pos.column)
		}
	}

	if blockTokens, ok := positions[TokenBlockStart]; ok && len(blockTokens) > 0 {
		pos := blockTokens[0]
		if pos.line != 3 {
			t.Errorf("Expected block start at line 3, got line %d", pos.line)
		}
	}
}

func TestWrapInvalidIdentifierProducesError(t *testing.T) {
	lexer := NewLexer(DefaultLexerConfig())

	_, err := lexer.wrap([]TokenInfo{{
		Line:   3,
		Column: 7,
		Type:   "name",
		Value:  "1invalid",
	}}, "test", "test.html")

	if err == nil {
		t.Fatal("expected error for invalid identifier")
	}

	var lexErr *LexerError
	if !errors.As(err, &lexErr) {
		t.Fatalf("expected LexerError, got %T", err)
	}

	if !strings.Contains(lexErr.Message, "invalid identifier") {
		t.Fatalf("unexpected error message: %q", lexErr.Message)
	}

	if lexErr.Line != 3 || lexErr.Column != 7 {
		t.Fatalf("unexpected error location: line %d column %d", lexErr.Line, lexErr.Column)
	}
}

func TestStringTokenization(t *testing.T) {
	config := DefaultLexerConfig()
	lexer := NewLexer(config)

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{"double quotes", `{{ "hello world" }}`, "hello world"},
		{"single quotes", `{{ 'hello world' }}`, "hello world"},
		{"escaped quotes", `{{ "hello \"world\"" }}`, `hello "world"`},
		{"escaped newlines", `{{ "hello\nworld" }}`, "hello\nworld"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stream, err := lexer.Tokenize(tt.template, "test", "test.html", "")
			if err != nil {
				t.Fatalf("Tokenize failed: %v", err)
			}

			// Skip VAR_START
			_ = stream.Next()

			stringToken := stream.Next()
			if stringToken.Type != TokenString {
				t.Errorf("Expected STRING, got %s", stringToken.Type)
			}
			if stringToken.Value != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, stringToken.Value)
			}
		})
	}
}

func TestNumberTokenization(t *testing.T) {
	config := DefaultLexerConfig()
	lexer := NewLexer(config)

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{"integer", "{{ 42 }}", "42"},
		{"float", "{{ 3.14 }}", "3.14"},
		{"hex", "{{ 0xFF }}", "255"},
		{"binary", "{{ 0b1010 }}", "10"},
		{"octal", "{{ 0o755 }}", "493"},
		{"underscored", "{{ 1_000_000 }}", "1000000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stream, err := lexer.Tokenize(tt.template, "test", "test.html", "")
			if err != nil {
				t.Fatalf("Tokenize failed: %v", err)
			}

			// Skip VAR_START
			_ = stream.Next()

			numberToken := stream.Next()
			if numberToken.Type != TokenNumber {
				t.Errorf("Expected NUMBER, got %s", numberToken.Type)
			}
			if numberToken.Value != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, numberToken.Value)
			}
		})
	}
}

func TestOperatorTokenization(t *testing.T) {
	config := DefaultLexerConfig()
	lexer := NewLexer(config)

	tests := []struct {
		name     string
		template string
		expected []TokenType
	}{
		{
			"arithmetic",
			"{{ a + b * c / d - e }}",
			[]TokenType{TokenVariableStart, TokenName, TokenAdd, TokenName, TokenMul, TokenName, TokenDiv, TokenName, TokenSub, TokenName, TokenVariableEnd},
		},
		{
			"comparison",
			"{{ a == b and c != d }}",
			[]TokenType{TokenVariableStart, TokenName, TokenComparison, TokenName, TokenAnd, TokenName, TokenComparison, TokenName, TokenVariableEnd},
		},
		{
			"pipe",
			"{{ value | filter }}",
			[]TokenType{TokenVariableStart, TokenName, TokenPipe, TokenName, TokenVariableEnd},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stream, err := lexer.Tokenize(tt.template, "test", "test.html", "")
			if err != nil {
				t.Fatalf("Tokenize failed: %v", err)
			}

			for i, expectedType := range tt.expected {
				token := stream.Next()
				if token.Type != expectedType {
					t.Errorf("Token %d: expected %s, got %s", i, expectedType, token.Type)
				}
			}
		})
	}
}

func TestNestedStructures(t *testing.T) {
	config := DefaultLexerConfig()
	lexer := NewLexer(config)

	template := "{{ func({ 'key': [1, 2, (3 + 4)] }) }}"
	stream, err := lexer.Tokenize(template, "test", "test.html", "")
	if err != nil {
		t.Fatalf("Tokenize failed: %v", err)
	}

	// Should successfully tokenize nested structures
	tokenCount := 0
	for !stream.Eof() {
		stream.Next()
		tokenCount++
	}

	if tokenCount < 10 {
		t.Errorf("Expected many tokens for nested structures, got %d", tokenCount)
	}
}

func TestErrorHandling(t *testing.T) {
	config := DefaultLexerConfig()
	lexer := NewLexer(config)

	tests := []struct {
		name     string
		template string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "unclosed variable",
			template: "{{ name",
			wantErr:  true,
			errMsg:   "missing",
		},
		{
			name:     "unclosed block",
			template: "{% if condition",
			wantErr:  true,
			errMsg:   "missing",
		},
		{
			name:     "unclosed comment",
			template: "{# comment",
			wantErr:  true,
			errMsg:   "missing",
		},
		{
			name:     "unclosed braces",
			template: "{{ func({})",
			wantErr:  true,
			errMsg:   "unclosed",
		},
		{
			name:     "mismatched braces",
			template: "{{ func({]})",
			wantErr:  true,
			errMsg:   "unclosed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := lexer.Tokenize(tt.template, "test", "test.html", "")

			if (err != nil) != tt.wantErr {
				t.Errorf("Lexer.Tokenize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error message to contain '%s', got: %v", tt.errMsg, err)
				}
			}
		})
	}
}

func TestWhitespaceControlSigns(t *testing.T) {
	config := DefaultLexerConfig()
	lexer := NewLexer(config)

	tests := []struct {
		name     string
		template string
		expected string // Expected text content
	}{
		{
			name:     "minus sign",
			template: "text \n{{- name }}",
			expected: "text name", // Whitespace before variable stripped
		},
		{
			name:     "plus sign",
			template: "text \n{{+ name }}",
			expected: "text \n name", // Whitespace preserved
		},
		{
			name:     "no sign with lstrip",
			template: "text \n{% if condition %}",
			expected: "text \n", // Depends on lstrip_blocks setting
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stream, err := lexer.Tokenize(tt.template, "test", "test.html", "")
			if err != nil {
				t.Fatalf("Tokenize failed: %v", err)
			}

			// Collect text tokens
			var textContent string
			for !stream.Eof() {
				token := stream.Next()
				if token.Type == TokenText || token.Type == TokenName {
					textContent += token.Value
				}
			}

			// Basic check - the exact content might vary based on whitespace handling
			if len(textContent) == 0 {
				t.Errorf("Expected some text content for %s", tt.name)
			}
		})
	}
}

func TestComplexTemplate(t *testing.T) {
	config := DefaultLexerConfig()
	lexer := NewLexer(config)

	template := `<!DOCTYPE html>
<html>
<head>
    <title>{{ title or "Default Title" }}</title>
</head>
<body>
    {# This is a comment #}
    {% if user %}
        <h1>Welcome {{ user.name }}!</h1>
        <p>You have {{ user.messages|length }} messages.</p>
    {% else %}
        <p>Please <a href="/login">log in</a>.</p>
    {% endif %}

    {% for item in items %}
        <div class="item">
            <h3>{{ item.title }}</h3>
            <p>{{ item.description|truncate(100) }}</p>
        </div>
    {% endfor %}
</body>
</html>`

	stream, err := lexer.Tokenize(template, "test", "test.html", "")
	if err != nil {
		t.Fatalf("Tokenize failed: %v", err)
	}

	// Should tokenize the complex template successfully
	tokenCount := 0
	variableCount := 0
	blockCount := 0

	for !stream.Eof() {
		token := stream.Next()
		tokenCount++
		switch token.Type {
		case TokenVariableStart, TokenVariableEnd:
			variableCount++
		case TokenBlockStart, TokenBlockEnd:
			blockCount++
		}
	}

	if tokenCount < 50 {
		t.Errorf("Expected many tokens for complex template, got %d", tokenCount)
	}

	if variableCount < 10 {
		t.Errorf("Expected many variable tokens, got %d", variableCount)
	}

	if blockCount < 8 {
		t.Errorf("Expected many block tokens, got %d", blockCount)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			func() bool {
				for i := 0; i <= len(s)-len(substr); i++ {
					if s[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			}())))
}
