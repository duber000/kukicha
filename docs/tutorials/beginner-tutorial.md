# Kukicha for Shell Scripters

You already know how to get things done in bash. You can pipe commands together, write functions, loop over files, and set variables. But your scripts are getting longer, harder to debug, and full of quoting nightmares. You want something more structured without losing the directness you're used to.

Kukicha is a programming language that compiles to Go. It keeps the things you like about shell scripting - pipes, readable flow, running commands - and gives you types, real error handling, and compiled binaries you can deploy anywhere.

This tutorial maps what you already know in bash to how it works in Kukicha.

## Table of Contents

1. [Setup](#setup)
2. [Hello World - Your First Script](#hello-world---your-first-script)
3. [Variables - No More Dollar Signs](#variables---no-more-dollar-signs)
4. [Comments](#comments)
5. [Types - What Kind of Data?](#types---what-kind-of-data)
6. [Strings - No More Quoting Hell](#strings---no-more-quoting-hell)
7. [Conditionals - If Without Then/Fi](#conditionals---if-without-thenfi)
8. [Functions - Like Shell Functions, But Better](#functions---like-shell-functions-but-better)
9. [Lists - Like Arrays, But They Actually Work](#lists---like-arrays-but-they-actually-work)
10. [Loops - For Without Do/Done](#loops---for-without-dodone)
11. [Pipes - You Already Know This One](#pipes---you-already-know-this-one)
12. [Running Commands](#running-commands)
13. [Error Handling - Better Than set -e](#error-handling---better-than-set--e)
14. [Putting It Together - A Log Analyzer](#putting-it-together---a-log-analyzer)
15. [What's Next?](#whats-next)

---

## Setup

Before writing code, set up a project folder:

```bash
mkdir my-kukicha-project
cd my-kukicha-project
go mod init my-project    # Create a Go module
kukicha init              # Extract Kukicha stdlib and configure go.mod
```

The `kukicha init` command sets up the Kukicha standard library in your project. This is needed when using `import "stdlib/..."` packages. For simple programs that don't use stdlib, you can skip this step.

---

## Hello World - Your First Script

**Bash:**
```bash
#!/bin/bash
echo "Hello, World!"
```

**Kukicha:**
```kukicha
function main()
    print("Hello, World!")
```

Save this as `hello.kuki` and run it:

```bash
kukicha run hello.kuki
```

Two differences to notice:
1. Every Kukicha program starts from a `main()` function, like a script's entry point
2. Indentation (4 spaces) defines blocks - no `fi`, `done`, `esac`, or curly braces

---

## Variables - No More Dollar Signs

In bash, variables are untyped strings with inconsistent syntax (`$var`, `${var}`, `"$var"`). In Kukicha, variables are straightforward.

**Bash:**
```bash
name="Alice"
age=25
echo "$name is $age"

# Reassign
age=26
```

**Kukicha:**
```kukicha
function main()
    name := "Alice"
    age := 25
    print("{name} is {age}")

    # Reassign
    age = 26
```

Save this as `variables.kuki` and run it:

```bash
kukicha run variables.kuki
```

Key differences:
- `:=` creates a new variable (like first assignment in bash)
- `=` updates an existing variable
- No `$` needed to read a variable - just use the name
- No quoting issues - spaces in strings just work

### Top-Level Variables

Sometimes you want a variable accessible across your whole file, like a config value at the top of a bash script. Use the `variable` keyword:

**Bash:**
```bash
APP_NAME="My Tool"
MAX_RETRIES=3
```

**Kukicha:**
```kukicha
variable APP_NAME string = "My Tool"
variable MAX_RETRIES int = 3

function main()
    print("Starting {APP_NAME}, max retries: {MAX_RETRIES}")
```

---

## Comments

Same as bash - lines starting with `#` are comments.

```kukicha
# This is a comment
function main()
    # Print a greeting
    print("Hello!")
```

---

## Types - What Kind of Data?

Bash treats everything as a string. Kukicha knows what kind of data you're working with, which catches mistakes before your script runs.

| Type | What it stores | Examples |
|------|----------------|----------|
| `int` | Whole numbers | `42`, `-10`, `0` |
| `float64` | Decimal numbers | `3.14`, `-0.5` |
| `string` | Text | `"Hello"`, `"/tmp/file"` |
| `bool` | True or false | `true`, `false` |

You usually don't need to write types explicitly - Kukicha figures them out:

```kukicha
function main()
    count := 25           # Kukicha knows this is int
    path := "/tmp/output" # Kukicha knows this is string
    verbose := true       # Kukicha knows this is bool
```

This means you can't accidentally do math on a filename or compare a number to a string. Kukicha catches those errors at compile time, not at 3am in production.

---

## Strings - No More Quoting Hell

If you've ever debugged a bash script where spaces in filenames broke everything, or wrestled with nested quotes, this section is for you.

### String Interpolation

**Bash:**
```bash
name="Alice"
count=5
echo "Hello $name, you have ${count} items"
echo "Path is: ${BASE_DIR}/${name}/config"
```

**Kukicha:**
```kukicha
function main()
    name := "Alice"
    count := 5
    print("Hello {name}, you have {count} items")

    baseDir := "/opt"
    print("Path is: {baseDir}/{name}/config")
```

Curly braces `{variable}` insert values into strings. No `$`, no worrying about when to use `${...}` vs `$...`.

### Expressions in Strings

You can put calculations inside `{}` too:

```kukicha
function main()
    x := 5
    y := 3
    print("The sum of {x} and {y} is {x + y}")
```

### String Comparison

**Bash:**
```bash
if [ "$status" = "ready" ]; then
    echo "Good to go"
fi
```

**Kukicha:**
```kukicha
function main()
    status := "ready"

    if status equals "ready"
        print("Good to go")
```

No brackets, no quoting the variable, no semicolons. `equals` and `not equals` do what you'd expect.

---

## Conditionals - If Without Then/Fi

**Bash:**
```bash
if [ "$score" -ge 90 ]; then
    echo "Grade: A"
elif [ "$score" -ge 80 ]; then
    echo "Grade: B"
elif [ "$score" -ge 70 ]; then
    echo "Grade: C"
else
    echo "Grade: F"
fi
```

**Kukicha:**
```kukicha
function main()
    score := 85

    if score >= 90
        print("Grade: A")
    else if score >= 80
        print("Grade: B")
    else if score >= 70
        print("Grade: C")
    else
        print("Grade: F")
```

Save this as `decisions.kuki` and run it:

```bash
kukicha run decisions.kuki
```

No `then`, no `fi`, no `[ ]`, no `-ge` / `-lt` / `-eq`. Just the comparison operators you'd expect: `>=`, `<=`, `>`, `<`, `equals`, `not equals`.

### Combining Conditions

**Bash:**
```bash
if [ "$age" -ge 18 ] && [ "$has_ticket" = "true" ]; then
    echo "Welcome"
fi
```

**Kukicha:**
```kukicha
function main()
    age := 25
    hasTicket := true

    if age >= 18 and hasTicket
        print("Welcome")
```

- `and` instead of `&&` or `-a`
- `or` instead of `||` or `-o`
- `not` instead of `!`

No need to remember which bracket syntax supports which operators.

---

## Functions - Like Shell Functions, But Better

Bash functions are basically named blocks of commands. They can't declare parameter types, can't return values (only exit codes), and communicate through global variables or stdout.

**Bash:**
```bash
greet() {
    local name=$1
    echo "Hello, $name!"
}
greet "Alice"
```

**Kukicha:**
```kukicha
function Greet(name string)
    print("Hello, {name}!")

function main()
    Greet("Alice")
```

Save this as `functions.kuki` and run it:

```bash
kukicha run functions.kuki
```

Parameters are named and typed - no more positional `$1`, `$2`, `$3`.

### Functions That Return Values

In bash, you'd capture stdout or use a global variable. In Kukicha, functions return values directly:

**Bash:**
```bash
add() {
    echo $(( $1 + $2 ))
}
result=$(add 5 3)
```

**Kukicha:**
```kukicha
function Add(a int, b int) int
    return a + b

function main()
    result := Add(5, 3)
    print(result)  # Prints: 8
```

The `int` after the parentheses is the return type. `return` sends a value back to the caller. No subshells, no stdout capture.

### Rules of Thumb

- Function parameters need explicit types: `name string`, `count int`
- Local variables inside functions get types inferred: `x := 5` just works
- Functions that produce a value declare a return type and use `return`

---

## Lists - Like Arrays, But They Actually Work

Bash arrays are notoriously inconsistent. Different syntax for declaration, access, length, iteration. Kukicha lists are uniform.

**Bash:**
```bash
fruits=("apple" "banana" "cherry")
echo "${fruits[0]}"
echo "${#fruits[@]}"
fruits+=("date")
```

**Kukicha:**
```kukicha
function main()
    fruits := list of string{"apple", "banana", "cherry"}
    print(fruits[0])        # apple
    print(len(fruits))      # 3
    fruits = append(fruits, "date")
```

Save this as `lists.kuki` and run it:

```bash
kukicha run lists.kuki
```

- `list of string` declares what the list holds
- Indexing starts at 0 (same as bash)
- `fruits[-1]` gets the last item (like Python, unlike bash)
- `len(fruits)` gives you the count (not `${#fruits[@]}`)
- `append(fruits, item)` returns a new list with the item added

### Other List Types

```kukicha
function main()
    ports := list of int{80, 443, 8080}
    flags := list of bool{true, false, true}
```

---

## Loops - For Without Do/Done

### Iterating Over a List

**Bash:**
```bash
for fruit in "${fruits[@]}"; do
    echo "I like $fruit"
done
```

**Kukicha:**
```kukicha
function main()
    fruits := list of string{"apple", "banana", "cherry"}

    for fruit in fruits
        print("I like {fruit}!")
```

Save this as `loops.kuki` and run it:

```bash
kukicha run loops.kuki
```

No `do`, no `done`, no `"${array[@]}"` incantation.

### With Index

**Bash:**
```bash
for i in "${!fruits[@]}"; do
    echo "$i: ${fruits[$i]}"
done
```

**Kukicha:**
```kukicha
function main()
    fruits := list of string{"apple", "banana", "cherry"}

    for i, fruit in fruits
        print("{i}: {fruit}")
```

### Counting Loops

**Bash:**
```bash
for i in $(seq 1 5); do
    echo "$i"
done
```

**Kukicha:**
```kukicha
function main()
    # 'to' is exclusive: prints 1, 2, 3, 4
    for i from 1 to 5
        print(i)

    # 'through' is inclusive: prints 1, 2, 3, 4, 5
    for i from 1 through 5
        print(i)
```

### While Loops

**Bash:**
```bash
count=5
while [ "$count" -gt 0 ]; do
    echo "$count"
    count=$((count - 1))
done
```

**Kukicha:**
```kukicha
function main()
    count := 5

    for count > 0
        print(count)
        count = count - 1
```

Kukicha uses `for condition` instead of `while` - one keyword for all loops.

### Increment and Decrement

Same as bash arithmetic - `++` adds 1, `--` subtracts 1:

```kukicha
function main()
    count := 0
    count++       # count is now 1
    count++       # count is now 2
    count--       # count is now 1
```

### Break and Continue

Same as bash - `break` exits the loop, `continue` skips to the next iteration:

```kukicha
function main()
    names := list of string{"Alice", "Bob", "Charlie", "Diana"}

    for name in names
        if name equals "Charlie"
            print("Found Charlie!")
            break
        print("Not {name}...")
```

---

## Pipes - You Already Know This One

This is where your shell instincts pay off. In bash, you pipe data between commands:

```bash
cat users.csv | grep "active" | sort | head -5
```

Kukicha has the same concept with `|>`, but for functions:

```kukicha
import "stdlib/string"

function main()
    text := "  HELLO world  "

    clean := text |> string.TrimSpace() |> string.ToLower() |> string.Title()
    print(clean)  # "Hello World"
```

The data flows left to right, just like a shell pipeline. The result of each step feeds into the next.

Without pipes, this would be nested function calls (like bash without pipes would be temp files everywhere):

```kukicha
# Without pipes - harder to read
clean := string.Title(string.ToLower(string.TrimSpace(text)))
```

### More Pipe Examples

```kukicha
import "stdlib/string"

function main()
    # Split a string and join with a different separator
    result := "apple,banana,cherry" |> string.Split(",") |> string.Join(" | ")
    print(result)  # "apple | banana | cherry"

    # Check if a string contains something
    if "Hello World" |> string.Contains("World")
        print("Found it!")
```

Pipes work with any function that takes an input and returns an output - they're not limited to strings. You'll see them used heavily with error handling and data processing in later tutorials.

---

## Running Commands

One of the biggest reasons to use a "real" language for scripts is better command execution. No more unquoted variables breaking your `rm` command.

```kukicha
import "stdlib/shell"

function main()
    # Run a command, capture the output
    status := shell.Output("git", "status", "--short") onerr ""

    if status equals ""
        print("Working directory clean (or not a git repo).")
        return

    print("Changed files:")
    print(status)
```

Save this as `git_check.kuki` and run it in a git repo:

```bash
kukicha run git_check.kuki
```

Each argument is a separate string - no word splitting, no glob expansion, no quoting issues. `shell.Output("rm", "-rf", path)` is always safe, even if `path` has spaces.

The `onerr ""` part is Kukicha's error handling, which we'll cover next.

---

## Error Handling - Better Than set -e

In bash, error handling is rough. You either ignore errors, use `set -e` (which has surprising behavior), or check `$?` after every command:

```bash
output=$(some_command 2>/dev/null)
if [ $? -ne 0 ]; then
    echo "Failed"
    exit 1
fi
```

Kukicha has `onerr` - inline error handling that's clear about what happens when something fails:

```kukicha
import "stdlib/shell"

function main()
    # If the command fails, use an empty string instead
    output := shell.Output("git", "log", "--oneline", "-5") onerr ""

    # If the command fails, stop the program with a message
    output := shell.Output("git", "log", "--oneline", "-5") onerr panic "git failed: {error}"

    # If the command fails, return from the function
    output := shell.Output("git", "log", "--oneline", "-5") onerr return
```

The `onerr` goes right at the end of the line that might fail. You can:
- Provide a fallback value: `onerr ""`, `onerr 0`, `onerr empty`
- Stop the program: `onerr panic "message"`
- Return from the function: `onerr return`
- Wrap error with a hint: `onerr explain "hint message"` or `onerr "default" explain "hint message"`
- Run a block of code:

```kukicha
    output := shell.Output("git", "status") onerr
        print("Command failed: {error}")
        return
```

The `{error}` variable is automatically available inside `onerr` blocks - it contains the error message.

This is clearer than `set -e` because you decide **per operation** what should happen on failure, instead of one global behavior for your whole script.

---

## Putting It Together - A Log Analyzer

Let's combine what you've learned into something practical - a script that processes log lines, counts errors, and reports results. This is the kind of thing you'd normally write as a bash one-liner that grew into an unmaintainable 200-line script.

Create a file called `log_analyzer.kuki`:

```kukicha
import "stdlib/string"

function Severity(line string) string
    if line |> string.Contains("ERROR:")
        return "ERROR"
    else if line |> string.Contains("WARN:")
        return "WARN"
    return "INFO"

function main()
    # Simulated log lines (in practice, you'd read a file)
    logs := list of string{
        "2024-01-15 ERROR: disk full on /dev/sda1",
        "2024-01-15 INFO: backup started",
        "2024-01-15 WARN: high memory usage (92%)",
        "2024-01-15 ERROR: connection timeout to db-primary",
        "2024-01-15 INFO: backup completed",
        "2024-01-15 WARN: certificate expires in 7 days",
        "2024-01-15 ERROR: failed to rotate logs",
    }

    errors := 0
    warns := 0
    infos := 0

    print("=== Log Analysis ===\n")

    for line in logs
        level := Severity(line)

        if level equals "ERROR"
            errors++
            print("[!] {line}")
        else if level equals "WARN"
            warns++

    print("\n=== Summary ===")
    print("Total lines: {len(logs)}")
    print("Errors:      {errors}")
    print("Warnings:    {warns}")
    print("Info:        {len(logs) - errors - warns}")

    if errors > 0
        print("\nAction required: {errors} error(s) found.")
```

Run it:

```bash
kukicha run log_analyzer.kuki
```

Expected output:
```
=== Log Analysis ===

[!] 2024-01-15 ERROR: disk full on /dev/sda1
[!] 2024-01-15 ERROR: connection timeout to db-primary
[!] 2024-01-15 ERROR: failed to rotate logs

=== Summary ===
Total lines: 7
Errors:      3
Warnings:    2
Info:        2

Action required: 3 error(s) found.
```

This is the kind of script that in bash would involve `grep -c`, `awk`, temp files, and careful quoting. In Kukicha, it's typed, structured, and compiles to a fast binary.

---

## What's Next?

You now know how to translate your shell scripting knowledge into Kukicha:

| Shell Concept | Kukicha Equivalent |
|---|---|
| `VAR="value"` | `name := "value"` |
| `echo "$VAR"` | `print("{name}")` |
| `$1`, `$2` (positional args) | Named, typed function parameters |
| `if [ ... ]; then ... fi` | `if condition` with indentation |
| `-eq`, `-gt`, `-lt` | `equals`, `>`, `<` |
| `&&`, `\|\|` | `and`, `or` |
| `for x in ...; do ... done` | `for x in ...` with indentation |
| `cmd1 \| cmd2 \| cmd3` | `val \|> func1() \|> func2()` |
| `result=$(command)` | `result := shell.Output(...)` |
| `set -e` / `$?` | `onerr` |
| `array=("a" "b")` | `list of string{"a", "b"}` |
| `${#array[@]}` | `len(list)` |

### Continue Your Journey

| # | Tutorial | What You'll Learn |
|---|----------|-------------------|
| 1 | **You are here** | Variables, functions, strings, conditionals, lists, loops, pipes |
| 2 | **[Data & AI Scripting](data-scripting-tutorial.md)** | Maps (key-value), parsing CSVs, shell commands, AI scripting |
| 3 | **[CLI Explorer](cli-explorer-tutorial.md)** | Custom types, methods, API data, arrow lambdas, error handling |
| 4 | **[Link Shortener](web-app-tutorial.md)** | HTTP servers, JSON, REST APIs, redirects |
| 5 | **[Production Patterns](production-patterns-tutorial.md)** | Databases, concurrency, Go conventions |

### Additional Resources

- **[Kukicha Grammar](../kukicha-grammar.ebnf.md)** - Complete language grammar reference
- **[Stdlib Reference](../kukicha-stdlib-reference.md)** - Standard library documentation
- **[Examples](../../examples/)** directory - More example programs
