# Kukicha Syntax Examples

## Complete Example: Todo App

```kukicha
petiole main

import "fmt"
import "stdlib/json"
import "os"
import "stdlib/files"

type Todo
    id int64
    title string
    completed bool
    tags list of string

func CreateTodo(id int64, title string) Todo
    return Todo
        id: id
        title: title
        completed: false
        tags: list of string{}

func Display on todo Todo string
    status := "[ ]"
    if todo.completed
        status = "[x]"
    return "{status} {todo.id}. {todo.title}"

func MarkComplete on todo reference Todo
    todo.completed = true

func AddTag on todo reference Todo tag string
    todo.tags = append(todo.tags, tag)

func SaveTodos(todos list of Todo, path string) error
    data := json.Marshal(todos) onerr return error
    files.Write(path, string(data)) onerr return error
    return empty

func LoadTodos(path string) list of Todo, error
    content := files.Read(path) onerr return empty list of Todo, error
    todos := list of Todo{}
    json.Unmarshal(list of byte(content), reference of todos) onerr return empty list of Todo, error
    return todos, empty

func main()
    todos := list of Todo{}

    todo1 := CreateTodo(1, "Learn Kukicha")
    todo2 := CreateTodo(2, "Build something cool")

    AddTag(reference of todo1, "learning")
    AddTag(reference of todo1, "programming")
    MarkComplete(reference of todo1)

    todos = append(todos, todo1, todo2)

    for todo in todos
        print(Display(todo))

    SaveTodos(todos, "todos.json") onerr panic "failed to save"
    print("Saved successfully!")
```

## Named Arguments and Default Parameters

### Default Parameter Values
```kukicha
# Parameters with defaults must come after required parameters
func Greet(name string, greeting string = "Hello", punctuation string = "!")
    print("{greeting}, {name}{punctuation}")

func main()
    Greet("Alice")                          # "Hello, Alice!"
    Greet("Bob", "Hi")                      # "Hi, Bob!"
    Greet("Charlie", "Welcome", ".")        # "Welcome, Charlie."
```

### Named Arguments
```kukicha
# Named arguments make code self-documenting
func CreateUser(name string, age int, active bool = true, role string = "user")
    # ...

func main()
    # Positional only
    CreateUser("Alice", 25, true, "admin")

    # With named arguments - much clearer!
    CreateUser("Bob", age: 30, role: "editor")

    # Named args can be in any order (after positional)
    CreateUser("Charlie", active: false, age: 22)
```

### Common Use: Configuration Functions
```kukicha
func Connect(host string, port int = 8080, timeout int = 30, retries int = 3)
    print("Connecting to {host}:{port} with {timeout}s timeout, {retries} retries")

func main()
    # Just specify what you need
    Connect("localhost")
    Connect("api.example.com", timeout: 60)
    Connect("db.local", port: 5432, retries: 5)
```

---

## Production-Ready Packages

### Validation with stdlib/validate
```kukicha
import "stdlib/validate"

func CreateUser(email string, age int, password string) (User, error)
    # Each validation returns (value, error) - perfect for onerr
    email |> validate.NotEmpty() onerr return User{}, error
    email |> validate.Email() onerr return User{}, error("invalid email format")

    age |> validate.InRange(18, 120) onerr return User{}, error("age must be 18-120")

    password |> validate.MinLength(8) onerr return User{}, error("password too short")
    password |> validate.Matches(`[A-Z]`) onerr return User{}, error("need uppercase")

    return User{email: email, age: age}, empty
```

### Environment Config with stdlib/env and stdlib/must
```kukicha
import "stdlib/env"
import "stdlib/must"

func main()
    # Startup config - panic is acceptable (use must)
    apiKey := must.Env("API_KEY")  # Panics if not set
    dbUrl := must.EnvOr("DATABASE_URL", "localhost:5432")

    # Runtime config - use onerr (use env)
    port := env.GetIntOr("PORT", 8080) onerr 8080
    debug := env.GetBoolOrDefault("DEBUG", false)
    hosts := env.GetListOr("ALLOWED_HOSTS", ",", empty list of string)
```

### HTTP Helpers with stdlib/http
```kukicha
import "stdlib/http" as httphelper

func HandleCreateUser(w http.ResponseWriter, r reference http.Request)
    # Parse JSON body easily
    input := UserInput{}
    httphelper.ReadJSON(r, reference of input) onerr
        return httphelper.JSONBadRequest(w, "Invalid JSON")

    # Validate
    user := CreateUser(input.email, input.age, input.password) onerr
        return httphelper.JSONBadRequest(w, "{error}")

    # Respond with JSON
    httphelper.JSONCreated(w, user)

func HandleGetUser(w http.ResponseWriter, r reference http.Request)
    # Query params
    id := httphelper.GetQueryInt(r, "id") onerr
        return httphelper.JSONBadRequest(w, "id parameter required")

    user := db.GetUser(id) onerr
        return httphelper.JSONNotFound(w, "User not found")

    httphelper.JSON(w, user)
```

### Time Handling with stdlib/datetime
```kukicha
import "stdlib/datetime"

func main()
    now := datetime.Now()

    # Named formats - no more "2006-01-02T15:04:05Z07:00"!
    print(datetime.Format(now, "iso8601"))   # 2024-01-15T14:30:00Z
    print(datetime.Format(now, "date"))      # 2024-01-15
    print(datetime.Format(now, "datetime"))  # 2024-01-15 14:30:00

    # Duration helpers
    timeout := datetime.Seconds(30)
    cacheTime := datetime.Hours(24)

    # Date arithmetic
    tomorrow := datetime.AddDays(now, 1)
    lastWeek := datetime.SubDays(now, 7)

    # Comparisons
    if datetime.IsPast(deadline)
        print("Deadline passed!")

    if datetime.IsToday(event)
        print("Event is today!")
```

### Safe Slice Access with stdlib/slice
```kukicha
import "stdlib/slice"

func main()
    items := list of string{"a", "b", "c"}

    # Safe access - never panics
    first := slice.FirstOr(items, "default")
    last := slice.LastOr(items, "default")

    # Get with negative indexing
    item := slice.GetOr(items, -1, "default")  # Last item or default

    # Find with predicate
    user := slice.Find(users, func(u User) bool
        return u.active
    ) onerr return defaultUser

    # Check if empty
    if slice.IsEmpty(items)
        return
```

---

## Error Handling Patterns

### Pattern 1: Panic on Critical Errors
```kukicha
func mustLoadConfig() Config
    data := os.ReadFile("config.json") onerr panic "config missing"
    config := Config{}
    json.Unmarshal(data, reference of config) onerr panic "invalid config"
    return config
```

### Pattern 2: Propagate Errors
```kukicha
func loadUser(id int64) User, error
    data := db.Query("SELECT * FROM users WHERE id = ?", id) onerr return empty User, error
    user := User{}
    parseResult(data, reference of user) onerr return empty User, error
    return user, empty
```

### Pattern 3: Default Values
```kukicha
func getPort() int
    portStr := os.Getenv("PORT") onerr "8080"
    port := strconv.Atoi(portStr) onerr 8080
    return port
```

### Pattern 4: Custom Error Messages
```kukicha
func validateAge(age int) error
    if age < 0
        return error "age cannot be negative"
    if age > 150
        return error "age seems unrealistic"
    return empty
```

## Pipe Operator Chains

### Data Processing Pipeline
```kukicha
import "stdlib/slice"
import "stdlib/slice"
import "stdlib/json"
import "stdlib/fetch"

type User
    name string
    age int
    active bool

func getActiveAdultNames(url string) list of string
    users := list of User{}
    url
        |> fetch.Get()
        |> fetch.CheckStatus()
        |> fetch.Bytes()
        |> json.Unmarshal(reference of users)
        onerr return empty list of User
    
    return users
        |> slice.Filter(func(u User) bool
            return u.active and u.age >= 18
        )
        |> slice.Map(func(u User) string
            return u.name
        )
        |> slice.Sort()
```

### Pipe with Placeholder (Advanced)

When the piped value isn't the first argument, use `_` as a placeholder:

```kukicha
import "encoding/json"
import "io"

type Todo
    id int64
    title string

# Without placeholder: data |> func() makes data the FIRST arg
# With placeholder: data |> func(x, _) puts data where _ appears

func writeTodoAsJson(w io.Writer, todo Todo) error
    # Placeholder puts 'todo' as second arg to json.NewEncoder().Encode()
    return todo |> json.NewEncoder(w).Encode(_)

# Another example: writing to specific position in args
func formatWithPrefix(prefix string, data string) string
    return data |> fmt.Sprintf("{prefix}: %s", _)

# Multiple args with placeholder in the middle
func processWithOptions(opts Options, data Data, format Format) Result
    return data |> transform(opts, _, format)
```

Transpilation:
- `todo |> json.NewEncoder(w).Encode(_)` → `json.NewEncoder(w).Encode(todo)`
- `data |> transform(opts, _, format)` → `transform(opts, data, format)`

### File Processing Pipeline
```kukicha
import "stdlib/files"
import "strings"

func processLogFile(path string) list of string
    return path
        |> files.Read()
        |> strings.Split("\n")
        |> slice.Filter(func(line string) bool
            return strings.Contains(line, "ERROR")
        )
        |> slice.Map(func(line string) string
            return strings.TrimSpace(line)
        )
```

## Interface Implementation

```kukicha
interface Formatter
    Format() string

type PlainFormatter
    prefix string

func Format on f PlainFormatter string
    return "{f.prefix}: plain format"

type JSONFormatter
    indent int

func Format on f JSONFormatter string
    return "{\"type\": \"json\", \"indent\": {f.indent}}"

func PrintFormatted(f Formatter)
    print(f.Format())

func main()
    plain := PlainFormatter{prefix: "LOG"}
    jsonFmt := JSONFormatter{indent: 2}

    PrintFormatted(plain)
    PrintFormatted(jsonFmt)
```

## Concurrency Patterns

### Worker Pool
```kukicha
func worker(id int, jobs channel of int, results channel of int)
    for job in jobs
        print("Worker {id} processing job {job}")
        send results, job * 2

func main()
    jobs := make channel of int, 100
    results := make channel of int, 100

    # Start 3 workers
    for w from 1 through 3
        go worker(w, jobs, results)

    # Send 5 jobs
    for j from 1 through 5
        send jobs, j
    close(jobs)

    # Collect results
    for a from 1 through 5
        result := receive from results
        print("Result: {result}")
```

### Fan-Out/Fan-In
```kukicha
func producer(out channel of int)
    for i from 0 to 10
        send out, i
    close(out)

func square(in channel of int, out channel of int)
    for n in in
        send out, n * n
    close(out)

func main()
    numbers := make channel of int
    squares := make channel of int

    go producer(numbers)
    go square(numbers, squares)

    for result in squares
        print(result)
```

## Generics via stdlib/iter and stdlib/slice

Kukicha uses special transpilation for Go 1.25+ generics without requiring you to write generic syntax:

### Functional Iteration (iter package)

```kukicha
import "stdlib/iter"
import "stdlib/slice"

func main()
    numbers := list of int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

    # Filter and transform
    result := numbers
        |> iter.FromSlice()
        |> iter.Filter(func(n int) bool
            return n % 2 equals 0
        )
        |> iter.Map(func(n int) int
            return n * n
        )
        |> iter.Take(3)
        |> slice.Collect()

    # result is [4, 16, 36]
```

### Grouping with Comparable Constraint (slice.GroupBy)

```kukicha
import "stdlib/slice"

type LogEntry
    timestamp int64
    level string      # ERROR, WARN, INFO
    message string

type DayStats
    day string
    count int

func analyzeLog(entries list of LogEntry) map of string to list of LogEntry
    # GroupBy automatically generates: GroupBy[LogEntry, string](items, keyFunc)
    # The key type (string) must be comparable - the transpiler ensures this!
    return entries
        |> slice.GroupBy(func(e LogEntry) string
            return e.level
        )

func main()
    logs := list of LogEntry{
        LogEntry{timestamp: 1000, level: "ERROR", message: "Connection failed"},
        LogEntry{timestamp: 1001, level: "WARN", message: "Retry attempt"},
        LogEntry{timestamp: 1002, level: "ERROR", message: "Timeout"},
        LogEntry{timestamp: 1003, level: "INFO", message: "Recovered"},
    }

    grouped := analyzeLog(logs)

    for level, entries in grouped
        print("Level {level}: {len(entries)} entries")
```

## Testing in Kukicha

```kukicha
petiole mypackage

import "testing"

func TestAddition(t reference testing.T)
    result := Add(2, 3)
    if result != 5
        t.Errorf("expected 5, got {result}")

func TestDivision(t reference testing.T)
    result, err := Divide(10, 2)
    if err != empty
        t.Fatalf("unexpected error: {err}")
    if result != 5
        t.Errorf("expected 5, got {result}")

func TestDivisionByZero(t reference testing.T)
    _, err := Divide(10, 0)
    if err equals empty
        t.Error("expected error for division by zero")
```

## Struct Embedding

```kukicha
type Animal
    name string
    age int

type Dog
    Animal           # Embedded struct
    breed string

func Speak on a Animal string
    return "{a.name} makes a sound"

func Speak on d Dog string
    return "{d.name} barks!"

func main()
    dog := Dog
        Animal: Animal
            name: "Buddy"
            age: 3
        breed: "Golden Retriever"

    print(dog.name)      # Access embedded field
    print(dog.Speak())   # Calls Dog's Speak method
```

## JSON Handling

```kukicha
import "encoding/json"

type APIResponse
    status string json:"status"
    data list of User json:"data"
    count int json:"count"

func parseResponse(jsonData list of byte) APIResponse, error
    response := APIResponse{}
    json.Unmarshal(jsonData, reference of response) onerr return empty APIResponse, error
    return response, empty

func toJSON(data APIResponse) string
    bytes := json.MarshalIndent(data, "", "  ") onerr return "{}"
    return string(bytes)
```

## Function Types (Callbacks & Higher-Order Functions)

### Map/Filter/Reduce Pattern

```kukicha
# Generic filter function
func Filter(items list of int, predicate func(int) bool) list of int
    result := list of int{}
    for item in items
        if predicate(item)
            result = append(result, item)
    return result

# Generic map function
func Map(items list of int, transform func(int) int) list of int
    result := list of int{}
    for item in items
        result = append(result, transform(item))
    return result

# Generic reduce function
func Reduce(items list of int, initial int, reducer func(int, int) int) int
    accumulator := initial
    for item in items
        accumulator = reducer(accumulator, item)
    return accumulator

func main()
    numbers := list of int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

    # Filter for even numbers
    evens := Filter(numbers, func(n int) bool
        return n % 2 equals 0
    )

    # Map to squares
    squares := Map(evens, func(n int) int
        return n * n
    )

    # Sum all squares
    total := Reduce(squares, 0, func(acc int, n int) int
        return acc + n
    )

    print(total)  # 4 + 16 + 36 + 64 + 100 = 220
```

### ForEach Pattern

```kukicha
func ForEach(items list of string, action func(string))
    for item in items
        action(item)

func main()
    names := list of string{"Alice", "Bob", "Charlie"}

    ForEach(names, func(name string)
        print("Hello, {name}!")
    )
```

### Complex Callback Example

```kukicha
type User
    name string
    age int

# Callback-based filtering
func FilterUsers(users list of User, check func(reference User) bool) list of User
    result := list of User{}
    for user in users
        if check(reference of user)
            result = append(result, user)
    return result

# Callback-based transformation
func TransformUsers(users list of User, transform func(reference User) string) list of string
    result := list of string{}
    for user in users
        result = append(result, transform(reference of user))
    return result

func main()
    users := list of User{
        User
            name: "Alice"
            age: 30
        User
            name: "Bob"
            age: 25
        User
            name: "Charlie"
            age: 35
    }

    # Filter adults (age >= 30)
    adults := FilterUsers(users, func(u reference User) bool
        return u.age >= 30
    )

    # Transform to display names
    names := TransformUsers(adults, func(u reference User) string
        return "{u.name} ({u.age})"
    )

    for name in names
        print(name)
```

### Async Operations with Callbacks

```kukicha
# Execute operation with success/error callbacks
func DoAsync(operation func() string, onSuccess func(string), onError func(error))
    result := operation()
    onSuccess(result)

func main()
    DoAsync(
        func() string
            return "Success!"
        ,
        func(msg string)
            print("Result: {msg}")
        ,
        func(err error)
            print("Error: {err}")
    )
```
