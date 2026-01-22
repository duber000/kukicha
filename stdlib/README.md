# Kukicha Standard Library

This directory contains the Kukicha standard library, written in Kukicha without generic syntax.

## Special Transpilation

Files in `stdlib/iter/` use **special transpilation rules** to generate generic Go code:

### How It Works

**Kukicha source (`stdlib/iter/iter.kuki`):**
```kukicha
func Filter(seq iter.Seq, keep func(any) bool) iter.Seq
    return func(yield func(any) bool) bool
        for item in seq
            if keep(item)
                if !yield(item)
                    return false
        return true
```

**Generated Go code:**
```go
func Filter[T any](seq iter.Seq[T], keep func(T) bool) iter.Seq[T] {
    return func(yield func(T) bool) bool {
        for item := range seq {
            if keep(item) {
                if !yield(item) {
                    return false
                }
            }
        }
        return true
    }
}
```

### Transformation Rules

1. `iter.Seq` → `iter.Seq[T any]` (adds generic type parameter)
2. `func(any)` → `func(T)` (replaces any with type parameter)
3. `func(any) any` → `func(T) U` (two type parameters for transformations)

### Why This Approach?

- ✅ **Stdlib written in Kukicha** - Readable, maintainable source code
- ✅ **No generic syntax** - Kukicha language stays simple
- ✅ **Type-safe** - Go's type inference ensures correctness
- ✅ **Works on all types** - Including user-defined types
- ✅ **Zero-cost** - Lazy evaluation with iterators

## Package Structure

```
stdlib/
├── iter/          # Iterator operations (special transpilation)
│   ├── iter.kuki  # Filter, Map, Take, Skip, etc.
│   └── README.md
├── slices/        # Slice operations (wrappers around Go stdlib)
│   └── slices.kuki
├── maps/          # Map operations (wrappers around Go stdlib)
│   └── maps.kuki
└── examples/      # Educational concrete implementations
    └── README.md
```

## Usage Example

```kukicha
import "slices"
import "stdlib/iter"

numbers := list of int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

# Lazy pipeline - zero allocations until Collect
result := numbers
    |> slices.Values()                                    # Go stdlib
    |> iter.Filter(func(n int) bool { return n > 3 })     # Kukicha stdlib
    |> iter.Map(func(n int) int { return n * 2 })         # Kukicha stdlib
    |> iter.Take(5)                                       # Kukicha stdlib
    |> slices.Collect()                                   # Go stdlib

# result = [8, 10, 12, 14, 16]
```

## Available Functions

### iter Package

- `Filter(seq, predicate)` - Keep only matching items
- `Map(seq, transform)` - Transform each item
- `Take(seq, n)` - First n items
- `Skip(seq, n)` - Skip first n items

### Coming Soon

- `FlatMap(seq, fn)` - Map and flatten
- `Zip(seq1, seq2)` - Combine two iterators
- `Enumerate(seq)` - Add indices
- `Chunk(seq, size)` - Group into chunks

## For Library Authors

If you need custom generic functions for your project, write them in Go and import them:

```go
// myproject/helpers/helpers.go
package helpers

func CustomGeneric[T any](items []T) []T {
    // Your implementation
    return items
}
```

```kukicha
# myproject/main.kuki
import "myproject/helpers"

result := helpers.CustomGeneric(myData)
```
