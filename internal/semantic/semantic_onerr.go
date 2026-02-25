package semantic

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/duber000/kukicha/internal/ast"
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
	// Parse string interpolations and validate expressions
	re := regexp.MustCompile(`\{([a-zA-Z_][^}]*)\}`)
	matches := re.FindAllStringSubmatch(lit.Value, -1)

	for _, match := range matches {
		exprStr := match[1]
		// For now, just validate it's not empty
		// Full expression parsing would require parsing the expression string
		if strings.TrimSpace(exprStr) == "" {
			a.error(lit.Pos(), "empty expression in string interpolation")
		}
		if a.inOnerr && strings.TrimSpace(exprStr) == "err" {
			hint := "use {error} not {err} inside onerr — the caught error is always named 'error'"
			if a.currentOnerrrAlias != "" {
				hint += fmt.Sprintf(", or {%s} via your 'onerr as %s' alias", a.currentOnerrrAlias, a.currentOnerrrAlias)
			} else {
				hint += ", or name it with 'onerr as e' and use {e}"
			}
			a.error(lit.Pos(), hint)
		}
	}
}
