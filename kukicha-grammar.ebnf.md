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
Program ::= LeafDeclaration { ImportDeclaration } { TopLevelDeclaration }

LeafDeclaration ::= "leaf" PackagePath NEWLINE

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
    "func" IDENTIFIER "(" [ ParameterList ] ")" [ TypeAnnotation ] NEWLINE
    INDENT StatementList DEDENT

MethodDeclaration ::=
    | "func" IDENTIFIER "on" TypeAnnotation [ "," ParameterList ] NEWLINE INDENT StatementList DEDENT
    | "func" "(" IDENTIFIER TypeAnnotation ")" IDENTIFIER "(" [ ParameterList ] ")" [ TypeAnnotation ] NEWLINE INDENT StatementList DEDENT

ParameterList ::= Parameter { "," Parameter }

Parameter ::= IDENTIFIER [ TypeAnnotation ]
```

---

## Type Annotations

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
Expression ::= OrExpression [ OrOperator ]

OrOperator ::= "or" ( Expression | NEWLINE INDENT StatementList DEDENT )

OrExpression ::= PipeExpression { ( "or" | "||" ) PipeExpression }

PipeExpression ::= AndExpression { "|>" AndExpression }

AndExpression ::= ComparisonExpression { ( "and" | "&&" ) ComparisonExpression }

ComparisonExpression ::= AdditiveExpression [ ComparisonOp AdditiveExpression ]

ComparisonOp ::=
    | "equals" | "=="
    | "not" "equals" | "!="
    | ">" | "<" | ">=" | "<="

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
        | "[" Expression ":" Expression "]"
    }

PrimaryExpression ::=
    | IDENTIFIER
    | Literal
    | "(" Expression ")"
    | StructLiteral
    | ListLiteral
    | MapLiteral
    | ChannelCreation
    | ReceiveExpression
    | RecoverExpression
    | TypeCast
    | "this"
    | "empty"

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

ListLiteral ::=
    | "empty" "list" "of" TypeAnnotation
    | "list" "of" TypeAnnotation
    | "[" "]" TypeAnnotation "{" "}"
    | "[" "]" TypeAnnotation "{" ExpressionList "}"

MapLiteral ::=
    | "empty" "map" "of" TypeAnnotation "to" TypeAnnotation
    | "map" "of" TypeAnnotation "to" TypeAnnotation NEWLINE INDENT MapEntryList DEDENT
    | "map" "[" TypeAnnotation "]" TypeAnnotation "{" "}"

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
or          not         return      go          defer
make        channel     send        receive     close
panic       recover     error       empty       reference
on          this        discard     true        false
equals      as
```

---

## Operators and Delimiters

```
+     -     *     /     %     
==    !=    <     <=    >     >=    equals
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

### Or Operator Auto-Unwrap

When the `or` operator is used with a function that returns `(T, error)`:
1. Automatically unwrap to `T` if no error
2. Execute the `or` clause if error is not empty

Example:
```kukicha
data := file.read(path) or panic "failed"
```

Desugars to:
```kukicha
data, err := file.read(path)
if err != empty
    panic "failed"
```

### Discard Keyword

The `discard` keyword is syntactic sugar for Go's `_` (blank identifier):
- Can appear in tuple unpacking
- Cannot be referenced as a variable

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
- Pipe has higher precedence than `or` operator

```kukicha
# Example
a + b |> double() or "default"

# Desugars to
((double(a + b)) or "default")
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

### Example 2: Method with Or Operator

```kukicha
func Load on Config, path string
    content := file.read(path) or return error "cannot read"
    this.data = json.parse(content) or return error "invalid json"
    return empty
```

Parse tree:
```
MethodDeclaration
├─ func
├─ Load
├─ on
├─ Config
├─ ParameterList
│  └─ Parameter
│     ├─ path
│     └─ string
└─ StatementList
   ├─ VariableDeclaration
   │  ├─ content
   │  ├─ :=
   │  └─ OrExpression
   │     ├─ FunctionCall: file.read(path)
   │     └─ or
   │        └─ ReturnStatement: return error "cannot read"
   ├─ Assignment
   │  ├─ this.data
   │  ├─ =
   │  └─ OrExpression
   │     ├─ FunctionCall: json.parse(content)
   │     └─ or
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

### 2. Or Operator Precedence

```kukicha
# Or as error handler (higher precedence)
result := calculate() or return error "failed"

# Or as boolean operator (lower precedence)
if a or b
    print "at least one is true"
```

**Resolution:** 
- After `:=` or `=`, `or` is error handler (higher precedence)
- After `if`, `for`, etc., `or` is boolean operator (lower precedence)

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

### Invalid Or Operator
```kukicha
x := 5 or 10  # Error: 'or' operator requires function returning (T, error)
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
   - Or operator desugaring
   - Dual syntax (Kukicha + Go forms)

3. **Semantic analyzer must:**
   - Check type compatibility
   - Resolve identifiers to declarations
   - Verify interface implementations
   - Check that `or` operator is used correctly
   - Validate method receivers

4. **Code generator must:**
   - Transform `or` operator to if/err checks
   - Convert methods to Go receiver syntax
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

# 2. Struct and Method
type User
    name string
    age int

func Display on User
    return "{this.name}, {this.age}"

# 3. Error Handling
func LoadConfig(path string)
    content := file.read(path) or return empty
    return json.parse(content) or return empty

# 4. Concurrency
func Fetch(urls list of string)
    ch := make channel of string
    for discard, url in urls
        go
            result := http.get(url) or return
            send ch, result
    
    for i from 0 to len(urls)
        print receive ch

# 5. Interface
interface Processor
    Process() string

func Run(p Processor)
    print p.Process()

# 6. Pipe Operator
func GetRepoStats(username string)
    return "https://api.github.com/users/{username}/repos"
        |> http.get()
        |> .json() as list of Repo
        |> filterByStars(10)
        |> sortByUpdated()
        or empty list of Repo
```

---

## Next Steps

With this grammar defined:
1. Implement lexer (tokenizer)
2. Implement parser (AST builder)
3. Implement semantic analyzer
4. Implement code generator (AST → Go)
5. Test with example programs

---

**Grammar Version:** 1.0.0
**Last Updated:** 2025-01-19
**Status:** Complete and ready for implementation
