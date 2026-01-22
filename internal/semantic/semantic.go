package semantic

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/duber000/kukicha/internal/ast"
)

// Analyzer performs semantic analysis on the AST
type Analyzer struct {
	program     *ast.Program
	symbolTable *SymbolTable
	errors      []error
	currentFunc *ast.FunctionDecl // Track current function for return type checking
}

// New creates a new semantic analyzer
func New(program *ast.Program) *Analyzer {
	return &Analyzer{
		program:     program,
		symbolTable: NewSymbolTable(),
		errors:      []error{},
	}
}

// Analyze performs semantic analysis on the program
func (a *Analyzer) Analyze() []error {
	// First pass: Collect all type and interface declarations
	a.collectDeclarations()

	// Second pass: Analyze function bodies and validate
	a.analyzeDeclarations()

	return a.errors
}

// collectDeclarations collects all top-level declarations
func (a *Analyzer) collectDeclarations() {
	// Collect imports
	for _, imp := range a.program.Imports {
		var name string
		if imp.Alias != nil {
			name = imp.Alias.Value
		} else {
			// Extract package name from path
			path := strings.Trim(imp.Path.Value, "\"")
			parts := strings.Split(path, "/")
			name = parts[len(parts)-1]
		}

		err := a.symbolTable.Define(&Symbol{
			Name:    name,
			Kind:    SymbolVariable, // Treat as variable for now
			Type:    &TypeInfo{Kind: TypeKindUnknown},
			Defined: imp.Pos(),
		})
		if err != nil {
			a.error(imp.Pos(), err.Error())
		}
	}

	for _, decl := range a.program.Declarations {
		switch d := decl.(type) {
		case *ast.TypeDecl:
			a.collectTypeDecl(d)
		case *ast.InterfaceDecl:
			a.collectInterfaceDecl(d)
		case *ast.FunctionDecl:
			a.collectFunctionDecl(d)
		}
	}
}

func (a *Analyzer) collectTypeDecl(decl *ast.TypeDecl) {
	// Check export rules: PascalCase = exported, camelCase = unexported
	if !isValidIdentifier(decl.Name.Value) {
		a.error(decl.Name.Pos(), fmt.Sprintf("invalid type name '%s'", decl.Name.Value))
		return
	}

	// Add type to symbol table
	symbol := &Symbol{
		Name:     decl.Name.Value,
		Kind:     SymbolType,
		Type:     &TypeInfo{Kind: TypeKindStruct, Name: decl.Name.Value},
		Defined:  decl.Name.Pos(),
		Exported: isExported(decl.Name.Value),
	}

	if err := a.symbolTable.Define(symbol); err != nil {
		a.error(decl.Name.Pos(), err.Error())
	}
}

func (a *Analyzer) collectInterfaceDecl(decl *ast.InterfaceDecl) {
	// Check export rules
	if !isValidIdentifier(decl.Name.Value) {
		a.error(decl.Name.Pos(), fmt.Sprintf("invalid interface name '%s'", decl.Name.Value))
		return
	}

	// Add interface to symbol table
	symbol := &Symbol{
		Name:     decl.Name.Value,
		Kind:     SymbolInterface,
		Type:     &TypeInfo{Kind: TypeKindInterface, Name: decl.Name.Value},
		Defined:  decl.Name.Pos(),
		Exported: isExported(decl.Name.Value),
	}

	if err := a.symbolTable.Define(symbol); err != nil {
		a.error(decl.Name.Pos(), err.Error())
	}
}

func (a *Analyzer) collectFunctionDecl(decl *ast.FunctionDecl) {
	// Check export rules
	if !isValidIdentifier(decl.Name.Value) {
		a.error(decl.Name.Pos(), fmt.Sprintf("invalid function name '%s'", decl.Name.Value))
		return
	}

	// Build function type
	params := make([]*TypeInfo, len(decl.Parameters))
	hasVariadic := false
	for i, param := range decl.Parameters {
		params[i] = a.typeAnnotationToTypeInfo(param.Type)
		if param.Variadic {
			hasVariadic = true
		}
	}

	returns := make([]*TypeInfo, len(decl.Returns))
	for i, ret := range decl.Returns {
		returns[i] = a.typeAnnotationToTypeInfo(ret)
	}

	funcType := &TypeInfo{
		Kind:     TypeKindFunction,
		Params:   params,
		Returns:  returns,
		Variadic: hasVariadic,
	}

	// Add function to symbol table
	symbol := &Symbol{
		Name:     decl.Name.Value,
		Kind:     SymbolFunction,
		Type:     funcType,
		Defined:  decl.Name.Pos(),
		Exported: isExported(decl.Name.Value),
	}

	if err := a.symbolTable.Define(symbol); err != nil {
		a.error(decl.Name.Pos(), err.Error())
	}
}

// analyzeDeclarations performs deep analysis of declarations
func (a *Analyzer) analyzeDeclarations() {
	for _, decl := range a.program.Declarations {
		switch d := decl.(type) {
		case *ast.TypeDecl:
			a.analyzeTypeDecl(d)
		case *ast.InterfaceDecl:
			a.analyzeInterfaceDecl(d)
		case *ast.FunctionDecl:
			a.analyzeFunctionDecl(d)
		}
	}
}

func (a *Analyzer) analyzeTypeDecl(decl *ast.TypeDecl) {
	// Validate field types exist
	for _, field := range decl.Fields {
		if !isValidIdentifier(field.Name.Value) {
			a.error(field.Name.Pos(), fmt.Sprintf("invalid field name '%s'", field.Name.Value))
		}

		// Check that field type exists
		a.validateTypeAnnotation(field.Type)
	}
}

func (a *Analyzer) analyzeInterfaceDecl(decl *ast.InterfaceDecl) {
	// Validate method signatures
	for _, method := range decl.Methods {
		if !isValidIdentifier(method.Name.Value) {
			a.error(method.Name.Pos(), fmt.Sprintf("invalid method name '%s'", method.Name.Value))
		}

		// Validate parameter types
		for _, param := range method.Parameters {
			a.validateTypeAnnotation(param.Type)
		}

		// Validate return types
		for _, ret := range method.Returns {
			a.validateTypeAnnotation(ret)
		}
	}
}

func (a *Analyzer) analyzeFunctionDecl(decl *ast.FunctionDecl) {
	// Enter new scope for function
	a.symbolTable.EnterScope()
	defer a.symbolTable.ExitScope()

	// Track current function for return checking
	a.currentFunc = decl

	// Add receiver if present (for methods)
	if decl.Receiver != nil {
		a.validateTypeAnnotation(decl.Receiver.Type)

		receiverSymbol := &Symbol{
			Name:    decl.Receiver.Name.Value,
			Kind:    SymbolParameter,
			Type:    a.typeAnnotationToTypeInfo(decl.Receiver.Type),
			Defined: decl.Receiver.Name.Pos(),
		}
		if err := a.symbolTable.Define(receiverSymbol); err != nil {
			a.error(decl.Receiver.Name.Pos(), err.Error())
		}
	}

	// Validate variadic parameters (must be last, only one)
	variadicCount := 0
	for i, param := range decl.Parameters {
		if param.Variadic {
			variadicCount++
			if variadicCount > 1 {
				a.error(param.Name.Pos(), "only one variadic parameter allowed per function")
			}
			if i != len(decl.Parameters)-1 {
				a.error(param.Name.Pos(), "variadic parameter must be the last parameter")
			}
		}
	}

	// Add parameters to scope
	for _, param := range decl.Parameters {
		if !isValidIdentifier(param.Name.Value) {
			a.error(param.Name.Pos(), fmt.Sprintf("invalid parameter name '%s'", param.Name.Value))
		}

		a.validateTypeAnnotation(param.Type)

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

	// Validate return types exist
	for _, ret := range decl.Returns {
		a.validateTypeAnnotation(ret)
	}

	// Analyze function body
	if decl.Body != nil {
		a.analyzeBlock(decl.Body)
	}

	a.currentFunc = nil
}

func (a *Analyzer) analyzeBlock(block *ast.BlockStmt) {
	for _, stmt := range block.Statements {
		a.analyzeStatement(stmt)
	}
}

func (a *Analyzer) analyzeStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.VarDeclStmt:
		a.analyzeVarDeclStmt(s)
	case *ast.AssignStmt:
		a.analyzeAssignStmt(s)
	case *ast.ReturnStmt:
		a.analyzeReturnStmt(s)
	case *ast.IfStmt:
		a.analyzeIfStmt(s)
	case *ast.ForRangeStmt:
		a.analyzeForRangeStmt(s)
	case *ast.ForNumericStmt:
		a.analyzeForNumericStmt(s)
	case *ast.ForConditionStmt:
		a.analyzeForConditionStmt(s)
	case *ast.DeferStmt:
		a.analyzeExpression(s.Call)
	case *ast.GoStmt:
		a.analyzeExpression(s.Call)
	case *ast.SendStmt:
		a.analyzeExpression(s.Value)
		a.analyzeExpression(s.Channel)
	case *ast.ExpressionStmt:
		a.analyzeExpression(s.Expression)
	}
}

func (a *Analyzer) analyzeVarDeclStmt(stmt *ast.VarDeclStmt) {
	// Analyze all value expressions
	valueTypes := make([]*TypeInfo, len(stmt.Values))
	for i, val := range stmt.Values {
		valueTypes[i] = a.analyzeExpression(val)
	}

	// Check that number of values matches number of names
	// For single function call that returns multiple values, we allow len(stmt.Values) == 1
	if len(stmt.Values) != len(stmt.Names) && len(stmt.Values) != 1 {
		a.error(stmt.Pos(), fmt.Sprintf("assignment mismatch: %d variables but %d values", len(stmt.Names), len(stmt.Values)))
	}

	// Type inference and validation
	for i, name := range stmt.Names {
		if !isValidIdentifier(name.Value) {
			a.error(name.Pos(), fmt.Sprintf("invalid variable name '%s'", name.Value))
			continue
		}

		// Determine the type for this variable
		var varType *TypeInfo
		if stmt.Type != nil {
			// Explicit type annotation applies to all variables
			a.validateTypeAnnotation(stmt.Type)
			varType = a.typeAnnotationToTypeInfo(stmt.Type)
		} else if len(stmt.Values) == len(stmt.Names) {
			// One value per variable: use corresponding value type
			varType = valueTypes[i]
		} else if len(stmt.Values) == 1 {
			// Single expression (likely multi-value function call)
			// For now, we use the first value type as a fallback
			varType = valueTypes[0]
		} else {
			varType = &TypeInfo{Kind: TypeKindUnknown}
		}

		// Check type compatibility if explicit type is specified
		if stmt.Type != nil && len(stmt.Values) == len(stmt.Names) {
			if !a.typesCompatible(varType, valueTypes[i]) {
				a.error(stmt.Pos(), fmt.Sprintf("cannot assign %s to %s", valueTypes[i], varType))
			}
		}

		// Add variable to symbol table
		symbol := &Symbol{
			Name:    name.Value,
			Kind:    SymbolVariable,
			Type:    varType,
			Defined: name.Pos(),
			Mutable: true,
		}
		if err := a.symbolTable.Define(symbol); err != nil {
			a.error(name.Pos(), err.Error())
		}
	}
}

func (a *Analyzer) analyzeAssignStmt(stmt *ast.AssignStmt) {
	// Analyze all target and value expressions
	targetTypes := make([]*TypeInfo, len(stmt.Targets))
	for i, target := range stmt.Targets {
		targetTypes[i] = a.analyzeExpression(target)
	}

	valueTypes := make([]*TypeInfo, len(stmt.Values))
	for i, val := range stmt.Values {
		valueTypes[i] = a.analyzeExpression(val)
	}

	// Check that number of values matches number of targets
	// For single function call that returns multiple values, we allow len(stmt.Values) == 1
	if len(stmt.Values) != len(stmt.Targets) && len(stmt.Values) != 1 {
		a.error(stmt.Pos(), fmt.Sprintf("assignment mismatch: %d variables but %d values", len(stmt.Targets), len(stmt.Values)))
		return
	}

	// Type compatibility checking
	if len(stmt.Values) == len(stmt.Targets) {
		// One value per target: check each pair
		for i := range stmt.Targets {
			if !a.typesCompatible(targetTypes[i], valueTypes[i]) {
				a.error(stmt.Pos(), fmt.Sprintf("cannot assign %s to %s", valueTypes[i], targetTypes[i]))
			}
		}
	}
	// If len(stmt.Values) == 1, it's likely a multi-value function call
	// We skip detailed checking for now as it requires tracking function return types
}

func (a *Analyzer) analyzeReturnStmt(stmt *ast.ReturnStmt) {
	if a.currentFunc == nil {
		a.error(stmt.Pos(), "return statement outside of function")
		return
	}

	// Check return value count
	if len(stmt.Values) != len(a.currentFunc.Returns) {
		a.error(stmt.Pos(), fmt.Sprintf("expected %d return values, got %d", len(a.currentFunc.Returns), len(stmt.Values)))
		return
	}

	// Check return value types
	for i, value := range stmt.Values {
		valueType := a.analyzeExpression(value)
		expectedType := a.typeAnnotationToTypeInfo(a.currentFunc.Returns[i])

		if !a.typesCompatible(expectedType, valueType) {
			a.error(stmt.Pos(), fmt.Sprintf("cannot return %s as %s", valueType, expectedType))
		}
	}
}

func (a *Analyzer) analyzeIfStmt(stmt *ast.IfStmt) {
	// Analyze condition
	condType := a.analyzeExpression(stmt.Condition)
	if condType.Kind != TypeKindBool && condType.Kind != TypeKindUnknown {
		a.error(stmt.Pos(), "if condition must be boolean")
	}

	// Analyze consequence
	a.symbolTable.EnterScope()
	a.analyzeBlock(stmt.Consequence)
	a.symbolTable.ExitScope()

	// Analyze alternative
	if stmt.Alternative != nil {
		a.symbolTable.EnterScope()
		switch alt := stmt.Alternative.(type) {
		case *ast.ElseStmt:
			a.analyzeBlock(alt.Body)
		case *ast.IfStmt:
			a.analyzeIfStmt(alt)
		}
		a.symbolTable.ExitScope()
	}
}

func (a *Analyzer) analyzeForRangeStmt(stmt *ast.ForRangeStmt) {
	// Analyze collection
	collType := a.analyzeExpression(stmt.Collection)

	a.symbolTable.EnterScope()
	defer a.symbolTable.ExitScope()

	// Add loop variables to scope
	if stmt.Index != nil {
		indexSymbol := &Symbol{
			Name:    stmt.Index.Value,
			Kind:    SymbolVariable,
			Type:    &TypeInfo{Kind: TypeKindInt},
			Defined: stmt.Index.Pos(),
			Mutable: true,
		}
		a.symbolTable.Define(indexSymbol)
	}

	// Determine element type from collection type
	var elemType *TypeInfo
	if collType.Kind == TypeKindList && collType.ElementType != nil {
		elemType = collType.ElementType
	} else {
		elemType = &TypeInfo{Kind: TypeKindUnknown}
	}

	varSymbol := &Symbol{
		Name:    stmt.Variable.Value,
		Kind:    SymbolVariable,
		Type:    elemType,
		Defined: stmt.Variable.Pos(),
		Mutable: true,
	}
	a.symbolTable.Define(varSymbol)

	// Analyze body
	a.analyzeBlock(stmt.Body)
}

func (a *Analyzer) analyzeForNumericStmt(stmt *ast.ForNumericStmt) {
	// Analyze start and end expressions
	startType := a.analyzeExpression(stmt.Start)
	endType := a.analyzeExpression(stmt.End)

	if startType.Kind != TypeKindInt && startType.Kind != TypeKindUnknown {
		a.error(stmt.Pos(), "for loop start must be int")
	}
	if endType.Kind != TypeKindInt && endType.Kind != TypeKindUnknown {
		a.error(stmt.Pos(), "for loop end must be int")
	}

	a.symbolTable.EnterScope()
	defer a.symbolTable.ExitScope()

	// Add loop variable to scope
	varSymbol := &Symbol{
		Name:    stmt.Variable.Value,
		Kind:    SymbolVariable,
		Type:    &TypeInfo{Kind: TypeKindInt},
		Defined: stmt.Variable.Pos(),
		Mutable: true,
	}
	a.symbolTable.Define(varSymbol)

	// Analyze body
	a.analyzeBlock(stmt.Body)
}

func (a *Analyzer) analyzeForConditionStmt(stmt *ast.ForConditionStmt) {
	// Analyze condition
	condType := a.analyzeExpression(stmt.Condition)
	if condType.Kind != TypeKindBool && condType.Kind != TypeKindUnknown {
		a.error(stmt.Pos(), "for condition must be boolean")
	}

	a.symbolTable.EnterScope()
	defer a.symbolTable.ExitScope()

	// Analyze body
	a.analyzeBlock(stmt.Body)
}

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
	case *ast.OnErrExpr:
		return a.analyzeOnErrExpr(e)
	case *ast.CallExpr:
		return a.analyzeCallExpr(e)
	case *ast.MethodCallExpr:
		return a.analyzeMethodCallExpr(e)
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
	default:
		return &TypeInfo{Kind: TypeKindUnknown}
	}
}

func (a *Analyzer) analyzeIdentifier(ident *ast.Identifier) *TypeInfo {
	symbol := a.symbolTable.Resolve(ident.Value)
	if symbol == nil {
		a.error(ident.Pos(), fmt.Sprintf("undefined identifier '%s'", ident.Value))
		return &TypeInfo{Kind: TypeKindUnknown}
	}
	return symbol.Type
}

func (a *Analyzer) analyzeBinaryExpr(expr *ast.BinaryExpr) *TypeInfo {
	leftType := a.analyzeExpression(expr.Left)
	rightType := a.analyzeExpression(expr.Right)

	switch expr.Operator {
	case "+", "-", "*", "/", "%":
		// Arithmetic operators
		if !isNumericType(leftType) || !isNumericType(rightType) {
			a.error(expr.Pos(), fmt.Sprintf("cannot apply %s to %s and %s", expr.Operator, leftType, rightType))
		}
		// Result type is the wider of the two
		if leftType.Kind == TypeKindFloat || rightType.Kind == TypeKindFloat {
			return &TypeInfo{Kind: TypeKindFloat}
		}
		return &TypeInfo{Kind: TypeKindInt}

	case "==", "!=", "<", ">", "<=", ">=":
		// Comparison operators
		if !a.typesCompatible(leftType, rightType) {
			a.error(expr.Pos(), fmt.Sprintf("cannot compare %s and %s", leftType, rightType))
		}
		return &TypeInfo{Kind: TypeKindBool}

	case "and", "or":
		// Logical operators
		if leftType.Kind != TypeKindBool || rightType.Kind != TypeKindBool {
			a.error(expr.Pos(), fmt.Sprintf("logical operator requires boolean operands"))
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
	// Left side is piped as first argument to right side (which must be a call)
	a.analyzeExpression(expr.Left)
	return a.analyzeExpression(expr.Right)
}

func (a *Analyzer) analyzeOnErrExpr(expr *ast.OnErrExpr) *TypeInfo {
	leftType := a.analyzeExpression(expr.Left)
	a.analyzeExpression(expr.Handler)
	// Returns the same type as left expression (the success case)
	return leftType
}

func (a *Analyzer) analyzeCallExpr(expr *ast.CallExpr) *TypeInfo {
	funcType := a.analyzeExpression(expr.Function)

	// If it's a known function, validate arguments
	if funcType.Kind == TypeKindFunction {
		// Validate argument count
		if funcType.Variadic {
			// Variadic: must have at least (params - 1) arguments
			minArgs := len(funcType.Params) - 1
			if len(expr.Arguments) < minArgs {
				a.error(expr.Pos(), fmt.Sprintf("expected at least %d arguments, got %d", minArgs, len(expr.Arguments)))
			}
		} else {
			// Non-variadic: must have exact number of arguments
			if len(expr.Arguments) != len(funcType.Params) {
				a.error(expr.Pos(), fmt.Sprintf("expected %d arguments, got %d", len(funcType.Params), len(expr.Arguments)))
			}
		}

		// Validate argument types
		for i, arg := range expr.Arguments {
			argType := a.analyzeExpression(arg)

			// For variadic, all args beyond params-1 match the last param type
			paramIndex := i
			if funcType.Variadic && i >= len(funcType.Params)-1 {
				paramIndex = len(funcType.Params) - 1
			}

			if paramIndex < len(funcType.Params) && !a.typesCompatible(funcType.Params[paramIndex], argType) {
				a.error(expr.Pos(), fmt.Sprintf("argument %d: cannot use %s as %s", i+1, argType, funcType.Params[paramIndex]))
			}
		}

		// Return first return type (if any)
		if len(funcType.Returns) > 0 {
			return funcType.Returns[0]
		}
	}

	return &TypeInfo{Kind: TypeKindUnknown}
}

func (a *Analyzer) analyzeMethodCallExpr(expr *ast.MethodCallExpr) *TypeInfo {
	// Analyze object
	a.analyzeExpression(expr.Object)

	// Analyze arguments
	for _, arg := range expr.Arguments {
		a.analyzeExpression(arg)
	}

	// For now, return unknown - full method resolution requires more complex type system
	return &TypeInfo{Kind: TypeKindUnknown}
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

func (a *Analyzer) analyzeStringInterpolation(lit *ast.StringLiteral) {
	// Parse string interpolations and validate expressions
	re := regexp.MustCompile(`\{([^}]+)\}`)
	matches := re.FindAllStringSubmatch(lit.Value, -1)

	for _, match := range matches {
		exprStr := match[1]
		// For now, just validate it's not empty
		// Full expression parsing would require parsing the expression string
		if strings.TrimSpace(exprStr) == "" {
			a.error(lit.Pos(), "empty expression in string interpolation")
		}
	}
}

// validateTypeAnnotation checks that a type annotation is valid
func (a *Analyzer) validateTypeAnnotation(typeAnn ast.TypeAnnotation) {
	switch t := typeAnn.(type) {
	case *ast.NamedType:
		// Allow built-in Go types
		builtInTypes := map[string]bool{
			"interface{}": true,
			"any":         true,
			"error":       true,
			"byte":        true,
			"rune":        true,
		}
		if builtInTypes[t.Name] {
			return // Built-in type is valid
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

func (a *Analyzer) error(pos ast.Position, message string) {
	err := fmt.Errorf("%s:%d:%d: %s", pos.File, pos.Line, pos.Column, message)
	a.errors = append(a.errors, err)
}

// Helper functions

func isValidIdentifier(name string) bool {
	if len(name) == 0 {
		return false
	}
	// Must start with letter and contain only letters, digits, underscores
	for i, r := range name {
		if i == 0 {
			if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_') {
				return false
			}
		} else {
			if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_') {
				return false
			}
		}
	}
	return true
}

func isExported(name string) bool {
	if len(name) == 0 {
		return false
	}
	// Exported if starts with uppercase letter
	r := rune(name[0])
	return r >= 'A' && r <= 'Z'
}

func isNumericType(t *TypeInfo) bool {
	return t.Kind == TypeKindInt || t.Kind == TypeKindFloat || t.Kind == TypeKindUnknown
}

func primitiveTypeFromString(name string) *TypeInfo {
	switch name {
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64":
		return &TypeInfo{Kind: TypeKindInt}
	case "float32", "float64":
		return &TypeInfo{Kind: TypeKindFloat}
	case "string":
		return &TypeInfo{Kind: TypeKindString}
	case "bool":
		return &TypeInfo{Kind: TypeKindBool}
	case "byte":
		return &TypeInfo{Kind: TypeKindInt}
	case "rune":
		return &TypeInfo{Kind: TypeKindInt}
	default:
		return &TypeInfo{Kind: TypeKindUnknown}
	}
}
