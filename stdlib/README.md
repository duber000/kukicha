# Kukicha Standard Library

The Kukicha standard library provides **value-add packages** that extend Go's capabilities. Following the CoffeeScript model ("It's just Go"), we only include packages that provide functionality Go lacks or makes awkward.

## Philosophy

Kukicha doesn't wrap Go's standard library - you use it directly with `onerr` syntax:

```kukicha
import "encoding/json"
import "net/http"

# Use Go stdlib directly
data := json.Marshal(user) onerr return error
resp := http.Get(url) onerr return nil, error
```

The Kukicha stdlib only exists where it adds genuine value.

## Packages

### iter - Functional Iterator Operations

Go's `iter.Seq` protocol is low-level. We provide higher-level functional operations:

```kukicha
import "slices"
import "stdlib/iter"

numbers := list of int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

# Lazy pipeline
result := slices.Values(numbers)
    |> iter.Filter(func(n int) bool { return n > 3 })
    |> iter.Map(func(n int) int { return n * 2 })
    |> iter.Take(5)
    |> iter.Collect()

# result = [8, 10, 12, 14, 16]
```

**Functions:** Filter, Map, FlatMap, Take, Skip, Enumerate, Zip, Chunk, Reduce, Collect, Any, All, Find

### slice - Slice Operations

Common slice operations Go lacks:

```kukicha
import "stdlib/slice"

firstThree := slice.First(items, 3)
lastTwo := slice.Last(items, 2)
reversed := slice.Reverse(items)
unique := slice.Unique(items)
chunked := slice.Chunk(items, 5)
```

**Functions:** First, Last, Drop, DropLast, Reverse, Unique, Chunk, Filter, Map, Contains, IndexOf, Concat

### string - String Utilities

Thin wrappers with minimal maintenance burden:

```kukicha
import "stdlib/string"

upper := string.ToUpper("hello")
trimmed := string.TrimSpace("  hello  ")
parts := string.Split("a,b,c", ",")
```

**Functions:** ToUpper, ToLower, TrimSpace, TrimPrefix, TrimSuffix, Split, Join, Fields, Lines, Contains, HasPrefix, HasSuffix, ReplaceAll, EqualFold

## Special Transpilation (iter package only)

The `iter` package uses special transpilation to generate generic Go code without requiring generic syntax in Kukicha:

**Kukicha source:**
```kukicha
func Filter(seq iter.Seq, keep func(any) bool) iter.Seq
    return func(yield func(any) bool) bool
        for item in seq
            if keep(item)
                if !yield(item)
                    return false
        return true
```

**Generated Go:**
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

This keeps Kukicha simple while enabling type-safe generic iteration.

## What's NOT Included

These packages were considered but are better used directly from Go:

| Package | Use Instead |
|---------|-------------|
| bytes | `import "bytes"` + `onerr` |
| io | `import "io"` + `onerr` |
| time | `import "time"` + `onerr` |
| context | `import "context"` |
| json | `import "encoding/json"` + `onerr` |
| http | `import "net/http"` + `onerr` |

See [kukicha-design-philosophy.md](../docs/kukicha-design-philosophy.md) for rationale.

## Package Structure

```
stdlib/
├── iter/          # Iterator operations (special transpilation)
│   ├── iter.kuki
│   └── iter_test.kuki
├── slice/         # Slice operations
│   └── slice.kuki
├── string/        # String utilities
│   └── string.kuki
└── README.md
```
