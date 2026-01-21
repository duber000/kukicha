# Proposal: English-First Generics and Variadic Syntax

**Status:** Implemented
**Date:** 2026-01-21
**Supersedes:** proposal-stdlib-language-features.md (syntax only, not goals)

## Problem

The original proposal uses Go-style syntax that creates cognitive load for beginners:

```kukicha
# Hard to understand - what is [T]? What is ...interface{}?
func Reverse[T](items list of T) list of T
func Print(args ...interface{})
```

These are familiar to Go/Rust/TypeScript developers but opaque to newcomers.

## Design Philosophy

Kukicha should be **readable as English**. A newcomer should understand code by reading it aloud:

> "Reverse takes a list of *some element* and returns a list of *that same element*"
> "Print takes *many values*"

This proposal introduces **semantic type placeholders** and **natural variadic syntax**.

---

## Part 1: Variadic Parameters with `many`

### Syntax

```kukicha
# Basic variadic (accepts any types)
func Print(many values)
    for i, value in values
        if i > 0
            write " "
        write value
    writeln

# Typed variadic (all arguments must be same type)
func Sum(many numbers int) int
    total := 0
    for _, n in numbers
        total = total + n
    return total

# Mixed parameters
func PrintLabeled(label string, many items)
    write label, ": "
    for i, item in items
        if i > 0
            write ", "
        write item
    writeln
```

### How It Reads

- `many values` reads as "takes many values"
- `many numbers int` reads as "takes many numbers (all integers)"

### Compilation

| Kukicha | Go |
|---------|-----|
| `many values` | `values ...interface{}` |
| `many numbers int` | `numbers ...int` |
| `many items string` | `items ...string` |

### Rules

1. `many` must be the **last parameter**
2. `many name` (no type) = accepts any types (`...interface{}`)
3. `many name type` = all arguments must match type

---

## Part 2: Generics with Semantic Type Placeholders

### Core Concept

Instead of abstract letters like `T`, `U`, `K`, `V`, kukicha uses **meaningful English words** as type placeholders:

| Placeholder | Meaning | Go Equivalent |
|-------------|---------|---------------|
| `element` | Element of a collection | `T any` |
| `item` | Alias for element | `T any` |
| `value` | A value (general purpose) | `T any` |
| `thing` | Informal alias for any | `T any` |
| `key` | Map key type | `K comparable` |
| `result` | Return/output type | `U any` |
| `number` | Numeric type | `T Numeric` |
| `comparable` | Types that can use == | `T comparable` |
| `ordered` | Types that can use < > | `T Ordered` |

### Basic Examples

```kukicha
# Reverse a list - preserves element type
func Reverse(items list of element) list of element
    result := make list of element with length len(items)
    for i, item in items
        result[len(items) - 1 - i] = item
    return result

# First n elements
func First(items list of element, n int) list of element
    if n > len(items)
        n = len(items)
    return items[0:n]

# Find in list - returns index or -1
func IndexOf(items list of comparable, target comparable) int
    for i, item in items
        if item == target
            return i
    return -1
```

### How It Reads

Reading `func Reverse(items list of element) list of element` aloud:

> "Reverse takes a list of elements and returns a list of elements"

The word "element" naturally suggests "whatever type of element you give me."

### Multiple Type Parameters

When you need different type parameters, use different placeholder words:

```kukicha
# Map transforms elements to results
func Map(items list of element, transform func(element) result) list of result
    output := make list of result with length len(items)
    for i, item in items
        output[i] = transform(item)
    return output

# Zip combines two lists
func Zip(left list of element, right list of value) list of pair of element and value
    # ... implementation
```

### How It Reads

Reading `func Map(items list of element, transform func(element) result) list of result`:

> "Map takes a list of elements, a transform function from element to result, and returns a list of results"

The distinct words (`element` vs `result`) clearly indicate different types.

### Constrained Placeholders

Some placeholders carry built-in constraints:

```kukicha
# Sum requires numeric types
func Sum(items list of number) number
    total := number(0)
    for _, item in items
        total = total + item
    return total

# Max requires ordered types (can compare with <, >)
func Max(items list of ordered) ordered
    if len(items) == 0
        panic "empty list"
    best := items[0]
    for _, item in items
        if item > best
            best = item
    return best

# Unique requires comparable types (can use ==)
func Unique(items list of comparable) list of comparable
    seen := empty map of comparable to bool
    result := empty list of comparable
    for _, item in items
        if not (item in seen)
            seen[item] = true
            result = append(result, item)
    return result
```

### Compilation to Go

| Kukicha | Go |
|---------|-----|
| `list of element` | `[]T` where `T any` |
| `list of number` | `[]T` where `T Numeric` |
| `list of comparable` | `[]T` where `T comparable` |
| `func(element) result` | `func(T) U` |

Full example:

```kukicha
func Filter(items list of element, keep func(element) bool) list of element
    result := empty list of element
    for _, item in items
        if keep(item)
            result = append(result, item)
    return result
```

Compiles to:

```go
func Filter[T any](items []T, keep func(T) bool) []T {
    result := []T{}
    for _, item := range items {
        if keep(item) {
            result = append(result, item)
        }
    }
    return result
}
```

---

## Part 3: Complete Standard Library Examples

### `stdlib/print/print.kuki`

```kukicha
leaf print

import "fmt"

# Print values with spaces between, newline at end
func Print(many values)
    for i, value in values
        if i > 0
            fmt.Print(" ")
        fmt.Print(value)
    fmt.Println()

# Print without newline
func Write(many values)
    for i, value in values
        if i > 0
            fmt.Print(" ")
        fmt.Print(value)

# Print with custom separator
func PrintJoined(separator string, many values)
    for i, value in values
        if i > 0
            fmt.Print(separator)
        fmt.Print(value)
    fmt.Println()

# Formatted print
func Printf(format string, many values)
    fmt.Printf(format, values...)
```

### `stdlib/slices/slices.kuki`

```kukicha
leaf slices

# Take first n elements
func First(items list of element, n int) list of element
    if n > len(items)
        n = len(items)
    return items[0:n]

# Take last n elements
func Last(items list of element, n int) list of element
    if n > len(items)
        return items
    return items[len(items) - n:]

# Drop first n elements
func Drop(items list of element, n int) list of element
    if n >= len(items)
        return empty list of element
    return items[n:]

# Reverse order
func Reverse(items list of element) list of element
    result := make list of element with length len(items)
    for i, item in items
        result[len(items) - 1 - i] = item
    return result

# Keep elements matching predicate
func Filter(items list of element, keep func(element) bool) list of element
    result := empty list of element
    for _, item in items
        if keep(item)
            result = append(result, item)
    return result

# Transform elements
func Map(items list of element, transform func(element) result) list of result
    output := make list of result with length len(items)
    for i, item in items
        output[i] = transform(item)
    return output

# Remove duplicates (requires equality check)
func Unique(items list of comparable) list of comparable
    seen := empty map of comparable to bool
    result := empty list of comparable
    for _, item in items
        if not (item in seen)
            seen[item] = true
            result = append(result, item)
    return result

# Sum all elements
func Sum(items list of number) number
    total := number(0)
    for _, n in items
        total = total + n
    return total

# Find maximum
func Max(items list of ordered) ordered
    if len(items) == 0
        panic "Max requires non-empty list"
    best := items[0]
    for _, item in items
        if item > best
            best = item
    return best

# Check if any matches
func Any(items list of element, check func(element) bool) bool
    for _, item in items
        if check(item)
            return true
    return false

# Check if all match
func All(items list of element, check func(element) bool) bool
    for _, item in items
        if not check(item)
            return false
    return true

# Find first matching element
func Find(items list of element, match func(element) bool) (element, bool)
    for _, item in items
        if match(item)
            return item, true
    zero := element{}
    return zero, false
```

### Usage Example

```kukicha
import print
import slices

func main()
    numbers := list of int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

    print.Print("Original:", numbers)

    # Pipeline operations
    result := numbers
        |> slices.Drop(2)
        |> slices.First(5)
        |> slices.Filter(func(n int) bool
            return n > 4
        )
        |> slices.Reverse()

    print.Print("Result:", result)

    total := slices.Sum(numbers)
    print.Print("Sum:", total)

    # Works with strings too!
    words := list of string{"apple", "banana", "cherry"}
    print.Print("Reversed words:", slices.Reverse(words))
```

---

## Part 4: Implementation Plan

### Phase 1: Lexer Updates

Add new tokens:

```go
// internal/lexer/token.go
TOKEN_MANY        // "many" keyword for variadic

// Semantic type placeholders (recognized as special identifiers)
// element, item, value, thing, key, result, number, comparable, ordered
```

### Phase 2: Parser Updates

1. Update `parseParameters()` to handle `many`:
   ```
   many name       → variadic interface{}
   many name type  → variadic type
   ```

2. Track semantic type placeholders in function context:
   - When parsing a function, track which placeholder words are used
   - Each unique placeholder becomes a type parameter in Go output

### Phase 3: AST Updates

```go
// internal/ast/ast.go

type Parameter struct {
    Name     *Identifier
    Type     TypeAnnotation
    Variadic bool           // NEW: true if "many" keyword used
}

type PlaceholderType struct {
    Name       string        // "element", "number", etc.
    Constraint string        // implicit constraint if any
}
```

### Phase 4: Semantic Analysis

1. Validate `many` is only used on last parameter
2. Track type placeholder usage within each function
3. Verify consistent use (same placeholder = same type throughout)

### Phase 5: Code Generation

1. Collect all placeholder types used in function signature
2. Generate Go type parameter list: `[T any, U any]`
3. Map placeholder names to generated type params
4. Output Go generic function

Example transformation:

```kukicha
func Map(items list of element, transform func(element) result) list of result
```

Analysis:
- `element` used 3 times → becomes `T any`
- `result` used 2 times → becomes `U any`

Generated:
```go
func Map[T any, U any](items []T, transform func(T) U) []U
```

---

## Part 5: Comparison

### Before (Go-style)

```kukicha
func Reverse[T](items list of T) list of T

func Map[T, U](items list of T, fn func(T) U) list of U

func Sum[T Numeric](items list of T) T

func Print(args ...interface{})
```

### After (English-first)

```kukicha
func Reverse(items list of element) list of element

func Map(items list of element, transform func(element) result) list of result

func Sum(items list of number) number

func Print(many values)
```

### Benefits

1. **No brackets or symbols** - reads as natural English
2. **Self-documenting** - placeholder names describe purpose
3. **Beginner-friendly** - no need to learn `[T]` notation
4. **Type-safe** - compiles to proper Go generics
5. **Consistent** - extends existing `list of`, `map of` patterns

---

## Design Decisions

### Decision 1: Type Inference Only

Types are **always inferred** from arguments. No explicit type specification syntax.

```kukicha
numbers := list of int{1, 2, 3}
reversed := Reverse(numbers)  # element inferred as int
```

**Rationale**: Simpler, matches the "read and understand" philosophy. If you need explicit types, the function signature documents them.

### Decision 2: Built-in Placeholders Only

Only the **reserved placeholder words** work as type parameters:

| Placeholder | Constraint | Use For |
|-------------|------------|---------|
| `element` | any | Collection elements |
| `item` | any | Alias for element |
| `value` | any | General values, map values |
| `thing` | any | Informal/teaching contexts |
| `key` | comparable | Map keys |
| `result` | any | Return/output types |
| `number` | numeric | Arithmetic operations |
| `comparable` | comparable | Equality checks |
| `ordered` | ordered | Comparison operations |

Custom words like `widget` are **not** type placeholders - they're treated as concrete type names.

**Rationale**: Predictable and learnable. When you see `element`, you know it's special.

### Decision 3: Generic Types Use `of`

Generic structs use `of` syntax to match collections:

```kukicha
# Generic struct
type Box of element
    value element

type Pair of element and value
    first element
    second value

# Usage
myBox := Box of int{value: 42}
myPair := Pair of string and int{first: "age", second: 30}
```

**Rationale**: Consistent with `list of element`, `map of key to value`.

### Decision 4: Standalone Constraint Words

Constrained placeholders are used directly without modifiers:

```kukicha
# 'number' is both placeholder AND constraint
func Sum(items list of number) number

# Not this (too verbose):
# func Sum(items list of any number) any number
```

**Rationale**: Simpler syntax. The word itself implies the constraint.

---

## Conclusion

This proposal replaces abstract symbols with meaningful English words:

| Concept | Go-Style | English-First |
|---------|----------|---------------|
| Variadic | `args ...interface{}` | `many values` |
| Type param | `[T any]` | `element` |
| Numeric constraint | `[T Numeric]` | `number` |
| Multiple params | `[T, U any]` | `element`, `result` |

The result is code that reads naturally:

```kukicha
func Reverse(items list of element) list of element
```

> "Reverse takes a list of elements and returns a list of elements"

This aligns with kukicha's mission: a language that newcomers can understand by reading examples.
