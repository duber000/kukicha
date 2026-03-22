package parser

import (
	"github.com/duber000/kukicha/internal/ast"
	"strings"
	"testing"
)

func TestParseSimpleFunction(t *testing.T) {
	input := `func Add(a int, b int) int
    return a + b
`

	program := mustParseProgram(t, input)

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

	program := mustParseProgram(t, input)

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

func TestParseTypeDeclarationFieldAlias(t *testing.T) {
	input := `type Repo
    Stars int as "stargazers_count"
`

	p, err := New(input, "test.kuki")
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}
	program, errors := p.Parse()
	if len(errors) > 0 {
		t.Fatalf("parser errors: %v", errors)
	}

	typeDecl, ok := program.Declarations[0].(*ast.TypeDecl)
	if !ok {
		t.Fatalf("expected TypeDecl, got %T", program.Declarations[0])
	}
	if len(typeDecl.Fields) != 1 {
		t.Fatalf("expected 1 field, got %d", len(typeDecl.Fields))
	}
	if typeDecl.Fields[0].Tag != `json:"stargazers_count"` {
		t.Fatalf("expected json tag from alias, got %q", typeDecl.Fields[0].Tag)
	}
}

func TestParseTypeDeclarationFieldAliasWithExplicitTagErrors(t *testing.T) {
	input := `type Repo
    Stars int as "stargazers_count" json:"stars"
`

	p, err := New(input, "test.kuki")
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}
	_, errors := p.Parse()
	if len(errors) == 0 {
		t.Fatal("expected parser errors for alias + explicit tag combination")
	}
	if !strings.Contains(errors[0].Error(), "cannot combine field alias and explicit struct tag") {
		t.Fatalf("unexpected parser error: %v", errors[0])
	}
}

func TestParseFunctionTypeAlias(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		typeName   string
		paramCount int
		retCount   int
	}{
		{
			name:       "basic func type",
			input:      "type Handler func(string)\n",
			typeName:   "Handler",
			paramCount: 1,
			retCount:   0,
		},
		{
			name:       "func type with return",
			input:      "type Transform func(string) string\n",
			typeName:   "Transform",
			paramCount: 1,
			retCount:   1,
		},
		{
			name:       "func type with multiple params and returns",
			input:      "type Callback func(string, int) (bool, error)\n",
			typeName:   "Callback",
			paramCount: 2,
			retCount:   2,
		},
		{
			name:       "func type with no params",
			input:      "type Factory func() error\n",
			typeName:   "Factory",
			paramCount: 0,
			retCount:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p, err := New(tt.input, "test.kuki")
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

			if typeDecl.Name.Value != tt.typeName {
				t.Errorf("expected type name %q, got %q", tt.typeName, typeDecl.Name.Value)
			}

			if typeDecl.AliasType == nil {
				t.Fatal("expected AliasType to be non-nil for function type alias")
			}

			funcType, ok := typeDecl.AliasType.(*ast.FunctionType)
			if !ok {
				t.Fatalf("expected FunctionType, got %T", typeDecl.AliasType)
			}

			if len(funcType.Parameters) != tt.paramCount {
				t.Errorf("expected %d parameters, got %d", tt.paramCount, len(funcType.Parameters))
			}

			if len(funcType.Returns) != tt.retCount {
				t.Errorf("expected %d return types, got %d", tt.retCount, len(funcType.Returns))
			}

			if typeDecl.Fields != nil {
				t.Errorf("expected Fields to be nil for type alias, got %v", typeDecl.Fields)
			}
		})
	}
}

func TestParseStructTypeStillWorks(t *testing.T) {
	input := `type Person
    Name string
    Age int
`

	program := mustParseProgram(t, input)

	typeDecl, ok := program.Declarations[0].(*ast.TypeDecl)
	if !ok {
		t.Fatalf("expected TypeDecl, got %T", program.Declarations[0])
	}

	if typeDecl.AliasType != nil {
		t.Error("expected AliasType to be nil for struct type")
	}

	if len(typeDecl.Fields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(typeDecl.Fields))
	}
}

func TestParseInterfaceDeclaration(t *testing.T) {
	input := `interface Writer
    Write(data string) (int, error)
`

	program := mustParseProgram(t, input)

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

	program := mustParseProgram(t, input)

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

	program := mustParseProgram(t, input)

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

	program := mustParseProgram(t, input)

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

	program := mustParseProgram(t, input)

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
