package parser

import (
	"github.com/duber000/kukicha/internal/ast"
	"github.com/duber000/kukicha/internal/lexer"
)

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
			nameToken := p.advance()     // consume identifier
			p.advance()                  // consume colon
			value := p.parseExpression() // parse value
			namedArgs = append(namedArgs, &ast.NamedArgument{
				Token: nameToken,
				Name:  &ast.Identifier{Token: nameToken, Value: nameToken.Lexeme},
				Value: value,
			})
			hasNamedArg = true
		} else {
			// Positional argument
			if hasNamedArg {
				p.error(p.peekToken(), "positional argument cannot follow named argument")
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
	case lexer.TOKEN_SWITCH:
		return p.parseSwitchOrTypeSwitchStmt()
	case lexer.TOKEN_SELECT:
		return p.parseSelectStmt()
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

func (p *Parser) parseSwitchOrTypeSwitchStmt() ast.Statement {
	token := p.advance() // consume 'switch'

	// Parse optional expression
	var expr ast.Expression
	if !p.check(lexer.TOKEN_NEWLINE) && !p.check(lexer.TOKEN_INDENT) && !p.isAtEnd() {
		expr = p.parseExpression()
	}

	// Check if this is a type switch: switch expr as binding
	// parseExpression will have parsed "expr as binding" as a TypeCastExpr
	// where TargetType is a simple NamedType (the binding name)
	if cast, ok := expr.(*ast.TypeCastExpr); ok {
		if named, ok := cast.TargetType.(*ast.NamedType); ok {
			return p.parseTypeSwitchBody(token, cast.Expression, &ast.Identifier{
				Token: named.Token,
				Value: named.Name,
			})
		}
	}

	// Regular switch statement
	return p.parseSwitchBody(token, expr)
}

func (p *Parser) parseSwitchBody(token lexer.Token, expr ast.Expression) *ast.SwitchStmt {
	stmt := &ast.SwitchStmt{
		Token:      token,
		Expression: expr,
		Cases:      []*ast.WhenCase{},
	}

	p.skipNewlines()
	if !p.match(lexer.TOKEN_INDENT) {
		p.error(p.peekToken(), "expected indented block after switch")
		return stmt
	}

	for !p.check(lexer.TOKEN_DEDENT) && !p.isAtEnd() {
		p.skipNewlines()
		if p.check(lexer.TOKEN_DEDENT) {
			break
		}

		if p.match(lexer.TOKEN_CASE) {
			caseToken := p.previousToken()
			if stmt.Otherwise != nil {
				p.error(caseToken, "'when' branch after 'otherwise' will never execute")
			}
			values := []ast.Expression{p.parseExpression()}
			for p.match(lexer.TOKEN_COMMA) {
				values = append(values, p.parseExpression())
			}

			p.skipNewlines()
			body := p.parseBlock()
			stmt.Cases = append(stmt.Cases, &ast.WhenCase{
				Token:  caseToken,
				Values: values,
				Body:   body,
			})
			continue
		}

		if p.match(lexer.TOKEN_DEFAULT) {
			otherwiseToken := p.previousToken()
			if stmt.Otherwise != nil {
				p.error(otherwiseToken, "switch can only have one otherwise branch")
			}

			p.skipNewlines()
			stmt.Otherwise = &ast.OtherwiseCase{
				Token: otherwiseToken,
				Body:  p.parseBlock(),
			}
			continue
		}

		p.error(p.peekToken(), "expected 'when' or 'otherwise' in switch block")
		p.advance()
	}

	p.consume(lexer.TOKEN_DEDENT, "expected dedent after switch block")
	p.skipNewlines()
	return stmt
}

func (p *Parser) parseTypeSwitchBody(token lexer.Token, expr ast.Expression, binding *ast.Identifier) *ast.TypeSwitchStmt {
	stmt := &ast.TypeSwitchStmt{
		Token:      token,
		Expression: expr,
		Binding:    binding,
		Cases:      []*ast.TypeCase{},
	}

	p.skipNewlines()
	if !p.match(lexer.TOKEN_INDENT) {
		p.error(p.peekToken(), "expected indented block after type switch")
		return stmt
	}

	for !p.check(lexer.TOKEN_DEDENT) && !p.isAtEnd() {
		p.skipNewlines()
		if p.check(lexer.TOKEN_DEDENT) {
			break
		}

		if p.match(lexer.TOKEN_CASE) {
			caseToken := p.previousToken()
			if stmt.Otherwise != nil {
				p.error(caseToken, "'when' branch after 'otherwise' will never execute")
			}
			typeAnn := p.parseTypeAnnotation()

			p.skipNewlines()
			body := p.parseBlock()
			stmt.Cases = append(stmt.Cases, &ast.TypeCase{
				Token: caseToken,
				Type:  typeAnn,
				Body:  body,
			})
			continue
		}

		if p.match(lexer.TOKEN_DEFAULT) {
			otherwiseToken := p.previousToken()
			if stmt.Otherwise != nil {
				p.error(otherwiseToken, "type switch can only have one otherwise branch")
			}

			p.skipNewlines()
			stmt.Otherwise = &ast.OtherwiseCase{
				Token: otherwiseToken,
				Body:  p.parseBlock(),
			}
			continue
		}

		p.error(p.peekToken(), "expected 'when' or 'otherwise' in type switch block")
		p.advance()
	}

	p.consume(lexer.TOKEN_DEDENT, "expected dedent after type switch block")
	p.skipNewlines()
	return stmt
}

func (p *Parser) parseSelectStmt() *ast.SelectStmt {
	token := p.advance() // consume 'select'

	stmt := &ast.SelectStmt{
		Token: token,
		Cases: []*ast.SelectCase{},
	}

	p.skipNewlines()
	if !p.match(lexer.TOKEN_INDENT) {
		p.error(p.peekToken(), "expected indented block after select")
		return stmt
	}

	for !p.check(lexer.TOKEN_DEDENT) && !p.isAtEnd() {
		p.skipNewlines()
		if p.check(lexer.TOKEN_DEDENT) {
			break
		}

		if p.match(lexer.TOKEN_CASE) { // 'when'
			caseToken := p.previousToken()
			if stmt.Otherwise != nil {
				p.error(caseToken, "'when' branch after 'otherwise' will never execute")
			}
			sc := p.parseSelectCase(caseToken)
			stmt.Cases = append(stmt.Cases, sc)
			continue
		}

		if p.match(lexer.TOKEN_DEFAULT) { // 'otherwise'
			otherwiseToken := p.previousToken()
			if stmt.Otherwise != nil {
				p.error(otherwiseToken, "select can only have one otherwise branch")
			}
			p.skipNewlines()
			stmt.Otherwise = &ast.OtherwiseCase{
				Token: otherwiseToken,
				Body:  p.parseBlock(),
			}
			continue
		}

		p.error(p.peekToken(), "expected 'when' or 'otherwise' in select block")
		p.advance()
	}

	p.consume(lexer.TOKEN_DEDENT, "expected dedent after select block") //nolint:errcheck
	p.skipNewlines()
	return stmt
}

func (p *Parser) parseSelectCase(caseToken lexer.Token) *ast.SelectCase {
	sc := &ast.SelectCase{Token: caseToken}

	switch {
	case p.check(lexer.TOKEN_RECEIVE):
		// Bare receive: "when receive from ch"
		sc.Recv = p.parseReceiveExpr()

	case p.check(lexer.TOKEN_SEND):
		// Send case: "when send val to ch"
		sc.Send = p.parseSendStmt()
		// parseSendStmt already consumes the newline via skipNewlines
		// Body will be parsed below; skip the extra newline skip
		sc.Body = p.parseBlock()
		return sc

	case p.check(lexer.TOKEN_IDENTIFIER):
		// Binding case: "when v := receive from ch" or "when v, ok := receive from ch"
		first := p.advance() // consume first identifier
		sc.Bindings = []string{first.Lexeme}

		if p.match(lexer.TOKEN_COMMA) {
			second, _ := p.consume(lexer.TOKEN_IDENTIFIER, "expected identifier after ',' in select binding")
			sc.Bindings = append(sc.Bindings, second.Lexeme)
		}

		p.consume(lexer.TOKEN_WALRUS, "expected ':=' after binding in select case") //nolint:errcheck
		sc.Recv = p.parseReceiveExpr()

	default:
		p.error(p.peekToken(), "expected 'receive', 'send', or binding in select case")
	}

	p.skipNewlines()
	sc.Body = p.parseBlock()
	return sc
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

	// Check for block form: go NEWLINE INDENT ... DEDENT
	// This desugars to go func() { ... }() in codegen
	if p.check(lexer.TOKEN_NEWLINE) || p.check(lexer.TOKEN_INDENT) {
		p.skipNewlines()
		if p.check(lexer.TOKEN_INDENT) {
			block := p.parseBlock()
			p.skipNewlines()
			return &ast.GoStmt{
				Token: token,
				Block: block,
			}
		}
	}

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
		p.error(token, "go must be followed by a function call or indented block")
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
	// Look ahead for: ident [, ident]+ := or =
	// Supports 2 or more identifiers on the left-hand side.
	// Examples: a, b := ...   or   _, ipNet, err := ...

	// Check if we have an identifier (or context-sensitive keyword) at current position
	currentToken := p.peekToken()
	if currentToken.Type != lexer.TOKEN_IDENTIFIER && currentToken.Type != lexer.TOKEN_EMPTY && currentToken.Type != lexer.TOKEN_ERROR {
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

	// Must have at least one comma after the first identifier
	idx, tok := skipIgnored(p.pos + 1)
	if tok.Type != lexer.TOKEN_COMMA {
		return false
	}

	// Consume (comma, identifier) pairs until we reach an assignment operator
	for tok.Type == lexer.TOKEN_COMMA {
		idx, tok = skipIgnored(idx + 1)
		if tok.Type != lexer.TOKEN_IDENTIFIER && tok.Type != lexer.TOKEN_EMPTY && tok.Type != lexer.TOKEN_ERROR {
			return false // Comma must be followed by an identifier
		}
		idx, tok = skipIgnored(idx + 1)
	}

	// After all identifiers, must be an assignment operator
	return tok.Type == lexer.TOKEN_ASSIGN || tok.Type == lexer.TOKEN_WALRUS
}

func (p *Parser) parseMultiValueAssignmentStmt() ast.Statement {
	// Parse left-hand side (comma-separated identifiers)
	var names []*ast.Identifier
	var targets []ast.Expression

	// Parse first identifier (also accept empty/error as identifiers)
	if !p.match(lexer.TOKEN_IDENTIFIER, lexer.TOKEN_EMPTY, lexer.TOKEN_ERROR) {
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
		if !p.match(lexer.TOKEN_IDENTIFIER, lexer.TOKEN_EMPTY, lexer.TOKEN_ERROR) {
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

// parseOnErrClause parses the onerr clause after a statement.
// Called when TOKEN_ONERR has already been detected (but not consumed).
//
// Forms:
//
//	onerr <handler>                          - handler only
//	onerr <handler> explain "hint"           - handler with explain
//	onerr explain "hint"                     - standalone explain (implies fmt.Errorf return)
//	onerr INDENT ... DEDENT                  - block handler
//	onerr return                             - shorthand: propagate error with zero-value returns
//	onerr as <ident> INDENT ... DEDENT       - block handler with named error alias
func (p *Parser) parseOnErrClause() *ast.OnErrClause {
	token := p.advance() // consume 'onerr'

	// Check for "onerr as <ident>" — block handler with named error alias.
	// Must appear before skipNewlines so we catch "as" on the same line as "onerr".
	if p.check(lexer.TOKEN_AS) {
		p.advance() // consume 'as'
		aliasToken := p.advance()
		if aliasToken.Type != lexer.TOKEN_IDENTIFIER {
			p.error(aliasToken, "expected identifier after 'onerr as'")
			return &ast.OnErrClause{Token: token}
		}
		alias := aliasToken.Lexeme
		p.skipNewlines()
		if !p.check(lexer.TOKEN_INDENT) {
			p.error(p.peekToken(), "'onerr as <ident>' must be followed by an indented block")
			return &ast.OnErrClause{Token: token}
		}
		block := p.parseBlock()
		return &ast.OnErrClause{
			Token: token,
			Alias: alias,
			Handler: &ast.BlockExpr{
				Token: block.Token,
				Body:  block,
			},
		}
	}

	// Check for block handler: onerr \n INDENT ...
	p.skipNewlines()
	if p.check(lexer.TOKEN_INDENT) {
		block := p.parseBlock()
		return &ast.OnErrClause{
			Token: token,
			Handler: &ast.BlockExpr{
				Token: block.Token,
				Body:  block,
			},
		}
	}

	// Check for bare "onerr return" shorthand.
	// Disambiguate from "onerr return empty, error ..." by peeking at the token
	// immediately after "return": if it is a newline, dedent, or EOF the user
	// wrote the shorthand form; otherwise fall through to parseExpression so the
	// existing verbose form continues to work.
	if p.check(lexer.TOKEN_RETURN) {
		next := p.peekNextToken()
		if next.Type == lexer.TOKEN_NEWLINE ||
			next.Type == lexer.TOKEN_DEDENT ||
			next.Type == lexer.TOKEN_EOF {
			p.advance() // consume 'return'
			return &ast.OnErrClause{
				Token:           token,
				ShorthandReturn: true,
			}
		}
	}

	// Check for standalone "onerr explain" (no handler before explain)
	if p.check(lexer.TOKEN_EXPLAIN) {
		p.advance() // consume 'explain'
		explainToken := p.advance()
		if explainToken.Type != lexer.TOKEN_STRING {
			p.error(explainToken, "expected string literal after 'explain'")
			return &ast.OnErrClause{Token: token}
		}
		// Standalone explain: implies return with fmt.Errorf wrapping
		return &ast.OnErrClause{
			Token:   token,
			Handler: nil, // nil handler signals standalone explain
			Explain: explainToken.Lexeme,
		}
	}

	handler := p.parseExpression()

	// Check for trailing "explain" after handler
	clause := &ast.OnErrClause{Token: token, Handler: handler}
	if p.check(lexer.TOKEN_EXPLAIN) {
		p.advance() // consume 'explain'
		explainToken := p.advance()
		if explainToken.Type != lexer.TOKEN_STRING {
			p.error(explainToken, "expected string literal after 'explain'")
		} else {
			clause.Explain = explainToken.Lexeme
		}
	}

	return clause
}

