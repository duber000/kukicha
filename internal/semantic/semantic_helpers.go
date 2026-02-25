package semantic

import (
	"strings"

	"github.com/duber000/kukicha/internal/ast"
)

func isValidIdentifier(name string) bool {
	if len(name) == 0 {
		return false
	}
	// Must start with letter and contain only letters, digits, underscores
	for i, r := range name {
		if i == 0 {
			if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_') {
				return false
			}
		} else {
			if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_') {
				return false
			}
		}
	}
	return true
}

func (a *Analyzer) extractPackageName(imp *ast.ImportDecl) string {
	if imp.Alias != nil {
		return imp.Alias.Value
	}

	// Extract package name from path
	path := strings.Trim(imp.Path.Value, "\"")

	// Rewrite stdlib imports to full module path before extracting package name
	if strings.HasPrefix(path, "stdlib/") {
		// Remap stdlib/iter to stdlib/iterator
		if path == "stdlib/iter" {
			path = "stdlib/iterator"
		}
		path = "github.com/duber000/kukicha/" + path
	}

	parts := strings.Split(path, "/")
	name := parts[len(parts)-1]

	// Handle version suffixes
	// 1. Dot-versions: gopkg.in/yaml.v3 → yaml
	if idx := strings.Index(name, ".v"); idx != -1 {
		name = name[:idx]
	}

	// 2. Slash-versions: encoding/json/v2 → use second-to-last segment
	//    This handles Go module major version suffixes
	if len(parts) >= 2 && len(name) >= 2 && name[0] == 'v' && name[1] >= '0' && name[1] <= '9' {
		// This looks like a version suffix (v2, v3, etc.)
		name = parts[len(parts)-2] // Use parent directory name

		// Handle gopkg.in dot-versions in parent too
		if idx := strings.Index(name, ".v"); idx != -1 {
			name = name[:idx]
		}
	}

	return name
}

func isExported(name string) bool {
	if len(name) == 0 {
		return false
	}
	// Exported if starts with uppercase letter
	r := rune(name[0])
	return r >= 'A' && r <= 'Z'
}

func isNumericType(t *TypeInfo) bool {
	return t.Kind == TypeKindInt || t.Kind == TypeKindFloat || t.Kind == TypeKindUnknown
}

func primitiveTypeFromString(name string) *TypeInfo {
	switch name {
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64":
		return &TypeInfo{Kind: TypeKindInt}
	case "float32", "float64":
		return &TypeInfo{Kind: TypeKindFloat}
	case "string":
		return &TypeInfo{Kind: TypeKindString}
	case "bool":
		return &TypeInfo{Kind: TypeKindBool}
	case "byte":
		return &TypeInfo{Kind: TypeKindInt}
	case "rune":
		return &TypeInfo{Kind: TypeKindInt}
	default:
		return &TypeInfo{Kind: TypeKindUnknown}
	}
}
