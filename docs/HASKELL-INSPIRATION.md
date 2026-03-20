# Haskell-Inspired Sum Types for Kukicha

## What Are Sum Types?

A **sum type** (also called a "tagged union" or "algebraic data type") is a type that can be *one of* several variants. Each variant can carry different data.

Think of it like a box that is labeled with what's inside:
- A `Shape` is **either** a `Circle` (with a radius) **or** a `Rectangle` (with width and height) — never both, never neither.

Sum types are powerful because the **compiler can check** that you handle every variant. If you add a new variant later and forget to update a `switch`, the compiler tells you.

## Proposed Syntax

### Inline form (short variants)

```kukicha
type Shape
    = Circle(radius float64)
    | Rectangle(width float64, height float64)
    | Triangle(a float64, b float64, c float64)
```

### With methods

```kukicha
type Shape
    = Circle(radius float64)
    | Rectangle(width float64, height float64)

func Area on s Shape float64
    return s |> switch
        when Circle
            return 3.14159 * s.radius * s.radius
        when Rectangle
            return s.width * s.height
```

### Optional: single-line form for simple cases

```kukicha
type Color = Red | Green | Blue
```

This form is useful for simple enumerations where variants carry no data.

## Before and After: Typed Piped Switch Example

The file `examples/typed-piped-switch.kuki` currently uses an interface plus separate struct types to model a `Shape`. Here is what it looks like today and what it could look like with sum types.

### Before (current Kukicha — interface + structs)

```kukicha
interface Shape
    Area() float64

type Circle
    radius float64

func Area on c Circle float64
    return 3.14159 * c.radius * c.radius

type Rectangle
    width float64
    height float64

func Area on r Rectangle float64
    return r.width * r.height

func Describe(s Shape) string
    return s |> switch as v
        when Circle
            return "Circle with radius {v.radius}, area {v.Area()}"
        when Rectangle
            return "{v.width}x{v.height} rectangle, area {v.Area()}"
        otherwise
            return "Unknown shape with area {s.Area()}"
```

Problems with this approach:
- The `interface` and structs are declared separately — nothing ties them together.
- The `otherwise` branch is required because the compiler doesn't know that `Circle` and `Rectangle` are the only possibilities.
- Adding a new variant (e.g., `Triangle`) compiles fine even if you forget to handle it in `Describe`.

### After (proposed — sum type)

```kukicha
type Shape
    = Circle(radius float64)
    | Rectangle(width float64, height float64)

func Describe(s Shape) string
    return s |> switch
        when Circle
            return "Circle with radius {s.radius}, area {s.Area()}"
        when Rectangle
            return "{s.width}x{s.height} rectangle, area {s.Area()}"

func main()
    shapes := list of Shape{
        Circle(5.0),                             # positional — one field
        Rectangle{width: 3.0, height: 4.0},     # named — clearer for two fields
    }

    for shape in shapes
        print(Describe(shape))
```

Improvements:
- **One declaration** defines `Shape` and all its variants.
- **No `otherwise` needed** — the compiler knows all variants are covered.
- **Adding a variant** (e.g., `| Triangle(...)`) produces a compile error everywhere the match is incomplete.

## Exhaustiveness Checking

This is the most valuable feature sum types bring. The compiler enforces that every `switch` on a sum type handles **all** variants.

### Missing branch = compile error

```kukicha
type Result
    = Ok(value string)
    | Err(message string)

func Handle(r Result) string
    return r |> switch
        when Ok
            return r.value
        # COMPILE ERROR: non-exhaustive switch on Result — missing variant: Err
```

### Redundant `otherwise` = warning

When all variants are covered, an `otherwise` branch is dead code:

```kukicha
func Handle(r Result) string
    return r |> switch
        when Ok
            return r.value
        when Err
            return "error: {r.message}"
        otherwise              # WARNING: unreachable — all variants of Result are handled
            return "impossible"
```

### The `otherwise` escape hatch

If you intentionally want to group several variants:

```kukicha
type Traffic = Red | Yellow | Green

func CanGo(light Traffic) bool
    return light |> switch
        when Green
            return true
        otherwise
            return false
```

This is fine — `otherwise` is only warned about when *all* variants already have explicit branches.

## Go Transpilation Strategy

Sum types transpile to an **interface + one struct per variant**, which is the idiomatic Go encoding.

### Kukicha source

```kukicha
type Shape
    = Circle(radius float64)
    | Rectangle(width float64, height float64)
```

### Generated Go

```go
type Shape interface {
    isShape() // sealed marker method
}

type Circle struct {
    Radius float64
}

func (Circle) isShape() {}

type Rectangle struct {
    Width  float64
    Height float64
}

func (Rectangle) isShape() {}
```

Key points:
- The **marker method** (`isShape()`) is unexported, making the interface *sealed* — only variants declared in this package can implement it.
- A `switch` on a sum type transpiles to a Go **type switch**.
- Exhaustiveness is checked at compile time by the Kukicha compiler, not by Go.

### Switch transpilation

Kukicha:
```kukicha
return s |> switch
    when Circle
        return "circle"
    when Rectangle
        return "rect"
```

Generated Go:
```go
switch s := s.(type) {
case Circle:
    return "circle"
case Rectangle:
    return "rect"
}
```

## Compiler Implementation Outline

Adding sum types touches every phase of the compiler. Here is a roadmap of the affected files.

### The `|` vs `|>` Distinction

Sum types use `|` to separate variants. Kukicha already uses `|>` as the pipe operator. These are different tokens that appear in different contexts:

| Token | Name | Where it appears | Example |
|-------|------|-----------------|---------|
| `\|` | Variant separator | Inside `type ... =` declarations only | `type Color = Red \| Green \| Blue` |
| `\|>` | Pipe operator | In expressions only | `data \|> parse() \|> transform()` |

They never overlap:
- `|` only follows `=` or another `| Variant` inside a type declaration
- `|>` only appears between expressions in executable code

The lexer already distinguishes them — `|>` is a two-character token, while `|` alone is a single character. No ambiguity for the compiler. For readers, the context makes it clear: if you see `|` after `=` inside a `type`, it's a sum type variant separator. If you see `|>` in an expression, it's a pipe.

**Alternatives considered:**

| Syntax | Example | Verdict |
|--------|---------|---------|
| `or` keyword | `type Color = Red or Green or Blue` | Reads like English and fits Kukicha's style, but `or` already means logical OR in expressions. Could work since the context (type declaration vs expression) disambiguates, but overloading a keyword risks confusion for beginners. |
| Indentation-only | Each variant on its own indented line | Eliminates the symbol entirely for multi-line forms, but prevents the useful single-line form `type Color = Red \| Green \| Blue`. |
| `variant` keyword | `variant Red, variant Green` | Unambiguous but verbose — defeats the conciseness goal. |

**Decision:** `|` is the right choice. It's the universal convention (Haskell, Rust, OCaml, TypeScript unions), it's already a token in the lexer, and the syntactic contexts are fully disjoint. The single-line form `type Color = Red | Green | Blue` is too valuable to give up, and `|` is the only separator that keeps it readable.

### `|` as Bitwise OR (Go Compatibility)

Go uses `|` for bitwise OR (`flags := os.O_CREATE | os.O_WRONLY`), and Kukicha preserves this. The same `|` token therefore has two meanings depending on context:

| Context | `\|` means | How the parser knows |
|---------|-----------|---------------------|
| `type Color = Red \| Green` | Variant separator | Follows `=` or another variant inside a `type` declaration |
| `flags := os.O_CREATE \| os.O_WRONLY` | Bitwise OR | Appears between expressions outside a type declaration |

The parser disambiguates structurally — when parsing a `type` declaration and it encounters `= Name`, it enters sum-type parsing mode where `|` separates variants. Everywhere else, `|` remains bitwise OR, exactly as Go users expect.

This is the same approach as OCaml and F#, which use `|` for both variant separators and bitwise OR without issue. Several languages (Elm, Gleam, ReScript) also use `|` for variants alongside `|>` for pipes — all three tokens (`|`, `|>`, and bitwise `|`) coexist cleanly because they appear in disjoint syntactic positions.

### 1. Lexer (`internal/lexer/`)

No new tokens required. The `=` and `|` tokens already exist. Variant names reuse `TOKEN_IDENTIFIER`. The lexer already distinguishes `|` (single character) from `|>` (two-character pipe token), so no changes are needed.

### 2. Parser (`internal/parser/`)

Extend `parseTypeDeclaration()` to detect the `= Variant(...)` pattern after the type name. Produce a new AST node (`SumTypeDecl`) containing a list of variants, each with a name and field list.

Single-line form (`type Color = Red | Green | Blue`) is parsed when all variants appear on the same line. Multi-line form uses standard indentation rules.

### 3. AST (`internal/ast/`)

New node types:

```go
type SumTypeDecl struct {
    Name     string
    Variants []SumVariant
}

type SumVariant struct {
    Name   string
    Fields []Field // reuse existing Field type
}
```

### 4. Semantic Analysis (`internal/semantic/`)

- **Registration**: register the sum type and each variant in the type table.
- **Variant constructors**: each variant name is valid as a constructor expression (e.g., `Circle{radius: 5.0}`).
- **Exhaustiveness check**: when a `switch` targets a sum-typed value and uses `when VariantName` branches, collect the set of matched variants. If the set does not cover all variants and there is no `otherwise`, emit a compile error. If all variants are covered and `otherwise` is present, emit a warning.
- **Field access**: inside a `when Circle` branch, the switched variable's type narrows to `Circle`, giving access to `radius`.

### 5. IR / Codegen (`internal/codegen/`)

- **Type emission** (`emit.go`): generate the sealed interface, one struct per variant, and the marker methods.
- **Switch lowering** (`lower.go`): translate a sum-type switch into a Go type switch. Each `when Variant` branch becomes a `case Variant:` with the switched variable re-bound to the concrete type.

### 6. Formatter (`internal/formatter/`)

Add formatting rules for the `= Variant(...) | Variant(...)` syntax, handling both single-line and multi-line forms.

## Design Decisions

Decisions informed by prior art in Coconut (Python), Rust, Haskell, OCaml, Elm, F#, Gleam, and Kotlin.

### 1. Construction: positional and named both supported

Variants with fields support both positional construction (parens) and named construction (braces):

```kukicha
# Positional — concise for 1-2 fields
c := Circle(5.0)
err := Err("something broke")

# Named — clearer for 2+ fields or when intent matters
r := Rectangle{width: 3.0, height: 4.0}
r := Rectangle(3.0, 4.0)    # also works, but less readable with multiple fields
```

Both forms are valid. Positional matches the declaration order in the sum type definition. Named uses the same struct-literal syntax Kukicha already has.

**Rationale:** Every language that started with only one form eventually added the other. Rust requires named fields for struct variants, Haskell uses positional only (and added `NamedFieldPuns` later), Kotlin/Swift support both. Supporting both from day one avoids a breaking change later. The linter/formatter can recommend named for 3+ fields.

### 2. Fieldless variants are bare values

Variants with no fields don't need braces or parentheses:

```kukicha
type Bump = Patch | Minor | Major
type Color = Red | Green | Blue

bump := Patch       # not Patch{} or Patch()
light := Green      # not Green{}
```

**Rationale:** Every language that required empty braces/parens for fieldless variants eventually removed the requirement (or never had it: Haskell, Rust, OCaml, Elm, Coconut). The `{}` is pure noise for enum-style types. Since fieldless variants carry no data, there's nothing to construct — they're values, not constructors.

In generated Go, `Patch` becomes `Patch{}` (Go requires the braces), but Kukicha hides that.

### 3. Field access through narrowed variable (no destructuring yet)

Inside a `when` branch, the switched variable's type narrows to the matched variant. Access fields through the variable:

```kukicha
s |> switch
    when Circle
        return 3.14159 * s.radius * s.radius    # s is narrowed to Circle
    when Rectangle
        return s.width * s.height                # s is narrowed to Rectangle
```

Destructuring in match arms (e.g., `when Circle(r)` binding `r` to the radius) is **not included in the initial implementation**. It can be added later as sugar without breaking existing code.

**Rationale:** Field access through the narrowed variable is consistent with how Kukicha structs already work — no new binding syntax needed. Destructuring is powerful (Coconut, Rust, and Haskell all have it), but it adds parser complexity and a new concept for beginners. Starting without it keeps the feature small and shippable. If demand emerges, `when Circle(r)` can be added in a future release as pure syntax sugar over field access.

### 4. Follow Go's mutability semantics (no immutability layer)

Variants are mutable, following Go's value semantics:

```kukicha
c := Circle(5.0)
c.radius = 10.0    # allowed — same as any struct field
```

**Rationale:** Coconut makes `data` types immutable by default, which makes sense for Python (where everything is a reference and accidental mutation is a common bug). Go's value semantics already protect against most mutation surprises — when you pass a struct by value, the callee gets a copy. Adding an immutability layer on top would diverge from Go conventions and complicate the transpiler for little practical benefit.

### 5. Auto-generated `String()` method

The transpiler generates a `String()` method for each sum type, producing readable output for debugging and logging:

```kukicha
c := Circle(5.0)
print(c)    # prints: Circle(radius: 5.0)

light := Green
print(light)    # prints: Green
```

Generated Go:

```go
func (c Circle) String() string {
    return fmt.Sprintf("Circle(radius: %v)", c.Radius)
}

func (g Green) String() string {
    return "Green"
}
```

**Rationale:** Coconut, Rust (`#[derive(Debug)]`), and Kotlin (`data class`) all auto-generate string representations. It's trivial for the transpiler to emit and immediately useful for `print()` debugging, logging, and error messages. Without it, printing a sum type variant shows the raw Go struct representation, which is less readable.

Fieldless variants print as just their name (`Green`). Variants with fields print as `Name(field: value, ...)`.

## Open Questions

1. **Recursive sum types?** e.g., `type Expr = Lit(value int) | Add(left Expr, right Expr)`. This requires the interface to be defined before the structs reference it — Go handles this naturally.
2. **Pattern matching beyond switch?** e.g., `if s is Circle` as a shorthand for single-variant checks. Could be added later as sugar.
3. **Methods on individual variants vs the sum type?** The `func Area on s Shape` syntax is clean, but should `func Radius on c Circle` also work?
4. **Visibility** — should variants always be exported when the sum type is exported?
