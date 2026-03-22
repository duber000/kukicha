# Enum Research: Design Proposal for Kukicha

## What Are Enums?

An **enumeration** (enum) is a type that restricts a variable to a fixed set of named values. Instead of using a raw `int` or `string` and hoping the programmer remembers which values are valid, an enum makes the compiler enforce it.

For example, instead of:

```go
status := 200  // What does 200 mean? Is 999 valid?
```

You write:

```kukicha
enum Status
    OK = 200
    NotFound = 404
    Error = 500

status := Status.OK  # Clear, discoverable, type-safe
```

Enums are valuable because they:
- **Prevent invalid values** — the compiler rejects `Status(999)` in well-typed code
- **Make code self-documenting** — `Status.NotFound` is clearer than `404`
- **Enable exhaustiveness checking** — the compiler can warn if a `switch` misses a case
- **Improve IDE support** — autocomplete shows all valid values

## How Go Handles Enums Today

Go has no `enum` keyword. The idiomatic pattern uses `const` blocks with `iota`:

```go
type Status int

const (
    StatusOK       Status = 200
    StatusNotFound Status = 404
    StatusError    Status = 500
)
```

Or with `iota` for auto-incrementing:

```go
type Color int

const (
    Red   Color = iota  // 0
    Green               // 1
    Blue                // 2
)
```

### Limitations of Go's Pattern

1. **No exhaustiveness checking** — `switch` on a `Status` compiles fine with missing cases
2. **No namespace** — constants are package-level (`StatusOK`, not `Status.OK`)
3. **No enforcement** — `Status(999)` is perfectly legal Go
4. **Verbose naming** — the type name must be prefixed manually (`StatusOK`, `StatusNotFound`)
5. **`iota` is implicit** — beginners find it confusing that `Green` has no explicit `= 1`

## What Developers Want

The **2025 Go Developer Survey** found that 65% of Go developers' favorite languages have type-safe enums, and enums are the **#2 most-wanted feature** (after better error handling, which Kukicha already addresses with `onerr`).

Go's core team has acknowledged the demand but has not committed to any timeline or syntax. Their conservative approach ("stability over novelty") means Go may not get enums for years, if ever.

## Proposed Kukicha Syntax

### Basic Value Enums

```kukicha
enum Status
    OK = 200
    NotFound = 404
    Error = 500
```

Rules:
- Each case must have an explicit value (no hidden `iota` magic — matches Kukicha's "no magic" philosophy)
- The underlying type is inferred from the values (`int` for integers, `string` for strings)
- Cases are accessed with dot syntax: `Status.OK`

### String Enums

```kukicha
enum LogLevel
    Debug = "debug"
    Info = "info"
    Warn = "warn"
    Error = "error"
```

### Usage with Switch/When

```kukicha
func handleStatus(s Status) string
    return s |> switch
        when Status.OK
            return "Success"
        when Status.NotFound
            return "Not found"
        when Status.Error
            return "Server error"
```

### Usage in Types

```kukicha
type Response
    status Status
    body string

resp := Response{status: Status.OK, body: "hello"}
```

## Generated Go Code

### Input (Kukicha)

```kukicha
enum Status
    OK = 200
    NotFound = 404
    Error = 500
```

### Output (Go)

```go
type Status int

const (
    StatusOK       Status = 200
    StatusNotFound Status = 404
    StatusError    Status = 500
)
```

The generated Go follows the standard idiomatic pattern. The type name is prepended to each case name (`OK` becomes `StatusOK`). This is invisible to Kukicha users, who always write `Status.OK`.

### Dot Access Translation

In codegen, `Status.OK` is translated to `StatusOK`. This is a straightforward rewrite in `exprToString` — when the left side of a field access is a known enum type, concatenate the type name with the field name.

## Integration with Existing Features

### Pattern Matching (switch/when)

Kukicha's existing `switch`/`when` syntax works naturally with enums:

```kukicha
status |> switch
    when Status.OK
        print("ok")
    when Status.NotFound
        print("not found")
    otherwise
        print("other")
```

### Piped Switch

```kukicha
result := status |> switch
    when Status.OK
        return "ok"
    when Status.Error
        return "error"
    otherwise
        return "unknown"
```

### Comparison

```kukicha
if status equals Status.OK
    print("success")
```

### Function Parameters

```kukicha
func process(level LogLevel)
    if level equals LogLevel.Debug
        print("debug mode")
```

## Implementation Scope

### 1. Lexer (`internal/lexer/token.go`)

Add a new token type:

```go
TOKEN_ENUM  // 'enum' keyword
```

Add to the `keywords` map:

```go
"enum": TOKEN_ENUM,
```

### 2. AST (`internal/ast/ast.go`)

Define two new nodes:

```go
// EnumDecl represents an enum declaration.
//
//  enum Status
//      OK = 200
//      NotFound = 404
type EnumDecl struct {
    Token      lexer.Token   // The 'enum' token
    Name       *Identifier
    Cases      []*EnumCase
    Directives []Directive
}

// EnumCase is a single named value in an enum.
type EnumCase struct {
    Name  *Identifier
    Value Expression  // Must be a literal (integer or string)
}
```

`EnumDecl` implements `Declaration` (like `TypeDecl` and `ConstDecl`).

### 3. Parser (`internal/parser/parser_decl.go`)

Add `case lexer.TOKEN_ENUM:` in `parseDeclaration()`, calling a new `parseEnumDecl()` method:

```
parseEnumDecl:
  1. consume TOKEN_ENUM
  2. parse name identifier
  3. consume TOKEN_INDENT
  4. loop: parse each case (identifier = literal)
  5. consume TOKEN_DEDENT
```

This follows the same pattern as `parseConstDecl()` for grouped const blocks.

### 4. Semantic Analysis (`internal/semantic/semantic_declarations.go`)

In `collectDeclarations()` and `analyzeDeclarations()`:

- Register the enum type as a named type in the symbol table
- Store enum case names and values for validation
- Validate: all values must be the same type (all int or all string)
- Validate: no duplicate case names
- Validate: no duplicate values
- Register enum cases as package-level constants

For field access validation in `semantic_calls.go`:
- When `expr.Object` resolves to an enum type and `expr.Field` is a valid case name, mark it as valid

### 5. Codegen (`internal/codegen/codegen_decl.go`)

Add `case *ast.EnumDecl:` in the declaration switch, calling `generateEnumDecl()`:

```go
func (g *Generator) generateEnumDecl(decl *ast.EnumDecl) {
    // Emit: type Status int  (or string)
    // Emit: const ( StatusOK Status = 200 ... )
}
```

In `exprToString` (for `FieldAccessExpr`):
- When the object resolves to an enum type, emit `TypeNameCaseName` instead of `TypeName.CaseName`

### 6. Formatter (`internal/formatter/printer.go`)

Add `case *ast.EnumDecl:` to print enum declarations with proper indentation.

### Key Files Summary

| File | Change |
|------|--------|
| `internal/lexer/token.go` | Add `TOKEN_ENUM`, add `"enum"` to keywords |
| `internal/ast/ast.go` | Add `EnumDecl`, `EnumCase` nodes |
| `internal/parser/parser_decl.go` | Add `parseEnumDecl()` |
| `internal/semantic/semantic_declarations.go` | Collect + analyze enum declarations |
| `internal/semantic/semantic_calls.go` | Resolve `EnumType.Case` field access |
| `internal/codegen/codegen_decl.go` | Generate Go `type` + `const` block |
| `internal/codegen/codegen_expr.go` | Translate `Status.OK` to `StatusOK` |
| `internal/codegen/codegen.go` | Add case in declaration dispatch |
| `internal/formatter/printer.go` | Format enum declarations |

## Risk Assessment

### Low Risk: Conflict with Go

Even if Go eventually adds enums, they will almost certainly generate `const`+`iota` patterns under the hood. Kukicha's generated Go code would just need codegen updates — the Kukicha syntax itself wouldn't need to change. Go has shown a strong preference for backward compatibility, so existing `const` patterns will remain valid.

### Medium Risk: Premature Design

If Go adds **data-carrying variants** (Rust-style `enum Result { Ok(T), Err(E) }`), Kukicha's simple value enums would need extension. However:

- Simple value enums are useful regardless — they cover the most common use case (status codes, log levels, directions, colors)
- Extension is additive — we can add data-carrying variants later without breaking existing enums
- Go's conservative approach makes complex enums unlikely in the near term

### Low Risk: Naming Collision

`enum` is not a Go keyword, so using it in generated code is not an issue. The generated output uses standard `type` + `const` patterns.

## Open Design Questions

### 1. Auto-Increment vs Explicit Values

**Current proposal:** All values must be explicit.

```kukicha
# This is the proposed syntax — explicit values required
enum Color
    Red = 0
    Green = 1
    Blue = 2
```

**Alternative:** Support auto-increment (like Go's `iota`):

```kukicha
enum Color
    Red      # 0
    Green    # 1
    Blue     # 2
```

**Recommendation:** Start with explicit values only. Kukicha's philosophy is "no magic" — auto-increment hides the actual values and makes insertion-order bugs possible. Auto-increment could be added later as a convenience feature if users request it.

### 2. Exhaustiveness Checking

Should the compiler warn when a `switch` on an enum type is missing cases?

```kukicha
# Should this warn about missing Status.Error case?
status |> switch
    when Status.OK
        print("ok")
    when Status.NotFound
        print("not found")
```

**Recommendation:** Add exhaustiveness checking, but only when there is no `otherwise` clause. If `otherwise` is present, the programmer has explicitly handled the remaining cases. This matches Rust's approach and provides real safety without being annoying.

**Implementation:** In `codegen_stmt.go` (or semantic), when generating a `switch` where the expression type is a known enum:
- Collect all `when` case values
- Compare against all enum cases
- If missing cases and no `otherwise`, emit a compiler warning

### 3. Underlying Type Specification

**Current proposal:** Infer from values (all integers → `int`, all strings → `string`).

**Alternative:** Allow explicit underlying type:

```kukicha
enum Port int
    HTTP = 80
    HTTPS = 443
```

**Recommendation:** Start with inference only. The inferred type covers most cases. Explicit types can be added later if needed (e.g., for `int32` or `uint8`).

### 4. String Representation

Should enums have an automatic `String()` method?

```kukicha
print(Status.OK)  # Should this print "OK" or "200"?
```

**Recommendation:** Generate a `String()` method that returns the case name. This matches user expectations and is more useful for debugging than the raw value. The raw value is still accessible since the enum is a typed constant.

### 5. Enum Methods

Should users be able to define methods on enum types?

```kukicha
func IsClientError on s Status bool
    return s >= 400 and s < 500
```

**Recommendation:** Yes — since enums generate a named type (`type Status int`), Go methods on that type work naturally. No special handling needed beyond what Kukicha already does for methods on named types.

### 6. JSON Serialization

Should enums serialize as their value or their name?

**Recommendation:** By default, serialize as the value (this is what Go does with typed constants). Consider adding a `# kuki:json name` directive later for name-based serialization.
