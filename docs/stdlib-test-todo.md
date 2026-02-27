# stdlib test TODO

Status of every `*_test.kuki` file and what needs doing.
Run `kukicha check stdlib/<pkg>/<pkg>_test.kuki` to verify after each fix.

Reference pattern: `stdlib/math/math_test.kuki`, `stdlib/slice/slice_test.kuki`.
See **stdlib/CLAUDE.md § Testing Stdlib Packages** for the full convention.

---

## Already done ✓

| Package | Status |
|---------|--------|
| `datetime` | table-driven, passes |
| `math`     | table-driven, passes |
| `slice`    | table-driven, passes |
| `string`   | table-driven, passes |

These are the reference implementations. When in doubt, look at one of them.

---

## Broken syntax — full rewrite needed

These files have parse errors and cannot compile at all.
Rewrite each as `func TestXxx(t reference testing.T)` with table-driven cases.

### Common issues found across these files

**`test <name>` block syntax** (ctx, encoding, env)
The files use an unimplemented `test Name` block form (originally written for a future test-DSL feature). Replace every `test FooBar` block with a standard test function:
```kukicha
# WRONG — unimplemented syntax
test BackgroundFunction
    h := ctx.Background()
    test.AssertNotEmpty(h)

# RIGHT
func TestBackground(t reference testing.T)
    h := ctx.Background()
    test.AssertNotEmpty(t, h)
```
Also add a missing `petiole <pkg>_test` declaration at the top if absent.

**Multi-return in `if` condition** (cast, validate)
`if val, err := f(); condition` is not valid Kukicha — split onto two lines:
```kukicha
# WRONG
if val, err := cast.SmartInt(5); err != empty or val not equals 5
    t.Fatalf("failed")

# RIGHT
val, err := cast.SmartInt(5)
if err != empty or val != 5
    t.Fatalf("SmartInt failed: {err}")
```

**`defer func() { ... }()` closures** (input, must, template)
Anonymous deferred closures for panic recovery don't parse. Use a named helper or
restructure the test to avoid recover-based panic testing entirely:
```kukicha
# WRONG — anonymous closure in defer
defer func()
    if recover() equals empty
        t.Error("Expected panic")
    return
}()

# RIGHT — use t.Run isolation or skip panic testing
# (functions that are supposed to panic are hard to unit-test; skip or document)
```

**`for _, item in list`** (maps — `containsValue` helper)
Kukicha for-in doesn't expose an index variable. Drop the `_,`:
```kukicha
# WRONG
for _, item in list

# RIGHT
for item in list
```

**Struct literals inside a list** (llm, mcp)
If the compiler rejects `list of Foo{Foo{field: val}, ...}`, use the element type
explicitly: `list of FooCase{FooCase{...}, FooCase{...}}`.

**Private field access** (ctx accesses `h.cancel`; obs accesses `logger.service`)
Tests must only use public API — remove assertions on unexported fields and test
the observable behaviour instead (e.g. call the function and check its return value).

---

### Per-package notes

| Package | First error | What to do |
|---------|-------------|------------|
| `cast` | `expected indented block` on multi-return `if` | Split multi-return conditions onto two lines; add table-driven cases |
| `ctx` | Missing `petiole`; `test Name` blocks | Add `petiole ctx_test`; replace `test X` blocks with `func TestX`; remove private-field assertions (`h.cancel`) |
| `encoding` | Missing `petiole`; `test Name` blocks | Add `petiole encoding_test`; replace `test X` blocks with `func TestX` |
| `env` | `test Name` blocks | Replace `test X` blocks with `func TestX` |
| `http` | `expected indented block` (multi-return `if`) | Split multi-return conditions; keep the `httptest.NewRecorder()` infrastructure |
| `input` | `defer func()` closure | Remove panic-recovery tests; test `input.Line` by injecting a fake stdin reader via `os.Pipe()` and a goroutine, or skip interactive tests with `t.Skip` |
| `llm` | Struct literal in `list of` | Fix struct-literal-in-list syntax; consider mocking the HTTP call with `httptest.NewServer` |
| `maps` | `for _, item in list` | Change `for _, item in list` → `for item in list` in the `containsValue` helper |
| `mcp` | Struct literal in `list of` | Fix struct-literal-in-list syntax |
| `must` | `defer func()` closure | Remove panic-recovery pattern; use `test.AssertEqual` on the non-panicking paths; document that panic paths are intentionally untested |
| `template` | `defer func()` closure | Remove panic-recovery; test `template.Execute` and `template.HTMLRenderSimple` with concrete inputs and `test.AssertEqual` |
| `validate` | `expected indented block` (multi-return `if`) | Split multi-return conditions onto two lines |

---

## Semantic errors — small targeted fixes

These parse correctly but fail type checking.

| Package | Error | Fix |
|---------|-------|-----|
| `net` | line 24: `assignment mismatch: 2 variables but 3 values` | `SplitHostPort` returns `(host, port, error)` — use `host, port, err := netutil.SplitHostPort(...)` and add `onerr` |
| `retry` | `undefined identifier 'calculateDelay'` | `calculateDelay` is a private implementation detail — remove those calls and test the public `retry.New()`/`retry.Sleep()` API instead |

---

## Go compile errors — logic/API mismatches

These pass `kukicha check` but fail `go build` or `go test`.

| Package | Error | Fix |
|---------|-------|-----|
| `a2a` | Value types compared with `nil`; private fields accessed | Use `empty` only for pointer/interface types; test the public builder API (`a2a.New(...)`) without inspecting private fields |
| `concurrent` | `comparison of function Parallel == nil is always false` | Don't compare function values to `nil` — call the functions and check results instead |
| `obs` | `logger.service undefined` (unexported field) | Test via public API: call `obs.New(...)` and check log output rather than inspecting struct fields |
| `parse` | `"fmt" imported and not used` in generated `.go` | The `.kuki` source imports `fmt` but never uses it — remove the import |
| `iterator` | `function type must have no type parameters` | The generated code produces invalid generic function types; inspect the generated `.go` and adjust the `.kuki` signatures so the generics emit correctly |

---

## Test logic bugs — compile OK but assertions fail

| Package | Failing test | Issue |
|---------|-------------|-------|
| `errors` | `TestJoin` | `errors.Join` preserves errors but the assertion checks the wrong condition — fix the expected value or the assertion logic |

---

## Passing but old pattern — upgrade candidates

These compile and pass but use bare `t.Errorf` without `t.Run` or `stdlib/test`.
Lower priority — fix only when touching a package for another reason.

| Package | Notes |
|---------|-------|
| `cli` | Simple flag tests; table-driven would help with multiple flag types |
| `fetch` | Integration-style tests with `httptest.NewServer`; subtests would improve failure messages |
| `files` | `TestAppend` calls `files.Append` (JSON-encodes) but expects raw string — fix the assertion or switch to `files.AppendString` |
| `json` | Pure function tests; straightforward table-driven conversion |
| `kube` | Mostly smoke tests against a fake server; lower value for table-driven |
| `netguard` | Small focused tests; table-driven for the allow/block cases |
| `pg` | Requires a live DB for most tests; limited scope for table-driven |
| `random` | Statistical distribution checks; not a natural fit for table-driven |
| `sandbox` | File I/O tests; could benefit from a path-case table |
| `shell` | Command execution tests; table-driven for input/output pairs |

---

## Quick-start checklist for fixing a broken file

1. Add `petiole <pkg>_test` at the top if missing
2. Add `import "stdlib/test"` and `import "testing"` if missing
3. Replace every `test FooBar` block with `func TestFooBar(t reference testing.T)`
4. Split multi-return-in-`if` into two statements
5. Remove assertions on private struct fields — test observable output only
6. Add a case type and `for tc in cases` / `t.Run(tc.name, ...)` loop
7. Replace bare `t.Errorf(...)` checks with `test.AssertEqual(t, got, want)`
8. Run `kukicha check stdlib/<pkg>/<pkg>_test.kuki`
9. Run `./kukicha build stdlib/<pkg>/<pkg>_test.kuki` (generates `_test.go`)
10. Run `go test ./stdlib/<pkg>/...`
