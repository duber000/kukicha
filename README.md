# Kukicha

**Write code that reads like English. Compile it to blazing-fast Go.**

Kukicha is a beginner-friendly programming language that transpiles to idiomatic Go code. No runtime overhead. No magic. Just cleaner syntax that becomes real Go.

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
VERSION=v0.0.2
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
go install github.com/duber000/kukicha/cmd/kukicha@v0.0.2
kukicha version
```

### First Run

```bash
kukicha init                # Extract stdlib + configure go.mod
kukicha run hello.kuki      # Transpile, build, run
```

---

## Why Kukicha?

### For Beginners

**Go is powerful but intimidating.** Pointers (`*`, `&`), error handling boilerplate, and cryptic symbols create a steep learning curve.

Kukicha fixes this:

| Go | Kukicha |
|----|---------|
| `&&`, `\|\|`, `!` | `and`, `or`, `not` |
| `*User`, `&user` | `reference User`, `reference of user` |
| `nil` | `empty` |
| `if err != nil { return err }` | `onerr return error "{error}"` |
| `break`, `continue` | `break`, `continue` |
| `for { ... }` | `for` |
| `v.(T)` | `v as T` |

**Learn programming concepts, not symbols.** When you're ready, the generated Go code teaches you Go itself.

### For DevOps And Go Developers

- **Single static binary** deployment
- **Zero runtime overhead** (transpiles to idiomatic Go)
- **Full Go ecosystem access** (import any Go package)

---

## AI Disclosure

Built with the assistance of AI. Review and test before production use.

---

## Quick Taste

### 1. Data Pipelines
Fetch and parse real data with zero boilerplate. The pipe operator (`|>`) and `onerr` make error handling elegant.

```kukicha
import "stdlib/fetch"
import "stdlib/json"

type Repo
    Name string json:"name"
    Stars int json:"stargazers_count"

func main()
    # Fetch, check status, and parse JSON in one pipeline
    repos := empty list of Repo
    fetch.Get("https://api.github.com/users/golang/repos")
        |> fetch.CheckStatus()
        |> fetch.Bytes()
        |> json.Unmarshal(reference repos)
        onerr panic "API call failed: {error}"

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
        |> llm.Ask(diff)
        onerr panic "AI Error: {error}"

    print("Suggested: {message}")
```

---

## Install

**Requirements:** Go 1.25+

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

- [Beginner Tutorial](docs/tutorial/beginner-tutorial.md)
- [Data & AI Scripting](docs/tutorial/data-scripting-tutorial.md)
- [CLI Repo Explorer](docs/tutorial/cli-explorer-tutorial.md)
- [Concurrent Health Checker](docs/tutorial/concurrent-url-health-checker.md)
- [FAQ](docs/faq.md)
- [Quick Reference](docs/kukicha-quick-reference.md)
- [Stdlib Reference](docs/kukicha-stdlib-reference.md)

---

## Status

**Version:** 0.0.2
**Status:** Ready for testing
**Go:** 1.25+ required

---

## License

See [LICENSE](LICENSE) for details.
