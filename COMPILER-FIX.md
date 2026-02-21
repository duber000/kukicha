For functions imported from other packages (like llm or shell), they must be registered in the knownExternalReturns map in 

internal/semantic/semantic.go
.

Here is the technical breakdown of why this is necessary in the current compiler architecture:

1. Local vs. Imported Functions
Local Functions: If you define a function in the same file or package, the compiler has the full AST (Abstract Syntax Tree). It can look at the func declaration, count the items in the return list, and know exactly what's happening.
Imported Functions: When you call llm.Ask(), the compiler (at least in its current MVP state) doesn't perform a "deep" cross-package analysis of every dependency's source code during the semantic pass of the main script. It sees llm.Ask as a "qualified identifier" and needs a way to know its signature without re-parsing the entire llm package.
2. The onerr Pipe Problem
The onerr keyword with pipes (e.g., x |> f() onerr ...) is actually quite complex for the transpiler. It has to "decompose" that single line into multiple Go statements:

Call the first part.
Check if it returns an error.
Pass the result to the next part.
If the compiler doesn't know that llm.Ask returns an error, it generates Go code like pipe_3 := llm.Ask(...) instead of pipe_3, err := llm.Ask(...). This leads to the "assignment mismatch" error you saw, because Go expects two variables for that assignment.

3. Future Improvements
Ideally, the compiler would:

Cache Metadata: Generate a "header" or "type info" file for each package so it can look up signatures without manual registration.
Automatic Inference: Automatically crawl the stdlib during the build to populate this registry.
For now, the manual registry is the "Source of Truth" that tells the transpiler how to correctly unwrap errors from standard library function
