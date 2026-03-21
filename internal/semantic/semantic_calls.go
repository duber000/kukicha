package semantic

import (
	"fmt"

	"github.com/duber000/kukicha/internal/ast"
)

// knownExternalReturns maps qualified function names to their return count.
// Built once from the auto-generated Kukicha stdlib registry (generatedStdlibRegistry)
// plus the auto-generated Go stdlib registry (generatedGoStdlib).
var knownExternalReturns map[string]int

func init() {
	knownExternalReturns = make(map[string]int, len(generatedStdlibRegistry)+len(generatedGoStdlib))
	for k, v := range generatedStdlibRegistry {
		knownExternalReturns[k] = v.Count
	}
	for k, v := range generatedGoStdlib {
		knownExternalReturns[k] = v.Count
	}
}

// hasUntypedParams reports whether the lambda has at least one parameter with no type annotation.
func hasUntypedParams(lambda *ast.ArrowLambda) bool {
	for _, param := range lambda.Parameters {
		if param.Type == nil {
			return true
		}
	}
	return false
}

// extractQualifiedName returns "pkg.Func" for a MethodCallExpr, or the identifier
// value for a plain Identifier.
func extractQualifiedName(fn ast.Expression) string {
	switch f := fn.(type) {
	case *ast.MethodCallExpr:
		if objID, ok := f.Object.(*ast.Identifier); ok {
			return objID.Value + "." + f.Method.Value
		}
	case *ast.Identifier:
		return f.Value
	}
	return ""
}

// resolveExpectedLambdaParams returns the expected []*TypeInfo for the j-th
// parameter of the lambda that appears at paramIdx in the calling function's
// signature. Three inference cases are handled:
//
//   - Case A: user-defined function with TypeKindFunction at paramIdx — uses
//     funcType.Params[paramIdx].Params directly.
//   - Cases B & C: Kukicha stdlib registry with ParamFuncParams entry for
//     paramIdx — substitutes placeholder names ("any", "any2", "ordered")
//     with elementType, and uses concrete names (e.g. "cli.Args") as-is.
func (a *Analyzer) resolveExpectedLambdaParams(
	qualName string, paramIdx int, funcType *TypeInfo, elementType *TypeInfo,
) []*TypeInfo {
	// Case A: user-defined function with a known func-typed parameter
	if funcType != nil && funcType.Kind == TypeKindFunction && paramIdx < len(funcType.Params) {
		paramType := funcType.Params[paramIdx]
		if paramType != nil && paramType.Kind == TypeKindFunction && len(paramType.Params) > 0 {
			return paramType.Params
		}
	}

	// Cases B & C: look up in the Kukicha stdlib registry
	if entry, ok := generatedStdlibRegistry[qualName]; ok && len(entry.ParamFuncParams) > 0 {
		if innerTypes, ok := entry.ParamFuncParams[paramIdx]; ok {
			result := make([]*TypeInfo, len(innerTypes))
			hasAny := false
			for j, gt := range innerTypes {
				switch {
				case gt.Kind == TypeKindNamed && (gt.Name == "any" || gt.Name == "any2" || gt.Name == "ordered" || gt.Name == "result"):
					if elementType != nil {
						result[j] = elementType
						hasAny = true
					}
				case gt.Kind != TypeKindUnknown:
					result[j] = &TypeInfo{Kind: gt.Kind, Name: gt.Name}
					hasAny = true
				}
			}
			if hasAny {
				return result
			}
		}
	}

	return nil
}

// inferLambdaParamTypes records inferred parameter types in exprTypes for untyped
// arrow lambda parameters appearing in a CallExpr's argument list.
// This enables codegen to emit typed Go func signatures without requiring the
// user to annotate every lambda parameter.
func (a *Analyzer) inferLambdaParamTypes(
	expr *ast.CallExpr,
	funcType *TypeInfo,
	providedArgTypes []*TypeInfo,
	pipedArg *TypeInfo,
	hasPlaceholder bool,
) {
	qualName := extractQualifiedName(expr.Function)

	// Resolve element type T (for generic Case B / list-element substitution)
	var elementType *TypeInfo
	if pipedArg != nil && !hasPlaceholder && pipedArg.ElementType != nil {
		elementType = pipedArg.ElementType
	} else if len(providedArgTypes) > 0 && providedArgTypes[0] != nil && providedArgTypes[0].ElementType != nil {
		elementType = providedArgTypes[0].ElementType
	}

	argOffset := 0
	if pipedArg != nil && !hasPlaceholder {
		argOffset = 1 // piped arg occupies parameter index 0
	}

	for i, arg := range expr.Arguments {
		lambda, ok := arg.(*ast.ArrowLambda)
		if !ok || !hasUntypedParams(lambda) {
			continue
		}
		paramIdx := i + argOffset

		inferredParamTypes := a.resolveExpectedLambdaParams(qualName, paramIdx, funcType, elementType)
		for j, param := range lambda.Parameters {
			if param.Type == nil && j < len(inferredParamTypes) && inferredParamTypes[j] != nil {
				a.recordType(param.Name, inferredParamTypes[j])
			}
		}
	}
}

// inferLambdaParamTypesMethod records inferred parameter types for untyped arrow
// lambda parameters in a MethodCallExpr's argument list.
func (a *Analyzer) inferLambdaParamTypesMethod(expr *ast.MethodCallExpr, pipedArg *TypeInfo) {
	qualName := ""
	if objID, ok := expr.Object.(*ast.Identifier); ok {
		qualName = objID.Value + "." + expr.Method.Value
	}
	if qualName == "" {
		return
	}

	// Resolve element type from pipedArg, or from the first non-lambda arg in exprTypes
	var elementType *TypeInfo
	if pipedArg != nil && pipedArg.ElementType != nil {
		elementType = pipedArg.ElementType
	} else {
		for _, arg := range expr.Arguments {
			if _, ok := arg.(*ast.ArrowLambda); ok {
				continue
			}
			if ti, ok := a.exprTypes[arg]; ok && ti != nil && ti.ElementType != nil {
				elementType = ti.ElementType
				break
			}
		}
	}

	argOffset := 0
	if pipedArg != nil {
		argOffset = 1 // piped arg occupies parameter index 0
	}

	for i, arg := range expr.Arguments {
		lambda, ok := arg.(*ast.ArrowLambda)
		if !ok || !hasUntypedParams(lambda) {
			continue
		}
		paramIdx := i + argOffset

		// funcType is unknown for stdlib method calls, so only Cases B & C apply
		inferredParamTypes := a.resolveExpectedLambdaParams(qualName, paramIdx, nil, elementType)
		for j, param := range lambda.Parameters {
			if param.Type == nil && j < len(inferredParamTypes) && inferredParamTypes[j] != nil {
				a.recordType(param.Name, inferredParamTypes[j])
			}
		}
	}
}

func (a *Analyzer) analyzeCallExpr(expr *ast.CallExpr, pipedArg *TypeInfo) []*TypeInfo {
	// Check for known Go stdlib functions (parsed as direct Identifier, e.g. os.LookupEnv)
	if id, ok := expr.Function.(*ast.Identifier); ok {
		if entry, ok := generatedGoStdlib[id.Value]; ok {
			types := goStdlibEntryToTypeInfos(entry)
			a.recordReturnCount(expr, entry.Count)
			return types
		}
	}

	// Check for known Go stdlib functions parsed as MethodCallExpr (pkg.Func form)
	if methodCall, ok := expr.Function.(*ast.MethodCallExpr); ok {
		if objID, ok := methodCall.Object.(*ast.Identifier); ok {
			qualifiedName := objID.Value + "." + methodCall.Method.Value
			if entry, ok := generatedGoStdlib[qualifiedName]; ok {
				types := goStdlibEntryToTypeInfos(entry)
				a.recordReturnCount(expr, entry.Count)
				return types
			}
		}
	}

	// Check for deprecated function calls
	if id, ok := expr.Function.(*ast.Identifier); ok {
		a.checkDeprecated(expr, id.Value, "")
		a.checkPanics(expr, id.Value, "")
	} else if methodCall, ok := expr.Function.(*ast.MethodCallExpr); ok {
		// Handles things like obj.method()() if that ever occurs, 
		// though typically pkg.Func() parses directly to MethodCallExpr
		if objID, ok := methodCall.Object.(*ast.Identifier); ok {
			qualifiedName := objID.Value + "." + methodCall.Method.Value
			a.checkDeprecated(expr, methodCall.Method.Value, qualifiedName)
			a.checkPanics(expr, methodCall.Method.Value, qualifiedName)
		}
	}

	funcType := a.analyzeExpression(expr.Function)

	// Analyze named arguments (check for duplicates)
	namedArgNames := make(map[string]bool)
	for _, namedArg := range expr.NamedArguments {
		if namedArgNames[namedArg.Name.Value] {
			a.error(namedArg.Pos(), fmt.Sprintf("duplicate named argument: %s", namedArg.Name.Value))
		}
		namedArgNames[namedArg.Name.Value] = true
		a.analyzeExpression(namedArg.Value)
	}

	// Check if any argument is a "_" placeholder (piped value position marker).
	// When present, the piped arg occupies the placeholder slot — don't double-count.
	hasPlaceholder := false
	for _, arg := range expr.Arguments {
		if ident, ok := arg.(*ast.Identifier); ok && ident.Value == "_" {
			hasPlaceholder = true
			break
		}
	}

	// Analyze non-lambda arguments first so their types are available for inference.
	var providedArgTypes []*TypeInfo
	if pipedArg != nil && !hasPlaceholder {
		providedArgTypes = append(providedArgTypes, pipedArg)
	}
	lambdaIndices := make(map[int]bool)
	for i, arg := range expr.Arguments {
		if hasPlaceholder && pipedArg != nil {
			if ident, ok := arg.(*ast.Identifier); ok && ident.Value == "_" {
				providedArgTypes = append(providedArgTypes, pipedArg)
				a.recordType(ident, pipedArg)
				continue
			}
		}
		if _, isLambda := arg.(*ast.ArrowLambda); isLambda {
			lambdaIndices[i] = true
			providedArgTypes = append(providedArgTypes, &TypeInfo{Kind: TypeKindUnknown})
		} else {
			providedArgTypes = append(providedArgTypes, a.analyzeExpression(arg))
		}
	}

	// Infer lambda param types before analyzing lambda bodies, so that
	// parameters have their types in scope during body analysis.
	a.inferLambdaParamTypes(expr, funcType, providedArgTypes, pipedArg, hasPlaceholder)

	// Now analyze lambda arguments — params are already typed from inference.
	for i, arg := range expr.Arguments {
		if lambdaIndices[i] {
			offset := 0
			if pipedArg != nil && !hasPlaceholder {
				offset = 1
			}
			providedArgTypes[i+offset] = a.analyzeExpression(arg)
		}
	}

	// Validate usage of named arguments
	if len(expr.NamedArguments) > 0 {
		if funcType.Kind != TypeKindFunction {
			// Check if the Kukicha stdlib registry has parameter names for this function
			qualifiedName := ""
			if methodCall, ok := expr.Function.(*ast.MethodCallExpr); ok {
				if objID, ok := methodCall.Object.(*ast.Identifier); ok {
					qualifiedName = objID.Value + "." + methodCall.Method.Value
				}
			}
			if entry, ok := generatedStdlibRegistry[qualifiedName]; ok && len(entry.ParamNames) > 0 {
				// Stdlib function with known param names — validate named args
				a.validateNamedArgs(expr, entry.ParamNames)
			} else {
				name := "function"
				if id, ok := expr.Function.(*ast.Identifier); ok {
					name = fmt.Sprintf("function '%s'", id.Value)
				}
				a.error(expr.Pos(), fmt.Sprintf("named arguments are not supported for imported or unknown %s (please use positional arguments)", name))
			}
		}
	}

	// If it's a known function, validate arguments
	if funcType.Kind == TypeKindFunction {
		// Validate argument count
		totalProvidedArgs := len(expr.Arguments) + len(expr.NamedArguments)
		if pipedArg != nil && !hasPlaceholder {
			totalProvidedArgs++
		}

		// Calculate required arguments (parameters without defaults)
		requiredParams := len(funcType.Params)
		if funcType.DefaultCount > 0 {
			requiredParams = len(funcType.Params) - funcType.DefaultCount
		}

		if funcType.Variadic {
			if expr.Variadic {
				// Spreading a slice into variadic: f(many args)
				// The spread argument replaces the entire variadic portion,
				// so we need at least (non-variadic params) + 1 (the spread) arguments.
				nonVariadicParams := len(funcType.Params) - 1
				if totalProvidedArgs < nonVariadicParams+1 {
					a.error(expr.Pos(), fmt.Sprintf("expected at least %d arguments, got %d", nonVariadicParams+1, totalProvidedArgs))
				}
			} else {
				// Variadic: must have at least (required params - 1) arguments
				minArgs := max(requiredParams-1, 0)
				if totalProvidedArgs < minArgs {
					a.error(expr.Pos(), fmt.Sprintf("expected at least %d arguments, got %d", minArgs, totalProvidedArgs))
				}
			}
		} else {
			// Non-variadic: must have between required and total params
			if totalProvidedArgs < requiredParams {
				a.error(expr.Pos(), fmt.Sprintf("expected at least %d arguments, got %d", requiredParams, totalProvidedArgs))
			}
			if totalProvidedArgs > len(funcType.Params) {
				a.error(expr.Pos(), fmt.Sprintf("expected at most %d arguments, got %d", len(funcType.Params), totalProvidedArgs))
			}
		}

		// Validate positional argument types
		for i, argType := range providedArgTypes {
			// For variadic, all args beyond params-1 match the last param type
			paramIndex := i
			if funcType.Variadic && i >= len(funcType.Params)-1 {
				paramIndex = len(funcType.Params) - 1
			}

			// When spreading a slice (expr.Variadic), the last argument is a
			// slice being unpacked. Check that its element type matches the
			// variadic parameter type instead of comparing directly.
			if expr.Variadic && funcType.Variadic && paramIndex == len(funcType.Params)-1 && i == len(providedArgTypes)-1 {
				variadicParamType := funcType.Params[paramIndex]
				if argType.Kind == TypeKindList {
					if argType.ElementType != nil {
						// list of T spread into ...T — check element type
						if !a.typesCompatible(variadicParamType, argType.ElementType) {
							a.error(expr.Pos(), fmt.Sprintf("argument %d: cannot use %s as []%s in variadic spread", i+1, argType, variadicParamType))
						}
					}
					// If ElementType is nil, we can't check — be lenient
				} else if argType.Kind != TypeKindUnknown {
					// Not a list — could still be valid for interface{} params or unknown types
					if !a.typesCompatible(variadicParamType, argType) {
						a.error(expr.Pos(), fmt.Sprintf("argument %d: cannot use %s as %s", i+1, argType, variadicParamType))
					}
				}
				continue
			}

			if paramIndex < len(funcType.Params) && !a.typesCompatible(funcType.Params[paramIndex], argType) {
				a.error(expr.Pos(), fmt.Sprintf("argument %d: cannot use %s as %s", i+1, argType, funcType.Params[paramIndex]))
			}
		}

		// Record expected param types on pipe placeholder "_" arguments
		// so that exprTypes contains typed info rather than TypeKindUnknown.
		a.typePlaceholderArgs(expr, funcType, pipedArg != nil && !hasPlaceholder)

		// Return all return types
		if len(funcType.Returns) > 0 {
			a.recordReturnCount(expr, len(funcType.Returns))
			return funcType.Returns
		}
		a.recordReturnCount(expr, 0)
	}

	return []*TypeInfo{{Kind: TypeKindUnknown}}
}

// goStdlibTypeToTypeInfo converts a goStdlibType to a TypeInfo, including nested
// element/key/value types for lists and maps.
func goStdlibTypeToTypeInfo(gt goStdlibType) *TypeInfo {
	ti := &TypeInfo{Kind: gt.Kind, Name: gt.Name}
	if gt.ElementType != nil {
		ti.ElementType = goStdlibTypeToTypeInfo(*gt.ElementType)
	}
	if gt.KeyType != nil {
		ti.KeyType = goStdlibTypeToTypeInfo(*gt.KeyType)
	}
	if gt.ValueType != nil {
		ti.ValueType = goStdlibTypeToTypeInfo(*gt.ValueType)
	}
	return ti
}

// goStdlibEntryToTypeInfos converts a generated Go stdlib entry to a slice of TypeInfo.
func goStdlibEntryToTypeInfos(entry goStdlibEntry) []*TypeInfo {
	types := make([]*TypeInfo, entry.Count)
	for i, gt := range entry.Types {
		types[i] = goStdlibTypeToTypeInfo(gt)
	}
	for i := len(entry.Types); i < entry.Count; i++ {
		types[i] = &TypeInfo{Kind: TypeKindUnknown}
	}
	return types
}

// isPlaceholderType returns true if the type is a generic placeholder (any, any2, ordered, result).
func isPlaceholderType(ti *TypeInfo) bool {
	return ti != nil && ti.Kind == TypeKindNamed &&
		(ti.Name == "any" || ti.Name == "any2" || ti.Name == "ordered" || ti.Name == "result")
}

// resolveGenericPlaceholders resolves placeholder element types in return types
// using the actual call-site arguments. For example, if a function returns
// "list of result" and a lambda argument returns RepoEntry, the element type
// is resolved to RepoEntry.
func resolveGenericPlaceholders(types []*TypeInfo, argTypes []*TypeInfo, pipedArg *TypeInfo) {
	// Resolve "any" placeholder from the first list argument's element type
	var anyType *TypeInfo
	if pipedArg != nil && pipedArg.ElementType != nil && !isPlaceholderType(pipedArg.ElementType) {
		anyType = pipedArg.ElementType
	} else {
		for _, at := range argTypes {
			if at != nil && at.Kind == TypeKindList && at.ElementType != nil && !isPlaceholderType(at.ElementType) {
				anyType = at.ElementType
				break
			}
		}
	}

	// Resolve "result" placeholder from lambda argument return types
	var resultType *TypeInfo
	for _, at := range argTypes {
		if at != nil && at.Kind == TypeKindFunction && len(at.Returns) > 0 {
			ret := at.Returns[0]
			if ret != nil && ret.Kind != TypeKindUnknown && !isPlaceholderType(ret) {
				resultType = ret
				break
			}
		}
	}

	// Apply resolutions to return types
	for _, ti := range types {
		if ti == nil {
			continue
		}
		if isPlaceholderType(ti.ElementType) {
			switch ti.ElementType.Name {
			case "any", "any2", "ordered":
				if anyType != nil {
					ti.ElementType = anyType
				}
			case "result":
				if resultType != nil {
					ti.ElementType = resultType
				}
			}
		}
	}
}

func (a *Analyzer) analyzeMethodCallExpr(expr *ast.MethodCallExpr, pipedArg *TypeInfo) []*TypeInfo {
	// Analyze object
	objType := pipedArg
	if expr.Object != nil {
		objType = a.analyzeExpression(expr.Object)
	}

	// Analyze named arguments
	for _, namedArg := range expr.NamedArguments {
		a.analyzeExpression(namedArg.Value)
	}

	// Analyze non-lambda arguments first so their types are available for inference.
	argTypes := make([]*TypeInfo, len(expr.Arguments))
	for i, arg := range expr.Arguments {
		if _, isLambda := arg.(*ast.ArrowLambda); !isLambda {
			argTypes[i] = a.analyzeExpression(arg)
		}
	}

	// Infer lambda param types before analyzing lambda bodies, so that
	// parameters have their types in scope during body analysis (e.g.,
	// e.name resolves correctly when e is inferred as RepoEntry).
	a.inferLambdaParamTypesMethod(expr, pipedArg)

	// Now analyze lambda arguments — params are already typed from inference.
	for i, arg := range expr.Arguments {
		if _, isLambda := arg.(*ast.ArrowLambda); isLambda {
			argTypes[i] = a.analyzeExpression(arg)
		}
	}

	// Handle known stdlib method return types
	methodName := expr.Method.Value

	// Known package-level functions parsed as MethodCallExpr (e.g., os.LookupEnv, fetch.Get)
	if objID, ok := expr.Object.(*ast.Identifier); ok {
		qualifiedName := objID.Value + "." + methodName

		// Security: detect string interpolation in SQL query arguments
		a.checkSQLInterpolation(qualifiedName, expr, pipedArg)

		// Security: detect XSS via http.HTML with non-literal content
		a.checkHTMLNonLiteral(qualifiedName, expr, pipedArg)

		// Security: detect fetch.Get/Post/New inside HTTP handlers (SSRF risk)
		a.checkFetchInHandler(qualifiedName, expr)

		// Security: detect files.* inside HTTP handlers (path traversal risk)
		a.checkFilesInHandler(qualifiedName, expr)

		// Security: detect shell.Run with non-literal argument (command injection)
		a.checkShellRunNonLiteral(qualifiedName, expr, pipedArg)

		// Security: detect http.Redirect with non-literal URL (open redirect)
		a.checkRedirectNonLiteral(qualifiedName, expr, pipedArg)

		// Check generated Go stdlib registry first (has full type info)
		if entry, ok := generatedGoStdlib[qualifiedName]; ok {
			a.checkDeprecated(expr, methodName, qualifiedName)
			a.checkPanics(expr, methodName, qualifiedName)

			types := goStdlibEntryToTypeInfos(entry)
			a.recordReturnCount(expr, entry.Count)
			return types
		}

		// Fall back to Kukicha stdlib registry (now has per-position type info)
		if entry, ok := generatedStdlibRegistry[qualifiedName]; ok {
			a.checkDeprecated(expr, methodName, qualifiedName)
			a.checkPanics(expr, methodName, qualifiedName)

			types := goStdlibEntryToTypeInfos(entry)
			resolveGenericPlaceholders(types, argTypes, pipedArg)
			a.recordReturnCount(expr, entry.Count)
			return types
		}
	}

	// Check generated Go stdlib registry for method calls
	if objType != nil && (objType.Kind == TypeKindNamed || objType.Kind == TypeKindReference) && objType.Name != "" {
		qualifiedMethodName := objType.Name + "." + methodName
		
		a.checkDeprecated(expr, methodName, qualifiedMethodName)
		a.checkPanics(expr, methodName, qualifiedMethodName)

		if entry, ok := generatedGoStdlib[qualifiedMethodName]; ok {
			types := goStdlibEntryToTypeInfos(entry)
			a.recordReturnCount(expr, entry.Count)
			return types
		}
	}

	// Method call on user-defined type: look up method signature
	if objType != nil {
		if objType.Kind == TypeKindNamed || objType.Kind == TypeKindStruct {
			qualifiedMethodName := objType.Name + "." + methodName
			a.checkDeprecated(expr, methodName, qualifiedMethodName)
			a.checkPanics(expr, methodName, qualifiedMethodName)
		} else {
			a.checkDeprecated(expr, methodName, "")
			a.checkPanics(expr, methodName, "")
		}

		methodType := a.resolveMethodType(objType, methodName)
		if methodType != nil && len(methodType.Returns) > 0 {
			a.recordReturnCount(expr, len(methodType.Returns))
			return methodType.Returns
		}
	}

	// Return count of 1 gives codegen's onerr path a safe default.
	a.recordReturnCount(expr, 1)
	return []*TypeInfo{{Kind: TypeKindUnknown}}
}

func (a *Analyzer) analyzeFieldAccessExpr(expr *ast.FieldAccessExpr, pipedArg *TypeInfo) *TypeInfo {
	objType := pipedArg
	if expr.Object != nil {
		objType = a.analyzeExpression(expr.Object)
	}

	if objType != nil {
		fieldType := a.resolveFieldType(objType, expr.Field.Value)
		if fieldType != nil {
			a.recordReturnCount(expr, 1)
			return fieldType
		}
	}

	a.recordReturnCount(expr, 1)
	return &TypeInfo{Kind: TypeKindUnknown}
}

// checkDeprecated emits a warning if the called function is marked # kuki:deprecated.
func (a *Analyzer) checkDeprecated(node ast.Node, name string, qualifiedName string) {
	// Check local deprecated functions (from same-file directives)
	if msg, ok := a.deprecatedFuncs[name]; ok {
		a.warn(node.Pos(), fmt.Sprintf("'%s' is deprecated: %s", name, msg))
		return
	}

	// Check stdlib deprecated functions (from generated registry)
	if qualifiedName != "" {
		if msg, ok := generatedStdlibDeprecated[qualifiedName]; ok {
			a.warn(node.Pos(), fmt.Sprintf("'%s' is deprecated: %s", qualifiedName, msg))
		}
	}
}

// checkPanics emits a warning if the called function is marked # kuki:panics.
func (a *Analyzer) checkPanics(node ast.Node, name string, qualifiedName string) {
	// Check local panicking functions (from same-file directives)
	if msg, ok := a.panickedFuncs[name]; ok {
		a.warn(node.Pos(), fmt.Sprintf("%s may panic: %q", name, msg))
		return
	}

	// Check stdlib panicking functions (from generated registry)
	if qualifiedName != "" {
		if msg, ok := generatedStdlibPanics[qualifiedName]; ok {
			a.warn(node.Pos(), fmt.Sprintf("%s may panic: %q", qualifiedName, msg))
		}
	}
}

// resolveFieldType looks up a field's type on a struct or reference-to-struct type.
func (a *Analyzer) resolveFieldType(objType *TypeInfo, fieldName string) *TypeInfo {
	typeInfo := objType
	// Dereference pointer/reference types
	if typeInfo.Kind == TypeKindReference && typeInfo.ElementType != nil {
		typeInfo = typeInfo.ElementType
	}

	// Look up in struct fields
	if typeInfo.Fields != nil {
		if fieldType, ok := typeInfo.Fields[fieldName]; ok {
			return fieldType
		}
	}

	// Try resolving via the symbol table (for named types)
	if typeInfo.Kind == TypeKindNamed || typeInfo.Kind == TypeKindStruct {
		name := typeInfo.Name
		if sym := a.symbolTable.Resolve(name); sym != nil && sym.Type != nil && sym.Type.Fields != nil {
			if fieldType, ok := sym.Type.Fields[fieldName]; ok {
				return fieldType
			}
		}
	}

	return nil
}

// resolveMethodType looks up a method's function type on a struct type.
func (a *Analyzer) resolveMethodType(objType *TypeInfo, methodName string) *TypeInfo {
	typeInfo := objType
	// Dereference pointer/reference types
	if typeInfo.Kind == TypeKindReference && typeInfo.ElementType != nil {
		typeInfo = typeInfo.ElementType
	}

	// Look up in type's methods
	if typeInfo.Methods != nil {
		if methodType, ok := typeInfo.Methods[methodName]; ok {
			return methodType
		}
	}

	// Try resolving via the symbol table (for named types)
	if typeInfo.Kind == TypeKindNamed || typeInfo.Kind == TypeKindStruct {
		name := typeInfo.Name
		if sym := a.symbolTable.Resolve(name); sym != nil && sym.Type != nil && sym.Type.Methods != nil {
			if methodType, ok := sym.Type.Methods[methodName]; ok {
				return methodType
			}
		}
	}

	return nil
}

// typePlaceholderArgs records expected parameter types on "_" placeholder
// arguments so exprTypes contains typed info rather than TypeKindUnknown.
func (a *Analyzer) typePlaceholderArgs(expr *ast.CallExpr, funcType *TypeInfo, hasPipedArg bool) {
	offset := 0
	if hasPipedArg {
		offset = 1 // piped arg occupies the first parameter position
	}
	for i, arg := range expr.Arguments {
		ident, ok := arg.(*ast.Identifier)
		if !ok || ident.Value != "_" {
			continue
		}
		paramIndex := i + offset
		if funcType.Variadic && paramIndex >= len(funcType.Params)-1 {
			paramIndex = len(funcType.Params) - 1
		}
		if paramIndex < len(funcType.Params) {
			paramType := funcType.Params[paramIndex]
			if paramType.Kind != TypeKindUnknown {
				a.recordType(ident, paramType)
			}
		}
	}
}

// validateNamedArgs checks that named arguments match known parameter names.
func (a *Analyzer) validateNamedArgs(expr *ast.CallExpr, paramNames []string) {
	paramSet := make(map[string]bool, len(paramNames))
	for _, name := range paramNames {
		paramSet[name] = true
	}
	for _, namedArg := range expr.NamedArguments {
		if !paramSet[namedArg.Name.Value] {
			a.error(namedArg.Pos(), fmt.Sprintf("unknown parameter name '%s'", namedArg.Name.Value))
		}
	}
}
