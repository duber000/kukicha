# Building a Web Todo App with Kukicha

**Level:** Intermediate  
**Time:** 30 minutes  
**Prerequisite:** [CLI Explorer Tutorial](cli-explorer-tutorial.md)

Welcome! You've built interactive CLI tools with custom types, methods, and pipes. Now let's build something even cooler: a **web application** that you can access from a browser!

## What You'll Learn

In this tutorial, you'll discover how to:
- Create a **web server** that responds to requests
- Send and receive **JSON data** (the language of web APIs)
- Build **endpoints** for creating, reading, updating, and deleting todos
- Handle **different request types** (GET, POST, PUT, DELETE)

By the end, you'll have a todo API that any web or mobile app could connect to!

---

## What We're Building

Our web app will be a **REST API** - a way for other programs (like websites or phone apps) to talk to our todo list.

| Action | Request | URL | Description |
|--------|---------|-----|-------------|
| List all todos | `GET` | `/todos` | Get all your todos |
| Create a todo | `POST` | `/todos` | Add a new todo |
| Get one todo | `GET` | `/todos/1` | Get todo with id 1 |
| Update a todo | `PUT` | `/todos/1` | Update todo with id 1 |
| Delete a todo | `DELETE` | `/todos/1` | Delete todo with id 1 |

**Why a web API?** A web API lets any device—browser, mobile app, or script—talk to the same todo data, which is a natural next step after a local CLI. Your todo list becomes accessible anywhere, not just from the terminal.

Don't worry if this looks complicated - we'll build it step by step!

---

## Step 0: Project Setup

If you haven't already, set up your project:

```bash
mkdir web-todo && cd web-todo
go mod init web-todo
kukicha init    # Extracts stdlib for JSON, fetch, etc.
```

---

## Step 1: Your First Web Server

Let's start with the simplest possible web server:

```kukicha
import "fmt"
import "net/http"

function main()
    # When someone visits the homepage, say hello
    http.HandleFunc("/", sayHello)
    
    print("Server starting on http://localhost:8080")
    http.ListenAndServe(":8080", empty) onerr panic "server failed to start"

# This function handles requests to "/"
function sayHello(response http.ResponseWriter, request reference http.Request)
    # Use pipe to send response!
    response |> fmt.Fprintln("Hello from Kukicha!")
```

**What's happening here?**

1. `http.HandleFunc("/", sayHello)` - When someone visits `/`, run the `sayHello` function
2. `http.ListenAndServe(":8080", empty)` - Start listening on port 8080
3. `sayHello` receives two things:
   - `response` - Where we write our reply
   - `request` - Information about what the user asked for

**Try it!**

Run the server:
```bash
kukicha run main.kuki
```

Then open your browser to `http://localhost:8080` - you should see "Hello from Kukicha!"

---

## Step 2: Understanding Handlers

A **handler** is a function that responds to web requests. Every handler receives:

```kukicha
function myHandler(response http.ResponseWriter, request reference http.Request)
    # response - write your reply here
    # request - contains info about the incoming request
```

We can check what **method** (GET, POST, etc.) the user is using:

```kukicha
function myHandler(response http.ResponseWriter, request reference http.Request)
    if request.Method equals "GET"
        response |> fmt.Fprintln("You used GET!")
    else if request.Method equals "POST"
        response |> fmt.Fprintln("You used POST!")
    else
        response |> fmt.Fprintln("You used something else!")
```

---

## Step 3: Sending JSON Responses

Web APIs typically send data as **JSON** (JavaScript Object Notation). It looks like this:

```json
{"id": 1, "title": "Buy groceries", "completed": false}
```

> **📚 Note: JSON in Kukicha with Go 1.25+**
>
> Kukicha's stdlib `json` package uses Go 1.25+ `encoding/json/v2` for 2-10x faster JSON parsing:
> ```kukicha
> import "stdlib/json"
> 
> # Parse JSON from file
> configData := files.Read("config.json")
> config := Config{}
> json.Unmarshal(configData as list of byte, reference config) onerr panic "parse failed"
> 
> # Or using the pipe pattern:
> config := Config{}
> files.Read("config.json") |> json.Unmarshal(_, reference config) onerr panic "parse failed"
> ```
>
> For web servers, you can use `encoding/json/v2` for streaming:
> ```kukicha
> import "encoding/json/v2"
> json.MarshalWrite(response, data)  # Write JSON directly to response
> json.UnmarshalRead(request.Body, reference result)  # Stream from request
>
> # With pipe placeholder (_), you can pipe data into any argument position:
> data |> json.MarshalWrite(response, _)  # _ marks where piped value goes
> ```
>
> **Rule of thumb:** Use `json.Unmarshal()` for parsing JSON data, use `json.Marshal()` for creating JSON, and use the streaming functions for web servers.

Kukicha makes sending JSON easy:

```kukicha
import "encoding/json/v2"  # Go 1.25+ for better performance

type Todo
    id int
    title string
    completed bool

function sendTodo(response http.ResponseWriter, request reference http.Request)
    # Create a todo with indented syntax
    todo := Todo
        id: 1
        title: "Learn Kukicha"
        completed: false
    
    # Tell the browser we're sending JSON using pipe chaining
    response |> .Header() |> .Set("Content-Type", "application/json")
    
    # Convert the todo to JSON and send it using pipe
    response |> json.NewEncoder() |> .Encode(todo) onerr return
```

**💡 Tip:** When piping into a method that belongs to the value itself, use the dot shorthand:
```kukicha
# Calling directly:
response.Header().Set(...)

# Same thing, using pipe:
response |> .Header() |> .Set(...)
```
This keeps the left-to-right data flow when chaining — and makes it clear the method belongs to the piped value, not an imported package.

When someone visits this endpoint, they'll receive:
```json
{"id":1,"title":"Learn Kukicha","completed":false}
```

---

## Step 4: Receiving JSON Data

When creating a new todo, the user sends JSON data to us. We need to read and parse it:

```kukicha
function createTodo(response http.ResponseWriter, request reference http.Request)
    # Create an empty todo to fill with the incoming data
    todo := Todo{}

    # Parse the JSON from the request body — pipe it through decoder
    # "reference of" gets a pointer so the decoder can fill in our todo
    decodeErr := request.Body
        |> json.NewDecoder()
        |> json.Decode(_, reference of todo)
    if decodeErr not equals empty
        response |> .WriteHeader(400)
        response |> fmt.Fprintln("Invalid JSON")
        return

    # Now 'todo' contains the data the user sent!
    print("Received todo: {todo.title}")

    # Send back a success response
    response |> .Header() |> .Set("Content-Type", "application/json")
    response |> .WriteHeader(201)  # 201 = Created
    response |> json.NewEncoder() |> .Encode(todo) onerr return
```

**What's `reference of`?**

When we write `reference of todo`, we're giving the JSON decoder a way to **fill in** our todo variable. Without it, the decoder would only have a copy and couldn't modify our actual todo.

This same idea shows up when you work with lists: looping over items gives you **copies**, so updating a todo requires mutating by index (or using a reference) rather than changing a loop variable. You'll see us find a todo's index and update it directly in later steps.

> **💡 Tip: The `_` placeholder.** By default, the piped value becomes the first argument. Use `_` to place it elsewhere:
> ```kukicha
> # Default: piped value is the first argument
> text |> string.ToLower()                      # → string.ToLower(text)
>
> # With _: piped value goes where _ is
> data |> json.MarshalWrite(response, _)        # → json.MarshalWrite(response, data)
> ```
> You'll see this pattern throughout this tutorial — it keeps data flowing left-to-right even when the API expects the value in a later position.

---

## Step 5: Building the Todo Storage

Let's create a type to hold our todo list and the next available ID. Wrapping state in a type keeps things organized — and as a bonus, we can pass our store to HTTP handlers using **method values** (we'll see that in Step 6):

```kukicha
# Our todo storage - bundled into a type
type TodoStore
    todos list of Todo
    nextId int
```

For finding todos by ID, we'll walk the list and return the index. We can get the todo from the index if needed — this avoids returning multiple values when we only need the index for updates and deletes.

---

## Step 6: The Complete Todo API

Now let's put it all together! Create `main.kuki`:

```kukicha
import "net/http"
import "encoding/json/v2"  # Go 1.25+ for better performance
import "strconv"
import "stdlib/string"

# --- Data Types ---

type Todo
    id int
    title string
    completed bool

# ErrorResponse is what the client receives when something goes wrong.
# The json tag maps the "err" field to the JSON key "error".
type ErrorResponse
    err string json:"error"

# Metadata can be used to store extra info (demonstrating map literals)
variable API_METADATA = map of string to string{
    "version": "1.0",
    "environment": "development",
}

# --- Store ---
# (In the Production tutorial, we'll use a database instead)

type TodoStore
    todos list of Todo
    nextId int

# --- Helper Functions ---

function findTodoIndex on store reference TodoStore(id int) int
    # Walk the list to find the matching ID
    for i, todo in store.todos
        if todo.id equals id
            return i
    return -1

function getIdFromPath(path string, prefix string) (int, bool)
    # Extract "1" from "/todos/1"
    idStr := path |> string.TrimPrefix(prefix)
    if idStr equals "" or idStr equals path
        return 0, false

    id := idStr |> strconv.Atoi() onerr return 0, false
    return id, true

function sendJSON on store reference TodoStore(response http.ResponseWriter, data any)
    # Helper to send any data as JSON with correct content-type header
    response |> .Header() |> .Set("Content-Type", "application/json")
    response |> json.NewEncoder() |> .Encode(data) onerr return

function sendError on store reference TodoStore(response http.ResponseWriter, status int, message string)
    # Helper to send error responses as JSON
    # The client will receive a JSON object like {"error":"message"} with the given status code
    response |> .Header() |> .Set("Content-Type", "application/json")
    response |> .WriteHeader(status)
    response |> json.NewEncoder() |> .Encode(ErrorResponse{err: message}) onerr return
```

> **💡 Pro Tip:** In production code, use `stdlib/http` helpers instead of writing these manually:
> ```kukicha
> import "stdlib/http" as httphelper
> httphelper.JSON(response, todo)           # Send JSON with correct headers
> httphelper.JSONError(response, 400, "...")  # Send error as JSON
> httphelper.ReadJSON(request, reference todo)  # Parse request body
> ```
> See the [Production Patterns Tutorial](production-patterns-tutorial.md) for more examples.

```kukicha
# --- API Handlers ---

# GET /todos - List all todos (with optional search)
function handleListTodos on store reference TodoStore(response http.ResponseWriter, request reference http.Request)
    # Note: In production, you'd add limit/offset query parameters for pagination
    # to handle large datasets efficiently
    search := request.URL.Query().Get("search")
    if search equals ""
        store.sendJSON(response, store.todos)
        return

    # Filter todos that contain the search string (case-insensitive)
    # Note: This is a simple substring match.
    # For production, regex or full-text search would be better for larger datasets
    filtered := empty list of Todo
    for todo in store.todos
        if todo.title |> string.ToLower() |> string.Contains(search |> string.ToLower())
            filtered = append(filtered, todo)

    store.sendJSON(response, filtered)

# POST /todos - Create a new todo
function handleCreateTodo on store reference TodoStore(response http.ResponseWriter, request reference http.Request)
    # Parse the incoming JSON — pipe through decoder
    # reference of todo gives the decoder a pointer so it can fill in the struct
    todo := Todo{}
    decodeErr := request.Body
        |> json.NewDecoder()
        |> json.Decode(_, reference of todo)
    if decodeErr not equals empty
        store.sendError(response, 400, "Invalid JSON")
        return

    # Validate
    if todo.title equals ""
        store.sendError(response, 400, "Title is required")
        return

    # Assign an ID and add to the list
    todo.id = store.nextId
    store.nextId = store.nextId + 1
    store.todos = append(store.todos, todo)

    # Send back the created todo
    response |> .WriteHeader(201)
    store.sendJSON(response, todo)

# GET /todos/{id} - Get a specific todo
function handleGetTodo on store reference TodoStore(response http.ResponseWriter, request reference http.Request)
    # Get the ID from the URL
    id, ok := getIdFromPath(request.URL.Path, "/todos/")
    if not ok
        store.sendError(response, 400, "Invalid todo ID")
        return

    # Find the todo
    index := store.findTodoIndex(id)
    if index equals -1
        store.sendError(response, 404, "Todo not found")
        return

    store.sendJSON(response, store.todos[index])

# PUT /todos/{id} - Update a todo
function handleUpdateTodo on store reference TodoStore(response http.ResponseWriter, request reference http.Request)
    # Get the ID from the URL
    id, ok := getIdFromPath(request.URL.Path, "/todos/")
    if not ok
        store.sendError(response, 400, "Invalid todo ID")
        return

    # Find the todo
    index := store.findTodoIndex(id)
    if index equals -1
        store.sendError(response, 404, "Todo not found")
        return

    # Parse the update — pipe through decoder
    updated := Todo{}
    updateErr := request.Body
        |> json.NewDecoder()
        |> json.Decode(_, reference of updated)
    if updateErr not equals empty
        store.sendError(response, 400, "Invalid JSON")
        return

    # Keep the original ID, update other fields
    updated.id = id
    store.todos[index] = updated

    store.sendJSON(response, updated)

# DELETE /todos/{id} - Delete a todo
function handleDeleteTodo on store reference TodoStore(response http.ResponseWriter, request reference http.Request)
    # Get the ID from the URL
    id, ok := getIdFromPath(request.URL.Path, "/todos/")
    if not ok
        store.sendError(response, 400, "Invalid todo ID")
        return

    # Find the todo
    index := store.findTodoIndex(id)
    if index equals -1
        store.sendError(response, 404, "Todo not found")
        return

    # Remove by creating a new list without this item
    # 'many' unpacks the second slice so append sees individual elements
    store.todos = append(store.todos[:index], many store.todos[index+1:])

    response |> .WriteHeader(204)  # 204 = No Content (success, nothing to return)

# --- Route Handler ---

function handleTodos on store reference TodoStore(response http.ResponseWriter, request reference http.Request)
    # Route to the right handler based on the path and method

    if request.URL.Path equals "/todos"
        # Collection routes: /todos
        if request.Method equals "GET"
            store.handleListTodos(response, request)
        else if request.Method equals "POST"
            store.handleCreateTodo(response, request)
        else
            store.sendError(response, 405, "Method not allowed")
    else
        # Item routes: /todos/{id}
        if request.Method equals "GET"
            store.handleGetTodo(response, request)
        else if request.Method equals "PUT"
            store.handleUpdateTodo(response, request)
        else if request.Method equals "DELETE"
            store.handleDeleteTodo(response, request)
        else
            store.sendError(response, 405, "Method not allowed")

# --- Main Entry Point ---

function main()
    # Create the store with an empty todo list using indented literal
    store := TodoStore
        todos: empty list of Todo
        nextId: 1

    # Set up routes — method values let us pass a method as a handler function
    http.HandleFunc("/todos", store.handleTodos)
    http.HandleFunc("/todos/", store.handleTodos)  # Trailing slash catches /todos/1

    print("=== Kukicha Todo API ===")
    print("Server running on http://localhost:8080")
    print("API endpoint: http://localhost:8080/todos")
    print("")
    print("Try these commands in another terminal:")
    print("  curl http://localhost:8080/todos")
    print("  curl -X POST -d '{\"title\":\"Learn Kukicha\"}' http://localhost:8080/todos")
    print("")

    # Note: In production, capture SIGINT and call http.Server.Shutdown()
    # for graceful shutdown instead of relying on panic
    http.ListenAndServe(":8080", empty) onerr panic "server failed to start"
```

---

## Step 7: Testing Your API

Run your server:
```bash
kukicha run main.kuki
```

Now test it with `curl` (open another terminal):

### Create todos:
```bash
curl -X POST -H "Content-Type: application/json" \
     -d '{"title":"Buy groceries"}' http://localhost:8080/todos
# Response: {"id":1,"title":"Buy groceries","completed":false}

curl -X POST -H "Content-Type: application/json" \
     -d '{"title":"Learn Kukicha"}' http://localhost:8080/todos
# Response: {"id":2,"title":"Learn Kukicha","completed":false}
```

### List all todos:
```bash
curl http://localhost:8080/todos
# Response: [{"id":1,"title":"Buy groceries","completed":false},{"id":2,"title":"Learn Kukicha","completed":false}]
```

### Search for todos:
```bash
curl "http://localhost:8080/todos?search=Kukicha"
# Response: [{"id":2,"title":"Learn Kukicha","completed":false}]
```

### Get a specific todo:
```bash
curl http://localhost:8080/todos/1
# Response: {"id":1,"title":"Buy groceries","completed":false}
```

### Update a todo:
```bash
curl -X PUT -H "Content-Type: application/json" \
     -d '{"title":"Buy groceries","completed":true}' http://localhost:8080/todos/1
# Response: {"id":1,"title":"Buy groceries","completed":true}
```

### Delete a todo:
```bash
curl -X DELETE http://localhost:8080/todos/2
# Response: (empty - 204 No Content)
```

🎉 **Congratulations!** You've built a working REST API!

---

## Understanding HTTP Status Codes

You may have noticed we use numbers like `200`, `201`, `404`. These are **status codes** that tell the client what happened:

| Code | Name | Meaning |
|------|------|---------|
| `200` | OK | Success! |
| `201` | Created | Successfully created something new |
| `204` | No Content | Success, but nothing to return |
| `400` | Bad Request | The client sent invalid data |
| `404` | Not Found | The requested item doesn't exist |
| `405` | Method Not Allowed | Wrong HTTP method for this endpoint |
| `500` | Internal Server Error | Something went wrong on the server |

---

## What You've Learned

Congratulations! You've built a real web API. Let's review:

| Concept | What It Does |
|---------|--------------|
| **HTTP Server** | `http.ListenAndServe()` starts a web server |
| **Pipe Operator** | Cleanly chain functions (like JSON encoders) with `|>` |
| **Method Values** | Pass `store.handleTodos` directly as an HTTP handler |
| **Handlers** | Methods that respond to web requests |
| **Request Methods** | GET (read), POST (create), PUT (update), DELETE (remove) |
| **JSON** | Data format for web APIs (`encoding/json` package) |
| **Status Codes** | Numbers that indicate success or failure |
| **URL Paths** | Routes like `/todos` and `/todos/1` |

---

## Current Limitations

Our todo API works, but it has some limitations:

1. **Data disappears when you restart** - We're storing in memory, not a database
2. **Not safe for multiple users** - Concurrent writes to `store.todos` could race and corrupt data. In production, you'd use a mutex (sync lock) to protect access
3. **No authentication** - Anyone can access it
4. **No input validation** - We only check that title isn't empty. In production, you'd validate title length, sanitize input, and enforce business rules
5. **Simple search only** - The substring search is slow on large datasets. You'd use regex or full-text search in production

We'll fix the first two limitations in the next tutorial!

---

## Practice Exercises

Before moving on, try these enhancements:

1. **Add a `priority` field** - Make todos have high/medium/low priority
2. **Add a search endpoint** - `GET /todos?search=groceries` 
3. **Count endpoint** - `GET /todos/count` returns the number of todos
4. **Filter by completed** - `GET /todos?completed=true`

---

## What's Next?

You now have a working web API! But it's not production-ready yet. In the next tutorial, you'll learn:

- **[Production Patterns Tutorial](production-patterns-tutorial.md)** (Advanced)
  - Store data in a **database** (SQLite)
  - Handle **multiple users safely** with locking
  - Learn **Go conventions** for larger applications
  - Add proper **logging** and **configuration**

---

**You've built a web API! 🚀**
