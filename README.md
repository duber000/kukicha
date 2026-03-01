# Kukicha

**AI-assisted coding. Not AI-replaced coding.**

You describe what you want. AI writes the code. Kukicha lets you actually *read* it.

Kukicha is a programming language designed to be read by humans and written by AI agents. It compiles to Go, so your programs run fast and deploy as a single binary with no dependencies.

## The Workflow

```
You describe what you want
        ↓
AI agent writes Kukicha
        ↓
You read and approve it  ← Kukicha makes this step possible
        ↓
kukicha build → single binary
        ↓
Ship it
```

You don't need to know how to write Kukicha — you just need to *read* it and spot when something looks wrong.

See the [Agent Workflow Tutorial](docs/tutorials/agent-workflow-tutorial.md) to get started immediately.

---

## Here's What AI-Generated Kukicha Looks Like

See if you can follow it:

```kukicha
import "stdlib/fetch"
import "stdlib/slice"

type Repo
    name string as "name"
    stars int as "stargazers_count"

func main()
    repos := fetch.Get("https://api.github.com/users/golang/repos")
        |> fetch.CheckStatus()
        |> fetch.Json(list of Repo) onerr panic "fetch failed: {error}"

    popular := repos |> slice.Filter((repo Repo) => repo.stars > 1000)

    for repo in popular
        print("{repo.name}: {repo.stars} stars")
```

**What just happened:** This program fetches a list of repositories from GitHub, keeps only the ones with more than 1,000 stars, and prints each name with its star count. The `|>` (pipe) passes the result of one step into the next. `onerr panic` means "if something goes wrong, stop and show this message."

---

## What to Read in Agent-Generated Code

When AI writes Kukicha for you, here's the decoder ring:

| You'll see | It means |
|-----------|---------|
| `onerr panic "message"` | If this fails, crash with message |
| `onerr return` | If this fails, pass the error up |
| `onerr 0` or `onerr "unknown"` | If this fails, use this default value |
| `\|>` | Then pass result to the next step |
| `list of string` | A collection of text values |
| `map of string to int` | A lookup table: text key → number |
| `reference User` | A reference to a User (like a bookmark) |
| `(x) => x + 1` | A shorthand function: given x, return x + 1 |
| `for item in items` | Do this for each item |
| `:=` | Create a new variable |

**Key question when reviewing:** Does each `onerr` say what to do when something fails? If it panics, is that appropriate? If it returns an error, will the caller handle it?

---

## Quickstart

### Prerequisites

**Go 1.26+** is required. If you don't have Go installed, [download it here](https://go.dev/dl/).

### Install Kukicha

**Option A — Binary download (no Go toolchain needed after install):**

Download a pre-built binary from [GitHub Releases](https://github.com/duber000/kukicha/releases).

**Option B — Install with Go:**

```bash
go install github.com/duber000/kukicha/cmd/kukicha@v0.0.11
kukicha version
```

### Your First Project

```bash
mkdir myapp && cd myapp
kukicha init          # sets up your project and stdlib
```

Create a file called `hello.kuki`:

```kukicha
# hello.kuki
func main()
    name := "World"
    print("Hello, {name}!")
```

Then run it:

```bash
kukicha run hello.kuki
```

### Which Command Do I Use?

| Command | What it does | When to use it |
|---------|-------------|----------------|
| `kukicha check file.kuki` | Validates syntax without running anything | Before committing — catches errors early |
| `kukicha run file.kuki` | Compiles and runs immediately | While developing and testing |
| `kukicha build file.kuki` | Compiles to a standalone binary | When you're ready to ship |

---

## What Can You Build?

### A Tool for Your AI Agent (MCP Server)

```kukicha
import "stdlib/mcp"
import "stdlib/fetch"

func getPrice(symbol string) string
    price := fetch.Get("https://api.example.com/price/{symbol}")
        |> fetch.CheckStatus()
        |> fetch.Text() onerr return "unavailable"
    return "{symbol}: {price}"

func main()
    server := mcp.NewServer()
    server |> mcp.Tool("get_price", "Get stock price by ticker symbol", getPrice)
    server |> mcp.Serve()
```

Compile to a single binary and register it with Claude Desktop or any MCP-compatible agent.

### A Simple Automation Script

```kukicha
import "stdlib/files"
import "stdlib/string"

func main()
    content := files.Read("notes.txt") onerr panic "can't read file: {error}"
    lines := content |> string.Split("\n")
    for line in lines
        if line |> string.Contains("TODO")
            print(line)
```

Read a file, find every line containing "TODO", and print them out.

**More examples:** [AI commit messages](docs/tutorials/data-scripting-tutorial.md), [Concurrent URL health checker](docs/tutorials/concurrent-url-health-checker.md), [REST API link shortener](docs/tutorials/web-app-tutorial.md), [CLI repo explorer](docs/tutorials/cli-explorer-tutorial.md)

---

## Beyond Readability

Kukicha compiles to Go — single binaries, goroutines, strong typing, memory safety, fast cold starts. No runtime dependencies.

But readability alone isn't enough. AI writes [nearly half of all committed code](https://shiftmag.dev/state-of-code-2025-7978/), yet [45% of it contains security flaws](https://www.veracode.com/blog/genai-code-security-report/) and AI-generated code introduces [1.7x more issues](https://www.coderabbit.ai/blog/state-of-ai-vs-human-code-generation-report). In 2026, [security debt hit 82% of organizations](https://www.veracode.com/resources/analyst-reports/state-of-software-security-2026/) and AI-generated code [introduces 15–18% more security vulnerabilities](https://opsera.ai/resources/report/ai-coding-impact-2025-benchmark-report/) at enterprise scale.

The Kukicha compiler catches common AI-generated vulnerabilities **at build time** — SQL injection, XSS, SSRF, path traversal, command injection, and open redirects — and tells both you and the AI agent what safe alternative to use instead.

AI is the writer. You are the editor. That's not a limitation — it's the architecture of trust.

---

## For Developers

If you already know Go or Python, here's how Kukicha compares:

| | Go | Python | Kukicha |
|-|----|---------|----|
| Reads like English | Partially | Yes | Yes |
| Classes / OOP required | No | Common | No |
| Special symbols (`&&`, `__`, `**`) | `&&` | `__`, `**` | No |
| Compiles to single binary | Yes | No | Yes (via Go) |
| Built for AI generation + human review | No | No | Yes |
| Transfers to Go/Python | — | — | 1:1 |

Every Kukicha concept maps 1:1 to Go and Python — see the [Quick Reference](docs/kukicha-quick-reference.md) for a full translation table.

---

## Standard Library

35+ packages, pipe-friendly, error-handled with `onerr`.

| Category | Packages |
|---------|---------|
| **Data** | `fetch`, `files`, `json`, `parse`, `encoding` |
| **Logic** | `slice`, `maps`, `string`, `math`, `iterator` |
| **Infrastructure** | `pg`, `kube`, `container`, `shell` |
| **AI & Agents** | `llm`, `mcp`, `a2a` |
| **Web** | `http`, `fetch`, `validate`, `netguard` |
| **Config & Ops** | `env`, `must`, `cli`, `obs`, `retry`, `ctx` |

See the full [Stdlib Reference](stdlib/AGENTS.md).

---

## Editor Support

**VS Code:** Search `kukicha-lang` in extensions, or download the `.vsix` from [GitHub Releases](https://github.com/duber000/kukicha/releases). See [`editors/vscode/README.md`](editors/vscode/README.md).

**Zed:** Open Zed → `zed: install dev extension` → point to `editors/zed/` in this repo.

**Other editors:** `make install-lsp` and configure your editor to run `kukicha-lsp` for `.kuki` files.

All editors get syntax highlighting, hover, go-to-definition, completions, and diagnostics via the LSP.

---

## Documentation

**New to Kukicha?**
- [Agent Workflow Tutorial](docs/tutorials/agent-workflow-tutorial.md) — prompt AI, read and approve, ship
- [Absolute Beginner Tutorial](docs/tutorials/absolute-beginner-tutorial.md) — first program, variables, functions, lists, loops

**Build something real:**
- [Data & AI Scripting](docs/tutorials/data-scripting-tutorial.md) — maps, CSV, shell, LLM
- [CLI Repo Explorer](docs/tutorials/cli-explorer-tutorial.md) — types, methods, API data
- [Link Shortener](docs/tutorials/web-app-tutorial.md) — HTTP servers, JSON, REST APIs
- [Concurrent Health Checker](docs/tutorials/concurrent-url-health-checker.md) — goroutines and channels

**Go deeper:**
- [Shell Scripters Guide](docs/tutorials/beginner-tutorial.md) — for bash users
- [Production Patterns](docs/tutorials/production-patterns-tutorial.md) — databases, validation, retry, auth
- [FAQ](docs/faq.md) — coming from bash, Python, or Go
- [Quick Reference](docs/kukicha-quick-reference.md) — Go-to-Kukicha translation table
- [Stdlib Reference](stdlib/AGENTS.md) — all packages

---

## Contributing

See [Contributing Guide](docs/contributing.md) for development setup, tests, and architecture.

---

## Status

**Version:** 0.0.11 — Ready for testing
**Go:** 1.26+ required
**License:** See [LICENSE](LICENSE)
