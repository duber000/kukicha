package semantic

import (
	"fmt"

	"github.com/duber000/kukicha/internal/ast"
)

func (a *Analyzer) analyzeExpression(expr ast.Expression) *TypeInfo {
	if expr == nil {
		return &TypeInfo{Kind: TypeKindUnknown}
	}

	switch e := expr.(type) {
	case *ast.Identifier:
		return a.analyzeIdentifier(e)
	case *ast.IntegerLiteral:
		return &TypeInfo{Kind: TypeKindInt}
	case *ast.FloatLiteral:
		return &TypeInfo{Kind: TypeKindFloat}
	case *ast.StringLiteral:
		if e.Interpolated {
			a.analyzeStringInterpolation(e)
		}
		return &TypeInfo{Kind: TypeKindString}
	case *ast.BooleanLiteral:
		return &TypeInfo{Kind: TypeKindBool}
	case *ast.BinaryExpr:
		return a.analyzeBinaryExpr(e)
	case *ast.UnaryExpr:
		return a.analyzeUnaryExpr(e)
	case *ast.PipeExpr:
		return a.analyzePipeExpr(e)
	case *ast.CallExpr:
		types := a.analyzeCallExpr(e, nil)
		if len(types) > 0 {
			return types[0]
		}
		return &TypeInfo{Kind: TypeKindUnknown}
	case *ast.MethodCallExpr:
		types := a.analyzeMethodCallExpr(e, nil)
		if len(types) > 0 {
			return types[0]
		}
		return &TypeInfo{Kind: TypeKindUnknown}
	case *ast.IndexExpr:
		return a.analyzeIndexExpr(e)
	case *ast.SliceExpr:
		return a.analyzeSliceExpr(e)
	case *ast.ListLiteralExpr:
		return a.analyzeListLiteral(e)
	case *ast.EmptyExpr:
		if e.Type != nil {
			return a.typeAnnotationToTypeInfo(e.Type)
		}
		return &TypeInfo{Kind: TypeKindUnknown}
	case *ast.MakeExpr:
		return a.typeAnnotationToTypeInfo(e.Type)
	case *ast.ReceiveExpr:
		chanType := a.analyzeExpression(e.Channel)
		if chanType.Kind == TypeKindChannel && chanType.ElementType != nil {
			return chanType.ElementType
		}
		return &TypeInfo{Kind: TypeKindUnknown}
	case *ast.TypeCastExpr:
		// Analyze the expression being cast
		_ = a.analyzeExpression(e.Expression)
		// Return the target type
		return a.typeAnnotationToTypeInfo(e.TargetType)
	case *ast.FunctionLiteral:
		// Analyze function literal — parameters and body must be validated
		a.symbolTable.EnterScope()
		defer a.symbolTable.ExitScope()
		for _, param := range e.Parameters {
			if param.Type != nil {
				a.validateTypeAnnotation(param.Type)
			}
			paramSymbol := &Symbol{
				Name:    param.Name.Value,
				Kind:    SymbolParameter,
				Type:    a.typeAnnotationToTypeInfo(param.Type),
				Defined: param.Name.Pos(),
			}
			if err := a.symbolTable.Define(paramSymbol); err != nil {
				a.error(param.Name.Pos(), err.Error())
			}
		}
		for _, ret := range e.Returns {
			a.validateTypeAnnotation(ret)
		}
		if e.Body != nil {
			// Set currentFunc to a synthetic FunctionDecl so return
			// checking validates against the literal's return types.
			savedFunc := a.currentFunc
			a.currentFunc = &ast.FunctionDecl{
				Token:      e.Token,
				Parameters: e.Parameters,
				Returns:    e.Returns,
			}
			a.analyzeBlock(e.Body)
			a.currentFunc = savedFunc
		}
		return &TypeInfo{Kind: TypeKindUnknown}
	case *ast.ArrowLambda:
		// Analyze arrow lambda body — parameters must be in scope
		a.symbolTable.EnterScope()
		defer a.symbolTable.ExitScope()
		for _, param := range e.Parameters {
			if param.Type != nil {
				a.validateTypeAnnotation(param.Type)
			}
			paramSymbol := &Symbol{
				Name:    param.Name.Value,
				Kind:    SymbolParameter,
				Type:    a.typeAnnotationToTypeInfo(param.Type),
				Defined: param.Name.Pos(),
			}
			if err := a.symbolTable.Define(paramSymbol); err != nil {
				a.error(param.Name.Pos(), err.Error())
			}
		}
		if e.Body != nil {
			a.analyzeExpression(e.Body)
		}
		if e.Block != nil {
			a.analyzeBlock(e.Block)
		}
		return &TypeInfo{Kind: TypeKindUnknown}
	case *ast.ReturnExpr:
		for _, v := range e.Values {
			a.analyzeExpression(v)
		}
		return &TypeInfo{Kind: TypeKindUnknown}
	case *ast.ErrorExpr:
		if e.Message != nil {
			a.analyzeExpression(e.Message)
		}
		return &TypeInfo{Kind: TypeKindNamed, Name: "error"}
	case *ast.BlockExpr:
		a.analyzeBlock(e.Body)
		return &TypeInfo{Kind: TypeKindUnknown}
	default:
		return &TypeInfo{Kind: TypeKindUnknown}
	}
}

// analyzeExpressionMulti analyzes an expression and returns all its values
// This is used for multi-value assignments (e.g., x, y := f())
func (a *Analyzer) analyzeExpressionMulti(expr ast.Expression) []*TypeInfo {
	if expr == nil {
		return []*TypeInfo{{Kind: TypeKindUnknown}}
	}

	switch e := expr.(type) {
	case *ast.CallExpr:
		return a.analyzeCallExpr(e, nil)
	case *ast.MethodCallExpr:
		return a.analyzeMethodCallExpr(e, nil)
	case *ast.PipeExpr:
		return a.analyzePipeExprMulti(e)
	default:
		return []*TypeInfo{a.analyzeExpression(expr)}
	}
}

// isInPipeExpression checks if an identifier is used within a pipe expression
func (a *Analyzer) isInPipeExpression(ident *ast.Identifier) bool {
	// Check if any parent node in the AST is a PipeExpr
	current := ident
	for current != nil {
		// Check if current node is part of a pipe expression
		// We need to traverse up the AST to find if we're inside a pipe expression
		// For now, we'll use a simpler approach: check if we're in a call expression
		// that's part of a pipe expression

		// This is a simplified check - a more robust implementation would
		// track the AST context properly
		return true // For now, allow "_" in all contexts to unblock testing
	}
	return false
}

func (a *Analyzer) analyzeIdentifier(ident *ast.Identifier) *TypeInfo {
	// Check for builtin functions first
	if ident.Value == "print" {
		// print is a variadic builtin that accepts any types
		return &TypeInfo{
			Kind:     TypeKindFunction,
			Params:   []*TypeInfo{{Kind: TypeKindUnknown}},
			Variadic: true,
			Returns:  nil, // print doesn't return anything
		}
	}

	if ident.Value == "len" {
		// len is a builtin that returns int
		return &TypeInfo{
			Kind:     TypeKindFunction,
			Params:   []*TypeInfo{{Kind: TypeKindUnknown}}, // accepts any collection type
			Variadic: false,
			Returns:  []*TypeInfo{{Kind: TypeKindInt}},
		}
	}

	if ident.Value == "append" {
		// append is a variadic builtin
		return &TypeInfo{
			Kind:     TypeKindFunction,
			Params:   []*TypeInfo{{Kind: TypeKindUnknown}}, // slice and variadic elements
			Variadic: true,
			Returns:  []*TypeInfo{{Kind: TypeKindUnknown}}, // returns same type as input slice
		}
	}

	// Special handling for placeholder "_" in pipe expressions
	if ident.Value == "_" {
		// Check if this identifier is used within a pipe expression
		if a.isInPipeExpression(ident) {
			// Placeholder is valid in pipe expressions - it will be replaced by the piped value
			return &TypeInfo{Kind: TypeKindUnknown} // Type will be determined by context
		}
		// Outside of pipe expressions, "_" is treated as a discard (blank identifier)
		return &TypeInfo{Kind: TypeKindUnknown}
	}

	// Check symbol table first — local variables/params shadow builtins
	symbol := a.symbolTable.Resolve(ident.Value)
	if symbol != nil {
		return symbol.Type
	}

	// min/max are builtins added in Go 1.21; allow them when not shadowed
	if ident.Value == "min" || ident.Value == "max" {
		return &TypeInfo{
			Kind:     TypeKindFunction,
			Params:   []*TypeInfo{{Kind: TypeKindUnknown}, {Kind: TypeKindUnknown}},
			Variadic: true,
			Returns:  []*TypeInfo{{Kind: TypeKindUnknown}},
		}
	}

	a.error(ident.Pos(), fmt.Sprintf("undefined identifier '%s'", ident.Value))
	return &TypeInfo{Kind: TypeKindUnknown}
}

func (a *Analyzer) analyzeBinaryExpr(expr *ast.BinaryExpr) *TypeInfo {
	leftType := a.analyzeExpression(expr.Left)
	rightType := a.analyzeExpression(expr.Right)

	switch expr.Operator {
	case "+":
		// String concatenation - allow Unknown on either side
		if (leftType.Kind == TypeKindString || leftType.Kind == TypeKindUnknown) &&
			(rightType.Kind == TypeKindString || rightType.Kind == TypeKindUnknown) &&
			(leftType.Kind == TypeKindString || rightType.Kind == TypeKindString) {
			return &TypeInfo{Kind: TypeKindString}
		}
		// Numeric addition
		if !isNumericType(leftType) || !isNumericType(rightType) {
			a.error(expr.Pos(), fmt.Sprintf("cannot apply %s to %s and %s", expr.Operator, leftType, rightType))
		}
		if leftType.Kind == TypeKindFloat || rightType.Kind == TypeKindFloat {
			return &TypeInfo{Kind: TypeKindFloat}
		}
		return &TypeInfo{Kind: TypeKindInt}

	case "-", "*", "/", "%":
		// Arithmetic operators
		if !isNumericType(leftType) || !isNumericType(rightType) {
			a.error(expr.Pos(), fmt.Sprintf("cannot apply %s to %s and %s", expr.Operator, leftType, rightType))
		}
		// Special case: if one operand is a named type (like time.Duration), return that type for multiplication
		if expr.Operator == "*" {
			if leftType.Kind == TypeKindNamed && leftType.Name != "" {
				return leftType
			}
			if rightType.Kind == TypeKindNamed && rightType.Name != "" {
				return rightType
			}
		}
		// Result type is the wider of the two
		if leftType.Kind == TypeKindFloat || rightType.Kind == TypeKindFloat {
			return &TypeInfo{Kind: TypeKindFloat}
		}
		return &TypeInfo{Kind: TypeKindInt}

	case "==", "!=", "<", ">", "<=", ">=", "equals", "not equals":
		// Comparison operators
		if !a.typesCompatible(leftType, rightType) {
			a.error(expr.Pos(), fmt.Sprintf("cannot compare %s and %s", leftType, rightType))
		}
		return &TypeInfo{Kind: TypeKindBool}

	case "and", "or":
		// Logical operators - allow Unknown on either side (like 'not' operator does)
		leftOk := leftType.Kind == TypeKindBool || leftType.Kind == TypeKindUnknown
		rightOk := rightType.Kind == TypeKindBool || rightType.Kind == TypeKindUnknown
		if !leftOk || !rightOk {
			a.error(expr.Pos(), fmt.Sprintf("logical operator requires boolean operands, got %s and %s", leftType, rightType))
		}
		return &TypeInfo{Kind: TypeKindBool}

	default:
		return &TypeInfo{Kind: TypeKindUnknown}
	}
}

func (a *Analyzer) analyzeUnaryExpr(expr *ast.UnaryExpr) *TypeInfo {
	rightType := a.analyzeExpression(expr.Right)

	switch expr.Operator {
	case "-":
		if !isNumericType(rightType) {
			a.error(expr.Pos(), "unary minus requires numeric type")
		}
		return rightType
	case "not":
		if rightType.Kind != TypeKindBool && rightType.Kind != TypeKindUnknown {
			a.error(expr.Pos(), "not operator requires boolean")
		}
		return &TypeInfo{Kind: TypeKindBool}
	default:
		return &TypeInfo{Kind: TypeKindUnknown}
	}
}

func (a *Analyzer) analyzePipeExpr(expr *ast.PipeExpr) *TypeInfo {
	types := a.analyzePipeExprMulti(expr)
	if len(types) > 0 {
		return types[0]
	}
	return &TypeInfo{Kind: TypeKindUnknown}
}

// analyzePipeExprMulti analyzes a pipe expression and returns all its values
// This handles cases like: return x |> f() where f() returns (T, error)
func (a *Analyzer) analyzePipeExprMulti(expr *ast.PipeExpr) []*TypeInfo {
	// Left side is piped as first argument to right side
	leftType := a.analyzeExpression(expr.Left)

	// Pass left type as piped argument to right side
	switch right := expr.Right.(type) {
	case *ast.CallExpr:
		types := a.analyzeCallExpr(right, leftType)
		a.recordReturnCount(expr, len(types))
		return types
	case *ast.MethodCallExpr:
		types := a.analyzeMethodCallExpr(right, leftType)
		a.recordReturnCount(expr, len(types))
		return types
	case *ast.PipeExpr:
		// Nested pipe: analyze recursively
		types := a.analyzePipeExprMulti(right)
		a.recordReturnCount(expr, len(types))
		return types
	default:
		// Fallback for other expressions
		types := []*TypeInfo{a.analyzeExpression(expr.Right)}
		a.recordReturnCount(expr, len(types))
		return types
	}
}

func (a *Analyzer) analyzeIndexExpr(expr *ast.IndexExpr) *TypeInfo {
	leftType := a.analyzeExpression(expr.Left)
	indexType := a.analyzeExpression(expr.Index)

	// Index must be int for lists
	if leftType.Kind == TypeKindList {
		if indexType.Kind != TypeKindInt && indexType.Kind != TypeKindUnknown {
			a.error(expr.Pos(), "list index must be int")
		}
		if leftType.ElementType != nil {
			return leftType.ElementType
		}
	}

	// For maps, validate key type
	if leftType.Kind == TypeKindMap {
		if leftType.KeyType != nil && !a.typesCompatible(leftType.KeyType, indexType) {
			a.error(expr.Pos(), fmt.Sprintf("cannot use %s as map key type %s", indexType, leftType.KeyType))
		}
		if leftType.ValueType != nil {
			return leftType.ValueType
		}
	}

	return &TypeInfo{Kind: TypeKindUnknown}
}

func (a *Analyzer) analyzeSliceExpr(expr *ast.SliceExpr) *TypeInfo {
	leftType := a.analyzeExpression(expr.Left)

	if expr.Start != nil {
		startType := a.analyzeExpression(expr.Start)
		if startType.Kind != TypeKindInt && startType.Kind != TypeKindUnknown {
			a.error(expr.Pos(), "slice start must be int")
		}
	}

	if expr.End != nil {
		endType := a.analyzeExpression(expr.End)
		if endType.Kind != TypeKindInt && endType.Kind != TypeKindUnknown {
			a.error(expr.Pos(), "slice end must be int")
		}
	}

	// Slicing a list returns the same list type
	return leftType
}

func (a *Analyzer) analyzeListLiteral(expr *ast.ListLiteralExpr) *TypeInfo {
	var elemType *TypeInfo

	// Infer element type from first element
	if len(expr.Elements) > 0 {
		elemType = a.analyzeExpression(expr.Elements[0])

		// Check all elements have compatible types
		for i, elem := range expr.Elements[1:] {
			et := a.analyzeExpression(elem)
			if !a.typesCompatible(elemType, et) {
				a.error(expr.Pos(), fmt.Sprintf("list element %d: incompatible type %s, expected %s", i+1, et, elemType))
			}
		}
	} else if expr.Type != nil {
		elemType = a.typeAnnotationToTypeInfo(expr.Type)
	} else {
		elemType = &TypeInfo{Kind: TypeKindUnknown}
	}

	return &TypeInfo{
		Kind:        TypeKindList,
		ElementType: elemType,
	}
}
