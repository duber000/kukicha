# Proposal: Language Features for Standard Library Implementation

**Status:** Draft
**Date:** 2026-01-21
**Version:** 1.1.0

## Executive Summary

To implement `print` and `slices` as standard library modules written in kukicha (not built-ins), we need to add two critical language features:

1. **Variadic Parameters** - for `print` to accept any number of arguments
2. **Generics/Type Parameters** - for `slices` to work with any type safely

## Current State Analysis

### What Kukicha Currently Supports ✅
- ✅ Interfaces (implicit implementation)
- ✅ Direct Go package imports
- ✅ Explicit type annotations for function signatures
- ✅ Local variable type inference
- ✅ String interpolation
- ✅ Slice syntax with negative indices

### What's Missing ❌
- ❌ **Variadic parameters** (`...args`)
- ❌ **Generic type parameters** (`[T any]`)
- ❌ **Type constraints** for generics

## Problem Statement

### Problem 1: Implementing `print` without Variadic Parameters

**Current state:** Users must use `fmt.Println()` directly from Go:
```kukicha
import "fmt"

func main()
    fmt.Println("Hello")
    fmt.Println("Name:", user.name, "Age:", user.age)
```

**Desired state:** Clean, built-in feeling `print` function:
```kukicha
func main()
    print("Hello")
    print("Name:", user.name, "Age:", user.age)
```

**Why we can't implement it now:**

Without variadic parameters, we'd need:
```kukicha
# ❌ Can't write this in kukicha today:
func Print(args ...interface{})
    # Print all args

# ❌ Would need separate overloads:
func Print1(a interface{})
func Print2(a interface{}, b interface{})
func Print3(a interface{}, b interface{}, c interface{})
# ... up to PrintN
```

This is unmaintainable and defeats the purpose of a clean stdlib.

### Problem 2: Implementing `slices` without Generics

**Current state:** Slice operations must be written for each type or use `interface{}` (losing type safety):

```kukicha
# ❌ Type-unsafe approach:
func Reverse(items list of interface{}) list of interface{}
    # Implementation

# Usage loses type information:
strings := list of string{"a", "b", "c"}
reversed := Reverse(strings)  # Returns list of interface{}, not list of string!
```

**Desired state:** Generic slice operations that preserve type safety:
```kukicha
import slices

# Usage with full type safety:
numbers := list of int{1, 2, 3, 4, 5}
result := numbers
    |> slices.drop(2)      # Still list of int
    |> slices.first(2)     # Still list of int
    |> slices.reverse()    # Still list of int
```

**Why we can't implement it now:**

Without generics, we'd need either:
1. **Type-unsafe implementation** using `interface{}` (loses static type checking)
2. **Code duplication** for every type:
   ```kukicha
   func ReverseInt(items list of int) list of int
   func ReverseString(items list of string) list of string
   func ReverseFloat(items list of float64) list of float64
   # ... for every possible type
   ```

Both approaches are unacceptable for a modern language.

## Proposed Solutions

### Solution 1: Add Variadic Parameters

#### Syntax Proposal

**Kukicha-style syntax (recommended):**
```kukicha
# Declaration
func Print(args ...interface{})
    for discard, arg in args
        fmt.Print(arg)
        fmt.Print(" ")
    fmt.Println()

# Usage
Print("Hello")
Print("Name:", name, "Age:", age)
Print(1, 2, 3, 4, 5)
```

**Alternative English-like syntax:**
```kukicha
# Declaration (more verbose but clearer for beginners)
func Print(args variadic interface{})
    for discard, arg in args
        fmt.Print(arg)
        fmt.Print(" ")
    fmt.Println()
```

#### Implementation Requirements

**Lexer changes:**
- Add `TOKEN_ELLIPSIS` for `...`
- Or add `TOKEN_VARIADIC` keyword for `variadic`

**Parser changes:**
- Extend `parseParameters()` to recognize variadic syntax
- Add `Variadic bool` field to `ast.Parameter`
- Validate: only last parameter can be variadic

**Semantic Analysis:**
- Type check: variadic parameter must be slice-compatible
- Validate: maximum one variadic parameter per function
- Validate: variadic parameter must be last

**Code Generation:**
- Generate Go's `...Type` syntax:
  ```kukicha
  func Print(args ...interface{})
  ```
  becomes:
  ```go
  func Print(args ...interface{}) {
      // ...
  }
  ```

### Solution 2: Add Generic Type Parameters

#### Syntax Proposal

**Kukicha-style syntax (recommended):**
```kukicha
# Single type parameter
func Reverse[T](items list of T) list of T
    result := make list of T with length len(items)
    for i, item in items
        result[len(items) - 1 - i] = item
    return result

# Multiple type parameters
func Map[T, U](items list of T, fn func(T) U) list of U
    result := empty list of U
    for discard, item in items
        result = append(result, fn(item))
    return result

# Type constraints
func Sum[T Numeric](items list of T) T
    total := T(0)
    for discard, item in items
        total = total + item
    return total
```

**Alternative English-like syntax:**
```kukicha
# Using "of" keyword (more verbose)
func Reverse of T (items list of T) list of T
    # Implementation

func Map of T, U (items list of T, fn func(T) U) list of U
    # Implementation
```

#### Built-in Type Constraints

```kukicha
# Comparable types (can use ==, !=)
interface Comparable
    # Built-in marker interface

# Numeric types (can use +, -, *, /)
interface Numeric
    # Built-in marker interface

# Ordered types (can use <, >, <=, >=)
interface Ordered
    # Built-in marker interface
```

#### Implementation Requirements

**Lexer changes:**
- Recognize `[` and `]` in type parameter context (already exists for slices)
- Context-sensitive: after function name, brackets mean type parameters

**Parser changes:**
- Add `TypeParameters []*TypeParameter` to `ast.FunctionDecl`
- Parse type parameter lists: `[T any]`, `[T, U any]`, `[T Comparable]`
- Update type annotation parsing to handle generic types

**AST additions:**
```go
type TypeParameter struct {
    Name       *Identifier      // T, U, etc.
    Constraint TypeAnnotation   // any, Comparable, Numeric, etc.
}
```

**Semantic Analysis:**
- Type constraint validation
- Generic type substitution during type checking
- Ensure type parameters are used consistently

**Code Generation:**
- Generate Go 1.25+ generic syntax:
  ```kukicha
  func Reverse[T](items list of T) list of T
  ```
  becomes:
  ```go
  func Reverse[T any](items []T) []T {
      // ...
  }
  ```

## Implementation Plan

### Phase 1: Variadic Parameters (Simpler)

**Estimated complexity:** Low-Medium
**Estimated time:** 2-3 days
**Priority:** High (needed for `print`)

1. Update lexer for `...` or `variadic` keyword
2. Update parser for variadic parameter syntax
3. Update semantic analysis for variadic validation
4. Update codegen to output Go variadic syntax
5. Add tests for variadic functions
6. Update documentation

### Phase 2: Generics (More Complex)

**Estimated complexity:** High
**Estimated time:** 1-2 weeks
**Priority:** High (needed for `slices` and future stdlib)

1. Update lexer for type parameter context
2. Update parser for generic function declarations
3. Extend AST with type parameter nodes
4. Update semantic analysis for:
   - Type parameter resolution
   - Type constraint checking
   - Generic type substitution
5. Update codegen for Go 1.25+ generic syntax
6. Add comprehensive generic function tests
7. Update documentation

### Phase 3: Standard Library Implementation

**Estimated complexity:** Low (once features are in place)
**Estimated time:** 2-3 days
**Priority:** High

1. Create `stdlib/` directory structure
2. Implement `stdlib/print/print.kuki`
3. Implement `stdlib/slices/slices.kuki`
4. Add stdlib tests
5. Update compiler to auto-include stdlib path
6. Update documentation

## Example Standard Library Files

### `stdlib/print/print.kuki`

```kukicha
leaf print

import "fmt"

# Print values separated by spaces, with newline
func Print(args ...interface{})
    for i, arg in args
        if i > 0
            fmt.Print(" ")
        fmt.Print(arg)
    fmt.Println()

# Print values without newline
func PrintNoNewline(args ...interface{})
    for i, arg in args
        if i > 0
            fmt.Print(" ")
        fmt.Print(arg)

# Print with custom separator
func PrintWith(separator string, args ...interface{})
    for i, arg in args
        if i > 0
            fmt.Print(separator)
        fmt.Print(arg)
    fmt.Println()

# Formatted print (like printf)
func Printf(format string, args ...interface{})
    fmt.Printf(format, args...)
```

### `stdlib/slices/slices.kuki`

```kukicha
leaf slices

# Take first n elements
func First[T](items list of T, n int) list of T
    if n > len(items)
        n = len(items)
    return items[0:n]

# Take last n elements
func Last[T](items list of T, n int) list of T
    if n > len(items)
        return items
    return items[len(items) - n:]

# Drop first n elements
func Drop[T](items list of T, n int) list of T
    if n >= len(items)
        return empty list of T
    return items[n:]

# Drop last n elements
func DropLast[T](items list of T, n int) list of T
    if n >= len(items)
        return empty list of T
    return items[0:len(items) - n]

# Reverse slice
func Reverse[T](items list of T) list of T
    result := make list of T with length len(items)
    for i, item in items
        result[len(items) - 1 - i] = item
    return result

# Remove duplicates (requires Comparable)
func Unique[T Comparable](items list of T) list of T
    seen := empty map of T to bool
    result := empty list of T
    for discard, item in items
        if not (item in seen)
            seen[item] = true
            result = append(result, item)
    return result

# Chunk slice into groups of size n
func Chunk[T](items list of T, size int) list of list of T
    if size <= 0
        panic "chunk size must be positive"

    result := empty list of list of T
    for i from 0 to len(items) step size
        end := i + size
        if end > len(items)
            end = len(items)
        result = append(result, items[i:end])
    return result

# Filter slice by predicate
func Filter[T](items list of T, predicate func(T) bool) list of T
    result := empty list of T
    for discard, item in items
        if predicate(item)
            result = append(result, item)
    return result

# Map slice to new type
func Map[T, U](items list of T, mapper func(T) U) list of U
    result := make list of U with length len(items)
    for i, item in items
        result[i] = mapper(item)
    return result

# Check if any element satisfies predicate
func Any[T](items list of T, predicate func(T) bool) bool
    for discard, item in items
        if predicate(item)
            return true
    return false

# Check if all elements satisfy predicate
func All[T](items list of T, predicate func(T) bool) bool
    for discard, item in items
        if not predicate(item)
            return false
    return true
```

### Usage Example

```kukicha
import print
import slices

func main()
    numbers := list of int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

    # Use print from stdlib
    print.Print("Original:", numbers)

    # Use slices utilities with pipeline
    result := numbers
        |> slices.Drop(2)           # [3, 4, 5, 6, 7, 8, 9, 10]
        |> slices.First(6)          # [3, 4, 5, 6, 7, 8]
        |> slices.Reverse()         # [8, 7, 6, 5, 4, 3]
        |> slices.Filter(func(n int) bool
            return n > 5
        )                           # [8, 7, 6]

    print.Print("Result:", result)
```

## Alternative: Minimal Viable Approach (Without Language Changes)

If we want to ship **something** quickly without language changes:

### Approach 1: Direct Go Package Wrapping

Create thin wrappers that just import Go packages:

```kukicha
# stdlib/print/print.kuki
leaf print

import "fmt"

# Note: Can only pass single interface{} without variadic support
func Print(value interface{})
    fmt.Println(value)

func Printf(format string, a interface{})
    fmt.Printf(format, a)

# For multiple values, users must use fmt directly
# Or we provide up to Print5, etc. (ugly but works)
```

```kukicha
# stdlib/slices/slices.kuki
leaf slices

import "slices" as goslices

# Note: Without generics, these are just thin wrappers around Go's generic slices
# Users must import the Go slices package directly for type safety

# We can't implement these in kukicha without generics
# So we just re-export Go's functions with documentation
```

**Problems with this approach:**
- ❌ Print only works with single values (or we make Print1, Print2, Print3...)
- ❌ Slices package can't be implemented in kukicha - just documentation
- ❌ Defeats the purpose of "stdlib leaves written in kukicha"

### Approach 2: Wait for Go stdlib to stabilize

For slices, Go 1.25's `slices` package already has:
- `slices.Reverse()`
- `slices.Clone()`
- `slices.Equal()`
- `slices.Sort()`
- etc.

We could document that users should import Go's slices package directly:

```kukicha
import "slices"

func main()
    items := list of int{1, 2, 3}
    slices.Reverse(items)  # Go's generic slices.Reverse works!
```

**This is actually viable for slices**, but doesn't help with `print` or custom stdlib functions.

## Recommendation

**Implement both variadic parameters and generics** for the following reasons:

1. **Future-proofing**: These features are essential for any modern standard library
2. **User expectations**: Developers coming from Go/Python/Rust expect these features
3. **Completeness**: Without them, kukicha stdlib will always feel incomplete
4. **Not that complex**: Both features map directly to Go 1.25+ constructs (no invention needed)

### Phased Rollout

**v1.1.0 - Variadic Parameters Only**
- Implement variadic parameters
- Ship basic `print` stdlib module
- Document that slices should use Go's `slices` package directly
- Takes ~3-5 days

**v1.2.0 - Add Generics**
- Implement generics with type constraints
- Ship full `slices` stdlib module
- Enable rich stdlib ecosystem
- Takes ~2-3 weeks

## Questions for Discussion

1. **Syntax preference**: Use `...args` (Go-like) or `variadic args` (English-like)?
2. **Generics syntax**: Use `[T any]` (Go-like) or `of T` (English-like)?
3. **Implementation timeline**: Should we ship v1.1.0 with just variadic parameters first?
4. **Stdlib organization**: Should stdlib be a separate repository or in main repo?

## Conclusion

To fulfill the vision of "standard library leaves written in kukicha," we must add:
1. ✅ **Variadic parameters** (essential for `print`)
2. ✅ **Generics** (essential for `slices` and most other stdlib packages)

Both features are well-understood, map directly to Go 1.25+ constructs, and are essential for a complete standard library. The implementation complexity is manageable, and the benefits are enormous.

**Without these features**, we can only create thin wrappers around Go packages, which defeats the purpose of having a kukicha standard library.

---

**Next Steps:**
1. Review and approve this proposal
2. Choose syntax for variadic parameters and generics
3. Begin implementation of Phase 1 (Variadic Parameters)
4. Follow with Phase 2 (Generics) once Phase 1 is stable
