# TODO: `concurrent.Map` and `concurrent.MapWithLimit`

## The Readability Case

The current `concurrent.Parallel` API fires zero-argument closures and returns nothing. For the common "transform every element in parallel and collect results" pattern, users must write the goroutine+channel boilerplate manually:

```kukicha
ch := make(channel of Result, len(urls))
for url in urls
    u := url
    go
        ch <- check(u)

results := list of Result{}
for _ from 0 to len(urls)
    results = append(results, receive from ch)
```

`concurrent.Map` collapses this to a single readable line:

```kukicha
results := concurrent.Map(urls, url => check(url))
```

That is a genuine readability win — same parallel execution, ordered results, no boilerplate. Worth implementing.

---

## Blocker: Generic Placeholder System

The current Kukicha generic placeholder system supports three names:

| Placeholder | Go constraint | Use |
|-------------|---------------|-----|
| `any`       | `T any`        | First unconstrained type param |
| `any2`      | `K comparable` | Second type param, map key |
| `ordered`   | `K cmp.Ordered`| Second type param, sort key |

`concurrent.Map` needs **two independent unconstrained type parameters** — input element type T and result type R:

```go
// Desired Go signature
func Map[T any, R any](items []T, fn func(T) R) []R
```

There is no `any3` or `result` placeholder for an unconstrained second type parameter. Writing the signature in `.kuki` source today:

```kukicha
# DOES NOT WORK — any2 has a comparable constraint, wrong for R
func Map(items list of any, fn func(any) any2) list of any2
```

This is the same limitation that affects `slice.Map` — its current signature `func Map(items list of any, transform func(any) any) list of any` forces input and output to be the same type (both collapse to `T`), making `slice.Map(repos, r => r.Name)` technically wrong at the Go generic level (though the Kukicha type checker accepts it).

---

## Solution Options

### Option A — Add a new `result` placeholder (recommended)

Add a fourth generic placeholder name `result` meaning "second unconstrained type parameter":

| Placeholder | Go constraint | Use |
|-------------|---------------|-----|
| `any`       | `T any`        | First type param |
| `any2`      | `K comparable` | Second type param, comparable |
| `ordered`   | `K cmp.Ordered`| Second type param, ordered |
| `result`    | `R any`        | Second unconstrained type param |

The `.kuki` signatures become:

```kukicha
func Map(items list of any, fn func(any) result) list of result
func MapWithLimit(items list of any, limit int, fn func(any) result) list of result
```

Generated Go:

```go
func Map[T any, R any](items []T, fn func(T) R) []R
func MapWithLimit[T any, R any](items []T, limit int, fn func(T) R) []R
```

This also fixes `slice.Map` — its signature should be updated to use `result` as well.

**Compiler changes required:**
- `cmd/genstdlibregistry/main.go` — add `"result"` to the placeholder set alongside `"any"`, `"any2"`, `"ordered"`; emit `R any` as the second type param when encountered
- `internal/semantic/semantic_calls.go` — `GetSliceGenericClass` (or an equivalent helper for concurrent) must recognize functions using `result` and resolve R from the transform's return type
- `internal/semantic/stdlib_types.go` — document the new placeholder in comments

### Option B — Pure Go implementation (simpler, bypasses the placeholder issue)

Write `Map` and `MapWithLimit` directly in Go, not generated from `.kuki` source. The generated `concurrent.go` file would contain the functions by hand.

**Downside:** Breaks the stdlib invariant ("never edit generated `.go` files; all code lives in `.kuki` source"). Not recommended unless Option A is deferred.

---

## Proposed API

```kukicha
# Map runs fn on every element of items concurrently.
# Results are returned in the same order as items.
# All goroutines run at once — use MapWithLimit for large lists.
func Map(items list of any, fn func(any) result) list of result
    results := make(list of result, len(items))
    wg := sync.WaitGroup{}
    wg.Add(len(items))
    for i, item in items
        idx := i
        it := item
        go func()
            results[idx] = fn(it)
            wg.Done()
        ()
    wg.Wait()
    return results

# MapWithLimit is like Map but runs at most `limit` goroutines at once.
func MapWithLimit(items list of any, limit int, fn func(any) result) list of result
    results := make(list of result, len(items))
    wg := sync.WaitGroup{}
    semaphore := make(channel of int, limit)
    wg.Add(len(items))
    for i, item in items
        idx := i
        it := item
        send 1 to semaphore
        go func()
            results[idx] = fn(it)
            receive from semaphore
            wg.Done()
        ()
    wg.Wait()
    return results
```

Usage:

```kukicha
import "stdlib/concurrent"

# Basic parallel map (inferred param type — url inferred as string)
results := concurrent.Map(urls, url => check(url))

# With concurrency cap (useful for rate-limited APIs)
results := concurrent.MapWithLimit(repos, 4, r => fetchDetails(r))
```

---

## Lambda Inference

Once the `result` placeholder is recognized by `genstdlibregistry`, `ParamFuncParams` for `concurrent.Map` will be:

```
ParamFuncParams: map[int][]goStdlibType{1: {{Kind: TypeKindNamed, Name: "any"}}}
```

This matches the existing pattern for `slice.Filter` and `sort.ByKey` — the lambda param is inferred as T (the element type of the piped list), and the Case B inference path in `resolveExpectedLambdaParams` handles it automatically. No additional inference changes are needed.

---

## What Also Needs Updating

If Option A is implemented, `slice.Map` should be updated at the same time:

```kukicha
# Before (input and output forced to same type T)
func Map(items list of any, transform func(any) any) list of any

# After (independent T and R)
func Map(items list of any, transform func(any) result) list of result
```

This is a correctness fix, not a breaking change — callers are unaffected at the Kukicha level.

---

## Scope Estimate

| Task | Files |
|------|-------|
| Add `result` placeholder to genstdlibregistry | `cmd/genstdlibregistry/main.go` |
| Emit `R any` second type param in generated Go | `cmd/genstdlibregistry/main.go` |
| Add `Map` and `MapWithLimit` to `concurrent.kuki` | `stdlib/concurrent/concurrent.kuki` |
| Run `make generate` | regenerates `concurrent.go`, `stdlib_registry_gen.go` |
| Fix `slice.Map` signature | `stdlib/slice/slice.kuki` |
| Add tests | `stdlib/concurrent/concurrent_test.kuki` |
| Update `docs/SKILL.md` concurrent section | `docs/SKILL.md` |

The compiler changes are isolated to `genstdlibregistry` — no parser, semantic, or codegen changes are required beyond what Phase 4 already delivered.
