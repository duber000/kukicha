# Kukicha Language Features

This document covers the complete feature set of Kukicha v1.0.0.

## Core Design Decisions

Kukicha v1.0.0 introduces key refinements that balance simplicity, performance, and consistency:

### 1. Optional Petiole Declarations

Folder-based package model with automatic Petiole (package) calculation from file path. No more header/directory sync issues!

```kukicha
# Explicit (optional)
petiole mypackage

# Or let Kukicha infer from directory structure
# A file at myproject/handlers/user.kuki automatically becomes package "handlers"
```

### 2. Signature-First Type Inference

Explicit types required for function parameters and returns; inference only for local variables. Maintains Go's performance while reducing boilerplate.

```kukicha
# Parameters and returns: explicit types required
func Calculate(x int, y int) int
    result := x + y    # Local variable: type inferred
    return result
```

### 3. Literal vs Dynamic Indexing

Negative indices with literal constants compile to zero-overhead code. Dynamic indices require explicit `.at()` method.

```kukicha
items := list of string{"a", "b", "c"}
last := items[-1]      # Compiles to items[len(items)-1] - zero overhead
# idx := -1
# item := items.at(idx)  # Dynamic index - runtime check
```

### 4. Indentation as Canonical

The `kukicha fmt` tool converts all code to standard 4-space indentation format, preventing "dialect drift" between coding styles.

### 5. Context-Sensitive Type Keywords

`list`, `map`, and `channel` are context-sensitive. In type contexts, they start composite types.

```kukicha
items list of string           # Type context: 'list' is keyword
list := getItems()             # Variable context: 'list' is identifier
```

### 6. Explicit Receiver Names

Methods use explicit receiver naming, following Go's philosophy that "methods are just functions".

```kukicha
func Display on todo Todo string
    return "{todo.id}: {todo.title}"

func SetTitle on t reference Todo title string
    t.title = title
```

### 7. Unified `func` Syntax

Use `func` for both declarations and types (like Go). Simple, consistent, one keyword to learn.

---

## Variables

### The Walrus Operator

```kukicha
# Create new binding
count := 42
name := "Alice"
active := true

# Reassign existing variable
count = 100
name = "Bob"
```

### Multiple Assignment

```kukicha
# Multiple variables
a, b := 1, 2

# Swap values
a, b = b, a

# From function returns
value, err := getData()
```

---

## Functions

### Basic Functions

```kukicha
# Function with explicit parameter and return types
func Greet(name string) string
    return "Hello {name}"

# Multiple parameters
func Add(a int, b int) int
    return a + b

# No return value
func PrintMessage(msg string)
    fmt.Println(msg)

# Multiple return values
func Divide(a int, b int) (int, error)
    if b equals 0
        return 0, error "division by zero"
    return a / b, empty
```

### Variadic Functions

```kukicha
func Sum(numbers many int) int
    total := 0
    for n in numbers
        total = total + n
    return total

# Call with multiple arguments
result := Sum(1, 2, 3, 4, 5)
```

### Methods

```kukicha
type Todo
    id int64
    title string
    completed bool

# Value receiver (cannot modify)
func Display on todo Todo string
    return "{todo.id}: {todo.title}"

# Pointer receiver (can modify)
func Complete on todo reference Todo
    todo.completed = true

# Usage
todo := Todo{id: 1, title: "Learn Kukicha", completed: false}
fmt.Println(todo.Display())
todo.Complete()
```

### Function Types (Callbacks)

```kukicha
# Function as parameter
func Filter(items list of int, predicate func(int) bool) list of int
    result := list of int{}
    for item in items
        if predicate(item)
            result = append(result, item)
    return result

# Lambda syntax
evens := Filter(numbers, func(n int) bool
    return n % 2 equals 0
)

# Arrow syntax for simple lambdas
evens := Filter(numbers, n -> n % 2 equals 0)
```

---

## Types

### Primitive Types

```kukicha
# Numbers
count int           # int, int8, int16, int32, int64
size uint           # uint, uint8, uint16, uint32, uint64
price float64       # float32, float64
char byte           # alias for uint8
code rune           # alias for int32 (Unicode code point)

# Boolean
active bool

# String
name string
```

### Struct Types

```kukicha
type User
    id int64
    name string
    email string
    active bool

# Struct literal
user := User
    id: 1
    name: "Alice"
    email: "alice@example.com"
    active: true

# Or inline
user := User{id: 1, name: "Alice", email: "alice@example.com", active: true}
```

### Struct Tags

Add metadata to struct fields using `key:"value"` syntax. Tags enable JSON marshaling, database mapping, and other Go reflection-based features.

```kukicha
type User
    ID int64 json:"id"
    Name string json:"name"
    Email string json:"email" db:"user_email"
    Active bool json:"active"

# Transpiles to Go with proper struct tags:
# type User struct {
#     ID     int64  `json:"id"`
#     Name   string `json:"name"`
#     Email  string `json:"email" db:"user_email"`
#     Active bool   `json:"active"`
# }

# Now compatible with encoding/json:
jsonData := `{"id": 1, "name": "Alice", "email": "alice@example.com", "active": true}`
user := User{}
json.Unmarshal([]byte(jsonData), &user)  # Works correctly with tags
```

Struct tags support any Go tag format:
- `json:"fieldname"` - JSON marshaling
- `xml:"fieldname"` - XML marshaling
- `db:"column_name"` - Database mapping
- `validate:"required"` - Validation rules
- Multiple tags: `json:"name" db:"user_name"`

### Collection Types

```kukicha
# List (slice)
items list of string
numbers list of int
users list of User

# Map
config map of string to string
scores map of string to int
cache map of int to User

# Channel
messages channel of string
results channel of int
```

### Pointer Types

```kukicha
# Pointer type declaration
userPtr reference User

# Address-of operator
user := User{name: "Alice"}
ptr := reference of user

# Dereference operator
val := dereference ptr
```

### Interface Types

```kukicha
interface Reader
    Read(p list of byte) (int, error)

interface Writer
    Write(p list of byte) (int, error)

interface ReadWriter
    Read(p list of byte) (int, error)
    Write(p list of byte) (int, error)
```

---

## Control Flow

### If Statements

```kukicha
# Basic if
if count > 0
    process()

# If-else
if active
    enable()
else
    disable()

# If-else if-else
if score >= 90
    grade := "A"
else if score >= 80
    grade := "B"
else if score >= 70
    grade := "C"
else
    grade := "F"

# Use 'equals' for equality
if status equals "active"
    proceed()
```

### For Loops

```kukicha
# Range over collection
for item in items
    process(item)

# Range with index
for index, item in items
    fmt.Println("{index}: {item}")

# Numeric range (exclusive end)
for i from 0 to 10
    fmt.Println(i)    # 0, 1, 2, ..., 9

# Numeric range (inclusive end)
for i from 1 through 5
    fmt.Println(i)    # 1, 2, 3, 4, 5

# Condition-based (while-style)
for count > 0
    count--

# Infinite loop
for true
    # ...
    if done
        break
```

---

## Error Handling

### The onerr Operator

```kukicha
# Panic on error
config := loadConfig() onerr panic "failed to load config"

# Return error to caller
data := fetchData() onerr return empty, error

# Provide default value
port := getPort() onerr 8080
name := getName() onerr "anonymous"

# Discard error (use sparingly)
result := riskyOperation() onerr discard
```

### Creating Errors

```kukicha
# Simple error
return error "something went wrong"

# Formatted error
return error "user {id} not found"

# Check for error
if err != empty
    return empty, err
```

### The empty Keyword

```kukicha
# empty is Kukicha's nil
if user equals empty
    return error "user not found"

# Return empty with type
return empty User, error "not found"

# Empty collections
return empty list of string
return empty map of string to int
```

---

## String Interpolation

```kukicha
name := "World"
greeting := "Hello {name}!"

# Expressions in interpolation
result := "Sum: {a + b}"
info := "User {user.name} has {len(user.items)} items"

# Compiles to fmt.Sprintf
# "Hello {name}!" -> fmt.Sprintf("Hello %s!", name)
```

---

## Operators

### Arithmetic

```kukicha
sum := a + b
diff := a - b
product := a * b
quotient := a / b
remainder := a % b
```

### Comparison

```kukicha
a equals b      # Equality (==)
a != b          # Not equal
a < b           # Less than
a <= b          # Less than or equal
a > b           # Greater than
a >= b          # Greater than or equal
```

### Logical

```kukicha
a and b         # Logical AND (&&)
a or b          # Logical OR (||)
not a           # Logical NOT (!)
```

### Membership

```kukicha
if item in collection
    # item exists in collection

if key in myMap
    # key exists in map
```

### Pipe

```kukicha
result := data
    |> parse()
    |> transform()
    |> validate()
    |> save()

# Equivalent to: save(validate(transform(parse(data))))
```

---

## Concurrency

### Goroutines

```kukicha
# Start a goroutine
go processInBackground(data)

# With anonymous function
go func()
    doWork()
()
```

### Channels

```kukicha
# Create channel
ch := make channel of string

# Buffered channel
ch := make channel of int, 100

# Send
send ch, "message"

# Receive
msg := receive ch

# Close
close(ch)

# Range over channel
for msg in ch
    process(msg)
```

### Defer

```kukicha
func processFile(path string)
    file := os.Open(path) onerr panic "cannot open"
    defer file.Close()
    # ... work with file
    # file.Close() called automatically when function returns
```

---

## Collections

### Lists (Slices)

```kukicha
# Create
items := list of string{"a", "b", "c"}
numbers := list of int{1, 2, 3, 4, 5}

# Empty with capacity
items := make(list of string, 0, 10)

# Access
first := items[0]
last := items[-1]       # Negative indexing

# Slice
subset := items[1:3]    # Elements 1 and 2
fromStart := items[:3]  # First 3 elements
toEnd := items[2:]      # From element 2 to end

# Append
items = append(items, "d")
items = append(items, "e", "f", "g")

# Length
count := len(items)
```

### Maps

```kukicha
# Create
config := map of string to string{
    "host": "localhost",
    "port": "8080",
}

# Empty
scores := make(map of string to int)

# Access
value := config["host"]

# Check existence
if key in config
    value := config[key]

# Set
config["timeout"] = "30"

# Delete
delete(config, "old_key")

# Length
count := len(config)
```

---

## Standard Library Usage

Kukicha uses Go's standard library directly:

```kukicha
import "fmt"
import "strings"
import "encoding/json"
import "net/http"

func main()
    # Use Go stdlib as normal
    fmt.Println("Hello!")

    upper := strings.ToUpper("hello")

    data := json.Marshal(user) onerr panic "marshal failed"
```

### Kukicha Standard Library

```kukicha
import "stdlib/iter"
import "stdlib/slice"
import "stdlib/string"
import "stdlib/fetch"
import "stdlib/files"
import "stdlib/parse"
```

See [Standard Library Roadmap](kukicha-stdlib-roadmap.md) for details.

---

## Complete Example

```kukicha
petiole main

import "fmt"
import "stdlib/slice"

type Task
    id int
    title string
    done bool

func Display on task Task string
    status := "[ ]"
    if task.done
        status = "[x]"
    return "{status} {task.id}. {task.title}"

func main()
    tasks := list of Task{
        Task{id: 1, title: "Learn Kukicha", done: true},
        Task{id: 2, title: "Build something", done: false},
        Task{id: 3, title: "Share with others", done: false},
    }

    fmt.Println("My Tasks:")
    for task in tasks
        fmt.Println(task.Display())

    pending := slice.Filter(tasks, t -> not t.done)
    fmt.Println("\nPending: {len(pending)} tasks")
```

**Output:**
```
My Tasks:
[x] 1. Learn Kukicha
[ ] 2. Build something
[ ] 3. Share with others

Pending: 2 tasks
```
