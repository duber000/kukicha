package parser

import (
	"fmt"
	"strconv"
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
	for _, t := range types {
		if p.check(t) {
			p.advance()
			return true
		}
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

// ============================================================================
// Declaration Parsing
// ============================================================================

func (p *Parser) parsePetioleDecl() *ast.PetioleDecl {
	token := p.advance() // consume 'petiole'
	p.skipNewlines()

	name := p.parseIdentifier()
	p.skipNewlines()

	return &ast.PetioleDecl{
		Token: token,
		Name:  name,
	}
}

func (p *Parser) parseImportDecl() *ast.ImportDecl {
	token := p.advance() // consume 'import'
	p.skipNewlines()

	pathToken := p.advance()
	if pathToken.Type != lexer.TOKEN_STRING {
		p.error(pathToken, "expected string literal for import path")
		return nil
	}

	decl := &ast.ImportDecl{
		Token: token,
		Path: &ast.StringLiteral{
			Token: pathToken,
			Value: pathToken.Lexeme,
		},
	}

	// Check for optional alias (as Name)
	if p.match(lexer.TOKEN_AS) {
		decl.Alias = p.parseIdentifier()
	}

	p.skipNewlines()
	return decl
}

func (p *Parser) parseDeclaration() ast.Declaration {
	p.skipNewlines()

	switch p.peekToken().Type {
	case lexer.TOKEN_TYPE:
		return p.parseTypeDecl()
	case lexer.TOKEN_INTERFACE:
		return p.parseInterfaceDecl()
	case lexer.TOKEN_FUNC:
		return p.parseFunctionDecl()
	default:
		if !p.isAtEnd() {
			p.error(p.peekToken(), fmt.Sprintf("unexpected token %s, expected declaration", p.peekToken().Type))
			p.advance() // Skip the problematic token
		}
		return nil
	}
}

func (p *Parser) parseTypeDecl() ast.Declaration {
	token := p.advance() // consume 'type'
	p.skipNewlines()

	name := p.parseIdentifier()
	p.skipNewlines()

	fields := []*ast.FieldDecl{}

	// Expect INDENT for fields
	if !p.match(lexer.TOKEN_INDENT) {
		p.error(p.peekToken(), "expected indented block for type fields")
		return nil
	}

	// Parse fields
	for !p.check(lexer.TOKEN_DEDENT) && !p.isAtEnd() {
		p.skipNewlines()
		if p.check(lexer.TOKEN_DEDENT) {
			break
		}

		fieldName := p.parseIdentifier()
		fieldType := p.parseTypeAnnotation()

		// Parse optional struct tag (e.g., json:"name")
		tag := p.parseStructTag()

		fields = append(fields, &ast.FieldDecl{
			Name: fieldName,
			Type: fieldType,
			Tag:  tag,
		})
		p.skipNewlines()
	}

	p.consume(lexer.TOKEN_DEDENT, "expected dedent after type fields")
	p.skipNewlines()

	return &ast.TypeDecl{
		Token:  token,
		Name:   name,
		Fields: fields,
	}
}

func (p *Parser) parseInterfaceDecl() *ast.InterfaceDecl {
	token := p.advance() // consume 'interface'
	p.skipNewlines()

	name := p.parseIdentifier()
	p.skipNewlines()

	methods := []*ast.MethodSignature{}

	// Expect INDENT for methods
	if !p.match(lexer.TOKEN_INDENT) {
		p.error(p.peekToken(), "expected indented block for interface methods")
		return nil
	}

	// Parse method signatures
	for !p.check(lexer.TOKEN_DEDENT) && !p.isAtEnd() {
		p.skipNewlines()
		if p.check(lexer.TOKEN_DEDENT) {
			break
		}

		methodName := p.parseIdentifier()

		// Parse parameters
		p.consume(lexer.TOKEN_LPAREN, "expected '(' for method parameters")
		params := p.parseParameters()
		p.consume(lexer.TOKEN_RPAREN, "expected ')' after method parameters")

		// Parse return types
		returns := []ast.TypeAnnotation{}
		if !p.check(lexer.TOKEN_NEWLINE) && !p.check(lexer.TOKEN_DEDENT) {
			returns = p.parseReturnTypes()
		}

		methods = append(methods, &ast.MethodSignature{
			Name:       methodName,
			Parameters: params,
			Returns:    returns,
		})
		p.skipNewlines()
	}

	p.consume(lexer.TOKEN_DEDENT, "expected dedent after interface methods")
	p.skipNewlines()

	return &ast.InterfaceDecl{
		Token:   token,
		Name:    name,
		Methods: methods,
	}
}

func (p *Parser) parseFunctionDecl() *ast.FunctionDecl {
	token := p.advance() // consume 'func'
	p.skipNewlines()

	decl := &ast.FunctionDecl{
		Token: token,
	}

	// Parse function name
	decl.Name = p.parseIdentifier()

	// Check for receiver (method declaration): func Name on receiverName Type
	if p.match(lexer.TOKEN_ON) {
		receiverName := p.parseIdentifier()
		receiverType := p.parseTypeAnnotation()
		decl.Receiver = &ast.Receiver{
			Name: receiverName,
			Type: receiverType,
		}
	}

	// Parse parameters (optional for methods with no parameters)
	if p.check(lexer.TOKEN_LPAREN) {
		p.advance() // consume '('
		decl.Parameters = p.parseParameters()
		p.consume(lexer.TOKEN_RPAREN, "expected ')' after function parameters")
	} else {
		decl.Parameters = []*ast.Parameter{}
	}

	// Parse return types
	if !p.check(lexer.TOKEN_NEWLINE) && !p.check(lexer.TOKEN_INDENT) {
		decl.Returns = p.parseReturnTypes()
	}

	p.skipNewlines()

	// Parse function body
	decl.Body = p.parseBlock()

	return decl
}

func (p *Parser) parseParameters() []*ast.Parameter {
	params := []*ast.Parameter{}
	hasDefaultValue := false // Track if we've seen a parameter with a default value

	if p.check(lexer.TOKEN_RPAREN) {
		return params
	}

	for {
		// Check for 'many' keyword (variadic parameter)
		variadic := false
		if p.check(lexer.TOKEN_MANY) {
			p.advance()
			variadic = true
		}

		paramName := p.parseIdentifier()

		// Type is optional for untyped variadic (many values)
		var paramType ast.TypeAnnotation
		if !p.check(lexer.TOKEN_COMMA) && !p.check(lexer.TOKEN_RPAREN) && !p.check(lexer.TOKEN_ASSIGN) {
			paramType = p.parseTypeAnnotation()
		}

		// Default untyped variadic to interface{}
		if variadic && paramType == nil {
			paramType = &ast.NamedType{
				Token: p.peekToken(),
				Name:  "interface{}",
			}
		}

		// Check for default value (e.g., count int = 10)
		var defaultValue ast.Expression
		if p.match(lexer.TOKEN_ASSIGN) {
			defaultValue = p.parseExpression()
			hasDefaultValue = true
		} else if hasDefaultValue {
			// Parameters with defaults must come after those without
			p.addError("parameter '%s' must have a default value (parameters with defaults must be contiguous at the end)", paramName.Value)
		}

		// Variadic parameters cannot have default values
		if variadic && defaultValue != nil {
			p.addError("variadic parameter '%s' cannot have a default value", paramName.Value)
		}

		params = append(params, &ast.Parameter{
			Name:         paramName,
			Type:         paramType,
			Variadic:     variadic,
			DefaultValue: defaultValue,
		})

		if !p.match(lexer.TOKEN_COMMA) {
			break
		}
	}

	return params
}

func (p *Parser) parseReturnTypes() []ast.TypeAnnotation {
	returns := []ast.TypeAnnotation{}

	// Single return type or multiple in parentheses
	if p.check(lexer.TOKEN_LPAREN) {
		p.advance() // consume '('
		for {
			returns = append(returns, p.parseTypeAnnotation())
			if !p.match(lexer.TOKEN_COMMA) {
				break
			}
		}
		p.consume(lexer.TOKEN_RPAREN, "expected ')' after return types")
	} else {
		returns = append(returns, p.parseTypeAnnotation())
	}

	return returns
}

// parseCallArguments parses function call arguments, supporting both positional
// and named arguments. Named arguments use the syntax: name: value
// Returns (positionalArgs, namedArgs, variadic)
func (p *Parser) parseCallArguments() ([]ast.Expression, []*ast.NamedArgument, bool) {
	args := []ast.Expression{}
	namedArgs := []*ast.NamedArgument{}
	variadic := false
	hasNamedArg := false

	if p.check(lexer.TOKEN_RPAREN) {
		return args, namedArgs, variadic
	}

	for {
		// Check for 'many' keyword (variadic argument)
		if p.match(lexer.TOKEN_MANY) {
			variadic = true
		}

		// Check if this is a named argument: identifier followed by colon
		// We need to look ahead to see if this is "name: value" syntax
		if p.check(lexer.TOKEN_IDENTIFIER) && p.peekNextToken().Type == lexer.TOKEN_COLON {
			// Named argument
			nameToken := p.advance()                  // consume identifier
			p.advance()                               // consume colon
			value := p.parseExpression()              // parse value
			namedArgs = append(namedArgs, &ast.NamedArgument{
				Token: nameToken,
				Name:  &ast.Identifier{Token: nameToken, Value: nameToken.Lexeme},
				Value: value,
			})
			hasNamedArg = true
		} else {
			// Positional argument
			if hasNamedArg {
				p.addError("positional argument cannot follow named argument")
			}
			args = append(args, p.parseExpression())
		}

		if !p.match(lexer.TOKEN_COMMA) {
			break
		}
	}

	return args, namedArgs, variadic
}

// ============================================================================
// Type Annotation Parsing
// ============================================================================

// parseTypeAnnotation parses Kukicha type syntax into AST TypeAnnotation nodes.
//
// ARCHITECTURE NOTE: This is where Kukicha's beginner-friendly type syntax
// is parsed. The English-like syntax maps to Go types:
//
//   Kukicha                   Go
//   -------                   --
//   list of string            []string
//   map of string to int      map[string]int
//   reference User            *User
//   channel of int            chan int
//   func(int) bool            func(int) bool
//
// Keywords `list`, `map`, `channel` are context-sensitive: they're only
// treated as type keywords when followed by `of`. This allows using them
// as variable names elsewhere (e.g., `list := getData()`).
func (p *Parser) parseTypeAnnotation() ast.TypeAnnotation {
	switch p.peekToken().Type {
	case lexer.TOKEN_REFERENCE:
		token := p.advance()
		elementType := p.parseTypeAnnotation()
		return &ast.ReferenceType{
			Token:       token,
			ElementType: elementType,
		}

	case lexer.TOKEN_LIST:
		token := p.advance()
		p.consume(lexer.TOKEN_OF, "expected 'of' after 'list'")
		elementType := p.parseTypeAnnotation()
		return &ast.ListType{
			Token:       token,
			ElementType: elementType,
		}

	case lexer.TOKEN_MAP:
		token := p.advance()
		p.consume(lexer.TOKEN_OF, "expected 'of' after 'map'")
		keyType := p.parseTypeAnnotation()
		p.consume(lexer.TOKEN_TO, "expected 'to' after map key type")
		valueType := p.parseTypeAnnotation()
		return &ast.MapType{
			Token:     token,
			KeyType:   keyType,
			ValueType: valueType,
		}

	case lexer.TOKEN_CHANNEL:
		token := p.advance()
		p.consume(lexer.TOKEN_OF, "expected 'of' after 'channel'")
		elementType := p.parseTypeAnnotation()
		return &ast.ChannelType{
			Token:       token,
			ElementType: elementType,
		}

	case lexer.TOKEN_FUNC:
		token := p.advance()
		p.consume(lexer.TOKEN_LPAREN, "expected '(' after 'func'")

		// Parse parameter types
		var parameters []ast.TypeAnnotation
		if p.peekToken().Type != lexer.TOKEN_RPAREN {
			parameters = append(parameters, p.parseTypeAnnotation())
			for p.peekToken().Type == lexer.TOKEN_COMMA {
				p.advance() // consume comma
				parameters = append(parameters, p.parseTypeAnnotation())
			}
		}

		p.consume(lexer.TOKEN_RPAREN, "expected ')' after function parameters")

		// Parse return types
		var returns []ast.TypeAnnotation
		if p.peekToken().Type != lexer.TOKEN_NEWLINE &&
			p.peekToken().Type != lexer.TOKEN_COMMA &&
			p.peekToken().Type != lexer.TOKEN_RPAREN &&
			p.peekToken().Type != lexer.TOKEN_EOF {
			returns = append(returns, p.parseTypeAnnotation())
		}

		return &ast.FunctionType{
			Token:      token,
			Parameters: parameters,
			Returns:    returns,
		}

	case lexer.TOKEN_ERROR:
		// Special case: 'error' is a keyword but also a valid type name
		token := p.advance()
		return &ast.NamedType{
			Token: token,
			Name:  "error",
		}

	case lexer.TOKEN_IDENTIFIER:
		token := p.advance()
		// Check for primitive types
		switch token.Lexeme {
		case "int", "int8", "int16", "int32", "int64",
			"uint", "uint8", "uint16", "uint32", "uint64",
			"float32", "float64", "string", "bool", "byte", "rune":
			return &ast.PrimitiveType{
				Token: token,
				Name:  token.Lexeme,
			}
		default:
			// Check for qualified type (package.Type)
			name := token.Lexeme
			if p.peekToken().Type == lexer.TOKEN_DOT {
				p.advance() // consume DOT
				typeIdent, _ := p.consume(lexer.TOKEN_IDENTIFIER, "expected type name after '.'")
				name = name + "." + typeIdent.Lexeme
			}
			return &ast.NamedType{
				Token: token,
				Name:  name,
			}
		}

	default:
		p.error(p.peekToken(), fmt.Sprintf("expected type annotation, got %s", p.peekToken().Type))
		return nil
	}
}

// ============================================================================
// Statement Parsing
// ============================================================================

func (p *Parser) parseBlock() *ast.BlockStmt {
	token := p.peekToken()
	statements := []ast.Statement{}

	if !p.match(lexer.TOKEN_INDENT) {
		p.error(token, "expected indented block")
		return &ast.BlockStmt{Token: token, Statements: statements}
	}

	for !p.check(lexer.TOKEN_DEDENT) && !p.isAtEnd() {
		p.skipNewlines()
		if p.check(lexer.TOKEN_DEDENT) {
			break
		}
		if stmt := p.parseStatement(); stmt != nil {
			statements = append(statements, stmt)
		}
	}

	p.consume(lexer.TOKEN_DEDENT, "expected dedent after block")

	return &ast.BlockStmt{
		Token:      token,
		Statements: statements,
	}
}

func (p *Parser) parseStatement() ast.Statement {
	p.skipNewlines()

	switch p.peekToken().Type {
	case lexer.TOKEN_RETURN:
		return p.parseReturnStmt()
	case lexer.TOKEN_IF:
		return p.parseIfStmt()
	case lexer.TOKEN_FOR:
		return p.parseForStmt()
	case lexer.TOKEN_DEFER:
		return p.parseDeferStmt()
	case lexer.TOKEN_GO:
		return p.parseGoStmt()
	case lexer.TOKEN_SEND:
		return p.parseSendStmt()
	case lexer.TOKEN_CONTINUE:
		return p.parseContinueStmt()
	case lexer.TOKEN_BREAK:
		return p.parseBreakStmt()
	default:
		return p.parseExpressionOrAssignmentStmt()
	}
}

func (p *Parser) parseReturnStmt() *ast.ReturnStmt {
	token := p.advance() // consume 'return'

	stmt := &ast.ReturnStmt{
		Token:  token,
		Values: []ast.Expression{},
	}

	// Check if there are return values
	if !p.check(lexer.TOKEN_NEWLINE) && !p.check(lexer.TOKEN_DEDENT) && !p.isAtEnd() {
		for {
			stmt.Values = append(stmt.Values, p.parseExpression())
			if !p.match(lexer.TOKEN_COMMA) {
				break
			}
		}
	}

	p.skipNewlines()
	return stmt
}

func (p *Parser) parseContinueStmt() *ast.ContinueStmt {
	token := p.advance()
	p.skipNewlines()
	return &ast.ContinueStmt{Token: token}
}

func (p *Parser) parseBreakStmt() *ast.BreakStmt {
	token := p.advance()
	p.skipNewlines()
	return &ast.BreakStmt{Token: token}
}

func (p *Parser) parseIfStmt() *ast.IfStmt {
	token := p.advance() // consume 'if'

	// Look ahead for if-init: if x := 1; x > 0
	var init ast.Statement
	var condition ast.Expression

	// We try to parse an expression or assignment.
	// If it's followed by a semicolon, it's an init statement.
	savePos := p.pos

	// Support both declarations (x := 1) and assignments (x = 1)
	// parseExpressionOrAssignmentStmt is appropriate but it usually consumes the newline.
	// Let's peek ahead for semicolon manually.

	expr := p.parseExpression()

	if p.match(lexer.TOKEN_SEMICOLON) {
		// It's an init statement. Convert expr to a statement.
		// If it's a binary expression with '=', it's an assignment.
		// If it's a walrus, it's a declaration.
		// But parseExpression already handled those?
		// Actually assignment is a statement in Kukicha, not an expression.
		// So parseExpression would have failed if it was an assignment.

		// Let's try again with a more direct approach.
		p.pos = savePos

		// We peek ahead for the semicolon to decide if we parse a statement first.
		hasSemicolon := false
		depth := 0
		for i := p.pos; i < len(p.tokens); i++ {
			t := p.tokens[i].Type
			if t == lexer.TOKEN_NEWLINE || t == lexer.TOKEN_EOF || t == lexer.TOKEN_INDENT || t == lexer.TOKEN_DEDENT {
				break
			}
			if t == lexer.TOKEN_LPAREN {
				depth++
			} else if t == lexer.TOKEN_RPAREN {
				depth--
			} else if t == lexer.TOKEN_SEMICOLON && depth == 0 {
				hasSemicolon = true
				break
			}
		}

		if hasSemicolon {
			// Parse it as a statement, but WITHOUT consuming the newline/dedent
			// We need a version of parseStatement that doesn't expect a newline if followed by ;
			// For now, let's just parse the expressionOrAssignment and then the semicolon.
			init = p.parseExpressionOrAssignmentStmt()
			// parseExpressionOrAssignmentStmt doesn't consume the semicolon if it was treated as stmt separator
			// But here it's an init separator.
			if p.previousToken().Type != lexer.TOKEN_SEMICOLON {
				p.match(lexer.TOKEN_SEMICOLON)
			}
			condition = p.parseExpression()
		} else {
			condition = expr
		}
	} else {
		condition = expr
	}

	stmt := &ast.IfStmt{
		Token:     token,
		Init:      init,
		Condition: condition,
	}

	p.skipNewlines()
	stmt.Consequence = p.parseBlock()
	p.skipNewlines()

	// Check for else/else if
	if p.check(lexer.TOKEN_ELSE) {
		elseToken := p.advance()
		p.skipNewlines()

		// Check for else if
		if p.check(lexer.TOKEN_IF) {
			stmt.Alternative = p.parseIfStmt()
		} else {
			stmt.Alternative = &ast.ElseStmt{
				Token: elseToken,
				Body:  p.parseBlock(),
			}
		}
	}

	p.skipNewlines()
	return stmt
}

func (p *Parser) parseForStmt() ast.Statement {
	token := p.advance() // consume 'for'

	// Look ahead to determine which type of for loop
	// for
	// for item in collection
	// for index, item in collection
	// for i from start to/through end
	// for condition

	// Bare for loop: for \n
	if p.check(lexer.TOKEN_NEWLINE) || p.check(lexer.TOKEN_INDENT) {
		p.skipNewlines()
		body := p.parseBlock()
		return &ast.ForConditionStmt{
			Token:     token,
			Condition: &ast.BooleanLiteral{Token: token, Value: true},
			Body:      body,
		}
	}

	savePos := p.pos

	if p.match(lexer.TOKEN_IDENTIFIER) {
		firstIdentToken := p.previousToken()
		firstIdent := &ast.Identifier{Token: firstIdentToken, Value: firstIdentToken.Lexeme}

		if p.match(lexer.TOKEN_IN) {
			// for item in collection
			collection := p.parseExpression()
			p.skipNewlines()
			body := p.parseBlock()
			return &ast.ForRangeStmt{
				Token:      token,
				Variable:   firstIdent,
				Collection: collection,
				Body:       body,
			}
		} else if p.match(lexer.TOKEN_COMMA) {
			// for index, item in collection
			secondIdent := p.parseIdentifier()
			p.consume(lexer.TOKEN_IN, "expected 'in' after variable list")
			collection := p.parseExpression()
			p.skipNewlines()
			body := p.parseBlock()
			return &ast.ForRangeStmt{
				Token:      token,
				Index:      firstIdent,
				Variable:   secondIdent,
				Collection: collection,
				Body:       body,
			}
		} else if p.match(lexer.TOKEN_FROM) {
			// for i from start to/through end
			startExpr := p.parseExpression()
			through := false
			if p.match(lexer.TOKEN_THROUGH) {
				through = true
			} else {
				p.consume(lexer.TOKEN_TO, "expected 'to' or 'through' after start value")
			}
			endExpr := p.parseExpression()
			p.skipNewlines()
			body := p.parseBlock()
			return &ast.ForNumericStmt{
				Token:    token,
				Variable: firstIdent,
				Start:    startExpr,
				End:      endExpr,
				Through:  through,
				Body:     body,
			}
		}
	}

	// Backtrack and parse as condition-based for loop
	p.pos = savePos
	condition := p.parseExpression()
	p.skipNewlines()
	body := p.parseBlock()
	return &ast.ForConditionStmt{
		Token:     token,
		Condition: condition,
		Body:      body,
	}
}

func (p *Parser) parseDeferStmt() *ast.DeferStmt {
	token := p.advance() // consume 'defer'

	expr := p.parseExpression()

	// Accept both regular function calls and method calls
	switch call := expr.(type) {
	case *ast.CallExpr:
		p.skipNewlines()
		return &ast.DeferStmt{
			Token: token,
			Call:  call,
		}
	case *ast.MethodCallExpr:
		// Use MethodCallExpr directly - no wrapping needed
		p.skipNewlines()
		return &ast.DeferStmt{
			Token: token,
			Call:  call,
		}
	default:
		p.error(token, "defer must be followed by a function call")
		return nil
	}
}

func (p *Parser) parseGoStmt() *ast.GoStmt {
	token := p.advance() // consume 'go'

	expr := p.parseExpression()

	// Accept both regular function calls and method calls
	switch call := expr.(type) {
	case *ast.CallExpr:
		p.skipNewlines()
		return &ast.GoStmt{
			Token: token,
			Call:  call,
		}
	case *ast.MethodCallExpr:
		// Use MethodCallExpr directly - no wrapping needed
		p.skipNewlines()
		return &ast.GoStmt{
			Token: token,
			Call:  call,
		}
	default:
		p.error(token, "go must be followed by a function call")
		return nil
	}
}

func (p *Parser) parseSendStmt() *ast.SendStmt {
	token := p.advance() // consume 'send'

	value := p.parseExpression()
	p.consume(lexer.TOKEN_TO, "expected 'to' after value in send statement")
	channel := p.parseExpression()

	p.skipNewlines()
	return &ast.SendStmt{
		Token:   token,
		Value:   value,
		Channel: channel,
	}
}

func (p *Parser) parseExpressionOrAssignmentStmt() ast.Statement {
	// Check if we have a multi-value assignment pattern
	if p.checkMultiValueAssignment() {
		return p.parseMultiValueAssignmentStmt()
	}

	expr := p.parseExpression()

	// Check for increment/decrement operators
	if p.match(lexer.TOKEN_PLUS_PLUS, lexer.TOKEN_MINUS_MINUS) {
		operator := p.previousToken()
		p.skipNewlines()
		return &ast.IncDecStmt{
			Token:    operator,
			Variable: expr,
			Operator: operator.Lexeme,
		}
	}

	// Check for assignment or walrus operator
	if p.match(lexer.TOKEN_ASSIGN) {
		// Regular assignment: x = value
		values := p.parseExpressionList()
		stmt := &ast.AssignStmt{
			Targets: []ast.Expression{expr},
			Values:  values,
			Token:   p.previousToken(),
		}
		// Check for onerr clause
		if p.check(lexer.TOKEN_ONERR) {
			stmt.OnErr = p.parseOnErrClause()
		}
		p.skipNewlines()
		return stmt
	} else if p.match(lexer.TOKEN_WALRUS) {
		// Variable declaration with inference: x := value
		ident, ok := expr.(*ast.Identifier)
		if !ok {
			p.error(p.previousToken(), "walrus operator can only be used with identifiers")
			return nil
		}
		values := p.parseExpressionList()
		stmt := &ast.VarDeclStmt{
			Names:  []*ast.Identifier{ident},
			Values: values,
			Token:  p.previousToken(),
		}
		// Check for onerr clause
		if p.check(lexer.TOKEN_ONERR) {
			stmt.OnErr = p.parseOnErrClause()
		}
		p.skipNewlines()
		return stmt
	}

	// ExpressionStmt — check for onerr clause
	if p.check(lexer.TOKEN_ONERR) {
		onErr := p.parseOnErrClause()
		p.skipNewlines()
		return &ast.ExpressionStmt{Expression: expr, OnErr: onErr}
	}

	p.skipNewlines()
	return &ast.ExpressionStmt{Expression: expr}
}

func (p *Parser) checkMultiValueAssignment() bool {
	// Look ahead to see if we have a pattern like: ident, ident := expr, expr
	// or: ident, ident = expr, expr

	// Check if we have an identifier at current position
	currentToken := p.peekToken()
	if currentToken.Type != lexer.TOKEN_IDENTIFIER {
		return false
	}

	// Helper function to skip ignored tokens and get next significant token
	skipIgnored := func(startIdx int) (int, lexer.Token) {
		idx := startIdx
		for idx < len(p.tokens) {
			tok := p.tokens[idx]
			if tok.Type != lexer.TOKEN_NEWLINE && tok.Type != lexer.TOKEN_COMMENT {
				return idx, tok
			}
			idx++
		}
		return idx, lexer.Token{Type: lexer.TOKEN_EOF}
	}

	// Look for pattern: identifier, identifier := or =
	// Start from pos + 1 (next token after current identifier)
	nextIdx, nextToken := skipIgnored(p.pos + 1)
	if nextToken.Type != lexer.TOKEN_COMMA {
		return false
	}

	// After comma, should be another identifier
	afterCommaIdx, afterCommaToken := skipIgnored(nextIdx + 1)
	if afterCommaToken.Type != lexer.TOKEN_IDENTIFIER {
		return false
	}

	// After second identifier, should be assignment operator
	_, assignmentToken := skipIgnored(afterCommaIdx + 1)
	if assignmentToken.Type != lexer.TOKEN_ASSIGN && assignmentToken.Type != lexer.TOKEN_WALRUS {
		return false
	}

	return true
}

func (p *Parser) parseMultiValueAssignmentStmt() ast.Statement {
	// Parse left-hand side (comma-separated identifiers)
	var names []*ast.Identifier
	var targets []ast.Expression

	// Parse first identifier
	if !p.match(lexer.TOKEN_IDENTIFIER) {
		p.error(p.peekToken(), "expected identifier in multi-value assignment")
		return nil
	}
	firstIdent := p.previousToken()
	firstName := &ast.Identifier{
		Token: firstIdent,
		Value: firstIdent.Lexeme,
	}
	names = append(names, firstName)
	targets = append(targets, firstName)

	// Parse additional identifiers separated by commas
	for p.match(lexer.TOKEN_COMMA) {
		if !p.match(lexer.TOKEN_IDENTIFIER) {
			p.error(p.peekToken(), "expected identifier after comma in multi-value assignment")
			return nil
		}
		identToken := p.previousToken()
		name := &ast.Identifier{
			Token: identToken,
			Value: identToken.Lexeme,
		}
		names = append(names, name)
		targets = append(targets, name)
	}

	// Check for assignment operator
	if p.match(lexer.TOKEN_WALRUS) {
		// Multi-value declaration: x, y := expr, expr
		values := p.parseExpressionList()
		stmt := &ast.VarDeclStmt{
			Names:  names,
			Values: values,
			Token:  p.previousToken(),
		}
		// Check for onerr clause
		if p.check(lexer.TOKEN_ONERR) {
			stmt.OnErr = p.parseOnErrClause()
		}
		p.skipNewlines()
		return stmt
	} else if p.match(lexer.TOKEN_ASSIGN) {
		// Multi-value assignment: x, y = expr, expr
		values := p.parseExpressionList()
		stmt := &ast.AssignStmt{
			Targets: targets,
			Values:  values,
			Token:   p.previousToken(),
		}
		// Check for onerr clause
		if p.check(lexer.TOKEN_ONERR) {
			stmt.OnErr = p.parseOnErrClause()
		}
		p.skipNewlines()
		return stmt
	} else {
		p.error(p.peekToken(), "expected assignment operator (= or :=) in multi-value assignment")
		return nil
	}
}

func (p *Parser) restorePosition(pos int) {
	p.pos = pos
}

// parseExpressionList parses a comma-separated list of expressions
// This is used for multi-value assignments like: x, y := 1, 2
// or function calls that return multiple values: x, y := iter.Pull(seq)
func (p *Parser) parseExpressionList() []ast.Expression {
	var expressions []ast.Expression

	// Parse first expression
	expressions = append(expressions, p.parseExpression())

	// Parse additional expressions separated by commas
	for p.match(lexer.TOKEN_COMMA) {
		expressions = append(expressions, p.parseExpression())
	}

	return expressions
}

// ============================================================================
// Expression Parsing with Operator Precedence
// ============================================================================

// Precedence levels (lowest to highest):
// 1. or
// 2. pipe (|>)
// 3. and
// 4. comparison (==, !=, <, >, <=, >=)
// 5. additive (+, -)
// 6. multiplicative (*, /, %)
// 7. unary (not, -)
// 8. postfix (call, index, slice, method call)
// 9. primary
//
// Note: onerr is NOT an expression operator. It is a statement-level clause
// attached to VarDeclStmt, AssignStmt, or ExpressionStmt.

func (p *Parser) parseExpression() ast.Expression {
	return p.parseOrExpr()
}

func (p *Parser) parseOrExpr() ast.Expression {
	left := p.parsePipeExpr()

	for p.match(lexer.TOKEN_OR) {
		operator := p.previousToken()
		right := p.parsePipeExpr()
		left = &ast.BinaryExpr{
			Token:    operator,
			Left:     left,
			Operator: operator.Lexeme,
			Right:    right,
		}
	}

	return left
}

func (p *Parser) parsePipeExpr() ast.Expression {
	left := p.parseAndExpr()

	for p.match(lexer.TOKEN_PIPE) {
		operator := p.previousToken()
		right := p.parseAndExpr()
		left = &ast.PipeExpr{
			Token: operator,
			Left:  left,
			Right: right,
		}
	}

	return left
}

// parseOnErrClause parses the onerr clause after a statement.
// Called when TOKEN_ONERR has already been detected (but not consumed).
func (p *Parser) parseOnErrClause() *ast.OnErrClause {
	token := p.advance() // consume 'onerr'
	handler := p.parseExpression()
	return &ast.OnErrClause{Token: token, Handler: handler}
}

func (p *Parser) parseAndExpr() ast.Expression {
	left := p.parseBitwiseOrExpr()

	for p.match(lexer.TOKEN_AND) {
		operator := p.previousToken()
		right := p.parseBitwiseOrExpr()
		left = &ast.BinaryExpr{
			Token:    operator,
			Left:     left,
			Operator: operator.Lexeme,
			Right:    right,
		}
	}

	return left
}

func (p *Parser) parseBitwiseOrExpr() ast.Expression {
	left := p.parseComparisonExpr()

	for p.match(lexer.TOKEN_BIT_OR) {
		operator := p.previousToken()
		right := p.parseComparisonExpr()
		left = &ast.BinaryExpr{
			Token:    operator,
			Left:     left,
			Operator: operator.Lexeme,
			Right:    right,
		}
	}

	return left
}

func (p *Parser) parseComparisonExpr() ast.Expression {
	left := p.parseAdditiveExpr()

	for p.match(lexer.TOKEN_DOUBLE_EQUALS, lexer.TOKEN_NOT_EQUALS, lexer.TOKEN_LT, lexer.TOKEN_GT, lexer.TOKEN_LTE, lexer.TOKEN_GTE) {
		operator := p.previousToken()
		right := p.parseAdditiveExpr()
		left = &ast.BinaryExpr{
			Token:    operator,
			Left:     left,
			Operator: operator.Lexeme,
			Right:    right,
		}
	}

	return left
}

func (p *Parser) parseAdditiveExpr() ast.Expression {
	left := p.parseMultiplicativeExpr()

	for p.match(lexer.TOKEN_PLUS, lexer.TOKEN_MINUS) {
		operator := p.previousToken()
		right := p.parseMultiplicativeExpr()
		left = &ast.BinaryExpr{
			Token:    operator,
			Left:     left,
			Operator: operator.Lexeme,
			Right:    right,
		}
	}

	return left
}

func (p *Parser) parseMultiplicativeExpr() ast.Expression {
	left := p.parseUnaryExpr()

	for p.match(lexer.TOKEN_STAR, lexer.TOKEN_SLASH, lexer.TOKEN_PERCENT) {
		operator := p.previousToken()
		right := p.parseUnaryExpr()
		left = &ast.BinaryExpr{
			Token:    operator,
			Left:     left,
			Operator: operator.Lexeme,
			Right:    right,
		}
	}

	return left
}

func (p *Parser) parseUnaryExpr() ast.Expression {
	if p.match(lexer.TOKEN_NOT, lexer.TOKEN_BANG, lexer.TOKEN_MINUS) {
		operator := p.previousToken()
		right := p.parseUnaryExpr()
		return &ast.UnaryExpr{
			Token:    operator,
			Operator: operator.Lexeme,
			Right:    right,
		}
	}

	// Handle "reference of expr" for address-of
	if p.match(lexer.TOKEN_REFERENCE) {
		refToken := p.previousToken()
		if p.match(lexer.TOKEN_OF) {
			operand := p.parseUnaryExpr()
			return &ast.AddressOfExpr{
				Token:   refToken,
				Operand: operand,
			}
		}
		// If not followed by 'of', we have an error - revert
		p.pos-- // Back up to before 'reference'
	}

	// Handle "dereference expr"
	if p.match(lexer.TOKEN_DEREFERENCE) {
		derefToken := p.previousToken()
		operand := p.parseUnaryExpr()
		return &ast.DerefExpr{
			Token:   derefToken,
			Operand: operand,
		}
	}

	return p.parsePostfixExpr()
}

func (p *Parser) parsePostfixExpr() ast.Expression {
	expr := p.parsePrimaryExpr()

	for {
		switch {
		case p.match(lexer.TOKEN_LPAREN):
			// Function call
			args, namedArgs, variadic := p.parseCallArguments()
			p.consume(lexer.TOKEN_RPAREN, "expected ')' after arguments")
			expr = &ast.CallExpr{
				Token:          p.previousToken(),
				Function:       expr,
				Arguments:      args,
				NamedArguments: namedArgs,
				Variadic:       variadic,
			}

		case p.match(lexer.TOKEN_DOT):
			dotToken := p.previousToken()

			// Check for type assertion: .(Type)
			if p.check(lexer.TOKEN_LPAREN) {
				p.advance() // consume '('
				targetType := p.parseTypeAnnotation()
				p.consume(lexer.TOKEN_RPAREN, "expected ')' after type assertion")
				expr = &ast.TypeAssertionExpr{
					Token:      dotToken,
					Expression: expr,
					TargetType: targetType,
				}
				continue
			}

			// Method call or field access
			method := p.parseIdentifier()

			if p.check(lexer.TOKEN_LPAREN) {
				// Method call
				p.advance() // consume '('
				args, namedArgs, variadic := p.parseCallArguments()
				p.consume(lexer.TOKEN_RPAREN, "expected ')' after arguments")
				expr = &ast.MethodCallExpr{
					Token:          dotToken,
					Object:         expr,
					Method:         method,
					Arguments:      args,
					NamedArguments: namedArgs,
					Variadic:       variadic,
					IsCall:         true,
				}
			} else if p.check(lexer.TOKEN_LBRACE) {
				// Qualified struct literal: pkg.Type{}
				// expr should be the package identifier
				if ident, ok := expr.(*ast.Identifier); ok {
					qualifiedName := ident.Value + "." + method.Value
					p.advance() // consume '{'

					// Parse struct literal fields
					fields := []*ast.FieldValue{}
					if !p.check(lexer.TOKEN_RBRACE) {
						for {
							fieldName := p.parseIdentifier()
							p.consume(lexer.TOKEN_COLON, "expected ':' after field name")
							fieldValue := p.parseExpression()
							fields = append(fields, &ast.FieldValue{
								Name:  fieldName,
								Value: fieldValue,
							})
							if p.match(lexer.TOKEN_COMMA) {
								continue
							}
							break
						}
					}
					p.consume(lexer.TOKEN_RBRACE, "expected '}' after struct literal")

					expr = &ast.StructLiteralExpr{
						Token: ident.Token,
						Type: &ast.NamedType{
							Token: ident.Token,
							Name:  qualifiedName,
						},
						Fields: fields,
					}
				} else {
					// Not a simple package.Type, treat as field access
					expr = &ast.MethodCallExpr{
						Token:     dotToken,
						Object:    expr,
						Method:    method,
						Arguments: []ast.Expression{},
					}
				}
			} else {
				// Field access - treat as method call with no args for now
				expr = &ast.MethodCallExpr{
					Token:     dotToken,
					Object:    expr,
					Method:    method,
					Arguments: []ast.Expression{},
				}
			}

		case p.match(lexer.TOKEN_LBRACKET):
			// Index or slice
			if p.check(lexer.TOKEN_COLON) {
				// Slice with no start: [:end]
				p.advance() // consume ':'
				end := p.parseExpression()
				p.consume(lexer.TOKEN_RBRACKET, "expected ']' after slice")
				expr = &ast.SliceExpr{
					Token: p.previousToken(),
					Left:  expr,
					Start: nil,
					End:   end,
				}
			} else {
				first := p.parseExpression()
				if p.match(lexer.TOKEN_COLON) {
					// Slice: [start:end] or [start:]
					var end ast.Expression
					if !p.check(lexer.TOKEN_RBRACKET) {
						end = p.parseExpression()
					}
					p.consume(lexer.TOKEN_RBRACKET, "expected ']' after slice")
					expr = &ast.SliceExpr{
						Token: p.previousToken(),
						Left:  expr,
						Start: first,
						End:   end,
					}
				} else {
					// Index: [index]
					p.consume(lexer.TOKEN_RBRACKET, "expected ']' after index")
					expr = &ast.IndexExpr{
						Token: p.previousToken(),
						Left:  expr,
						Index: first,
					}
				}
			}

		case p.match(lexer.TOKEN_AS):
			// Type cast
			asToken := p.previousToken()
			targetType := p.parseTypeAnnotation()
			expr = &ast.TypeCastExpr{
				Token:      asToken,
				Expression: expr,
				TargetType: targetType,
			}

		default:
			return expr
		}
	}
}

func (p *Parser) parsePrimaryExpr() ast.Expression {
	switch p.peekToken().Type {
	case lexer.TOKEN_INTEGER:
		return p.parseIntegerLiteral()
	case lexer.TOKEN_FLOAT:
		return p.parseFloatLiteral()
	case lexer.TOKEN_STRING:
		return p.parseStringLiteral()
	case lexer.TOKEN_RUNE:
		return p.parseRuneLiteral()
	case lexer.TOKEN_TRUE, lexer.TOKEN_FALSE:
		return p.parseBooleanLiteral()
	case lexer.TOKEN_IDENTIFIER:
		return p.parseIdentifierOrStructLiteral()
	case lexer.TOKEN_EMPTY:
		return p.parseEmptyExpr()
	case lexer.TOKEN_DISCARD:
		token := p.advance()
		return &ast.DiscardExpr{Token: token}
	case lexer.TOKEN_ERROR:
		return p.parseErrorExpr()
	case lexer.TOKEN_MAKE:
		return p.parseMakeExpr()
	case lexer.TOKEN_CLOSE:
		return p.parseCloseExpr()
	case lexer.TOKEN_PANIC:
		return p.parsePanicExpr()
	case lexer.TOKEN_RECOVER:
		token := p.advance()
		return &ast.RecoverExpr{Token: token}
	case lexer.TOKEN_RECEIVE:
		return p.parseReceiveExpr()
	case lexer.TOKEN_LIST:
		return p.parseTypedListLiteral()
	case lexer.TOKEN_LBRACKET:
		return p.parseListLiteral()
	case lexer.TOKEN_LPAREN:
		return p.parseGroupedExpression()
	case lexer.TOKEN_FUNC:
		return p.parseFunctionLiteral()
	case lexer.TOKEN_DOT:
		return p.parseShorthandMethodCall()
	case lexer.TOKEN_RETURN:
		return p.parseReturnExpr()
	default:
		p.error(p.peekToken(), fmt.Sprintf("unexpected token in expression: %s", p.peekToken().Type))
		p.advance()
		return nil
	}
}

func (p *Parser) parseIdentifier() *ast.Identifier {
	token := p.advance()
	if token.Type != lexer.TOKEN_IDENTIFIER {
		p.error(token, "expected identifier")
		return nil
	}
	return &ast.Identifier{
		Token: token,
		Value: token.Lexeme,
	}
}

func (p *Parser) parseIntegerLiteral() *ast.IntegerLiteral {
	token := p.advance()
	// Use base 0 to auto-detect: 0x=hex, 0o/0=octal, 0b=binary, otherwise decimal
	value, err := strconv.ParseInt(token.Lexeme, 0, 64)
	if err != nil {
		p.error(token, fmt.Sprintf("could not parse integer: %s", err))
		return nil
	}
	return &ast.IntegerLiteral{
		Token: token,
		Value: value,
	}
}

func (p *Parser) parseFloatLiteral() *ast.FloatLiteral {
	token := p.advance()
	value, err := strconv.ParseFloat(token.Lexeme, 64)
	if err != nil {
		p.error(token, fmt.Sprintf("could not parse float: %s", err))
		return nil
	}
	return &ast.FloatLiteral{
		Token: token,
		Value: value,
	}
}

func (p *Parser) parseStringLiteral() *ast.StringLiteral {
	token := p.advance()

	// Check for string interpolation
	if strings.Contains(token.Lexeme, "{") {
		// Has interpolation - we'll parse this in semantic analysis
		return &ast.StringLiteral{
			Token:        token,
			Value:        token.Lexeme,
			Interpolated: true,
		}
	}

	return &ast.StringLiteral{
		Token:        token,
		Value:        token.Lexeme,
		Interpolated: false,
	}
}

func (p *Parser) parseRuneLiteral() *ast.RuneLiteral {
	token := p.advance()
	// The lexeme contains the character as a string
	var value rune
	if len(token.Lexeme) > 0 {
		value = []rune(token.Lexeme)[0]
	}
	return &ast.RuneLiteral{
		Token: token,
		Value: value,
	}
}

func (p *Parser) parseBooleanLiteral() *ast.BooleanLiteral {
	token := p.advance()
	return &ast.BooleanLiteral{
		Token: token,
		Value: token.Type == lexer.TOKEN_TRUE,
	}
}

func (p *Parser) parseIdentifierOrStructLiteral() ast.Expression {
	// Could be an identifier or a struct literal (TypeName{field: value})
	ident := p.parseIdentifier()

	// Check for struct literal
	if p.check(lexer.TOKEN_LBRACE) {
		p.advance() // consume '{'

		// Parse type from identifier
		var typ ast.TypeAnnotation
		switch ident.Value {
		case "int", "int8", "int16", "int32", "int64",
			"uint", "uint8", "uint16", "uint32", "uint64",
			"float32", "float64", "string", "bool", "byte", "rune":
			typ = &ast.PrimitiveType{
				Token: ident.Token,
				Name:  ident.Value,
			}
		default:
			typ = &ast.NamedType{
				Token: ident.Token,
				Name:  ident.Value,
			}
		}

		// Check for map literal: {key: value}
		// vs struct literal: TypeName{field: value}
		// We'll treat this as struct literal for now
		fields := []*ast.FieldValue{}

		if !p.check(lexer.TOKEN_RBRACE) {
			for {
				fieldName := p.parseIdentifier()
				p.consume(lexer.TOKEN_COLON, "expected ':' after field name")
				fieldValue := p.parseExpression()
				fields = append(fields, &ast.FieldValue{
					Name:  fieldName,
					Value: fieldValue,
				})

				if p.match(lexer.TOKEN_COMMA) {
					continue
				}
				break
			}
		}

		p.consume(lexer.TOKEN_RBRACE, "expected '}' after struct literal")

		return &ast.StructLiteralExpr{
			Token:  ident.Token,
			Type:   typ,
			Fields: fields,
		}
	}

	return ident
}

func (p *Parser) parseEmptyExpr() *ast.EmptyExpr {
	token := p.advance() // consume 'empty'

	expr := &ast.EmptyExpr{Token: token}

	// Check for typed empty: empty Type
	// Be careful not to consume logical operators or other delimiters as type annotations
	next := p.peekToken().Type
	if !p.check(lexer.TOKEN_NEWLINE) && !p.check(lexer.TOKEN_COMMA) && !p.check(lexer.TOKEN_RPAREN) &&
		!p.check(lexer.TOKEN_AND) && !p.check(lexer.TOKEN_OR) && !p.check(lexer.TOKEN_NOT_EQUALS) &&
		!p.check(lexer.TOKEN_DOUBLE_EQUALS) && !p.check(lexer.TOKEN_BANG) && !p.check(lexer.TOKEN_PIPE) &&
		!p.isAtEnd() {
		// Only parse if it looks like a type name or keywords like 'map', 'list', 'func', 'channel'
		if next == lexer.TOKEN_IDENTIFIER || next == lexer.TOKEN_MAP || next == lexer.TOKEN_LIST ||
			next == lexer.TOKEN_FUNC || next == lexer.TOKEN_CHANNEL || next == lexer.TOKEN_REFERENCE {
			expr.Type = p.parseTypeAnnotation()
		}
	}

	return expr
}

func (p *Parser) parseErrorExpr() *ast.ErrorExpr {
	token := p.advance() // consume 'error'
	message := p.parseExpression()
	return &ast.ErrorExpr{
		Token:   token,
		Message: message,
	}
}

func (p *Parser) parseMakeExpr() *ast.MakeExpr {
	token := p.advance() // consume 'make'
	p.consume(lexer.TOKEN_LPAREN, "expected '(' after 'make'")

	typ := p.parseTypeAnnotation()
	args := []ast.Expression{}

	if p.match(lexer.TOKEN_COMMA) {
		for {
			args = append(args, p.parseExpression())
			if !p.match(lexer.TOKEN_COMMA) {
				break
			}
		}
	}

	p.consume(lexer.TOKEN_RPAREN, "expected ')' after make arguments")

	return &ast.MakeExpr{
		Token: token,
		Type:  typ,
		Args:  args,
	}
}

func (p *Parser) parseCloseExpr() *ast.CloseExpr {
	token := p.advance() // consume 'close'
	channel := p.parseExpression()
	return &ast.CloseExpr{
		Token:   token,
		Channel: channel,
	}
}

func (p *Parser) parsePanicExpr() *ast.PanicExpr {
	token := p.advance() // consume 'panic'
	message := p.parseExpression()
	return &ast.PanicExpr{
		Token:   token,
		Message: message,
	}
}

func (p *Parser) parseReceiveExpr() *ast.ReceiveExpr {
	token := p.advance() // consume 'receive'
	p.consume(lexer.TOKEN_FROM, "expected 'from' after 'receive'")
	channel := p.parseExpression()
	return &ast.ReceiveExpr{
		Token:   token,
		Channel: channel,
	}
}

func (p *Parser) parseListLiteral() *ast.ListLiteralExpr {
	token := p.advance() // consume '['

	elements := []ast.Expression{}

	if !p.check(lexer.TOKEN_RBRACKET) {
		for {
			elements = append(elements, p.parseExpression())
			if !p.match(lexer.TOKEN_COMMA) {
				break
			}
		}
	}

	p.consume(lexer.TOKEN_RBRACKET, "expected ']' after list elements")

	return &ast.ListLiteralExpr{
		Token:    token,
		Elements: elements,
	}
}

func (p *Parser) parseTypedListLiteral() *ast.ListLiteralExpr {
	token := p.advance() // consume 'list'
	p.consume(lexer.TOKEN_OF, "expected 'of' after 'list'")

	elementType := p.parseTypeAnnotation()

	p.consume(lexer.TOKEN_LBRACE, "expected '{' after list type")

	elements := []ast.Expression{}
	if !p.check(lexer.TOKEN_RBRACE) {
		for {
			elements = append(elements, p.parseExpression())
			if !p.match(lexer.TOKEN_COMMA) {
				break
			}
		}
	}

	p.consume(lexer.TOKEN_RBRACE, "expected '}' after list elements")

	return &ast.ListLiteralExpr{
		Token:    token,
		Type:     elementType,
		Elements: elements,
	}
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.advance() // consume '('
	expr := p.parseExpression()
	p.consume(lexer.TOKEN_RPAREN, "expected ')' after expression")
	return expr
}

func (p *Parser) parseFunctionLiteral() *ast.FunctionLiteral {
	token := p.advance() // consume 'func'
	p.consume(lexer.TOKEN_LPAREN, "expected '(' after 'func'")

	// Parse parameters (same as function declaration)
	params := p.parseParameters()
	p.consume(lexer.TOKEN_RPAREN, "expected ')' after parameters")

	// Parse return types (optional)
	returns := []ast.TypeAnnotation{}
	if !p.check(lexer.TOKEN_NEWLINE) && !p.check(lexer.TOKEN_INDENT) {
		returns = p.parseReturnTypes()
	}

	// Parse body
	p.skipNewlines()
	body := p.parseBlock()

	return &ast.FunctionLiteral{
		Token:      token,
		Parameters: params,
		Returns:    returns,
		Body:       body,
	}
}

// parseStructTag parses a struct tag like json:"name" or empty string if none present
// Format: identifier:stringLiteral
func (p *Parser) parseStructTag() string {
	// Check if next token is an identifier (tag name like "json", "xml", etc.)
	if !p.check(lexer.TOKEN_IDENTIFIER) {
		return ""
	}

	// Look ahead to see if there's a colon
	// Save current position
	savedPos := p.pos
	tagKeyToken := p.advance() // consume identifier

	if !p.check(lexer.TOKEN_COLON) {
		// Not a tag, restore position and return empty
		p.pos = savedPos
		return ""
	}

	// We have a tag - continue parsing
	tagKey := tagKeyToken.Lexeme
	p.consume(lexer.TOKEN_COLON, "expected ':' in struct tag")

	if !p.check(lexer.TOKEN_STRING) {
		p.error(p.peekToken(), "expected string value in struct tag")
		return ""
	}

	tagValueToken := p.advance() // consume string
	tagValue := tagValueToken.Lexeme

	// Return formatted tag: json:"name"
	return tagKey + ":" + `"` + tagValue + `"`
}
func (p *Parser) parseReturnExpr() ast.Expression {
	token := p.advance() // consume 'return'

	expr := &ast.ReturnExpr{
		Token:  token,
		Values: []ast.Expression{},
	}

	// Check if there are return values
	// Semicolon, newline, or dedent end the expression in onerr context
	if !p.check(lexer.TOKEN_NEWLINE) && !p.check(lexer.TOKEN_DEDENT) && !p.check(lexer.TOKEN_SEMICOLON) && !p.isAtEnd() {
		for {
			expr.Values = append(expr.Values, p.parseExpression())
			if !p.match(lexer.TOKEN_COMMA) {
				break
			}
		}
	}

	return expr
}

func (p *Parser) parseShorthandMethodCall() ast.Expression {
	token := p.advance() // consume '.'
	methodName := p.parseIdentifier()

	expr := &ast.MethodCallExpr{
		Token:  token,
		Object: nil, // shorthand
		Method: methodName,
		IsCall: false,
	}

	if p.match(lexer.TOKEN_LPAREN) {
		expr.IsCall = true
		if !p.check(lexer.TOKEN_RPAREN) {
			expr.Arguments = p.parseExpressionList()
		} else {
			expr.Arguments = []ast.Expression{}
		}
		p.consume(lexer.TOKEN_RPAREN, "expected ')' after method arguments")
	}

	return expr
}
