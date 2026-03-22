// gengostdlib generates internal/semantic/go_stdlib_gen.go by inspecting
// Go standard library function signatures via go/importer. This replaces
// the hand-maintained knownExternalReturns entries in semantic_calls.go.
//
// Run via "make gengostdlib" or "go run ./cmd/gengostdlib".
package main

import (
	"fmt"
	"go/format"
	"go/importer"
	"go/types"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// funcSpec defines which functions to extract from each Go stdlib package.
// An empty list means "extract all exported functions".
type funcSpec struct {
	pkg   string
	funcs []string // specific functions; empty = all exported
}

// knownNamedTypes lists qualified type names that should be preserved as
// TypeKindNamed in the generated registry. These are types with hand-coded
// method entries in semantic_calls.go (e.g., time.Time has method resolution
// for .Unix(), .Before(), etc.).
var knownNamedTypes = map[string]bool{
	"time.Time": true,
}

// packages lists Go stdlib packages and the specific functions we need.
// Only functions actually used in Kukicha programs or stdlib wrappers
// need to be here — this is intentionally curated, not exhaustive.
var packages = []funcSpec{
	{pkg: "os", funcs: []string{
		"ReadFile", "ReadDir", "Create", "Open", "OpenFile",
		"WriteFile", "Remove", "RemoveAll", "Rename",
		"Mkdir", "MkdirAll", "Chmod", "Chown",
		"Setenv", "Unsetenv", "LookupEnv",
		"Stat", "Getwd", "UserHomeDir", "Executable",
	}},
	{pkg: "strconv", funcs: []string{
		"Atoi", "ParseInt", "ParseFloat", "ParseBool",
		"Itoa", "FormatInt", "FormatFloat", "FormatBool",
	}},
	{pkg: "fmt", funcs: []string{
		"Sprintf", "Fprintf", "Printf", "Errorf",
		"Sscanf", "Fscanf",
	}},
	{pkg: "io", funcs: []string{
		"Copy", "ReadAll", "ReadFull", "WriteString",
	}},
	{pkg: "net", funcs: []string{
		"ParseCIDR", "Dial", "Listen",
	}},
	{pkg: "net/url", funcs: []string{
		"Parse", "ParseRequestURI",
	}},
	{pkg: "net/http", funcs: []string{
		"Get", "Post", "NewRequest",
	}},
	{pkg: "encoding/json", funcs: []string{
		"Marshal", "Unmarshal", "MarshalIndent",
	}},
	{pkg: "path/filepath", funcs: []string{
		"Abs", "Glob", "Rel", "EvalSymlinks",
	}},
	{pkg: "regexp", funcs: []string{
		"Compile", "MustCompile", "MatchString",
	}},
	{pkg: "time", funcs: []string{
		"Now", "Since", "Parse", "ParseDuration", "LoadLocation",
	}},
	{pkg: "sync", funcs: []string{}}, // no top-level funcs we need
	{pkg: "context", funcs: []string{
		"WithCancel", "WithTimeout", "WithDeadline",
	}},
	{pkg: "strings", funcs: []string{
		"NewReader", "NewReplacer",
	}},
	{pkg: "bytes", funcs: []string{
		"NewBuffer", "NewReader",
	}},
	{pkg: "bufio", funcs: []string{
		"NewScanner", "NewReader", "NewWriter",
	}},
	{pkg: "sort", funcs: []string{
		"Search",
	}},
	{pkg: "math", funcs: []string{
		"Abs", "Ceil", "Floor", "Round", "Max", "Min", "Pow", "Sqrt", "Log",
	}},
	{pkg: "os/exec", funcs: []string{
		"Command", "CommandContext",
	}},
}

func main() {
	imp := importer.Default()
	entries, errs := extractSignatures(imp, packages)
	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintln(os.Stderr, e)
		}
		// Warnings, not fatal — some packages might not be available
	}

	ifaces := extractInterfaces(imp, packages)

	src := formatOutput(entries, ifaces)
	outPath := filepath.Join("internal", "semantic", "go_stdlib_gen.go")
	if err := os.WriteFile(outPath, src, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write %s: %v\n", outPath, err)
		os.Exit(1)
	}
	fmt.Printf("Generated %s with %d entries, %d interfaces.\n", outPath, len(entries), len(ifaces))
}

// entry represents a single Go stdlib function's return signature.
type entry struct {
	QualifiedName string     // e.g. "os.ReadFile"
	ReturnCount   int        // number of return values
	Returns       []typeRepr // type info per return position
}

// typeRepr is a serializable representation of a TypeInfo.
type typeRepr struct {
	Kind string // TypeKind constant name: "TypeKindString", "TypeKindInt", etc.
	Name string // For TypeKindNamed: qualified name (e.g. "error", "os.FileInfo")
}

func extractSignatures(imp types.Importer, specs []funcSpec) ([]entry, []error) {
	var entries []entry
	var errs []error

	for _, spec := range specs {
		pkg, err := imp.Import(spec.pkg)
		if err != nil {
			errs = append(errs, fmt.Errorf("import %s: %v", spec.pkg, err))
			continue
		}

		// Build the set of functions to extract
		wantAll := len(spec.funcs) == 0
		wanted := make(map[string]bool, len(spec.funcs))
		for _, f := range spec.funcs {
			wanted[f] = true
		}

		scope := pkg.Scope()
		for _, name := range scope.Names() {
			obj := scope.Lookup(name)

			if fn, ok := obj.(*types.Func); ok {
				if !wantAll && !wanted[name] {
					continue
				}

				sig := fn.Type().(*types.Signature)
				results := sig.Results()
				if results.Len() == 0 {
					continue // skip void functions
				}

				// Build qualified name using the alias Kukicha uses
				qualName := kukichaAlias(spec.pkg) + "." + name

				e := entry{
					QualifiedName: qualName,
					ReturnCount:   results.Len(),
				}
				for v := range results.Variables() {
					e.Returns = append(e.Returns, goTypeToRepr(v.Type()))
				}
				entries = append(entries, e)
			} else if typeName, ok := obj.(*types.TypeName); ok {
				if !typeName.Exported() {
					continue
				}
				named, ok := typeName.Type().(*types.Named)
				if !ok {
					continue
				}

				// Extract all exported methods for this type
				// Use MethodSet on pointer type to get both pointer and value methods
				ptr := types.NewPointer(named)
				mset := types.NewMethodSet(ptr)
				for method := range mset.Methods() {
					m := method.Obj().(*types.Func)
					if !m.Exported() {
						continue
					}
					sig := m.Type().(*types.Signature)
					results := sig.Results()
					if results.Len() == 0 {
						continue // skip void methods
					}

					// Register under both value type and pointer type
					// This matches how analyzeMethodCallExpr handles objType.Name
					pkgAlias := kukichaAlias(spec.pkg)
					valName := pkgAlias + "." + named.Obj().Name() + "." + m.Name()
					ptrName := "*" + valName

					eVal := entry{
						QualifiedName: valName,
						ReturnCount:   results.Len(),
					}
					ePtr := entry{
						QualifiedName: ptrName,
						ReturnCount:   results.Len(),
					}

					for v := range results.Variables() {
						repr := goTypeToRepr(v.Type())
						eVal.Returns = append(eVal.Returns, repr)
						ePtr.Returns = append(ePtr.Returns, repr)
					}

					entries = append(entries, eVal, ePtr)
				}
			}
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].QualifiedName < entries[j].QualifiedName
	})
	return entries, errs
}

// kukichaAlias maps full Go import paths to the short name Kukicha uses.
func kukichaAlias(pkg string) string {
	switch pkg {
	case "net/url":
		return "url"
	case "net/http":
		return "http"
	case "encoding/json":
		return "json"
	case "path/filepath":
		return "filepath"
	default:
		// Last segment: "os" → "os", "strconv" → "strconv", etc.
		parts := strings.Split(pkg, "/")
		return parts[len(parts)-1]
	}
}

// goTypeToRepr converts a Go type to our serializable type representation.
func goTypeToRepr(t types.Type) typeRepr {
	// Preserve type identity for specific named struct types that have
	// hand-coded method entries in semantic_calls.go (e.g., time.Time).
	// Other named types resolve to their underlying kind to avoid
	// cascading type compatibility issues.
	if named, ok := t.(*types.Named); ok && named.Obj().Pkg() != nil {
		qualName := kukichaAlias(named.Obj().Pkg().Path()) + "." + named.Obj().Name()
		if knownNamedTypes[qualName] {
			return typeRepr{Kind: "TypeKindNamed", Name: qualName}
		}
	}
	t = t.Underlying()

	switch u := t.(type) {
	case *types.Basic:
		switch u.Kind() {
		case types.String:
			return typeRepr{Kind: "TypeKindString"}
		case types.Bool:
			return typeRepr{Kind: "TypeKindBool"}
		case types.Int, types.Int8, types.Int16, types.Int32, types.Int64,
			types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64:
			return typeRepr{Kind: "TypeKindInt"}
		case types.Float32, types.Float64:
			return typeRepr{Kind: "TypeKindFloat"}
		default:
			return typeRepr{Kind: "TypeKindUnknown"}
		}
	case *types.Slice:
		return typeRepr{Kind: "TypeKindList"}
	case *types.Map:
		return typeRepr{Kind: "TypeKindMap"}
	case *types.Chan:
		return typeRepr{Kind: "TypeKindChannel"}
	case *types.Pointer:
		if named, ok := u.Elem().(*types.Named); ok && named.Obj().Pkg() != nil {
			return typeRepr{Kind: "TypeKindReference", Name: "*" + named.Obj().Pkg().Name() + "." + named.Obj().Name()}
		}
		return typeRepr{Kind: "TypeKindReference"}
	case *types.Interface:
		// Check if this is the error interface
		if isErrorType(t) {
			return typeRepr{Kind: "TypeKindNamed", Name: "error"}
		}
		return typeRepr{Kind: "TypeKindInterface"}
	case *types.Signature:
		return typeRepr{Kind: "TypeKindFunction"}
	case *types.Struct:
		return typeRepr{Kind: "TypeKindStruct"}
	default:
		return typeRepr{Kind: "TypeKindUnknown"}
	}
}

// isErrorType checks if a Go type is the error interface.
// We check the original (non-underlying) type first for named "error".
func isErrorType(t types.Type) bool {
	if named, ok := t.(*types.Named); ok {
		return named.Obj().Name() == "error" && named.Obj().Pkg() == nil
	}
	// Also check if the underlying is an interface with Error() string
	iface, ok := t.Underlying().(*types.Interface)
	if !ok {
		return false
	}
	for m := range iface.Methods() {
		if m.Name() == "Error" {
			sig := m.Type().(*types.Signature)
			if sig.Params().Len() == 0 && sig.Results().Len() == 1 {
				if basic, ok := sig.Results().At(0).Type().(*types.Basic); ok && basic.Kind() == types.String {
					return true
				}
			}
		}
	}
	return false
}

// extractInterfaces collects all exported interface type names from the
// scanned Go stdlib packages. Returns qualified names like "io.Reader".
func extractInterfaces(imp types.Importer, specs []funcSpec) []string {
	seen := map[string]bool{}
	for _, spec := range specs {
		pkg, err := imp.Import(spec.pkg)
		if err != nil {
			continue
		}
		alias := kukichaAlias(spec.pkg)
		scope := pkg.Scope()
		for _, name := range scope.Names() {
			obj := scope.Lookup(name)
			tn, ok := obj.(*types.TypeName)
			if !ok || !tn.Exported() {
				continue
			}
			if _, isIface := tn.Type().Underlying().(*types.Interface); isIface {
				seen[alias+"."+name] = true
			}
		}
	}
	result := make([]string, 0, len(seen))
	for k := range seen {
		result = append(result, k)
	}
	sort.Strings(result)
	return result
}

func formatOutput(entries []entry, ifaces []string) []byte {
	var lines []string
	for _, e := range entries {
		var retParts []string
		for _, r := range e.Returns {
			if r.Name != "" {
				retParts = append(retParts, fmt.Sprintf("{Kind: %s, Name: %q}", r.Kind, r.Name))
			} else {
				retParts = append(retParts, fmt.Sprintf("{Kind: %s}", r.Kind))
			}
		}
		lines = append(lines, fmt.Sprintf(
			"\t%q: {Count: %d, Types: []goStdlibType{%s}},",
			e.QualifiedName, e.ReturnCount, strings.Join(retParts, ", "),
		))
	}

	var ifaceLines []string
	for _, name := range ifaces {
		ifaceLines = append(ifaceLines, fmt.Sprintf("\t%q: true,", name))
	}

	src := fmt.Sprintf(`// Code generated by cmd/gengostdlib; DO NOT EDIT.
// Run "make gengostdlib" to regenerate after updating the function list.
//
// Source: Go standard library function signatures extracted via go/importer.
// This replaces the hand-maintained knownExternalReturns entries.
//
// Type definitions (goStdlibType, goStdlibEntry) live in stdlib_types.go.

package semantic

// generatedGoStdlib maps qualified Go stdlib function names to their return
// signature. Consumed by analyzeMethodCallExpr to determine return counts
// and type info without hand-maintained switch statements.
var generatedGoStdlib = map[string]goStdlibEntry{
%s
}

// generatedGoInterfaces lists qualified Go stdlib type names that are interfaces.
// Used by codegen to decide between type assertion (x.(T)) and type conversion (T(x)).
var generatedGoInterfaces = map[string]bool{
%s
}
`, strings.Join(lines, "\n"), strings.Join(ifaceLines, "\n"))

	formatted, err := format.Source([]byte(src))
	if err != nil {
		return []byte(src)
	}
	return formatted
}
