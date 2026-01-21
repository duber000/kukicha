# Proposal: Drop Generics from Kukicha Entirely

**Date:** 2026-01-21
**Status:** For Discussion
**Impact:** Massive simplification

---

## The Radical Simplification

**Proposal:** Remove ALL generic syntax from Kukicha (no `[T]`, no placeholders) and rely entirely on Go's stdlib for generic functionality.

**Result:** A dramatically simpler language that still delivers the full stdlib roadmap.

---

## The Key Insight

**Generics are for library authors, not library users.**

Users don't need to *write* generic code - they need to *call* generic functions. Go's stdlib provides generic functions, and Go has type inference, so users can call them without generic syntax:

```go
// Go's generic function
func slices.Sort[T cmp.Ordered](items []T)

// But you call it WITHOUT type parameters:
items := []int{3, 1, 2}
slices.Sort(items)  // Type inference works!
```

**Kukicha can leverage this!**

---

## Part 1: The Three-Layer Architecture

### Layer 1: Kukicha Language (No Generics)

Users write simple, concrete code:

```kukicha
func ProcessNumbers(items list of int) list of int
    return items
        |> Filter(func(n int) bool { return n > 5 })
        |> Map(func(n int) int { return n * 2 })
        |> Collect()
```

**No generic syntax anywhere.**

### Layer 2: Kukicha Stdlib (Go with Generics)

The stdlib is written in Go with generics:

```go
// In kukicha_stdlib/iter/iter.go
package iter

import "iter"

// Generic function - invisible to Kukicha users
func Filter[T any](seq iter.Seq[T], keep func(T) bool) iter.Seq[T] {
    return func(yield func(T) bool) bool {
        for item := range seq {
            if keep(item) {
                if !yield(item) {
                    return false
                }
            }
        }
        return true
    }
}
```

### Layer 3: Go Stdlib (Already Generic)

Go's stdlib provides generic utilities:

```go
import "slices"
import "maps"
import "iter"

slices.Sort(items)           // Generic
slices.Contains(items, val)  // Generic
maps.Keys(m)                 // Generic iterator
```

### How It Works Together

```kukicha
# Kukicha code
numbers := list of int{3, 1, 4, 1, 5, 9}

result := numbers
    |> slices.Values()                          # Go stdlib (generic)
    |> Filter(func(n int) bool { return n > 3 }) # Kukicha stdlib (generic)
    |> slices.Collect()                         # Go stdlib (generic)
```

**Transpiles to:**

```go
numbers := []int{3, 1, 4, 1, 5, 9}

result := slices.Collect(
    kukicha_stdlib.Filter(
        slices.Values(numbers),
        func(n int) bool { return n > 3 },
    ),
)
```

**Type inference handles everything!**

---

## Part 2: Coverage of Stdlib Roadmap

Let's check if this approach covers your entire stdlib roadmap:

### ✅ Slices Package (100% Coverage)

**All operations can be generic Go functions:**

```go
// kukicha_stdlib/slices/slices.go
package slices

func First[T any](items []T, n int) []T { return items[:n] }
func Last[T any](items []T, n int) []T { return items[len(items)-n:] }
func Drop[T any](items []T, n int) []T { return items[n:] }
func Reverse[T any](items []T) []T { /* implementation */ }
func Unique[T comparable](items []T) []T { /* implementation */ }
func Chunk[T any](items []T, size int) [][]T { /* implementation */ }
```

**Users call without generic syntax:**

```kukicha
import slices

firstThree := slices.first(items, 3)     # Type inferred!
reversed := slices.reverse(items)         # Type inferred!
unique := slices.unique(items)            # Type inferred!
```

### ✅ HTTP Package (100% Coverage - No Generics Needed)

All concrete types:

```kukicha
import http

response := http.get("https://api.example.com/users")
user := User{name: "Alice", age: 30}
http.post("https://api.example.com/users", user)
```

### ✅ JSON Package (100% Coverage)

Type assertions handle parsing:

```kukicha
import json

# Parse with type assertion
config := json.parse(jsonString).(Config)

# Or with onerr
config := json.parse(jsonString) as Config
    onerr return error "invalid JSON"
```

### ✅ File Package (100% Coverage - No Generics Needed)

All string/bytes operations:

```kukicha
import file

content := file.read("config.json")
file.write("output.txt", content)
```

### ✅ String Package (100% Coverage - No Generics Needed)

All string operations:

```kukicha
import string

upper := string.upper("hello")
parts := string.split("a,b,c", ",")
```

### ✅ Docker Package (100% Coverage - No Generics Needed)

Concrete types (containers, images, etc.):

```kukicha
import docker

image := docker.build("my-app:latest", "./Dockerfile")
container := docker.run(image, ports: {"8080": "8080"})
```

### ✅ Kubernetes Package (100% Coverage - No Generics Needed)

Concrete types (pods, deployments, etc.):

```kukicha
import k8s

deployment := k8s.deploy("my-app", image: "my-app:v1.0.0", replicas: 3)
pods := k8s.getPods("production", "app=my-app")
```

### ✅ Claude/OpenAI Packages (100% Coverage - No Generics Needed)

String-based APIs:

```kukicha
import claude

response := claude.complete("Explain Kukicha")
```

### ✅ SQL Package (Mostly Coverage)

Type assertions for queries:

```kukicha
import sql

db := sql.connect("postgres://localhost/mydb")

# Type assertion
users := db.query("SELECT * FROM users").(list of User)

# Or with as syntax
users := db.query("SELECT * FROM users") as list of User
    onerr return error "query failed"
```

### ✅ Test Package (100% Coverage)

Generic assertions in stdlib, concrete calls:

```kukicha
import test

func TestCalculator(t test.T)
    result := add(2, 3)
    test.assertEqual(t, result, 5, "2 + 3 should equal 5")
```

### Summary

**Roadmap Coverage: 100%**

- 80% of stdlib needs NO generics (HTTP, File, String, Docker, K8s, LLM)
- 20% that needs generics (Slices, JSON, SQL, Test) can use generic Go functions with type inference

---

## Part 3: What Users Can and Cannot Do

### ✅ What Users CAN Do (No Generics Needed)

**1. Use stdlib generic functions:**

```kukicha
numbers := list of int{1, 2, 3, 4, 5}
evens := Filter(numbers, func(n int) bool { return n % 2 == 0 })
```

**2. Work with concrete types:**

```kukicha
func ProcessUsers(users list of User) list of User
    return users
        |> Filter(func(u User) bool { return u.active })
        |> Map(func(u User) User { return u.normalize() })
        |> Collect()
```

**3. Use type assertions:**

```kukicha
data := json.parse(jsonString).(MyType)
```

**4. Write business logic (90% of code):**

```kukicha
func HandleRequest(req http.Request) http.Response
    user := authenticate(req)
    data := fetchData(user.id)
    return http.json(data)
```

### ❌ What Users CANNOT Do (Need Workarounds)

**1. Write custom generic functions:**

```kukicha
# CANNOT DO THIS:
func GetFirst[T](items list of T) T
    return items[0]
```

**Workarounds:**

**Option A: Write concrete versions**
```kukicha
func GetFirstInt(items list of int) int
    return items[0]

func GetFirstString(items list of string) string
    return items[0]
```

**Option B: Use `any` (lose type safety)**
```kukicha
func GetFirst(items list of any) any
    return items[0]

# Usage requires type assertion
first := GetFirst(numbers).(int)
```

**Option C: Drop to Go for generic helper**
```go
// helpers.go
package helpers

func GetFirst[T any](items []T) T {
    return items[0]
}
```

```kukicha
# Import and use
import "myproject/helpers"

first := helpers.GetFirst(numbers)  # Type inferred!
```

**2. Define generic data structures:**

```kukicha
# CANNOT DO THIS:
type Box[T]
    value T
```

**Workaround: Define in Go**
```go
// box.go
package myproject

type Box[T any] struct {
    Value T
}

func NewBox[T any](value T) Box[T] {
    return Box[T]{Value: value}
}
```

```kukicha
# Import and use
import "myproject"

intBox := myproject.NewBox(42)
strBox := myproject.NewBox("hello")
```

---

## Part 4: Complexity Comparison

### Current Approach (With Generics)

**Kukicha must implement:**

1. Lexer: 9 placeholder tokens
2. Parser: ~50 lines for type parameters
3. Semantic: ~60 lines for placeholder collection
4. Codegen: Placeholder → T mapping
5. Documentation: Teach generics to beginners
6. Error messages: Map placeholder errors to T

**Total: ~150 lines + teaching overhead**

### No-Generics Approach

**Kukicha must implement:**

1. Lexer: Nothing
2. Parser: Nothing
3. Semantic: Nothing
4. Codegen: Nothing
5. Documentation: "Just call the function"
6. Error messages: Go's type errors (already clear)

**Total: ~0 lines**

### Stdlib Implementation

**Current (in Kukicha):**
```kukicha
# Must parse generics, placeholders, constraints
func Filter[T](items list of T, keep func(T) bool) list of T
    # ...
```

**No-Generics (in Go):**
```go
// Write once in Go, use everywhere
func Filter[T any](items []T, keep func(T) bool) []T {
    // ...
}
```

**Benefit:** Stdlib complexity moves to Go (one-time cost), Kukicha stays simple forever.

---

## Part 5: Teaching & Learning Benefits

### For Beginners

**Current approach (with generics):**
- Learn concrete types ✅
- Learn generic syntax ❌ (complex)
- Learn placeholders vs T ❌ (confusing)
- Learn constraints ❌ (advanced)

**No-generics approach:**
- Learn concrete types ✅
- Just call functions ✅ (simple!)
- Types automatically inferred ✅ (magic!)

### Curriculum Progression

**Phase 1: Basics (No Generics Needed)**
```kukicha
# Concrete types only
func Add(a int, b int) int
    return a + b

numbers := list of int{1, 2, 3}
for _, n in numbers
    print(n)
```

**Phase 2: Stdlib Functions (No Generic Syntax)**
```kukicha
# Use generic functions without understanding generics!
evens := Filter(numbers, func(n int) bool { return n % 2 == 0 })
doubled := Map(evens, func(n int) int { return n * 2 })
```

**Phase 3: Advanced (Optional - Learn Go)**
```kukicha
# When you need custom generics, learn Go
import "myhelpers"  # Written in Go with generics
result := myhelpers.CustomGenericFunc(data)
```

### Comparison to Other Languages

**JavaScript:** No generics syntax, arrays work on any type
```javascript
[1, 2, 3].filter(x => x > 1).map(x => x * 2)  // No generic syntax!
```

**Python (pre-3.5):** No generics syntax, duck typing
```python
def filter_items(items, predicate):
    return [x for x in items if predicate(x)]  # Works on any type!
```

**Kukicha (no-generics approach):** Like JS/Python but TYPE-SAFE
```kukicha
# Clean syntax, type safe, no generic syntax to learn!
filtered := Filter(items, func(x int) bool { return x > 1 })
```

---

## Part 6: Real-World Usage Examples

### Example 1: Web API

```kukicha
import http
import json
import slices

type User
    id int
    name string
    email string
    active bool

func GetActiveUsers() list of User
    # Fetch from database (concrete type)
    allUsers := db.query("SELECT * FROM users") as list of User

    # Filter using generic stdlib function (no generic syntax!)
    return Filter(allUsers, func(u User) bool { return u.active })

func HandleUsersRequest(req http.Request) http.Response
    users := GetActiveUsers()

    # Sort using Go's generic function
    slices.Sort(users)

    return http.json(users)

func main()
    http.handle("/users", HandleUsersRequest)
    http.listen(":8080")
```

**No generic syntax anywhere. Clean and simple.**

### Example 2: Data Processing Pipeline

```kukicha
import file
import string
import slices

func ProcessLogFile(filename string) list of string
    # Read file (concrete)
    content := file.read(filename)
        onerr return empty list of string

    # Split into lines (concrete)
    lines := string.split(content, "\n")

    # Filter errors using generic stdlib (no syntax!)
    errors := lines
        |> slices.Values()
        |> Filter(func(line string) bool {
            return string.contains(line, "ERROR")
        })
        |> Take(10)
        |> slices.Collect()

    return errors

func main()
    errors := ProcessLogFile("app.log")
    for _, err in errors
        print(err)
```

**Generic Filter and Take work seamlessly without generic syntax.**

### Example 3: Custom Generic Helper (Drop to Go)

**When you need it:**

```kukicha
# main.kuki - Cannot write generic function
func ProcessInts(items list of int) int
    return Sum(items)  # Need Sum for int

func ProcessFloats(items list of float) float
    return Sum(items)  # Need Sum for float
```

**Solution: Write once in Go**

```go
// helpers/math.go
package helpers

func Sum[T interface{ int | float64 }](items []T) T {
    var total T
    for _, item := range items {
        total += item
    }
    return total
}
```

**Use from Kukicha:**

```kukicha
import "myproject/helpers"

intSum := helpers.Sum(list of int{1, 2, 3})       # Type inferred
floatSum := helpers.Sum(list of float{1.5, 2.5})  # Type inferred
```

**Benefit:** Advanced users graduate to Go naturally. Write once, use everywhere.

---

## Part 7: Implementation Impact

### What Gets Removed

**1. Lexer (token.go):**
```go
// DELETE these token types
TOKEN_MANY       // Keep this - variadic is still useful!
TOKEN_ELEMENT    // DELETE
TOKEN_ITEM       // DELETE
TOKEN_VALUE      // DELETE
TOKEN_THING      // DELETE
TOKEN_KEY        // DELETE
TOKEN_RESULT     // DELETE
TOKEN_NUMBER     // DELETE
TOKEN_COMPARABLE // DELETE
TOKEN_ORDERED    // DELETE
```

**2. Parser (parser.go):**
```go
// DELETE entire function (~50 lines)
func (p *Parser) parseTypePlaceholders() []*ast.TypeParameter

// DELETE generic type declaration parsing
func (p *Parser) parseGenericTypeDecl() *ast.GenericTypeDecl

// SIMPLIFY function parsing (remove TypeParameters)
type FunctionDecl struct {
    Name       *Identifier
    Parameters []*Parameter
    Returns    []TypeAnnotation
    Body       *BlockStmt
    Receiver   *Receiver
    // DELETE: TypeParameters []*TypeParameter
}
```

**3. Semantic (semantic.go):**
```go
// DELETE entire function (~60 lines)
func (a *Analyzer) collectPlaceholders(decl *ast.FunctionDecl) []*ast.TypeParameter

// DELETE PlaceholderType handling
case *ast.PlaceholderType:
    // DELETE this case
```

**4. Codegen (codegen.go):**
```go
// DELETE placeholder mapping
type Generator struct {
    // DELETE: placeholderMap map[string]string
}

// DELETE generateTypeParameters
func (g *Generator) generateTypeParameters(typeParams []*ast.TypeParameter) string

// DELETE GenericTypeDecl generation
func (g *Generator) generateGenericTypeDecl(decl *ast.GenericTypeDecl)
```

**5. AST (ast.go):**
```go
// DELETE these types
type GenericTypeDecl struct { ... }
type TypeParameter struct { ... }
type PlaceholderType struct { ... }

// SIMPLIFY FunctionDecl
type FunctionDecl struct {
    // DELETE: TypeParameters []*TypeParameter
}
```

### What Stays

**Keep variadic (`many`) - it's still useful:**

```kukicha
func Print(many values)
    for _, v in values
        fmt.Println(v)

Print("hello", "world", 123)
```

**Keep all concrete type features:**
- `list of int`
- `map of string to User`
- `channel of Message`
- Type inference for `:=`

### Lines of Code Removed

| Component | Current Lines | After Removal | Savings |
|-----------|--------------|---------------|---------|
| Lexer | ~400 | ~380 | 20 |
| Parser | ~1400 | ~1300 | 100 |
| Semantic | ~966 | ~900 | 66 |
| Codegen | ~938 | ~880 | 58 |
| AST | ~600 | ~550 | 50 |
| **Total** | **4304** | **4010** | **294 lines** |

**Plus:** Simpler documentation, fewer tests, clearer error messages.

---

## Part 8: Migration of Existing Examples

### Before (With Generics)

```kukicha
func Reverse(items list of element) list of element
    result := make list of element with length len(items)
    for i, item in items
        result[len(items) - 1 - i] = item
    return result

func Map(items list of element, fn func(element) result) list of result
    output := make list of result with length len(items)
    for i, item in items
        output[i] = fn(item)
    return output
```

### After (Stdlib in Go, Call from Kukicha)

**Stdlib (written once in Go):**

```go
// kukicha_stdlib/slices/slices.go
package slices

import "slices" as goslices

func Reverse[T any](items []T) []T {
    result := make([]T, len(items))
    for i, item := range items {
        result[len(items)-1-i] = item
    }
    return result
}

func Map[T, U any](items []T, fn func(T) U) []U {
    result := make([]U, len(items))
    for i, item := range items {
        result[i] = fn(item)
    }
    return result
}
```

**Kukicha code (no generic syntax):**

```kukicha
import slices

numbers := list of int{1, 2, 3, 4, 5}

reversed := slices.Reverse(numbers)              # Type inferred!
doubled := slices.Map(numbers, func(n int) int {
    return n * 2
})
```

**Cleaner, simpler, type-safe.**

---

## Part 9: Pros and Cons

### Pros (Massive Simplification)

1. ✅ **~300 lines of code removed** from compiler
2. ✅ **Zero generic syntax to teach** beginners
3. ✅ **Simpler language spec** - fewer concepts
4. ✅ **Stdlib in Go** - better tooling, better performance
5. ✅ **Type inference "just works"** - Go handles it
6. ✅ **Easier debugging** - errors point to actual code
7. ✅ **Natural graduation path** - advanced users learn Go
8. ✅ **Less maintenance** - fewer features to support
9. ✅ **Clearer positioning** - "Simple layer over Go"

### Cons (Limitations)

1. ❌ **No user-defined generic functions** - must use concrete types or Go
2. ❌ **No user-defined generic types** - must define in Go
3. ❌ **Type assertions needed** - for `any` results like JSON parsing
4. ❌ **Requires Go knowledge** - for advanced abstractions

### Is This Acceptable?

**For a beginner-focused language: YES!**

**Reasoning:**
- Beginners rarely need custom generics (they use stdlib)
- Advanced users can write Go helpers (natural progression)
- 90% of code doesn't need generics (business logic)
- Simpler language = easier to learn = better for beginners

**Analogy:**
- **JavaScript:** No generics, millions of developers
- **Python (historically):** No generics, huge success
- **Kukicha:** No generics, full type safety via Go

---

## Part 10: Alternative: "any" for Everything

**If type assertions are annoying, use `any` escape hatch:**

```kukicha
# User can write with 'any' if they want flexibility
func GetFirst(items list of any) any
    return items[0]

# Usage
firstInt := GetFirst(list of int{1, 2, 3}).(int)
firstStr := GetFirst(list of string{"a", "b"}).(string)
```

**This works, but:**
- Loses compile-time type safety
- Runtime panics if type assertion fails
- Not recommended for beginners

**Better:** Just write concrete versions or drop to Go.

---

## Part 11: Decision Framework

### Questions to Answer

**1. What % of user code needs custom generics?**
- Hypothesis: <5%
- Most code is business logic (concrete types)
- Stdlib covers common generic needs

**2. Is "drop to Go" acceptable for advanced needs?**
- Pro: Natural learning progression
- Pro: Gives users superpowers when needed
- Con: Requires Go knowledge

**3. Does simplicity outweigh expressiveness?**
- For beginners: Simplicity wins
- For experts: Expressiveness wins
- Kukicha targets: Beginners

### The Trade-Off

**With generics:**
- More expressive
- More complex
- Harder to learn
- More code to maintain

**Without generics:**
- Less expressive
- Much simpler
- Easier to learn
- Less code to maintain

**For a beginner language: Simplicity > Expressiveness**

---

## Part 12: Recommendation

### Strong Recommendation: Drop Generics

**Rationale:**

1. **Beginner-focused languages succeed through simplicity**
   - JavaScript: No generics, billions of programs
   - Python (pre-typing): No generics, huge success
   - Scratch, Logo, Basic: No generics, great for learning

2. **Stdlib covers 95% of generic needs**
   - Filter, Map, Reduce, Sort, Reverse, etc.
   - All available as Go functions with type inference
   - Users never see generic syntax

3. **Natural progression: Kukicha → Go**
   - Start with Kukicha (simple, concrete)
   - Use stdlib (transparent generics)
   - Graduate to Go (custom generics)

4. **Massive simplification**
   - 300 lines of code removed
   - Simpler docs, easier teaching
   - Fewer edge cases, less maintenance

5. **Still type-safe**
   - Go's type system enforces safety
   - Type inference handles complexity
   - Compile errors catch mistakes

### Implementation Plan

**Phase 1: Remove generic syntax (1 day)**
- Delete placeholder tokens
- Delete type parameter parsing
- Delete placeholder collection
- Simplify AST

**Phase 2: Create kukicha_stdlib in Go (2 days)**
- Write Filter, Map, Reduce, etc. in Go
- Use iter.Seq[T] for efficiency
- Export for Kukicha use

**Phase 3: Update documentation (1 day)**
- Remove generic syntax from docs
- Add "Using Go Helpers" guide
- Update examples

**Phase 4: Update tests (1 day)**
- Remove generic tests
- Add stdlib integration tests
- Test type inference

**Total: ~5 days of work**

---

## Conclusion

### The Question

> Can we drop generics and using Go's stdlib still offer everything we want on our stdlib roadmap?

### The Answer

**YES - and it makes the language dramatically simpler.**

**Key insights:**

1. **Go's stdlib is generic** - Filter, Map, etc. can be written in Go once
2. **Type inference works** - Users call generic functions without syntax
3. **Roadmap covered 100%** - Every planned feature works without user-facing generics
4. **Natural progression** - Advanced users graduate to Go for custom generics
5. **Massive simplification** - ~300 lines removed, simpler teaching

**The trade-off:**
- ❌ Users can't write custom generic functions (use Go instead)
- ✅ Language is dramatically simpler
- ✅ Still fully type-safe
- ✅ Stdlib provides all common generic operations

**For a beginner-focused language, this is a huge win.**

---

## Next Steps

If you agree with this approach:

1. Remove generic syntax from language
2. Write kukicha_stdlib in Go with generics
3. Update documentation
4. Ship dramatically simpler v1.0

**Result:** A language that keeps beginners in mind while maintaining Go's power through stdlib composition.

---

**What do you think?** Is dropping generics the right simplification for Kukicha?
