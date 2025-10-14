package lexer

import (
	"fmt"
	"testing"
)

func TestDebugLexing(t *testing.T) {
	config := DefaultLexerConfig()
	lexer := NewLexer(config)

	// Test a simple case: just "name"
	fmt.Printf("Name regex matches 'name': %v\n", NameRegex.FindStringSubmatch("name"))

	// Test a simple variable with no spaces
	variableTemplate2 := "{{name}}"
	fmt.Printf("\nTesting variable template (no spaces): %q\n", variableTemplate2)

	variableStream2, err := lexer.Tokenize(variableTemplate2, "test", "test.html", "")
	if err != nil {
		t.Fatalf("Variable tokenize failed: %v", err)
	}

	fmt.Println("\nVariable tokens (no spaces):")
	for {
		token := variableStream2.Next()
		if token.Type == TokenEOF {
			break
		}
		fmt.Printf("  %s\n", token)
	}

	// Test just variable content
	variableTemplate := "{{ name }}"
	fmt.Printf("\nTesting variable template: %q\n", variableTemplate)

	variableStream, err := lexer.Tokenize(variableTemplate, "test", "test.html", "")
	if err != nil {
		t.Fatalf("Variable tokenize failed: %v", err)
	}

	fmt.Println("\nVariable tokens:")
	for {
		token := variableStream.Next()
		if token.Type == TokenEOF {
			break
		}
		fmt.Printf("  %s\n", token)
	}
}

func TestDebugVariableStateOnly(t *testing.T) {
	config := DefaultLexerConfig()
	lexer := NewLexer(config)

	// Test tokenizing "name }}" starting in variable state
	fmt.Printf("\n=== Testing 'name }}' in variable state ===\n")
	stream, err := lexer.Tokenize("name }}", "test", "test.html", "variable_begin")
	if err != nil {
		t.Fatalf("Variable state tokenize failed: %v", err)
	}

	fmt.Println("\nTokens from variable state ('name }}'):")
	for {
		token := stream.Next()
		if token.Type == TokenEOF {
			break
		}
		fmt.Printf("  %s\n", token)
	}

	// Test tokenizing " name }}" starting in variable state
	fmt.Printf("\n=== Testing ' name }}' in variable state ===\n")
	stream2, err := lexer.Tokenize(" name }}", "test", "test.html", "variable_begin")
	if err != nil {
		t.Fatalf("Variable state tokenize failed: %v", err)
	}

	fmt.Println("\nTokens from variable state (' name }}'):")
	for {
		token := stream2.Next()
		if token.Type == TokenEOF {
			break
		}
		fmt.Printf("  %s\n", token)
	}

	// Test more complex expressions
	fmt.Printf("\n=== Testing 'user.name + 42 }}' in variable state ===\n")
	stream3, err := lexer.Tokenize("user.name + 42 }}", "test", "test.html", "variable_begin")
	if err != nil {
		t.Fatalf("Variable state tokenize failed: %v", err)
	}

	fmt.Println("\nTokens from variable state ('user.name + 42 }}'):")
	for {
		token := stream3.Next()
		if token.Type == TokenEOF {
			break
		}
		fmt.Printf("  %s\n", token)
	}

	// Test string and operators
	fmt.Printf("\n=== Testing '\"hello\" + world }}' in variable state ===\n")
	stream4, err := lexer.Tokenize("\"hello\" + world }}", "test", "test.html", "variable_begin")
	if err != nil {
		t.Fatalf("Variable state tokenize failed: %v", err)
	}

	fmt.Println("\nTokens from variable state ('\"hello\" + world }}'):")
	for {
		token := stream4.Next()
		if token.Type == TokenEOF {
			break
		}
		fmt.Printf("  %s\n", token)
	}
}