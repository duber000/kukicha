# Kukicha Language Syntax Reference v1.0

## Overview

Kukicha is a high-level language that compiles to Go 1.25+ with the Green Tea garbage collector. It prioritizes readability and English-like syntax while maintaining Go's explicit type system and performance.

---

## Module Structure (Botanical Metaphor)

Kukicha uses a simple botanical hierarchy that maps directly to Go's structure:

```
Twig (Go module)
  └── Stem (Go package)
      └── Leaf (Go file)
```

### Mapping to Go

| Kukicha | Go Equivalent | File/Directory |
|---------|---------------|----------------|
| **Twig** | Go module | Root directory with `twig.toml` |
| **Stem** | Go package | Subdirectory (package) |
| **Leaf** | Go file | Individual `.kuki` file |

### Example Directory Structure

```
myapp/                      # Twig (module root)
  twig.toml                # Module configuration
  database/                # Stem (package)
    models.kuki            # Leaf (file)
    queries.kuki           # Leaf (file)
  api/                     # Stem (package)
    handlers.kuki          # Leaf (file)
    middleware.kuki        # Leaf (file)
  todo/                    # Stem (package)
    todo.kuki              # Leaf (file)
```

---

## Leaf Declaration (Optional)

Leaf declarations are **optional**. If not provided, the compiler automatically calculates the Stem (package) name based on the file's path relative to `twig.toml`.

### Implicit Stem Calculation

Files in a directory automatically belong to the stem named after that directory:

```
myapp/                      # Twig (module root)
  twig.toml                # Module configuration
  src/
    auth/
      login.kuki           # Belongs to "auth" stem (implicit)
      session.kuki         # Belongs to "auth" stem (implicit)
    database/
      models.kuki          # Belongs to "database" stem (implicit)
```

**Benefit**: Reduces boilerplate and prevents "file-move" errors where the directory and header get out of sync.

### Explicit Leaf Declaration (Optional)

You can still explicitly declare the stem if needed:

```kukicha
leaf database.models
```

This is useful for:
- Multi-level package paths (e.g., `database.models.user`)
- Overriding the default path-based stem name
- Self-documenting code when the file structure is complex

**Rule**: If a leaf declaration is present, it takes precedence over path-based inference. If absent, the compiler uses the directory path relative to `twig.toml`.

---

## Imports

Import standard library packages, GitHub packages, or local stems:

```kukicha
import time
import github.com/user/repo/package as alias
import github.com/user/repo@v1.2.3
import todo
import database.models as models
```

---

## Export Control

Kukicha uses Go's case convention for exports:

```kukicha
# Exported (uppercase first letter)
type Todo              # Available outside the stem
func CreateTodo()      # Available outside the stem

# Unexported (lowercase first letter)
type internalCache     # Private to this stem
func validateInput()   # Private to this stem
```

**Rules:**
- **Types**: `PascalCase` = exported, `camelCase` = unexported
- **Functions**: `PascalCase` = exported, `camelCase` = unexported
- **Variables**: Always `camelCase` (exported via functions if needed)

---

## Types

### Type Declaration

Types are declared with fields and their Go-style explicit types:

```kukicha
type Todo
    id int64
    title string
    description string
    completed bool
    created_at time.Time
    completed_at time.Time
```

No `struct` keyword needed — the type definition is implicit.

### Supported Type Syntax

- **Primitives**: `int64`, `int32`, `int`, `string`, `bool`, `float64`, `float32`
- **References (Pointers)**: `reference Type`
- **Collections**: `list of Type`, `map of KeyType to ValueType`
- **Other packages**: `time.Time`, `sync.RWMutex`

### Types with References

```kukicha
type User
    id int64
    name string
    settings reference Settings
    cache reference sync.RWMutex
```

Compiles to Go as:
```go
type User struct {
    ID int64
    Name string
    Settings *Settings
    Cache *sync.RWMutex
}
```

---

## Variables & Assignment

### Binding (Walrus Operator ⭐)

Create new bindings with `:=`:

```kukicha
result := expensiveOperation()
todos := empty list of Todo
config := map of string to string
user := empty reference User
```

### Reassignment

Update existing variables with `=`:

```kukicha
result = newValue
todos = append(todos, newTodo)
completed = true
user = reference to newUser
```

---

## Functions

### Basic Function Declaration

**Function parameters and return types must have explicit type annotations** to maintain Go's performance and static safety. Type inference is only used for local variables inside function bodies.

```kukicha
func CreateTodo(id int64, title string, description string) Todo
    return Todo
        id: id
        title: title
        description: description
        completed: false
        created_at: time.now()
        completed_at: empty
```

**Signature-First Type Inference Rules:**
- ✅ **Explicit**: Function parameters must declare types
- ✅ **Explicit**: Function return types must be declared
- ✅ **Explicit**: Struct fields must declare types
- ✅ **Inferred**: Local variables inside function bodies use `:=` for inference

**Example of local variable inference:**
```kukicha
func ProcessData(input string) int
    # Local variables are inferred
    result := parseInput(input)        # Type inferred from parseInput's return
    count := len(result)               # Type inferred from len() return (int)
    doubled := count * 2               # Type inferred from arithmetic (int)
    return doubled
```

### Single Return

```kukicha
func MarkDone(todo Todo) Todo
    todo.completed = true
    todo.completed_at = time.now()
    return todo
```

### Multiple Return (Tuple)

```kukicha
func Display(todo Todo) (int64, string, string)
    status := "pending"
    if todo.completed
        status = "done"
    return todo.id, todo.title, status
```

Multiple returns are separated by commas and form a tuple. Return types must be explicitly declared in the signature.

### Using Discard for Unwanted Returns

```kukicha
# Ignore the first return value
discard, exists := config at "host"

# Ignore the second return value
value, discard := multiReturn()

# Ignore multiple values
discard, result, discard := threeReturns()
```

`discard` replaces Go's `_` underscore for ignored values.

### Struct Initialization in Returns

Indentation-based, explicit field assignment:

```kukicha
return Todo
    id: 1
    title: "buy milk"
    description: "at store"
    completed: false
    created_at: time.now()
    completed_at: empty
```

---

## Control Flow

### If/Else Blocks

```kukicha
if todo.completed
    print "done"
else
    print "pending"
```

```kukicha
status := "pending"
if todo.completed
    status = "done"
```

### For Loops

**Range-based iteration (exclusive):**
```kukicha
for i from 0 to 10
    print i
# Prints 0, 1, 2, 3, 4, 5, 6, 7, 8, 9 (0-9, excludes 10)
```

**Range-based iteration (inclusive):**
```kukicha
for i from 1 through 10
    print i
# Prints 1, 2, 3, 4, 5, 6, 7, 8, 9, 10 (includes 10)
```

**Range rules:**
- `from X to Y` - Excludes Y (like Go's `< Y`)
- `from X through Y` - Includes Y (like `<= Y`)

**Collection iteration:**
```kukicha
for todo in todos
    print Display(todo)

for key, value in config
    print key + ": " + value
```

**With discard:**
```kukicha
# Iterate without using the index
for discard, todo in todos
    process(todo)

# Iterate without using the value
for key, discard in config
    print "key exists: " + key
```

**Go-style loops also work:**
```kukicha
for i := 0; i < 10; i++
    print i

for i, todo := range todos
    print i, todo
```

---

## Error Handling

Kukicha provides ergonomic error handling with the `onerr` operator while maintaining Go's explicit style.

### The `onerr` Operator

The `onerr` operator automatically unwraps `(value, error)` tuples and handles errors inline:

```kukicha
# Panic on error
content := file.read("config.json") onerr panic "missing config file"

# Return error up the call stack
data := http.get(url) onerr return error "failed to fetch data"

# Provide default value
port := env.get("PORT") onerr "8080"

# Custom error handling
config := file.read("config.json") onerr
    print "Warning: config not found, using defaults"
    loadDefaults()
```

**How it works:**
```kukicha
# This line
result := someFunc() onerr panic "failed"

# Is equivalent to
result, err := someFunc()
if err != empty
    panic "failed"
```

**Design Note:** The `onerr` keyword is distinct from the `or` keyword used for boolean logic. This separation makes code more readable - you can tell at a glance whether an expression handles errors or performs boolean operations.

### Boolean OR vs Error Handling

```kukicha
# Boolean OR - for logical operations
if active or pending or draft
    process()

# Error handling - with onerr
data := fetch() onerr return error "failed"
```

### Explicit Error Handling

Traditional Go-style error handling is also supported:

```kukicha
data, err := file.read("config.json")
if err != empty
    print "Error reading config: {err}"
    return err

config, err := json.parse(data)
if err != empty
    return error "invalid JSON: {err}"
```

### Error Chaining

```kukicha
# Try multiple sources
config := file.read("config.json")
    onerr file.read("config.yaml")
    onerr file.read("config.toml")
    onerr panic "no config file found"
```

---

## Methods

Methods are functions that operate on types. Use the `on this` syntax with an explicit `this` receiver.

### Method Declaration

```kukicha
type Todo
    id int64
    title string
    completed bool

# Value receiver - uses explicit 'this'
func Display on this Todo string
    status := "pending"
    if this.completed
        status = "done"
    return "{status}: {this.title}"

# Reference receiver - for mutation
func MarkDone on this reference Todo
    this.completed = true
    this.completed_at = time.now()

# Method with parameters
func UpdateTitle on this reference Todo, newTitle string
    this.title = newTitle
```

**Design Note:** The `this` keyword is explicit in the method signature (`on this Todo`), making it clear that `this` refers to the receiver. This improves discoverability for beginners - you can see where `this` comes from just by reading the function declaration.

### Method Usage

```kukicha
todo := CreateTodo(1, "Buy milk", "At store")

# Call methods
message := todo.Display()
print message

# Modify with reference receiver
todo.MarkDone()
todo.UpdateTitle("Buy organic milk")
```

### Go-Style Syntax (Also Works)

```kukicha
# Explicit receiver parameter
func (todo Todo) Display() string
    status := "pending"
    if todo.completed
        status = "done"
    return "{status}: {todo.title}"

func (todo *Todo) MarkDone()
    todo.completed = true
```

---

## Interfaces

Interfaces define contracts that types can implement. Implementation is implicit (like Go).

### Interface Declaration

```kukicha
interface Displayable
    Display() string
    GetTitle() string

interface Serializable
    ToJSON() string
    FromJSON(data string)
```

### Implicit Implementation

Types implement interfaces automatically by having the required methods:

```kukicha
type Todo
    id int64
    title string
    completed bool

# Implementing Displayable interface
func Display on this Todo string
    return "{this.id}. {this.title}"

func GetTitle on this Todo string
    return this.title

# Todo now implements Displayable automatically!
```

### Using Interfaces

```kukicha
func ShowItem(item Displayable)
    print item.Display()

func ShowAll(items list of Displayable)
    for discard, item in items
        print item.Display()

# Usage
todo := Todo
    id: 1
    title: "Buy milk"
    completed: false

ShowItem(todo)  # Works because Todo implements Displayable
```

---

## Concurrency

Kukicha uses Go's powerful concurrency model with more readable syntax.

### Goroutines

Launch concurrent tasks with `go`:

```kukicha
# Start goroutine
go doWork()
go fetchData(url)

# With anonymous function
go
    result := expensiveOperation()
    print "Done: {result}"

# Pass parameters
go processItem(item)
```

### Channels

Channels provide communication between goroutines:

**Creating channels:**
```kukicha
# Unbuffered channel
ch := make channel of string

# Buffered channel
ch := make channel of int, 100

# Go syntax also works
ch := make(chan string)
ch := make(chan int, 100)
```

**Sending and receiving:**
```kukicha
# Primary syntax
send ch, "message"
msg := receive ch

# Go syntax also works
ch <- "message"
msg := <-ch
```

**Closing channels:**
```kukicha
close ch
```

### Concurrency Examples

**Parallel fetching:**
```kukicha
func fetchAll(urls list of string)
    results := make channel of string, len(urls)
    
    for discard, url in urls
        go
            response := http.get(url) onerr return
            send results, response.body
    
    allData := empty list of string
    for i from 0 to len(urls)
        data := receive results
        allData = append(allData, data)
    
    return allData
```

**Worker pool:**
```kukicha
func processJobs(jobs list of Job, workers int)
    jobCh := make channel of Job, len(jobs)
    resultCh := make channel of Result, len(jobs)
    
    # Start workers
    for i from 0 to workers
        go worker(jobCh, resultCh)
    
    # Send jobs
    for discard, job in jobs
        send jobCh, job
    close jobCh
    
    # Collect results
    results := empty list of Result
    for i from 0 to len(jobs)
        result := receive resultCh
        results = append(results, result)
    
    return results
```

---

## Defer and Recover

### Defer (Cleanup)

`defer` ensures code runs when the function exits, useful for cleanup:

```kukicha
func processFile(path string)
    file := file.open(path) onerr return error "cannot open"
    defer file.close()  # Always closes, even on error
    
    # Work with file
    content := file.read()
    return content

func complexOperation()
    defer print "Operation complete"
    defer connection.close()
    defer file.close()
    
    # Multiple defers execute in reverse order (LIFO)
    # Output: file closes, then connection, then prints
```

### Recover (Panic Handling)

`recover()` catches panics and allows graceful handling:

```kukicha
func safeOperation()
    defer
        if r := recover(); r != empty
            print "Recovered from panic: {r}"
            # Can return error or handle gracefully
    
    riskyOperation()  # Might panic

func safeDivide(a, b int)
    defer
        if r := recover(); r != empty
            print "Division error: {r}"
            return 0
    
    return a / b  # Panics if b == 0
```

**Combined with error handling:**
```kukicha
func robustOperation() error
    defer
        if r := recover(); r != empty
            return error "operation panicked: {r}"
    
    result := dangerousFunc() onerr return error "failed"
    return empty  # No error
```

---

## Collections

### Lists

**Declaration:**
```kukicha
todos := empty list of Todo
tags := list of string
numbers := list of int64
```

**Adding elements:**
```kukicha
todos = append(todos, newTodo)
```

**Access by index:**
```kukicha
# Positive indexing
first := items at 0
first := items[0]

# Negative indexing (from end)
last := items at -1
secondLast := items[-2]
```

**Slicing:**
```kukicha
# Slicing uses Go syntax
subset := items[2:7]

# Negative indices in slices
lastThree := items[-3:]
allButLast := items[:-1]
middle := items[1:-1]

# Empty bounds
fromThird := items[3:]
firstFive := items[:5]
```

**Membership testing:**
```kukicha
# Check if item exists in list
if user in admins
    grantAccess()

if item not in blacklist
    process(item)
```

**Length:**
```kukicha
count := len(items)
```

### Maps

**Declaration:**
```kukicha
config := map of string to string
    host: "localhost"
    port: "5432"
    debug: "true"
```

**Access:**
```kukicha
value := config at "host"
config at "port" = "8080"
```

**Existence check:**
```kukicha
# Traditional check
value, exists := config at "host"
if exists
    print "host configured"

# Membership operator (recommended)
if "host" in config
    connect(config at "host")

if "api_key" not in config
    print "Warning: API key not configured"
```

---

## Membership Testing

Kukicha provides intuitive membership operators for checking if an item exists in a collection.

### The `in` Operator

Check if an element exists in a list, key exists in a map, or substring exists in a string:

```kukicha
# List membership
if user in admins
    grantAccess()

if "active" in statusList
    proceed()

# Map key existence
if "host" in config
    connect(config at "host")

if "DEBUG" in environment
    enableDebugMode()

# String containment
if "error" in logMessage
    alertOps()
```

### The `not in` Operator

Check if an element does NOT exist:

```kukicha
# List exclusion
if item not in blacklist
    process(item)

if user not in bannedUsers
    allowAccess()

# Map key absence
if "api_key" not in config
    print "Warning: API key missing"
    return error "configuration incomplete"

# String exclusion
if "success" not in response
    retryOperation()
```

### How It Works

**For lists/slices:**
```kukicha
# Kukicha
if item in items
    print "found"

# Compiles to Go (using slices package)
import "slices"
if slices.Contains(items, item) {
    fmt.Println("found")
}
```

**For maps:**
```kukicha
# Kukicha
if key in config
    print "exists"

# Compiles to Go (using Go's map idiom)
if _, exists := config[key]; exists {
    fmt.Println("exists")
}
```

**For strings:**
```kukicha
# Kukicha
if "hello" in text
    print "found"

# Compiles to Go (using strings package)
import "strings"
if strings.Contains(text, "hello") {
    fmt.Println("found")
}
```

### Practical Examples

**Validation:**
```kukicha
func ValidatePermissions(user User, action string)
    allowedActions := list of string{"read", "write", "delete"}

    if action not in allowedActions
        return error "invalid action"

    if user.role not in list of string{"admin", "moderator"}
        return error "insufficient permissions"

    return empty
```

**Configuration checking:**
```kukicha
func LoadDatabase(config map of string to string)
    requiredKeys := list of string{"host", "port", "database", "user"}

    for discard, key in requiredKeys
        if key not in config
            return error "missing config: {key}"

    # All keys present, proceed
    connectDatabase(config)
```

---

## Negative Indexing & Slicing

Kukicha supports Python-style negative indexing with **compile-time optimization for literals** to maintain Go's performance.

### Negative Indexing with Literal Optimization

Access elements from the end using negative numbers. When the index is a **literal constant**, the compiler generates optimized code at compile-time:

```kukicha
items := list of string{"a", "b", "c", "d", "e"}

# Literal negative indices (compile-time optimized)
last := items at -1           # "e" - compiled to items[len(items)-1]
secondLast := items[-2]       # "d" - compiled to items[len(items)-2]
thirdLast := items[-3]        # "c" - compiled to items[len(items)-3]

# Both syntaxes work
last := items at -1
last := items[-1]
```

**Literal Optimization (No Runtime Overhead):**
```kukicha
# Kukicha (literal index)
last := items at -1

# Compiles directly to Go (calculated at compile-time)
last := items[len(items)-1]
```

### Dynamic Negative Indexing

For **variable-based** negative indices (dynamic at runtime), use the `at` method which includes bounds checking:

```kukicha
# Dynamic negative indexing (runtime calculation)
index := -2
element := items.at(index)    # Bounds-checked, handles negative indices at runtime

# Or with explicit function
element := at(items, index)   # Equivalent explicit function call
```

**Performance Trade-off:**
- **Literal indices** (e.g., `-1`, `-2`): Zero runtime overhead, compiled to `len(items)-N`
- **Dynamic indices** (e.g., variables): Requires `at()` method with safety checks

This design preserves raw Go speed for standard loops while supporting negative indexing where needed.

### Slicing with Negative Indices (Literal Optimization)

Negative indices in slice operations are also compile-time optimized when using **literal constants**:

```kukicha
items := list of int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

# Literal negative indices in slices (compile-time optimized)
lastThree := items[-3:]        # [7, 8, 9]
lastFive := items[-5:]         # [5, 6, 7, 8, 9]

# Remove last N elements
allButLast := items[:-1]       # [0, 1, 2, 3, 4, 5, 6, 7, 8]
allButLastTwo := items[:-2]    # [0, 1, 2, 3, 4, 5, 6, 7]

# Extract middle (remove first and last)
middle := items[1:-1]          # [1, 2, 3, 4, 5, 6, 7, 8]

# Mix positive and negative literals
fromThirdToSecondLast := items[3:-2]  # [3, 4, 5, 6, 7]
```

**Literal Compilation (Zero Runtime Overhead):**
```kukicha
# Kukicha (literal indices)
lastThree := items[-3:]
allButLast := items[:-1]
middle := items[1:-1]

# Compiles directly to Go (calculated at compile-time)
lastThree := items[len(items)-3:]
allButLast := items[:len(items)-1]
middle := items[1:len(items)-1]
```

**Dynamic Slicing:**

For variable-based slice bounds, use explicit slice methods:

```kukicha
start := -5
end := -2
subset := items.slice(start, end)  # Handles negative indices at runtime
```

### Practical Examples

**Processing recent logs:**
```kukicha
func AnalyzeRecentLogs(logs list of LogEntry)
    # Get last 100 log entries
    recent := logs[-100:]

    for discard, entry in recent
        if "ERROR" in entry.message
            alertOps(entry)
```

**Removing header/footer:**
```kukicha
func ExtractBody(lines list of string)
    # Remove first line (header) and last line (footer)
    body := lines[1:-1]
    return body
```

**Tail of data:**
```kukicha
func GetTail(data list of int, n int)
    # Return last n elements safely
    if len(data) < n
        return data
    return data[-n:]
```

**Trimming:**
```kukicha
func TrimEdges(items list of string)
    # Remove first 2 and last 2 elements
    if len(items) <= 4
        return empty list of string

    trimmed := items[2:-2]
    return trimmed
```

---

## Pipe Operator

The pipe operator `|>` enables clean data pipelines by passing the result of one expression as the first argument to the next function.

### Basic Usage

```kukicha
# Instead of nested calls
result := process(transform(parse(data)))

# Use pipes (reads left to right)
result := data
    |> parse()
    |> transform()
    |> process()
```

### Piping Rules

**Basic usage — result becomes first argument:**

```kukicha
# These are equivalent
x |> func(y, z)
func(x, y, z)
```

**Method calls — use leading dot to call method on piped value:**

```kukicha
response |> .json()  # Same as: response.json()
```

**Chaining both styles:**

```kukicha
users := http.get(url)
    |> .json()              # Call .json() on response
    |> filterActive()       # Pass result to filterActive()
    |> sortByName()
```

### Multiline Pipes (Recommended Style)

```kukicha
config := file.read("config.json")
    |> json.parse()
    |> validate()
    |> applyDefaults()
    onerr return error "invalid config"
```

### Pipe with Standard Library

```kukicha
# HTTP data processing
users := http.get("https://api.com/users")
    |> .json() as list of User
    |> filterByAge(18)
    |> sortByName()

# File pipeline
content := file.read("data.csv")
    |> string.split("\n")
    |> parseCSV()
    |> filterValid()
    |> toJSON()

# LLM workflow
analysis := file.read("logs.txt")
    |> extractErrors()
    |> claude.complete("Analyze these errors")
    |> formatReport()
```

### Operator Precedence

Pipe has low precedence (only `or` is lower):

```kukicha
# Arithmetic binds tighter
result := a + b |> double()
# Same as: (a + b) |> double()

# Works with or operator
data := fetch() |> parse() onerr return error "failed"
# Same as: (fetch() |> parse()) onerr return error "failed"
```

### Single-Line Pipes

```kukicha
# Short chains can be single-line
id := getUserId() |> int64() |> validate()

# But multiline is more readable for longer chains
result := data
    |> step1()
    |> step2()
    |> step3()
```

### Mixing with Method Calls

```kukicha
# Methods and pipes work together
todo := CreateTodo(1, "Buy milk")
    |> .MarkDone()           # Method call
    |> save()                # Function call
    |> logAction("created")  # Function call
```

### Real-World Examples

**GitHub API Pipeline:**
```kukicha
func GetTopRepos(username string)
    return "https://api.github.com/users/{username}/repos"
        |> http.get()
        |> .json() as list of Repo
        |> filterByStars(10)
        |> sortByUpdated()
        |> take(5)
        onerr empty list of Repo
```

**Data Transformation Pipeline:**
```kukicha
func ProcessCSV(path string)
    return file.read(path)
        |> string.split("\n")
        |> parseCSVLines()
        |> filterValid()
        |> aggregateByCategory()
        |> toJSON()
        onerr return error "processing failed"
```

**Container Deployment Pipeline:**
```kukicha
func DeployApp(version string)
    image := "my-app:{version}"
        |> docker.build()
        |> docker.test()
        |> docker.push()
        onerr return error "build failed"
    
    return image
        |> k8s.deploy("production")
        |> k8s.waitReady()
        onerr return error "deployment failed"
```

---

## Strings

### String Interpolation Rules

**Quoted strings are always string literals:**
```kukicha
message := "hello {name}"                    # String
status := "{todo.id}. {todo.title}"          # String
return "{todo.id} {todo.title} {status}"     # Returns single string
```

**Unquoted comma-separated values in returns are tuples:**
```kukicha
return todo.id, todo.title, status           # Returns 3 values (tuple)
```

**In assignments, interpolated strings are always strings:**
```kukicha
x := "{a} {b}"           # String
y := a, b                # Syntax error (use tuples only in returns)
```

### String Operations

**Concatenation:**
```kukicha
result := str1 + str2
greeting := "hello " + name
```

**String functions:**
```kukicha
upper := string.upper("hello")
parts := string.split("a,b,c", ",")
joined := string.join(parts, "|")
```

---

## References (Pointers)

### Declaration

```kukicha
type Node
    value int64
    next reference Node
    parent reference Node
```

### Creating References

```kukicha
# Empty reference (nil)
user := empty reference User

# Reference to new value
settings := reference to Settings
    theme: "dark"
    font_size: 14
```

### Dereferencing

References are automatically dereferenced when accessing fields:

```kukicha
if user.name equals "admin"
    print "welcome admin"
```

Compiles to Go's automatic dereferencing: `if user.Name == "admin"`

---

## Data Types & Zero Values

### Primitive Types

- `int64`, `int32`, `int`, `uint64`, `uint32`, `uint`
- `float64`, `float32`
- `string`
- `bool`

### Package Types

- `time.Time`
- `sync.RWMutex`
- Any imported type

### Zero Values

```kukicha
empty                          # Zero value, nil equivalent
nil                            # Accepted as alias for empty
empty reference Type           # nil pointer
false                          # boolean zero
0                              # numeric zero
""                             # empty string
time.now()                     # current time
empty list of Type             # empty list
empty map of KeyType to ValueType  # empty map
```

**Note:** `nil` is accepted as a direct alias for `empty` for Go developers' familiarity. Both compile to the same zero value.

---

## Special Keywords

| Keyword | Purpose |
|---------|---------|
| `leaf` | Module/file declaration |
| `import` | Import packages or stems |
| `type` | Type declaration |
| `interface` | Interface declaration |
| `func` | Function declaration |
| `return` | Return from function |
| `if`, `else` | Control flow |
| `for` | Loop |
| `in` | Loop iteration / map membership |
| `from`, `to` | Range specification (exclusive) |
| `through` | Range specification (inclusive) |
| `at` | Element access |
| `of` | Collection type declaration |
| `and`, `or`, `not` | Boolean operators |
| `empty` | Zero value |
| `reference` | Pointer type |
| `discard` | Ignore return value |
| `on` | Method declaration |
| `this` | Implicit receiver in methods |
| `go` | Spawn goroutine |
| `channel` | Channel type |
| `send`, `receive` | Channel operations |
| `make` | Create channel |
| `close` | Close channel |
| `defer` | Defer execution until function exits |
| `recover` | Recover from panic |
| `panic` | Trigger panic |
| `error` | Create/return error |

---

## Operators

### Assignment

| Operator | Usage |
|----------|-------|
| `:=` | Binding (create new variable) ⭐ |
| `=` | Reassignment (update existing) |

### Arithmetic & String

| Operator | Usage |
|----------|-------|
| `+` | Addition, concatenation |
| `-` | Subtraction |
| `*` | Multiplication |
| `/` | Division |
| `%` | Modulo |

### Comparison

| Operator | Usage |
|----------|-------|
| `equals` | Equality |
| `not equals` | Inequality |
| `!=` | Inequality (alternative) |
| `>` | Greater than |
| `<` | Less than |
| `>=` | Greater than or equal |
| `<=` | Less than or equal |
| `in` | Membership test |
| `not in` | Negative membership test |

### Boolean

| Operator | Usage |
|----------|-------|
| `and` | Logical AND |
| `or` | Logical OR |
| `not` | Logical NOT |

### Access

| Operator | Usage |
|----------|-------|
| `.` | Field/method access |
| `:` | Struct field assignment |
| `,` | Tuple/argument separator |
| `()` | Function call grouping |
| `|>` | Pipe operator (pass result to next function) |

---

## Configuration

Kukicha supports TOML-based configuration:

**twig.toml:**
```toml
[project]
name = "my-app"
version = "0.1.0"

[build]
experiment = "greenteagc"
goVersion = "1.25"

[config]
storageType = "memory"
debug = true
```

**Usage in code:**
```kukicha
config settings from "config.toml"

storageType := settings.config.storageType
debugMode := settings.config.debug or false
```

---

## Symbol Minimalism

### Avoided Symbols

- `{}` — braces (use indentation)
- `;` — semicolons (use newlines)
- `_` — underscores for ignoring (use `discard`)
- `[]` — bracket indexing (use `at` keyword)
- `[Type]` — array syntax (use `list of Type`)
- `map[K]V` — map syntax (use `map of K to V`)
- `&&`, `||` — boolean operators (use `and`, `or`)
- `==` — equality (use `equals`)
- `*` — pointer prefix (use `reference`)

### Kept Symbols

- `()` — function arguments and grouping
- `:` — struct field assignment
- `.` — field/method access
- `:=` — walrus operator (binding) ⭐
- `=` — reassignment
- `,` — tuple/argument separator
- `+` — concatenation and arithmetic
- `-`, `*`, `/`, `%` — arithmetic operators
- `>`, `<`, `>=`, `<=` — comparison
- `!=` — inequality
- `equals` — equality (keyword, not symbol)

---

## Builtins

Kukicha provides built-in functions that are available without imports:

### Core Functions

```kukicha
print(value)                    # Print to stdout
len(collection)                 # Get length of list, map, string, or channel
append(list, item)              # Add item to list, returns new list
```

### Channel Operations

```kukicha
make(channel of Type)           # Create unbuffered channel
make(channel of Type, size)     # Create buffered channel
close(channel)                  # Close channel
```

### Error Handling

```kukicha
panic(message)                  # Trigger panic with message
recover()                       # Recover from panic (use in defer)
```

### Type Casting

```kukicha
int64(value)                    # Convert to int64
int32(value)                    # Convert to int32
int(value)                      # Convert to int
uint64(value)                   # Convert to uint64
uint32(value)                   # Convert to uint32
uint(value)                     # Convert to uint
float64(value)                  # Convert to float64
float32(value)                  # Convert to float32
string(value)                   # Convert to string
bool(value)                     # Convert to bool
```

---

## Example: Complete Todo App Leaf

```kukicha
leaf todo

import time

type Todo
    id int64
    title string
    description string
    completed bool
    created_at time.Time
    completed_at time.Time

type TodoList
    items list of Todo
    count int64

interface Displayable
    Display() string

# Constructor function
func CreateTodo(id int64, title string, description string) Todo
    return Todo
        id: id
        title: title
        description: description
        completed: false
        created_at: time.now()
        completed_at: empty

# Method with value receiver
func Display on this Todo string
    status := "○"
    if this.completed
        status = "✓"
    return "{status} {this.id}. {this.title}"

# Method with reference receiver
func MarkDone on this reference Todo
    this.completed = true
    this.completed_at = time.now()

func UpdateTitle on this reference Todo, newTitle string
    this.title = newTitle

# Function with error handling
func FindById(todos list of Todo, id int64) (Todo, error)
    for discard, todo in todos
        if todo.id equals id
            return todo, empty  # Found, no error
    return empty, error "todo not found"

# Function with or operator
func LoadTodos(path string) list of Todo
    content := file.read(path) onerr return empty list of Todo
    todos := json.parse(content) onerr return empty list of Todo
    return todos

# Concurrent processing
func ProcessAll(todos list of Todo)
    results := make channel of string, len(todos)

    for discard, todo in todos
        go
            result := processTodo(todo)
            send results, result

    for i from 0 to len(todos)
        result := receive results
        print result

# Function with defer
func SaveTodos(path string, todos list of Todo) error
    file := file.open(path) onerr return error "cannot open file"
    defer file.close()

    data := json.encode(todos)
    file.write(data)
    return empty  # No error

# Function with recover
func SafeProcess(todo Todo)
    defer
        if r := recover(); r != empty
            print "Error processing todo: {r}"

    riskyOperation(todo)
```

---

## Compilation

Build Kukicha with Green Tea GC:

```bash
kukicha build --experiment greenteagc
```

Or set environment:

```bash
GOEXPERIMENT=greenteagc kukicha build
```

---

## Version History

- **v1.1.0** — Core Design Refinements (2026)
  - ✅ **Optional Leaf Declarations**: Folder-based package model with automatic Stem calculation from file path
  - ✅ **Signature-First Type Inference**: Explicit types for function parameters/returns and struct fields; inference only for local variables
  - ✅ **Literal vs Dynamic Indexing**: Compile-time optimization for literal negative indices; explicit methods for dynamic indices
  - ✅ **Indentation as Canonical**: `kuki fmt` tool converts brace-based syntax to standard indentation format

- **v1.0.0** — Complete language specification
  - ✅ Error handling with `or` operator (auto-unwraps tuples)
  - ✅ Methods with `on` keyword and implicit `this`
  - ✅ Interfaces (implicit implementation like Go)
  - ✅ Concurrency: `go` for goroutines, `send`/`receive` for channels
  - ✅ Defer and recover for cleanup and panic handling
  - ✅ Range inclusivity: `to` (exclusive), `through` (inclusive)
  - ✅ Pipe operator `|>` for data pipelines
  - ✅ Indentation: 4 spaces (tabs rejected)
  - ✅ Dual syntax support throughout (Kukicha + Go)
  - ✅ File extension: `.kuki` (茎 = stem in Japanese)

- **v0.2.0** — Cohesive syntax refinement
  - Simplified module structure: Twig → Stem → Leaf
  - `reference` keyword replaces `*` for pointers
  - `discard` keyword replaces `_` for ignored values
  - Clarified string interpolation rules
  - Export control via Go's case convention
  - Zero value syntax: `empty reference Type`
  - Type casting: `int64(value)` instead of annotations
  - Removed if/then/else expressions (use standard if/else blocks)
  - Keep `!=` for inequality, `equals` for equality
  - Keep `at` for indexing (more readable for newbies)
  - Dual syntax support: Go syntax accepted where compatible

- **v0.1.0** — Initial syntax specification
  - Module structure (Leaf/Petiole/Stem/Stalk/Twig)
  - English-like syntax (no `{}`, `;`)
  - Walrus operator (`:=`)
  - Type inference for function parameters
  - String interpolation
  - Collection literals with `list of` and `map of`
  - Single-line if/then/else expressions
  - TOML configuration support

---

## Design Philosophy

Kukicha smooths Go's rough edges while preserving its power:

✅ **Keep**: Explicit types, static typing, performance, Go's stdlib
✅ **Smooth**: Symbols minimized, English-like keywords, consistent syntax
✅ **Star**: The walrus operator `:=` for clean variable binding
✅ **Simple**: Three-level module hierarchy that maps 1:1 to Go

---

## Notes

- Kukicha compiles to Go 1.25+
- **Explicit types required**: Function parameters, returns, and struct fields must have type annotations
- **Type inference only for locals**: Local variables inside function bodies use `:=` for inference
- No implicit type conversions (use casting)
- **Indentation is canonical**: `kuki fmt` enforces 4-space indentation standard
- **Optional leaf declarations**: Stem calculated from file path; explicit declaration optional
- Focus on readability without sacrificing Go's performance

---

## Syntax Standard: Indentation is Canonical

**Kukicha uses indentation-based syntax as the canonical standard.** This prevents "Dialect Drift" between Python-style and Go-style formatting.

### The `kuki fmt` Tool

To ensure consistency, the `kuki fmt` formatter will automatically convert brace-based syntax to the standard indentation-based format:

```bash
# Format a single file
kuki fmt myfile.kuki

# Format all files in directory
kuki fmt ./src/

# Check formatting without modifying
kuki fmt --check ./src/
```

**Example conversion:**

```kukicha
# Before: Go-style braces (non-standard)
if count == 5 {
    print("five")
}

# After: Standard Kukicha (kuki fmt output)
if count equals 5
    print "five"
```

### Indentation Rules

- **4 spaces per indentation level** (strict requirement)
- **Tabs are rejected** with error message
- **Newlines define statement boundaries** (no semicolons)
- **INDENT/DEDENT tokens** generated by lexer for block structure

### Alternative Syntax Support (Go Compatibility)

While indentation is canonical, Kukicha accepts **some Go syntax** for compatibility. However, **`kuki fmt` will convert these to the standard form:**

#### Comparisons

**Primary (Recommended):**
```kukicha
if count equals 5
    print "five"
```

**Go Syntax (Also Works):**
```kukicha
if count == 5
    print "five"
```

Both `equals` and `==` compile to identical Go code.

---

#### Boolean Logic

**Primary (Recommended):**
```kukicha
if completed and not expired or override
    proceed()
```

**Go Syntax (Also Works):**
```kukicha
if completed && !expired || override
    proceed()
```

Both `and`/`or`/`not` and `&&`/`||`/`!` are accepted.

---

#### Indexing and Access

**Primary (Recommended):**
```kukicha
first := items at 0
value := config at "host"
config at "port" = "8080"
```

**Go Syntax (Also Works):**
```kukicha
first := items[0]
value := config["host"]
config["port"] = "8080"
```

Both `at` keyword and `[]` brackets work for collections.

---

#### Type Declarations - Collections

**Primary (Recommended):**
```kukicha
todos := empty list of Todo
config := map of string to int
```

**Go Syntax (Also Works):**
```kukicha
todos := []Todo{}
config := map[string]int{}
```

---

#### Type Declarations - References

**Primary (Recommended):**
```kukicha
user := empty reference User
next := reference to Node
```

**Go Syntax (Also Works):**
```kukicha
user := (*User)(nil)
next := &Node{}
```

---

#### Struct Field Types

**Primary (Recommended):**
```kukicha
type Node
    value int64
    next reference Node
    items list of string
    data map of string to int
```

**Go Syntax (Also Works):**
```kukicha
type Node
    value int64
    next *Node
    items []string
    data map[string]int
```

---

#### Loop Iteration

**Collection iteration:**

| Kukicha | Go Syntax |
|---------|-----------|
| `for x in items` | `for _, x := range items` |
| `for i, x in items` | `for i, x := range items` |
| `for discard, x in items` | `for _, x := range items` |

**Range loops:**
```kukicha
# Exclusive range
for i from 0 to 10
    print i  # 0-9

# Inclusive range
for i from 1 through 10
    print i  # 1-10
```

**Go-style C loops:**
```kukicha
for i := 0; i < 10; i++
    print i
```

Note: With Go syntax, you must use `:=` in the for loop declaration.

---

#### Methods

**Primary (Recommended):**
```kukicha
func Display on this Todo
    return "{this.title}"

func MarkDone on this reference Todo
    this.completed = true
```

**Go Syntax (Also Works):**
```kukicha
func (todo Todo) Display() string
    return "{todo.title}"

func (todo *Todo) MarkDone()
    todo.completed = true
```

---

#### Goroutines

**Primary (Recommended):**
```kukicha
go doWork()
go fetchData(url)
```

**Go Syntax (Same):**
```kukicha
go doWork()
go fetchData(url)
```

Both use `go` keyword - it's already intuitive!

---

#### Channels

**Primary (Recommended):**
```kukicha
ch := make channel of string
send ch, "message"
msg := receive ch
close ch
```

**Go Syntax (Also Works):**
```kukicha
ch := make(chan string)
ch <- "message"
msg := <-ch
close(ch)
```

---

#### Error Handling

**Primary (Recommended):**
```kukicha
# Or operator (auto-unwrap)
data := file.read(path) onerr panic "file not found"
user := parseUser(data) onerr return error "invalid user"
port := env.get("PORT") onerr "8080"
```

**Go Syntax (Also Works):**
```kukicha
data, err := file.read(path)
if err != empty
    panic "file not found"
```

---

### Alternative Syntax Converted by `kuki fmt`

#### Assignment Operators (Walrus is Core)

```kukicha
# Kukicha only - these have distinct meanings
count := 0      # Create new binding
count = 5       # Update existing
```

The distinction between `:=` (create) and `=` (update) is a core Kukicha feature that does not accept Go's shadowing semantics.

---

#### Braces, Semicolons (Converted to Standard)

```kukicha
# Accepted but converted by kuki fmt
if x == 5 { print "yes" }       # Braces converted to indentation
var x int = 5;                  # Semicolons removed

# Standard Kukicha (after kuki fmt)
if x equals 5
    print "yes"
x := int(5)
```

**Formatting Policy:**
- **Braces `{}`**: Accepted by parser, but `kuki fmt` converts to indentation
- **Semicolons `;`**: Removed by `kuki fmt`
- **Indentation**: 4 spaces (canonical)
- **Snake_case**: Discouraged (use camelCase), but not enforced by formatter

---

### Complete Examples: Side by Side

#### Example 1: Todo Processing with Error Handling

**Newbie-Friendly Syntax (Recommended):**
```kukicha
func ProcessTodos(path string)
    todos := file.read(path) 
        onerr return empty list of Todo
    
    parsed := json.parse(todos) as list of Todo
        onerr return empty list of Todo
    
    results := empty list of Todo
    for discard, todo in parsed
        if todo.completed and not todo.deleted
            results = append(results, todo)
    return results

func Display on this Todo
    status := "○"
    if this.completed
        status = "✓"
    return "{status} {this.title}"
```

**Go-Style Syntax (Also Works):**
```kukicha
func ProcessTodos(path string) []Todo
    todos, err := file.read(path)
    if err != nil
        return []Todo{}
    
    var parsed []Todo
    err = json.parse(todos, &parsed)
    if err != nil
        return []Todo{}
    
    results := []Todo{}
    for _, todo := range parsed
        if todo.completed && !todo.deleted
            results = append(results, todo)
    return results

func (todo Todo) Display() string
    status := "○"
    if todo.completed
        status = "✓"
    return status + " " + todo.title
```

**Both compile to identical Go code.**

---

#### Example 2: Concurrent Data Fetching

**Newbie-Friendly Syntax (Recommended):**
```kukicha
func FetchAll(urls list of string)
    results := make channel of string, len(urls)
    
    for discard, url in urls
        go
            response := http.get(url) onerr return
            send results, response.body
    
    allData := empty list of string
    for i from 0 to len(urls)
        data := receive results
        allData = append(allData, data)
    
    return allData
```

**Go-Style Syntax (Also Works):**
```kukicha
func FetchAll(urls []string) []string
    results := make(chan string, len(urls))
    
    for _, url := range urls
        go func(u string)
            response, err := http.get(u)
            if err != nil
                return
            results <- response.body
        (url)
    
    allData := []string{}
    for i := 0; i < len(urls); i++
        data := <-results
        allData = append(allData, data)
    
    return allData
```

---

#### Example 3: Safe File Processing with Defer

**Newbie-Friendly Syntax (Recommended):**
```kukicha
interface FileProcessor
    Process(content string) string

func SafeProcess(path string, processor FileProcessor)
    file := file.open(path) onerr return error "cannot open file"
    defer file.close()
    
    defer
        if r := recover(); r != empty
            print "Processing failed: {r}"
    
    content := file.read() onerr return error "cannot read file"
    result := processor.Process(content)
    return result, empty
```

**Go-Style Syntax (Also Works):**
```kukicha
type FileProcessor interface
    Process(content string) string

func SafeProcess(path string, processor FileProcessor) (string, error)
    file, err := file.open(path)
    if err != nil
        return "", fmt.Errorf("cannot open file")
    defer file.close()
    
    defer func()
        if r := recover(); r != nil
            fmt.Println("Processing failed:", r)
    ()
    
    content, err := file.read()
    if err != nil
        return "", fmt.Errorf("cannot read file")
    result := processor.Process(content)
    return result, nil
```

---

### Documentation Strategy

Throughout this specification, **primary syntax** (newbie-friendly) is featured because:
- It's designed for programming beginners
- It's more readable and self-documenting
- It teaches programming concepts clearly

However, **Go syntax compatibility** means:
- Experienced Go developers can use familiar patterns
- Copy-paste from Go tutorials often works
- Teams can mix experience levels
- You can transition gradually as you learn

**Start with readable syntax. Graduate to Go syntax when you're ready. Both work perfectly.**

---

## For Go Developers: Quick Migration

If you're coming from Go, you can paste Go-style code and use `kukicha fmt` to convert it to canonical Kukicha syntax:

```bash
kukicha fmt -w myfile.kuki
```

**What gets converted:**

| Go Style | Kukicha Style |
|----------|---------------|
| `if x == 5 {` ... `}` | `if (x equals 5)` + indentation |
| `&&` | `and` |
| `||` | `or` |
| `==` | `equals` |
| `!=` | `not equals` |
| `!x` | `not x` |
| Trailing `;` | removed |
| Braces `{}` | indentation blocks |

**Example:**

```go
// Go-style input
func greet(name string) {
    if name != "" {
        fmt.Println("Hello " + name);
    }
}
```

After `kukicha fmt -w`:

```kukicha
func greet(name string)
    if (name not equals "")
        fmt.Println("Hello " + name)
```

This lets you copy-paste Go snippets from tutorials and convert them to idiomatic Kukicha.

---

## See Also

- [Quick Reference](kukicha-quick-reference.md) - Fast lookup cheat sheet for developers
- [Grammar (EBNF)](kukicha-grammar.ebnf.md) - Formal grammar specification
- [Compiler Architecture](kukicha-compiler-architecture.md) - Transpiler implementation details
- [Standard Library Roadmap](kukicha-stdlib-roadmap.md) - Planned features and extensions
