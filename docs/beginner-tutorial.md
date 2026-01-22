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
import "fmt"

func main()
    fmt.Println("Hello, World!")
```

**What's happening here?**

1. `import "fmt"` - We're bringing in the "fmt" package, which lets us print text to the screen
2. `func main()` - This defines a function named "main". Every Kukicha program starts by running the `main` function
3. `fmt.Println("Hello, World!")` - This prints the text "Hello, World!" to the screen
4. Notice: **No semicolons!** Kukicha uses indentation (spaces) to understand where code blocks begin and end

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
import "fmt"

# Comments help you remember what your code does
func main()
    # Print a greeting to the screen
    fmt.Println("Hello!")
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
import "fmt"

func main()
    # Create a variable named 'age' and store 25 in it
    age := 25

    # Create a variable named 'name' and store "Alice" in it
    name := "Alice"

    # Use the variables
    fmt.Println(name)
    fmt.Println(age)
```

**Output:**
```
Alice
25
```

### Updating Variables

Once a variable exists, use a single `=` to change its value:

```kukicha
import "fmt"

func main()
    score := 0          # Create score, set to 0
    fmt.Println(score)  # Prints: 0

    score = 10          # Update score to 10
    fmt.Println(score)  # Prints: 10

    score = score + 5   # Add 5 to current score
    fmt.Println(score)  # Prints: 15
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
import "fmt"

# Define a function named Greet
func Greet()
    fmt.Println("Hello!")

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
import "fmt"

# This function takes one parameter: a string named 'name'
func Greet(name string)
    fmt.Println("Hello, {name}!")

func main()
    Greet("Alice")  # Prints: Hello, Alice!
    Greet("Bob")    # Prints: Hello, Bob!
```

**Important:** For function parameters, you **must** specify the type. Here, `name string` means "name is a string".

### Functions that Return Values

Functions can give back (return) a value:

```kukicha
import "fmt"

# This function takes two ints and returns their sum (also an int)
func Add(a int, b int) int
    return a + b

func main()
    result := Add(5, 3)
    fmt.Println(result)  # Prints: 8
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
import "fmt"

func main()
    greeting := "Hello"
    name := "World"
    sentence := "Programming is fun!"

    fmt.Println(greeting)
    fmt.Println(name)
    fmt.Println(sentence)
```

### Combining Strings

Use the `+` operator to join (concatenate) strings:

```kukicha
import "fmt"

func main()
    firstName := "Alice"
    lastName := "Johnson"

    # Combine strings
    fullName := firstName + " " + lastName

    fmt.Println(fullName)  # Prints: Alice Johnson
```

### String Comparisons

Compare strings using English words:

```kukicha
import "fmt"

func main()
    password := "secret123"

    if password equals "secret123"
        fmt.Println("Access granted!")
    else
        fmt.Println("Access denied!")
```

**String comparison operators:**
- `equals` - checks if two strings are the same
- `not equals` - checks if two strings are different

---

## String Interpolation - Combining Text and Data

**String interpolation** lets you insert variable values directly into strings using curly braces `{variable}`.

### Basic Interpolation

```kukicha
import "fmt"

func main()
    name := "Alice"
    age := 25

    # Insert variables into the string using {variable}
    message := "My name is {name} and I am {age} years old"

    fmt.Println(message)
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
import "fmt"

func Greet(name string, time string) string
    return "Good {time}, {name}!"

func main()
    morning := Greet("Alice", "morning")
    evening := Greet("Bob", "evening")

    fmt.Println(morning)  # Prints: Good morning, Alice!
    fmt.Println(evening)  # Prints: Good evening, Bob!
```

### Interpolation with Expressions

You can put more than just variables in `{}`!

```kukicha
import "fmt"

func main()
    x := 5
    y := 3

    # You can do calculations inside {}
    result := "The sum of {x} and {y} is {x + y}"

    fmt.Println(result)
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
import "fmt"
import "stdlib/string"

func main()
    text := "hello world"

    upper := string.ToUpper(text)
    lower := string.ToLower(text)
    title := string.Title(text)

    fmt.Println(upper)  # Prints: HELLO WORLD
    fmt.Println(lower)  # Prints: hello world
    fmt.Println(title)  # Prints: Hello World
```

**Real-world use case:** Converting user input to a consistent format before comparing it.

### Trimming Whitespace

Remove extra spaces from the beginning and end of strings:

```kukicha
import "fmt"
import "stdlib/string"

func main()
    messy := "   hello   "
    clean := string.TrimSpace(messy)

    fmt.Println("Messy: [{messy}]")   # Prints: Messy: [   hello   ]
    fmt.Println("Clean: [{clean}]")   # Prints: Clean: [hello]
```

**Real-world use case:** Cleaning up user input from forms.

### Removing Prefixes and Suffixes

```kukicha
import "fmt"
import "stdlib/string"

func main()
    url := "https://example.com"
    filename := "document.pdf"

    # Remove the "https://" prefix
    domain := string.TrimPrefix(url, "https://")
    fmt.Println(domain)  # Prints: example.com

    # Remove the ".pdf" suffix
    name := string.TrimSuffix(filename, ".pdf")
    fmt.Println(name)  # Prints: document
```

### Splitting Strings

Break a string into pieces:

```kukicha
import "fmt"
import "stdlib/string"

func main()
    # Split a comma-separated list
    colors := "red,green,blue"
    parts := string.Split(colors, ",")

    # parts is now a list: ["red", "green", "blue"]
    fmt.Println(parts[0])  # Prints: red
    fmt.Println(parts[1])  # Prints: green
    fmt.Println(parts[2])  # Prints: blue
```

**Real-world use case:** Parsing CSV data or command-line arguments.

### Joining Strings

Combine a list of strings into one string:

```kukicha
import "fmt"
import "stdlib/string"

func main()
    words := ["Hello", "World", "from", "Kukicha"]

    # Join with spaces
    sentence := string.Join(words, " ")
    fmt.Println(sentence)  # Prints: Hello World from Kukicha

    # Join with dashes
    dashed := string.Join(words, "-")
    fmt.Println(dashed)  # Prints: Hello-World-from-Kukicha
```

### Searching Within Strings

Check if a string contains another string:

```kukicha
import "fmt"
import "stdlib/string"

func main()
    message := "Error: File not found"

    # Check if the message contains "Error"
    if string.Contains(message, "Error")
        fmt.Println("This is an error message!")

    # Check if the message starts with "Error:"
    if string.HasPrefix(message, "Error:")
        fmt.Println("Error detected!")

    # Check if a filename ends with .txt
    filename := "data.txt"
    if string.HasSuffix(filename, ".txt")
        fmt.Println("This is a text file!")
```

### The 'in' Operator - Membership Testing

Kukicha has a super convenient shortcut for checking if text contains something:

```kukicha
import "fmt"

func main()
    message := "Error: File not found"

    # Check if "Error" is in the message
    if "Error" in message
        fmt.Println("This is an error message!")

    # Check if "Success" is NOT in the message
    if "Success" not in message
        fmt.Println("Operation did not succeed")
```

This is easier than using `string.Contains`!

### Finding Positions

Find where a substring appears:

```kukicha
import "fmt"
import "stdlib/string"

func main()
    text := "Hello, World!"

    # Find the position of "World"
    position := string.Index(text, "World")
    fmt.Println(position)  # Prints: 7

    # If not found, returns -1
    notFound := string.Index(text, "Kukicha")
    fmt.Println(notFound)  # Prints: -1
```

### Counting Occurrences

Count how many times a substring appears:

```kukicha
import "fmt"
import "stdlib/string"

func main()
    text := "apple banana apple cherry apple"

    count := string.Count(text, "apple")
    fmt.Println("The word 'apple' appears {count} times")
    # Prints: The word 'apple' appears 3 times
```

### Replacing Text

Replace parts of a string:

```kukicha
import "fmt"
import "stdlib/string"

func main()
    text := "I love cats. Cats are great!"

    # Replace all occurrences of "cats" with "dogs"
    newText := string.ReplaceAll(text, "cats", "dogs")

    fmt.Println(newText)
    # Prints: I love dogs. dogs are great!
```

---

## Building Your First Real Program

Let's combine everything we've learned to build a practical program: a **name formatter** that takes messy user input and formats it nicely.

```kukicha
import "fmt"
import "stdlib/string"

# Clean and format a person's name
func FormatName(rawName string) string
    # Remove extra whitespace
    cleaned := string.TrimSpace(rawName)

    # Convert to title case (First Letter Caps)
    formatted := string.Title(cleaned)

    return formatted

# Create a greeting message
func CreateGreeting(name string, age int) string
    return "Welcome, {name}! You are {age} years old."

# Main program
func main()
    fmt.Println("=== Name Formatter ===")

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
    fmt.Println(greeting1)
    fmt.Println(greeting2)
    fmt.Println(greeting3)

    # Demonstrate string searching
    fmt.Println("\n=== Name Search ===")

    if "Alice" in name1
        fmt.Println("Found Alice!")

    if string.Contains(name2, "Bob")
        fmt.Println("Found Bob!")
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
- ✅ How to use the string petiole for advanced text operations

### Continue Your Journey

Here's what to explore next:

1. **Control Flow** - Learn `if`, `else`, and loops (`for`)
2. **Collections** - Work with lists and maps (arrays and dictionaries)
3. **Error Handling** - Use the `onerr` operator to handle errors gracefully
4. **Other Petioles** - Explore `iter` (iterators), `slice` (list operations), and more
5. **Build Projects** - Make a todo app, a simple calculator, or a text-based game

### Additional Resources

- **kukicha-syntax-v1.0.md** - Complete language reference
- **kukicha-quick-reference.md** - Cheat sheet for quick lookups
- **examples/** directory - More example programs

### Practice Exercises

Try building these programs to practice your skills:

1. **Email Validator** - Check if an email contains "@" and ends with a domain
2. **Word Counter** - Count how many words are in a sentence (hint: use `string.Fields`)
3. **URL Parser** - Extract the domain from a URL (hint: use `string.TrimPrefix` and `string.Split`)
4. **Password Checker** - Verify a password is at least 8 characters and contains both letters and numbers

---

## Summary of String Petiole Functions

Here's a quick reference of all 28 functions in the string package:

### Case Conversion
- `string.ToUpper(s)` - Convert to UPPERCASE
- `string.ToLower(s)` - Convert to lowercase
- `string.Title(s)` - Convert To Title Case

### Trimming
- `string.TrimSpace(s)` - Remove leading/trailing spaces
- `string.TrimPrefix(s, prefix)` - Remove prefix if present
- `string.TrimSuffix(s, suffix)` - Remove suffix if present
- `string.Trim(s, cutset)` - Remove leading/trailing characters from cutset
- `string.TrimLeft(s, cutset)` - Remove leading characters from cutset
- `string.TrimRight(s, cutset)` - Remove trailing characters from cutset

### Splitting & Joining
- `string.Split(s, sep)` - Split string by separator
- `string.SplitN(s, sep, n)` - Split string by separator, max n parts
- `string.Fields(s)` - Split by whitespace
- `string.Lines(s)` - Split by newlines
- `string.Join(list, sep)` - Join list of strings with separator

### Searching
- `string.Contains(s, substr)` - Check if s contains substr
- `string.HasPrefix(s, prefix)` - Check if s starts with prefix
- `string.HasSuffix(s, suffix)` - Check if s ends with suffix
- `string.Index(s, substr)` - Find first position of substr (-1 if not found)
- `string.LastIndex(s, substr)` - Find last position of substr
- `string.Count(s, substr)` - Count occurrences of substr

### Replacement
- `string.ReplaceAll(s, old, new)` - Replace all occurrences
- `string.Replace(s, old, new, n)` - Replace first n occurrences (-1 for all)

### Other
- `string.Repeat(s, count)` - Repeat string count times
- `string.Compare(a, b)` - Compare two strings (-1, 0, or 1)
- `string.EqualFold(a, b)` - Case-insensitive equality check
- `string.Clone(s)` - Create a copy of string
- `string.Cut(s, sep)` - Split at first separator, return before/after/found
- `string.CutPrefix(s, prefix)` - Remove prefix, return result and whether it was present
- `string.CutSuffix(s, suffix)` - Remove suffix, return result and whether it was present

---

**Welcome to the world of programming with Kukicha! Happy coding! 🌱**
