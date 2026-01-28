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
    Petiole     *PetioleDecl   // Optional: nil if not declared
    FilePath    string         // Required: used to calculate implicit Petiole
    StemRoot    string         // Required: path to stem.toml directory
    Imports     []*ImportDecl
    Declarations []Declaration
}

// CalculatePetiole computes the package name from file path if Petiole is nil
func (p *Program) CalculatePetiole() string {
    if p.Petiole != nil {
        return p.Petiole.Name  // Explicit declaration takes precedence
    }

    // Calculate petiole from file path relative to stem.toml
    relPath := filepath.Rel(p.StemRoot, filepath.Dir(p.FilePath))
    // Convert path to package name (e.g., "src/auth" -> "auth")
    petiole := filepath.Base(relPath)
    return petiole
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
    OnErr *OnErrClause // Optional onerr clause
    Pos   Position
}

type AssignStmt struct {
    Target Expr
    Value  Expr
    OnErr  *OnErrClause // Optional onerr clause
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

// OnErrClause is not an AST node — it's a helper struct attached to statements
type OnErrClause struct {
    Token   Token      // The 'onerr' token
    Handler Expression // Error handler (panic, error, empty, discard, or default value)
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

type TypeAssertionExpr struct {
    Expression Expr
    TargetType TypeExpr
    Pos        Position
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
    
    // Parse petiole declaration
    program.Petiole = p.parsePetioleDecl()
    
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
    case *PipeExpr:
        return sa.checkPipeExpr(e)
    case *CallExpr:
        return sa.checkCallExpr(e)
    // ... more cases
    }

    return nil
}

// analyzeOnErrClause analyzes the handler expression in an OnErr clause.
// Called from statement analyzers (VarDeclStmt, AssignStmt, ExpressionStmt).
func (sa *SemanticAnalyzer) analyzeOnErrClause(clause *OnErrClause) {
    if clause != nil {
        sa.checkExpr(clause.Handler)
    }
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
Convert AST to idiomatic Go code, with special handling for stdlib generic type parameters.

### Stdlib Generic Type Parameters

Kukicha transparently generates Go 1.25+ generic type parameters for stdlib functions without requiring users to write generic syntax.

#### How It Works

1. **Automatic Detection**: When generating code in `stdlib/iterator/` or `stdlib/slice/`, the codegen detects functions that need type parameters
2. **Type Inference**: Uses placeholder names in function signatures (`"any"` → `"T"`, `"any2"` → `"K"`) to infer generic structure
3. **Constraint Application**: Functions like `GroupBy` get proper constraints (`comparable` for map keys)
4. **Transparent to Users**: End-users write normal Kukicha code with explicit parameter types; generics are added automatically

#### Example: GroupBy

**User Code (Kukicha):**
```kukicha
errors := logs
    |> slice.GroupBy(func(e LogEntry) string {
        return e.Level
    })
```

**Generated Code (Go 1.25+):**
```go
errors := slice.GroupBy(logs, func(e LogEntry) string {
    return e.Level
})

// Where GroupBy is generated as:
func GroupBy[T any, K comparable](items []T, keyFunc func(T) K) map[K][]T {
    result := make(map[K][]T)
    for item := range items {
        key := keyFunc(item)
        result[key] = append(result[key], item)
    }
    return result
}
```

#### Code Generation Flow for Stdlib Generics

```
Stdlib Function AST
    ↓
Detect stdlib/slice or stdlib/iterator context
    ↓
inferSliceTypeParameters() or inferStdlibTypeParameters()
    ↓
Build TypeParameter mapping: "any" → "T", "any2" → "K", etc.
    ↓
generateTypeParameters([T any, K comparable])
    ↓
Apply placeholder substitution recursively through types
    ↓
Generate final Go signature with generics
```

#### Supported Generic Functions

| Package | Function | Type Parameters | Constraint |
|---------|----------|-----------------|------------|
| **iteatorr** | Map | `[T any, U any]` | None |
| **iterator** | FlatMap | `[T any, U any]` | None |
| **iterator** | Filter | `[T any]` | None |
| **iterator** | Enumerate | `[T any]` (returns `Seq2[int, T]`) | None |
| **slice** | GroupBy | `[T any, K comparable]` | K must be comparable for map keys |

#### Future Extension

The same mechanism can be extended to other stdlib/slice functions or new packages. The key is detecting the function at code generation time and inferring the type parameters from signature patterns.

### Code Generator Structure

```go
type Generator struct {
    program              *ast.Program
    output               strings.Builder
    indent               int
    placeholderMap       map[string]string // Maps placeholder names to type param names (e.g., "element" → "T")
    autoImports          map[string]bool   // Tracks auto-imports needed (e.g., "cmp" for generic constraints)
    isStdlibIter         bool              // True if generating stdlib/iterator code (enables special transpilation)
    sourceFile           string            // Source file path for detecting stdlib
    currentFuncName      string            // Current function being generated (for context-aware decisions)
    processingReturnType bool              // Whether we are currently generating return types
    tempCounter          int               // Counter for generating unique temporary variable names
}

func New(program *ast.Program) *Generator {
    return &Generator{
        program:     program,
        indent:      0,
        autoImports: make(map[string]bool),
    }
}

// SetSourceFile sets the source file path and detects if special transpilation is needed
func (g *Generator) SetSourceFile(path string) {
    g.sourceFile = path
    // Enable special transpilation for stdlib/iterator files
    g.isStdlibIter = strings.Contains(path, "stdlib/iterator/")
    // Note: stdlib/slice uses a different approach - type parameters are detected per-function
}

func (g *Generator) Generate() (string, error) {
    g.output.Reset()

    // Pre-scan declarations to collect auto-imports (e.g. net/http for fetch wrappers)
    g.scanForAutoImports()

    // Generate header comment
    g.writeLine("// Generated by Kukicha v1.0.0 (requires Go 1.25+)")
    g.writeLine("//")
    g.writeLine("// Performance options:")
    g.writeLine("//   GOEXPERIMENT=greenteagc  - Green Tea GC (10-40% faster)")
    g.writeLine("//   GOEXPERIMENT=jsonv2      - Faster JSON parsing (2-10x improvement)")
    g.writeLine("")

    // Generate package declaration
    g.generatePackage()

    // Generate imports (including auto-imports for generics like cmp)
    needsFmt := g.needsStringInterpolation() || g.needsPrintBuiltin()
    if len(g.program.Imports) > 0 || needsFmt || len(g.autoImports) > 0 {
        g.writeLine("")
        g.generateImports()
    }

    // Generate declarations
    for _, decl := range g.program.Declarations {
        g.writeLine("")
        g.generateDeclaration(decl)
    }

    return g.output.String(), nil
}

For detailed code generator implementation, see `internal/codegen/codegen.go`. Key generation methods:

- `generateFunctionDecl()` - Function generation with generic type parameter support
- `generateTypeAnnotation()` - Type annotation transpilation with placeholder substitution
- `generateTypeParameters()` - Generate Go type parameter list `[T any, K comparable]`
- `inferStdlibTypeParameters()` - Infer generics for stdlib/iter functions
- `inferSliceTypeParameters()` - Infer generics for stdlib/slice functions like GroupBy

**Generic Type Parameter Generation:**

When processing function declarations in stdlib packages, the code generator:
1. Detects the source file context (stdlib/iterator or stdlib/slice)
2. Calls the appropriate type inference function
3. Builds a placeholder mapping for type substitution
4. Recursively applies the mapping through all type annotations
5. Generates proper Go generic syntax in the output

**Typed Empty Handling for Generics:**
For expressions like `empty T` where `T` is a generic type parameter, the generator produces `*new(T)` instead of invalid `T{}` syntax, ensuring zero values work correctly with generics. This is primarily used in `stdlib/iterator` context.

This ensures that stdlib functions automatically receive proper Go 1.25+ generic type parameters without users needing to write generic syntax in their Kukicha code.

### Pipe Operator with Placeholder Strategy

The pipe operator (`|>`) supports two strategies for argument injection:

#### Strategy A: Default (Data First)
When no placeholder is present, the piped value becomes the first argument:
```kukicha
data |> transform(opts)
```
Generates:
```go
transform(data, opts)
```

**Special Case: Multi-Return Functions**
If the left side is a function call that returns multiple values (e.g., `fetch.Get()`), the generator automatically wraps it in an IIFE (Immediately Invoked Function Expression) to extract just the first return value for the pipe:
```kukicha
fetch.Get(url) |> fetch.CheckStatus()
```
Generates:
```go
fetch.CheckStatus(func() *http.Response { val, _ := fetch.Get(url); return val }())
```

#### Strategy B: Explicit Placeholder
When `_` appears in the argument list, the piped value replaces it:
```kukicha
todo |> json.MarshalWrite(writer, _)
data |> encode(opts, _, format)
```
Generates:
```go
json.MarshalWrite(writer, todo)
encode(opts, data, format)
```

#### Strategy C: Context-First
If the piped value is a context (variable named `ctx` or a `context.*` call), it is always prepended to the argument list, even without a placeholder. This allows fluent chaining of context-heavy operations:
```kukicha
ctx |> db.FetchUser(userID)
context.WithTimeout(ctx, 5s) |> db.Save(data)
```
Generates:
```go
db.FetchUser(ctx, userID)
ctx2, cancel := context.WithTimeout(ctx, 5*time.Second) 
db.Save(ctx2, data) // (Pseudo-code: assuming ctx piping)
```

#### Strategy D: Shorthand Method Pipe
Piping into a method starting with `.` (e.g., `|> .Output()`) automatically treats the left side as the receiver object:
```kukicha
ctx |> exec.CommandContext("ls") |> .Output()
```
Generates:
```go
exec.CommandContext(ctx, "ls").Output()
```

**Implementation:** `generatePipeExpr()` scans arguments for `_`. If found, Strategy B is used. If the right side is a shorthand method (starts with `.`), Strategy D is used. If the left side is detected as a context by `isContextExpr()`, Strategy C is used. Otherwise, Strategy A (Data First) is the fallback.

### Error Handling Code Generation (onerr)

The `onerr` clause is a **statement-level** construct attached to `VarDeclStmt`, `AssignStmt`, or `ExpressionStmt`. It is not an expression operator — this makes nested onerr structurally unrepresentable in the AST.

#### Pattern 1: Default Value (VarDeclStmt)
```kukicha
port := getPort() onerr "8080"
```
Generates:
```go
port, err_1 := getPort()
if err_1 != nil {
    port = "8080"
}
```

#### Pattern 2: Panic (VarDeclStmt)
```kukicha
config := loadConfig() onerr panic "config missing"
```
Generates:
```go
config, err_1 := loadConfig()
if err_1 != nil {
    panic("config missing")
}
```

#### Pattern 3: Discard (VarDeclStmt)
```kukicha
result := riskyOp() onerr discard
```
Generates (optimized - no error check needed):
```go
result, _ := riskyOp()
```

#### Pattern 4: Assignment (AssignStmt)
```kukicha
x = f() onerr panic "failed"
```
Generates:
```go
x, err_1 = f()
if err_1 != nil {
    panic("failed")
}
```

#### Pattern 5: Expression Statement (ExpressionStmt)
```kukicha
todo |> json.MarshalWrite(w, _) onerr panic "marshal failed"
```
Generates:
```go
if err_1 := json.MarshalWrite(w, todo); err_1 != nil {
    panic("marshal failed")
}
```

**Implementation Details:**
- `generateVarDeclStmt()` checks `stmt.OnErr != nil` and delegates to `generateOnErrVarDecl()`
- `generateAssignStmt()` checks `stmt.OnErr != nil` and delegates to `generateOnErrAssign()`
- `generateStatement()` checks `ExpressionStmt.OnErr != nil` and delegates to `generateOnErrStmt()`
- `uniqueId()` generates unique error variable names (`err_1`, `err_2`) to prevent shadowing
- `generateOnErrHandler()` generates appropriate handler code based on handler type (PanicExpr, ErrorExpr, ReturnExpr, EmptyExpr, or default value)
- `ReturnExpr` allows returning from the parent function directly from an `onerr` handler, supporting multiple return values: `onerr return empty, error "failed"`
- Pipe expressions are fully resolved before the onerr clause — no restructuring needed


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
