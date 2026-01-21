# Go 1.22-1.25 Language Features: Deep Impact on Kukicha Design

**Date:** 2026-01-21
**Status:** Design Analysis

---

## Critical Insight I Missed

My initial review focused on "can these features simplify the generics *implementation*" but missed the bigger question: **"Do these features change how Kukicha should be designed?"**

The answer is **YES** - especially for **iterators** and **range improvements**.

---

## Part 1: Iterator Revolution (Go 1.23) - THIS IS HUGE

### What Changed in Go 1.23

Go added **range-over-function** - functions can now act as iterators:

```go
// Go 1.23 - function returns an iterator
func Fibonacci(max int) iter.Seq[int] {
    return func(yield func(int) bool) {
        a, b := 0, 1
        for a <= max {
            if !yield(a) {
                return
            }
            a, b = b, a+b
        }
    }
}

// Use it naturally
for n := range Fibonacci(100) {
    fmt.Println(n)
}
```

### Why This Changes Everything for Kukicha

**Current Approach (materializes lists):**
```kukicha
# Current - builds entire list in memory
func Filter[T](items list of T, keep func(T) bool) list of T
    result := empty list of T
    for _, item in items
        if keep(item)
            result = append(result, item)
    return result

# Problem: intermediate allocations
numbers := list of int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
result := numbers
    |> Filter(func(n int) bool { return n > 3 })      # allocates list
    |> Map(func(n int) int { return n * 2 })          # allocates list
    |> Filter(func(n int) bool { return n < 20 })     # allocates list
```

**Iterator Approach (lazy, composable):**
```kukicha
# New - returns iterator, no allocation until collected
func Filter[T](items iter.Seq[T], keep func(T) bool) iter.Seq[T]
    return func(yield func(T) bool) bool
        for item in items
            if keep(item)
                if !yield(item)
                    return false
        return true

# No intermediate lists!
numbers := list of int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
result := numbers
    |> iter.All()                                      # zero-cost
    |> Filter(func(n int) bool { return n > 3 })       # zero-cost
    |> Map(func(n int) int { return n * 2 })           # zero-cost
    |> Filter(func(n int) bool { return n < 20 })      # zero-cost
    |> slices.Collect()                                # single allocation
```

### Benefits of Iterator-Based Stdlib

1. **Memory Efficient**: No intermediate slices
2. **Composable**: Chain operations naturally
3. **Early Exit**: Can break/return from iterator chain
4. **Infinite Sequences**: `Range(0, infinity)` works
5. **Go 1.23+ Native**: Uses stdlib `iter.Seq[T]`

### Example: Iterator-Based Kukicha Stdlib

```kukicha
# Iterator functions (lazy)
func Filter[T](seq iter.Seq[T], keep func(T) bool) iter.Seq[T]
func Map[T, U](seq iter.Seq[T], transform func(T) U) iter.Seq[U]
func Take[T](seq iter.Seq[T], n int) iter.Seq[T]
func Skip[T](seq iter.Seq[T], n int) iter.Seq[T]

# Terminal operations (evaluate)
func Collect[T](seq iter.Seq[T]) list of T
func Count[T](seq iter.Seq[T]) int
func First[T](seq iter.Seq[T]) T
func Any[T](seq iter.Seq[T], predicate func(T) bool) bool

# Usage
numbers := list of int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

# Lazy pipeline
evens := numbers
    |> iter.All()
    |> Filter(func(n int) bool { return n % 2 == 0 })
    |> Map(func(n int) int { return n * n })
    |> Take(3)
    |> Collect()  # Only evaluates what's needed

# Result: [4, 16, 36] - only processed 6 elements, not 10
```

### This Simplifies Your Language Design

**Before (materialized lists):**
- Every operation allocates
- Can't represent infinite sequences
- Performance concerns with long chains
- Need to document allocation behavior

**After (iterators):**
- Zero-cost abstractions
- Infinite sequences work naturally
- Performance is excellent
- Matches Go 1.23+ best practices

**Recommendation:** Design your stdlib around `iter.Seq[T]` from day one.

---

## Part 2: Range Over Integer (Go 1.22) - Adopt Immediately

### What It Is

```go
// Go 1.22+
for i := range 10 {
    fmt.Println(i)  // 0, 1, 2, ..., 9
}
```

### Kukicha Should Support This

```kukicha
# Traditional (verbose)
for i := 0; i < 10; i = i + 1
    print(i)

# New (clean)
for i := range 10
    print(i)

# Even cleaner with underscore
for _ := range 10
    doSomething()
```

### Use Cases

**Repeat N times:**
```kukicha
for _ := range 5
    print("Hello")
```

**Generate sequences:**
```kukicha
squares := empty list of int
for i := range 10
    squares = append(squares, i * i)
```

**With iterators:**
```kukicha
func Range(n int) iter.Seq[int]
    return func(yield func(int) bool) bool
        for i := range n
            if !yield(i)
                return false
        return true

# Use it
for i in Range(100)
    print(i)
```

**Recommendation:** Add `for i := range N` syntax. It's simpler than traditional for loops for "do N times" logic.

---

## Part 3: Loop Variable Safety (Go 1.22) - Automatic Win

### The Old Bug (Pre-Go 1.22)

```go
// Old Go - BUG!
for _, user := range users {
    go func() {
        fmt.Println(user.name)  // All print the SAME user!
    }()
}

// Workaround
for _, user := range users {
    user := user  // Make a copy (wtf?)
    go func() {
        fmt.Println(user.name)  // Now works
    }()
}
```

### Go 1.22+ Fix

```go
// Go 1.22+ - WORKS CORRECTLY
for _, user := range users {
    go func() {
        fmt.Println(user.name)  // Each goroutine gets its own copy
    }()
}
```

### Kukicha Automatically Benefits

Since Kukicha transpiles to Go 1.24+, this bug **never exists** in Kukicha:

```kukicha
# Kukicha - always safe!
for _, user in users
    go func()
        print(user.name)  # Works correctly - no workaround needed
```

**Documentation benefit:** "Unlike old Go, Kukicha's loop variables always work correctly in goroutines. No `user := user` workaround needed."

**Recommendation:** Highlight this in docs as a selling point. Kukicha is "safe by default."

---

## Part 4: slices/maps Stdlib Packages - Use, Don't Rebuild

### What Go 1.23+ Provides

**slices package:**
```go
slices.Sort(items)
slices.Reverse(items)
slices.Contains(items, target)
slices.Equal(a, b)
slices.Index(items, target)
slices.Min(items)
slices.Max(items)

// Iterator support
slices.All(items)        // iter.Seq2[int, T]
slices.Values(items)     // iter.Seq[T]
slices.Collect(iter)     // []T
slices.Sorted(iter)      // []T (sorted)
```

**maps package:**
```go
maps.Clone(m)
maps.Equal(m1, m2)
maps.Keys(m)       // iter.Seq[K]
maps.Values(m)     // iter.Seq[V]
maps.Collect(iter) // map[K]V
```

### Kukicha Should Wrap These, Not Reimplement

**Bad approach (reimplementation):**
```kukicha
# DON'T rebuild what Go provides
func Sort[T ordered](items list of T) list of T
    # ... bubble sort implementation ...
    return items
```

**Good approach (thin wrapper):**
```kukicha
# Just expose Go's optimized implementation
func Sort[T ordered](items list of T)
    slices.Sort(items)  # Uses Go's optimized sort

func Sorted[T ordered](items list of T) list of T
    result := slices.Clone(items)
    slices.Sort(result)
    return result
```

### Example Kukicha Stdlib Using Go Packages

```kukicha
# slices.kuki - thin wrappers around Go stdlib

import "slices"
import "iter"

# Sorting
func Sort[T ordered](items list of T)
    slices.Sort(items)

func Sorted[T ordered](items list of T) list of T
    result := slices.Clone(items)
    slices.Sort(result)
    return result

# Searching
func Contains[T comparable](items list of T, target T) bool
    return slices.Contains(items, target)

func Index[T comparable](items list of T, target T) int
    return slices.Index(items, target)

# Iterators
func All[T](items list of T) iter.Seq[T]
    return slices.Values(items)

func Collect[T](seq iter.Seq[T]) list of T
    return slices.Collect(seq)
```

**Benefits:**
- ✅ Use Go's optimized implementations
- ✅ Less code to maintain
- ✅ Automatically get performance improvements from Go updates
- ✅ Battle-tested stdlib functions

**Recommendation:** Design Kukicha stdlib as thin wrappers around Go 1.23+ packages, not reimplementations.

---

## Part 5: Generic Type Aliases (Go 1.24) - User-Facing Feature

### What It Enables

```go
// Go 1.24+
type List[T any] []T
type StringList = List[string]  // Generic alias!

var names StringList = []string{"Alice", "Bob"}
```

### Kukicha Benefit

Users can create readable type aliases:

```kukicha
# Define generic alias
type List[T] = list of T
type StringList = List[string]

# Use it
names := StringList{"Alice", "Bob"}

# More examples
type IntMap = map of int to string
type UserList = list of User
type Result[T] = tuple of T and error
```

**Benefit:** Improves readability in user code. Doesn't simplify your compiler, but makes Kukicha more expressive.

**Recommendation:** Support this syntax - it's a one-line addition to type parsing.

---

## Part 6: Core Types Removed (Go 1.25) - Simpler Mental Model

### What Changed

Go 1.25 simplified the generics spec by removing "core types" concept. Now type constraints are explained purely via type sets.

### Impact on Kukicha

**Before (confusing):**
```kukicha
# What's the "core type" of ordered?
func Min[T ordered](a T, b T) T
    # ...
```

**After (clearer):**
```kukicha
# T must be in the set of ordered types
func Min[T ordered](a T, b T) T
    if a < b
        return a
    return b
```

**Benefit:** Simpler to explain to users. "T must be comparable" is easier than "T's core type must support comparison."

**Recommendation:** Leverage this in documentation - explain constraints as type sets, not core types.

---

## Part 7: Revised Recommendations

Based on **core language features**, here's what Kukicha should do:

### Priority 1: Adopt Iterator-Based Stdlib (Critical)

Design your entire stdlib around `iter.Seq[T]`:

```kukicha
# Core iterator operations
func Filter[T](seq iter.Seq[T], keep func(T) bool) iter.Seq[T]
func Map[T, U](seq iter.Seq[T], fn func(T) U) iter.Seq[U]
func FlatMap[T, U](seq iter.Seq[T], fn func(T) iter.Seq[U]) iter.Seq[U]
func Take[T](seq iter.Seq[T], n int) iter.Seq[T]
func Skip[T](seq iter.Seq[T], n int) iter.Seq[T]
func Zip[T, U](seq1 iter.Seq[T], seq2 iter.Seq[U]) iter.Seq[tuple of T and U]

# Terminal operations
func Collect[T](seq iter.Seq[T]) list of T
func Reduce[T, U](seq iter.Seq[T], initial U, fn func(U, T) U) U
func Count[T](seq iter.Seq[T]) int
func Any[T](seq iter.Seq[T], pred func(T) bool) bool
func All[T](seq iter.Seq[T], pred func(T) bool) bool
```

**Impact:** Makes functional programming in Kukicha zero-cost and composable.

### Priority 2: Add Range-Over-Integer Syntax (Easy)

```kukicha
for i := range 10
    print(i)
```

**Impact:** Cleaner "do N times" loops.

### Priority 3: Wrap Go Stdlib, Don't Reimplement (Important)

Use `slices`, `maps`, `iter` packages as foundation:

```kukicha
import "slices"
import "maps"
import "iter"

# Expose Go's optimized functions
func Sort[T ordered](items list of T)
    slices.Sort(items)

func Keys[K comparable, V](m map of K to V) iter.Seq[K]
    return maps.Keys(m)
```

**Impact:** Less code to maintain, better performance.

### Priority 4: Support Generic Type Aliases (Nice-to-have)

```kukicha
type StringList = list of string
type IntMap = map of int to string
```

**Impact:** Improves user code readability.

### Priority 5: Document Loop Variable Safety (Marketing)

Highlight that Kukicha is "safe by default":

```kukicha
# Works correctly - no workarounds needed!
for _, user in users
    go func()
        print(user.name)
```

**Impact:** Positions Kukicha as "Go, but safer and cleaner."

---

## Part 8: Updated Stdlib Design

### Before (Materialized Lists)

```kukicha
# Old approach - allocates intermediate lists
func Filter[T](items list of T, keep func(T) bool) list of T
func Map[T, U](items list of T, fn func(T) U) list of U
func FlatMap[T, U](items list of T, fn func(T) list of U) list of U

# Pipeline allocates 3 times
result := data
    |> Filter(isEven)
    |> Map(square)
    |> Filter(isSmall)
```

### After (Iterator-Based)

```kukicha
# New approach - lazy evaluation
func Filter[T](seq iter.Seq[T], keep func(T) bool) iter.Seq[T]
func Map[T, U](seq iter.Seq[T], fn func(T) U) iter.Seq[U]
func FlatMap[T, U](seq iter.Seq[T], fn func(T) iter.Seq[U]) iter.Seq[U]

# Pipeline allocates once (at Collect)
result := data
    |> slices.Values()
    |> Filter(isEven)
    |> Map(square)
    |> Filter(isSmall)
    |> slices.Collect()
```

**Performance difference:**
- Old: O(3n) space, 3 allocations
- New: O(n) space, 1 allocation

---

## Part 9: Example - Rewritten Stdlib

```kukicha
# slices.kuki - Kukicha iterator stdlib

import "iter"
import "slices" as go_slices
import "cmp"

# ============================================================================
# Iterator Generators
# ============================================================================

func All[T](items list of T) iter.Seq[T]
    return go_slices.Values(items)

func Range(start int, end int) iter.Seq[int]
    return func(yield func(int) bool) bool
        for i := start; i < end; i = i + 1
            if !yield(i)
                return false
        return true

func Repeat[T](value T, count int) iter.Seq[T]
    return func(yield func(T) bool) bool
        for _ := range count
            if !yield(value)
                return false
        return true

# ============================================================================
# Transformations (Lazy)
# ============================================================================

func Filter[T](seq iter.Seq[T], keep func(T) bool) iter.Seq[T]
    return func(yield func(T) bool) bool
        for item in seq
            if keep(item)
                if !yield(item)
                    return false
        return true

func Map[T, U](seq iter.Seq[T], transform func(T) U) iter.Seq[U]
    return func(yield func(U) bool) bool
        for item in seq
            if !yield(transform(item))
                return false
        return true

func Take[T](seq iter.Seq[T], n int) iter.Seq[T]
    return func(yield func(T) bool) bool
        count := 0
        for item in seq
            if count >= n
                return true
            if !yield(item)
                return false
            count = count + 1
        return true

func Skip[T](seq iter.Seq[T], n int) iter.Seq[T]
    return func(yield func(T) bool) bool
        count := 0
        for item in seq
            if count >= n
                if !yield(item)
                    return false
            count = count + 1
        return true

# ============================================================================
# Terminal Operations (Evaluate)
# ============================================================================

func Collect[T](seq iter.Seq[T]) list of T
    return go_slices.Collect(seq)

func Count[T](seq iter.Seq[T]) int
    count := 0
    for _ in seq
        count = count + 1
    return count

func Reduce[T, U](seq iter.Seq[T], initial U, fn func(U, T) U) U
    acc := initial
    for item in seq
        acc = fn(acc, item)
    return acc

func Any[T](seq iter.Seq[T], predicate func(T) bool) bool
    for item in seq
        if predicate(item)
            return true
    return false

func All[T](seq iter.Seq[T], predicate func(T) bool) bool
    for item in seq
        if !predicate(item)
            return false
    return true

func First[T](seq iter.Seq[T]) option of T
    for item in seq
        return some(item)
    return none

# ============================================================================
# Sorting (Use Go's optimized implementations)
# ============================================================================

func Sort[T cmp.Ordered](items list of T)
    go_slices.Sort(items)

func Sorted[T cmp.Ordered](seq iter.Seq[T]) list of T
    items := go_slices.Collect(seq)
    go_slices.Sort(items)
    return items

func SortBy[T](items list of T, compare func(T, T) int)
    go_slices.SortFunc(items, compare)

# ============================================================================
# Example Usage
# ============================================================================

func example()
    numbers := list of int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

    # Lazy pipeline - only one allocation
    result := numbers
        |> All()
        |> Filter(func(n int) bool { return n % 2 == 0 })
        |> Map(func(n int) int { return n * n })
        |> Take(3)
        |> Collect()

    # result = [4, 16, 36]

    # Infinite sequence (only takes what's needed)
    firstTenSquares := Range(0, 1000000)
        |> Map(func(n int) int { return n * n })
        |> Take(10)
        |> Collect()
```

---

## Conclusion

### You Were Right to Ask

The core language improvements **DO** significantly impact Kukicha's design:

1. **Iterators (Go 1.23)** change the entire stdlib design
2. **Range-over-integer** should be adopted directly
3. **Loop safety** is automatic - marketing benefit
4. **stdlib packages** should be wrapped, not rebuilt

### Updated Priority Order

1. **Iterator-based stdlib** (Priority 1) - Fundamental design change
2. **Range-over-integer syntax** (Priority 2) - Easy win
3. **Wrap slices/maps packages** (Priority 3) - Less code to maintain
4. **Generics syntax** (Priority 4) - Still recommend `[T]` over placeholders
5. **Generic type aliases** (Priority 5) - Nice-to-have

### The Generics Question Remains

The iterator revelation **doesn't change** my generics syntax recommendation:

```kukicha
# Still recommend this
func Filter[T](seq iter.Seq[T], keep func(T) bool) iter.Seq[T]

# Over this
func Filter(seq iter.Seq of element, keep func(element) bool) iter.Seq of element
```

Because:
- `[T]` is still simpler to implement
- Iterator functions still need clear type parameters
- The `iter.Seq[T]` syntax uses brackets anyway

But **iterators are the bigger opportunity** - they fundamentally improve the language's performance and composability.

---

## Next Steps

Should we:
1. ✅ Redesign stdlib around iterators (big win)
2. ✅ Add range-over-integer syntax (easy win)
3. ✅ Simplify generics to `[T]` (cleanup)
4. ✅ Wrap Go packages (maintenance win)

All four together would significantly simplify and improve Kukicha.
