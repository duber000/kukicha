# Kukicha Standard Library Roadmap

**Version:** 1.0.0
**Status:** Planning Phase

This document outlines planned standard library packages and features for future Kukicha releases.

---

## Core Libraries (Priority 1)

### Slices Package

Extended slice operations for common patterns:

```kukicha
import slices

# Take/drop operations
firstThree := slices.first(items, 3)
lastTwo := slices.last(items, 2)
tail := slices.drop(items, 3)
head := slices.dropLast(items, 1)

# Pipeline-friendly
result := items
    |> slices.drop(2)
    |> slices.first(10)
    |> process()

# Additional operations
reversed := slices.reverse(items)
unique := slices.unique(items)
chunked := slices.chunk(items, 5)
```

### HTTP Package

Simple HTTP client and server:

```kukicha
import http

# Client
response := http.get("https://api.example.com/users")
    |> .json() as list of User
    onerr return error "fetch failed"

# POST with JSON
user := User{name: "Alice", age: 30}
response := http.post("https://api.example.com/users", user)

# Server
http.handle("/users", func(req http.Request) http.Response
    users := getUsers()
    return http.json(users)
)

http.listen(":8080")
```

### JSON Package

JSON encoding and decoding:

```kukicha
import json

# Parse JSON string
config := json.parse(jsonString) as Config
    onerr return error "invalid JSON"

# Encode to JSON
jsonString := json.encode(config)

# Pretty print
prettyJson := json.pretty(config)
```

### File Package

File I/O operations:

```kukicha
import file

# Read entire file
content := file.read("config.json")
    onerr return error "cannot read file"

# Write file
file.write("output.txt", content)
    onerr return error "cannot write file"

# Append to file
file.append("log.txt", logEntry)

# File existence
if file.exists("config.json")
    loadConfig()

# Directory operations
files := file.list("./data/")
file.mkdir("./output/")
file.remove("temp.txt")
```

### String Package

String manipulation utilities:

```kukicha
import string

# Case conversion
upper := string.upper("hello")
lower := string.lower("WORLD")
title := string.title("hello world")

# Splitting and joining
parts := string.split("a,b,c", ",")
joined := string.join(parts, "|")

# Trimming
trimmed := string.trim("  hello  ")
trimmedLeft := string.trimLeft("  hello")
trimmedRight := string.trimRight("hello  ")

# Searching
if string.contains(text, "error")
    handleError()

startsWith := string.hasPrefix(text, "http://")
endsWith := string.hasSuffix(filename, ".kuki")
```

---

## Cloud & Infrastructure (Priority 2)

### Docker Package

Docker container management:

```kukicha
import docker

# Build image
image := docker.build("my-app:latest", "./Dockerfile")
    onerr return error "build failed"

# Run container
container := docker.run(image,
    ports: map of string to string{"8080": "8080"}
    env: map of string to string{"ENV": "production"}
)

# Container operations
docker.stop(container)
docker.logs(container)
docker.remove(container)

# List containers
containers := docker.list()
for discard, c in containers
    print "{c.name}: {c.status}"
```

### Kubernetes Package

Kubernetes cluster management:

```kukicha
import k8s

# Deploy application
deployment := k8s.deploy("my-app",
    image: "my-app:v1.0.0"
    replicas: 3
    namespace: "production"
)

# Wait for rollout
k8s.waitReady(deployment)
    onerr return error "deployment failed"

# Get pods
pods := k8s.getPods("production", "app=my-app")
for discard, pod in pods
    print "{pod.name}: {pod.status}"

# Scale deployment
k8s.scale(deployment, 5)

# Get logs
logs := k8s.logs("my-app-pod-123")
```

---

## AI & LLM Integration (Priority 2)

### Claude Package

Integration with Anthropic's Claude API:

```kukicha
import claude

# Simple completion
response := claude.complete("Explain Kukicha in one sentence")
    onerr return error "API call failed"

# Structured conversation
conversation := claude.newConversation()
conversation.addMessage("user", "What is Go?")
response := conversation.send()

# Analysis with context
analysis := file.read("logs.txt")
    |> extractErrors()
    |> claude.complete("Analyze these errors and suggest fixes")
    |> formatReport()
```

### OpenAI Package

Integration with OpenAI API:

```kukicha
import openai

# GPT completion
response := openai.complete("gpt-4", "Translate to Spanish: Hello")

# Image generation
image := openai.generateImage("A sunset over mountains")
file.write("sunset.png", image)
```

---

## Database (Priority 3)

### SQL Package

Database operations:

```kukicha
import sql

# Connect to database
db := sql.connect("postgres://localhost/mydb")
    onerr return error "connection failed"

defer db.close()

# Query
users := db.query("SELECT * FROM users WHERE active = true") as list of User
    onerr return error "query failed"

# Execute
db.exec("UPDATE users SET last_login = NOW() WHERE id = $1", userId)

# Transactions
db.transaction(func(tx sql.Transaction)
    tx.exec("INSERT INTO orders (user_id, total) VALUES ($1, $2)", userId, total)
    tx.exec("UPDATE inventory SET stock = stock - $1 WHERE id = $2", quantity, itemId)
)
```

---

## Testing (Priority 3)

### Test Package

Unit testing framework:

```kukicha
import test

func TestCalculator(t test.T)
    result := add(2, 3)
    test.assertEqual(t, result, 5, "2 + 3 should equal 5")

    result = divide(10, 2)
    test.assertEqual(t, result, 5, "10 / 2 should equal 5")

func TestErrorHandling(t test.T)
    discard, err := divide(10, 0)
    test.assertError(t, err, "divide by zero should return error")
```

---

## Tooling Enhancements

### Package Manager

```bash
# Install dependencies
kuki get github.com/user/package@v1.0.0

# Update dependencies
kuki update

# Vendor dependencies
kuki vendor
```

### IDE Support

- **VS Code Extension**
  - Syntax highlighting
  - IntelliSense/autocomplete
  - Go-to-definition
  - Error highlighting
  - Code formatting on save

- **LSP Server**
  - Language Server Protocol implementation
  - Works with any LSP-compatible editor

### Debugger

```bash
# Debug a program
kuki debug main.kuki

# Set breakpoints
(kuki-db) break main.kuki:15
(kuki-db) run
(kuki-db) print count
(kuki-db) step
(kuki-db) continue
```

### Formatter Enhancements

```bash
# Format with style options
kuki fmt --style=compact main.kuki
kuki fmt --style=expanded main.kuki

# Auto-fix common issues
kuki fmt --fix main.kuki

# Format on save (IDE integration)
```

---

## Contributing

Want to help build the Kukicha standard library? See [CONTRIBUTING.md](../CONTRIBUTING.md) for guidelines.

**Priority areas:**
1. Core libraries (slices, http, json, file, string)
2. Cloud infrastructure (docker, k8s)
3. AI/LLM integration (claude, openai)

---

**Last Updated:** 2026-01-20
**For Questions:** Open an issue on GitHub
