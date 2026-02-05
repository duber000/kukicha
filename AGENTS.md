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
```
> **Note:** `error` always requires a message string. Use `error "{error}"` to re-wrap the implicit onerr error variable. Multi-statement error handling is supported via indented blocks following `onerr`.

### Types
```kukicha
type Todo
    id int64
    title string json:"title"       # Struct tags supported
    tags list of string
    meta map of string to string
```

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
```

### Pipes
```kukicha
result := data |> parse() |> transform()

# Placeholder _ for non-first position
todo |> json.MarshalWrite(w, _)   # becomes: json.MarshalWrite(w, todo)
```

### Concurrency
```kukicha
ch := make channel of string
send ch, "message"
msg := receive from ch
go doWork()
```

## Build & Test Commands

```bash
make build              # Build the kukicha compiler
make test               # Run all tests (sets GOEXPERIMENT)
make generate           # Regenerate stdlib .go from .kuki sources
kukicha check file.kuki # Validate syntax without compiling
kukicha build file.kuki # Transpile and compile to binary
kukicha run file.kuki   # Transpile, compile, and run
kukicha fmt -w file.kuki # Format in place
```

## File Map

```
cmd/kukicha/          # CLI entry point
internal/
  lexer/              # Tokenization (INDENT/DEDENT handling)
  parser/             # Recursive descent parser → AST
  ast/                # AST node definitions
  semantic/           # Type checking, validation
  codegen/            # AST → Go code generation
  formatter/          # Code formatting
stdlib/               # Standard library (.kuki source files)
  slice/              # Filter, Map, GroupBy, etc.
  json/               # jsonv2 wrapper
  fetch/              # HTTP client
  files/              # File I/O
  shell/              # Command execution
  ...
examples/             # Example programs
docs/                 # Documentation
```

## Critical Rules

1. **Never edit `stdlib/*/*.go` directly** - Edit the `.kuki` files, then run `make generate`
2. **Always validate** - Run `kukicha check` before committing `.kuki` changes
3. **4-space indentation only** - Tabs are not allowed in Kukicha
4. **Explicit function signatures** - Parameters and return types must be declared
5. **Test with `make test`** - Sets required `GOEXPERIMENT=jsonv2,greenteagc`

### Parser Constraints (Last Updated: 2026-02-04)
The following limitations still exist in the compiler:

- **Semantic limit on multi-value pipe return** — `return x |> f()` where `f` returns `(T, error)` parses correctly but currently fails semantic analysis/codegen. **Workaround:** Capture to a variable first: `val, err := x |> f() \n return val, err`.

## Adding Features to the Compiler

Typical workflow for new syntax:
1. **Lexer** (`internal/lexer/`) - Add token type if new keyword/operator
2. **Parser** (`internal/parser/`) - Add parsing logic, create AST nodes
3. **AST** (`internal/ast/`) - Define new node types if needed
4. **Codegen** (`internal/codegen/`) - Generate corresponding Go code
5. **Tests** - Add tests in each modified package

## Stdlib Packages

| Package | Purpose |
|---------|---------|
| `stdlib/slice` | Filter, Map, GroupBy, GetOr, FirstOr, Find, Pop |
| `stdlib/json` | jsonv2 wrapper (Marshal, Unmarshal, streaming) |
| `stdlib/fetch` | HTTP client with builder pattern |
| `stdlib/files` | Read, Write, Watch file operations |
| `stdlib/shell` | Safe command execution |
| `stdlib/cli` | CLI argument parsing |
| `stdlib/concurrent` | Parallel, ParallelWithLimit |
| `stdlib/validate` | Input validation (Email, URL, InRange, NotEmpty) |
| `stdlib/must` | Panic-on-error helpers for startup (Env, EnvInt) |
| `stdlib/env` | Typed env vars with onerr (Get, GetInt, GetBool) |
| `stdlib/datetime` | Named formats, duration helpers (Format, Seconds) |
| `stdlib/http` | HTTP helpers (JSON, JSONError, ReadJSON, GetQueryInt) |

Import with: `import "stdlib/slice"`

### Common Patterns

```kukicha
# Validation (returns error for onerr)
import "stdlib/validate"
email |> validate.Email() onerr return error "{error}"
age |> validate.InRange(18, 120) onerr return error "{error}"

# Startup config (panics if missing/invalid)
import "stdlib/must"
apiKey := must.Env("API_KEY")
port := must.EnvIntOr("PORT", 8080)

# Runtime config (returns error for onerr)
import "stdlib/env"
debug := env.GetBoolOrDefault("DEBUG", false)

# HTTP responses
import "stdlib/http" as httphelper
httphelper.JSON(w, data)
httphelper.JSONNotFound(w, "User not found")

# Time formatting
import "stdlib/datetime"
datetime.Format(t, "iso8601")  # Not "2006-01-02T15:04:05Z07:00"!
timeout := datetime.Seconds(30)
```

## More Documentation

- `.agent/skills/kukicha/` - Comprehensive syntax reference, examples, and troubleshooting (for all AI tools)
- `.claude/skills/kukicha/` - Same content, Claude Code-specific location
- `docs/kukicha-grammar.ebnf.md` - Formal grammar
- `docs/kukicha-compiler-architecture.md` - Compiler internals
- `docs/tutorial/` - Progressive tutorials
