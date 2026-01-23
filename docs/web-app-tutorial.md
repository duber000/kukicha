# Building a Todo Web App with Kukicha

**Level:** Intermediate  
**Time:** 30 minutes  
**Goal:** Build a simple REST API for managing todos using Go's standard library, written entirely in Kukicha

In this tutorial, you'll learn how to:
- Use Go's `net/http` package from Kukicha
- Handle JSON encoding/decoding with `reference of`
- Build HTTP handlers and routes
- Run a web server
- Use error handling with `onerr`

---

## What We're Building

A simple REST API with these endpoints:

| Method | Path | Purpose |
|--------|------|---------|
| `GET` | `/todos` | List all todos |
| `POST` | `/todos` | Create a new todo |
| `GET` | `/todos/{id}` | Get a specific todo |
| `PUT` | `/todos/{id}` | Update a todo |
| `DELETE` | `/todos/{id}` | Delete a todo |

---

## Step 1: Project Setup

Create a new Kukicha project:

```bash
mkdir todo-app
cd todo-app
go mod init github.com/username/todo-app
```

Create `stem.toml`:

```toml
[stem]
name = "todo-app"
version = "0.1.0"
```

Create `main/main.kuki`:

```kukicha
petiole main

import "net/http"
import "encoding/json"
import "sync"

type Todo
    id int64
    title string
    completed bool

type Server
    todos list of Todo
    nextId int64
    mu reference sync.RWMutex

func main()
    server := Server
        todos: empty list of Todo
        nextId: 1
        mu: reference to sync.RWMutex{}
    
    http.HandleFunc("/todos", func(w http.ResponseWriter, r reference http.Request)
        if r.Method equals "GET"
            server.GetTodos(w, r)
        else if r.Method equals "POST"
            server.CreateTodo(w, r)
    )
    
    http.HandleFunc("/todos/", func(w http.ResponseWriter, r reference http.Request)
        if r.Method equals "GET"
            server.GetTodo(w, r)
        else if r.Method equals "PUT"
            server.UpdateTodo(w, r)
        else if r.Method equals "DELETE"
            server.DeleteTodo(w, r)
    )
    
    print "Server starting on :8080"
    http.ListenAndServe(":8080", empty) onerr panic "server error"

func (s reference Server) GetTodos(w http.ResponseWriter, r reference http.Request)
    dereference s.mu .RLock()
    defer dereference s.mu .RUnlock()
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(dereference s .todos) onerr return

func (s reference Server) CreateTodo(w http.ResponseWriter, r reference http.Request)
    todo := Todo{}
    json.NewDecoder(r.Body).Decode(reference of todo) onerr
        w.WriteHeader(400)
        return
    
    todo.id = dereference s .nextId
    dereference s .nextId = dereference s .nextId + 1
    
    dereference s .mu .Lock()
    dereference s .todos = append(dereference s .todos, todo)
    dereference s .mu .Unlock()
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(201)
    json.NewEncoder(w).Encode(todo) onerr return

func (s reference Server) GetTodo(w http.ResponseWriter, r reference http.Request)
    idStr := r.URL.Path[len("/todos/"):]
    id := atoi(idStr) onerr
        w.WriteHeader(404)
        return
    
    dereference s.mu .RLock()
    defer dereference s.mu .RUnlock()
    
    for todo in dereference s .todos
        if todo.id equals int64(id)
            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(todo) onerr return
            return
    
    w.WriteHeader(404)

func (s reference Server) UpdateTodo(w http.ResponseWriter, r reference http.Request)
    idStr := r.URL.Path[len("/todos/"):]
    id := atoi(idStr) onerr
        w.WriteHeader(404)
        return
    
    updated := Todo{}
    json.NewDecoder(r.Body).Decode(reference of updated) onerr
        w.WriteHeader(400)
        return
    
    dereference s.mu .Lock()
    defer dereference s.mu .Unlock()
    
    for i, todo in dereference s .todos
        if todo.id equals int64(id)
            updated.id = todo.id
            dereference s .todos[i] = updated
            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(updated) onerr return
            return
    
    w.WriteHeader(404)

func (s reference Server) DeleteTodo(w http.ResponseWriter, r reference http.Request)
    idStr := r.URL.Path[len("/todos/"):]
    id := atoi(idStr) onerr
        w.WriteHeader(404)
        return
    
    dereference s.mu .Lock()
    defer dereference s.mu .Unlock()
    
    for i, todo in dereference s .todos
        if todo.id equals int64(id)
            dereference s .todos = append(dereference s .todos[0:i], dereference s .todos[i+1:]...)
            w.WriteHeader(204)
            return
    
    w.WriteHeader(404)

func atoi(s string) (int, error)
    n := 0
    for c in s
        if c < "0" or c > "9"
            return 0, error("invalid number")
        n = n * 10 + int(c - "0")
    return n, empty
```

---

## Step 2: Build and Run

Build the application:

```bash
kukicha build main/main.kuki
```

Run the server:

```bash
./main
```

You should see:
```
Server starting on :8080
```

---

## Step 3: Test the API

### Create a Todo

```bash
curl -X POST http://localhost:8080/todos \
  -H "Content-Type: application/json" \
  -d '{"title":"Learn Kukicha","completed":false}'
```

Response:
```json
{"id":1,"title":"Learn Kukicha","completed":false}
```

### Get All Todos

```bash
curl http://localhost:8080/todos
```

Response:
```json
[{"id":1,"title":"Learn Kukicha","completed":false}]
```

### Update a Todo

```bash
curl -X PUT http://localhost:8080/todos/1 \
  -H "Content-Type: application/json" \
  -d '{"title":"Learn Kukicha","completed":true}'
```

### Delete a Todo

```bash
curl -X DELETE http://localhost:8080/todos/1
```

---

## Key Concepts Demonstrated

### 1. Using Go Stdlib Directly

```kukicha
import "net/http"
import "encoding/json"
import "sync"

# Use Go packages directly - no wrappers needed
http.HandleFunc("/todos", handler)
json.NewEncoder(w).Encode(data)
sync.RWMutex{}
```

### 2. Reference-of for Stdlib Functions

```kukicha
# json.Unmarshal needs a pointer to the destination
json.NewDecoder(r.Body).Decode(reference of todo)

# Lock operations on pointer fields
dereference s.mu .Lock()
```

### 3. Pointers for Mutable State

```kukicha
# Method receiver is a pointer to modify server state
func (s reference Server) CreateTodo(w http.ResponseWriter, r reference http.Request)
    # Modify todos list through pointer
    dereference s .todos = append(dereference s .todos, todo)
```

### 4. Error Handling with onerr

```kukicha
# Handle errors from stdlib functions
json.NewDecoder(r.Body).Decode(reference of todo) onerr
    w.WriteHeader(400)
    return

# Parse with error handling
id := atoi(idStr) onerr
    w.WriteHeader(404)
    return
```

### 5. Defer for Cleanup

```kukicha
# Locks are released even if function panics
dereference s.mu .RLock()
defer dereference s.mu .RUnlock()
```

---

## Real-World Enhancements

Once you understand the basics, you can add:

### 1. Persistent Storage (JSON file)

```kukicha
import "os"

func LoadTodos() list of Todo
    data := os.ReadFile("todos.json") onerr return empty list of Todo
    todos := empty list of Todo
    json.Unmarshal(data, reference of todos) onerr return empty list of Todo
    return todos

func SaveTodos(todos list of Todo)
    data := json.Marshal(todos) onerr return
    os.WriteFile("todos.json", data, 0644) onerr return
```

### 2. Database (Using `database/sql`)

```kukicha
import "database/sql"
import _ "github.com/mattn/go-sqlite3"

func NewDB() reference sql.DB
    db := sql.Open("sqlite3", "todos.db") onerr return empty
    return reference of db
```

### 3. Middleware

```kukicha
func LoggingMiddleware(next http.Handler) http.Handler
    return http.HandlerFunc(func(w http.ResponseWriter, r reference http.Request)
        print "{r.Method} {r.URL.Path}"
        next.ServeHTTP(w, r)
    )
```

### 4. Input Validation

```kukicha
func (todo reference Todo) Validate() (bool, string)
    if todo.title equals ""
        return false, "title is required"
    if len(todo.title) > 255
        return false, "title too long"
    return true, ""
```

---

## Important Notes

### Memory Management

This example uses in-memory storage, so todos are lost when the server stops. For production:
- Use a database (SQLite, PostgreSQL)
- Use file persistence
- Add proper transaction handling

### Concurrency

The example uses `sync.RWMutex` for thread-safe access. For higher performance:
- Consider using channels
- Use worker pools
- Implement proper request queuing

### Error Handling

This example is simplified. Production code should:
- Log errors properly
- Return meaningful error messages
- Handle partial failures gracefully
- Implement request validation

---

## Summary

You've learned how to:

✅ Import and use Go stdlib packages directly  
✅ Use `reference of` and `dereference` for pointer operations  
✅ Build HTTP handlers and routes  
✅ Handle JSON encoding/decoding  
✅ Manage concurrent access with mutexes  
✅ Use `onerr` for error handling  
✅ Write production-like Kukicha code  

This demonstrates Kukicha's core philosophy: **"It's just Go"** - you have direct access to the entire Go ecosystem without wrappers or special syntax.

---

## Next Steps

- Deploy to a cloud platform (Heroku, AWS, Google Cloud)
- Add authentication (JWT tokens)
- Connect to a real database
- Build a frontend (HTML/CSS/JavaScript)
- Add unit tests
- Deploy with Docker

Happy coding! 🌱
