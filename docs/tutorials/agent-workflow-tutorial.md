# Agent-Assisted Development with Kukicha

Describe what you want. AI writes the code. You can read it.

## Who is this for?

- **Non-programmers** — Build custom tools by directing an AI, no programming language required.
- **SREs / DevOps** — Automate workflows without C-style syntax.
- **AI-First Developers** — Use Claude Code, Cursor, or similar tools as your primary dev interface.
- **Go/Python Devs** — Cleaner, more readable target for AI code generation.

## Prerequisites

1. **Kukicha compiler installed**
```bash
go install github.com/duber000/kukicha/cmd/kukicha@latest
kukicha version
```

2. **An AI coding agent** — [Claude Code](https://claude.com/claude-code) (recommended), [OpenCode](https://opencode.ai/), or [Cursor](https://www.cursor.com/)

3. **Initialize your project:**
   ```bash
   mkdir go_agent_go && cd go_agent_go
   kukicha init
   ```

   `kukicha init` writes a `## Writing Kukicha` section into `AGENTS.md` (creating it if needed), covering syntax, error handling, pipes, and stdlib packages. Most AI coding agents read `AGENTS.md` automatically.

   If your project has a `CLAUDE.md`, `kukicha init` also appends `@AGENTS.md` to it so Claude Code loads the reference at startup.

   Commit `AGENTS.md`. Add only `.kukicha/` to your `.gitignore`.

No prior programming experience needed.

---

## Step 1: Describe Your Intent

You don't write code — you describe what the program should do.

> "Fetch the Go repositories from GitHub, sort them by stars descending, and print a formatted table showing name, stars, and language."

That's it.

---

## Step 2: Agent Generates Kukicha

Here's what the agent generates:

```kukicha
import "stdlib/fetch"
import "stdlib/slice"
import "stdlib/sort"
import "stdlib/table"

type Repo
    name string as "name"
    stars int as "stargazers_count"
    language string as "language"

function main()
    repos := fetch.Get("https://api.github.com/users/golang/repos")
        |> fetch.CheckStatus()
        |> fetch.Json(list of Repo) onerr panic "fetch failed: {error}"

    popular := repos
        |> slice.Filter(r => r.stars > 100)
        |> sort.ByKey(r => r.stars)
        |> slice.Reverse()

    t := table.New("Name", "Stars", "Language")
    for repo in popular
        lang := repo.language
        if lang equals ""
            lang = "—"
        t |> table.AddRow(repo.name, "{repo.stars}", lang)
    t |> table.Print()
```

Four stdlib packages, zero boilerplate. The output is a formatted terminal table, not raw text.

### Prompt templates

**Claude Code or OpenCode:**
```
Write a Kukicha program that fetches Go repos from GitHub, filters to >100 stars,
sorts by stars descending, and prints a table with name, stars, and language.
Use stdlib/fetch, stdlib/slice, stdlib/sort, and stdlib/table.
```

**Cursor or VSCode with Roo Code Extension** (create `repos.kuki`, then in composer):
```
Generate a Kukicha program (not Go) that fetches golang repos from the GitHub API,
filters by star count > 100, sorts by stars descending, and displays a table.
```


---

## Step 3: Read and Approve

You don't need to understand every detail — verify the program does what you intended.

### Decoder ring

| You'll see | It means |
|-----------|---------|
| `onerr panic "message"` | If this fails, stop with this message |
| `onerr return` | If this fails, pass the error to the caller |
| `onerr "default"` | If this fails, use this value instead |
| `\|>` | Pass the result to the next step |
| `list of string` | A collection of text values |
| `map of string to int` | A lookup table: text key → number |
| `reference User` | A reference/pointer to a User value |
| `func main()` | Where the program starts |
| `for item in items` | Do this for each item |
| `type Repo` | A data shape — describes what a Repo looks like |
| `:=` | Create a new variable |
| `# comment` | A human note — the compiler ignores it |

### Review checklist

- [ ] Does it do what I described? Read through `main()`.
- [ ] Are errors handled? Every operation that can fail should have `onerr`.
- [ ] Does `onerr panic` make sense here? Fine for tools; prefer `onerr return` for production.
- [ ] Any hardcoded secrets? (Use `stdlib/env` or `stdlib/must` for those.)
- [ ] Are external URLs correct?

---

## Step 4: Compile and Run

```bash
kukicha check repos.kuki    # validate syntax
kukicha run repos.kuki       # compile and run
kukicha build repos.kuki -o repos   # build a binary to ship
./repos
```

If `kukicha check` reports errors, paste the error into your agent and ask it to fix.

---

## Iterating With Your Agent

### Add retry + logging

```
Add retry logic (3 attempts, 500ms delay) and structured logging.
```

```kukicha
import "stdlib/obs"
import "stdlib/fetch"

function main()
    log := obs.New("repos")
    log |> obs.Info("fetching repositories")

    repos := fetch.Get("https://api.github.com/users/golang/repos")
        |> fetch.Retry(3, 500)
        |> fetch.CheckStatus()
        |> fetch.Json(list of Repo) onerr panic "fetch failed: {error}"

    log |> obs.Info("found {len(repos)} repos")
```

### Add a CLI with flags

```
Add a --min-stars flag (default 100) and a --format flag (table or json).
```

```kukicha
import "stdlib/cli"
import "stdlib/json"
import "stdlib/table"

function run(args cli.Args)
    minStars := cli.GetInt(args, "min-stars")
    format := cli.GetString(args, "format")

    # ... fetch and filter repos ...

    if format equals "json"
        repos |> json.MarshalPretty() |> print
    else
        t := table.New("Name", "Stars")
        for repo in popular
            t |> table.AddRow(repo.name, "{repo.stars}")
        t |> table.Print()

function main()
    app := cli.New("repos")
        |> cli.Description("Browse GitHub repos by star count")
        |> cli.AddFlag("min-stars", "Minimum star count", "100")
        |> cli.AddFlag("format", "Output format: table or json", "table")
        |> cli.Action(run)
    cli.RunApp(app) onerr panic "{error}"
```

### Fetch in parallel

```
Fetch repos from multiple GitHub users concurrently.
```

```kukicha
import "stdlib/concurrent"
import "stdlib/fetch"
import "stdlib/slice"

function fetchUser(user string) list of Repo
    repos := fetch.Get("https://api.github.com/users/{user}/repos")
        |> fetch.Retry(3, 500)
        |> fetch.CheckStatus()
        |> fetch.Json(list of Repo) onerr panic "fetch {user} failed: {error}"
    return repos

function main()
    users := list of string{"golang", "googlecloudplatform", "kubernetes"}

    allRepos := users
        |> concurrent.Map(u => fetchUser(u))
        |> slice.Concat()
        |> slice.Filter(r => r.stars > 100)
        |> sort.ByKey(r => r.stars)
        |> slice.Reverse()

    # ... print table ...
```

`concurrent.Map` fetches all three users in parallel and collects the results.

### Write results to CSV

```
Write the results to results.csv instead of printing a table.
```

```kukicha
import "stdlib/files"
import "stdlib/string"

    header := "name,stars,language"
    lines := popular |> slice.Map(r => "{r.name},{r.stars},{r.language}")
    content := string.Join(slice.Concat(list of string{header}, lines), "\n")
    files.Write("results.csv", content) onerr panic "write failed: {error}"
```

### Debug a problem

Paste the error output into your agent:
```
Running repos.kuki failed with this error:
<paste the error here>
Fix it.
```

---

## Building an MCP Server

MCP (Model Context Protocol) lets you extend AI agents with custom tools. Build a single binary that your agent can call.

### Example: DNS lookup tool

```kukicha
import "stdlib/mcp"
import "stdlib/net"
import "stdlib/string"

function lookupHost(hostname string) string
    ips := net.LookupHost(hostname) onerr return "lookup failed: {error}"
    return string.Join(ips, ", ")

function main()
    server := mcp.NewServer()
    server |> mcp.Tool("dns_lookup", "Resolve a hostname to IP addresses", lookupHost)
    server |> mcp.Serve()
```

### Example: secure file reader (sandboxed)

```kukicha
import "stdlib/mcp"
import "stdlib/sandbox"

function readFile(path string) string
    sb := sandbox.New("/var/data") onerr return "sandbox error: {error}"
    defer sb |> sandbox.Close()
    content := sb |> sandbox.ReadString(path) onerr return "read failed: {error}"
    return content

function main()
    server := mcp.NewServer()
    server |> mcp.Tool("read_data", "Read a file from the data directory (sandboxed)", readFile)
    server |> mcp.Serve()
```

### Compile and register

```bash
kukicha build dns-tool.kuki -o dns-tool
```

Add to Claude Code's MCP config:
```json
{
  "mcpServers": {
    "dns": {
      "command": "/absolute/path/to/dns-tool"
    }
  }
}
```

### MCP review checklist

- [ ] Tool functions take only simple types (string, int, bool)?
- [ ] Tool returns a string or simple value?
- [ ] Errors handled so the server doesn't crash on bad input?
- [ ] Tool description clear enough for the AI to know when to use it?

---

## Where Concepts Transfer

Everything you learn maps directly to Go and Python.

| Concept | Kukicha | Go | Python |
|---------|---------|-----|--------|
| Variable | `count := 42` | `count := 42` | `count = 42` |
| List | `list of int{1, 2}` | `[]int{1, 2}` | `[1, 2]` |
| Loop | `for item in items` | `for _, item := range items` | `for item in items:` |
| Error handling | `result onerr panic "msg"` | `if err != nil { panic("msg") }` | `try: ... except: raise` |
| Function | `func Add(a int, b int) int` | `func Add(a int, b int) int` | `def add(a: int, b: int) -> int:` |
| Null check | `if x equals empty` | `if x == nil` | `if x is None:` |
| Pipe/chain | `data \|> f() \|> g()` | `g(f(data))` | `g(f(data))` |

Kukicha is Go with English keywords and indentation instead of symbols and braces. If you can read Kukicha, you can read Go.

---

## Next Steps

- [Absolute Beginner Tutorial](absolute-beginner-tutorial.md) — syntax from scratch
- [Data & AI Scripting](data-scripting-tutorial.md) — maps, CSV, LLM integration
- [Production Patterns](production-patterns-tutorial.md) — databases, validation, retry, auth

---

## Tips for Effective Agent Prompting

1. **Be specific about intent, not implementation**
   - Bad: "Add error handling"
   - Good: "If the API call fails, print the error and exit"

2. **Name the stdlib packages you want**
   - "Use stdlib/fetch for HTTP, stdlib/slice for filtering, stdlib/table for output"

3. **Iterate in small steps**
   - Bad: "Make this production-ready with logging, metrics, and retries"
   - Good: "Add structured logging using stdlib/obs"

4. **Always validate before shipping**
   ```bash
   kukicha check file.kuki
   kukicha run file.kuki
   kukicha build file.kuki
   ```

5. **If the agent generates Go syntax by mistake:**
   ```
   You wrote Go syntax. Rewrite in Kukicha:
   - "and", "or", "not" instead of &&, ||, !
   - "list of string" instead of []string
   - "equals" instead of ==
   - 4-space indentation, no braces
   - "onerr" instead of "if err != nil"
   ```
