package codegen

import (
	"fmt"
	"slices"
	"strings"

	"github.com/duber000/kukicha/internal/ast"
)

// generateOnErrVarDecl handles variable declarations with onerr
// e.g., val := foo() onerr panic "error" → val, err := foo(); if err != nil { panic("error") }
// e.g., port := getPort() onerr "8080" → port, err := getPort(); if err != nil { port = "8080" }
// e.g., val := foo() onerr explain "hint" → val, err := foo(); if err != nil { return ..., fmt.Errorf("hint: %w", err) }
func (g *Generator) generateOnErrVarDecl(names []*ast.Identifier, values []ast.Expression, clause *ast.OnErrClause) {
	// Build the value expression string (typically a single call expression)
	valueExpr := strings.Join(g.exprStrings(values), ", ")

	if len(names) == 1 && len(values) == 1 {
		if pipe, ok := values[0].(*ast.PipeExpr); ok {
			// Variable declarations do not have declared targets available inside
			// onerr handlers yet, so use handler forms that don't assign to names.
			if finalVar, ok := g.generateOnErrPipeChain(pipe, clause, []*ast.Identifier{}); ok {
				g.writeLine(fmt.Sprintf("%s := %s", names[0].Value, finalVar))
				return
			}
		}

		// Handle: result := A |> B() |> switch ... onerr handler
		if ps, ok := values[0].(*ast.PipedSwitchExpr); ok {
			l := newLowerer(g)
			block := l.lowerPipedSwitchVarDecl(names[0].Value, ps, clause, names)
			if block != nil {
				g.emitIR(block)
				return
			}
		}
	}

	// Check for discard case first - we can skip error handling entirely
	if g.emitOnErrDiscard(clause, identNames(names), ":=", valueExpr, len(names) > 1 && len(values) == 1, nil) {
		return
	}

	l := newLowerer(g)

	// Special case: if we have multiple names (e.g., x, err) and a single value expression,
	// the single expression likely returns multiple values including an error.
	// In this case, use the names as-is without adding an extra error variable.
	// This handles cases like: x, err := Parse() onerr handler
	// where Parse() returns (int, error) and names are [x, err]
	if len(names) > 1 && len(values) == 1 {
		block := l.lowerOnErrWithExplicitErr(identNames(names), valueExpr, clause, true)
		g.emitIR(block)
		return
	}

	// Single-return with auto-generated error variable
	block := l.lowerOnErr(valueExpr, 0, clause, identNames(names), true)
	g.emitIR(block)
}

// generateOnErrHandler generates code for the onerr handler expression
func (g *Generator) generateOnErrHandler(names []*ast.Identifier, handler ast.Expression, errVar string) {
	// If handler is nil, the explain wrapping already generated the return
	if handler == nil {
		return
	}
	switch h := handler.(type) {
	case *ast.PanicExpr:
		// onerr panic "message"
		// If message contains {error}, replace it with the actual error variable
		msg := ""
		if strLit, ok := h.Message.(*ast.StringLiteral); ok {
			msg = strLit.Value
		} else {
			msg = g.exprToString(h.Message)
		}

		if strings.Contains(msg, "{error}") {
			msg = strings.ReplaceAll(msg, "{error}", fmt.Sprintf("{%s}", errVar))
		}
		g.writeLine(fmt.Sprintf("panic(%s)", g.generateStringInterpolation(msg)))
	case *ast.ErrorExpr:
		// onerr return empty, error - generate return with error
		// This assumes the function returns (T, error)
		errExpr := g.errorValueExpr(h.Message, errVar)
		if len(names) > 0 {
			// Return the first value and the error
			g.writeLine(fmt.Sprintf("return %s, %s", names[0].Value, errExpr))
		} else {
			g.writeLine(fmt.Sprintf("return %s", errExpr))
		}
	case *ast.ReturnExpr:
		// onerr return empty, error "{error}"
		// If any value is identifier "error", replace with errVar
		// If any value is an ErrorExpr, use errorValueExpr to substitute {error}
		values := make([]string, len(h.Values))
		for i, v := range h.Values {
			if id, ok := v.(*ast.Identifier); ok && id.Value == "error" {
				values[i] = errVar
			} else if errExpr, ok := v.(*ast.ErrorExpr); ok {
				values[i] = g.errorValueExpr(errExpr.Message, errVar)
			} else {
				values[i] = g.exprToString(v)
			}
		}
		g.writeLine(fmt.Sprintf("return %s", strings.Join(values, ", ")))
	case *ast.BlockExpr:
		// onerr block handler: generate the block body.
		// The caller (renderHandler) sets currentOnErrVar so that
		// exprToString resolves "error" / alias to the actual error variable.
		g.generateBlock(h.Body)
		return
	case *ast.EmptyExpr:
		// onerr return empty - generate bare return (for named return values)
		g.writeLine("return")
	default:
		// onerr expression (default value case)
		// e.g., port := getPort() onerr "8080"
		// Assign the default value to the first variable
		if len(names) > 0 {
			g.writeLine(fmt.Sprintf("%s = %s", names[0].Value, g.exprToString(handler)))
		}
	}
}

// emitOnErrDiscard handles the discard case for all three onerr forms.
// lhsParts: pre-built target strings (nil for statement-level); op: ":=" or "=";
// valueExpr: RHS string; isMultiReturn: whether len(targets) > 1 && len(values) == 1;
// expr: the RHS expression (used for inferReturnCount at statement level).
// Returns true if discard was handled.
func (g *Generator) emitOnErrDiscard(clause *ast.OnErrClause, lhsParts []string, op string, valueExpr string, isMultiReturn bool, expr ast.Expression) bool {
	if clause.Handler == nil {
		return false
	}
	if _, isDiscard := clause.Handler.(*ast.DiscardExpr); !isDiscard {
		return false
	}

	// Statement-level (no named targets): use inferReturnCount to determine blank count
	if lhsParts == nil {
		if count, ok := g.inferReturnCount(expr); ok {
			switch count {
			case 0:
				g.writeLine(g.exprToString(expr))
			case 1:
				g.writeLine(fmt.Sprintf("_ = %s", g.exprToString(expr)))
			default:
				blanks := make([]string, count)
				for i := range blanks {
					blanks[i] = "_"
				}
				g.writeLine(fmt.Sprintf("%s = %s", strings.Join(blanks, ", "), g.exprToString(expr)))
			}
		} else {
			// Fallback: when return count inference fails, default to a single blank assignment.
			g.writeLine(fmt.Sprintf("_ = %s", g.exprToString(expr)))
		}
		return true
	}

	// Multi-value returns (last value is error): use targets as-is
	if isMultiReturn {
		g.writeLine(fmt.Sprintf("%s %s %s", strings.Join(lhsParts, ", "), op, valueExpr))
		return true
	}

	// Single-value: append _ to ignore the error
	lhsParts = append(lhsParts, "_")
	g.writeLine(fmt.Sprintf("%s %s %s", strings.Join(lhsParts, ", "), op, valueExpr))
	return true
}

// identNames extracts the string values from a slice of identifiers.
func identNames(idents []*ast.Identifier) []string {
	parts := make([]string, len(idents))
	for i, id := range idents {
		parts[i] = id.Value
	}
	return parts
}

// exprStrings converts a slice of expressions to their string representations.
func (g *Generator) exprStrings(exprs []ast.Expression) []string {
	parts := make([]string, len(exprs))
	for i, e := range exprs {
		parts[i] = g.exprToString(e)
	}
	return parts
}

// extractDefaultValue returns the Go literal string for a handler's default value
// when the handler is a simple literal (string, int, float, bool) and the returned
// type matches. Returns empty string for terminating handlers (panic, return, block).
func (g *Generator) extractDefaultValue(clause *ast.OnErrClause, _ string) string {
	if clause == nil || clause.Handler == nil {
		return ""
	}
	switch h := clause.Handler.(type) {
	case *ast.StringLiteral:
		return g.exprToString(h)
	case *ast.IntegerLiteral:
		return g.exprToString(h)
	case *ast.FloatLiteral:
		return g.exprToString(h)
	case *ast.BooleanLiteral:
		return g.exprToString(h)
	case *ast.EmptyExpr:
		return "nil"
	}
	return ""
}

func flattenPipeChain(expr ast.Expression) (ast.Expression, []ast.Expression, bool) {
	pipe, ok := expr.(*ast.PipeExpr)
	if !ok {
		return nil, nil, false
	}

	var steps []ast.Expression
	var walk func(ast.Expression) (ast.Expression, bool)
	walk = func(e ast.Expression) (ast.Expression, bool) {
		if p, ok := e.(*ast.PipeExpr); ok {
			base, ok := walk(p.Left)
			if !ok {
				return nil, false
			}
			steps = append(steps, p.Right)
			return base, true
		}
		return e, true
	}

	base, ok := walk(pipe)
	return base, steps, ok
}

func (g *Generator) generatePipedStepCall(right ast.Expression, leftExpr string) (string, bool) {
	var funcName string
	var arguments []ast.Expression
	var isVariadic bool

	if call, ok := right.(*ast.CallExpr); ok {
		funcName = g.exprToString(call.Function)
		arguments = call.Arguments
		isVariadic = call.Variadic
	} else if method, ok := right.(*ast.MethodCallExpr); ok {
		objStr := g.exprToString(method.Object)
		if alias, ok := g.pkgAliases[objStr]; ok {
			objStr = alias
		}
		funcName = objStr + "." + method.Method.Value

		if method.Object == nil {
			funcName = leftExpr + "." + method.Method.Value
			if !method.IsCall {
				return funcName, true
			}
			arguments = method.Arguments
			isVariadic = method.Variadic
		} else {
			if !method.IsCall {
				return funcName, true
			}
			arguments = method.Arguments
			isVariadic = method.Variadic
		}
	} else {
		return "", false
	}

	args := g.buildPipeArgs(leftExpr, arguments)

	if isVariadic {
		return fmt.Sprintf("%s(%s...)", funcName, strings.Join(args, ", ")), true
	}
	return fmt.Sprintf("%s(%s)", funcName, strings.Join(args, ", ")), true
}

// buildPipeArgs builds the argument list for a piped function call.
// It handles placeholder substitution (_), context-first insertion, and default data-first insertion.
func (g *Generator) buildPipeArgs(leftExpr string, arguments []ast.Expression) []string {
	placeholderIndex := -1
	for i, arg := range arguments {
		if ident, isIdent := arg.(*ast.Identifier); isIdent && ident.Value == "_" {
			placeholderIndex = i
			break
		}
		if _, isDiscard := arg.(*ast.DiscardExpr); isDiscard {
			placeholderIndex = i
			break
		}
	}

	var args []string
	if placeholderIndex != -1 {
		for i, arg := range arguments {
			if i == placeholderIndex {
				args = append(args, leftExpr)
			} else {
				args = append(args, g.exprToString(arg))
			}
		}
	} else {
		args = append(args, leftExpr)
		for _, arg := range arguments {
			args = append(args, g.exprToString(arg))
		}
	}
	return args
}

func (g *Generator) generateOnErrPipeChain(pipe *ast.PipeExpr, clause *ast.OnErrClause, names []*ast.Identifier) (string, bool) {
	l := newLowerer(g)
	nameStrs := identNames(names)
	block, finalVar := l.lowerOnErrPipeChain(pipe, clause, nameStrs)
	if block == nil {
		return "", false
	}
	g.emitIR(block)
	return finalVar, true
}

// generateOnErrStmt handles statement-level onerr
// e.g., todo |> json.MarshalWrite(w, _) onerr panic("failed")
// Generates: if err := json.MarshalWrite(w, todo); err != nil { panic("failed") }
func (g *Generator) generateOnErrStmt(expr ast.Expression, clause *ast.OnErrClause) {
	if pipe, ok := expr.(*ast.PipeExpr); ok {
		if finalVar, ok := g.generateOnErrPipeChain(pipe, clause, []*ast.Identifier{}); ok {
			g.writeLine(fmt.Sprintf("_ = %s", finalVar))
			return
		}
	} else if ps, ok := expr.(*ast.PipedSwitchExpr); ok {
		l := newLowerer(g)
		block := l.lowerPipedSwitchStmt(ps, clause)
		if block != nil {
			g.emitIR(block)
		}
		return
	}

	// Check for discard case - just execute and ignore error
	if g.emitOnErrDiscard(clause, nil, "", "", false, expr) {
		return
	}

	// Lower statement-level onerr via IR
	l := newLowerer(g)
	block := l.lowerOnErrStmt(g.exprToString(expr), expr, clause)
	g.emitIR(block)
}

// generateOnErrAssign handles assignment statements with onerr
// e.g., x = foo() onerr panic "error" → x, err = foo(); if err != nil { panic("error") }
func (g *Generator) generateOnErrAssign(stmt *ast.AssignStmt) {
	clause := stmt.OnErr

	// Build value expression
	valueExpr := strings.Join(g.exprStrings(stmt.Values), ", ")

	// Build target names for handler (convert targets to identifiers where possible)
	var names []*ast.Identifier
	for _, t := range stmt.Targets {
		if ident, ok := t.(*ast.Identifier); ok {
			names = append(names, ident)
		}
	}

	if len(stmt.Targets) == 1 && len(stmt.Values) == 1 {
		if pipe, ok := stmt.Values[0].(*ast.PipeExpr); ok {
			if finalVar, ok := g.generateOnErrPipeChain(pipe, clause, names); ok {
				g.writeLine(fmt.Sprintf("%s = %s", g.exprToString(stmt.Targets[0]), finalVar))
				return
			}
		}
	}

	// Check for discard case
	if g.emitOnErrDiscard(clause, g.exprStrings(stmt.Targets), "=", valueExpr, len(stmt.Targets) > 1 && len(stmt.Values) == 1, nil) {
		return
	}

	l := newLowerer(g)

	// Special case: if we have multiple targets (e.g., x, err) and a single value expression,
	// the single expression likely returns multiple values including an error.
	if len(stmt.Targets) > 1 && len(stmt.Values) == 1 {
		lhsParts := g.exprStrings(stmt.Targets)
		block := l.lowerOnErrWithExplicitErr(lhsParts, valueExpr, clause, false)
		g.emitIR(block)
		return
	}

	// Single-return with auto-generated error variable
	block := l.lowerOnErr(valueExpr, 0, clause, g.exprStrings(stmt.Targets), false)
	g.emitIR(block)
}

func (g *Generator) needsExplain() bool {
	for _, decl := range g.program.Declarations {
		fn, ok := decl.(*ast.FunctionDecl)
		if !ok || fn.Body == nil {
			continue
		}
		if g.blockHasExplain(fn.Body) {
			return true
		}
	}
	return false
}

func (g *Generator) blockHasExplain(block *ast.BlockStmt) bool {
	return slices.ContainsFunc(block.Statements, g.stmtHasExplain)
}

func (g *Generator) stmtHasExplain(stmt ast.Statement) bool {
	switch s := stmt.(type) {
	case *ast.VarDeclStmt:
		if s.OnErr != nil && s.OnErr.Explain != "" {
			return true
		}
	case *ast.AssignStmt:
		if s.OnErr != nil && s.OnErr.Explain != "" {
			return true
		}
	case *ast.ExpressionStmt:
		if s.OnErr != nil && s.OnErr.Explain != "" {
			return true
		}
	case *ast.IfStmt:
		if s.Consequence != nil && g.blockHasExplain(s.Consequence) {
			return true
		}
	case *ast.ForRangeStmt:
		if s.Body != nil && g.blockHasExplain(s.Body) {
			return true
		}
	case *ast.ForNumericStmt:
		if s.Body != nil && g.blockHasExplain(s.Body) {
			return true
		}
	case *ast.ForConditionStmt:
		if s.Body != nil && g.blockHasExplain(s.Body) {
			return true
		}
	}
	return false
}
