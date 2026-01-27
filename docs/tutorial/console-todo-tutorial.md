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
# 'list' is the receiver (the TodoList instance we're working on)
func Add on list reference TodoList, title string
    todo := CreateTodo(list.nextId, title)
    list.items = append(list.items, todo)
    list.nextId = list.nextId + 1
    print("Added: {title}")
```

**Note on receiver naming:** We use `list` as the receiver variable name (referring to the `TodoList` instance), which is clearer than a generic name. Some codebases use `tl` or `self` - pick whatever makes your code most readable!

```kukicha
# ShowAll displays all todos in the list
func ShowAll on list TodoList
    if len(list.items) equals 0
        print("No todos yet! Use 'add' to create one.")
        return
    
    print("\n=== Your Todos ===")
    for todo in list.items
        print(todo.Display())
    print("")
```

```kukicha
# Complete marks a todo as done by its id
func Complete on list reference TodoList, id int
    for i, todo in list.items
        if todo.id equals id
            list.items[i].completed = true
            print("Completed: {todo.title}")
            return true
    print("Todo #{id} not found")
    return false
```

---

## Step 5: Error Handling with `onerr`

Real programs need to handle errors. What if the user types something that isn't a number? What if a file doesn't exist?

Kukicha makes error handling readable with the `onerr` keyword.

### The Old Way (tedious)

```kukicha
# Without onerr - lots of repetitive code
result, err := somethingThatMightFail()
if err not equals empty
    print("Something went wrong!")
    return
```

### The Kukicha Way (clean)

```kukicha
# With onerr - handle the error inline
result := somethingThatMightFail() onerr
    print("Something went wrong!")
    return
```

### Common `onerr` Patterns

```kukicha
# Pattern 1: Provide a default value
name := getUserInput() onerr "Anonymous"

# Pattern 2: Print an error and return
data := loadFile() onerr
    print("Could not load file")
    return

# Pattern 3: Panic (crash) with a message
config := loadConfig() onerr panic "Missing config file!"

# Pattern 4: Exit with an error (Advanced)
# If your function returns an error, you can pass it along:
func DoWork() error
    data := loadFile() onerr return error
    return empty
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
func Save on list TodoList, filename string error
    lines := empty list of string

    for todo in list.items
        # Format: id|title|completed
        # We use pipe (|) as a delimiter because titles can contain commas or spaces
        # The format is simple: id|title|completed
        completed := "false"
        if todo.completed
            completed = "true"
        line := "{todo.id}|{todo.title}|{completed}"
        lines = append(lines, line)

    # Join all lines and write to file using pipe operator
    content := lines
        |> string.Join("\n")
        |> files.WriteString(filename)
        onerr return error

    print("Saved {len(list.items)} todos to {filename}")
    return empty

**Pipe Operator Shorthand:** Notice we use `|> files.WriteString(filename)` instead of `|> .WriteString(filename)`. Both work, but the full form is clearer when calling functions from imported packages. Use the dot shorthand (`.Display()`) when calling methods directly on the value itself.

# Load reads todos from a file
func Load(filename string) (TodoList, error)
    list := TodoList
        items: empty list of Todo
        nextId: 1

    # Read the file and split into lines in one pipeline
    lines := filename
        |> files.Read()
        onerr return list, error
        |> string.Split("\n")
    
    # Skip if file is empty
    if len(lines) equals 0 or (len(lines) equals 1 and lines[0] equals "")
        return list, empty
    maxId := 0
    
    for line in lines
        if line equals ""
            continue
        
        parts := string.Split(line, "|")
        if len(parts) not equals 3
            continue
        
        id := parts[0] |> strconv.Atoi() onerr continue
        title := parts[1]
        completed := parts[2] equals "true"
        
        todo := Todo
            id: id
            title: title
            completed: completed
        
        list.items = append(list.items, todo)
        
        if id > maxId
            maxId = id
    
    list.nextId = maxId + 1
    print("Loaded {len(list.items)} todos from {filename}")
    return list, empty
```

Notice how `onerr` makes the file operations clean:
- `onerr return error` - Pass the error to the caller
- `onerr continue` - Skip this line and try the next one

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

func Add on list reference TodoList, title string
    todo := Todo
        id: list.nextId
        title: title
        completed: false
    list.items = append(list.items, todo)
    list.nextId = list.nextId + 1
    print("Added: {title}")

func ShowAll on list TodoList
    if len(list.items) equals 0
        print("\nNo todos yet! Use 'add <task>' to create one.\n")
        return
    
    print("\n=== Your Todos ===")
    for todo in list.items
        print(todo.Display())
    print("")

func Complete on list reference TodoList, id int
    for i, todo in list.items
        if todo.id equals id
            list.items[i].completed = true
            print("Completed: {todo.title}")
            return
    print("Todo #{id} not found")

func Save on list TodoList, filename string
    lines := empty list of string

    for todo in list.items
        completed := "false"
        if todo.completed
            completed = "true"
        lines = append(lines, "{todo.id}|{todo.title}|{completed}")

    content := lines |> string.Join("\n")
    content |> files.WriteString(filename) onerr
        print("Error saving: could not write file")
        return

    print("Saved {len(list.items)} todos to {filename}")

# --- Helper Functions ---

func LoadTodos(filename string) TodoList
    list := TodoList
        items: empty list of Todo
        nextId: 1

    lines := filename
        |> files.Read()
        onerr return list
        |> string.Split("\n")
    
    maxId := 0
    for line in lines
        if line equals ""
            continue
        
        parts := string.Split(line, "|")
        if len(parts) not equals 3
            continue
        
        id := parts[0] |> strconv.Atoi() onerr continue
        title := parts[1]
        completed := parts[2] equals "true"
        
        list.items = append(list.items, Todo
            id: id
            title: title
            completed: completed
        )
        
        if id > maxId
            maxId = id
    
    list.nextId = maxId + 1
    return list

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
    # This works the same on Windows, macOS, and Linux
    
    # Try to load existing todos
    list := LoadTodos(filename)
    
    print("=== Kukicha Todo App ===")
    print("Type 'help' for commands\n")
    
    # Show existing todos if any
    if len(list.items) > 0
        list.ShowAll()
    
    # Create a reader for user input
    reader := bufio.NewReader(os.Stdin)
    
    # Main loop
    for
        fmt.Print("> ")
        
        # Read user input
        input := reader.ReadString('\n') onerr
            print("Error reading input")
            continue
        
        # Clean up the input using pipe
        input = input 
            |> string.TrimSpace()
        
        if input equals ""
            continue
        
        # Parse the command using pipe
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
            list.ShowAll()
        
        else if command equals "add"
            if len(parts) < 2
                print("Usage: add <task description>")
                continue
            title := parts[1]
            list.Add(title)
        
        else if command equals "done" or command equals "complete"
            if len(parts) < 2
                print("Usage: done <id>")
                continue
            id := parts[1] |> strconv.Atoi() onerr
                print("Invalid id: {parts[1]}")
                continue
            list.Complete(id)
        
        else if command equals "save"
            list.Save(filename)
        
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

- **[Kukicha Syntax Reference](kukicha-syntax-v1.0.md)** - Complete language guide
- **[Quick Reference](kukicha-quick-reference.md)** - Cheat sheet
- **[Standard Library](kukicha-stdlib-reference.md)** - iter, slice, and more

---

**Great job! You've built a complete application with Kukicha! 🎉**
