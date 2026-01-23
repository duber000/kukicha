# Kukicha Design Philosophy

**Version:** 1.0.0
**Status:** Active

This document establishes the core design principles guiding Kukicha's development.

---

## Core Principle: "It's Just Go"

Kukicha is syntactic sugar for Go, not a parallel universe.

Like CoffeeScript's golden rule ("It's just JavaScript"), Kukicha's compiled output IS Go. You can use any Go library directly, and there is no interpretation at runtime—just 1:1 compilation.

This principle guides every design decision:
- **We smooth syntax, not semantics**
- **We extend Go's capabilities, not replace them**
- **We trust Go's ecosystem, not duplicate it**

---

## What Kukicha Provides

### 1. Syntax Sugar (Compile-Time Transforms)

These features compile to idiomatic Go with zero runtime overhead:

| Kukicha | Go |
|---------|-----|
| `list of int` | `[]int` |
| `map of string to int` | `map[string]int` |
| `reference User` | `*User` |
| `reference of x` | `&x` (address-of) |
| `dereference ptr` | `*ptr` (dereference) |
| `empty` | `nil` |
| `and`, `or`, `not` | `&&`, `\|\|`, `!` |
| `"Hello {name}"` | `fmt.Sprintf("Hello %s", name)` |
| `items[-1]` | `items[len(items)-1]` |
| Indentation blocks | `{ }` braces |

**Goal: Complete Syntax Coverage**

A beginner should be able to write any program using only Kukicha syntax, without needing to learn Go's symbols (`&`, `*`, `nil`, `{}`, etc.).

### 2. Error Handling Sugar (`onerr`)

The `onerr` operator handles Go's (value, error) tuple returns at the call site:

```kukicha
# Kukicha
content := os.ReadFile("config.json") onerr panic "missing file"

# Compiles to Go
content, err := os.ReadFile("config.json")
if err != nil {
    panic("missing file")
}
```

This is the CoffeeScript approach: sugar at the call site, not wrappers around the function.

### 3. Value-Add Standard Library

Kukicha only provides stdlib packages where they add genuine value Go lacks:

| Package | Purpose | Why It Exists |
|---------|---------|---------------|
| `iter` | Functional iteration | Go's `iter.Seq` is low-level; we add Filter, Map, Reduce |
| `slice` | Slice operations | Go lacks First, Last, Drop, Unique, Chunk |
| `string` | String utilities | Convenience wrappers, minimal overhead |

---

## What Kukicha Does NOT Provide

### No Wrappers for Go's Stdlib

We do not wrap Go's standard library packages. Instead, use them directly with Kukicha syntax:

```kukicha
import "encoding/json"

type User struct
    Name  string
    Email string

# Marshal - pure Kukicha syntax
func SaveUser(user User) (list of byte, error)
    return json.Marshal(user)

# Unmarshal - using 'reference of' for address-of
func LoadUser(data list of byte) (User, error)
    user := User{}
    json.Unmarshal(data, reference of user) onerr return User{}, error
    return user, empty
```

**Why not wrappers?**
- Maintenance burden: Go updates would break wrappers
- Incomplete coverage: Always missing something
- Documentation overhead: Two sets of docs to maintain
- No real benefit: `onerr` handles error tuples already

### No Parallel Type System

Kukicha uses Go's types directly. We don't create alternative abstractions:
- Use `time.Duration`, not a Kukicha Duration type
- Use `context.Context`, not a Kukicha Context type
- Use `io.Reader`, not a Kukicha Reader interface

---

## Design Decisions

### Direct Go Imports

Any Go package works in Kukicha:

```kukicha
import "encoding/json"
import "net/http"
import "github.com/gorilla/mux"  # Third-party works too

# Use them directly
router := mux.NewRouter()
```

### Error Handling Philosophy

Instead of hiding Go's error handling, we make it ergonomic:

```kukicha
# Default value on error
port := os.Getenv("PORT") onerr "8080"

# Panic on error
config := loadConfig() onerr panic "startup failed"

# Return error to caller
data := fetchData() onerr return nil, error

# Log and use default
result := parse(input) onerr
    log.Printf("parse failed: {error}")
    continue with defaultValue
```

### Type Inference Boundaries

Kukicha uses signature-first type inference:
- **Function signatures**: Explicit types required
- **Local variables**: Inferred from context

This matches Go's philosophy while reducing verbosity inside functions.

---

## Learning from Compile-to-X Languages

**CoffeeScript's insight:** Don't create a parallel universe. CoffeeScript made JavaScript more pleasant to write while keeping full compatibility with the JS ecosystem.

**TypeScript's insight:** Add value through types and tooling, not by replacing the underlying language.

Kukicha takes the best of both approaches:
- **Complete syntax coverage** - Beginners never need to see `&`, `*`, `nil`, or `{}`
- **Full Go compatibility** - Any Go library works, compiled output is idiomatic Go
- **No parallel ecosystem** - Use Go's stdlib directly, not wrappers
- **Smooth the rough edges** - English-like syntax, `onerr` for error handling

The goal: A beginner can learn programming with Kukicha, then transition to Go when ready, recognizing all the concepts they already know.

---

## Summary

| Principle | Implication |
|-----------|-------------|
| "It's just Go" | Compiled output is idiomatic Go |
| Syntax sugar, not semantics | No runtime abstractions |
| Trust the ecosystem | Use Go stdlib directly |
| Add value selectively | Only create what Go lacks |
| Error handling sugar | `onerr` at call sites, not wrappers |

Kukicha makes Go more approachable without creating a separate world to learn.
