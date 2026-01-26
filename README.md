# Kukicha

**Write code that reads like English. Compile it to blazing-fast Go.**

Kukicha is a beginner-friendly programming language that transpiles to idiomatic Go code. No runtime overhead. No magic. Just cleaner syntax that becomes real Go.

```kukicha
func main()
    users := fetchUsers() onerr panic "failed to fetch"

    active := users
        |> slice.Filter(u -> u.active)
        |> slice.Map(u -> u.name)

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
| `if err != nil { return err }` | `onerr return error` |
| `break`, `continue` | `break`, `continue` |
| `for { ... }` | `for` |

**Learn programming concepts, not symbols.** When you're ready, the generated Go code teaches you Go itself.

### For DevOps Engineers

**Automate infrastructure with Go's reliability, without Go's verbosity.**

```kukicha
# Fetch pod status, filter failures, alert
pods := k8s.ListPods(namespace) onerr panic "k8s unavailable"

failing := pods
    |> slice.Filter(p -> p.status != "Running")
    |> slice.Map(p -> "{p.name}: {p.status}")

if len(failing) > 0
    slack.Alert(channel, "Pods failing:\n" + strings.Join(failing, "\n"))
```

Kukicha compiles to a **single static binary**. No Python dependencies. No Node.js runtime. Just copy and run.

### For Go Developers

**It's still Go.** Kukicha is syntactic sugar, not a new language.

- **Zero runtime overhead** - compiles to idiomatic Go
- **Full Go stdlib access** - import any Go package
- **Gradual adoption** - mix with existing Go code
- **Green Tea GC optimized** - designed for Go 1.25+ performance

```kukicha
# This IS Go, just friendlier
import "net/http"
import "encoding/json/v2"  # Go 1.25+ jsonv2 for 2-10x faster JSON

func HandleUser(w http.ResponseWriter, r reference http.Request)
    user := parseUser(r.Body) onerr
        http.Error(w, "invalid request", 400)
        return

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

Kukicha is optimized for Go's **Green Tea GC** - the experimental garbage collector in Go 1.25 that will become default in Go 1.26. Your Kukicha code is ready for the future.

---

## Quick Taste

```kukicha
type Todo
    id int json:"id"
    title string json:"title"
    done bool json:"done"

func Display on todo Todo string
    status := "[x]" if todo.done else "[ ]"
    return "{status} {todo.title}"

func main()
    todos := list of Todo{
        Todo{id: 1, title: "Learn Kukicha", done: true},
        Todo{id: 2, title: "Build something", done: false},
    }

    for todo in todos
        print(todo.Display())

---

## Smart Pipe Logic 

Kukicha's pipe operator (`|>`) isn't just a simple transformation; it understands Go's common API patterns.

- **Data-First**: `data |> process()` becomes `process(data)`
- **Shorthand Methods**: `result |> .JSON()` becomes `result.JSON()`
- **Context-Aware**: `ctx |> db.Fetch()` becomes `db.Fetch(ctx)`
- **Placeholders**: `user |> json.MarshalWrite(w, _)` becomes `json.MarshalWrite(w, user)`
```

**Transpiles to clean Go:**

```go
type Todo struct {
    ID    int
    Title string
    Done  bool
}

func (todo Todo) Display() string {
    status := "[x]"
    if !todo.Done {
        status = "[ ]"
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

```bash
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

### Adding Features

1. Update docs and grammar specification
2. Implement in the appropriate compiler phase
3. Add tests
4. Submit PR

See [Contributing Guide](docs/contributing.md) for full details.

---

## Documentation

### Learn Kukicha

- [Beginner Tutorial](docs/beginner-tutorial.md) - Start here
- [Language Features](docs/language-features.md) 
- [Quick Reference](docs/kukicha-quick-reference.md) - Cheat sheet

### Tutorials

- [Console Todo App](docs/console-todo-tutorial.md) - Build a CLI app
- [Web App Tutorial](docs/web-app-tutorial.md) - Build a REST API
- [Production Patterns](docs/production-patterns-tutorial.md) - Best practices

### Technical

- [Design Philosophy](docs/kukicha-design-philosophy.md) - Why Kukicha works this way
- [Compiler Architecture](docs/kukicha-compiler-architecture.md) - How the transpiler works
- [Grammar (EBNF)](docs/kukicha-grammar.ebnf.md) - Formal specification

### Standard Library

- [stdlib Roadmap](docs/kukicha-stdlib-roadmap.md) - Current and planned packages

---

## Status

**Version:** 1.0.0
**Status:** Ready for testing
**Go:** 1.25+ required

---

## License

See [LICENSE](LICENSE) for details.

---

<p align="center">
  <strong>Kukicha</strong> - Smooth syntax. Go performance. Green tea vibes.
</p>
