package parser

import (
	"fmt"
	"slices"

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
	tokens []lexer.Token
	pos    int
	errors []error // Collected errors - parsing continues after errors for better diagnostics
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

// skipIgnoredTokens advances past comments and semicolons
func (p *Parser) skipIgnoredTokens() {
	for p.pos < len(p.tokens) {
		t := p.tokens[p.pos]
		if t.Type == lexer.TOKEN_COMMENT || t.Type == lexer.TOKEN_SEMICOLON {
			p.pos++
		} else {
			break
		}
	}
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
	// Note: skipIgnoredTokens is already called by peekToken/peekNextToken but
	// for safety we don't call it here to avoid complex side effects if used with large offsets.
	// Actually, most peek methods call it.
	if p.pos+offset >= len(p.tokens) {
		return lexer.Token{Type: lexer.TOKEN_EOF}
	}
	return p.tokens[p.pos+offset]
}

func (p *Parser) peekTokenAt(index int) lexer.Token {
	p.skipIgnoredTokens()
	if index >= len(p.tokens) {
		return lexer.Token{Type: lexer.TOKEN_EOF}
	}
	return p.tokens[index]
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

