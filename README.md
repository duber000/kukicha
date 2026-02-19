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
VERSION=v0.0.5
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
go install github.com/duber000/kukicha/cmd/kukicha@v0.0.5
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
import "stdlib/fetch"

func check(url string, results channel of string)
    resp := fetch.Get(url) onerr
        send "{url} is DOWN ({error})" to results
        return
    resp |> fetch.CheckStatus() onerr
        send "{url} returned {resp.StatusCode}" to results
        return
    send "{url} is UP" to results

func main()
    urls := list of string{"https://google.com", "https://github.com", "https://go.dev"}
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

## Standard Library

35+ packages built pipe-first. Chain operations left-to-right, handle errors inline with `onerr`.

```kukicha
import "stdlib/fetch"
import "stdlib/slice"
import "stdlib/string"

# Fetch → check → decode → filter → map — one pipeline
repos := fetch.Get("https://api.github.com/users/golang/repos")
    |> fetch.CheckStatus()
    |> fetch.Json(list of Repo) onerr panic "{error}"

names := repos
    |> slice.Filter((r Repo) => r.Stars > 1000)
    |> slice.Map((r Repo) => r.Name |> string.ToUpper())
```

```kukicha
import "stdlib/pg"
import "stdlib/kube"
import "stdlib/must"
import "stdlib/env"

# Startup config — panic if required vars are missing
apiKey := must.Env("API_KEY")
port := env.GetOr("PORT", ":8080")

# Databases and clusters with built-in retry
pool := pg.New(dbURL) |> pg.Retry(5, 500) |> pg.Open() onerr panic "{error}"
cluster := kube.New() |> kube.Retry(3, 1000) |> kube.Open() onerr panic "{error}"
```

```kukicha
import "stdlib/validate"
import "stdlib/cli"
import "stdlib/concurrent"

# Validate inputs — each returns an error for onerr
email |> validate.Email() onerr return error "{error}"
age |> validate.InRange(18, 120) onerr return error "{error}"

# Run tasks in parallel with one line
results := concurrent.Map(urls, (u string) => fetch.Get(u))

# Parse CLI args with a builder
app := cli.New("myapp") |> cli.Arg("name", "User name") |> cli.Action(run)
```

See the full [Stdlib Reference](docs/kukicha-stdlib-reference.md) for all packages.

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
- [Absolute Beginner](docs/tutorials/absolute-beginner-tutorial.md) - first program, variables, functions, lists, loops
- [Shell Scripters Guide](docs/tutorials/beginner-tutorial.md) - for bash users moving to Kukicha
- [Data & AI Scripting](docs/tutorials/data-scripting-tutorial.md) - maps, CSV parsing, shell commands, LLM integration
- [CLI Repo Explorer](docs/tutorials/cli-explorer-tutorial.md) - custom types, methods, API data
- [Link Shortener](docs/tutorials/web-app-tutorial.md) - HTTP servers, JSON, REST APIs, redirects
- [Concurrent Health Checker](docs/tutorials/concurrent-url-health-checker.md) - goroutines and channels
- [Production Patterns](docs/tutorials/production-patterns-tutorial.md) - databases, validation, retry, auth

**Reference:**
- [FAQ](docs/faq.md) - coming from bash, Python, or Go
- [Quick Reference](docs/kukicha-quick-reference.md) - Go-to-Kukicha translation table
- [Stdlib Reference](docs/kukicha-stdlib-reference.md) - standard library documentation

---

## Status

**Version:** 0.0.5
**Status:** Ready for testing
**Go:** 1.26+ required

---

## License

See [LICENSE](LICENSE) for details.
