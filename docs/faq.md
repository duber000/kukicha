# FAQ

## Coming from Bash / Shell Scripting

**Why not just keep writing bash scripts?**

Bash is great for quick one-liners and gluing commands together. But once your script hits a few hundred lines, you start running into problems: quoting issues, no real data structures, error handling with `set -e` that surprises you, no type safety, and debugging with `echo` everywhere.

Kukicha keeps the parts of shell scripting that work well - pipes, running commands, readable flow - and gives you real types, proper error handling, and compiled binaries.

| Bash Pain Point | Kukicha Solution |
|---|---|
| Quoting hell (`"${var}"`) | Just use `{var}` in strings |
| `set -e` surprises | `onerr` per operation |
| No real data types | `int`, `string`, `bool`, `list of`, `map of` |
| `$1`, `$2` positional args | Named, typed function parameters |
| `if [ ... ]; then ... fi` | `if condition` with indentation |
| Arrays (`"${arr[@]}"`) | `list of string{"a", "b"}` |
| Word splitting breaks things | Arguments are separate strings, always safe |

**What about Python as a bash replacement?**

Python is a solid option. But Kukicha compiles to a static binary - no runtime, no `pip install`, no virtualenv on the target machine. `scp` the binary and run it. For scripts that need to run on remote servers, in containers, or as CLI tools you distribute, that matters.

**Where to start:** The [Beginner Tutorial](tutorials/beginner-tutorial.md) is written specifically for shell scripters. Every section shows the bash way first, then the Kukicha equivalent.

---

## Coming from Python

**Why use Kukicha instead of Python?**

Kukicha was designed to feel familiar to Python developers. You already know a lot of the syntax:

| Python | Kukicha | Notes |
|---|---|---|
| `and`, `or`, `not` | `and`, `or`, `not` | Identical |
| `if x == y:` | `if x equals y` | English keyword |
| `for x in items:` | `for x in items` | Same iteration style |
| `# comment` | `# comment` | Identical |
| Indentation (4 spaces) | Indentation (4 spaces) | Identical |
| f-strings `f"{name}"` | `"{name}"` | No prefix needed |
| `def greet(name):` | `func Greet(name string)` | Types required |
| Default params | `func F(x int = 10)` | Same concept |
| `**kwargs` / named args | `F(x: 10)` | Clean syntax |

### Key Differences

1. **Static types** - function parameters require explicit types. Local variables are inferred.
   ```kukicha
   func Greet(name string)      # 'name' must specify type
       message := "Hi"          # 'message' type is inferred
   ```

2. **No implicit returns** - use `return` explicitly.

3. **Error handling** - Python uses exceptions; Kukicha uses `onerr` for inline handling.
   ```kukicha
   data := files.Read("config.json") onerr panic "failed"
   ```

4. **The pipe operator** - chain operations in a data-flow style.
   ```kukicha
   result := text |> string.TrimSpace() |> string.ToLower()
   ```

### When to Stay with Python

Python is improving rapidly - Python 3.13+ has experimental free-threading (no GIL), and **uv** makes dependency management much smoother. If your workloads are I/O-bound or ML-focused, and you're already productive in Python, it may not be worth switching.

### When Kukicha Makes Sense

| Consideration | Python | Kukicha |
|---|---|---|
| **Deployment** | Better with `uv` + containers, still needs a runtime | Single static binary, zero dependencies |
| **Concurrency** | Free-threading is experimental; async has limits | Goroutines are production-proven |
| **Type safety** | Optional (mypy, pyright) - runtime errors still possible | Mandatory compile-time checking |
| **Performance** | Slow for CPU-bound tasks | 10-100x faster, compiles to native code |
| **Ecosystem** | Massive, unparalleled for ML/data science | Full access to Go's ecosystem |

**Where to start:** The [Beginner Tutorial](tutorials/beginner-tutorial.md) covers the fundamentals. If you're comfortable with programming concepts already, you may want to skim it and jump to the [Quick Reference](kukicha-quick-reference.md) to see the full syntax mapping.

---

## Coming from Go

**Why not just write Go?**

You already know Go, so you already know Kukicha's underlying power. The question is whether the syntax improvements are worth it for you.

| Go | Kukicha | Notes |
|---|---|---|
| `if err != nil { ... }` | `onerr panic "msg"` | Inline, one line |
| `&&`, `\|\|`, `!` | `and`, `or`, `not` | English keywords |
| `*Type`, `&var` | `reference Type`, `reference of var` | Readable pointers |
| `[]string{"a", "b"}` | `list of string{"a", "b"}` | Readable collections |
| `map[string]int{}` | `map of string to int{}` | Readable maps |
| `func(s string) bool { return ... }` | `(s string) => ...` | Arrow lambdas |
| Curly braces + semicolons | Indentation (4 spaces) | Python-style blocks |
| `go func() { ... }()` | `go` with indented block | No IIFE pattern |
| `case` / `default` | `when` / `otherwise` | English switch |

### What Stays the Same

- Full access to the Go standard library and third-party packages
- `go mod` for dependency management
- Goroutines and channels (with friendlier syntax: `channel of Type`, `send val to ch`, `receive from ch`)
- All Go types and interfaces
- Compiles to idiomatic Go, then to a native binary

### The Pipe Operator

This is probably the biggest feature Go doesn't have. Instead of:

```go
result := strings.Title(strings.ToLower(strings.TrimSpace(text)))
```

You write:

```kukicha
result := text |> string.TrimSpace() |> string.ToLower() |> string.Title()
```

For Go's "writer-first" or "context-first" APIs, Kukicha uses smart pipe logic:

```kukicha
# Context is automatically prepended as first arg
ctx |> db.FetchUser(userID)  # Becomes: db.FetchUser(ctx, userID)

# Use _ to specify argument position
todo |> json.MarshalWrite(response, _)  # Becomes: json.MarshalWrite(response, todo)
```

**Where to start:** The [Quick Reference](kukicha-quick-reference.md) is a direct Go-to-Kukicha translation table. That's likely all you need to get going.

---

## General Questions

### Does the Kukicha standard library depend on third-party packages?

Wherever possible, Kukicha's stdlib is built on Go's standard library alone. Most packages —
`slice`, `string`, `json`, `files`, `net`, `errors`, `encoding`, `fetch`, `llm`, `shell`,
`env`, `validate`, and others — use only packages that ship with Go itself.

The exceptions are packages that wrap functionality Go's stdlib simply does not provide:

| Package | Third-party dependency | Why no stdlib alternative |
|---------|----------------------|--------------------------|
| `stdlib/parse` | `gopkg.in/yaml.v3` | Go has no built-in YAML parser |
| `stdlib/pg` | `github.com/jackc/pgx/v5` | Go has no built-in PostgreSQL driver |
| `stdlib/container` | `github.com/docker/docker/client` | Go has no built-in Docker/Podman SDK |
| `stdlib/kube` | `k8s.io/client-go` | Go has no built-in Kubernetes client |
| `stdlib/mcp` | `github.com/modelcontextprotocol/go-sdk/mcp` | Go has no built-in MCP support |
| `stdlib/a2a` | `github.com/a2aproject/a2a-go` | Go has no built-in A2A protocol |

`stdlib/json` uses `encoding/json/v2`, which is part of Go 1.26+ (enabled via
`GOEXPERIMENT=jsonv2`) rather than a third-party package.

When you import one of the exception packages, `go mod tidy` will pull in the corresponding
dependency automatically.

---

### Does Kukicha have a runtime?

No. Kukicha has zero runtime overhead. The compiler transpiles your code into standard, idiomatic Go. Once compiled by the Go toolchain, there is no trace of Kukicha left - just a native Go binary.

### Can I use existing Go libraries?

Yes. You can import any Go package (standard library or third-party) and use it directly in Kukicha. If the compiler hasn't seen the type before, it trusts the external package, allowing you to use the entire Go ecosystem immediately.

### Does Kukicha support named arguments and default parameters?

Yes. Default parameters let you specify fallback values:

```kukicha
func Greet(name string, greeting string = "Hello")
    print("{greeting}, {name}!")

Greet("Alice")          # "Hello, Alice!"
Greet("Bob", "Hi")      # "Hi, Bob!"
```

Named arguments let you specify argument names at the call site:

```kukicha
func Connect(host string, port int = 8080, timeout int = 30)
    # ...

Connect("localhost", timeout: 60)
Connect("api.example.com", port: 443, timeout: 120)
```

Named arguments must come after positional arguments, and parameters with defaults must come after those without.

### Doesn't XGo (formerly Go+) already do this?

XGo is an excellent project, but it serves a different niche:

- **Semantic keywords** - Kukicha replaces symbols with English words (`reference of user` instead of `&user`, `and`/`or` instead of `&&`/`||`). XGo stays closer to standard Go syntax.
- **The pipe operator** - Kukicha is built around a data-flow philosophy. The smart pipe (`|>`) supports placeholders (`_`), making it readable for DevOps and API logic.
- **Error handling** - Kukicha's `onerr` keyword removes the `if err != nil` boilerplate without hiding errors behind exceptions.

### What editor support is available?

**Zed** is currently supported with full language support:

- Syntax highlighting (Tree-sitter grammar)
- Real-time diagnostics from the parser and semantic analyzer
- Hover information for functions, types, interfaces, and builtins
- Go-to-definition for functions, types, interfaces, and fields
- Code completions for keywords, builtins, types, and declarations
- Document symbols (outline view)

**Installation:**

```bash
# 1. Build and install the LSP server
make install-lsp

# 2. Ensure GOPATH/bin is in your PATH
export PATH="$PATH:$(go env GOPATH)/bin"

# 3. Install the Zed extension (from the kukicha repo)
# In Zed, run: "zed: install dev extension"
# Select the editors/zed directory
```
