# Building a Console Todo App with Kukicha

**Level:** Intermediate  
**Time:** 25 minutes  
**Prerequisite:** [Beginner Tutorial](beginner-tutorial.md)

Welcome back! In the beginner tutorial, you learned about variables, functions, and strings. Now we're going to build something real: a **todo list application** that runs in your terminal.

## What You'll Learn

In this tutorial, you'll discover how to:
- Store multiple items in **lists** (collections)
- Create **types** to organize related data
- Write **methods** that belong to types
- Use the **Pipe Operator (`|>`)** for clean data flow
- Handle **errors** gracefully with `onerr`
- **Save and load** data from files

Let's build something useful!

---

## What We're Building

Our todo app will let you:
- **Add** new tasks
- **View** all your tasks
- **Mark** tasks as done
- **Save** your tasks to a file
- **Load** them when you restart

Here's what it will look like when running:

```
=== Todo App ===
1. [ ] Buy groceries
2. [✓] Learn Kukicha
3. [ ] Call mom

Commands: add, done, list, save, quit
> 
```

---

## Step 0: Project Setup

If you haven't already, set up your project:

```bash
mkdir todo-app && cd todo-app
go mod init todo-app
kukicha init    # Extracts stdlib for imports like "stdlib/json"
```

---

## Step 1: Creating a Todo Type

> **📝 Reminder:** This tutorial builds on the beginner tutorial. Here are the key concepts you'll need:
> - **`:=`** creates a new variable, **`=`** updates an existing one
> - **String interpolation:** Use `{variable}` inside strings to insert values
> - **`print()`** outputs to the console
> - **Functions** take parameters and can return values
> - **Comments** start with `#`
>
> If you need a refresher, [revisit the beginner tutorial](beginner-tutorial.md)!

In the beginner tutorial, you learned about basic types like `string`, `int`, and `bool`. Now let's create our own type to represent a todo item.

Create a file called `todo.kuki`:

```kukicha
# A Todo represents a single task
# It has an id to identify it, a title describing the task,
# and completed tells us if it's done or not
type Todo
    id int
    title string
    completed bool
```

**What's happening here?**

We're defining a new **type** called `Todo`. Think of it as a blueprint for creating todo items. Each todo has:
- `id` - A number to identify this todo
- `title` - The text describing what needs to be done  
- `completed` - Whether the task is finished (`true` or `false`)

### Creating a Todo

Now let's write a function to create a new todo:

```kukicha
# CreateTodo makes a new todo with the given id and title
func CreateTodo(id int, title string) Todo
    return Todo
        id: id
        title: title
        completed: false
```

This function takes an `id` number and a `title` string, and returns a brand new `Todo` with `completed` set to `false` (because new tasks aren't done yet!).

---

## Step 2: Writing Methods

A **method** is a function that belongs to a type. Methods let you define what actions a type can perform.

In Kukicha, we use the `on` keyword to attach a method to a type:

```kukicha
# Display shows the todo in a nice format
# This method works "on" a Todo
func Display on todo Todo string
    status := "[ ]"
    if todo.completed
        status = "[✓]"
    return "{todo.id}. {status} {todo.title}"
```

**Reading this method:**
- `func Display` - We're creating a method called "Display"
- `on todo Todo` - This method works on a `Todo` (syntax: receiver name first, then the type). Inside the method, we call it `todo`
- `string` - The method returns a string

The method checks if the todo is completed. If so, it shows a checkmark. Otherwise, it shows empty brackets.

**💡 Tip:** When piping into a method that belongs to the value itself, use the dot shorthand:
```kukicha
# Calling directly:
message := todo.Display()

# Same thing, using pipe:
message := todo |> .Display()
```
This keeps the left-to-right data flow when chaining — and makes it clear the method belongs to the piped value, not an imported package.

### A Method That Changes Things

What if we want to mark a todo as done? We need a method that can **modify** the todo. For that, we use `reference`:

```kukicha
# MarkDone sets the todo as completed
# We use "reference" because we're changing the todo
func MarkDone on todo reference Todo
    todo.completed = true
```

**Why `reference`?**

Without `reference`, the method would get a **copy** of the todo. Any changes would only affect the copy, not the original. Using `reference` means we're working with the **actual** todo, so our changes stick.

---

## Step 3: Working with Lists

One todo is nice, but a todo **list** is what we really want! In Kukicha, we use `list of Type` to create a collection:

```kukicha
# Create an empty list of todos
todos := empty list of Todo

# Create some todos
todo1 := CreateTodo(1, "Buy groceries")
todo2 := CreateTodo(2, "Learn Kukicha")
todo3 := CreateTodo(3, "Call mom")

# Add them to the list
todos = append(todos, todo1)
todos = append(todos, todo2)
todos = append(todos, todo3)
```

### Looping Through a List

To go through each item in a list, use `for item in list`:

```kukicha
# Print all todos
for todo in todos
    message := todo.Display()
    print(message)
```

**Output:**
```
1. [ ] Buy groceries
2. [ ] Learn Kukicha
3. [ ] Call mom
```

### Finding Items in a List

Let's write a function to find a todo by its id:

```kukicha
# FindTodo searches for a todo by id
# Returns the todo and true if found, or empty and false if not
func FindTodo(todos list of Todo, id int) (Todo, bool)
    for todo in todos
        if todo.id equals id
            return todo, true
    return empty, false
```

Notice the **two return values**: the todo (if found) and a `bool` indicating success. This is a common pattern in Kukicha for operations that might fail.

---

## Step 4: Building the Todo List Type

Let's create a type to manage our entire todo list:

```kukicha
# TodoList manages a collection of todos
type TodoList
    items list of Todo
    nextId int
```

Now let's add methods to this type:

```kukicha
# Add creates a new todo and adds it to the list
# 'tl' is the receiver (the TodoList instance we're working on)
# Extra parameters go in parentheses after the receiver type
func Add on tl reference TodoList(title string)
    todo := CreateTodo(tl.nextId, title)
    tl.items = append(tl.items, todo)
    tl.nextId = tl.nextId + 1
    print("Added: {title}")
```

**Note on receiver naming:** We use `tl` as the receiver variable name (referring to the `TodoList` instance). Avoid using `list` — it's a keyword in Kukicha (as in `list of int`).

```kukicha
# ShowAll displays all todos in the list
func ShowAll on tl TodoList
    if len(tl.items) equals 0
        print("No todos yet! Use 'add' to create one.")
        return

    print("\n=== Your Todos ===")
    for todo in tl.items
        print(todo.Display())
    print("")
```

```kukicha
# Complete marks a todo as done by its id
func Complete on tl reference TodoList(id int)
    for i, todo in tl.items
        if todo.id equals id
            tl.items[i].completed = true
            print("Completed: {todo.title}")
            return
    print("Todo #{id} not found")
```

---

## Step 5: Error Handling with `onerr`

Real programs need to handle errors. What if the user types something that isn't a number? What if a file doesn't exist?

Kukicha makes error handling readable with the `onerr` keyword.

### Without `onerr`

```kukicha
# Manual error check - explicit but verbose
result, err := somethingThatMightFail()
if err not equals empty
    print("Something went wrong!")
    return
```

### With `onerr` (single expression)

```kukicha
# onerr handles the error in one line — best for simple cases
result := somethingThatMightFail() onerr "default value"
```

For multi-statement error handlers (like logging + returning), use the manual check above. `onerr` shines when the handler is a single expression: a default value, a `panic`, or a `return`.

### Common `onerr` Patterns

```kukicha
# Pattern 1: Provide a default value
name := getUserInput() onerr "Anonymous"

# Pattern 2: Panic (crash) with a message — good for startup config
config := loadConfig() onerr panic "Missing config file!"

# Pattern 3: Multi-statement block handler — for reporting and recovering
user := fetchUser(id) onerr
    log.Printf("Error fetching user {id}: {error}")
    return empty

# Pattern 4: Propagate the error to the caller
# error "{error}" wraps the original error in a new one
func DoWork() (string, error)
    data := loadFile() onerr return empty, error "{error}"
    return data, empty

# Pattern 4: When you need to do something before returning,
# use a manual error check instead:
func DoWorkVerbose() (string, error)
    data, err := loadFile()
    if err not equals empty
        print("Could not load file")
        return empty, err
    return data, empty
```

---

## Step 6: Saving and Loading from Files

Let's add the ability to save our todos to a file and load them back!

We'll use the `files` package from Kukicha's stdlib for easy file operations.

```kukicha
import "stdlib/files"
import "stdlib/string"
import "strconv"

# Save writes all todos to a file
func Save on tl TodoList(filename string) error
    lines := empty list of string

    for todo in tl.items
        # Format: id|title|completed
        # We use pipe (|) as a delimiter because titles can contain commas or spaces
        completed := "false"
        if todo.completed
            completed = "true"
        line := "{todo.id}|{todo.title}|{completed}"
        lines = append(lines, line)

    # Join all lines and write to file — pipe the data left to right
    lines
        |> string.Join("\n")
        |> files.WriteString(filename) onerr return error "{error}"

    print("Saved {len(tl.items)} todos to {filename}")
    return empty
```

**Pipe Operator:** Notice the clean pipeline: join the lines, then write to file, with each step on its own line. The `onerr return error "{error}"` propagates any write failure to the caller.

```kukicha
# Load reads todos from a file
func Load(filename string) (TodoList, error)
    tl := TodoList{items: empty list of Todo, nextId: 1}

    # Read the file content as bytes, convert to string, and split into lines
    data, err := files.Read(filename)
    if err != empty
        return tl, err
    lines := string.Split(string(data), "\n")

    # Skip if file is empty
    if len(lines) equals 0 or (len(lines) equals 1 and lines[0] equals "")
        return tl, empty
    maxId := 0

    for line in lines
        if line equals ""
            continue

        parts := string.Split(line, "|")
        if len(parts) not equals 3
            continue

        # strconv.Atoi can fail on bad data — skip the line if so
        id, parseErr := strconv.Atoi(parts[0])
        if parseErr not equals empty
            continue
        title := parts[1]
        completed := parts[2] equals "true"

        tl.items = append(tl.items, Todo{id: id, title: title, completed: completed})

        if id > maxId
            maxId = id

    tl.nextId = maxId + 1
    print("Loaded {len(tl.items)} todos from {filename}")
    return tl, empty
```

Notice the error handling approaches:
- `onerr return error "{error}"` (in Save) — propagates the error to the caller in one line
- Manual `if err not equals empty` (in Load) — use this when you need to do something other than return before continuing

---

## Step 7: The Complete Program

Now let's put it all together into a working application!

Create a file called `main.kuki`:

```kukicha
import "fmt"
import "os"
import "bufio"
import "stdlib/files"
import "stdlib/string"
import "strconv"

# --- Type Definitions ---

type Todo
    id int
    title string
    completed bool

type TodoList
    items list of Todo
    nextId int

# --- Todo Methods ---

func Display on todo Todo string
    status := "[ ]"
    if todo.completed
        status = "[✓]"
    return "{todo.id}. {status} {todo.title}"

# --- TodoList Methods ---

func Add on tl reference TodoList(title string)
    todo := Todo{id: tl.nextId, title: title, completed: false}
    tl.items = append(tl.items, todo)
    tl.nextId = tl.nextId + 1
    print("Added: {title}")

func ShowAll on tl TodoList
    if len(tl.items) equals 0
        print("\nNo todos yet! Use 'add <task>' to create one.\n")
        return

    print("\n=== Your Todos ===")
    for todo in tl.items
        print(todo.Display())
    print("")

func Complete on tl reference TodoList(id int)
    for i, todo in tl.items
        if todo.id equals id
            tl.items[i].completed = true
            print("Completed: {todo.title}")
            return
    print("Todo #{id} not found")

func Save on tl TodoList(filename string)
    lines := empty list of string

    for todo in tl.items
        completed := "false"
        if todo.completed
            completed = "true"
        lines = append(lines, "{todo.id}|{todo.title}|{completed}")

    # Join lines and write to file — pipe the data left to right
    lines
        |> string.Join("\n")
        |> files.WriteString(filename) onerr return

    print("Saved {len(tl.items)} todos to {filename}")

# --- Helper Functions ---

func LoadTodos(filename string) TodoList
    tl := TodoList{items: empty list of Todo, nextId: 1}

    data, err := files.Read(filename)
    if err != empty
        return tl
    lines := string.Split(string(data), "\n")

    maxId := 0
    for line in lines
        if line equals ""
            continue

        parts := string.Split(line, "|")
        if len(parts) not equals 3
            continue

        # Atoi can fail on bad data — skip the line
        id, parseErr := strconv.Atoi(parts[0])
        if parseErr not equals empty
            continue
        title := parts[1]
        completed := parts[2] equals "true"

        tl.items = append(tl.items, Todo{id: id, title: title, completed: completed})

        if id > maxId
            maxId = id

    tl.nextId = maxId + 1
    return tl

func PrintHelp()
    print("Commands:")
    print("  add <task>  - Add a new todo")
    print("  done <id>   - Mark a todo as complete")
    print("  list        - Show all todos")
    print("  save        - Save todos to file")
    print("  help        - Show this help")
    print("  quit        - Exit the app")

# --- Main Program ---

func main()
    filename := "todos.txt"
    # Note: This file will be created in the current working directory

    # Try to load existing todos
    tl := LoadTodos(filename)

    print("=== Kukicha Todo App ===")
    print("Type 'help' for commands\n")

    # Show existing todos if any
    if len(tl.items) > 0
        tl.ShowAll()

    # Create a reader for user input
    reader := bufio.NewReader(os.Stdin)

    # Main loop
    for
        fmt.Print("> ")

        # Read user input — default to empty string on error
        input := reader.ReadString('\n') onerr ""
        input = input |> string.TrimSpace()

        if input equals ""
            continue

        # SplitN(" ", 2) splits into at most 2 parts, so "add Buy milk" becomes ["add", "Buy milk"]
        # This protects against titles containing spaces
        parts := input |> string.SplitN(" ", 2)
        command := parts[0] |> string.ToLower()

        if command equals "quit" or command equals "exit" or command equals "q"
            print("Goodbye!")
            break

        else if command equals "help" or command equals "?"
            PrintHelp()

        else if command equals "list" or command equals "ls"
            tl.ShowAll()

        else if command equals "add"
            if len(parts) < 2
                print("Usage: add <task description>")
                continue
            title := parts[1]
            tl.Add(title)

        else if command equals "done" or command equals "complete"
            if len(parts) < 2
                print("Usage: done <id>")
                continue
            # Parse the id — print a message and skip if it's not a number
            id, idErr := strconv.Atoi(parts[1])
            if idErr not equals empty
                print("Invalid id: {parts[1]}")
                continue
            tl.Complete(id)

        else if command equals "save"
            tl.Save(filename)

        else
            print("Unknown command: {command}")
            print("Type 'help' for available commands")
```

---

## Step 8: Running Your App

Build and run your todo app:

```bash
kukicha run main.kuki
```

**Try these commands:**

```
> add Buy groceries
Added: Buy groceries

> add Learn Kukicha
Added: Learn Kukicha

> add Call mom
Added: Call mom

> list

=== Your Todos ===
1. [ ] Buy groceries
2. [ ] Learn Kukicha
3. [ ] Call mom

> done 2
Completed: Learn Kukicha

> list

=== Your Todos ===
1. [ ] Buy groceries
2. [✓] Learn Kukicha
3. [ ] Call mom

> save
Saved 3 todos to todos.txt

> quit
Goodbye!
```

If you run the app again, your todos will be loaded from the file!

---

## What You've Learned

Congratulations! You've built a real, working application. Let's review what you learned:

| Concept | What It Does |
|---------|--------------|
| **Custom Types** | Define your own data structures with `type Name` |
| **Methods** | Attach functions to types with `func Name on receiver Type` |
| **Pipe Operator** | Cleanly chain functions together with `|>` |
| **Lists** | Store multiple items with `list of Type` |
| **Loops** | Process each item with `for item in list` |
| **Error Handling** | Handle failures gracefully with `onerr` |
| **File I/O** | Save and load data with `files.Read()` and `files.WriteString()` |

---

## Practice Exercises

Ready for a challenge? Try these enhancements:

1. **Delete Command** - Add a `delete <id>` command to remove todos
2. **Priority Levels** - Add a `priority` field (high, medium, low) to todos
3. **Due Dates** - Add a `due` field and show overdue items
4. **Categories** - Add tags or categories to organize todos
5. **Search** - Add a `find <text>` command to search todos

---

## What's Next?

You now have solid programming skills with Kukicha! Continue with the tutorial series:

### Tutorial Path

1. ✅ **Beginner Tutorial** - Variables, functions, strings *(completed)*
2. ✅ **Console Todo** - Types, methods, lists, file I/O *(you are here)*
3. **[Web Todo Tutorial](web-app-tutorial.md)** ← Next step!
   - Build an HTTP server with REST APIs
   - Learn about JSON and web requests
4. **[Production Patterns](production-patterns-tutorial.md)** (Advanced)
   - Add database storage
   - Learn Go conventions and best practices

### Reference Docs

- **[Kukicha Grammar](../kukicha-grammar.ebnf.md)** - Complete language grammar
- **[Standard Library](../kukicha-stdlib-reference.md)** - iterator, slice, and more

---

**Great job! You've built a complete application with Kukicha! 🎉**
