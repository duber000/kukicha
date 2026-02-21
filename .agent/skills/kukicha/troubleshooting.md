# Kukicha Troubleshooting Guide

## Common Errors and Solutions

### "missing type annotation for parameter"

**Cause:** Function parameters must have explicit types.

```kukicha
# Wrong
func Add(a, b)
    return a + b

# Correct
func Add(a int, b int) int
    return a + b
```

### "missing return type"

**Cause:** Functions that return values must declare return types.

```kukicha
# Wrong
func GetName()
    return "Alice"

# Correct
func GetName() string
    return "Alice"
```

### "expected INDENT"

**Cause:** Missing or incorrect indentation after function/control flow declarations.

```kukicha
# Wrong (no indentation)
func Test()
return 42

# Correct (4-space indent)
func Test() int
    return 42
```

### "unexpected token in type context"

**Cause:** Using Go syntax instead of Kukicha type syntax.

```kukicha
# Wrong (Go syntax)
func Process(items []string)

# Correct (Kukicha syntax)
func Process(items list of string)
```

### "undefined: nil"

**Cause:** Kukicha uses `empty` instead of `nil`.

```kukicha
# Wrong
if user == nil
    return nil

# Correct
if user equals empty
    return empty
```

### "invalid operator &&"

**Cause:** Kukicha uses English boolean operators.

```kukicha
# Wrong
if a && b || !c

# Correct
if a and b or not c
```

### "expected 'of' after 'list'"

**Cause:** Collection types require the full syntax.

```kukicha
# Wrong
items list string

# Correct
items list of string
```

### "assignment mismatch: 1 variable but X returns 2 values"

**Cause:** Using `onerr` or pipe `|>` with a stdlib function that returns `(T, error)`, but the compiler doesn't recognize the function's return count. The semantic analyzer uses a generated registry of known external function return counts.

**Context:** This affects pipe chains with `onerr` where the function is from a stdlib package (e.g., `parse.CsvWithHeader`, `json.MarshalPretty`). The compiler needs to know the function returns 2 values so it can split the assignment into `val, err := f()`.

**Solution:** The registry is now auto-generated from stdlib `.kuki` sources — do **not** edit `semantic.go` manually.

- If the function is in a Kukicha stdlib package: make sure its return type is declared in the `.kuki` file, then run `make genstdlibregistry` (or `make generate`, which runs it automatically). Commit the updated `internal/semantic/stdlib_registry_gen.go`.
- If the function is from an external Go package: add it to the Go stdlib block in `analyzeMethodCallExpr` in `internal/semantic/semantic.go` (just below where `generatedStdlibRegistry` is merged).

### "onerr requires error-returning expression"

**Cause:** Using `onerr` on a function that doesn't return an error.

```kukicha
# Wrong (len doesn't return error)
length := len(items) onerr 0

# Correct
length := len(items)
```

### "cannot use reference without 'of'"

**Cause:** Address-of syntax requires `reference of`.

```kukicha
# Wrong
ptr := &user

# Correct
ptr := reference of user
```

### "expected 'when' or 'otherwise' in switch block"

**Cause:** Using `case` or `default` instead of Kukicha keywords.

```kukicha
# Wrong (Go syntax)
switch command
    case "help"
        showHelp()
    default
        print("unknown")

# Correct (Kukicha syntax)
switch command
    when "help"
        showHelp()
    otherwise
        print("unknown")
```

**Note:** `default` is accepted as an alias for `otherwise`, but `case` is not a keyword — use `when`.

### "unexpected token in expression: REFERENCE" (inside switch)

**Cause:** Using type-style `when` branches (`when reference T`) in a regular value switch.

```kukicha
# Wrong — this is parsed as a value switch
switch event
    when reference a2a.TaskStatusUpdateEvent
        print("status")

# Correct — type switch form
switch event as e
    when reference a2a.TaskStatusUpdateEvent
        print("status: {e.Status.State}")
```

### "'when' branch after 'otherwise' will never execute"

**Cause:** Placing a `when` branch after `otherwise`. The `otherwise` branch catches everything, so later `when` branches are unreachable.

```kukicha
# Wrong — "help" branch will never run
switch command
    when "fetch"
        fetchData()
    otherwise
        print("unknown")
    when "help"
        showHelp()

# Correct — otherwise goes last
switch command
    when "fetch"
        fetchData()
    when "help"
        showHelp()
    otherwise
        print("unknown")
```

### "switch condition branch must be bool"

**Cause:** Using a non-boolean expression in a condition switch (bare `switch` without an expression).

```kukicha
# Wrong — 42 is not a boolean
switch
    when 42
        print("bad")

# Correct — use a comparison
switch
    when score >= 42
        print("good")
```

## Indentation Issues

### Mixed Tabs and Spaces

Kukicha requires 4-space indentation. Tabs cause errors.

```bash
# Fix with formatter
kukicha fmt -w myfile.kuki
```

### Inconsistent Block Levels

Each nested block must increase indentation by exactly 4 spaces.

```kukicha
# Wrong (8-space jump)
func Test()
        return 42

# Correct
func Test() int
    return 42

# Nested blocks
func Process()
    if condition
        for item in items
            process(item)
```

## Type Inference Limits

### Where Inference Works
```kukicha
func Example()
    x := 42              # Inferred as int
    name := "Alice"      # Inferred as string
    items := list of string{"a", "b"}  # Explicit type in literal
```

### Where Inference Doesn't Work
```kukicha
# Function parameters - must be explicit
func Process(x int)     # Required

# Function returns - must be explicit
func GetValue() int     # Required
    return 42

# Empty collections need type
items := list of string{}   # Type required
```

## Error Handling Edge Cases

### Multiple Return Values with onerr
```kukicha
# When function returns (T, error)
value := getData() onerr return empty, error "{error}"

# When function returns (T1, T2, error)
# Use tuple unpacking first
a, b, err := getMultiple()
if err != empty
    return empty, empty, err
```

### onerr with return
```kukicha
# Must match function's return signature
func LoadConfig() Config, error
    data := readFile() onerr return empty Config, error "{error}"  # Explicit empty type
    # ...
```

### Block-style onerr
```kukicha
# Multi-statement error handling with indented block
data := fetchData() onerr
    print("Error occurred: {error}")   # {error} references the caught error
    return

# {error} only works inside the onerr block - it maps to the generated error variable
# Outside onerr blocks, "error" is the Go type, not a variable
```

### onerr explain (error wrapping with hint)
```kukicha
# Standalone explain - wraps error and returns it
func FetchUser(id int64) User, error
    data := db.Query("SELECT * FROM users WHERE id = ?", id) onerr explain "database query failed"
    # ...

# Explain with handler - wraps error before running handler
func GetConfig() int
    port := os.Getenv("PORT") onerr 0 explain "PORT must be set"
    return port
```

## String Interpolation Gotchas

### Escaping Braces
```kukicha
# To include literal braces, double them
msg := "Use {{name}} for variables"  # Outputs: Use {name} for variables
```

### Complex Expressions
```kukicha
# Expressions in interpolation must be valid
msg := "Sum: {a + b}"           # OK
msg := "Cond: {if x then y}"    # Wrong - no if expressions

# Use intermediate variable for complex logic
result := if condition then "yes" else "no"  # Wrong - no ternary
result := "yes"
if not condition
    result = "no"
msg := "Result: {result}"
```

## Import Path Issues

### stdlib Imports
```kukicha
# Correct stdlib path
import "stdlib/slice"
import "stdlib/iter"

# Not this
import "slice"           # Wrong
import "./stdlib/slice"  # Wrong
```

### Go Standard Library
```kukicha
# Use exactly as in Go
import "fmt"
import "encoding/json"
import "net/http"
```

## Method Receiver Mistakes

### Forgetting Receiver Name
```kukicha
# Wrong (missing receiver name)
func Display on Todo string

# Correct
func Display on todo Todo string
    return todo.title
```

### Value vs Reference Receiver
```kukicha
# Value receiver (cannot modify)
func GetTitle on todo Todo string
    return todo.title

# Reference receiver (can modify)
func SetTitle on todo reference Todo title string
    todo.title = title
```

## Debugging Tips

### Check Generated Go Code
```bash
# View transpiled output
kukicha build myfile.kuki
cat myfile.go  # or check build output
```

### Verbose Type Checking
```bash
kukicha check myfile.kuki
```

### Common Build Errors

| Error | Likely Cause |
|-------|--------------|
| "unexpected NEWLINE" | Missing expression or extra blank line |
| "expected identifier" | Using reserved word as variable name |
| "type mismatch" | Wrong type in assignment/return |
| "undeclared name" | Variable used before declaration |
| "not enough arguments" | Missing function arguments |

## Function Type Errors

### "undefined function type"

**Cause:** Using function type syntax incorrectly or with wrong parameter/return types.

```kukicha
# Wrong - mixed Go and Kukicha syntax
callback func(int) -> int

# Correct
callback func(int) int
```

### "function expects N arguments, got M"

**Cause:** Passing function literal with wrong number of parameters.

```kukicha
func Filter(items list of int, predicate func(int) bool) list of int
    # ...

# Note: inline closures don't compile — extract to named top-level functions

# Wrong - takes 2 parameters instead of 1
func wrongPred(a int, b int) bool
    return a > b

result := Filter(numbers, wrongPred)

# Correct - takes 1 parameter matching func(int) bool
func aboveFive(n int) bool
    return n > 5

result := Filter(numbers, aboveFive)
```

### "function literal must return type"

**Cause:** Function type requires return type but function literal doesn't specify it.

```kukicha
# Wrong - no return type specified
callback := func(n int)
    return n * 2

# Correct
callback := func(n int) int
    return n * 2
```

### "cannot use func with wrong signature"

**Cause:** Function signature doesn't match the parameter type.

```kukicha
func Process(handler func(string) int)
    # ...

# Wrong - returns string, not int
Process(func(s string) string
    return s
)

# Correct - returns int
Process(func(s string) int
    return len(s)
)
```

## Import and Package Issues

### "undefined: fmt" in generated Go (interpolated errors)

**Cause:** The compiler generates `errors.New(fmt.Sprintf(...))` for `error "... {var} ..."` literals but does NOT auto-import `fmt`. You must add the import yourself.

```kukicha
# Wrong — causes "undefined: fmt" in generated Go
func Load(path string) Config, error
    data := os.ReadFile(path) onerr return empty Config, error "cannot read {path}"

# Correct — add import "fmt"
import "fmt"
import "os"

func Load(path string) Config, error
    data := os.ReadFile(path) onerr return empty Config, error "cannot read {path}"
```

### Package name conflicts (e.g., "ctx redeclared in this block")

**Cause:** Local variable name matches the imported package name.

```kukicha
# Wrong — local 'ctx' variable shadows the package
import "stdlib/ctx"
func Handle(ctx context.Context)
    c := ctx.Background()   # Error: ctx is the parameter, not the package

# Correct — alias the package
import "stdlib/ctx" as ctxpkg
func Handle(ctx context.Context)
    c := ctxpkg.Background() |> ctxpkg.WithTimeout(30)
    defer ctxpkg.Cancel(c)
```

**Common conflict cases and recommended aliases:**

| Import | Alias |
|--------|-------|
| `stdlib/ctx` | `ctxpkg` |
| `stdlib/errors` | `errs` |
| `stdlib/json` | `jsonpkg` |
| `stdlib/container` | `ctr` |
| `stdlib/string` | `strutil` |

### "variadic argument mismatch" or spreading a slice

**Cause:** Forgetting to use `many` when spreading a slice into a variadic function.

```kukicha
# Wrong — passing slice directly to variadic func
func Sum(many numbers int) int
    # ...

nums := list of int{1, 2, 3}
result := Sum(nums)      # Type error — expects int, not list of int

# Correct — spread with "many" keyword
result := Sum(many nums)
```

## Performance Considerations

### Negative Indexing
```kukicha
# Literal negative index - zero overhead
last := items[-1]  # Compiles to items[len(items)-1]

# Dynamic negative index - requires runtime check
idx := getIndex()  # Might be negative
item := items.at(idx)  # Use .at() method for safety
```

### String Interpolation in Loops
```kukicha
# Avoid in hot loops - each creates new string
for i from 0 to 1000000
    msg := "Item {i}"  # fmt.Sprintf overhead

# Better for hot paths
builder := strings.Builder{}
for i from 0 to 1000000
    builder.WriteString("Item ")
    builder.WriteString(strconv.Itoa(i))
```
