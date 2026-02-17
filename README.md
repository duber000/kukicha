# Kukicha

**Write code that reads like English. Compile it to blazing-fast Go.**

Kukicha is a programming language that transpiles to idiomatic Go code. If you're comfortable with shell scripts but find Go's symbols intimidating, or you want Python's readability with Go's performance and deployment story, Kukicha is for you. No runtime overhead. No magic. Just cleaner syntax that becomes real Go.

```kukicha
import "stdlib/slice"

func isActive(user User) bool
    return user.active

func getName(user User) string
    return user.name

func main()
    users := fetchUsers() onerr panic "failed to fetch"

    active := users
        |> slice.Filter(isActive)
        |> slice.Map(getName)

    for name in active
        print("Hello {name}!")
```

---

## Quickstart

### Download A Release Binary

Pick your OS/arch from the GitHub releases, download, and run:

```bash
VERSION=v0.0.4
OS=linux
ARCH=amd64
curl -L -o kukicha.tar.gz \
  "https://github.com/duber000/kukicha/releases/download/${VERSION}/kukicha_${VERSION}_${OS}_${ARCH}.tar.gz"
tar -xzf kukicha.tar.gz
./kukicha version
```

Windows uses `.zip` archives and `kukicha.exe`.

### Or Install With Go

```bash
go install github.com/duber000/kukicha/cmd/kukicha@v0.0.4
kukicha version
```

### First Run

```bash
kukicha init                # Extract stdlib + configure go.mod
kukicha run hello.kuki      # Transpile, build, run
```

---

## Why Kukicha?

### For Shell Scripters

Your bash scripts work, but they're getting harder to maintain. Kukicha keeps what you like - pipes, running commands, readable flow - and adds real types, proper error handling, and compiled binaries.

| Bash | Kukicha |
|------|---------|
| `echo "$name is $age"` | `print("{name} is {age}")` |
| `if [ ... ]; then ... fi` | `if condition` with indentation |
| `cmd1 \| cmd2 \| cmd3` | `val \|> func1() \|> func2()` |
| `result=$(command)` | `result := shell.Output(...)` |
| `set -e` / `$?` | `onerr` per operation |

### For Python Developers

You already know the syntax - `and`/`or`/`not`, indentation, `for x in items`, `# comments`. Kukicha adds compile-time type checking, real concurrency, and single-binary deployment.

### For Go Developers

Same ecosystem, same performance, less boilerplate:

| Go | Kukicha |
|----|---------|
| `&&`, `\|\|`, `!` | `and`, `or`, `not` |
| `*User`, `&user` | `reference User`, `reference of user` |
| `if err != nil { return err }` | `onerr return error "{error}"` |
| `[]string{"a", "b"}` | `list of string{"a", "b"}` |
| Curly braces everywhere | Indentation (4 spaces) |

See the [FAQ](docs/faq.md) for detailed comparisons and migration paths.

---

## AI Disclosure

Built with the assistance of AI. Review and test before production use.

---

## Quick Taste

### 1. Data Pipelines
Fetch and parse real data with zero boilerplate. The pipe operator (`|>`) and `onerr` make error handling elegant.

```kukicha
import "stdlib/fetch"

type Repo
    Name string as "name"
    Stars int as "stargazers_count"

func main()
    # Fetch, check status, and decode JSON in one pipeline
    repos := empty list of Repo
    fetch.Get("https://api.github.com/users/golang/repos")
        |> fetch.CheckStatus()
        |> fetch.JsonAs(_, reference of repos) onerr panic "API call failed: {error}"

    for repo in repos[:5]
        print("- {repo.Name}: {repo.Stars} stars")
```

### 2. Concurrency
Run thousands of tasks in parallel using Goroutines and Channels. High performance, zero complexity.

```kukicha
import "time"

func check(url string, results channel of string)
    # Background work...
    send "{url} is UP" to results

func main()
    urls := list of string{"google.com", "github.com", "go.dev"}
    results := make channel of string

    for url in urls
        go check(url, results)

    for i from 0 to len(urls)
        print(receive from results)
```

### 3. AI Scripting
Native LLM support lets you build AI-powered tools as easily as writing a "Hello World".

```kukicha
import "stdlib/llm"
import "stdlib/shell"

func main()
    # Pipe a git diff directly into an LLM for commit message generation
    diff := shell.Output("git", "diff", "--staged") onerr return

    message := llm.New("gpt-5-nano")
        |> llm.System("Write a concise git commit message for this diff.")
        |> llm.Ask(diff) onerr panic "AI Error: {error}"

    print("Suggested: {message}")
```

---

## Install

**Requirements:** Go 1.26+

```bash
git clone https://github.com/duber000/kukicha.git
cd kukicha
go build -o kukicha ./cmd/kukicha
```

---

## Usage

```bash
kukicha init
kukicha build myapp.kuki
kukicha run myapp.kuki
kukicha check myapp.kuki
kukicha fmt myapp.kuki
```

---

## Contributing

See [Contributing Guide](docs/contributing.md) for development setup, tests, and architecture.

---

## Documentation

**Tutorials:**
- [Beginner Tutorial](docs/tutorials/beginner-tutorial.md) - for shell scripters moving to Kukicha
- [Data & AI Scripting](docs/tutorials/data-scripting-tutorial.md) - maps, CSV parsing, shell commands, LLM integration
- [CLI Repo Explorer](docs/tutorials/cli-explorer-tutorial.md) - custom types, methods, API data
- [Concurrent Health Checker](docs/tutorials/concurrent-url-health-checker.md) - goroutines and channels

**Reference:**
- [FAQ](docs/faq.md) - coming from bash, Python, or Go
- [Quick Reference](docs/kukicha-quick-reference.md) - Go-to-Kukicha translation table
- [Stdlib Reference](docs/kukicha-stdlib-reference.md) - standard library documentation

---

## Status

**Version:** 0.0.4
**Status:** Ready for testing
**Go:** 1.26+ required

---

## License

See [LICENSE](LICENSE) for details.
