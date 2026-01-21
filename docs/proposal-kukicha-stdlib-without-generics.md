# Kukicha Stdlib Without Generics: Multiple Approaches

**Date:** 2026-01-21
**Question:** Can we write Kukicha stdlib in Kukicha (not Go) without generic syntax?
**Answer:** YES - Multiple viable approaches!

---

## The Core Challenge

You want:
1. ✅ No generic syntax in Kukicha language
2. ✅ Stdlib written in Kukicha (not Go)
3. ✅ Type-safe operations (Filter, Map, Sort, etc.)
4. ✅ Leverage Go 1.22+ features

Let's explore **five approaches** to achieve this.

---

## Approach 1: Direct Import of Go's Generic Stdlib (Simplest)

### The Insight

Go 1.23+ has generic functions in `slices`, `maps`, and `iter` packages. Kukicha can **import and use them directly** without any wrapper!

### How It Works

**Kukicha code:**
```kukicha
import "slices"

numbers := list of int{3, 1, 4, 1, 5, 9}

# Call Go's generic functions - no Kukicha stdlib needed!
slices.Sort(numbers)
slices.Reverse(numbers)

evens := slices.DeleteFunc(numbers, func(n int) bool {
    return n % 2 != 0
})

if slices.Contains(numbers, 5)
    print("Found 5!")
```

**Transpiles directly to:**
```go
import "slices"

numbers := []int{3, 1, 4, 1, 5, 9}

slices.Sort(numbers)
slices.Reverse(numbers)

evens := slices.DeleteFunc(numbers, func(n int) bool {
    return n%2 != 0
})

if slices.Contains(numbers, 5) {
    fmt.Println("Found 5!")
}
```

**Go's type inference handles everything!**

### Available Functions

**From `slices` package (Go 1.21+):**
- `Sort(items)` - Sort any comparable slice
- `Reverse(items)` - Reverse any slice
- `Contains(items, target)` - Check membership
- `Index(items, target)` - Find index
- `Equal(a, b)` - Compare slices
- `Clone(items)` - Deep copy
- `Compact(items)` - Remove consecutive duplicates
- `Delete(items, i, j)` - Remove range
- `DeleteFunc(items, keep)` - Filter in-place
- `Insert(items, i, values...)` - Insert elements
- `Min(items)` - Find minimum
- `Max(items)` - Find maximum

**From `slices` package (Go 1.23+ with iterators):**
- `All(items)` - Iterator with index and value
- `Values(items)` - Iterator of values only
- `Collect(iter)` - Collect iterator to slice
- `Sorted(iter)` - Collect and sort
- `SortedFunc(iter, cmp)` - Collect and sort with comparator

**From `maps` package (Go 1.23+):**
- `Keys(m)` - Iterator of keys
- `Values(m)` - Iterator of values
- `Clone(m)` - Deep copy
- `Equal(m1, m2)` - Compare maps
- `DeleteFunc(m, keep)` - Filter in-place
- `Collect(iter)` - Build map from iterator

### Pros and Cons

**Pros:**
- ✅ **Zero Kukicha stdlib code needed**
- ✅ **Go's optimized implementations** (battle-tested)
- ✅ **Fully type-safe** via Go's type inference
- ✅ **Always up-to-date** with Go improvements
- ✅ **Works today** - no development needed

**Cons:**
- ❌ **Not "written in Kukicha"** - it's Go stdlib
- ❌ **Go naming conventions** - `slices.Sort` not `Sort`
- ❌ **Some functions missing** - no `Filter`, `Map` (only `DeleteFunc`)
- ❌ **Learning materials** reference Go docs

### Verdict

**Best for:** Production code, minimal maintenance

---

## Approach 2: Thin Kukicha Wrappers Around Go Stdlib

### The Idea

Write minimal Kukicha functions that call Go stdlib, providing nicer names and filling gaps.

### Implementation

**File: `stdlib/slices/slices.kuki`**
```kukicha
import "slices" as goslices

# ============================================================================
# Sorting (pass-through to Go)
# ============================================================================

func Sort(items list of any)
    goslices.Sort(items)

func SortDesc(items list of any)
    goslices.SortFunc(items, func(a any, b any) int {
        if a > b
            return -1
        if a < b
            return 1
        return 0
    })

func Reverse(items list of any)
    goslices.Reverse(items)

# ============================================================================
# Searching (pass-through to Go)
# ============================================================================

func Contains(items list of any, target any) bool
    return goslices.Contains(items, target)

func Index(items list of any, target any) int
    return goslices.Index(items, target)

# ============================================================================
# Operations not in Go stdlib - implement in Kukicha
# ============================================================================

func Filter(items list of any, keep func(any) bool) list of any
    result := empty list of any
    for _, item in items
        if keep(item)
            result = append(result, item)
    return result

func Map(items list of any, transform func(any) any) list of any
    result := make list of any with length len(items)
    for i, item in items
        result[i] = transform(item)
    return result

func Take(items list of any, n int) list of any
    if n >= len(items)
        return items
    return items[0:n]

func Drop(items list of any, n int) list of any
    if n >= len(items)
        return empty list of any
    return items[n:]

func First(items list of any) any
    if len(items) == 0
        panic "empty list"
    return items[0]

func Last(items list of any) any
    if len(items) == 0
        panic "empty list"
    return items[-1]
```

### Usage

```kukicha
import "stdlib/slices"

numbers := list of int{3, 1, 4, 1, 5, 9}

# Use Kukicha stdlib wrappers
slices.Sort(numbers)
slices.Reverse(numbers)

evens := slices.Filter(numbers, func(n int) bool {
    return n % 2 == 0
})

doubled := slices.Map(evens, func(n int) int {
    return n * 2
})
```

### How This Transpiles

**Kukicha stdlib functions using `any`:**
```kukicha
func Filter(items list of any, keep func(any) bool) list of any
```

**Transpile to Go:**
```go
func Filter(items []any, keep func(any) bool) []any
```

**Problem:** This loses type safety! The function signature uses `any`.

**BUT:** When users call it:
```kukicha
evens := slices.Filter(numbers, func(n int) bool { return n % 2 == 0 })
```

The Go compiler sees:
```go
evens := slices.Filter(numbers, func(n int) bool { return n % 2 == 0 })
```

And this **won't compile** because:
- `numbers` is `[]int`
- `Filter` expects `[]any`
- Need explicit conversion

**This approach has a fatal flaw!**

---

## Approach 3: Type-Specific Implementations (Code Generation)

### The Idea

Write stdlib once in Kukicha, generate versions for each common type.

### Template Implementation

**File: `stdlib/slices/slices.kuki.template`**
```kukicha
# TYPE will be replaced: int, string, float, bool, etc.

func Filter_TYPE(items list of TYPE, keep func(TYPE) bool) list of TYPE
    result := empty list of TYPE
    for _, item in items
        if keep(item)
            result = append(result, item)
    return result

func Map_TYPE_TYPE(items list of TYPE, transform func(TYPE) TYPE) list of TYPE
    result := make list of TYPE with length len(items)
    for i, item in items
        result[i] = transform(item)
    return result

func Take_TYPE(items list of TYPE, n int) list of TYPE
    if n >= len(items)
        return items
    return items[0:n]
```

### Code Generation

**Build script generates:**
```kukicha
# slices/int.kuki
func FilterInt(items list of int, keep func(int) bool) list of int
    result := empty list of int
    for _, item in items
        if keep(item)
            result = append(result, item)
    return result

# slices/string.kuki
func FilterString(items list of string, keep func(string) bool) list of string
    result := empty list of string
    for _, item in items
        if keep(item)
            result = append(result, item)
    return result

# slices/float.kuki
func FilterFloat(items list of float, keep func(float) bool) list of float
    result := empty list of float
    for _, item in items
        if keep(item)
            result = append(result, item)
    return result

# ... etc for all common types
```

### Usage

```kukicha
import "stdlib/slices"

numbers := list of int{1, 2, 3, 4, 5}
names := list of string{"Alice", "Bob", "Charlie"}

# Type-specific versions
evens := slices.FilterInt(numbers, func(n int) bool { return n % 2 == 0 })
shortNames := slices.FilterString(names, func(s string) bool { return len(s) < 5 })
```

### Pros and Cons

**Pros:**
- ✅ **Written in Kukicha** - source code is readable
- ✅ **Type-safe** - each version has concrete types
- ✅ **Educational** - users can read Kukicha implementations

**Cons:**
- ❌ **Code explosion** - 10 types × 20 functions = 200 functions
- ❌ **Verbose API** - users must know `FilterInt` vs `FilterString`
- ❌ **Limited types** - only works for pre-generated types
- ❌ **Custom types not supported** - can't filter `list of User`

### Verdict

**Not practical** - too many generated functions, doesn't work for user types.

---

## Approach 4: Iterator-Based Lazy Operations (BEST APPROACH!)

### The Key Insight

Go 1.23's `iter.Seq[T]` makes this work! We can write Kukicha stdlib that:
1. Uses Go's `iter.Seq` type
2. Implements lazy operations
3. Works with ANY type through Go's type inference

### How It Works

**The Magic:** `iter.Seq` is already generic in Go, and Kukicha doesn't need to know that!

### Implementation

**File: `stdlib/iter/iter.kuki`**
```kukicha
import "iter"

# ============================================================================
# Lazy Transformations - Work on ANY type via iter.Seq
# ============================================================================

# Filter returns an iterator that yields only matching items
# Note: We don't specify the type - Go's iter.Seq handles it!
func Filter(seq iter.Seq, keep func(any) bool) iter.Seq
    return func(yield func(any) bool) bool
        for item in seq
            if keep(item)
                if !yield(item)
                    return false
        return true

# Map returns an iterator that transforms each item
func Map(seq iter.Seq, transform func(any) any) iter.Seq
    return func(yield func(any) bool) bool
        for item in seq
            if !yield(transform(item))
                return false
        return true

# Take returns an iterator of the first n items
func Take(seq iter.Seq, n int) iter.Seq
    return func(yield func(any) bool) bool
        count := 0
        for item in seq
            if count >= n
                return true
            if !yield(item)
                return false
            count = count + 1
        return true

# Skip returns an iterator skipping the first n items
func Skip(seq iter.Seq, n int) iter.Seq
    return func(yield func(any) bool) bool
        count := 0
        for item in seq
            if count >= n
                if !yield(item)
                    return false
            count = count + 1
        return true
```

**Wait... this still uses `any`!**

**BUT HERE'S THE TRICK:** When transpiling, we can make these **generic in the Go output!**

### Special Transpilation Rule for Stdlib

When transpiling `stdlib/iter/iter.kuki`:

**Kukicha source:**
```kukicha
func Filter(seq iter.Seq, keep func(any) bool) iter.Seq
```

**Transpiled Go (with special stdlib rule):**
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

### The Transpilation Rules

**For stdlib packages only:**
1. `iter.Seq` → `iter.Seq[T any]` (add type parameter)
2. `func(any)` → `func(T)` (replace any with T)
3. `func(any) any` → `func(T) U` (add second type parameter if different return)

**This is ONLY for stdlib** - user code doesn't get this treatment.

### Usage

```kukicha
import "slices"
import "stdlib/iter"

numbers := list of int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

# Lazy pipeline - written in Kukicha stdlib!
result := numbers
    |> slices.Values()                              # Go stdlib: iter.Seq[int]
    |> iter.Filter(func(n int) bool { return n > 3 })  # Kukicha stdlib: iter.Seq[int]
    |> iter.Map(func(n int) int { return n * 2 })      # Kukicha stdlib: iter.Seq[int]
    |> iter.Take(5)                                     # Kukicha stdlib: iter.Seq[int]
    |> slices.Collect()                              # Go stdlib: []int

# Fully type-safe, zero allocations until Collect!
```

### Pros and Cons

**Pros:**
- ✅ **Written in Kukicha** - readable, maintainable
- ✅ **Type-safe** - Go's type inference handles it
- ✅ **Works for ALL types** - including user-defined types
- ✅ **Zero-cost abstractions** - lazy evaluation
- ✅ **Leverages Go 1.23+** - built on `iter.Seq`
- ✅ **No code explosion** - one implementation for all types
- ✅ **Clean API** - `Filter`, `Map`, not `FilterInt`, `FilterString`

**Cons:**
- ⚠️ **Requires special transpilation rule** for stdlib
- ⚠️ **Only works for iterator-based operations**
- ⚠️ **Some magic** - `any` in source becomes `T` in output

### Verdict

**⭐ RECOMMENDED** - This is the sweet spot!

---

## Approach 5: Hybrid Approach (Pragmatic)

### Combine the Best of Multiple Approaches

1. **Import Go stdlib directly** for what it provides well
2. **Kukicha stdlib with special transpilation** for iterator operations
3. **Concrete Kukicha functions** for educational examples

### Structure

```
stdlib/
├── iter/
│   ├── iter.kuki          # Iterator operations (special transpilation)
│   └── README.md          # Explain lazy evaluation
├── slices/
│   ├── slices.kuki        # Thin wrappers + additions
│   └── README.md          # Guide to Go's slices package
├── maps/
│   ├── maps.kuki          # Thin wrappers + additions
│   └── README.md          # Guide to Go's maps package
└── examples/
    ├── filter_int.kuki    # Educational: concrete Filter for ints
    ├── map_string.kuki    # Educational: concrete Map for strings
    └── README.md          # Learning path
```

### Iterator Operations (Special Transpilation)

**File: `stdlib/iter/iter.kuki`**
```kukicha
import "iter"

# These use special transpilation to become generic in Go
func Filter(seq iter.Seq, keep func(any) bool) iter.Seq
func Map(seq iter.Seq, transform func(any) any) iter.Seq
func Take(seq iter.Seq, n int) iter.Seq
func Skip(seq iter.Seq, n int) iter.Seq
func FlatMap(seq iter.Seq, fn func(any) iter.Seq) iter.Seq
func Zip(seq1 iter.Seq, seq2 iter.Seq) iter.Seq
```

### Slice Operations (Direct Go Import)

**File: `stdlib/slices/slices.kuki`**
```kukicha
# Re-export Go's functions for convenience
import "slices" as goslices

# Sorting
func Sort(items list of any)
    goslices.Sort(items)

func Reverse(items list of any)
    goslices.Reverse(items)

# Searching
func Contains(items list of any, target any) bool
    return goslices.Contains(items, target)

func Index(items list of any, target any) int
    return goslices.Index(items, target)

# Iterators
func Values(items list of any) iter.Seq
    return goslices.Values(items)

func Collect(seq iter.Seq) list of any
    return goslices.Collect(seq)
```

### Educational Examples (Concrete Types)

**File: `stdlib/examples/filter_int.kuki`**
```kukicha
# Educational implementation - shows how Filter works
# Users can read this to learn, but use iter.Filter in production

func FilterIntExample(items list of int, keep func(int) bool) list of int
    result := empty list of int
    for _, item in items
        if keep(item)
            result = append(result, item)
    return result
```

### Usage

```kukicha
# Production code - use iter for efficiency
import "stdlib/iter"
import "slices"

numbers := list of int{1, 2, 3, 4, 5}

result := numbers
    |> slices.Values()                              # Go stdlib
    |> iter.Filter(func(n int) bool { return n > 2 })  # Kukicha stdlib
    |> iter.Map(func(n int) int { return n * 2 })      # Kukicha stdlib
    |> slices.Collect()                              # Go stdlib

# Learning - read concrete examples
import "stdlib/examples"

evens := examples.FilterIntExample(numbers, func(n int) bool {
    return n % 2 == 0
})
```

### Pros and Cons

**Pros:**
- ✅ **Best of both worlds** - efficiency + education
- ✅ **Written in Kukicha** - core stdlib is Kukicha code
- ✅ **Leverages Go** - uses stdlib where appropriate
- ✅ **Educational path** - concrete examples for learning
- ✅ **Type-safe** - through Go's type inference

**Cons:**
- ⚠️ **Requires special transpilation** for iter package
- ⚠️ **Two APIs** - `iter.Filter` (lazy) vs `examples.FilterInt` (eager)

### Verdict

**⭐ HIGHLY RECOMMENDED** - Balances pragmatism with education.

---

## Comparison Table

| Approach | Kukicha Code | Type Safe | Works on User Types | Maintenance | Educational |
|----------|--------------|-----------|---------------------|-------------|-------------|
| **1. Direct Go Import** | No | ✅ | ✅ | ✅ Minimal | ❌ Go docs |
| **2. Thin Wrappers** | Yes | ❌ (`any`) | ❌ | ⚠️ Medium | ⚠️ Some |
| **3. Code Generation** | Yes | ✅ | ❌ | ❌ High | ✅ Good |
| **4. Iterator + Special Transpile** | Yes | ✅ | ✅ | ✅ Low | ✅ Good |
| **5. Hybrid** | Yes | ✅ | ✅ | ✅ Low | ✅ Excellent |

---

## Recommendation: Approach 4 or 5

### Primary Recommendation: Approach 5 (Hybrid)

**Structure:**
1. **Core iterator operations** in Kukicha with special transpilation
2. **Direct imports** of Go's slices/maps packages where applicable
3. **Educational examples** showing concrete implementations

**Benefits:**
- ✅ Stdlib is written in Kukicha (user-facing goal)
- ✅ Type-safe through Go's inference
- ✅ Works on all types including user-defined
- ✅ Educational path for learners
- ✅ Leverages Go 1.23+ features (iterators)
- ✅ Low maintenance

**Implementation:**

1. **Add special transpilation rule** for stdlib iter package:
   - `iter.Seq` → `iter.Seq[T]`
   - `func(any)` → `func(T)`
   - Apply ONLY to files in `stdlib/iter/`

2. **Write Kukicha stdlib**:
   - `stdlib/iter/iter.kuki` - Iterator operations
   - `stdlib/slices/slices.kuki` - Thin wrappers
   - `stdlib/maps/maps.kuki` - Thin wrappers
   - `stdlib/examples/` - Educational concrete examples

3. **Documentation**:
   - Explain that stdlib uses special transpilation
   - Show both lazy (iter) and eager (examples) approaches
   - Guide users from examples → production stdlib

---

## The Special Transpilation Rule

### What It Does

For **stdlib packages only** (specifically `stdlib/iter/*.kuki`):

**Input (Kukicha):**
```kukicha
func Filter(seq iter.Seq, keep func(any) bool) iter.Seq
    return func(yield func(any) bool) bool
        for item in seq
            if keep(item)
                if !yield(item)
                    return false
        return true
```

**Output (Go):**
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

### Rules

1. Detect `iter.Seq` type → Add generic type parameter `[T any]`
2. Replace `func(any)` with `func(T)` in that function
3. If return type is different, use `[T any, U any]`
4. ONLY apply to `stdlib/iter/*.kuki` files

### Is This Acceptable?

**Arguments for:**
- ✅ Keeps Kukicha syntax simple (no generics for users)
- ✅ Stdlib is still readable Kukicha code
- ✅ One-time implementation cost (special case in codegen)
- ✅ Type-safe output
- ✅ Works for all types

**Arguments against:**
- ❌ "Magic" transpilation - source doesn't match output
- ❌ Only works for stdlib (users can't do this)
- ❌ Adds complexity to compiler

**Verdict:** The trade-off is worth it for a drastically simpler language.

---

## Complete Example: Hybrid Approach in Action

### Stdlib Implementation

**File: `stdlib/iter/filter.kuki`**
```kukicha
import "iter"

func Filter(seq iter.Seq, keep func(any) bool) iter.Seq
    return func(yield func(any) bool) bool
        for item in seq
            if keep(item)
                if !yield(item)
                    return false
        return true
```

**Transpiles to (with special rule):**
```go
package iter

import "iter"

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

### User Code

```kukicha
import "slices"
import "stdlib/iter"

type User
    id int
    name string
    active bool

users := list of User{
    User{id: 1, name: "Alice", active: true},
    User{id: 2, name: "Bob", active: false},
    User{id: 3, name: "Charlie", active: true},
}

# Type-safe pipeline using Kukicha stdlib!
activeNames := users
    |> slices.Values()
    |> iter.Filter(func(u User) bool { return u.active })
    |> iter.Map(func(u User) string { return u.name })
    |> slices.Collect()

# activeNames is list of string: ["Alice", "Charlie"]
```

**Transpiles to:**
```go
import "slices"
import "stdlib/iter"

type User struct {
    id int
    name string
    active bool
}

users := []User{
    {id: 1, name: "Alice", active: true},
    {id: 2, name: "Bob", active: false},
    {id: 3, name: "Charlie", active: true},
}

activeNames := slices.Collect(
    iter.Map(
        iter.Filter(
            slices.Values(users),
            func(u User) bool { return u.active },
        ),
        func(u User) string { return u.name },
    ),
)
```

**Type inference makes it work seamlessly!**

---

## Answer to Your Question

### "Can we write stdlib in Kukicha without generics?"

**YES!** Using Approach 5 (Hybrid):

1. **Core stdlib in Kukicha** - `stdlib/iter/*.kuki`
2. **Special transpilation rule** - Makes it generic in Go output
3. **Leverages Go 1.23+** - Uses `iter.Seq[T]` seamlessly
4. **Type-safe through inference** - Users never see generic syntax
5. **Educational path** - Concrete examples show how it works

**What you get:**
- ✅ Stdlib written in Kukicha (readable, maintainable)
- ✅ No generic syntax in Kukicha language
- ✅ Fully type-safe (Go's type system)
- ✅ Works on all types (including user types)
- ✅ Zero-cost abstractions (lazy iterators)
- ✅ Leverages Go stdlib where appropriate

**The one caveat:**
- ⚠️ Stdlib uses special transpilation (only for `stdlib/iter/`)
- ⚠️ Users can't write their own generic functions (must drop to Go)

**Is this acceptable?** For a beginner-focused language, YES! It provides:
- Simple language (no generics to learn)
- Powerful stdlib (written in Kukicha)
- Type safety (through Go)
- Escape hatch (drop to Go for advanced needs)

---

## Implementation Plan

If you choose Approach 5:

**Phase 1: Add Special Transpilation (2 days)**
- Detect `stdlib/iter/*.kuki` files
- Add generic type parameters when transpiling
- Replace `any` with type parameter in those files

**Phase 2: Write Kukicha Stdlib (3 days)**
- `stdlib/iter/iter.kuki` - Filter, Map, Take, Skip, etc.
- `stdlib/slices/slices.kuki` - Wrappers around Go's slices
- `stdlib/maps/maps.kuki` - Wrappers around Go's maps

**Phase 3: Educational Examples (1 day)**
- `stdlib/examples/` - Concrete implementations for learning

**Phase 4: Documentation (1 day)**
- Explain the approach
- Show usage patterns
- Document the special transpilation

**Total: ~7 days**

---

## Conclusion

**You CAN have stdlib written in Kukicha without generic syntax!**

The key is **special transpilation for stdlib** that generates generic Go code from Kukicha source using `any` as a placeholder.

**Recommended approach:** Hybrid (Approach 5)
- Core operations in Kukicha (with special transpilation)
- Direct Go stdlib imports where appropriate
- Educational examples for learning

This achieves your goals:
- ✅ No generics in Kukicha language
- ✅ Stdlib written in Kukicha
- ✅ Type-safe and efficient
- ✅ Leverages Go 1.23+ features

**Next step:** Implement special transpilation rule and write stdlib!
