---
name: kukicha
description: Help write, debug, and understand Kukicha code - a beginner-friendly language that transpiles to Go. Use when working with .kuki files, discussing Kukicha syntax, error handling with onerr, pipe operators, or the Kukicha compiler/transpiler.
---

# Kukicha Language Skill

Kukicha (茎) transpiles to idiomatic Go. Full language reference is in `AGENTS.md`; stdlib API and patterns are in `stdlib/AGENTS.md` — both always available. This skill adds troubleshooting, gotchas, and content not covered in those files.

## Troubleshooting — Common AI Mistakes

### `{error}` vs `{err}` in onerr blocks

Inside any `onerr` handler (block or inline), the caught error variable is always named `error`, never `err`. Using `{err}` is a **compile-time error** — the compiler rejects it with `use {error} not {err} inside onerr`.

```kukicha
# CORRECT
result := fetch.Get(url) onerr
    print("failed: {error}")
    return

# COMPILE-TIME ERROR — the compiler rejects {err} inside onerr
result := fetch.Get(url) onerr
    print("failed: {err}")    # error: use {error} not {err} inside onerr
    return
```

| onerr form | Error variable |
|------------|----------------|
| `x := f() onerr 0` | — |
| `x := f() onerr panic "msg"` | — |
| `x := f() onerr return empty, error "{error}"` | `{error}` in string |
| Block form (indented body) | `{error}` in interpolation |

### `kukicha init` required before stdlib imports

`kukicha build`/`run` auto-extract stdlib if needed, but `kukicha init` is recommended for new projects — it extracts the embedded stdlib to `.kukicha/stdlib/` and adds the `replace` directive to `go.mod`.

```bash
kukicha init    # run once per project before using import "stdlib/..."
```

### `import "fmt"` for interpolated `error ""` literals

The compiler generates `errors.New(fmt.Sprintf(...))` for `error "msg with {var}"` but does NOT auto-import `fmt`. Add it manually when any `error ""` literal contains interpolation.

```kukicha
import "fmt"    # required if any error "" contains {interpolation}

func doThing(name string) error
    return error "failed for {name}"    # needs fmt imported
```

`print(...)` auto-imports `fmt`. Only the `error ""` literal does not.

### `in` only works in `for` loops

`in` is not a membership operator. Use `slices.Contains` for membership checks.

```kukicha
# WRONG
if item in items
    ...

# CORRECT
if slices.Contains(items, item)
    ...

# 'in' only works as a for-loop keyword
for item in items
    process(item)
```

### `fetch.Json` sample parameter — compile-time type hint, not a runtime value

The argument to `fetch.Json` tells the compiler what type to decode into. It is NOT evaluated at runtime.

| Argument | Decodes | Use when |
|----------|---------|----------|
| `fetch.Json(list of Repo)` | JSON array → `[]Repo` | API returns a JSON array |
| `fetch.Json(empty Repo)` | JSON object → `Repo` | API returns a single JSON object |
| `fetch.Json(map of string to string)` | JSON object → `map[string]string` | Dynamic key-value response |

Passing the wrong shape (e.g., `list of Repo` when the API returns an object) produces a runtime decode error with no compile-time warning. Match the shape to the actual API response.

### Struct literals must be inline — no multiline form

```kukicha
# CORRECT
todo := Todo{id: 1, title: "Learn Kukicha", completed: false}

# WRONG — multiline struct literals do not parse
todo := Todo{
    id: 1,
    title: "Learn Kukicha",
}
```

### `any2` in stdlib source is a compiler placeholder — not user syntax

When reading stdlib `.kuki` files you will see `any2` in function signatures. It is a compiler-reserved name for a second generic type parameter with a `comparable` constraint. Do not use it in application code.

```kukicha
# This is stdlib source — NOT application code syntax
func GroupBy(items list of any, keyFunc func(any) any2) map of any2 to list of any
# Compiler generates: func GroupBy[T any, K comparable](items []T, keyFunc func(T) K) map[K][]T

# In application code, just call it normally — no generics syntax needed
grouped := logs |> slice.GroupBy(getLevel)
```

`any2` cannot be removed without compiler changes — it is the only mechanism Kukicha has to express a second type parameter with a `comparable` constraint in `.kuki` source.

---

## Quick Reference — Not in AGENTS.md

### Transpilation Patterns

| Kukicha | Go |
|---------|----|
| `list of int` | `[]int` |
| `map of string to int` | `map[string]int` |
| `reference User` | `*User` |
| `reference of x` | `&x` |
| `dereference ptr` | `*ptr` |
| `empty` | `nil` |
| `and`, `or`, `not` | `&&`, `\|\|`, `!` |
| `equals` | `==` |
| `print(...)` | `fmt.Println(...)` |
| `"Hello {name}"` | `fmt.Sprintf("Hello %s", name)` |
| `items[-1]` | `items[len(items)-1]` |
| `json:"name"` (struct tag) | `` `json:"name"` `` |
| `a \|> f(b)` | `f(a, b)` |
| `a \|> f(b, _)` | `f(b, a)` (placeholder) |
| `a \|> print` | `fmt.Println(a)` (bare identifier) |
| `x := f() onerr "default"` | `x, err := f(); if err != nil { x = "default" }` |
| `x := f() onerr discard` | `x, _ := f()` |
| `x := f() onerr explain "hint"` | wraps error with hint, propagates |
| `(r Repo) => r.Stars > 100` | `func(r Repo) bool { return r.Stars > 100 }` |
| `go` + indented block | `go func() { ... }()` |
| `switch x as v` / `when reference T` | `switch v := x.(type) { case *T: ... }` |

`onerr` can appear on a continuation line after a pipe chain:

```kukicha
result := fetch.Get(url)
    |> fetch.CheckStatus()
    |> fetch.Json(list of Repo)
    onerr return empty list of Repo
```

### Type Switch

```kukicha
switch event as e
    when reference a2a.TaskStatusUpdateEvent
        print(e.Status.State)
    when reference a2a.Task
        print(e.ID)
    when string
        print(e)
    otherwise
        print("Unknown event")
```

### `print` Builtin

`print` auto-imports `fmt` and transpiles to `fmt.Println()`. Accepts multiple arguments of any type.

```kukicha
print("Hello World")
print("Value:", count, "items")    # variadic
print(user.Name, user.Age)
```

---

## Security — Compiler Checks and Safe APIs

The Kukicha compiler rejects common security anti-patterns at **compile time**. The table below lists what triggers each error and the safe replacement.

| Anti-pattern | Compiler error | Safe replacement |
|---|---|---|
| `pg.Query(pool, "... {var}")` | SQL injection risk | Use `$1` parameters: `pg.Query(pool, "... WHERE x = $1", val)` |
| `httphelper.HTML(w, nonLiteral)` | XSS risk | `httphelper.SafeHTML(w, content)` |
| `fetch.Get(url)` inside HTTP handler | SSRF risk | `fetch.SafeGet(url)` |
| `files.Read(path)` inside HTTP handler | Path traversal risk | `sandbox.Read(box, path)` with `sandbox.New(root)` |
| `shell.Run("cmd {var}")` | Command injection risk | `shell.Output("cmd", arg1, arg2)` |
| `httphelper.Redirect(w, r, nonLiteral)` | Open redirect risk | `httphelper.SafeRedirect(w, r, url, "allowed.host")` |
| `template.Execute(tmpl, data)` for HTML | (warning in docs) | `template.HTMLExecute` / `template.HTMLRenderSimple` |

### New security APIs (stdlib/http, stdlib/fetch, stdlib/template)

```kukicha
# HTML output — escape user content
httphelper.SafeHTML(w, userInput)

# SSRF-protected HTTP fetch (use inside server code / HTTP handlers)
resp := fetch.SafeGet(url) onerr return
# With retry + SSRF protection (builder pattern):
import "stdlib/netguard"
resp := fetch.New(url)
    |> fetch.Transport(netguard.HTTPTransport(netguard.NewSSRFGuard()))
    |> fetch.Retry(3, 500)
    |> fetch.Do() onerr return

# Limit request/response body size (prevent OOM)
httphelper.ReadJSONLimit(r, 1 << 20, reference of input) onerr return
resp := fetch.New(url) |> fetch.MaxBodySize(1 << 20) |> fetch.Do() onerr return

# Safe redirect — allowlist-based host validation (relative URLs always pass)
httphelper.SafeRedirect(w, r, returnURL, "example.com", "api.example.com") onerr return
# For intentional redirects to arbitrary URLs (e.g., a link shortener), set header directly:
w.Header().Set("Location", link.url)
w.WriteHeader(301)

# Security headers — middleware (preferred) or per-handler
http.ListenAndServe(":8080", httphelper.SecureHeaders(mux))
httphelper.SetSecureHeaders(w)  # per-handler alternative

# HTML templates with auto-escaping (html/template instead of text/template)
result := template.HTMLRenderSimple(tmplStr, map of string to any{"name": username}) onerr return
httphelper.HTML(w, result)
# Multi-template workflow:
td := template.TemplateData{Name: "page", Text: tmplStr, Data: data}
html, err := template.HTMLExecute(td) onerr return
httphelper.HTML(w, html)
```

### Detecting HTTP handlers

The compiler's SSRF, path-traversal, and related checks activate when the **enclosing function** has an `http.ResponseWriter` parameter. Any function matching that signature is treated as a handler:

```kukicha
# ← SSRF / path-traversal checks active inside this function
function handleSearch(w http.ResponseWriter, r reference http.Request)
    resp := fetch.Get(url)  # compile error: SSRF risk
    resp := fetch.SafeGet(url)  # OK
```

---

## DevOps & SRE Patterns

```kukicha
# 1. Resource validation
"user@domain.com" |> validate.Email() onerr panic "invalid contact"
env.GetInt("REPLICA_COUNT") onerr 3 |> validate.InRange(1, 10) onerr panic

# 2. Resilient retry loop
func deploy()
    cfg := retry.New() |> retry.Attempts(5)
    attempt := 0
    for attempt < cfg.MaxAttempts
        shell.New("kubectl", "apply", "-f", "manifest.yaml") |> shell.Execute() onerr
            retry.Sleep(cfg, attempt)
            attempt = attempt + 1
            continue
        return

# 3. Concurrent health checks
tasks := list of func(){}
for url in endpoints
    u := url
    tasks = append(tasks, func()
        fetch.Get(u) |> fetch.CheckStatus() onerr print "FAILED: {u}"
    )
concurrent.Parallel(tasks...)
```

---

## When Writing Kukicha Code

1. Always use explicit types for function parameters and returns
2. Use `onerr` for error handling — not manual `if err != nil`
3. Prefer pipe operators for data transformation chains
4. Use English keywords: `and`, `or`, `not`, `equals`, `empty`
5. Use 4-space indentation (tabs not allowed)
6. Use `reference of` / `dereference` instead of `&` / `*`
7. **Add `import "fmt"`** to any file using `{interpolation}` inside `error ""` literals
8. Use `import "stdlib/pkg" as alias` for packages that clash with local names — see canonical alias table in `AGENTS.md`
9. Never edit `stdlib/*/*.go` directly — edit `.kuki` source and run `make generate`
10. Never edit `internal/semantic/stdlib_registry_gen.go` directly — run `make genstdlibregistry` (auto-run by `make generate`)
11. Run `kukicha check file.kuki` to validate syntax before committing
