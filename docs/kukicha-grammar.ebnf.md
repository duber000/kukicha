# Kukicha Language Grammar (EBNF)

**Version:** 1.0.0
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
Program ::= [ LeafDeclaration ] { ImportDeclaration } { TopLevelDeclaration }

LeafDeclaration ::= "leaf" PackagePath NEWLINE
    # Optional: if absent, package name is calculated from file path relative to twig.toml

PackagePath ::= IDENTIFIER { "." IDENTIFIER }

ImportDeclaration ::= "import" ImportSpec NEWLINE

ImportSpec ::= 
    | PackagePath
    | PackagePath "as" IDENTIFIER
    | URL
    | URL "as" IDENTIFIER
    | URL "@" VERSION

URL ::= DOMAIN "/" PATH
VERSION ::= "v" NUMBER "." NUMBER "." NUMBER
```

---

## Top-Level Declarations

```ebnf
TopLevelDeclaration ::=
    | TypeDeclaration
    | InterfaceDeclaration
    | FunctionDeclaration
    | MethodDeclaration

TypeDeclaration ::= "type" IDENTIFIER NEWLINE INDENT FieldList DEDENT

FieldList ::= Field { Field }

Field ::= IDENTIFIER TypeAnnotation NEWLINE

InterfaceDeclaration ::= "interface" IDENTIFIER NEWLINE INDENT MethodSignatureList DEDENT

MethodSignatureList ::= MethodSignature { MethodSignature }

MethodSignature ::= IDENTIFIER "(" [ ParameterList ] ")" [ TypeAnnotation ] NEWLINE

FunctionDeclaration ::=
    "func" IDENTIFIER "(" [ ParameterList ] ")" TypeAnnotation NEWLINE
    INDENT StatementList DEDENT
    # Return type TypeAnnotation is REQUIRED for functions that return values

MethodDeclaration ::=
    # Kukicha syntax - readable, uses explicit 'this' receiver
    | "func" IDENTIFIER "on" "this" [ "reference" ] TypeAnnotation [ "," ParameterList ] [ TypeAnnotation ] NEWLINE INDENT StatementList DEDENT
    # Go-compatible syntax - for copy-paste from Go code
    | "func" "(" IDENTIFIER TypeAnnotation ")" IDENTIFIER "(" [ ParameterList ] ")" [ TypeAnnotation ] NEWLINE INDENT StatementList DEDENT

ParameterList ::= Parameter { "," Parameter }

Parameter ::= IDENTIFIER TypeAnnotation
    # Type annotation is REQUIRED for all function/method parameters (signature-first inference)
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
    | QualifiedType
    | GoStyleType

PrimitiveType ::=
    | "int" | "int32" | "int64"
    | "uint" | "uint32" | "uint64"
    | "float32" | "float64"
    | "string" | "bool"

ReferenceType ::= "reference" ( TypeAnnotation | "to" TypeAnnotation )

ListType ::= "list" "of" TypeAnnotation

MapType ::= "map" "of" TypeAnnotation "to" TypeAnnotation

ChannelType ::= "channel" "of" TypeAnnotation

QualifiedType ::= IDENTIFIER "." IDENTIFIER

GoStyleType ::=
    | "*" TypeAnnotation
    | "[" "]" TypeAnnotation
    | "map" "[" TypeAnnotation "]" TypeAnnotation
    | "chan" TypeAnnotation
```

**Parser Implementation Note**: When in type annotation context, if the parser sees `list`, `map`, or `channel`, it MUST be followed by `of`. This is not ambiguous because the parser knows when it expects a type.

---

## Statements

```ebnf
StatementList ::= Statement { Statement }

Statement ::=
    | VariableDeclaration
    | Assignment
    | ReturnStatement
    | IfStatement
    | ForStatement
    | DeferStatement
    | GoStatement
    | SendStatement
    | ExpressionStatement
    | NEWLINE

VariableDeclaration ::= IDENTIFIER ":=" Expression NEWLINE

Assignment ::= 
    | IDENTIFIER "=" Expression NEWLINE
    | Expression "=" Expression NEWLINE

ReturnStatement ::= "return" [ ExpressionList ] NEWLINE

IfStatement ::=
    "if" Expression NEWLINE
    INDENT StatementList DEDENT
    [ ElseClause ]

ElseClause ::=
    | "else" NEWLINE INDENT StatementList DEDENT
    | "else" IfStatement

ForStatement ::=
    | ForRangeLoop
    | ForCollectionLoop
    | ForGoStyleLoop

ForRangeLoop ::=
    "for" IDENTIFIER "from" Expression ( "to" | "through" ) Expression NEWLINE
    INDENT StatementList DEDENT

ForCollectionLoop ::=
    "for" [ IDENTIFIER "," ] IDENTIFIER "in" Expression NEWLINE
    INDENT StatementList DEDENT

ForGoStyleLoop ::=
    "for" [ IDENTIFIER ":=" Expression ";" ] Expression [ ";" Expression ] NEWLINE
    INDENT StatementList DEDENT

DeferStatement ::=
    | "defer" Expression NEWLINE
    | "defer" NEWLINE INDENT StatementList DEDENT

GoStatement ::= "go" ( Expression | NEWLINE INDENT StatementList DEDENT ) NEWLINE

SendStatement ::= "send" Expression "," Expression NEWLINE

ExpressionStatement ::= Expression NEWLINE
```

---

## Expressions

```ebnf
Expression ::= OrExpression [ OnErrOperator ]

OnErrOperator ::= "onerr" ( Expression | NEWLINE INDENT StatementList DEDENT )

OrExpression ::= PipeExpression { ( "or" | "||" ) PipeExpression }

PipeExpression ::= AndExpression { "|>" AndExpression }

AndExpression ::= ComparisonExpression { ( "and" | "&&" ) ComparisonExpression }

ComparisonExpression ::= AdditiveExpression [ ComparisonOp AdditiveExpression ]

ComparisonOp ::=
    | "equals" | "=="
    | "not" "equals" | "!="
    | ">" | "<" | ">=" | "<="
    | "in"
    | "not" "in"

AdditiveExpression ::= MultiplicativeExpression { ( "+" | "-" ) MultiplicativeExpression }

MultiplicativeExpression ::= UnaryExpression { ( "*" | "/" | "%" ) UnaryExpression }

UnaryExpression ::= 
    | ( "not" | "!" | "-" ) UnaryExpression
    | PostfixExpression

PostfixExpression ::=
    PrimaryExpression {
        | "." IDENTIFIER
        | "(" [ ExpressionList ] ")"
        | "at" Expression
        | "[" Expression "]"
        | "[" [ Expression ] ":" [ Expression ] "]"
    }

PrimaryExpression ::=
    | IDENTIFIER
    | Literal
    | "(" Expression ")"
    | StructLiteral
    | EmptyLiteral       # 'empty' with optional type (uses 1-token lookahead)
    | ListLiteral        # 'list of Type' with initial values
    | MapLiteral         # 'map of K to V' with initial entries
    | ChannelCreation
    | ReceiveExpression
    | RecoverExpression
    | TypeCast
    | "this"

ExpressionList ::= Expression { "," Expression }
```

---

## Literals

```ebnf
Literal ::=
    | IntegerLiteral
    | FloatLiteral
    | StringLiteral
    | BooleanLiteral

IntegerLiteral ::= DIGIT { DIGIT }

FloatLiteral ::= DIGIT { DIGIT } "." DIGIT { DIGIT }

StringLiteral ::= 
    | '"' { StringChar | Interpolation } '"'
    | "'" { StringChar } "'"

StringChar ::= /* any character except ", ', newline, or { */

Interpolation ::= "{" Expression "}"

BooleanLiteral ::= "true" | "false"

StructLiteral ::=
    TypeAnnotation NEWLINE
    INDENT FieldInitList DEDENT

FieldInitList ::= FieldInit { FieldInit }

FieldInit ::= IDENTIFIER ":" Expression NEWLINE

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
ListLiteral ::=
    | "list" "of" TypeAnnotation NEWLINE INDENT ExpressionList DEDENT
    | "[" "]" TypeAnnotation "{" "}"
    | "[" "]" TypeAnnotation "{" ExpressionList "}"

# Non-empty map literal (map with initial entries)
MapLiteral ::=
    | "map" "of" TypeAnnotation "to" TypeAnnotation NEWLINE INDENT MapEntryList DEDENT
    | "map" "[" TypeAnnotation "]" TypeAnnotation "{" "}"
    | "map" "[" TypeAnnotation "]" TypeAnnotation "{" MapEntryList "}"

MapEntryList ::= MapEntry { MapEntry }

MapEntry ::= IDENTIFIER ":" Expression NEWLINE

ChannelCreation ::= "make" ( ChannelType | "(" "chan" TypeAnnotation [ "," Expression ] ")" )

ReceiveExpression ::= 
    | "receive" Expression
    | "<-" Expression

RecoverExpression ::= "recover" "(" ")"

TypeCast ::=
    | TypeAnnotation "(" Expression ")"
    | Expression "as" TypeAnnotation
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
leaf        import      type        interface   func
if          else        for         in          from
to          through     at          of          and
or          onerr       not         return      go
defer       make        list        map         channel
send        receive     close       panic       recover
error       empty       reference   on          this
discard     true        false       equals      as
```

**Note:** The keywords `list`, `map`, and `channel` are context-sensitive and may also be used as identifiers in certain contexts.

---

## Operators and Delimiters

```
+     -     *     /     %
==    !=    <     <=    >     >=    equals    in
&&    ||    !     and   or    not
:=    =     :     .     ,
(     )     [     ]     {     }
<-    ->    |>
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

### OnErr Operator Auto-Unwrap

The `onerr` operator provides ergonomic error handling for functions that return `(T, error)` tuples:
1. Automatically unwrap to `T` if no error
2. Execute the `onerr` clause if error is not empty

**Important:** The `onerr` operator is distinct from the boolean `or` operator. This separation makes code more readable - you can tell at a glance whether an expression handles errors or performs boolean logic.

Example:
```kukicha
# Error handling with onerr
data := file.read(path) onerr panic "failed"

# Boolean logic with or
if active or pending
    process()
```

The `onerr` operator desugars to:
```kukicha
data, err := file.read(path)
if err != empty
    panic "failed"
```

### Discard Keyword

The `discard` keyword is syntactic sugar for Go's `_` (blank identifier):
- Can appear in tuple unpacking
- Cannot be referenced as a variable

### Membership Operators

The `in` and `not in` operators test for membership in collections.

**Disambiguation:**
- In `for` loops: `for x in items` — iteration syntax
- In expressions: `x in items` — membership test operator

**Code Generation:**

For lists/slices:
```kukicha
# Source
if item in items
    print "found"

# Generates Go
import "slices"
if slices.Contains(items, item) {
    fmt.Println("found")
}
```

For maps:
```kukicha
# Source
if key in config
    print "exists"

# Generates Go
if _, exists := config[key]; exists {
    fmt.Println("exists")
}
```

For strings:
```kukicha
# Source
if "hello" in text
    print "found"

# Generates Go
import "strings"
if strings.Contains(text, "hello") {
    fmt.Println("found")
}
```

The `not in` operator negates the result:
```kukicha
# Source
if item not in blacklist
    process(item)

# Generates Go
if !slices.Contains(blacklist, item) {
    process(item)
}
```

### Negative Indexing

Kukicha supports negative indices for accessing elements from the end of a collection.

**Single element access:**
```kukicha
# Source
last := items at -1
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
- Both `at` keyword and bracket `[]` syntax support negative indices

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
- Pipe has higher precedence than `onerr` operator

```kukicha
# Example
a + b |> double() onerr "default"

# Desugars to
((double(a + b)) onerr "default")
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

### Example 2: Method with OnErr Operator

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
├─ this         # explicit receiver name
├─ Config       # receiver type
├─ ParameterList
│  └─ Parameter
│     ├─ path
│     └─ string
└─ StatementList
   ├─ VariableDeclaration
   │  ├─ content
   │  ├─ :=
   │  └─ OnErrExpression
   │     ├─ FunctionCall: file.read(path)
   │     └─ onerr
   │        └─ ReturnStatement: return error "cannot read"
   ├─ Assignment
   │  ├─ this.data
   │  ├─ =
   │  └─ OnErrExpression
   │     ├─ FunctionCall: json.parse(content)
   │     └─ onerr
   │        └─ ReturnStatement: return error "invalid json"
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
        result := receive results
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
   │  └─ ChannelCreation
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
         ├─ VariableDeclaration: result := receive results
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

### 2. OnErr vs Or Operators

```kukicha
# OnErr for error handling
result := calculate() onerr return error "failed"

# Or for boolean logic
if a or b
    print "at least one is true"
```

**Resolution:**
- `onerr` is used exclusively for error handling (functions returning `(T, error)`)
- `or` is used exclusively for boolean logic
- No ambiguity - the keywords are distinct and self-documenting

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

### Invalid OnErr Operator
```kukicha
x := 5 onerr 10  # Error: 'onerr' operator requires function returning (T, error)
```

### Missing Type Annotation
```kukicha
func Process(data)  # Warning: Type inference may fail. Consider explicit type.
```

---

## Grammar Completeness Checklist

- [x] Program structure (leaf, imports)
- [x] Type declarations (structs, interfaces)
- [x] Function declarations (functions, methods)
- [x] Control flow (if/else, for loops)
- [x] Error handling (or operator)
- [x] Concurrency (go, channels, send/receive)
- [x] Expressions (arithmetic, boolean, comparison)
- [x] Pipe operator (|> for data pipelines)
- [x] Literals (all types including string interpolation)
- [x] Type annotations (all forms)
- [x] Defer/recover
- [x] Lexical elements (identifiers, keywords, operators)
- [x] Indentation handling (4 spaces, tabs rejected)
- [x] Dual syntax support (Kukicha + Go)
- [x] Ambiguity resolution rules
- [x] Error productions

---

## Implementation Notes

### For Parser Implementers

1. **Lexer must handle:**
   - Indentation tracking (INDENT/DEDENT tokens)
   - String interpolation (split into literal + expression segments)
   - Keywords vs identifiers
   - Comments (strip from token stream)

2. **Parser must handle:**
   - Operator precedence (use precedence climbing)
   - Type inference contexts
   - OnErr operator desugaring
   - Explicit `this` receiver in method declarations
   - Dual syntax (Kukicha + Go forms)

3. **Semantic analyzer must:**
   - Check type compatibility
   - Resolve identifiers to declarations
   - Verify interface implementations
   - Check that `onerr` operator is used correctly (function returns `(T, error)`)
   - Verify `this` is only referenced within method bodies
   - Validate method receivers

4. **Code generator must:**
   - Transform `onerr` operator to if/err checks
   - Convert methods to Go receiver syntax (`on this T` → `(this T)`)
   - Handle string interpolation (fmt.Sprintf)
   - Generate proper Go package structure

---

## Grammar Testing

Recommended test cases:

```kukicha
# 1. Hello World
leaf main
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
    return json.parse(content) onerr return empty

# 4. Concurrency
func Fetch(urls list of string)
    ch := make channel of string
    for discard, url in urls
        go
            result := http.get(url) onerr return
            send ch, result

    for i from 0 to len(urls)
        print receive ch

# 5. Interface
interface Processor
    Process() string

func Run(p Processor)
    print p.Process()

# 6. Pipe Operator with typed empty
func GetRepoStats(username string) list of Repo
    return "https://api.github.com/users/{username}/repos"
        |> http.get()
        |> .json() as list of Repo
        |> filterByStars(10)
        |> sortByUpdated()
        onerr empty list of Repo

# 7. Empty literal variants
func EmptyExamples()
    nilValue := empty                      # standalone nil
    emptyList := empty list of Todo        # typed empty list
    emptyMap := empty map of string to int # typed empty map
    nilPtr := empty reference User         # nil pointer
```

---

**Grammar Version:** 1.0.0
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
