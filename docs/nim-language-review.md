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
