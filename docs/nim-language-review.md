# Nim Language Feature Review for Kukicha

An analysis of Nim programming language features and their applicability to Kukicha (v0.0.15), a beginner-friendly language that transpiles to Go.

## Feature-by-Feature Assessment

### 1. UFCS (Uniform Function Call Syntax) — SKIP

**Nim**: `foo.bar()` is equivalent to `bar(foo)`. Methods and free functions are interchangeable.

**Kukicha already has the pipe operator** (`|>`) which covers 95% of the same practical benefit: `data |> parse()` becomes `parse(data)`. The pipe operator also supports shorthand method access and the placeholder `_` for non-first-arg placement.

Adding full UFCS (`myList.Filter(fn)` resolving to `slice.Filter(myList, fn)`) would require the semantic analyzer to search all imported packages for matching free functions when a method lookup fails — significant complexity for marginal gain.

**Verdict**: Skip. The pipe operator is Kukicha's answer to UFCS, and it is more explicit.

---

### 2. Result Types — SKIP

**Nim**: Migrating toward `Result[T, string]` for public APIs, replacing exceptions.

**Kukicha already has `onerr`**, which directly desugars Go's idiomatic `(T, error)` return pattern. The 10+ `onerr` forms (panic, return, default, discard, continue, break, block, explain, etc.) cover every practical use case. Adding a Result type would create a competing error-handling paradigm.

**Verdict**: Skip. `onerr` is already superior for Go's error model.

---

### 3. Compile-Time Evaluation — CONSIDER (Constrained Form)

**Nim**: Any function can run at compile time with `static` blocks and `const` evaluation.

**Go constraint**: Go has no compile-time function execution. However, Kukicha's transpiler CAN evaluate expressions at transpile time before emitting Go code.

**What would work**: Const string interpolation — evaluating `{expr}` in const contexts at transpile time:

```kukicha
const AppName = "myapp"
const Greeting = "Hello {AppName}"     # Evaluated at transpile time
const ApiVersion = "v{1 + 1}"         # Arithmetic in const context
```

**Implementation complexity**: Medium. Requires a const evaluator pass in semantic analysis.

**Verdict**: Nice-to-have. Low priority relative to core features but a genuine ergonomic win since Go doesn't support const string interpolation.

---

### 4. Templates/Macros — SKIP

**Nim**: Hygienic macros and templates for compile-time code generation.

Macros are notoriously hard to understand and debug. Go deliberately excludes them, and transpiling macro-generated code to Go would produce unreadable output. Kukicha's `# kuki:` directive system already provides targeted compile-time annotations without macro complexity.

**Verdict**: Skip. Directly contradicts Kukicha's beginner-friendly philosophy.

---

### 5. Enhanced Type Inference — RECOMMENDED

**Nim**: Extensive type inference reduces boilerplate for variables, return types, and generic parameters.

**Kukicha's current state**: Variables use `:=` inference (like Go). Function signatures require explicit types (deliberate design choice). Lambda parameters can be untyped in single-param form (`n => n > 0`).

**Recommended improvement — Lambda parameter type inference from context**:

```kukicha
# Today (verbose)
repos |> slice.Filter((r Repo) => r.Stars > 100)

# With context-based inference (proposed)
repos |> slice.Filter(r => r.Stars > 100)
```

The compiler knows `repos` is `list of Repo` and `slice.Filter` expects `func(T) bool`, so `r` must be `Repo`. This also extends to multi-param lambdas: `(a, b) => a + b`.

**Implementation touches**:
- `internal/semantic/semantic_expressions.go` — infer lambda param types from call context
- `internal/parser/parser_expr.go` — extend multi-param untyped lambda support
- `internal/codegen/codegen_decl.go` — emit inferred types in lambda output

**Verdict**: Highest-value feature from this analysis. Reduces the most common source of verbosity in pipe chains.

---

### 6. Effect System — CONSIDER (Simplified Form)

**Nim**: `{.raises: [IOError, ValueError].}` tracks which exceptions a function can raise at compile time.

**Kukicha context**: Go doesn't have exceptions but does have `panic`. The compiler already tracks `error` returns via `exprReturnCounts` and `funcReturnsError`.

**Proposed**: A `# kuki:panics` directive that warns callers:

```kukicha
# kuki:panics "when input is empty"
func MustParse(s string) Config
    if s equals ""
        panic "empty input"
    return parse(s)
```

Callers would get a compile-time warning encouraging use of a non-panicking alternative.

**Verdict**: Low complexity, moderate value. Fits naturally into the existing directive infrastructure.

---

### 7. Identifier Equality — SKIP

**Nim**: `myVariable`, `my_variable`, and `myvariable` are all the same identifier.

Directly conflicts with Go's case-based export visibility (`Exported` vs `unexported`). Beginners benefit from consistency, not from multiple spellings being equivalent.

**Verdict**: Skip.

---

### 8. Multi-Backend — SKIP

Kukicha transpiles to Go specifically. Multi-backend would fundamentally change the language's identity and lose Go ecosystem integration.

**Verdict**: Not applicable.

---

### 9. Async/Await — SKIP

**Nim**: Future-based async with `async`/`await` keywords and event loop.

Go's goroutines + channels are already superior to async/await. Kukicha already exposes them with clean syntax (`go` blocks, channels, `select`, `send`/`receive`). Adding async/await would create a competing concurrency model.

**Verdict**: Skip.

---

### 10. Pragmas (Extended Directives) — WORTH EXTENDING

**Nim**: `{.pragma.}` annotations control compiler behavior: `{.inline.}`, `{.deprecated.}`, `{.raises.}`, etc.

**Kukicha already has** the `# kuki:` directive system with `deprecated` and `security`. Architecturally the same concept.

**Practical additions**:

| Directive | Purpose | Go Output |
|-----------|---------|-----------|
| `# kuki:deprecated "msg"` | Already exists | Warning at call sites |
| `# kuki:security "cat"` | Already exists | Compile-time checks |
| `# kuki:panics "msg"` | Warn about panic risk | Warning at call sites |
| `# kuki:todo "msg"` | Compile-time reminder | Warning during build |

**Verdict**: Incrementally extend the directive system as needs arise. The architecture is already in place.

---

## Ranked Recommendations

### Tier 1: High Value, Feasible

1. **Lambda Parameter Type Inference from Context** — Reduces the most common verbosity in pipe chains. Medium-high complexity but highest impact.

### Tier 2: Moderate Value, Low Complexity

2. **`# kuki:panics` Directive** — Helps beginners distinguish safe functions from ones that can crash. Trivial extension of existing infrastructure.

3. **`# kuki:todo` Directive** — Useful for AI-generated code that flags sections for human review. Very low complexity.

### Tier 3: Nice-to-Have

4. **Compile-Time Const String Interpolation** — Genuine ergonomic win since Go doesn't support this. Medium complexity.

### Skip: Does Not Fit Kukicha

| Feature | Reason |
|---------|--------|
| Full UFCS | Pipe operator already covers this |
| Result types | `onerr` is already superior |
| Templates/Macros | Contradicts beginner-friendliness |
| Identifier equality | Conflicts with Go's export model |
| Multi-backend | Changes language identity |
| Async/await | Go goroutines are already better |

---

## Nim Stdlib Features → Kukicha Stdlib Additions

Beyond language-level features, Nim's standard library has several modules that map to gaps in Kukicha's stdlib. These are pure Go implementations — no language changes needed.

### Tier 1: Fill Existing Gaps (Highest Value)

#### 1. Expand `stdlib/maps` — inspired by Nim's `tables`

Nim's `tables` module provides ordered tables, count tables, merge, iteration, and transformation. Kukicha's `stdlib/maps` currently only has 3 functions (`Keys`, `Values`, `Contains`).

**Proposed additions**:

```kukicha
# Transformation
maps.Filter(m, (k, v) => v > 0)          # Filter entries by predicate
maps.Map(m, (k, v) => v * 2)             # Transform values
maps.MapKeys(m, (k) => string.ToUpper(k)) # Transform keys

# Combination
maps.Merge(m1, m2)                        # Merge two maps (m2 wins conflicts)
maps.MergeWith(m1, m2, (a, b) => a + b)  # Merge with conflict resolver

# Iteration
maps.ForEach(m, (k, v) => print("{k}: {v}"))

# Access
maps.GetOr(m, key, defaultVal)            # Safe get with default
maps.Pop(m, key)                           # Remove and return value

# Conversion
maps.FromPairs(list of Pair)              # List of key-value pairs → map
maps.ToPairs(m)                            # Map → list of key-value pairs
maps.Invert(m)                             # Swap keys ↔ values
```

**Go compatibility**: All trivially implementable with Go's built-in `map` type and generics.

#### 2. Expand `stdlib/random` — inspired by Nim's `random`

Nim's `random` module provides seeded RNG, integer/float ranges, weighted choice, and shuffling. Kukicha's `stdlib/random` only has 2 functions (`String`, `Alphanumeric`).

**Proposed additions**:

```kukicha
random.Int(min, max int) int              # Random int in range [min, max]
random.Float(min, max float64) float64    # Random float in range
random.Bool() bool                         # Random boolean
random.Choice(items list of any) any      # Pick random element
random.Shuffle(items list of any)          # Shuffle in place
random.Sample(items list of any, n int)   # Pick n unique random elements
random.UUID() string                       # Generate UUID v4
random.Seed(n int64)                       # Set seed for reproducibility
```

**Go compatibility**: All use `math/rand/v2` or `crypto/rand`. UUID uses `crypto/rand` + formatting.

#### 3. Expand `stdlib/concurrent` — inspired by Nim's `threadpool`

Nim provides thread pools, spawn, and flow variables. Kukicha's `stdlib/concurrent` only has `Parallel`, `ParallelWithLimit`, and `Go`.

**Proposed additions**:

```kukicha
# Worker pool
pool := concurrent.NewPool(workers: 4)
pool.Submit(task func())
pool.Wait()
pool.Close()

# Fan-out/fan-in
results := concurrent.FanOut(items, workerCount: 4, (item) => process(item))

# Rate limiting
limiter := concurrent.NewLimiter(rate: 10, per: "second")
limiter.Wait()

# Once (run exactly once, thread-safe)
concurrent.Once(initFunc)

# Mutex helpers
mu := concurrent.NewMutex()
concurrent.WithLock(mu, () => doWork())
```

**Go compatibility**: Maps directly to `sync.WaitGroup`, `sync.Pool`, `sync.Once`, `sync.Mutex`, `golang.org/x/time/rate`.

### Tier 2: New Packages (Moderate Value)

#### 4. `stdlib/option` — inspired by Nim's `options`

Nim's `options` module provides `Option[T]` with `some`, `none`, `isSome`, `get`, `map`, `filter`. This is useful when `empty` (nil) is a valid data value and you need to distinguish "absent" from "present but zero."

```kukicha
import "stdlib/option"

opt := option.Some("hello")
opt2 := option.None(string)

if option.IsSome(opt)
    val := option.Get(opt)

# Chaining
result := option.Some(42)
    |> option.Map((n int) => n * 2)
    |> option.Filter((n int) => n > 50)
    |> option.GetOr(0)
```

**Go compatibility**: Implemented as a generic struct `Option[T]` with `Valid bool` and `Value T`. Go 1.22+ generics make this clean.

#### 5. `stdlib/stats` — inspired by Nim's `stats`

Nim provides `mean`, `variance`, `standardDeviation`, and running statistics. Kukicha's `stdlib/math` covers basic math but no statistics.

```kukicha
import "stdlib/stats"

data := list of float64{1.0, 2.0, 3.0, 4.0, 5.0}

stats.Mean(data)               # 3.0
stats.Median(data)             # 3.0
stats.StdDev(data)             # ~1.414
stats.Variance(data)           # 2.0
stats.Min(data)                # 1.0
stats.Max(data)                # 5.0
stats.Sum(data)                # 15.0
stats.Percentile(data, 90)     # 90th percentile
```

**Go compatibility**: Pure arithmetic — no external dependencies needed.

#### 6. `stdlib/deque` — inspired by Nim's `deques`

Nim provides a double-ended queue. Useful for BFS, sliding windows, and job queues.

```kukicha
import "stdlib/deque"

d := deque.New(int)
deque.PushBack(d, 1)
deque.PushFront(d, 0)
val := deque.PopFront(d)     # 0
val2 := deque.PopBack(d)     # 1
deque.Len(d)
deque.IsEmpty(d)
```

**Go compatibility**: Implemented as a ring buffer backed by a slice. Pure Go, no dependencies.

#### 7. `stdlib/heap` — inspired by Nim's `heapqueue`

Priority queue backed by a binary heap. Useful for scheduling, Dijkstra's algorithm, and top-N queries.

```kukicha
import "stdlib/heap"

h := heap.New((a, b int) => a < b)    # Min-heap
heap.Push(h, 5)
heap.Push(h, 1)
heap.Push(h, 3)
val := heap.Pop(h)                     # 1 (smallest)
top := heap.Peek(h)                    # 3
```

**Go compatibility**: Wraps Go's `container/heap` with a generic, beginner-friendly API.

### Tier 3: Nice-to-Have

#### 8. `stdlib/string` additions — inspired by Nim's `strutils` + `editdistance` + `wordwrap`

Nim's string handling includes edit distance, word wrapping, and more parsing utilities. Some additions to the existing package:

```kukicha
string.WordWrap(text, width: 80)           # Wrap at word boundaries
string.EditDistance(a, b string) int        # Levenshtein distance
string.Slugify(text) string                # "Hello World!" → "hello-world"
string.Truncate(text, maxLen: 50) string   # Truncate with "..." suffix
string.Reverse(text) string                # Reverse a string (Unicode-safe)
string.IsNumeric(text) bool                # All digits?
string.RemovePrefix(text, prefix) string   # Remove if present
string.RemoveSuffix(text, suffix) string   # Remove if present
```

**Go compatibility**: Pure string operations, some using `unicode/utf8`.

#### 9. `stdlib/sort` additions — inspired by Nim's `algorithm`

Nim's `algorithm` module includes binary search, reverse, rotation, and deduplication.

```kukicha
sort.IsSorted(items, less func(a, b) bool) bool
sort.BinarySearch(items, target) int, bool
sort.Deduplicate(items) list of any            # Remove consecutive dupes (sorted input)
sort.Stable(items, less func(a, b) bool)       # Stable sort
sort.Reversed(items) list of any               # Return reversed copy
```

**Go compatibility**: Maps to Go's `sort` and `slices` packages.

### Summary Table

| Nim Module | Kukicha Target | Priority | Complexity |
|------------|---------------|----------|------------|
| `tables` | Expand `stdlib/maps` | High | Low |
| `random` | Expand `stdlib/random` | High | Low |
| `threadpool` | Expand `stdlib/concurrent` | High | Medium |
| `options` | New `stdlib/option` | Medium | Low |
| `stats` | New `stdlib/stats` | Medium | Low |
| `deques` | New `stdlib/deque` | Medium | Low |
| `heapqueue` | New `stdlib/heap` | Medium | Low |
| `strutils`/`editdistance` | Expand `stdlib/string` | Low | Low |
| `algorithm` | Expand `stdlib/sort` | Low | Low |
