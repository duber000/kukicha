package parser

import (
	"github.com/duber000/kukicha/internal/ast"
	"testing"
)

func TestParseForRangeLoop(t *testing.T) {
	input := `func Test(items list of int)
    for item in items
        print(item)
`

	program := mustParseProgram(t, input)

	fn := program.Declarations[0].(*ast.FunctionDecl)
	forStmt, ok := fn.Body.Statements[0].(*ast.ForRangeStmt)
	if !ok {
		t.Fatalf("expected ForRangeStmt, got %T", fn.Body.Statements[0])
	}

	if forStmt.Variable.Value != "item" {
		t.Errorf("expected variable 'item', got '%s'", forStmt.Variable.Value)
	}
}

func TestParseForNumericLoop(t *testing.T) {
	input := `func Test()
    for i from 0 to 10
        print(i)
`

	program := mustParseProgram(t, input)

	fn := program.Declarations[0].(*ast.FunctionDecl)
	forStmt, ok := fn.Body.Statements[0].(*ast.ForNumericStmt)
	if !ok {
		t.Fatalf("expected ForNumericStmt, got %T", fn.Body.Statements[0])
	}

	if forStmt.Variable.Value != "i" {
		t.Errorf("expected variable 'i', got '%s'", forStmt.Variable.Value)
	}

	if forStmt.Through {
		t.Error("expected 'to' loop, got 'through'")
	}
}

func TestParseBinaryExpression(t *testing.T) {
	tests := []struct {
		input    string
		operator string
	}{
		{`func Test() int
    return 1 + 2
`, "+"},
		{`func Test() int
    return 1 - 2
`, "-"},
		{`func Test() int
    return 1 * 2
`, "*"},
		{`func Test() int
    return 1 / 2
`, "/"},
		{`func Test() bool
    return 1 == 2
`, "=="},
		{`func Test() bool
    return 1 != 2
`, "!="},
	}

	for _, tt := range tests {
		t.Run(tt.operator, func(t *testing.T) {
			t.Parallel()

			p, err := New(tt.input, "test.kuki")
			if err != nil {
				t.Fatalf("lexer error: %v", err)
			}
			program, errors := p.Parse()

			if len(errors) > 0 {
				t.Fatalf("parser errors for operator %s: %v", tt.operator, errors)
			}

			fn := program.Declarations[0].(*ast.FunctionDecl)
			retStmt := fn.Body.Statements[0].(*ast.ReturnStmt)
			binExpr, ok := retStmt.Values[0].(*ast.BinaryExpr)
			if !ok {
				t.Fatalf("expected BinaryExpr, got %T", retStmt.Values[0])
			}

			if binExpr.Operator != tt.operator {
				t.Errorf("expected operator '%s', got '%s'", tt.operator, binExpr.Operator)
			}
		})
	}
}

func TestParsePipeExpression(t *testing.T) {
	input := `func Test() string
    return "hello" |> ToUpper()
`

	program := mustParseProgram(t, input)

	fn := program.Declarations[0].(*ast.FunctionDecl)
	retStmt := fn.Body.Statements[0].(*ast.ReturnStmt)
	pipeExpr, ok := retStmt.Values[0].(*ast.PipeExpr)
	if !ok {
		t.Fatalf("expected PipeExpr, got %T", retStmt.Values[0])
	}

	if pipeExpr.Left == nil {
		t.Error("expected left expression, got nil")
	}

	if pipeExpr.Right == nil {
		t.Error("expected right expression, got nil")
	}
}

func TestParsePipeExpressionMultiLine(t *testing.T) {
	input := `func Test() string
    return "hello" |>
        ToUpper() |>
        TrimSpace()
`

	p, err := New(input, "test.kuki")
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}
	program, errors := p.Parse()
	if len(errors) > 0 {
		t.Fatalf("parser errors: %v", errors)
	}

	fn := program.Declarations[0].(*ast.FunctionDecl)
	retStmt := fn.Body.Statements[0].(*ast.ReturnStmt)

	// The outer pipe: (_ |> TrimSpace())
	outerPipe, ok := retStmt.Values[0].(*ast.PipeExpr)
	if !ok {
		t.Fatalf("expected outer PipeExpr, got %T", retStmt.Values[0])
	}

	// The inner pipe: ("hello" |> ToUpper())
	innerPipe, ok := outerPipe.Left.(*ast.PipeExpr)
	if !ok {
		t.Fatalf("expected inner PipeExpr on Left, got %T", outerPipe.Left)
	}

	if innerPipe.Left == nil || innerPipe.Right == nil {
		t.Error("inner pipe has nil Left or Right")
	}
	if outerPipe.Right == nil {
		t.Error("outer pipe has nil Right")
	}
}

func TestParseOnErrStatement(t *testing.T) {
	input := `func Test()
    val := ReadFile("test.txt") onerr 0
`

	program := mustParseProgram(t, input)

	fn := program.Declarations[0].(*ast.FunctionDecl)
	varDecl, ok := fn.Body.Statements[0].(*ast.VarDeclStmt)
	if !ok {
		t.Fatalf("expected VarDeclStmt, got %T", fn.Body.Statements[0])
	}

	if varDecl.OnErr == nil {
		t.Fatal("expected OnErr clause on VarDeclStmt, got nil")
	}

	if varDecl.OnErr.Handler == nil {
		t.Error("expected handler expression in OnErr clause, got nil")
	}

	if len(varDecl.Values) != 1 {
		t.Fatalf("expected 1 value expression, got %d", len(varDecl.Values))
	}

	// The value should be the call expression (ReadFile("test.txt")), not an OnErrExpr
	if _, ok := varDecl.Values[0].(*ast.CallExpr); !ok {
		t.Errorf("expected CallExpr as value, got %T", varDecl.Values[0])
	}
}

func TestParseListType(t *testing.T) {
	input := `func Test(items list of string)
    return items
`

	program := mustParseProgram(t, input)

	fn := program.Declarations[0].(*ast.FunctionDecl)
	param := fn.Parameters[0]

	listType, ok := param.Type.(*ast.ListType)
	if !ok {
		t.Fatalf("expected ListType, got %T", param.Type)
	}

	elemType, ok := listType.ElementType.(*ast.PrimitiveType)
	if !ok {
		t.Fatalf("expected PrimitiveType for element, got %T", listType.ElementType)
	}

	if elemType.Name != "string" {
		t.Errorf("expected element type 'string', got '%s'", elemType.Name)
	}
}

func TestParseMapType(t *testing.T) {
	input := `func Test(m map of string to int)
    return m
`

	program := mustParseProgram(t, input)

	fn := program.Declarations[0].(*ast.FunctionDecl)
	param := fn.Parameters[0]

	mapType, ok := param.Type.(*ast.MapType)
	if !ok {
		t.Fatalf("expected MapType, got %T", param.Type)
	}

	keyType, ok := mapType.KeyType.(*ast.PrimitiveType)
	if !ok {
		t.Fatalf("expected PrimitiveType for key, got %T", mapType.KeyType)
	}

	if keyType.Name != "string" {
		t.Errorf("expected key type 'string', got '%s'", keyType.Name)
	}

	valType, ok := mapType.ValueType.(*ast.PrimitiveType)
	if !ok {
		t.Fatalf("expected PrimitiveType for value, got %T", mapType.ValueType)
	}

	if valType.Name != "int" {
		t.Errorf("expected value type 'int', got '%s'", valType.Name)
	}
}

func TestParseReferenceType(t *testing.T) {
	input := `func Test(p reference Person)
    return p
`

	program := mustParseProgram(t, input)

	fn := program.Declarations[0].(*ast.FunctionDecl)
	param := fn.Parameters[0]

	refType, ok := param.Type.(*ast.ReferenceType)
	if !ok {
		t.Fatalf("expected ReferenceType, got %T", param.Type)
	}

	elemType, ok := refType.ElementType.(*ast.NamedType)
	if !ok {
		t.Fatalf("expected NamedType for element, got %T", refType.ElementType)
	}

	if elemType.Name != "Person" {
		t.Errorf("expected element type 'Person', got '%s'", elemType.Name)
	}
}

func TestParseImportDeclaration(t *testing.T) {
	input := `import "fmt"
import "strings" as str

func Test()
    print("hello")
`

	program := mustParseProgram(t, input)

	if len(program.Imports) != 2 {
		t.Fatalf("expected 2 imports, got %d", len(program.Imports))
	}

	imp1 := program.Imports[0]
	if imp1.Path.Value != "fmt" {
		t.Errorf("expected import path 'fmt', got '%s'", imp1.Path.Value)
	}

	if imp1.Alias != nil {
		t.Errorf("expected no alias for first import, got '%s'", imp1.Alias.Value)
	}

	imp2 := program.Imports[1]
	if imp2.Path.Value != "strings" {
		t.Errorf("expected import path 'strings', got '%s'", imp2.Path.Value)
	}

	if imp2.Alias == nil {
		t.Error("expected alias for second import, got nil")
	} else if imp2.Alias.Value != "str" {
		t.Errorf("expected alias 'str', got '%s'", imp2.Alias.Value)
	}
}

func TestParsePetioleDeclaration(t *testing.T) {
	input := `petiole main

func Main()
    print("hello")
`

	program := mustParseProgram(t, input)

	if program.PetioleDecl == nil {
		t.Fatal("expected petiole declaration, got nil")
	}

	if program.PetioleDecl.Name.Value != "main" {
		t.Errorf("expected petiole name 'main', got '%s'", program.PetioleDecl.Name.Value)
	}
}

func TestParseComplexExpression(t *testing.T) {
	input := `func Test() bool
    return a + b * c > d and e or f
`

	program := mustParseProgram(t, input)

	fn := program.Declarations[0].(*ast.FunctionDecl)
	retStmt := fn.Body.Statements[0].(*ast.ReturnStmt)

	// Should parse as: ((a + (b * c)) > d) and e) or f
	// Top level should be 'or'
	orExpr, ok := retStmt.Values[0].(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr at top level, got %T", retStmt.Values[0])
	}

	if orExpr.Operator != "or" {
		t.Errorf("expected top-level operator 'or', got '%s'", orExpr.Operator)
	}
}

func TestParseWalrusOperator(t *testing.T) {
	input := `func Test()
    x := 42
    print(x)
`

	program := mustParseProgram(t, input)

	fn := program.Declarations[0].(*ast.FunctionDecl)
	varDecl, ok := fn.Body.Statements[0].(*ast.VarDeclStmt)
	if !ok {
		t.Fatalf("expected VarDeclStmt, got %T", fn.Body.Statements[0])
	}

	if len(varDecl.Names) != 1 || varDecl.Names[0].Value != "x" {
		t.Errorf("expected variable name 'x', got '%v'", varDecl.Names)
	}

	intLit, ok := varDecl.Values[0].(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("expected IntegerLiteral, got %T", varDecl.Values[0])
	}

	if intLit.Value != 42 {
		t.Errorf("expected value 42, got %d", intLit.Value)
	}
}

func TestParseMethodCall(t *testing.T) {
	input := `func Test(s string) int
    return s.Length()
`

	program := mustParseProgram(t, input)

	fn := program.Declarations[0].(*ast.FunctionDecl)
	retStmt := fn.Body.Statements[0].(*ast.ReturnStmt)

	methodCall, ok := retStmt.Values[0].(*ast.MethodCallExpr)
	if !ok {
		t.Fatalf("expected MethodCallExpr, got %T", retStmt.Values[0])
	}

	if methodCall.Method.Value != "Length" {
		t.Errorf("expected method name 'Length', got '%s'", methodCall.Method.Value)
	}
}

func TestParseFieldAccess(t *testing.T) {
	input := `func Test(user User) string
    return user.name
`

	program := mustParseProgram(t, input)

	fn := program.Declarations[0].(*ast.FunctionDecl)
	retStmt := fn.Body.Statements[0].(*ast.ReturnStmt)

	fieldAccess, ok := retStmt.Values[0].(*ast.FieldAccessExpr)
	if !ok {
		t.Fatalf("expected FieldAccessExpr, got %T", retStmt.Values[0])
	}

	if fieldAccess.Field.Value != "name" {
		t.Errorf("expected field name 'name', got '%s'", fieldAccess.Field.Value)
	}
}

func TestParsePipedFieldAccess(t *testing.T) {
	input := `func Test(user User) string
    return user |> .name
`

	program := mustParseProgram(t, input)

	fn := program.Declarations[0].(*ast.FunctionDecl)
	retStmt := fn.Body.Statements[0].(*ast.ReturnStmt)

	pipeExpr, ok := retStmt.Values[0].(*ast.PipeExpr)
	if !ok {
		t.Fatalf("expected PipeExpr, got %T", retStmt.Values[0])
	}

	fieldAccess, ok := pipeExpr.Right.(*ast.FieldAccessExpr)
	if !ok {
		t.Fatalf("expected FieldAccessExpr on pipe right side, got %T", pipeExpr.Right)
	}

	if fieldAccess.Object != nil {
		t.Fatalf("expected shorthand field access to have nil object, got %T", fieldAccess.Object)
	}

	if fieldAccess.Field.Value != "name" {
		t.Errorf("expected field name 'name', got '%s'", fieldAccess.Field.Value)
	}
}

func TestParseIndexExpression(t *testing.T) {
	input := `func Test(items list of int) int
    return items[0]
`

	program := mustParseProgram(t, input)

	fn := program.Declarations[0].(*ast.FunctionDecl)
	retStmt := fn.Body.Statements[0].(*ast.ReturnStmt)

	indexExpr, ok := retStmt.Values[0].(*ast.IndexExpr)
	if !ok {
		t.Fatalf("expected IndexExpr, got %T", retStmt.Values[0])
	}

	intLit, ok := indexExpr.Index.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("expected IntegerLiteral for index, got %T", indexExpr.Index)
	}

	if intLit.Value != 0 {
		t.Errorf("expected index 0, got %d", intLit.Value)
	}
}

func TestParseSliceExpression(t *testing.T) {
	input := `func Test(items list of int) list of int
    return items[1:3]
`

	program := mustParseProgram(t, input)

	fn := program.Declarations[0].(*ast.FunctionDecl)
	retStmt := fn.Body.Statements[0].(*ast.ReturnStmt)

	sliceExpr, ok := retStmt.Values[0].(*ast.SliceExpr)
	if !ok {
		t.Fatalf("expected SliceExpr, got %T", retStmt.Values[0])
	}

	if sliceExpr.Start == nil {
		t.Error("expected start index, got nil")
	}

	if sliceExpr.End == nil {
		t.Error("expected end index, got nil")
	}
}
