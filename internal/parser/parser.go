package parser

import (
	"fmt"
	"slices"
	"strings"

	"github.com/duber000/kukicha/internal/ast"
	"github.com/duber000/kukicha/internal/lexer"
)

// Parser parses tokens into an AST using recursive descent.
//
// ARCHITECTURE NOTE: The parser uses error collection (not fail-fast).
// When an error is encountered, it's appended to p.errors and parsing continues.
// This allows reporting multiple errors in a single parse, improving UX.
//
// The parser handles Kukicha's "context-sensitive keywords" - words like
// `list`, `map`, and `channel` are keywords only when followed by `of` in a
// type context. This lets users use these as variable names in expressions.
type Parser struct {
	tokens            []lexer.Token
	pos               int
	errors            []error         // Collected errors - parsing continues after errors for better diagnostics
	pendingDirectives []ast.Directive // Directives collected before the next declaration
}

// New creates a new parser from a source string
func New(source string, filename string) (*Parser, error) {
	l := lexer.NewLexer(source, filename)
	tokens, err := l.ScanTokens()
	if err != nil {
		return nil, err
	}
	return &Parser{
		tokens: tokens,
		pos:    0,
		errors: []error{},
	}, nil
}

// NewFromTokens creates a new parser from a slice of tokens
func NewFromTokens(tokens []lexer.Token) *Parser {
	return &Parser{
		tokens: tokens,
		pos:    0,
		errors: []error{},
	}
}

// Parse parses the tokens into a Program AST
func (p *Parser) Parse() (*ast.Program, []error) {
	program := &ast.Program{
		Imports:      []*ast.ImportDecl{},
		Declarations: []ast.Declaration{},
	}

	// Skip leading newlines (may follow comments at file start)
	p.skipNewlines()

	// Parse optional petiole declaration
	if p.peekToken().Type == lexer.TOKEN_PETIOLE {
		program.PetioleDecl = p.parsePetioleDecl()
	}

	p.skipNewlines()

	// Parse optional skill declaration (simple form: skill name)
	if p.peekToken().Type == lexer.TOKEN_SKILL {
		program.SkillDecl = p.parseSkillDecl()
	}

	p.skipNewlines()

	// Parse imports
	for p.peekToken().Type == lexer.TOKEN_IMPORT {
		program.Imports = append(program.Imports, p.parseImportDecl())
		p.skipNewlines()
	}

	// Parse top-level declarations
	for !p.isAtEnd() {
		if decl := p.parseDeclaration(); decl != nil {
			program.Declarations = append(program.Declarations, decl)
		}
	}

	return program, p.errors
}

// Errors returns the parsing errors
func (p *Parser) Errors() []error {
	return p.errors
}

// ============================================================================
// Helper Methods
// ============================================================================

func (p *Parser) isAtEnd() bool {
	return p.pos >= len(p.tokens) || p.peekToken().Type == lexer.TOKEN_EOF
}

// skipIgnoredTokens advances past comments, semicolons, and collects directives.
// Directive tokens are parsed and accumulated in pendingDirectives for the next declaration.
func (p *Parser) skipIgnoredTokens() {
	for p.pos < len(p.tokens) {
		t := p.tokens[p.pos]
		if t.Type == lexer.TOKEN_COMMENT || t.Type == lexer.TOKEN_SEMICOLON {
			p.pos++
		} else if t.Type == lexer.TOKEN_DIRECTIVE {
			p.pendingDirectives = append(p.pendingDirectives, parseDirective(t))
			p.pos++
		} else {
			break
		}
	}
}

// parseDirective extracts the directive name and arguments from a TOKEN_DIRECTIVE lexeme.
// Format: "# kuki:name arg1 arg2 ..." or "# kuki:name \"quoted arg\""
func parseDirective(t lexer.Token) ast.Directive {
	// Strip "# kuki:" prefix
	content := strings.TrimPrefix(t.Lexeme, "# kuki:")
	content = strings.TrimSpace(content)

	// Split into name and remaining args
	name := content
	var args []string
	if idx := strings.IndexByte(content, ' '); idx >= 0 {
		name = content[:idx]
		argStr := strings.TrimSpace(content[idx+1:])
		if argStr != "" {
			// Parse quoted strings as single args, unquoted as space-split
			args = parseDirectiveArgs(argStr)
		}
	}

	return ast.Directive{
		Token: t,
		Name:  name,
		Args:  args,
	}
}

// parseDirectiveArgs splits a directive argument string, respecting quoted strings.
func parseDirectiveArgs(s string) []string {
	var args []string
	for len(s) > 0 {
		s = strings.TrimLeft(s, " \t")
		if len(s) == 0 {
			break
		}
		if s[0] == '"' {
			// Find closing quote
			end := strings.IndexByte(s[1:], '"')
			if end < 0 {
				args = append(args, s[1:]) // unterminated quote — take rest
				break
			}
			args = append(args, s[1:end+1])
			s = s[end+2:]
		} else {
			end := strings.IndexByte(s, ' ')
			if end < 0 {
				args = append(args, s)
				break
			}
			args = append(args, s[:end])
			s = s[end+1:]
		}
	}
	return args
}

func (p *Parser) peekToken() lexer.Token {
	p.skipIgnoredTokens()
	if p.pos >= len(p.tokens) {
		return lexer.Token{Type: lexer.TOKEN_EOF}
	}
	return p.tokens[p.pos]
}

func (p *Parser) peekNextToken() lexer.Token {
	p.skipIgnoredTokens()
	if p.pos+1 >= len(p.tokens) {
		return lexer.Token{Type: lexer.TOKEN_EOF}
	}
	return p.tokens[p.pos+1]
}

func (p *Parser) peekAt(offset int) lexer.Token {
	// Skip ignored tokens (comments, semicolons, directives) when counting
	// the offset so that comments between meaningful tokens don't break
	// lookahead patterns like struct literal detection.
	p.skipIgnoredTokens()
	i := p.pos
	seen := 0
	for i < len(p.tokens) {
		t := p.tokens[i]
		if t.Type == lexer.TOKEN_COMMENT || t.Type == lexer.TOKEN_SEMICOLON || t.Type == lexer.TOKEN_DIRECTIVE {
			i++
			continue
		}
		if seen == offset {
			return t
		}
		seen++
		i++
	}
	return lexer.Token{Type: lexer.TOKEN_EOF}
}

func (p *Parser) previousToken() lexer.Token {
	if p.pos == 0 {
		return lexer.Token{Type: lexer.TOKEN_EOF}
	}
	return p.tokens[p.pos-1]
}

func (p *Parser) advance() lexer.Token {
	if !p.isAtEnd() {
		p.pos++
	}
	return p.previousToken()
}

func (p *Parser) check(tokenType lexer.TokenType) bool {
	if p.isAtEnd() {
		return false
	}
	return p.peekToken().Type == tokenType
}

func (p *Parser) match(types ...lexer.TokenType) bool {
	if slices.ContainsFunc(types, p.check) {
		p.advance()
		return true
	}
	return false
}

func (p *Parser) consume(tokenType lexer.TokenType, message string) (lexer.Token, error) {
	if p.check(tokenType) {
		return p.advance(), nil
	}
	err := p.error(p.peekToken(), message)
	return lexer.Token{}, err
}

func (p *Parser) error(token lexer.Token, message string) error {
	err := fmt.Errorf("%s:%d:%d: %s", token.File, token.Line, token.Column, message)
	p.errors = append(p.errors, err)
	return err
}

func (p *Parser) skipNewlines() {
	for p.match(lexer.TOKEN_NEWLINE) {
	}
}

// drainDirectives returns any pending directives and clears the buffer.
func (p *Parser) drainDirectives() []ast.Directive {
	if len(p.pendingDirectives) == 0 {
		return nil
	}
	dirs := p.pendingDirectives
	p.pendingDirectives = nil
	return dirs
}

// isIdentifierFollower returns true if the next token indicates that the current
// token (empty/error) is being used as an identifier rather than a keyword.
// Tokens that follow identifiers: assignment, postfix, member access, indexing,
// operators, delimiters, end-of-line.
func (p *Parser) isIdentifierFollower() bool {
	next := p.peekNextToken().Type
	switch next {
	case lexer.TOKEN_WALRUS, lexer.TOKEN_ASSIGN,
		lexer.TOKEN_BIT_AND, lexer.TOKEN_BIT_AND_ASSIGN,
		lexer.TOKEN_DOT, lexer.TOKEN_LBRACKET,
		lexer.TOKEN_COMMA, lexer.TOKEN_RPAREN, lexer.TOKEN_RBRACKET, lexer.TOKEN_RBRACE,
		lexer.TOKEN_PLUS_PLUS, lexer.TOKEN_MINUS_MINUS,
		lexer.TOKEN_SEMICOLON,
		lexer.TOKEN_STRING_MID, lexer.TOKEN_STRING_TAIL:
		return true
	default:
		return false
	}
}
