# Agent-Assisted Development with Kukicha

Kukicha is designed for a specific workflow: you describe what you want, an AI agent writes the code, you read and approve it, then you compile and ship. This works because Kukicha's syntax is intentionally readable — no cryptic symbols, no class hierarchies, no boilerplate.

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

## Who is this for?

- **Non-programmers** — People who want to build custom tools by directing an AI, without needing to learn a full programming language.
- **SREs / Sysadmins / DevOps engineers** — Operators who want to automate workflows without getting bogged down in C-style syntax.
- **AI-First Developers** — People who use Claude Code, Cursor, or similar tools as their primary development interface.
- **Go/Python Devs** — Developers who want a cleaner, more readable target for AI code generation.

## What You'll Learn

- How to prompt AI agents to write Kukicha code
- How to read and approve agent-generated Kukicha (no programming background needed)
- How to iterate: ask your agent to add features, fix bugs, or refactor
- How to build MCP servers that extend your AI agent
- How Kukicha concepts transfer to Go and Python

## Prerequisites

1. **Kukicha compiler installed**
   ```bash
   go install github.com/duber000/kukicha/cmd/kukicha@v0.0.11
   kukicha version  # Confirm it works
   ```

2. **An AI coding agent** — one of:
   - [Claude Code](https://claude.com/claude-code) (recommended)
   - [Cursor](https://www.cursor.com/)
   - [ChatGPT](https://chatgpt.com/) with Code Interpreter

3. **Initialize your project** — this sets up the stdlib and the language reference your agent needs:
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

The first step is writing a clear prompt. You don't write code — you describe what the program should do.

**Example intent:**
> "Fetch the Go repositories from GitHub, keep only the ones with more than 1000 stars, and print the name and star count of each."

That's it. The AI handles the rest.

---

## Step 2: Agent Generates Kukicha

Here's what the agent should generate for the prompt above:

```kukicha
import "stdlib/fetch"
import "stdlib/slice"

type Repo
    name string as "name"
    stars int as "stargazers_count"

func main()
    # Fetch the repos list from the GitHub API
    repos := fetch.Get("https://api.github.com/users/golang/repos")
        |> fetch.CheckStatus()
        |> fetch.Json(list of Repo) onerr panic "fetch failed: {error}"

    # Filter to repos with more than 1000 stars
    popular := repos |> slice.Filter((r Repo) => r.stars > 1000)

    # Print each result
    for repo in popular
        print("{repo.name}: {repo.stars} stars")
```

### Prompt templates to try

**Claude Code** (run in your terminal):
```
Write a Kukicha program that fetches the latest Go repositories from GitHub,
filters ones with >1000 stars, and prints the name and star count for each.
Use stdlib/fetch and stdlib/slice.
```

**Cursor** (create `repos.kuki`, then in composer):
```
Generate a Kukicha program (not Go) that fetches golang repos from the GitHub API,
filters by star count > 1000, and prints name + stars for each result.
```

**ChatGPT or other LLMs**:
```
Write a complete, working Kukicha program (not Go) that:
1. Imports stdlib/fetch and stdlib/slice
2. Defines a type Repo with fields name (JSON: "name") and stars (JSON: "stargazers_count")
3. Fetches repos from https://api.github.com/users/golang/repos
4. Filters repos with > 1000 stars using slice.Filter
5. Prints the name and star count of each result
Use fetch.Json(list of Repo) for JSON decoding and onerr for error handling.
```

---

## Step 3: Read and Approve

This is the most important step. You don't need to understand every detail — you need to verify the program does what you intended.

### Decoder ring

| You'll see | It means |
|-----------|---------|
| `onerr panic "message"` | If this fails, stop the program with this message |
| `onerr return` | If this fails, pass the error to the caller |
| `onerr "default"` | If this fails, use this value instead |
| `\|>` | Take the result and pass it to the next step |
| `list of string` | A collection of text values |
| `map of string to int` | A lookup table: text key → number |
| `reference User` | A reference/pointer to a User value |
| `func main()` | Where the program starts |
| `for item in items` | Do this for each item |
| `type Repo` | A data shape — describes what a Repo looks like |
| `:=` | Create a new variable |
| `# comment` | A human note — the compiler ignores it |

### Review checklist

Before running the code, ask yourself:

- [ ] Does it do what I described? Read through the steps in `main()`.
- [ ] Are errors handled? Every operation that can fail should have `onerr`.
- [ ] Does `onerr panic` make sense here? Panicking is fine for prototype tools; for production you'd prefer `onerr return`.
- [ ] Are there any hardcoded secrets or credentials? (There shouldn't be — use `stdlib/env` or `stdlib/must` for those.)
- [ ] Does the output format match what you want? Check any `print()` calls.
- [ ] Are external URLs correct? Verify any API endpoints match what you intended.

**In the example above:**
- `fetch.Get(...)` — fetches from GitHub. Correct URL?
- `|> fetch.CheckStatus()` — fails if the API returns an error. Good.
- `|> fetch.Json(list of Repo)` — decodes the JSON as a list of Repos. Matches the `type Repo` definition?
- `onerr panic "fetch failed: {error}"` — crashes with an error message if anything in the pipeline fails. Acceptable for a one-off tool.
- `slice.Filter((r Repo) => r.stars > 1000)` — keeps repos with > 1000 stars. Matches the intent?

---

## Step 4: Compile and Run

```bash
# Validate syntax first
kukicha check repos.kuki

# Run it
kukicha run repos.kuki

# Or build a binary to ship
kukicha build repos.kuki -o repos
./repos
```

If `kukicha check` reports errors, paste the error into your agent and ask it to fix them.

---

## Iterating With Your Agent

Your agent doesn't just generate code once — you can refine, add features, and debug together.

### Add retry logic

Ask:
```
Add retry logic to the fetch call. Retry 3 times with 500ms delay between attempts.
```

Agent adds one line:
```kukicha
    repos := fetch.Get("https://api.github.com/users/golang/repos")
        |> fetch.Retry(3, 500)      # ← Agent added this
        |> fetch.CheckStatus()
        |> fetch.Json(list of Repo) onerr panic "fetch failed: {error}"
```

### Add a CLI flag

Ask:
```
Add a --min-stars flag (default 1000) so the user can control the minimum star count.
```

Agent adds:
```kukicha
import "stdlib/cli"

func run(args cli.Args)
    minStars := cli.GetInt(args, "min-stars")
    # ... use minStars instead of the hardcoded 1000 ...

func main()
    app := cli.New("repos")
        |> cli.AddFlag("min-stars", "Minimum star count", "1000")
        |> cli.Action(run)
    cli.RunApp(app) onerr panic "failed: {error}"
```

### Add output to a file

Ask:
```
Write the results to a file called output.txt instead of printing to the screen.
```

Agent will add `import "stdlib/files"` and replace the `print()` calls with `files.Append("output.txt", ...)`.

### Debug a problem

If something goes wrong, paste the error output into your agent:
```
Running repos.kuki failed with this error:
<paste the error here>
Fix it.
```

---

## Building an MCP Server

MCP (Model Context Protocol) lets you extend AI agents with custom tools written in Kukicha. You build a single binary that the agent can call.

### Example: stock price tool

Ask your agent:
```
Write a Kukicha MCP server with a tool called "get_price" that takes a stock ticker symbol
as a string and returns a hardcoded price as a string. Register it with mcp.Tool and serve it.
```

Agent output:
```kukicha
import "stdlib/mcp"

func getPrice(symbol string) string
    # Replace with a real API call in production
    if symbol equals "GOOG"
        return "GOOG: $180.00"
    if symbol equals "AAPL"
        return "AAPL: $220.00"
    return "{symbol}: price unavailable"

func main()
    server := mcp.NewServer()
    server |> mcp.Tool("get_price", "Get stock price by ticker symbol", getPrice)
    server |> mcp.Serve()
```

### Compile and register

```bash
kukicha build prices.kuki -o prices-server
./prices-server
```

Add to `~/.claude/config.json` (Claude Desktop):
```json
{
  "mcpServers": {
    "prices": {
      "command": "/absolute/path/to/prices-server"
    }
  }
}
```

Restart Claude Desktop. Claude now has a `get_price` tool it can call.

### Review checklist for MCP servers

- [ ] Does each tool function take only simple types (string, int, bool)?
- [ ] Does the tool return a string or simple value?
- [ ] Are errors handled so the server doesn't crash on bad input?
- [ ] Is the tool description (second argument to `mcp.Tool`) clear enough for the AI to know when to use it?

---

## Where Concepts Transfer

Everything you learn in Kukicha maps directly to Go and Python. You're not learning a dead end.

| Concept | Kukicha | Go | Python |
|---------|---------|-----|--------|
| Variable | `count := 42` | `count := 42` | `count = 42` |
| List | `list of int{1, 2}` | `[]int{1, 2}` | `[1, 2]` |
| Loop | `for item in items` | `for _, item := range items` | `for item in items:` |
| Error handling | `result onerr panic "msg"` | `if err != nil { panic("msg") }` | `try: ... except: raise` |
| Function | `func Add(a int, b int) int` | `func Add(a int, b int) int` | `def add(a: int, b: int) -> int:` |
| Null check | `if x equals empty` | `if x == nil` | `if x is None:` |
| Struct | `type User` | `type User struct { ... }` | `@dataclass\nclass User:` |
| Pointer | `reference User` | `*User` | *(implicit reference)* |
| Pipe/chain | `data \|> f() \|> g()` | `g(f(data))` | `g(f(data))` |

**The key insight:** Kukicha is Go with English keywords and indentation instead of symbols and braces. If you can read Kukicha, you can read Go. The logic is identical.

---

## Next Steps

- [Absolute Beginner Tutorial](absolute-beginner-tutorial.md) — learn Kukicha syntax from scratch
- [Data & AI Scripting](data-scripting-tutorial.md) — maps, CSV, LLM integration
- [Production Patterns](production-patterns-tutorial.md) — databases, validation, retry, auth
- [Stdlib Reference](../../stdlib/AGENTS.md) — all available packages

---

## Tips for Effective Agent Prompting

1. **Be specific about intent, not implementation**
   - ❌ "Add error handling"
   - ✅ "If the API call fails, print the error and exit"

2. **Name the stdlib packages you want used**
   - ✅ "Use stdlib/fetch for HTTP and stdlib/slice for filtering"

3. **Request testable code**
   - ✅ "Include a main() that demonstrates the function with example data"

4. **Iterate in small steps**
   - ❌ "Make this production-ready with logging, metrics, and retries"
   - ✅ "Add structured logging using stdlib/obs"

5. **Always validate before shipping**
   ```bash
   kukicha check file.kuki   # syntax check
   kukicha run file.kuki     # test it
   kukicha build file.kuki   # compile for distribution
   ```

6. **If the agent generates Go syntax by mistake**, remind it:
   ```
   You wrote Go syntax. Rewrite it in Kukicha:
   - Use "and", "or", "not" instead of &&, ||, !
   - Use "list of string" instead of []string
   - Use "equals" instead of ==
   - Use 4-space indentation instead of curly braces
   - Use "onerr" instead of "if err != nil"
   ```
