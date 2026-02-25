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

	// Special case: time.Duration is compatible with int64 (Duration is defined as int64 in Go)
	if (t1.Kind == TypeKindNamed && t1.Name == "time.Duration" && t2.Kind == TypeKindInt) ||
		(t2.Kind == TypeKindNamed && t2.Name == "time.Duration" && t1.Kind == TypeKindInt) {
		return true
	}

	// Must be same kind
	if t1.Kind != t2.Kind {
		return false
	}

	// Check nested types for compound types
	switch t1.Kind {
	case TypeKindList, TypeKindChannel, TypeKindReference:
		return a.typesCompatible(t1.ElementType, t2.ElementType)
	case TypeKindMap:
		return a.typesCompatible(t1.KeyType, t2.KeyType) && a.typesCompatible(t1.ValueType, t2.ValueType)
	case TypeKindNamed:
		return t1.Name == t2.Name
	default:
		return true
	}
}
