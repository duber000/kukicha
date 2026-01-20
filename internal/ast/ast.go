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
	LeafDecl     *LeafDecl      // Optional leaf declaration
	Imports      []*ImportDecl  // Import declarations
	Declarations []Declaration  // Top-level declarations (types, interfaces, functions)
}

func (p *Program) TokenLiteral() string {
	if p.LeafDecl != nil {
		return p.LeafDecl.TokenLiteral()
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
	if p.LeafDecl != nil {
		return p.LeafDecl.Pos()
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

type LeafDecl struct {
	Token lexer.Token // The 'leaf' token
	Name  *Identifier
}

func (d *LeafDecl) TokenLiteral() string {
	return d.Token.Lexeme
}
func (d *LeafDecl) Pos() Position {
	return Position{Line: d.Token.Line, Column: d.Token.Column, File: d.Token.File}
}
func (d *LeafDecl) declNode()            {}

type ImportDecl struct {
	Token lexer.Token // The 'import' token
	Path  *StringLiteral
	Alias *Identifier // Optional alias
}

func (d *ImportDecl) TokenLiteral() string { return d.Token.Lexeme }
func (d *ImportDecl) Pos() Position        { return Position{Line: d.Token.Line, Column: d.Token.Column, File: d.Token.File} }
func (d *ImportDecl) declNode()            {}

type TypeDecl struct {
	Token  lexer.Token // The 'type' token
	Name   *Identifier
	Fields []*FieldDecl
}

func (d *TypeDecl) TokenLiteral() string { return d.Token.Lexeme }
func (d *TypeDecl) Pos() Position        { return Position{Line: d.Token.Line, Column: d.Token.Column, File: d.Token.File} }
func (d *TypeDecl) declNode()            {}

type FieldDecl struct {
	Name *Identifier
	Type TypeAnnotation
}

type InterfaceDecl struct {
	Token   lexer.Token // The 'interface' token
	Name    *Identifier
	Methods []*MethodSignature
}

func (d *InterfaceDecl) TokenLiteral() string { return d.Token.Lexeme }
func (d *InterfaceDecl) Pos() Position        { return Position{Line: d.Token.Line, Column: d.Token.Column, File: d.Token.File} }
func (d *InterfaceDecl) declNode()            {}

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
func (d *FunctionDecl) Pos() Position        { return Position{Line: d.Token.Line, Column: d.Token.Column, File: d.Token.File} }
func (d *FunctionDecl) declNode()            {}

type Parameter struct {
	Name *Identifier
	Type TypeAnnotation
}

type Receiver struct {
	Name *Identifier // Can be 'this' or a variable name
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
func (t *PrimitiveType) Pos() Position        { return Position{Line: t.Token.Line, Column: t.Token.Column, File: t.Token.File} }
func (t *PrimitiveType) typeNode()            {}

type NamedType struct {
	Token lexer.Token // The identifier token
	Name  string
}

func (t *NamedType) TokenLiteral() string { return t.Token.Lexeme }
func (t *NamedType) Pos() Position        { return Position{Line: t.Token.Line, Column: t.Token.Column, File: t.Token.File} }
func (t *NamedType) typeNode()            {}

type ReferenceType struct {
	Token       lexer.Token // The 'reference' token
	ElementType TypeAnnotation
}

func (t *ReferenceType) TokenLiteral() string { return t.Token.Lexeme }
func (t *ReferenceType) Pos() Position        { return Position{Line: t.Token.Line, Column: t.Token.Column, File: t.Token.File} }
func (t *ReferenceType) typeNode()            {}

type ListType struct {
	Token       lexer.Token // The 'list' token
	ElementType TypeAnnotation
}

func (t *ListType) TokenLiteral() string { return t.Token.Lexeme }
func (t *ListType) Pos() Position        { return Position{Line: t.Token.Line, Column: t.Token.Column, File: t.Token.File} }
func (t *ListType) typeNode()            {}

type MapType struct {
	Token     lexer.Token // The 'map' token
	KeyType   TypeAnnotation
	ValueType TypeAnnotation
}

func (t *MapType) TokenLiteral() string { return t.Token.Lexeme }
func (t *MapType) Pos() Position        { return Position{Line: t.Token.Line, Column: t.Token.Column, File: t.Token.File} }
func (t *MapType) typeNode()            {}

type ChannelType struct {
	Token       lexer.Token // The 'channel' token
	ElementType TypeAnnotation
}

func (t *ChannelType) TokenLiteral() string { return t.Token.Lexeme }
func (t *ChannelType) Pos() Position        { return Position{Line: t.Token.Line, Column: t.Token.Column, File: t.Token.File} }
func (t *ChannelType) typeNode()            {}

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
func (s *BlockStmt) Pos() Position        { return Position{Line: s.Token.Line, Column: s.Token.Column, File: s.Token.File} }
func (s *BlockStmt) stmtNode()            {}

type VarDeclStmt struct {
	Name  *Identifier
	Type  TypeAnnotation // Optional (can be nil for inference)
	Value Expression
	Token lexer.Token // The identifier token or walrus token
}

func (s *VarDeclStmt) TokenLiteral() string { return s.Token.Lexeme }
func (s *VarDeclStmt) Pos() Position        { return Position{Line: s.Token.Line, Column: s.Token.Column, File: s.Token.File} }
func (s *VarDeclStmt) stmtNode()            {}

type AssignStmt struct {
	Target Expression  // Can be identifier, index expr, etc.
	Value  Expression
	Token  lexer.Token // The '=' token
}

func (s *AssignStmt) TokenLiteral() string { return s.Token.Lexeme }
func (s *AssignStmt) Pos() Position        { return Position{Line: s.Token.Line, Column: s.Token.Column, File: s.Token.File} }
func (s *AssignStmt) stmtNode()            {}

type ReturnStmt struct {
	Token  lexer.Token // The 'return' token
	Values []Expression
}

func (s *ReturnStmt) TokenLiteral() string { return s.Token.Lexeme }
func (s *ReturnStmt) Pos() Position        { return Position{Line: s.Token.Line, Column: s.Token.Column, File: s.Token.File} }
func (s *ReturnStmt) stmtNode()            {}

type IfStmt struct {
	Token       lexer.Token // The 'if' token
	Condition   Expression
	Consequence *BlockStmt
	Alternative Statement // Can be ElseStmt or another IfStmt (else if)
}

func (s *IfStmt) TokenLiteral() string { return s.Token.Lexeme }
func (s *IfStmt) Pos() Position        { return Position{Line: s.Token.Line, Column: s.Token.Column, File: s.Token.File} }
func (s *IfStmt) stmtNode()            {}

type ElseStmt struct {
	Token lexer.Token // The 'else' token
	Body  *BlockStmt
}

func (s *ElseStmt) TokenLiteral() string { return s.Token.Lexeme }
func (s *ElseStmt) Pos() Position        { return Position{Line: s.Token.Line, Column: s.Token.Column, File: s.Token.File} }
func (s *ElseStmt) stmtNode()            {}

// ForRangeStmt: for item in collection
type ForRangeStmt struct {
	Token      lexer.Token // The 'for' token
	Variable   *Identifier
	Index      *Identifier // Optional (for index, item in collection)
	Collection Expression
	Body       *BlockStmt
}

func (s *ForRangeStmt) TokenLiteral() string { return s.Token.Lexeme }
func (s *ForRangeStmt) Pos() Position        { return Position{Line: s.Token.Line, Column: s.Token.Column, File: s.Token.File} }
func (s *ForRangeStmt) stmtNode()            {}

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
func (s *ForNumericStmt) Pos() Position        { return Position{Line: s.Token.Line, Column: s.Token.Column, File: s.Token.File} }
func (s *ForNumericStmt) stmtNode()            {}

// ForConditionStmt: for condition
type ForConditionStmt struct {
	Token     lexer.Token // The 'for' token
	Condition Expression
	Body      *BlockStmt
}

func (s *ForConditionStmt) TokenLiteral() string { return s.Token.Lexeme }
func (s *ForConditionStmt) Pos() Position        { return Position{Line: s.Token.Line, Column: s.Token.Column, File: s.Token.File} }
func (s *ForConditionStmt) stmtNode()            {}

type DeferStmt struct {
	Token lexer.Token // The 'defer' token
	Call  *CallExpr
}

func (s *DeferStmt) TokenLiteral() string { return s.Token.Lexeme }
func (s *DeferStmt) Pos() Position        { return Position{Line: s.Token.Line, Column: s.Token.Column, File: s.Token.File} }
func (s *DeferStmt) stmtNode()            {}

type GoStmt struct {
	Token lexer.Token // The 'go' token
	Call  *CallExpr
}

func (s *GoStmt) TokenLiteral() string { return s.Token.Lexeme }
func (s *GoStmt) Pos() Position        { return Position{Line: s.Token.Line, Column: s.Token.Column, File: s.Token.File} }
func (s *GoStmt) stmtNode()            {}

type SendStmt struct {
	Token   lexer.Token // The 'send' token
	Value   Expression
	Channel Expression
}

func (s *SendStmt) TokenLiteral() string { return s.Token.Lexeme }
func (s *SendStmt) Pos() Position        { return Position{Line: s.Token.Line, Column: s.Token.Column, File: s.Token.File} }
func (s *SendStmt) stmtNode()            {}

type ExpressionStmt struct {
	Expression Expression
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
func (e *Identifier) Pos() Position        { return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File} }
func (e *Identifier) exprNode()            {}

type IntegerLiteral struct {
	Token lexer.Token
	Value int64
}

func (e *IntegerLiteral) TokenLiteral() string { return e.Token.Lexeme }
func (e *IntegerLiteral) Pos() Position        { return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File} }
func (e *IntegerLiteral) exprNode()            {}

type FloatLiteral struct {
	Token lexer.Token
	Value float64
}

func (e *FloatLiteral) TokenLiteral() string { return e.Token.Lexeme }
func (e *FloatLiteral) Pos() Position        { return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File} }
func (e *FloatLiteral) exprNode()            {}

type StringLiteral struct {
	Token        lexer.Token
	Value        string
	Interpolated bool                   // True if contains {expr}
	Parts        []*StringInterpolation // For interpolated strings
}

func (e *StringLiteral) TokenLiteral() string { return e.Token.Lexeme }
func (e *StringLiteral) Pos() Position        { return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File} }
func (e *StringLiteral) exprNode()            {}

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
func (e *BooleanLiteral) Pos() Position        { return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File} }
func (e *BooleanLiteral) exprNode()            {}

type BinaryExpr struct {
	Token    lexer.Token // The operator token
	Left     Expression
	Operator string
	Right    Expression
}

func (e *BinaryExpr) TokenLiteral() string { return e.Token.Lexeme }
func (e *BinaryExpr) Pos() Position        { return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File} }
func (e *BinaryExpr) exprNode()            {}

type UnaryExpr struct {
	Token    lexer.Token // The operator token
	Operator string
	Right    Expression
}

func (e *UnaryExpr) TokenLiteral() string { return e.Token.Lexeme }
func (e *UnaryExpr) Pos() Position        { return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File} }
func (e *UnaryExpr) exprNode()            {}

type PipeExpr struct {
	Token lexer.Token // The '|>' token
	Left  Expression
	Right Expression // Must be a function call
}

func (e *PipeExpr) TokenLiteral() string { return e.Token.Lexeme }
func (e *PipeExpr) Pos() Position        { return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File} }
func (e *PipeExpr) exprNode()            {}

type OnErrExpr struct {
	Token   lexer.Token // The 'onerr' token
	Left    Expression  // Expression that might error
	Handler Expression  // Error handler (can be discard or expression)
}

func (e *OnErrExpr) TokenLiteral() string { return e.Token.Lexeme }
func (e *OnErrExpr) Pos() Position        { return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File} }
func (e *OnErrExpr) exprNode()            {}

type CallExpr struct {
	Token     lexer.Token // The '(' token or identifier
	Function  Expression
	Arguments []Expression
}

func (e *CallExpr) TokenLiteral() string { return e.Token.Lexeme }
func (e *CallExpr) Pos() Position        { return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File} }
func (e *CallExpr) exprNode()            {}

type MethodCallExpr struct {
	Token     lexer.Token // The '.' token
	Object    Expression
	Method    *Identifier
	Arguments []Expression
}

func (e *MethodCallExpr) TokenLiteral() string { return e.Token.Lexeme }
func (e *MethodCallExpr) Pos() Position        { return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File} }
func (e *MethodCallExpr) exprNode()            {}

type IndexExpr struct {
	Token lexer.Token // The '[' token
	Left  Expression
	Index Expression
}

func (e *IndexExpr) TokenLiteral() string { return e.Token.Lexeme }
func (e *IndexExpr) Pos() Position        { return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File} }
func (e *IndexExpr) exprNode()            {}

type SliceExpr struct {
	Token lexer.Token // The '[' token
	Left  Expression
	Start Expression // Can be nil
	End   Expression // Can be nil
}

func (e *SliceExpr) TokenLiteral() string { return e.Token.Lexeme }
func (e *SliceExpr) Pos() Position        { return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File} }
func (e *SliceExpr) exprNode()            {}

type StructLiteralExpr struct {
	Token  lexer.Token // The type identifier
	Type   TypeAnnotation
	Fields []*FieldValue
}

func (e *StructLiteralExpr) TokenLiteral() string { return e.Token.Lexeme }
func (e *StructLiteralExpr) Pos() Position        { return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File} }
func (e *StructLiteralExpr) exprNode()            {}

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
func (e *ListLiteralExpr) Pos() Position        { return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File} }
func (e *ListLiteralExpr) exprNode()            {}

type MapLiteralExpr struct {
	Token   lexer.Token // The '{' token or 'map' keyword
	KeyType TypeAnnotation
	ValType TypeAnnotation
	Pairs   []*KeyValuePair
}

func (e *MapLiteralExpr) TokenLiteral() string { return e.Token.Lexeme }
func (e *MapLiteralExpr) Pos() Position        { return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File} }
func (e *MapLiteralExpr) exprNode()            {}

type KeyValuePair struct {
	Key   Expression
	Value Expression
}

type ReceiveExpr struct {
	Token   lexer.Token // The 'receive' token
	Channel Expression
}

func (e *ReceiveExpr) TokenLiteral() string { return e.Token.Lexeme }
func (e *ReceiveExpr) Pos() Position        { return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File} }
func (e *ReceiveExpr) exprNode()            {}

type TypeCastExpr struct {
	Token      lexer.Token // The 'as' token
	Expression Expression
	TargetType TypeAnnotation
}

func (e *TypeCastExpr) TokenLiteral() string { return e.Token.Lexeme }
func (e *TypeCastExpr) Pos() Position        { return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File} }
func (e *TypeCastExpr) exprNode()            {}

type ThisExpr struct {
	Token lexer.Token // The 'this' token
}

func (e *ThisExpr) TokenLiteral() string { return e.Token.Lexeme }
func (e *ThisExpr) Pos() Position        { return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File} }
func (e *ThisExpr) exprNode()            {}

type EmptyExpr struct {
	Token lexer.Token // The 'empty' token
	Type  TypeAnnotation
}

func (e *EmptyExpr) TokenLiteral() string { return e.Token.Lexeme }
func (e *EmptyExpr) Pos() Position        { return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File} }
func (e *EmptyExpr) exprNode()            {}

type DiscardExpr struct {
	Token lexer.Token // The 'discard' token
}

func (e *DiscardExpr) TokenLiteral() string { return e.Token.Lexeme }
func (e *DiscardExpr) Pos() Position        { return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File} }
func (e *DiscardExpr) exprNode()            {}

type ErrorExpr struct {
	Token   lexer.Token // The 'error' token
	Message Expression  // Usually a string literal
}

func (e *ErrorExpr) TokenLiteral() string { return e.Token.Lexeme }
func (e *ErrorExpr) Pos() Position        { return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File} }
func (e *ErrorExpr) exprNode()            {}

type MakeExpr struct {
	Token lexer.Token // The 'make' token
	Type  TypeAnnotation
	Args  []Expression // Size/capacity for slices, channels
}

func (e *MakeExpr) TokenLiteral() string { return e.Token.Lexeme }
func (e *MakeExpr) Pos() Position        { return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File} }
func (e *MakeExpr) exprNode()            {}

type CloseExpr struct {
	Token   lexer.Token // The 'close' token
	Channel Expression
}

func (e *CloseExpr) TokenLiteral() string { return e.Token.Lexeme }
func (e *CloseExpr) Pos() Position        { return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File} }
func (e *CloseExpr) exprNode()            {}

type PanicExpr struct {
	Token   lexer.Token // The 'panic' token
	Message Expression
}

func (e *PanicExpr) TokenLiteral() string { return e.Token.Lexeme }
func (e *PanicExpr) Pos() Position        { return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File} }
func (e *PanicExpr) exprNode()            {}

type RecoverExpr struct {
	Token lexer.Token // The 'recover' token
}

func (e *RecoverExpr) TokenLiteral() string { return e.Token.Lexeme }
func (e *RecoverExpr) Pos() Position        { return Position{Line: e.Token.Line, Column: e.Token.Column, File: e.Token.File} }
func (e *RecoverExpr) exprNode()            {}
