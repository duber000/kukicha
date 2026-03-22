package semantic

import (
	"github.com/duber000/kukicha/internal/parser"
	"strings"
	"testing"
)

func TestSQLInterpolationDetection(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		expectErr string // substring expected in error, empty = no error expected
	}{
		{
			name: "interpolation in pg.Query is rejected",
			source: `import "stdlib/pg"

func Bad(pool pg.Pool, id int)
    rows := pg.Query(pool, "SELECT * FROM users WHERE id = {id}") onerr return
    _ = rows
`,
			expectErr: "SQL injection risk",
		},
		{
			name: "interpolation in pg.Exec is rejected",
			source: `import "stdlib/pg"

func Bad(pool pg.Pool, name string)
    pg.Exec(pool, "INSERT INTO users (name) VALUES ('{name}')") onerr return
`,
			expectErr: "SQL injection risk",
		},
		{
			name: "interpolation in pg.QueryRow is rejected",
			source: `import "stdlib/pg"

func Bad(pool pg.Pool, id int)
    row := pg.QueryRow(pool, "SELECT name FROM users WHERE id = {id}") onerr return
    _ = row
`,
			expectErr: "SQL injection risk",
		},
		{
			name: "interpolation in pg.TxExec is rejected",
			source: `import "stdlib/pg"

func Bad(tx pg.Tx, table string)
    pg.TxExec(tx, "DELETE FROM {table}") onerr return
`,
			expectErr: "SQL injection risk",
		},
		{
			name: "parameterized pg.Query is allowed",
			source: `import "stdlib/pg"

func Good(pool pg.Pool, id int)
    rows := pg.Query(pool, "SELECT * FROM users WHERE id = $1", id) onerr return
    _ = rows
`,
			expectErr: "",
		},
		{
			name: "plain string literal pg.Exec is allowed",
			source: `import "stdlib/pg"

func Good(pool pg.Pool)
    pg.Exec(pool, "DELETE FROM sessions WHERE expired = true") onerr return
`,
			expectErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p, err := parser.New(tt.source, "test.kuki")
			if err != nil {
				t.Fatalf("parser error: %v", err)
			}

			program, parseErrors := p.Parse()
			if len(parseErrors) > 0 {
				t.Fatalf("parse errors: %v", parseErrors)
			}

			analyzer := New(program)
			errors := analyzer.Analyze()

			if tt.expectErr == "" {
				// Expect no SQL injection errors (other semantic errors are OK)
				for _, e := range errors {
					if strings.Contains(e.Error(), "SQL injection") {
						t.Fatalf("unexpected SQL injection error: %v", e)
					}
				}
			} else {
				found := false
				for _, e := range errors {
					if strings.Contains(e.Error(), tt.expectErr) {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("expected error containing %q, got errors: %v", tt.expectErr, errors)
				}
			}
		})
	}
}

func TestHTMLNonLiteralDetection(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		expectErr string
	}{
		{
			name: "http.HTML with variable argument is rejected",
			source: `import "net/http"
import "stdlib/http" as httphelper

func Handle(w http.ResponseWriter, r reference http.Request)
    body := "<h1>hello</h1>"
    httphelper.HTML(w, body) onerr return
`,
			expectErr: "XSS risk",
		},
		{
			name: "http.HTML with literal argument is allowed",
			source: `import "net/http"
import "stdlib/http" as httphelper

func Handle(w http.ResponseWriter, r reference http.Request)
    httphelper.HTML(w, "<h1>Hello</h1>") onerr return
`,
			expectErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p, err := parser.New(tt.source, "test.kuki")
			if err != nil {
				t.Fatalf("parser error: %v", err)
			}

			program, parseErrors := p.Parse()
			if len(parseErrors) > 0 {
				t.Fatalf("parse errors: %v", parseErrors)
			}

			analyzer := New(program)
			errors := analyzer.Analyze()

			if tt.expectErr == "" {
				for _, e := range errors {
					if strings.Contains(e.Error(), "XSS risk") {
						t.Fatalf("unexpected XSS error: %v", e)
					}
				}
			} else {
				found := false
				for _, e := range errors {
					if strings.Contains(e.Error(), tt.expectErr) {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("expected error containing %q, got errors: %v", tt.expectErr, errors)
				}
			}
		})
	}
}

func TestFetchInHandlerDetection(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		expectErr string
	}{
		{
			name: "fetch.Get inside HTTP handler is rejected",
			source: `import "net/http"
import "stdlib/fetch"

func Handle(w http.ResponseWriter, r reference http.Request)
    url := "https://example.com"
    resp, err := fetch.Get(url)
    _ = resp
    _ = err
`,
			expectErr: "SSRF risk",
		},
		{
			name: "fetch.Post inside HTTP handler is rejected",
			source: `import "net/http"
import "stdlib/fetch"

func Handle(w http.ResponseWriter, r reference http.Request)
    resp, err := fetch.Post("data", "https://example.com")
    _ = resp
    _ = err
`,
			expectErr: "SSRF risk",
		},
		{
			name: "fetch.Get outside HTTP handler is allowed",
			source: `import "stdlib/fetch"

func FetchData(url string) (string, error)
    resp, err := fetch.Get(url)
    _ = resp
    return "", err
`,
			expectErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p, err := parser.New(tt.source, "test.kuki")
			if err != nil {
				t.Fatalf("parser error: %v", err)
			}

			program, parseErrors := p.Parse()
			if len(parseErrors) > 0 {
				t.Fatalf("parse errors: %v", parseErrors)
			}

			analyzer := New(program)
			errors := analyzer.Analyze()

			if tt.expectErr == "" {
				for _, e := range errors {
					if strings.Contains(e.Error(), "SSRF risk") {
						t.Fatalf("unexpected SSRF error: %v", e)
					}
				}
			} else {
				found := false
				for _, e := range errors {
					if strings.Contains(e.Error(), tt.expectErr) {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("expected error containing %q, got errors: %v", tt.expectErr, errors)
				}
			}
		})
	}
}

func TestFilesInHandlerDetection(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		expectErr string
	}{
		{
			name: "files.Read inside HTTP handler is rejected",
			source: `import "net/http"
import "stdlib/files"

func Handle(w http.ResponseWriter, r reference http.Request)
    path := r.URL.Path
    data, err := files.Read(path)
    _ = data
    _ = err
`,
			expectErr: "path traversal risk",
		},
		{
			name: "files.Write inside HTTP handler is rejected",
			source: `import "net/http"
import "stdlib/files"

func Handle(w http.ResponseWriter, r reference http.Request)
    files.Write("hello", "/tmp/out.txt") onerr return
`,
			expectErr: "path traversal risk",
		},
		{
			name: "files.Read outside HTTP handler is allowed",
			source: `import "stdlib/files"

func LoadConfig(path string) (list of byte, error)
    return files.Read(path)
`,
			expectErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p, err := parser.New(tt.source, "test.kuki")
			if err != nil {
				t.Fatalf("parser error: %v", err)
			}

			program, parseErrors := p.Parse()
			if len(parseErrors) > 0 {
				t.Fatalf("parse errors: %v", parseErrors)
			}

			analyzer := New(program)
			errors := analyzer.Analyze()

			if tt.expectErr == "" {
				for _, e := range errors {
					if strings.Contains(e.Error(), "path traversal risk") {
						t.Fatalf("unexpected path traversal error: %v", e)
					}
				}
			} else {
				found := false
				for _, e := range errors {
					if strings.Contains(e.Error(), tt.expectErr) {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("expected error containing %q, got errors: %v", tt.expectErr, errors)
				}
			}
		})
	}
}

func TestShellRunNonLiteralDetection(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		expectErr string
	}{
		{
			name: "shell.Run with variable argument is rejected",
			source: `import "stdlib/shell"

func RunCmd(cmd string) (string, error)
    return shell.Run(cmd)
`,
			expectErr: "command injection risk",
		},
		{
			name: "shell.Run with string literal is allowed",
			source: `import "stdlib/shell"

func RunStatus() (string, error)
    return shell.Run("git status")
`,
			expectErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p, err := parser.New(tt.source, "test.kuki")
			if err != nil {
				t.Fatalf("parser error: %v", err)
			}
			program, parseErrors := p.Parse()
			if len(parseErrors) > 0 {
				t.Fatalf("parse errors: %v", parseErrors)
			}
			analyzer := New(program)
			errors := analyzer.Analyze()
			if tt.expectErr == "" {
				for _, e := range errors {
					if strings.Contains(e.Error(), "command injection risk") {
						t.Fatalf("unexpected command injection error: %v", e)
					}
				}
			} else {
				found := false
				for _, e := range errors {
					if strings.Contains(e.Error(), tt.expectErr) {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("expected error containing %q, got errors: %v", tt.expectErr, errors)
				}
			}
		})
	}
}

func TestRedirectNonLiteralDetection(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		expectErr string
	}{
		{
			name: "httphelper.Redirect with variable URL is rejected",
			source: `import "net/http"
import "stdlib/http" as httphelper

func Handle(w http.ResponseWriter, r reference http.Request)
    returnURL := r.URL.Query().Get("return")
    httphelper.Redirect(w, r, returnURL)
`,
			expectErr: "open redirect risk",
		},
		{
			name: "httphelper.RedirectPermanent with variable URL is rejected",
			source: `import "net/http"
import "stdlib/http" as httphelper

func Handle(w http.ResponseWriter, r reference http.Request)
    target := r.URL.Query().Get("to")
    httphelper.RedirectPermanent(w, r, target)
`,
			expectErr: "open redirect risk",
		},
		{
			name: "httphelper.Redirect with literal URL is allowed",
			source: `import "net/http"
import "stdlib/http" as httphelper

func Handle(w http.ResponseWriter, r reference http.Request)
    httphelper.Redirect(w, r, "/dashboard")
`,
			expectErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p, err := parser.New(tt.source, "test.kuki")
			if err != nil {
				t.Fatalf("parser error: %v", err)
			}
			program, parseErrors := p.Parse()
			if len(parseErrors) > 0 {
				t.Fatalf("parse errors: %v", parseErrors)
			}
			analyzer := New(program)
			errors := analyzer.Analyze()
			if tt.expectErr == "" {
				for _, e := range errors {
					if strings.Contains(e.Error(), "open redirect risk") {
						t.Fatalf("unexpected open redirect error: %v", e)
					}
				}
			} else {
				found := false
				for _, e := range errors {
					if strings.Contains(e.Error(), tt.expectErr) {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("expected error containing %q, got errors: %v", tt.expectErr, errors)
				}
			}
		})
	}
}
