package lexer

import (
	"fmt"
	"strings"
	"unicode"
)

// Lexer tokenizes Kukicha source code
type Lexer struct {
	source             []rune
	start              int
	current            int
	line               int
	column             int
	file               string
	tokens             []Token
	indentStack        []int // Track indentation levels
	pendingDedents     int   // Dedents to emit
	atLineStart        bool  // Whether we're at the start of a line
	indentationHandled bool  // Whether indentation has been handled for the current line
	errors             []error
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
		l.addToken(TOKEN_NEWLINE)
		l.line++
		l.column = 0
		l.atLineStart = true
		l.indentationHandled = false
	case '\r':
		if l.peek() == '\n' {
			l.advance()
		}
		l.addToken(TOKEN_NEWLINE)
		l.line++
		l.column = 0
		l.atLineStart = true
		l.indentationHandled = false
	case '#':
		l.scanComment()
	case ';':
		l.addToken(TOKEN_SEMICOLON)
	case '"', '\'':
		l.scanString(c)
	case '(':
		l.addToken(TOKEN_LPAREN)
	case ')':
		l.addToken(TOKEN_RPAREN)
	case '[':
		l.addToken(TOKEN_LBRACKET)
	case ']':
		l.addToken(TOKEN_RBRACKET)
	case '{':
		l.addToken(TOKEN_LBRACE)
	case '}':
		l.addToken(TOKEN_RBRACE)
	case ',':
		l.addToken(TOKEN_COMMA)
	case '.':
		l.addToken(TOKEN_DOT)
	case '+':
		l.addToken(TOKEN_PLUS)
	case '-':
		l.addToken(TOKEN_MINUS)
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
		} else {
			if l.match('|') {
				l.addToken(TOKEN_OR_OR)
			} else {
				l.error("Unexpected character '|'. Did you mean '|>' for pipe operator?")
			}
		}
	case '&':
		if l.match('&') {
			l.addToken(TOKEN_AND_AND)
		} else {
			l.error("Unexpected character '&'. Did you mean '&&'?")
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

// handleIndentation handles indentation at the start of a line
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

// scanString scans a string literal with optional interpolation
func (l *Lexer) scanString(quote rune) {
	value := strings.Builder{}

	for !l.isAtEnd() && l.peek() != quote {
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
		} else if l.peek() == '{' && quote == '"' {
			// String interpolation (only in double-quoted strings)
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
