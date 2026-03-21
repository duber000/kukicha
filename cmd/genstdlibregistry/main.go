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
	count           int
	types           []typeRepr        // per-position return types
	paramNames      []string          // parameter names (for named argument support)
	defaultValues   []string          // Go expression strings for default values; "" = no default
	paramFuncParams map[int][]typeRepr // func-typed param index → inner param types (for lambda inference)
}

// typeRepr is the generator's representation of a type, emitted as goStdlibType.
type typeRepr struct {
	kind        string     // TypeKind constant name (e.g., "TypeKindInt")
	name        string     // For TypeKindNamed (e.g., "error", "time.Time")
	elementType *typeRepr  // element type for TypeKindList
	keyType     *typeRepr  // key type for TypeKindMap
	valueType   *typeRepr  // value type for TypeKindMap
}

// scanResult holds the scanned registry and deprecated map.
type scanResult struct {
	registry     map[string]registryEntry
	deprecated   map[string]string // qualified name → deprecation message
	genericClass map[string]string // qualified name → generic class ("T", "K", or "TK")
	security     map[string]string // qualified name → security category (sql, html, fetch, files, redirect, shell)
	interfaces   map[string]bool   // qualified interface names (e.g., "mcp.Server")
	panics       map[string]string // qualified name → panics message
}

// scanRegistry reads and parses all .kuki files in paths, returning a map of
// qualified function name → signature info plus a map of deprecated functions.
// Skips _test.kuki files, unexported functions, methods, and void functions.
// Returns accumulated errors.
func scanRegistry(paths []string) (scanResult, []error) {
	result := scanResult{
		registry:     map[string]registryEntry{},
		deprecated:   map[string]string{},
		genericClass: map[string]string{},
		security:     map[string]string{},
		interfaces:   map[string]bool{},
		panics:       map[string]string{},
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
			// Collect exported interface declarations.
			if iface, ok := decl.(*ast.InterfaceDecl); ok {
				name := iface.Name.Value
				if len(name) > 0 && name[0] >= 'A' && name[0] <= 'Z' {
					result.interfaces[pkgName+"."+name] = true
				}
				continue
			}

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

			// Check for directives
			for _, dir := range fd.Directives {
				switch dir.Name {
				case "deprecated":
					msg := "deprecated"
					if len(dir.Args) > 0 {
						msg = dir.Args[0]
					}
					result.deprecated[key] = msg
				case "security":
					if len(dir.Args) > 0 {
						result.security[key] = dir.Args[0]
					}
				case "panics":
					msg := "panics"
					if len(dir.Args) > 0 {
						msg = dir.Args[0]
					}
					result.panics[key] = msg
				}
			}

			// Detect generic placeholder usage for stdlib functions.
			// This drives codegen's type parameter inference so new functions
			// don't need to be manually added to hardcoded allowlists.
			if pkgName == "slice" || pkgName == "sort" || pkgName == "concurrent" {
				usesAny := signatureContainsPlaceholder(fd, "any")
				usesAny2 := signatureContainsPlaceholder(fd, "any2")
				usesOrdered := signatureContainsPlaceholder(fd, "ordered")
				usesResult := signatureContainsPlaceholder(fd, "result")
				if usesAny && usesResult {
					result.genericClass[key] = "TR"
				} else if usesAny && usesOrdered {
					result.genericClass[key] = "TO"
				} else if usesAny && usesAny2 {
					result.genericClass[key] = "TK"
				} else if usesAny2 {
					result.genericClass[key] = "K"
				} else if usesOrdered {
					result.genericClass[key] = "O"
				} else if usesAny {
					result.genericClass[key] = "T"
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

			// Extract parameter names, default values, and func-typed param inner types
			paramNames := make([]string, len(fd.Parameters))
			defaultValues := make([]string, len(fd.Parameters))
			hasDefaults := false
			var paramFuncParams map[int][]typeRepr
			for i, param := range fd.Parameters {
				paramNames[i] = param.Name.Value
				if param.DefaultValue != nil {
					defaultValues[i] = defaultValueToGo(param.DefaultValue)
					hasDefaults = true
				}
				// Detect function-typed parameters and extract their inner param types.
				// This enables lambda parameter type inference at the call site.
				// Placeholder names ("any", "any2", "ordered") are kept as-is for
				// substitution at inference time; other unqualified named types are
				// prefixed with the package name so codegen emits "cli.Args" etc.
				if ft, ok := param.Type.(*ast.FunctionType); ok && len(ft.Parameters) > 0 {
					innerTypes := make([]typeRepr, len(ft.Parameters))
					for j, innerParam := range ft.Parameters {
						tr := typeAnnotationToRepr(innerParam)
						// Qualify unqualified named types with the package name.
						if tr.kind == "TypeKindNamed" && tr.name != "" &&
							!strings.Contains(tr.name, ".") &&
							tr.name != "any" && tr.name != "any2" && tr.name != "ordered" && tr.name != "result" && tr.name != "error" {
							tr.name = pkgName + "." + tr.name
						}
						innerTypes[j] = tr
					}
					if paramFuncParams == nil {
						paramFuncParams = make(map[int][]typeRepr)
					}
					paramFuncParams[i] = innerTypes
				}
			}
			if !hasDefaults {
				defaultValues = nil
			}

			if existing, exists := result.registry[key]; !exists || returnCount > existing.count {
				result.registry[key] = registryEntry{
					count:           returnCount,
					types:           types,
					paramNames:      paramNames,
					defaultValues:   defaultValues,
					paramFuncParams: paramFuncParams,
				}
			}
		}
	}

	return result, errs
}

// signatureContainsPlaceholder checks if a function's parameters or return types
// contain a placeholder name (e.g., "any" or "any2").
func signatureContainsPlaceholder(fd *ast.FunctionDecl, placeholder string) bool {
	for _, param := range fd.Parameters {
		if typeContainsPlaceholder(param.Type, placeholder) {
			return true
		}
	}
	for _, ret := range fd.Returns {
		if typeContainsPlaceholder(ret, placeholder) {
			return true
		}
	}
	return false
}

// typeContainsPlaceholder recursively checks if a type annotation tree
// contains the given placeholder name.
func typeContainsPlaceholder(typeAnn ast.TypeAnnotation, placeholder string) bool {
	if typeAnn == nil {
		return false
	}
	switch t := typeAnn.(type) {
	case *ast.PrimitiveType:
		return t.Name == placeholder
	case *ast.NamedType:
		return t.Name == placeholder
	case *ast.ListType:
		return typeContainsPlaceholder(t.ElementType, placeholder)
	case *ast.MapType:
		return typeContainsPlaceholder(t.KeyType, placeholder) || typeContainsPlaceholder(t.ValueType, placeholder)
	case *ast.ChannelType:
		return typeContainsPlaceholder(t.ElementType, placeholder)
	case *ast.ReferenceType:
		return typeContainsPlaceholder(t.ElementType, placeholder)
	case *ast.FunctionType:
		for _, param := range t.Parameters {
			if typeContainsPlaceholder(param, placeholder) {
				return true
			}
		}
		for _, ret := range t.Returns {
			if typeContainsPlaceholder(ret, placeholder) {
				return true
			}
		}
	}
	return false
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
		elem := typeAnnotationToRepr(t.ElementType)
		return typeRepr{kind: "TypeKindList", elementType: &elem}
	case *ast.MapType:
		key := typeAnnotationToRepr(t.KeyType)
		val := typeAnnotationToRepr(t.ValueType)
		return typeRepr{kind: "TypeKindMap", keyType: &key, valueType: &val}
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
	parts := []string{fmt.Sprintf("Kind: %s", tr.kind)}
	if tr.name != "" {
		parts = append(parts, fmt.Sprintf("Name: %q", tr.name))
	}
	if tr.elementType != nil {
		parts = append(parts, fmt.Sprintf("ElementType: &goStdlibType{%s}", formatTypeReprInner(*tr.elementType)))
	}
	if tr.keyType != nil {
		parts = append(parts, fmt.Sprintf("KeyType: &goStdlibType{%s}", formatTypeReprInner(*tr.keyType)))
	}
	if tr.valueType != nil {
		parts = append(parts, fmt.Sprintf("ValueType: &goStdlibType{%s}", formatTypeReprInner(*tr.valueType)))
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

// formatTypeReprInner formats a typeRepr for use inside a nested &goStdlibType{}.
func formatTypeReprInner(tr typeRepr) string {
	parts := []string{fmt.Sprintf("Kind: %s", tr.kind)}
	if tr.name != "" {
		parts = append(parts, fmt.Sprintf("Name: %q", tr.name))
	}
	return strings.Join(parts, ", ")
}

// defaultValueToGo converts an AST default value expression to a Go expression string.
func defaultValueToGo(expr ast.Expression) string {
	switch e := expr.(type) {
	case *ast.StringLiteral:
		return fmt.Sprintf("%q", e.Value)
	case *ast.IntegerLiteral:
		return fmt.Sprintf("%d", e.Value)
	case *ast.FloatLiteral:
		return fmt.Sprintf("%g", e.Value)
	case *ast.BooleanLiteral:
		if e.Value {
			return "true"
		}
		return "false"
	case *ast.EmptyExpr:
		return "nil"
	default:
		return ""
	}
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

		// Build the default values slice literal (only if there are defaults)
		defaultsLiteral := ""
		if len(v.defaultValues) > 0 {
			defaultParts := make([]string, len(v.defaultValues))
			for i, dv := range v.defaultValues {
				defaultParts[i] = fmt.Sprintf("%q", dv)
			}
			defaultsLiteral = strings.Join(defaultParts, ", ")
		}

		// Build the ParamFuncParams map literal (only if any param is func-typed)
		paramFuncLiteral := ""
		if len(v.paramFuncParams) > 0 {
			// Sort keys for deterministic output
			pfpKeys := make([]int, 0, len(v.paramFuncParams))
			for idx := range v.paramFuncParams {
				pfpKeys = append(pfpKeys, idx)
			}
			sort.Ints(pfpKeys)
			var pfpEntries []string
			for _, idx := range pfpKeys {
				innerTypes := v.paramFuncParams[idx]
				innerParts := make([]string, len(innerTypes))
				for j, tr := range innerTypes {
					innerParts[j] = formatTypeRepr(tr)
				}
				pfpEntries = append(pfpEntries, fmt.Sprintf("%d: {%s}", idx, strings.Join(innerParts, ", ")))
			}
			paramFuncLiteral = strings.Join(pfpEntries, ", ")
		}

		entry := fmt.Sprintf("{Count: %d, Types: []goStdlibType{%s}, ParamNames: []string{%s}", v.count, typesLiteral, paramsLiteral)
		if defaultsLiteral != "" {
			entry += fmt.Sprintf(", DefaultValues: []string{%s}", defaultsLiteral)
		}
		if paramFuncLiteral != "" {
			entry += fmt.Sprintf(", ParamFuncParams: map[int][]goStdlibType{%s}", paramFuncLiteral)
		}
		entry += "}"
		entries = append(entries, fmt.Sprintf("\t%q: %s,", k, entry))
	}
	sort.Strings(entries)

	securityEntries := make([]string, 0, len(result.security))
	for k, v := range result.security {
		securityEntries = append(securityEntries, fmt.Sprintf("\t%q: %q,", k, v))
	}
	sort.Strings(securityEntries)

	genericEntries := make([]string, 0, len(result.genericClass))
	for k, v := range result.genericClass {
		genericEntries = append(genericEntries, fmt.Sprintf("\t%q: %q,", k, v))
	}
	sort.Strings(genericEntries)

	depEntries := make([]string, 0, len(result.deprecated))
	for k, v := range result.deprecated {
		depEntries = append(depEntries, fmt.Sprintf("\t%q: %q,", k, v))
	}
	sort.Strings(depEntries)

	panicsEntries := make([]string, 0, len(result.panics))
	for k, v := range result.panics {
		panicsEntries = append(panicsEntries, fmt.Sprintf("\t%q: %q,", k, v))
	}
	sort.Strings(panicsEntries)

	ifaceEntries := make([]string, 0, len(result.interfaces))
	for k := range result.interfaces {
		ifaceEntries = append(ifaceEntries, fmt.Sprintf("\t%q: true,", k))
	}
	sort.Strings(ifaceEntries)

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

// generatedStdlibPanics maps qualified Kukicha stdlib function names to their
// panic messages. Populated from # kuki:panics directives in stdlib .kuki files.
var generatedStdlibPanics = map[string]string{
%s
}

// generatedSecurityFunctions maps qualified stdlib function names to their
// security check category. Populated from # kuki:security directives in .kuki files.
// Categories: "sql", "html", "fetch", "files", "redirect", "shell"
var generatedSecurityFunctions = map[string]string{
%s
}

// generatedSliceGenericClass maps stdlib function names to their generic
// classification based on placeholder usage in their signatures:
//   - "T"  = uses "any" placeholder only  → emits [T any]
//   - "K"  = uses "any2" placeholder only → emits [K comparable]
//   - "TK" = uses both                    → emits [T any, K comparable]
//   - "TR" = uses "any" and "result"      → emits [T any, R any]
//
// Functions not in this map do not use placeholders and are not made generic.
var generatedSliceGenericClass = map[string]string{
%s
}

// generatedStdlibInterfaces lists qualified Kukicha stdlib type names that are interfaces.
// Used by codegen to decide between type assertion (x.(T)) and type conversion (T(x)).
var generatedStdlibInterfaces = map[string]bool{
%s
}
`, strings.Join(entries, "\n"), strings.Join(depEntries, "\n"), strings.Join(panicsEntries, "\n"), strings.Join(securityEntries, "\n"), strings.Join(genericEntries, "\n"), strings.Join(ifaceEntries, "\n"))

	formatted, fmtErr := format.Source([]byte(src))
	if fmtErr != nil {
		// Fall back to unformatted — still valid Go.
		return []byte(src)
	}
	return formatted
}
