package semantic

import (
	"strings"
	"testing"

	"github.com/duber000/kukicha/internal/parser"
)

// =============================================================================
// Security check edge cases — piped arguments, additional variants, and
// boundary conditions NOT covered by semantic_test.go.
// =============================================================================

// --- SQL injection: piped variant and Tx functions ---

func TestSQLInterpolation_PipedCallShiftsArgIndex(t *testing.T) {
	// When the pool is piped in, the SQL string is at index 0 (not 1).
	source := `import "stdlib/pg"

func Bad(pool pg.Pool, id int)
    rows := pool |> pg.Query("SELECT * FROM users WHERE id = {id}") onerr return
    _ = rows
`
	assertSecurityError(t, source, "SQL injection risk")
}

func TestSQLInterpolation_TxQueryRejected(t *testing.T) {
	source := `import "stdlib/pg"

func Bad(tx pg.Tx, id int)
    rows := pg.TxQuery(tx, "SELECT * FROM t WHERE id = {id}") onerr return
    _ = rows
`
	assertSecurityError(t, source, "SQL injection risk")
}

func TestSQLInterpolation_TxQueryRowRejected(t *testing.T) {
	source := `import "stdlib/pg"

func Bad(tx pg.Tx, id int)
    row := pg.TxQueryRow(tx, "SELECT name FROM t WHERE id = {id}") onerr return
    _ = row
`
	assertSecurityError(t, source, "SQL injection risk")
}

func TestSQLInterpolation_NonSQLFunctionIgnored(t *testing.T) {
	// Calling a non-pg function with interpolation must NOT flag SQL injection.
	source := `func Format(name string) string
    return "Hello {name}"
`
	assertNoSecurityError(t, source, "SQL injection")
}

func TestSQLInterpolation_NoArgs(t *testing.T) {
	// Edge case: pg.Query with too few arguments should not panic.
	source := `import "stdlib/pg"

func Bad(pool pg.Pool)
    rows := pg.Query(pool) onerr return
    _ = rows
`
	analyzeIgnoringNonSecurity(t, source, "SQL injection")
}

// --- XSS: piped variant ---

func TestHTMLNonLiteral_PipedResponseWriter(t *testing.T) {
	// When w is piped, content arg shifts to index 0.
	source := `import "net/http"
import "stdlib/http" as httphelper

func Handle(w http.ResponseWriter, r reference http.Request)
    body := "<h1>hello</h1>"
    w |> httphelper.HTML(body) onerr return
`
	assertSecurityError(t, source, "XSS risk")
}

func TestHTMLNonLiteral_PipedLiteralAllowed(t *testing.T) {
	source := `import "net/http"
import "stdlib/http" as httphelper

func Handle(w http.ResponseWriter, r reference http.Request)
    w |> httphelper.HTML("<h1>Safe</h1>") onerr return
`
	assertNoSecurityError(t, source, "XSS risk")
}

func TestHTMLNonLiteral_MissingContentArg(t *testing.T) {
	// Edge case: not enough args should not panic.
	source := `import "net/http"
import "stdlib/http" as httphelper

func Handle(w http.ResponseWriter, r reference http.Request)
    httphelper.HTML(w) onerr return
`
	analyzeIgnoringNonSecurity(t, source, "XSS risk")
}

// --- SSRF: fetch.New variant ---

func TestFetchInHandler_FetchNew(t *testing.T) {
	source := `import "net/http"
import "stdlib/fetch"

func Handle(w http.ResponseWriter, r reference http.Request)
    client := fetch.New()
    _ = client
`
	assertSecurityError(t, source, "SSRF risk")
}

func TestFetchInHandler_NotAHandler(t *testing.T) {
	// A function without http.ResponseWriter is NOT a handler.
	source := `import "stdlib/fetch"

func Background()
    resp, err := fetch.Get("https://example.com")
    _ = resp
    _ = err
`
	assertNoSecurityError(t, source, "SSRF risk")
}

// --- Path traversal: additional file operations ---

func TestFilesInHandler_Delete(t *testing.T) {
	source := `import "net/http"
import "stdlib/files"

func Handle(w http.ResponseWriter, r reference http.Request)
    files.Delete("/tmp/file") onerr return
`
	assertSecurityError(t, source, "path traversal risk")
}

func TestFilesInHandler_DeleteAll(t *testing.T) {
	source := `import "net/http"
import "stdlib/files"

func Handle(w http.ResponseWriter, r reference http.Request)
    files.DeleteAll("/tmp/dir") onerr return
`
	assertSecurityError(t, source, "path traversal risk")
}

func TestFilesInHandler_List(t *testing.T) {
	source := `import "net/http"
import "stdlib/files"

func Handle(w http.ResponseWriter, r reference http.Request)
    entries, err := files.List("/tmp")
    _ = entries
    _ = err
`
	assertSecurityError(t, source, "path traversal risk")
}

func TestFilesInHandler_ListRecursive(t *testing.T) {
	source := `import "net/http"
import "stdlib/files"

func Handle(w http.ResponseWriter, r reference http.Request)
    entries, err := files.ListRecursive("/tmp")
    _ = entries
    _ = err
`
	assertSecurityError(t, source, "path traversal risk")
}

func TestFilesInHandler_Append(t *testing.T) {
	source := `import "net/http"
import "stdlib/files"

func Handle(w http.ResponseWriter, r reference http.Request)
    files.Append("data", "/tmp/file") onerr return
`
	assertSecurityError(t, source, "path traversal risk")
}

func TestFilesInHandler_AppendString(t *testing.T) {
	source := `import "net/http"
import "stdlib/files"

func Handle(w http.ResponseWriter, r reference http.Request)
    files.AppendString("data", "/tmp/file") onerr return
`
	assertSecurityError(t, source, "path traversal risk")
}

func TestFilesInHandler_ReadBytes(t *testing.T) {
	source := `import "net/http"
import "stdlib/files"

func Handle(w http.ResponseWriter, r reference http.Request)
    data, err := files.ReadBytes("/tmp/file")
    _ = data
    _ = err
`
	assertSecurityError(t, source, "path traversal risk")
}

func TestFilesInHandler_WriteString(t *testing.T) {
	source := `import "net/http"
import "stdlib/files"

func Handle(w http.ResponseWriter, r reference http.Request)
    files.WriteString("data", "/tmp/file") onerr return
`
	assertSecurityError(t, source, "path traversal risk")
}

func TestFilesOutsideHandler_AllAllowed(t *testing.T) {
	source := `import "stdlib/files"

func Cleanup()
    files.Delete("/tmp/old") onerr return
    files.DeleteAll("/tmp/dir") onerr return
`
	assertNoSecurityError(t, source, "path traversal risk")
}

// --- Command injection: piped arg edge case ---

func TestShellRun_PipedArgSkipped(t *testing.T) {
	// When argument is piped, we can't verify origin — should NOT error.
	source := `import "stdlib/shell"

func Run(cmd string) (string, error)
    return cmd |> shell.Run()
`
	assertNoSecurityError(t, source, "command injection risk")
}

func TestShellRun_NoArgs(t *testing.T) {
	// Edge case: shell.Run with no arguments should not panic.
	source := `import "stdlib/shell"

func Run() (string, error)
    return shell.Run()
`
	analyzeIgnoringNonSecurity(t, source, "command injection")
}

// --- Open redirect: piped variant ---

func TestRedirectNonLiteral_PipedResponseWriter(t *testing.T) {
	// When w is piped, URL index shifts from 2 to 1.
	source := `import "net/http"
import "stdlib/http" as httphelper

func Handle(w http.ResponseWriter, r reference http.Request)
    target := r.URL.Query().Get("to")
    w |> httphelper.Redirect(r, target)
`
	assertSecurityError(t, source, "open redirect risk")
}

func TestRedirectNonLiteral_PipedLiteralAllowed(t *testing.T) {
	source := `import "net/http"
import "stdlib/http" as httphelper

func Handle(w http.ResponseWriter, r reference http.Request)
    w |> httphelper.Redirect(r, "/safe")
`
	assertNoSecurityError(t, source, "open redirect risk")
}

func TestRedirectNonLiteral_StdlibExempt(t *testing.T) {
	// Stdlib source files are exempt from redirect checks.
	source := `import "net/http"
import "stdlib/http" as httphelper

func Handle(w http.ResponseWriter, r reference http.Request)
    target := "/dashboard"
    httphelper.Redirect(w, r, target)
`
	p, err := parser.New(source, "stdlib/http/redirect.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}
	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	analyzer := NewWithFile(program, "stdlib/http/redirect.kuki")
	errors := analyzer.Analyze()

	for _, e := range errors {
		if strings.Contains(e.Error(), "open redirect risk") {
			t.Fatalf("stdlib files should be exempt from redirect check, got: %v", e)
		}
	}
}

func TestRedirectNonLiteral_TooFewArgs(t *testing.T) {
	// Edge case: not enough arguments should not panic.
	source := `import "net/http"
import "stdlib/http" as httphelper

func Handle(w http.ResponseWriter, r reference http.Request)
    httphelper.Redirect(w, r)
`
	analyzeIgnoringNonSecurity(t, source, "open redirect risk")
}

// --- isInHTTPHandler edge cases ---

func TestIsInHTTPHandler_NoFunc(t *testing.T) {
	// A top-level type decl should not trigger handler detection.
	source := `type Config
    Port int
`
	assertNoSecurityError(t, source, "risk")
}

// =============================================================================
// Test helpers
// =============================================================================

// assertSecurityError parses source and asserts an error containing substr.
func assertSecurityError(t *testing.T, source string, substr string) {
	t.Helper()
	p, err := parser.New(source, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}
	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}
	analyzer := New(program)
	errors := analyzer.Analyze()

	for _, e := range errors {
		if strings.Contains(e.Error(), substr) {
			return
		}
	}
	t.Fatalf("expected error containing %q, got: %v", substr, errors)
}

// assertNoSecurityError parses source and asserts NO error containing substr.
func assertNoSecurityError(t *testing.T, source string, substr string) {
	t.Helper()
	p, err := parser.New(source, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}
	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}
	analyzer := New(program)
	errors := analyzer.Analyze()

	for _, e := range errors {
		if strings.Contains(e.Error(), substr) {
			t.Fatalf("unexpected error containing %q: %v", substr, e)
		}
	}
}

// analyzeIgnoringNonSecurity parses source and ensures it doesn't panic,
// ignoring all errors except those containing the security substr.
func analyzeIgnoringNonSecurity(t *testing.T, source string, securitySubstr string) {
	t.Helper()
	p, err := parser.New(source, "test.kuki")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}
	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}
	analyzer := New(program)
	errors := analyzer.Analyze()

	for _, e := range errors {
		if strings.Contains(e.Error(), securitySubstr) {
			t.Fatalf("unexpected security error: %v", e)
		}
	}
}
