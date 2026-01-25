# Frequently Asked Questions

## What are the main differences between Kukicha and Go+ (formerly Goplus)?

While both Kukicha and Go+ aim to make Go more accessible and productive, they have different design philosophies and target audiences. Here’s a high-level comparison:

| Feature | Kukicha | Go+ (Goplus) |
|---|---|---|
| **Core Philosophy** | **"Go for Scripts"**: Aims to be a powerful, readable scripting language that transpiles to idiomatic Go. | **"Go for STEM"**: Aims to be a simpler, more beginner-friendly language for engineering, data science, and education. |
| **Syntax Style** | Indentation-based, English-like keywords (`equals`, `and`, `or`), and a pipe operator (`|>`). | Go-like, but with simplifications like list comprehensions and optional struct type declarations. |
| **Error Handling** | The `onerr` keyword provides a consistent way to handle errors from functions that return `(T, error)`. | `ErrWrap` expressions (`!`, `?`, `?:defval`) simplify error handling by allowing for panicking, returning, or providing a default value. |
| **Standard Library**| A rich standard library focused on scripting tasks, with packages for iteration, slice manipulation, file operations, and more. | Adds features for data science and STEM, like native rational numbers and operator overloading. |
| **Target Audience** | Developers who want a more expressive and readable language for scripting, automation, and building CLI tools. | Beginners, students, data scientists, and engineers who want a simpler introduction to the Go ecosystem. |

### Key Differences in a Nutshell

*   **Kukicha is a scripting-friendly superset of Go.** It focuses on providing a rich set of tools and a readable syntax for common scripting tasks, while still allowing you to use any existing Go package. The pipe operator (`|>`) is a central feature, enabling elegant data processing pipelines.

*   **Go+ is a simplified version of Go.** It adds features from other languages (like Python's list comprehensions) and introduces new concepts (like native rational numbers) to make the language more approachable for a wider audience, particularly in the fields of data science and education.

In summary, if you're looking for a powerful and expressive language for scripting and automation with a focus on readability and a rich standard library, **Kukicha** is a great choice. If you're a beginner, a data scientist, or an educator looking for a simpler and more accessible entry point into the Go ecosystem, **Go+** might be a better fit.
