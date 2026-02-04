package lexer

import (
	"fmt"
	"strings"
	"unicode"
)

// Lexer tokenizes Kukicha source code.
//
// ARCHITECTURE NOTE: Kukicha uses Python-style indentation-based blocks.
// The lexer converts 4-space indentation changes into INDENT and DEDENT tokens,
// which the parser then uses to determine block structure.
//
// The indentStack tracks nesting levels. When indentation increases by 4 spaces,
// an INDENT token is emitted and the level is pushed. When it decreases, DEDENT
// tokens are emitted (possibly multiple) until the stack matches the new level.
//
// Why 4 spaces only (no tabs)?
//   - Consistency: Eliminates "tabs vs spaces" debates
//   - Beginner-friendly: One clear rule, no configuration needed
//   - Error prevention: Mixed tabs/spaces is a common Python mistake
type Lexer struct {
	source             []rune
	start              int
	current            int
	line               int
	column             int
	file               string
	tokens             []Token
	indentStack        []int // Stack of indentation levels (in spaces). Always starts with [0].
	pendingDedents     int   // Dedents to emit
	atLineStart        bool  // Whether we're at the start of a line
	indentationHandled bool  // Whether indentation has been handled for the current line
	errors             []error

	// Pipe continuation support: a trailing |> at end of line suppresses the
	// NEWLINE token and causes the next line's indentation to be consumed
	// without emitting INDENT/DEDENT, so the pipe RHS is parsed as part of
	// the same expression regardless of how it is indented.
	lastTokenType    TokenType // last emitted token type (TOKEN_COMMENT excluded)
	continuationLine bool      // true when the next line is a |> continuation
	braceDepth       int       // current nesting level of (), [], {}
}

// NewLexer creates a new lexer for the given source code
func NewLexer(source string, filename string) *Lexer {
	return &Lexer{
		source:             []rune(source),
		file:               filename,
		line:               1,
		column:             1,
		indentStack:        []int{0},
		atLineStart:        true,
		indentationHandled: false,
		braceDepth:         0,
	}
}

// ScanTokens scans all tokens from the source
func (l *Lexer) ScanTokens() ([]Token, error) {
	for !l.isAtEnd() {
		l.start = l.current
		l.scanToken()
	}

	// Emit remaining dedents
	for len(l.indentStack) > 1 {
		l.addToken(TOKEN_DEDENT)
		l.indentStack = l.indentStack[:len(l.indentStack)-1]
	}

	l.addToken(TOKEN_EOF)

	if len(l.errors) > 0 {
		return nil, fmt.Errorf("lexer errors: %v", l.errors)
	}

	return l.tokens, nil
}

// scanToken scans a single token
func (l *Lexer) scanToken() {
	// Pipe continuation: the previous line ended with |> so this line's
	// leading whitespace is consumed without emitting INDENT/DEDENT.  The
	// indent stack is left untouched; when the pipe chain ends and a normal
	// NEWLINE is emitted the next line's indentation is compared against the
	// unchanged stack and DEDENT tokens are emitted as usual.
	if l.atLineStart && l.continuationLine && !l.indentationHandled {
		l.continuationLine = false
		for !l.isAtEnd() && (l.peek() == ' ' || l.peek() == '\t') {
			l.advance()
		}
		l.start = l.current // move past consumed whitespace so it is not included in the next token
		l.indentationHandled = true
		// Fall through — the next token on this line is scanned normally.
	}

	// Handle indentation at line start
	if l.atLineStart && !l.indentationHandled {
		c := l.peek()

		// If it's space or tab, we definitely need to handle indentation
		if c == ' ' || c == '\t' {
			l.indentationHandled = true
			l.handleIndentation()
			return
		}

		// Check for implicit dedent to 0 level (no indentation)
		// Don't process for newlines or comments which handle their own flow
		if c != '\n' && c != '\r' && c != '#' {
			if len(l.indentStack) > 1 {
				l.indentationHandled = true
				l.handleIndentation()
				return
			}
			// Mark indentation as handled even if we don't change indentation
			l.indentationHandled = true
		}
	}

	c := l.advance()

	l.atLineStart = false

	switch c {
	case ' ', '\t':
		// Skip whitespace (not at line start)
		for !l.isAtEnd() && (l.peek() == ' ' || l.peek() == '\t') {
			l.advance()
		}
	case '\n':
		// Implicit continuation if:
		// 1. We are inside braces (braceDepth > 0)
		// 2. The *previous* token was a pipe
		// 3. The *next* token (on the new line) is a pipe
		if l.braceDepth > 0 || l.lastTokenType == TOKEN_PIPE || l.isPipeAtStartOfNextLine() {
			l.continuationLine = true
		} else {
			l.addToken(TOKEN_NEWLINE)
		}
		l.line++
		l.column = 0
		l.atLineStart = true
		l.indentationHandled = false
	case '\r':
		if l.peek() == '\n' {
			l.advance()
		}
		if l.braceDepth > 0 || l.lastTokenType == TOKEN_PIPE || l.isPipeAtStartOfNextLine() {
			l.continuationLine = true
		} else {
			l.addToken(TOKEN_NEWLINE)
		}
		l.line++
		l.column = 0
		l.atLineStart = true
		l.indentationHandled = false
	case '#':
		l.scanComment()
	case ';':
		l.addToken(TOKEN_SEMICOLON)
	case '"':
		l.scanString()
	case '\'':
		l.scanRune()
	case '(':
		l.braceDepth++
		l.addToken(TOKEN_LPAREN)
	case ')':
		if l.braceDepth > 0 {
			l.braceDepth--
		}
		l.addToken(TOKEN_RPAREN)
	case '[':
		l.braceDepth++
		l.addToken(TOKEN_LBRACKET)
	case ']':
		if l.braceDepth > 0 {
			l.braceDepth--
		}
		l.addToken(TOKEN_RBRACKET)
	case '{':
		l.braceDepth++
		l.addToken(TOKEN_LBRACE)
	case '}':
		if l.braceDepth > 0 {
			l.braceDepth--
		}
		l.addToken(TOKEN_RBRACE)
	case ',':
		l.addToken(TOKEN_COMMA)
	case '.':
		l.addToken(TOKEN_DOT)
	case '+':
		if l.match('+') {
			l.addToken(TOKEN_PLUS_PLUS)
		} else {
			l.addToken(TOKEN_PLUS)
		}
	case '-':
		if l.match('-') {
			l.addToken(TOKEN_MINUS_MINUS)
		} else {
			l.addToken(TOKEN_MINUS)
		}
	case '*':
		l.addToken(TOKEN_STAR)
	case '/':
		l.addToken(TOKEN_SLASH)
	case '%':
		l.addToken(TOKEN_PERCENT)
	case ':':
		if l.match('=') {
			l.addToken(TOKEN_WALRUS)
		} else {
			l.addToken(TOKEN_COLON)
		}
	case '=':
		if l.match('=') {
			l.addToken(TOKEN_DOUBLE_EQUALS)
		} else {
			l.addToken(TOKEN_ASSIGN)
		}
	case '!':
		if l.match('=') {
			l.addToken(TOKEN_NOT_EQUALS)
		} else {
			l.addToken(TOKEN_BANG)
		}
	case '<':
		if l.match('-') {
			l.addToken(TOKEN_ARROW_LEFT)
		} else if l.match('=') {
			l.addToken(TOKEN_LTE)
		} else {
			l.addToken(TOKEN_LT)
		}
	case '>':
		if l.match('=') {
			l.addToken(TOKEN_GTE)
		} else {
			l.addToken(TOKEN_GT)
		}
	case '|':
		if l.match('>') {
			l.addToken(TOKEN_PIPE)
		} else if l.match('|') {
			l.addToken(TOKEN_OR_OR)
		} else {
			l.addToken(TOKEN_BIT_OR)
		}
	case '&':
		if l.match('&') {
			l.addToken(TOKEN_AND_AND)
		} else {
			l.error("Unexpected '&'. Use 'and' for logical AND, or '&&' for Go-style AND. Bitwise AND is not supported.")
		}
	default:
		if isDigit(c) {
			l.scanNumber()
		} else if isAlpha(c) {
			l.scanIdentifier()
		} else {
			l.error(fmt.Sprintf("Unexpected character: %c", c))
		}
	}
}

// handleIndentation handles indentation at the start of a line.
//
// ARCHITECTURE NOTE: The algorithm works as follows:
//  1. Count leading spaces (tabs are rejected with error)
//  2. Skip blank lines and comment-only lines (no tokens emitted)
//  3. Validate spacing is a multiple of 4
//  4. Compare to current indent level:
//     - If greater: push level, emit INDENT
//     - If smaller: pop levels, emit DEDENT for each
//     - If equal: no token (same level continues)
//
// The indentStack ensures dedents always return to a valid prior level.
// For example, going from 8 spaces directly to 0 emits two DEDENT tokens.
func (l *Lexer) handleIndentation() {
	spaces := 0
	tabs := 0

	// Count spaces and tabs
	for !l.isAtEnd() && (l.peek() == ' ' || l.peek() == '\t') {
		if l.peek() == ' ' {
			spaces++
		} else {
			tabs++
		}
		l.advance()
	}

	// Check for tabs
	if tabs > 0 {
		l.error("Use 4 spaces for indentation, not tabs")
		return
	}

	// Skip blank lines and comment-only lines
	if l.isAtEnd() || l.peek() == '\n' || l.peek() == '\r' || l.peek() == '#' {
		return
	}

	// Must be multiple of 4
	if spaces%4 != 0 {
		l.error(fmt.Sprintf("Indentation must be multiple of 4 spaces, got %d", spaces))
		return
	}

	currentIndent := l.indentStack[len(l.indentStack)-1]

	if spaces > currentIndent {
		// Indent
		if spaces != currentIndent+4 {
			l.error(fmt.Sprintf("Indentation can only increase by 4 spaces, got increase of %d", spaces-currentIndent))
			return
		}
		l.indentStack = append(l.indentStack, spaces)
		l.addToken(TOKEN_INDENT)
	} else if spaces < currentIndent {
		// Dedent (possibly multiple levels)
		for len(l.indentStack) > 1 && l.indentStack[len(l.indentStack)-1] > spaces {
			l.addToken(TOKEN_DEDENT)
			l.indentStack = l.indentStack[:len(l.indentStack)-1]
		}

		// Verify we landed on a valid indentation level
		if l.indentStack[len(l.indentStack)-1] != spaces {
			l.error("Indentation mismatch")
		}
	}
}

// scanString scans a double-quoted string literal with optional interpolation
func (l *Lexer) scanString() {
	value := strings.Builder{}

	for !l.isAtEnd() && l.peek() != '"' {
		if l.peek() == '\n' {
			l.error("Unterminated string")
			return
		}

		if l.peek() == '\\' {
			// Handle escape sequences
			l.advance() // consume \
			if !l.isAtEnd() {
				escaped := l.advance()
				switch escaped {
				case 'n':
					value.WriteRune('\n')
				case 't':
					value.WriteRune('\t')
				case 'r':
					value.WriteRune('\r')
				case '\\':
					value.WriteRune('\\')
				case '"':
					value.WriteRune('"')
				case '\'':
					value.WriteRune('\'')
				default:
					value.WriteRune(escaped)
				}
			}
		} else if l.peek() == '{' {
			// String interpolation
			value.WriteRune(l.advance())
		} else {
			value.WriteRune(l.advance())
		}
	}

	if l.isAtEnd() {
		l.error("Unterminated string")
		return
	}

	l.advance() // consume closing quote

	// For now, store the entire string including interpolation markers
	// The parser will handle breaking it down into segments
	l.addTokenWithLexeme(TOKEN_STRING, value.String())
}

// scanRune scans a single-quoted character/rune literal
func (l *Lexer) scanRune() {
	if l.isAtEnd() {
		l.error("Unterminated character literal")
		return
	}

	var char rune
	if l.peek() == '\\' {
		// Handle escape sequences
		l.advance() // consume \
		if l.isAtEnd() {
			l.error("Unterminated escape sequence in character literal")
			return
		}
		escaped := l.advance()
		switch escaped {
		case 'n':
			char = '\n'
		case 't':
			char = '\t'
		case 'r':
			char = '\r'
		case '\\':
			char = '\\'
		case '\'':
			char = '\''
		case '"':
			char = '"'
		case '0':
			char = '\x00'
		default:
			char = escaped
		}
	} else if l.peek() == '\'' {
		l.error("Empty character literal")
		return
	} else {
		char = l.advance()
	}

	if l.isAtEnd() || l.peek() != '\'' {
		l.error("Unterminated character literal (use double quotes for strings)")
		return
	}

	l.advance() // consume closing quote

	// Store the rune as a string (the value will be the character)
	l.addTokenWithLexeme(TOKEN_RUNE, string(char))
}

// scanNumber scans a number (integer or float)
func (l *Lexer) scanNumber() {
	for isDigit(l.peek()) {
		l.advance()
	}

	// Look for decimal point
	if l.peek() == '.' && isDigit(l.peekNext()) {
		l.advance() // consume .

		for isDigit(l.peek()) {
			l.advance()
		}

		l.addToken(TOKEN_FLOAT)
	} else {
		l.addToken(TOKEN_INTEGER)
	}
}

// scanIdentifier scans an identifier or keyword
func (l *Lexer) scanIdentifier() {
	for isAlphaNumeric(l.peek()) {
		l.advance()
	}

	text := string(l.source[l.start:l.current])
	tokenType := LookupKeyword(text)
	l.addToken(tokenType)
}

// scanComment scans a comment and emits a TOKEN_COMMENT
func (l *Lexer) scanComment() {
	// Consume the rest of the comment line
	for !l.isAtEnd() && l.peek() != '\n' {
		l.advance()
	}
	// The lexeme includes the # and the comment text
	l.addToken(TOKEN_COMMENT)
}

// Helper methods

func (l *Lexer) isAtEnd() bool {
	return l.current >= len(l.source)
}

func (l *Lexer) advance() rune {
	if l.isAtEnd() {
		return 0
	}
	c := l.source[l.current]
	l.current++
	l.column++
	return c
}

func (l *Lexer) peek() rune {
	if l.isAtEnd() {
		return 0
	}
	return l.source[l.current]
}

func (l *Lexer) peekNext() rune {
	if l.current+1 >= len(l.source) {
		return 0
	}
	return l.source[l.current+1]
}

func (l *Lexer) match(expected rune) bool {
	if l.isAtEnd() {
		return false
	}
	if l.source[l.current] != expected {
		return false
	}
	l.current++
	l.column++
	return true
}

func (l *Lexer) addToken(tokenType TokenType) {
	l.addTokenWithLexeme(tokenType, string(l.source[l.start:l.current]))
}

func (l *Lexer) addTokenWithLexeme(tokenType TokenType, lexeme string) {
	token := Token{
		Type:   tokenType,
		Lexeme: lexeme,
		Line:   l.line,
		Column: l.column - len([]rune(lexeme)),
		File:   l.file,
	}
	l.tokens = append(l.tokens, token)
	// Track last emitted type for pipe-continuation logic.  Comments are
	// excluded so that a comment on the same line as a trailing |> does not
	// break the continuation (the parser already skips TOKEN_COMMENT).
	if tokenType != TOKEN_COMMENT {
		l.lastTokenType = tokenType
	}
}

func (l *Lexer) error(message string) {
	err := fmt.Errorf("%s:%d:%d: %s", l.file, l.line, l.column, message)
	l.errors = append(l.errors, err)
}

// Character classification helpers

func isDigit(c rune) bool {
	return c >= '0' && c <= '9'
}

func isAlpha(c rune) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		c == '_'
}

func isAlphaNumeric(c rune) bool {
	return isAlpha(c) || isDigit(c)
}

func isWhitespace(c rune) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

// IsKeyword checks if a string is a keyword
func IsKeyword(s string) bool {
	_, ok := keywords[s]
	return ok
}

// Helper to check if a rune is a letter (for identifiers)
func isLetter(c rune) bool {
	return unicode.IsLetter(c) || c == '_'
}

// isPipeAtStartOfNextLine checks if the next non-whitespace characters
// on the upcoming line form a pipe operator "|>".  Called from the '\n'
// (or '\r') case after advance() has already consumed the newline, so
// l.current points to the first character of the next line.
func (l *Lexer) isPipeAtStartOfNextLine() bool {
	idx := l.current

	for idx < len(l.source) {
		c := l.source[idx]
		if c == ' ' || c == '\t' {
			idx++
			continue
		}
		// Found non-whitespace
		if c == '|' && idx+1 < len(l.source) && l.source[idx+1] == '>' {
			return true
		}
		return false
	}
	return false
}
