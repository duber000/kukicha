package parser

import (
	"strings"
	"testing"

	"github.com/duber000/kukicha/internal/ast"
)

func TestParseSimpleFunction(t *testing.T) {
	input := `func Add(a int, b int) int
    return a + b
`

	p, err := New(input, "test.kuki")
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}
	program, errors := p.Parse()

	if len(errors) > 0 {
		t.Fatalf("parser errors: %v", errors)
	}

	if len(program.Declarations) != 1 {
		t.Fatalf("expected 1 declaration, got %d", len(program.Declarations))
	}

	fn, ok := program.Declarations[0].(*ast.FunctionDecl)
	if !ok {
		t.Fatalf("expected FunctionDecl, got %T", program.Declarations[0])
	}

	if fn.Name.Value != "Add" {
		t.Errorf("expected function name 'Add', got '%s'", fn.Name.Value)
	}

	if len(fn.Parameters) != 2 {
		t.Errorf("expected 2 parameters, got %d", len(fn.Parameters))
	}

	if len(fn.Returns) != 1 {
		t.Errorf("expected 1 return type, got %d", len(fn.Returns))
	}
}

func TestParseTypeDeclaration(t *testing.T) {
	input := `type Person
    Name string
    Age int
`

	p, err := New(input, "test.kuki")
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}
	program, errors := p.Parse()

	if len(errors) > 0 {
		t.Fatalf("parser errors: %v", errors)
	}

	if len(program.Declarations) != 1 {
		t.Fatalf("expected 1 declaration, got %d", len(program.Declarations))
	}

	typeDecl, ok := program.Declarations[0].(*ast.TypeDecl)
	if !ok {
		t.Fatalf("expected TypeDecl, got %T", program.Declarations[0])
	}

	if typeDecl.Name.Value != "Person" {
		t.Errorf("expected type name 'Person', got '%s'", typeDecl.Name.Value)
	}

	if len(typeDecl.Fields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(typeDecl.Fields))
	}
}

func TestParseInterfaceDeclaration(t *testing.T) {
	input := `interface Writer
    Write(data string) (int, error)
`

	p, err := New(input, "test.kuki")
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}
	program, errors := p.Parse()

	if len(errors) > 0 {
		t.Fatalf("parser errors: %v", errors)
	}

	if len(program.Declarations) != 1 {
		t.Fatalf("expected 1 declaration, got %d", len(program.Declarations))
	}

	ifaceDecl, ok := program.Declarations[0].(*ast.InterfaceDecl)
	if !ok {
		t.Fatalf("expected InterfaceDecl, got %T", program.Declarations[0])
	}

	if ifaceDecl.Name.Value != "Writer" {
		t.Errorf("expected interface name 'Writer', got '%s'", ifaceDecl.Name.Value)
	}

	if len(ifaceDecl.Methods) != 1 {
		t.Fatalf("expected 1 method, got %d", len(ifaceDecl.Methods))
	}

	method := ifaceDecl.Methods[0]
	if method.Name.Value != "Write" {
		t.Errorf("expected method name 'Write', got '%s'", method.Name.Value)
	}

	if len(method.Parameters) != 1 {
		t.Errorf("expected 1 parameter, got %d", len(method.Parameters))
	}

	if len(method.Returns) != 2 {
		t.Errorf("expected 2 return types, got %d", len(method.Returns))
	}
}

func TestParseMethodDeclaration(t *testing.T) {
	input := `func Display on p Person
    print("Name: {p.Name}")
`

	p, err := New(input, "test.kuki")
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}
	program, errors := p.Parse()

	if len(errors) > 0 {
		t.Fatalf("parser errors: %v", errors)
	}

	if len(program.Declarations) != 1 {
		t.Fatalf("expected 1 declaration, got %d", len(program.Declarations))
	}

	fn, ok := program.Declarations[0].(*ast.FunctionDecl)
	if !ok {
		t.Fatalf("expected FunctionDecl, got %T", program.Declarations[0])
	}

	if fn.Name.Value != "Display" {
		t.Errorf("expected method name 'Display', got '%s'", fn.Name.Value)
	}

	if fn.Receiver == nil {
		t.Fatal("expected receiver, got nil")
	}

	if fn.Receiver.Name.Value != "p" {
		t.Errorf("expected receiver name 'p', got '%s'", fn.Receiver.Name.Value)
	}
}

func TestParseIfStatement(t *testing.T) {
	input := `func Test(x int) string
    if x > 10
        return "big"
    else
        return "small"
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
	if len(fn.Body.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(fn.Body.Statements))
	}

	ifStmt, ok := fn.Body.Statements[0].(*ast.IfStmt)
	if !ok {
		t.Fatalf("expected IfStmt, got %T", fn.Body.Statements[0])
	}

	if ifStmt.Condition == nil {
		t.Error("expected condition, got nil")
	}

	if ifStmt.Consequence == nil {
		t.Error("expected consequence, got nil")
	}

	if ifStmt.Alternative == nil {
		t.Error("expected alternative, got nil")
	}
}

func TestParseSwitchStatement(t *testing.T) {
	input := `func Route(command string) string
    switch command
        when "fetch", "pull"
            return "fetching"
        when "help"
            return "help"
        otherwise
            return "unknown"
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
	switchStmt, ok := fn.Body.Statements[0].(*ast.SwitchStmt)
	if !ok {
		t.Fatalf("expected SwitchStmt, got %T", fn.Body.Statements[0])
	}

	if switchStmt.Expression == nil {
		t.Fatal("expected switch expression, got nil")
	}

	if len(switchStmt.Cases) != 2 {
		t.Fatalf("expected 2 when branches, got %d", len(switchStmt.Cases))
	}

	if len(switchStmt.Cases[0].Values) != 2 {
		t.Fatalf("expected 2 values in first when branch, got %d", len(switchStmt.Cases[0].Values))
	}

	if switchStmt.Otherwise == nil {
		t.Fatal("expected otherwise branch, got nil")
	}
}

func TestParseConditionSwitchStatement(t *testing.T) {
	input := `func Label(stars int) string
    switch
        when stars >= 1000
            return "popular"
        when stars >= 100
            return "growing"
        otherwise
            return "new"
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
	switchStmt, ok := fn.Body.Statements[0].(*ast.SwitchStmt)
	if !ok {
		t.Fatalf("expected SwitchStmt, got %T", fn.Body.Statements[0])
	}

	if switchStmt.Expression != nil {
		t.Fatal("expected condition switch with nil expression")
	}
}

func TestParseWhenAfterOtherwiseIsError(t *testing.T) {
	input := `func Route(command string) string
    switch command
        otherwise
            return "default"
        when "help"
            return "help"
`

	p, err := New(input, "test.kuki")
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}
	_, errors := p.Parse()

	if len(errors) == 0 {
		t.Fatal("expected parser error for 'when' after 'otherwise'")
	}

	found := false
	for _, e := range errors {
		if strings.Contains(e.Error(), "will never execute") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected 'will never execute' error, got: %v", errors)
	}
}

func TestParseForRangeLoop(t *testing.T) {
	input := `func Test(items list of int)
    for item in items
        print(item)
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

	p, err := New(input, "test.kuki")
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}
	program, errors := p.Parse()

	if len(errors) > 0 {
		t.Fatalf("parser errors: %v", errors)
	}

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
	}
}

func TestParsePipeExpression(t *testing.T) {
	input := `func Test() string
    return "hello" |> ToUpper()
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

	p, err := New(input, "test.kuki")
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}
	program, errors := p.Parse()

	if len(errors) > 0 {
		t.Fatalf("parser errors: %v", errors)
	}

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

	p, err := New(input, "test.kuki")
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}
	program, errors := p.Parse()

	if len(errors) > 0 {
		t.Fatalf("parser errors: %v", errors)
	}

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

	p, err := New(input, "test.kuki")
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}
	program, errors := p.Parse()

	if len(errors) > 0 {
		t.Fatalf("parser errors: %v", errors)
	}

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

	p, err := New(input, "test.kuki")
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}
	program, errors := p.Parse()

	if len(errors) > 0 {
		t.Fatalf("parser errors: %v", errors)
	}

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

	p, err := New(input, "test.kuki")
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}
	program, errors := p.Parse()

	if len(errors) > 0 {
		t.Fatalf("parser errors: %v", errors)
	}

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

	p, err := New(input, "test.kuki")
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}
	program, errors := p.Parse()

	if len(errors) > 0 {
		t.Fatalf("parser errors: %v", errors)
	}

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

	p, err := New(input, "test.kuki")
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}
	program, errors := p.Parse()

	if len(errors) > 0 {
		t.Fatalf("parser errors: %v", errors)
	}

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

	methodCall, ok := retStmt.Values[0].(*ast.MethodCallExpr)
	if !ok {
		t.Fatalf("expected MethodCallExpr, got %T", retStmt.Values[0])
	}

	if methodCall.Method.Value != "Length" {
		t.Errorf("expected method name 'Length', got '%s'", methodCall.Method.Value)
	}
}

func TestParseIndexExpression(t *testing.T) {
	input := `func Test(items list of int) int
    return items[0]
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

// Tests for new generic features

func TestParseVariadicParameter(t *testing.T) {
	input := `func Print(many values)
    return values
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
	if len(fn.Parameters) != 1 {
		t.Fatalf("expected 1 parameter, got %d", len(fn.Parameters))
	}

	param := fn.Parameters[0]
	if param.Name.Value != "values" {
		t.Errorf("expected parameter name 'values', got '%s'", param.Name.Value)
	}

	if !param.Variadic {
		t.Error("expected parameter to be variadic")
	}
}

func TestParseTypedVariadicParameter(t *testing.T) {
	input := `func Sum(many numbers int) int
    return 0
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
	param := fn.Parameters[0]

	if !param.Variadic {
		t.Error("expected parameter to be variadic")
	}

	primType, ok := param.Type.(*ast.PrimitiveType)
	if !ok {
		t.Fatalf("expected PrimitiveType, got %T", param.Type)
	}

	if primType.Name != "int" {
		t.Errorf("expected type 'int', got '%s'", primType.Name)
	}
}

// REMOVED: Old generics tests - generics syntax has been removed from Kukicha
// Generic functionality is now provided by the stdlib (written in Go) with special transpilation
