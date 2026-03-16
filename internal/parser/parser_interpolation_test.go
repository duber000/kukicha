package parser

import (
	"testing"

	"github.com/duber000/kukicha/internal/ast"
)

func TestStringInterpolationParts_SimpleIdentifier(t *testing.T) {
	prog := mustParseProgram(t, `func Greet(name string) string
    return "Hello {name}!"
`)

	fn := prog.Declarations[0].(*ast.FunctionDecl)
	ret := fn.Body.Statements[0].(*ast.ReturnStmt)
	lit := ret.Values[0].(*ast.StringLiteral)

	if !lit.Interpolated {
		t.Fatal("expected Interpolated to be true")
	}
	if len(lit.Parts) != 3 {
		t.Fatalf("expected 3 parts, got %d", len(lit.Parts))
	}

	// Part 0: literal "Hello "
	if !lit.Parts[0].IsLiteral || lit.Parts[0].Literal != "Hello " {
		t.Errorf("part 0: expected literal 'Hello ', got %+v", lit.Parts[0])
	}

	// Part 1: expression (Identifier "name")
	if lit.Parts[1].IsLiteral {
		t.Errorf("part 1: expected expression, got literal")
	}
	ident, ok := lit.Parts[1].Expr.(*ast.Identifier)
	if !ok {
		t.Fatalf("part 1: expected Identifier, got %T", lit.Parts[1].Expr)
	}
	if ident.Value != "name" {
		t.Errorf("part 1: expected 'name', got '%s'", ident.Value)
	}

	// Part 2: literal "!"
	if !lit.Parts[2].IsLiteral || lit.Parts[2].Literal != "!" {
		t.Errorf("part 2: expected literal '!', got %+v", lit.Parts[2])
	}
}

func TestStringInterpolationParts_PipeExpr(t *testing.T) {
	prog := mustParseProgram(t, `func Show(data string) string
    return "{data |> slice.First()}"
`)

	fn := prog.Declarations[0].(*ast.FunctionDecl)
	ret := fn.Body.Statements[0].(*ast.ReturnStmt)
	lit := ret.Values[0].(*ast.StringLiteral)

	if len(lit.Parts) != 1 {
		t.Fatalf("expected 1 part, got %d", len(lit.Parts))
	}
	if lit.Parts[0].IsLiteral {
		t.Error("expected expression part, got literal")
	}
	// Should be a PipeExpr
	if _, ok := lit.Parts[0].Expr.(*ast.PipeExpr); !ok {
		t.Errorf("expected PipeExpr, got %T", lit.Parts[0].Expr)
	}
}

func TestStringInterpolationParts_MultipleExprs(t *testing.T) {
	prog := mustParseProgram(t, `func Fmt(a int, b int) string
    return "{a} and {b}"
`)

	fn := prog.Declarations[0].(*ast.FunctionDecl)
	ret := fn.Body.Statements[0].(*ast.ReturnStmt)
	lit := ret.Values[0].(*ast.StringLiteral)

	if len(lit.Parts) != 3 {
		t.Fatalf("expected 3 parts, got %d", len(lit.Parts))
	}

	// expr, literal, expr
	if lit.Parts[0].IsLiteral {
		t.Error("part 0: expected expression")
	}
	if !lit.Parts[1].IsLiteral || lit.Parts[1].Literal != " and " {
		t.Errorf("part 1: expected literal ' and ', got %+v", lit.Parts[1])
	}
	if lit.Parts[2].IsLiteral {
		t.Error("part 2: expected expression")
	}
}

func TestStringInterpolationParts_TypeCast(t *testing.T) {
	prog := mustParseProgram(t, `func Show(x int) string
    return "{x as string}"
`)

	fn := prog.Declarations[0].(*ast.FunctionDecl)
	ret := fn.Body.Statements[0].(*ast.ReturnStmt)
	lit := ret.Values[0].(*ast.StringLiteral)

	if len(lit.Parts) != 1 {
		t.Fatalf("expected 1 part, got %d", len(lit.Parts))
	}
	if lit.Parts[0].IsLiteral {
		t.Error("expected expression part")
	}
	if _, ok := lit.Parts[0].Expr.(*ast.TypeCastExpr); !ok {
		t.Errorf("expected TypeCastExpr, got %T", lit.Parts[0].Expr)
	}
}

func TestStringInterpolationParts_EscapedBraces(t *testing.T) {
	// \{ and \} produce PUA sentinels, not real braces — should NOT be interpolated
	prog := mustParseProgram(t, `func Format() string
    return "\{key\}"
`)

	fn := prog.Declarations[0].(*ast.FunctionDecl)
	ret := fn.Body.Statements[0].(*ast.ReturnStmt)
	lit := ret.Values[0].(*ast.StringLiteral)

	// Escaped braces should not trigger interpolation — Parts should be nil or empty
	if lit.Interpolated {
		t.Error("escaped braces should not set Interpolated to true")
	}
}

func TestStringInterpolationParts_FieldAccess(t *testing.T) {
	prog := mustParseProgram(t, `type User
    name string

func Show(u User) string
    return "Hello {u.name}!"
`)

	fn := prog.Declarations[1].(*ast.FunctionDecl)
	ret := fn.Body.Statements[0].(*ast.ReturnStmt)
	lit := ret.Values[0].(*ast.StringLiteral)

	if len(lit.Parts) != 3 {
		t.Fatalf("expected 3 parts, got %d", len(lit.Parts))
	}
	if lit.Parts[1].IsLiteral {
		t.Error("part 1: expected expression")
	}
	if _, ok := lit.Parts[1].Expr.(*ast.FieldAccessExpr); !ok {
		t.Errorf("part 1: expected FieldAccessExpr, got %T", lit.Parts[1].Expr)
	}
}
