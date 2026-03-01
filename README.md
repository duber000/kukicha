# Kukicha

**AI-assisted coding. Not AI-replaced coding.**

Kukicha is a programming language designed to be read by humans and written by AI agents. It compiles to Go, so your programs run fast and deploy as a single binary with no dependencies.

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

No classes. No `__init__`. No `**kwargs`. No `||`, or `*ptr`. If you can read the code above, you can review what an AI wrote for you.

---

## Why This Matters

AI writes nearly half of all committed code, yet [45% of it contains security flaws](https://www.veracode.com) and AI-assisted codebases show [1.7x more bugs](https://www.coderabbit.ai). The industry is sprinting toward full autonomy — agent loops that generate, deploy, and patch code with no human in the loop.

**Kukicha exists because the answer to "AI writes all the code" shouldn't be "and nobody reads any of it." It should be "and a human can still understand every line."**

AI is the writer. You are the editor. That's not a limitation — it's the architecture of trust.

---

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

## Why Not Just Use Go or Python?

| | Go | Python | Kukicha |
|-|----|---------|----|
| Reads like English | Partially | Yes | Yes |
| Classes / OOP required | No | Common | No |
| Special symbols (`&&`, `__`, `**`) | `&&` | `__`, `**` | No |
| Compiles to single binary | Yes | No | Yes (via Go) |
| Built for AI generation + human review | No | No | Yes |
| Transfers to Go/Python | — | — | 1:1 |

Go and Python were designed for humans to *write*. The bottleneck has shifted from writing to reviewing. Every Kukicha concept maps 1:1 to Go and Python — see the [Quick Reference](docs/kukicha-quick-reference.md) for a full translation table.

---

## Quickstart

### Install

```bash
go install github.com/duber000/kukicha/cmd/kukicha@v0.0.11
kukicha version
```

Or download a release binary from [GitHub Releases](https://github.com/duber000/kukicha/releases).

**Requirements:** Go 1.26+

### First Project

```bash
mkdir myapp && cd myapp
kukicha init          # go mod init + sets up stdlib
kukicha run hello.kuki
```

```kukicha
# hello.kuki
func main()
    name := "World"
    print("Hello, {name}!")
```

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
| `for item in items` | Do this for each item |
| `:=` | Create a new variable |

**Key question when reviewing:** Does each `onerr` say what to do when something fails? If it panics, is that appropriate? If it returns an error, will the caller handle it?

---

## What Can You Build?

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

**More examples:** [AI commit messages](docs/tutorials/data-scripting-tutorial.md), [Concurrent URL health checker](docs/tutorials/concurrent-url-health-checker.md), [REST API link shortener](docs/tutorials/web-app-tutorial.md), [CLI repo explorer](docs/tutorials/cli-explorer-tutorial.md)

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

**Start here:**
- [Agent Workflow Tutorial](docs/tutorials/agent-workflow-tutorial.md) — prompt AI, read and approve, ship
- [Absolute Beginner Tutorial](docs/tutorials/absolute-beginner-tutorial.md) — first program, variables, functions, lists, loops

**Go deeper:**
- [Shell Scripters Guide](docs/tutorials/beginner-tutorial.md) — for bash users
- [Data & AI Scripting](docs/tutorials/data-scripting-tutorial.md) — maps, CSV, shell, LLM
- [CLI Repo Explorer](docs/tutorials/cli-explorer-tutorial.md) — types, methods, API data
- [Link Shortener](docs/tutorials/web-app-tutorial.md) — HTTP servers, JSON, REST APIs
- [Concurrent Health Checker](docs/tutorials/concurrent-url-health-checker.md) — goroutines and channels
- [Production Patterns](docs/tutorials/production-patterns-tutorial.md) — databases, validation, retry, auth

**Reference:**
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
