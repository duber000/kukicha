# AGENTS.md

Kukicha is a beginner-friendly programming language that **transpiles to Go**.
When editing `.kuki` files, write **Kukicha syntax, NOT Go**.

## Kukicha vs Go Syntax (Common AI Mistakes)

| Go | Kukicha |
|----|---------|
| `&&`, `\|\|`, `!` | `and`, `or`, `not` |
| `[]string` | `list of string` |
| `map[string]int` | `map of string to int` |
| `*User` | `reference User` |
| `&user` | `reference of user` |
| `*ptr` | `dereference ptr` |
| `nil` | `empty` |
| `{ }` braces | 4-space indentation |
| `==` | `equals` (or `==`) |
| `func (t T) Method()` | `func Method on t T` |
| `func(x T) T { return expr }` | `(x T) => expr` |
| `go func() { ... }()` | `go` + indented block |

## Keyword Aliases (English-Friendly Forms)

Kukicha accepts English-word aliases for two common keywords:

| Short form | English alias | When to use |
|-----------|--------------|-------------|
| `func`    | `function`   | Beginner-facing code and tutorials |
| `var`     | `variable`   | Top-level variable declarations in beginner-facing code |

Both forms compile identically. Use `func`/`var` in idiomatic/production code, and `function`/`variable` when writing beginner tutorials or agent-generated code aimed at non-programmers.

```kukicha
# These are identical to the compiler:
func Add(a int, b int) int
function Add(a int, b int) int

# Top-level variable (file scope):
var AppName string = "myapp"
variable AppName string = "myapp"
```

**For AI agents generating beginner-facing code:** prefer `function` and `variable`.
**For all other code generation:** use `func` and `var`.

## Generic Type Placeholders (stdlib authoring only)

Kukicha uses reserved placeholder names to express generic type parameters in stdlib `.kuki` source files. **Do not use these in application code** — they are only meaningful inside stdlib function signatures.

| Placeholder | Go equivalent | Constraint | Used for |
|-------------|---------------|------------|----------|
| `any` | `T` | `any` (unconstrained) | First type parameter |
| `any2` | `K` | `comparable` | Second type parameter (e.g., map key) |

Example: `slice.GroupBy` uses `any` for element type and `any2` for the map key type:
```kukicha
# stdlib signature (you read this; you do NOT write it in app code)
func GroupBy(items list of any, keyFunc func(any) any2) map of any2 to list of any
```
The compiler generates: `func GroupBy[T any, K comparable](items []T, keyFunc func(T) K) map[K][]T`

Application code just calls `logs |> slice.GroupBy(getLevel)` — no generics syntax needed.

## Kukicha Syntax Quick Reference

### Variables
```kukicha
count := 42              # Type inferred
count = 100              # Reassignment
```

### Functions (explicit types required)
```kukicha
func Add(a int, b int) int
    return a + b

func Divide(a int, b int) int, error
    if b equals 0
        return 0, error "division by zero"
    return a / b, empty

# Default parameter values
func Greet(name string, greeting string = "Hello") string
    return "{greeting}, {name}!"

# Named arguments (at call site)
result := Greet("Alice", greeting: "Hi")
files.Copy(from: source, to: dest)
```

### Methods (receiver after `on`)
```kukicha
func Display on todo Todo string
    return "{todo.id}: {todo.title}"

func SetDone on todo reference Todo       # Pointer receiver
    todo.done = true
```

### Error Handling (`onerr`)
```kukicha
data := fetchData() onerr panic "failed"              # Panic on error
data := fetchData() onerr return empty, error "{error}" # Propagate error
port := getPort() onerr 8080                          # Default value
_ := riskyOp() onerr discard                          # Ignore error

# Explain syntax - wrap error with hint message
data := fetchData() onerr explain "failed to fetch data"  # Standalone: returns wrapped error
data := fetchData() onerr 0 explain "fetch failed"        # With handler: wraps error, then runs handler

# Block-style onerr (multi-statement error handling)
users := csvData |> parse.CsvWithHeader() onerr
    print("Failed to parse CSV: {error}")    # {error} refers to the caught error
    return
```
> **`{error}` in `onerr` — critical:** The caught error is always named `error`, never `err`. Use `{error}` in string interpolation to reference it. Writing `{err}` inside any `onerr` handler is a **compile-time error** — the compiler will reject it with `use {error} not {err} inside onerr`.

| onerr form | Example | Error variable available |
|------------|---------|--------------------------|
| Default value | `x := f() onerr 0` | — |
| Panic | `x := f() onerr panic "msg"` | — |
| Propagate inline | `x := f() onerr return empty, error "{error}"` | `{error}` in string |
| Block (multi-stmt) | `x := f() onerr` + indented body | `{error}` in interpolation |

Use the **block form** when the error handler needs more than one statement; use inline forms for everything else.

> **Note:** `error "msg"` always requires a message string. Use `error "{error}"` to include the original error text when propagating.

### Types
```kukicha
type Todo
    id int64
    title string as "title"         # JSON alias sugar
    tags list of string
    meta map of string to string

# Function type aliases
type Handler func(string)
type Transform func(int) (string, error)
```

```kukicha
# Typed JSON decode (preferred over bytes + unmarshal boilerplate)
items := fetch.Get(url) |> fetch.CheckStatus() |> fetch.Json(list of Todo) onerr panic "{error}"
```

> **`fetch.Json` sample parameter:** The argument is a typed zero value that tells the compiler what to decode into — it is NOT passed at runtime.
> - `fetch.Json(list of Todo)` → decodes a JSON array into `[]Todo`
> - `fetch.Json(empty Todo)` → decodes a JSON object into `Todo`
> - `fetch.Json(map of string to string)` → decodes a JSON object into `map[string]string`
>
> Passing the wrong shape (e.g., `list of Todo` when the API returns an object) produces a runtime decode error with no compile-time warning.

### Collections
```kukicha
items := list of string{"a", "b", "c"}
config := map of string to int{"port": 8080}
last := items[-1]                      # Negative indexing
```

### Control Flow
```kukicha
if count equals 0
    return "empty"
else if count < 10
    return "small"

for item in items
    process(item)

for i from 0 to 10        # 0..9 (exclusive)
for i from 0 through 10   # 0..10 (inclusive)

switch command
    when "fetch", "pull"
        fetchRepos()
    when "help"
        showHelp()
    otherwise
        print "Unknown: {command}"

switch                     # condition switch (bare)
    when stars >= 1000
        print "Popular"
    otherwise
        print "New"
```

### Pipes
```kukicha
result := data |> parse() |> transform()

# Placeholder _ for non-first position
todo |> json.MarshalWrite(w, _)   # becomes: json.MarshalWrite(w, todo)

# Bare identifier as pipe target (no parentheses needed)
data |> print                     # becomes: fmt.Println(data)
```

### Arrow Lambdas
```kukicha
# Expression lambda (auto-return)
repos |> slice.Filter((r Repo) => r.Stars > 100)

# Single untyped param (no parens)
numbers |> slice.Filter(n => n > 0)

# Zero params
button.OnClick(() => print("clicked"))

# Block lambda (multi-statement, explicit return)
repos |> slice.Filter((r Repo) =>
    name := r.Name |> string.ToLower()
    return name |> string.Contains("go")
)
```

### Variadic Arguments (`many`)
```kukicha
# Declare: "many" before param name
func Sum(many numbers int) int
    total := 0
    for n in numbers
        total = total + n
    return total

# Call with individual args
result := Sum(1, 2, 3)

# Spread a slice with "many" at call site
args := list of int{1, 2, 3}
result := Sum(many args)
```

### Type Assertions
```kukicha
# Two-value form (safe)
result, ok := value.(string)
if ok
    print("string: {result}")

# Direct assertion (panics if wrong type)
s := value.(string)
```

### Multi-Value Destructuring
```kukicha
# 2-value (common)
data, err := os.ReadFile(path)

# 3-value (supported)
_, ipNet, err := net.ParseCIDR("192.168.0.0/16")
```

### Concurrency
```kukicha
ch := make channel of string
send "message" to ch
msg := receive from ch
go doWork()

# Go block (multi-statement goroutine)
go
    mu.Lock()
    doWork()
    mu.Unlock()

# Select (channel multiplexing)
select
    when receive from done           # bare receive (no assignment)
        return
    when msg := receive from ch      # assign one var
        print(msg)
    when msg, ok := receive from ch  # assign two vars (ok check)
        if ok
            print(msg)
    when send "ping" to out          # send case
        print("sent")
    otherwise                        # default (non-blocking)
        print("nothing ready")
```

## Security Checks (Compiler-Enforced)

The compiler enforces SQL injection, XSS, SSRF, path traversal, command injection, and open redirect checks at compile time. See **[`stdlib/AGENTS.md`](stdlib/AGENTS.md)** for the full check table and safe alternatives.

## Build & Test Commands

```bash
make build                # Build the kukicha compiler
make test                 # Run all tests
make generate             # Regenerate stdlib_registry_gen.go + all stdlib .go files
make genstdlibregistry    # Regenerate only internal/semantic/stdlib_registry_gen.go
kukicha check file.kuki   # Validate syntax without compiling
kukicha build file.kuki   # Transpile and compile to binary
kukicha run file.kuki     # Transpile, compile, and run
kukicha fmt -w file.kuki  # Format in place
```

## File Map

```
cmd/kukicha/              # CLI entry point
cmd/genstdlibregistry/    # Generator: scans stdlib/*.kuki → stdlib_registry_gen.go
internal/
  lexer/                  # Tokenization (INDENT/DEDENT handling)
  parser/                 # Recursive descent parser → AST
  ast/                    # AST node definitions
  semantic/               # Type checking, validation
    stdlib_registry_gen.go  # GENERATED — run "make genstdlibregistry" to update
  codegen/                # AST → Go code generation
  formatter/              # Code formatting
stdlib/                   # Standard library (.kuki source files)
  slice/                  # Filter, Map, GroupBy, etc.
  json/                   # encoding/json wrapper
  fetch/                  # HTTP client (Auth, Sessions)
  files/                  # File I/O
  shell/                  # Command execution
  ...
examples/                 # Example programs
docs/                     # Documentation
```

## Imports

```kukicha
import "stdlib/slice"                   # standard package
import "stdlib/ctx" as ctxpkg          # alias — use when the package name conflicts with a local variable
import "github.com/jackc/pgx/v5" as pgx  # external package with alias
```

Use `as alias` whenever the package's last path segment clashes with a local variable name. See **[`stdlib/AGENTS.md`](stdlib/AGENTS.md)** for the canonical alias table.

## Critical Rules

1. **Always validate** - Run `kukicha check` before committing `.kuki` changes
2. **4-space indentation only** - Tabs are not allowed in Kukicha
3. **Explicit function signatures** - Parameters and return types must be declared
4. **Test with `make test`** - Run the full test suite

## Adding Features to the Compiler

Typical workflow for new syntax:
1. **Lexer** (`internal/lexer/`) - Add token type if new keyword/operator
2. **Parser** (`internal/parser/`) - Add parsing logic, create AST nodes
3. **AST** (`internal/ast/`) - Define new node types if needed
4. **Codegen** (`internal/codegen/`) - Generate corresponding Go code
5. **Tests** - Add tests in each modified package

## Stdlib Packages

See **[`stdlib/AGENTS.md`](stdlib/AGENTS.md)** for the full package reference, API details, and common usage patterns.

Import with: `import "stdlib/slice"`

## More Documentation

- `.agent/skills/kukicha/` - Comprehensive syntax reference, examples, and troubleshooting (for all AI tools)
- `.claude/skills/kukicha/` - Same content, Claude Code-specific location
- `docs/kukicha-grammar.ebnf.md` - Formal grammar
- `docs/kukicha-compiler-architecture.md` - Compiler internals
- `docs/tutorials/` - Progressive tutorials
