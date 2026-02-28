package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/duber000/kukicha/internal/ast"
	"github.com/duber000/kukicha/internal/lexer"
)

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

	for {
		var operator lexer.Token
		if p.match(lexer.TOKEN_DOUBLE_EQUALS, lexer.TOKEN_NOT_EQUALS, lexer.TOKEN_LT, lexer.TOKEN_GT, lexer.TOKEN_LTE, lexer.TOKEN_GTE, lexer.TOKEN_EQUALS) {
			operator = p.previousToken()
		} else if p.check(lexer.TOKEN_NOT) && p.peekNextToken().Type == lexer.TOKEN_EQUALS {
			operator = p.advance() // consume NOT
			operator.Lexeme = "not equals"
			p.advance() // consume EQUALS
		} else if p.match(lexer.TOKEN_IN) {
			operator = p.previousToken()
		} else if p.check(lexer.TOKEN_NOT) && p.peekNextToken().Type == lexer.TOKEN_IN {
			operator = p.advance() // consume NOT
			operator.Lexeme = "not in"
			p.advance() // consume IN
		} else {
			break
		}

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
								if p.check(lexer.TOKEN_RBRACE) {
									break
								}
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
		// Check for single-param untyped arrow lambda: x => expr
		if p.peekNextToken().Type == lexer.TOKEN_FAT_ARROW {
			return p.parseArrowLambda()
		}
		return p.parseIdentifierOrStructLiteral()
	case lexer.TOKEN_EMPTY:
		// empty is almost always a literal unless it's an assignment or member access
		next := p.peekNextToken().Type
		if next == lexer.TOKEN_WALRUS || next == lexer.TOKEN_ASSIGN ||
			next == lexer.TOKEN_DOT || next == lexer.TOKEN_LBRACKET ||
			next == lexer.TOKEN_COLON {
			token := p.advance()
			return &ast.Identifier{Token: token, Value: token.Lexeme}
		}
		return p.parseEmptyExpr()
	case lexer.TOKEN_DISCARD:
		token := p.advance()
		return &ast.DiscardExpr{Token: token}
	case lexer.TOKEN_ERROR:
		if p.isIdentifierFollower() || p.check(lexer.TOKEN_RPAREN) || p.check(lexer.TOKEN_COMMA) || p.check(lexer.TOKEN_COLON) {
			token := p.advance()
			return &ast.Identifier{Token: token, Value: token.Lexeme}
		}
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
		if p.peekNextToken().Type == lexer.TOKEN_OF {
			return p.parseTypedListLiteral()
		}
		token := p.advance()
		return &ast.Identifier{Token: token, Value: token.Lexeme}
	case lexer.TOKEN_MAP:
		if p.peekNextToken().Type == lexer.TOKEN_OF {
			return p.parseMapLiteral()
		}
		token := p.advance()
		return &ast.Identifier{Token: token, Value: token.Lexeme}
	case lexer.TOKEN_LBRACKET:
		return p.parseListLiteral()
	case lexer.TOKEN_LPAREN:
		// Check if this is an arrow lambda: () => ..., (x Type) => ..., (x, y) => ...
		if p.isArrowLambda() {
			return p.parseArrowLambda()
		}
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
	if token.Type != lexer.TOKEN_IDENTIFIER && token.Type != lexer.TOKEN_EMPTY && token.Type != lexer.TOKEN_ERROR {
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
	var fields []*ast.FieldValue
	isIndented := false
	isBraced := false

	if p.check(lexer.TOKEN_LBRACE) {
		isBraced = true
		p.advance() // consume '{'
	} else if p.peekToken().Type == lexer.TOKEN_NEWLINE &&
		p.peekNextToken().Type == lexer.TOKEN_INDENT &&
		p.peekAt(2).Type == lexer.TOKEN_IDENTIFIER &&
		p.peekAt(3).Type == lexer.TOKEN_COLON {
		isIndented = true
		p.advance() // consume NEWLINE
		p.advance() // consume INDENT
	}

	if isBraced || isIndented {
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

		fields = []*ast.FieldValue{}

		if isBraced {
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
						if p.check(lexer.TOKEN_RBRACE) {
							break
						}
						continue
					}
					break
				}
			}
			p.consume(lexer.TOKEN_RBRACE, "expected '}' after struct literal")
		} else {
			// Indented
			for !p.check(lexer.TOKEN_DEDENT) && !p.isAtEnd() {
				p.skipNewlines()
				if p.check(lexer.TOKEN_DEDENT) {
					break
				}

				fieldName := p.parseIdentifier()
				p.consume(lexer.TOKEN_COLON, "expected ':' after field name")
				fieldValue := p.parseExpression()
				fields = append(fields, &ast.FieldValue{Name: fieldName, Value: fieldValue})

				if p.check(lexer.TOKEN_COMMA) {
					p.advance()
				}
				p.skipNewlines()
			}
			p.consume(lexer.TOKEN_DEDENT, "expected dedent after struct fields")
		}

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
			if p.check(lexer.TOKEN_RBRACKET) {
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

func (p *Parser) parseTypedListLiteral() ast.Expression {
	token := p.advance() // consume 'list'
	p.consume(lexer.TOKEN_OF, "expected 'of' after 'list'")

	elementType := p.parseTypeAnnotation()

	// Allow "list of T" in expression position as typed-empty shorthand.
	// This is useful for APIs like fetch.Json(list of Repo).
	if !p.check(lexer.TOKEN_LBRACE) {
		return &ast.EmptyExpr{
			Token: token,
			Type: &ast.ListType{
				Token:       token,
				ElementType: elementType,
			},
		}
	}

	p.consume(lexer.TOKEN_LBRACE, "expected '{' after list type")

	elements := []ast.Expression{}
	if !p.check(lexer.TOKEN_RBRACE) {
		for {
			elements = append(elements, p.parseExpression())
			if !p.match(lexer.TOKEN_COMMA) {
				break
			}
			if p.check(lexer.TOKEN_RBRACE) {
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

// isArrowLambda performs lookahead to determine if the current position starts
// an arrow lambda expression. Called when peekToken is TOKEN_LPAREN.
// It scans forward to find the matching ')' and checks if '=>' follows.
func (p *Parser) isArrowLambda() bool {
	// We're at '(' — scan forward to find matching ')'
	depth := 0
	i := p.pos
	for i < len(p.tokens) {
		tok := p.tokens[i]
		switch tok.Type {
		case lexer.TOKEN_LPAREN:
			depth++
		case lexer.TOKEN_RPAREN:
			depth--
			if depth == 0 {
				// Found matching ')'. Check if '=>' follows.
				i++
				// Skip any comments
				for i < len(p.tokens) && p.tokens[i].Type == lexer.TOKEN_COMMENT {
					i++
				}
				return i < len(p.tokens) && p.tokens[i].Type == lexer.TOKEN_FAT_ARROW
			}
		case lexer.TOKEN_NEWLINE, lexer.TOKEN_EOF, lexer.TOKEN_INDENT, lexer.TOKEN_DEDENT:
			// Newlines inside parens shouldn't occur (lexer suppresses them)
			// but if we hit EOF or indent tokens, it's not a lambda
			return false
		}
		i++
	}
	return false
}

// parseArrowLambda parses an arrow lambda expression.
// Forms:
//
//	x => expr                          single untyped param
//	(x Type) => expr                   single typed param
//	(x Type, y Type) => expr           multiple typed params
//	(x, y) => expr                     multiple untyped params
//	() => expr                         zero params
//	<any of the above> => NEWLINE INDENT ... DEDENT   block form
func (p *Parser) parseArrowLambda() *ast.ArrowLambda {
	var params []*ast.Parameter

	if p.check(lexer.TOKEN_IDENTIFIER) && p.peekNextToken().Type == lexer.TOKEN_FAT_ARROW {
		// Single untyped param: x => ...
		paramToken := p.advance()
		params = append(params, &ast.Parameter{
			Name: &ast.Identifier{Token: paramToken, Value: paramToken.Lexeme},
		})
	} else if p.check(lexer.TOKEN_LPAREN) {
		p.advance() // consume '('
		if !p.check(lexer.TOKEN_RPAREN) {
			params = p.parseArrowLambdaParams()
		}
		p.consume(lexer.TOKEN_RPAREN, "expected ')' after arrow lambda parameters")
	}

	arrowToken, _ := p.consume(lexer.TOKEN_FAT_ARROW, "expected '=>' in arrow lambda")

	lambda := &ast.ArrowLambda{
		Token:      arrowToken,
		Parameters: params,
	}

	// Check if block form or expression form
	if p.check(lexer.TOKEN_NEWLINE) || p.check(lexer.TOKEN_INDENT) {
		p.skipNewlines()
		if p.check(lexer.TOKEN_INDENT) {
			lambda.Block = p.parseBlock()
		} else {
			// Newline but no indent — parse as expression
			lambda.Body = p.parseExpression()
		}
	} else {
		lambda.Body = p.parseExpression()
	}

	return lambda
}

// parseArrowLambdaParams parses arrow lambda parameters.
// Supports both typed (x int, y string) and untyped (x, y) params.
func (p *Parser) parseArrowLambdaParams() []*ast.Parameter {
	var params []*ast.Parameter

	for {
		paramName := p.parseIdentifier()

		// Determine if this is typed or untyped by checking what follows:
		// - comma or ')' means untyped
		// - anything else means it's a type annotation
		var paramType ast.TypeAnnotation
		if !p.check(lexer.TOKEN_COMMA) && !p.check(lexer.TOKEN_RPAREN) && !p.check(lexer.TOKEN_ASSIGN) {
			paramType = p.parseTypeAnnotation()
		}

		// Check for default value
		var defaultValue ast.Expression
		if p.match(lexer.TOKEN_ASSIGN) {
			defaultValue = p.parseExpression()
		}

		params = append(params, &ast.Parameter{
			Name:         paramName,
			Type:         paramType,
			DefaultValue: defaultValue,
		})

		if !p.match(lexer.TOKEN_COMMA) {
			break
		}
	}

	return params
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

func (p *Parser) parseMapLiteral() ast.Expression {
	token := p.advance() // consume 'map'
	p.consume(lexer.TOKEN_OF, "expected 'of' after 'map'")
	keyType := p.parseTypeAnnotation()
	p.consume(lexer.TOKEN_TO, "expected 'to' after key type")
	valType := p.parseTypeAnnotation()

	// Allow "map of K to V" in expression position as typed-empty shorthand.
	if !p.check(lexer.TOKEN_LBRACE) {
		return &ast.EmptyExpr{
			Token: token,
			Type: &ast.MapType{
				Token:     token,
				KeyType:   keyType,
				ValueType: valType,
			},
		}
	}

	p.consume(lexer.TOKEN_LBRACE, "expected '{' after map type")

	pairs := []*ast.KeyValuePair{}
	if !p.check(lexer.TOKEN_RBRACE) {
		for {
			// Newlines are suppressed inside braces by lexer, but we can verify
			key := p.parseExpression()
			p.consume(lexer.TOKEN_COLON, "expected ':' after map key")
			val := p.parseExpression()

			pairs = append(pairs, &ast.KeyValuePair{Key: key, Value: val})

			if p.match(lexer.TOKEN_COMMA) {
				if p.check(lexer.TOKEN_RBRACE) {
					break
				}
				continue
			}
			break
		}
	}

	p.consume(lexer.TOKEN_RBRACE, "expected '}' after map literal")

	return &ast.MapLiteralExpr{
		Token:   token,
		KeyType: keyType,
		ValType: valType,
		Pairs:   pairs,
	}
}
