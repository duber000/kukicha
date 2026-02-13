# Building a Console Todo App with Kukicha

**Level:** Intermediate  
**Time:** 15-18 minutes  
**Prerequisite:** [Beginner Tutorial](beginner-tutorial.md)

Welcome back! In the beginner tutorial, you learned about variables, functions, strings, decisions, lists, and loops. Now we're going to build something real: a **todo list application** that runs in your terminal.

## What You'll Learn

In this tutorial, you'll discover how to:
- Create **custom types** to organize related data
- Write **methods** that belong to types
- Use `reference` to modify data in place
- Handle errors gracefully with **`onerr`**
- Use the **Pipe Operator (`|>`)** for clean data flow
- Build a simple **command loop** for a console app

Let's build something useful!

---

## What We're Building

Our todo app will let you:
- **Add** new tasks
- **View** all your tasks
- **Mark** tasks as done

Here's what it will look like when running:

```
=== Todo App ===
1. [ ] Buy groceries
2. [✓] Learn Kukicha
3. [ ] Call mom

Commands: add, done, list, help, quit
> 
```

---

## Step 0: Project Setup

If you haven't already, set up your project:

```bash
mkdir todo-app && cd todo-app
go mod init todo-app
kukicha init    # Extracts stdlib for imports like "stdlib/string"
```

---

## Step 1: Creating a Todo Type

> **📝 Reminder:** This tutorial builds on the beginner tutorial. Here are the key concepts you'll need:
> - **`:=`** creates a new variable, **`=`** updates an existing one
> - **String interpolation:** Use `{variable}` inside strings to insert values
> - **`print()`** outputs to the console
> - **Functions** (starting with `function`) take parameters and can return values
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
function CreateTodo(id int, title string) Todo
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
function Display on todo Todo string
    status := "[ ]"
    if todo.completed
        status = "[✓]"
    return "{todo.id}. {status} {todo.title}"
```

**Reading this method:**
- `function Display` - We're creating a method called "Display"
- `on todo Todo` - This method works on a `Todo` (syntax: receiver name first, then the type). Inside the method, we call it `todo`
- `string` - The method returns a string

The method checks if the todo is completed. If so, it shows a checkmark. Otherwise, it shows empty brackets.

### A Method That Changes Things

What if we want to mark a todo as done? We need a method that can **modify** the todo. For that, we use `reference`:

```kukicha
# MarkDone sets the todo as completed
# We use "reference" because we're changing the todo
function MarkDone on todo reference Todo
    todo.completed = true
```

**Why `reference`?**

Without `reference`, the method would get a **copy** of the todo. Any changes would only affect the copy, not the original. Using `reference` means we're working with the **actual** todo, so our changes stick.

Think of it like a shared document: without `reference`, you'd get a photocopy - you can scribble on it all day, but the original won't change. With `reference`, you're editing the original document itself.

### Let's Try It

Add a `main` function to `todo.kuki` so we can run what we've built so far:

```kukicha
function main()
    todo := CreateTodo(1, "Learn Kukicha")

    # Call Display directly
    print(todo.Display())

    # Or use the pipe operator — data flows left to right
    todo |> .Display() |> print()

    # Mark it done and display again
    todo.MarkDone()
    todo |> .Display() |> print()
```

Run it with `kukicha run todo.kuki`:

```
1. [ ] Learn Kukicha
1. [ ] Learn Kukicha
1. [✓] Learn Kukicha
```

**💡 Pipe Dot Shorthand:** When piping into a method that belongs to the value itself, use `.Method()`. Like we did in .Display() above. This keeps the left-to-right data flow and makes it clear the method belongs to the piped value, not an imported package.

---

## Step 3: Building the Todo List Type

Next, we'll create a type that manages a list of todos and provides methods to add, complete, and display items. Add a `TodoList` type to `todo.kuki`, above `main`:

```kukicha
# TodoList manages a collection of todos
type TodoList
    items list of Todo
    nextId int
```

Now add methods for it:

```kukicha
# Add creates a new todo and adds it to the list
# 'tl' is the receiver (the TodoList instance we're working on)
# Extra parameters go in parentheses after the receiver type
function Add on tl reference TodoList(title string)
    todo := CreateTodo(tl.nextId, title)
    tl.items = append(tl.items, todo)
    tl.nextId = tl.nextId + 1
    print("Added: {title}")
```

**Note on receiver naming:** We use `tl` as the receiver variable name (referring to the `TodoList` instance). Avoid using `list` — it's a keyword in Kukicha (as in `list of int`).

```kukicha
# ShowAll displays all todos in the list
function ShowAll on tl TodoList
    if len(tl.items) equals 0
        print("No todos yet! Use 'add' to create one.")
        return

    print("\n=== Your Todos ===")
    for todo in tl.items
        todo |> .Display() |> print()
    print("")
```

### Finding and Modifying Items

To complete a todo, we need to find it in our list. List items are accessed by their **index** (their position in the list, starting at 0).

> **💡 Recall from the beginner tutorial:** `for i, todo in tl.items` loops through the list with both the index and the item. The names `i` and `todo` are your choice - `i` gets the position number (starting at 0), and `todo` gets the item at that position. See [Loops - Repeating Actions](beginner-tutorial.md#loops---repeating-actions) for a refresher.

**Why return -1?** Valid list indices start at 0, so -1 is an impossible index. This is a common programming convention called a **sentinel value** - a special value that signals "not found." When the caller sees -1, they know the search failed.

```kukicha
# FindIndex returns the invalid index -1 if not found
function FindIndex on tl TodoList(id int) int
    for i, todo in tl.items
        if todo.id equals id
            return i
    return -1

# Complete marks a todo as done by its id
function Complete on tl reference TodoList(id int)
    index := tl.FindIndex(id)
    if index equals -1
        print("Todo #{id} not found")
        return

    tl.items[index].completed = true
    print("Completed: {tl.items[index].title}")
```

Notice `Complete` modifies items by index (`tl.items[i].completed = true`) — this changes the actual item in the list, solving the copy problem from the previous step.

Update `main` to use the new `TodoList`:

```kukicha
function main()
    tl := TodoList{items: empty list of Todo, nextId: 1}

    tl.Add("Buy groceries")
    tl.Add("Learn Kukicha")
    tl.Add("Call mom")

    tl.ShowAll()

    tl.Complete(2)
    tl.ShowAll()
```

**Note:** `empty list of Todo` creates an empty list that's typed to hold `Todo` items. Since the list starts empty, Kukicha can't infer the type from the contents, so we specify it explicitly.

Run it:

```
Added: Buy groceries
Added: Learn Kukicha
Added: Call mom

=== Your Todos ===
1. [ ] Buy groceries
2. [ ] Learn Kukicha
3. [ ] Call mom

Completed: Learn Kukicha

=== Your Todos ===
1. [ ] Buy groceries
2. [✓] Learn Kukicha
3. [ ] Call mom
```

---

## Step 4: The Complete Program

Now let's put it all together into a working application!

> **Note:** The final program imports `bufio` and `os` only for reading console input. Everything else is pure Kukicha and its standard library.

Create a file called `main.kuki`:

```kukicha
import "bufio"
import "os"
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

function Display on todo Todo string
    status := "[ ]"
    if todo.completed
        status = "[✓]"
    return "{todo.id}. {status} {todo.title}"

# --- TodoList Methods ---

function Add on tl reference TodoList(title string)
    todo := Todo{id: tl.nextId, title: title, completed: false}
    tl.items = append(tl.items, todo)
    tl.nextId = tl.nextId + 1
    print("Added: {title}")

function ShowAll on tl TodoList
    if len(tl.items) equals 0
        print("\nNo todos yet! Use 'add <task>' to create one.\n")
        return

    print("\n=== Your Todos ===")
    for todo in tl.items
        print(todo.Display())
    print("")

function FindIndex on tl TodoList(id int) int
    for i, todo in tl.items
        if todo.id equals id
            return i
    return -1

function Complete on tl reference TodoList(id int)
    index := tl.FindIndex(id)
    if index equals -1
        print("Todo #{id} not found")
        return

    tl.items[index].completed = true
    print("Completed: {tl.items[index].title}")

function PrintHelp()
    print("Commands:")
    print("  add <task>  - Add a new todo")
    print("  done <id>   - Mark a todo as complete")
    print("  list        - Show all todos")
    print("  help        - Show this help")
    print("  quit        - Exit the app")

# --- Main Program ---

function main()
    tl := TodoList{items: empty list of Todo, nextId: 1}

    print("=== Kukicha Todo App ===")
    print("Type 'help' for commands\n")

    # Create a reader for user input
    reader := bufio.NewReader(os.Stdin)

    # Main loop
    for
        print("> ")

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

        else
            print("Unknown command: {command}")
            print("Type 'help' for available commands")
```

---

## Step 5: Running Your App

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

> quit
Goodbye!
```

---

## Understanding the New Concepts

The final program introduced several concepts that deserve a closer look. Let's walk through them.

### Bare `for` - The Infinite Loop

```kukicha
for
    # ... read input and process commands ...
```

A `for` with no condition runs forever. This is the standard pattern for programs that wait for user input - the loop keeps running until something inside calls `break`. You saw in the beginner tutorial that `for condition` runs while the condition is true; a bare `for` is just the extreme case where the condition is always true.

### `onerr` - Graceful Error Handling

```kukicha
input := reader.ReadString('\n') onerr ""
```

Some operations can fail - reading input might hit an error if the terminal closes unexpectedly. The **`onerr`** clause says "if this fails, use this value instead." Here, if `ReadString` fails, `input` gets set to an empty string `""` and the program continues normally instead of crashing.

You can use `onerr` with different fallback strategies:
- `onerr ""` or `onerr 0` - use a default value
- `onerr return` - exit the current function
- `onerr panic "message"` - crash with an error message (for truly unexpected failures)

### `continue` in Context

```kukicha
if input equals ""
    continue
```

When the user presses Enter without typing anything, `continue` skips the rest of the loop body and goes straight back to the `>` prompt. Without `continue`, the empty input would fall through to the command parsing logic and print "Unknown command."

### `empty` for Null Checking

```kukicha
if idErr not equals empty
    print("Invalid id: {parts[1]}")
```

In Kukicha, **`empty`** represents "no value" (called `nil` in many other languages). When a function can fail, it returns an error value alongside the result. If the error `not equals empty`, something went wrong. Here we check whether `strconv.Atoi` (which converts text to a number) failed - if so, the user typed something that isn't a valid number.

---

## What You've Learned

Congratulations! You've built a real, working application. Let's review what you learned:

| Concept | What It Does |
|---------|--------------|
| **Custom Types** | Define your own data structures with `type Name` |
| **Methods** | Attach functions to types with `function Name on receiver Type` |
| **`reference`** | Modify the original value, not a copy |
| **`onerr`** | Handle errors gracefully with fallback values |
| **Pipe Operator** | Cleanly chain functions together with `|>` |
| **`empty`** | Check for null/missing values |
| **Command Loop** | Read input, bare `for`, `break`, and `continue` |

---

## Practice Exercises

Ready for a challenge? Try these enhancements:

1. **Delete Command** - Add a `delete <id>` command to remove todos
2. **Priority Levels** - Add a `priority` field (high, medium, low) to todos
3. **Categories** - Add tags or categories to organize todos
4. **Search** - Add a `find <text>` command to search todos

---

## What's Next?

You now have solid programming skills with Kukicha! Continue with the tutorial series:

### Tutorial Path

1. ✅ **Beginner Tutorial** - Variables, functions, strings, decisions, lists, loops *(completed)*
2. ✅ **Console Todo** - Custom types, methods, error handling, command loops *(you are here)*
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
