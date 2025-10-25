package lexer

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// LexerError represents a lexing error
type LexerError struct {
	Message string
	Line    int
	Column  int
	Pos     int
}

func (e LexerError) Error() string {
	return fmt.Sprintf("%s at line %d, column %d", e.Message, e.Line, e.Column)
}

// Rule represents a lexing rule with regex pattern and associated token(s)
type Rule struct {
	Regex    *regexp.Regexp
	Tokens   interface{} // string or []string
	NewState *string
}

// LexerState represents the current parsing state
type LexerState string

const (
	StateRoot          LexerState = "root"
	StateVariableBegin LexerState = "variable_begin"
	StateBlockBegin    LexerState = "block_begin"
	StateCommentBegin  LexerState = "comment_begin"
	StateRawBegin      LexerState = "raw_begin"
	StateLineStatement LexerState = "linestatement_begin"
	StateLineComment   LexerState = "linecomment_begin"
)

// LexerConfig holds configuration for the lexer
type LexerConfig struct {
	Delimiters          Delimiters
	TrimBlocks          bool
	LstripBlocks        bool
	NewlineSequence     string
	KeepTrailingNewline bool
}

func DefaultLexerConfig() LexerConfig {
	return LexerConfig{
		Delimiters:          DefaultDelimiters(),
		TrimBlocks:          false,
		LstripBlocks:        false,
		NewlineSequence:     "\n",
		KeepTrailingNewline: false,
	}
}

// Lexer implements the Jinja2 template lexer
type Lexer struct {
	config LexerConfig
	rules  map[LexerState][]*Rule
}

// NewLexer creates a new lexer with the given configuration
func NewLexer(config LexerConfig) *Lexer {
	lexer := &Lexer{
		config: config,
		rules:  make(map[LexerState][]*Rule),
	}
	lexer.buildRules()
	return lexer
}

// buildRules constructs the lexing rules for different states
func (l *Lexer) buildRules() {
	// Tag rules for use inside blocks/variables
	tagRules := l.buildTagRules()

	// Build root parts regex
	rootParts := l.buildRootParts()

	// Block suffix regex for trim blocks
	blockSuffix := ""
	if l.config.TrimBlocks {
		blockSuffix = "\\n?"
	}

	// Compile regex patterns
	c := func(pattern string) *regexp.Regexp {
		return regexp.MustCompile(pattern)
	}

	// Root state rules - process text before delimiters
	l.rules[StateRoot] = []*Rule{
		{
			Regex:    c("(.*?)(" + rootParts + ")"),
			Tokens:   "#bygroup",
			NewState: strPtr("#bygroup"),
		},
		{
			Regex:    c("[^{]+"),
			Tokens:   "data",
			NewState: nil,
		},
		{
			Regex:    c("."),
			Tokens:   "data",
			NewState: nil,
		},
	}

	// Comment state rules
	commentEndPattern := l.buildCommentEndPattern(blockSuffix)
	l.rules[StateCommentBegin] = []*Rule{
		{
			Regex:    c("(.*?)(" + commentEndPattern + ")"),
			Tokens:   []string{"comment"},
			NewState: strPtr("#pop"),
		},
		{
			Regex:    c("(.)"),
			Tokens:   []string{"error:missing_end_comment"},
			NewState: nil,
		},
	}

	// Block state rules
	blockEndPattern := l.buildBlockEndPattern(blockSuffix)
	blockRules := []*Rule{
		{
			Regex:    c(blockEndPattern),
			Tokens:   "block_end",
			NewState: strPtr("#pop"),
		},
	}
	blockRules = append(blockRules, tagRules...)
	l.rules[StateBlockBegin] = blockRules

	// Variable state rules
	variableEndPattern := l.buildVariableEndPattern()
	variableRules := []*Rule{
		{
			Regex:    c(variableEndPattern),
			Tokens:   "variable_end",
			NewState: strPtr("#pop"),
		},
	}
	variableRules = append(variableRules, tagRules...)
	l.rules[StateVariableBegin] = variableRules

	// Raw block state rules
	rawEndPattern := l.buildRawEndPattern(blockSuffix)
	l.rules[StateRawBegin] = []*Rule{
		{
			Regex:    c("(.*?)(" + rawEndPattern + ")"),
			Tokens:   []string{"raw_data"},
			NewState: strPtr("#pop"),
		},
		{
			Regex:    c("(.)"),
			Tokens:   []string{"error:missing_end_raw"},
			NewState: nil,
		},
	}

	// Line statement state rules
	linestatementRules := []*Rule{
		{
			Regex:    c("\\s*(\\n|$)"),
			Tokens:   "block_end",
			NewState: strPtr("#pop"),
		},
	}
	linestatementRules = append(linestatementRules, tagRules...)
	l.rules[StateLineStatement] = linestatementRules

	// Line comment state rules
	l.rules[StateLineComment] = []*Rule{
		{
			Regex:    c("(.*?)()"),
			Tokens:   []string{"linecomment", "linecomment_end"},
			NewState: strPtr("#pop"),
		},
	}
}

// buildTagRules creates rules for parsing content within blocks/variables
func (l *Lexer) buildTagRules() []*Rule {
	return []*Rule{
		{
			Regex:    WhitespaceRegex,
			Tokens:   "whitespace",
			NewState: nil,
		},
		{
			Regex:    FloatRegex,
			Tokens:   "float",
			NewState: nil,
		},
		{
			Regex:    IntegerRegex,
			Tokens:   "integer",
			NewState: nil,
		},
		{
			Regex:    NameRegex,
			Tokens:   "name",
			NewState: nil,
		},
		{
			Regex:    StringRegex,
			Tokens:   "string",
			NewState: nil,
		},
		{
			Regex:    OperatorRegex,
			Tokens:   "operator",
			NewState: nil,
		},
	}
}

// buildRootParts constructs the regex pattern for root-level delimiters
func (l *Lexer) buildRootParts() string {
	e := regexp.QuoteMeta

	// Simple patterns for delimiters without complex whitespace control
	variableBegin := e(l.config.Delimiters.VariableStart)
	blockBegin := e(l.config.Delimiters.BlockStart)
	commentBegin := e(l.config.Delimiters.CommentStart)

	// Raw/verbatim block pattern - support whitespace-control strip markers around tag name
	rawBegin := e(l.config.Delimiters.BlockStart) + "\\s*[-+]?\\s*(?:raw|verbatim)\\s*[-+]?\\s*" + e(l.config.Delimiters.BlockEnd)

	// Combine in priority order
	parts := []string{
		rawBegin,
		commentBegin,
		blockBegin,
		variableBegin,
	}

	// Line statement pattern
	// Line statements/comments handled explicitly in tokeniter to preserve newline semantics

	return strings.Join(parts, "|")
}

// buildCommentEndPattern builds the regex pattern for comment end
func (l *Lexer) buildCommentEndPattern(blockSuffix string) string {
	e := regexp.QuoteMeta
	return e(l.config.Delimiters.CommentEnd)
}

// buildBlockEndPattern builds the regex pattern for block end
func (l *Lexer) buildBlockEndPattern(blockSuffix string) string {
	e := regexp.QuoteMeta
	return e(l.config.Delimiters.BlockEnd)
}

// buildVariableEndPattern builds the regex pattern for variable end
func (l *Lexer) buildVariableEndPattern() string {
	e := regexp.QuoteMeta
	return e(l.config.Delimiters.VariableEnd)
}

// buildRawEndPattern builds the regex pattern for raw block end
func (l *Lexer) buildRawEndPattern(blockSuffix string) string {
	e := regexp.QuoteMeta
	return fmt.Sprintf("(?:%s(\\-|\\+|))\\s*end(?:raw|verbatim)\\s*(?:\\+%s|\\-%s\\s*|%s%s)",
		e(l.config.Delimiters.BlockStart),
		e(l.config.Delimiters.BlockEnd),
		e(l.config.Delimiters.BlockEnd),
		e(l.config.Delimiters.BlockEnd),
		blockSuffix)
}

// Tokenize tokenizes the given source string and returns a stream of tokens
func (l *Lexer) Tokenize(source, name, filename string, initialState LexerState) (*TokenStream, error) {
	tokens, err := l.tokeniter(source, name, filename, initialState)
	if err != nil {
		return nil, err
	}

	wrappedTokens, err := l.wrap(tokens, name, filename)
	if err != nil {
		return nil, err
	}
	return NewTokenStream(wrappedTokens), nil
}

// tokeniter implements the core tokenization logic based on Python's tokeniter
func (l *Lexer) tokeniter(source, name, filename string, initialState LexerState) ([]TokenInfo, error) {
	// Normalize newlines to ensure consistent line counting
	source = l.normalizeNewlines(source)

	// Handle keep_trailing_newline configuration
	if !l.config.KeepTrailingNewline && strings.HasSuffix(source, "\n") {
		source = strings.TrimSuffix(source, "\n")
	}

	// Initialize lexer state
	pos := 0
	lineno := 1
	column := 1
	sourceLen := len(source)

	// State stack for nested parsing
	stack := []LexerState{StateRoot}
	if initialState != "" && initialState != StateRoot {
		if initialState != StateVariableBegin && initialState != StateBlockBegin {
			return nil, fmt.Errorf("invalid initial state: %s", initialState)
		}
		stack = append(stack, initialState)
	}

	// Get rules for current state
	statetokens := l.rules[stack[len(stack)-1]]

	// Stack for tracking balanced braces/parentheses/brackets
	balancingStack := []rune{}

	// Track if we're at the start of a line for lstrip_blocks
	lineStarting := true
	suppressNextNewline := false

	var tokens []TokenInfo

	// Add infinite loop protection - limit maximum iterations
	maxIterations := sourceLen * 10 // Reasonable upper bound
	iterations := 0

	for pos < sourceLen && iterations < maxIterations {
		matched := false

		// Handle line statements/comments explicitly when at line start
		currentState := stack[len(stack)-1]
		if currentState == StateRoot && lineStarting {
			if wsLen, ok := matchLinePrefix(source[pos:], l.config.Delimiters.LineStatement); ok {
				prefix := l.config.Delimiters.LineStatement
				whitespaceRunes := utf8.RuneCountInString(source[pos : pos+wsLen])
				tokenCol := column + whitespaceRunes
				tokens = append(tokens, TokenInfo{
					Line:   lineno,
					Column: tokenCol,
					Type:   "block_begin",
					Value:  prefix,
				})

				advance := wsLen + len(prefix)
				pos += advance
				column += whitespaceRunes + utf8.RuneCountInString(prefix)
				stack = append(stack, StateLineStatement)
				statetokens = l.rules[stack[len(stack)-1]]
				lineStarting = false
				matched = true
				iterations++
				continue
			}

			if wsLen, ok := matchLinePrefix(source[pos:], l.config.Delimiters.LineComment); ok {
				prefix := l.config.Delimiters.LineComment
				commentStart := pos + wsLen + len(prefix)
				commentLen := 0
				if commentStart < len(source) {
					if idx := strings.Index(source[commentStart:], "\n"); idx >= 0 {
						commentLen = idx
					} else {
						commentLen = len(source) - commentStart
					}
				}

				consumed := wsLen + len(prefix) + commentLen
				column += utf8.RuneCountInString(source[pos : pos+consumed])
				pos += consumed
				lineStarting = false
				if pos < len(source) && source[pos] == '\n' {
					pos++
					lineno++
					column = 1
					lineStarting = true
				}
				matched = true
				iterations++
				continue
			}
		}

		if currentState == StateRoot {
			boundary := findLineControlBoundary(
				source[pos:],
				l.config.Delimiters.LineStatement,
				l.config.Delimiters.LineComment,
				l.config.Delimiters.VariableStart,
				l.config.Delimiters.BlockStart,
				l.config.Delimiters.CommentStart,
			)
			if boundary > 0 {
				rawSegment := source[pos : pos+boundary]
				outputSegment := rawSegment
				if suppressNextNewline && strings.HasSuffix(rawSegment, "\n") {
					outputSegment = strings.TrimSuffix(rawSegment, "\n")
					suppressNextNewline = false
				} else if suppressNextNewline {
					suppressNextNewline = false
				}

				if outputSegment != "" {
					tokens = append(tokens, TokenInfo{
						Line:   lineno,
						Column: column,
						Type:   "data",
						Value:  outputSegment,
					})
				}

				newlines := strings.Count(rawSegment, "\n")
				lineno += newlines
				if newlines > 0 {
					lastNewlinePos := strings.LastIndex(rawSegment, "\n")
					column = utf8.RuneCountInString(rawSegment[lastNewlinePos+1:]) + 1
				} else {
					column += utf8.RuneCountInString(rawSegment)
				}

				pos += boundary
				lineStarting = true
				matched = true
				iterations++
				continue
			}
		}

		// Emit standalone newlines to preserve output and reset line state
		if currentState == StateRoot && strings.HasPrefix(source[pos:], "\n") {
			prevLine := lineno
			prevColumn := column
			pos++
			lineno++
			column = 1
			if suppressNextNewline {
				suppressNextNewline = false
			} else {
				tokens = append(tokens, TokenInfo{
					Line:   prevLine,
					Column: prevColumn,
					Type:   "data",
					Value:  "\n",
				})
			}
			lineStarting = true
			matched = true
			iterations++
			continue
		}

		// Try each rule for the current state
		for _, rule := range statetokens {
			loc := rule.Regex.FindStringSubmatchIndex(source[pos:])
			if loc == nil || loc[0] != 0 {
				continue
			}

			matchStart := pos
			matchEnd := pos + loc[1]
			matchText := source[matchStart:matchEnd]

			// Process the match and generate tokens
			newTokens, strippedCount, err := l.processMatch(rule, source, loc, pos, lineno, column, filename, balancingStack, lineStarting)
			if err != nil {
				return nil, err
			}

			// Update position tracking using the full match text
			newlines := strings.Count(matchText, "\n")
			lineno += newlines + strippedCount
			if newlines > 0 {
				lastNewlinePos := strings.LastIndex(matchText, "\n")
				column = utf8.RuneCountInString(matchText[lastNewlinePos+1:]) + 1
			} else {
				column += utf8.RuneCountInString(matchText)
			}

			lineStarting = strings.HasSuffix(matchText, "\n")

			// Handle state transitions
			if rule.NewState != nil {
				if *rule.NewState == "#pop" {
					if len(stack) > 1 {
						stack = stack[:len(stack)-1]
					}
				} else if *rule.NewState == "#bygroup" {
					// Determine which group matched and transition to appropriate state
					newState := l.determineStateByGroup(rule, source[pos:])
					if newState != "" {
						stack = append(stack, newState)
						// For #bygroup, we need to adjust the position
						// to be after the delimiter, not the full match
						groups := rule.Regex.FindStringSubmatch(source[pos:])
						if len(groups) >= 2 {
							// Find which delimiter group matched
							for i := 2; i < len(groups); i++ {
								if groups[i] != "" {
									// The delimiter starts after the text content (group 1)
									// Position should be at the start of delimiter + delimiter length
									textLen := len(groups[1])  // Length of text before delimiter
									delimLen := len(groups[i]) // Length of matched delimiter
									pos = pos + textLen + delimLen
									break
								}
							}
						}
					} else {
						// If no specific state determined, stay in current state
						// This prevents popping to root prematurely
						// BUT we still need to advance position to avoid infinite loops
						pos = matchEnd
					}
				} else {
					// Explicit state transition
					stack = append(stack, LexerState(*rule.NewState))
				}
				statetokens = l.rules[stack[len(stack)-1]]
			}

			// Update balancing stack for operators and delimiters
			l.updateBalancingStackForMatch(&balancingStack, rule, matchText, lineno, column, name, filename)

			// Add generated tokens
			tokens = append(tokens, newTokens...)
			if l.config.TrimBlocks && currentState == StateBlockBegin {
				for _, tok := range newTokens {
					if tok.Type == "block_end" {
						suppressNextNewline = true
						break
					}
				}
			}
			if rule.NewState == nil || *rule.NewState != "#bygroup" {
				pos = matchEnd
			}
			matched = true
			iterations++
			break
		}

		if !matched {
			if pos >= sourceLen {
				break
			}
			// If no rule matched, we have an unexpected character
			rune, _ := utf8.DecodeRuneInString(source[pos:])
			return nil, &LexerError{
				Message: fmt.Sprintf("unexpected character %q", rune),
				Line:    lineno,
				Column:  column,
				Pos:     pos,
			}
		}
	}

	// Check if we exited due to infinite loop protection
	if iterations >= maxIterations {
		return nil, &LexerError{
			Message: "lexer infinite loop detected - possible regex or state machine issue",
			Line:    lineno,
			Column:  column,
			Pos:     pos,
		}
	}

	// Final validation - check for unclosed constructs
	if len(balancingStack) > 0 {
		return nil, &LexerError{
			Message: fmt.Sprintf("unclosed %q", string(balancingStack[len(balancingStack)-1])),
			Line:    lineno,
			Column:  column,
			Pos:     pos,
		}
	}

	// Gracefully close unterminated line statements/comments at EOF
	for len(stack) > 1 {
		currentState := stack[len(stack)-1]
		switch currentState {
		case StateLineStatement:
			tokens = append(tokens, TokenInfo{
				Line:   lineno,
				Column: column,
				Type:   "block_end",
				Value:  "",
			})
			stack = stack[:len(stack)-1]
		case StateLineComment:
			tokens = append(tokens, TokenInfo{
				Line:   lineno,
				Column: column,
				Type:   "comment_end",
				Value:  "",
			})
			stack = stack[:len(stack)-1]
		default:
			// For other states we break to report a proper error below
			stack = stack[:len(stack)-1]
			// Push back to allow error message to reference this state
			stack = append(stack, currentState)
			goto validateRemaining
		}
	}

validateRemaining:
	if len(stack) > 1 {
		currentState := stack[len(stack)-1]
		var expectedTag string
		switch currentState {
		case StateVariableBegin:
			expectedTag = l.config.Delimiters.VariableEnd
		case StateBlockBegin:
			expectedTag = l.config.Delimiters.BlockEnd
		case StateCommentBegin:
			expectedTag = l.config.Delimiters.CommentEnd
		case StateRawBegin:
			expectedTag = "endraw"
		case StateLineStatement:
			expectedTag = "end of line"
		case StateLineComment:
			expectedTag = "end of line"
		}
		return nil, &LexerError{
			Message: fmt.Sprintf("unclosed construct - missing %s", expectedTag),
			Line:    lineno,
			Column:  column,
			Pos:     pos,
		}
	}

	return tokens, nil
}

// TokenInfo represents intermediate token information
type TokenInfo struct {
	Line   int
	Column int
	Type   string
	Value  string
}

// processMatch processes a regex match and generates tokens with proper whitespace handling
func (l *Lexer) processMatch(rule *Rule, source string, loc []int, pos, lineno, column int, filename string, balancingStack []rune, lineStarting bool) ([]TokenInfo, int, error) {
	var tokens []TokenInfo
	var newlinesStripped int

	switch t := rule.Tokens.(type) {
	case string:
		if strings.HasPrefix(t, "error:") {
			return nil, 0, &LexerError{
				Message: strings.TrimPrefix(t, "error:"),
				Line:    lineno,
				Column:  column,
				Pos:     pos,
			}
		}

		if t == "#bygroup" {
			// Handle bygroup matching - this can be either a single match or token array
			groups := rule.Regex.FindStringSubmatch(source[pos:])
			if len(groups) > 0 {
				// Check if we have multiple groups (text + delimiter)
				if len(groups) >= 2 {
					// First group is text before delimiter
					if groups[1] != "" {
						tokens = append(tokens, TokenInfo{
							Line:   lineno,
							Column: column,
							Type:   "data",
							Value:  groups[1],
						})
					}

					// Find which delimiter group matched (after the text group)
					for i := 2; i < len(groups); i++ {
						if groups[i] != "" {
							tokenType := l.mapMatchToTokenType(groups[i])
							if tokenType != "" {
								tokenColumn := column + utf8.RuneCountInString(groups[1])
								tokens = append(tokens, TokenInfo{
									Line:   lineno,
									Column: tokenColumn,
									Type:   tokenType,
									Value:  groups[i],
								})
							}
							break
						}
					}
				} else {
					// Single group - just the delimiter
					delimMatch := groups[0]
					tokenType := l.mapMatchToTokenType(delimMatch)
					if tokenType != "" {
						tokens = append(tokens, TokenInfo{
							Line:   lineno,
							Column: column,
							Type:   tokenType,
							Value:  delimMatch,
						})
					}
				}
			}
		} else {
			value := source[pos+loc[0] : pos+loc[1]]
			tokens = append(tokens, TokenInfo{
				Line:   lineno,
				Column: column,
				Type:   t,
				Value:  value,
			})
		}

	case []string:
		groups := rule.Regex.FindStringSubmatch(source[pos:])
		if len(groups) > 0 {
			// Handle optional lstrip for token arrays
			if l.supportsLStrip(rule) && len(groups) > 1 {
				tokenTypes, _ := rule.Tokens.([]string)
				processedTokens, stripped := l.handleLStrip(groups, tokenTypes, nil, source, pos, lineno, column, lineStarting)
				tokens = append(tokens, processedTokens...)
				newlinesStripped = stripped
			} else {
				// Regular token array processing
				currentColumn := column
				groupIndex := 1
				for _, tokenType := range t {
					if groupIndex >= len(groups) {
						break
					}
					value := groups[groupIndex]
					groupIndex++
					if value != "" || !l.shouldIgnoreToken(tokenType) {
						tokens = append(tokens, TokenInfo{
							Line:   lineno,
							Column: currentColumn,
							Type:   tokenType,
							Value:  value,
						})
					}
					currentColumn += utf8.RuneCountInString(value)
				}
			}
		}
	}

	return tokens, newlinesStripped, nil
}

func (l *Lexer) isRawLikeTag(match string) bool {
	lower := strings.ToLower(match)
	blockStart := strings.ToLower(l.config.Delimiters.BlockStart)
	blockEnd := strings.ToLower(l.config.Delimiters.BlockEnd)

	lower = strings.TrimPrefix(lower, blockStart)
	lower = strings.TrimPrefix(lower, "-")
	lower = strings.TrimSpace(lower)
	if strings.HasSuffix(lower, blockEnd) {
		lower = strings.TrimSuffix(lower, blockEnd)
	}
	lower = strings.TrimSuffix(lower, "-")
	lower = strings.TrimSpace(lower)
	return strings.HasPrefix(lower, "raw") || strings.HasPrefix(lower, "verbatim")
}

// determineStateByGroup determines the next state based on which regex matched
func (l *Lexer) determineStateByGroup(rule *Rule, source string) LexerState {
	groups := rule.Regex.FindStringSubmatch(source)
	if len(groups) == 0 {
		return ""
	}

	// Find which delimiter group matched (groups 2+ are delimiters)
	for i := 2; i < len(groups); i++ {
		if groups[i] != "" {
			match := groups[i]

			// Map the matched delimiter to a state
			if strings.HasPrefix(match, l.config.Delimiters.CommentStart) {
				return StateCommentBegin
			}
			if strings.HasPrefix(match, l.config.Delimiters.BlockStart) {
				if l.isRawLikeTag(match) {
					return StateRawBegin
				}
				return StateBlockBegin
			}
			if strings.HasPrefix(match, l.config.Delimiters.VariableStart) {
				return StateVariableBegin
			}
			if l.config.Delimiters.LineStatement != "" && strings.Contains(match, l.config.Delimiters.LineStatement) {
				return StateLineStatement
			}
			if l.config.Delimiters.LineComment != "" && strings.Contains(match, l.config.Delimiters.LineComment) {
				return StateLineComment
			}
		}
	}

	return ""
}

// mapMatchToTokenType maps a matched delimiter to the appropriate token type
func (l *Lexer) mapMatchToTokenType(match string) string {
	// Check which delimiter pattern matched
	if strings.HasPrefix(match, l.config.Delimiters.CommentStart) {
		return "comment_begin"
	}
	if strings.HasPrefix(match, l.config.Delimiters.BlockStart) {
		if l.isRawLikeTag(match) {
			return "raw_begin"
		}
		return "block_begin"
	}
	if strings.HasPrefix(match, l.config.Delimiters.VariableStart) {
		return "variable_begin"
	}
	if l.config.Delimiters.LineStatement != "" && strings.Contains(match, l.config.Delimiters.LineStatement) {
		return "linestatement_begin"
	}
	if l.config.Delimiters.LineComment != "" && strings.Contains(match, l.config.Delimiters.LineComment) {
		return "linecomment_begin"
	}
	return ""
}

// supportsLStrip checks if a rule supports lstrip functionality
func (l *Lexer) supportsLStrip(rule *Rule) bool {
	if tokenArray, ok := rule.Tokens.([]string); ok {
		if len(tokenArray) <= 1 {
			return false
		}
	}
	// Rules that contain delimiter patterns support lstrip
	pattern := rule.Regex.String()
	return strings.Contains(pattern, l.config.Delimiters.BlockStart) ||
		strings.Contains(pattern, l.config.Delimiters.CommentStart)
}

// handleLStrip processes lstrip functionality for rules that support it
func (l *Lexer) handleLStrip(groups []string, tokenTypes []string, subexpNames []string, source string, pos, lineno, column int, lineStarting bool) ([]TokenInfo, int) {
	var tokens []TokenInfo
	var newlinesStripped int

	// Extract the text before the delimiter
	text := ""
	if len(groups) > 1 {
		text = groups[1]
	} else if len(groups) > 0 {
		text = groups[0]
	}

	// Find the strip sign (-, +, or empty)
	stripSign := ""
	for i := 2; i < len(groups); i++ {
		if groups[i] == "-" || groups[i] == "+" {
			stripSign = groups[i]
			break
		}
	}

	if stripSign == "-" {
		// Strip all whitespace between the text and the tag
		stripped := strings.TrimRight(text, " \t\r\n")
		newlinesStripped = strings.Count(text[len(stripped):], "\n")
		text = stripped
	} else if stripSign != "+" && l.config.LstripBlocks {
		// Check if we should apply lstrip
		lastNewline := strings.LastIndex(text, "\n")
		if lastNewline >= 0 || lineStarting {
			// Check if there's only whitespace after the last newline
			afterNewline := text[lastNewline+1:]
			if strings.TrimSpace(afterNewline) == "" {
				text = text[:lastNewline+1]
			}
		}
	}

	// Add the text token if it's not empty
	if len(tokenTypes) > 0 && (text != "" || !l.shouldIgnoreToken(tokenTypes[0])) {
		tokens = append(tokens, TokenInfo{
			Line:   lineno,
			Column: column,
			Type:   tokenTypes[0],
			Value:  text,
		})
	}

	// Calculate column position for delimiter tokens
	delimiterColumn := column + utf8.RuneCountInString(text)

	// Process the delimiter tokens
	if subexpNames != nil {
		// Bygroup case - find the matched group
		for i := 2; i < len(groups); i++ {
			if groups[i] != "" && i < len(subexpNames) {
				groupName := subexpNames[i]
				if groupName != "" && groupName != "data" {
					tokens = append(tokens, TokenInfo{
						Line:   lineno,
						Column: delimiterColumn,
						Type:   groupName,
						Value:  groups[i],
					})
					break
				}
			}
		}
	} else {
		// Token array case
		currentColumn := delimiterColumn
		groupIndex := 2
		for i := 1; i < len(tokenTypes) && groupIndex < len(groups); i++ {
			value := groups[groupIndex]
			groupIndex++
			if value != "" || !l.shouldIgnoreToken(tokenTypes[i]) {
				tokens = append(tokens, TokenInfo{
					Line:   lineno,
					Column: currentColumn,
					Type:   tokenTypes[i],
					Value:  value,
				})
			}
			currentColumn += utf8.RuneCountInString(value)
		}
	}

	return tokens, newlinesStripped
}

// wrap converts raw token info into processed tokens
func (l *Lexer) wrap(tokenInfos []TokenInfo, name, filename string) ([]Token, error) {
	var tokens []Token

	for _, info := range tokenInfos {
		if l.shouldIgnoreToken(info.Type) {
			continue
		}

		token := Token{
			Line:   info.Line,
			Column: info.Column,
			Value:  info.Value,
		}

		// Convert token type string to TokenType
		switch info.Type {
		case "data", TokenText.String():
			token.Type = TokenText
			token.Value = l.normalizeNewlines(info.Value)
		case "variable_begin":
			token.Type = TokenVariableStart
		case "variable_end":
			token.Type = TokenVariableEnd
		case "block_begin", "linestatement_begin":
			token.Type = TokenBlockStart
		case "block_end", "linestatement_end":
			token.Type = TokenBlockEnd
		case "comment_begin", "linecomment_begin":
			token.Type = TokenCommentStart
		case "comment_end", "linecomment_end":
			token.Type = TokenCommentEnd
		case "raw_begin", "raw_end":
			continue // Skip raw tokens
		case "raw_data":
			token.Type = TokenText
			token.Value = l.normalizeNewlines(info.Value)
		case "name":
			// Check for logical operators and keywords
			switch strings.ToLower(info.Value) {
			case "and":
				token.Type = TokenAnd
			case "or":
				token.Type = TokenOr
			case "not":
				token.Type = TokenNot
			case "in":
				token.Type = TokenComparison
			case "is":
				token.Type = TokenComparison
			default:
				token.Type = TokenName
				if !isValidIdentifier(info.Value) {
					return nil, &LexerError{
						Message: fmt.Sprintf("invalid identifier %q", info.Value),
						Line:    info.Line,
						Column:  info.Column,
					}
				}
			}
		case "string":
			token.Type = TokenString
			token.Value = l.unescapeString(info.Value)
		case "integer", "float":
			token.Type = TokenNumber
			// Parse numbers
			if info.Type == "integer" {
				if val, err := strconv.ParseInt(strings.ReplaceAll(info.Value, "_", ""), 0, 64); err == nil {
					token.Value = fmt.Sprintf("%d", val)
				}
			} else {
				if val, err := strconv.ParseFloat(strings.ReplaceAll(info.Value, "_", ""), 64); err == nil {
					token.Value = fmt.Sprintf("%g", val)
				}
			}
		case "operator":
			token.Type = l.mapOperator(info.Value)
		default:
			// Try to map to known token types
			if tt, ok := mapStringToTokenType(info.Type); ok {
				token.Type = tt
			} else {
				token.Type = TokenText
			}
		}

		tokens = append(tokens, token)
	}

	return tokens, nil
}

// Helper functions

func strPtr(s string) *string {
	return &s
}

func (l *Lexer) getGroupName(regex *regexp.Regexp, groupIndex int) string {
	names := regex.SubexpNames()
	if groupIndex < len(names) {
		return names[groupIndex]
	}
	return ""
}

// updateBalancingStackForMatch updates the balancing stack based on the matched rule and text
func (l *Lexer) updateBalancingStackForMatch(stack *[]rune, rule *Rule, matchText string, lineno, column int, name, filename string) {
	// Only operator tokens should affect brace balancing
	// String tokens should NOT affect balancing because braces inside strings are literal
	if tokensStr, ok := rule.Tokens.(string); ok && tokensStr == "operator" {
		l.updateBalancingStack(stack, matchText, lineno, column)
	}
}

// updateBalancingStack updates the stack for balanced braces/parentheses/brackets
func (l *Lexer) updateBalancingStack(stack *[]rune, text string, lineno, column int) {
	for _, char := range text {
		switch char {
		case '{':
			*stack = append(*stack, '}')
		case '(':
			*stack = append(*stack, ')')
		case '[':
			*stack = append(*stack, ']')
		case '}', ')', ']':
			if len(*stack) == 0 {
				// This should not happen in normal operation, but we'll handle it gracefully
				continue
			}
			expected := (*stack)[len(*stack)-1]
			*stack = (*stack)[:len(*stack)-1]
			if expected != char {
				// Mismatched closing bracket - this will be caught by final validation
				*stack = append(*stack, expected) // Put it back for proper error reporting
			}
		}
	}
}

func (l *Lexer) shouldIgnoreToken(tokenType string) bool {
	ignored := map[string]bool{
		"comment_begin":     true,
		"comment":           true,
		"comment_end":       true,
		"whitespace":        true,
		"linecomment_begin": true,
		"linecomment":       true,
		"linecomment_end":   true,
	}
	return ignored[tokenType]
}

func matchLinePrefix(src, prefix string) (int, bool) {
	if prefix == "" {
		return 0, false
	}

	idx := 0
	for idx < len(src) {
		r, size := utf8.DecodeRuneInString(src[idx:])
		if r == ' ' || r == '\t' || r == '\v' {
			idx += size
			continue
		}
		break
	}

	if strings.HasPrefix(src[idx:], prefix) {
		return idx, true
	}
	return 0, false
}

func findLineControlBoundary(src, linePrefix, commentPrefix, variableStart, blockStart, commentStart string) int {
	if linePrefix == "" && commentPrefix == "" {
		return 0
	}

	offset := 0
	for {
		idx := strings.Index(src[offset:], "\n")
		if idx < 0 {
			return 0
		}

		newlinePos := offset + idx
		nextPos := newlinePos + 1
		remainder := src[nextPos:]
		segment := src[:nextPos]
		segmentContent := segment
		if strings.HasSuffix(segmentContent, "\n") {
			segmentContent = segmentContent[:len(segmentContent)-1]
		}

		if linePrefix != "" {
			if _, ok := matchLinePrefix(remainder, linePrefix); ok {
				if !containsTemplateDelimiter(segmentContent, variableStart, blockStart, commentStart) {
					return nextPos
				}
			}
		}

		if commentPrefix != "" {
			if _, ok := matchLinePrefix(remainder, commentPrefix); ok {
				if !containsTemplateDelimiter(segmentContent, variableStart, blockStart, commentStart) {
					return nextPos
				}
			}
		}

		offset = nextPos
		if offset >= len(src) {
			return 0
		}
	}
}

func containsTemplateDelimiter(segment, variableStart, blockStart, commentStart string) bool {
	if segment == "" {
		return false
	}

	if variableStart != "" && strings.Contains(segment, variableStart) {
		return true
	}
	if blockStart != "" && strings.Contains(segment, blockStart) {
		return true
	}
	if commentStart != "" && strings.Contains(segment, commentStart) {
		return true
	}
	return false
}

func (l *Lexer) normalizeNewlines(value string) string {
	return NewlineRegex.ReplaceAllString(value, l.config.NewlineSequence)
}

func (l *Lexer) unescapeString(value string) string {
	// Remove quotes and unescape
	if len(value) >= 2 && (value[0] == '\'' || value[0] == '"') {
		value = value[1 : len(value)-1]
	}

	// Handle common escape sequences
	result := value
	result = strings.ReplaceAll(result, "\\n", "\n")
	result = strings.ReplaceAll(result, "\\r", "\r")
	result = strings.ReplaceAll(result, "\\t", "\t")
	result = strings.ReplaceAll(result, "\\\\", "\\")
	result = strings.ReplaceAll(result, "\\\"", "\"")
	result = strings.ReplaceAll(result, "\\'", "'")
	result = strings.ReplaceAll(result, "\\b", "\b")
	result = strings.ReplaceAll(result, "\\f", "\f")
	result = strings.ReplaceAll(result, "\\v", "\v")

	return result
}

func (l *Lexer) mapOperator(op string) TokenType {
	operatorMap := map[string]TokenType{
		"+":  TokenAdd,
		"-":  TokenSub,
		"*":  TokenMul,
		"/":  TokenDiv,
		"//": TokenFloorDiv,
		"%":  TokenMod,
		"**": TokenPow,
		"==": TokenComparison,
		"!=": TokenComparison,
		">":  TokenComparison,
		">=": TokenComparison,
		"<":  TokenComparison,
		"<=": TokenComparison,
		"=":  TokenAssign,
		".":  TokenDot,
		":":  TokenColon,
		"|":  TokenPipe,
		",":  TokenComma,
		";":  TokenSemicolon,
		"(":  TokenLeftParen,
		")":  TokenRightParen,
		"[":  TokenLeftBracket,
		"]":  TokenRightBracket,
		"{":  TokenLeftCurly,
		"}":  TokenRightCurly,
		"!":  TokenNot,
		"&":  TokenAnd,
		"?":  TokenTernary,
		"~":  TokenAdd, // Concatenation operator maps to Add for now
	}
	if tt, ok := operatorMap[op]; ok {
		return tt
	}
	return TokenOperator
}

func mapStringToTokenType(s string) (TokenType, bool) {
	tokenMap := map[string]TokenType{
		"data":                TokenText,
		"variable_begin":      TokenVariableStart,
		"variable_end":        TokenVariableEnd,
		"block_begin":         TokenBlockStart,
		"block_end":           TokenBlockEnd,
		"comment_begin":       TokenCommentStart,
		"comment_end":         TokenCommentEnd,
		"name":                TokenName,
		"string":              TokenString,
		"integer":             TokenNumber,
		"float":               TokenNumber,
		"operator":            TokenOperator,
		"whitespace":          TokenWhitespace,
		"linestatement_begin": TokenBlockStart,
		"linestatement_end":   TokenBlockEnd,
		"linecomment_begin":   TokenCommentStart,
		"linecomment_end":     TokenCommentEnd,
	}
	tt, ok := tokenMap[s]
	return tt, ok
}

// applyWhitespaceControl applies lstrip/trim whitespace control based on the sign
func (l *Lexer) applyWhitespaceControl(text, stripSign string, lineStarting bool) (string, int) {
	var newlinesStripped int

	switch stripSign {
	case "-":
		// Strip all whitespace to the left including newlines
		stripped := strings.TrimRight(text, " \t\r\n")
		newlinesStripped = strings.Count(text[len(stripped):], "\n")
		return stripped, newlinesStripped

	case "+":
		// Preserve all whitespace
		return text, 0

	default:
		// No explicit sign - apply lstrip_blocks if enabled
		if l.config.LstripBlocks {
			lastNewline := strings.LastIndex(text, "\n")
			if lastNewline >= 0 || lineStarting {
				// Check if there's only whitespace after the last newline
				afterNewline := text[lastNewline+1:]
				if strings.TrimSpace(afterNewline) == "" {
					newlinesStripped = strings.Count(afterNewline, "\n")
					return text[:lastNewline+1], newlinesStripped
				}
			}
		}
		return text, 0
	}
}

// isDelimiterGroup checks if a matched group is a delimiter
func (l *Lexer) isDelimiterGroup(group string) bool {
	return strings.HasPrefix(group, l.config.Delimiters.CommentStart) ||
		strings.HasPrefix(group, l.config.Delimiters.BlockStart) ||
		strings.HasPrefix(group, l.config.Delimiters.VariableStart) ||
		(l.config.Delimiters.LineStatement != "" && strings.Contains(group, l.config.Delimiters.LineStatement)) ||
		(l.config.Delimiters.LineComment != "" && strings.Contains(group, l.config.Delimiters.LineComment))
}

func isValidIdentifier(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		if i == 0 && !unicode.IsLetter(r) && r != '_' {
			return false
		}
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			return false
		}
	}
	return true
}
