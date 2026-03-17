package codegen

import (
	"fmt"
	"strings"

	"github.com/duber000/kukicha/internal/ast"
	"github.com/duber000/kukicha/internal/semantic"
)

func (g *Generator) exprToString(expr ast.Expression) string {
	if expr == nil {
		return ""
	}

	switch e := expr.(type) {
	case *ast.Identifier:
		// Inside block-style onerr, replace "error" (or the named alias) with the actual error variable.
		if g.currentOnErrVar != "" {
			if e.Value == "error" {
				return g.currentOnErrVar
			}
			if g.currentOnErrAlias != "" && e.Value == g.currentOnErrAlias {
				return g.currentOnErrVar
			}
		}

		// Check if this is the "empty" keyword used as an identifier (e.g. passed as argument)
		// If semantic analysis resolved it to TypeKindNil, it means it's not shadowed, so emit "nil".
		if e.Value == "empty" {
			if t, ok := g.exprTypes[e]; ok && t.Kind == semantic.TypeKindNil {
				// In generic stdlib context, use *new(T) or *new(K) for zero value instead of nil
				// But only if the return type at this position actually uses a placeholder type
				if (g.isStdlibIter || g.isStdlibSlice() || g.isStdlibSort()) && g.placeholderMap != nil {
					// If we're in a return statement and know the return type, check if it's a placeholder
					if g.currentReturnIndex >= 0 && g.currentReturnIndex < len(g.currentReturnTypes) {
						retType := g.currentReturnTypes[g.currentReturnIndex]
						if _, hasT := g.placeholderMap["any"]; hasT && g.typeContainsPlaceholder(retType, "any") {
							return "*new(T)"
						}
						if _, hasK := g.placeholderMap["any2"]; hasK && g.typeContainsPlaceholder(retType, "any2") {
							return "*new(K)"
						}
						if _, hasK := g.placeholderMap["ordered"]; hasK && g.typeContainsPlaceholder(retType, "ordered") {
							return "*new(K)"
						}
						// Return type doesn't use a placeholder — fall through to nil
					} else {
						// Not in a return statement context — use *new(T) as default
						if _, hasT := g.placeholderMap["any"]; hasT {
							return "*new(T)"
						}
						if _, hasK := g.placeholderMap["any2"]; hasK {
							return "*new(K)"
						}
						if _, hasK := g.placeholderMap["ordered"]; hasK {
							return "*new(K)"
						}
					}
				}
				return "nil"
			}
			return "empty"
		}

		return e.Value
	case *ast.IntegerLiteral:
		// Preserve original representation for octal (0...), hex (0x...), binary (0b...)
		lexeme := e.Token.Lexeme
		if len(lexeme) > 1 && lexeme[0] == '0' {
			return lexeme // Keep original format
		}
		return fmt.Sprintf("%d", e.Value)
	case *ast.FloatLiteral:
		return e.Token.Lexeme
	case *ast.RuneLiteral:
		return fmt.Sprintf("'%s'", g.escapeRune(e.Value))
	case *ast.StringLiteral:
		return g.generateStringLiteral(e)
	case *ast.BooleanLiteral:
		if e.Value {
			return "true"
		}
		return "false"
	case *ast.BinaryExpr:
		return g.generateBinaryExpr(e)
	case *ast.UnaryExpr:
		return g.generateUnaryExpr(e)
	case *ast.PipeExpr:
		return g.generatePipeExpr(e)
	case *ast.CallExpr:
		return g.generateCallExpr(e)
	case *ast.MethodCallExpr:
		return g.generateMethodCallExpr(e)
	case *ast.FieldAccessExpr:
		return g.generateFieldAccessExpr(e)
	case *ast.IndexExpr:
		left := g.exprToString(e.Left)
		if u, ok := isNegativeExpr(e.Index); ok {
			absIndex := g.exprToString(u.Right)
			return fmt.Sprintf("%s[len(%s)-%s]", left, left, absIndex)
		}
		index := g.exprToString(e.Index)
		return fmt.Sprintf("%s[%s]", left, index)
	case *ast.SliceExpr:
		return g.generateSliceExpr(e)
	case *ast.StructLiteralExpr:
		return g.generateStructLiteral(e)
	case *ast.ListLiteralExpr:
		return g.generateListLiteral(e)
	case *ast.MapLiteralExpr:
		return g.generateMapLiteral(e)
	case *ast.ReceiveExpr:
		channel := g.exprToString(e.Channel)
		return fmt.Sprintf("<-%s", channel)
	case *ast.TypeCastExpr:
		targetType := g.generateTypeAnnotation(e.TargetType)
		expr := g.exprToString(e.Expression)
		// Use type assertion for interface types, conversion for concrete types.
		if g.isLikelyInterfaceType(targetType) {
			return fmt.Sprintf("%s.(%s)", expr, targetType)
		}
		return fmt.Sprintf("%s(%s)", targetType, expr)
	case *ast.EmptyExpr:
		if e.Type != nil {
			targetType := g.generateTypeAnnotation(e.Type)
			// Check if targetType is a generic type parameter (T, U, K)
			if g.placeholderMap != nil {
				for _, typeParam := range g.placeholderMap {
					if targetType == typeParam {
						return fmt.Sprintf("*new(%s)", targetType)
					}
				}
			}
			return g.zeroValueForType(e.Type)
		}
		// In generic stdlib context, use *new(T) or *new(K) for zero value instead of nil
		// But only if the return type at this position actually uses a placeholder type
		if (g.isStdlibIter || g.isStdlibSlice() || g.isStdlibSort()) && g.placeholderMap != nil {
			// If we're in a return statement and know the return type, check if it's a placeholder
			if g.currentReturnIndex >= 0 && g.currentReturnIndex < len(g.currentReturnTypes) {
				retType := g.currentReturnTypes[g.currentReturnIndex]
				if _, hasT := g.placeholderMap["any"]; hasT && g.typeContainsPlaceholder(retType, "any") {
					return "*new(T)"
				}
				if _, hasK := g.placeholderMap["any2"]; hasK && g.typeContainsPlaceholder(retType, "any2") {
					return "*new(K)"
				}
				if _, hasK := g.placeholderMap["ordered"]; hasK && g.typeContainsPlaceholder(retType, "ordered") {
					return "*new(K)"
				}
				// Return type doesn't use a placeholder — fall through to nil
			} else {
				// Not in a return statement context — use *new(T) as default
				if _, hasT := g.placeholderMap["any"]; hasT {
					return "*new(T)"
				}
				if _, hasK := g.placeholderMap["any2"]; hasK {
					return "*new(K)"
				}
				if _, hasK := g.placeholderMap["ordered"]; hasK {
					return "*new(K)"
				}
			}
		}
		return "nil"
	case *ast.DiscardExpr:
		return "_"
	case *ast.ErrorExpr:
		message := g.exprToString(e.Message)
		return fmt.Sprintf("errors.New(%s)", message)
	case *ast.ReturnExpr:
		return g.generateReturnExpr(e)
	case *ast.MakeExpr:
		return g.generateMakeExpr(e)
	case *ast.CloseExpr:
		channel := g.exprToString(e.Channel)
		return fmt.Sprintf("close(%s)", channel)
	case *ast.PanicExpr:
		message := g.exprToString(e.Message)
		return fmt.Sprintf("panic(%s)", message)
	case *ast.RecoverExpr:
		return "recover()"
	case *ast.FunctionLiteral:
		return g.generateFunctionLiteral(e)
	case *ast.ArrowLambda:
		return g.generateArrowLambda(e)
	case *ast.AddressOfExpr:
		return g.generateAddressOfExpr(e)
	case *ast.DerefExpr:
		return g.generateDerefExpr(e)
	case *ast.TypeAssertionExpr:
		targetType := g.generateTypeAnnotation(e.TargetType)
		expr := g.exprToString(e.Expression)
		return fmt.Sprintf("%s.(%s)", expr, targetType)
	case *ast.PipedSwitchExpr:
		return g.generatePipedSwitchExpr(e)
	default:
		pos := expr.Pos()
		panic(fmt.Sprintf("codegen: unhandled expression type %T at %s:%d:%d", expr, pos.File, pos.Line, pos.Column))
	}
}

func (g *Generator) generatePipedSwitchExpr(expr *ast.PipedSwitchExpr) string {
	left := g.exprToString(expr.Left)
	tempGen := &Generator{
		program:        g.program,
		output:         strings.Builder{},
		indent:         g.indent + 1,
		placeholderMap: g.placeholderMap,
		autoImports:    g.autoImports,
		pkgAliases:     g.pkgAliases,
		funcDefaults:   g.funcDefaults,
		isStdlibIter:   g.isStdlibIter,
		sourceFile:     g.sourceFile,
		exprTypes:      g.exprTypes,
	}

	switch stmt := expr.Switch.(type) {
	case *ast.SwitchStmt:
		originalExpr := stmt.Expression
		stmt.Expression = &ast.Identifier{Value: left}
		tempGen.generateSwitchStmt(stmt)
		stmt.Expression = originalExpr
	case *ast.TypeSwitchStmt:
		originalExpr := stmt.Expression
		stmt.Expression = &ast.Identifier{Value: left}
		tempGen.generateTypeSwitchStmt(stmt)
		stmt.Expression = originalExpr
	default:
		return ""
	}

	returnType := g.pipedSwitchReturnType(expr)

	var result strings.Builder
	if returnType != "" {
		result.WriteString(fmt.Sprintf("func() %s {\n", returnType))
	} else {
		result.WriteString("func() {\n")
	}
	result.WriteString(tempGen.output.String())
	for i := 0; i < g.indent; i++ {
		result.WriteString("\t")
	}
	result.WriteString("}()")
	return result.String()
}

func (g *Generator) pipedSwitchReturnType(expr *ast.PipedSwitchExpr) string {
	if g.exprTypes != nil {
		if ti, ok := g.exprTypes[expr]; ok && ti != nil && ti.Kind != semantic.TypeKindUnknown {
			return g.typeInfoToGoString(ti)
		}
	}
	return g.inferPipedSwitchReturnType(expr.Switch)
}

// inferPipedSwitchReturnType scans a switch body for return statements and
// determines the Go return type for the IIFE wrapper. Returns empty string
// when no cases return a value (void IIFE). Returns "any" when returns exist
// but types cannot be inferred consistently. Uses exprTypes populated by the
// semantic analyzer to resolve expression types.
func (g *Generator) inferPipedSwitchReturnType(stmt ast.PipedSwitchBody) string {
	var returnExprs []ast.Expression
	inferTypedReturn := func(expr ast.Expression, binding string, typeAnn ast.TypeAnnotation) string {
		if id, ok := expr.(*ast.Identifier); ok && id.Value == binding && typeAnn != nil {
			return g.generateTypeAnnotation(typeAnn)
		}
		return g.inferExprType(expr)
	}

	collectReturns := func(body *ast.BlockStmt) {
		if body == nil {
			return
		}
		for _, s := range body.Statements {
			if ret, ok := s.(*ast.ReturnStmt); ok && len(ret.Values) > 0 {
				returnExprs = append(returnExprs, ret.Values[0])
			}
		}
	}

	switch s := stmt.(type) {
	case *ast.SwitchStmt:
		for _, c := range s.Cases {
			collectReturns(c.Body)
		}
		if s.Otherwise != nil {
			collectReturns(s.Otherwise.Body)
		}
	case *ast.TypeSwitchStmt:
		var inferredType string
		collectTypedReturns := func(body *ast.BlockStmt, typeAnn ast.TypeAnnotation) bool {
			if body == nil {
				return true
			}
			for _, stmt := range body.Statements {
				ret, ok := stmt.(*ast.ReturnStmt)
				if !ok || len(ret.Values) == 0 {
					continue
				}
				ts := inferTypedReturn(ret.Values[0], s.Binding.Value, typeAnn)
				if ts == "" {
					inferredType = "any"
					return false
				}
				if inferredType == "" {
					inferredType = ts
				} else if ts != inferredType {
					inferredType = "any"
					return false
				}
			}
			return true
		}
		for _, c := range s.Cases {
			if !collectTypedReturns(c.Body, c.Type) {
				return "any"
			}
		}
		if s.Otherwise != nil && !collectTypedReturns(s.Otherwise.Body, nil) {
			return "any"
		}
		if inferredType == "" {
			return ""
		}
		return inferredType
	}

	if len(returnExprs) == 0 {
		return "" // void IIFE — no return type needed
	}

	var inferredType string
	for _, expr := range returnExprs {
		ts := g.inferExprType(expr)
		if ts == "" {
			return "any" // can't determine type for this return expression
		}
		if inferredType == "" {
			inferredType = ts
		} else if ts != inferredType {
			return "any" // inconsistent return types across cases
		}
	}

	if inferredType == "" {
		return "any"
	}
	return inferredType
}

// inferExprType returns the Go type string for an expression, consulting exprTypes
// first, then falling back to direct AST literal inspection for common cases.
func (g *Generator) inferExprType(expr ast.Expression) string {
	if g.exprTypes != nil {
		if ti, ok := g.exprTypes[expr]; ok && ti != nil {
			return g.typeInfoToGoString(ti)
		}
	}
	// Fall back to AST literal type inspection
	switch expr.(type) {
	case *ast.StringLiteral:
		return "string"
	case *ast.IntegerLiteral:
		return "int"
	case *ast.FloatLiteral:
		return "float64"
	case *ast.BooleanLiteral:
		return "bool"
	}
	return ""
}

// escapeRune returns the Go escape sequence for a rune
func (g *Generator) escapeRune(r rune) string {
	switch r {
	case '\n':
		return "\\n"
	case '\t':
		return "\\t"
	case '\r':
		return "\\r"
	case '\\':
		return "\\\\"
	case '\'':
		return "\\'"
	case '\x00':
		return "\\x00"
	default:
		return string(r)
	}
}

// escapeString returns a string with special characters escaped for Go string literals
func (g *Generator) escapeString(s string) string {
	var result strings.Builder
	for _, r := range s {
		switch r {
		case '\n':
			result.WriteString("\\n")
		case '\t':
			result.WriteString("\\t")
		case '\r':
			result.WriteString("\\r")
		case '\\':
			result.WriteString("\\\\")
		case '"':
			result.WriteString("\\\"")
		case '\x00':
			result.WriteString("\\x00")
		case '\uE000':
			result.WriteRune('{') // PUA sentinel → literal {
		case '\uE001':
			result.WriteRune('}') // PUA sentinel → literal }
		default:
			result.WriteRune(r)
		}
	}
	return result.String()
}

func (g *Generator) generateStringLiteral(lit *ast.StringLiteral) string {
	if !lit.Interpolated && !strings.ContainsRune(lit.Value, '\uE002') {
		return fmt.Sprintf("\"%s\"", g.escapeString(lit.Value))
	}

	// Non-interpolated string with \sep sentinel — handle inline since Parts is empty
	if len(lit.Parts) == 0 {
		return g.generateSepOnlyString(lit.Value)
	}

	return g.generateStringFromParts(lit)
}

// generateSepOnlyString handles non-interpolated strings that contain \uE002 (\sep) sentinels.
func (g *Generator) generateSepOnlyString(value string) string {
	g.addImport("path/filepath")
	segments := strings.Split(value, "\uE002")
	if len(segments) == 2 && segments[0] == "" && segments[1] == "" {
		// String is just "\sep"
		return "string(filepath.Separator)"
	}
	var format strings.Builder
	var args []string
	for i, seg := range segments {
		format.WriteString(g.escapeString(seg))
		if i < len(segments)-1 {
			format.WriteString("%v")
			args = append(args, "string(filepath.Separator)")
		}
	}
	if len(args) == 0 {
		return fmt.Sprintf("\"%s\"", format.String())
	}
	return fmt.Sprintf("fmt.Sprintf(\"%s\", %s)", format.String(), strings.Join(args, ", "))
}

// generateStringFromParts generates a Go string expression from pre-parsed interpolation parts.
func (g *Generator) generateStringFromParts(lit *ast.StringLiteral) string {
	var format strings.Builder
	var args []string

	for _, part := range lit.Parts {
		if part.IsLiteral {
			literal := part.Literal
			// Handle \uE002 (\sep) in literal parts
			if strings.ContainsRune(literal, '\uE002') {
				g.addImport("path/filepath")
				segments := strings.Split(literal, "\uE002")
				for i, seg := range segments {
					format.WriteString(g.escapeString(seg))
					if i < len(segments)-1 {
						format.WriteString("%v")
						args = append(args, "string(filepath.Separator)")
					}
				}
			} else {
				format.WriteString(g.escapeString(literal))
			}
		} else {
			format.WriteString("%v")
			// Check onerr substitution
			if g.currentOnErrVar != "" {
				if ident, ok := part.Expr.(*ast.Identifier); ok {
					if ident.Value == "error" || (g.currentOnErrAlias != "" && ident.Value == g.currentOnErrAlias) {
						args = append(args, g.currentOnErrVar)
						continue
					}
				}
			}
			args = append(args, g.exprToString(part.Expr))
		}
	}

	if len(args) == 0 {
		return fmt.Sprintf("\"%s\"", format.String())
	}
	argsStr := strings.Join(args, ", ")
	return fmt.Sprintf("fmt.Sprintf(\"%s\", %s)", format.String(), argsStr)
}

// parseStringPartsOrInterpolation returns a format string and args from a StringLiteral.
func (g *Generator) parseStringPartsOrInterpolation(lit *ast.StringLiteral) (string, []string) {
	var format strings.Builder
	var args []string
	for _, part := range lit.Parts {
		if part.IsLiteral {
			literal := part.Literal
			if strings.ContainsRune(literal, '\uE002') {
				g.addImport("path/filepath")
				segments := strings.Split(literal, "\uE002")
				for i, seg := range segments {
					format.WriteString(g.escapeString(seg))
					if i < len(segments)-1 {
						format.WriteString("%v")
						args = append(args, "string(filepath.Separator)")
					}
				}
			} else {
				format.WriteString(g.escapeString(literal))
			}
		} else {
			format.WriteString("%v")
			if g.currentOnErrVar != "" {
				if ident, ok := part.Expr.(*ast.Identifier); ok {
					if ident.Value == "error" || (g.currentOnErrAlias != "" && ident.Value == g.currentOnErrAlias) {
						args = append(args, g.currentOnErrVar)
						continue
					}
				}
			}
			args = append(args, g.exprToString(part.Expr))
		}
	}
	return format.String(), args
}

func (g *Generator) generateBinaryExpr(expr *ast.BinaryExpr) string {
	left := g.exprToString(expr.Left)
	right := g.exprToString(expr.Right)

	// Map Kukicha operators to Go operators
	op := expr.Operator
	switch op {
	case "and":
		op = "&&"
	case "or":
		op = "||"
	case "equals":
		op = "=="
	case "not equals":
		op = "!="
	}

	return fmt.Sprintf("(%s %s %s)", left, op, right)
}

func (g *Generator) generateUnaryExpr(expr *ast.UnaryExpr) string {
	right := g.exprToString(expr.Right)

	op := expr.Operator
	if op == "not" {
		op = "!"
	}

	return fmt.Sprintf("%s%s", op, right)
}

func (g *Generator) generateAddressOfExpr(expr *ast.AddressOfExpr) string {
	operand := g.exprToString(expr.Operand)
	if isNonAddressable(expr.Operand) {
		return fmt.Sprintf("new(%s)", operand)
	}
	return fmt.Sprintf("&%s", operand)
}

// isNonAddressable returns true for expressions whose results cannot be
// addressed with & in Go (e.g. function/method call return values).
// Go 1.26's new(expr) syntax handles these cases.
func isNonAddressable(expr ast.Expression) bool {
	switch expr.(type) {
	case *ast.CallExpr, *ast.MethodCallExpr:
		return true
	default:
		return false
	}
}

func (g *Generator) generateDerefExpr(expr *ast.DerefExpr) string {
	operand := g.exprToString(expr.Operand)
	return fmt.Sprintf("*%s", operand)
}

// generatePipeExpr transforms pipe expressions into function calls.
//
// ARCHITECTURE NOTE: Kukicha's pipe operator (|>) supports three strategies
// to determine where the piped value is inserted:
//
//	Strategy A (Placeholder): User explicitly marks position with "_"
//	  data |> json.MarshalWrite(w, _)  →  json.MarshalWrite(w, data)
//
//	Strategy B (Data-First): Default - piped value becomes first argument
//	  users |> slice.Filter(fn)  →  slice.Filter(users, fn)
//
//	Strategy C (Context-First): If piped value is a context.Context, special handling
//	  ctx |> db.Query(sql)  →  db.Query(ctx, sql)
//
// The placeholder strategy (A) takes precedence. This design lets users handle
// APIs where the "data" isn't the first parameter, without requiring Kukicha
// to know every function signature in the ecosystem.
func (g *Generator) generatePipeExpr(expr *ast.PipeExpr) string {
	// Transform a |> b() into b(a)
	// Supports placeholder strategy: a |> b(x, _) becomes b(x, a)
	// Supports context-first strategy: ctx |> b(x) becomes b(ctx, x)

	// Calculate Left expression first, handling multi-return values if needed
	leftExpr := g.exprToString(expr.Left)
	if count, ok := g.inferReturnCount(expr.Left); ok && count >= 2 {
		// Wrap in a function call to only take the first return value
		// e.g., func() any { val, _ := fetch.Get(...); return val }()
		blanks := make([]string, count-1)
		for i := range blanks {
			blanks[i] = "_"
		}
		leftExpr = fmt.Sprintf("func() any { val, %s := %s; return val }()", strings.Join(blanks, ", "), leftExpr)
	}

	// Right side can be a CallExpr or MethodCallExpr
	var funcName string
	var arguments []ast.Expression
	var isVariadic bool

	if call, ok := expr.Right.(*ast.CallExpr); ok {
		funcName = g.exprToString(call.Function)
		// Check if this is a print() builtin - transpile to fmt.Println() or fmt.Fprintln(os.Stderr)
		if id, ok := call.Function.(*ast.Identifier); ok && id.Value == "print" {
			if g.mcpTarget {
				funcName = "fmt.Fprintln"
			} else {
				funcName = "fmt.Println"
			}
		}
		arguments = call.Arguments
		isVariadic = call.Variadic
	} else if method, ok := expr.Right.(*ast.MethodCallExpr); ok {
		objStr := g.exprToString(method.Object)
		if alias, ok := g.pkgAliases[objStr]; ok {
			objStr = alias
		}
		funcName = objStr + "." + method.Method.Value
		if method.Object == nil {
			// Shorthand: .Method()
			funcName = leftExpr + "." + method.Method.Value

			// Method call: obj.Method(args)
			arguments = method.Arguments
			isVariadic = method.Variadic

			args := make([]string, len(arguments))
			for i, arg := range arguments {
				args[i] = g.exprToString(arg)
			}
			if isVariadic {
				return fmt.Sprintf("%s(%s...)", funcName, strings.Join(args, ", "))
			}
			return fmt.Sprintf("%s(%s)", funcName, strings.Join(args, ", "))
		}
		arguments = method.Arguments
		isVariadic = method.Variadic
	} else if field, ok := expr.Right.(*ast.FieldAccessExpr); ok {
		if field.Object == nil {
			return leftExpr + "." + field.Field.Value
		}
		return g.generateFieldAccessExpr(field)
	} else if id, ok := expr.Right.(*ast.Identifier); ok {
		// Bare identifier on right side of pipe: treat as function call with piped value
		// e.g., data |> print  →  fmt.Println(data)
		funcName := id.Value
		if funcName == "print" {
			if g.mcpTarget {
				return fmt.Sprintf("fmt.Fprintln(os.Stderr, %s)", leftExpr)
			}
			return fmt.Sprintf("fmt.Println(%s)", leftExpr)
		}
		return fmt.Sprintf("%s(%s)", funcName, leftExpr)
	} else {
		pos := expr.Right.Pos()
		panic(fmt.Sprintf("codegen: unhandled pipe target %T at %s:%d:%d", expr.Right, pos.File, pos.Line, pos.Column))
	}

	// Build the argument list using the shared helper
	args := g.buildPipeArgs(leftExpr, arguments)

	// Fill in missing trailing args from stdlib registry defaults.
	// funcName may have been aliased (e.g., "kukistring.PadRight"), so
	// also try the original Kukicha name for registry lookup.
	g.fillStdlibDefaults(funcName, expr.Right, &args)

	// MCP special case: prepend os.Stderr for fmt.Fprintln
	if g.mcpTarget && funcName == "fmt.Fprintln" {
		args = append([]string{"os.Stderr"}, args...)
	}

	if isVariadic {
		return fmt.Sprintf("%s(%s...)", funcName, strings.Join(args, ", "))
	}

	return fmt.Sprintf("%s(%s)", funcName, strings.Join(args, ", "))
}

func (g *Generator) generateCallExpr(expr *ast.CallExpr) string {
	funcName := g.exprToString(expr.Function)

	// Check if this is a print() builtin - transpile to fmt.Println() or fmt.Fprintln(os.Stderr)
	isPrintCall := false
	if id, ok := expr.Function.(*ast.Identifier); ok {
		if id.Value == "print" {
			isPrintCall = true
			if g.mcpTarget {
				funcName = "fmt.Fprintln"
			} else {
				funcName = "fmt.Println"
			}
		}
	}

	// If there are no named arguments and no defaults need filling, use the simple path
	if len(expr.NamedArguments) == 0 {
		needsDefaults := false
		if id, ok := expr.Function.(*ast.Identifier); ok {
			if fd := g.funcDefaults[id.Value]; fd != nil && len(expr.Arguments) < len(fd.ParamNames) {
				needsDefaults = true
			}
		}

		if !needsDefaults {
			args := make([]string, 0, len(expr.Arguments)+1)
			if g.mcpTarget && isPrintCall {
				args = append(args, "os.Stderr")
			}
			for _, arg := range expr.Arguments {
				args = append(args, g.exprToString(arg))
			}

			if expr.Variadic {
				return fmt.Sprintf("%s(%s...)", funcName, strings.Join(args, ", "))
			}
			return fmt.Sprintf("%s(%s)", funcName, strings.Join(args, ", "))
		}
	}

	// Handle named arguments - reorder based on function parameter order
	// Look up function defaults to get parameter names
	var funcDef *FuncDefaults
	if id, ok := expr.Function.(*ast.Identifier); ok {
		funcDef = g.funcDefaults[id.Value]
	}

	if funcDef == nil {
		// Can't resolve function - emit named arguments in order they appear
		args := make([]string, 0, len(expr.Arguments)+len(expr.NamedArguments))
		for _, arg := range expr.Arguments {
			args = append(args, g.exprToString(arg))
		}
		for _, namedArg := range expr.NamedArguments {
			args = append(args, g.exprToString(namedArg.Value))
		}
		if expr.Variadic {
			return fmt.Sprintf("%s(%s...)", funcName, strings.Join(args, ", "))
		}
		return fmt.Sprintf("%s(%s)", funcName, strings.Join(args, ", "))
	}

	// Build argument map from named arguments
	namedArgMap := make(map[string]ast.Expression)
	for _, namedArg := range expr.NamedArguments {
		namedArgMap[namedArg.Name.Value] = namedArg.Value
	}

	// Build final argument list in parameter order
	args := make([]string, len(funcDef.ParamNames))
	positionalIdx := 0

	for i, paramName := range funcDef.ParamNames {
		if positionalIdx < len(expr.Arguments) {
			// Use positional argument
			args[i] = g.exprToString(expr.Arguments[positionalIdx])
			positionalIdx++
		} else if namedVal, ok := namedArgMap[paramName]; ok {
			// Use named argument
			args[i] = g.exprToString(namedVal)
		} else if funcDef.DefaultValues[i] != nil {
			// Use default value
			args[i] = g.exprToString(funcDef.DefaultValues[i])
		} else if i == len(funcDef.ParamNames)-1 && funcDef.HasVariadic {
			// Last parameter is variadic with no args provided - omit it
			args = args[:i]
			break
		} else {
			// Missing argument - this should be caught by semantic analysis
			// For safety, use empty placeholder
			args[i] = "/* missing argument */"
		}
	}

	if expr.Variadic {
		return fmt.Sprintf("%s(%s...)", funcName, strings.Join(args, ", "))
	}
	return fmt.Sprintf("%s(%s)", funcName, strings.Join(args, ", "))
}

func (g *Generator) generateMethodCallExpr(expr *ast.MethodCallExpr) string {
	object := g.exprToString(expr.Object)
	method := expr.Method.Value

	// Rewrite package name if it was auto-aliased due to collision
	if alias, ok := g.pkgAliases[object]; ok {
		object = alias
	}

	// Check if this is a printf-style method (Errorf, Fatalf, Logf, Skipf, Printf, etc.)
	// These methods require a constant format string in Go 1.26+
	if g.isPrintfStyleMethod(method) && len(expr.Arguments) > 0 {
		if strLit, ok := expr.Arguments[0].(*ast.StringLiteral); ok {
			format, formatArgs := g.parseStringPartsOrInterpolation(strLit)
			if len(formatArgs) > 0 {
				// Generate: t.Errorf("format %v", args...) instead of t.Errorf(fmt.Sprintf(...))
				allArgs := make([]string, 0, len(formatArgs)+len(expr.Arguments)-1)
				allArgs = append(allArgs, formatArgs...)
				// Add remaining arguments after the format string
				for i := 1; i < len(expr.Arguments); i++ {
					allArgs = append(allArgs, g.exprToString(expr.Arguments[i]))
				}
				if expr.Variadic {
					return fmt.Sprintf("%s.%s(\"%s\", %s...)", object, method, format, strings.Join(allArgs, ", "))
				}
				return fmt.Sprintf("%s.%s(\"%s\", %s)", object, method, format, strings.Join(allArgs, ", "))
			}
		}
	}

	// Collect all arguments: positional first, then named (in their declaration order)
	args := make([]string, 0, len(expr.Arguments)+len(expr.NamedArguments))

	// Add positional arguments
	for _, arg := range expr.Arguments {
		args = append(args, g.exprToString(arg))
	}

	// Add named argument values (in the order they appear)
	for _, namedArg := range expr.NamedArguments {
		args = append(args, g.exprToString(namedArg.Value))
	}

	// Fill in missing trailing args from stdlib registry defaults
	goName := object + "." + method
	g.fillStdlibDefaults(goName, expr, &args)

	if expr.Variadic {
		return fmt.Sprintf("%s.%s(%s...)", object, method, strings.Join(args, ", "))
	}
	return fmt.Sprintf("%s.%s(%s)", object, method, strings.Join(args, ", "))
}

func (g *Generator) generateFieldAccessExpr(expr *ast.FieldAccessExpr) string {
	object := g.exprToString(expr.Object)
	field := expr.Field.Value

	if alias, ok := g.pkgAliases[object]; ok {
		object = alias
	}

	return fmt.Sprintf("%s.%s", object, field)
}

// fillStdlibDefaults appends missing trailing default values from the stdlib
// registry to args. goName is the Go-aliased function name (e.g., "kukistring.PadRight").
// exprNode is the AST node used to extract the original Kukicha package name.
func (g *Generator) fillStdlibDefaults(goName string, exprNode ast.Expression, args *[]string) {
	// Try the Go name first (works for non-aliased packages)
	kukichaName := goName

	// If Go name not in registry, derive the Kukicha name from the AST
	if _, ok := semantic.GetStdlibEntry(kukichaName); !ok {
		switch n := exprNode.(type) {
		case *ast.MethodCallExpr:
			if objID, ok := n.Object.(*ast.Identifier); ok {
				kukichaName = objID.Value + "." + n.Method.Value
			}
		case *ast.CallExpr:
			if mc, ok := n.Function.(*ast.MethodCallExpr); ok {
				if objID, ok := mc.Object.(*ast.Identifier); ok {
					kukichaName = objID.Value + "." + mc.Method.Value
				}
			}
		}
	}

	entry, ok := semantic.GetStdlibEntry(kukichaName)
	if !ok || len(entry.DefaultValues) == 0 {
		return
	}
	for len(*args) < len(entry.ParamNames) {
		idx := len(*args)
		if idx < len(entry.DefaultValues) && entry.DefaultValues[idx] != "" {
			*args = append(*args, entry.DefaultValues[idx])
		} else {
			break
		}
	}
}

// printfMethods lists printf-style methods that expect a format string as their first argument.
var printfMethods = map[string]bool{
	"Errorf":  true,
	"Fatalf":  true,
	"Logf":    true,
	"Skipf":   true,
	"Printf":  true,
	"Sprintf": true,
	"Fprintf": true,
	"Panicf":  true,
	"Warnf":   true,
	"Infof":   true,
	"Debugf":  true,
}

// isPrintfStyleMethod returns true if the method name is a printf-style method
// that expects a format string as its first argument.
func (g *Generator) isPrintfStyleMethod(method string) bool {
	return printfMethods[method]
}

func (g *Generator) generateSliceExpr(expr *ast.SliceExpr) string {
	left := g.exprToString(expr.Left)

	var start, end string
	if expr.Start != nil {
		if u, ok := isNegativeExpr(expr.Start); ok {
			absIndex := g.exprToString(u.Right)
			start = fmt.Sprintf("len(%s)-%s", left, absIndex)
		} else {
			start = g.exprToString(expr.Start)
		}
	}
	if expr.End != nil {
		if u, ok := isNegativeExpr(expr.End); ok {
			absIndex := g.exprToString(u.Right)
			end = fmt.Sprintf("len(%s)-%s", left, absIndex)
		} else {
			end = g.exprToString(expr.End)
		}
	}

	return fmt.Sprintf("%s[%s:%s]", left, start, end)
}

// isNegativeExpr checks if an expression is a unary minus (negative index).
func isNegativeExpr(expr ast.Expression) (*ast.UnaryExpr, bool) {
	u, ok := expr.(*ast.UnaryExpr)
	return u, ok && u.Operator == "-"
}

func (g *Generator) generateStructLiteral(expr *ast.StructLiteralExpr) string {
	typeName := g.generateTypeAnnotation(expr.Type)

	if len(expr.Fields) == 0 {
		return fmt.Sprintf("%s{}", typeName)
	}

	fields := make([]string, len(expr.Fields))
	for i, field := range expr.Fields {
		value := g.exprToString(field.Value)
		fields[i] = fmt.Sprintf("%s: %s", field.Name.Value, value)
	}

	return fmt.Sprintf("%s{%s}", typeName, strings.Join(fields, ", "))
}

func (g *Generator) generateListLiteral(expr *ast.ListLiteralExpr) string {
	if len(expr.Elements) == 0 {
		if expr.Type != nil {
			elemType := g.generateTypeAnnotation(expr.Type)
			return fmt.Sprintf("[]%s{}", elemType)
		}
		return "[]any{}"
	}

	elements := make([]string, len(expr.Elements))
	for i, elem := range expr.Elements {
		elements[i] = g.exprToString(elem)
	}

	typePrefix := ""
	if expr.Type != nil {
		elemType := g.generateTypeAnnotation(expr.Type)
		typePrefix = fmt.Sprintf("[]%s", elemType)
	} else {
		typePrefix = "[]any"
	}

	return fmt.Sprintf("%s{%s}", typePrefix, strings.Join(elements, ", "))
}

func (g *Generator) generateMapLiteral(expr *ast.MapLiteralExpr) string {
	keyType := g.generateTypeAnnotation(expr.KeyType)
	valType := g.generateTypeAnnotation(expr.ValType)

	if len(expr.Pairs) == 0 {
		return fmt.Sprintf("map[%s]%s{}", keyType, valType)
	}

	pairs := make([]string, len(expr.Pairs))
	for i, pair := range expr.Pairs {
		key := g.exprToString(pair.Key)
		value := g.exprToString(pair.Value)
		pairs[i] = fmt.Sprintf("%s: %s", key, value)
	}

	return fmt.Sprintf("map[%s]%s{%s}", keyType, valType, strings.Join(pairs, ", "))
}

func (g *Generator) generateMakeExpr(expr *ast.MakeExpr) string {
	targetType := g.generateTypeAnnotation(expr.Type)

	if len(expr.Args) == 0 {
		// Slices require a size argument, maps and channels don't
		if strings.HasPrefix(targetType, "[]") {
			return fmt.Sprintf("make(%s, 0)", targetType)
		}
		return fmt.Sprintf("make(%s)", targetType)
	}

	args := make([]string, len(expr.Args))
	for i, arg := range expr.Args {
		args[i] = g.exprToString(arg)
	}

	return fmt.Sprintf("make(%s, %s)", targetType, strings.Join(args, ", "))
}



func (g *Generator) generateReturnExpr(expr *ast.ReturnExpr) string {
	values := make([]string, len(expr.Values))
	for i, v := range expr.Values {
		values[i] = g.exprToString(v)
	}
	return "return " + strings.Join(values, ", ")
}
