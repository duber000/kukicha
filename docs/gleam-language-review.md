# Gleam Language Review: Applicable Innovations for Kukicha

**Date:** 2026-01-23
**Status:** Analysis for potential features
**Goal:** Identify Gleam innovations that align with Kukicha's "It's Just Go" philosophy

---

## Executive Summary

**Verdict:** Gleam has interesting innovations, but **pickings are indeed slim** for a Go-aligned language. Most of Gleam's power comes from the BEAM runtime and functional paradigm, which don't translate to Go's imperative, statically-compiled model.

**Viable candidates:**
1. ✅ **Labeled arguments** - Could work well with Go
2. 🤔 **Label shorthand syntax** - Depends on labeled arguments
3. ❌ **Use expressions** - Conflicts with existing `onerr`
4. ❌ **Exhaustive pattern matching** - Go doesn't enforce this
5. ❌ **Result-everywhere** - Go uses (value, error) tuples

---

## What Kukicha Currently Provides

### Current Features (v1.0.0)
| Feature | Kukicha Syntax | Go Equivalent | Notes |
|---------|---------------|---------------|-------|
| English types | `list of int` | `[]int` | ✅ Complete |
| Error handling | `onerr` | if err != nil {...} | ✅ Elegant solution |
| Pipe operator | `\|>` | Nested calls | ✅ Already has this |
| String interpolation | `"{name}"` | `fmt.Sprintf` | ✅ Already has this |
| Negative indexing | `items[-1]` | `items[len(items)-1]` | ✅ With optimization |
| Named receivers | `func f on x X` | `func (x X) f()` | ✅ Explicit naming |
| Walrus operator | `:=` | `:=` | ✅ Keeps Go's best feature |

### Design Philosophy
- **"It's Just Go"** - Transpiles to idiomatic Go
- **Syntax sugar, not semantics** - No runtime overhead
- **Trust the ecosystem** - Use Go stdlib directly
- **Beginner-friendly** - English-like syntax

---

## Gleam Language Innovations

### 1. Type System Features

#### What Gleam Has
- **No null values** - Uses `Option(T)` and `Result(T, E)` types
- **Exhaustive pattern matching** - Compiler enforces all cases handled
- **Type inference** - Full Hindley-Milner across entire program
- **Algebraic data types** - Custom types with pattern matching

#### Go Reality Check
❌ **Can't adopt these** because:
- Go has `nil` - it's fundamental to Go semantics
- Go's switch doesn't require exhaustiveness
- Go uses explicit types in signatures (Kukicha already does this)
- Go's type system is nominal, not structural

**Verdict:** These are BEAM/functional paradigm features that don't translate to Go.

---

### 2. Error Handling: The `use` Expression

#### What Gleam Has

Gleam's `use` expression is their "killer feature" for error handling:

```gleam
pub fn example() -> Result(String, Error) {
  use user <- result.try(fetch_user())
  use profile <- result.try(fetch_profile(user.id))
  use settings <- result.try(fetch_settings(profile.id))
  Ok(format_data(user, profile, settings))
}
```

This desugars to nested pattern matching on Result types.

#### Kukicha Already Has Better (for Go)

Kukicha's `onerr` is **more Go-aligned**:

```kukicha
func example() (string, error)
    user := fetchUser() onerr return "", error
    profile := fetchProfile(user.id) onerr return "", error
    settings := fetchSettings(profile.id) onerr return "", error
    return formatData(user, profile, settings), empty
```

**Why `onerr` is better for Go:**
- ✅ Works with Go's (value, error) tuples directly
- ✅ No Result type wrapper needed
- ✅ Compiles to idiomatic Go error handling
- ✅ Supports multiple patterns: panic, return, default values
- ✅ Interops with all Go libraries without wrapping

**Verdict:** ❌ Don't adopt `use` - Kukicha's `onerr` is already ideal for Go.

---

### 3. Labeled Arguments ⭐ VIABLE CANDIDATE

#### What Gleam Has

```gleam
pub fn replace(
  in string: String,
  each target: String,
  with replacement: String
) -> String {
  // first word is label, second is variable name
}

// Call with any order
replace(in: "🍔🍔🍔", each: "🍔", with: "🍕")
replace(with: "🍕", each: "🍔", in: "🍔🍔🍔")
```

#### How This Could Work in Kukicha

```kukicha
# Function definition - label before parameter name
func replace(in string string, each target string, with replacement string) string
    # Implementation uses 'target' and 'replacement' variables
    # ...

# Calling - can use labels in any order
result := replace(in: "🍔🍔🍔", each: "🍔", with: "🍕")
result := replace(with: "🍕", each: "🍔", in: "🍔🍔🍔")

# Or use positional (labels optional)
result := replace("🍔🍔🍔", "🍔", "🍕")
```

#### Go Compatibility

This **can work** because:
- ✅ Go functions have named parameters
- ✅ Transpiles to normal Go function calls (arguments reordered at compile-time)
- ✅ Zero runtime overhead
- ✅ Improves readability for functions with many parameters
- ✅ Especially valuable for optional-style patterns

#### Example Use Cases

```kukicha
# HTTP requests become self-documenting
response := fetch.Get(
    url: "https://api.github.com/users/duber000",
    timeout: 30,
    headers: map of string to string{"Accept": "application/json"}
)

# Database queries
users := db.Query(
    table: "users",
    where: "age > 21",
    limit: 100,
    order: "created_at DESC"
)

# File operations
files.Copy(
    from: "/tmp/source.txt",
    to: "/data/dest.txt",
    overwrite: true
)
```

#### Implementation Challenges

**Syntax Decision:**
```kukicha
# Option A: Space-separated (Gleam style)
func replace(in string string, each target string, with replacement string) string

# Option B: Colon-separated (more familiar?)
func replace(in: string, each: target string, with: replacement string) string

# Option C: 'as' keyword
func replace(in as string string, each as target string, with as replacement string) string
```

**Recommendation:** Option A (space-separated) - most similar to existing Kukicha syntax.

**Rules:**
- All unlabeled parameters must come before labeled ones
- Labels are always optional at call site
- Compiler reorders arguments to match function definition

**Verdict:** ✅ **STRONG CANDIDATE** - Adds value, zero overhead, Go-compatible

---

### 4. Label Shorthand Syntax ⭐ COMPLEMENTARY FEATURE

#### What Gleam Has

When variable names match label names, you can omit the value:

```gleam
let name = "Lucy"
let age = 35

// Instead of:
create_user(name: name, age: age)

// You can write:
create_user(name:, age:)
```

#### How This Could Work in Kukicha

```kukicha
name := "Lucy"
age := 35

# Instead of
createUser(name: name, age: age)

# Could write
createUser(name:, age:)
```

#### Benefits
- ✅ Reduces repetition
- ✅ Common pattern in modern languages (JavaScript, Rust)
- ✅ Zero runtime cost
- ✅ Only works if labeled arguments are implemented

**Verdict:** 🤔 **CONDITIONAL** - Only implement if labeled arguments are added first.

---

### 5. Exhaustive Pattern Matching

#### What Gleam Has

```gleam
case user.role {
  Admin -> grant_full_access()
  Moderator -> grant_moderate_access()
  User -> grant_basic_access()
  // Compiler error if you forget a case
}
```

#### Go Reality Check

Go's switch **doesn't require exhaustiveness**:
- Default case is optional
- Missing cases just fall through to nothing
- This is intentional - Go trusts developers

**Kukicha's current approach:**
```kukicha
# Switch in Kukicha (transpiles to Go)
switch user.role
    case "admin":
        grantFullAccess()
    case "moderator":
        grantModerateAccess()
    default:
        grantBasicAccess()
```

**Could we add exhaustiveness checking?**
- ❌ Would diverge from Go semantics
- ❌ Go doesn't have ADTs/enums that make this feasible
- ❌ Would require tracking all possible values (not practical with strings/ints)
- ❌ Breaks "It's Just Go" principle

**Verdict:** ❌ Don't adopt - conflicts with Go's design philosophy.

---

### 6. Result Type Everywhere

#### What Gleam Has

```gleam
// All fallible functions return Result(T, E)
pub fn parse_int(s: String) -> Result(Int, Nil) {
  // ...
}

// Even "no error detail" uses Result(T, Nil)
// NOT Option(T)
```

#### Go's Approach

```go
// Go uses (value, error) tuples
func ParseInt(s string) (int, error) {
    // ...
}

// Or just error for no value
func SaveFile(path string) error {
    // ...
}
```

#### Kukicha Already Handles This

```kukicha
# Kukicha uses Go's tuple returns
func parseInt(s string) (int, error)
    # ...

# With onerr for ergonomics
value := parseInt("123") onerr return 0, error
```

**Why Kukicha's approach is better for Go:**
- ✅ Matches Go stdlib conventions
- ✅ Works with all Go libraries out of the box
- ✅ `onerr` provides sugar without wrapping
- ✅ No performance overhead from Result wrapper types

**Verdict:** ❌ Don't adopt - Go's (value, error) is the right pattern.

---

### 7. Other Gleam Features (Not Applicable)

| Feature | Why Not Applicable |
|---------|-------------------|
| **Actor Model** | BEAM-specific, Go uses goroutines/channels |
| **Immutability by default** | Go is imperative with mutable variables |
| **No exceptions** | Go already has this via error returns |
| **First-class functions** | Kukicha already has this |
| **Pipe operator** | Kukicha already has `\|>` |
| **String interpolation** | Kukicha already has `"{var}"` |
| **Type inference** | Kukicha uses signature-first (Go-aligned) |

---

## Recommendations

### ✅ Implement: Labeled Arguments

**Rationale:**
- Go-compatible (zero runtime overhead)
- Solves real readability problems
- Doesn't conflict with existing features
- Especially valuable for:
  - HTTP/network requests
  - Database queries
  - File operations
  - Configuration functions

**Syntax Proposal:**
```kukicha
# Definition
func fetch(url string, timeout int, headers map of string to string) Response
    # ...

# Usage - any of these work
response := fetch(
    url: "https://api.example.com",
    timeout: 30,
    headers: defaultHeaders
)

response := fetch(
    timeout: 30,
    url: "https://api.example.com",
    headers: defaultHeaders
)

response := fetch("https://api.example.com", 30, defaultHeaders)
```

**Implementation:**
1. Parser recognizes `label: value` syntax in function calls
2. Semantic analysis matches labels to parameter names
3. Codegen reorders arguments to match function signature
4. Compiles to normal Go function call

### 🤔 Consider: Label Shorthand (If Labeled Arguments Added)

Only if labeled arguments prove valuable:
```kukicha
url := "https://api.example.com"
timeout := 30

# Instead of
response := fetch(url: url, timeout: timeout, headers: defaultHeaders)

# Could write
response := fetch(url:, timeout:, headers: defaultHeaders)
```

### ❌ Don't Implement

| Feature | Reason |
|---------|--------|
| `use` expressions | `onerr` is better for Go |
| Exhaustive matching | Conflicts with Go semantics |
| Result types | Go uses (value, error) tuples |
| No-nil guarantee | Go has nil, can't change this |
| Immutability | Go is imperative |

---

## Conclusion

### The Bottom Line

**You were right:** Pickings are slim because Gleam's innovations are deeply tied to:
1. The BEAM runtime (actor model, fault tolerance)
2. Functional paradigm (immutability, ADTs, exhaustive matching)
3. Elixir/Erlang compatibility

**One viable feature:** Labeled arguments
- ✅ Go-compatible
- ✅ Zero overhead
- ✅ Solves real readability problems
- ✅ Doesn't conflict with "It's Just Go"

### What Kukicha Already Does Better (for Go)

| Kukicha Feature | Gleam Equivalent | Why Kukicha's Approach is Better |
|----------------|------------------|----------------------------------|
| `onerr` operator | `use` expressions | Works with Go tuples, no wrappers needed |
| Pipe operator `\|>` | Pipe operator `\|>` | Already implemented! |
| String interpolation | String interpolation | Already implemented! |
| English-like types | Type syntax | More beginner-friendly than Gleam |
| Named receivers | Implicit self | Explicit is better |

### Final Assessment

**Gleam is an excellent language** - but for the BEAM ecosystem. Most of its innovations don't translate to Go's imperative, compiled, nil-friendly world.

**Kukicha has already made the right choices** for a Go-aligned language:
- ✅ `onerr` for error handling
- ✅ Pipe operator for data transformation
- ✅ English-like syntax for types
- ✅ Direct Go stdlib usage (no wrappers)

**One addition worth considering:** Labeled arguments would be a natural fit and solve real readability problems, especially in Kukicha's target domain (HTTP, files, databases).

---

## Sources

### Gleam Language Information
- [Gleam: The Rising Star of Functional Programming in 2026](https://pulse-scope.ovidgame.com/2026-01-14-17-54/gleam-the-rising-star-of-functional-programming-in-2026)
- [Gleam programming language](https://gleam.run/)
- [GitHub - gleam-lang/gleam](https://github.com/gleam-lang/gleam)
- [Gleam: The new programming language for building typesafe systems](https://daily.dev/blog/gleam-the-new-programming-language-for-building-typesafe-systems)
- [Exploring Gleam, a type-safe language on the BEAM!](https://christopher.engineering/en/blog/gleam-overview/)
- [Things I like about Gleam's Syntax](https://erikarow.land/notes/gleam-syntax)

### Gleam Features
- [Frequently asked questions | Gleam programming language](https://gleam.run/frequently-asked-questions/)
- [Gleam vs Elixir | What are the differences?](https://stackshare.io/stackups/elixir-vs-gleam)
- [First impressions of Gleam: lots of joys and some rough edges](https://ntietz.com/blog/first-impressions-of-gleam/)

### Specific Feature Documentation
- [My Favorite Gleam Feature](https://erikarow.land/notes/gleam-favorite-feature) (use expressions)
- [Fault tolerant Gleam](https://gleam.run/news/fault-tolerant-gleam/)
- [Result(a, Nil) vs Option(a)](https://github.com/gleam-lang/gleam/discussions/1265)
- [Learn Gleam in Y Minutes](https://learnxinyminutes.com/gleam/)

---

**Next Steps:**
1. Discuss labeled arguments feasibility with team
2. If approved, design syntax and grammar changes
3. Implement in parser → semantic → codegen
4. Add comprehensive tests
5. Update documentation
