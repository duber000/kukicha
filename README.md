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

No classes. No `__init__`. No `**kwargs`. No `&&`, `||`, or `*ptr`. If you can read the code above, you can review what an AI wrote for you.

---

## Why This Matters Right Now

As of February 2026, AI writes nearly half of all committed code. The industry is sprinting toward full autonomy — autonomous agent loops that generate, deploy, and patch code with no human in the loop.

Here's what that's producing:

- AI-generated code contains **2.74x more security vulnerabilities** than human-written code (Veracode 2025 GenAI Code Security Report)
- **1.7x more logical and correctness bugs** in AI-assisted codebases (CodeRabbit, December 2025)
- **45% of AI-generated code** contains security flaws — and the rate isn't improving with larger models (Veracode, 100+ LLMs tested)
- In January 2026, a vibe-coded startup leaked **1.5 million API keys in three days** because no one read what the AI wrote
- Only **3% of developers** report high trust in AI-generated output (Stack Overflow 2025 Developer Survey) — yet the industry keeps removing them from the review process

The trajectory is clear: AI generates the app, deploys it, monitors it, patches it — and your entire production system becomes code that nobody at your company has ever read.

**Kukicha exists because the answer to "AI writes all the code" shouldn't be "and nobody reads any of it." It should be "and a human can still understand every line."**

The key word is **assisted**. AI is the writer. You are the editor. That's not a limitation — it's the architecture of trust.

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

## Where We Are and Where This Is Going

**Today:** You describe what you want → AI generates code across six languages and config formats you can't fully review → you deploy and hope.

**Kukicha (now):** You describe what you want → AI generates one readable language → you review every line and ship a real binary.

**The future:** AI-native runtimes will collapse the entire stack — from intent to running application — with far fewer intermediate artifacts. The languages, frameworks, and configuration layers that exist today were designed for humans to *write*. That assumption is already breaking down.

Within three years, the best AI-generated applications won't be written in languages designed for humans to type — they'll target the describe-generate-review-ship workflow end to end.

Kukicha is that bridge. As AI capabilities grow, the human review step may get faster — but it should never disappear.

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

Go and Python were designed for humans to *write*. The bottleneck has shifted from writing to reviewing.

Kukicha uses plain English for every operator:

| Other languages | Kukicha |
|----------------|---------|
| `&&`, `\|\|`, `!` | `and`, `or`, `not` |
| `==` (equality) | `equals` |
| `nil`, `None`, `null` | `empty` |
| `[]string`, `list[str]` | `list of string` |
| `*User` (Go pointer) | `reference User` |
| `if err != nil { return err }` | `onerr return empty, error "{error}"` |
| Curly braces `{ }` | 4-space indentation |

---

## Quickstart

### Install

```bash
go install github.com/duber000/kukicha/cmd/kukicha@v0.0.8
kukicha version
```

Or download a release binary for your OS/arch from [GitHub Releases](https://github.com/duber000/kukicha/releases):

```bash
VERSION=v0.0.8
OS=linux   # or darwin, windows
ARCH=amd64 # or arm64
curl -L -o kukicha.tar.gz \
  "https://github.com/duber000/kukicha/releases/download/${VERSION}/kukicha_${VERSION}_${OS}_${ARCH}.tar.gz"
tar -xzf kukicha.tar.gz
./kukicha version
```

**Requirements:** Go 1.26+

### First Project

```bash
mkdir myapp && cd myapp
go mod init myapp
kukicha init          # sets up stdlib
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
| `onerr return empty, error "{error}"` | If this fails, pass the error up |
| `onerr 0` or `onerr "unknown"` | If this fails, use this default value |
| `\|>` | Then pass result to the next step |
| `list of string` | A collection of text values |
| `map of string to int` | A lookup table: text key → number |
| `reference User` | A reference to a User (like a bookmark) |
| `func main()` | Where the program starts |
| `for item in items` | Do this for each item |
| `type Repo` | A data shape definition |
| `:=` | Create a new variable |
| `# comment` | A note — the computer ignores this |

**Key question to ask yourself when reviewing:** Does each `onerr` say what to do when something fails? If it panics, is that appropriate? If it returns an error, will the caller handle it?

---

## What Can You Build?

### Ask an AI to summarize something

```kukicha
import "stdlib/llm"
import "stdlib/shell"

func main()
    diff := shell.Output("git", "diff", "--staged") onerr return

    message := llm.New("openai:gpt-5-nano")
        |> llm.System("Write a concise git commit message for this diff.")
        |> llm.Ask(diff) onerr panic "AI Error: {error}"

    print("Suggested commit message: {message}")
```

### Build a custom AI tool (MCP server)

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

**More examples:** [Concurrent URL health checker](docs/tutorials/concurrent-url-health-checker.md), [REST API link shortener](docs/tutorials/web-app-tutorial.md), [CLI repo explorer](docs/tutorials/cli-explorer-tutorial.md)

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

```kukicha
import "stdlib/fetch"
import "stdlib/slice"
import "stdlib/string"

repos := fetch.Get("https://api.github.com/users/golang/repos")
    |> fetch.CheckStatus()
    |> fetch.Json(list of Repo) onerr panic "{error}"

names := repos
    |> slice.Filter((r Repo) => r.stars > 1000)          # r is each repo
    |> slice.Map((r Repo) => r.name |> string.ToUpper())
```

See the full [Stdlib Reference](stdlib/AGENTS.md).

---

## When You're Ready to Graduate

Every Kukicha concept maps directly to Go and Python. Learning Kukicha first is not a detour.

| Concept | Kukicha | Go | Python |
|---------|---------|-----|--------|
| Variable | `count := 42` | `count := 42` | `count = 42` |
| List | `list of string{"a","b"}` | `[]string{"a","b"}` | `["a","b"]` |
| Loop | `for item in items` | `for _, item := range items` | `for item in items:` |
| Function | `func Add(a int, b int) int` | `func Add(a int, b int) int` | `def add(a: int, b: int) -> int:` |
| Error handling | `result onerr panic "msg"` | `if err != nil { panic("msg") }` | `try: ... except: raise` |
| Null check | `if x equals empty` | `if x == nil` | `if x is None:` |
| Struct | `type User` | `type User struct { ... }` | `@dataclass\nclass User:` |
| Method | `func Greet on u User` | `func (u User) Greet()` | `def greet(self):` |
| Pointer | `reference User` | `*User` | *(implicit reference)* |
| Pipe/chain | `data \|> f() \|> g()` | `g(f(data))` | `g(f(data))` |

Once you understand these patterns in Kukicha, reading Go or Python code will feel familiar.

---

## Usage

```bash
kukicha init                  # Set up stdlib in a new project
kukicha run myapp.kuki        # Compile and run
kukicha build myapp.kuki      # Compile to binary
kukicha check myapp.kuki      # Validate syntax only
kukicha fmt -w myapp.kuki     # Format in place
```

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

**Version:** 0.0.8 — Ready for testing
**Go:** 1.26+ required
**License:** See [LICENSE](LICENSE)
