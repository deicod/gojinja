package lexer

import (
	"fmt"
	"regexp"
	"testing"
)

func TestRootStateProcessing(t *testing.T) {
	config := DefaultLexerConfig()
	lexer := NewLexer(config)

	// Test what rootParts pattern actually matches
	fmt.Printf("=== Testing root parts pattern ===\n")
	rootParts := lexer.buildRootParts()
	fmt.Printf("Root parts pattern: %q\n", rootParts)

	// Test if the root parts regex can find our delimiters
	c := func(pattern string) *regexp.Regexp {
		return regexp.MustCompile(pattern)
	}

	rootRegex := c("(.*?)(" + rootParts + ")")

	testString := "Hello {{ user.name }}, world"
	fmt.Printf("Test string: %q\n", testString)

	matches := rootRegex.FindStringSubmatch(testString)
	fmt.Printf("Root regex matches: %v\n", matches)
	if len(matches) > 0 {
		for i, match := range matches {
			fmt.Printf("  Group %d: %q\n", i, match)
		}
	}

	// Test what the individual delimiter patterns look like
	e := regexp.QuoteMeta
	variableBegin := e(config.Delimiters.VariableStart)
	fmt.Printf("Variable begin pattern: %q\n", variableBegin)

	// Test if our delimiter matches
	varRegex := regexp.MustCompile(variableBegin)
	if varRegex.MatchString(testString) {
		fmt.Printf("Variable begin matches test string!\n")
		loc := varRegex.FindStringIndex(testString)
		fmt.Printf("  Found at index: %v\n", loc)
		fmt.Printf("  Matched text: %q\n", testString[loc[0]:loc[1]])
	}

	// Now test a simple case
	fmt.Printf("\n=== Testing simple root state processing ===\n")
	simpleTemplate := "Hello {{ name }}"
	fmt.Printf("Simple template: %q\n", simpleTemplate)

	simpleStream, err := lexer.Tokenize(simpleTemplate, "test", "test.html", "")
	if err != nil {
		t.Fatalf("Simple template tokenize failed: %v", err)
	}

	fmt.Println("\nSimple template tokens:")
	for {
		token := simpleStream.Next()
		if token.Type == TokenEOF {
			break
		}
		fmt.Printf("  %s\n", token)
	}
}