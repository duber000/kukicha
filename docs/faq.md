# FAQ

## Coming from Bash / Shell Scripting

**Why not just keep writing bash scripts?**

Bash is great for quick one-liners. But once your script hits a few hundred lines, you start running into problems: quoting issues, no real data structures, `set -e` surprises, no type safety.

Kukicha keeps the parts of shell scripting that work — pipes, running commands, readable flow — and adds real types, proper error handling, and compiled binaries.

| Bash Pain Point | Kukicha Solution |
|---|---|
| Quoting hell (`"${var}"`) | `{var}` in strings |
| `set -e` surprises | `onerr` per operation |
| No real data types | `int`, `string`, `bool`, `list of`, `map of` |
| `$1`, `$2` positional args | Named, typed function parameters |
| `if [ ... ]; then ... fi` | `if condition` with indentation |
| Arrays (`"${arr[@]}"`) | `list of string{"a", "b"}` |

**What about Python as a bash replacement?**

Python is a solid option. But Kukicha compiles to a static binary — no runtime, no `pip install`, no virtualenv. `scp` the binary and run it. For scripts that run on remote servers or in containers that matters.

**Where to start:** [Shell Scripters Guide](tutorials/beginner-tutorial.md)

---

## Coming from Python

**Why use Kukicha instead of Python?**

The syntax will feel familiar — `and`/`or`/`not`, indentation, `for x in items`, `# comments`. The differences are:

1. **Static types** — function parameters require explicit types; local variables are inferred.
2. **No implicit returns** — use `return` explicitly.
3. **Error handling** — `onerr` instead of exceptions.
4. **Single binary deployment** — no runtime, no virtualenv, no `pip install` on the target.

| Python | Kukicha |
|---|---|
| `f"{name}"` | `"{name}"` (no prefix) |
| `def greet(name: str) -> str:` | `func Greet(name string) string` |
| `**kwargs` / named args | `F(x: 10)` |
| `try: ... except Exception as e:` | `result onerr return empty, error "{error}"` |
| `[x for x in items if pred(x)]` | `items \|> slice.Filter((x T) => pred(x))` |

If your work is ML/data science, Python's ecosystem (numpy, pandas, PyTorch) has no equivalent here. Kukicha is aimed at infrastructure automation, CLI tools, and AI agent tooling.

**Where to start:** [Beginner Tutorial](tutorials/absolute-beginner-tutorial.md) or skim and jump to the [Quick Reference](kukicha-quick-reference.md).

---

## Coming from Go

**Why not just write Go?**

You already know Go's power. The question is whether the syntax improvements are worth it.

| Go | Kukicha |
|---|---|
| `if err != nil { ... }` | `onerr panic "msg"` |
| `&&`, `\|\|`, `!` | `and`, `or`, `not` |
| `*Type`, `&var` | `reference Type`, `reference of var` |
| `[]string{"a", "b"}` | `list of string{"a", "b"}` |
| `func(s string) bool { return ... }` | `(s string) => ...` |
| `go func() { ... }()` | `go` with indented block |
| `case` / `default` | `when` / `otherwise` |

Everything else is the same: full Go stdlib access, `go mod`, goroutines, channels, interfaces.

**The pipe operator** is probably the biggest addition. Instead of:

```go
result := strings.Title(strings.ToLower(strings.TrimSpace(text)))
```

You write:

```kukicha
result := text |> string.TrimSpace() |> string.ToLower() |> string.Title()
```

Use `_` to control where the piped value lands when it isn't the first argument:

```kukicha
todo |> json.MarshalWrite(response, _)  # Becomes: json.MarshalWrite(response, todo)
```

**Where to start:** [Quick Reference](kukicha-quick-reference.md) — it's a direct Go-to-Kukicha translation table.

---

## General Questions

### I don't know how to program. Can I use Kukicha?

Yes — that's one of the primary use cases. The workflow is:

1. Describe what you want to an AI agent (Claude Code, Cursor, ChatGPT)
2. The agent writes Kukicha code
3. You read and approve it — the English-like syntax makes this possible without a programming background
4. Run `kukicha build` to compile a single binary

See the [Agent Workflow Tutorial](tutorials/agent-workflow-tutorial.md) to get started.

### Does Kukicha have a runtime?

No. The compiler transpiles your code to standard, idiomatic Go. Once compiled by the Go toolchain, there is no trace of Kukicha — just a native Go binary with zero runtime overhead.

### Can I use existing Go libraries?

Yes. Import any Go package (standard library or third-party) and use it directly. The compiler trusts external packages it hasn't seen before, giving you the full Go ecosystem.

### Does Kukicha support named arguments and default parameters?

Yes.

```kukicha
func Connect(host string, port int = 8080, timeout int = 30)
    # ...

Connect("localhost", timeout: 60)
Connect("api.example.com", port: 443, timeout: 120)
```

Named arguments must come after positional arguments. Parameters with defaults must come after those without.

### Does the Kukicha standard library depend on third-party packages?

Most packages use only Go's standard library. The exceptions are packages that wrap functionality Go simply doesn't provide:

| Package | Dependency | Reason |
|---------|-----------|--------|
| `stdlib/parse` | `gopkg.in/yaml.v3` | No built-in YAML parser |
| `stdlib/pg` | `github.com/jackc/pgx/v5` | No built-in PostgreSQL driver |
| `stdlib/container` | `github.com/docker/docker/client` | No built-in Docker SDK |
| `stdlib/kube` | `k8s.io/client-go` | No built-in Kubernetes client |
| `stdlib/mcp` | `github.com/modelcontextprotocol/go-sdk/mcp` | No built-in MCP support |
| `stdlib/a2a` | `github.com/a2aproject/a2a-go` | No built-in A2A protocol |

`go mod tidy` pulls in the relevant dependency when you import one of these packages.

### What editor support is available?

**Zed** has full support:

- Syntax highlighting (Tree-sitter grammar)
- Real-time diagnostics
- Hover information for functions, types, and builtins
- Go-to-definition
- Code completions
- Document symbols (outline view)

```bash
# Build and install the LSP server
make install-lsp

# Ensure GOPATH/bin is in your PATH
export PATH="$PATH:$(go env GOPATH)/bin"

# In Zed: "zed: install dev extension" → select the editors/zed directory
```

### Why bother making an AI Agent friendly language? In the future we won't be able read what they generate anyway

Yes, applications will become specialized neural micro-models whose weights encode behavior; vector embeddings will replace rigid syntax, control flow is differentiable and self-adjusting, and correctness is ensured through formal mathematical proofs instead of tests.

But in the meantime let's try and keep a human in the loop!
