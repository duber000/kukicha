package semantic

import (
	"fmt"

	"github.com/duber000/kukicha/internal/ast"
)

func (a *Analyzer) analyzeCallExpr(expr *ast.CallExpr, pipedArg *TypeInfo) []*TypeInfo {
	// Check for known stdlib functions first
	if id, ok := expr.Function.(*ast.Identifier); ok {
		switch id.Value {
		case "os.LookupEnv":
			types := []*TypeInfo{
				{Kind: TypeKindString},
				{Kind: TypeKindBool},
			}
			a.recordReturnCount(expr, len(types))
			return types
		}
	}

	// Check for known stdlib method calls (e.g., os.LookupEnv)
	// This might be parsed as a MethodCallExpr in some cases
	if methodCall, ok := expr.Function.(*ast.MethodCallExpr); ok {
		if objID, ok := methodCall.Object.(*ast.Identifier); ok {
			methodName := methodCall.Method.Value
			qualifiedName := objID.Value + "." + methodName
			switch qualifiedName {
			case "os.LookupEnv":
				types := []*TypeInfo{
					{Kind: TypeKindString},
					{Kind: TypeKindBool},
				}
				a.recordReturnCount(expr, len(types))
				return types
			// bufio package functions
			case "bufio.NewScanner":
				types := []*TypeInfo{{Kind: TypeKindNamed, Name: "bufio.Scanner"}}
				a.recordReturnCount(expr, len(types))
				return types
			}
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

	// Validate usage of named arguments (only supported for local functions)
	if len(expr.NamedArguments) > 0 {
		if funcType.Kind != TypeKindFunction {
			name := "function"
			if id, ok := expr.Function.(*ast.Identifier); ok {
				name = fmt.Sprintf("function '%s'", id.Value)
			}
			a.error(expr.Pos(), fmt.Sprintf("named arguments are not supported for imported or unknown %s (please use positional arguments)", name))
		}
	}

	// If it's a known function, validate arguments
	if funcType.Kind == TypeKindFunction {
		// Validate argument count
		totalProvidedArgs := len(expr.Arguments) + len(expr.NamedArguments)
		if pipedArg != nil {
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

		// Collect all provided argument types in order
		var providedArgTypes []*TypeInfo
		if pipedArg != nil {
			providedArgTypes = append(providedArgTypes, pipedArg)
		}
		for _, arg := range expr.Arguments {
			providedArgTypes = append(providedArgTypes, a.analyzeExpression(arg))
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

		// Named arguments validation would require parameter name information
		// which is tracked in ParamNames field

		// Return all return types
		if len(funcType.Returns) > 0 {
			a.recordReturnCount(expr, len(funcType.Returns))
			return funcType.Returns
		}
		a.recordReturnCount(expr, 0)
	}

	return []*TypeInfo{{Kind: TypeKindUnknown}}
}

func (a *Analyzer) analyzeMethodCallExpr(expr *ast.MethodCallExpr, pipedArg *TypeInfo) []*TypeInfo {
	// Analyze object
	objType := a.analyzeExpression(expr.Object)

	// Method support is currently limited to positional arguments
	if len(expr.NamedArguments) > 0 {
		a.error(expr.Pos(), "named arguments are not supported for method calls (please use positional arguments)")
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

		// Build the registry from the auto-generated Kukicha stdlib entries,
		// then layer in Go stdlib functions (which have no .kuki source).
		knownExternalReturns := make(map[string]int, len(generatedStdlibRegistry)+20)
		for k, v := range generatedStdlibRegistry {
			knownExternalReturns[k] = v
		}
		// Go stdlib (not derived from .kuki files)
		knownExternalReturns["os.ReadFile"] = 2
		knownExternalReturns["os.Create"] = 2
		knownExternalReturns["os.Open"] = 2
		knownExternalReturns["os.LookupEnv"] = 2
		knownExternalReturns["strconv.Atoi"] = 2
		knownExternalReturns["strconv.ParseInt"] = 2
		knownExternalReturns["strconv.ParseFloat"] = 2
		knownExternalReturns["url.Parse"] = 2
		knownExternalReturns["fmt.Fprintf"] = 2
		knownExternalReturns["fmt.Sprintf"] = 1
		knownExternalReturns["net.ParseCIDR"] = 3

		if count, ok := knownExternalReturns[qualifiedName]; ok {
			types := make([]*TypeInfo, count)
			for i := range types {
				types[i] = &TypeInfo{Kind: TypeKindUnknown}
			}
			// Provide specific type info for well-known cases
			switch qualifiedName {
			case "os.LookupEnv":
				types[0] = &TypeInfo{Kind: TypeKindString}
				types[1] = &TypeInfo{Kind: TypeKindBool}
			case "fmt.Sprintf":
				types[0] = &TypeInfo{Kind: TypeKindString}
			case "strconv.Atoi":
				types[0] = &TypeInfo{Kind: TypeKindInt}
				types[1] = &TypeInfo{Kind: TypeKindNamed, Name: "error"}
			}
			a.recordReturnCount(expr, count)
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

	// Handle pipedArg for method calls too?
	// Currently method analysis is mostly "Unknown", but we should at least not crash/error on count if we implemented it.
	// Since we return Unknown anyway, ignoring pipedArg here is safe for now,
	// UNLESS we add argument validation logic for methods later.
	// But wait, the previous code didn't validate method arguments at all (loop just calls analyzeExpression).
	// So just updating signature is enough.

	// For now, return unknown - full method resolution requires more complex type system
	// Record a return count of 1 so codegen's onerr discard path has a safe default
	a.recordReturnCount(expr, 1)
	return []*TypeInfo{{Kind: TypeKindUnknown}}
}
