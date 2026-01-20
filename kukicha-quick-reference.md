# Kukicha Quick Reference Guide

**Version:** 1.0.0 | **File Extension:** `.kuki` | **Target:** Go 1.25+

---

## Basics

### Module Structure

```
myapp/              # Twig (module root)
  twig.toml        # Module config
  models/          # Stem (package)
    user.kuki      # Leaf (file)
    todo.kuki
  api/
    handlers.kuki
```

### Leaf Declaration

```kukicha
leaf models.todo
import time
import github.com/user/repo
```

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
func CreateTodo(id, title, description)
    return Todo
        id: id
        title: title
        completed: false

func ProcessData(data list of User)
    # Explicit types when needed
```

### Methods

```kukicha
# Value receiver
func Display on Todo
    return "{this.id}. {this.title}"

# Reference receiver
func MarkDone on reference Todo
    this.completed = true

# With parameters
func UpdateTitle on reference Todo, newTitle string
    this.title = newTitle
```

### Interfaces

```kukicha
interface Displayable
    Display() string
    GetTitle() string

# Implicit implementation - just add methods
func Display on Todo
    return this.title
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

### Or Operator (Recommended)

```kukicha
# Auto-unwrap and handle errors
data := file.read("config.json") or panic "missing config"
user := parseUser(data) or return error "invalid user"
port := env.get("PORT") or "8080"

# Chaining
config := file.read("config.json") 
    or file.read("config.yaml")
    or panic "no config found"
```

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
numbers := list of int

# Access
first := items at 0
second := items[1]  # Go syntax also works

# Slice
subset := items from 0 to 5
last := items from end to end

# Append
items = append(items, newItem)

# Length
count := len(items)
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

# Check existence
value, exists := config at "host"
if exists
    print value
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

# Left side becomes first argument
x |> func(y, z)  # Same as: func(x, y, z)
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
    or return error "invalid"
```

### Precedence

```kukicha
# Operators bind tighter than pipe
a + b |> double()  # Same as: (a + b) |> double()

# Or operator has lower precedence
fetch() |> parse() or "default"
# Same as: (fetch() |> parse()) or "default"
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
parts := "a,b,c" split ","
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
    file := file.open(path) or return error
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

## Comparison & Logic

### Operators

```kukicha
# Comparison
if x equals 5           # Equality
if x == 5              # Go syntax
if x != 5              # Inequality
if x > 5, x < 5
if x >= 5, x <= 5

# Boolean logic
if completed and not expired
if active or archived
if ready && !paused   # Go syntax
```

---

## Zero Values

```kukicha
empty                        # nil/zero
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
    file := file.open(path) or return error "cannot open"
    defer file.close()
    
    data := file.read() or return error "cannot read"
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
| `list of Type` | `[]Type` |
| `map of K to V` | `map[K]V` |
| `reference Type` | `*Type` |
| `items at 0` | `items[0]` |
| `discard` | `_` |
| `send ch, val` | `ch <- val` |
| `receive ch` | `<-ch` |
| `make channel of T` | `make(chan T)` |
| `for x in items` | `for _, x := range items` |

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

## Getting Help

```bash
kukicha help
kukicha version
kukicha doc [topic]
```

---

**Learn more:** https://kukicha.dev
**Report issues:** https://github.com/yourusername/kukicha
