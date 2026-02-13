# Proposal: Fix Syntax Friction Discovered Writing Tutorials

**Date:** 2026-02-13
**Status:** Draft
**References:** [ROADMAP-v0.0.2 — Syntax Friction](../ROADMAP-v0.0.2.md#syntax-friction-discovered-writing-tutorials)

---

## Summary

Three syntax pain points were identified while writing tutorials for Kukicha v0.0.1. This proposal addresses each with concrete syntax designs, grammar changes, AST additions, and transpilation rules. The goals are:

1. Make pipe + functional helper chains feel natural instead of verbose (**High priority**)
2. Eliminate the confusing `go func()...()` IIFE pattern (**Low priority**)
3. Implement `switch/case/default` to replace if/else dispatch chains (**High priority**)

---

## 1. Arrow Lambdas (Short Inline Functions)

### Problem

Every `slice.Filter` or `slice.Map` call requires a full `function(param Type) ReturnType` signature with an indented body and explicit `return`. A simple predicate becomes 3+ lines:

```kukicha
# Current: 3 lines for a one-expression predicate
repos |> slice.Filter(function(r Repo) bool
    return r.Stars > 100
)

# Current: 3 lines for a one-expression transform
repos |> slice.Map(function(r Repo) string
    return r.Name
)
```

This is the **main friction point** when composing pipes with functional helpers. Beginners expect something concise; instead they get ceremony.

### Design

Introduce **arrow lambda expressions** using `=>`. Two forms:

#### A. Expression Lambda (single expression — auto-returned)

```
( Parameters ) => Expression
```

The return type is inferred from the expression. The `return` is implicit.

#### B. Block Lambda (multi-statement body — explicit return)

```
( Parameters ) => NEWLINE INDENT StatementList DEDENT
```

Same as a full `function` literal, but shorter header.

### Syntax Examples

```kukicha
# ── Expression lambdas (single expression, auto-return) ──────────

# Typed parameters (always works, no inference needed)
repos |> slice.Filter((r Repo) => r.Stars > 100)
repos |> slice.Map((r Repo) => r.Name)
names |> slice.Filter((s string) => s |> string.Contains("go"))

# Untyped parameters (type inferred from calling context)
repos |> slice.Filter(r => r.Stars > 100)
repos |> slice.Map(r => r.Name)
numbers |> slice.Filter(n => n > 0)

# Multiple parameters
pairs |> slice.Map((k string, v int) => "{k}: {v}")

# No parameters
button.OnClick(() => print("clicked"))


# ── Block lambdas (multi-statement body) ─────────────────────────

repos |> slice.Filter((r Repo) =>
    name := r.Name |> string.ToLower()
    return name |> string.Contains("go")
)

# Untyped block lambda
repos |> slice.Filter(r =>
    name := r.Name |> string.ToLower()
    return name |> string.Contains("go")
)
```

### Comparison: Before and After

```kukicha
# ── BEFORE (v0.0.1) ─────────────────────────────────────────────

# Filter repos with >100 stars
active := repos |> slice.Filter(function(r Repo) bool
    return r.Stars > 100
)

# Get repo names
names := repos |> slice.Map(function(r Repo) string
    return r.Name
)

# Full pipeline: filter → map → sort
result := repos
    |> slice.Filter(function(r Repo) bool
        return r.Stars > 100
    )
    |> slice.Map(function(r Repo) string
        return r.Name
    )
    |> slice.Sort(function(a string, b string) bool
        return a < b
    )


# ── AFTER (v0.0.2 with arrow lambdas) ───────────────────────────

# Filter repos with >100 stars
active := repos |> slice.Filter((r Repo) => r.Stars > 100)

# Get repo names
names := repos |> slice.Map((r Repo) => r.Name)

# Full pipeline: filter → map → sort
result := repos
    |> slice.Filter((r Repo) => r.Stars > 100)
    |> slice.Map((r Repo) => r.Name)
    |> slice.Sort((a string, b string) => a < b)
```

**Line reduction:** 17 lines → 7 lines for the pipeline example. Each pipe stage fits on one line.

### Grammar Changes (EBNF)

Add to `PrimaryExpression`:

```ebnf
PrimaryExpression ::=
    | ... (existing productions)
    | ArrowLambda

ArrowLambda ::=
    | "(" [ TypedParameterList ] ")" "=>" LambdaBody        # zero or 2+ params, typed
    | "(" [ UntypedParameterList ] ")" "=>" LambdaBody      # zero or 2+ params, untyped
    | IDENTIFIER "=>" LambdaBody                             # single untyped param (no parens)

LambdaBody ::=
    | Expression                                             # expression lambda (auto-return)
    | NEWLINE INDENT StatementList DEDENT                    # block lambda (explicit return)

TypedParameterList ::= TypedParameter { "," TypedParameter }
TypedParameter ::= IDENTIFIER TypeAnnotation

UntypedParameterList ::= IDENTIFIER { "," IDENTIFIER }
```

**Parsing notes:**
- A single identifier followed by `=>` is an untyped single-param lambda: `r => r.Stars > 100`
- A parenthesized group followed by `=>` is a multi-param or zero-param lambda: `(r Repo) => ...`, `() => ...`
- The parser can distinguish `(r Repo) => ...` (lambda) from `(r Repo)` (parenthesized expression) by looking for `=>` after the closing paren
- `=>` is a new token (`TOKEN_FAT_ARROW`)

### AST Changes

Add a new expression node:

```go
type ArrowLambda struct {
    Token      lexer.Token   // The '=>' token (or first param token)
    Parameters []*Parameter  // May have nil Type for untyped params
    Body       Expression    // Expression lambda: the single expression
    Block      *BlockStmt    // Block lambda: multi-statement body (mutually exclusive with Body)
}
```

### Parser Changes

1. Add `TOKEN_FAT_ARROW` (`=>`) to the lexer
2. In `parsePrimaryExpression`:
   - When seeing `IDENTIFIER` followed by `=>`: parse as single-param untyped lambda
   - When seeing `(`: attempt to parse lambda by looking ahead for `)` followed by `=>`
   - Fallback: parse as parenthesized expression (existing behavior)
3. New function `parseArrowLambda()`:
   - Parse parameter list (typed or untyped)
   - Consume `=>`
   - If next token is `NEWLINE` + `INDENT`: parse block body
   - Otherwise: parse single expression as body

### Codegen Changes

Arrow lambdas transpile to Go anonymous functions:

```kukicha
# Expression lambda
(r Repo) => r.Stars > 100
```
```go
func(r Repo) bool {
    return r.Stars > 100
}
```

```kukicha
# Block lambda
(r Repo) =>
    name := strings.ToLower(r.Name)
    return strings.Contains(name, "go")
```
```go
func(r Repo) bool {
    name := strings.ToLower(r.Name)
    return strings.Contains(name, "go")
}
```

**Return type inference:**
- For expression lambdas, the return type is inferred from the expression:
  - Field access (`r.Stars > 100`) → `bool`
  - Field access (`r.Name`) → type of field
  - String literal → `string`
  - Arithmetic → numeric type
- For block lambdas, the return type is inferred from `return` statements
- The semantic analyzer already has type-checking infrastructure; this extends it to lambda bodies

**Untyped parameter inference:**
- When a lambda is passed as an argument to a function with a known `func(T) U` parameter type, the parameter types are inferred from the calling context
- Example: `slice.Filter` expects `func(T) bool` — so `r => r.Stars > 100` gets `r` typed as `T` (the element type of the piped list)
- The semantic analyzer resolves this during the type-checking pass using the signature-first approach already in place

### Type Inference Strategy

Untyped lambdas require **contextual type inference**. This is feasible because Kukicha already has a signature-first semantic analysis pass. The approach:

1. When analyzing a `CallExpr` whose argument is an `ArrowLambda` with untyped parameters:
2. Look up the function's signature to find the expected parameter type (e.g., `func(T) bool`)
3. Unify `T` with the actual generic type from the piped value
4. Assign the resolved types to the lambda's parameters
5. Proceed with body type-checking as normal

**Scope of inference:** Initially, type inference should be limited to **direct call arguments** — the lambda must be an immediate argument to a function whose signature is known. This covers the primary use case (`slice.Filter`, `slice.Map`, etc.) without requiring full Hindley-Milner inference.

**Fallback:** If inference fails (e.g., the lambda is assigned to a variable, or the calling function's signature is unknown), the compiler should emit a clear error:

```
Error in main.kuki:12:30

   12 | repos |> slice.Filter(r => r.Stars > 100)
      |                        ^
      | Cannot infer type of parameter 'r'
      |
Help: Add an explicit type: (r Repo) => r.Stars > 100
```

---

## 2. `go` Block Syntax (Eliminate IIFE Pattern)

### Problem

Spawning a goroutine with an inline block requires the Go IIFE pattern, which is confusing for beginners:

```kukicha
# Current: IIFE pattern — the trailing () is non-obvious
go func()
    s.mu.Lock()
    s.db.IncrementClicks(code)
    s.mu.Unlock()
()
```

The trailing `()` invokes the anonymous function. Beginners don't understand why it's there.

### Design

The grammar **already specifies** `go` with a block form:

```ebnf
GoStatement ::= "go" ( Expression | NEWLINE INDENT StatementList DEDENT ) NEWLINE
```

But the parser only accepts `CallExpr` or `MethodCallExpr`. The fix is to implement what the grammar already describes.

### Syntax

```kukicha
# ── PROPOSED: go with block ──────────────────────────────────────

go
    s.mu.Lock()
    s.db.IncrementClicks(code)
    s.mu.Unlock()


# ── STILL VALID: go with function call ───────────────────────────

go processItem(item)
go s.handleRequest(req)
```

### Comparison: Before and After

```kukicha
# BEFORE (IIFE — confusing)
go func()
    s.mu.Lock()
    s.db.IncrementClicks(code)
    s.mu.Unlock()
()

# AFTER (block — clear)
go
    s.mu.Lock()
    s.db.IncrementClicks(code)
    s.mu.Unlock()
```

### Parser Changes

Modify `parseGoStmt()` to check for `NEWLINE` + `INDENT` after `go`:

```go
func (p *Parser) parseGoStmt() *ast.GoStmt {
    token := p.advance() // consume 'go'

    // Check for block form: go NEWLINE INDENT ... DEDENT
    if p.check(lexer.TOKEN_NEWLINE) {
        p.advance() // consume newline
        if p.check(lexer.TOKEN_INDENT) {
            block := p.parseBlock()
            return &ast.GoStmt{
                Token: token,
                Block: block,   // new field
            }
        }
    }

    // Expression form (existing behavior)
    expr := p.parseExpression()
    // ... existing validation ...
}
```

### AST Changes

The `GoStmt` node already has a `Block` field (defined in the architecture doc but not used in the parser):

```go
type GoStmt struct {
    Token lexer.Token
    Call  Expression  // Expression form: go f()
    Block *BlockStmt  // Block form: go NEWLINE INDENT ... DEDENT
}
```

Exactly one of `Call` or `Block` should be non-nil.

### Codegen Changes

When `GoStmt.Block` is non-nil, generate a Go IIFE:

```kukicha
go
    s.mu.Lock()
    s.db.IncrementClicks(code)
    s.mu.Unlock()
```

Generates:

```go
go func() {
    s.mu.Lock()
    s.db.IncrementClicks(code)
    s.mu.Unlock()
}()
```

The codegen wraps the block in `func() { ... }()` automatically — the user never sees the IIFE.

### Variable Capture

When variables from the outer scope are used inside a `go` block, they are captured by the closure (same as Go). The compiler should **not** add any special behavior — Go's closure semantics apply directly. This matches the existing behavior of `go func()...()`.

For loop variable capture, Go 1.22+ changed loop variables to be per-iteration, so the classic closure-in-loop bug is already resolved by the target Go version (1.25+).

---

## 3. `switch/case/default` Implementation

### Problem

`switch`, `case`, and `default` are reserved keywords in the lexer but have **no parser, AST, or compiler support**. The CLI Explorer tutorial uses verbose if/else chains for command dispatch:

```kukicha
# Current: if/else chain for command dispatch
if command equals "search"
    handleSearch(ex, term)
else if command equals "list"
    handleList(ex)
else if command equals "filter"
    handleFilter(ex, term)
else if command equals "stats"
    handleStats(ex)
else if command equals "help"
    handleHelp()
else if command equals "quit" or command equals "exit"
    print "Goodbye!"
    break
else
    print "Unknown command. Type 'help' for usage."
```

### Design

Kukicha's switch follows Go semantics (no fallthrough by default) with indentation-based blocks:

#### A. Expression Switch (most common)

```kukicha
switch command
    case "search"
        handleSearch(ex, term)
    case "list"
        handleList(ex)
    case "filter"
        handleFilter(ex, term)
    case "stats"
        handleStats(ex)
    case "help"
        handleHelp()
    case "quit", "exit"
        print "Goodbye!"
        break
    default
        print "Unknown command. Type 'help' for usage."
```

#### B. Tagless Switch (replaces if/else if chains)

```kukicha
switch
    case score >= 90
        grade := "A"
    case score >= 80
        grade := "B"
    case score >= 70
        grade := "C"
    default
        grade := "F"
```

#### C. Switch with Initializer

```kukicha
switch value := computeValue()
    case 1
        print "one"
    case 2
        print "two"
    default
        print "other: {value}"
```

### Comparison: Before and After

```kukicha
# ── BEFORE (if/else chain) ──────────────────────────────────────

if command equals "search"
    handleSearch(ex, term)
else if command equals "list"
    handleList(ex)
else if command equals "filter"
    handleFilter(ex, term)
else if command equals "stats"
    handleStats(ex)
else if command equals "help"
    handleHelp()
else if command equals "quit" or command equals "exit"
    print "Goodbye!"
    break
else
    print "Unknown command. Type 'help' for usage."


# ── AFTER (switch/case) ─────────────────────────────────────────

switch command
    case "search"
        handleSearch(ex, term)
    case "list"
        handleList(ex)
    case "filter"
        handleFilter(ex, term)
    case "stats"
        handleStats(ex)
    case "help"
        handleHelp()
    case "quit", "exit"
        print "Goodbye!"
        break
    default
        print "Unknown command. Type 'help' for usage."
```

### Grammar Changes (EBNF)

Add to `Statement`:

```ebnf
Statement ::=
    | ... (existing productions)
    | SwitchStatement

SwitchStatement ::=
    "switch" [ SimpleStatement ";" ] [ Expression ] NEWLINE
    INDENT CaseClause { CaseClause } [ DefaultClause ] DEDENT

CaseClause ::=
    "case" ExpressionList NEWLINE
    INDENT StatementList DEDENT

DefaultClause ::=
    "default" NEWLINE
    INDENT StatementList DEDENT
```

**Notes:**
- Multiple expressions per case: `case "quit", "exit"` (comma-separated, same as Go)
- No fallthrough by default (same as Go)
- Tagless switch: omit the expression after `switch` for boolean case conditions
- Optional initializer: `switch x := f()` followed by the tag expression after `;`

### AST Changes

```go
type SwitchStmt struct {
    Token   lexer.Token   // The 'switch' token
    Init    Statement     // Optional initializer (e.g., x := f())
    Tag     Expression    // Optional tag expression (nil for tagless switch)
    Cases   []*CaseClause
    Default *BlockStmt    // Optional default block
}

type CaseClause struct {
    Token       lexer.Token   // The 'case' token
    Expressions []Expression  // One or more match expressions
    Body        *BlockStmt
}
```

### Codegen

Direct mapping to Go switch:

```kukicha
switch command
    case "search"
        handleSearch(ex, term)
    case "quit", "exit"
        print "Goodbye!"
    default
        print "Unknown"
```

Generates:

```go
switch command {
case "search":
    handleSearch(ex, term)
case "quit", "exit":
    fmt.Println("Goodbye!")
default:
    fmt.Println("Unknown")
}
```

---

## Implementation Plan

### Phase 1: `go` Block Syntax (Smallest change, immediate impact)

**Effort:** Small — the grammar and AST already support it.

1. Modify `parseGoStmt()` to detect `NEWLINE INDENT` and parse a block
2. Modify `generateGoStmt()` to emit `go func() { ... }()` for block form
3. Add parser tests for `go` with block
4. Add codegen tests for `go` with block
5. Update quick reference and production tutorial

### Phase 2: `switch/case/default` (High value, moderate effort)

**Effort:** Moderate — new statement type through all compiler phases.

1. Add `TOKEN_SWITCH`, `TOKEN_CASE`, `TOKEN_DEFAULT` recognition in lexer (already reserved)
2. Add `SwitchStmt` and `CaseClause` AST nodes
3. Implement `parseSwitchStmt()` in parser
4. Implement `generateSwitchStmt()` in codegen
5. Add semantic analysis for switch (type checking case expressions against tag)
6. Add tests for all switch variants (expression, tagless, initializer, multi-value case)
7. Update CLI Explorer tutorial to use switch
8. Update quick reference and grammar docs

### Phase 3: Arrow Lambdas (Highest value, largest effort)

**Effort:** Large — new expression type, new token, type inference extensions.

1. Add `TOKEN_FAT_ARROW` (`=>`) to lexer
2. Add `ArrowLambda` AST node
3. Implement `parseArrowLambda()` in parser with lookahead for `=>`
4. Implement `generateArrowLambda()` in codegen (transpile to Go `func` literal)
5. Implement return type inference for expression lambdas
6. Implement contextual parameter type inference for untyped lambdas
7. Add comprehensive parser tests (typed, untyped, expression, block, 0/1/N params)
8. Add codegen tests
9. Add semantic analysis tests for type inference
10. Update all tutorials to use arrow lambdas where appropriate
11. Update quick reference, grammar, and architecture docs

### Phase Ordering Rationale

| Phase | Feature | Priority | Effort | Why this order |
|-------|---------|----------|--------|----------------|
| 1 | `go` block | Low | Small | Minimal risk, grammar already defined, quick win |
| 2 | `switch/case` | High | Moderate | High tutorial impact, no type inference complexity |
| 3 | Arrow lambdas | High | Large | Most impactful but needs type inference, benefits from lessons learned in phases 1-2 |

---

## Design Decisions & Alternatives Considered

### Why `=>` over other lambda syntaxes?

| Syntax | Example | Pros | Cons |
|--------|---------|------|------|
| **`=>` (chosen)** | `r => r.Stars > 100` | Widely known (JS, C#, Dart, Kotlin), clean | New token |
| `\|r\|` (Rust) | `\|r\| r.Stars > 100` | Familiar to Rust devs | Ambiguous with pipe `\|>`, confusing for beginners |
| `r -> r.Stars` | `r -> r.Stars > 100` | Familiar (Java, Haskell) | `->` could be confused with channel arrow |
| `{ r in ... }` (Swift) | `{ r in r.Stars > 100 }` | Readable | Braces conflict with Kukicha's indentation philosophy |
| `it.Stars > 100` (Kotlin) | `slice.Filter(it.Stars > 100)` | Ultra-concise | Implicit variable feels "magical", only single-param |

**Decision:** `=>` is the best fit because:
- It's the most widely recognized lambda syntax across languages
- No ambiguity with existing Kukicha operators
- Supports both typed and untyped parameters naturally
- Works for 0, 1, and N parameters consistently

### Why not full type inference everywhere?

Kukicha's philosophy is **explicit types at boundaries** (function signatures, struct fields) with inference only inside function bodies (`:=`). Arrow lambdas extend this by allowing inference **only when the calling context provides unambiguous type information** (e.g., `slice.Filter` knows it needs `func(T) bool`).

This keeps the language predictable:
- Top-level functions: always typed
- Lambda arguments to known functions: optionally untyped
- Standalone lambdas assigned to variables: must be typed

### Why not add `fallthrough` to switch?

Go has `fallthrough` but it's rarely used and considered a footgun. Kukicha omits it for simplicity. If a user needs fallthrough behavior, they can use if/else chains or combine cases: `case "quit", "exit"`.

---

## Impact on Existing Code

All three changes are **purely additive**. No existing valid Kukicha code is affected:

- `function(r Repo) bool` / `func(r Repo) bool` literals continue to work unchanged
- `go func()...()` IIFE pattern continues to work unchanged
- `if/else if` chains continue to work unchanged

The new syntax forms are alternatives, not replacements. Existing tutorials and code need no migration, though they can be updated to use the cleaner forms.

---

## Open Questions

1. **Should expression lambdas support multiple return values?** Proposal: No — use block form for `(T, error)` returns. Expression lambdas should be simple and single-valued.

2. **Should `switch` support type switching (`switch v as type`)?** Proposal: Defer to a future version. Expression and tagless switch cover the immediate tutorial needs. Type switch can follow the interface tutorial work.

3. **Should untyped lambdas be limited to stdlib contexts?** Proposal: No — allow them anywhere the compiler can infer the type from the calling function's signature. But start the implementation with stdlib functions as the test cases.
