# Function Type Syntax

**Status:** ✅ Implemented
**Date:** 2026-01-21

---

## Overview

Kukicha supports **function types** (also called function signatures or callback types) using the `func` keyword. This allows you to specify that a parameter or variable should be a function.

## Syntax

```kukicha
func(parameter_types...) return_type
```

- **`func`** - keyword that declares a function type
- **`(parameter_types...)`** - parameter types (comma-separated)
- **`return_type`** - optional return type

## Examples

### Basic Function Types

```kukicha
# Simple callback: takes int, returns bool
func Filter(items list of int, predicate func(int) bool) list of int
    result := list of int{}
    for item in items
        if predicate(item)
            result = append(result, item)
    return result

# Multiple parameters: takes two ints, returns int
func Reduce(items list of int, reducer func(int, int) int) int
    result := 0
    for item in items
        result = reducer(result, item)
    return result

# No return type: takes string, returns nothing
func ForEach(items list of string, action func(string))
    for item in items
        action(item)

# Complex types: takes string, returns list of int
func Parse(data string, parser func(string) list of int) list of int
    return parser(data)
```

## What It Compiles To

Kukicha function types compile directly to Go function types:

| Kukicha | Go |
|---------|-----|
| `func(int) bool` | `func(int) bool` |
| `func(int, int) int` | `func(int, int) int` |
| `func(string)` | `func(string)` |
| `func(string) list of int` | `func(string) []int` |

## Complete Example

Here's a real-world example using function types:

```kukicha
# Higher-order function that takes a predicate
func Filter(items list of int, keep func(int) bool) list of int
    result := list of int{}
    for item in items
        if keep(item)
            result = append(result, item)
    return result

# Use it with different predicates
func main()
    numbers := list of int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

    # Filter for even numbers
    evens := Filter(numbers, func(n int) bool
        return n % 2 == 0
    )
    # evens = [2, 4, 6, 8, 10]

    # Filter for numbers > 5
    big := Filter(numbers, func(n int) bool
        return n > 5
    )
    # big = [6, 7, 8, 9, 10]

    # Filter for multiples of 3
    threes := Filter(numbers, func(n int) bool
        return n % 3 == 0
    )
    # threes = [3, 6, 9]
```

## Why `func` for Both?

Kukicha uses `func` for **both** function declarations and function types:

```kukicha
func Add(a int, b int) int              # func for declaration
    return a + b

func Apply(x int, f func(int) int) int  # func for type
    return f(x)
```

This is:
- **Simple** - One keyword to learn, not two
- **Consistent** - Matches Go exactly
- **Clear** - Less mental overhead
- **Concise** - Shorter to type

## Usage in Stdlib

Function types will be used extensively in the Kukicha stdlib for iterator operations:

```kukicha
# From stdlib/iter/iter.kuki
func Filter(seq iter.Seq, keep func(any) bool) iter.Seq
func Map(seq iter.Seq, transform func(any) any) iter.Seq
func Reduce(seq iter.Seq, initial any, reducer func(any, any) any) any
```

When transpiled for stdlib, the codegen intelligently converts `any` to type parameters:

```go
// Generated Go code
func Filter[T any](seq iter.Seq[T], keep func(T) bool) iter.Seq[T]
func Map[T any, U any](seq iter.Seq[T], transform func(T) U) iter.Seq[U]
func Reduce[T any, U any](seq iter.Seq[T], initial U, reducer func(U, T) U) U
```

This gives us:
- ✅ Simple Kukicha syntax (no generics)
- ✅ Powerful Go generics (type-safe)
- ✅ Type inference "just works"

## Common Patterns

### Map Pattern

```kukicha
func Map(items list of int, transform func(int) int) list of int
    result := list of int{}
    for item in items
        result = append(result, transform(item))
    return result
```

### Filter Pattern

```kukicha
func Filter(items list of int, predicate func(int) bool) list of int
    result := list of int{}
    for item in items
        if predicate(item)
            result = append(result, item)
    return result
```

### Reduce Pattern

```kukicha
func Reduce(items list of int, initial int, reducer func(int, int) int) int
    accumulator := initial
    for item in items
        accumulator = reducer(accumulator, item)
    return accumulator
```

### ForEach Pattern

```kukicha
func ForEach(items list of string, action func(string))
    for item in items
        action(item)
```

## Comparison with Other Languages

| Language | Function Type Syntax |
|----------|---------------------|
| **Kukicha** | `func(int) bool` |
| Go | `func(int) bool` |
| TypeScript | `(n: number) => boolean` |
| Rust | `fn(i32) -> bool` |
| Python | `Callable[[int], bool]` |
| Java | `Function<Integer, Boolean>` |

Kukicha's syntax matches Go exactly, making it familiar and easy to learn.

## Tips

1. **Keep it simple** - Function types are just type annotations, not implementations
2. **Use descriptive names** - Name your callback parameters clearly (e.g., `predicate`, `transform`, `action`)
3. **Document expectations** - Use comments to explain what the callback should do
4. **Type safety** - Let the compiler catch errors with proper types

## See Also

- [Kukicha Syntax Reference](kukicha-syntax-v1.0.md)
- [Stdlib Special Transpilation](stdlib-special-transpilation.md)
