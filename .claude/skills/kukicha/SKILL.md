---
name: kukicha
description: Help write, debug, and understand Kukicha code - a beginner-friendly language that transpiles to Go. Use when working with .kuki files, discussing Kukicha syntax, error handling with onerr, pipe operators, or the Kukicha compiler/transpiler.
---

# Kukicha Language Skill

Kukicha (茎) transpiles to idiomatic Go. Full language reference is in `AGENTS.md`; stdlib API and patterns are in `stdlib/AGENTS.md` — both always available.

**For compiler errors and diagnostics**, read `.claude/skills/kukicha/troubleshooting.md`.

## Common AI Mistakes (Gotchas Not in AGENTS.md)

### `{error}` vs `{err}` in onerr blocks

Inside any `onerr` handler, the caught error is always named `error`, never `err`. Using `{err}` is a **compile-time error**. To use a custom name, write `onerr as <ident>` — then both `{error}` and `{<ident>}` are valid inside that block.

```kukicha
# CORRECT — canonical name
result := fetch.Get(url) onerr
    print("failed: {error}")
    return

# CORRECT — named alias (onerr as e)
result := fetch.Get(url) onerr as e
    print("failed: {e}")    # {e} and {error} both work here
    return

# WRONG — compiler rejects {err} inside onerr
result := fetch.Get(url) onerr
    print("failed: {err}")    # error: use {error} not {err} inside onerr
    return
```

### `kukicha init` required before stdlib imports

```bash
kukicha init    # run once per project before using import "stdlib/..."
```

### `import "fmt"` for interpolated `error ""` literals

The compiler generates `errors.New(fmt.Sprintf(...))` for `error "msg with {var}"` but does NOT auto-import `fmt`. Add it manually.

```kukicha
import "fmt"    # required if any error "" contains {interpolation}

func doThing(name string) error
    return error "failed for {name}"
```

`print(...)` auto-imports `fmt`. Only `error ""` literals do not.

### `in` is not a membership operator

```kukicha
# WRONG
if item in items
    ...

# CORRECT
if slices.Contains(items, item)
    ...

# 'in' only works in for loops
for item in items
    process(item)
```

### `fetch.Json` — compile-time type hint, not a runtime value

| Argument | Decodes |
|----------|---------|
| `fetch.Json(list of Repo)` | JSON array → `[]Repo` |
| `fetch.Json(empty Repo)` | JSON object → `Repo` |
| `fetch.Json(map of string to string)` | JSON object → `map[string]string` |

Wrong shape = runtime decode error with no compile-time warning.

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

When reading stdlib `.kuki` files you will see `any2` in function signatures. Do not use it in application code — it is a compiler-reserved name for a second generic type parameter.
