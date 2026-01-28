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

## Part 1: Go-Style Method Receivers

In the previous tutorials, we used Kukicha's `on` syntax for methods:

```kukicha
# Kukicha style - English-like
func Display on todo Todo string
    return "{todo.id}. {todo.title}"
```

Go uses a different syntax that you'll see in most Go code:

```kukicha
# Go style - what you'll see in codebases
func (todo Todo) Display() string
    return "{todo.id}. {todo.title}"
```

Both compile to the same Go code! The Go style puts the receiver in parentheses before the function name.

### When to Use Which?

| Style | Best For |
|-------|----------|
| `func Name on receiver Type` | Learning, readability |
| `func (receiver Type) Name()` | Go compatibility, team projects |

For this tutorial, we'll use **Go-style receivers** since that's what you'll encounter in real Go projects.

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
func (s reference Server) GetAllTodos() list of Todo
    s.mu.RLock()           # Start reading
    defer s.mu.RUnlock()   # Unlock when done (even if there's an error)
    
    # Return a copy using standard slices.Clone
    return slices.Clone(s.todos)

# Create a new todo - uses a write lock
func (s reference Server) CreateTodo(title string) Todo
    s.mu.Lock()           # Start writing (exclusive access)
    defer s.mu.Unlock()   # Unlock when done
    
    todo := Todo
        id: s.nextId
        title: title
        completed: false
    
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
    db.Exec(createTable) onerr return empty, error
    
    return Database{db: db}, empty

# Close the database
func (d Database) Close()
    # empty is equivalent to nil in Go
    if d.db not equals empty
        d.db.Close()
```

### CRUD Operations

```kukicha
# Create a new todo
func (d Database) CreateTodo(title string) (Todo, error)
    result := d.db.Exec("INSERT INTO todos (title, completed) VALUES (?, FALSE)", title) 
        onerr return Todo{}, error
    
    # Note: 'onerr' creates an implicit 'error' variable. If we used it again here,
    # it would "shadow" (hide) the outer error from the function signature.
    # Always explicitly handle or propagate the error to avoid confusion.
    id := result.LastInsertId() onerr return Todo{}, error
    
    todo := Todo
        id: int(id)
        title: title
        completed: false
    return todo, empty

# Get all todos
func (d Database) GetAllTodos() (list of Todo, error)
    # Returns a cloned slice to prevent external mutations
    # Note: This is a shallow copy; deep copy needed if fields become reference types
    rows := d.db.Query("SELECT id, title, completed FROM todos") onerr
        return empty, error
    defer rows.Close()
    
    todos := empty list of Todo
    
    for rows.Next()
        todo := Todo{}
        rows.Scan(reference of todo.id, reference of todo.title, reference of todo.completed) onerr
            continue
        todos = append(todos, todo)
    
    return todos, empty

# Get a single todo by ID
func (d Database) GetTodo(id int) (Todo, error)
    row := d.db.QueryRow("SELECT id, title, completed FROM todos WHERE id = ?", id)
    
    todo := Todo{}
    row.Scan(reference of todo.id, reference of todo.title, reference of todo.completed) onerr
        return Todo{}, error
    
    return todo, empty

# Update a todo
func (d Database) UpdateTodo(id int, title string, completed bool) error
    d.db.Exec("UPDATE todos SET title = ?, completed = ? WHERE id = ?", title, completed, id) onerr
        return error
    return empty

# Delete a todo
func (d Database) DeleteTodo(id int) error
    d.db.Exec("DELETE FROM todos WHERE id = ?", id) onerr
        return error
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
    error string json:"error"

type CreateTodoInput
    title string

type UpdateTodoInput
    title string
    completed bool

# --- Server Constructor ---

func NewServer(dbPath string) (reference Server, error)
    db := OpenDatabase(dbPath) onerr return empty, error
    
    server := Server
        db: db
    
    return reference of server, empty

# --- HTTP Handlers (Go-style receivers) ---

func (s reference Server) HandleTodos(w http.ResponseWriter, r reference http.Request)
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

func (s reference Server) handleListTodos(w http.ResponseWriter, r reference http.Request)
    todos, err := s.db.GetAllTodos() onerr
        log.Printf("Error fetching todos: %v", err)
        s.sendError(w, 500, "Failed to fetch todos")
        return
    
    s.sendJSON(w, 200, todos)

func (s reference Server) handleCreateTodo(w http.ResponseWriter, r reference http.Request)
    # Parse request body using pipe
    input := CreateTodoInput{}

    r.Body |> json.NewDecoder() |> json.Decode(_, reference of input) onerr
        return s.sendError(w, 400, "Invalid JSON")

    if input.title equals ""
        s.sendError(w, 400, "Title is required")
        return
    
    todo := s.db.CreateTodo(input.title) onerr
        log.Printf("Error creating todo: %v", error)
        s.sendError(w, 500, "Failed to create todo")
        return
    
    s.sendJSON(w, 201, todo)

func (s reference Server) handleGetTodo(w http.ResponseWriter, r reference http.Request)
    id := s.getIdFromPath(r.URL.Path, "/todos/") onerr
        s.sendError(w, 400, "Invalid ID")
        return
    
    todo := s.db.GetTodo(id) onerr
        return s.sendError(w, 404, "Todo not found")
    
    s.sendJSON(w, 200, todo)

func (s reference Server) handleUpdateTodo(w http.ResponseWriter, r reference http.Request)
    id := s.getIdFromPath(r.URL.Path, "/todos/") onerr
        s.sendError(w, 400, "Invalid ID")
        return
    
    # Parse request body using pipe
    # Note: This is a full update (PUT). For partial updates, use PATCH with optional fields
    input := UpdateTodoInput{}
    
    r.Body |> json.NewDecoder() |> json.Decode(_, reference of input) onerr
        return s.sendError(w, 400, "Invalid JSON")
    
    s.db.UpdateTodo(id, input.title, input.completed) onerr
        log.Printf("Error updating todo: %v", error)
        s.sendError(w, 500, "Failed to update todo")
        return
    
    # Fetch the updated todo to return
    todo := s.db.GetTodo(id) onerr
        return s.sendError(w, 404, "Todo not found")
    
    s.sendJSON(w, 200, todo)

func (s reference Server) handleDeleteTodo(w http.ResponseWriter, r reference http.Request)
    id := s.getIdFromPath(r.URL.Path, "/todos/") onerr
        s.sendError(w, 400, "Invalid ID")
        return
    
    # In a production API, you might check if the todo was actually deleted (rows affected)
    # and return 404 if not found for better idempotency semantics
    s.db.DeleteTodo(id) onerr
        log.Printf("Error deleting todo: %v", error)
        s.sendError(w, 500, "Failed to delete todo")
        return
    
    w.WriteHeader(204)

# --- Helper Methods ---

func (s reference Server) getIdFromPath(path string, prefix string) (int, error)
    # When called with onerr, any error returned triggers the error handler block
    # E.g., id := s.getIdFromPath(...) onerr { ... } executes the block if path is invalid
    idStr := path |> string.TrimPrefix(prefix)
    if idStr equals "" or idStr equals path
        return 0, fmt.Errorf("invalid path")
    return idStr |> strconv.Atoi()

func (s reference Server) sendJSON(w http.ResponseWriter, status int, data any)
    # Set header before WriteHeader; after WriteHeader, header changes are ignored
    # The Encode call writes the response body; any error after WriteHeader can't change status
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    w |> json.NewEncoder() |> .Encode(data) onerr return

func (s reference Server) sendError(w http.ResponseWriter, status int, message string)
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    w |> json.NewEncoder() |> .Encode(ErrorResponse{error: message}) onerr return

# --- Main Entry Point ---

func main()
    # Configuration (could come from environment variables)
    dbPath := "todos.db"
    port := ":8080"
    
    # Create the server
    # Using panic for unrecoverable startup errors is acceptable here.
    # For more graceful error handling, use: log.Fatalf("Failed to open database: %v", error)
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
| **Storage** | Global `var todos` | SQLite database |
| **Safety** | None | Database handles concurrency |
| **Method style** | Standalone functions | `func (s reference Server)` |
| **Error handling** | Basic | Logging + proper responses |
| **Lifecycle** | None | `NewServer()` constructor, `Close()` cleanup |
| **State** | Global variables | Encapsulated in `Server` type |

---

## Part 7: Go Conventions You've Learned

### Pointer Receivers

```kukicha
# Kukicha: "reference Server"
# Go convention: "*Server"
func (s reference Server) Method()
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
    resource := Acquire() onerr return error
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

## Summary: The Kukicha Learning Path

You've completed the full Kukicha tutorial series!

| Tutorial | What You Learned |
|----------|-----------------|
| ✅ **1. Beginner** | Variables, functions, strings, string petiole |
| ✅ **2. Console Todo** | Types, methods (`on`), lists, `onerr`, file I/O |
| ✅ **3. Web Todo** | HTTP servers, JSON, REST APIs |
| ✅ **4. Production** | Databases, Go conventions, proper architecture |

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
