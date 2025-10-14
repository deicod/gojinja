package lexer

import (
	"fmt"
	"regexp"
	"testing"
)

func TestVariableState(t *testing.T) {
	config := DefaultLexerConfig()
	lexer := NewLexer(config)

	// Test if name regex works
	fmt.Printf("Name regex matches 'name': %v\n", NameRegex.FindStringSubmatch("name"))

	// Test if variable end regex works
	fmt.Printf("Variable end regex matches '}}': %v\n", regexp.MustCompile(lexer.buildVariableEndPattern()).FindStringSubmatch("}}"))

	// Test direct tokenization in variable state
	// Start in variable state
	stream, err := lexer.Tokenize("name }}", "test", "test.html", "variable_begin")
	if err != nil {
		t.Fatalf("Variable state tokenize failed: %v", err)
	}

	fmt.Println("\nVariable state tokens (with end):")
	for {
		token := stream.Next()
		if token.Type == TokenEOF {
			break
		}
		fmt.Printf("  %s\n", token)
	}
}