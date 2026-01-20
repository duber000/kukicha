# Kukicha Compiler Architecture

**Version:** 1.0.0
**Target:** Go 1.25+ with Green Tea GC
**Approach:** Transpiler (Kukicha → Go → Native Binary)

---

## Overview

The Kukicha compiler is a source-to-source transpiler that converts `.kuki` files into clean, idiomatic Go code, which is then compiled with the Go toolchain.

```
┌─────────────┐      ┌────────────┐      ┌──────────────┐      ┌─────────────┐
│ .kuki files │ ───▶ │   Lexer    │ ───▶ │    Parser    │ ───▶ │     AST     │
└─────────────┘      │ (Tokenize) │      │ (Build Tree) │      │  (Memory)   │
                     └────────────┘      └──────────────┘      └─────────────┘
                                                                       │
                                                                       ▼
┌─────────────┐      ┌────────────┐      ┌──────────────┐      ┌─────────────┐
│   Binary    │ ◀─── │ Go Compile │ ◀─── │ Code Gen     │ ◀─── │  Semantic   │
│ Executable  │      │ (go build) │      │ (AST → Go)   │      │  Analysis   │
└─────────────┘      └────────────┘      └──────────────┘      └─────────────┘
```

---

## Phase 1: Lexer (Tokenizer)

### Purpose
Convert raw source code into a stream of tokens.

### Input
```kukicha
func Greet(name string)
    print "Hello {name}"
```

### Output
```
FUNC, IDENTIFIER(Greet), LPAREN, IDENTIFIER(name), IDENTIFIER(string), RPAREN, NEWLINE
INDENT, IDENTIFIER(print), STRING("Hello {name}"), NEWLINE
DEDENT, EOF
```

### Token Types

```go
type TokenType int

const (
    // Literals
    TOKEN_IDENTIFIER TokenType = iota
    TOKEN_INTEGER
    TOKEN_FLOAT
    TOKEN_STRING
    TOKEN_TRUE
    TOKEN_FALSE
    
    // Keywords
    TOKEN_LEAF
    TOKEN_IMPORT
    TOKEN_TYPE
    TOKEN_INTERFACE
    TOKEN_FUNC
    TOKEN_RETURN
    TOKEN_IF
    TOKEN_ELSE
    TOKEN_FOR
    TOKEN_IN
    TOKEN_FROM
    TOKEN_TO
    TOKEN_THROUGH
    TOKEN_GO
    TOKEN_DEFER
    TOKEN_MAKE
    TOKEN_CHANNEL
    TOKEN_SEND
    TOKEN_RECEIVE
    TOKEN_CLOSE
    TOKEN_PANIC
    TOKEN_RECOVER
    TOKEN_ERROR
    TOKEN_EMPTY
    TOKEN_REFERENCE
    TOKEN_ON
    TOKEN_THIS
    TOKEN_DISCARD
    TOKEN_AT
    TOKEN_OF
    
    // Operators
    TOKEN_WALRUS        // :=
    TOKEN_ASSIGN        // =
    TOKEN_EQUALS        // equals
    TOKEN_DOUBLE_EQUALS // ==
    TOKEN_NOT_EQUALS    // !=
    TOKEN_LT            // <
    TOKEN_GT            // >
    TOKEN_LTE           // <=
    TOKEN_GTE           // >=
    TOKEN_PLUS          // +
    TOKEN_MINUS         // -
    TOKEN_STAR          // *
    TOKEN_SLASH         // /
    TOKEN_PERCENT       // %
    TOKEN_AND           // and
    TOKEN_AND_AND       // &&
    TOKEN_OR            // or
    TOKEN_OR_OR         // ||
    TOKEN_NOT           // not
    TOKEN_BANG          // !
    TOKEN_PIPE          // |>
    TOKEN_ARROW_LEFT    // <-
    
    // Delimiters
    TOKEN_LPAREN        // (
    TOKEN_RPAREN        // )
    TOKEN_LBRACKET      // [
    TOKEN_RBRACKET      // ]
    TOKEN_LBRACE        // {
    TOKEN_RBRACE        // }
    TOKEN_COMMA         // ,
    TOKEN_DOT           // .
    TOKEN_COLON         // :
    
    // Special
    TOKEN_NEWLINE
    TOKEN_INDENT
    TOKEN_DEDENT
    TOKEN_EOF
    TOKEN_ERROR
)

type Token struct {
    Type    TokenType
    Lexeme  string
    Line    int
    Column  int
    File    string
}
```

### Lexer Structure

```go
type Lexer struct {
    source          []rune
    start           int
    current         int
    line            int
    column          int
    file            string
    tokens          []Token
    indentStack     []int  // Track indentation levels
    pendingDedents  int    // Dedents to emit
    
    // String interpolation tracking
    braceDepth      int
    inString        bool
}

func NewLexer(source string, filename string) *Lexer {
    return &Lexer{
        source:      []rune(source),
        file:        filename,
        line:        1,
        column:      1,
        indentStack: []int{0},
    }
}

func (l *Lexer) ScanTokens() ([]Token, error) {
    for !l.isAtEnd() {
        l.start = l.current
        l.scanToken()
    }
    
    // Emit remaining dedents
    for len(l.indentStack) > 1 {
        l.addToken(TOKEN_DEDENT)
        l.indentStack = l.indentStack[:len(l.indentStack)-1]
    }
    
    l.addToken(TOKEN_EOF)
    return l.tokens, nil
}

func (l *Lexer) scanToken() {
    c := l.advance()
    
    switch c {
    case ' ', '\t':
        // Handle at line start for indentation
        if l.column == 1 {
            l.handleIndentation()
        }
    case '\n', '\r':
        l.addToken(TOKEN_NEWLINE)
        l.line++
        l.column = 0
    case '#':
        l.skipComment()
    case '"', '\'':
        l.scanString(c)
    case '(':
        l.addToken(TOKEN_LPAREN)
    case ')':
        l.addToken(TOKEN_RPAREN)
    // ... more cases
    default:
        if isDigit(c) {
            l.scanNumber()
        } else if isAlpha(c) {
            l.scanIdentifier()
        } else {
            l.error("Unexpected character")
        }
    }
}
```

### Indentation Handling

```go
func (l *Lexer) handleIndentation() {
    spaces := 0
    
    // Count spaces (reject tabs)
    for !l.isAtEnd() && l.peek() == ' ' {
        spaces++
        l.advance()
    }
    
    // Check for tabs
    if !l.isAtEnd() && l.peek() == '\t' {
        l.error("Use 4 spaces for indentation, not tabs")
        return
    }
    
    // Must be multiple of 4
    if spaces%4 != 0 {
        l.error(fmt.Sprintf("Indentation must be multiple of 4 spaces, got %d", spaces))
        return
    }
    
    currentIndent := l.indentStack[len(l.indentStack)-1]
    
    if spaces > currentIndent {
        // Indent
        l.indentStack = append(l.indentStack, spaces)
        l.addToken(TOKEN_INDENT)
    } else if spaces < currentIndent {
        // Dedent (possibly multiple levels)
        for len(l.indentStack) > 1 && l.indentStack[len(l.indentStack)-1] > spaces {
            l.addToken(TOKEN_DEDENT)
            l.indentStack = l.indentStack[:len(l.indentStack)-1]
        }
        
        // Verify we landed on a valid indentation level
        if l.indentStack[len(l.indentStack)-1] != spaces {
            l.error("Indentation mismatch")
        }
    }
}
```

### String Interpolation

```go
func (l *Lexer) scanString(quote rune) {
    segments := []StringSegment{}
    current := strings.Builder{}
    
    for !l.isAtEnd() && l.peek() != quote {
        if l.peek() == '{' {
            // Save literal segment
            if current.Len() > 0 {
                segments = append(segments, StringSegment{
                    Type: SEGMENT_LITERAL,
                    Value: current.String(),
                })
                current.Reset()
            }
            
            // Scan interpolation expression
            l.advance() // consume {
            expr := l.scanInterpolation()
            segments = append(segments, StringSegment{
                Type: SEGMENT_INTERPOLATION,
                Value: expr,
            })
        } else {
            current.WriteRune(l.advance())
        }
    }
    
    // Save remaining literal
    if current.Len() > 0 {
        segments = append(segments, StringSegment{
            Type: SEGMENT_LITERAL,
            Value: current.String(),
        })
    }
    
    if l.isAtEnd() {
        l.error("Unterminated string")
        return
    }
    
    l.advance() // consume closing quote
    
    l.addStringToken(segments)
}
```

---

## Phase 2: Parser (AST Builder)

### Purpose
Convert token stream into Abstract Syntax Tree (AST).

### AST Node Types

```go
type Node interface {
    node()
    Position() Position
}

type Position struct {
    Line   int
    Column int
    File   string
}

// Program root
type Program struct {
    Leaf        *LeafDecl      // Optional: nil if not declared
    FilePath    string         // Required: used to calculate implicit Stem
    TwigRoot    string         // Required: path to twig.toml directory
    Imports     []*ImportDecl
    Declarations []Declaration
}

// CalculateStem computes the package name from file path if Leaf is nil
func (p *Program) CalculateStem() string {
    if p.Leaf != nil {
        return p.Leaf.Name  // Explicit declaration takes precedence
    }

    // Calculate stem from file path relative to twig.toml
    relPath := filepath.Rel(p.TwigRoot, filepath.Dir(p.FilePath))
    // Convert path to package name (e.g., "src/auth" -> "auth")
    stem := filepath.Base(relPath)
    return stem
}

// Declarations
type Declaration interface {
    Node
    declaration()
}

type TypeDecl struct {
    Name   string
    Fields []*Field
    Pos    Position
}

type Field struct {
    Name string
    Type TypeExpr
    Pos  Position
}

type InterfaceDecl struct {
    Name    string
    Methods []*MethodSignature
    Pos     Position
}

type FunctionDecl struct {
    Name       string
    Params     []*Parameter
    ReturnType TypeExpr
    Body       *BlockStmt
    Pos        Position
}

type MethodDecl struct {
    Name        string
    Receiver    *Receiver
    Params      []*Parameter
    ReturnType  TypeExpr
    Body        *BlockStmt
    Pos         Position
}

type Receiver struct {
    Name       string
    Type       TypeExpr
    IsPointer  bool
    Pos        Position
}

// Type Expressions
type TypeExpr interface {
    Node
    typeExpr()
}

type SimpleType struct {
    Name string
    Pos  Position
}

type ReferenceType struct {
    Inner TypeExpr
    Pos   Position
}

type ListType struct {
    Element TypeExpr
    Pos     Position
}

type MapType struct {
    Key   TypeExpr
    Value TypeExpr
    Pos   Position
}

type ChannelType struct {
    Element TypeExpr
    Pos     Position
}

// Statements
type Statement interface {
    Node
    statement()
}

type BlockStmt struct {
    Statements []Statement
    Pos        Position
}

type VarDeclStmt struct {
    Name  string
    Value Expr
    Pos   Position
}

type AssignStmt struct {
    Target Expr
    Value  Expr
    Pos    Position
}

type ReturnStmt struct {
    Values []Expr
    Pos    Position
}

type IfStmt struct {
    Condition Expr
    Then      *BlockStmt
    Else      Statement // Can be another IfStmt or BlockStmt
    Pos       Position
}

type ForStmt struct {
    // Range loop
    Var   string
    From  Expr
    To    Expr
    Inclusive bool // true for "through", false for "to"
    
    // Collection loop
    Index  string // can be "discard"
    Value  string
    Collection Expr
    
    // Go-style loop
    Init      Statement
    Condition Expr
    Post      Statement
    
    Body *BlockStmt
    Type ForType
    Pos  Position
}

type ForType int
const (
    FOR_RANGE ForType = iota
    FOR_COLLECTION
    FOR_GOSTYLE
)

type DeferStmt struct {
    Call Expr // or Block for defer with recover
    Block *BlockStmt
    Pos  Position
}

type GoStmt struct {
    Call  Expr
    Block *BlockStmt
    Pos   Position
}

type SendStmt struct {
    Channel Expr
    Value   Expr
    Pos     Position
}

// Expressions
type Expr interface {
    Node
    expr()
}

type BinaryExpr struct {
    Left     Expr
    Operator TokenType
    Right    Expr
    Pos      Position
}

type UnaryExpr struct {
    Operator TokenType
    Operand  Expr
    Pos      Position
}

type PipeExpr struct {
    Left  Expr
    Right Expr // Function call
    Pos   Position
}

type OnErrExpr struct {
    Left    Expr
    Handler Expr // Can be panic, return, or default value
    Pos     Position
}

type CallExpr struct {
    Function Expr
    Args     []Expr
    Pos      Position
}

type MethodCallExpr struct {
    Object Expr
    Method string
    Args   []Expr
    Pos    Position
}

type IndexExpr struct {
    Object Expr
    Index  Expr
    Pos    Position
}

type SliceExpr struct {
    Object    Expr
    Start     Expr
    End       Expr
    Inclusive bool
    Pos       Position
}

type IdentifierExpr struct {
    Name string
    Pos  Position
}

type LiteralExpr struct {
    Type  LiteralType
    Value interface{}
    Pos   Position
}

type LiteralType int
const (
    LIT_INTEGER LiteralType = iota
    LIT_FLOAT
    LIT_STRING
    LIT_BOOL
)

type StringLiteralExpr struct {
    Segments []StringSegment
    Pos      Position
}

type StringSegment struct {
    Type  SegmentType
    Value string // literal text or expression string
}

type SegmentType int
const (
    SEGMENT_LITERAL SegmentType = iota
    SEGMENT_INTERPOLATION
)

type StructLiteralExpr struct {
    Type   TypeExpr
    Fields []*FieldInit
    Pos    Position
}

type FieldInit struct {
    Name  string
    Value Expr
    Pos   Position
}

type ListLiteralExpr struct {
    Type     TypeExpr
    Elements []Expr
    Pos      Position
}

type MapLiteralExpr struct {
    Type    TypeExpr
    Entries []*MapEntry
    Pos     Position
}

type MapEntry struct {
    Key   Expr
    Value Expr
    Pos   Position
}

type ReceiveExpr struct {
    Channel Expr
    Pos     Position
}

type TypeCastExpr struct {
    Type  TypeExpr
    Value Expr
    Pos   Position
}

type ThisExpr struct {
    Pos Position
}

type EmptyExpr struct {
    Type TypeExpr // Can be nil for untyped empty
    Pos  Position
}
```

### Parser Structure

```go
type Parser struct {
    tokens  []Token
    current int
    file    string
    errors  []error
}

func NewParser(tokens []Token, filename string) *Parser {
    return &Parser{
        tokens: tokens,
        file:   filename,
    }
}

func (p *Parser) Parse() (*Program, error) {
    program := &Program{}
    
    // Parse leaf declaration
    program.Leaf = p.parseLeafDecl()
    
    // Parse imports
    for p.match(TOKEN_IMPORT) {
        program.Imports = append(program.Imports, p.parseImport())
    }
    
    // Parse top-level declarations
    for !p.isAtEnd() {
        program.Declarations = append(program.Declarations, p.parseDeclaration())
    }
    
    if len(p.errors) > 0 {
        return nil, fmt.Errorf("parse errors: %v", p.errors)
    }
    
    return program, nil
}

func (p *Parser) parseDeclaration() Declaration {
    switch {
    case p.match(TOKEN_TYPE):
        return p.parseTypeDecl()
    case p.match(TOKEN_INTERFACE):
        return p.parseInterfaceDecl()
    case p.match(TOKEN_FUNC):
        return p.parseFunctionOrMethod()
    default:
        p.error("Expected declaration")
        return nil
    }
}

func (p *Parser) parseExpression() Expr {
    return p.parseOrExpression()
}

func (p *Parser) parseOrExpression() Expr {
    left := p.parsePipeExpression()
    
    if p.match(TOKEN_OR) {
        // Or operator for error handling
        handler := p.parseOrHandler()
        return &OrExpr{
            Left:    left,
            Handler: handler,
            Pos:     left.Position(),
        }
    }
    
    // Or as boolean operator
    for p.match(TOKEN_OR, TOKEN_OR_OR) {
        op := p.previous().Type
        right := p.parsePipeExpression()
        left = &BinaryExpr{
            Left:     left,
            Operator: op,
            Right:    right,
            Pos:      left.Position(),
        }
    }
    
    return left
}

func (p *Parser) parsePipeExpression() Expr {
    left := p.parseAndExpression()
    
    for p.match(TOKEN_PIPE) {
        right := p.parseAndExpression()
        left = &PipeExpr{
            Left:  left,
            Right: right,
            Pos:   left.Position(),
        }
    }
    
    return left
}
```

---

## Phase 3: Semantic Analysis

### Purpose
Type checking, name resolution, and validation.

### Symbol Table

```go
type SymbolTable struct {
    scopes []*Scope
}

type Scope struct {
    parent  *Scope
    symbols map[string]*Symbol
}

type Symbol struct {
    Name       string
    Type       TypeInfo
    Kind       SymbolKind
    Defined    Position
    Mutable    bool
}

type SymbolKind int
const (
    SYM_VARIABLE SymbolKind = iota
    SYM_PARAMETER
    SYM_FUNCTION
    SYM_METHOD
    SYM_TYPE
    SYM_FIELD
)

type TypeInfo interface {
    String() string
    Equals(other TypeInfo) bool
}

// Type representations
type PrimitiveTypeInfo struct {
    Kind PrimitiveKind
}

type StructTypeInfo struct {
    Name   string
    Fields map[string]TypeInfo
}

type InterfaceTypeInfo struct {
    Name    string
    Methods map[string]*MethodSignature
}

type ListTypeInfo struct {
    Element TypeInfo
}

type MapTypeInfo struct {
    Key   TypeInfo
    Value TypeInfo
}

type ReferenceTypeInfo struct {
    Inner TypeInfo
}

type ChannelTypeInfo struct {
    Element TypeInfo
}

type FunctionTypeInfo struct {
    Params  []TypeInfo
    Returns []TypeInfo
}
```

### Semantic Analyzer

```go
type SemanticAnalyzer struct {
    symbolTable *SymbolTable
    currentFunc *FunctionDecl
    errors      []error
}

func NewSemanticAnalyzer() *SemanticAnalyzer {
    return &SemanticAnalyzer{
        symbolTable: NewSymbolTable(),
    }
}

func (sa *SemanticAnalyzer) Analyze(program *Program) error {
    // First pass: collect all type declarations
    sa.collectTypeDeclarations(program)

    // Second pass: SIGNATURE-FIRST - collect all function/method signatures
    // This maps all function inputs/outputs BEFORE analyzing bodies
    sa.collectFunctionSignatures(program)

    // Third pass: check interfaces implementation
    sa.checkInterfaces(program)

    // Fourth pass: type check function bodies
    // Local variables are inferred using := within function bodies
    sa.checkFunctionBodies(program)

    if len(sa.errors) > 0 {
        return fmt.Errorf("semantic errors: %v", sa.errors)
    }

    return nil
}

// Signature-First Type Checking Pass
func (sa *SemanticAnalyzer) collectFunctionSignatures(program *Program) {
    for _, decl := range program.Declarations {
        switch d := decl.(type) {
        case *FunctionDecl:
            // Verify parameters have explicit types
            for _, param := range d.Params {
                if param.Type == nil {
                    sa.error(param.Pos, "Function parameters must have explicit type annotations")
                }
            }

            // Verify return type is explicit
            if d.ReturnType == nil && sa.functionReturnsValue(d) {
                sa.error(d.Pos, "Function return types must be explicit")
            }

            // Register function signature in symbol table
            sa.registerFunctionSignature(d)

        case *MethodDecl:
            // Same verification for methods
            for _, param := range d.Params {
                if param.Type == nil {
                    sa.error(param.Pos, "Method parameters must have explicit type annotations")
                }
            }

            if d.ReturnType == nil && sa.methodReturnsValue(d) {
                sa.error(d.Pos, "Method return types must be explicit")
            }

            sa.registerMethodSignature(d)
        }
    }
}

func (sa *SemanticAnalyzer) checkExpr(expr Expr) TypeInfo {
    switch e := expr.(type) {
    case *BinaryExpr:
        return sa.checkBinaryExpr(e)
    case *OrExpr:
        return sa.checkOrExpr(e)
    case *PipeExpr:
        return sa.checkPipeExpr(e)
    case *CallExpr:
        return sa.checkCallExpr(e)
    // ... more cases
    }
    
    return nil
}

func (sa *SemanticAnalyzer) checkOnErrExpr(expr *OnErrExpr) TypeInfo {
    leftType := sa.checkExpr(expr.Left)

    // Check if left is a function returning (T, error)
    if funcType, ok := leftType.(*FunctionTypeInfo); ok {
        if len(funcType.Returns) == 2 {
            // Valid onerr operator usage
            return funcType.Returns[0]
        }
    }

    sa.error(expr.Pos, "'onerr' operator requires function returning (T, error)")
    return nil
}

func (sa *SemanticAnalyzer) checkPipeExpr(expr *PipeExpr) TypeInfo {
    leftType := sa.checkExpr(expr.Left)
    
    // Right must be a function call
    call, ok := expr.Right.(*CallExpr)
    if !ok {
        sa.error(expr.Pos, "Pipe right side must be function call")
        return nil
    }
    
    // Insert left as first argument
    call.Args = append([]Expr{expr.Left}, call.Args...)
    
    return sa.checkCallExpr(call)
}
```

---

## Phase 4: Code Generation

### Purpose
Convert AST to idiomatic Go code.

### Code Generator

```go
type CodeGenerator struct {
    output      strings.Builder
    indentLevel int
    packageName string
}

func NewCodeGenerator() *CodeGenerator {
    return &CodeGenerator{}
}

func (cg *CodeGenerator) Generate(program *Program) (string, error) {
    // Generate package declaration
    cg.writeLine("package %s", program.Leaf.Name)
    cg.writeLine("")
    
    // Generate imports
    cg.generateImports(program.Imports)
    cg.writeLine("")
    
    // Generate type declarations
    for _, decl := range program.Declarations {
        switch d := decl.(type) {
        case *TypeDecl:
            cg.generateTypeDecl(d)
        case *InterfaceDecl:
            cg.generateInterfaceDecl(d)
        case *FunctionDecl:
            cg.generateFunctionDecl(d)
        case *MethodDecl:
            cg.generateMethodDecl(d)
        }
        cg.writeLine("")
    }
    
    return cg.output.String(), nil
}

func (cg *CodeGenerator) generateTypeDecl(decl *TypeDecl) {
    cg.writeLine("type %s struct {", decl.Name)
    cg.indent()
    
    for _, field := range decl.Fields {
        cg.writeLine("%s %s", capitalize(field.Name), cg.typeToGo(field.Type))
    }
    
    cg.dedent()
    cg.writeLine("}")
}

func (cg *CodeGenerator) generateMethodDecl(decl *MethodDecl) {
    receiver := cg.receiverToGo(decl.Receiver)
    params := cg.paramsToGo(decl.Params)
    returnType := cg.returnTypeToGo(decl.ReturnType)
    
    cg.writeLine("func %s %s(%s) %s {", receiver, decl.Name, params, returnType)
    cg.indent()
    cg.generateBlock(decl.Body)
    cg.dedent()
    cg.writeLine("}")
}

func (cg *CodeGenerator) generateExpr(expr Expr) string {
    switch e := expr.(type) {
    case *BinaryExpr:
        return cg.generateBinaryExpr(e)
    case *OrExpr:
        return cg.generateOrExpr(e)
    case *PipeExpr:
        return cg.generatePipeExpr(e)
    case *CallExpr:
        return cg.generateCallExpr(e)
    case *StringLiteralExpr:
        return cg.generateStringLiteral(e)
    // ... more cases
    }
    
    return ""
}

func (cg *CodeGenerator) generateOnErrExpr(expr *OnErrExpr) string {
    // Desugar onerr operator
    // result := func() onerr handler
    // Becomes:
    // result, err := func()
    // if err != nil { handler }

    tmpVar := cg.generateTempVar()

    cg.writeLine("%s, err := %s", tmpVar, cg.generateExpr(expr.Left))
    cg.writeLine("if err != nil {")
    cg.indent()
    cg.generateExpr(expr.Handler)
    cg.dedent()
    cg.writeLine("}")

    return tmpVar
}

func (cg *CodeGenerator) generatePipeExpr(expr *PipeExpr) string {
    // Desugar pipe operator
    // a |> f() |> g(x)
    // Becomes: g(f(a), x)
    
    // Insert left as first arg to right (which must be CallExpr)
    call := expr.Right.(*CallExpr)
    call.Args = append([]Expr{expr.Left}, call.Args...)
    
    return cg.generateCallExpr(call)
}

func (cg *CodeGenerator) generateBinaryExpr(expr *BinaryExpr) string {
    // Handle special operators that need transformation

    switch expr.Operator {
    case TOKEN_IN:
        // Membership test: item in collection
        // For slices: slices.Contains(collection, item)
        // For maps: _, exists := map[key]; exists
        // For strings: strings.Contains(string, substring)

        collectionType := cg.typeOf(expr.Right)

        if collectionType.IsMap() {
            // Map: generate idiom _, exists := map[key]; exists
            return fmt.Sprintf("func() bool { _, exists := %s[%s]; return exists }()",
                cg.generateExpr(expr.Right),
                cg.generateExpr(expr.Left))
        } else if collectionType.IsString() {
            // String: strings.Contains(string, substring)
            return fmt.Sprintf("strings.Contains(%s, %s)",
                cg.generateExpr(expr.Right),
                cg.generateExpr(expr.Left))
        } else {
            // Slice: slices.Contains(slice, item)
            return fmt.Sprintf("slices.Contains(%s, %s)",
                cg.generateExpr(expr.Right),
                cg.generateExpr(expr.Left))
        }

    case TOKEN_NOT, TOKEN_BANG:
        if expr.Right.(*BinaryExpr).Operator == TOKEN_IN {
            // Handle "not in" as negated membership
            innerExpr := expr.Right.(*BinaryExpr)
            return "!" + cg.generateBinaryExpr(innerExpr)
        }
    }

    // Standard binary operators
    left := cg.generateExpr(expr.Left)
    right := cg.generateExpr(expr.Right)
    op := cg.operatorToGo(expr.Operator)

    return fmt.Sprintf("%s %s %s", left, op, right)
}

func (cg *CodeGenerator) generateIndexExpr(expr *IndexExpr) string {
    // LITERAL NEGATIVE INDEX OPTIMIZATION
    // Compile-time transformation for literal negative indices
    // items[-1] becomes items[len(items)-1] with zero runtime overhead

    if unary, ok := expr.Index.(*UnaryExpr); ok && unary.Operator == TOKEN_MINUS {
        // Check if the index is a LITERAL constant
        if isLiteralExpr(unary.Operand) {
            // LITERAL: Compile-time optimized path
            object := cg.generateExpr(expr.Object)
            index := cg.generateExpr(unary.Operand)
            return fmt.Sprintf("%s[len(%s)-%s]", object, object, index)
        } else {
            // DYNAMIC: Variable-based negative index requires .at() method
            // This will be caught in semantic analysis and require explicit syntax
            cg.error(expr.Pos, "Dynamic negative indexing requires .at() method: use items.at(index)")
            return ""
        }
    }

    // Standard positive indexing (always allowed)
    return fmt.Sprintf("%s[%s]",
        cg.generateExpr(expr.Object),
        cg.generateExpr(expr.Index))
}

// Helper: Check if expression is a literal constant
func isLiteralExpr(expr Expr) bool {
    switch expr.(type) {
    case *IntegerLiteralExpr, *FloatLiteralExpr:
        return true
    default:
        return false
    }
}

func (cg *CodeGenerator) generateSliceExpr(expr *SliceExpr) string {
    // LITERAL NEGATIVE SLICE OPTIMIZATION
    // Compile-time transformation for literal negative slice indices
    // items[-3:] becomes items[len(items)-3:]
    // items[:-1] becomes items[:len(items)-1]
    // items[1:-1] becomes items[1:len(items)-1]

    object := cg.generateExpr(expr.Object)
    var start, end string

    if expr.Start != nil {
        if unary, ok := expr.Start.(*UnaryExpr); ok && unary.Operator == TOKEN_MINUS {
            // Check if start is a LITERAL
            if isLiteralExpr(unary.Operand) {
                // LITERAL: Compile-time optimized
                startVal := cg.generateExpr(unary.Operand)
                start = fmt.Sprintf("len(%s)-%s", object, startVal)
            } else {
                // DYNAMIC: Requires .slice() method
                cg.error(expr.Pos, "Dynamic negative slicing requires .slice() method: use items.slice(start, end)")
                return ""
            }
        } else {
            start = cg.generateExpr(expr.Start)
        }
    }

    if expr.End != nil {
        if unary, ok := expr.End.(*UnaryExpr); ok && unary.Operator == TOKEN_MINUS {
            // Check if end is a LITERAL
            if isLiteralExpr(unary.Operand) {
                // LITERAL: Compile-time optimized
                endVal := cg.generateExpr(unary.Operand)
                end = fmt.Sprintf("len(%s)-%s", object, endVal)
            } else {
                // DYNAMIC: Requires .slice() method
                cg.error(expr.Pos, "Dynamic negative slicing requires .slice() method: use items.slice(start, end)")
                return ""
            }
        } else {
            end = cg.generateExpr(expr.End)
        }
    }

    return fmt.Sprintf("%s[%s:%s]", object, start, end)
}

func (cg *CodeGenerator) generateStringLiteral(expr *StringLiteralExpr) string {
    // Convert string interpolation to fmt.Sprintf

    if len(expr.Segments) == 1 && expr.Segments[0].Type == SEGMENT_LITERAL {
        // Simple string, no interpolation
        return fmt.Sprintf("%q", expr.Segments[0].Value)
    }

    // Build format string and args
    format := strings.Builder{}
    args := []string{}

    for _, seg := range expr.Segments {
        if seg.Type == SEGMENT_LITERAL {
            format.WriteString(seg.Value)
        } else {
            format.WriteString("%v")
            args = append(args, seg.Value) // Expression
        }
    }

    if len(args) == 0 {
        return fmt.Sprintf("%q", format.String())
    }

    return fmt.Sprintf("fmt.Sprintf(%q, %s)", format.String(), strings.Join(args, ", "))
}

func (cg *CodeGenerator) typeToGo(typ TypeExpr) string {
    switch t := typ.(type) {
    case *SimpleType:
        return t.Name
    case *ReferenceType:
        return "*" + cg.typeToGo(t.Inner)
    case *ListType:
        return "[]" + cg.typeToGo(t.Element)
    case *MapType:
        return fmt.Sprintf("map[%s]%s", cg.typeToGo(t.Key), cg.typeToGo(t.Value))
    case *ChannelType:
        return "chan " + cg.typeToGo(t.Element)
    }
    
    return ""
}
```

---

## Phase 5: Code Formatting (`kuki fmt`)

### Purpose
Ensure consistent, canonical indentation-based syntax across all Kukicha code.

### Design Goal
**Indentation is the source of truth.** To prevent "Dialect Drift" between Python-style and Go-style formatting, `kuki fmt` automatically converts brace-based syntax to the standard indentation-based format.

### Formatter Architecture

```go
type Formatter struct {
    lexer       *Lexer
    tokens      []Token
    output      strings.Builder
    indentLevel int
    needsIndent bool
}

func NewFormatter() *Formatter {
    return &Formatter{}
}

func (f *Formatter) Format(source string) (string, error) {
    // Tokenize the source
    f.tokens = f.lexer.ScanTokens(source)

    // Rewrite tokens in canonical format
    for i := 0; i < len(f.tokens); i++ {
        token := f.tokens[i]

        switch token.Type {
        case TOKEN_LBRACE:
            // Convert { to INDENT
            f.output.WriteString("\n")
            f.indentLevel++
            f.needsIndent = true
            // Skip the brace

        case TOKEN_RBRACE:
            // Convert } to DEDENT
            f.indentLevel--
            f.needsIndent = true
            // Skip the brace

        case TOKEN_SEMICOLON:
            // Remove semicolons
            f.output.WriteString("\n")
            f.needsIndent = true

        case TOKEN_DOUBLE_EQUALS:
            // Convert == to equals
            if f.needsIndent {
                f.writeIndent()
                f.needsIndent = false
            }
            f.output.WriteString(" equals ")

        case TOKEN_AND_AND:
            // Convert && to and
            f.output.WriteString(" and ")

        case TOKEN_OR_OR:
            // Convert || to or
            f.output.WriteString(" or ")

        case TOKEN_BANG:
            // Convert ! to not
            f.output.WriteString("not ")

        case TOKEN_NEWLINE:
            f.output.WriteString("\n")
            f.needsIndent = true

        default:
            // Write token as-is
            if f.needsIndent && token.Type != TOKEN_NEWLINE {
                f.writeIndent()
                f.needsIndent = false
            }
            f.output.WriteString(token.Lexeme)
        }
    }

    return f.output.String(), nil
}

func (f *Formatter) writeIndent() {
    for i := 0; i < f.indentLevel; i++ {
        f.output.WriteString("    ") // 4 spaces
    }
}
```

### Formatting Rules

1. **Braces to Indentation**
   ```go
   // Input (Go-style)
   if count == 5 {
       print("five")
   }

   // Output (Kukicha canonical)
   if count equals 5
       print "five"
   ```

2. **Semicolons Removed**
   ```go
   // Input
   x := 5;
   y := 10;

   // Output
   x := 5
   y := 10
   ```

3. **Operators Normalized**
   ```go
   // Input
   if x == 5 && y != 10 || !z

   // Output
   if x equals 5 and y not equals 10 or not z
   ```

4. **Indentation Enforced**
   - 4 spaces per level (strict)
   - Tabs converted to spaces
   - Trailing whitespace removed

### CLI Usage

```bash
# Format a single file
kuki fmt myfile.kuki

# Format all files in directory recursively
kuki fmt ./src/

# Check formatting without modifying (exit 1 if not formatted)
kuki fmt --check ./src/

# Format and write back to file
kuki fmt -w myfile.kuki

# Format all .kuki files in project
kuki fmt -w $(find . -name "*.kuki")
```

### Integration with Build Pipeline

```bash
# Pre-commit hook
#!/bin/sh
kuki fmt --check $(git diff --cached --name-only --diff-filter=ACM | grep '\.kuki$')
if [ $? -ne 0 ]; then
    echo "Error: Code not formatted. Run 'kuki fmt -w .' to fix."
    exit 1
fi

# CI/CD pipeline
- name: Check Kukicha Formatting
  run: kuki fmt --check ./src/
```

### Formatting Guarantees

**After running `kuki fmt`:**
- ✅ All code uses 4-space indentation
- ✅ No braces `{}` (converted to indentation)
- ✅ No semicolons `;`
- ✅ Operators use English keywords (`equals`, `and`, `or`, `not`)
- ✅ Consistent newlines and spacing
- ✅ Trailing whitespace removed
- ✅ Tabs converted to spaces

---

## Project Structure

```
kukicha/
├── cmd/
│   └── kukicha/
│       └── main.go              # CLI entry point
├── internal/
│   ├── lexer/
│   │   ├── lexer.go             # Lexer implementation
│   │   ├── token.go             # Token types
│   │   └── lexer_test.go
│   ├── parser/
│   │   ├── parser.go            # Parser implementation
│   │   ├── ast.go               # AST node definitions
│   │   └── parser_test.go
│   ├── semantic/
│   │   ├── analyzer.go          # Semantic analysis
│   │   ├── symbol.go            # Symbol table
│   │   ├── types.go             # Type system
│   │   └── semantic_test.go
│   ├── codegen/
│   │   ├── generator.go         # Code generation
│   │   └── codegen_test.go
│   ├── compiler/
│   │   └── compiler.go          # Orchestrates all phases
│   └── errors/
│       └── errors.go            # Error formatting
├── pkg/
│   └── kukicha/
│       └── kukicha.go           # Public API
├── stdlib/                       # Kukicha standard library (Phase 2)
│   ├── http/
│   ├── json/
│   ├── file/
│   ├── docker/
│   ├── k8s/
│   └── llm/
├── examples/                     # Example programs
│   ├── hello.kuki
│   ├── todo.kuki
│   └── github-cli.kuki
├── testdata/                     # Test fixtures
│   ├── valid/
│   └── invalid/
├── docs/                         # Documentation
│   ├── spec.md
│   ├── grammar.ebnf
│   └── stdlib/
├── go.mod
├── go.sum
├── README.md
└── LICENSE
```

---

## CLI Design

```go
// cmd/kukicha/main.go
package main

func main() {
    app := &cli.App{
        Name:  "kukicha",
        Usage: "The Kukicha programming language compiler",
        Commands: []*cli.Command{
            {
                Name:  "build",
                Usage: "Compile a .kuki file",
                Flags: []cli.Flag{
                    &cli.StringFlag{
                        Name:  "output",
                        Alias: "o",
                        Usage: "Output binary name",
                    },
                    &cli.BoolFlag{
                        Name:  "experiment",
                        Usage: "Enable Green Tea GC",
                        Value: true,
                    },
                },
                Action: buildCommand,
            },
            {
                Name:   "run",
                Usage:  "Compile and run a .kuki file",
                Action: runCommand,
            },
            {
                Name:   "fmt",
                Usage:  "Format .kuki files",
                Action: fmtCommand,
            },
            {
                Name:   "test",
                Usage:  "Run tests",
                Action: testCommand,
            },
        },
    }
    
    app.Run(os.Args)
}

func buildCommand(c *cli.Context) error {
    if c.NArg() == 0 {
        return fmt.Errorf("no input file")
    }
    
    inputFile := c.Args().Get(0)
    
    // Read source
    source, err := os.ReadFile(inputFile)
    if err != nil {
        return err
    }
    
    // Compile
    compiler := compiler.NewCompiler()
    goCode, err := compiler.Compile(string(source), inputFile)
    if err != nil {
        return err
    }
    
    // Write Go file
    goFile := strings.TrimSuffix(inputFile, ".kuki") + ".go"
    if err := os.WriteFile(goFile, []byte(goCode), 0644); err != nil {
        return err
    }
    
    // Run go build
    output := c.String("output")
    if output == "" {
        output = strings.TrimSuffix(inputFile, ".kuki")
    }
    
    cmd := exec.Command("go", "build", "-o", output, goFile)
    if c.Bool("experiment") {
        cmd.Env = append(os.Environ(), "GOEXPERIMENT=greenteagc")
    }
    
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    
    return cmd.Run()
}
```

---

## Testing Strategy

### Unit Tests

```go
// internal/lexer/lexer_test.go
func TestLexer(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected []TokenType
    }{
        {
            name:  "simple function",
            input: "func Hello()\n    print \"hi\"\n",
            expected: []TokenType{
                TOKEN_FUNC, TOKEN_IDENTIFIER, TOKEN_LPAREN, TOKEN_RPAREN, TOKEN_NEWLINE,
                TOKEN_INDENT, TOKEN_IDENTIFIER, TOKEN_STRING, TOKEN_NEWLINE,
                TOKEN_DEDENT, TOKEN_EOF,
            },
        },
        // ... more tests
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            lexer := NewLexer(tt.input, "test.kuki")
            tokens, err := lexer.ScanTokens()
            
            require.NoError(t, err)
            require.Equal(t, len(tt.expected), len(tokens))
            
            for i, expected := range tt.expected {
                assert.Equal(t, expected, tokens[i].Type)
            }
        })
    }
}
```

### Integration Tests

```go
// internal/compiler/compiler_test.go
func TestCompiler(t *testing.T) {
    tests := []struct {
        name     string
        kuki     string
        expected string // Expected Go output
    }{
        {
            name: "hello world",
            kuki: `leaf main
func main()
    print "Hello, World!"
`,
            expected: `package main

func main() {
    fmt.Println("Hello, World!")
}
`,
        },
        // ... more tests
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            compiler := NewCompiler()
            output, err := compiler.Compile(tt.kuki, "test.kuki")
            
            require.NoError(t, err)
            assert.Equal(t, normalize(tt.expected), normalize(output))
        })
    }
}
```

### Golden Tests

```
testdata/
├── valid/
│   ├── hello.kuki
│   ├── hello.go.golden
│   ├── todo.kuki
│   └── todo.go.golden
└── invalid/
    ├── bad-indent.kuki
    └── bad-indent.error.golden
```

---

## Error Messages

### Good Error Format

```
Error in main.kuki:12:10

  10 │ func Process(data)
  11 │     result := calculate()
  12 │     count := "hello"
     │              ^^^^^^^
     │ 
     │ Type mismatch: expected int, got string
     │ 
     │ Help: count is declared as int on line 8
     │       If you want to convert, use: int64(value)
```

### Error Reporter

```go
type ErrorReporter struct {
    file   string
    source []string
}

func (er *ErrorReporter) Report(pos Position, message string, help string) {
    fmt.Fprintf(os.Stderr, "Error in %s:%d:%d\n\n", pos.File, pos.Line, pos.Column)
    
    // Show context
    start := max(0, pos.Line-2)
    end := min(len(er.source), pos.Line+1)
    
    for i := start; i < end; i++ {
        prefix := "  "
        if i == pos.Line-1 {
            prefix = "→ "
        }
        fmt.Fprintf(os.Stderr, "%s%3d │ %s\n", prefix, i+1, er.source[i])
        
        if i == pos.Line-1 {
            // Show pointer
            spaces := strings.Repeat(" ", pos.Column+6)
            arrows := strings.Repeat("^", 7)
            fmt.Fprintf(os.Stderr, "%s%s\n", spaces, arrows)
        }
    }
    
    fmt.Fprintf(os.Stderr, "\n│ %s\n", message)
    
    if help != "" {
        fmt.Fprintf(os.Stderr, "│ \n│ Help: %s\n", help)
    }
    
    fmt.Fprintln(os.Stderr)
}
```

---

## Implementation Status

### Completed Features (v1.0.0)

- ✅ **Lexer** - Full tokenization with indentation support
- ✅ **Parser** - Complete AST generation
- ✅ **Semantic Analysis** - Type checking and validation
- ✅ **Code Generation** - Idiomatic Go transpilation
- ✅ **CLI Tool** - Build, run, and transpile commands
- ✅ **Test Suite** - Comprehensive tests for all phases

### Future Enhancements

See the [Standard Library Roadmap](kukicha-stdlib-roadmap.md) for planned features and extensions.

For the complete language specification, see:
- [Language Syntax Reference](kukicha-syntax-v1.0.md)
- [Quick Reference](kukicha-quick-reference.md)
- [Grammar (EBNF)](kukicha-grammar.ebnf.md)
