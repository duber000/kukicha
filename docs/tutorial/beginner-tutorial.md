# Kukicha Programming Tutorial for Complete Beginners

Welcome! This tutorial will teach you programming from scratch using **Kukicha** (茎), a beginner-friendly language. By the end, you'll understand the basics and be able to work with text (strings) in your programs.

## Table of Contents

1. [What is Programming?](#what-is-programming)
2. [What is Kukicha?](#what-is-kukicha)
3. [Your First Program](#your-first-program)
4. [Comments - Leaving Notes for Yourself](#comments)
5. [Variables - Storing Information](#variables)
6. [Types - What Kind of Data?](#types)
7. [Functions - Reusable Recipes](#functions)
8. [Strings - Working with Text](#strings)
9. [String Interpolation - Combining Text and Data](#string-interpolation)
10. [The String Petiole - Text Superpowers](#the-string-petiole)
11. [Building Your First Real Program](#building-your-first-real-program)
12. [What's Next?](#whats-next)

---

## What is Programming?

**Programming** is giving instructions to a computer. Just like you might follow a recipe to bake a cake, computers follow programs (sets of instructions) to perform tasks.

When you write a program, you're teaching the computer:
- What information to remember
- What calculations to perform
- What decisions to make
- What actions to take

Computers are very literal - they do exactly what you tell them, nothing more, nothing less!

---

## What is Kukicha?

**Kukicha** is a programming language designed specifically for beginners. Unlike many languages that use lots of symbols (`&&`, `||`, `!=`, etc.), Kukicha uses plain English words:

- Instead of `==`, we write `equals`
- Instead of `&&`, we write `and`
- Instead of `||`, we write `or`
- Instead of `!`, we write `not`

We also prefer full English words for definitions:
- `function` (instead of `func`)
- `variable` (instead of `var`)
- `reference` (instead of pointers)

Kukicha compiles to Go (another programming language), which means your Kukicha programs run fast and can use Go's huge ecosystem of tools.

**The Botanical Metaphor:**
Kukicha uses plant terms to organize code:
- **Stem** = Your whole project (like a "module")
- **Petiole** = A package (a collection of related code)

Don't worry if this seems confusing now - it'll make sense as we go!

---

## Your First Program

Let's start with the traditional "Hello, World!" program. This is usually the first program anyone writes in a new language.

### Setting Up Your Project

Before writing code, let's set up a project folder:

```bash
mkdir my-kukicha-project
cd my-kukicha-project
go mod init my-project    # Create a Go module
kukicha init              # Extract Kukicha stdlib and configure go.mod
```

The `kukicha init` command sets up the Kukicha standard library in your project. This is needed when using `import "stdlib/..."` packages. For simple programs that don't use stdlib, you can skip this step.

### Writing Your First Program

Create a file called `hello.kuki` with this content:

```kukicha
function main()
    print("Hello, World!")
```

**What's happening here?**

1. `function main()` - This defines a function named "main". Every Kukicha program starts by running the `main` function
2. `print("Hello, World!")` - This built-in function prints the text "Hello, World!" to the screen 
3. Kukicha uses indentation (spaces) to understand where code blocks begin and end

**Try it yourself:**

```bash
kukicha run hello.kuki
```

You should see:
```
Hello, World!
```

Congratulations! You're now a programmer! 🎉

---

## Comments - Leaving Notes for Yourself

As you write programs, you'll want to leave notes explaining what your code does. These notes are called **comments**.

In Kukicha, any line starting with `#` is a comment - the computer ignores it completely.

Let's update our `hello.kuki` file to include some comments:

```kukicha
# This is a comment - the computer skips this line

# Comments help you remember what your code does
function main()
    # Print a greeting to the screen
    print("Hello!")
```

**Try it yourself:**

```bash
kukicha run hello.kuki
```

**When to use comments:**
- Explain *why* you wrote code a certain way
- Leave reminders for yourself
- Help other people understand your code

**Pro tip:** Good code should be clear enough that it doesn't need too many comments. Comments should explain the "why", not the "what".

---

## Variables - Storing Information

A **variable** is like a labeled box where you store information. You give it a name and put data in it.

### Creating Variables

Create a file called `variables.kuki`:

```kukicha
function main()
    # Create a variable named 'age' and store 25 in it
    age := 25

    # Create a variable named 'name' and store "Alice" in it
    name := "Alice"

    # Use the variables
    print(name)
    print(age)
```

**Try it yourself:**

```bash
kukicha run variables.kuki
```

**Output:**
```
Alice
25
```

### Updating Variables

Once a variable exists, use a single `=` to change its value. Let's update `variables.kuki`:

```kukicha
function main()
    score := 0          # Create score, set to 0
    print(score)  # Prints: 0

    score = 10          # Update score to 10
    print(score)  # Prints: 10

    score = score + 5   # Add 5 to current score
    print(score)  # Prints: 15
```

**Try it yourself:**

```bash
kukicha run variables.kuki
```

**Key difference:**
- `:=` creates a **new** variable
- `=` updates an **existing** variable

### Top-level Variables
Sometimes you want a variable to be accessible throughout your whole file, like a configuration setting. For this, you use the `variable` keyword at the top level. Let's update `variables.kuki` again:

```kukicha
variable APP_NAME string = "My Awesome App"
variable MAX_STRENGTH int = 100

function main()
    print("Welcome to {APP_NAME}!")
    print("Max strength is {MAX_STRENGTH}")
```

**Try it yourself:**

```bash
kukicha run variables.kuki
```

> **💡 Note:** Kukicha is designed to read like English. While you might see `func` or `var` in some advanced code (shortcuts), we recommend using `function` and `variable` to keep your code readable and friendly.

---

## Types - What Kind of Data?

Every piece of data has a **type** - it tells the computer what kind of information it is.

### Common Types

| Type | What it stores | Examples |
|------|----------------|----------|
| `int` | Whole numbers | `42`, `-10`, `0` |
| `float64` | Decimal numbers | `3.14`, `-0.5`, `2.0` |
| `string` | Text | `"Hello"`, `"Kukicha"` |
| `bool` | True or false | `true`, `false` |

### Type Inference

Kukicha is smart - when you create a local variable, it figures out the type automatically. Let's create a new file `functions.kuki` to see this:

```kukicha
function main()
    age := 25              # Kukicha knows this is int
    price := 19.99         # Kukicha knows this is float64
    name := "Bob"          # Kukicha knows this is string
    isStudent := true      # Kukicha knows this is bool
```

**Try it yourself:**

```bash
kukicha run functions.kuki
```

### Why Types Matter

Types prevent mistakes. If you try to do something that doesn't make sense (like divide text by a number), Kukicha will catch the error before your program runs!

---

## Functions - Reusable Recipes

A **function** is a named block of code that performs a specific task. Think of it like a recipe you can use over and over.

### 1. The `function` (or `func`) keyword
This tells Kukicha we are starting a new function.

### Basic Function

Update `functions.kuki`:

```kukicha
# Define a function named Greet
function Greet()
    print("Hello!")

# The main function - where your program starts
function main()
    Greet()  # Call the Greet function
    Greet()  # Call it again!
```

**Try it yourself:**

```bash
kukicha run functions.kuki
```

**Output:**
```
Hello!
Hello!
```

### Functions with Parameters

Functions can accept **parameters** (inputs). Update `functions.kuki`:

```kukicha
# This function takes one parameter: a string named 'name'
function Greet(name string)
    print("Hello, {name}!")

function main()
    Greet("Alice")  # Prints: Hello, Alice!
    Greet("Bob")    # Prints: Hello, Bob!
```

**Try it yourself:**

```bash
kukicha run functions.kuki
```

**Important:** For function parameters, you **must** specify the type. Here, `name string` means "name is a string".

### Functions that Return Values

Functions can give back (return) a value. Update `functions.kuki`:

```kukicha
# This function takes two ints and returns their sum (also an int)
function Add(a int, b int) int
    return a + b

function main()
    result := Add(5, 3)
    print(result)  # Prints: 8
```

**Try it yourself:**

```bash
kukicha run functions.kuki
```

**Key points:**
- The type after the parentheses (`int`) is the **return type**
- `return` sends a value back to whoever called the function
- Parameters and return types must have **explicit types** (you write them out)
- Local variables inside functions use **type inference** (Kukicha figures it out)

---

## Strings - Working with Text

A **string** is text - any sequence of characters. Strings are surrounded by double quotes.

### Creating Strings

Create a file called `strings.kuki`:

```kukicha
function main()
    greeting := "Hello"
    name := "World"
    sentence := "Programming is fun!"

    print(greeting)
    print(name)
    print(sentence)
```

**Try it yourself:**

```bash
kukicha run strings.kuki
```

### Combining Strings

Use the `+` operator to join (concatenate) strings. Update `strings.kuki`:

```kukicha
function main()
    firstName := "Alice"
    lastName := "Johnson"

    # Combine strings
    fullName := firstName + " " + lastName

    print(fullName)  # Prints: Alice Johnson
```

**Try it yourself:**

```bash
kukicha run strings.kuki
```

### String Comparisons

Compare strings using English words. Update `strings.kuki`:

```kukicha
function main()
    password := "secret123"

    if password equals "secret123"
        print("Access granted!")
    else
        print("Access denied!")
```

**Try it yourself:**

```bash
kukicha run strings.kuki
```

**String comparison operators:**
- `equals` - checks if two strings are the same
- `not equals` - checks if two strings are different

---

## String Interpolation - Combining Text and Data

**String interpolation** lets you insert variable values directly into strings using curly braces `{variable}`.

### Basic Interpolation

Update `strings.kuki`:

```kukicha
function main()
    name := "Alice"
    age := 25

    # Insert variables into the string using {variable}
    message := "My name is {name} and I am {age} years old"

    print(message)
    # Prints: My name is Alice and I am 25 years old
```

**Try it yourself:**

```bash
kukicha run strings.kuki
```

### Why Interpolation is Awesome

**Without interpolation (the old way):**
```kukicha
function main()
    name := "Alice"
    age := 25
    message := "My name is " + name + " and I am " + age + " years old"
    print(message)
```

**With interpolation (the Kukicha way):**
```kukicha
function main()
    name := "Alice"
    age := 25
    message := "My name is {name} and I am {age} years old"
    print(message)
```

**Try it yourself:**

```bash
kukicha run strings.kuki
```

### Interpolation in Functions

```kukicha
function Greet(name string, time string) string
    return "Good {time}, {name}!"

function main()
    morning := Greet("Alice", "morning")
    evening := Greet("Bob", "evening")

    print(morning)  # Prints: Good morning, Alice!
    print(evening)  # Prints: Good evening, Bob!
```

### Interpolation with Expressions

You can put more than just variables in `{}`! Update `strings.kuki` one last time:

```kukicha
function main()
    x := 5
    y := 3

    # You can do calculations inside {}
    result := "The sum of {x} and {y} is {x + y}"

    print(result)
    # Prints: The sum of 5 and 3 is 8
```

**Try it yourself:**

```bash
kukicha run strings.kuki
```

---

## The String Petiole - Text Superpowers

Now comes the exciting part! Kukicha includes a **string petiole** (package) with powerful functions for working with text.

A **petiole** is just a collection of related functions. The string petiole contains functions specifically for manipulating text.

### Importing the String Petiole

To use the string package, you need to import it:

```kukicha
import "stdlib/string"
```

Now you have access to the essential string functions covered in this tutorial!

### Converting Case

Change text to lowercase or Title Case. Create a file called `string_petiole.kuki`:

```kukicha
import "stdlib/string"

function main()
    text := "hello world"

    # Pipe syntax: value |> function
    # It reads like: "take text, then convert to lower case"
    lower := text |> string.ToLower()
    title := text |> string.Title()

    print(lower)  # Prints: hello world
    print(title)  # Prints: Hello World
```

**Try it yourself:**

```bash
kukicha run string_petiole.kuki
```

**Real-world use case:** Converting user input to a consistent format before comparing it.

### The Pipe Operator (`|>`) - Cleaning Up Data

Sometimes you want to perform multiple operations on the same piece of text. Kukicha has a special tool called the **pipe operator** (`|>`) that lets you pass the result of one function directly into the next.

Instead of this:
```kukicha
cleaned := string.TrimSpace(text)
lower := string.ToLower(cleaned)
```

You can do this:
```kukicha
lower := text |> string.TrimSpace() |> string.ToLower()
```

It's called a "pipe" because it acts like a pipe at a construction site - data goes in one end and comes out the other end, transformed!

### Trimming Whitespace

Remove extra spaces from the beginning and end of strings. This is a perfect job for the pipe we just learned — trim the whitespace, then normalize the case, all in one line. Update `string_petiole.kuki`:

```kukicha
import "stdlib/string"

function main()
    messy := "  HELLO  "

    # Pipe: trim whitespace, then lowercase — no temp variable needed
    clean := messy |> string.TrimSpace() |> string.ToLower()

    print("Messy: [{messy}]")   # Prints: Messy: [  HELLO  ]
    print("Clean: [{clean}]")   # Prints: Clean: [hello]
```

**Try it yourself:**

```bash
kukicha run string_petiole.kuki
```

**Real-world use case:** Cleaning up user input from forms.

### Removing Prefixes and Suffixes

Update `string_petiole.kuki`:

```kukicha
import "stdlib/string"

function main()
    url := "https://example.com/"
    filename := "document.pdf"

    # URLs often have both a prefix and trailing slash — pipe strips both
    domain := url |> string.TrimPrefix("https://") |> string.TrimSuffix("/")
    print(domain)  # Prints: example.com

    # Single operation — Use pipe for consistency!
    name := filename |> string.TrimSuffix(".pdf")
    print(name)  # Prints: document
```

**Try it yourself:**

```bash
kukicha run string_petiole.kuki
```

### Splitting Strings

Break a string into pieces. Update `string_petiole.kuki`:

```kukicha
import "stdlib/string"

function main()
    # Split a comma-separated list
    colors := "red,green,blue"
    
    # Use pipe to split
    parts := colors |> string.Split(",")

    # parts is now a list: ["red", "green", "blue"]
    print(parts[0])  # Prints: red
    print(parts[1])  # Prints: green
    print(parts[2])  # Prints: blue

    # Often when splitting strings, you have extra spaces
    # Let's see a messier example
    servers := "api1.example.com,  api2.example.com  , api3.example.com "
    serverList := servers |> string.Split(",")

    print(serverList[0])  # Prints: api1.example.com
    print(serverList[1])  # Prints:   api2.example.com   (with spaces!)
```

**Try it yourself:**

```bash
kukicha run string_petiole.kuki
```

**Real-world use case:** Parsing CSV data or command-line arguments.

### Advanced: Splitting AND Trimming

When you split strings with messy spacing, you often need to trim each piece. Kukicha provides a utility function that does both in one step! Add this to `string_petiole.kuki`:

```kukicha
import "stdlib/string"
import "stdlib/env"

function main()
    # Messy comma-separated list with inconsistent spacing
    servers := "api1.example.com,  api2.example.com  , api3.example.com "

    # env.SplitAndTrim does split + trim in one operation!
    # Pipes make it readable: 
    clean := servers |> env.SplitAndTrim(",")

    print(clean[0])  # Prints: api1.example.com (no spaces!)
    print(clean[1])  # Prints: api2.example.com (no spaces!)
    print(clean[2])  # Prints: api3.example.com (no spaces!)

    # It also skips empty parts - useful for trailing commas
    messy := "one, two, , three,  "
    cleaned := messy |> env.SplitAndTrim(",")
    print(cleaned)
    # Result: ["one", "two", "three"] - empty part removed!
```

**Try it yourself:**

```bash
kukicha run string_petiole.kuki
```

**Why is this useful?**
- Handles messy user input gracefully
- Saves you from writing loops to trim each piece
- Automatically removes empty entries
- Though it's in the `env` package, it's a general-purpose utility you can use anywhere!

**Real-world use case:** Parsing comma-separated lists from config files, user input, or database fields.

### Joining Strings

Combine a list of strings into one string. Update `string_petiole.kuki`:

```kukicha
import "stdlib/string"

function main()
    words := list of string{"Hello", "World", "from", "Kukicha"}

    # Join with spaces
    sentence := words |> string.Join(" ")
    print(sentence)  # Prints: Hello World from Kukicha

    # Join with dashes
    dashed := words |> string.Join("-")
    print(dashed)  # Prints: Hello-World-from-Kukicha
```

**Try it yourself:**

```bash
kukicha run string_petiole.kuki
```

### Searching Within Strings

Check if a string contains another string. Update `string_petiole.kuki`:

```kukicha
import "stdlib/string"

function main()
    message := "Error: File not found"

    # Check if the message contains "Error"
    if message |> string.Contains("Error")
        print("This is an error message!")

    # Check if the message starts with "Error:"
    if message |> string.HasPrefix("Error:")
        print("Error detected!")

    # Check if a filename ends with .txt
    filename := "data.txt"
    if filename |> string.HasSuffix(".txt")
        print("This is a text file!")
```

**Try it yourself:**

```bash
kukicha run string_petiole.kuki
```

### Membership Testing with Contains

Check whether a string contains a substring using `string.Contains`:

```kukicha
import "stdlib/string"

function main()
    message := "Error: File not found"

    # Check if the message contains "Error"
    if message |> string.Contains("Error")
        print("This is an error message!")

    # Check if "Success" is NOT in the message
    # 'not' works great with pipes!
    if not (message |> string.Contains("Success"))
        print("Operation did not succeed")
```

You already saw `string.Contains` earlier — it's your go-to for this kind of check.



### Replacing Text

Replace parts of a string. Need to make multiple replacements? Pipe them, one step per line. Update `string_petiole.kuki` one last time:

```kukicha
import "stdlib/string"

function main()
    text := "I love cats and dogs"

    # Each replacement feeds into the next — line up the pipes to see the flow
    newText := text
        |> string.ReplaceAll("cats", "kittens")
        |> string.ReplaceAll("dogs", "puppies")

    print(newText)
    # Prints: I love kittens and puppies
```

**Try it yourself:**

```bash
kukicha run string_petiole.kuki
```

---

## Building Your First Real Program

Let's combine everything we've learned to build a practical program: a **name formatter** that takes messy user input and formats it nicely.

Create a file called `name_formatter.kuki`:

```kukicha
import "stdlib/string"

# Clean and format a person's name using pipes
func FormatName(rawName string) string
    # We take the rawName, trim the space, convert any uppercase to lowercase and then convert to Title Case
    return rawName 
        |> string.TrimSpace()
        |> string.ToLower()
        |> string.Title()

# Create a greeting message
func CreateGreeting(name string, age int) string
    return "Welcome, {name}! You are {age} years old."

# Main program
func main()
    print("=== Name Formatter ===")

    # Simulate messy user input
    messyName1 := "  alice johnson  "
    messyName2 := "BOB SMITH"
    messyName3 := "charlie brown   "

    # Format the names
    name1 := FormatName(messyName1)
    name2 := FormatName(messyName2)
    name3 := FormatName(messyName3)

    # Create greetings
    greeting1 := CreateGreeting(name1, 25)
    greeting2 := CreateGreeting(name2, 30)
    greeting3 := CreateGreeting(name3, 22)

    # Print results
    print(greeting1)
    print(greeting2)
    print(greeting3)

    # Demonstrate string searching
    print("\n=== Name Search ===")

    if string.Contains(name1, "Alice")
        print("Found Alice!")

    if string.Contains(name2, "Bob")
        print("Found Bob!")
```

**Try it yourself:**

```bash
kukicha run name_formatter.kuki
```

**Output:**
```
=== Name Formatter ===
Welcome, Alice Johnson! You are 25 years old.
Welcome, Bob Smith! You are 30 years old.
Welcome, Charlie Brown! You are 22 years old.

=== Name Search ===
Found Alice!
Found Bob!
```

**What this program demonstrates:**
1. ✅ Importing multiple packages
2. ✅ Creating reusable functions
3. ✅ Using parameters and return values
4. ✅ String interpolation
5. ✅ Using the string petiole (TrimSpace, Title, Contains)
6. ✅ `string.Contains` for string searching
7. ✅ Combining everything into a working program

---

## What's Next?

Congratulations! You now know:

- ✅ What programming is and why it matters
- ✅ How to write and run Kukicha programs
- ✅ How to use variables to store data
- ✅ How to create functions to organize code
- ✅ How to work with strings (text)
- ✅ How to use string interpolation
- ✅ How to use the Pipe Operator (`|>`) to chain functions
- ✅ How to use the string petiole for advanced text operations

### Continue Your Journey

Ready for the next step? Follow this learning path:

| # | Tutorial | What You'll Learn |
|---|----------|-------------------|
| 1 | ✅ *You are here* | Variables, functions, strings, pipes, shorthands |
| 2 | **[Console Todo](console-todo-tutorial.md)** ← Next! | Types, methods, default parameters, named arguments, lists, file I/O, more pipes, error handling |
| 3 | **[Web Todo](web-app-tutorial.md)** | HTTP servers, JSON, REST APIs, expert piping |
| 4 | **[Production Patterns](production-patterns-tutorial.md)** | Databases, Go conventions |

### Additional Resources

- **[Kukicha Grammar](../kukicha-grammar.ebnf.md)** - Complete language grammar reference
- **[Stdlib Reference](../kukicha-stdlib-reference.md)** - Standard library documentation - additional functions to make your life easier!
- **[Examples](../../examples/)** directory - More example programs

---

**Welcome to the world of programming with Kukicha! Happy coding!**
