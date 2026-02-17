# Kukicha Language Grammar (EBNF)

**Version:** 0.0.1
**Notation:** Extended Backus-Naur Form (EBNF)

---

## Notation Guide

```
::=     Definition
|       Alternative
()      Grouping
[]      Optional (zero or one)
{}      Repetition (zero or more)
+       One or more
?       Optional
"text"  Terminal (literal)
UPPER   Non-terminal
```

---

## Program Structure

```ebnf
Program ::= [ PetioleDeclaration ] { ImportDeclaration } { TopLevelDeclaration }

PetioleDeclaration ::= "petiole" IDENTIFIER NEWLINE
    # Optional: if absent, package name is calculated from file path relative to stem.toml

ImportDeclaration ::= "import" STRING [ "as" IDENTIFIER ] NEWLINE
```

---

## Top-Level Declarations

```ebnf
TopLevelDeclaration ::=
    | TypeDeclaration
    | InterfaceDeclaration
    | FunctionDeclaration
    | MethodDeclaration

TypeDeclaration ::=
    | "type" IDENTIFIER NEWLINE INDENT FieldList DEDENT
    | "type" IDENTIFIER FunctionType NEWLINE

FieldList ::= Field { Field }

Field ::= IDENTIFIER TypeAnnotation [ FieldAlias ] { StructTag } NEWLINE

FieldAlias ::= "as" StringLiteral
    # Sugar for JSON mapping, e.g., Stars int as "stargazers_count"
    # Equivalent generated Go tag: `json:"stargazers_count"`
    # Constraint: do not combine FieldAlias with explicit StructTag on the same field

StructTag ::= IDENTIFIER ":" StringLiteral
    # e.g., json:"id" or db:"user_name"

InterfaceDeclaration ::= "interface" IDENTIFIER NEWLINE INDENT MethodSignatureList DEDENT

MethodSignatureList ::= MethodSignature { MethodSignature }

MethodSignature ::= IDENTIFIER "(" [ ParameterList ] ")" [ TypeAnnotation ] NEWLINE

FunctionDeclaration ::=
    "func" IDENTIFIER [ "(" [ ParameterList ] ")" ] [ ReturnTypeList ] NEWLINE
    INDENT StatementList DEDENT
    # Return types are optional, but required for functions that return values

MethodDeclaration ::=
    # Kukicha syntax - explicit receiver name
    "func" IDENTIFIER "on" IDENTIFIER TypeAnnotation [ "," ParameterList | "(" [ ParameterList ] ")" ] [ ReturnTypeList ] NEWLINE
    INDENT StatementList DEDENT
    # Additional params may be comma-separated after receiver type (no parens needed):
    #   func Load on cfg Config, path string

ParameterList ::= Parameter { "," Parameter }

Parameter ::= [ "many" ] IDENTIFIER [ TypeAnnotation ] [ "=" Expression ]
    # Type annotation is required except for untyped variadic ("many x")
    # Optional default value (parameters with defaults must come after those without)
    # Examples:
    #   name string                     # required parameter
    #   count int = 10                  # parameter with default value
    #   many values                     # variadic (no default allowed)

ReturnTypeList ::= TypeAnnotation | "(" TypeAnnotation { "," TypeAnnotation } ")"
```

---

## Type Annotations

**Context-Sensitive Keywords**: The keywords `list`, `map`, and `channel` are context-sensitive.
- In **type annotation context** (function parameters, struct fields, variable type hints), they begin composite types.
- In **expression context**, they may be used as identifiers (though this is discouraged for clarity).

The parser determines context based on position. Type annotations appear after:
- Parameter names in function signatures
- Field names in struct definitions
- The `reference` keyword
- The `as` keyword in type casts
- The `:=` operator when followed by a type constructor

```ebnf
TypeAnnotation ::=
    | PrimitiveType
    | ReferenceType
    | ListType
    | MapType
    | ChannelType
    | FunctionType
    | QualifiedType

PrimitiveType ::=
    | "int" | "int8" | "int16" | "int32" | "int64"
    | "uint" | "uint8" | "uint16" | "uint32" | "uint64"
    | "float32" | "float64"
    | "string" | "bool" | "byte" | "rune" | "error"

ReferenceType ::= "reference" ( TypeAnnotation | "to" TypeAnnotation )

ListType ::= "list" "of" TypeAnnotation

MapType ::= "map" "of" TypeAnnotation "to" TypeAnnotation

ChannelType ::= "channel" "of" TypeAnnotation

QualifiedType ::= IDENTIFIER "." IDENTIFIER

FunctionType ::= "func" "(" [ FunctionTypeParameterList ] ")" [ TypeAnnotation ]

FunctionTypeParameterList ::= TypeAnnotation { "," TypeAnnotation }
```

**Parser Implementation Note**: When in type annotation context, if the parser sees `list`, `map`, or `channel`, it MUST be followed by `of`. This is not ambiguous because the parser knows when it expects a type.

---

## Statements

```ebnf
StatementList ::= Statement { Statement }

Statement ::=
    | VariableDeclaration
    | Assignment
    | IncDecStatement
    | ReturnStatement
    | IfStatement
    | SwitchStatement
    | ForStatement
    | DeferStatement
    | GoStatement
    | SendStatement
    | PrintStatement
    | ContinueStatement
    | BreakStatement
    | ExpressionStatement
    | NEWLINE

PrintStatement ::= "print" ExpressionList NEWLINE
    # 'print' is a built-in keyword that transpiles to fmt.Println()

VariableDeclaration ::= IdentifierList ":=" ExpressionList [ OnErrClause ] StatementTerminator

Assignment ::=
    | IdentifierList "=" ExpressionList [ OnErrClause ] StatementTerminator
    | Expression "=" ExpressionList [ OnErrClause ] StatementTerminator

IncDecStatement ::= Expression ("++" | "--") StatementTerminator

ReturnStatement ::= "return" [ ExpressionList ] NEWLINE

ContinueStatement ::= "continue" NEWLINE

BreakStatement ::= "break" NEWLINE

IfStatement ::=
    "if" [ SimpleStatement ";" ] Expression NEWLINE
    INDENT StatementList DEDENT
    [ ElseClause ]

ElseClause ::=
    | "else" NEWLINE INDENT StatementList DEDENT
    | "else" IfStatement

SwitchStatement ::=
    | RegularSwitchStatement
    | TypeSwitchStatement

RegularSwitchStatement ::=
    "switch" [ Expression ] NEWLINE
    INDENT { WhenClause } [ OtherwiseClause ] DEDENT

TypeSwitchStatement ::=
    "switch" Expression "as" IDENTIFIER NEWLINE
    INDENT { TypeWhenClause } [ OtherwiseClause ] DEDENT

WhenClause ::=
    "when" Expression { "," Expression } NEWLINE
    INDENT StatementList DEDENT

TypeWhenClause ::=
    "when" TypeAnnotation NEWLINE
    INDENT StatementList DEDENT

OtherwiseClause ::=
    ( "otherwise" | "default" ) NEWLINE
    INDENT StatementList DEDENT

ForStatement ::=
    | ForBareLoop
    | ForRangeLoop
    | ForCollectionLoop
    | ForNumericLoop
    | ForConditionLoop

ForBareLoop ::=
    "for" NEWLINE
    INDENT StatementList DEDENT

ForRangeLoop ::=
    "for" IDENTIFIER "from" Expression ( "to" | "through" ) Expression NEWLINE
    INDENT StatementList DEDENT

ForCollectionLoop ::=
    "for" [ IDENTIFIER "," ] IDENTIFIER "in" Expression NEWLINE
    INDENT StatementList DEDENT

ForNumericLoop ::=
    "for" IDENTIFIER "from" Expression ( "to" | "through" ) Expression NEWLINE
    INDENT StatementList DEDENT

ForConditionLoop ::=
    "for" Expression NEWLINE
    INDENT StatementList DEDENT

DeferStatement ::=
    | "defer" Expression NEWLINE
    | "defer" NEWLINE INDENT StatementList DEDENT

GoStatement ::= "go" ( Expression | NEWLINE INDENT StatementList DEDENT ) NEWLINE

SendStatement ::= "send" Expression "," Expression NEWLINE

ExpressionStatement ::= Expression [ OnErrClause ] StatementTerminator

OnErrClause ::= "onerr" ( Expression | NEWLINE INDENT StatementList DEDENT ) [ "explain" STRING ]
    # Single expression: onerr panic "failed"
    # Block form:
    #   onerr
    #       log.Printf("Error: {err}")
    #       return
    # With explain hint:
    #   onerr explain "hint message"           # Standalone: wraps error, returns
    #   onerr "default" explain "hint message" # With handler: wraps error, then runs handler

SimpleStatement ::=
    | IdentifierList ":=" ExpressionList [ OnErrClause ]
    | IdentifierList "=" ExpressionList [ OnErrClause ]
    | Expression "=" ExpressionList [ OnErrClause ]
    | Expression [ OnErrClause ]

StatementTerminator ::= NEWLINE | ";"

IdentifierList ::= IDENTIFIER { "," IDENTIFIER }
```

---

## Expressions

```ebnf
Expression ::= OrExpression

OrExpression ::= PipeExpression { "or" PipeExpression }

PipeExpression ::= AndExpression { "|>" AndExpression }

AndExpression ::= BitwiseOrExpression { "and" BitwiseOrExpression }

BitwiseOrExpression ::= ComparisonExpression { "|" ComparisonExpression }

ComparisonExpression ::= AdditiveExpression [ ComparisonOp AdditiveExpression | "in" AdditiveExpression | "not" "in" AdditiveExpression ]

ComparisonOp ::=
    | "==" | "!=" | "equals" | "not" "equals"
    | ">" | "<" | ">=" | "<="

AdditiveExpression ::= MultiplicativeExpression { ( "+" | "-" ) MultiplicativeExpression }

MultiplicativeExpression ::= UnaryExpression { ( "*" | "/" | "%" ) UnaryExpression }

UnaryExpression ::=
    | ( "not" | "!" | "-" ) UnaryExpression
    | "reference" "of" UnaryExpression
    | "dereference" UnaryExpression
    | PostfixExpression

PostfixExpression ::=
    PrimaryExpression {
        | "." IDENTIFIER
        | "(" [ ArgumentList ] ")"
        | "[" Expression "]"
        | "[" [ Expression ] ":" [ Expression ] "]"
    }

PrimaryExpression ::=
    | IDENTIFIER
    | Literal
    | "(" Expression ")"
    | StructLiteral
    | ListLiteral
    | TypedListLiteral
    | EmptyLiteral       # 'empty' with optional type (uses 1-token lookahead)
    | MakeExpression
    | CloseExpression
    | PanicExpression
    | RecoverExpression
    | ReceiveExpression
    | ErrorExpression
    | DiscardExpression
    | TypeCast
    | TypeAssertionExpression
    | FunctionLiteral
    | ArrowLambda
    | ReturnExpression
    | ShorthandMethodCall

ExpressionList ::= Expression { "," Expression }

ArgumentList ::= Argument { "," Argument }
Argument ::= [ "many" ] ( NamedArgument | Expression )
NamedArgument ::= IDENTIFIER ":" Expression
    # Named arguments allow explicit parameter binding: foo(name: "value", count: 5)
    # Named arguments must come after positional arguments
    # Named arguments can appear in any order relative to each other
```

---

## Literals

```ebnf
Literal ::=
    | IntegerLiteral
    | FloatLiteral
    | StringLiteral
    | RuneLiteral
    | BooleanLiteral

IntegerLiteral ::= DIGIT { DIGIT }

FloatLiteral ::= DIGIT { DIGIT } "." DIGIT { DIGIT }

StringLiteral ::= '"' { StringChar | Interpolation } '"'

StringChar ::= /* any character except ", newline, or { */

RuneChar ::= /* any character except ', newline, or escape */

Interpolation ::= "{" Expression "}"

BooleanLiteral ::= "true" | "false"

StructLiteral ::=
    | TypeAnnotation "{" [ FieldInitList ] "}"
    | TypeAnnotation NEWLINE INDENT FieldInitBlock DEDENT

FieldInitList ::= FieldInit { "," FieldInit }

FieldInit ::= IDENTIFIER ":" Expression

FieldInitBlock ::= FieldInitLine { FieldInitLine }

FieldInitLine ::= IDENTIFIER ":" Expression NEWLINE
    # Indentation-based struct literal:
    #   todo := Todo
    #       id: 1
    #       title: "Learn Kukicha"
    #       completed: false

# EmptyLiteral uses 1-token lookahead after 'empty' to determine the type.
# If 'empty' is followed by 'list', 'map', 'channel', or 'reference', parse as typed empty.
# Otherwise, 'empty' is a standalone nil/zero-value literal.
EmptyLiteral ::=
    | "empty" "list" "of" TypeAnnotation          # empty list of Todo
    | "empty" "map" "of" TypeAnnotation "to" TypeAnnotation  # empty map of string to int
    | "empty" "channel" "of" TypeAnnotation       # empty channel of Result
    | "empty" "reference" TypeAnnotation          # empty reference User (nil pointer)
    | "empty"                                      # standalone nil/zero-value

# Non-empty list literal (list with initial values)
ListLiteral ::= "[" [ ExpressionList ] "]"

# Typed list literal with explicit element type
TypedListLiteral ::= "list" "of" TypeAnnotation "{" [ ExpressionList ] "}"
    # e.g., list of int{1, 2, 3} or list of Todo{}

MakeExpression ::=
    | "make" "(" TypeAnnotation [ "," ExpressionList ] ")"
    | "make" TypeAnnotation [ "," ExpressionList ]
    # Both forms valid: make(channel of string) or make channel of string, 100

ReceiveExpression ::= "receive" "from" Expression

RecoverExpression ::= "recover" "(" ")"

TypeCast ::= Expression "as" TypeAnnotation

TypeAssertionExpression ::= "." "(" TypeAnnotation ")"
    # Postfix usage: value.(Type)

ReturnExpression ::= "return" [ ExpressionList ]

ShorthandMethodCall ::= "." IDENTIFIER [ "(" [ ArgumentList ] ")" ]

FunctionLiteral ::= "func" "(" [ ParameterList ] ")" [ ReturnTypeList ] NEWLINE INDENT StatementList DEDENT

ArrowLambda ::=
    | "(" [ ParameterList ] ")" "=>" LambdaBody
    | IDENTIFIER "=>" LambdaBody

LambdaBody ::=
    | Expression
    | NEWLINE INDENT StatementList DEDENT

ErrorExpression ::= "error" Expression
DiscardExpression ::= "discard"
CloseExpression ::= "close" Expression
PanicExpression ::= "panic" Expression
RuneLiteral ::= "'" RuneChar "'"
```

---

## Lexical Elements

```ebnf
IDENTIFIER ::= LETTER { LETTER | DIGIT }

LETTER ::= "a" | "b" | ... | "z" | "A" | "B" | ... | "Z"

DIGIT ::= "0" | "1" | "2" | "3" | "4" | "5" | "6" | "7" | "8" | "9"

DOMAIN ::= IDENTIFIER { "." IDENTIFIER }

PATH ::= IDENTIFIER { "/" IDENTIFIER }

NUMBER ::= DIGIT { DIGIT }

NEWLINE ::= "\n" | "\r\n"

INDENT ::= /* Increase in indentation level */

DEDENT ::= /* Decrease in indentation level */

COMMENT ::= "#" { any character except NEWLINE } NEWLINE
```

---

## Keywords (Reserved)

```
petiole      import      type        interface   func
if          else        for         in          from
to          through     of          and
or          onerr       not         return      go
defer       make        list        map         channel
send        receive     close       panic       recover
error       empty       nil         reference   dereference
on          discard     true        false       equals
as          many        continue    break
switch      when        otherwise   default
```

**Note:** The keywords `list`, `map`, and `channel` are context-sensitive and may also be used as identifiers in certain contexts.

---

## Operators and Delimiters

```
+     -     *     /     %
==    !=    <     <=    >     >=
!     and   or    not
|     |>    =>    ++    --
:=    =     :     .     ,     ;
(     )     [     ]     {     }
```

---

## Special Handling

### Indentation Sensitivity

Kukicha uses significant whitespace (like Python). The lexer must:
1. Track indentation level at start of each line
2. Generate `INDENT` token when indentation increases
3. Generate `DEDENT` token when indentation decreases
4. **Use 4 spaces per indentation level (tabs are rejected)**
5. Indentation must be consistent (multiples of 4 spaces)

**Indentation Rules:**
- Each indentation level = 4 spaces
- Tabs are not allowed (lexer error)
- Mixed spaces/tabs within a file = error
- Indentation must increase/decrease by 4 spaces at a time

**Example:**
```kukicha
func Process()
····if condition        # 4 spaces (1 level)
········doSomething()   # 8 spaces (2 levels)
····else                # 4 spaces (back to 1 level)
········doOther()       # 8 spaces (2 levels)
```

**Lexer Error for Tabs:**
```
Error in main.kuki:5:1

    5 |→→if condition
      |^^ Use 4 spaces for indentation, not tabs

Help: Configure your editor to use spaces.
      VSCode: Set "editor.insertSpaces": true
```

### String Interpolation

String literals with `{}` must be processed to extract:
1. Literal string segments
2. Expression segments (inside `{}`)

Example:
```kukicha
"Hello {name}, you have {count} messages"
```

Parsed as:
- Literal: "Hello "
- Expression: `name`
- Literal: ", you have "
- Expression: `count`
- Literal: " messages"

### OnErr Clause (Statement-Level Error Handling)

The `onerr` clause provides ergonomic error handling for functions that return `(T, error)` tuples. It attaches to `VarDeclStmt`, `AssignStmt`, or `ExpressionStmt` — it is **not** an expression operator.

1. Automatically unwrap to `T` if no error
2. Execute the `onerr` handler if error is not empty

**Important:** The `onerr` keyword is distinct from the boolean `or` operator. This separation makes code more readable - you can tell at a glance whether a statement handles errors or an expression performs boolean logic.

Example:
```kukicha
# Error handling with onerr (statement-level clause)
data := file.read(path) onerr panic "failed"

# Boolean logic with or (expression-level operator)
if active or pending
    process()
```

The `onerr` clause desugars to:
```kukicha
data, err := file.read(path)
if err != empty
    panic "failed"
```

### Discard Keyword

The `discard` keyword is syntactic sugar for Go's `_` (blank identifier):
- Can appear in tuple unpacking
- Cannot be referenced as a variable

### Negative Indexing

Kukicha supports negative indices for accessing elements from the end of a collection.

**Single element access:**
```kukicha
# Source
last := items[-1]
secondLast := items[-2]

# Generates Go
last := items[len(items)-1]
secondLast := items[len(items)-2]
```

**Slicing with negative indices:**
```kukicha
# Source
lastThree := items[-3:]
allButLast := items[:-1]
middle := items[1:-1]

# Generates Go
lastThree := items[len(items)-3:]
allButLast := items[:len(items)-1]
middle := items[1:len(items)-1]
```

**How it works:**
- The parser recognizes negative numbers as `UnaryExpression` with `-` operator
- The code generator detects negative indices and transforms them to `len(collection) - N`

### Pipe Operator

The pipe operator `|>` passes the left-hand side as the first argument to the right-hand side function call.

**Desugaring rule:**
```kukicha
# Source
a |> f() |> g(x, y)

# Desugars to
g(f(a), x, y)
```

**Multiple arguments:**
```kukicha
# Source
data |> process(option1, option2)

# Desugars to
process(data, option1, option2)
```

**With method calls:**
```kukicha
# Source
response |> .json() |> filterActive()

# Desugars to
filterActive(response.json())
```

**Precedence:**
- Pipe has lower precedence than arithmetic/comparison operators
- `onerr` is a statement-level clause, not an expression operator

```kukicha
# Example: pipe with onerr clause on variable declaration
result := a + b |> double() onerr "default"

# The pipe resolves to: double(a + b)
# The onerr clause attaches to the VarDeclStmt, not the expression
```

---

## Grammar Production Examples

### Example 1: Simple Function

```kukicha
func Greet(name string)
    print "Hello {name}"
```

Parse tree:
```
FunctionDeclaration
├─ func
├─ Greet
├─ ParameterList
│  └─ Parameter
│     ├─ name
│     └─ string
└─ StatementList
   └─ ExpressionStatement
      └─ FunctionCall
         ├─ print
         └─ StringLiteral: "Hello {name}"
            ├─ "Hello "
            ├─ Interpolation: name
            └─ ""
```

### Example 2: Method with OnErr Clause

```kukicha
func Load on cfg Config, path string
    content := file.read(path) onerr return error "cannot read"
    cfg.data = json.parse(content) onerr return error "invalid json"
    return empty
```

Parse tree:
```
MethodDeclaration
├─ func
├─ Load
├─ on
├─ cfg          # explicit receiver name
├─ Config       # receiver type
├─ ParameterList
│  └─ Parameter
│     ├─ path
│     └─ string
└─ StatementList
   ├─ VariableDeclaration
   │  ├─ content
   │  ├─ :=
   │  ├─ FunctionCall: file.read(path)
   │  └─ OnErrClause
   │     └─ ReturnStatement: return error "cannot read"
   ├─ Assignment
   │  ├─ cfg.data
   │  ├─ =
   │  ├─ FunctionCall: json.parse(content)
   │  └─ OnErrClause
   │     └─ ReturnStatement: return error "invalid json"
   └─ ReturnStatement
      └─ empty
```

### Example 3: Concurrent Processing

```kukicha
func ProcessAll(items list of Item)
    results := make channel of Result, len(items)
    
    for discard, item in items
        go
            result := process(item)
            send results, result
    
    for i from 0 to len(items)
        result := receive from results
        print result
```

Parse tree:
```
FunctionDeclaration
├─ func
├─ ProcessAll
├─ ParameterList
│  └─ Parameter
│     ├─ items
│     └─ ListType
│        └─ Item
└─ StatementList
   ├─ VariableDeclaration
   │  ├─ results
   │  ├─ :=
   │  └─ MakeExpression
   │     ├─ make
   │     ├─ ChannelType: channel of Result
   │     └─ len(items)
   ├─ ForCollectionLoop
   │  ├─ for
   │  ├─ discard
   │  ├─ item
   │  ├─ in
   │  ├─ items
   │  └─ GoStatement
   │     └─ StatementList
   │        ├─ VariableDeclaration: result := process(item)
   │        └─ SendStatement: send results, result
   └─ ForRangeLoop
      ├─ for
      ├─ i
      ├─ from
      ├─ 0
      ├─ to
      ├─ len(items)
      └─ StatementList
         ├─ VariableDeclaration: result := receive from results
         └─ ExpressionStatement: print result
```

---

## Ambiguity Resolution

### 1. Method vs Function Call

```kukicha
# Method call
todo.Display()

# Function call with method syntax (not allowed)
Display(todo)
```

**Resolution:** If expression before `()` contains `.`, it's a method call.

### 2. OnErr Clause vs Or Operator

```kukicha
# OnErr for error handling (statement-level clause)
result := calculate() onerr return error "failed"

# Or for boolean logic (expression-level operator)
if a or b
    print "at least one is true"
```

**Resolution:**
- `onerr` is a statement-level clause on VarDeclStmt, AssignStmt, or ExpressionStmt
- `or` is an expression-level boolean operator
- No ambiguity - `onerr` cannot appear in expressions, only after statements

### 3. Type Annotation vs Expression

```kukicha
# Type annotation in parameter
func Process(data list of User)

# Expression in function call
Process(getUserList())
```

**Resolution:** Context-dependent
- After parameter name → Type annotation
- In function call → Expression

### 4. Reference Creation vs Reference Type

```kukicha
# Type annotation
user reference User

# Reference creation
user := reference to User { ... }
```

**Resolution:**
- After `:` in field/parameter → Type annotation
- After `:=` or `return` → Expression

### 5. Empty Literal (Typed vs Standalone)

```kukicha
# Standalone empty (nil/zero-value)
result := empty

# Typed empty literals
todos := empty list of Todo
config := empty map of string to int
ptr := empty reference User
```

**Resolution:** 1-token lookahead after `empty`:
- If followed by `list`, `map`, `channel`, or `reference` → Typed empty literal
- Otherwise → Standalone nil/zero-value

### 6. Method Syntax

```kukicha
# Methods use explicit receiver names (no special 'this' or 'self')
func Display on todo Todo string
    return todo.title

func MarkDone on todo reference Todo
    todo.completed = true

# The receiver name is explicit - like any other parameter
func Summary on t Todo string
    return "{t.id}: {t.title}"
```

**Design Philosophy:** Following Go's "Zen", methods are just functions where the receiver is the first parameter. The `on` keyword makes this explicit and readable. There's no magic `this` or `self` - the receiver is named in the function signature just like any other parameter.

**Conversion to Go:**
| Kukicha Syntax | Go Equivalent |
|---------------|---------------|
| `func F on r T` | `func (r T) F()` |
| `func F on r reference T` | `func (r *T) F()` |

---

## Error Productions

The grammar should provide helpful errors for common mistakes:

### Missing Indentation
```kukicha
if condition
print "wrong"  # Error: Expected INDENT after if statement
```

### Mixed Assignment Operators
```kukicha
x := 5
x := 10  # Error: Variable 'x' already declared. Use '=' to reassign.
```

### Invalid OnErr Clause
```kukicha
x := 5 onerr 10  # Error: 'onerr' clause requires function returning (T, error)
```

### Missing Type Annotation
```kukicha
func Process(data)  # Warning: Type inference may fail. Consider explicit type.
```

---

## Grammar Completeness Checklist

- [x] Program structure (petiole, imports)
- [x] Type declarations (structs, interfaces)
- [x] Function declarations (functions, methods)
- [x] Function types (callbacks, higher-order functions)
- [x] Control flow (if/else, for loops)
- [x] Error handling (onerr clause)
- [x] Concurrency (go, channels, send/receive)
- [x] Expressions (arithmetic, boolean, comparison)
- [x] Pipe operator (|> for data pipelines)
- [x] Literals (all types including string interpolation)
- [x] Type annotations (all forms including function types)
- [x] Defer/recover
- [x] Lexical elements (identifiers, keywords, operators)
- [x] Indentation handling (4 spaces, tabs rejected)
- [x] Dual syntax support (Kukicha + Go)
- [x] Ambiguity resolution rules
- [x] Error productions

---

## Transparent Generic Type Parameters

**Important:** Kukicha users do NOT write generic syntax. Generic type parameters are automatically generated by the transpiler for stdlib functions.

### How It Works

When you write:
```kukicha
errors := logs |> slice.GroupBy(func(e LogEntry) string
    return e.Level
)
```

The transpiler generates:
```go
errors := slice.GroupBy(logs, func(e LogEntry) string {
    return e.Level
})

// Where GroupBy is defined as:
func GroupBy[T any, K comparable](items []T, keyFunc func(T) K) map[K][]T {
    // ...
}
```

The generic types `[T any, K comparable]` are inferred from:
1. The placeholder types in the Kukicha stdlib code (`list of any`, `map of any2 to list of any`)
2. The type constraints needed for the operation (e.g., `K comparable` is required for map keys)
3. Context-specific inference (e.g., `Enumerate` returns `iter.Seq2[int, T]`)

### Stdlib Functions with Generics (Go 1.26+)

| Function | Type Parameters | Example |
|----------|-----------------|---------|
| `slice.GroupBy` | `[T any, K comparable]` | Groups items by key type |
| `iter.Map` | `[T any, U any]` | Transforms element type |
| `iter.FlatMap` | `[T any, U any]` | Flattens and transforms |
| `iter.Filter` | `[T any]` | Filters by predicate |

### For Users

**You don't need to:**
- Write generic syntax
- Understand type parameters
- Think about constraints

**You only need to:**
- Write explicit types in function callbacks
- Let the transpiler handle the rest

## Implementation Notes

### For Transpiler Implementers

1. **Lexer must handle:**
   - Indentation tracking (INDENT/DEDENT tokens)
   - String interpolation (split into literal + expression segments)
   - Keywords vs identifiers
   - Comments (strip from token stream)

2. **Parser must handle:**
   - Operator precedence (use precedence climbing)
   - Type inference contexts
   - OnErr clause parsing (statement-level only)
   - Explicit receiver names in method declarations
   - Dual syntax (Kukicha + Go forms)

3. **Semantic analyzer must:**
   - Check type compatibility
   - Resolve identifiers to declarations
   - Verify interface implementations
   - Check that `onerr` clause is used correctly (function returns `(T, error)`)
   - Verify receiver names are only referenced within method bodies
   - Validate method receivers

4. **Code generator must:**
   - Transform `onerr` clause to if/err checks
   - Convert methods to Go receiver syntax (`on r T` → `(r T)`)
   - Handle string interpolation (fmt.Sprintf)
   - Generate proper Go package structure
   - **Generate generic type parameters for stdlib functions** (Go 1.26+ syntax)
     - Detect stdlib package context (stdlib/iter, stdlib/slice)
     - Infer type parameters from placeholder names and function signatures
     - Apply constraints where needed (e.g., `comparable` for map keys)
     - Recursively substitute types in all annotations

---

## Grammar Testing

Recommended test cases:

```kukicha
# 1. Hello World
petiole main
func main()
    print "Hello, World!"

# 2. Struct and Method (explicit receiver names)
type User
    name string
    age int

func Display on user User string
    return "{user.name}, {user.age}"

func UpdateName on user reference User, newName string
    user.name = newName

# 3. Error Handling
func LoadConfig(path string) Config
    content := file.read(path) onerr return empty
    config := json.parse(content) onerr return empty
    return config

# 4. Concurrency
func Fetch(urls list of string)
    ch := make(channel of string)
    for discard, url in urls
        go
            result := http.get(url) onerr return
            send ch, result

    for i from 0 to len(urls)
        print receive from ch

# 5. Interface
interface Processor
    Process() string

func Run(p Processor)
    print p.Process()

# 6. Pipe Operator with typed empty
func GetRepoStats(username string) list of Repo
    repos := "https://api.github.com/users/{username}/repos"
        |> http.get()
        |> .json() as list of Repo
        |> filterByStars(10)
        |> sortByUpdated()
        onerr empty list of Repo
    return repos

# 7. Empty literal variants
func EmptyExamples()
    nilValue := empty                      # standalone nil
    emptyList := empty list of Todo        # typed empty list
    emptyMap := empty map of string to int # typed empty map
    nilPtr := empty reference User         # nil pointer

# 8. Function types (callbacks and higher-order functions)
func Filter(items list of int, predicate func(int) bool) list of int
    result := empty list of int
    for item in items
        if predicate(item)
            result = append(result, item)
    return result

func ForEach(items list of string, action func(string))
    for item in items
        action(item)

func main()
    numbers := [1, 2, 3, 4, 5]

    # Pass a function literal as callback
    evens := Filter(numbers, func(n int) bool
        return n % 2 == 0
    )

    # Pass another callback
    ForEach(["a", "b", "c"], func(s string)
        print s
    )
```

---

**Grammar Version:** 0.0.1
**Last Updated:** 2026-01-20
**Status:** ✅ Implemented and Production Ready

**Implementation Notes:**
- Full transpiler implementation complete
- All grammar productions supported
- Comprehensive test coverage

**Related Documentation:**
- [Compiler Architecture](kukicha-compiler-architecture.md) - Implementation details
- [Language Syntax Reference](kukicha-syntax-v1.0.md) - Complete syntax guide with examples
- [Quick Reference](kukicha-quick-reference.md) - Developer cheat sheet
