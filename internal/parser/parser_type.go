package parser

import (
	"fmt"

	"github.com/duber000/kukicha/internal/ast"
	"github.com/duber000/kukicha/internal/lexer"
)

// ============================================================================
// Type Annotation Parsing
// ============================================================================

// parseTypeAnnotation parses Kukicha type syntax into AST TypeAnnotation nodes.
//
// ARCHITECTURE NOTE: This is where Kukicha's beginner-friendly type syntax
// is parsed. The English-like syntax maps to Go types:
//
//	Kukicha                   Go
//	-------                   --
//	list of string            []string
//	map of string to int      map[string]int
//	reference User            *User
//	channel of int            chan int
//	func(int) bool            func(int) bool
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

		// Parse return types (single or parenthesized multiple)
		var returns []ast.TypeAnnotation
		if p.peekToken().Type != lexer.TOKEN_NEWLINE &&
			p.peekToken().Type != lexer.TOKEN_COMMA &&
			p.peekToken().Type != lexer.TOKEN_RPAREN &&
			p.peekToken().Type != lexer.TOKEN_EOF &&
			p.peekToken().Type != lexer.TOKEN_INDENT &&
			p.peekToken().Type != lexer.TOKEN_DEDENT {
			if p.check(lexer.TOKEN_LPAREN) {
				// Multiple return types: (T, error)
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
