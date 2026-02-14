# Production Patterns with Kukicha (Advanced)

**Level:** Advanced
**Time:** 45 minutes
**Prerequisite:** [Link Shortener Tutorial](web-app-tutorial.md)

Welcome to the advanced tutorial! You've built a working link shortener, but it's not ready for real users yet. In this tutorial, we'll add:

- **Database storage** (so links persist across restarts)
- **Random codes** (so links aren't guessable)
- **Safe concurrent access** (so multiple users don't corrupt data)
- **Go conventions** (patterns you'll see in real Go codebases)
- **Proper configuration and validation**

This tutorial bridges Kukicha's beginner-friendly syntax with real-world Go patterns.

---

## What's Wrong with Our Current App?

Our link shortener from the previous tutorial has four problems:

| Problem | Why It Matters |
|---------|----------------|
| **Memory storage** | Links disappear when the server restarts |
| **No locking** | Two users shortening at once could corrupt data |
| **Predictable codes** | Sequential codes like `1`, `2`, `3` are guessable |
| **Global variables** | Makes testing hard and code messy |

Let's fix all four!

---

## Optional: File Persistence (Stepping Stone)

If you want a quick way to persist links without a database, you can save them to a file. This is fine for small, single-user tools, but it's **not safe** for concurrent web requests. That's why this tutorial moves to a database.

```kukicha
import "stdlib/files"
import "stdlib/json"

function SaveLinks(links map of string to Link, filename string) error
    data := links |> json.Marshal() onerr return error "{error}"
    files.Write(filename, data) onerr return error "{error}"
    return empty

function LoadLinks(filename string) (map of string to Link, error)
    data := files.Read(filename) onerr return map of string to Link{}, error "{error}"
    links := map of string to Link{}
    data |> json.Unmarshal(_, reference of links) onerr return map of string to Link{}, error "{error}"
    return links, empty
```

**Why not use this for production?**
- File writes aren't atomic across concurrent requests
- No locking or transactions
- Hard to query efficiently (search by URL, analytics, etc.)

We'll use SQLite because it solves these problems and teaches real-world patterns.

---

## Part 1: Method Receivers

In the previous tutorials, we used Kukicha's `on` syntax for methods:

```kukicha
# Kukicha style — English-like
function Display on link Link() string
    return "{link.code}: {link.url} ({link.clicks} clicks)"
```

This is the **only** method syntax Kukicha supports. When you read Go code, you'll see a different syntax (`func (link Link) Display() string`), but in Kukicha it maps directly to the `on` form. The translation table at the end of this tutorial covers the full mapping.

### Understanding `reference` vs `reference of`

As you read through the code, you'll see two pointer-related keywords:
- **`reference Type`** — Declares a pointer type (e.g., `reference Server` means "pointer to Server")
- **`reference of value`** — Takes the address of an existing value (e.g., `reference of server` converts `server` into a pointer)

Both are correct Kukicha syntax; they're just used in different contexts (declarations vs. operations).

---

## Part 2: Creating a Server Type

Instead of global variables, let's create a proper `Server` type that holds all our state:

```kukicha
import "sync"

type Server
    db Database
    mu sync.RWMutex    # A lock for safe access
    baseURL string     # e.g., "http://localhost:8080"
```

**What's a `sync.RWMutex`?**

It's a "read-write lock" that prevents data corruption:
- **Read Lock** (`RLock`) — Multiple readers can access at once
- **Write Lock** (`Lock`) — Only one writer at a time, blocks everyone else

Think of it like a library book:
- Many people can read the same book at once
- But if someone is writing in it, everyone else has to wait

### Why We Wrap State in a Struct

Instead of using a `LinkStore` with methods, we encapsulate all server state in a `Server` type. This enables:
- **Testability** — Create multiple test instances with different states
- **Dependency injection** — Pass the server instance where needed
- **Concurrency safety** — The mutex lives with the data it protects
- **Composability** — Adding the database is just another field

---

## Part 3: Thread-Safe Methods

Now let's write methods that use locking. We'll also add random code generation:

```kukicha
import "stdlib/random"
```

Random codes solve the "guessable" problem from the previous tutorial. Codes like `"x7km2p"` are much harder to guess than `"1"`, `"2"`, `"3"`. The `random.String(6)` call from `stdlib/random` generates a 6-character alphanumeric code — no boilerplate required.

```kukicha
# CreateLink generates a random code, stores the link, and returns it
function CreateLink on s reference Server(url string) (Link, error)
    s.mu.Lock()              # Exclusive access for writing
    defer s.mu.Unlock()      # Unlock when done (even if there's an error)

    # Generate a unique code (retry if collision)
    code := random.String(6)
    for i := 0 to 10
        _, exists := s.db.GetLink(code) onerr empty
        if not exists
            break
        code = random.String(6)

    link, err := s.db.InsertLink(code, url)
    if err not equals empty
        return Link{}, err

    return link, empty

# GetLink retrieves a link by code
function GetLink on s reference Server(code string) (Link, bool)
    s.mu.RLock()             # Shared access for reading
    defer s.mu.RUnlock()

    link, err := s.db.GetLink(code)
    if err not equals empty
        return Link{}, false
    return link, true

# RecordClick increments the click counter for a link
function RecordClick on s reference Server(code string)
    s.mu.Lock()
    defer s.mu.Unlock()
    s.db.IncrementClicks(code)
```

**Why `reference Server`?**

We use `reference` (a pointer) because:
1. We need to **modify** the server's data
2. Locking only works if everyone uses the **same** lock

---

## Part 4: Adding a Database

Let's store links in SQLite so they persist across restarts.

### Installing the Driver

```bash
go get github.com/mattn/go-sqlite3
```

### Database Helper Type

```kukicha
import "database/sql"
import _ "github.com/mattn/go-sqlite3"

type Database
    db reference sql.DB

type Link
    code string
    url string
    clicks int
    createdAt string json:"created_at"

# Open the database and create the table if needed
function OpenDatabase(filename string) (Database, error)
    db, err := sql.Open("sqlite3", filename)
    if err not equals empty
        return empty, err

    # Create the links table
    createTable := `
        CREATE TABLE IF NOT EXISTS links (
            code TEXT PRIMARY KEY,
            url TEXT NOT NULL,
            clicks INTEGER DEFAULT 0,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )
    `
    createTable |> db.Exec() onerr return empty, error "{error}"

    return Database{db: db}, empty

# Close the database
function Close on d Database()
    if d.db not equals empty
        d.db.Close()
```

### CRUD Operations

```kukicha
# InsertLink creates a new link in the database
function InsertLink on d Database(code string, url string) (Link, error)
    _, execErr := d.db.Exec(
        "INSERT INTO links (code, url) VALUES (?, ?)", code, url)
    if execErr not equals empty
        return Link{}, execErr

    return d.GetLink(code)

# GetLink retrieves a link by its code
function GetLink on d Database(code string) (Link, error)
    row := d.db.QueryRow(
        "SELECT code, url, clicks, created_at FROM links WHERE code = ?", code)

    link := Link{}
    scanErr := row.Scan(
        reference of link.code,
        reference of link.url,
        reference of link.clicks,
        reference of link.createdAt)
    if scanErr not equals empty
        return Link{}, scanErr

    return link, empty

# GetAllLinks returns all links, newest first
function GetAllLinks on d Database() (list of Link, error)
    rows, queryErr := d.db.Query(
        "SELECT code, url, clicks, created_at FROM links ORDER BY created_at DESC")
    if queryErr not equals empty
        return empty, queryErr
    defer rows.Close()

    links := empty list of Link
    for rows.Next()
        link := Link{}
        scanErr := rows.Scan(
            reference of link.code,
            reference of link.url,
            reference of link.clicks,
            reference of link.createdAt)
        if scanErr not equals empty
            continue
        links = append(links, link)

    return links, empty

# IncrementClicks adds 1 to the click counter (called on every redirect)
function IncrementClicks on d Database(code string) error
    "UPDATE links SET clicks = clicks + 1 WHERE code = ?"
        |> d.db.Exec(code) onerr return error "{error}"
    return empty

# DeleteLink removes a link by its code
function DeleteLink on d Database(code string) error
    "DELETE FROM links WHERE code = ?" |> d.db.Exec(code) onerr return error "{error}"
    return empty
```

---

## Part 5: The Production Server

Now let's put it all together into a production-ready server:

```kukicha
# Standard library
import "fmt"
import "log"
import "net/http"
import "sync"
import "database/sql"
import "encoding/json/v2"

# Kukicha stdlib
import "stdlib/string"
import "stdlib/validate"
import "stdlib/http" as httphelper
import "stdlib/must"
import "stdlib/env"
import "stdlib/random"

# Third-party
import _ "github.com/mattn/go-sqlite3"

# --- Types ---

type Link
    code string
    url string
    clicks int
    createdAt string json:"created_at"

type Server
    db Database
    mu sync.RWMutex
    baseURL string

type ShortenRequest
    url string

type ShortenResponse
    code string
    url string
    shortUrl string json:"short_url"
    clicks int

type ErrorResponse
    err string json:"error"

# --- Server Constructor ---

function NewServer(dbPath string, baseURL string) (reference Server, error)
    db, dbErr := OpenDatabase(dbPath)
    if dbErr not equals empty
        return empty, dbErr

    server := Server{db: db, baseURL: baseURL}
    return reference of server, empty

# --- HTTP Handlers ---

# POST /shorten — Create a new short link
function handleShorten on s reference Server(w http.ResponseWriter, r reference http.Request)
    if r.Method not equals "POST"
        httphelper.MethodNotAllowed(w)
        return

    # Parse request body
    input := ShortenRequest{}
    readErr := r |> httphelper.ReadJSON(reference of input)
    if readErr not equals empty
        httphelper.JSONBadRequest(w, "Invalid JSON")
        return

    # Validate URL
    _, emptyErr := input.url |> validate.NotEmpty()
    if emptyErr not equals empty
        httphelper.JSONBadRequest(w, "URL is required")
        return

    _, urlErr := input.url |> validate.URL()
    if urlErr not equals empty
        httphelper.JSONBadRequest(w, "Invalid URL — must start with http:// or https://")
        return

    # Create the link
    s.mu.Lock()
    code := random.String(6)
    # Retry on collision (unlikely with 6 random chars, but be safe)
    for i := 0 to 10
        _, getErr := s.db.GetLink(code)
        if getErr not equals empty
            break
        code = random.String(6)
    link, createErr := s.db.InsertLink(code, input.url)
    s.mu.Unlock()

    if createErr not equals empty
        log.Printf("Error creating link: %v", createErr)
        httphelper.JSONError(w, 500, "Failed to create link")
        return

    result := ShortenResponse
        code: link.code
        url: link.url
        shortUrl: "{s.baseURL}/r/{link.code}"
        clicks: 0

    httphelper.JSONStatus(w, 201, result)

# GET /r/{code} — Redirect to original URL
function handleRedirect on s reference Server(w http.ResponseWriter, r reference http.Request)
    code := r.URL.Path |> string.TrimPrefix("/r/")
    if code equals "" or code equals r.URL.Path
        httphelper.JSONBadRequest(w, "Missing link code")
        return

    # Look up the link
    s.mu.RLock()
    link, getErr := s.db.GetLink(code)
    s.mu.RUnlock()

    if getErr not equals empty
        httphelper.JSONNotFound(w, "Link not found")
        return

    # Record the click (async-safe with its own lock)
    go
        s.mu.Lock()
        s.db.IncrementClicks(code)
        s.mu.Unlock()

    http.Redirect(w, r, link.url, 301)

# GET /links — List all links
function handleListLinks on s reference Server(w http.ResponseWriter, r reference http.Request)
    if r.Method not equals "GET"
        httphelper.MethodNotAllowed(w)
        return

    s.mu.RLock()
    links, err := s.db.GetAllLinks()
    s.mu.RUnlock()

    if err not equals empty
        log.Printf("Error fetching links: %v", err)
        httphelper.JSONError(w, 500, "Failed to fetch links")
        return

    httphelper.JSON(w, links)

# /links/{code} — Get info or delete a link
function handleLinkDetail on s reference Server(w http.ResponseWriter, r reference http.Request)
    code := r.URL.Path |> string.TrimPrefix("/links/")
    if code equals "" or code equals r.URL.Path
        httphelper.JSONBadRequest(w, "Missing link code")
        return

    switch r.Method
        when "GET"
            s.mu.RLock()
            link, err := s.db.GetLink(code)
            s.mu.RUnlock()
            if err not equals empty
                httphelper.JSONNotFound(w, "Link not found")
                return
            httphelper.JSON(w, link)

        when "DELETE"
            s.mu.Lock()
            deleteErr := s.db.DeleteLink(code)
            s.mu.Unlock()
            if deleteErr not equals empty
                log.Printf("Error deleting link: %v", deleteErr)
                httphelper.JSONError(w, 500, "Failed to delete link")
                return
            w |> .WriteHeader(204)

        otherwise
            httphelper.MethodNotAllowed(w)

# --- Main Entry Point ---

function main()
    # Configuration from environment variables (production best practice)
    dbPath := must.EnvOr("DATABASE_URL", "links.db")
    port := env.GetOr("PORT", ":8080")
    baseURL := env.GetOr("BASE_URL", "http://localhost{port}")

    # Create the server
    server := NewServer(dbPath, baseURL) onerr panic "Failed to open database: {error}"
    defer server.db.Close()

    # Register routes
    http.HandleFunc("/shorten", server.handleShorten)
    http.HandleFunc("/r/", server.handleRedirect)
    http.HandleFunc("/links", server.handleListLinks)
    http.HandleFunc("/links/", server.handleLinkDetail)

    log.Printf("Link shortener starting on %s", port)
    log.Printf("Database: %s", dbPath)
    log.Printf("Base URL: %s", baseURL)

    http.ListenAndServe(port, empty) onerr panic "Server failed: {error}"
```

---

## Part 6: What Changed?

Let's compare the web tutorial version with this production version:

| Aspect | Web Tutorial | Production |
|--------|--------------|------------|
| **Storage** | In-memory map | SQLite database |
| **Codes** | Sequential (`1`, `2`, ...) | Random 6-character (`x7km2p`) |
| **Safety** | None | `sync.RWMutex` on every access |
| **Clicks** | Lost on restart | Persisted, tracked with `go func()` |
| **Validation** | Manual string checks | `stdlib/validate` (URL, NotEmpty) |
| **Config** | Hardcoded | Environment variables (`PORT`, `DATABASE_URL`) |
| **Errors** | Manual JSON encoding | `stdlib/http` helpers |
| **Optional/Result values** | Not modeled explicitly | `stdlib/result` (`Some`/`None`, `Ok`/`Err`) |
| **HTTP retries** | Single attempt only | `stdlib/retry` config + manual retry loop |
| **Lifecycle** | `LinkStore` struct | `NewServer()` constructor, `defer Close()` |

---

## Part 7: Go Conventions You've Learned

### Pointer Receivers

```kukicha
# Kukicha: "reference Server"  →  Go: "*Server"
function Method on s reference Server()
```

Use pointer receivers when:
- The method modifies the receiver
- The receiver is large (avoids copying)
- You need consistency (if one method needs a pointer, use pointers for all)

### Constructors

Go doesn't have constructors, so we use functions named `New<Type>`:

```kukicha
function NewServer(config string) (reference Server, error)
    # Initialize and return
```

### Defer for Cleanup

```kukicha
function DoWork() error
    resource := Acquire() onerr return error "{error}"
    defer resource.Close()  # Guaranteed to run when function exits

    # Do work...
    return empty
```

### Goroutines for Background Work

```kukicha
# Fire-and-forget click tracking
go
    s.mu.Lock()
    s.db.IncrementClicks(code)
    s.mu.Unlock()
```

The `go` keyword launches code in a separate goroutine. When followed by an indented block, Kukicha wraps it in an anonymous function for you (no need for the `go func()...()` pattern from Go). We use it for click tracking so the redirect response isn't delayed by a database write.

You can still use `go` with a direct function call for single operations:
```kukicha
go processItem(item)
```

---

## Part 8: Production-Ready Packages

Kukicha includes several packages designed for production code:

### Configuration with `env` and `must`

```kukicha
import "stdlib/env"
import "stdlib/must"

function main()
    # Required config (panic if missing)
    apiKey := must.Env("API_KEY")

    # Optional config with defaults
    port := env.GetOr("PORT", ":8080")
    debug := env.GetBoolOrDefault("DEBUG", false)
    timeout := env.GetIntOr("TIMEOUT", 30) onerr 30

    # Parse a comma-separated list from any source
    allowedOrigins := env.GetOr("ALLOWED_ORIGINS", "http://localhost:3000")
        |> env.SplitAndTrim(",")
```

### Input Validation with `validate`

```kukicha
import "stdlib/validate"

function ValidateShortenRequest(url string) error
    url |> validate.NotEmpty() onerr return error "URL is required"
    url |> validate.URL() onerr return error "Invalid URL"
    url |> validate.MaxLength(2048) onerr return error "URL too long"
    return empty
```

### HTTP Helpers

```kukicha
import "stdlib/http" as httphelper

function HandleRequest(w http.ResponseWriter, r reference http.Request)
    # Read JSON body
    input := ShortenRequest{}
    readErr := r |> httphelper.ReadJSON(reference of input)
    if readErr not equals empty
        httphelper.JSONBadRequest(w, "Invalid JSON")
        return

    # Send JSON responses
    httphelper.JSON(w, link)                        # 200 OK
    httphelper.JSONStatus(w, 201, link)             # 201 Created
    httphelper.JSONNotFound(w, "Link not found")    # 404
    httphelper.JSONError(w, 500, "Server error")    # Any status

    # Query parameters
    page := httphelper.GetQueryIntOr(r, "page", 1)
    search := httphelper.GetQueryParam(r, "q")
```

### Rust-Style Optionals with `result`

```kukicha
import "stdlib/result"

# Pattern 1: Optional for nullable cache lookups
function FindCachedUser(id string) result.Optional
    user, exists := userCache[id]
    if not exists
        return result.None()
    return result.Some(user)

# Usage
opt := FindCachedUser(id)
if result.IsSome(opt)
    user := result.Unwrap(opt)
    print("Found: {user}")
```

```kukicha
# Pattern 2: Result for operations that can fail
function FetchLinkResult on s reference Server(code string) result.Result
    link, err := s.db.GetLink(code)
    if err not equals empty
        return result.Err(err)
    return result.Ok(link)

# Usage with Match for clean dispatch
result.Match(
    s.FetchLinkResult(code),
    (link any) => httphelper.JSON(w, link),
    (err error) => httphelper.JSONNotFound(w, "Link not found")
)
```

```kukicha
# Pattern 3: AndThen for chaining fallible steps
s.FetchLinkResult(code)
    |> result.AndThen((link any) => ValidateLinkResult(link))
    |> result.UnwrapOrResult(Link{})
```

Use `result` when you want success/failure as a first-class value you can return, pass, or store, instead of only using multiple return values.

### Resilient HTTP Calls with `retry`

```kukicha
import "stdlib/retry"

function FetchReposResilient(username string) list of Repo
    url := "https://api.github.com/users/{username}/repos?per_page=30&sort=stars"
    cfg := retry.New()
        |> retry.Attempts(3)
        |> retry.Delay(500)
        |> retry.Backoff(1)   # 1 = exponential: 500ms, 1000ms, 2000ms

    attempt := 0
    for attempt < cfg.MaxAttempts
        repos := empty list of Repo
        fetchOk := true

        fetch.Get(url)
            |> fetch.CheckStatus()
            |> fetch.Bytes()
            |> json.Unmarshal(reference repos)
            onerr
                fetchOk = false

        if fetchOk
            return repos

        if attempt < cfg.MaxAttempts - 1
            print("Attempt {attempt + 1} failed, retrying...")
            retry.Sleep(cfg, attempt)

        attempt = attempt + 1

    print("Failed to fetch repos for '{username}' after {cfg.MaxAttempts} attempts")
    return empty list of Repo
```

Notes:
- `retry.New()` defaults to 3 attempts, 1000ms delay, exponential backoff.
- `retry.Backoff(0)` is linear (constant delay), `retry.Backoff(1)` is exponential.
- `retry.Sleep(cfg, attempt)` computes the correct pause for each attempt.
- `retry.Do()` is intentionally not provided; in Kukicha, a manual loop is the recommended pattern.

---

## Part 9: Panic and Recover

In production, you want your server to stay alive even if a bug causes a crash. In Go (and Kukicha), a crash is called a **panic**.

You can "catch" a panic using `recover`. This is usually done in **middleware** (code that wraps every request) or at the top of a background job.

### Middleware Example

Here's how to write a middleware that recovers from panics and logs the error instead of crashing the server:

```kukicha
import "log"
import "net/http"
import "stdlib/env"

function RecoveryMiddleware(next http.Handler) http.Handler
    return http.HandlerFunc(function(w http.ResponseWriter, r reference http.Request)
        # Defer a function that calls recover()
        defer function()
            err := recover()
            if err not equals empty
                log.Printf("PANIC RECOVERED: %v", err)
                http.Error(w, "Internal Server Error", 500)
        () # Call the deferred function

        # Call the next handler
        next.ServeHTTP(w, r)
    )
```

**Key points:**
- `panic("message")` stops normal execution immediately.
- `recover()` regains control, but **only** if called inside a `defer` function.
- If you don't recover, the program exits.

---

## Summary: The Kukicha Learning Path

You've completed the full Kukicha tutorial series!

| Tutorial | What You Learned |
|----------|-----------------|
| ✅ **1. Beginner** | Variables, functions, strings, loops, pipes |
| ✅ **2. CLI Explorer** | Types, methods (`on`), API data, arrow lambdas, `fetch` + `json` |
| ✅ **3. Link Shortener** | HTTP servers, JSON, REST APIs, maps, redirects |
| ✅ **4. Production** | Databases, concurrency, Go conventions, `env`/`must`, `validate`, `http`, `result`, `retry` |
|    **Bonus: LLM Scripting** | Shell + LLM + pipes — [try it!](llm-pipe-tutorial.md) |

---

## Where to Go From Here

### Explore More

- **[Kukicha Grammar](../kukicha-grammar.ebnf.md)** — Complete language grammar
- **[Standard Library](../kukicha-stdlib-reference.md)** — iterator, slice, and more
- **[LLM Scripting Tutorial](llm-pipe-tutorial.md)** — Combine shell + LLM + pipes

### Build Projects

Ideas for your next project:
- **Paste Bin** — Share code snippets with syntax highlighting
- **Webhook Relay** — Receive, log, and forward webhooks
- **Health Checker** — Monitor URLs and alert on failures
- **Chat Application** — WebSockets, real-time messaging

### Learn More Go

Now that you know Kukicha, learning Go will be easy:
- [Go Tour](https://go.dev/tour/) — Official interactive tutorial
- [Effective Go](https://go.dev/doc/effective_go) — Go best practices
- [Go by Example](https://gobyexample.com/) — Practical examples

---

## Kukicha to Go Translation

Here's a quick reference for translating between Kukicha and Go:

| Kukicha | Go |
|---------|-----|
| `list of int` | `[]int` |
| `map of string to int` | `map[string]int` |
| `reference Type` | `*Type` |
| `reference of x` | `&x` |
| `empty` | `nil` |
| `equals` | `==` |
| `not equals` | `!=` |
| `and` | `&&` |
| `or` | `\|\|` |
| `not` | `!` |
| `for item in list` | `for _, item := range list` |
| `function Name on x Type` | `func (x Type) Name()` |
| `result onerr default` | `if err != nil { ... }` |
| `a \|> f(b)` | `f(a, b)` |
| `a \|> f(b, _)` | `f(b, a)` (placeholder) |
| `(r Repo) => r.Stars > 100` | `func(r Repo) bool { return r.Stars > 100 }` |
| `go` + indented block | `go func() { ... }()` |
| `switch x` / `when a` / `otherwise` | `switch x { case a: ... default: ... }` |

---

**Congratulations! You're now a Kukicha developer! 🎉🌱**
