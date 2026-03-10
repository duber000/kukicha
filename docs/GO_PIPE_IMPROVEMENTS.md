# Addressing the Procedural vs. Functional Pipe Split in Kukicha

> **Status (v0.0.12):** Items 1 (pipe-aware iterators), 2 (pipeline-level onerr), and 4 (piped switch) are implemented. Item 3 (implicit `it`) was prototyped and removed — explicit typed lambdas are preferred for type safety.

While Kukicha's pipe operator (`|>`) offers excellent readability for data transformations, the language's reliance on Go's procedural control flow and strict error handling can lead to a "split identity" in codebases.

To remain as Go-compatible as possible while leaning fully into the pipe-forward architecture, the following language design improvements could be explored:

## 1. Pipe-Aware Iterators (Leveraging Go 1.23)
Go 1.23 introduced the `iter.Seq` type, standardizing looping. Kukicha could automatically compile piped `list of` and `map of` collections to iterators, allowing procedural loops to become natural pipeline terminators.

Instead of breaking a pipe to proceduralize it:

```kukicha
popular := repos |> slice.Filter((r Repo) => r.stars > 1000)
for repo in popular
    print(repo.name)
```

You could terminate the pipe with an iterator consumer:

```kukicha
repos 
    |> slice.Filter((r Repo) => r.stars > 1000)
    |> slice.Each((r Repo) => print(r.name))
```

This maps perfectly to Go's new `range` over func capabilities, maintaining a zero-cost abstraction while satisfying the functional pipeline style.

## 2. Pipeline-Level Error Catching
Currently, Kukicha puts `onerr` on specific steps:

```kukicha
prices := fetch.Get(url) |> fetch.CheckStatus() onerr return error
```

If you chain 10 operations that can fail, putting `onerr` on every single one mixes structural error handling with data flow. 

Kukicha could allow an `onerr` block to terminate an entire pipe group, behaving like a transparent Result monad. If *any* function in the pipe returns a Go `error`, the pipeline short-circuits to the `onerr` block at the end:

```kukicha
processed := data 
    |> parse.Json(list of User) 
    |> fetch.EnrichWithDB() 
    |> validate.Safe()
onerr as e
    log.Error(e)
    return empty
```

Under the hood, the compiler generates standard Go `if err != nil` checks between every single function call in the pipeline, maintaining 100% Go semantics while keeping the `.kuki` code entirely focused on the "happy path" data flow.

## 3. Implicit `it` for Go-Native Methods
When piping data into a standard Go library function that wasn't designed for pipes, the chain often breaks. Kukicha currently allows `_` as a placeholder (`todo |> json.MarshalWrite(response, _)`). 

To lean all the way into pipelines, Kukicha could adopt an implicit `it` variable (similar to Kotlin) inside block lambdas, removing the need to declare procedural boilerplate just to map Go structs:

```kukicha
# Currently:
result := data |> slice.Map((req HttpRequest) => req.Header.Get("Auth"))

# With 'it':
result := data |> slice.Map(() => it.Header.Get("Auth"))
```

## 4. Piped Control Flow
To prevent developers from having to assign a pipeline to a variable just to use an `if` statement, Kukicha could allow procedural control flow keywords to act as pipeline consumers.

```kukicha
user.Role 
    |> switch
        when "admin"
            grantAccess()
        when "guest"
            denyAccess()
        otherwise
            checkPermissions()
```

This is purely syntactic sugar. The compiler assigns the left side of the pipe to a hidden Go variable (e.g., `_kuki_var_1 := user.Role`) and generates a standard Go `switch _kuki_var_1 { case ... }`. 

### Summary
The way to fix the split identity is to **make the Go ecosystem conform to the pipe in syntax, rather than making the pipe conform to Go**. By using the compiler to generate the tedious boilerplate (like `if err != nil` between pipe stages, or wrapping variables in `switch` statements), Kukicha could offer a purely functional, top-to-bottom reading experience while still outputting safe, procedural, idiomatic Go code on the back end.
