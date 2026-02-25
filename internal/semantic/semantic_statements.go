package semantic

import (
	"fmt"

	"github.com/duber000/kukicha/internal/ast"
)

func (a *Analyzer) analyzeBlock(block *ast.BlockStmt) {
	for _, stmt := range block.Statements {
		a.analyzeStatement(stmt)
	}
}

func (a *Analyzer) analyzeStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.VarDeclStmt:
		a.analyzeVarDeclStmt(s)
		a.analyzeOnErrClause(s.OnErr)
	case *ast.AssignStmt:
		a.analyzeAssignStmt(s)
		a.analyzeOnErrClause(s.OnErr)
	case *ast.ReturnStmt:
		a.analyzeReturnStmt(s)
	case *ast.IfStmt:
		a.analyzeIfStmt(s)
	case *ast.SwitchStmt:
		a.analyzeSwitchStmt(s)
	case *ast.TypeSwitchStmt:
		a.analyzeTypeSwitchStmt(s)
	case *ast.ForRangeStmt:
		a.analyzeForRangeStmt(s)
	case *ast.ForNumericStmt:
		a.analyzeForNumericStmt(s)
	case *ast.ForConditionStmt:
		a.analyzeForConditionStmt(s)
	case *ast.DeferStmt:
		a.analyzeExpression(s.Call)
	case *ast.GoStmt:
		if s.Call != nil {
			a.analyzeExpression(s.Call)
		}
		if s.Block != nil {
			a.analyzeBlock(s.Block)
		}
	case *ast.SendStmt:
		a.analyzeExpression(s.Value)
		a.analyzeExpression(s.Channel)
	case *ast.SelectStmt:
		a.switchDepth++ // reuse switchDepth so break works inside select
		defer func() { a.switchDepth-- }()
		for _, c := range s.Cases {
			a.symbolTable.EnterScope()
			if c.Recv != nil {
				a.analyzeExpression(c.Recv)
				for _, binding := range c.Bindings {
					sym := &Symbol{
						Name:    binding,
						Kind:    SymbolVariable,
						Type:    &TypeInfo{Kind: TypeKindUnknown},
						Defined: ast.Position{Line: c.Token.Line, Column: c.Token.Column, File: c.Token.File},
					}
					a.symbolTable.Define(sym)
				}
			}
			if c.Send != nil {
				a.analyzeExpression(c.Send.Value)
				a.analyzeExpression(c.Send.Channel)
			}
			a.analyzeBlock(c.Body)
			a.symbolTable.ExitScope()
		}
		if s.Otherwise != nil {
			a.analyzeBlock(s.Otherwise.Body)
		}
	case *ast.ExpressionStmt:
		a.analyzeExpression(s.Expression)
		a.analyzeOnErrClause(s.OnErr)
	case *ast.ContinueStmt:
		if a.loopDepth == 0 {
			a.error(s.Pos(), "continue statement outside of loop")
		}
	case *ast.BreakStmt:
		if a.loopDepth == 0 && a.switchDepth == 0 {
			a.error(s.Pos(), "break statement outside of loop")
		}
	}
}

func (a *Analyzer) analyzeSwitchStmt(stmt *ast.SwitchStmt) {
	if stmt.Expression != nil {
		a.analyzeExpression(stmt.Expression)
	}

	a.switchDepth++
	defer func() { a.switchDepth-- }()

	for _, c := range stmt.Cases {
		for _, val := range c.Values {
			valType := a.analyzeExpression(val)
			if stmt.Expression == nil && valType != nil && valType.Kind != TypeKindBool && valType.Kind != TypeKindUnknown {
				a.error(val.Pos(), "switch condition branch must be bool")
			}
		}
		a.analyzeBlock(c.Body)
	}

	if stmt.Otherwise != nil {
		a.analyzeBlock(stmt.Otherwise.Body)
	}
}

func (a *Analyzer) analyzeTypeSwitchStmt(stmt *ast.TypeSwitchStmt) {
	a.analyzeExpression(stmt.Expression)

	a.switchDepth++
	defer func() { a.switchDepth-- }()

	for _, c := range stmt.Cases {
		// Define the binding variable in a new scope for each case body
		a.symbolTable.EnterScope()
		bindingSymbol := &Symbol{
			Name:    stmt.Binding.Value,
			Kind:    SymbolVariable,
			Type:    &TypeInfo{Kind: TypeKindUnknown},
			Defined: stmt.Binding.Pos(),
		}
		a.symbolTable.Define(bindingSymbol)
		a.analyzeBlock(c.Body)
		a.symbolTable.ExitScope()
	}

	if stmt.Otherwise != nil {
		a.symbolTable.EnterScope()
		bindingSymbol := &Symbol{
			Name:    stmt.Binding.Value,
			Kind:    SymbolVariable,
			Type:    &TypeInfo{Kind: TypeKindUnknown},
			Defined: stmt.Binding.Pos(),
		}
		a.symbolTable.Define(bindingSymbol)
		a.analyzeBlock(stmt.Otherwise.Body)
		a.symbolTable.ExitScope()
	}
}

func (a *Analyzer) analyzeVarDeclStmt(stmt *ast.VarDeclStmt) {
	// Analyze all value expressions
	valueTypes := make([]*TypeInfo, len(stmt.Values))
	for i, val := range stmt.Values {
		valueTypes[i] = a.analyzeExpression(val)
	}

	// Special handling for multi-value return from single function call or type assertion
	var multiValueTypes []*TypeInfo
	if len(stmt.Values) == 1 && len(stmt.Names) > 1 {
		// Check if this is a type assertion (e.g., value, ok := expr as Type)
		if len(stmt.Names) == 2 {
			if typeCast, ok := stmt.Values[0].(*ast.TypeCastExpr); ok {
				// Type assertion returns (value, bool)
				targetType := a.typeAnnotationToTypeInfo(typeCast.TargetType)
				multiValueTypes = []*TypeInfo{
					targetType,
					{Kind: TypeKindBool},
				}
			} else {
				// Regular multi-value return
				multiValueTypes = a.analyzeExpressionMulti(stmt.Values[0])
			}
		} else {
			// Regular multi-value return
			multiValueTypes = a.analyzeExpressionMulti(stmt.Values[0])
		}

		if len(multiValueTypes) != len(stmt.Names) {
			// If we can't match exact count, check if it's dynamic/unknown
			if len(multiValueTypes) == 1 && multiValueTypes[0].Kind == TypeKindUnknown {
				// Allow assignment of Unknown to multiple variables
			} else {
				a.error(stmt.Pos(), fmt.Sprintf("assignment mismatch: %d variables but %d values", len(stmt.Names), len(multiValueTypes)))
			}
		}
	} else {
		// Check that number of values matches number of names
		if len(stmt.Values) != len(stmt.Names) {
			a.error(stmt.Pos(), fmt.Sprintf("assignment mismatch: %d variables but %d values", len(stmt.Names), len(stmt.Values)))
		}
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
			if multiValueTypes != nil {
				if i < len(multiValueTypes) {
					varType = multiValueTypes[i]
				} else if len(multiValueTypes) == 1 && multiValueTypes[0].Kind == TypeKindUnknown {
					varType = multiValueTypes[0]
				} else {
					varType = &TypeInfo{Kind: TypeKindUnknown}
				}
			} else {
				// Fallback (shouldn't happen with correct logic above)
				varType = valueTypes[0]
			}
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

	// Special handling for multi-value return from single function call or type assertion
	var multiValueTypes []*TypeInfo
	if len(stmt.Values) == 1 && len(stmt.Targets) > 1 {
		// Check if this is a type assertion (e.g., value, ok := expr as Type)
		if len(stmt.Targets) == 2 {
			if typeCast, ok := stmt.Values[0].(*ast.TypeCastExpr); ok {
				// Type assertion returns (value, bool)
				targetType := a.typeAnnotationToTypeInfo(typeCast.TargetType)
				multiValueTypes = []*TypeInfo{
					targetType,
					{Kind: TypeKindBool},
				}
			} else {
				// Regular multi-value return
				multiValueTypes = a.analyzeExpressionMulti(stmt.Values[0])
			}
		} else {
			// Regular multi-value return
			multiValueTypes = a.analyzeExpressionMulti(stmt.Values[0])
		}

		if len(multiValueTypes) != len(stmt.Targets) {
			// If we can't match exact count, check if it's dynamic/unknown
			if len(multiValueTypes) == 1 && multiValueTypes[0].Kind == TypeKindUnknown {
				// Allow assignment of Unknown to multiple variables
			} else {
				a.error(stmt.Pos(), fmt.Sprintf("assignment mismatch: %d variables but %d values", len(stmt.Targets), len(multiValueTypes)))
				return
			}
		}

		// Check types for multi-value assignment
		for i := range stmt.Targets {
			var valType *TypeInfo
			if i < len(multiValueTypes) {
				valType = multiValueTypes[i]
			} else {
				valType = multiValueTypes[0] // Fallback for Unknown
			}

			if !a.typesCompatible(targetTypes[i], valType) {
				a.error(stmt.Pos(), fmt.Sprintf("cannot assign %s to %s", valType, targetTypes[i]))
			}
		}
		return
	}

	// Check that number of values matches number of targets
	if len(stmt.Values) != len(stmt.Targets) {
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
}

func (a *Analyzer) analyzeReturnStmt(stmt *ast.ReturnStmt) {
	if a.currentFunc == nil {
		a.error(stmt.Pos(), "return statement outside of function")
		return
	}

	// Special handling for multi-value return from single expression (e.g., pipe expression)
	var valueTypes []*TypeInfo
	if len(stmt.Values) == 1 && len(a.currentFunc.Returns) > 1 {
		valueTypes = a.analyzeExpressionMulti(stmt.Values[0])

		if len(valueTypes) != len(a.currentFunc.Returns) {
			// If we can't match exact count, check if it's dynamic/unknown
			if len(valueTypes) == 1 && valueTypes[0].Kind == TypeKindUnknown {
				// Allow return of Unknown to multiple return positions
			} else {
				a.error(stmt.Pos(), fmt.Sprintf("expected %d return values, got %d", len(a.currentFunc.Returns), len(valueTypes)))
				return
			}
		}

		// Check types for multi-value return
		for i := range a.currentFunc.Returns {
			var valType *TypeInfo
			if i < len(valueTypes) {
				valType = valueTypes[i]
			} else {
				valType = valueTypes[0] // Fallback for Unknown
			}
			expectedType := a.typeAnnotationToTypeInfo(a.currentFunc.Returns[i])

			if !a.typesCompatible(expectedType, valType) {
				a.error(stmt.Pos(), fmt.Sprintf("cannot return %s as %s", valType, expectedType))
			}
		}
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
	a.loopDepth++
	defer func() { a.loopDepth-- }()

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
	a.loopDepth++
	defer func() { a.loopDepth-- }()

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
	a.loopDepth++
	defer func() { a.loopDepth-- }()

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
