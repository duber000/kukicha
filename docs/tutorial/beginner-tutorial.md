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

Kukicha compiles to Go (another programming language), which means your Kukicha programs run fast and can use Go's huge ecosystem of tools.

**The Botanical Metaphor:**
Kukicha uses plant terms to organize code:
- **Stem** = Your whole project (like a "module")
- **Petiole** = A package (a collection of related code)

Don't worry if this seems confusing now - it'll make sense as we go!

---

## Your First Program

Let's start with the traditional "Hello, World!" program. This is usually the first program anyone writes in a new language.

Create a file called `hello.kuki` with this content:

```kukicha
func main()
    print("Hello, World!")
```

**What's happening here?**

1. `func main()` - This defines a function named "main". Every Kukicha program starts by running the `main` function
2. `print("Hello, World!")` - This built-in function prints the text "Hello, World!" to the screen (and automatically imports `fmt` for you)
3. Notice: **No semicolons!** Kukicha uses indentation (spaces) to understand where code blocks begin and end

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

In Kukicha, any line starting with `#` is a comment - the computer ignores it completely:

```kukicha
# This is a comment - the computer skips this line

# Comments help you remember what your code does
func main()
    # Print a greeting to the screen
    print("Hello!")
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

Use the **walrus operator** `:=` to create a new variable:

```kukicha
func main()
    # Create a variable named 'age' and store 25 in it
    age := 25

    # Create a variable named 'name' and store "Alice" in it
    name := "Alice"

    # Use the variables
    print(name)
    print(age)
```

**Output:**
```
Alice
25
```

### Updating Variables

Once a variable exists, use a single `=` to change its value:

```kukicha
func main()
    score := 0          # Create score, set to 0
    print(score)  # Prints: 0

    score = 10          # Update score to 10
    print(score)  # Prints: 10

    score = score + 5   # Add 5 to current score
    print(score)  # Prints: 15
```

**Key difference:**
- `:=` creates a **new** variable
- `=` updates an **existing** variable

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

Kukicha is smart - when you create a local variable, it figures out the type automatically:

```kukicha
func main()
    age := 25              # Kukicha knows this is int
    price := 19.99         # Kukicha knows this is float64
    name := "Bob"          # Kukicha knows this is string
    isStudent := true      # Kukicha knows this is bool
```

### Why Types Matter

Types prevent mistakes. If you try to do something that doesn't make sense (like divide text by a number), Kukicha will catch the error before your program runs!

---

## Functions - Reusable Recipes

A **function** is a named block of code that performs a specific task. Think of it like a recipe you can use over and over.

### Basic Function

```kukicha
# Define a function named Greet
func Greet()
    print("Hello!")

# The main function - where your program starts
func main()
    Greet()  # Call the Greet function
    Greet()  # Call it again!
```

**Output:**
```
Hello!
Hello!
```

### Functions with Parameters

Functions can accept **parameters** (inputs):

```kukicha
# This function takes one parameter: a string named 'name'
func Greet(name string)
    print("Hello, {name}!")

func main()
    Greet("Alice")  # Prints: Hello, Alice!
    Greet("Bob")    # Prints: Hello, Bob!
```

**Important:** For function parameters, you **must** specify the type. Here, `name string` means "name is a string".

### Functions that Return Values

Functions can give back (return) a value:

```kukicha
# This function takes two ints and returns their sum (also an int)
func Add(a int, b int) int
    return a + b

func main()
    result := Add(5, 3)
    print(result)  # Prints: 8
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

```kukicha
func main()
    greeting := "Hello"
    name := "World"
    sentence := "Programming is fun!"

    print(greeting)
    print(name)
    print(sentence)
```

### Combining Strings

Use the `+` operator to join (concatenate) strings:

```kukicha
func main()
    firstName := "Alice"
    lastName := "Johnson"

    # Combine strings
    fullName := firstName + " " + lastName

    print(fullName)  # Prints: Alice Johnson
```

### String Comparisons

Compare strings using English words:

```kukicha
func main()
    password := "secret123"

    if password equals "secret123"
        print("Access granted!")
    else
        print("Access denied!")
```

**String comparison operators:**
- `equals` - checks if two strings are the same
- `not equals` - checks if two strings are different

---

## String Interpolation - Combining Text and Data

**String interpolation** lets you insert variable values directly into strings using curly braces `{variable}`.

### Basic Interpolation

```kukicha
func main()
    name := "Alice"
    age := 25

    # Insert variables into the string using {variable}
    message := "My name is {name} and I am {age} years old"

    print(message)
    # Prints: My name is Alice and I am 25 years old
```

### Why Interpolation is Awesome

**Without interpolation (the old way):**
```kukicha
message := "My name is " + name + " and I am " + age + " years old"
# Messy! Hard to read!
```

**With interpolation (the Kukicha way):**
```kukicha
message := "My name is {name} and I am {age} years old"
# Clean! Easy to read!
```

### Interpolation in Functions

```kukicha
func Greet(name string, time string) string
    return "Good {time}, {name}!"

func main()
    morning := Greet("Alice", "morning")
    evening := Greet("Bob", "evening")

    print(morning)  # Prints: Good morning, Alice!
    print(evening)  # Prints: Good evening, Bob!
```

### Interpolation with Expressions

You can put more than just variables in `{}`!

```kukicha
func main()
    x := 5
    y := 3

    # You can do calculations inside {}
    result := "The sum of {x} and {y} is {x + y}"

    print(result)
    # Prints: The sum of 5 and 3 is 8
```

---

## The String Petiole - Text Superpowers

Now comes the exciting part! Kukicha includes a **string petiole** (package) with 28 powerful functions for working with text.

A **petiole** is just a collection of related functions. The string petiole contains functions specifically for manipulating text.

### Importing the String Petiole

To use the string package, you need to import it:

```kukicha
import "stdlib/string"
```

Now you have access to all 28 string functions!

### Converting Case

Change text to UPPERCASE, lowercase, or Title Case:

```kukicha
import "stdlib/string"

func main()
    text := "hello world"

    upper := string.ToUpper(text)
    lower := string.ToLower(text)
    title := string.Title(text)

    print(upper)  # Prints: HELLO WORLD
    print(lower)  # Prints: hello world
    print(title)  # Prints: Hello World
```

**Real-world use case:** Converting user input to a consistent format before comparing it.

### The Pipe Operator (`|>`) - Cleaning Up Data

Sometimes you want to perform multiple operations on the same piece of text. Kukicha has a special tool called the **pipe operator** (`|>`) that lets you pass the result of one function directly into the next.

Instead of this:
```kukicha
cleaned := string.TrimSpace(text)
upper := string.ToUpper(cleaned)
```

You can do this:
```kukicha
upper := text |> string.TrimSpace() |> string.ToUpper()
```

It's called a "pipe" because it acts like a pipe at a construction site - data goes in one end and comes out the other end, transformed!

**Advanced tip:** By default, the piped value becomes the first argument. If you need it elsewhere, use `_` as a placeholder:
```kukicha
# Put "text" as the second argument instead of first
text |> process(config, _)  # Becomes: process(config, text)
```

### Trimming Whitespace

Remove extra spaces from the beginning and end of strings:

```kukicha
import "stdlib/string"

func main()
    messy := "   hello   "
    clean := string.TrimSpace(messy)

    print("Messy: [{messy}]")   # Prints: Messy: [   hello   ]
    print("Clean: [{clean}]")   # Prints: Clean: [hello]
```

**Real-world use case:** Cleaning up user input from forms.

### Removing Prefixes and Suffixes

```kukicha
import "stdlib/string"

func main()
    url := "https://example.com"
    filename := "document.pdf"

    # Remove the "https://" prefix
    domain := string.TrimPrefix(url, "https://")
    print(domain)  # Prints: example.com

    # Remove the ".pdf" suffix
    name := string.TrimSuffix(filename, ".pdf")
    print(name)  # Prints: document
```

### Splitting Strings

Break a string into pieces:

```kukicha
import "stdlib/string"

func main()
    # Split a comma-separated list
    colors := "red,green,blue"
    parts := string.Split(colors, ",")

    # parts is now a list: ["red", "green", "blue"]
    print(parts[0])  # Prints: red
    print(parts[1])  # Prints: green
    print(parts[2])  # Prints: blue
```

**Real-world use case:** Parsing CSV data or command-line arguments.

### Joining Strings

Combine a list of strings into one string:

```kukicha
import "stdlib/string"

func main()
    words := ["Hello", "World", "from", "Kukicha"]

    # Join with spaces
    sentence := string.Join(words, " ")
    print(sentence)  # Prints: Hello World from Kukicha

    # Join with dashes
    dashed := string.Join(words, "-")
    print(dashed)  # Prints: Hello-World-from-Kukicha
```

### Searching Within Strings

Check if a string contains another string:

```kukicha
import "stdlib/string"

func main()
    message := "Error: File not found"

    # Check if the message contains "Error"
    if string.Contains(message, "Error")
        print("This is an error message!")

    # Check if the message starts with "Error:"
    if string.HasPrefix(message, "Error:")
        print("Error detected!")

    # Check if a filename ends with .txt
    filename := "data.txt"
    if string.HasSuffix(filename, ".txt")
        print("This is a text file!")
```

### The 'in' Operator - Membership Testing

Kukicha has a super convenient shortcut for checking if text contains something:

```kukicha
func main()
    message := "Error: File not found"

    # Check if "Error" is in the message
    if "Error" in message
        print("This is an error message!")

    # Check if "Success" is NOT in the message
    if "Success" not in message
        print("Operation did not succeed")
```

This is easier than using `string.Contains`!

### Finding Positions

Find where a substring appears:

```kukicha
import "stdlib/string"

func main()
    text := "Hello, World!"

    # Find the position of "World"
    position := string.Index(text, "World")
    print(position)  # Prints: 7

    # If not found, returns -1
    notFound := string.Index(text, "Kukicha")
    print(notFound)  # Prints: -1
```

### Counting Occurrences

Count how many times a substring appears:

```kukicha
import "stdlib/string"

func main()
    text := "apple banana apple cherry apple"

    count := string.Count(text, "apple")
    print("The word 'apple' appears {count} times")
    # Prints: The word 'apple' appears 3 times
```

### Replacing Text

Replace parts of a string:

```kukicha
import "stdlib/string"

func main()
    text := "I love cats. Cats are great!"

    # Replace all occurrences of "cats" with "dogs"
    newText := string.ReplaceAll(text, "cats", "dogs")

    print(newText)
    # Prints: I love dogs. dogs are great!
```

---

## Building Your First Real Program

Let's combine everything we've learned to build a practical program: a **name formatter** that takes messy user input and formats it nicely.

```kukicha
import "stdlib/string"

# Clean and format a person's name using pipes
func FormatName(rawName string) string
    # We take the rawName, trim the space, then convert to Title Case
    return rawName 
        |> string.TrimSpace() 
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

    if "Alice" in name1
        print("Found Alice!")

    if string.Contains(name2, "Bob")
        print("Found Bob!")
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
6. ✅ The `in` operator for string searching
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
| 1 | ✅ *You are here* | Variables, functions, strings, pipes |
| 2 | **[Console Todo](console-todo-tutorial.md)** ← Next! | Types, methods, lists, file I/O, more pipes |
| 3 | **[Web Todo](web-app-tutorial.md)** | HTTP servers, JSON, REST APIs, expert piping |
| 4 | **[Production Patterns](production-patterns-tutorial.md)** | Databases, Go conventions |

### Additional Resources

- **[Kukicha Syntax Reference](../kukicha-syntax-v1.0.md)** - Complete language reference
- **[Quick Reference](../kukicha-quick-reference.md)** - Cheat sheet for quick lookups
- **examples/** directory - More example programs

### Practice Exercises

Try building these programs to practice your skills:

1. **Email Validator** - Check if an email contains "@" and ends with a domain
2. **Word Counter** - Count how many words are in a sentence (hint: use `string.Fields`)
3. **URL Parser** - Extract the domain from a URL (hint: use `string.TrimPrefix` and `string.Split`)
4. **Password Checker** - Verify a password is at least 8 characters and contains both letters and numbers

---

**Welcome to the world of programming with Kukicha! Happy coding!**
