# Build a Link Shortener with Kukicha

**Level:** Intermediate
**Time:** 30 minutes
**Prerequisite:** [CLI Explorer Tutorial](cli-explorer-tutorial.md)

Welcome! You've built interactive CLI tools with custom types, methods, and pipes. Now let's build something even cooler: a **web service** you can access from a browser — a link shortener, like bit.ly.

## What You'll Learn

In this tutorial, you'll discover how to:
- Create a **web server** that responds to requests
- Send and receive **JSON data** (the language of web APIs)
- Build **endpoints** for creating, reading, and deleting links
- Handle **different request types** (GET, POST, DELETE) in web handlers
- Perform **HTTP redirects** — the core of a link shortener

By the end, you'll have a working link shortener API that anyone can use!

---

## What We're Building

A **link shortener** takes long URLs and gives back short ones. When someone visits the short URL, they get redirected to the original.

Our API:

| Action | Request | URL | Description |
|--------|---------|-----|-------------|
| Shorten a URL | `POST` | `/shorten` | Submit a URL, get back a short code |
| Follow a link | `GET` | `/r/{code}` | Redirects to the original URL |
| List all links | `GET` | `/links` | See all shortened links |
| Get link info | `GET` | `/links/{code}` | Get details and click count |
| Delete a link | `DELETE` | `/links/{code}` | Remove a shortened link |

**Why a link shortener?** It's a real tool people actually use, it teaches all the core web concepts (routing, JSON, status codes, redirects), and you get to see HTTP redirects in action — something most tutorials skip.

Don't worry if this looks complicated — we'll build it step by step!

---

## Step 0: Project Setup

```bash
mkdir link-shortener && cd link-shortener
kukicha init    # go mod init + extracts stdlib for JSON, etc.
```

---

## Step 1: Your First Web Server

Let's start with the smallest possible web server so we can focus on HTTP flow before adding routing and JSON:

```kukicha
import "fmt"
import "net/http"

function main()
    # When someone visits the homepage, say hello
    http.HandleFunc("/", sayHello)

    print("Server starting on http://localhost:8080")
    http.ListenAndServe(":8080", empty) onerr panic "server failed to start"

# This function handles requests to "/"
function sayHello(response http.ResponseWriter, request reference http.Request)
    response |> fmt.Fprintln("Hello from Kukicha!")
```

**What's happening here?**

1. `http.HandleFunc("/", sayHello)` — When someone visits `/`, run the `sayHello` function
2. `http.ListenAndServe(":8080", empty)` — Start listening on port 8080
3. `sayHello` receives two things:
   - `response` — Where we write our reply
   - `request` — Information about what the user asked for

**Try it!**

Run the server:
```bash
kukicha run main.kuki
```

Then open your browser to `http://localhost:8080` — you should see "Hello from Kukicha!"

---

## Step 2: HTTP Handlers and Methods

A **handler** is a function that responds to web requests. Every handler receives:

```kukicha
function myHandler(response http.ResponseWriter, request reference http.Request)
    # response - write your reply here
    # request - contains info about the incoming request
```

Inside a handler, branch on request method to decide behavior:

```kukicha
function myHandler(response http.ResponseWriter, request reference http.Request)
    if request.Method equals "GET"
        response |> fmt.Fprintln("You want to read something!")
    else if request.Method equals "POST"
        response |> fmt.Fprintln("You want to create something!")
    else
        response |> fmt.Fprintln("You used something else!")
```

> **💡 Naming handler signatures.** If you find yourself repeating the same function signature, you can give it a name with a **function type alias**:
> ```kukicha
> type HandlerFunc func(http.ResponseWriter, reference http.Request)
> ```
> This is exactly how Go's `http.HandlerFunc` is defined under the hood. You'll see this pattern more in the [Production Patterns Tutorial](production-patterns-tutorial.md).

---

## Step 3: Sending JSON Responses

Web APIs typically send data as **JSON** (JavaScript Object Notation). It looks like this:

```json
{"code": "k7f", "url": "https://go.dev", "clicks": 42}
```

Kukicha's `stdlib/json` wraps Go's JSON encoder with pipe-friendly helpers. Let's define our `Link` type and send one as JSON:

```kukicha
import "stdlib/json"

type Link
    code string
    url string
    clicks int

function sendLink(response http.ResponseWriter, request reference http.Request)
    link := Link
        code: "k7f"
        url: "https://go.dev"
        clicks: 42

    # Tell the browser we're sending JSON using pipe chaining
    response |> .Header() |> .Set("Content-Type", "application/json")

    # Convert the link to JSON and write it to the response
    json.MarshalWrite(response, link) onerr return
```

**💡 Tip:** When piping into a method that belongs to the value itself, use the dot shorthand:
```kukicha
# Calling directly:
response.Header().Set(...)

# Same thing, using pipe:
response |> .Header() |> .Set(...)
```
This keeps the left-to-right data flow — and makes it clear the method belongs to the piped value, not an imported package.

When someone hits this endpoint, they'll receive:
```json
{"code":"k7f","url":"https://go.dev","clicks":42}
```

---

## Step 4: Receiving JSON Data

When someone wants to shorten a URL, they'll send us JSON. We need to read and parse it:

```kukicha
type ShortenRequest
    url string

function handleShorten(response http.ResponseWriter, request reference http.Request)
    # Parse the incoming JSON — UnmarshalRead reads from any io.Reader
    input := ShortenRequest{}
    json.UnmarshalRead(request.Body, reference of input) onerr
        response |> .WriteHeader(400)
        response |> fmt.Fprintln("Invalid JSON")
        return

    # Now 'input' contains the URL the user wants to shorten!
    print("Received URL: {input.url}")

    # We'll generate a short code and send it back (next step)
```

`json.UnmarshalRead` reads JSON from any `io.Reader` (like a request body) and decodes it into the target. The `onerr` block handles invalid JSON gracefully.

---

## Step 5: Building the Link Store

Let's create a type to hold our links. We'll use a map keyed by short code for constant-time lookups:

```kukicha
type LinkStore
    links map of string to Link    # code → Link
    nextId int
```

Wrapping state in a type keeps things organized — and as a bonus, we can pass our store to HTTP handlers using **method values** (we'll see that in the main function).

We also need a way to generate short codes. For now, we'll use a simple counter converted to base-36 (which uses letters and numbers):

```kukicha
import "strconv"

function generateCode on store reference LinkStore() string
    store.nextId = store.nextId + 1
    return strconv.FormatInt(int64(store.nextId), 36)
```

This gives codes like `"1"`, `"2"`, ..., `"a"`, `"b"`, ..., `"10"`, `"11"`. Short, URL-safe, and predictable. The production tutorial will add proper random codes.

---

## Step 6: The Complete Link Shortener

Now let's put it all together! Create `main.kuki`:

```kukicha
import "net/http"
import "strconv"
import "stdlib/json"
import "stdlib/string"
import "stdlib/maps"
import "stdlib/http" as httphelper

# --- Data Types ---

type Link
    code string
    url string
    clicks int

type ShortenRequest
    url string

type ShortenResponse
    code string
    url string
    shortUrl string as "short_url"

# --- Store ---
# (In the Production tutorial, we'll replace this with a database)

type LinkStore
    links map of string to Link
    nextId int

# --- Helper Functions ---

function generateCode on store reference LinkStore() string
    store.nextId = store.nextId + 1
    return strconv.FormatInt(int64(store.nextId), 36)
```

> **💡 Notice:** We don't need manual `sendJSON` or `sendError` helpers anymore. `stdlib/http` (imported as `httphelper`) provides `JSON()`, `JSONError()`, `ReadJSON()`, and more — with correct headers, status codes, and content types built in.

```kukicha
# --- API Handlers ---

# POST /shorten — Create a shortened link
function handleShorten on store reference LinkStore(response http.ResponseWriter, request reference http.Request)
    if request.Method not equals "POST"
        httphelper.MethodNotAllowed(response)
        return

    # Parse the incoming JSON — limit to 64 KB to prevent OOM from huge bodies
    input := ShortenRequest{}
    httphelper.ReadJSONLimit(request, 65536, reference of input) onerr
        httphelper.JSONBadRequest(response, "Invalid JSON")
        return

    # Validate the URL
    if input.url equals ""
        httphelper.JSONBadRequest(response, "URL is required")
        return

    if not (input.url |> string.HasPrefix("http://")) and not (input.url |> string.HasPrefix("https://"))
        httphelper.JSONBadRequest(response, "URL must start with http:// or https://")
        return

    # Generate a short code and store the link
    code := store.generateCode()
    link := Link{code: code, url: input.url, clicks: 0}
    store.links[code] = link

    # Send back the shortened link info
    result := ShortenResponse
        code: code
        url: input.url
        shortUrl: "http://localhost:8080/r/{code}"

    httphelper.JSONCreated(response, result)

# GET /r/{code} — Redirect to the original URL
# This is the core of a link shortener!
function handleRedirect on store reference LinkStore(response http.ResponseWriter, request reference http.Request)
    # Extract the code from the URL path: "/r/abc" → "abc"
    code := request.URL.Path |> string.TrimPrefix("/r/")
    if code equals "" or code equals request.URL.Path
        httphelper.JSONBadRequest(response, "Missing link code")
        return

    # Look up the link
    link, exists := store.links[code]
    if not exists
        httphelper.JSONNotFound(response, "Link not found")
        return

    # Increment the click counter
    link.clicks = link.clicks + 1
    store.links[code] = link

    # Redirect! Set the Location header and return 301 Moved Permanently.
    # Note: a link shortener intentionally redirects to arbitrary user-supplied URLs —
    # that's the whole point. We set the header directly rather than using
    # httphelper.RedirectPermanent because the compiler (correctly) warns about
    # non-literal redirect targets to help catch *accidental* open redirects.
    # We've already validated that link.url starts with http:// or https://.
    response.Header().Set("Location", link.url)
    response.WriteHeader(301)

# GET /links — List all links
function handleListLinks on store reference LinkStore(response http.ResponseWriter, request reference http.Request)
    if request.Method not equals "GET"
        httphelper.MethodNotAllowed(response)
        return

    # Convert map values to a list for JSON output
    result := empty list of Link
    for _, link in store.links
        result = append(result, link)

    httphelper.JSON(response, result)

# /links/{code} — Get info or delete a specific link
function handleLinkDetail on store reference LinkStore(response http.ResponseWriter, request reference http.Request)
    # Extract the code from the URL path
    code := request.URL.Path |> string.TrimPrefix("/links/")
    if code equals "" or code equals request.URL.Path
        httphelper.JSONBadRequest(response, "Missing link code")
        return

    switch request.Method
        when "GET"
            link, exists := store.links[code]
            if not exists
                httphelper.JSONNotFound(response, "Link not found")
                return
            httphelper.JSON(response, link)

        when "DELETE"
            if not maps.Contains(store.links, code)
                httphelper.JSONNotFound(response, "Link not found")
                return
            delete(store.links, code)
            httphelper.NoContent(response)

        otherwise
            httphelper.MethodNotAllowed(response)

# --- Main Entry Point ---

function main()
    store := LinkStore
        links: map of string to Link{}
        nextId: 0

    # Set up routes — method values let us pass methods as handler functions
    http.HandleFunc("/shorten", store.handleShorten)
    http.HandleFunc("/r/", store.handleRedirect)
    http.HandleFunc("/links", store.handleListLinks)
    http.HandleFunc("/links/", store.handleLinkDetail)

    print("=== Kukicha Link Shortener ===")
    print("Server running on http://localhost:8080")
    print("")
    print("Try these commands in another terminal:")
    print("  curl -X POST -d '{\"url\":\"https://go.dev\"}' http://localhost:8080/shorten")
    print("  curl -L http://localhost:8080/r/1")
    print("")

    http.ListenAndServe(":8080", empty) onerr panic "server failed to start"
```

---

## Step 7: Testing Your Link Shortener

Run your server:
```bash
kukicha run main.kuki
```

Now test it with `curl` in another terminal:

### Shorten some URLs:
```bash
curl -X POST -H "Content-Type: application/json" \
     -d '{"url":"https://go.dev"}' http://localhost:8080/shorten
# {"code":"1","url":"https://go.dev","short_url":"http://localhost:8080/r/1"}

curl -X POST -H "Content-Type: application/json" \
     -d '{"url":"https://github.com/golang/go"}' http://localhost:8080/shorten
# {"code":"2","url":"https://github.com/golang/go","short_url":"http://localhost:8080/r/2"}
```

### Follow a short link (the magic moment!):
```bash
curl -L http://localhost:8080/r/1
# Redirects to https://go.dev and shows the Go website HTML!

# Without -L, you can see the redirect itself:
curl -I http://localhost:8080/r/1
# HTTP/1.1 301 Moved Permanently
# Location: https://go.dev
```

### List all links:
```bash
curl http://localhost:8080/links
# [{"code":"1","url":"https://go.dev","clicks":2},
#  {"code":"2","url":"https://github.com/golang/go","clicks":0}]
```

### Get info about a specific link:
```bash
curl http://localhost:8080/links/1
# {"code":"1","url":"https://go.dev","clicks":2}
```

### Delete a link:
```bash
curl -X DELETE http://localhost:8080/links/2
# (empty - 204 No Content)
```

### Try an invalid URL:
```bash
curl -X POST -H "Content-Type: application/json" \
     -d '{"url":"not-a-url"}' http://localhost:8080/shorten
# {"error":"URL must start with http:// or https://"}
```

You've built a working link shortener!

---

## Understanding HTTP Status Codes

You may have noticed we use numbers like `200`, `201`, `301`, `404`. These are **status codes** that tell the client what happened:

| Code | Name | Meaning |
|------|------|---------|
| `200` | OK | Success! |
| `201` | Created | Successfully created something new |
| `204` | No Content | Success, but nothing to return |
| `301` | Moved Permanently | Redirect — go to this other URL instead |
| `400` | Bad Request | The client sent invalid data |
| `404` | Not Found | The requested item doesn't exist |
| `405` | Method Not Allowed | Wrong HTTP method for this endpoint |
| `500` | Internal Server Error | Something went wrong on the server |

The `301` is the workhorse of our link shortener — it tells browsers "this URL has moved, go here instead."

---

## What You've Learned

Congratulations! You've built a real web service. Let's review:

| Concept | What It Does |
|---------|--------------|
| **HTTP Server** | `http.ListenAndServe()` starts a web server |
| **Method Values** | Pass `store.handleShorten` directly as an HTTP handler |
| **Handlers** | Functions that respond to web requests |
| **`stdlib/json`** | Encode and decode JSON with `MarshalWrite`, `UnmarshalRead` |
| **`stdlib/http`** | Response helpers: `JSON()`, `JSONBadRequest()`, `ReadJSONLimit()`, `NoContent()` |
| **Status Codes** | Numbers that indicate success, failure, or redirect |
| **Map-backed Store** | Use `map of string to Link` for fast code lookup |
| **HTTP Redirects** | Set `Location` header + `WriteHeader(301)` — or use `httphelper.SafeRedirect` for known hosts |

---

## Current Limitations

Our link shortener works, but it has some limitations:

1. **Data disappears when you restart** — We're storing links in memory, not a database
2. **Not safe for multiple users** — Concurrent writes to `store.links` could race and corrupt data
3. **Predictable codes** — Sequential codes (`1`, `2`, `3`...) are guessable. Real shorteners use random codes
4. **No analytics** — Click counts don't persist across restarts
5. **No expiration** — Links live forever

We'll fix all of these in the next tutorial!

---

## Step 8: Server-Side Rendering (Optional)

Web APIs are great, but sometimes you want to serve HTML pages directly. Kukicha's `stdlib/template` makes this easy.

Let's add a simple homepage so users can shorten links from their browser.

### Importing the Template Package

Add `import "stdlib/template"` to your `main.kuki`.

### Creating the Handler

Add this handler to your `LinkStore` (or `Server`):

```kukicha
function handleHome on store reference LinkStore(response http.ResponseWriter, request reference http.Request)
    if request.URL.Path not equals "/"
        http.NotFound(response, request)
        return

    html := `
<!DOCTYPE html>
<html>
<head><title>Kukicha Shortener</title></head>
<body>
    <h1>Shorten Your Link</h1>
    <form action="/shorten" method="POST">
        <input type="text" name="url" placeholder="https://example.com" required>
        <button type="submit">Shorten</button>
    </form>
</body>
</html>
`
    # Parse the HTML string into a template, then write it to the response.
    # template.Parse() returns (Template, error). Since we're parsing a hardcoded
    # string that we know is valid, we use _ to discard the error. For templates
    # loaded from files, you'd want onerr instead.
    tmpl, _ := template.New("home") |> .Parse(html)
    tmpl |> .Execute(response, empty) onerr return
```

Then register it in `main()`:

```kukicha
http.HandleFunc("/", store.handleHome)
```

Now visiting `http://localhost:8080` shows a real HTML form!

**Why `stdlib/template`?**
- Safe against XSS attacks (auto-escaping via `html/template`)
- Powerful logic (`{{if .}}`, loops, etc.)
- Familiar syntax for Go developers (it wraps `html/template`)

> **Rendering dynamic data safely.** The example above uses a static HTML string with no user data, so `tmpl.Execute` is fine. If your template renders user-supplied values (names, URLs, etc.), use `template.HTMLExecute` or `template.HTMLRenderSimple` instead — they use `html/template` which auto-escapes `{{ }}` values:
> ```kukicha
> # One-shot render with auto-escaping (html/template)
> html := template.HTMLRenderSimple(tmplStr, map of string to any{"name": username}) onerr return
> httphelper.HTML(w, html)
> ```
> `template.Execute` uses `text/template` and performs **no** HTML escaping — never pass user input through it for HTML responses.

---

## Practice Exercises

Before moving on, try these enhancements:

1. **Custom codes** — Let users pick their own short code: `{"url":"...", "code":"my-link"}`
2. **Search** — Add `GET /links?search=github` to filter links by URL
3. **Stats endpoint** — `GET /stats` returns total links created and total clicks
4. **Duplicate detection** — If the same URL is submitted twice, return the existing short link

---

## What's Next?

You now have a working web service! But it's not production-ready yet. In the next tutorial, you'll learn:

### Tutorial Path

| # | Tutorial | What You'll Learn |
|---|----------|-------------------|
| 1 | **[Beginner Tutorial](beginner-tutorial.md)** | Variables, functions, strings, decisions, lists, loops, pipes |
| 2 | **[Data & AI Scripting](data-scripting-tutorial.md)** | Maps (Key-Value), parsing CSVs, shell commands, AI scripting |
| 3 | **[CLI Explorer](cli-explorer-tutorial.md)** | Custom types, methods, API data, arrow lambdas, error handling |
| 4 | ✅ **Link Shortener** | HTTP servers, JSON, REST APIs, redirects |
| 5 | **[Health Checker](concurrent-url-health-checker.md)** ← Next! | Concurrency (Goroutines, Channels), Interfaces |
| 6 | **[Production Patterns](production-patterns-tutorial.md)** | Databases, advanced patterns |

---

You've built a link shortener.
