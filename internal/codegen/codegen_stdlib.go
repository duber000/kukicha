package codegen

import (
	"fmt"
	"strings"

	"github.com/duber000/kukicha/internal/ast"
)

// inferExprReturnType tries to infer the return type of an expression lambda body.
// Returns empty string if it can't determine the type.
func (g *Generator) inferExprReturnType(expr ast.Expression) string {
	switch e := expr.(type) {
	case *ast.BinaryExpr:
		switch e.Operator {
		case "==", "!=", "<", ">", "<=", ">=", "equals", "not equals",
			"and", "or", "&&", "||", "in", "not in":
			return "bool"
		case "+", "-", "*", "/", "%":
			// Arithmetic — try to infer from operands
			leftType := g.inferExprReturnType(e.Left)
			if leftType != "" {
				return leftType
			}
			return g.inferExprReturnType(e.Right)
		}
	case *ast.UnaryExpr:
		if e.Operator == "not" || e.Operator == "!" {
			return "bool"
		}
	case *ast.BooleanLiteral:
		return "bool"
	case *ast.IntegerLiteral:
		return "int"
	case *ast.FloatLiteral:
		return "float64"
	case *ast.StringLiteral:
		return "string"
	case *ast.CallExpr:
		// Can't easily determine return type of arbitrary call
		return ""
	}
	// For field access, method calls, etc. — can't determine without full type info
	return ""
}

// inferBlockReturnType scans a block's return statements to infer return type.
func (g *Generator) inferBlockReturnType(block *ast.BlockStmt) string {
	for _, stmt := range block.Statements {
		if ret, ok := stmt.(*ast.ReturnStmt); ok {
			if len(ret.Values) == 1 {
				return g.inferExprReturnType(ret.Values[0])
			}
		}
	}
	return ""
}

// inferStdlibTypeParameters infers type parameters for stdlib/iterator functions
// This enables special transpilation where iter.Seq → iter.Seq[T]
func (g *Generator) inferStdlibTypeParameters(decl *ast.FunctionDecl) []*TypeParameter {
	var typeParams []*TypeParameter
	usesIterSeq := false
	needsTwoTypes := false

	// Check if function uses iter.Seq and scan for any2 placeholder
	for _, param := range decl.Parameters {
		if g.isIterSeqType(param.Type) {
			usesIterSeq = true
		}
		if g.typeContainsPlaceholder(param.Type, "any2") {
			needsTwoTypes = true
		}
	}

	for _, ret := range decl.Returns {
		if g.isIterSeqType(ret) {
			usesIterSeq = true
		}
		if g.typeContainsPlaceholder(ret, "any2") {
			needsTwoTypes = true
		}
	}

	// Generate type parameters
	if usesIterSeq {
		typeParams = append(typeParams, &TypeParameter{
			Name:        "T",
			Placeholder: "any",
			Constraint:  "any",
		})

		if needsTwoTypes {
			typeParams = append(typeParams, &TypeParameter{
				Name:        "U",
				Placeholder: "any2",
				Constraint:  "any",
			})
		}
	}

	return typeParams
}

// isStdlibSlice checks if we're generating code in stdlib/slice
func (g *Generator) isStdlibSlice() bool {
	return strings.Contains(g.sourceFile, "stdlib/slice/") || strings.Contains(g.sourceFile, "stdlib\\slice\\")
}

// isStdlibFetch checks if we're generating code in stdlib/fetch.
func (g *Generator) isStdlibFetch() bool {
	return strings.Contains(g.sourceFile, "stdlib/fetch/") || strings.Contains(g.sourceFile, "stdlib\\fetch\\")
}

// isStdlibJSON checks if we're generating code in stdlib/json.
func (g *Generator) isStdlibJSON() bool {
	return strings.Contains(g.sourceFile, "stdlib/json/") || strings.Contains(g.sourceFile, "stdlib\\json\\")
}

// inferSliceTypeParameters infers type parameters for stdlib/slice functions.
// GroupBy gets [T any, K comparable]; a curated set of other functions get [T any].
//
// Functions are excluded from generic treatment when they either:
//   - return `empty` as a value of type T (which becomes nil — invalid for unconstrained generics), or
//   - use T as a map key (requires comparable), or
//   - delegate to Go stdlib functions that require comparable (slices.Contains, slices.Index).
//
// The excluded functions (Get, FirstOne, LastOne, Find, FindLast, Pop, Shift, Unique,
// Contains, IndexOf) retain []any signatures until the codegen can emit proper zero
// values for generic type parameters.
func (g *Generator) inferSliceTypeParameters(decl *ast.FunctionDecl) []*TypeParameter {
	var typeParams []*TypeParameter

	// GroupBy has two type parameters: T (element) and K (comparable map key)
	if decl.Name.Value == "GroupBy" {
		typeParams = append(typeParams, &TypeParameter{
			Name:        "T",
			Placeholder: "any",
			Constraint:  "any",
		})
		typeParams = append(typeParams, &TypeParameter{
			Name:        "K",
			Placeholder: "any2",
			Constraint:  "comparable",
		})
		return typeParams
	}

	// Allowlist: functions that are safe to parameterise with [T any].
	// These never return `empty` as a value of type T, never use T as a map key,
	// and do not delegate to comparable-constrained Go stdlib helpers.
	genericSafe := map[string]bool{
		"Filter":     true,
		"Map":        true,
		"First":      true,
		"Last":       true,
		"Drop":       true,
		"DropLast":   true,
		"Reverse":    true,
		"Chunk":      true,
		"IsEmpty":    true,
		"IsNotEmpty": true,
		"Concat":     true,
		"GetOr":      true,
		"FirstOr":    true,
		"LastOr":     true,
		"FindOr":     true,
		"FindIndex":  true,
		"FindLastOr": true,
	}
	if !genericSafe[decl.Name.Value] {
		return typeParams
	}

	// If the signature references the "any" placeholder, emit [T any] and substitute
	// "any" → T throughout parameters and return types.
	usesAny := false
	for _, param := range decl.Parameters {
		if g.typeContainsPlaceholder(param.Type, "any") {
			usesAny = true
			break
		}
	}
	if !usesAny {
		for _, ret := range decl.Returns {
			if g.typeContainsPlaceholder(ret, "any") {
				usesAny = true
				break
			}
		}
	}
	if usesAny {
		typeParams = append(typeParams, &TypeParameter{
			Name:        "T",
			Placeholder: "any",
			Constraint:  "any",
		})
	}

	return typeParams
}

// inferFetchTypeParameters infers type parameters for selected stdlib/fetch helpers.
// Json uses placeholders to produce: func Json[T any](resp *http.Response, sample T) (T, error)
func (g *Generator) inferFetchTypeParameters(decl *ast.FunctionDecl) []*TypeParameter {
	if decl.Name == nil || decl.Name.Value != "Json" {
		return nil
	}

	usesAny := false
	for _, param := range decl.Parameters {
		if g.typeContainsPlaceholder(param.Type, "any") {
			usesAny = true
			break
		}
	}
	if !usesAny {
		for _, ret := range decl.Returns {
			if g.typeContainsPlaceholder(ret, "any") {
				usesAny = true
				break
			}
		}
	}
	if !usesAny {
		return nil
	}

	return []*TypeParameter{
		{
			Name:        "T",
			Placeholder: "any",
			Constraint:  "any",
		},
	}
}

// inferJSONTypeParameters infers type parameters for selected stdlib/json helpers.
// DecodeRead uses placeholders to produce: func DecodeRead[T any](reader io.Reader, sample T) (T, error)
func (g *Generator) inferJSONTypeParameters(decl *ast.FunctionDecl) []*TypeParameter {
	if decl.Name == nil || decl.Name.Value != "DecodeRead" {
		return nil
	}

	usesAny := false
	for _, param := range decl.Parameters {
		if g.typeContainsPlaceholder(param.Type, "any") {
			usesAny = true
			break
		}
	}
	if !usesAny {
		for _, ret := range decl.Returns {
			if g.typeContainsPlaceholder(ret, "any") {
				usesAny = true
				break
			}
		}
	}
	if !usesAny {
		return nil
	}

	return []*TypeParameter{
		{
			Name:        "T",
			Placeholder: "any",
			Constraint:  "any",
		},
	}
}

// isIterSeqType checks if a type is iter.Seq (to be made generic)
func (g *Generator) isIterSeqType(typeAnn ast.TypeAnnotation) bool {
	if namedType, ok := typeAnn.(*ast.NamedType); ok {
		// Check for "iter.Seq" or just "Seq" in iter context
		return namedType.Name == "iter.Seq" ||
			(g.isStdlibIter && namedType.Name == "Seq")
	}
	return false
}

// typeContainsPlaceholder recursively checks if a type annotation tree
// contains the given placeholder name (e.g., "any2")
func (g *Generator) typeContainsPlaceholder(typeAnn ast.TypeAnnotation, placeholder string) bool {
	if typeAnn == nil {
		return false
	}
	switch t := typeAnn.(type) {
	case *ast.PrimitiveType:
		return t.Name == placeholder
	case *ast.NamedType:
		return t.Name == placeholder
	case *ast.ListType:
		return g.typeContainsPlaceholder(t.ElementType, placeholder)
	case *ast.MapType:
		return g.typeContainsPlaceholder(t.KeyType, placeholder) || g.typeContainsPlaceholder(t.ValueType, placeholder)
	case *ast.ChannelType:
		return g.typeContainsPlaceholder(t.ElementType, placeholder)
	case *ast.ReferenceType:
		return g.typeContainsPlaceholder(t.ElementType, placeholder)
	case *ast.FunctionType:
		for _, param := range t.Parameters {
			if g.typeContainsPlaceholder(param, placeholder) {
				return true
			}
		}
		for _, ret := range t.Returns {
			if g.typeContainsPlaceholder(ret, placeholder) {
				return true
			}
		}
	}
	return false
}

// isLikelyInterfaceType checks if a Go type name is likely an interface type.
// Used to determine whether empty Type should generate nil (interface) vs Type{} (struct).
func (g *Generator) isLikelyInterfaceType(typeName string) bool {
	// "error" is always an interface
	if typeName == "error" {
		return true
	}

	// Check current program's declarations for interface types
	for _, decl := range g.program.Declarations {
		if iface, ok := decl.(*ast.InterfaceDecl); ok {
			if iface.Name.Value == typeName {
				return true
			}
		}
	}

	// Common standard library interfaces
	knownInterfaces := map[string]bool{
		"io.Reader":           true,
		"io.Writer":           true,
		"io.Closer":           true,
		"io.ReadCloser":       true,
		"io.WriteCloser":      true,
		"io.ReadWriter":       true,
		"io.ReadWriteCloser":  true,
		"io.ReaderFrom":       true,
		"io.WriterTo":         true,
		"io.Seeker":           true,
		"io.ReadSeeker":       true,
		"io.ReadWriteSeeker":  true,
		"fmt.Stringer":        true,
		"fmt.Scanner":         true,
		"http.Handler":        true,
		"http.ResponseWriter": true,
		"context.Context":     true,
		"sort.Interface":      true,
		"net.Conn":            true,
		"net.Listener":        true,
		"net.Error":           true,
	}

	return knownInterfaces[typeName]
}

// zeroValueForType returns a Go expression for the zero value of a type annotation.
func (g *Generator) zeroValueForType(typeAnn ast.TypeAnnotation) string {
	switch t := typeAnn.(type) {
	case *ast.PrimitiveType:
		switch t.Name {
		case "string":
			return "\"\""
		case "bool":
			return "false"
		default:
			return "0"
		}
	case *ast.ListType:
		return g.generateTypeAnnotation(t) + "{}"
	case *ast.MapType:
		return g.generateTypeAnnotation(typeAnn) + "{}"
	case *ast.ReferenceType, *ast.ChannelType, *ast.FunctionType:
		return "nil"
	case *ast.NamedType:
		typeName := g.generateTypeAnnotation(t)
		if g.isLikelyInterfaceType(typeName) {
			return "nil"
		}
		return fmt.Sprintf("*new(%s)", typeName)
	default:
		return "nil"
	}
}

func (g *Generator) errorValueExpr(expr ast.Expression, errVar string) string {
	// Defensive guard: the parser lexes `error` as TOKEN_ERROR, not TOKEN_IDENTIFIER,
	// so this branch is currently unreachable. Kept as documentation of intent.
	if id, ok := expr.(*ast.Identifier); ok && id.Value == "error" {
		return errVar
	}
	if strLit, ok := expr.(*ast.StringLiteral); ok {
		msg := strLit.Value
		if strings.Contains(msg, "{error}") {
			msg = strings.ReplaceAll(msg, "{error}", fmt.Sprintf("{%s}", errVar))
		}
		return fmt.Sprintf("errors.New(%s)", g.generateStringInterpolation(msg))
	}
	return fmt.Sprintf("errors.New(%s)", g.exprToString(expr))
}

func (g *Generator) inferReturnCount(expr ast.Expression) (int, bool) {
	if g.exprReturnCounts != nil {
		if count, ok := g.exprReturnCounts[expr]; ok {
			return count, true
		}
	}
	switch e := expr.(type) {
	case *ast.PipeExpr:
		return g.inferReturnCount(e.Right)
	case *ast.CallExpr:
		if id, ok := e.Function.(*ast.Identifier); ok {
			return g.returnCountForFunctionName(id.Value)
		}
	case *ast.MethodCallExpr:
		// Method return counts require type info; skip for now.
		return 0, false
	}
	return 0, false
}

func (g *Generator) returnCountForFunctionName(name string) (int, bool) {
	for _, decl := range g.program.Declarations {
		if fn, ok := decl.(*ast.FunctionDecl); ok {
			if fn.Receiver == nil && fn.Name.Value == name {
				return len(fn.Returns), true
			}
		}
	}
	return 0, false
}

