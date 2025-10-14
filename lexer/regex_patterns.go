package lexer

import (
	"regexp"
	"strings"
)

// Precompiled regular expressions for tokenizing
var (
	// Whitespace detection
	WhitespaceRegex = regexp.MustCompile(`\s+`)

	// Newline detection (handles \r\n, \r, \n)
	NewlineRegex = regexp.MustCompile(`(\r\n|\r|\n)`)

	// String literals (single and double quoted, with escape sequences)
	StringRegex = regexp.MustCompile(`('([^'\\]*(?:\\.[^'\\]*)*)'|"([^"\\]*(?:\\.[^"\\]*)*)")`)

	// Integer literals (binary, octal, hex, decimal)
	IntegerRegex = regexp.MustCompile(`(?i)(0b(_?[0-1])+|0o(_?[0-7])+|0x(_?[\da-f])+|[1-9](_?\d)*|0(_?0)*)`)

	// Float literals (Go doesn't support lookbehind, so we use a different approach)
	FloatRegex = regexp.MustCompile(`(?i)(?:(?:^|[^.])((\d+_)*\d+((\.(\d+_)*\d+)?e[+\-]?(\d+_)*\d+|\.(\d+_)*\d+)))`)

	// Identifier/names
	NameRegex = regexp.MustCompile(`[a-zA-Z_][a-zA-Z0-9_]*`)

	// Operators (sorted by length for correct matching)
	OperatorPatterns = []string{
		"//", "**", "==", "!=", ">=", "<=", "=", "+", "-", "*", "/", "%", "~",
		"[", "]", "(", ")", ">", "<", ".", ":", "|", ",", ";", "{", "}",
	}

	// Combined operator regex
	OperatorRegex = func() *regexp.Regexp {
		var escaped []string
		for _, op := range OperatorPatterns {
			escaped = append(escaped, regexp.QuoteMeta(op))
		}
		// Sort by length descending to match longer operators first
		for i := 0; i < len(escaped); i++ {
			for j := i + 1; j < len(escaped); j++ {
				if len(escaped[i]) < len(escaped[j]) {
					escaped[i], escaped[j] = escaped[j], escaped[i]
				}
			}
		}
		pattern := "(" + strings.Join(escaped, "|") + ")"
		return regexp.MustCompile(pattern)
	}()
)

// Environment-specific delimiters
type Delimiters struct {
	BlockStart     string
	BlockEnd       string
	VariableStart  string
	VariableEnd    string
	CommentStart   string
	CommentEnd     string
	LineStatement  string
	LineComment    string
}

func DefaultDelimiters() Delimiters {
	return Delimiters{
		BlockStart:    "{%",
		BlockEnd:      "%}",
		VariableStart: "{{",
		VariableEnd:   "}}",
		CommentStart:  "{#",
		CommentEnd:    "#}",
	}
}

// BuildEnvironmentRules is deprecated - use the internal rule building in Lexer
func BuildEnvironmentRules(delims Delimiters) []*Rule {
	// This function is kept for compatibility but the actual rule building
	// is now handled internally by the Lexer
	return []*Rule{}
}