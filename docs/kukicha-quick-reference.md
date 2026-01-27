# Kukicha Quick Reference

A cheat sheet for developers moving from Go to Kukicha.

## Unique Kukicha Syntax

### 1. Keyword Operators
Kukicha replaces many symbolic operators with English words for better readability.

| Operator | Usage | Description |
|----------|-------|-------------|
| `and` | `a and b` | Logical AND (`&&`) |
| `or` | `a or b` | Logical OR (`||`) |
| `not` | `not a` | Logical NOT (`!`) |
| `equals` | `a equals b` | Equality (`==`) |
| `not equals` | `a not equals b` | Inequality (`!=`) |
| `in` | `item in collection` | Membership test |
| `not in` | `item not in collection` | Inverse membership test |
| `discard` | `onerr discard` | Ignore error in `onerr` clause |

### 2. The Discard Keyword vs Underscore
Kukicha distinguishes between the `discard` keyword and the `_` identifier.

- **Use `_`** for discarding values in `for` loops and multi-value assignments (same as Go).
- **Use `discard`** in `onerr` clauses to explicitly ignore an error.
- **Both `_` and `discard`** can be used as placeholders in the pipe operator (`|>`).

```kukicha
# Use _ in loops
for _, item in items
    print(item)

# Use discard in onerr
data := fetch() onerr discard

# Both work in pipes
user |> json.MarshalWrite(w, _)
user |> json.MarshalWrite(w, discard)
```

### 3. The Pipe Operator (`|>`)
Chain functions and methods in a data-flow style.

```kukicha
# Basic pipe: data is passed as the first argument
users |> slice.Filter(u -> u.active) |> slice.Map(u -> u.name)

# Method shorthand: pipe into a method of the object
response |> .JSON()

# Explicit placeholder: use _ to specify argument position
user |> json.MarshalWrite(w, _)
```

### 4. Error Handling (`onerr`)
Inline error handling for functions that return `(T, error)`.

```kukicha
# Panic on error
data := files.Read("config.json") onerr panic "failed to read"

# Return default value
config := parse(data) onerr DefaultConfig

# Block handler
user := fetchUser(id) onerr
    log.Printf("Error fetching user {id}")
    return empty
```

### 5. References and Pointers
Kukicha uses explicit keywords instead of symbols for pointers.

```kukicha
# Type annotation
func Update(u reference User)

# Address of
userPtr := reference of user

# Dereference
userValue := dereference userPtr
```

### 6. String Interpolation
Insert expressions directly into strings using curly braces.

```kukicha
name := "Kukicha"
version := 1.0
print("Welcome to {name} v{version}!")
print("Math: 1 + 1 = {1 + 1}")
```

### 7. Indentation-based Blocks
Kukicha uses 4-space indentation instead of curly braces for all blocks.

```kukicha
func main()
    if active
        for item in items
            print(item)
    else
        print("Inactive")
```

### 8. Collection Types
Construct composite types with a readable syntax.

```kukicha
# Lists
names := list of string{"Alice", "Bob"}
emptyList := empty list of int

# Maps
scores := map of string to int{"Alice": 100}
emptyMap := empty map of string to int

# Channels
ch := make channel of string, 10
```

### 9. Methods
Methods are defined with an explicit receiver name and the `on` keyword.

```kukicha
type User
    name string

func Greet on u User string
    return "Hello, {u.name}!"

# Pointer receiver
func SetName on u reference User, name string
    u.name = name
```

### 10. Control Flow Variations
```kukicha
# Range loops
for i from 0 to 10          # 0 to 9
for i from 0 through 10     # 0 to 10

# Collection loops
for item in items           # Values only
for i, item in items        # Index and value

# Ternary-like expressions
status := "Active" if user.active else "Inactive"
```

---

## Go to Kukicha Translation Table

| Go | Kukicha |
|----|---------|
| `// comment` | `# comment` |
| `{ ... }` | (Indentation - 4 spaces) |
| `&&`, `\|\|`, `!` | `and`, `or`, `not` |
| `==`, `!=` | `equals`, `not equals` |
| `*T` | `reference T` |
| `&v` | `reference of v` |
| `*v` | `dereference v` |
| `nil` | `empty` or `nil` |
| `if err != nil { return err }` | `onerr return error` |
| `fmt.Println(...)` | `print(...)` |
| `fmt.Sprintf("Hello %s", name)` | `"Hello {name}"` |
| `[]T` | `list of T` |
| `map[K]V` | `map of K to V` |
| `chan T` | `channel of T` |
| `func (r T) Name()` | `func Name on r T` |
| `for _, v := range slice` | `for v in slice` |
| `for i, v := range slice` | `for i, v in slice` |
| `for i := 0; i < 10; i++` | `for i from 0 to 10` |
| `ch <- v` | `send ch, v` |
| `v := <-ch` | `v := receive from ch` |
| `_` | `_` or `discard` (see section 2) |
| `v.(T)` | `v as T` |
| `func F(v ...T)` | `func F(many v T)` |
| `v[len(v)-1]` | `v at -1` or `v[-1]` |
| `v[1:len(v)-1]` | `v[1:-1]` |
| `struct { Key string }` | `type T \n    Key string` |
| `append(slice, item)` | `append(slice, item)` |
| `make([]T, len)` | `make list of T, len` |
| `defer f()` | `defer f()` |
| `go f()` | `go f()` |

---

## Botanical Glossary
Kukicha uses a plant-based metaphor for its module system.

| Term | Go Equivalent | Description |
|------|---------------|-------------|
| **Stem** | Module | The root of your project (`go.mod` location). |
| **Petiole** | Package | A directory of related Kukicha/Go files. |
| **Kukicha** | Language | The "stems and veins" that make Go smoother. |
