# Fetch GitHub Repos Example

This example demonstrates using Kukicha's `fetch` and `files` standard library packages to:

1. Fetch data from a REST API (GitHub)
2. Parse JSON responses
3. Filter and transform data using pipes
4. Save results to a file

## What it does

The example fetches repositories from GitHub's API for the `golang` user, filters for repos with more than 100 stars, and saves the results to `repos.json`.

## Running the example

```bash
# From the example directory
kuki run main.kuki
```

## Pipeline breakdown

```kukicha
# Don't forget to import the json package
import "stdlib/json"

repos := fetch.Get("https://api.github.com/users/golang/repos")
    |> fetch.CheckStatus()                    # Verify HTTP 200
    |> fetch.Bytes()                          # Get response bytes
    |> json.Unmarshal(_, reference repos)     # Parse JSON into repos
    |> slice.Filter(r -> r.Stars > 100)      # Filter popular repos
    |> files.Write("repos.json")              # Save to file
    onerr panic "failed"                      # Handle errors
```

This showcases Kukicha's pipe-first design where data flows naturally through transformations.

## Output

The program creates a `repos.json` file containing an array of repository objects with fields like:
- Name
- Description
- Stars count
- Language
- URL

## Advanced example

The file also includes an `advancedExample()` function that demonstrates:
- Filtering out archived repositories
- Mapping to a simplified RepoSummary type
- Combining multiple transformations in a single pipeline

This pattern is perfect for quick automation scripts and data processing tasks.
