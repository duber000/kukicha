package lexer

import (
	"fmt"
	"strings"
	"unicode"
	"unique"
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
	atLineStart        bool  // Whether we're at the start of a line
	indentationHandled bool  // Whether indentation has been handled for the current line
	errors             []error

	// Pipe continuation support: a trailing |> at end of line suppresses the
	// NEWLINE token and causes the next line's indentation to be consumed
	// without emitting INDENT/DEDENT, so the pipe RHS is parsed as part of
	// the same expression regardless of how it is indented.
	lastTokenType     TokenType // last emitted token type (TOKEN_COMMENT excluded)
	continuationLine  bool      // true when the next line is a |> continuation
	braceDepth        int       // current nesting level of [], {} (used for continuations)
	parenDepth        int       // current nesting level of () (used for closures)
	inFunctionLiteral bool      // true when we've just seen 'func' and are parsing its body

	// String interpolation support: when scanning a string and encountering
	// {expr}, the lexer emits TOKEN_STRING_HEAD, returns to normal tokenization
	// for the expression, and resumes string scanning when the matching } is found.
	// Each entry in interpStack is the brace depth within that interpolation level.
	interpStack []int
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
		parenDepth:         0,
		inFunctionLiteral:  false,
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
		// 1. We are inside non-paren braces (braceDepth > 0) AND not in a closure
		// 2. The *previous* token was a pipe
		// 3. The *next* token (on the new line) is a pipe
		//
		// NOTE: parenDepth > 0 does NOT suppress indentation when we're in a function
		// literal (closure), because closures need INDENT/DEDENT tokens for their body.
		isLineContinuation := (l.braceDepth > 0) || l.lastTokenType == TOKEN_PIPE || l.isPipeAtStartOfNextLine() || l.isOnErrAtStartOfNextLine()
		if isLineContinuation {
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
		isLineContinuation := (l.braceDepth > 0) || l.lastTokenType == TOKEN_PIPE || l.isPipeAtStartOfNextLine() || l.isOnErrAtStartOfNextLine()
		if isLineContinuation {
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
		l.parenDepth++
		l.addToken(TOKEN_LPAREN)
	case ')':
		if l.parenDepth > 0 {
			l.parenDepth--
		}
		// When closing the parameter list of a function literal (parenDepth becomes 0),
		// we know the next tokens will be the return type annotations and then the body.
		// Keep inFunctionLiteral true; it will be reset when the body block is done.
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
		if len(l.interpStack) > 0 {
			l.interpStack[len(l.interpStack)-1]++
		}
		l.braceDepth++
		l.addToken(TOKEN_LBRACE)
	case '}':
		if len(l.interpStack) > 0 && l.interpStack[len(l.interpStack)-1] == 0 {
			// End of string interpolation expression — resume string scanning
			l.interpStack = l.interpStack[:len(l.interpStack)-1]
			l.start = l.current
			l.scanStringContinuation()
			return
		}
		if len(l.interpStack) > 0 {
			l.interpStack[len(l.interpStack)-1]--
		}
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
		} else if l.match('>') {
			l.addToken(TOKEN_FAT_ARROW)
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
		} else if l.match('=') {
			l.addToken(TOKEN_BIT_AND_ASSIGN)
		} else {
			l.addToken(TOKEN_BIT_AND)
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
		l.error("indentation error: tabs are not allowed — use 4 spaces per indent level")
		return
	}

	// Skip blank lines and comment-only lines
	if l.isAtEnd() || l.peek() == '\n' || l.peek() == '\r' || l.peek() == '#' {
		return
	}

	// Must be multiple of 4
	if spaces%4 != 0 {
		nearest := ((spaces + 2) / 4) * 4
		if nearest == 0 {
			nearest = 4
		}
		l.error(fmt.Sprintf("indentation error: found %d spaces, but Kukicha requires multiples of 4 spaces (nearest valid: %d)", spaces, nearest))
		return
	}

	currentIndent := l.indentStack[len(l.indentStack)-1]

	if spaces > currentIndent {
		// Indent
		if spaces != currentIndent+4 {
			l.error(fmt.Sprintf("indentation error: indentation can only increase by 4 spaces at a time (jumped from %d to %d)", currentIndent, spaces))
			return
		}
		l.indentStack = append(l.indentStack, spaces)
		l.addToken(TOKEN_INDENT)
	} else if spaces < currentIndent {
		// Capture valid levels before popping, for use in the error message.
		validLevels := make([]int, len(l.indentStack))
		copy(validLevels, l.indentStack)

		// Dedent (possibly multiple levels)
		for len(l.indentStack) > 1 && l.indentStack[len(l.indentStack)-1] > spaces {
			l.addToken(TOKEN_DEDENT)
			l.indentStack = l.indentStack[:len(l.indentStack)-1]
		}

		// Verify we landed on a valid indentation level
		if l.indentStack[len(l.indentStack)-1] != spaces {
			parts := make([]string, len(validLevels))
			for i, v := range validLevels {
				parts[i] = fmt.Sprintf("%d", v)
			}
			l.error(fmt.Sprintf("indentation error: dedent does not match any outer indent level (found %d spaces, expected one of: %s)", spaces, strings.Join(parts, ", ")))
		}
	}
}

// scanString scans a double-quoted string literal.
//
// If the next two characters are also '"', this is a triple-quoted multi-line
// string ("""...""") and is handled by scanTripleQuoteString.
//
// For non-interpolated strings, emits a single TOKEN_STRING.
// For interpolated strings (containing {expr}), emits TOKEN_STRING_HEAD for the
// leading literal, then returns so the main scanToken loop can tokenize the
// expression normally. When the matching } is found, scanStringContinuation
// resumes scanning, emitting TOKEN_STRING_MID or TOKEN_STRING_TAIL.
func (l *Lexer) scanString() {
	// Check for triple-quote: "" could be an empty string or the start of """
	if l.peek() == '"' {
		if l.current+1 < len(l.source) && l.source[l.current+1] == '"' {
			// Triple-quote: consume the two extra '"' and scan multi-line
			l.advance() // consume second "
			l.advance() // consume third "
			l.scanTripleQuoteString()
			return
		}
		// Empty string "" — consume closing quote, emit empty TOKEN_STRING
		l.advance()
		l.addTokenWithLexeme(TOKEN_STRING, "")
		return
	}
	l.scanStringBody(TOKEN_STRING_HEAD, TOKEN_STRING)
}

// scanTripleQuoteString scans a triple-quoted string literal ("""...""").
//
// Behavior:
//   - Content spans multiple lines until the closing """
//   - Leading common indentation (dedent) is stripped: the minimum non-empty
//     leading-whitespace count across all content lines is removed from each line
//   - The first newline after the opening """ is not part of the content
//   - The last newline before the closing """ is not part of the content
//   - String interpolation ({expr}) works inside, same as regular strings
//
// The resulting content is passed through the same TOKEN_STRING / TOKEN_STRING_HEAD
// path as regular strings, so the rest of the pipeline (parser, codegen) is unchanged.
func (l *Lexer) scanTripleQuoteString() {
	startLine := l.line // Save start line for the token position
	raw := strings.Builder{}

	for !l.isAtEnd() {
		// Check for closing """
		if l.peek() == '"' && l.current+1 < len(l.source) && l.source[l.current+1] == '"' &&
			l.current+2 < len(l.source) && l.source[l.current+2] == '"' {
			l.advance() // consume first "
			l.advance() // consume second "
			l.advance() // consume third "
			break
		}
		ch := l.advance()
		if ch == '\n' {
			l.line++
			l.column = 0
		}
		raw.WriteRune(ch)
	}

	// Save the post-scan line (after closing """) so we can restore it after
	// the content injection. This ensures tokens following the triple-quote
	// string get the correct line number.
	endLine := l.line
	endColumn := l.column

	content := dedentTripleQuote(raw.String())

	// Set line to start so the emitted string token has the correct position.
	l.line = startLine

	// Now re-scan content string through the interpolation machinery by
	// injecting it as if it were scanned from a regular "..." string.
	l.scanStringFromContent(content)

	// Restore line/column to the position after the closing """ so subsequent
	// tokens get correct positions.
	l.line = endLine
	l.column = endColumn
}

// dedentTripleQuote strips the first newline (if any), the last newline (if any),
// and the common leading indentation from all non-empty lines.
func dedentTripleQuote(raw string) string {
	// Strip leading newline (the one right after opening """)
	if len(raw) > 0 && raw[0] == '\n' {
		raw = raw[1:]
	} else if len(raw) > 1 && raw[0] == '\r' && raw[1] == '\n' {
		raw = raw[2:]
	}

	// Strip trailing newline (the one right before closing """)
	if len(raw) > 0 && raw[len(raw)-1] == '\n' {
		raw = raw[:len(raw)-1]
		if len(raw) > 0 && raw[len(raw)-1] == '\r' {
			raw = raw[:len(raw)-1]
		}
	}

	// Find minimum indentation (spaces only) across non-empty lines
	lines := strings.Split(raw, "\n")
	minIndent := -1
	for _, line := range lines {
		if strings.TrimRight(line, " \t\r") == "" {
			continue // skip blank / whitespace-only lines
		}
		indent := 0
		for _, ch := range line {
			if ch == ' ' {
				indent++
			} else if ch == '\t' {
				indent += 4
			} else {
				break
			}
		}
		if minIndent < 0 || indent < minIndent {
			minIndent = indent
		}
	}
	if minIndent <= 0 {
		return raw
	}

	// Strip minIndent spaces from the front of each line
	out := strings.Builder{}
	for i, line := range lines {
		stripped := 0
		j := 0
		for j < len(line) && stripped < minIndent {
			if line[j] == ' ' {
				stripped++
				j++
			} else if line[j] == '\t' {
				stripped += 4
				j++
			} else {
				break
			}
		}
		out.WriteString(line[j:])
		if i < len(lines)-1 {
			out.WriteRune('\n')
		}
	}
	return out.String()
}

// scanStringFromContent injects pre-extracted string content back into the
// source stream so the existing scanStringBody machinery handles it naturally.
// Bare `"` characters (not preceded by `\`) are escaped to `\"` so they
// don't prematurely terminate the scan. All other escape sequences are kept
// as-is so scanStringEscape can process them normally.
func (l *Lexer) scanStringFromContent(content string) {
	runes := []rune(content)
	inject := make([]rune, 0, len(runes)+4)

	for i := 0; i < len(runes); i++ {
		ch := runes[i]
		if ch == '\\' && i+1 < len(runes) {
			// Keep escape sequence intact (two characters)
			inject = append(inject, ch, runes[i+1])
			i++ // skip next char (already consumed)
		} else if ch == '"' {
			// Bare quote — escape it so scanStringBody doesn't stop here
			inject = append(inject, '\\', '"')
		} else if ch == '\n' {
			// Newline — inject as \n escape so scanStringBody doesn't error
			inject = append(inject, '\\', 'n')
		} else if ch == '\r' {
			// Skip bare CR (CR+LF was already normalized to LF during extraction)
		} else {
			inject = append(inject, ch)
		}
	}
	// Append synthetic closing quote
	inject = append(inject, '"')

	// Splice into l.source at l.current
	newSource := make([]rune, 0, len(l.source)+len(inject))
	newSource = append(newSource, l.source[:l.current]...)
	newSource = append(newSource, inject...)
	newSource = append(newSource, l.source[l.current:]...)
	l.source = newSource
	l.scanStringBody(TOKEN_STRING_HEAD, TOKEN_STRING)
}

// scanStringContinuation resumes string scanning after a } closes an
// interpolation expression. Emits TOKEN_STRING_MID if another interpolation
// follows, or TOKEN_STRING_TAIL at the closing quote.
func (l *Lexer) scanStringContinuation() {
	l.scanStringBody(TOKEN_STRING_MID, TOKEN_STRING_TAIL)
}

// scanStringBody is the shared string scanning logic.
// interpTokenType is emitted when a {expr} interpolation is found (HEAD or MID).
// endTokenType is emitted when the string ends with " (STRING or TAIL).
func (l *Lexer) scanStringBody(interpTokenType TokenType, endTokenType TokenType) {
	value := strings.Builder{}

	for !l.isAtEnd() && l.peek() != '"' {
		if l.peek() == '\n' {
			l.error("Unterminated string")
			return
		}

		if l.peek() == '\\' {
			l.scanStringEscape(&value)
		} else if l.peek() == '{' && l.isInterpStart() {
			// String interpolation: emit the accumulated literal, push
			// interp state, and return so the expression gets tokenized.
			l.advance() // consume '{'
			l.addTokenWithLexeme(interpTokenType, value.String())
			l.interpStack = append(l.interpStack, 0)
			return
		} else {
			value.WriteRune(l.advance())
		}
	}

	if l.isAtEnd() {
		l.error("Unterminated string")
		return
	}

	l.advance() // consume closing quote
	l.addTokenWithLexeme(endTokenType, value.String())
}

// isInterpStart checks whether { at the current position starts a string
// interpolation. Requires the character after { to be an identifier-start
// character (letter or underscore). This avoids treating regex quantifiers
// like {2,} as interpolation.
func (l *Lexer) isInterpStart() bool {
	// peek() is '{', check the character after it
	nextIdx := l.current + 1
	if nextIdx >= len(l.source) {
		return false
	}
	c := l.source[nextIdx]
	return isAlpha(c)
}

// scanStringEscape handles a single escape sequence inside a string literal,
// writing the result into the provided builder.
func (l *Lexer) scanStringEscape(value *strings.Builder) {
	l.advance() // consume '\'
	if l.isAtEnd() {
		return
	}
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
	case '{':
		value.WriteRune('\uE000') // PUA sentinel for literal {
	case '}':
		value.WriteRune('\uE001') // PUA sentinel for literal }
	case 's':
		// Check for \sep (filepath separator)
		if l.peek() == 'e' && l.peekNext() == 'p' {
			l.advance()               // consume 'e'
			l.advance()               // consume 'p'
			value.WriteRune('\uE002') // PUA sentinel for filepath.Separator
		} else {
			value.WriteRune('s')
		}
	case 'x':
		// Hex escape: \xHH
		if !l.isAtEnd() {
			h1 := l.advance()
			if !l.isAtEnd() {
				h2 := l.advance()
				hi, ok1 := hexDigit(h1)
				lo, ok2 := hexDigit(h2)
				if ok1 && ok2 {
					value.WriteRune(rune(hi*16 + lo))
				} else {
					value.WriteString(`\x`)
					value.WriteRune(h1)
					value.WriteRune(h2)
				}
			}
		}
	default:
		value.WriteRune(escaped)
	}
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

	text := unique.Make(string(l.source[l.start:l.current])).Value()
	tokenType := LookupKeyword(text)
	l.addTokenWithLexeme(tokenType, text)

	// Track when we enter a function literal context
	if tokenType == TOKEN_FUNC {
		l.inFunctionLiteral = true
	}
}

// scanComment scans a comment. If the comment starts with "# kuki:", it is
// emitted as TOKEN_DIRECTIVE so the parser can attach it to a declaration.
// Otherwise it is emitted as a regular TOKEN_COMMENT.
func (l *Lexer) scanComment() {
	// Consume the rest of the comment line
	for !l.isAtEnd() && l.peek() != '\n' {
		l.advance()
	}
	// Check if this is a directive comment (# kuki:...)
	lexeme := string(l.source[l.start:l.current])
	if strings.HasPrefix(lexeme, "# kuki:") {
		l.addToken(TOKEN_DIRECTIVE)
	} else {
		l.addToken(TOKEN_COMMENT)
	}
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
	if tokenType != TOKEN_COMMENT && tokenType != TOKEN_DIRECTIVE {
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

func hexDigit(c rune) (int, bool) {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0'), true
	case c >= 'a' && c <= 'f':
		return int(c-'a') + 10, true
	case c >= 'A' && c <= 'F':
		return int(c-'A') + 10, true
	default:
		return 0, false
	}
}

func isAlpha(c rune) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		c == '_'
}

func isAlphaNumeric(c rune) bool {
	return isAlpha(c) || isDigit(c)
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
	idx, indent := l.nextNonWhitespaceWithIndent()
	if indent < l.indentStack[len(l.indentStack)-1] {
		return false
	}
	if idx+1 < len(l.source) && l.source[idx] == '|' && l.source[idx+1] == '>' {
		return true
	}
	return false
}

func (l *Lexer) isOnErrAtStartOfNextLine() bool {
	idx, indent := l.nextNonWhitespaceWithIndent()
	if indent < l.indentStack[len(l.indentStack)-1] {
		return false
	}
	if idx+5 <= len(l.source) && string(l.source[idx:idx+5]) == "onerr" {
		if idx+5 == len(l.source) || !isLetter(l.source[idx+5]) && !isDigit(l.source[idx+5]) {
			return true
		}
	}
	return false
}

func (l *Lexer) nextNonWhitespaceWithIndent() (int, int) {
	idx := l.current
	indent := 0
	for idx < len(l.source) {
		c := l.source[idx]
		if c == ' ' {
			indent++
			idx++
		} else if c == '\t' {
			indent += 4
			idx++
		} else if c == '\n' || c == '\r' {
			idx++
			indent = 0
			if c == '\r' && idx < len(l.source) && l.source[idx] == '\n' {
				idx++
			}
		} else {
			break
		}
	}
	return idx, indent
}
