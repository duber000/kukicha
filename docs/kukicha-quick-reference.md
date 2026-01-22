# Kukicha Quick Reference Guide

**Version:** 1.0.0 | **File Extension:** `.kuki` | **Target:** Go 1.25+

> This is a quick reference for developers already familiar with Kukicha. For a complete guide with detailed explanations, see the [Language Syntax Reference](kukicha-syntax-v1.0.md).

---

## Basics

### Module Structure

```
myapp/              # Stem (module root)
  stem.toml        # Module config
  models/          # Petiole (package)
    user.kuki      # Source file
    todo.kuki
  api/
    handlers.kuki
```

### Petiole Declaration (Optional)

```kukicha
# Optional - if absent, petiole calculated from file path
petiole models.todo

import time
import github.com/user/repo
```

**New:** Petiole declarations are optional! Package name is automatically calculated from the directory path relative to `stem.toml`.

---

## Types & Variables

### Type Declaration

```kukicha
type Todo
    id int64
    title string
    completed bool
    settings reference Settings
    tags list of string
    metadata map of string to string
```

### Function Types

```kukicha
# Function type syntax: func(params) return_type
func Filter(items list of int, predicate func(int) bool) list of int
func Map(items list of int, transform func(int) int) list of int
func ForEach(items list of string, action func(string))

# Use with lambdas
evens := Filter(numbers, func(n int) bool
    return n % 2 == 0
)
```

**Note:** Use `func` for both declarations and types (consistent with Go).

### Variables

```kukicha
# Create (walrus operator)
result := calculate()
todos := empty list of Todo
user := empty reference User

# Update
result = newValue
user = reference to newUser
```

### Exports

```kukicha
type Todo              # Exported (PascalCase)
type internalCache     # Private (camelCase)

func CreateTodo()      # Exported
func validateInput()   # Private
```

---

## Functions & Methods

### Functions

```kukicha
# EXPLICIT types required for params and returns
func CreateTodo(id int64, title string, description string) Todo
    return Todo
        id: id
        title: title
        completed: false

func ProcessData(data list of User) int
    # Local variables inferred
    count := len(data)
    return count
```

**New:** Signature-first type inference - explicit types on parameters/returns, inference only for locals.

### Methods

```kukicha
# Value receiver - explicit receiver name
func Display on todo Todo string
    return "{todo.id}. {todo.title}"

# Reference receiver - for mutation
func MarkDone on todo reference Todo
    todo.completed = true

# With parameters
func UpdateTitle on todo reference Todo, newTitle string
    todo.title = newTitle

# Receiver is just a parameter - no special 'this' or 'self'
func Summary on t Todo string
    return t.title
```

### Interfaces

```kukicha
interface Displayable
    Display() string
    GetTitle() string

# Implicit implementation - just add methods
func Display on todo Todo string
    return todo.title
# Todo now implements Displayable
```

---

## Control Flow

### If/Else

```kukicha
if condition
    doSomething()
else if otherCondition
    doOther()
else
    doDefault()
```

### Loops

```kukicha
# Exclusive range (0-9)
for i from 0 to 10
    print i

# Inclusive range (1-10)
for i from 1 through 10
    print i

# Collection iteration
for item in items
    process(item)

for key, value in config
    print "{key}: {value}"

# With discard
for discard, item in items
    process(item)
```

---

## Error Handling

### OnErr Operator (Recommended)

```kukicha
# Auto-unwrap and handle errors
data := file.read("config.json") onerr panic "missing config"
user := parseUser(data) onerr return error "invalid user"
port := env.get("PORT") onerr "8080"

# Chaining
config := file.read("config.json")
    onerr file.read("config.yaml")
    onerr panic "no config found"
```

**Note:** The `onerr` keyword is distinct from `or` (boolean logic), making error handling visually clear.

### Explicit Handling

```kukicha
data, err := file.read("config.json")
if err != empty
    return err
```

---

## Collections

### Lists

```kukicha
# Create
todos := empty list of Todo
numbers := empty list of int

# Access by index (positive)
first := items at 0
first := items[0]

# LITERAL negative indices (compile-time optimized)
last := items at -1         # Zero runtime overhead
secondLast := items[-2]      # Compiled to len(items)-2

# DYNAMIC negative indices (runtime)
index := -2
element := items.at(index)   # Use .at() method for variables

# Slicing uses Go syntax
subset := items[2:7]

# Slicing with LITERAL negative indices (compile-time optimized)
lastThree := items[-3:]      # Zero overhead
allButLast := items[:-1]
middle := items[1:-1]

# Append
items = append(items, newItem)

# Length
count := len(items)

# Membership test
if item in items
    print "found"

if item not in blacklist
    process(item)
```

### Maps

```kukicha
# Create
config := map of string to string
    host: "localhost"
    port: "8080"

# Access
value := config at "host"
value := config["host"]  # Go syntax

# Set
config at "port" = "9000"

# Check existence (traditional)
value, exists := config at "host"
if exists
    print value

# Membership test (recommended)
if "host" in config
    connect(config at "host")

if "api_key" not in config
    print "Warning: missing API key"
```

---

## Pipe Operator

### Data Pipelines

```kukicha
# Pass result to next function
result := data
    |> parse()
    |> transform()
    |> process()

# Basic usage — result becomes first argument
x |> func(y, z)  # Same as: func(x, y, z)

# Method calls — use leading dot
response |> .json()  # Same as: response.json()
```

### Common Patterns

```kukicha
# HTTP pipeline
users := http.get(url)
    |> .json() as list of User
    |> filterActive()
    |> sortByName()

# File processing
content := file.read("data.csv")
    |> string.split("\n")
    |> parseCSV()
    |> aggregate()

# With error handling
config := file.read("config.json")
    |> json.parse()
    |> validate()
    onerr return error "invalid"
```

### Precedence

```kukicha
# Operators bind tighter than pipe
a + b |> double()  # Same as: (a + b) |> double()

# OnErr operator has lower precedence
fetch() |> parse() onerr "default"
# Same as: (fetch() |> parse()) onerr "default"
```

---

## Strings

### Interpolation

```kukicha
# Variables in strings
message := "Hello {name}, you have {count} messages"
status := "{id}: {title} - {completed}"

# String operations
upper := string.upper("hello")
parts := string.split("a,b,c", ",")
joined := string.join(parts, "|")
```

---

## Concurrency

### Goroutines

```kukicha
# Launch concurrent task
go doWork()
go fetchData(url)

# Anonymous function
go
    result := calculate()
    print result
```

### Channels

```kukicha
# Create
ch := make channel of string
buffered := make channel of int, 100

# Send/Receive (Recommended)
send ch, "message"
msg := receive ch

# Go syntax also works
ch <- "message"
msg := <-ch

# Close
close ch
```

---

## Defer & Recover

### Defer (Cleanup)

```kukicha
func processFile(path)
    file := file.open(path) onerr return error
    defer file.close()  # Always runs
    
    # Work with file
```

### Recover (Panic Handling)

```kukicha
func safeOperation()
    defer
        if r := recover(); r != empty
            print "Recovered: {r}"
    
    riskyOperation()
```

---

## Membership Testing

```kukicha
# Check if item exists in list
if user in admins
    grantAccess()

if item not in blacklist
    process(item)

# Check if key exists in map
if "host" in config
    connect(config at "host")

if "DEBUG" not in environment
    print "Production mode"

# Check if substring in string
if "error" in logMessage
    alertOps()
```

---

## Negative Indexing & Slicing

```kukicha
# Access from end
last := items at -1
secondLast := items[-2]

# Slicing with negatives
lastThree := items[-3:]       # Last 3 elements
allButLast := items[:-1]      # All except last
middle := items[1:-1]         # Remove first and last
mixed := items[2:-3]          # Mix positive and negative
```

---

## Comparison & Logic

### Operators

```kukicha
# Comparison
if x equals 5           # Equality
if x == 5              # Go syntax
if x not equals 5      # Inequality
if x != 5              # Inequality (alternative)
if x > 5, x < 5
if x >= 5, x <= 5

# Membership
if item in items        # Check existence
if key not in map       # Check absence

# Boolean logic
if completed and not expired
if active or archived
if ready && !paused   # Go syntax
```

---

## Zero Values

```kukicha
empty                        # nil/zero
nil                          # Alias for empty
empty reference Type         # nil pointer
empty list of Type          # empty slice
empty map of K to V         # empty map
false                       # bool zero
0                           # numeric zero
""                          # empty string
```

---

## Type Casting

```kukicha
id := int64(rawValue)
count := int32(value)
text := string(bytes)
num := float64(integer)
```

---

## Common Patterns

### Constructor Function

```kukicha
func CreateTodo(id, title)
    return Todo
        id: id
        title: title
        completed: false
        created_at: time.now()
```

### Error with Cleanup

```kukicha
func process(path)
    file := file.open(path) onerr return error "cannot open"
    defer file.close()
    
    data := file.read() onerr return error "cannot read"
    return data, empty
```

### Concurrent Processing

```kukicha
func processAll(items list of Item)
    results := make channel of Result, len(items)
    
    for discard, item in items
        go
            result := process(item)
            send results, result
    
    for i from 0 to len(items)
        result := receive results
        print result
```

### Interface Usage

```kukicha
interface Processor
    Process() string

func RunAll(processors list of Processor)
    for discard, p in processors
        result := p.Process()
        print result
```

---

## Dual Syntax Support

Most Kukicha syntax has a Go equivalent that also works:

| Kukicha | Go Syntax |
|---------|-----------|
| `equals` | `==` |
| `and`, `or`, `not` | `&&`, `||`, `!` |
| `in` | `slices.Contains()` (lists), map idiom |
| `not in` | Negated membership |
| `list of Type` | `[]Type` |
| `map of K to V` | `map[K]V` |
| `reference Type` | `*Type` |
| `items at 0` | `items[0]` |
| `items at -1` | `items[len(items)-1]` |
| `items[-3:]` | `items[len(items)-3:]` |
| `items[:-1]` | `items[:len(items)-1]` |
| `discard` | `_` |
| `send ch, val` | `ch <- val` |
| `receive ch` | `<-ch` |
| `make channel of T` | `make(chan T)` |
| `for x in items` | `for _, x := range items` |
| `for i, x in items` | `for i, x := range items` |
| `for discard, x in items` | `for _, x := range items` |

**Note:** The walrus operator `:=` is always create-only in Kukicha (unlike Go's shadowing behavior).

---

## File Extension

**Kukicha files:** `.kuki` (茎 = "stem" in Japanese)

```
myapp/
  models/
    user.kuki
    todo.kuki
  api/
    handlers.kuki
```

---

## Build & Run

```bash
# Compile
kukicha build main.kuki

# With Green Tea GC
kukicha build --experiment greenteagc

# Run
kukicha run main.kuki

# Test
kukicha test
```

---

## Code Formatting

```bash
# Format a single file (output to stdout)
kukicha fmt myfile.kuki

# Format and write back to file
kukicha fmt -w myfile.kuki

# Format all files in directory
kukicha fmt -w ./src/

# Check formatting (CI/CD) - exit 1 if not formatted
kukicha fmt --check ./src/
```

`kukicha fmt` converts all code to canonical indentation-based syntax (4 spaces, no braces, no semicolons) and preserves comments.

---

## Getting Help

```bash
kukicha help
kukicha version
kukicha doc [topic]
```

---

## See Also

- [Language Syntax Reference](kukicha-syntax-v1.0.md) - Complete syntax guide with detailed explanations
- [Grammar (EBNF)](kukicha-grammar.ebnf.md) - Formal grammar specification
- [Compiler Architecture](kukicha-compiler-architecture.md) - Implementation details
- [Standard Library Roadmap](kukicha-stdlib-roadmap.md) - Planned features
