# Production Patterns with Kukicha (Advanced)

**Level:** Advanced  
**Time:** 45 minutes  
**Prerequisite:** [Web Todo Tutorial](web-app-tutorial.md)

Welcome to the advanced tutorial! You've built a working web API, but it's not ready for real users yet. In this tutorial, we'll add:

- **Database storage** (so data persists)
- **Safe concurrent access** (so multiple users don't corrupt data)
- **Go conventions** (patterns you'll see in real Go codebases)
- **Proper project structure** (organized code)

This tutorial bridges Kukicha's beginner-friendly syntax with real-world Go patterns.

---

## What's Wrong with Our Current App?

Our web API from the previous tutorial has three problems:

| Problem | Why It Matters |
|---------|----------------|
| **Memory storage** | Data disappears when the server restarts |
| **No locking** | Two users updating at once could corrupt data |
| **Global variables** | Makes testing hard and code messy |

Let's fix all three!

---

## Part 1: Method Receivers

In the previous tutorials, we used Kukicha's `on` syntax for methods:

```kukicha
# Kukicha style - English-like
func Display on todo Todo() string
    return "{todo.id}. {todo.title}"
```

This is the **only** method syntax Kukicha supports. When you read Go code, you'll see a different syntax (`func (todo Todo) Display() string`), but in Kukicha that maps directly to the `on` form. The translation table at the end of this tutorial covers the full mapping.

For this tutorial, we'll use `on`-style receivers throughout — the same pattern you learned in the Console Todo tutorial, but now with pointer receivers for mutation.

### Understanding `reference` vs `reference of`

As you read through the code, you'll see two pointer-related keywords:
- **`reference Type`** - Declares a pointer type (e.g., `reference Server` means "pointer to Server")
- **`reference of value`** - Takes the address of an existing value (e.g., `reference of server` converts `server` into a pointer)

Both are correct Kukicha syntax; they're just used in different contexts (declarations vs. operations).

---

## Part 2: Creating a Server Type

Instead of global variables, let's create a proper `Server` type that holds all our state:

```kukicha
import "sync"
import "slices"

type Server
    todos list of Todo
    nextId int
    mu sync.RWMutex  # A lock for safe access
```

**What's a `sync.RWMutex`?**

It's a "read-write lock" that prevents data corruption:
- **Read Lock** (`RLock`) - Multiple readers can access at once
- **Write Lock** (`Lock`) - Only one writer at a time, blocks everyone else

Think of it like a library book:
- Many people can read the same book at once
- But if someone is writing in it, everyone else has to wait

### Why We Wrap State in a Struct

Instead of using global variables (like we did in the web tutorial), we encapsulate all server state in the `Server` type. This design choice enables:
- **Testability** - You can create multiple test instances with different states
- **Dependency injection** - Pass the server instance where needed instead of relying on globals
- **Concurrency safety** - The mutex lives with the data it protects
- **Composability** - Future features can be added as new fields without touching global state

---

## Part 3: Thread-Safe Methods

Now let's write methods that use locking:

```kukicha
# Get all todos - uses a read lock
func GetAllTodos on s reference Server() list of Todo
    s.mu.RLock()           # Start reading
    defer s.mu.RUnlock()   # Unlock when done (even if there's an error)

    # Return a copy using standard slices.Clone
    return slices.Clone(s.todos)

# Create a new todo - uses a write lock
func CreateTodo on s reference Server(title string) Todo
    s.mu.Lock()           # Start writing (exclusive access)
    defer s.mu.Unlock()   # Unlock when done

    todo := Todo{id: s.nextId, title: title, completed: false}

    s.nextId = s.nextId + 1
    s.todos = append(s.todos, todo)

    return todo
```

**Why `reference Server`?**

We use `reference` (a pointer) because:
1. We need to **modify** the server's data
2. Locking only works if everyone uses the **same** lock

---

## Part 4: Adding a Database

Let's store our todos in SQLite so they persist across restarts.

### Installing the Driver

First, you need the SQLite driver:
```bash
go get github.com/mattn/go-sqlite3
```

### Database Helper Type

```kukicha
import "database/sql"
import _ "github.com/mattn/go-sqlite3"

type Database
    db reference sql.DB

# Open the database and create the table if needed
func OpenDatabase(filename string) (Database, error)
    db, err := sql.Open("sqlite3", filename)
    if err not equals empty
        return empty, err

    # Create the todos table
    createTable := `
        CREATE TABLE IF NOT EXISTS todos (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            title TEXT NOT NULL,
            completed BOOLEAN DEFAULT FALSE
        )
    `
    db.Exec(createTable) onerr return empty, error "{error}"

    return Database{db: db}, empty

# Close the database
func Close on d Database()
    # empty is equivalent to nil in Go
    if d.db not equals empty
        d.db.Close()
```

### CRUD Operations

```kukicha
# Create a new todo
func CreateTodo on d Database(title string) (Todo, error)
    result, execErr := d.db.Exec("INSERT INTO todos (title, completed) VALUES (?, FALSE)", title)
    if execErr not equals empty
        return Todo{}, execErr

    # Note: 'onerr' creates an implicit 'error' variable. If we used it again here,
    # it would "shadow" (hide) the outer error from the function signature.
    # Always explicitly handle or propagate the error to avoid confusion.
    id, idErr := result.LastInsertId()
    if idErr not equals empty
        return Todo{}, idErr

    return Todo{id: int(id), title: title, completed: false}, empty

# Get all todos
func GetAllTodos on d Database() (list of Todo, error)
    # Returns a cloned slice to prevent external mutations
    # Note: This is a shallow copy; deep copy needed if fields become reference types
    rows, queryErr := d.db.Query("SELECT id, title, completed FROM todos")
    if queryErr not equals empty
        return empty, queryErr
    defer rows.Close()

    todos := empty list of Todo

    for rows.Next()
        todo := Todo{}
        scanErr := rows.Scan(reference of todo.id, reference of todo.title, reference of todo.completed)
        if scanErr not equals empty
            continue
        todos = append(todos, todo)

    return todos, empty

# Get a single todo by ID
func GetTodo on d Database(id int) (Todo, error)
    row := d.db.QueryRow("SELECT id, title, completed FROM todos WHERE id = ?", id)

    todo := Todo{}
    scanErr := row.Scan(reference of todo.id, reference of todo.title, reference of todo.completed)
    if scanErr not equals empty
        return Todo{}, scanErr

    return todo, empty

# Update a todo
func UpdateTodo on d Database(id int, title string, completed bool) error
    d.db.Exec("UPDATE todos SET title = ?, completed = ? WHERE id = ?", title, completed, id) onerr return error "{error}"
    return empty

# Delete a todo
func DeleteTodo on d Database(id int) error
    d.db.Exec("DELETE FROM todos WHERE id = ?", id) onerr return error "{error}"
    return empty
```

---

## Part 5: The Production Server

Now let's put it all together into a production-ready server:

```kukicha
# Standard library - core functionality
import "fmt"
import "log"
import "net/http"
import "strconv"
import "sync"

# Standard library - data structures and encoding
import "slices"
import "database/sql"
import "encoding/json/v2"

# Kukicha stdlib
import "stdlib/string"
import "stdlib/validate"
import "stdlib/http" as httphelper
import "stdlib/must"
import "stdlib/env"

# Third-party packages
import _ "github.com/mattn/go-sqlite3"   # SQLite driver (requires CGO)

# --- Types ---

type Todo
    id int
    title string
    completed bool

type Server
    db Database

type ErrorResponse
    err string json:"error"

type CreateTodoInput
    title string

type UpdateTodoInput
    title string
    completed bool

# --- Server Constructor ---

func NewServer(dbPath string) (reference Server, error)
    db, dbErr := OpenDatabase(dbPath)
    if dbErr not equals empty
        return empty, dbErr

    server := Server{db: db}

    return reference of server, empty

# --- HTTP Handlers ---

func HandleTodos on s reference Server(w http.ResponseWriter, r reference http.Request)
    if r.URL.Path equals "/todos"
        if r.Method equals "GET"
            s.handleListTodos(w, r)
        else if r.Method equals "POST"
            s.handleCreateTodo(w, r)
        else
            s.sendError(w, 405, "Method not allowed")
    else
        if r.Method equals "GET"
            s.handleGetTodo(w, r)
        else if r.Method equals "PUT"
            s.handleUpdateTodo(w, r)
        else if r.Method equals "DELETE"
            s.handleDeleteTodo(w, r)
        else
            s.sendError(w, 405, "Method not allowed")

func handleListTodos on s reference Server(w http.ResponseWriter, r reference http.Request)
    todos, err := s.db.GetAllTodos()
    if err not equals empty
        log.Printf("Error fetching todos: %v", err)
        s.sendError(w, 500, "Failed to fetch todos")
        return

    s.sendJSON(w, 200, todos)

func handleCreateTodo on s reference Server(w http.ResponseWriter, r reference http.Request)
    # Parse request body using the http helper
    input := CreateTodoInput{}
    readErr := httphelper.ReadJSON(r, reference of input)
    if readErr not equals empty
        httphelper.JSONBadRequest(w, "Invalid JSON")
        return

    # Validate input using the validate package
    _, titleErr := input.title |> validate.NotEmpty()
    if titleErr not equals empty
        httphelper.JSONBadRequest(w, "Title is required")
        return

    _, lenErr := input.title |> validate.MaxLength(200)
    if lenErr not equals empty
        httphelper.JSONBadRequest(w, "Title must be 200 characters or less")
        return

    todo, createErr := s.db.CreateTodo(input.title)
    if createErr not equals empty
        log.Printf("Error creating todo: %v", createErr)
        httphelper.JSONInternalError(w, "Failed to create todo")
        return

    httphelper.JSONCreated(w, todo)

func handleGetTodo on s reference Server(w http.ResponseWriter, r reference http.Request)
    id, idErr := s.getIdFromPath(r.URL.Path, "/todos/")
    if idErr not equals empty
        s.sendError(w, 400, "Invalid ID")
        return

    todo, err := s.db.GetTodo(id)
    if err not equals empty
        s.sendError(w, 404, "Todo not found")
        return

    s.sendJSON(w, 200, todo)

func handleUpdateTodo on s reference Server(w http.ResponseWriter, r reference http.Request)
    id, idErr := s.getIdFromPath(r.URL.Path, "/todos/")
    if idErr not equals empty
        s.sendError(w, 400, "Invalid ID")
        return

    # Parse request body using pipe
    # Note: This is a full update (PUT). For partial updates, use PATCH with optional fields
    input := UpdateTodoInput{}
    decodeErr := r.Body
        |> json.NewDecoder()
        |> json.Decode(_, reference of input)
    if decodeErr not equals empty
        s.sendError(w, 400, "Invalid JSON")
        return

    updateErr := s.db.UpdateTodo(id, input.title, input.completed)
    if updateErr not equals empty
        log.Printf("Error updating todo: %v", updateErr)
        s.sendError(w, 500, "Failed to update todo")
        return

    # Fetch the updated todo to return
    todo, getErr := s.db.GetTodo(id)
    if getErr not equals empty
        s.sendError(w, 404, "Todo not found")
        return

    s.sendJSON(w, 200, todo)

func handleDeleteTodo on s reference Server(w http.ResponseWriter, r reference http.Request)
    id, idErr := s.getIdFromPath(r.URL.Path, "/todos/")
    if idErr not equals empty
        s.sendError(w, 400, "Invalid ID")
        return

    # In a production API, you might check if the todo was actually deleted (rows affected)
    # and return 404 if not found for better idempotency semantics
    deleteErr := s.db.DeleteTodo(id)
    if deleteErr not equals empty
        log.Printf("Error deleting todo: %v", deleteErr)
        s.sendError(w, 500, "Failed to delete todo")
        return

    w.WriteHeader(204)

# --- Helper Methods ---

func getIdFromPath on s reference Server(path string, prefix string) (int, error)
    # When called with manual error check, any error returned lets the caller decide
    # what to do — e.g., send a 400 response if path is invalid
    idStr := path |> string.TrimPrefix(prefix)
    if idStr equals "" or idStr equals path
        return 0, fmt.Errorf("invalid path")
    id := idStr |> strconv.Atoi() onerr return 0, error "{error}"
    return id, empty

func sendJSON on s reference Server(w http.ResponseWriter, status int, data any)
    # Set header before WriteHeader; after WriteHeader, header changes are ignored
    # The Encode call writes the response body; any error after WriteHeader can't change status
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    w |> json.NewEncoder() |> .Encode(data) onerr return

**💡 Tip:** Note the use of `.Encode(data)` above. The dot shorthand keeps the focus on the data being piped — even when calling methods!

func sendError on s reference Server(w http.ResponseWriter, status int, message string)
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    w |> json.NewEncoder() |> .Encode(ErrorResponse{err: message}) onerr return

# --- Main Entry Point ---

func main()
    # Configuration from environment variables (production best practice)
    # Using the 'must' package for startup config - panic is acceptable here

    # Required: DATABASE_URL must be set
    dbPath := must.EnvOr("DATABASE_URL", "todos.db")

    # Optional: PORT with default
    port := env.GetOr("PORT", "8080")

    # Create the server
    server := NewServer(dbPath) onerr panic "Failed to open database: {error}"

    # Ensures the SQLite file is closed when the program exits
    defer server.db.Close()

    # Register routes
    # Trailing slash catches /todos/1 and other ID-based routes
    http.HandleFunc("/todos", server.HandleTodos)
    http.HandleFunc("/todos/", server.HandleTodos)

    # Start server
    log.Printf("Server starting on http://localhost%s", port)
    log.Printf("Database: %s", dbPath)

    # Note: In production, capture SIGINT (Ctrl+C) and call http.Server.Shutdown()
    # for graceful shutdown instead of relying on panic
    http.ListenAndServe(port, empty) onerr panic "Server failed: {error}"
```

---

## Part 6: What Changed?

Let's compare the web tutorial version with this production version:

| Aspect | Web Tutorial | Production |
|--------|--------------|------------|
| **Storage** | In-memory `TodoStore` | SQLite database |
| **Safety** | None | Database handles concurrency |
| **Method style** | Methods on `TodoStore` | Methods on `Server` with DB |
| **Error handling** | Basic | Logging + proper responses |
| **Lifecycle** | None | `NewServer()` constructor, `Close()` cleanup |
| **State** | `TodoStore` type | `Server` type with `Database` |

---

## Part 7: Go Conventions You've Learned

### Pointer Receivers

```kukicha
# Kukicha: "reference Server"  →  Go: "*Server"
func Method on s reference Server()
```

Use pointer receivers when:
- The method modifies the receiver
- The receiver is large (avoids copying)
- You need consistency (if one method needs a pointer, use pointers for all)

### Constructors

Go doesn't have constructors, so we use functions named `New<Type>`:

```kukicha
func NewServer(config string) (reference Server, error)
    # Initialize and return
```

### Defer for Cleanup

```kukicha
func DoWork() error
    resource := Acquire() onerr return error "{error}"
    defer resource.Close()  # Guaranteed to run when function exits

    # Do work...
    return empty
```

### Error Wrapping

```kukicha
# Kukicha makes this easy with onerr
result := operation() onerr return fmt.Errorf("operation failed: {error}")
```

---

## Part 8: Production-Ready Packages

Kukicha includes several packages designed for production code:

### Configuration with `env` and `must`

```kukicha
import "stdlib/env"
import "stdlib/must"

func main()
    # Required config (panic if missing)
    apiKey := must.Env("API_KEY")

    # Optional config with defaults
    port := env.GetOr("PORT", "8080")
    debug := env.GetBoolOrDefault("DEBUG", false)
    timeout := env.GetIntOr("TIMEOUT", 30) onerr 30

    # Parse configuration from other sources (not just environment variables)
    # These utilities are useful when reading config files, command-line flags, etc.

    # Example: Parsing boolean from a config file value
    configValue := "yes"  # Could come from a file, database, etc.
    enabled := env.ParseBool(configValue) onerr false
    # Accepts: "true", "false", "1", "0", "yes", "no", "on", "off"

    # Example: Parsing a comma-separated list from any source
    serverList := "prod-1.example.com, prod-2.example.com,  prod-3.example.com"
    servers := env.SplitAndTrim(serverList, ",")
    # Returns: ["prod-1.example.com", "prod-2.example.com", "prod-3.example.com"]
    # Automatically trims whitespace and removes empty entries

    print("Servers: {len(servers)} configured")
```

**Why use these utilities?**
- `env.ParseBool()` handles multiple formats (true/false, yes/no, 1/0, on/off)
- `env.SplitAndTrim()` combines splitting and trimming in one operation
- Though they're in the `env` package, they're general-purpose utilities
- Useful for parsing configuration from any source (files, databases, APIs)

### Input Validation with `validate`

```kukicha
import "stdlib/validate"
import "stdlib/env"

func CreateUser(email string, age int) (User, error)
    # Chain validations - each returns (value, error) for onerr
    email |> validate.NotEmpty() onerr return User{}, error "{error}"
    email |> validate.Email() onerr return User{}, error "{error}"

    age |> validate.InRange(18, 120) onerr return User{}, error "{error}"

    return User{email: email, age: age}, empty

# Example: Parsing and validating a boolean setting from user input
func UpdateSettings(enableNotificationsStr string) (Settings, error)
    # First parse the boolean string, then use it
    enableNotifications := env.ParseBool(enableNotificationsStr) onerr
        return Settings{}, error "enableNotifications must be true/false/yes/no/1/0"

    return Settings{notifications: enableNotifications}, empty

# Example: Parsing and validating a list of email addresses
func AddAllowedEmails(emailListStr string) error
    # Split and trim the list
    emails := env.SplitAndTrim(emailListStr, ",")

    # Validate each email
    for email in emails
        email |> validate.Email() onerr
            return error "invalid email: {email}"

    return empty
```

**Combining utilities with validation:**
- Use `env.ParseBool()` to handle various boolean formats before validation
- Use `env.SplitAndTrim()` to clean up lists before validating each item
- These utilities make your validation code more robust and user-friendly

### HTTP Helpers

```kukicha
import "stdlib/http" as httphelper

func HandleUser(w http.ResponseWriter, r reference http.Request)
    # Read JSON body
    input := UserInput{}
    readErr := httphelper.ReadJSON(r, reference of input)
    if readErr not equals empty
        httphelper.JSONBadRequest(w, "Invalid JSON")
        return

    # Send JSON responses
    httphelper.JSON(w, user)              # 200 OK
    httphelper.JSONCreated(w, user)       # 201 Created
    httphelper.JSONNotFound(w, "User not found")  # 404

    # Query parameters
    page := httphelper.GetQueryIntOr(r, "page", 1)
    search := httphelper.GetQueryParam(r, "q")
```

---

## Summary: The Kukicha Learning Path

You've completed the full Kukicha tutorial series!

| Tutorial | What You Learned |
|----------|-----------------|
| ✅ **1. Beginner** | Variables, functions, strings, string petiole |
| ✅ **2. Console Todo** | Types, methods (`on`), default parameters, named arguments, lists, `onerr`, file I/O |
| ✅ **3. Web Todo** | HTTP servers, JSON, REST APIs |
| ✅ **4. Production** | Databases, Go conventions, validation, env config |

---

## Where to Go From Here

### Explore More

- **[Kukicha Grammar](../kukicha-grammar.ebnf.md)** - Complete language grammar
- **[Standard Library](../kukicha-stdlib-reference.md)** - iterator, slice, and more

### Build Projects

Ideas for your next project:
- **Blog Engine** - Posts, comments, user authentication
- **Chat Application** - WebSockets, real-time messaging
- **File Upload Service** - Handle file uploads and downloads
- **Task Queue** - Background job processing

### Learn More Go

Now that you know Kukicha, learning Go will be easy:
- [Go Tour](https://go.dev/tour/) - Official interactive tutorial
- [Effective Go](https://go.dev/doc/effective_go) - Go best practices
- [Go by Example](https://gobyexample.com/) - Practical examples

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
| `func Name on x Type` | `func (x Type) Name()` |
| `result onerr default` | `if err != nil { ... }` |
| `a \|> f(b)` | `f(a, b)` |
| `a \|> f(b, _)` | `f(b, a)` (placeholder) |

---

**Congratulations! You're now a Kukicha developer! 🎉🌱**
