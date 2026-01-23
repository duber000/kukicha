# Kukicha Pipe-First Simplification Review

**Date:** 2026-01-23
**Focus:** Variables, Pipes, Functions — Basics Only
**Goals:** Type safety, Go alignment, ease of use

---

## Executive Summary

After reviewing Kukicha's syntax and stdlib roadmap, here's a proposal to **simplify to the essentials** while maintaining type safety, error handling, and Go compatibility.

### Core Philosophy

**Three Pillars:**
1. **Variables** - Clear binding (`:=`) and assignment (`=`)
2. **Pipes** - Data flow with `|>` operator
3. **Functions** - Explicit types, simple signatures

**Everything else is sugar** - Keep only what directly supports these three pillars.

---

## Current State Analysis

### ✅ What's Working Well

**1. Pipe Operator (`|>`)**
- Clean left-to-right data flow
- Works seamlessly with Go stdlib
- Makes error handling chains readable
- Perfect for scripting workflows

```kukicha
# Excellent pipe usage
users := http.Get(url)
    |> .Body
    |> io.ReadAll()
    |> parse.Json() as list of User
    |> slice.Filter(u -> u.Active)
    onerr empty list of User
```

**2. Error Handling (`onerr`)**
- Separate keyword from boolean `or` (clearer intent)
- Auto-unwraps (value, error) tuples
- Works naturally in pipes
- Provides fallback values or panic/return

```kukicha
# Clean error handling
config := file.read("config.json")
    |> json.parse()
    |> validate()
    onerr return error "invalid config"
```

**3. Type Safety via Go**
- Explicit types on function signatures
- Local inference inside function bodies
- No type annotations on variables (`:=` infers)
- Direct Go stdlib access

```kukicha
# Signature-first inference
func ProcessData(input string) int     # Explicit params & return
    result := parseInput(input)        # Inferred locally
    count := len(result)               # Inferred
    return count
```

**4. Stdlib Design**
- Pipe-first function signatures
- Thin wrappers over Go stdlib
- Focuses on scripting conveniences Go lacks
- "It's Just Go" underneath

---

## Simplification Opportunities

### 🎯 Core Simplifications

#### 1. **Reduce Syntax Variations**

**Current:** Multiple ways to do the same thing
- `equals` vs `==`
- `and`/`or`/`not` vs `&&`/`||`/`!`
- `at` vs `[]`
- `list of Type` vs `[]Type`

**Proposed:** Pick ONE canonical syntax, support Go syntax for compatibility

```kukicha
# Canonical Kukicha (what kuki fmt outputs)
if count equals 5 and active
    items at 0
    todos := list of Todo{}

# Go syntax (accepted, converted by formatter)
if count == 5 && active
    items[0]
    todos := []Todo{}
```

**Benefit:**
- Beginners learn ONE way
- Go devs can use familiar syntax
- `kuki fmt` normalizes everything
- Less cognitive load

---

#### 2. **Simplify Collections**

**Current:** Many collection operations in `slice` package

**Proposed:** Focus on pipe-essential operations only

**Keep (Essential for pipes):**
- `Filter` - Core pipe operation
- `Map` - Core pipe operation
- `First(n)` / `Last(n)` - Common pipe patterns
- `Drop(n)` / `DropLast(n)` - Common pipe patterns

**Remove/Use Go stdlib:**
- `Reverse` → Use Go's `slices.Reverse()`
- `Unique` → Use Go's `slices.Compact()` + `slices.Sort()`
- `Chunk` → Write inline when needed
- `Contains` → Use Go's `slices.Contains()`
- `IndexOf` → Use Go's `slices.Index()`

**Rationale:**
- Avoid duplicating what Go already does well
- Keep stdlib minimal and focused
- Let users learn Go stdlib directly

```kukicha
# Before: Kukicha wrapper
if slice.Contains(items, value)
    process(value)

# After: Direct Go usage (clearer intent)
import "slices"
if slices.Contains(items, value)
    process(value)

# OR: Use membership operator for lists
if value in items
    process(value)
```

---

#### 3. **Simplify String Package**

**Current:** 40+ functions wrapping Go's strings package

**Proposed:** Keep only pipe-optimized helpers

**Keep (Pipe-friendly):**
- `Split(sep)` - Returns list for pipes
- `Join(sep)` - Joins list in pipes
- `TrimSpace()` - Common pipe step
- `ToUpper()` / `ToLower()` - Common transformations

**Remove/Use Go directly:**
- `Trim`, `TrimLeft`, `TrimRight` → `strings.Trim*()`
- `Replace`, `ReplaceAll` → `strings.Replace*()`
- `Contains`, `HasPrefix`, `HasSuffix` → `strings.*`
- `Index`, `LastIndex`, `Count` → `strings.*`

**Rationale:**
- Teach users Go stdlib from day one
- Less maintenance burden
- Users can read Go docs directly

```kukicha
# Pipe-optimized stays
text := rawInput
    |> string.TrimSpace()
    |> string.ToLower()
    |> string.Split("\n")

# But for complex string ops, use Go directly
import "strings"
if strings.Contains(text, "error")
    handleError()
```

---

#### 4. **Remove Advanced Features (For Now)**

**Current:** Many advanced features competing for attention

**Proposed:** Defer until core is solid

**Remove from v1.0:**
- Negative indexing (`items[-1]`) - Nice but adds complexity
- Membership operators (`in`, `not in`) - Can use Go stdlib
- Multiple range syntaxes (`from...to`, `from...through`) - Just use Go's `range`
- `discard` keyword - Just use `_` (already familiar to Go devs)
- Channel syntax sugar (`send`, `receive`) - Use Go's `<-`

**Keep Go-native syntax:**
```kukicha
# Use standard Go patterns
for i := 0; i < 10; i++          # Standard loop
for _, item := range items       # Use _ directly
ch <- value                      # Go channel syntax
msg := <-ch
```

**Rationale:**
- Every feature adds cognitive load
- Go syntax is already simple
- Focus on what makes Kukicha unique (pipes + onerr)
- Can add these back later if needed

---

### 🚀 What Makes Kukicha Special

After simplification, focus energy on these **unique selling points:**

#### 1. **Pipe Operator**
The killer feature that makes data transformations readable:

```kukicha
# This is uniquely Kukicha
result := data
    |> validate()
    |> transform()
    |> save()
    onerr return error "failed"
```

#### 2. **Error Handling (onerr)**
Simple, clear error handling without try/catch:

```kukicha
# Beautiful error handling
config := file.read("config.json")
    onerr file.read("default.json")
    onerr panic "no config found"
```

#### 3. **Type Inference**
Write less boilerplate while keeping type safety:

```kukicha
func ProcessUsers(users list of User) int
    # No type annotations needed inside
    active := filterActive(users)
    count := len(active)
    return count
```

#### 4. **Scripting-Optimized Stdlib**
High-level helpers that Go lacks:

```kukicha
# fetch.Get - simple HTTP
users := fetch.Get(apiUrl)
    |> fetch.Json() as list of User
    onerr empty list of User

# parse.Json - easy parsing
config := file.read("config.json")
    |> parse.Json() as Config
    onerr defaultConfig()
```

---

## Recommended Minimal Syntax

### Core Language Elements

```kukicha
# 1. VARIABLES
result := calculate()      # Create (walrus)
result = newValue          # Update

# 2. FUNCTIONS
func Add(a int, b int) int
    return a + b

# 3. TYPES
type User
    id int64
    name string
    email string

# 4. PIPES
result := data
    |> transform()
    |> process()

# 5. ERROR HANDLING
data := fetch() onerr return error "failed"

# 6. CONTROL FLOW (Use Go syntax)
if condition
    doSomething()

for i := 0; i < 10; i++
    process(i)

for _, item := range items
    handle(item)
```

### What to Remove

```kukicha
# ❌ REMOVE THESE (use Go syntax instead)
equals           → Use ==
not equals       → Use !=
and, or, not     → Use &&, ||, !
discard          → Use _
at               → Use []
from...to        → Use range or i := 0; i < n; i++
send/receive     → Use <-
items[-1]        → Use items[len(items)-1] or stdlib helper
in/not in        → Use slices.Contains() or custom func
```

**Benefit:**
- Smaller language surface area
- Go devs feel at home immediately
- Less to learn, less to document
- Can always add sugar later

---

## Simplified Stdlib Roadmap

### Phase 1: Pipe-Essential (Now)

**iter** - Lazy evaluation helpers
- `Filter(predicate)` - Essential
- `Map(transform)` - Essential
- `Reduce(initial, reducer)` - Essential
- `Take(n)`, `Skip(n)` - Common patterns

**slice** - Eager helpers (minimal)
- `Filter(predicate)` - For when you need a list back
- `Map(transform)` - For when you need a list back
- `First(n)`, `Last(n)` - Convenience
- Everything else: use Go's `slices` package

**string** - Text pipe helpers (minimal)
- `Split(sep)` - Common in pipes
- `Join(sep)` - Common in pipes
- `TrimSpace()`, `ToUpper()`, `ToLower()` - Common transforms
- Everything else: use Go's `strings` package

### Phase 2: Scripting Power (Next)

**fetch** - HTTP made easy
```kukicha
users := fetch.Get(url)
    |> fetch.Json() as list of User
    onerr empty list of User
```

**parse** - Parsing made easy
```kukicha
config := file.read("config.json")
    |> parse.Json() as Config
    onerr defaultConfig()
```

**files** - File ops for pipes
```kukicha
output := file.read("input.txt")
    |> string.Split("\n")
    |> slice.Filter(notEmpty)
    |> string.Join("\n")
    |> file.write("output.txt")
    onerr panic "processing failed"
```

### Phase 3: Advanced (Later)

Everything else can wait:
- `cli` - Argument parsing
- `shell` - Command execution
- `template` - Text templating
- `retry` - Retry logic
- `result` - Optional/Result types

---

## Type Safety & Error Handling Strategy

### Keep It Simple

**1. Explicit Function Signatures**
```kukicha
# Always explicit on the signature
func ProcessData(input string, limit int) (list of Result, error)
    # Inferred inside
    results := transform(input)
    filtered := applyLimit(results, limit)
    return filtered, empty
```

**2. Error Tuples + onerr**
```kukicha
# Functions return (value, error) tuples
func LoadConfig(path string) (Config, error)
    data := file.read(path) onerr return Config{}, error
    config := parse.Json(data) onerr return Config{}, error
    return config, empty

# Callers use onerr
config := LoadConfig("config.json") onerr defaultConfig()
```

**3. Type Casts via Functions**
```kukicha
# Use Go-style casts
id := int64(rawValue)
text := string(bytes)
```

**4. Let Go Do the Heavy Lifting**
```kukicha
# Use Go's type system directly
import "encoding/json"

func SaveUser(user User) error
    data, err := json.Marshal(user)
    if err != empty
        return err
    return file.write("user.json", data)
```

---

## Go Alignment: What to Embrace

### Use Go Directly (Don't Wrap)

```kukicha
# ✅ GOOD: Use Go stdlib with onerr
import "encoding/json"
import "os"

user := User{id: 1, name: "Alice"}
data := json.Marshal(user) onerr return error
os.WriteFile("user.json", data, 0644) onerr return error

# ❌ BAD: Don't create Kukicha wrappers
user := User{id: 1, name: "Alice"}
data := kuki.toJson(user) onerr return error
kuki.writeFile("user.json", data) onerr return error
```

### Leverage Go's Ecosystem

```kukicha
# Use third-party Go packages directly
import "github.com/gorilla/mux"
import "github.com/lib/pq"

router := mux.NewRouter()
db := sql.Open("postgres", connStr) onerr panic "db connection failed"
```

---

## Recommended Changes

### Immediate (v1.0)

1. **Remove syntax variations**
   - Pick ONE canonical syntax (English-like)
   - Accept Go syntax for compatibility
   - Let `kuki fmt` normalize everything

2. **Trim stdlib packages**
   - Keep only pipe-essential functions
   - Remove wrappers that duplicate Go stdlib
   - Focus on scripting conveniences

3. **Simplify type syntax**
   - Keep: `list of Type`, `map of K to V`, `func(T) R`
   - Everything else: use Go syntax directly

4. **Remove advanced features**
   - Negative indexing
   - Membership operators
   - Multiple range syntaxes
   - Channel syntax sugar
   - Keep focus on pipes + onerr

### Phase 2 (Post-v1.0)

5. **Add scripting packages**
   - `fetch` - HTTP client
   - `parse` - JSON/YAML/CSV
   - `files` - File operations

6. **Evaluate feature requests**
   - Add back negative indexing if highly requested
   - Add membership operators if they prove essential
   - Data-driven decisions

---

## Example: Before vs After

### Before (Too Many Features)

```kukicha
# Multiple ways to do everything
if count equals 5 and active and not expired
    first := items at 0
    last := items at -1

    for i from 0 to 10
        process(i)

    if "error" in message
        handle()

    send ch, value
    msg := receive ch
```

### After (Minimal & Clear)

```kukicha
# One clear way
if count == 5 && active && !expired
    first := items[0]
    last := items[len(items)-1]

    for i := 0; i < 10; i++
        process(i)

    if strings.Contains(message, "error")
        handle()

    ch <- value
    msg := <-ch
```

**What we gain:**
- Smaller learning curve
- Go developers feel at home
- Less to maintain
- Can focus on what makes Kukicha unique

**What we keep that's special:**
```kukicha
# Pipes + error handling = Kukicha's secret sauce
result := fetch.Get(url)
    |> fetch.Json() as list of User
    |> slice.Filter(u -> u.Active)
    |> slice.Map(u -> u.Name)
    |> string.Join(", ")
    onerr "no users found"
```

---

## Metrics for Success

### Language Simplicity
- ✅ **Total keywords:** < 30 (currently ~35)
- ✅ **Stdlib packages in v1.0:** 3-5 max (iter, slice, string + fetch, parse)
- ✅ **Syntax variations:** 1 canonical + Go compatibility
- ✅ **Special operators:** 2 (`|>` and `onerr` only)

### Go Alignment
- ✅ **Go stdlib usage:** Direct, no wrappers
- ✅ **Type system:** 100% Go-compatible
- ✅ **Compiled output:** Idiomatic Go
- ✅ **Third-party packages:** Work out of the box

### Ease of Use
- ✅ **Pipes make data flow obvious**
- ✅ **Error handling is clear (onerr)**
- ✅ **Type inference reduces boilerplate**
- ✅ **Minimal syntax to learn**

---

## Summary

### The Three Pillars

**1. Variables**
- `:=` for create
- `=` for update
- Clear and simple

**2. Pipes**
- `|>` for data flow
- Works with any function
- Left-to-right reading

**3. Functions**
- Explicit types on signatures
- Inference inside bodies
- Direct Go compatibility

### Everything Else

- Use **Go syntax** directly (loops, conditionals, channels)
- Use **Go stdlib** directly (no wrappers)
- Add **scripting helpers** only where Go is verbose
- Keep **stdlib minimal** (3-5 packages in v1.0)

### The Result

A language that's:
- **Smaller** - Less to learn, less to maintain
- **Faster** - Less compilation complexity
- **Clearer** - Obvious what's Kukicha vs Go
- **Better** - Focus on unique strengths (pipes + onerr)

---

## Next Steps

1. **Audit syntax** - Remove redundant keywords
2. **Trim stdlib** - Keep only pipe-essential functions
3. **Update docs** - Focus on variables, pipes, functions
4. **Implement Phase 1** - iter, slice, string (minimal)
5. **Add Phase 2** - fetch, parse (scripting power)
6. **Ship v1.0** - Simple, focused, powerful

**The goal:** A beginner can learn "variables, pipes, and functions" in an afternoon and be productive immediately.
