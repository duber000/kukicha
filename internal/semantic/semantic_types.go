package semantic

import (
	"fmt"
	"strings"

	"github.com/duber000/kukicha/internal/ast"
)

// validateTypeAnnotation checks that a type annotation is valid
func (a *Analyzer) validateTypeAnnotation(typeAnn ast.TypeAnnotation) {
	switch t := typeAnn.(type) {
	case *ast.NamedType:
		// Allow built-in Go types
		builtInTypes := map[string]bool{
			"interface{}": true,
			"any":         true,
			"any2":        true, // Placeholder for second generic type parameter
			"error":       true,
			"byte":        true,
			"rune":        true,
		}
		if builtInTypes[t.Name] {
			return // Built-in type is valid
		}

		// Check for qualified type (package.Type)
		if strings.Contains(t.Name, ".") {
			parts := strings.Split(t.Name, ".")
			if len(parts) != 2 {
				a.error(t.Pos(), fmt.Sprintf("invalid qualified type '%s'", t.Name))
				return
			}

			pkgName := parts[0]

			// Verify the package is imported
			pkgSymbol := a.symbolTable.Resolve(pkgName)
			if pkgSymbol == nil {
				a.error(t.Pos(), fmt.Sprintf("package '%s' not imported (for type '%s')", pkgName, t.Name))
				return
			}

			// Package is imported - trust that the type exists
			// We can't validate external package types at Kukicha compile time
			return
		}

		// Check that the type exists in symbol table
		symbol := a.symbolTable.Resolve(t.Name)
		if symbol == nil || (symbol.Kind != SymbolType && symbol.Kind != SymbolInterface) {
			a.error(t.Pos(), fmt.Sprintf("undefined type '%s'", t.Name))
		}

		// Warn if the type is deprecated
		if msg, ok := a.deprecatedTypes[t.Name]; ok {
			a.warn(t.Pos(), fmt.Sprintf("'%s' is deprecated: %s", t.Name, msg))
		}
	case *ast.ReferenceType:
		a.validateTypeAnnotation(t.ElementType)
	case *ast.ListType:
		a.validateTypeAnnotation(t.ElementType)
	case *ast.MapType:
		a.validateTypeAnnotation(t.KeyType)
		a.validateTypeAnnotation(t.ValueType)
	case *ast.ChannelType:
		a.validateTypeAnnotation(t.ElementType)
	case *ast.FunctionType:
		// Validate parameter types
		for _, param := range t.Parameters {
			a.validateTypeAnnotation(param)
		}
		// Validate return types
		for _, ret := range t.Returns {
			a.validateTypeAnnotation(ret)
		}
	}
}

// typeAnnotationToTypeInfo converts AST type annotation to TypeInfo
func (a *Analyzer) typeAnnotationToTypeInfo(typeAnn ast.TypeAnnotation) *TypeInfo {
	if typeAnn == nil {
		return &TypeInfo{Kind: TypeKindUnknown}
	}

	switch t := typeAnn.(type) {
	case *ast.PrimitiveType:
		return primitiveTypeFromString(t.Name)
	case *ast.NamedType:
		return &TypeInfo{Kind: TypeKindNamed, Name: t.Name}
	case *ast.ReferenceType:
		return &TypeInfo{
			Kind:        TypeKindReference,
			ElementType: a.typeAnnotationToTypeInfo(t.ElementType),
		}
	case *ast.ListType:
		return &TypeInfo{
			Kind:        TypeKindList,
			ElementType: a.typeAnnotationToTypeInfo(t.ElementType),
		}
	case *ast.MapType:
		return &TypeInfo{
			Kind:      TypeKindMap,
			KeyType:   a.typeAnnotationToTypeInfo(t.KeyType),
			ValueType: a.typeAnnotationToTypeInfo(t.ValueType),
		}
	case *ast.ChannelType:
		return &TypeInfo{
			Kind:        TypeKindChannel,
			ElementType: a.typeAnnotationToTypeInfo(t.ElementType),
		}
	case *ast.FunctionType:
		var params []*TypeInfo
		for _, param := range t.Parameters {
			params = append(params, a.typeAnnotationToTypeInfo(param))
		}
		var returns []*TypeInfo
		for _, ret := range t.Returns {
			returns = append(returns, a.typeAnnotationToTypeInfo(ret))
		}
		return &TypeInfo{
			Kind:    TypeKindFunction,
			Params:  params,
			Returns: returns,
		}
	default:
		return &TypeInfo{Kind: TypeKindUnknown}
	}
}

// isReferenceType checks if a type can be nil
func (a *Analyzer) isReferenceType(t *TypeInfo) bool {
	if t == nil {
		return false
	}
	switch t.Kind {
	case TypeKindReference, TypeKindList, TypeKindMap, TypeKindChannel, TypeKindFunction, TypeKindInterface:
		return true
	case TypeKindNamed:
		if t.Name == "any" || t.Name == "any2" || t.Name == "error" || t.Name == "interface{}" {
			return true
		}
		
		// If qualified name (contains dot), assume it might be a reference/interface (e.g. io.Reader)
		// We can't verify external types, so be lenient and let Go compiler catch errors.
		if strings.Contains(t.Name, ".") {
			return true
		}

		// Resolve named type to see if it's an interface or reference type alias
		sym := a.symbolTable.Resolve(t.Name)
		if sym != nil {
			if sym.Kind == SymbolInterface {
				return true
			}
			if sym.Kind == SymbolType {
				return a.isReferenceType(sym.Type)
			}
		}
		return false
	case TypeKindUnknown:
		return true // Allow leniently
	default:
		return false
	}
}

// typesCompatible checks if two types are compatible
func (a *Analyzer) typesCompatible(t1, t2 *TypeInfo) bool {
	if t1 == nil || t2 == nil {
		return false
	}

	// Unknown types are compatible with anything
	if t1.Kind == TypeKindUnknown || t2.Kind == TypeKindUnknown {
		return true
	}

	// interface{} and any accept any type
	if t1.Kind == TypeKindNamed && (t1.Name == "interface{}" || t1.Name == "any") {
		return true
	}
	if t2.Kind == TypeKindNamed && (t2.Name == "interface{}" || t2.Name == "any") {
		return true
	}

	// error interface accepts structs and named types (we defer implementation check to Go compiler)
	if t1.Kind == TypeKindNamed && t1.Name == "error" {
		if t2.Kind == TypeKindStruct || t2.Kind == TypeKindNamed || t2.Kind == TypeKindReference {
			return true
		}
	}
	if t2.Kind == TypeKindNamed && t2.Name == "error" {
		if t1.Kind == TypeKindStruct || t1.Kind == TypeKindNamed || t1.Kind == TypeKindReference {
			return true
		}
	}

	// Special case: time.Duration is compatible with int64 (Duration is defined as int64 in Go)
	if (t1.Kind == TypeKindNamed && t1.Name == "time.Duration" && t2.Kind == TypeKindInt) ||
		(t2.Kind == TypeKindNamed && t2.Name == "time.Duration" && t1.Kind == TypeKindInt) {
		return true
	}

	// Must be same kind
	if t1.Kind != t2.Kind {
		// Nil is compatible with reference types
		if t1.Kind == TypeKindNil {
			return a.isReferenceType(t2)
		}
		if t2.Kind == TypeKindNil {
			return a.isReferenceType(t1)
		}

		return false
	}

	// Check nested types for compound types
	switch t1.Kind {
	case TypeKindList, TypeKindChannel, TypeKindReference:
		// If either side has no element type info (e.g., from Go stdlib registry),
		// treat as compatible — the Go compiler will catch any real mismatch.
		if t1.ElementType == nil || t2.ElementType == nil {
			return true
		}
		return a.typesCompatible(t1.ElementType, t2.ElementType)
	case TypeKindMap:
		if t1.KeyType == nil || t2.KeyType == nil || t1.ValueType == nil || t2.ValueType == nil {
			return true
		}
		return a.typesCompatible(t1.KeyType, t2.KeyType) && a.typesCompatible(t1.ValueType, t2.ValueType)
	case TypeKindNamed:
		return t1.Name == t2.Name
	default:
		return true
	}
}
