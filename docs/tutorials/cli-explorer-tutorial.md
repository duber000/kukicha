# Build a GitHub Repo Explorer with Kukicha

**Level:** Intermediate
**Time:** 15-18 minutes
**Prerequisite:** [Data & AI Scripting](data-scripting-tutorial.md)

Welcome back! In the beginner tutorial, you learned about variables, functions, strings, decisions, lists, and loops. Now we're going to build something genuinely useful: a **GitHub Repo Explorer** that fetches real data from the internet and lets you browse it interactively.

## What You'll Learn

In this tutorial, you'll discover how to:
- Create **custom types** to organize related data
- Write **methods** that belong to types
- Use `reference` to modify data in place
- Apply familiar **`onerr` + pipeline** patterns to larger programs
- Write concise inline functions with **arrow lambdas (`=>`)** for filtering and mapping
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
kukicha init    # go mod init + extracts stdlib for imports like "stdlib/fetch"
```

---

## Step 1: Creating a Repo Type

> **Reminder:** This tutorial assumes Tutorials 1 and 2. If you need a refresher on core syntax, [revisit the beginner tutorial](beginner-tutorial.md).

In the beginner tutorial, you learned about basic types like `string`, `int`, and `bool`. Now let's create our own type to represent a GitHub repository.

Create a file called `explorer.kuki`:

```kukicha
import "stdlib/fetch"
import "stdlib/string"
import "stdlib/slice"

# A Repo represents a GitHub repository
# The `as "..."` aliases map API JSON field names to readable field names
type Repo
    Name string as "name"
    Description string as "description"
    Stars int as "stargazers_count"
    Language string as "language"
    URL string as "html_url"
```

**What's new here?**

We're defining a custom **type** called `Repo` — a blueprint for GitHub repository data. Each repo has:
- `Name` — The repository name
- `Description` — What it's about
- `Stars` — How many people have starred it
- `Language` — Primary programming language
- `URL` — Link to view it on GitHub

The **`as "..."`** part after each field is a JSON alias. GitHub's API returns fields like `"stargazers_count"`, and the alias maps that data into your readable `Stars` field.

---

## Step 2: Fetching Repos from GitHub

This is where Kukicha shines. Let's fetch real data from GitHub's public API:

```kukicha
# FetchRepos gets repositories for a GitHub user or organization
function FetchRepos(username string) list of Repo
    url := "https://api.github.com/users/{username}/repos?per_page=30&sort=stars"
    repos := fetch.Get(url)
        |> fetch.CheckStatus()
        |> fetch.Json(list of Repo) onerr
        print("Failed to fetch repos for '{username}': {error}")
        return empty list of Repo

    return repos
```

**What's happening here?**

This reuses the same pipeline pattern from Tutorial 2, now applied to live API data:

1. `fetch.Get(url)` — Make an HTTP request to GitHub's API
2. `|> fetch.CheckStatus()` — Verify we got a success response (not a 404)
3. `|> fetch.Json(list of Repo)` — Decode JSON directly into our `list of Repo`

**`onerr` in action:** If *any* step fails (network error, bad status code, invalid JSON), execution jumps to the `onerr` block.

### Let's Try It

Add a `main` function to test our fetcher:

```kukicha
function main()
    print("Fetching repos for 'golang'...")
    repos := FetchRepos("golang")
    print("Found {len(repos)} repos!\n")

    for repo in repos[:5]  # [:5] is a slice — it gives us just the first 5 items (see list slicing in the beginner tutorial)
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
    name := repo.Name |> string.PadRight(30)
    lang = lang |> string.PadRight(12)
    return "  {index}. {name}  ⭐ {repo.Stars}  {lang}  {repo.Description}"
```

**Reading this method:**
- `function Summary` — We're creating a method called "Summary"
- `on repo Repo` — This method works on a `Repo` (syntax: receiver name first, then the type). Inside the method, we call it `repo`
- `(index int)` — The method also takes an index number for display numbering
- `string` — The method returns a string
- `string.PadRight(30)` — pads a string to exactly 30 characters wide by adding spaces on the right (the default pad character), so every repo name takes up the same width and columns line up cleanly in the output

### Filtering with Pipes and Arrow Lambdas

Now let's write functions to filter repos. This is where pipes and **arrow lambdas** shine together:

```kukicha
# FilterByLanguage returns repos matching a language (case-insensitive)
function FilterByLanguage(repos list of Repo, language string) list of Repo
    return repos |> slice.Filter(r => r.Language |> string.ToLower() |> string.Contains(language |> string.ToLower()))
```

**What's new here?**

The `r => ...` form is an **arrow lambda** — a concise inline function used heavily with `slice.Filter` and `slice.Map`. The compiler infers the parameter type from context, so you don't need to write `(r Repo) =>` explicitly.

### Let's Try It

Add import "stdlib/string" and import "stdlib/slice" to the imports at the top

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

> **Note:** The final program uses `stdlib/input` for reading console input. This replaces the `bufio`/`os` boilerplate you'd need in plain Go. The package also provides `input.Confirm(prompt)` for yes/no questions and `input.Choose(prompt, options)` for numbered menus — both return errors you can handle with `onerr`.

Replace the contents of `explorer.kuki` with the complete program:

```kukicha
import "stdlib/fetch"
import "stdlib/string"
import "stdlib/slice"
import "stdlib/input"
import "stdlib/cast"

# --- Type Definitions ---

type Repo
    Name string as "name"
    Description string as "description"
    Stars int as "stargazers_count"
    Language string as "language"
    URL string as "html_url"

type Explorer
    repos list of Repo
    favorites list of Repo
    username string

# --- Repo Methods ---

function Summary on repo Repo(index int) string
    lang := repo.Language
    if lang equals ""
        lang = "n/a"
    name := repo.Name |> string.PadRight(30)
    lang = lang |> string.PadRight(12)
    return "  {index}. {name}  ⭐ {repo.Stars}  {lang}  {repo.Description}"

# --- Data Fetching ---

function FetchRepos(username string) list of Repo
    url := "https://api.github.com/users/{username}/repos?per_page=30&sort=stars"
    repos := fetch.Get(url)
        |> fetch.CheckStatus()
        |> fetch.Json(list of Repo) onerr
            print("Failed to fetch repos for '{username}': {error}")
            return empty list of Repo
    return repos

# --- Filter Functions ---

function FilterByLanguage(repos list of Repo, language string) list of Repo
    return repos |> slice.Filter(r => r.Language |> string.ToLower() |> string.Contains(language |> string.ToLower()))

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

    # Main loop
    for
        # Read user input — default to empty string on error
        line := input.ReadLine("> ") onerr ""

        if line equals ""
            continue

        # SplitN(" ", 2) splits into at most 2 parts
        parts := line |> string.SplitN(" ", 2)
        command := parts[0] |> string.ToLower()

        command |> switch
            when "quit", "exit", "q"
                print("Goodbye!")
                break

            when "help", "?"
                PrintHelp()

            when "list", "ls"
                ex.ShowList()

            when "fetch"
                if len(parts) < 2
                    print("Usage: fetch <username>")
                    continue
                ex.Fetch(parts[1])

            when "filter"
                if len(parts) < 2
                    # No argument — build a menu from languages in the current repo list
                    languages := make(list of string, 0)
                    for repo in ex.repos
                        if repo.Language not equals "" and not slice.Contains(languages, repo.Language)
                            languages = append(languages, repo.Language)
                    if slice.IsEmpty(languages)
                        print("No language data in current repos.")
                        continue
                    idx := input.Choose("Filter by language:", languages) onerr
                        print("Cancelled.")
                        continue
                    filtered := FilterByLanguage(ex.repos, languages[idx])
                    print("\nShowing {len(filtered)} repos in {languages[idx]}")
                    for i, repo in filtered
                        print(repo.Summary(i + 1))
                    print("")
                    continue
                filtered := FilterByLanguage(ex.repos, parts[1])
                print("\nShowing {len(filtered)} repos matching '{parts[1]}'")
                for i, repo in filtered
                    print(repo.Summary(i + 1))
                print("")

            when "search"
                if len(parts) < 2
                    print("Usage: search <text>")
                    continue
                term := parts[1] |> string.ToLower()
                results := ex.repos |> slice.Filter(r =>
                    name := r.Name |> string.ToLower()
                    desc := r.Description |> string.ToLower()
                    return name |> string.Contains(term) or desc |> string.Contains(term)
                )
                print("\nFound {len(results)} repos matching '{parts[1]}'")
                for i, repo in results
                    print(repo.Summary(i + 1))
                print("")

            when "fav"
                if len(parts) < 2
                    print("Usage: fav <number>")
                    continue
                # Parse the number — print a message and skip if it's not valid
                id := cast.ToInt(parts[1]) onerr
                    print("Invalid number: {parts[1]}")
                    continue
                ex.AddFavorite(id)

            when "favs", "favorites"
                ex.ShowFavorites()

            otherwise
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

> filter
Filter by language:

  1) Go
  2) Python
  3) Shell

Enter number (or q to quit): 1
Showing 3 repos in Go
  1. go  ⭐ 125000  Go  The Go programming language
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

The final program introduced several concepts that deserve a closer look.

### JSON Aliases — Mapping API Data to Types

```kukicha
type Repo
    Stars int as "stargazers_count"
```

GitHub's API returns `"stargazers_count"` in its JSON response. The alias `as "stargazers_count"` tells the JSON parser: "when you see `stargazers_count` in the data, put it in the `Stars` field." This keeps your type fields readable while matching API conventions.

### Bare `for` — The Infinite Loop

```kukicha
for
    # ... read input and process commands ...
```

A `for` with no condition runs forever. This is the standard pattern for programs that wait for user input — the loop keeps running until something inside calls `break`. You saw in the beginner tutorial that `for condition` runs while the condition is true; a bare `for` is just the extreme case where the condition is always true.

### `switch`/`when`/`otherwise` — Command Dispatch

```kukicha
switch command
    when "quit", "exit", "q"
        print("Goodbye!")
        break
    when "help"
        PrintHelp()
    otherwise
        print("Unknown command")
```

Instead of a long chain of `if`/`else if`/`else`, `switch` with `when` branches makes command dispatch clean and readable. Each `when` can match **multiple values** separated by commas — so `when "quit", "exit", "q"` handles all three in one branch. The `otherwise` branch catches anything that doesn't match (like `default` in other languages).

You can also use `switch` without an expression for condition-based branching:

```kukicha
switch
    when stars >= 1000
        print("Popular!")
    when stars >= 100
        print("Growing")
    otherwise
        print("New")
```

### Piped Switch — Routing Values Directly

In the main loop, we piped the command directly into a `switch`:

```kukicha
command |> switch
    when "quit", "exit", "q"
        print("Goodbye!")
        break
    when "help"
        PrintHelp()
```

This is equivalent to `switch command`, but uses the pipe `|>` to feed the value in. This is especially useful when the value comes from a chain of transformations — you can pipe and switch in one flow without an intermediate variable.

### Typed Piped Switch (`|> switch as v`) — Branching on Type

Sometimes you have a value whose exact type isn't known — for example, data from a JSON API where a field might be a string, a number, or a list. A **typed piped switch** lets you branch on what the value actually is:

```kukicha
response |> switch as v
    when string
        print("Got text: {v}")
    when int
        print("Got number: {v}")
    when list of string
        for item in v
            print("  - {item}")
    otherwise
        print("Unknown type")
```

The `as v` binding gives you a correctly-typed variable in each branch — so in the `when string` branch, `v` is a `string` you can use directly. This avoids manual type assertions and makes the code easier to read.

### Arrow Lambdas (`=>`) — Concise Inline Functions

```kukicha
repos |> slice.Filter(r => r.Stars > 100)
```

An **arrow lambda** is a short inline function. The part before `=>` declares the parameters, and the part after is the body. For single-expression lambdas, the result is returned automatically — no `return` needed.

When calling stdlib functions like `slice.Filter` or `sort.ByKey`, the compiler **infers the parameter type** from the list being filtered/sorted — so you don't need to write `(r Repo) =>`, just `r =>`.

Arrow lambdas come in two forms:
- **Expression form** (one line, auto-return): `r => r.Stars > 100`
- **Block form** (multi-statement, explicit `return`):
  ```kukicha
  repos |> slice.Filter(r =>
      name := r.Name |> string.ToLower()
      desc := r.Description |> string.ToLower()
      return name |> string.Contains(term) or desc |> string.Contains(term)
  )
  ```

In the block form, you **must use `return`** explicitly because the lambda contains multiple statements. The expression form works without `return` because there's only one expression — the result is automatically returned.

They're especially useful with `slice.Filter`, `slice.Map`, and other functional helpers where a full `function(...)` literal would be verbose.

### `continue` in Context

```kukicha
if input equals ""
    continue
```

When the user presses Enter without typing anything, `continue` skips the rest of the loop body and goes straight back to the `>` prompt.

### `cast.ToInt` and `stdlib/cast`

```kukicha
id := cast.ToInt(parts[1]) onerr
    print("Invalid number: {parts[1]}")
    continue
```

`stdlib/cast` provides type conversion utilities that return errors instead of panicking. `cast.ToInt` converts any value to `int` — strings, floats, booleans. It's cleaner than calling `strconv.Atoi` directly and works with `onerr` naturally.

---

## Step 7: Ship It!

One of Kukicha's superpowers is that it compiles to a **single, standalone binary**.

With Python, you need the interpreter installed. Tools like `uv` make this easier (`uv run script.py`), but the user still needs `uv` installed.

A compiled Kukicha binary needs **nothing**. No runtime, no `pip`, no `node_modules`, no containers. You can email the file to a friend, and it just runs. 

### Build Your Tool

Run this in your terminal:

```bash
kukicha build explorer.kuki
```

You'll see a new file named `explorer` (or `explorer.exe` on Windows). You can run it directly:

```bash
./explorer
# === GitHub Repo Explorer ===
# Type 'help' for commands
```

### Cross-Compilation (The Killer Feature)

Want to send your tool to a friend who uses Linux while you're on a Mac? No problem. Just tell Kukicha where you're sending it:

```bash
# Build for Linux
GOOS=linux kukicha build explorer.kuki

# Build for Windows
GOOS=windows kukicha build explorer.kuki
```

That's it. You've just created software you can ship to anyone.

---

## Step 8: Make it Fast (Bonus)

Let's use **concurrency** to do multiple things at once. We'll add a `fetch-all` command that grabs repos for multiple users simultaneously.

### The Power of `stdlib/concurrent`

In most languages, doing things in parallel is hard. In Kukicha, it's one line.

First, import the concurrent package in `explorer.kuki`:

```kukicha
import "stdlib/concurrent"
import "stdlib/slice"
```

Now add a new `fetch-all` command to your `switch` block in `main`:

```kukicha
            when "fetch-all"
                # Usage: fetch-all golang google facebook
                if len(parts) < 2
                    print("Usage: fetch-all <user1> <user2> ...")
                    continue

                users := parts[1] |> string.Split(" ")
                print("Fetching {len(users)} users in parallel...")

                # Fetch all users concurrently and flatten into one list
                repoLists := concurrent.Map(users, u => FetchRepos(u))
                ex.repos = slice.Concat(repoLists)
                print("Fetched {len(ex.repos)} total repos from {len(users)} users.")
```

**How it works:**
- `concurrent.Map` runs `FetchRepos` for every username **at the same time** — each in its own goroutine — and collects results in order.
- `slice.Concat` flattens the list of lists into a single list.
- For large lists, use `concurrent.MapWithLimit(users, 3, u => FetchRepos(u))` to cap concurrency.

Want to understand how goroutines and channels work under the hood? See the [Concurrent URL Health Checker](concurrent-url-health-checker.md) tutorial.

---

## Step 9: Subcommand Mode with `cli` (Bonus)

The interactive loop is great for exploration, but sometimes you want a tool you can script:

```bash
./explorer fetch golang
./explorer list --filter=go --limit=10
./explorer search "web framework"
```

For that, use `stdlib/cli` with **subcommands**. Each interactive command becomes its own subcommand with typed flags.

First, add the import:

```kukicha
import "stdlib/cli"
```

Then build a subcommand-based entrypoint:

```kukicha
function main()
    _ = cli.New("explorer")
        |> cli.Description("Explore GitHub repos from the command line")
        |> cli.GlobalFlag("limit", "Max repos to show", "10")
        |> cli.Command("fetch", "Fetch repos for a GitHub user or org")
        |> cli.CommandFlag("fetch", "user", "GitHub username", "")
        |> cli.CommandAction("fetch", doFetch)
        |> cli.Command("list", "List fetched repos")
        |> cli.CommandFlag("list", "filter", "Filter by language", "")
        |> cli.CommandAction("list", doList)
        |> cli.Command("search", "Search repos by keyword")
        |> cli.CommandFlag("search", "query", "Search term", "")
        |> cli.CommandAction("search", doSearch)
        |> cli.RunApp() onerr panic "CLI error: {error}"

function doFetch(args cli.Args)
    user := cli.GetString(args, "user")
    if user equals ""
        print("Usage: explorer fetch --user USERNAME")
        return
    repos := FetchRepos(user)
    print("Fetched {len(repos)} repos for {user}")

function doList(args cli.Args)
    filter := cli.GetString(args, "filter")
    limit := cli.GetInt(args, "limit") onerr 10
    # ... filter and display repos
```

Running `./explorer --help` auto-generates:

```
Explore GitHub repos from the command line

Usage:
  explorer <command> [flags]

Commands:
  fetch         Fetch repos for a GitHub user or org
  list          List fetched repos
  search        Search repos by keyword

Global Flags:
  --limit             Max repos to show (default: 10)

Use "explorer <command> --help" for more information about a command.
```

**What this adds:**
- `cli.Command(...)` registers a named subcommand with a description.
- `cli.CommandFlag(...)` adds flags specific to one subcommand.
- `cli.GlobalFlag(...)` adds flags shared across all subcommands.
- `cli.CommandAction(...)` wires a handler function per subcommand.
- `--help` is auto-generated for the app and each subcommand.

**When to use which style:**
- Use the **interactive loop** when you want guided exploration with multiple commands in one session.
- Use **subcommand CLI mode** when you want shell scripts, automation, and pipelines.
- Use the **flat API** (`cli.Arg` + `cli.AddFlag` + `cli.Action`) for simple single-purpose tools.

---

## What You've Learned

Congratulations! You've built a real tool that talks to the internet. Let's review what you learned:

| Concept | What It Does |
|---------|--------------|
| **Single Binary** | Compile your app for any OS with `kukicha build` |
| **Concurrency** | Run tasks concurrently with goroutines and channels |
| **Custom Types** | Define your own data structures with `type Name` |
| **JSON Aliases** | Map API fields to your type's fields with `as "..."` |
| **Methods** | Attach functions to types with `function Name on receiver Type` |
| **`reference`** | Modify the original value, not a copy |
| **`onerr`** | Handle errors gracefully with fallback values |
| **Pipe Operator** | Chain operations into readable data pipelines |
| **Arrow Lambdas** | Concise inline functions with `r => expr` (type inferred from context) |
| **`empty`** | Check for null/missing values |
| **`switch`/`when`** | Clean command dispatch with multiple matches per branch |
| **`\|> switch as v`** | Branch on a value's type with typed piped switch |
| **Command Loop** | Read input, bare `for`, `break`, and `continue` |
| **`fetch` + `json`** | Fetch and parse data from web APIs |
| **`cli`** | Build CLI tools with subcommands, flags, and auto-generated help |
| **`cast`** | Convert strings/values to typed numbers with `onerr` |

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
| 1 | **[Beginner Tutorial](beginner-tutorial.md)** | Variables, functions, strings, decisions, lists, loops, pipes |
| 2 | **[Data & AI Scripting](data-scripting-tutorial.md)** | Maps (Key-Value), parsing CSVs, shell commands, AI scripting |
| 3 | ✅ **CLI Explorer** | Custom types, methods, API data, arrow lambdas, error handling *(you are here)* |
| 4 | **[Link Shortener](web-app-tutorial.md)** ← Next! | HTTP servers, JSON, REST APIs, redirects |
| 5 | **[Production Patterns](production-patterns-tutorial.md)** | Databases, concurrency, Go conventions |

### Reference Docs

- **[Kukicha Grammar](../kukicha-grammar.ebnf.md)** — Complete language grammar
- **[Standard Library](../../stdlib/AGENTS.md)** — iterator, slice, fetch, and more

---

Great job. You've built a real tool that talks to the internet.
