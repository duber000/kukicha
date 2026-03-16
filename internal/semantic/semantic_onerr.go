package semantic

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/duber000/kukicha/internal/ast"
	"github.com/duber000/kukicha/internal/parser"
)

var (
	interpolationPattern = regexp.MustCompile(`\{([a-zA-Z_][^}]*)\}`)
	identifierPattern    = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
)

// analyzeOnErrClause analyzes the onerr clause on a statement
func (a *Analyzer) analyzeOnErrClause(clause *ast.OnErrClause) {
	if clause == nil {
		return
	}

	pos := ast.Position{Line: clause.Token.Line, Column: clause.Token.Column, File: clause.Token.File}

	// Validate bare "onerr return" shorthand: enclosing function must return an error.
	if clause.ShorthandReturn {
		if a.currentFunc == nil {
			a.error(pos, "'onerr return' used outside of a function")
		} else if !funcReturnsError(a.currentFunc) {
			a.error(pos, "'onerr return' requires the enclosing function to return an error; use an explicit handler instead")
		}
		return
	}

	// Lint: onerr discard outside test files silently swallows errors.
	if _, isDiscard := clause.Handler.(*ast.DiscardExpr); isDiscard {
		if !strings.HasSuffix(a.sourceFile, "_test.kuki") {
			a.warn(pos, "onerr discard silently swallows errors; prefer an explicit handler (use in test files only)")
		}
	}

	// Lint: onerr panic in library (non-main) packages terminates the program.
	if _, isPanic := clause.Handler.(*ast.PanicExpr); isPanic {
		if a.program.PetioleDecl != nil && a.program.PetioleDecl.Name != nil &&
			a.program.PetioleDecl.Name.Value != "main" {
			a.warn(pos, "onerr panic in library code terminates the entire program; prefer returning an error to the caller")
		}
	}

	// Warn if the onerr error variable name shadows a user-declared variable.
	// The implicit name is "error"; an explicit alias overrides it.
	onerrrName := "error"
	if clause.Alias != "" {
		onerrrName = clause.Alias
	}
	if sym := a.symbolTable.Resolve(onerrrName); sym != nil {
		a.warn(pos, fmt.Sprintf("onerr variable '%s' shadows declaration at %s:%d", onerrrName, sym.Defined.File, sym.Defined.Line))
	}

	prev := a.inOnerr
	prevAlias := a.currentOnerrrAlias
	a.inOnerr = true
	if clause.Alias != "" {
		a.currentOnerrrAlias = clause.Alias
	}
	a.analyzeExpression(clause.Handler)
	a.inOnerr = prev
	a.currentOnerrrAlias = prevAlias
}

// funcReturnsError reports whether the function's last return type is "error".
func funcReturnsError(decl *ast.FunctionDecl) bool {
	if len(decl.Returns) == 0 {
		return false
	}
	last := decl.Returns[len(decl.Returns)-1]
	named, ok := last.(*ast.NamedType)
	return ok && named.Name == "error"
}

func (a *Analyzer) analyzeStringInterpolation(lit *ast.StringLiteral) {
	// Fast path: use pre-parsed Parts from the parser
	if len(lit.Parts) > 0 {
		for _, part := range lit.Parts {
			if part.IsLiteral {
				continue
			}
			// Check for {err} in onerr context
			if a.inOnerr {
				if ident, ok := part.Expr.(*ast.Identifier); ok && ident.Value == "err" {
					hint := "use {error} not {err} inside onerr — the caught error is always named 'error'"
					if a.currentOnerrrAlias != "" {
						hint += fmt.Sprintf(", or {%s} via your 'onerr as %s' alias", a.currentOnerrrAlias, a.currentOnerrrAlias)
					} else {
						hint += ", or name it with 'onerr as e' and use {e}"
					}
					a.error(lit.Pos(), hint)
					continue
				}
			}
			// Skip known onerr error variables — they're injected by codegen, not user-defined
			if a.inOnerr {
				if ident, ok := part.Expr.(*ast.Identifier); ok {
					if ident.Value == "error" || ident.Value == a.currentOnerrrAlias {
						continue
					}
				}
			}
			// Patch position info for better error reporting
			patchExprPosition(part.Expr, lit.Token.File, lit.Token.Line, lit.Token.Column)
			a.analyzeExpression(part.Expr)
		}
		return
	}

	// Fallback: regex-based parsing when Parts is empty (parse failure)
	matches := interpolationPattern.FindAllStringSubmatchIndex(lit.Value, -1)

	for _, match := range matches {
		exprStr := strings.TrimSpace(lit.Value[match[2]:match[3]])
		if exprStr == "" {
			a.error(lit.Pos(), "empty expression in string interpolation")
			continue
		}
		if a.inOnerr && exprStr == "err" {
			hint := "use {error} not {err} inside onerr — the caught error is always named 'error'"
			if a.currentOnerrrAlias != "" {
				hint += fmt.Sprintf(", or {%s} via your 'onerr as %s' alias", a.currentOnerrrAlias, a.currentOnerrrAlias)
			} else {
				hint += ", or name it with 'onerr as e' and use {e}"
			}
			a.error(lit.Pos(), hint)
			continue
		}

		// Skip known onerr error variables — they're injected by codegen, not user-defined
		if a.inOnerr && (exprStr == "error" || exprStr == a.currentOnerrrAlias) {
			continue
		}

		expr, err := parseInterpolationExpression(exprStr, lit.Token.File)
		if err != nil {
			// Only report for expressions that look intentional (not format specifiers etc.)
			if identifierPattern.MatchString(exprStr) || strings.Contains(exprStr, ".") || strings.Contains(exprStr, "(") {
				a.error(lit.Pos(), fmt.Sprintf("invalid expression in string interpolation: %s", exprStr))
			}
			continue
		}

		// Patch position info for better error reporting
		patchExprPosition(expr, lit.Token.File, lit.Token.Line, lit.Token.Column+match[2])

		a.analyzeExpression(expr)
	}
}

// patchExprPosition updates position info on an expression for error reporting.
func patchExprPosition(expr ast.Expression, file string, line, column int) {
	switch e := expr.(type) {
	case *ast.Identifier:
		e.Token.File = file
		e.Token.Line = line
		e.Token.Column = column
	case *ast.MethodCallExpr:
		if obj, ok := e.Object.(*ast.Identifier); ok {
			obj.Token.File = file
			obj.Token.Line = line
			obj.Token.Column = column
		}
	case *ast.FieldAccessExpr:
		if obj, ok := e.Object.(*ast.Identifier); ok {
			obj.Token.File = file
			obj.Token.Line = line
			obj.Token.Column = column
		}
	}
}

func parseInterpolationExpression(exprStr string, filename string) (ast.Expression, error) {
	source := fmt.Sprintf("func __interp__()\n    print(%s)\n", exprStr)

	p, err := parser.New(source, filename)
	if err != nil {
		return nil, err
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		return nil, parseErrors[0]
	}
	if len(program.Declarations) == 0 {
		return nil, fmt.Errorf("missing interpolation wrapper function")
	}

	fn, ok := program.Declarations[0].(*ast.FunctionDecl)
	if !ok || fn.Body == nil || len(fn.Body.Statements) == 0 {
		return nil, fmt.Errorf("missing interpolation wrapper body")
	}

	stmt, ok := fn.Body.Statements[0].(*ast.ExpressionStmt)
	if !ok {
		return nil, fmt.Errorf("missing interpolation wrapper statement")
	}

	call, ok := stmt.Expression.(*ast.CallExpr)
	if !ok || len(call.Arguments) != 1 {
		return nil, fmt.Errorf("missing interpolation wrapper argument")
	}

	return call.Arguments[0], nil
}
