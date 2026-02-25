package codegen

import (
	"github.com/duber000/kukicha/internal/ast"
)

// walkProgram calls visit for every expression reachable from any function
// body in the program. Returns true (and stops early) the moment visit returns
// true for any expression.
func (g *Generator) walkProgram(visit func(ast.Expression) bool) bool {
	for _, decl := range g.program.Declarations {
		if fn, ok := decl.(*ast.FunctionDecl); ok {
			if fn.Body != nil && g.walkBlock(fn.Body, visit) {
				return true
			}
		}
	}
	return false
}

// walkBlock walks all statements in block, short-circuiting on the first true.
func (g *Generator) walkBlock(block *ast.BlockStmt, visit func(ast.Expression) bool) bool {
	for _, stmt := range block.Statements {
		if g.walkStmt(stmt, visit) {
			return true
		}
	}
	return false
}

// walkStmt walks all expressions reachable from stmt.
func (g *Generator) walkStmt(stmt ast.Statement, visit func(ast.Expression) bool) bool {
	if stmt == nil {
		return false
	}
	switch s := stmt.(type) {
	case *ast.VarDeclStmt:
		for _, v := range s.Values {
			if g.walkExpr(v, visit) {
				return true
			}
		}
		if s.OnErr != nil && g.walkExpr(s.OnErr.Handler, visit) {
			return true
		}
	case *ast.AssignStmt:
		for _, t := range s.Targets {
			if g.walkExpr(t, visit) {
				return true
			}
		}
		for _, v := range s.Values {
			if g.walkExpr(v, visit) {
				return true
			}
		}
		if s.OnErr != nil && g.walkExpr(s.OnErr.Handler, visit) {
			return true
		}
	case *ast.ReturnStmt:
		for _, v := range s.Values {
			if g.walkExpr(v, visit) {
				return true
			}
		}
	case *ast.IncDecStmt:
		if g.walkExpr(s.Variable, visit) {
			return true
		}
	case *ast.IfStmt:
		if s.Init != nil && g.walkStmt(s.Init, visit) {
			return true
		}
		if g.walkExpr(s.Condition, visit) {
			return true
		}
		if s.Consequence != nil && g.walkBlock(s.Consequence, visit) {
			return true
		}
		if s.Alternative != nil && g.walkStmt(s.Alternative, visit) {
			return true
		}
	case *ast.ElseStmt:
		if s.Body != nil && g.walkBlock(s.Body, visit) {
			return true
		}
	case *ast.SwitchStmt:
		if s.Expression != nil && g.walkExpr(s.Expression, visit) {
			return true
		}
		for _, c := range s.Cases {
			for _, v := range c.Values {
				if g.walkExpr(v, visit) {
					return true
				}
			}
			if c.Body != nil && g.walkBlock(c.Body, visit) {
				return true
			}
		}
		if s.Otherwise != nil && s.Otherwise.Body != nil && g.walkBlock(s.Otherwise.Body, visit) {
			return true
		}
	case *ast.SelectStmt:
		for _, c := range s.Cases {
			if c.Recv != nil && g.walkExpr(c.Recv, visit) {
				return true
			}
			if c.Send != nil {
				if g.walkExpr(c.Send.Value, visit) {
					return true
				}
				if g.walkExpr(c.Send.Channel, visit) {
					return true
				}
			}
			if c.Body != nil && g.walkBlock(c.Body, visit) {
				return true
			}
		}
		if s.Otherwise != nil && s.Otherwise.Body != nil && g.walkBlock(s.Otherwise.Body, visit) {
			return true
		}
	case *ast.TypeSwitchStmt:
		if g.walkExpr(s.Expression, visit) {
			return true
		}
		for _, c := range s.Cases {
			if c.Body != nil && g.walkBlock(c.Body, visit) {
				return true
			}
		}
		if s.Otherwise != nil && s.Otherwise.Body != nil && g.walkBlock(s.Otherwise.Body, visit) {
			return true
		}
	case *ast.ForRangeStmt:
		if g.walkExpr(s.Collection, visit) {
			return true
		}
		if s.Body != nil && g.walkBlock(s.Body, visit) {
			return true
		}
	case *ast.ForNumericStmt:
		if g.walkExpr(s.Start, visit) {
			return true
		}
		if g.walkExpr(s.End, visit) {
			return true
		}
		if s.Body != nil && g.walkBlock(s.Body, visit) {
			return true
		}
	case *ast.ForConditionStmt:
		if g.walkExpr(s.Condition, visit) {
			return true
		}
		if s.Body != nil && g.walkBlock(s.Body, visit) {
			return true
		}
	case *ast.DeferStmt:
		if g.walkExpr(s.Call, visit) {
			return true
		}
	case *ast.GoStmt:
		if s.Call != nil && g.walkExpr(s.Call, visit) {
			return true
		}
		if s.Block != nil && g.walkBlock(s.Block, visit) {
			return true
		}
	case *ast.SendStmt:
		if g.walkExpr(s.Value, visit) {
			return true
		}
		if g.walkExpr(s.Channel, visit) {
			return true
		}
	case *ast.ExpressionStmt:
		if g.walkExpr(s.Expression, visit) {
			return true
		}
		if s.OnErr != nil && g.walkExpr(s.OnErr.Handler, visit) {
			return true
		}
	}
	return false
}

// walkExpr calls visit(expr) first; if visit returns true, walkExpr returns
// true immediately (short-circuit). Otherwise it recurses into all
// sub-expressions and returns true as soon as any recursive call does.
func (g *Generator) walkExpr(expr ast.Expression, visit func(ast.Expression) bool) bool {
	if expr == nil {
		return false
	}
	if visit(expr) {
		return true
	}
	switch e := expr.(type) {
	case *ast.BinaryExpr:
		return g.walkExpr(e.Left, visit) || g.walkExpr(e.Right, visit)
	case *ast.UnaryExpr:
		return g.walkExpr(e.Right, visit)
	case *ast.PipeExpr:
		return g.walkExpr(e.Left, visit) || g.walkExpr(e.Right, visit)
	case *ast.CallExpr:
		if g.walkExpr(e.Function, visit) {
			return true
		}
		for _, arg := range e.Arguments {
			if g.walkExpr(arg, visit) {
				return true
			}
		}
		for _, na := range e.NamedArguments {
			if g.walkExpr(na.Value, visit) {
				return true
			}
		}
	case *ast.MethodCallExpr:
		if g.walkExpr(e.Object, visit) {
			return true
		}
		for _, arg := range e.Arguments {
			if g.walkExpr(arg, visit) {
				return true
			}
		}
		for _, na := range e.NamedArguments {
			if g.walkExpr(na.Value, visit) {
				return true
			}
		}
	case *ast.IndexExpr:
		return g.walkExpr(e.Left, visit) || g.walkExpr(e.Index, visit)
	case *ast.SliceExpr:
		if g.walkExpr(e.Left, visit) {
			return true
		}
		if g.walkExpr(e.Start, visit) {
			return true
		}
		if g.walkExpr(e.End, visit) {
			return true
		}
	case *ast.ErrorExpr:
		return g.walkExpr(e.Message, visit)
	case *ast.PanicExpr:
		return g.walkExpr(e.Message, visit)
	case *ast.ReturnExpr:
		for _, v := range e.Values {
			if g.walkExpr(v, visit) {
				return true
			}
		}
	case *ast.MakeExpr:
		for _, arg := range e.Args {
			if g.walkExpr(arg, visit) {
				return true
			}
		}
	case *ast.CloseExpr:
		return g.walkExpr(e.Channel, visit)
	case *ast.ReceiveExpr:
		return g.walkExpr(e.Channel, visit)
	case *ast.AddressOfExpr:
		return g.walkExpr(e.Operand, visit)
	case *ast.DerefExpr:
		return g.walkExpr(e.Operand, visit)
	case *ast.TypeCastExpr:
		return g.walkExpr(e.Expression, visit)
	case *ast.TypeAssertionExpr:
		return g.walkExpr(e.Expression, visit)
	case *ast.StructLiteralExpr:
		for _, f := range e.Fields {
			if g.walkExpr(f.Value, visit) {
				return true
			}
		}
	case *ast.ListLiteralExpr:
		for _, elem := range e.Elements {
			if g.walkExpr(elem, visit) {
				return true
			}
		}
	case *ast.MapLiteralExpr:
		for _, pair := range e.Pairs {
			if g.walkExpr(pair.Key, visit) || g.walkExpr(pair.Value, visit) {
				return true
			}
		}
	case *ast.FunctionLiteral:
		if e.Body != nil && g.walkBlock(e.Body, visit) {
			return true
		}
	case *ast.ArrowLambda:
		if e.Body != nil && g.walkExpr(e.Body, visit) {
			return true
		}
		if e.Block != nil && g.walkBlock(e.Block, visit) {
			return true
		}
	case *ast.BlockExpr:
		if e.Body != nil && g.walkBlock(e.Body, visit) {
			return true
		}
	}
	return false
}

// needsStringInterpolation returns true if any string literal in the program
// uses interpolation (e.g., "hello {name}") in a context that requires
// fmt.Sprintf. Printf-style method call format strings (e.g., t.Errorf("hello {name}"))
// are handled inline without fmt.Sprintf and are excluded from this check.
func (g *Generator) needsStringInterpolation() bool {
	for _, decl := range g.program.Declarations {
		fn, ok := decl.(*ast.FunctionDecl)
		if !ok || fn.Body == nil {
			continue
		}
		if g.blockHasNonPrintfInterpolation(fn.Body) {
			return true
		}
	}
	return false
}

// blockHasNonPrintfInterpolation returns true if any interpolated string in the
// block would require fmt.Sprintf (i.e., is not the format string of a printf-style
// method call).
func (g *Generator) blockHasNonPrintfInterpolation(block *ast.BlockStmt) bool {
	for _, stmt := range block.Statements {
		if g.stmtHasNonPrintfInterpolation(stmt) {
			return true
		}
	}
	return false
}

// stmtHasNonPrintfInterpolation recurses through a statement checking for
// interpolated strings that require fmt.Sprintf.
func (g *Generator) stmtHasNonPrintfInterpolation(stmt ast.Statement) bool {
	switch s := stmt.(type) {
	case *ast.VarDeclStmt:
		for _, v := range s.Values {
			if g.exprHasNonPrintfInterpolation(v) {
				return true
			}
		}
		if s.OnErr != nil && g.exprHasNonPrintfInterpolation(s.OnErr.Handler) {
			return true
		}
	case *ast.AssignStmt:
		for _, v := range s.Values {
			if g.exprHasNonPrintfInterpolation(v) {
				return true
			}
		}
		if s.OnErr != nil && g.exprHasNonPrintfInterpolation(s.OnErr.Handler) {
			return true
		}
	case *ast.ReturnStmt:
		for _, v := range s.Values {
			if g.exprHasNonPrintfInterpolation(v) {
				return true
			}
		}
	case *ast.IfStmt:
		if g.exprHasNonPrintfInterpolation(s.Condition) {
			return true
		}
		if s.Consequence != nil && g.blockHasNonPrintfInterpolation(s.Consequence) {
			return true
		}
		if s.Alternative != nil {
			if g.stmtHasNonPrintfInterpolation(s.Alternative) {
				return true
			}
		}
	case *ast.ExpressionStmt:
		if s.Expression != nil && g.exprHasNonPrintfInterpolation(s.Expression) {
			return true
		}
		if s.OnErr != nil && g.exprHasNonPrintfInterpolation(s.OnErr.Handler) {
			return true
		}
	case *ast.ForRangeStmt:
		if g.exprHasNonPrintfInterpolation(s.Collection) {
			return true
		}
		if s.Body != nil && g.blockHasNonPrintfInterpolation(s.Body) {
			return true
		}
	case *ast.ForNumericStmt:
		if g.exprHasNonPrintfInterpolation(s.Start) || g.exprHasNonPrintfInterpolation(s.End) {
			return true
		}
		if s.Body != nil && g.blockHasNonPrintfInterpolation(s.Body) {
			return true
		}
	case *ast.ForConditionStmt:
		if g.exprHasNonPrintfInterpolation(s.Condition) {
			return true
		}
		if s.Body != nil && g.blockHasNonPrintfInterpolation(s.Body) {
			return true
		}
	}
	return false
}

// exprHasNonPrintfInterpolation returns true if expr contains any interpolated
// string literal that would require fmt.Sprintf. Format strings that are the
// first argument to a printf-style method are excluded.
func (g *Generator) exprHasNonPrintfInterpolation(expr ast.Expression) bool {
	if expr == nil {
		return false
	}
	switch e := expr.(type) {
	case *ast.StringLiteral:
		return e.Interpolated
	case *ast.BinaryExpr:
		return g.exprHasNonPrintfInterpolation(e.Left) || g.exprHasNonPrintfInterpolation(e.Right)
	case *ast.UnaryExpr:
		return g.exprHasNonPrintfInterpolation(e.Right)
	case *ast.CallExpr:
		if g.exprHasNonPrintfInterpolation(e.Function) {
			return true
		}
		for _, arg := range e.Arguments {
			if g.exprHasNonPrintfInterpolation(arg) {
				return true
			}
		}
	case *ast.MethodCallExpr:
		// For printf-style methods, skip argument 0 (the format string) — it is
		// rendered inline via %v substitution, not via fmt.Sprintf.
		startIdx := 0
		if g.isPrintfStyleMethod(e.Method.Value) && len(e.Arguments) > 0 {
			startIdx = 1
		}
		for i := startIdx; i < len(e.Arguments); i++ {
			if g.exprHasNonPrintfInterpolation(e.Arguments[i]) {
				return true
			}
		}
	case *ast.PipeExpr:
		return g.exprHasNonPrintfInterpolation(e.Left) || g.exprHasNonPrintfInterpolation(e.Right)
	case *ast.ErrorExpr:
		return g.exprHasNonPrintfInterpolation(e.Message)
	case *ast.FunctionLiteral:
		if e.Body != nil {
			return g.blockHasNonPrintfInterpolation(e.Body)
		}
	case *ast.ArrowLambda:
		if e.Body != nil && g.exprHasNonPrintfInterpolation(e.Body) {
			return true
		}
		if e.Block != nil && g.blockHasNonPrintfInterpolation(e.Block) {
			return true
		}
	}
	return false
}

// needsPrintBuiltin returns true if any call to the print() builtin exists in
// the program.
func (g *Generator) needsPrintBuiltin() bool {
	return g.walkProgram(func(e ast.Expression) bool {
		call, ok := e.(*ast.CallExpr)
		if !ok {
			return false
		}
		id, ok := call.Function.(*ast.Identifier)
		return ok && id.Value == "print"
	})
}

// needsErrorsPackage returns true if any error() expression (which generates a
// call to errors.New) is used in the program.
func (g *Generator) needsErrorsPackage() bool {
	return g.walkProgram(func(e ast.Expression) bool {
		_, ok := e.(*ast.ErrorExpr)
		return ok
	})
}
