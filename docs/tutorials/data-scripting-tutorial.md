# Tutorial 2: Data & AI Scripting

**Level:** Beginner/Intermediate
**Time:** 15 minutes
**Prerequisite:** [Beginner Tutorial](beginner-tutorial.md)

In the first tutorial, you learned the core syntax: variables, lists, loops, pipes, shell commands, and `onerr`. This tutorial builds on that foundation for data-heavy scripting workflows.

In this tutorial, you will learn:
1.  **Variadic Functions**: Functions that take any number of items (`many`).
2.  **Data Cleaning Pipelines**: Practical text-cleaning flows for messy inputs.
3.  **Maps**: How to store Key-Value data (like a dictionary).
4.  **Parsing**: How to turn raw text (like CSVs) into structured data.
5.  **Error Context**: How to wrap lower-level errors with useful messages.
6.  **AI Scripting**: How to combine shell output + LLM calls to automate tasks.

This is the "glue" code that makes Kukicha a powerful scripting language.

---

## Part 1: Many Items (Variadic Functions)

In the beginner tutorial, every function you wrote took a fixed number of inputs — `Greet(name string)` always takes exactly one string. But sometimes you don't know ahead of time how many values someone will pass in. Think about `print()`: you've been calling it with one argument, two arguments, whatever you need. That's a **variadic** function — one that accepts any number of arguments.

Kukicha uses the `many` keyword to declare a variadic parameter. Let's see how it works by building a `Sum` function that adds up however many numbers you give it.

Create `sum.kuki`:

```kukicha
function Sum(many numbers int) int
    total := 0
    for n in numbers
        total = total + n
    return total

function main()
    print(Sum(1, 2, 3))       # Prints: 6
    print(Sum(10, 20))        # Prints: 30
    print(Sum())              # Prints: 0
```

Walk through what's happening:

1. `many numbers int` in the function signature means "this function takes any number of `int` arguments." The caller can pass one, five, or zero — all valid.
2. Inside `Sum`, `numbers` behaves exactly like a `list of int`. You loop over it, index into it, check its length — everything you learned about lists in Tutorial 1 works here.
3. The difference is on the **caller's side**: instead of building a list and passing it in, you just write the values directly: `Sum(1, 2, 3)`.

### Spreading a list into a variadic call

What if you already have a list and want to pass its elements as the variadic arguments? Use `many` at the call site:

```kukicha
    values := list of int{4, 5, 6}
    print(Sum(many values))   # Prints: 15
```

Without `many` here, the compiler would think you're passing one argument (a whole list) instead of three individual ints. The `many` keyword "spreads" the list out into separate arguments.

**Key points:**
- Declare a variadic parameter with `many` before the name: `function F(many items T)`.
- Inside the function, the parameter is just a `list of T` — nothing special.
- Callers pass values directly: `F(a, b, c)` — no list construction needed.
- To pass an existing list, spread it with `many`: `F(many myList)`.

---

## Part 2: Data Cleaning Pipelines

You already know pipes from Tutorial 1. Here we'll use them for realistic cleaning tasks.

### Cleaning Delimited Input

Create `strings.kuki`:

```kukicha
import "stdlib/env"
import "stdlib/string"

function main()
    raw := "analytics,  batch jobs,ops , ai "
    tags := raw |> env.SplitAndTrim(",")
    print(tags |> string.Join(" | "))  # analytics | batch jobs | ops | ai
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

## Part 4: Parsing Data (CSV)

Real data often comes in messy formats like CSV (Comma Separated Values). Let's parse some user data using `stdlib/parse`.

Create `parser.kuki`:

```kukicha
import "stdlib/parse"
import "stdlib/json"  # We'll use this to pretty-print

function main()
    # Raw CSV data (simulating reading a file)
    csvData := "Name,Role,Score\nAlice,Admin,95\nBob,User,82\nCharlie,User,45"

    # Parse into a list of maps
    # Each row becomes a map: {"Name": "Alice", "Role": "Admin", ...}
    users := csvData |> parse.CsvWithHeader() onerr
        print("Failed to parse CSV: {error}")
        return

    # Print the first user's name and role.
    # Note: map key lookups use double quotes, which can't be nested inside a "..." string.
    # Assign to a variable first, then interpolate.
    firstUser := users[0]
    firstName := firstUser["Name"]
    firstRole := firstUser["Role"]
    print("First user: {firstName} is a {firstRole}")

    # Print everything as JSON to see the structure
    users |> json.MarshalPretty() |> print
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

## Part 5: Wrapping Errors with `stdlib/errors`

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

## Part 6: AI Scripting (The Fun Part)

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
        |> llm.Ask(diff) onerr
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
This composes a simple data pipeline:
`Data (diff)` -> `LLM` -> `Result`.

You can use this pattern for anything:
-   **Summarize logs**: Pipe `tail loop.log` into `llm.Ask("Find errors")`.
-   **Explain code**: Pipe `files.Read("main.kuki")` into `llm.Ask("Explain this")`.
-   **Translate**: Pipe text into `llm.Ask("Translate to Spanish")`.

---

## Part 7: Putting It All Together

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
    rows := csvData |> parse.CsvWithHeader() onerr
        print("Failed to parse CSV: {error}")
        return

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
