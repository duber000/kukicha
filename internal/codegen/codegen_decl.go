package codegen

import (
	"fmt"
	"maps"
	"strings"

	"github.com/duber000/kukicha/internal/ast"
)

func (g *Generator) generateTypeDecl(decl *ast.TypeDecl) {
	// Type alias (e.g., type Handler func(string))
	if decl.AliasType != nil {
		g.writeLine(fmt.Sprintf("type %s %s", decl.Name.Value, g.generateTypeAnnotation(decl.AliasType)))
		return
	}

	g.write(fmt.Sprintf("type %s struct {", decl.Name.Value))
	g.writeLine("")
	g.indent++

	for _, field := range decl.Fields {
		fieldType := g.generateTypeAnnotation(field.Type)
		line := fmt.Sprintf("%s %s", field.Name.Value, fieldType)
		if field.Tag != "" {
			line += fmt.Sprintf(" `%s`", field.Tag)
		}
		g.writeLine(line)
	}

	g.indent--
	g.writeLine("}")
}

func (g *Generator) generateInterfaceDecl(decl *ast.InterfaceDecl) {
	g.write(fmt.Sprintf("type %s interface {", decl.Name.Value))
	g.writeLine("")
	g.indent++

	for _, method := range decl.Methods {
		// Generate method signature
		params := g.generateParameters(method.Parameters)
		returns := g.generateReturnTypes(method.Returns)

		if returns != "" {
			g.writeLine(fmt.Sprintf("%s(%s) %s", method.Name.Value, params, returns))
		} else {
			g.writeLine(fmt.Sprintf("%s(%s)", method.Name.Value, params))
		}
	}

	g.indent--
	g.writeLine("}")
}

func (g *Generator) generateGlobalVarDecl(stmt *ast.VarDeclStmt) {
	if len(stmt.Names) == 0 {
		return
	}

	// Build comma-separated list of names
	names := make([]string, len(stmt.Names))
	for i, n := range stmt.Names {
		names[i] = n.Value
	}
	namesStr := strings.Join(names, ", ")

	// Generate type if present
	if stmt.Type != nil {
		varType := g.generateTypeAnnotation(stmt.Type)
		if len(stmt.Values) > 0 {
			// With initializer
			values := make([]string, len(stmt.Values))
			for i, v := range stmt.Values {
				values[i] = g.exprToString(v)
			}
			valuesStr := strings.Join(values, ", ")
			g.writeLine(fmt.Sprintf("var %s %s = %s", namesStr, varType, valuesStr))
		} else {
			// Without initializer
			g.writeLine(fmt.Sprintf("var %s %s", namesStr, varType))
		}
	} else if len(stmt.Values) > 0 {
		// No explicit type, but with initializer
		values := make([]string, len(stmt.Values))
		for i, v := range stmt.Values {
			values[i] = g.exprToString(v)
		}
		valuesStr := strings.Join(values, ", ")
		g.writeLine(fmt.Sprintf("var %s = %s", namesStr, valuesStr))
	} else {
		// No type, no initializer - this is unusual for a global variable
		// but we'll generate it anyway (will be zero-valued)
		g.writeLine(fmt.Sprintf("var %s any", namesStr))
	}
}

// generateFunctionDecl generates a Go function from a Kukicha function declaration.
//
// ARCHITECTURE NOTE: For stdlib/iterator and stdlib/slice packages, this function
// performs "generic inference" - it scans the function's parameter and return types
// for placeholder types ("any", "any2") and generates proper Go type parameters.
//
// Example: A Kukicha function like:
//
//	func Filter(items list of any, predicate func(any) bool) list of any
//
// Becomes Go code like:
//
//	func Filter[T any](items []T, predicate func(T) bool) []T
//
// This happens automatically for stdlib packages. User code doesn't need this
// because users import the stdlib and call its generic functions; the Go compiler
// handles type inference for callers.
func (g *Generator) generateFunctionDecl(decl *ast.FunctionDecl) {
	// Set up placeholder mapping for this function
	g.placeholderMap = make(map[string]string)

	g.currentFuncName = decl.Name.Value

	// Check if this is a stdlib function that needs special transpilation
	// (generic type parameter inference from placeholder types)
	var typeParams []*TypeParameter
	if g.isStdlibIter {
		// Generate type parameters from function signature for iter
		typeParams = g.inferStdlibTypeParameters(decl)
		for _, tp := range typeParams {
			g.placeholderMap[tp.Placeholder] = tp.Name
		}
	} else if g.isStdlibSlice() {
		// Generate type parameters from function signature for slice
		typeParams = g.inferSliceTypeParameters(decl)
		for _, tp := range typeParams {
			g.placeholderMap[tp.Placeholder] = tp.Name
		}
	} else if g.isStdlibFetch() {
		// Generate type parameters for selected fetch helpers (e.g., Json)
		typeParams = g.inferFetchTypeParameters(decl)
		for _, tp := range typeParams {
			g.placeholderMap[tp.Placeholder] = tp.Name
		}
	} else if g.isStdlibJSON() {
		// Generate type parameters for selected json helpers (e.g., DecodeRead)
		typeParams = g.inferJSONTypeParameters(decl)
		for _, tp := range typeParams {
			g.placeholderMap[tp.Placeholder] = tp.Name
		}
	}

	// Generate function signature
	signature := "func "

	// Add receiver for methods
	if decl.Receiver != nil {
		receiverType := g.generateTypeAnnotation(decl.Receiver.Type)
		receiverName := decl.Receiver.Name.Value
		signature += fmt.Sprintf("(%s %s) ", receiverName, receiverType)
	}

	// Add function name
	signature += decl.Name.Value

	// Add type parameters if present
	if len(typeParams) > 0 {
		signature += g.generateTypeParameters(typeParams)
	}

	// Add parameters
	params := g.generateFunctionParameters(decl.Parameters)
	signature += fmt.Sprintf("(%s)", params)

	// Add return types
	g.processingReturnType = true
	returns := g.generateReturnTypes(decl.Returns)
	g.processingReturnType = false

	if returns != "" {
		signature += " " + returns
	}

	g.write(signature + " {")
	g.writeLine("")

	// Set return types for type coercion in return statements
	g.currentReturnTypes = decl.Returns

	// Generate body
	if decl.Body != nil {
		g.indent++
		g.generateBlock(decl.Body)
		g.indent--
	}

	g.writeLine("}")

	// Clear function context
	g.placeholderMap = nil
	g.currentFuncName = ""
	g.currentReturnTypes = nil
}

func (g *Generator) generateFunctionLiteral(lit *ast.FunctionLiteral) string {
	// Save current placeholder map and create new one for this literal
	oldPlaceholderMap := g.placeholderMap
	g.placeholderMap = make(map[string]string)

	// Inherit placeholders from parent scope
	maps.Copy(g.placeholderMap, oldPlaceholderMap)

	// Check if this is a stdlib/iterator function literal that needs special transpilation
	var typeParams []*TypeParameter
	if g.isStdlibIter {
		// Create a temporary function decl to reuse the inference logic
		tempDecl := &ast.FunctionDecl{
			Name:       &ast.Identifier{Value: ""}, // dummy name for inference
			Parameters: lit.Parameters,
			Returns:    lit.Returns,
		}
		typeParams = g.inferStdlibTypeParameters(tempDecl)
		for _, tp := range typeParams {
			g.placeholderMap[tp.Placeholder] = tp.Name
		}
	}

	// Generate function signature
	signature := "func"

	// Add type parameters if present
	if len(typeParams) > 0 {
		signature += g.generateTypeParameters(typeParams)
	}

	// Add parameters
	params := g.generateFunctionParameters(lit.Parameters)
	signature += fmt.Sprintf("(%s)", params)

	// Add return types
	returns := g.generateReturnTypes(lit.Returns)
	if returns != "" {
		signature += " " + returns
	}

	// Generate body inline - create temporary generator to capture output
	tempGen := &Generator{
		program:        g.program,
		output:         strings.Builder{},
		indent:         g.indent + 1,
		placeholderMap: g.placeholderMap,
		autoImports:    g.autoImports,
		isStdlibIter:   g.isStdlibIter,
		sourceFile:     g.sourceFile,
	}

	var result strings.Builder
	result.WriteString(signature + " {\n")

	if lit.Body != nil {
		// Generate body statements using temporary generator
		for _, stmt := range lit.Body.Statements {
			tempGen.generateStatement(stmt)
		}
		result.WriteString(tempGen.output.String())
	}

	// Add proper indentation for closing brace
	for i := 0; i < g.indent; i++ {
		result.WriteString("\t")
	}
	result.WriteString("}")

	// Restore placeholder mapping
	g.placeholderMap = oldPlaceholderMap

	return result.String()
}

// generateArrowLambda transpiles an arrow lambda to a Go anonymous function.
// Expression form: (r Repo) => r.Stars > 100  →  func(r Repo) bool { return r.Stars > 100 }
// Block form:      (r Repo) => BLOCK           →  func(r Repo) ReturnType { BLOCK }
func (g *Generator) generateArrowLambda(lambda *ast.ArrowLambda) string {
	// Build parameter string
	var paramParts []string
	for _, param := range lambda.Parameters {
		if param.Type != nil {
			paramParts = append(paramParts, param.Name.Value+" "+g.generateTypeAnnotation(param.Type))
		} else {
			// Untyped parameter — type must be inferred from context.
			// For now, we emit as-is; the Go compiler will catch type errors.
			// Full contextual inference is a semantic analysis extension.
			paramParts = append(paramParts, param.Name.Value)
		}
	}
	params := strings.Join(paramParts, ", ")

	if lambda.Body != nil {
		// Expression lambda: auto-return the expression
		bodyStr := g.exprToString(lambda.Body)

		// Infer return type from the expression for the Go func signature.
		// For typed params, we can determine the return type.
		// For the common case, we omit the return type and let Go infer it
		// from the context (e.g., when passed to a generic function).
		returnType := g.inferExprReturnType(lambda.Body)

		if returnType != "" {
			return fmt.Sprintf("func(%s) %s { return %s }", params, returnType, bodyStr)
		}
		return fmt.Sprintf("func(%s) { return %s }", params, bodyStr)
	}

	if lambda.Block != nil {
		// Block lambda: generate as multi-line anonymous function
		returnType := g.inferBlockReturnType(lambda.Block)

		// Create temporary generator to capture body output
		tempGen := &Generator{
			program:        g.program,
			output:         strings.Builder{},
			indent:         g.indent + 1,
			placeholderMap: g.placeholderMap,
			autoImports:    g.autoImports,
			isStdlibIter:   g.isStdlibIter,
			sourceFile:     g.sourceFile,
		}

		for _, stmt := range lambda.Block.Statements {
			tempGen.generateStatement(stmt)
		}

		var result string
		if returnType != "" {
			result = fmt.Sprintf("func(%s) %s {\n", params, returnType)
		} else {
			result = fmt.Sprintf("func(%s) {\n", params)
		}
		result += tempGen.output.String()
		for i := 0; i < g.indent; i++ {
			result += "\t"
		}
		result += "}"
		return result
	}

	// Shouldn't happen — at least one of Body or Block must be set
	return fmt.Sprintf("func(%s) {}", params)
}

// generateTypeParameters generates Go generic type parameter list
func (g *Generator) generateTypeParameters(typeParams []*TypeParameter) string {
	if len(typeParams) == 0 {
		return ""
	}

	parts := make([]string, len(typeParams))
	for i, tp := range typeParams {
		constraint := tp.Constraint
		if constraint == "cmp.Ordered" {
			g.addImport("cmp")
		}
		parts[i] = fmt.Sprintf("%s %s", tp.Name, constraint)
	}

	return "[" + strings.Join(parts, ", ") + "]"
}

func (g *Generator) generateFunctionParameters(params []*ast.Parameter) string {
	if len(params) == 0 {
		return ""
	}

	parts := make([]string, len(params))
	for i, param := range params {
		paramType := g.generateTypeAnnotation(param.Type)
		if param.Variadic {
			// Variadic parameter: use ...Type syntax
			parts[i] = fmt.Sprintf("%s ...%s", param.Name.Value, paramType)
		} else {
			parts[i] = fmt.Sprintf("%s %s", param.Name.Value, paramType)
		}
	}

	return strings.Join(parts, ", ")
}

func (g *Generator) generateParameters(params []*ast.Parameter) string {
	return g.generateFunctionParameters(params)
}

func (g *Generator) generateReturnTypes(returns []ast.TypeAnnotation) string {
	if len(returns) == 0 {
		return ""
	}

	if len(returns) == 1 {
		return g.generateTypeAnnotation(returns[0])
	}

	// Multiple return types
	parts := make([]string, len(returns))
	for i, ret := range returns {
		parts[i] = g.generateTypeAnnotation(ret)
	}

	return "(" + strings.Join(parts, ", ") + ")"
}

func (g *Generator) generateTypeAnnotation(typeAnn ast.TypeAnnotation) string {
	if typeAnn == nil {
		return ""
	}

	switch t := typeAnn.(type) {
	case *ast.PrimitiveType:
		if g.placeholderMap != nil {
			if typeParam, ok := g.placeholderMap[t.Name]; ok {
				return typeParam
			}
		}
		return t.Name
	case *ast.NamedType:
		if g.placeholderMap != nil {
			if typeParam, ok := g.placeholderMap[t.Name]; ok {
				return typeParam
			}
		}
		// Rewrite package-qualified type names if the package was auto-aliased
		if dotIdx := strings.Index(t.Name, "."); dotIdx > 0 {
			pkgPart := t.Name[:dotIdx]
			typePart := t.Name[dotIdx:]
			if alias, ok := g.pkgAliases[pkgPart]; ok {
				return alias + typePart
			}
		}
		// Special handling for iter.Seq in stdlib mode
		if g.isStdlibIter && g.placeholderMap != nil {
			if g.isIterSeqType(t) {
				// Transform iter.Seq → iter.Seq[T]
				if _, ok := g.placeholderMap["any"]; ok {
					typeParam := "T"
					// If this is a return type and U is declared, use U
					if g.processingReturnType {
						if _, hasU := g.placeholderMap["any2"]; hasU {
							typeParam = "U"
						}
					}
					return "iter.Seq[" + typeParam + "]"
				}
			}

			// iter.SeqU → iter.Seq[U]
			if t.Name == "iter.SeqU" {
				return "iter.Seq[U]"
			}

			// iter.Seq2 → iter.Seq2[T, U] or iter.Seq2[int, T] (for Enumerate) or iter.Seq2[T, T]
			if t.Name == "iter.Seq2" {
				if g.currentFuncName == "Enumerate" {
					return "iter.Seq2[int, T]"
				}
				// Only use U if it's actually declared as a type parameter
				if _, hasU := g.placeholderMap["any2"]; hasU {
					return "iter.Seq2[T, U]"
				}
				return "iter.Seq2[T, T]"
			}

			// iter.SeqSlice → iter.Seq[[]T] (for Chunk)
			if t.Name == "iter.SeqSlice" {
				return "iter.Seq[[]T]"
			}
		}
		return t.Name
	case *ast.ReferenceType:
		return "*" + g.generateTypeAnnotation(t.ElementType)
	case *ast.ListType:
		return "[]" + g.generateTypeAnnotation(t.ElementType)
	case *ast.MapType:
		keyType := g.generateTypeAnnotation(t.KeyType)
		valueType := g.generateTypeAnnotation(t.ValueType)
		return fmt.Sprintf("map[%s]%s", keyType, valueType)
		// Note: keyType and valueType already have placeholders substituted via recursion
	case *ast.ChannelType:
		return "chan " + g.generateTypeAnnotation(t.ElementType)
	case *ast.FunctionType:
		// Generate Go function type: func(params) returns
		var paramTypes []string
		for _, param := range t.Parameters {
			paramTypes = append(paramTypes, g.generateTypeAnnotation(param))
		}

		result := "func(" + strings.Join(paramTypes, ", ") + ")"

		if len(t.Returns) == 1 {
			result += " " + g.generateTypeAnnotation(t.Returns[0])
		} else if len(t.Returns) > 1 {
			var returnTypes []string
			for _, ret := range t.Returns {
				returnTypes = append(returnTypes, g.generateTypeAnnotation(ret))
			}
			result += " (" + strings.Join(returnTypes, ", ") + ")"
		}

		return result
	default:
		return "any"
	}
}
