# Tutorial 2: Data & AI Scripting

**Level:** Beginner/Intermediate
**Time:** 15 minutes
**Prerequisite:** [Beginner Tutorial](beginner-tutorial.md)

In the first tutorial, you learned about lists—ordered collections of items. But often, data isn't just a list; it's structured. Detailed records, configuration settings, and data from files often come as "Key-Value" pairs.

In this tutorial, you will learn:
1.  **Variadic Functions**: Functions that take any number of items (`many`).
2.  **String Superpowers**: Chaining operations with the Pipe Operator (`|>`).
3.  **Maps**: How to store Key-Value data (like a dictionary).
4.  **Parsing**: How to turn raw text (like CSVs) into structured data.
5.  **Shell Commands**: How to run system tools from Kukicha.
6.  **AI Scripting**: How to pipe data into an LLM to automate tasks.

This is the "glue" code that makes Kukicha a powerful scripting language.

---

## Part 1: Many Items (Variadic Functions)

Sometimes you want a function that can take any number of arguments, like `print()`. In Kukicha, use the `many` keyword. This is perfect for data aggregation tools.

Create `sum.kuki`:

```kukicha
# 'many numbers int' means 'numbers' is a list of int
function Sum(many numbers int) int
    total := 0
    for n in numbers
        total = total + n
    return total

function main()
    print(Sum(1, 2, 3))       # Prints: 6
    print(Sum(10, 20))        # Prints: 30
    print(Sum())              # Prints: 0

    # Spread an existing list with 'many' at the call site
    values := list of int{4, 5, 6}
    print(Sum(many values))   # Prints: 15
```

**Key points:**
- `many` goes before the parameter name in the declaration.
- Inside the function, the parameter acts like a standard list.
- To spread an existing list into a variadic call, use `many` at the call site too: `Sum(many values)`.

---

## Part 2: String Superpowers & Pipes

Kukicha includes a **string petiole** (package) with powerful functions for working with text. When combined with the **pipe operator** (`|>`), it makes data cleaning a breeze.

### The Pipe Operator (`|>`)

The pipe operator lets you pass the result of one function directly into the next. It reads like a natural flow: "take X, then do Y, then do Z."

Instead of nesting functions:
`lower := string.ToLower(string.TrimSpace(text))`

You write a pipeline:
`lower := text |> string.TrimSpace() |> string.ToLower()`

### Common String Operations

Create `strings.kuki`:

```kukicha
import "stdlib/string"

function main()
    text := "  HELLO world  "

    # 1. Cleaning up
    clean := text |> string.TrimSpace() |> string.ToLower() |> string.Title()
    print("Clean: [{clean}]") # Prints: Clean: [Hello World]

    # 2. Searching
    if clean |> string.Contains("Hello")
        print("Found greeting!")

    # 3. Splitting and Joining
    parts := "apple,banana,cherry" |> string.Split(",")
    joined := parts |> string.Join(" | ")
    print(joined) # Prints: apple | banana | cherry
```

### Advanced: Splitting AND Trimming

When you split strings with messy spacing, you often need to trim each piece. The `env` package has a utility for this:

```kukicha
import "stdlib/env"

function main()
    messy := "one,  two , three "
    cleanList := messy |> env.SplitAndTrim(",")
    # Result: ["one", "two", "three"] (no spaces!)
```

---

## Part 3: Maps - Key-Value Pairs

A **Map** is like a dictionary or a phone book. You look up a "Key" (like a name) to find a "Value" (like a phone number).

### Creating a Map

Create `maps.kuki`:

```kukicha
function main()
    # Create a map where Keys are strings and Values are strings
    capitals := map of string to string{
        "France": "Paris",
        "Japan": "Tokyo",
        "Egypt": "Cairo",
    }

    print(capitals["Japan"])  # Prints: Tokyo
```

**Try it:**
```bash
kukicha run maps.kuki
```

### Adding and Changing Items

```kukicha
    # Add a new one
    capitals["Brazil"] = "Brasilia"

    # Change an existing one
    capitals["Japan"] = "Kyoto"  # Wait, that's the old capital!
    capitals["Japan"] = "Tokyo"  # Fixed.

    print(capitals)
```

### Checking for Existence

What if you look up a key that doesn't exist?

```kukicha
    city := capitals["Mars"]
    if city equals ""
        print("Capital of Mars not found.")
```

For `map of string to string`, a missing key returns `""` (empty string). For `map of string to int`, it returns `0`.

---

## Part 2: Parsing Data (CSV)

Real data often comes in messy formats like CSV (Comma Separated Values). Let's parse some user data using `stdib/parse`.

Create `parser.kuki`:

```kukicha
import "stdlib/parse"
import "stdlib/json"  # We'll use this to pretty-print

function main()
    # Raw CSV data (simulating reading a file)
    csvData := "Name,Role,Score\nAlice,Admin,95\nBob,User,82\nCharlie,User,45"

    # Parse into a list of maps
    # Each row becomes a map: {"Name": "Alice", "Role": "Admin", ...}
    users := csvData |> parse.CsvWithHeader() onerr empty

    if users equals empty
        print("Failed to parse CSV")
        return

    # Print the first user's Role
    firstUser := users[0]
    print("First user: {firstUser["Name"]} is a {firstUser["Role"]}")

    # Print everything as JSON to see the structure
    users |> json.MarshalIndent("", "  ") |> print
```

**Try it:**
```bash
kukicha run parser.kuki
```

**What happened?**
1.  `parse.CsvWithHeader()` took the text.
2.  It used the first line (`Name,Role,Score`) as keys.
3.  It turned each row into a `map of string to string`.
4.  Result: A `list of map of string to string`.

This is incredibly powerful for processing spreadsheets or data exports!

---

## Part 3: Shell Commands

Kukicha can run any command you can type in your terminal. This is great for automating workflows.

Create `git_check.kuki`:

```kukicha
import "stdlib/shell"
import "stdlib/string"

function main()
    # Run 'git status' and capture the output
    status := shell.Output("git", "status", "--short") onerr ""

    if status equals ""
        print("Directory is clean (or not a git repo).")
        return

    print("Changed files:")
    print(status)
```

**Try it** (in a git folder):
```bash
kukicha run git_check.kuki
```

---

## Part 3b: Wrapping Errors with `stdlib/errors`

When a shell command fails, you often want to add context to the error before surfacing it to the caller. The `stdlib/errors` package makes this clean:

```kukicha
import "stdlib/shell"
import "stdlib/errors"

function GitDiff() (string, error)
    diff := shell.Output("git", "diff", "--staged") onerr return "", errors.Wrap(error, "git diff failed")
    return diff, empty
```

`errors.Wrap(err, "message")` produces `"message: <original error>"` — the same pattern as Go's `fmt.Errorf("message: %w", err)` but without the format string boilerplate.

You can also check whether a specific error occurred deep in a call stack:

```kukicha
import "stdlib/errors"
import "io"

data := readSomething() onerr
    if errors.Is(error, io.EOF)
        return "", empty  # EOF is normal — treat as empty
    return "", errors.Wrap(error, "read failed")
```

---

## Part 4: AI Scripting (The Fun Part)

Now, let's combine **Shell** commands with **AI**. We'll build a script that looks at your code changes and writes a commit message for you.

**Prerequisite:** You need an API key (OpenAI or Anthropic).
```bash
export OPENAI_API_KEY="sk-..."
```

Create `autocommit.kuki`:

```kukicha
import "stdlib/shell"
import "stdlib/llm"
import "stdlib/string"

function main()
    # 1. Get the staged changes
    diff := shell.Output("git", "diff", "--staged") onerr ""

    if diff |> string.TrimSpace() equals ""
        print("No staged changes. Run 'git add' first.")
        return

    # 2. Pipe the diff to the LLM
    print("Analyzing changes...")
    
    message := llm.New("gpt-5-nano")
        |> llm.System("You are a helpful assistant. Write a concise git commit message for this diff. Format: 'feat: description' or 'fix: description'. One line only.")
        |> llm.Ask(diff)
        onerr
            print("LLM Error: {error}")
            return

    # 3. Print the result
    print("\nSuggested Commit Message:")
    print(message)
```

**Run it:**
1.  Make a change to a file.
2.  `git add file`
3.  `kukicha run autocommit.kuki`

### How it works
This shows the power of the **Pipe Operator (`|>`)**:
`Data (diff)` -> `LLM` -> `Result`.

You can use this pattern for anything:
-   **Summarize logs**: Pipe `tail loop.log` into `llm.Ask("Find errors")`.
-   **Explain code**: Pipe `files.Read("main.kuki")` into `llm.Ask("Explain this")`.
-   **Translate**: Pipe text into `llm.Ask("Translate to Spanish")`.

---

## Part 5: Putting It All Together

Let's build a "Data Cleaner". It will:
1.  Read a CSV of names.
2.  "Clean" them (fix capitalization).
3.  Output the clean list.

Create `cleaner.kuki`:

```kukicha
import "stdlib/parse"
import "stdlib/string"

function main()
    # Messy data
    csvData := "name,id\nalice smith,1\nBOB JONES,2\ncharlie brown,3"

    print("--- Raw Data ---")
    print(csvData)

    # Parse
    rows := csvData |> parse.CsvWithHeader() onerr empty

    print("\n--- Cleaning ---")
    for row in rows
        # Get the name
        name := row["name"]
        
        # Clean it: Trim spaces, Title Case
        cleanName := name |> string.TrimSpace() |> string.Title()
        
        # Update the map
        row["name"] = cleanName
        
        print("Fixed: {name} -> {cleanName}")

    print("\n--- Done ---")
```

**Try it:**
```bash
kukicha run cleaner.kuki
```

---

## What's Next?

You now have the tools to fetch data, organize it, and even use AI to understand it.

Next, we'll build a full interactive application that fetches data from the internet:

👉 **[Tutorial 3: CLI Explorer](cli-explorer-tutorial.md)**

---
**Summary of New/Updated Concepts:**

| Concept | Syntax | Example |
| :--- | :--- | :--- |
| **Map** | `map of K to V` | `m := map of string to int{"A": 1}` |
| **Access** | `m[key]` | `val := m["A"]` |
| **Parse CSV** | `parse.CsvWithHeader()` | Turns string keys into field names |
| **Shell** | `shell.Output()` | API to run `git`, `ls`, etc. |
| **LLM** | `llm.New() \|> ...` | Easy AI integration |
| **Variadic spread** | `Sum(many values)` | Spread a list into a variadic call |
| **Error wrapping** | `errors.Wrap(err, "msg")` | Add context to errors |
