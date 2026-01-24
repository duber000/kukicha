# Building a Web Todo App with Kukicha

**Level:** Intermediate  
**Time:** 30 minutes  
**Prerequisite:** [Console Todo Tutorial](console-todo-tutorial.md)

Welcome! You've built a console app that saves to files. Now let's build something even cooler: a **web application** that you can access from a browser!

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

Don't worry if this looks complicated - we'll build it step by step!

---

## Step 1: Your First Web Server

Let's start with the simplest possible web server:

```kukicha
import "fmt"
import "net/http"

func main()
    # When someone visits the homepage, say hello
    http.HandleFunc("/", sayHello)
    
    print("Server starting on http://localhost:8080")
    http.ListenAndServe(":8080", empty) onerr panic "server failed to start"

# This function handles requests to "/"
func sayHello(response http.ResponseWriter, request reference http.Request)
    fmt.Fprintln(response, "Hello from Kukicha!")
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
func myHandler(response http.ResponseWriter, request reference http.Request)
    # response - write your reply here
    # request - contains info about the incoming request
```

We can check what **method** (GET, POST, etc.) the user is using:

```kukicha
func myHandler(response http.ResponseWriter, request reference http.Request)
    if request.Method equals "GET"
        fmt.Fprintln(response, "You used GET!")
    else if request.Method equals "POST"
        fmt.Fprintln(response, "You used POST!")
    else
        fmt.Fprintln(response, "You used something else!")
```

---

## Step 3: Sending JSON Responses

Web APIs typically send data as **JSON** (JavaScript Object Notation). It looks like this:

```json
{"id": 1, "title": "Buy groceries", "completed": false}
```

> **📚 Note: JSON in Kukicha with Go 1.25+**
>
> Kukicha's stdlib `parse` and `fetch` packages use Go 1.25+ `encoding/json/v2` for 2-10x faster JSON parsing:
> ```kukicha
> config := files.Read("config.json") |> parse.Json() as Config
> response := fetch.Get(url) |> fetch.Json() as Data  # Streams efficiently!
> ```
>
> For web servers, you can use `encoding/json/v2` for streaming:
> ```kukicha
> import "encoding/json/v2"
> json.MarshalWrite(response, data)  # Write JSON directly to response
> json.UnmarshalRead(request.Body, reference result)  # Stream from request
> ```
>
> **Rule of thumb:** Use `parse.Json()`/`fetch.Json()` for convenience, use `encoding/json/v2` directly for custom streaming needs.

Kukicha makes sending JSON easy:

```kukicha
import "encoding/json/v2"  # Go 1.25+ for better performance

type Todo
    id int
    title string
    completed bool

func sendTodo(response http.ResponseWriter, request reference http.Request)
    # Create a todo
    todo := Todo
        id: 1
        title: "Learn Kukicha"
        completed: false
    
    # Tell the browser we're sending JSON
    response.Header().Set("Content-Type", "application/json")
    
    # Convert the todo to JSON and send it using pipe
    response |> json.NewEncoder() |> .Encode(todo) onerr return
```

When someone visits this endpoint, they'll receive:
```json
{"id":1,"title":"Learn Kukicha","completed":false}
```

---

## Step 4: Receiving JSON Data

When creating a new todo, the user sends JSON data to us. We need to read and parse it:

```kukicha
func createTodo(response http.ResponseWriter, request reference http.Request)
    # Create an empty todo to fill with the incoming data
    todo := Todo{}
    
    # Parse the JSON from the request body using pipe
    # "reference of" gets a pointer so the decoder can fill in our todo
    request.Body |> json.NewDecoder() |> .Decode(reference of todo) onerr
        response.WriteHeader(400)  # 400 = Bad Request
        fmt.Fprintln(response, "Invalid JSON")
        return
    
    # Now 'todo' contains the data the user sent!
    print("Received todo: {todo.title}")
    
    # Send back a success response
    response.Header().Set("Content-Type", "application/json")
    response.WriteHeader(201)  # 201 = Created
    json.NewEncoder(response).Encode(todo) onerr return
```

**What's `reference of`?**

When we write `reference of todo`, we're giving the JSON decoder a way to **fill in** our todo variable. Without it, the decoder would only have a copy and couldn't modify our actual todo.

---

## Step 5: Building the Todo Storage

Let's create a simple way to store our todos. We'll use a **global variable** for now (we'll learn a better way in the next tutorial):

```kukicha
# Our todo storage - a list of todos and the next available ID
var todos list of Todo
var nextId int = 1

# Helper to find a todo by ID
func findTodoById(id int) (Todo, int, bool)
    for index, todo in todos
        if todo.id equals id
            return todo, index, true
    return Todo{}, -1, false
```

---

## Step 6: The Complete Todo API

Now let's put it all together! Create `main.kuki`:

```kukicha
import "net/http"
import "encoding/json/v2"  # Go 1.25+ for better performance
import "strconv"
import "slices"
import "stdlib/string"
import "stdlib/iter"

# --- Data Types ---

type Todo
    id int
    title string
    completed bool

# --- Storage ---
# (In the Production tutorial, we'll use a database instead)

var todos list of Todo
var nextId int = 1

# --- Helper Functions ---

func findTodoIndex(id int) int
    # Use standard slices.IndexFunc to find the item
    return todos |> slices.IndexFunc(func(t Todo) bool: return t.id equals id)

func getIdFromPath(path string, prefix string) (int, bool)
    # Extract "1" from "/todos/1"
    idStr := string.TrimPrefix(path, prefix)
    if idStr equals "" or idStr equals path
        return 0, false
    
    id := idStr |> strconv.Atoi() onerr return 0, false
    return id, true

    response.Header().Set("Content-Type", "application/json")
    response |> json.NewEncoder() |> .Encode(data) onerr return

func sendError(response http.ResponseWriter, status int, message string)
    response.Header().Set("Content-Type", "application/json")
    response.WriteHeader(status)
    
    errorResponse := map of string to string
        error: message
    response |> json.NewEncoder() |> .Encode(errorResponse) onerr return

# --- API Handlers ---

# GET /todos - List all todos (with optional search)
func handleListTodos(response http.ResponseWriter, request reference http.Request)
    search := request.URL.Query().Get("search")
    if search equals ""
        sendJSON(response, todos)
        return
    
    # Supercharge the data flow with iterators!
    filtered := todos
        |> slices.Values()
        |> iter.Filter(func(t Todo) bool
            return string.Contains(string.ToLower(t.title), string.ToLower(search))
        )
        |> iter.Collect()
    
    sendJSON(response, filtered)

# POST /todos - Create a new todo
func handleCreateTodo(response http.ResponseWriter, request reference http.Request)
    # Parse the incoming JSON using pipe
    todo := Todo{}
    request.Body |> json.NewDecoder() |> .Decode(reference of todo) onerr
        sendError(response, 400, "Invalid JSON")
        return
    
    # Validate
    if todo.title equals ""
        sendError(response, 400, "Title is required")
        return
    
    # Assign an ID and add to the list
    todo.id = nextId
    nextId = nextId + 1
    todos = append(todos, todo)
    
    # Send back the created todo
    response.WriteHeader(201)
    sendJSON(response, todo)

# GET /todos/{id} - Get a specific todo
func handleGetTodo(response http.ResponseWriter, request reference http.Request)
    # Get the ID from the URL
    id, ok := getIdFromPath(request.URL.Path, "/todos/")
    if not ok
        sendError(response, 400, "Invalid todo ID")
        return
    
    # Find the todo
    index := findTodoIndex(id)
    if index equals -1
        sendError(response, 404, "Todo not found")
        return
    
    sendJSON(response, todos[index])

# PUT /todos/{id} - Update a todo
func handleUpdateTodo(response http.ResponseWriter, request reference http.Request)
    # Get the ID from the URL
    id, ok := getIdFromPath(request.URL.Path, "/todos/")
    if not ok
        sendError(response, 400, "Invalid todo ID")
        return
    
    # Find the todo
    index := findTodoIndex(id)
    if index equals -1
        sendError(response, 404, "Todo not found")
        return
    
    # Parse the update using pipe
    updated := Todo{}
    request.Body |> json.NewDecoder() |> .Decode(reference of updated) onerr
        sendError(response, 400, "Invalid JSON")
        return
    
    # Keep the original ID, update other fields
    updated.id = id
    todos[index] = updated
    
    sendJSON(response, updated)

# DELETE /todos/{id} - Delete a todo
func handleDeleteTodo(response http.ResponseWriter, request reference http.Request)
    # Get the ID from the URL
    id, ok := getIdFromPath(request.URL.Path, "/todos/")
    if not ok
        sendError(response, 400, "Invalid todo ID")
        return
    
    # Find the todo
    index := findTodoIndex(id)
    if index equals -1
        sendError(response, 404, "Todo not found")
        return
    
    # Remove by creating a new list without this item
    todos = append(todos[:index], todos[index+1:]...)
    
    response.WriteHeader(204)  # 204 = No Content (success, nothing to return)

# --- Route Handler ---

func handleTodos(response http.ResponseWriter, request reference http.Request)
    # Route to the right handler based on the path and method
    
    if request.URL.Path equals "/todos"
        # Collection routes: /todos
        if request.Method equals "GET"
            handleListTodos(response, request)
        else if request.Method equals "POST"
            handleCreateTodo(response, request)
        else
            sendError(response, 405, "Method not allowed")
    else
        # Item routes: /todos/{id}
        if request.Method equals "GET"
            handleGetTodo(response, request)
        else if request.Method equals "PUT"
            handleUpdateTodo(response, request)
        else if request.Method equals "DELETE"
            handleDeleteTodo(response, request)
        else
            sendError(response, 405, "Method not allowed")

# --- Main Entry Point ---

func main()
    # Set up routes
    http.HandleFunc("/todos", handleTodos)
    http.HandleFunc("/todos/", handleTodos)
    
    print("=== Kukicha Todo API ===")
    print("Server running on http://localhost:8080")
    print("")
    print("Try these commands in another terminal:")
    print("  curl http://localhost:8080/todos")
    print("  curl -X POST -d '{\"title\":\"Learn Kukicha\"}' http://localhost:8080/todos")
    print("")
    
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
curl -X POST -d '{"title":"Buy groceries"}' http://localhost:8080/todos
# Response: {"id":1,"title":"Buy groceries","completed":false}

curl -X POST -d '{"title":"Learn Kukicha"}' http://localhost:8080/todos
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
curl -X PUT -d '{"title":"Buy groceries","completed":true}' http://localhost:8080/todos/1
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
| **Handlers** | Functions that respond to web requests |
| **Request Methods** | GET (read), POST (create), PUT (update), DELETE (remove) |
| **JSON** | Data format for web APIs (`encoding/json` package) |
| **Status Codes** | Numbers that indicate success or failure |
| **URL Paths** | Routes like `/todos` and `/todos/1` |

---

## Current Limitations

Our todo API works, but it has some limitations:

1. **Data disappears when you restart** - We're storing in memory, not a database
2. **Not safe for multiple users** - If two people use it at once, data could get corrupted
3. **No authentication** - Anyone can access it

We'll fix all of these in the next tutorial!

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
