# Kukicha

**Write code that reads like English. Compile it to blazing-fast Go.**

Kukicha is a beginner-friendly programming language that transpiles to idiomatic Go code. No runtime overhead. No magic. Just cleaner syntax that becomes real Go.

```kukicha
import "stdlib/slice"

func isActive(user User) bool
    return user.active

func getName(user User) string
    return user.name

func main()
    users := fetchUsers() onerr panic "failed to fetch"

    active := users
        |> slice.Filter(isActive)
        |> slice.Map(getName)

    for name in active
        print("Hello {name}!")
```

---

## Why Kukicha?

### For Beginners

**Go is powerful but intimidating.** Pointers (`*`, `&`), error handling boilerplate, and cryptic symbols create a steep learning curve.

Kukicha fixes this:

| Go | Kukicha |
|----|---------|
| `&&`, `\|\|`, `!` | `and`, `or`, `not` |
| `*User`, `&user` | `reference User`, `reference of user` |
| `nil` | `empty` |
| `if err != nil { return err }` | `onerr return error "{error}"` |
| `break`, `continue` | `break`, `continue` |
| `for { ... }` | `for` |
| `v.(T)` | `v as T` |

**Learn programming concepts, not symbols.** When you're ready, the generated Go code teaches you Go itself.

### For DevOps Engineers

**Automate infrastructure with Go's reliability, without Go's verbosity.**

```kukicha
import "stdlib/fetch"
import "stdlib/env"
import "stdlib/json"
import "stdlib/retry"
import "stdlib/slice"

# Fetch pod status with retries, filter failures, and alert
func main()
    namespace := env.Get("K8S_NAMESPACE") onerr "default"
    
    cfg := retry.New() |> retry.Attempts(3)
    
    pods := list of Pod{}
    fetch.Get("https://api.k8s.local/pods/{namespace}")
        |> fetch.CheckStatus()
        |> fetch.Bytes()
        |> json.Unmarshal(reference pods)
        onerr panic "k8s unavailable"

    failing := pods
        |> slice.Filter(func(p Pod) bool
            return p.status != "Running"
        )
        |> slice.Map(func(p Pod) string
            return "{p.name}: {p.status}"
        )

    if len(failing) > 0
        print "Pods failing: {len(failing)}"
```

Kukicha compiles to a **single static binary**. No Python dependencies. No Node.js runtime. Just copy and run.

### For Go Developers

**It's still Go.** Kukicha is syntactic sugar, not a new language.

- **Zero runtime overhead** - compiles to idiomatic Go
- **Full Go stdlib access** - import any Go package
- **Gradual adoption** - mix with existing Go code

```kukicha
# This IS Go, just friendlier
import "stdlib/http"
import "stdlib/json"

func HandleUser(w http.ResponseWriter, r reference http.Request)
    user := User{}
    r.Body |> json.UnmarshalRead(reference user) onerr return
    
    # Process user...
    
    user |> json.MarshalWrite(w, _)
```

---

## AI Disclosure

Built with the assistance of AI. Review and test before production use.

---

## The Green Tea Ecosystem

Kukicha (茎茶) is Japanese green tea made from **stems and leaf veins** - the parts usually discarded. We take the "rough edges" of Go and brew something smooth.

| Term | Meaning | Go Equivalent |
|------|---------|---------------|
| **Kukicha** | The language (green tea stems) | - |
| **Stem** | Your project root | Go module |
| **Petiole** | A package (leaf stem) | Go package |
| **Green Tea GC** | Go 1.25+ garbage collector | 10-40% faster GC |

---

## Quick Taste

```kukicha
type Todo
    id int json:"id"
    title string json:"title"
    done bool json:"done"

func Display on todo Todo string
    status := "[ ]"
    if todo.done
        status = "[x]"
    return "{status} {todo.title}"

func main()
    todos := list of Todo{
        Todo
            id: 1
            title: "Learn Kukicha"
            done: true
        Todo
            id: 2
            title: "Build something"
            done: false
    }

    for todo in todos
        print(todo.Display())

```

---

## Smart Pipe Logic 

Kukicha's pipe operator (`|>`) isn't just a simple transformation; it understands Go's common API patterns.

- **Data-First**: `data |> process()` becomes `process(data)`
- **Shorthand Methods**: `result |> .JSON()` becomes `result.JSON()`
- **Context-Aware**: `ctx |> db.Fetch()` becomes `db.Fetch(ctx)`
- **Placeholders**: `user |> json.MarshalWrite(w, _)` becomes `json.MarshalWrite(w, user)`

**Transpiles to clean Go:**

```go
type Todo struct {
    ID    int
    Title string
    Done  bool
}

func (todo Todo) Display() string {
    status := "[ ]"
    if todo.Done {
        status = "[x]"
    }
    return fmt.Sprintf("%s %s", status, todo.Title)
}
```

---

## Install

**Requirements:** Go 1.25+

```bash
# Clone and build
git clone https://github.com/duber000/kukicha.git
cd kukicha
go build -o kukicha ./cmd/kukicha

# Optional: install globally
go install ./cmd/kukicha
```

---

## Usage

```bash
# Initialize a project (extracts stdlib, configures go.mod)
kukicha init

# Compile to binary
kukicha build myapp.kuki

# Compile and run immediately
kukicha run myapp.kuki

# Type-check without compiling
kukicha check myapp.kuki

# Format source files
kukicha fmt myapp.kuki
kukicha fmt -w myapp.kuki      # Write changes
kukicha fmt --check src/       # CI check

# Show version
kukicha version
```

---

## Development

### Run Tests

Kukicha uses Go 1.25+ experimental features (`jsonv2`, `greenteagc`). Use `make` to set the environment automatically, or export `GOEXPERIMENT` yourself:

```bash
make test                        # Sets GOEXPERIMENT automatically
# or manually:
export GOEXPERIMENT=jsonv2,greenteagc
go test ./...                    # All tests
go test ./internal/parser/... -v # Specific package
go test ./... -cover             # With coverage
```

### Project Structure

```
kukicha/
├── cmd/kukicha/     # CLI
├── internal/        # Compiler (lexer, parser, semantic, codegen)
├── stdlib/          # Standard library (iter, slice, fetch, parse, concurrent, http, etc.)
├── docs/            # Documentation
└── examples/        # Example programs
```

### Generated Files

The `stdlib/` `.go` files are **generated** from `.kuki` sources by the transpiler. Always edit the `.kuki` file, then regenerate:

```bash
make generate        # Rebuild kukicha, then regenerate all stdlib .go files
make check-generate  # CI: verify .go files match .kuki sources
```

Do not edit `stdlib/*/*.go` by hand — your changes will be overwritten on the next `make generate`.

### Versioning

The version is defined in a single source of truth: `internal/version/version.go`. To update the version:

1. Edit `internal/version/version.go`
2. Run `make generate` to update all generated file headers

### Adding Features

1. Update docs and grammar specification
2. Implement in the appropriate compiler phase
3. Add tests
4. Submit PR

See [Contributing Guide](docs/contributing.md) for full details.

---

## Documentation

### Learn Kukicha

- [Beginner Tutorial](docs/tutorial/beginner-tutorial.md) - Start here
- [FAQ](docs/faq.md) - Why?
- [Quick Reference](docs/kukicha-quick-reference.md) - Cheat sheet

### Tutorials

- [Console Todo App](docs/tutorial/console-todo-tutorial.md) - Build a CLI app
- [Web App Tutorial](docs/tutorial/web-app-tutorial.md) - Build a REST API
- [Production Patterns](docs/tutorial/production-patterns-tutorial.md) - Best practices

### Technical

- [Compiler Architecture](docs/kukicha-compiler-architecture.md) - How the transpiler works
- [Grammar (EBNF)](docs/kukicha-grammar.ebnf.md) - Formal specification

### Standard Library

- [stdlib Roadmap](docs/kukicha-stdlib-reference.md) - Current packages

---

## Status

**Version:** 0.0.1
**Status:** Ready for testing
**Go:** 1.25+ required

---

## License

See [LICENSE](LICENSE) for details.

