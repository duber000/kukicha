# DuckDB Sales Analytics Example

This example demonstrates using [DuckDB](https://duckdb.org/) with Kukicha for in-memory analytical queries via Go's `database/sql` interface.

## What it does

1. Opens an in-memory DuckDB database
2. Creates a sales table and inserts sample data
3. Runs four analytical queries showcasing DuckDB's strengths:
   - **Revenue by Product** - GROUP BY with aggregations
   - **Sales by Region** - COUNT, SUM, and sorting
   - **Monthly Trend** - DuckDB's STRFTIME date function
   - **Top Sale per Region** - DuckDB's QUALIFY clause with window functions

## Features demonstrated

- **Blank imports**: `import "github.com/marcboeker/go-duckdb" as _` for driver registration
- **`onerr` error handling**: Clean error propagation with `onerr panic`
- **String interpolation**: `"  {product}: {totalQty} units, ${revenue}"`
- **Pointer passing**: `rows.Scan(reference of product, reference of totalQty, reference of revenue)`
- **Variadic helper**: `func mustExec(db reference sql.DB, query string, many args)`

## Prerequisites

```bash
# Add the DuckDB Go driver dependency
go get github.com/marcboeker/go-duckdb
```

DuckDB requires CGO. Make sure you have a C compiler installed (`gcc` or `clang`).

## Running the example

```bash
kukicha run examples/duckdb/main.kuki
```

## Expected output

```
=== DuckDB Sales Analytics ===

--- Revenue by Product ---
  Widget: 295 units, $2948.05
  Gadget: 145 units, $3623.55
  Gizmo: 380 units, $1896.2

--- Sales by Region ---
  North: 3 sales, $2746.5
  South: 3 sales, $3021.45
  East: 2 sales, $2823.15

--- Monthly Revenue Trend ---
  2024-01: $2248.5
  2024-02: $1747.05
  2024-03: $3595.55

--- Top Sale per Region (Window Function) ---
  East: Widget ($1198.8)
  South: Gizmo ($898.2)
  North: Gizmo ($998.0)

Analytics complete!
```

## Key pattern: mustExec helper

The example uses a `mustExec` helper to avoid repeating error handling for INSERT/DDL statements:

```kukicha
func mustExec(db reference sql.DB, query string, many args) sql.Result
    result := db.Exec(query, many args) onerr panic "exec failed: {error}"
    return result
```

This lets you write clean one-liners for data setup:

```kukicha
mustExec(db, "CREATE TABLE sales (...)")
mustExec(db, ins, "Widget", "North", 100, 9.99, "2024-01-15")
```
