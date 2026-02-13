# Build a GitHub Repo Explorer with Kukicha

**Level:** Intermediate
**Time:** 15-18 minutes
**Prerequisite:** [Beginner Tutorial](beginner-tutorial.md)

Welcome back! In the beginner tutorial, you learned about variables, functions, strings, decisions, lists, and loops. Now we're going to build something genuinely useful: a **GitHub Repo Explorer** that fetches real data from the internet and lets you browse it interactively.

## What You'll Learn

In this tutorial, you'll discover how to:
- Create **custom types** to organize related data
- Write **methods** that belong to types
- Use `reference` to modify data in place
- Handle errors gracefully with **`onerr`**
- Use the **Pipe Operator (`|>`)** to build data transformation pipelines
- **Fetch data** from a web API and parse **JSON**
- Build a simple **command loop** for a console app

Let's build something useful!

---

## What We're Building

A CLI tool that talks to GitHub's public API and lets you:
- **Fetch** repositories for any GitHub user or organization
- **Display** them in a clean, formatted list
- **Filter** by programming language
- **Search** by name or description
- **Save favorites** across multiple users

Here's what it will look like when running:

```
=== GitHub Repo Explorer ===
Fetching repos for 'golang'...
Found 30 repos!

Commands: list, filter, search, fav, favs, fetch, help, quit
> list
  1. go           ⭐ 125000  Go   The Go programming language
  2. tools        ⭐  15200  Go   Go Tools
  3. protobuf     ⭐  10500  Go   Go support for Protocol Buffers
...

> filter python
Showing 2 repos matching 'python'

> fav 1
Saved to favorites: go

> fetch torvalds
Fetching repos for 'torvalds'...
Found 8 repos!

> quit
Goodbye!
```

---

## Step 0: Project Setup

If you haven't already, set up your project:

```bash
mkdir repo-explorer && cd repo-explorer
go mod init repo-explorer
kukicha init    # Extracts stdlib for imports like "stdlib/fetch"
```

---

## Step 1: Creating a Repo Type

> **📝 Reminder:** This tutorial builds on the beginner tutorial. Here are the key concepts you'll need:
> - **`:=`** creates a new variable, **`=`** updates an existing one
> - **String interpolation:** Use `{variable}` inside strings to insert values
> - **`print()`** outputs to the console
> - **Functions** (starting with `function`) take parameters and can return values
> - **Comments** start with `#`
>
> If you need a refresher, [revisit the beginner tutorial](beginner-tutorial.md)!

In the beginner tutorial, you learned about basic types like `string`, `int`, and `bool`. Now let's create our own type to represent a GitHub repository.

Create a file called `explorer.kuki`:

```kukicha
import "stdlib/fetch"
import "stdlib/json"
import "stdlib/string"
import "stdlib/slice"

# A Repo represents a GitHub repository
# The json:"..." tags tell the JSON parser which API field maps to which field
type Repo
    Name string json:"name"
    Description string json:"description"
    Stars int json:"stargazers_count"
    Language string json:"language"
    URL string json:"html_url"
```

**What's new here?**

We're defining a custom **type** called `Repo` — a blueprint for GitHub repository data. Each repo has:
- `Name` — The repository name
- `Description` — What it's about
- `Stars` — How many people have starred it
- `Language` — Primary programming language
- `URL` — Link to view it on GitHub

The **`json:"..."`** part after each field is a **struct tag**. GitHub's API returns JSON with fields like `"stargazers_count"`. The tag tells Kukicha: "when you see `stargazers_count` in the JSON data, put it in the `Stars` field." This lets you use clean names in your code while still matching the API's naming conventions.

---

## Step 2: Fetching Repos from GitHub

This is where Kukicha shines. Let's fetch real data from GitHub's public API:

```kukicha
# FetchRepos gets repositories for a GitHub user or organization
function FetchRepos(username string) list of Repo
    url := "https://api.github.com/users/{username}/repos?per_page=30&sort=stars"

    repos := empty list of Repo

    fetch.Get(url)
        |> fetch.CheckStatus()
        |> fetch.Bytes()
        |> json.Unmarshal(reference repos)
        onerr
            print("Failed to fetch repos for '{username}': {error}")
            return empty list of Repo

    return repos
```

**What's happening here?**

This is a **data pipeline** using the pipe operator `|>`. Read it left to right:

1. `fetch.Get(url)` — Make an HTTP request to GitHub's API
2. `|> fetch.CheckStatus()` — Verify we got a success response (not a 404)
3. `|> fetch.Bytes()` — Read the response body as raw bytes
4. `|> json.Unmarshal(reference repos)` — Parse the JSON bytes into our `list of Repo`

Each step's output flows into the next step's input. Without pipes, you'd need to store each intermediate result in a temporary variable and check for errors at every step — roughly 12 lines instead of 4.

**`onerr` in action:** If *any* step in the pipeline fails (network error, bad status code, invalid JSON), execution jumps to the `onerr` block. One clause handles errors from four operations.

**`reference repos`** is needed because `json.Unmarshal` needs to *modify* the `repos` variable (fill it with data). We'll explore `reference` more in Step 4.

### Let's Try It

Add a `main` function to test our fetcher:

```kukicha
function main()
    print("Fetching repos for 'golang'...")
    repos := FetchRepos("golang")
    print("Found {len(repos)} repos!\n")

    for repo in repos[:5]
        print("- {repo.Name}: {repo.Stars} stars")
```

Run it with `kukicha run explorer.kuki`:

```
Fetching repos for 'golang'...
Found 30 repos!

- go: 125000 stars
- tools: 15200 stars
- protobuf: 10500 stars
- net: 9800 stars
- mock: 9600 stars
```

You just fetched real data from the internet in about 10 lines of code.

---

## Step 3: Writing Methods

A **method** is a function that belongs to a type. Methods let you define what actions a type can perform.

In Kukicha, we use the `on` keyword to attach a method to a type:

```kukicha
# Summary returns a formatted one-line display of the repo
function Summary on repo Repo(index int) string
    lang := repo.Language
    if lang equals ""
        lang = "n/a"
    return "  {index}. {repo.Name}  ⭐ {repo.Stars}  {lang}  {repo.Description}"
```

**Reading this method:**
- `function Summary` — We're creating a method called "Summary"
- `on repo Repo` — This method works on a `Repo` (syntax: receiver name first, then the type). Inside the method, we call it `repo`
- `(index int)` — The method also takes an index number for display numbering
- `string` — The method returns a string

### Filtering with Pipes

Now let's write functions to filter repos. This is where pipes become essential:

```kukicha
# FilterByLanguage returns repos matching a language (case-insensitive)
function FilterByLanguage(repos list of Repo, language string) list of Repo
    return repos |> slice.Filter(function(r Repo) bool
        return r.Language |> string.ToLower() |> string.Contains(language |> string.ToLower())
    )
```

**💡 Pipe Pipelines:** Notice how `r.Language |> string.ToLower() |> string.Contains(...)` reads like a sentence: "take the language, make it lowercase, check if it contains our search term." This is cleaner than nesting function calls like `string.Contains(string.ToLower(r.Language), ...)`.

### Let's Try It

Update `main` to display and filter repos:

```kukicha
function main()
    repos := FetchRepos("golang")
    print("Found {len(repos)} repos!\n")

    # Display first 5
    for i, repo in repos[:5]
        print(repo.Summary(i + 1))

    # Filter by language
    print("\n--- Repos written in Go ---")
    goRepos := FilterByLanguage(repos, "go")
    for i, repo in goRepos[:3]
        print(repo.Summary(i + 1))
```

Run it:

```
Found 30 repos!

  1. go  ⭐ 125000  Go  The Go programming language
  2. tools  ⭐ 15200  Go  Go Tools
  3. protobuf  ⭐ 10500  Go  Go support for Protocol Buffers
  4. net  ⭐ 9800  Go  Go supplementary network libraries
  5. mock  ⭐ 9600  Go  GoMock is a mocking framework

--- Repos written in Go ---
  1. go  ⭐ 125000  Go  The Go programming language
  2. tools  ⭐ 15200  Go  Go Tools
  3. protobuf  ⭐ 10500  Go  Go support for Protocol Buffers
```

---

## Step 4: Building the Explorer

Now let's create an `Explorer` type that tracks state — which repos we've fetched and which we've favorited. This is where `reference` becomes important.

```kukicha
# Explorer manages our browsing session
type Explorer
    repos list of Repo
    favorites list of Repo
    username string
```

Now add methods for it:

```kukicha
# Fetch loads repos for a GitHub user
function Fetch on ex reference Explorer(username string)
    ex.username = username
    print("Fetching repos for '{username}'...")
    ex.repos = FetchRepos(username)
    print("Found {len(ex.repos)} repos!")

# ShowList displays all loaded repos
function ShowList on ex Explorer
    if len(ex.repos) equals 0
        print("\nNo repos loaded. Use 'fetch <username>' first.\n")
        return

    print("\n=== Repos for {ex.username} ===")
    for i, repo in ex.repos
        print(repo.Summary(i + 1))
    print("")
```

**Note on receiver naming:** We use `ex` as the receiver variable name (short for "explorer"). Keep receiver names short and consistent.

### A Method That Changes Things

What if we want to save a repo as a favorite? We need a method that can **modify** the explorer. For that, we use `reference`:

```kukicha
# AddFavorite saves a repo to favorites by its display number
function AddFavorite on ex reference Explorer(index int)
    if index < 1 or index > len(ex.repos)
        print("Invalid number. Use 1-{len(ex.repos)}")
        return

    repo := ex.repos[index - 1]

    # Check if already favorited
    for fav in ex.favorites
        if fav.Name equals repo.Name
            print("'{repo.Name}' is already in your favorites")
            return

    ex.favorites = append(ex.favorites, repo)
    print("Saved to favorites: {repo.Name}")

# ShowFavorites displays saved repos
function ShowFavorites on ex Explorer
    if len(ex.favorites) equals 0
        print("\nNo favorites yet! Use 'fav <number>' to save one.\n")
        return

    print("\n=== Your Favorites ===")
    for i, repo in ex.favorites
        print(repo.Summary(i + 1))
    print("")
```

**Why `reference`?**

Without `reference`, the method would get a **copy** of the explorer. Any changes would only affect the copy, not the original. Using `reference` means we're working with the **actual** explorer, so our changes stick.

Think of it like a shared document: without `reference`, you'd get a photocopy — scribble on it all day, but the original won't change. With `reference`, you're editing the original document itself.

`Fetch` and `AddFavorite` use `reference Explorer` because they modify the explorer. `ShowList` and `ShowFavorites` use plain `Explorer` (no reference) because they only read data — this signals to readers "this method won't change anything."

---

## Step 5: The Complete Program

Now let's put it all together into a working application!

> **Note:** The final program imports `bufio` and `os` only for reading console input. Everything else uses Kukicha's standard library.

Replace the contents of `explorer.kuki` with the complete program:

```kukicha
import "bufio"
import "os"
import "stdlib/fetch"
import "stdlib/json"
import "stdlib/string"
import "stdlib/slice"
import "strconv"

# --- Type Definitions ---

type Repo
    Name string json:"name"
    Description string json:"description"
    Stars int json:"stargazers_count"
    Language string json:"language"
    URL string json:"html_url"

type Explorer
    repos list of Repo
    favorites list of Repo
    username string

# --- Repo Methods ---

function Summary on repo Repo(index int) string
    lang := repo.Language
    if lang equals ""
        lang = "n/a"
    return "  {index}. {repo.Name}  ⭐ {repo.Stars}  {lang}  {repo.Description}"

# --- Data Fetching ---

function FetchRepos(username string) list of Repo
    url := "https://api.github.com/users/{username}/repos?per_page=30&sort=stars"
    repos := empty list of Repo
    fetch.Get(url)
        |> fetch.CheckStatus()
        |> fetch.Bytes()
        |> json.Unmarshal(reference repos)
        onerr
            print("Failed to fetch repos for '{username}': {error}")
            return empty list of Repo
    return repos

# --- Filter Functions ---

function FilterByLanguage(repos list of Repo, language string) list of Repo
    return repos |> slice.Filter(function(r Repo) bool
        return r.Language |> string.ToLower() |> string.Contains(language |> string.ToLower())
    )

# --- Explorer Methods ---

function Fetch on ex reference Explorer(username string)
    ex.username = username
    print("Fetching repos for '{username}'...")
    ex.repos = FetchRepos(username)
    print("Found {len(ex.repos)} repos!")

function ShowList on ex Explorer
    if len(ex.repos) equals 0
        print("\nNo repos loaded. Use 'fetch <username>' first.\n")
        return

    print("\n=== Repos for {ex.username} ===")
    for i, repo in ex.repos
        print(repo.Summary(i + 1))
    print("")

function AddFavorite on ex reference Explorer(index int)
    if index < 1 or index > len(ex.repos)
        print("Invalid number. Use 1-{len(ex.repos)}")
        return

    repo := ex.repos[index - 1]
    for fav in ex.favorites
        if fav.Name equals repo.Name
            print("'{repo.Name}' is already in your favorites")
            return

    ex.favorites = append(ex.favorites, repo)
    print("Saved to favorites: {repo.Name}")

function ShowFavorites on ex Explorer
    if len(ex.favorites) equals 0
        print("\nNo favorites yet! Use 'fav <number>' to save one.\n")
        return

    print("\n=== Your Favorites ===")
    for i, repo in ex.favorites
        print(repo.Summary(i + 1))
    print("")

function PrintHelp()
    print("Commands:")
    print("  fetch <user>  - Fetch repos for a GitHub user/org")
    print("  list          - Show all fetched repos")
    print("  filter <lang> - Filter repos by programming language")
    print("  search <text> - Search repos by name or description")
    print("  fav <number>  - Save a repo to your favorites")
    print("  favs          - Show your saved favorites")
    print("  help          - Show this help")
    print("  quit          - Exit the explorer")

# --- Main Program ---

function main()
    ex := Explorer
        repos: empty list of Repo
        favorites: empty list of Repo
        username: ""

    print("=== GitHub Repo Explorer ===")
    print("Type 'help' for commands\n")

    # Start with some repos to explore
    ex.Fetch("golang")
    print("")

    # Create a reader for user input
    reader := bufio.NewReader(os.Stdin)

    # Main loop
    for
        print("> ")

        # Read user input — default to empty string on error
        input := reader.ReadString('\n') onerr ""
        input = input |> string.TrimSpace()

        if input equals ""
            continue

        # SplitN(" ", 2) splits into at most 2 parts
        parts := input |> string.SplitN(" ", 2)
        command := parts[0] |> string.ToLower()

        if command equals "quit" or command equals "exit" or command equals "q"
            print("Goodbye!")
            break

        else if command equals "help" or command equals "?"
            PrintHelp()

        else if command equals "list" or command equals "ls"
            ex.ShowList()

        else if command equals "fetch"
            if len(parts) < 2
                print("Usage: fetch <username>")
                continue
            ex.Fetch(parts[1])

        else if command equals "filter"
            if len(parts) < 2
                print("Usage: filter <language>")
                continue
            filtered := FilterByLanguage(ex.repos, parts[1])
            print("\nShowing {len(filtered)} repos matching '{parts[1]}'")
            for i, repo in filtered
                print(repo.Summary(i + 1))
            print("")

        else if command equals "search"
            if len(parts) < 2
                print("Usage: search <text>")
                continue
            term := parts[1] |> string.ToLower()
            results := ex.repos |> slice.Filter(function(r Repo) bool
                name := r.Name |> string.ToLower()
                desc := r.Description |> string.ToLower()
                return name |> string.Contains(term) or desc |> string.Contains(term)
            )
            print("\nFound {len(results)} repos matching '{parts[1]}'")
            for i, repo in results
                print(repo.Summary(i + 1))
            print("")

        else if command equals "fav"
            if len(parts) < 2
                print("Usage: fav <number>")
                continue
            # Parse the number — print a message and skip if it's not valid
            id, idErr := strconv.Atoi(parts[1])
            if idErr not equals empty
                print("Invalid number: {parts[1]}")
                continue
            ex.AddFavorite(id)

        else if command equals "favs" or command equals "favorites"
            ex.ShowFavorites()

        else
            print("Unknown command: {command}")
            print("Type 'help' for available commands")
```

---

## Step 6: Running Your Explorer

Build and run your explorer:

```bash
kukicha run explorer.kuki
```

**Try these commands:**

```
=== GitHub Repo Explorer ===
Type 'help' for commands

Fetching repos for 'golang'...
Found 30 repos!

> list

=== Repos for golang ===
  1. go  ⭐ 125000  Go  The Go programming language
  2. tools  ⭐ 15200  Go  Go Tools
  3. protobuf  ⭐ 10500  Go  Go support for Protocol Buffers
...

> filter python
Showing 2 repos matching 'python'
  1. example  ⭐ 200  Python  Example Python bindings
...

> fav 1
Saved to favorites: go

> fetch torvalds
Fetching repos for 'torvalds'...
Found 8 repos!

> list

=== Repos for torvalds ===
  1. linux  ⭐ 185000  C  Linux kernel source tree
  2. subsurface-for-dirk  ⭐ 2100  C++  Divelog program
...

> fav 1
Saved to favorites: linux

> favs

=== Your Favorites ===
  1. go  ⭐ 125000  Go  The Go programming language
  2. linux  ⭐ 185000  C  Linux kernel source tree

> quit
Goodbye!
```

---

## Understanding the New Concepts

The final program introduced several concepts that deserve a closer look. Let's walk through them.

### JSON Tags — Mapping API Data to Types

```kukicha
type Repo
    Stars int json:"stargazers_count"
```

GitHub's API returns `"stargazers_count"` in its JSON response. The tag `json:"stargazers_count"` tells the JSON parser: "when you see `stargazers_count` in the data, put it in the `Stars` field." This lets you use clean names in your code while mapping to the API's conventions.

### Pipe Pipelines — Real Data Transformation

```kukicha
fetch.Get(url) |> fetch.CheckStatus() |> fetch.Bytes() |> json.Unmarshal(reference repos) onerr ...
```

This four-step pipeline is the heart of the program. Each `|>` passes the result of one operation to the next. Without pipes, you'd write:

```kukicha
resp, err1 := fetch.Get(url)
# check err1...
resp2, err2 := fetch.CheckStatus(resp)
# check err2...
bytes, err3 := fetch.Bytes(resp2)
# check err3...
err4 := json.Unmarshal(bytes, reference repos)
# check err4...
```

With pipes + `onerr`, four operations and their error handling compress into a readable pipeline.

### Bare `for` — The Infinite Loop

```kukicha
for
    # ... read input and process commands ...
```

A `for` with no condition runs forever. This is the standard pattern for programs that wait for user input — the loop keeps running until something inside calls `break`. You saw in the beginner tutorial that `for condition` runs while the condition is true; a bare `for` is just the extreme case where the condition is always true.

### `onerr` — Graceful Error Handling

```kukicha
input := reader.ReadString('\n') onerr ""
```

Some operations can fail — reading input might hit an error if the terminal closes. The **`onerr`** clause says "if this fails, use this value instead." Here, if `ReadString` fails, `input` gets set to an empty string.

You can use `onerr` with different fallback strategies:
- `onerr ""` or `onerr 0` — use a default value
- `onerr return` — exit the current function
- `onerr panic "message"` — crash with an error message (for truly unexpected failures)

### `continue` in Context

```kukicha
if input equals ""
    continue
```

When the user presses Enter without typing anything, `continue` skips the rest of the loop body and goes straight back to the `>` prompt.

### `empty` for Null Checking

```kukicha
if idErr not equals empty
    print("Invalid number: {parts[1]}")
```

In Kukicha, **`empty`** represents "no value" (called `nil` in many other languages). When `strconv.Atoi` fails to convert text to a number, it returns an error. If the error `not equals empty`, something went wrong.

---

## What You've Learned

Congratulations! You've built a real tool that talks to the internet. Let's review what you learned:

| Concept | What It Does |
|---------|--------------|
| **Custom Types** | Define your own data structures with `type Name` |
| **JSON Tags** | Map API fields to your type's fields with `json:"..."` |
| **Methods** | Attach functions to types with `function Name on receiver Type` |
| **`reference`** | Modify the original value, not a copy |
| **`onerr`** | Handle errors gracefully with fallback values |
| **Pipe Operator** | Chain operations into readable data pipelines |
| **`empty`** | Check for null/missing values |
| **Command Loop** | Read input, bare `for`, `break`, and `continue` |
| **`fetch` + `json`** | Fetch and parse data from web APIs |

---

## Practice Exercises

Ready for a challenge? Try these enhancements:

1. **Star Sort** — Add a `sort` command that orders repos by star count (highest first)
2. **Save Favorites** — Write favorites to a JSON file so they persist between sessions (hint: `stdlib/files` + `stdlib/json`)
3. **Rate Limit** — GitHub limits unauthenticated requests to 60/hour. Show remaining requests using the `X-RateLimit-Remaining` response header
4. **Compare Users** — Add a `compare <user1> <user2>` command that shows stats side-by-side

---

## What's Next?

You now have solid programming skills with Kukicha! Continue with the tutorial series:

### Tutorial Path

| # | Tutorial | What You'll Learn |
|---|----------|-------------------|
| 1 | ✅ **Beginner Tutorial** — Variables, functions, strings, decisions, lists, loops *(completed)* |
| 2 | ✅ **CLI Explorer** — Custom types, methods, API data, pipes, error handling *(you are here)* |
| 3 | **[Link Shortener](web-app-tutorial.md)** ← Next step! |
|   | Build a web service with HTTP, JSON, redirects |
| 4 | **[Production Patterns](production-patterns-tutorial.md)** (Advanced) |
|   | Add database, concurrency, Go conventions |
|   | **[LLM Scripting](llm-pipe-tutorial.md)** (Bonus) |
|   | Combine shell commands + LLM calls + pipes — a force multiplier |

### Reference Docs

- **[Kukicha Grammar](../kukicha-grammar.ebnf.md)** — Complete language grammar
- **[Standard Library](../kukicha-stdlib-reference.md)** — iterator, slice, fetch, and more

---

**Great job! You've built a real tool that talks to the internet! 🎉**
