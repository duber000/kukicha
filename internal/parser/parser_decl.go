package parser

import (
	"fmt"

	"github.com/duber000/kukicha/internal/ast"
	"github.com/duber000/kukicha/internal/lexer"
)

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

func (p *Parser) parseSkillDecl() *ast.SkillDecl {
	token := p.advance() // consume 'skill'
	p.skipNewlines()

	name := p.parseIdentifier()

	decl := &ast.SkillDecl{
		Token: token,
		Name:  name,
	}

	p.skipNewlines()

	// Check for indented block with description/version fields
	if p.match(lexer.TOKEN_INDENT) {
		for !p.check(lexer.TOKEN_DEDENT) && !p.isAtEnd() {
			p.skipNewlines()
			if p.check(lexer.TOKEN_DEDENT) {
				break
			}

			// Parse field name (contextual identifier: "description" or "version")
			fieldToken := p.advance()
			if fieldToken.Type != lexer.TOKEN_IDENTIFIER {
				p.error(fieldToken, fmt.Sprintf("expected 'description' or 'version' in skill block, got %s", fieldToken.Type))
				p.skipNewlines()
				continue
			}

			p.consume(lexer.TOKEN_COLON, fmt.Sprintf("expected ':' after '%s'", fieldToken.Lexeme))

			// Parse string literal value
			valueToken := p.advance()
			if valueToken.Type != lexer.TOKEN_STRING {
				p.error(valueToken, fmt.Sprintf("expected string value for '%s'", fieldToken.Lexeme))
				p.skipNewlines()
				continue
			}

			switch fieldToken.Lexeme {
			case "description":
				decl.Description = valueToken.Lexeme
			case "version":
				decl.Version = valueToken.Lexeme
			default:
				p.error(fieldToken, fmt.Sprintf("unknown skill field '%s' (expected 'description' or 'version')", fieldToken.Lexeme))
			}

			p.skipNewlines()
		}

		p.consume(lexer.TOKEN_DEDENT, "expected dedent after skill block")
		p.skipNewlines()
	}

	return decl
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
	case lexer.TOKEN_VAR:
		return p.parseVarDeclaration()
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

	// Check for type alias: type Name func(...) ...
	if p.check(lexer.TOKEN_FUNC) {
		aliasType := p.parseTypeAnnotation()
		p.skipNewlines()
		return &ast.TypeDecl{
			Token:     token,
			Name:      name,
			AliasType: aliasType,
		}
	}

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
		alias := p.parseFieldAlias()

		// Parse optional struct tag (e.g., json:"name")
		tag := p.parseStructTag()
		if alias != "" && tag != "" {
			p.error(p.peekToken(), "cannot combine field alias and explicit struct tag on the same field")
		} else if alias != "" {
			tag = `json:"` + alias + `"`
		}

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
			p.error(paramName.Token, fmt.Sprintf("parameter '%s' must have a default value (parameters with defaults must be contiguous at the end)", paramName.Value))
		}

		// Variadic parameters cannot have default values
		if variadic && defaultValue != nil {
			p.error(paramName.Token, fmt.Sprintf("variadic parameter '%s' cannot have a default value", paramName.Value))
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

// parseFieldAlias parses optional field alias syntax: as "json_name"
// Returns empty string when no alias is present.
func (p *Parser) parseFieldAlias() string {
	if !p.match(lexer.TOKEN_AS) {
		return ""
	}

	if !p.check(lexer.TOKEN_STRING) {
		p.error(p.peekToken(), "expected string value after 'as' in field alias")
		return ""
	}

	return p.advance().Lexeme
}

func (p *Parser) parseVarDeclaration() ast.Declaration {
	token := p.advance() // consume 'var'
	p.skipNewlines()

	// Parse identifiers
	var names []*ast.Identifier
	firstIdent := p.parseIdentifier()
	if firstIdent == nil {
		return nil
	}
	names = append(names, firstIdent)

	for p.match(lexer.TOKEN_COMMA) {
		ident := p.parseIdentifier()
		if ident == nil {
			return nil
		}
		names = append(names, ident)
	}

	// Parse type (optional)
	var typeAnnot ast.TypeAnnotation
	// Check if next is assignment or implicit newline/EOF (if allowed?)
	// If not assignment, try to parse type.
	if !p.check(lexer.TOKEN_ASSIGN) {
		typeAnnot = p.parseTypeAnnotation()
	}

	// Parse values
	var values []ast.Expression
	if p.match(lexer.TOKEN_ASSIGN) {
		values = p.parseExpressionList()
	}

	p.skipNewlines()

	return &ast.VarDeclStmt{
		Token:  token,
		Names:  names,
		Type:   typeAnnot,
		Values: values,
	}
}

