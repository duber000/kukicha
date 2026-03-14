// genstdlibregistry generates internal/semantic/stdlib_registry_gen.go by
// scanning all stdlib/*.kuki source files and extracting exported function
// signatures: return counts, per-position return types, and parameter names.
// Run via "make genstdlibregistry" or "go run ./cmd/genstdlibregistry".
//
// This implements the "Automatic Inference" improvement described in COMPILER-FIX.md:
// instead of maintaining a hand-written registry in semantic.go, the registry is
// derived directly from the .kuki source files so it stays in sync automatically.
package main

import (
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/duber000/kukicha/internal/ast"
	"github.com/duber000/kukicha/internal/parser"
)

func main() {
	// Expect to be run from the repo root (where stdlib/ lives).
	matches, err := filepath.Glob("stdlib/*/*.kuki")
	if err != nil {
		fmt.Fprintf(os.Stderr, "glob error: %v\n", err)
		os.Exit(1)
	}

	result, scanErrs := scanRegistry(matches)
	if len(scanErrs) > 0 {
		for _, e := range scanErrs {
			fmt.Fprintln(os.Stderr, e)
		}
		fmt.Fprintln(os.Stderr, "aborting: stdlib scan had errors; registry not updated")
		os.Exit(1)
	}

	formatted := formatRegistry(result)

	outPath := filepath.Join("internal", "semantic", "stdlib_registry_gen.go")
	if err := os.WriteFile(outPath, formatted, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write %s: %v\n", outPath, err)
		os.Exit(1)
	}

	fmt.Printf("Generated %s with %d entries (%d deprecated).\n", outPath, len(result.registry), len(result.deprecated))
}

// registryEntry holds the full signature info for a stdlib function.
type registryEntry struct {
	count      int
	types      []typeRepr   // per-position return types
	paramNames []string     // parameter names (for named argument support)
}

// typeRepr is the generator's representation of a type, emitted as goStdlibType.
type typeRepr struct {
	kind string // TypeKind constant name (e.g., "TypeKindInt")
	name string // For TypeKindNamed (e.g., "error", "time.Time")
}

// scanResult holds the scanned registry and deprecated map.
type scanResult struct {
	registry   map[string]registryEntry
	deprecated map[string]string // qualified name → deprecation message
}

// scanRegistry reads and parses all .kuki files in paths, returning a map of
// qualified function name → signature info plus a map of deprecated functions.
// Skips _test.kuki files, unexported functions, methods, and void functions.
// Returns accumulated errors.
func scanRegistry(paths []string) (scanResult, []error) {
	result := scanResult{
		registry:   map[string]registryEntry{},
		deprecated: map[string]string{},
	}
	var errs []error

	for _, path := range paths {
		base := filepath.Base(path)
		if strings.HasSuffix(base, "_test.kuki") {
			continue
		}

		src, err := os.ReadFile(path)
		if err != nil {
			errs = append(errs, fmt.Errorf("read %s: %v", path, err))
			continue
		}

		p, err := parser.New(string(src), path)
		if err != nil {
			errs = append(errs, fmt.Errorf("lex %s: %v", path, err))
			continue
		}

		prog, parseErrs := p.Parse()
		if len(parseErrs) > 0 {
			for _, e := range parseErrs {
				errs = append(errs, fmt.Errorf("parse error %s: %v", path, e))
			}
		}
		if prog == nil || prog.PetioleDecl == nil {
			continue
		}

		pkgName := prog.PetioleDecl.Name.Value

		for _, decl := range prog.Declarations {
			fd, ok := decl.(*ast.FunctionDecl)
			if !ok {
				continue
			}

			// Skip unexported functions (start with lowercase).
			name := fd.Name.Value
			if len(name) == 0 || name[0] < 'A' || name[0] > 'Z' {
				continue
			}

			// Skip methods — they have a receiver and are called as value.Method(),
			// not as pkg.Method(), so they don't belong in this registry.
			if fd.Receiver != nil {
				continue
			}

			key := pkgName + "." + name

			// Check for # kuki:deprecated directive
			for _, dir := range fd.Directives {
				if dir.Name == "deprecated" {
					msg := "deprecated"
					if len(dir.Args) > 0 {
						msg = dir.Args[0]
					}
					result.deprecated[key] = msg
					break
				}
			}

			returnCount := len(fd.Returns)
			if returnCount == 0 {
				// Void functions don't need an entry; the codegen handles them fine.
				continue
			}

			// Extract per-position return types
			types := make([]typeRepr, returnCount)
			for i, ret := range fd.Returns {
				types[i] = typeAnnotationToRepr(ret)
			}

			// Extract parameter names
			paramNames := make([]string, len(fd.Parameters))
			for i, param := range fd.Parameters {
				paramNames[i] = param.Name.Value
			}

			if existing, exists := result.registry[key]; !exists || returnCount > existing.count {
				result.registry[key] = registryEntry{
					count:      returnCount,
					types:      types,
					paramNames: paramNames,
				}
			}
		}
	}

	return result, errs
}

// typeAnnotationToRepr converts a Kukicha AST type annotation to a typeRepr.
func typeAnnotationToRepr(ann ast.TypeAnnotation) typeRepr {
	switch t := ann.(type) {
	case *ast.PrimitiveType:
		switch t.Name {
		case "int", "int8", "int16", "int32", "int64",
			"uint", "uint8", "uint16", "uint32", "uint64",
			"byte", "rune":
			return typeRepr{kind: "TypeKindInt"}
		case "float32", "float64":
			return typeRepr{kind: "TypeKindFloat"}
		case "string":
			return typeRepr{kind: "TypeKindString"}
		case "bool":
			return typeRepr{kind: "TypeKindBool"}
		default:
			return typeRepr{kind: "TypeKindUnknown"}
		}
	case *ast.NamedType:
		return typeRepr{kind: "TypeKindNamed", name: t.Name}
	case *ast.ListType:
		return typeRepr{kind: "TypeKindList"}
	case *ast.MapType:
		return typeRepr{kind: "TypeKindMap"}
	case *ast.ChannelType:
		return typeRepr{kind: "TypeKindChannel"}
	case *ast.ReferenceType:
		return typeRepr{kind: "TypeKindReference"}
	case *ast.FunctionType:
		return typeRepr{kind: "TypeKindFunction"}
	default:
		return typeRepr{kind: "TypeKindUnknown"}
	}
}

// formatTypeRepr formats a typeRepr as a Go source literal for goStdlibType.
func formatTypeRepr(tr typeRepr) string {
	if tr.name != "" {
		return fmt.Sprintf("{Kind: %s, Name: %q}", tr.kind, tr.name)
	}
	return fmt.Sprintf("{Kind: %s}", tr.kind)
}

// formatRegistry generates the Go source code for stdlib_registry_gen.go from
// the scan result. Returns gofmt'd source.
func formatRegistry(result scanResult) []byte {
	entries := make([]string, 0, len(result.registry))
	for k, v := range result.registry {
		// Build the types slice literal
		typeParts := make([]string, len(v.types))
		for i, tr := range v.types {
			typeParts[i] = formatTypeRepr(tr)
		}
		typesLiteral := strings.Join(typeParts, ", ")

		// Build the param names slice literal
		paramParts := make([]string, len(v.paramNames))
		for i, name := range v.paramNames {
			paramParts[i] = fmt.Sprintf("%q", name)
		}
		paramsLiteral := strings.Join(paramParts, ", ")

		entries = append(entries, fmt.Sprintf(
			"\t%q: {Count: %d, Types: []goStdlibType{%s}, ParamNames: []string{%s}},",
			k, v.count, typesLiteral, paramsLiteral,
		))
	}
	sort.Strings(entries)

	depEntries := make([]string, 0, len(result.deprecated))
	for k, v := range result.deprecated {
		depEntries = append(depEntries, fmt.Sprintf("\t%q: %q,", k, v))
	}
	sort.Strings(depEntries)

	src := fmt.Sprintf(`// Code generated by cmd/genstdlibregistry; DO NOT EDIT.
// Run "make genstdlibregistry" to regenerate after changing stdlib/*.kuki files.

package semantic

// generatedStdlibRegistry maps qualified Kukicha stdlib function names (pkg.Func)
// to their return signature (count, per-position types, and parameter names).
// The semantic analyzer uses this to correctly decompose pipe expressions and
// onerr clauses, provide type info for imported functions, and support named arguments.
//
// Generated from: stdlib/*/*.kuki (excludes *_test.kuki, methods, unexported funcs,
// and void functions).
var generatedStdlibRegistry = map[string]goStdlibEntry{
%s
}

// generatedStdlibDeprecated maps qualified Kukicha stdlib function names to their
// deprecation messages. Populated from # kuki:deprecated directives in stdlib .kuki files.
var generatedStdlibDeprecated = map[string]string{
%s
}
`, strings.Join(entries, "\n"), strings.Join(depEntries, "\n"))

	formatted, fmtErr := format.Source([]byte(src))
	if fmtErr != nil {
		// Fall back to unformatted — still valid Go.
		return []byte(src)
	}
	return formatted
}
