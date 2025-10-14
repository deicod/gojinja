package lexer

import (
	"fmt"
	"regexp"
	"testing"
)

func TestVariableStateProcessing(t *testing.T) {
	config := DefaultLexerConfig()
	lexer := NewLexer(config)

	// Test what the variable end pattern looks like
	fmt.Printf("=== Testing variable state patterns ===\n")
	variableEndPattern := lexer.buildVariableEndPattern()
	fmt.Printf("Variable end pattern: %q\n", variableEndPattern)

	// Test the variable end regex
	c := func(pattern string) *regexp.Regexp {
		return regexp.MustCompile(pattern)
	}

	varEndRegex := c(variableEndPattern)
	testContent := "name }}"
	fmt.Printf("Test content: %q\n", testContent)

	if varEndRegex.MatchString(testContent) {
		fmt.Printf("Variable end pattern matches!\n")
		loc := varEndRegex.FindStringIndex(testContent)
		fmt.Printf("  Found at index: %v\n", loc)
		fmt.Printf("  Matched text: %q\n", testContent[loc[0]:loc[1]])

		// Test if the pattern can find just the name part
		nameRegex := NameRegex
		if nameRegex.MatchString(testContent) {
			nameLoc := nameRegex.FindStringIndex(testContent)
			fmt.Printf("  Name regex matches at: %v\n", nameLoc)
			fmt.Printf("  Name text: %q\n", testContent[nameLoc[0]:nameLoc[1]])
		}
	}

	// Test step by step tokenization starting in variable state
	fmt.Printf("\n=== Step by step variable state processing ===\n")

	// Test just "name" first - this should fail with unclosed construct error
	fmt.Printf("Testing 'name' in variable state:\n")
	stream1, err := lexer.Tokenize("name", "test", "test.html", "variable_begin")
	if err != nil {
		fmt.Printf("  Expected error: %v\n", err)
	} else {
		t.Errorf("Expected error for unclosed variable, but got tokens:")
		for {
			token := stream1.Next()
			if token.Type == TokenEOF {
				break
			}
			fmt.Printf("  %s\n", token)
		}
	}

	// Test "name }}"
	fmt.Printf("\nTesting 'name }}' in variable state:\n")
	stream2, err := lexer.Tokenize("name }}", "test", "test.html", "variable_begin")
	if err != nil {
		t.Fatalf("Failed: %v", err)
	}
	for {
		token := stream2.Next()
		if token.Type == TokenEOF {
			break
		}
		fmt.Printf("  %s\n", token)
	}

	// Test what the actual variable state rules look like
	fmt.Printf("\n=== Variable state rules ===\n")
	variableRules := lexer.rules[StateVariableBegin]
	for i, rule := range variableRules {
		fmt.Printf("Rule %d: Pattern='%s' Tokens='%v'\n", i, rule.Regex.String(), rule.Tokens)

		// Test if this rule matches our content
		testStr := "name }}"
		if rule.Regex.MatchString(testStr) {
			loc := rule.Regex.FindStringSubmatchIndex(testStr)
			if loc != nil && loc[0] == 0 {
				fmt.Printf("  -> Matches at start! match_len=%d\n", loc[1])
				fmt.Printf("  -> Matched text: %q\n", testStr[0:loc[1]])
			}
		}
	}
}