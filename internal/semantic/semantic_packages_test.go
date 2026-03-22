//go:build go1.25

package semantic

import (
	"github.com/duber000/kukicha/internal/parser"
	"strings"
	"testing"
	"testing/synctest"
)

// TestConcurrentSemanticAnalysis tests that the semantic analyzer is thread-safe
// and multiple analyzers can run concurrently using synctest
func TestConcurrentSemanticAnalysis(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// Test that semantic analyzer is thread-safe
		// Multiple analyzers should be able to run concurrently

		programs := []string{
			`func Add(a int, b int) int
    return a + b`,
			`type User
    name string
    age int`,
			`func Test() bool
    x := 5
    return x > 3`,
		}

		results := make(chan bool, len(programs))

		for _, src := range programs {
			go func(source string) {
				p, err := parser.New(source, "test.kuki")
				if err != nil {
					t.Errorf("parser error: %v", err)
					results <- false
					return
				}
				program, parseErrors := p.Parse()
				if len(parseErrors) > 0 {
					t.Errorf("parse errors: %v", parseErrors)
					results <- false
					return
				}
				analyzer := New(program)
				errors := analyzer.Analyze()
				if len(errors) > 0 {
					t.Errorf("semantic errors: %v", errors)
					results <- false
					return
				}
				results <- true
			}(src)
		}

		synctest.Wait()

		// Verify all completed successfully
		successCount := 0
		for range programs {
			select {
			case success := <-results:
				if success {
					successCount++
				}
			default:
				t.Error("Expected result not received")
			}
		}

		if successCount != len(programs) {
			t.Errorf("Expected %d successful analyses, got %d", len(programs), successCount)
		}
	})
}

func TestQualifiedTypes(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid qualified type in struct field",
			source: `
import "io"

type Writer
    output io.Writer
`,
			wantErr: false,
		},
		{
			name: "valid qualified type in function parameter",
			source: `
import "io"

func Write(w io.Writer, data string)
    return
`,
			wantErr: false,
		},
		{
			name: "valid qualified type in function return",
			source: `
import "io"

func GetWriter() io.Writer
    return empty
`,
			wantErr: false,
		},
		{
			name: "multiple qualified types",
			source: `
import "io"
import "bytes"

type Wrapper
    writer io.Writer
    reader io.Reader
    buffer bytes.Buffer
`,
			wantErr: false,
		},
		{
			name: "unimported package",
			source: `
type Writer
    output io.Writer
`,
			wantErr: true,
			errMsg:  "package 'io' not imported",
		},
		{
			name: "qualified type in list",
			source: `
import "io"

type Readers
    readers list of io.Reader
`,
			wantErr: false,
		},
		{
			name: "qualified type in map",
			source: `
import "io"

type WriterMap
    writers map of string to io.Writer
`,
			wantErr: false,
		},
		{
			name: "qualified type as pointer",
			source: `
import "bytes"

type BufferPtr
    buf reference bytes.Buffer
`,
			wantErr: false,
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

			if tt.wantErr {
				if len(errors) == 0 {
					t.Fatalf("expected error containing '%s', but got no errors", tt.errMsg)
				}
				found := false
				for _, err := range errors {
					if strings.Contains(err.Error(), tt.errMsg) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error containing '%s', got: %v", tt.errMsg, errors)
				}
			} else {
				if len(errors) > 0 {
					t.Errorf("expected no errors, got: %v", errors)
				}
			}
		})
	}
}

func TestVersionedPackageNameInference(t *testing.T) {
	tests := []struct {
		name        string
		importPath  string
		expectedPkg string
		source      string
	}{
		{
			name:        "slash-version suffix v2",
			importPath:  "encoding/json/v2",
			expectedPkg: "json",
			source: `import "encoding/json/v2"

type Config
    Name string

func main()
    cfg := Config{}
    cfg.Name = "test"
    data, _ := json.Marshal(cfg)
`,
		},
		{
			name:        "slash-version suffix v3",
			importPath:  "google.golang.org/protobuf/v3",
			expectedPkg: "protobuf",
			source: `import "google.golang.org/protobuf/v3"

func main()
    protobuf.NewMessage()
`,
		},
		{
			name:        "slash-version suffix v10",
			importPath:  "example.com/pkg/v10",
			expectedPkg: "pkg",
			source: `import "example.com/pkg/v10"

func main()
    pkg.DoSomething()
`,
		},
		{
			name:        "dot-version suffix (gopkg.in style)",
			importPath:  "gopkg.in/yaml.v3",
			expectedPkg: "yaml",
			source: `import "gopkg.in/yaml.v3"

type Data
    Value string

func main()
    d := Data{}
    yaml.Marshal(d)
`,
		},
		{
			name:        "no version suffix",
			importPath:  "encoding/json",
			expectedPkg: "json",
			source: `import "encoding/json"

type Data
    Value string

func main()
    d := Data{}
    json.Marshal(d)
`,
		},
		{
			name:        "package named vendor (not a version)",
			importPath:  "github.com/company/vendor",
			expectedPkg: "vendor",
			source: `import "github.com/company/vendor"

func main()
    vendor.DoSomething()
`,
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

			// We expect no errors because the package name should be inferred correctly
			// and the functions should be resolved
			if len(errors) > 0 {
				t.Errorf("expected no errors for package %s (inferred as %s), got: %v", tt.importPath, tt.expectedPkg, errors)
			}

			// Verify the package was added to the symbol table with the correct name
			pkgSymbol := analyzer.symbolTable.Resolve(tt.expectedPkg)
			if pkgSymbol == nil {
				t.Errorf("expected package %s to be in symbol table, but it wasn't found", tt.expectedPkg)
			}
			if pkgSymbol != nil && pkgSymbol.Kind != SymbolVariable {
				t.Errorf("expected symbol %s to be a variable (imports are stored as variables), got kind: %v", tt.expectedPkg, pkgSymbol.Kind)
			}
		})
	}
}

func TestExternalInterfaceTypeAsParameter(t *testing.T) {
	// http.Handler should be accepted as a parameter and return type
	// when net/http is imported
	input := `import "net/http"

func Wrap(handler http.Handler) http.Handler
    return handler
`

	analyzer, errors := analyzeSource(t, input)
	_ = analyzer

	// Should not produce any type-related errors
	for _, e := range errors {
		if strings.Contains(e.Error(), "undefined type") ||
			strings.Contains(e.Error(), "not imported") {
			t.Fatalf("unexpected type error for http.Handler: %v", e)
		}
	}
}
