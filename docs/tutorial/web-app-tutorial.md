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
- Handle **different request types** (GET, POST, DELETE) with `switch`/`when`
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
go mod init link-shortener
kukicha init    # Extracts stdlib for JSON, etc.
```

---

## Step 1: Your First Web Server

Let's start with the simplest possible web server:

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
    # Use pipe to send response!
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

## Step 2: Understanding Handlers

A **handler** is a function that responds to web requests. Every handler receives:

```kukicha
function myHandler(response http.ResponseWriter, request reference http.Request)
    # response - write your reply here
    # request - contains info about the incoming request
```

We can check what **method** (GET, POST, etc.) the user is using:

```kukicha
function myHandler(response http.ResponseWriter, request reference http.Request)
    if request.Method equals "GET"
        response |> fmt.Fprintln("You want to read something!")
    else if request.Method equals "POST"
        response |> fmt.Fprintln("You want to create something!")
    else
        response |> fmt.Fprintln("You used something else!")
```

---

## Step 3: Sending JSON Responses

Web APIs typically send data as **JSON** (JavaScript Object Notation). It looks like this:

```json
{"code": "k7f", "url": "https://go.dev", "clicks": 42}
```

> **Note: JSON in Kukicha with Go 1.25+**
>
> Kukicha uses Go 1.25+ `encoding/json/v2` for faster JSON:
> ```kukicha
> import "encoding/json/v2"
> json.MarshalWrite(response, data)            # Write JSON directly to response
> json.UnmarshalRead(request.Body, reference result)  # Read JSON from request
>
> # With pipe placeholder (_), pipe data into any argument position:
> data |> json.MarshalWrite(response, _)       # _ marks where piped value goes
> ```
>
> **Rule of thumb:** Use `json.NewEncoder()` / `json.NewDecoder()` for streaming (web servers), and `json.Marshal()` / `json.Unmarshal()` for in-memory conversion.

Let's define our `Link` type and send one as JSON:

```kukicha
import "encoding/json/v2"

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

    # Convert the link to JSON and send it using pipe
    response |> json.NewEncoder() |> .Encode(link) onerr return
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
    # Create an empty request to fill with the incoming data
    input := ShortenRequest{}

    # Parse the JSON from the request body — pipe it through decoder
    # "reference of" gets a pointer so the decoder can fill in our struct
    decodeErr := request.Body
        |> json.NewDecoder()
        |> json.Decode(_, reference of input)
    if decodeErr not equals empty
        response |> .WriteHeader(400)
        response |> fmt.Fprintln("Invalid JSON")
        return

    # Now 'input' contains the URL the user wants to shorten!
    print("Received URL: {input.url}")

    # We'll generate a short code and send it back (next step)
```

**What's `reference of`?**

When we write `reference of input`, we're giving the JSON decoder a way to **fill in** our struct variable. Without it, the decoder would only have a copy and couldn't modify our actual `input`.

> **💡 Tip: The `_` placeholder.** By default, the piped value becomes the first argument. Use `_` to place it elsewhere:
> ```kukicha
> # Default: piped value is first argument
> text |> string.ToLower()                      # → string.ToLower(text)
>
> # With _: piped value goes where _ is
> data |> json.MarshalWrite(response, _)        # → json.MarshalWrite(response, data)
> ```
> You'll see this pattern throughout this tutorial.

---

## Step 5: Building the Link Store

Let's create a type to hold our links. Instead of a list, we'll use a **map** — a key-value store where the key is the short code and the value is the `Link`. This gives us instant lookup by code:

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
import "fmt"
import "net/http"
import "encoding/json/v2"
import "strconv"
import "stdlib/string"

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
    shortUrl string json:"short_url"

type ErrorResponse
    err string json:"error"

# --- Store ---
# (In the Production tutorial, we'll replace this with a database)

type LinkStore
    links map of string to Link
    nextId int

# --- Helper Functions ---

function generateCode on store reference LinkStore() string
    store.nextId = store.nextId + 1
    return strconv.FormatInt(int64(store.nextId), 36)

function sendJSON on store reference LinkStore(response http.ResponseWriter, data any)
    response |> .Header() |> .Set("Content-Type", "application/json")
    response |> json.NewEncoder() |> .Encode(data) onerr return

function sendError on store reference LinkStore(response http.ResponseWriter, status int, message string)
    response |> .Header() |> .Set("Content-Type", "application/json")
    response |> .WriteHeader(status)
    response |> json.NewEncoder() |> .Encode(ErrorResponse{err: message}) onerr return
```

> **💡 Pro Tip:** In production code, use `stdlib/http` helpers instead of writing these manually:
> ```kukicha
> import "stdlib/http" as httphelper
> httphelper.JSON(response, link)              # Send JSON with correct headers
> httphelper.JSONError(response, 400, "...")   # Send error as JSON
> httphelper.ReadJSON(request, reference link) # Parse request body
> ```
> See the [Production Patterns Tutorial](production-patterns-tutorial.md) for more.

```kukicha
# --- API Handlers ---

# POST /shorten — Create a shortened link
function handleShorten on store reference LinkStore(response http.ResponseWriter, request reference http.Request)
    if request.Method not equals "POST"
        store.sendError(response, 405, "Method not allowed")
        return

    # Parse the incoming JSON
    input := ShortenRequest{}
    decodeErr := request.Body
        |> json.NewDecoder()
        |> json.Decode(_, reference of input)
    if decodeErr not equals empty
        store.sendError(response, 400, "Invalid JSON")
        return

    # Validate the URL
    if input.url equals ""
        store.sendError(response, 400, "URL is required")
        return

    if not (input.url |> string.HasPrefix("http://")) and not (input.url |> string.HasPrefix("https://"))
        store.sendError(response, 400, "URL must start with http:// or https://")
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

    response |> .WriteHeader(201)
    store.sendJSON(response, result)

# GET /r/{code} — Redirect to the original URL
# This is the core of a link shortener!
function handleRedirect on store reference LinkStore(response http.ResponseWriter, request reference http.Request)
    # Extract the code from the URL path: "/r/abc" → "abc"
    code := request.URL.Path |> string.TrimPrefix("/r/")
    if code equals "" or code equals request.URL.Path
        store.sendError(response, 400, "Missing link code")
        return

    # Look up the link
    link, exists := store.links[code]
    if not exists
        store.sendError(response, 404, "Link not found")
        return

    # Increment the click counter
    link.clicks = link.clicks + 1
    store.links[code] = link

    # Redirect! The browser will automatically follow this to the original URL
    http.Redirect(response, request, link.url, 301)

# GET /links — List all links
function handleListLinks on store reference LinkStore(response http.ResponseWriter, request reference http.Request)
    if request.Method not equals "GET"
        store.sendError(response, 405, "Method not allowed")
        return

    # Convert map values to a list for JSON output
    result := empty list of Link
    for _, link in store.links
        result = append(result, link)

    store.sendJSON(response, result)

# /links/{code} — Get info or delete a specific link
function handleLinkDetail on store reference LinkStore(response http.ResponseWriter, request reference http.Request)
    # Extract the code from the URL path
    code := request.URL.Path |> string.TrimPrefix("/links/")
    if code equals "" or code equals request.URL.Path
        store.sendError(response, 400, "Missing link code")
        return

    switch request.Method
        when "GET"
            link, exists := store.links[code]
            if not exists
                store.sendError(response, 404, "Link not found")
                return
            store.sendJSON(response, link)

        when "DELETE"
            _, exists := store.links[code]
            if not exists
                store.sendError(response, 404, "Link not found")
                return
            delete(store.links, code)
            response |> .WriteHeader(204)

        otherwise
            store.sendError(response, 405, "Method not allowed")

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
| **Pipe Operator** | Cleanly chain functions (like JSON encoders) with `|>` |
| **Method Values** | Pass `store.handleShorten` directly as an HTTP handler |
| **Handlers** | Functions that respond to web requests |
| **Request Methods** | GET (read), POST (create), DELETE (remove) — dispatched with `switch`/`when` |
| **JSON** | Data format for web APIs (`encoding/json/v2`) |
| **Status Codes** | Numbers that indicate success, failure, or redirect |
| **Maps** | Key-value storage with `map of string to Link` |
| **HTTP Redirects** | `http.Redirect()` sends browsers to another URL |

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

## Practice Exercises

Before moving on, try these enhancements:

1. **Custom codes** — Let users pick their own short code: `{"url":"...", "code":"my-link"}`
2. **Search** — Add `GET /links?search=github` to filter links by URL
3. **Stats endpoint** — `GET /stats` returns total links created and total clicks
4. **Duplicate detection** — If the same URL is submitted twice, return the existing short link

---

## What's Next?

You now have a working web service! But it's not production-ready yet. In the next tutorial, you'll learn:

- **[Production Patterns Tutorial](production-patterns-tutorial.md)** (Advanced)
  - Store links in a **database** (SQLite) so they persist
  - Generate **random codes** that aren't guessable
  - Handle **multiple users safely** with locking
  - Add proper **analytics**, **validation**, and **configuration**

---

**You've built a link shortener! 🔗**
