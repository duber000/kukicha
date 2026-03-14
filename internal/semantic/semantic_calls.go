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
	a.checkDeprecatedCall(expr)

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

	// Always analyze positional arguments to ensure their types are recorded
	var providedArgTypes []*TypeInfo
	if pipedArg != nil && !hasPlaceholder {
		providedArgTypes = append(providedArgTypes, pipedArg)
	}
	for _, arg := range expr.Arguments {
		if hasPlaceholder && pipedArg != nil {
			if ident, ok := arg.(*ast.Identifier); ok && ident.Value == "_" {
				// Placeholder takes the piped arg's type and records it in exprTypes
				providedArgTypes = append(providedArgTypes, pipedArg)
				a.recordType(ident, pipedArg)
				continue
			}
		}
		providedArgTypes = append(providedArgTypes, a.analyzeExpression(arg))
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
				if argType.Kind == TypeKindList && argType.ElementType != nil {
					// list of T spread into ...T — check element type
					if !a.typesCompatible(variadicParamType, argType.ElementType) {
						a.error(expr.Pos(), fmt.Sprintf("argument %d: cannot use %s as []%s in variadic spread", i+1, argType, variadicParamType))
					}
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

// goStdlibEntryToTypeInfos converts a generated Go stdlib entry to a slice of TypeInfo.
func goStdlibEntryToTypeInfos(entry goStdlibEntry) []*TypeInfo {
	types := make([]*TypeInfo, entry.Count)
	for i, gt := range entry.Types {
		types[i] = &TypeInfo{Kind: gt.Kind, Name: gt.Name}
	}
	for i := len(entry.Types); i < entry.Count; i++ {
		types[i] = &TypeInfo{Kind: TypeKindUnknown}
	}
	return types
}

func (a *Analyzer) analyzeMethodCallExpr(expr *ast.MethodCallExpr, pipedArg *TypeInfo) []*TypeInfo {
	// Analyze object
	objType := a.analyzeExpression(expr.Object)

	// Analyze named arguments
	for _, namedArg := range expr.NamedArguments {
		a.analyzeExpression(namedArg.Value)
	}

	// Analyze arguments
	for _, arg := range expr.Arguments {
		a.analyzeExpression(arg)
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
			types := goStdlibEntryToTypeInfos(entry)
			a.recordReturnCount(expr, entry.Count)
			return types
		}

		// Fall back to Kukicha stdlib registry (now has per-position type info)
		if entry, ok := generatedStdlibRegistry[qualifiedName]; ok {
			types := goStdlibEntryToTypeInfos(entry)
			a.recordReturnCount(expr, entry.Count)
			return types
		}
	}

	// time.Time methods with known return types
	if objType != nil && objType.Kind == TypeKindNamed && objType.Name == "time.Time" {
		switch methodName {
		case "Equal", "Before", "After":
			types := []*TypeInfo{{Kind: TypeKindBool}}
			a.recordReturnCount(expr, len(types))
			return types
		case "Year":
			types := []*TypeInfo{{Kind: TypeKindInt}}
			a.recordReturnCount(expr, len(types))
			return types
		case "Month":
			types := []*TypeInfo{{Kind: TypeKindNamed, Name: "time.Month"}}
			a.recordReturnCount(expr, len(types))
			return types
		case "Day", "Hour", "Minute", "Second":
			types := []*TypeInfo{{Kind: TypeKindInt}}
			a.recordReturnCount(expr, len(types))
			return types
		case "Weekday":
			types := []*TypeInfo{{Kind: TypeKindNamed, Name: "time.Weekday"}}
			a.recordReturnCount(expr, len(types))
			return types
		}
	}

	// exec.ExitError methods used by typed piped switch result inference
	if objType != nil && objType.Kind == TypeKindNamed {
		if objType.Name == "exec.ExitError" || objType.Name == "*exec.ExitError" {
			switch methodName {
			case "ExitCode":
				types := []*TypeInfo{{Kind: TypeKindInt}}
				a.recordReturnCount(expr, len(types))
				return types
			case "Error":
				types := []*TypeInfo{{Kind: TypeKindString}}
				a.recordReturnCount(expr, len(types))
				return types
			}
		}
	}

	// bufio.Scanner methods (needed for SSE streaming in llm.kuki)
	if objType != nil && objType.Kind == TypeKindNamed {
		if objType.Name == "bufio.Scanner" || objType.Name == "*bufio.Scanner" {
			switch methodName {
			case "Scan":
				types := []*TypeInfo{{Kind: TypeKindBool}}
				a.recordReturnCount(expr, len(types))
				return types
			case "Text":
				types := []*TypeInfo{{Kind: TypeKindString}}
				a.recordReturnCount(expr, len(types))
				return types
			case "Bytes":
				types := []*TypeInfo{{Kind: TypeKindList, ElementType: &TypeInfo{Kind: TypeKindNamed, Name: "byte"}}}
				a.recordReturnCount(expr, len(types))
				return types
			case "Err":
				types := []*TypeInfo{{Kind: TypeKindNamed, Name: "error"}}
				a.recordReturnCount(expr, len(types))
				return types
			}
		}
	}

	// regexp.Regexp instance methods
	if objType != nil && objType.Kind == TypeKindReference &&
		(objType.Name == "*regexp.Regexp" || objType.Name == "regexp.Regexp") {
		switch methodName {
		case "MatchString", "Match", "MatchReader":
			types := []*TypeInfo{{Kind: TypeKindBool}, {Kind: TypeKindNamed, Name: "error"}}
			a.recordReturnCount(expr, len(types))
			return types
		case "FindString", "ReplaceAllString", "ReplaceAllLiteralString", "String":
			types := []*TypeInfo{{Kind: TypeKindString}}
			a.recordReturnCount(expr, 1)
			return types
		case "FindStringSubmatch", "FindAllString":
			types := []*TypeInfo{{Kind: TypeKindList, ElementType: &TypeInfo{Kind: TypeKindString}}}
			a.recordReturnCount(expr, 1)
			return types
		}
	}

	// Field access (no parentheses): resolve struct field type
	if !expr.IsCall && objType != nil {
		fieldType := a.resolveFieldType(objType, methodName)
		if fieldType != nil {
			a.recordReturnCount(expr, 1)
			a.recordType(expr, fieldType)
			return []*TypeInfo{fieldType}
		}
	}

	// Method call on user-defined type: look up method signature
	if expr.IsCall && objType != nil {
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

// checkDeprecatedCall emits a warning if the called function is marked # kuki:deprecated.
func (a *Analyzer) checkDeprecatedCall(expr *ast.CallExpr) {
	var name, qualifiedName string
	switch fn := expr.Function.(type) {
	case *ast.Identifier:
		name = fn.Value
	case *ast.MethodCallExpr:
		name = fn.Method.Value
		// Build qualified name for stdlib lookup (e.g., "slice.Filter")
		if objID, ok := fn.Object.(*ast.Identifier); ok {
			qualifiedName = objID.Value + "." + fn.Method.Value
		}
	default:
		return
	}

	// Check local deprecated functions (from same-file directives)
	if msg, ok := a.deprecatedFuncs[name]; ok {
		a.warn(expr.Pos(), fmt.Sprintf("'%s' is deprecated: %s", name, msg))
		return
	}

	// Check stdlib deprecated functions (from generated registry)
	if qualifiedName != "" {
		if msg, ok := generatedStdlibDeprecated[qualifiedName]; ok {
			a.warn(expr.Pos(), fmt.Sprintf("'%s' is deprecated: %s", qualifiedName, msg))
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
