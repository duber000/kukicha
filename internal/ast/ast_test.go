package ast

import (
	"testing"

	"github.com/duber000/kukicha/internal/lexer"
)

// =============================================================================
// Program.TokenLiteral() — conditional dispatch across three sources
// =============================================================================

func TestProgram_TokenLiteral_WithPetiole(t *testing.T) {
	p := &Program{
		PetioleDecl: &PetioleDecl{
			Token: lexer.Token{Lexeme: "petiole"},
			Name:  &Identifier{Value: "main"},
		},
		Imports: []*ImportDecl{
			{Token: lexer.Token{Lexeme: "import"}},
		},
		Declarations: []Declaration{
			&TypeDecl{Token: lexer.Token{Lexeme: "type"}, Name: &Identifier{Value: "T"}},
		},
	}
	if got := p.TokenLiteral(); got != "petiole" {
		t.Errorf("expected 'petiole', got %q", got)
	}
}

func TestProgram_TokenLiteral_NoPetioleWithImports(t *testing.T) {
	p := &Program{
		Imports: []*ImportDecl{
			{Token: lexer.Token{Lexeme: "import"}},
		},
		Declarations: []Declaration{
			&TypeDecl{Token: lexer.Token{Lexeme: "type"}, Name: &Identifier{Value: "T"}},
		},
	}
	if got := p.TokenLiteral(); got != "import" {
		t.Errorf("expected 'import', got %q", got)
	}
}

func TestProgram_TokenLiteral_NoPetioleNoImportsWithDecls(t *testing.T) {
	p := &Program{
		Declarations: []Declaration{
			&FunctionDecl{Token: lexer.Token{Lexeme: "func"}, Name: &Identifier{Value: "main"}},
		},
	}
	if got := p.TokenLiteral(); got != "func" {
		t.Errorf("expected 'func', got %q", got)
	}
}

func TestProgram_TokenLiteral_Empty(t *testing.T) {
	p := &Program{}
	if got := p.TokenLiteral(); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

// =============================================================================
// Program.Pos() — same conditional dispatch, returns Position
// =============================================================================

func TestProgram_Pos_WithPetiole(t *testing.T) {
	p := &Program{
		PetioleDecl: &PetioleDecl{
			Token: lexer.Token{Line: 1, Column: 1, File: "main.kuki"},
		},
	}
	pos := p.Pos()
	if pos.Line != 1 || pos.Column != 1 || pos.File != "main.kuki" {
		t.Errorf("expected {1 1 main.kuki}, got %+v", pos)
	}
}

func TestProgram_Pos_NoPetioleWithImports(t *testing.T) {
	p := &Program{
		Imports: []*ImportDecl{
			{Token: lexer.Token{Line: 3, Column: 1, File: "test.kuki"}},
		},
	}
	pos := p.Pos()
	if pos.Line != 3 {
		t.Errorf("expected line 3, got %d", pos.Line)
	}
}

func TestProgram_Pos_NoPetioleNoImportsWithDecls(t *testing.T) {
	p := &Program{
		Declarations: []Declaration{
			&FunctionDecl{
				Token: lexer.Token{Line: 5, Column: 1, File: "test.kuki"},
				Name:  &Identifier{Value: "main"},
			},
		},
	}
	pos := p.Pos()
	if pos.Line != 5 {
		t.Errorf("expected line 5, got %d", pos.Line)
	}
}

func TestProgram_Pos_Empty(t *testing.T) {
	p := &Program{}
	pos := p.Pos()
	if pos.Line != 0 || pos.Column != 0 || pos.File != "" {
		t.Errorf("expected zero Position, got %+v", pos)
	}
}

// =============================================================================
// Node interface compliance — verify all node types implement Node
// =============================================================================

func TestNodeInterface_Declarations(t *testing.T) {
	nodes := []Node{
		&PetioleDecl{Token: lexer.Token{Lexeme: "petiole", Line: 1, Column: 1}},
		&SkillDecl{Token: lexer.Token{Lexeme: "skill", Line: 2, Column: 1}},
		&ImportDecl{Token: lexer.Token{Lexeme: "import", Line: 3, Column: 1}},
		&TypeDecl{Token: lexer.Token{Lexeme: "type", Line: 4, Column: 1}, Name: &Identifier{Value: "T"}},
		&InterfaceDecl{Token: lexer.Token{Lexeme: "interface", Line: 5, Column: 1}, Name: &Identifier{Value: "I"}},
		&FunctionDecl{Token: lexer.Token{Lexeme: "func", Line: 6, Column: 1}, Name: &Identifier{Value: "F"}},
	}
	for _, n := range nodes {
		if n.TokenLiteral() == "" {
			t.Errorf("node %T returned empty TokenLiteral()", n)
		}
		pos := n.Pos()
		if pos.Line == 0 {
			t.Errorf("node %T returned zero Line in Pos()", n)
		}
	}
}

func TestNodeInterface_Statements(t *testing.T) {
	nodes := []Node{
		&BlockStmt{Token: lexer.Token{Lexeme: "{", Line: 1, Column: 1}},
		&VarDeclStmt{Token: lexer.Token{Lexeme: ":=", Line: 2, Column: 1}},
		&AssignStmt{Token: lexer.Token{Lexeme: "=", Line: 3, Column: 1}},
		&ReturnStmt{Token: lexer.Token{Lexeme: "return", Line: 4, Column: 1}},
		&ContinueStmt{Token: lexer.Token{Lexeme: "continue", Line: 5, Column: 1}},
		&BreakStmt{Token: lexer.Token{Lexeme: "break", Line: 6, Column: 1}},
		&IfStmt{Token: lexer.Token{Lexeme: "if", Line: 7, Column: 1}},
		&ElseStmt{Token: lexer.Token{Lexeme: "else", Line: 8, Column: 1}},
		&SwitchStmt{Token: lexer.Token{Lexeme: "switch", Line: 9, Column: 1}},
		&SelectStmt{Token: lexer.Token{Lexeme: "select", Line: 10, Column: 1}},
		&TypeSwitchStmt{Token: lexer.Token{Lexeme: "switch", Line: 11, Column: 1}},
		&ForRangeStmt{Token: lexer.Token{Lexeme: "for", Line: 12, Column: 1}},
		&ForNumericStmt{Token: lexer.Token{Lexeme: "for", Line: 13, Column: 1}},
		&ForConditionStmt{Token: lexer.Token{Lexeme: "for", Line: 14, Column: 1}},
		&DeferStmt{Token: lexer.Token{Lexeme: "defer", Line: 15, Column: 1}},
		&GoStmt{Token: lexer.Token{Lexeme: "go", Line: 16, Column: 1}},
		&SendStmt{Token: lexer.Token{Lexeme: "send", Line: 17, Column: 1}},
		&IncDecStmt{Token: lexer.Token{Lexeme: "++", Line: 18, Column: 1}},
	}
	for _, n := range nodes {
		if n.TokenLiteral() == "" {
			t.Errorf("node %T returned empty TokenLiteral()", n)
		}
		pos := n.Pos()
		if pos.Line == 0 {
			t.Errorf("node %T returned zero Line in Pos()", n)
		}
	}
}

func TestNodeInterface_Expressions(t *testing.T) {
	nodes := []Node{
		&Identifier{Token: lexer.Token{Lexeme: "x", Line: 1, Column: 1}, Value: "x"},
		&IntegerLiteral{Token: lexer.Token{Lexeme: "42", Line: 2, Column: 1}},
		&FloatLiteral{Token: lexer.Token{Lexeme: "3.14", Line: 3, Column: 1}},
		&RuneLiteral{Token: lexer.Token{Lexeme: "'a'", Line: 4, Column: 1}},
		&StringLiteral{Token: lexer.Token{Lexeme: `"hello"`, Line: 5, Column: 1}},
		&BooleanLiteral{Token: lexer.Token{Lexeme: "true", Line: 6, Column: 1}},
		&BinaryExpr{Token: lexer.Token{Lexeme: "+", Line: 7, Column: 1}},
		&UnaryExpr{Token: lexer.Token{Lexeme: "-", Line: 8, Column: 1}},
		&PipeExpr{Token: lexer.Token{Lexeme: "|>", Line: 9, Column: 1}},
		&CallExpr{Token: lexer.Token{Lexeme: "(", Line: 10, Column: 1}},
		&MethodCallExpr{Token: lexer.Token{Lexeme: ".", Line: 11, Column: 1}},
		&IndexExpr{Token: lexer.Token{Lexeme: "[", Line: 12, Column: 1}},
		&SliceExpr{Token: lexer.Token{Lexeme: "[", Line: 13, Column: 1}},
		&EmptyExpr{Token: lexer.Token{Lexeme: "empty", Line: 14, Column: 1}},
		&DiscardExpr{Token: lexer.Token{Lexeme: "discard", Line: 15, Column: 1}},
		&ErrorExpr{Token: lexer.Token{Lexeme: "error", Line: 16, Column: 1}},
		&MakeExpr{Token: lexer.Token{Lexeme: "make", Line: 17, Column: 1}},
		&CloseExpr{Token: lexer.Token{Lexeme: "close", Line: 18, Column: 1}},
		&PanicExpr{Token: lexer.Token{Lexeme: "panic", Line: 19, Column: 1}},
		&RecoverExpr{Token: lexer.Token{Lexeme: "recover", Line: 20, Column: 1}},
		&ReceiveExpr{Token: lexer.Token{Lexeme: "receive", Line: 21, Column: 1}},
		&TypeCastExpr{Token: lexer.Token{Lexeme: "as", Line: 22, Column: 1}},
		&TypeAssertionExpr{Token: lexer.Token{Lexeme: ".", Line: 23, Column: 1}},
		&AddressOfExpr{Token: lexer.Token{Lexeme: "reference", Line: 24, Column: 1}},
		&DerefExpr{Token: lexer.Token{Lexeme: "dereference", Line: 25, Column: 1}},
		&ArrowLambda{Token: lexer.Token{Lexeme: "=>", Line: 26, Column: 1}},
		&ReturnExpr{Token: lexer.Token{Lexeme: "return", Line: 27, Column: 1}},
		&FunctionLiteral{Token: lexer.Token{Lexeme: "func", Line: 28, Column: 1}},
		&NamedArgument{Token: lexer.Token{Lexeme: "name", Line: 29, Column: 1}},
		&StructLiteralExpr{Token: lexer.Token{Lexeme: "User", Line: 30, Column: 1}},
		&ListLiteralExpr{Token: lexer.Token{Lexeme: "[", Line: 31, Column: 1}},
		&MapLiteralExpr{Token: lexer.Token{Lexeme: "{", Line: 32, Column: 1}},
		&BlockExpr{Token: lexer.Token{Lexeme: "INDENT", Line: 33, Column: 1}},
	}
	for _, n := range nodes {
		if n.TokenLiteral() == "" {
			t.Errorf("node %T returned empty TokenLiteral()", n)
		}
		pos := n.Pos()
		if pos.Line == 0 {
			t.Errorf("node %T returned zero Line in Pos()", n)
		}
	}
}

func TestNodeInterface_TypeAnnotations(t *testing.T) {
	nodes := []Node{
		&PrimitiveType{Token: lexer.Token{Lexeme: "int", Line: 1, Column: 1}},
		&NamedType{Token: lexer.Token{Lexeme: "User", Line: 2, Column: 1}},
		&ReferenceType{Token: lexer.Token{Lexeme: "reference", Line: 3, Column: 1}},
		&ListType{Token: lexer.Token{Lexeme: "list", Line: 4, Column: 1}},
		&MapType{Token: lexer.Token{Lexeme: "map", Line: 5, Column: 1}},
		&ChannelType{Token: lexer.Token{Lexeme: "channel", Line: 6, Column: 1}},
		&FunctionType{Token: lexer.Token{Lexeme: "func", Line: 7, Column: 1}},
	}
	for _, n := range nodes {
		if n.TokenLiteral() == "" {
			t.Errorf("node %T returned empty TokenLiteral()", n)
		}
		pos := n.Pos()
		if pos.Line == 0 {
			t.Errorf("node %T returned zero Line in Pos()", n)
		}
	}
}

// =============================================================================
// ExpressionStmt delegates to its Expression
// =============================================================================

func TestExpressionStmt_DelegatesToExpression(t *testing.T) {
	expr := &Identifier{
		Token: lexer.Token{Lexeme: "foo", Line: 10, Column: 5, File: "test.kuki"},
		Value: "foo",
	}
	stmt := &ExpressionStmt{Expression: expr}

	if got := stmt.TokenLiteral(); got != "foo" {
		t.Errorf("expected 'foo', got %q", got)
	}
	pos := stmt.Pos()
	if pos.Line != 10 || pos.Column != 5 {
		t.Errorf("expected {10 5}, got {%d %d}", pos.Line, pos.Column)
	}
}

// =============================================================================
// VarDeclStmt implements both Statement and Declaration
// =============================================================================

func TestVarDeclStmt_ImplementsBothInterfaces(t *testing.T) {
	stmt := &VarDeclStmt{
		Token: lexer.Token{Lexeme: ":=", Line: 1, Column: 1},
		Names: []*Identifier{{Value: "x"}},
	}

	// Verify it satisfies Statement
	var _ Statement = stmt

	// Verify it satisfies Declaration
	var _ Declaration = stmt
}
