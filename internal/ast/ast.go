package ast

import "github.com/duber000/kukicha/internal/lexer"

// Node is the base interface for all AST nodes
type Node interface {
	TokenLiteral() string
	Pos() Position
}

// Position represents a location in the source code
type Position struct {
	Line   int
	Column int
	File   string
}

// ============================================================================
// Program - Root node
// ============================================================================

type Program struct {
	PetioleDecl  *PetioleDecl  // Optional petiole declaration
	Imports      []*ImportDecl // Import declarations
	Declarations []Declaration // Top-level declarations (types, interfaces, functions)
}

func (p *Program) TokenLiteral() string {
	if p.PetioleDecl != nil {
		return p.PetioleDecl.TokenLiteral()
	}
	if len(p.Imports) > 0 {
		return p.Imports[0].TokenLiteral()
	}
	if len(p.Declarations) > 0 {
		return p.Declarations[0].TokenLiteral()
	}
	return ""
}

func (p *Program) Pos() Position {
	if p.PetioleDecl != nil {
		return p.PetioleDecl.Pos()
	}
	if len(p.Imports) > 0 {
		return p.Imports[0].Pos()
	}
	if len(p.Declarations) > 0 {
		return p.Declarations[0].Pos()
	}
	return Position{}
}

// ============================================================================
// Declarations
// ============================================================================

type Declaration interface {
	Node
	declNode()
}

type PetioleDecl struct {
	Token lexer.Token // The 'petiole' token
	Name  *Identifier
}

func (d *PetioleDecl) TokenLiteral() string {
	return d.Token.Lexeme
}
func (d *PetioleDecl) Pos() Position {
	return Position{Line: d.Token.Line, Column: d.Token.Column, File: d.Token.File}
}
func (d *PetioleDecl) declNode() {}

type ImportDecl struct {
	Token lexer.Token // The 'import' token
	Path  *StringLiteral
	Alias *Identifier // Optional alias
}

func (d *ImportDecl) TokenLiteral() string { return d.Token.Lexeme }
func (d *ImportDecl) Pos() Position {
	return Position{Line: d.Token.Line, Column: d.Token.Column, File: d.Token.File}
}
func (d *ImportDecl) declNode() {}

type TypeDecl struct {
	Token  lexer.Token // The 'type' token
	Name   *Identifier
	Fields []*FieldDecl
}

func (d *TypeDecl) TokenLiteral() string { return d.Token.Lexeme }
func (d *TypeDecl) Pos() Position {
	return Position{Line: d.Token.Line, Column: d.Token.Column, File: d.Token.File}
}
func (d *TypeDecl) declNode() {}

type FieldDecl struct {
	Name *Identifier
	Type TypeAnnotation
	Tag  string // Struct tag (e.g., `json:"name"`)
}

type InterfaceDecl struct {
	Token   lexer.Token // The 'interface' token
	Name    *Identifier
	Methods []*MethodSignature
}

func (d *InterfaceDecl) TokenLiteral() string { return d.Token.Lexeme }
func (d *InterfaceDecl) Pos() Position {
	return Position{Line: d.Token.Line, Column: d.Token.Column, File: d.Token.File}
}
func (d *InterfaceDecl) declNode() {}

type MethodSignature struct {
	Name       *Identifier
	Parameters []*Parameter
	Returns    []TypeAnnotation
}

type FunctionDecl struct {
	Token      lexer.Token // The 'func' token
	Name       *Identifier
	Parameters []*Parameter
	Returns    []TypeAnnotation
	Body       *BlockStmt
	Receiver   *Receiver // For methods (optional)
}

func (d *FunctionDecl) TokenLiteral() string { return d.Token.Lexeme }
func (d *FunctionDecl) Pos() Position {
	return Position{Line: d.Token.Line, Column: d.Token.Column, File: d.Token.File}
}
func (d *FunctionDecl) declNode() {}

type Parameter struct {
	Name         *Identifier
	Type         TypeAnnotation
	Variadic     bool       // true if "many" keyword used
	DefaultValue Expression // Optional default value (e.g., count int = 10)
}

type Receiver struct {
	Name *Identifier // The receiver variable name
	Type TypeAnnotation
}

// ============================================================================
// Type Annotations
// ============================================================================

type TypeAnnotation interface {
	Node
	typeNode()
}

type PrimitiveType struct {
	Token lexer.Token // The type token
	Name  string      // int, float, string, bool, etc.
}

func (t *PrimitiveType) TokenLiteral() string { return t.Token.Lexeme }
func (t *PrimitiveType) Pos() Position {
	return Position{Line: t.Token.Line, Column: t.Token.Column, File: t.Token.File}
}
func (t *PrimitiveType) typeNode() {}

type NamedType struct {
	Token lexer.Token // The identifier token
	Name  string
}

func (t *NamedType) TokenLiteral() string { return t.Token.Lexeme }
func (t *NamedType) Pos() Position {
	return Position{Line: t.Token.Line, Column: t.Token.Column, File: t.Token.File}
}
func (t *NamedType) typeNode() {}

type ReferenceType struct {
	Token       lexer.Token // The 'reference' token
	ElementType TypeAnnotation
}

func (t *ReferenceType) TokenLiteral() string { return t.Token.Lexeme }
func (t *ReferenceType) Pos() Position {
	return Position{Line: t.Token.Line, Column: t.Token.Column, File: t.Token.File}
}
func (t *ReferenceType) typeNode() {}

type ListType struct {
	Token       lexer.Token // The 'list' token
	ElementType TypeAnnotation
}

func (t *ListType) TokenLiteral() string { return t.Token.Lexeme }
func (t *ListType) Pos() Position {
	return Position{Line: t.Token.Line, Column: t.Token.Column, File: t.Token.File}
}
func (t *ListType) typeNode() {}

type MapType struct {
	Token     lexer.Token // The 'map' token
	KeyType   TypeAnnotation
	ValueType TypeAnnotation
}

func (t *MapType) TokenLiteral() string { return t.Token.Lexeme }
func (t *MapType) Pos() Position {
	return Position{Line: t.Token.Line, Column: t.Token.Column, File: t.Token.File}
}
func (t *MapType) typeNode() {}

type ChannelType struct {
	Token       lexer.Token // The 'channel' token
	ElementType TypeAnnotation
}

func (t *ChannelType) TokenLiteral() string { return t.Token.Lexeme }
func (t *ChannelType) Pos() Position {
	return Position{Line: t.Token.Line, Column: t.Token.Column, File: t.Token.File}
}
func (t *ChannelType) typeNode() {}

// FunctionType represents a function type annotation
// e.g., func(int, string) bool
type FunctionType struct {
	Token      lexer.Token      // The 'func' token
	Parameters []TypeAnnotation // Parameter types
	Returns    []TypeAnnotation // Return types
}

func (t *FunctionType) TokenLiteral() string { return t.Token.Lexeme }
func (t *FunctionType) Pos() Position {
	return Position{Line: t.Token.Line, Column: t.Token.Column, File: t.Token.File}
}
func (t *FunctionType) typeNode() {}

// ============================================================================
// OnErr Clause (attached to statement nodes, not a standalone node)
// ============================================================================

// OnErrClause represents the error handling part of an onerr statement.
// It is not an AST node itself — it is a field on VarDeclStmt, AssignStmt, and ExpressionStmt.
type OnErrClause struct {
	Token   lexer.Token // The 'onerr' token
	Handler Expression  // Error handler (panic, error, empty, discard, or default value)
}

// ============================================================================
// Statements
// ============================================================================

type Statement interface {
	Node
	stmtNode()
}

type BlockStmt struct {
	Token      lexer.Token // The '{' or INDENT token
	Statements []Statement
}

func (s *BlockStmt) TokenLiteral() string { return s.Token.Lexeme }
func (s *BlockStmt) Pos() Position {
	return Position{Line: s.Token.Line, Column: s.Token.Column, File: s.Token.File}
}
func (s *BlockStmt) stmtNode() {}

type VarDeclStmt struct {
	Names  []*Identifier
	Type   TypeAnnotation // Optional (can be nil for inference)
	Values []Expression   // Right-hand side values (can be single or multiple)
	Token  lexer.Token    // The identifier token or walrus token
	OnErr  *OnErrClause   // Optional onerr clause (e.g., x := f() onerr panic "msg")
}

func (s *VarDeclStmt) TokenLiteral() string { return s.Token.Lexeme }
func (s *VarDeclStmt) Pos() Position {
	return Position{Line: s.Token.Line, Column: s.Token.Column, File: s.Token.File}
}
func (s *VarDeclStmt) stmtNode() {}
func (s *VarDeclStmt) declNode() {}

type AssignStmt struct {
	Targets []Expression // Can be single or multiple targets
	Values  []Expression // Right-hand side values (can be single or multiple)
	Token   lexer.Token  // The '=' token
	OnErr   *OnErrClause // Optional onerr clause (e.g., x = f() onerr panic "msg")
}

func (s *AssignStmt) TokenLiteral() string { return s.Token.Lexeme }
func (s *AssignStmt) Pos() Position {
	return Position{Line: s.Token.Line, Column: s.Token.Column, File: s.Token.File}
}
func (s *AssignStmt) stmtNode() {}

type IncDecStmt struct {
	Token    lexer.Token // The '++' or '--' token
	Variable Expression
	Operator string // "++" or "--"
}

func (s *IncDecStmt) TokenLiteral() string { return s.Token.Lexeme }
func (s *IncDecStmt) Pos() Position {
	return Position{Line: s.Token.Line, Column: s.Token.Column, File: s.Token.File}
}
func (s *IncDecStmt) stmtNode() {}

type ReturnStmt struct {
	Token  lexer.Token // The 'return' token
	Values []Expression
}

func (s *ReturnStmt) TokenLiteral() string { return s.Token.Lexeme }
func (s *ReturnStmt) Pos() Position {
	return Position{Line: s.Token.Line, Column: s.Token.Column, File: s.Token.File}
}
func (s *ReturnStmt) stmtNode() {}

type ContinueStmt struct {
	Token lexer.Token // The 'continue' token
}

func (s *ContinueStmt) TokenLiteral() string { return s.Token.Lexeme }
func (s *ContinueStmt) Pos() Position {
	return Position{Line: s.Token.Line, Column: s.Token.Column, File: s.Token.File}
}
func (s *ContinueStmt) stmtNode() {}

type BreakStmt struct {
	Token lexer.Token // The 'break' token
}

func (s *BreakStmt) TokenLiteral() string { return s.Token.Lexeme }
func (s *BreakStmt) Pos() Position {
	return Position{Line: s.Token.Line, Column: s.Token.Column, File: s.Token.File}
}
func (s *BreakStmt) stmtNode() {}

type IfStmt struct {
	Token       lexer.Token // The 'if' token
	Init        Statement   // Optional initialization statement (can be nil)
	Condition   Expression
	Consequence *BlockStmt
	Alternative Statement // Can be ElseStmt or another IfStmt (else if)
}

func (s *IfStmt) TokenLiteral() string { return s.Token.Lexeme }
func (s *IfStmt) Pos() Position {
	return Position{Line: s.Token.Line, Column: s.Token.Column, File: s.Token.File}
}
func (s *IfStmt) stmtNode() {}

type ElseStmt struct {
	Token lexer.Token // The 'else' token
	Body  *BlockStmt
}

func (s *ElseStmt) TokenLiteral() string { return s.Token.Lexeme }
func (s *ElseStmt) Pos() Position {
	return Position{Line: s.Token.Line, Column: s.Token.Column, File: s.Token.File}
}
func (s *ElseStmt) stmtNode() {}

type SwitchStmt struct {
	Token      lexer.Token // The 'switch' token
	Expression Expression  // Optional (nil for condition switch)
	Cases      []*WhenCase
	Otherwise  *OtherwiseCase // Optional
}

func (s *SwitchStmt) TokenLiteral() string { return s.Token.Lexeme }
func (s *SwitchStmt) Pos() Position {
	return Position{Line: s.Token.Line, Column: s.Token.Column, File: s.Token.File}
}
func (s *SwitchStmt) stmtNode() {}

type WhenCase struct {
	Token  lexer.Token // The 'when' or 'case' token
	Values []Expression
	Body   *BlockStmt
}

type OtherwiseCase struct {
	Token lexer.Token // The 'otherwise' or 'default' token
	Body  *BlockStmt
}

// ForRangeStmt: for item in collection
type ForRangeStmt struct {
	Token      lexer.Token // The 'for' token
	Variable   *Identifier
	Index      *Identifier // Optional (for index, item in collection)
	Collection Expression
	Body       *BlockStmt
}

func (s *ForRangeStmt) TokenLiteral() string { return s.Token.Lexeme }
func (s *ForRangeStmt) Pos() Position {
	return Position{Line: s.Token.Line, Column: s.Token.Column, File: s.Token.File}
}
func (s *ForRangeStmt) stmtNode() {}

// ForNumericStmt: for i from start to end / for i from start through end
type ForNumericStmt struct {
	Token    lexer.Token // The 'for' token
	Variable *Identifier
	Start    Expression
	End      Expression
	Through  bool // true for 'through', false for 'to'
	Body     *BlockStmt
}

func (s *ForNumericStmt) TokenLiteral() string { return s.Token.Lexeme }
func (s *ForNumericStmt) Pos() Position {
	return Position{Line: s.Token.Line, Column: s.Token.Column, File: s.Token.File}
}
func (s *ForNumericStmt) stmtNode() {}

// ForConditionStmt: for condition
type ForConditionStmt struct {
	Token     lexer.Token // The 'for' token
	Condition Expression
	Body      *BlockStmt
}

func (s *ForConditionStmt) TokenLiteral() string { return s.Token.Lexeme }
func (s *ForConditionStmt) Pos() Position {
	return Position{Line: s.Token.Line, Column: s.Token.Column, File: s.Token.File}
}
func (s *ForConditionStmt) stmtNode() {}

type DeferStmt struct {
	Token lexer.Token // The 'defer' token
	Call  Expression  // Can be CallExpr or MethodCallExpr
}

func (s *DeferStmt) TokenLiteral() string { return s.Token.Lexeme }
func (s *DeferStmt) Pos() Position {
	return Position{Line: s.Token.Line, Column: s.Token.Column, File: s.Token.File}
}
func (s *DeferStmt) stmtNode() {}

type GoStmt struct {
	Token lexer.Token // The 'go' token
	Call  Expression  // Can be CallExpr or MethodCallExpr
}

func (s *GoStmt) TokenLiteral() string { return s.Token.Lexeme }
func (s *GoStmt) Pos() Position {
	return Position{Line: s.Token.Line, Column: s.Token.Column, File: s.Token.File}
}
func (s *GoStmt) stmtNode() {}

type SendStmt struct {
	Token   lexer.Token // The 'send' token
	Value   Expression
	Channel Expression
}

func (s *SendStmt) TokenLiteral() string { return s.Token.Lexeme }
func (s *SendStmt) Pos() Position {
	return Position{Line: s.Token.Line, Column: s.Token.Column, File: s.Token.File}
}
func (s *SendStmt) stmtNode() {}

type ExpressionStmt struct {
	Expression Expression
	OnErr      *OnErrClause // Optional onerr clause (e.g., f() onerr panic "msg")
}

func (s *ExpressionStmt) TokenLiteral() string { return s.Expression.TokenLiteral() }
func (s *ExpressionStmt) Pos() Position        { return s.Expression.Pos() }
func (s *ExpressionStmt) stmtNode()            {}

// ============================================================================
// Expressions
// ============================================================================

type Expression interface {
	Node
	exprNode()
}

type Identifier struct {
	Token lexer.Token
	Value string
}

func (e *Identifier) TokenLiteral() string { return e.Token.Lexeme }
func (e *Identifier) Pos() Position {
	return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File}
}
func (e *Identifier) exprNode() {}

type IntegerLiteral struct {
	Token lexer.Token
	Value int64
}

func (e *IntegerLiteral) TokenLiteral() string { return e.Token.Lexeme }
func (e *IntegerLiteral) Pos() Position {
	return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File}
}
func (e *IntegerLiteral) exprNode() {}

type FloatLiteral struct {
	Token lexer.Token
	Value float64
}

func (e *FloatLiteral) TokenLiteral() string { return e.Token.Lexeme }
func (e *FloatLiteral) Pos() Position {
	return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File}
}
func (e *FloatLiteral) exprNode() {}

type RuneLiteral struct {
	Token lexer.Token
	Value rune
}

func (e *RuneLiteral) TokenLiteral() string { return e.Token.Lexeme }
func (e *RuneLiteral) Pos() Position {
	return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File}
}
func (e *RuneLiteral) exprNode() {}

type StringLiteral struct {
	Token        lexer.Token
	Value        string
	Interpolated bool                   // True if contains {expr}
	Parts        []*StringInterpolation // For interpolated strings
}

func (e *StringLiteral) TokenLiteral() string { return e.Token.Lexeme }
func (e *StringLiteral) Pos() Position {
	return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File}
}
func (e *StringLiteral) exprNode() {}

type StringInterpolation struct {
	IsLiteral bool       // True for literal parts, false for expressions
	Literal   string     // For literal parts
	Expr      Expression // For expression parts
}

type BooleanLiteral struct {
	Token lexer.Token
	Value bool
}

func (e *BooleanLiteral) TokenLiteral() string { return e.Token.Lexeme }
func (e *BooleanLiteral) Pos() Position {
	return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File}
}
func (e *BooleanLiteral) exprNode() {}

type BinaryExpr struct {
	Token    lexer.Token // The operator token
	Left     Expression
	Operator string
	Right    Expression
}

func (e *BinaryExpr) TokenLiteral() string { return e.Token.Lexeme }
func (e *BinaryExpr) Pos() Position {
	return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File}
}
func (e *BinaryExpr) exprNode() {}

type UnaryExpr struct {
	Token    lexer.Token // The operator token
	Operator string
	Right    Expression
}

func (e *UnaryExpr) TokenLiteral() string { return e.Token.Lexeme }
func (e *UnaryExpr) Pos() Position {
	return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File}
}
func (e *UnaryExpr) exprNode() {}

type PipeExpr struct {
	Token lexer.Token // The '|>' token
	Left  Expression
	Right Expression // Must be a function call
}

func (e *PipeExpr) TokenLiteral() string { return e.Token.Lexeme }
func (e *PipeExpr) Pos() Position {
	return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File}
}
func (e *PipeExpr) exprNode() {}

// NamedArgument represents a named argument in a function call
// e.g., foo(name: "value", count: 5)
type NamedArgument struct {
	Token lexer.Token // The identifier token for the name
	Name  *Identifier
	Value Expression
}

func (e *NamedArgument) TokenLiteral() string { return e.Token.Lexeme }
func (e *NamedArgument) Pos() Position {
	return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File}
}
func (e *NamedArgument) exprNode() {}

type CallExpr struct {
	Token          lexer.Token // The '(' token or identifier
	Function       Expression
	Arguments      []Expression     // Positional arguments
	NamedArguments []*NamedArgument // Named arguments (e.g., name: value)
	Variadic       bool             // true if 'many' used: f(many args)
}

func (e *CallExpr) TokenLiteral() string { return e.Token.Lexeme }
func (e *CallExpr) Pos() Position {
	return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File}
}
func (e *CallExpr) exprNode() {}

type MethodCallExpr struct {
	Token          lexer.Token // The '.' token
	Object         Expression  // Can be nil for shorthand pipes: |> .Method()
	Method         *Identifier
	Arguments      []Expression     // Positional arguments
	NamedArguments []*NamedArgument // Named arguments (e.g., name: value)
	Variadic       bool             // true if 'many' used: obj.f(many args)
	IsCall         bool             // true if explicit parentheses were used: obj.method() vs obj.field
}

func (e *MethodCallExpr) TokenLiteral() string { return e.Token.Lexeme }
func (e *MethodCallExpr) Pos() Position {
	return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File}
}
func (e *MethodCallExpr) exprNode() {}

type IndexExpr struct {
	Token lexer.Token // The '[' token
	Left  Expression
	Index Expression
}

func (e *IndexExpr) TokenLiteral() string { return e.Token.Lexeme }
func (e *IndexExpr) Pos() Position {
	return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File}
}
func (e *IndexExpr) exprNode() {}

type SliceExpr struct {
	Token lexer.Token // The '[' token
	Left  Expression
	Start Expression // Can be nil
	End   Expression // Can be nil
}

func (e *SliceExpr) TokenLiteral() string { return e.Token.Lexeme }
func (e *SliceExpr) Pos() Position {
	return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File}
}
func (e *SliceExpr) exprNode() {}

type StructLiteralExpr struct {
	Token  lexer.Token // The type identifier
	Type   TypeAnnotation
	Fields []*FieldValue
}

func (e *StructLiteralExpr) TokenLiteral() string { return e.Token.Lexeme }
func (e *StructLiteralExpr) Pos() Position {
	return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File}
}
func (e *StructLiteralExpr) exprNode() {}

type FieldValue struct {
	Name  *Identifier
	Value Expression
}

type ListLiteralExpr struct {
	Token    lexer.Token // The '[' token or 'list' keyword
	Type     TypeAnnotation
	Elements []Expression
}

func (e *ListLiteralExpr) TokenLiteral() string { return e.Token.Lexeme }
func (e *ListLiteralExpr) Pos() Position {
	return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File}
}
func (e *ListLiteralExpr) exprNode() {}

type MapLiteralExpr struct {
	Token   lexer.Token // The '{' token or 'map' keyword
	KeyType TypeAnnotation
	ValType TypeAnnotation
	Pairs   []*KeyValuePair
}

func (e *MapLiteralExpr) TokenLiteral() string { return e.Token.Lexeme }
func (e *MapLiteralExpr) Pos() Position {
	return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File}
}
func (e *MapLiteralExpr) exprNode() {}

type KeyValuePair struct {
	Key   Expression
	Value Expression
}

type ReceiveExpr struct {
	Token   lexer.Token // The 'receive' token
	Channel Expression
}

func (e *ReceiveExpr) TokenLiteral() string { return e.Token.Lexeme }
func (e *ReceiveExpr) Pos() Position {
	return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File}
}
func (e *ReceiveExpr) exprNode() {}

type TypeCastExpr struct {
	Token      lexer.Token // The 'as' token
	Expression Expression
	TargetType TypeAnnotation
}

func (e *TypeCastExpr) TokenLiteral() string { return e.Token.Lexeme }
func (e *TypeCastExpr) Pos() Position {
	return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File}
}
func (e *TypeCastExpr) exprNode() {}

type TypeAssertionExpr struct {
	Token      lexer.Token // The '.' token
	Expression Expression
	TargetType TypeAnnotation
}

func (e *TypeAssertionExpr) TokenLiteral() string { return e.Token.Lexeme }
func (e *TypeAssertionExpr) Pos() Position {
	return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File}
}
func (e *TypeAssertionExpr) exprNode() {}

type EmptyExpr struct {
	Token lexer.Token // The 'empty' token
	Type  TypeAnnotation
}

func (e *EmptyExpr) TokenLiteral() string { return e.Token.Lexeme }
func (e *EmptyExpr) Pos() Position {
	return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File}
}
func (e *EmptyExpr) exprNode() {}

type DiscardExpr struct {
	Token lexer.Token // The 'discard' token
}

func (e *DiscardExpr) TokenLiteral() string { return e.Token.Lexeme }
func (e *DiscardExpr) Pos() Position {
	return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File}
}
func (e *DiscardExpr) exprNode() {}

type ErrorExpr struct {
	Token   lexer.Token // The 'error' token
	Message Expression  // Usually a string literal
}

func (e *ErrorExpr) TokenLiteral() string { return e.Token.Lexeme }
func (e *ErrorExpr) Pos() Position {
	return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File}
}
func (e *ErrorExpr) exprNode() {}

type ReturnExpr struct {
	Token  lexer.Token // The 'return' token
	Values []Expression
}

func (e *ReturnExpr) TokenLiteral() string { return e.Token.Lexeme }
func (e *ReturnExpr) Pos() Position {
	return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File}
}
func (e *ReturnExpr) exprNode() {}

type MakeExpr struct {
	Token lexer.Token // The 'make' token
	Type  TypeAnnotation
	Args  []Expression // Size/capacity for slices, channels
}

func (e *MakeExpr) TokenLiteral() string { return e.Token.Lexeme }
func (e *MakeExpr) Pos() Position {
	return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File}
}
func (e *MakeExpr) exprNode() {}

type CloseExpr struct {
	Token   lexer.Token // The 'close' token
	Channel Expression
}

func (e *CloseExpr) TokenLiteral() string { return e.Token.Lexeme }
func (e *CloseExpr) Pos() Position {
	return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File}
}
func (e *CloseExpr) exprNode() {}

type PanicExpr struct {
	Token   lexer.Token // The 'panic' token
	Message Expression
}

func (e *PanicExpr) TokenLiteral() string { return e.Token.Lexeme }
func (e *PanicExpr) Pos() Position {
	return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File}
}
func (e *PanicExpr) exprNode() {}

type RecoverExpr struct {
	Token lexer.Token // The 'recover' token
}

func (e *RecoverExpr) TokenLiteral() string { return e.Token.Lexeme }
func (e *RecoverExpr) Pos() Position {
	return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File}
}
func (e *RecoverExpr) exprNode() {}

type FunctionLiteral struct {
	Token      lexer.Token // The 'func' token
	Parameters []*Parameter
	Returns    []TypeAnnotation
	Body       *BlockStmt
}

func (e *FunctionLiteral) TokenLiteral() string { return e.Token.Lexeme }
func (e *FunctionLiteral) Pos() Position {
	return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File}
}
func (e *FunctionLiteral) exprNode() {}

type AddressOfExpr struct {
	Token   lexer.Token // The 'reference' token
	Operand Expression  // The expression to take address of
}

func (e *AddressOfExpr) TokenLiteral() string { return e.Token.Lexeme }
func (e *AddressOfExpr) Pos() Position {
	return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File}
}
func (e *AddressOfExpr) exprNode() {}

type DerefExpr struct {
	Token   lexer.Token // The 'dereference' token
	Operand Expression  // The expression to dereference
}

func (e *DerefExpr) TokenLiteral() string { return e.Token.Lexeme }
func (e *DerefExpr) Pos() Position {
	return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File}
}
func (e *DerefExpr) exprNode() {}

type BlockExpr struct {
	Token lexer.Token // The INDENT token
	Body  *BlockStmt
}

func (e *BlockExpr) TokenLiteral() string { return e.Token.Lexeme }
func (e *BlockExpr) Pos() Position {
	return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File}
}
func (e *BlockExpr) exprNode() {}
