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
VERSION=v0.0.1
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
go install github.com/duber000/kukicha/cmd/kukicha@v0.0.1
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

```kukicha
type Todo
    id int json:"id"
    title string json:"title"
    done bool json:"done"

func Display on todo Todo string
    status := "[ ]"
    if todo.done
        status = "[x]"
    return "{status} {todo.title}"

func main()
    todos := list of Todo{
        Todo
            id: 1
            title: "Learn Kukicha"
            done: true
        Todo
            id: 2
            title: "Build something"
            done: false
    }

    for todo in todos
        print(todo.Display())

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
- [FAQ](docs/faq.md)
- [Quick Reference](docs/kukicha-quick-reference.md)
- [Stdlib Reference](docs/kukicha-stdlib-reference.md)

---

## Status

**Version:** 0.0.1
**Status:** Ready for testing
**Go:** 1.25+ required

---

## License

See [LICENSE](LICENSE) for details.
