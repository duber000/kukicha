# Scripting Superpowers: LLM + Shell + Pipes

**Level:** Intermediate
**Time:** 10 minutes
**Prerequisite:** [Beginner Tutorial](beginner-tutorial.md)

Kukicha's pipe operator `|>` isn't just for formatting strings. Combined with the `shell` and `llm` packages, it turns into a force multiplier — connecting system commands to AI in a handful of lines.

This tutorial shows four practical scripts. None is longer than 30 lines.

---

## Prerequisites

You'll need an API key for at least one LLM provider:

```bash
# Option A: OpenAI
export OPENAI_API_KEY="sk-..."

# Option B: Anthropic
export ANTHROPIC_API_KEY="sk-ant-..."
```

```bash
mkdir llm-scripts && cd llm-scripts
go mod init llm-scripts
kukicha init
```

---

## Example 1: AI Commit Message Generator

You've staged changes with `git add`. Let this script write your commit message.

Create `commitmsg.kuki`:

```kukicha
import "stdlib/shell"
import "stdlib/llm"
import "stdlib/string"

function main()
    # Get the staged diff from git
    result := shell.New("git", "diff", "--staged") |> shell.Execute()
    diff := string(shell.GetOutput(result))

    if diff |> string.TrimSpace() equals ""
        print("No staged changes. Run 'git add' first.")
        return

    # Pipe the diff into an LLM with a focused prompt
    message := llm.New("gpt-4o-mini")
        |> llm.System("Write a concise git commit message for this diff. Use conventional commit format (feat:, fix:, refactor:, etc). One line, under 72 characters. No explanation, just the message.")
        |> llm.Ask(diff)
        onerr
            print("LLM error: {error}")
            return

    print("Suggested commit message:")
    print("  {message}")
```

Run it:

```bash
kukicha run commitmsg.kuki
# => Suggested commit message:
# =>   feat: add user authentication middleware
```

**What's happening:** Three tools — `shell`, `llm`, and `string` — chained together. The diff flows from git, through an LLM, out to your terminal. That's a script you'd actually use every day.

---

## Example 2: Explain This File

Point this at any file and get a plain-English explanation.

Create `explain.kuki`:

```kukicha
import "os"
import "stdlib/files"
import "stdlib/llm"

function main()
    # Get filename from command line args
    if len(os.Args) < 2
        print("Usage: kukicha run explain.kuki <filename>")
        return

    filename := os.Args[1]

    code := files.Read(filename) onerr
        print("Can't read '{filename}': {error}")
        return

    explanation := llm.New("gpt-4o-mini")
        |> llm.System("Explain this code to a junior developer. Be concise — bullet points, not paragraphs. Focus on what it does, not line-by-line narration.")
        |> llm.Ask(string(code))
        onerr
            print("LLM error: {error}")
            return

    print("=== {filename} ===\n")
    print(explanation)
```

Run it:

```bash
kukicha run explain.kuki ../my-project/main.go
# => === ../my-project/main.go ===
# =>
# => - Sets up an HTTP server on port 8080
# => - Defines two routes: /health for status checks, /api/users for user data
# => - Uses middleware for request logging and authentication
# => ...
```

---

## Example 3: Smart Log Analyzer

Tail your logs and ask an LLM to spot the problem.

Create `logcheck.kuki`:

```kukicha
import "stdlib/shell"
import "stdlib/llm"
import "stdlib/string"

function main()
    # Grab recent logs (swap the command for your log source)
    result := shell.New("journalctl", "--no-pager", "-n", "50", "--output", "short-iso")
        |> shell.Execute()

    logs := string(shell.GetOutput(result))

    if logs |> string.TrimSpace() equals ""
        print("No log entries found.")
        return

    analysis := llm.New("gpt-4o-mini")
        |> llm.System("You are a sysadmin. Analyze these logs. Summarize: (1) any errors or warnings, (2) likely root cause, (3) suggested fix. Be terse.")
        |> llm.Ask(logs)
        onerr
            print("LLM error: {error}")
            return

    print(analysis)
```

Swap `journalctl` for `kubectl logs`, `docker logs`, or `tail` — same pattern, different source.

---

## The Pattern

Every script above follows the same shape:

```
get data (shell/files) → |> transform with LLM → print or save
```

That's the force multiplier. Kukicha's pipes let you compose system tools and AI the same way Unix composes commands with `|`. But instead of `sed` and `awk`, your "filter" understands natural language.

---

## Example 4: Putting It Together — A Branch Reviewer

Let's combine everything into a single useful tool: a local code reviewer that examines your branch, asks an LLM for feedback, and saves the result.

Create `review.kuki`:

```kukicha
import "stdlib/shell"
import "stdlib/llm"
import "stdlib/files"
import "stdlib/string"
import "stdlib/datetime"

function main()
    # Get the diff between this branch and main
    result := shell.New("git", "diff", "main...HEAD") |> shell.Execute()
    diff := string(shell.GetOutput(result))

    if diff |> string.TrimSpace() equals ""
        print("No changes compared to main.")
        return

    # Get commit messages for context
    logResult := shell.New("git", "log", "main..HEAD", "--oneline") |> shell.Execute()
    commits := string(shell.GetOutput(logResult))

    # Build the review prompt
    prompt := "## Commits\n{commits}\n\n## Diff\n{diff}"

    review := llm.New("gpt-4o-mini")
        |> llm.System("You are a senior code reviewer. Review this PR diff. Format as:\n## Summary\n(1-2 sentences)\n## Issues\n(bulleted list, or 'None found')\n## Suggestions\n(bulleted list of improvements)")
        |> llm.Ask(prompt)
        onerr
            print("LLM error: {error}")
            return

    # Save the review
    timestamp := datetime.Now() |> datetime.Format("date")
    review |> files.Write("review-{timestamp}.md") onerr
        print("Couldn't save file: {error}")

    print(review)
    print("\nSaved to review-{timestamp}.md")
```

Run it from any feature branch:

```bash
kukicha run review.kuki
# => ## Summary
# => Adds rate limiting middleware to the API server with configurable thresholds.
# =>
# => ## Issues
# => - The rate limit counter isn't thread-safe; concurrent requests could bypass limits
# =>
# => ## Suggestions
# => - Use sync.Mutex or atomic operations for the counter
# => - Add rate limit headers (X-RateLimit-Remaining) to responses
# => - Consider using a sliding window instead of fixed intervals
# =>
# => Saved to review-2026-02-13.md
```

In ~30 lines: a local PR reviewer that reads your git history, feeds it to an LLM, and saves a formatted review. No SaaS subscription, no browser tabs.

---

## Using Anthropic Instead

All examples above use OpenAI. Swap to Anthropic by changing the builder:

```kukicha
# Instead of:
message := llm.New("gpt-4o-mini")
    |> llm.System("...")
    |> llm.Ask("...")
    onerr ...

# Use:
message := llm.NewMessages("claude-sonnet-4-5-20250929")
    |> llm.MSystem("...")
    |> llm.MMaxTokens(1000)
    |> llm.MAsk("...")
    onerr ...
```

Same pipes, same data flow, different model.

---

## What You've Learned

| Concept | What It Does |
|---------|--------------|
| **`stdlib/shell`** | Run system commands safely (no shell injection) |
| **`stdlib/llm`** | Call LLM APIs with a pipe-friendly builder pattern |
| **`stdlib/files`** | Read and write files in pipelines |
| **Pipe Composition** | Chain shell, LLM, and file I/O in a single readable flow |
| **`onerr` blocks** | Handle errors at the end of a pipeline |

---

## Ideas to Build

- **Changelog generator** — `git log` between tags → LLM → formatted CHANGELOG.md
- **README drafter** — read source files → LLM → draft README
- **SQL explainer** — `files.Read("query.sql")` → LLM → plain-English explanation
- **Config auditor** — read YAML/JSON config → LLM → security review
- **Test writer** — read a function → LLM → generate test cases

The pattern is always the same: get data, pipe it through intelligence, do something with the result.

---

### Tutorial Path

| # | Tutorial | What You'll Learn |
|---|----------|-------------------|
| 1 | **[Beginner Tutorial](beginner-tutorial.md)** | Variables, functions, strings, decisions, lists, loops, pipes |
| 2 | **[CLI Explorer](cli-explorer-tutorial.md)** | Custom types, methods, API data, error handling |
| 3 | **[Link Shortener](web-app-tutorial.md)** | HTTP servers, JSON, redirects |
| 4 | **[Production Patterns](production-patterns-tutorial.md)** | Databases, concurrency, Go conventions |
|   | ✅ **LLM Scripting** *(you are here)* | Shell + LLM + pipes |
