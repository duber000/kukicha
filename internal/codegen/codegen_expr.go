package codegen

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/duber000/kukicha/internal/ast"
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
	case *ast.IndexExpr:
		left := g.exprToString(e.Left)
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
		// Use type assertion syntax for interface types (contains a dot like http.Handler)
		// or when likely asserting from any/interface
		// Exception: iter.Seq types are functions, so use conversion
		if strings.Contains(targetType, ".") && !strings.Contains(targetType, "iter.Seq") {
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
		if (g.isStdlibIter || g.isStdlibSlice()) && g.placeholderMap != nil {
			// If we're in a return statement and know the return type, check if it's a placeholder
			if g.currentReturnIndex >= 0 && g.currentReturnIndex < len(g.currentReturnTypes) {
				retType := g.currentReturnTypes[g.currentReturnIndex]
				if _, hasT := g.placeholderMap["any"]; hasT && g.typeContainsPlaceholder(retType, "any") {
					return "*new(T)"
				}
				if _, hasK := g.placeholderMap["any2"]; hasK && g.typeContainsPlaceholder(retType, "any2") {
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
	default:
		return ""
	}
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
		default:
			result.WriteRune(r)
		}
	}
	return result.String()
}

func (g *Generator) generateStringLiteral(lit *ast.StringLiteral) string {
	if !lit.Interpolated {
		return fmt.Sprintf("\"%s\"", g.escapeString(lit.Value))
	}

	// Parse string interpolation
	return g.generateStringInterpolation(lit.Value)
}

func (g *Generator) generateStringInterpolation(str string) string {
	format, args := g.parseStringInterpolation(str)
	if len(args) == 0 {
		return fmt.Sprintf("\"%s\"", format)
	}
	argsStr := strings.Join(args, ", ")
	return fmt.Sprintf("fmt.Sprintf(\"%s\", %s)", format, argsStr)
}

// parseStringInterpolation extracts the format string and arguments from
// a Kukicha string with {expr} interpolation patterns.
// Returns the format string (with %v placeholders) and the list of argument expressions.
func (g *Generator) parseStringInterpolation(str string) (string, []string) {
	// Find all {expr} patterns where expr starts with an identifier character.
	// This avoids matching regex quantifiers like {2,} or {3,5}.
	re := regexp.MustCompile(`\{([a-zA-Z_][^}]*)\}`)
	matches := re.FindAllStringSubmatchIndex(str, -1)

	if len(matches) == 0 {
		return g.escapeString(str), nil
	}

	// Build format string and args
	var format strings.Builder
	args := []string{}
	lastIndex := 0

	for _, match := range matches {
		// Add literal part before the interpolation (escaped)
		if match[0] > lastIndex {
			format.WriteString(g.escapeString(str[lastIndex:match[0]]))
		}

		// Add format specifier
		format.WriteString("%v")

		// Extract expression and transform Kukicha syntax to Go
		expr := str[match[2]:match[3]]
		expr = g.transformInterpolatedExpr(expr)
		args = append(args, expr)

		lastIndex = match[1]
	}

	// Add remaining literal part (escaped)
	if lastIndex < len(str) {
		format.WriteString(g.escapeString(str[lastIndex:]))
	}

	return format.String(), args
}

// transformInterpolatedExpr converts Kukicha expression syntax in string
// interpolation to valid Go syntax.
func (g *Generator) transformInterpolatedExpr(expr string) string {
	// Inside block-style onerr, replace "error" (or the named alias) with the actual error variable.
	trimmed := strings.TrimSpace(expr)
	if g.currentOnErrVar != "" {
		if trimmed == "error" {
			return g.currentOnErrVar
		}
		if g.currentOnErrAlias != "" && trimmed == g.currentOnErrAlias {
			return g.currentOnErrVar
		}
	}

	// Handle "X as Type" -> "Type(X)" for type conversions
	// This is a simple string-based transformation for common cases
	asRe := regexp.MustCompile(`^(.+)\s+as\s+(\w+)$`)
	if matches := asRe.FindStringSubmatch(strings.TrimSpace(expr)); matches != nil {
		value := strings.TrimSpace(matches[1])
		targetType := matches[2]
		return fmt.Sprintf("%s(%s)", targetType, value)
	}
	return expr
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

func (g *Generator) isContextExpr(expr ast.Expression) bool {
	// Simple type detection for Context
	// 1. Literal 'ctx' identifier
	if id, ok := expr.(*ast.Identifier); ok {
		return id.Value == "ctx"
	}
	// 2. Call to context package (e.g., context.Background(), context.WithTimeout())
	if call, ok := expr.(*ast.CallExpr); ok {
		if id, ok := call.Function.(*ast.Identifier); ok {
			return strings.HasPrefix(id.Value, "context.")
		}
	}
	return false
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
			// Shorthand: .Method() or .Field
			// We will prepend expr.Left as the object
			funcName = leftExpr + "." + method.Method.Value

			if !method.IsCall {
				// Field access: obj.Field
				return funcName
			}

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
		// Normal method call (already has object)
		if !method.IsCall {
			return funcName
		}
		arguments = method.Arguments
		isVariadic = method.Variadic
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
		// Fallback: If piping into something that isn't a call
		return leftExpr + " |> " + g.exprToString(expr.Right)
	}

	// Build the argument list using the shared helper
	args := g.buildPipeArgs(leftExpr, arguments)

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

	// If no parentheses were used, it's field access
	if !expr.IsCall {
		return fmt.Sprintf("%s.%s", object, method)
	}

	// Check if this is a printf-style method (Errorf, Fatalf, Logf, Skipf, Printf, etc.)
	// These methods require a constant format string in Go 1.26+
	if g.isPrintfStyleMethod(method) && len(expr.Arguments) > 0 {
		if strLit, ok := expr.Arguments[0].(*ast.StringLiteral); ok {
			format, formatArgs := g.parseStringInterpolation(strLit.Value)
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

	if expr.Variadic {
		return fmt.Sprintf("%s.%s(%s...)", object, method, strings.Join(args, ", "))
	}
	return fmt.Sprintf("%s.%s(%s)", object, method, strings.Join(args, ", "))
}

// isPrintfStyleMethod returns true if the method name is a printf-style method
// that expects a format string as its first argument.
func (g *Generator) isPrintfStyleMethod(method string) bool {
	// Common printf-style methods from testing, fmt, and log packages
	printfMethods := map[string]bool{
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
	return printfMethods[method]
}

func (g *Generator) generateSliceExpr(expr *ast.SliceExpr) string {
	left := g.exprToString(expr.Left)

	var start, end string
	if expr.Start != nil {
		start = g.exprToString(expr.Start)
	}
	if expr.End != nil {
		end = g.exprToString(expr.End)
	}

	return fmt.Sprintf("%s[%s:%s]", left, start, end)
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

func (g *Generator) isPrintBuiltin(expr ast.Expression) bool {
	if id, ok := expr.(*ast.Identifier); ok {
		return id.Value == "print"
	}
	return false
}

func (g *Generator) generateReturnExpr(expr *ast.ReturnExpr) string {
	values := make([]string, len(expr.Values))
	for i, v := range expr.Values {
		values[i] = g.exprToString(v)
	}
	return "return " + strings.Join(values, ", ")
}

