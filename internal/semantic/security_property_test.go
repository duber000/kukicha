package semantic

import (
	"strings"
	"testing"
)

// TestProperty_SecurityMonotonicity verifies that adding innocuous code to a
// program never removes an existing security error. This is a key property:
// if check(P) reports error E, then check(P + safe_code) must also report E.
func TestProperty_SecurityMonotonicity(t *testing.T) {
	cases := []struct {
		name   string
		source string
		errSub string
	}{
		{
			name: "sql-injection",
			source: `import "stdlib/pg"

func Bad(pool pg.Pool, id int)
    rows := pg.Query(pool, "SELECT * FROM users WHERE id = {id}") onerr return
    _ = rows
`,
			errSub: "SQL injection risk",
		},
		{
			name: "xss",
			source: `import "stdlib/http"

func Bad(w http.ResponseWriter, r reference http.Request)
    body := "user input"
    http.HTML(w, body)
`,
			errSub: "XSS risk",
		},
		{
			name: "ssrf",
			source: `import "stdlib/fetch"

func Bad(w http.ResponseWriter, r reference http.Request)
    data := fetch.Get("http://example.com") onerr return
    _ = data
`,
			errSub: "SSRF risk",
		},
		{
			name: "path-traversal",
			source: `import "stdlib/files"

func Bad(w http.ResponseWriter, r reference http.Request)
    data := files.Read("secret.txt") onerr return
    _ = data
`,
			errSub: "path traversal risk",
		},
		{
			name: "command-injection",
			source: `import "stdlib/shell"

func Bad()
    cmd := "ls -la"
    shell.Run(cmd)
`,
			errSub: "command injection risk",
		},
		{
			name: "open-redirect",
			source: `import "stdlib/http"

func Bad(w http.ResponseWriter, r reference http.Request)
    url := "/somewhere"
    http.Redirect(w, r, url)
`,
			errSub: "open redirect risk",
		},
	}

	// Innocuous code to append — should never remove an existing error
	appendCode := `

func SafeHelper() string
    return "safe"

func AnotherHelper(x int) int
    return x + 1
`

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Verify the base program has the expected error
			_, errs1 := analyzeSource(t, tc.source)
			if !containsError(errs1, tc.errSub) {
				t.Fatalf("base program should have error containing %q, got: %v", tc.errSub, errs1)
			}

			// Verify the base + extra code still has the error
			_, errs2 := analyzeSource(t, tc.source+appendCode)
			if !containsError(errs2, tc.errSub) {
				t.Fatalf("extended program lost error containing %q, got: %v", tc.errSub, errs2)
			}
		})
	}
}

// TestProperty_SecurityMonotonicity_ExtraFunction verifies monotonicity when
// an additional safe function is added alongside the vulnerable one.
func TestProperty_SecurityMonotonicity_ExtraFunction(t *testing.T) {
	source := `import "stdlib/pg"

func SafeHelper() string
    return "safe"

func Bad(pool pg.Pool, id int)
    rows := pg.Query(pool, "SELECT * FROM users WHERE id = {id}") onerr return
    _ = rows
`

	_, errs := analyzeSource(t, source)
	if !containsError(errs, "SQL injection risk") {
		t.Fatalf("program with extra safe function lost SQL injection error: %v", errs)
	}
}

func containsError(errs []error, substr string) bool {
	for _, e := range errs {
		if strings.Contains(e.Error(), substr) {
			return true
		}
	}
	return false
}
