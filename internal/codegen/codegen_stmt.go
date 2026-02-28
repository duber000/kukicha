package codegen

import (
	"fmt"
	"strings"

	"github.com/duber000/kukicha/internal/ast"
)

func (g *Generator) generateBlock(block *ast.BlockStmt) {
	for _, stmt := range block.Statements {
		g.generateStatement(stmt)
	}
}

func (g *Generator) generateStatement(stmt ast.Statement) {
	g.emitLineDirective(stmt.Pos())
	switch s := stmt.(type) {
	case *ast.VarDeclStmt:
		g.generateVarDeclStmt(s)
	case *ast.AssignStmt:
		g.generateAssignStmt(s)
	case *ast.IncDecStmt:
		g.generateIncDecStmt(s)
	case *ast.ReturnStmt:
		g.generateReturnStmt(s)
	case *ast.IfStmt:
		g.generateIfStmt(s)
	case *ast.SwitchStmt:
		g.generateSwitchStmt(s)
	case *ast.SelectStmt:
		g.generateSelectStmt(s)
	case *ast.TypeSwitchStmt:
		g.generateTypeSwitchStmt(s)
	case *ast.ForRangeStmt:
		g.generateForRangeStmt(s)
	case *ast.ForNumericStmt:
		g.generateForNumericStmt(s)
	case *ast.ForConditionStmt:
		g.generateForConditionStmt(s)
	case *ast.DeferStmt:
		g.writeLine("defer " + g.exprToString(s.Call))
	case *ast.GoStmt:
		if s.Block != nil {
			// Block form: go NEWLINE INDENT ... DEDENT
			// Generates: go func() { ... }()
			g.write(g.indentStr() + "go func() {\n")
			g.indent++
			for _, stmt := range s.Block.Statements {
				g.generateStatement(stmt)
			}
			g.indent--
			g.write(g.indentStr() + "}()\n")
		} else {
			g.writeLine("go " + g.exprToString(s.Call))
		}
	case *ast.SendStmt:
		channel := g.exprToString(s.Channel)
		value := g.exprToString(s.Value)
		g.writeLine(fmt.Sprintf("%s <- %s", channel, value))
	case *ast.ContinueStmt:
		g.writeLine("continue")
	case *ast.BreakStmt:
		g.writeLine("break")
	case *ast.ExpressionStmt:
		if s.OnErr != nil {
			g.generateOnErrStmt(s.Expression, s.OnErr)
		} else {
			g.writeLine(g.exprToString(s.Expression))
		}
	}
}

func (g *Generator) generateVarDeclStmt(stmt *ast.VarDeclStmt) {
	// Check for onerr clause on the statement
	if stmt.OnErr != nil {
		g.generateOnErrVarDecl(stmt.Names, stmt.Values, stmt.OnErr)
		return
	}

	// Special case: typed empty with interface type needs var declaration
	// e.g., x := empty io.Reader → var x io.Reader (nil by default)
	if len(stmt.Names) == 1 && len(stmt.Values) == 1 {
		if emptyExpr, ok := stmt.Values[0].(*ast.EmptyExpr); ok {
			if emptyExpr.Type != nil {
				targetType := g.generateTypeAnnotation(emptyExpr.Type)
				if g.isLikelyInterfaceType(targetType) {
					g.writeLine(fmt.Sprintf("var %s %s", stmt.Names[0].Value, targetType))
					return
				}
			} else {
				// Untyped empty → var x any
				g.writeLine(fmt.Sprintf("var %s any", stmt.Names[0].Value))
				return
			}
		}
	}

	// Build comma-separated list of names
	names := make([]string, len(stmt.Names))
	for i, n := range stmt.Names {
		names[i] = n.Value
	}
	namesStr := strings.Join(names, ", ")

	// Build comma-separated list of values
	values := make([]string, len(stmt.Values))
	for i, v := range stmt.Values {
		// Special case: multi-value declaration with TypeCastExpr should use assertion syntax
		// e.g., val, ok := x as Type -> val, ok := x.(Type)
		if len(stmt.Names) == 2 && len(stmt.Values) == 1 {
			if typeCast, ok := v.(*ast.TypeCastExpr); ok {
				targetType := g.generateTypeAnnotation(typeCast.TargetType)
				expr := g.exprToString(typeCast.Expression)
				values[i] = fmt.Sprintf("%s.(%s)", expr, targetType)
				continue
			}
		}
		values[i] = g.exprToString(v)
	}
	valuesStr := strings.Join(values, ", ")

	if stmt.Type != nil {
		// Explicit type declaration
		varType := g.generateTypeAnnotation(stmt.Type)
		g.writeLine(fmt.Sprintf("var %s %s = %s", namesStr, varType, valuesStr))
	} else {
		// Type inference with :=
		g.writeLine(fmt.Sprintf("%s := %s", namesStr, valuesStr))
	}
}

func (g *Generator) generateAssignStmt(stmt *ast.AssignStmt) {
	// Check for onerr clause on assignment
	if stmt.OnErr != nil {
		g.generateOnErrAssign(stmt)
		return
	}

	// Build comma-separated list of targets
	targets := make([]string, len(stmt.Targets))
	for i, t := range stmt.Targets {
		targets[i] = g.exprToString(t)
	}
	targetsStr := strings.Join(targets, ", ")

	// Build comma-separated list of values
	values := make([]string, len(stmt.Values))
	for i, v := range stmt.Values {
		// Special case: multi-value assignment with TypeCastExpr should use assertion syntax
		// e.g., val, ok := x as Type -> val, ok := x.(Type)
		if len(stmt.Targets) == 2 && len(stmt.Values) == 1 {
			if typeCast, ok := v.(*ast.TypeCastExpr); ok {
				targetType := g.generateTypeAnnotation(typeCast.TargetType)
				expr := g.exprToString(typeCast.Expression)
				values[i] = fmt.Sprintf("%s.(%s)", expr, targetType)
				continue
			}
		}
		values[i] = g.exprToString(v)
	}
	valuesStr := strings.Join(values, ", ")

	g.writeLine(fmt.Sprintf("%s = %s", targetsStr, valuesStr))
}

func (g *Generator) generateIncDecStmt(stmt *ast.IncDecStmt) {
	variable := g.exprToString(stmt.Variable)
	g.writeLine(fmt.Sprintf("%s%s", variable, stmt.Operator))
}

func (g *Generator) generateReturnStmt(stmt *ast.ReturnStmt) {
	if len(stmt.Values) == 0 {
		g.writeLine("return")
		return
	}

	values := make([]string, len(stmt.Values))
	for i, val := range stmt.Values {
		g.currentReturnIndex = i
		valStr := g.exprToString(val)

		// Apply type coercion if we have matching return types
		// This handles cases like: return n * 1000 -> return time.Duration(n * 1000)
		if i < len(g.currentReturnTypes) {
			valStr = g.coerceReturnValue(valStr, val, g.currentReturnTypes[i])
		}

		values[i] = valStr
	}

	g.currentReturnIndex = -1
	g.writeLine(fmt.Sprintf("return %s", strings.Join(values, ", ")))
}

// coerceReturnValue wraps a return value in a type conversion if needed
// This handles cases where Go requires explicit conversion to named types
func (g *Generator) coerceReturnValue(valStr string, val ast.Expression, returnType ast.TypeAnnotation) string {
	// Only coerce for named types (like time.Duration)
	namedType, ok := returnType.(*ast.NamedType)
	if !ok {
		return valStr
	}

	typeName := g.generateTypeAnnotation(returnType)

	// Don't wrap if it's already a type cast to this type
	if cast, ok := val.(*ast.TypeCastExpr); ok {
		castType := g.generateTypeAnnotation(cast.TargetType)
		if castType == typeName {
			return valStr
		}
	}

	// Don't wrap if it's a function call that likely returns the right type
	// (the function's return type should match)
	if _, ok := val.(*ast.CallExpr); ok {
		return valStr
	}
	if _, ok := val.(*ast.MethodCallExpr); ok {
		return valStr
	}

	// Don't wrap identifiers - they might already be the right type
	if _, ok := val.(*ast.Identifier); ok {
		return valStr
	}

	// Don't wrap if it's an empty expression (like time.Time{})
	if _, ok := val.(*ast.EmptyExpr); ok {
		return valStr
	}

	// For arithmetic expressions on numeric types returning a named numeric type,
	// wrap in the type conversion (e.g., time.Duration)
	if _, ok := val.(*ast.BinaryExpr); ok {
		// Check if this is a stdlib named type that needs wrapping
		if strings.Contains(namedType.Name, ".") {
			return fmt.Sprintf("%s(%s)", typeName, valStr)
		}
	}

	return valStr
}

func (g *Generator) generateIfStmt(stmt *ast.IfStmt) {
	if stmt.Init != nil {
		g.write("if ")
		// Use a separate generator to avoid adding newline to main output
		tempGen := New(g.program)
		tempGen.indent = 0
		tempGen.generateStatement(stmt.Init)
		initStr := strings.TrimSpace(tempGen.output.String())
		g.write(initStr)
		g.write("; ")
		g.write(g.exprToString(stmt.Condition))
		g.writeLine(" {")
	} else {
		condition := g.exprToString(stmt.Condition)
		g.writeLine(fmt.Sprintf("if %s {", condition))
	}

	g.indent++
	g.generateBlock(stmt.Consequence)
	g.indent--

	if stmt.Alternative != nil {
		switch alt := stmt.Alternative.(type) {
		case *ast.ElseStmt:
			g.writeLine("} else {")
			g.indent++
			g.generateBlock(alt.Body)
			g.indent--
			g.writeLine("}")
		case *ast.IfStmt:
			g.write(g.indentStr() + "} else ")
			g.generateIfStmtContinued(alt)
			return // Don't write closing brace, it's handled recursively
		}
	} else {
		g.writeLine("}")
	}
}

func (g *Generator) generateIfStmtContinued(stmt *ast.IfStmt) {
	condition := g.exprToString(stmt.Condition)
	g.output.WriteString(fmt.Sprintf("if %s {\n", condition))

	g.indent++
	g.generateBlock(stmt.Consequence)
	g.indent--

	if stmt.Alternative != nil {
		switch alt := stmt.Alternative.(type) {
		case *ast.ElseStmt:
			g.writeLine("} else {")
			g.indent++
			g.generateBlock(alt.Body)
			g.indent--
			g.writeLine("}")
		case *ast.IfStmt:
			g.write(g.indentStr() + "} else ")
			g.generateIfStmtContinued(alt)
			return
		}
	} else {
		g.writeLine("}")
	}
}

func (g *Generator) generateSwitchStmt(stmt *ast.SwitchStmt) {
	if stmt.Expression != nil {
		g.writeLine(fmt.Sprintf("switch %s {", g.exprToString(stmt.Expression)))
	} else {
		g.writeLine("switch {")
	}

	g.indent++
	for _, c := range stmt.Cases {
		caseValues := make([]string, len(c.Values))
		for i, value := range c.Values {
			caseValues[i] = g.exprToString(value)
		}
		g.writeLine(fmt.Sprintf("case %s:", strings.Join(caseValues, ", ")))

		g.indent++
		g.generateBlock(c.Body)
		g.indent--
	}

	if stmt.Otherwise != nil {
		g.writeLine("default:")
		g.indent++
		g.generateBlock(stmt.Otherwise.Body)
		g.indent--
	}
	g.indent--
	g.writeLine("}")
}

func (g *Generator) generateSelectStmt(stmt *ast.SelectStmt) {
	g.writeLine("select {")
	g.indent++
	for _, c := range stmt.Cases {
		var commStr string
		if c.Recv != nil {
			ch := g.exprToString(c.Recv.Channel)
			switch len(c.Bindings) {
			case 0:
				commStr = fmt.Sprintf("case <-%s:", ch)
			case 1:
				commStr = fmt.Sprintf("case %s := <-%s:", c.Bindings[0], ch)
			case 2:
				commStr = fmt.Sprintf("case %s, %s := <-%s:", c.Bindings[0], c.Bindings[1], ch)
			}
		} else if c.Send != nil {
			ch := g.exprToString(c.Send.Channel)
			val := g.exprToString(c.Send.Value)
			commStr = fmt.Sprintf("case %s <- %s:", ch, val)
		}
		g.writeLine(commStr)
		g.indent++
		g.generateBlock(c.Body)
		g.indent--
	}
	if stmt.Otherwise != nil {
		g.writeLine("default:")
		g.indent++
		g.generateBlock(stmt.Otherwise.Body)
		g.indent--
	}
	g.indent--
	g.writeLine("}")
}

func (g *Generator) generateTypeSwitchStmt(stmt *ast.TypeSwitchStmt) {
	expr := g.exprToString(stmt.Expression)
	binding := stmt.Binding.Value
	g.writeLine(fmt.Sprintf("switch %s := %s.(type) {", binding, expr))

	g.indent++
	for _, c := range stmt.Cases {
		typeStr := g.generateTypeAnnotation(c.Type)
		g.writeLine(fmt.Sprintf("case %s:", typeStr))

		g.indent++
		g.generateBlock(c.Body)
		g.indent--
	}

	if stmt.Otherwise != nil {
		g.writeLine("default:")
		g.indent++
		g.generateBlock(stmt.Otherwise.Body)
		g.indent--
	}
	g.indent--
	g.writeLine("}")
}

func (g *Generator) generateForRangeStmt(stmt *ast.ForRangeStmt) {
	collection := g.exprToString(stmt.Collection)

	if stmt.Index != nil {
		if stmt.Variable.Value == "_" {
			g.writeLine(fmt.Sprintf("for %s := range %s {", stmt.Index.Value, collection))
		} else {
			g.writeLine(fmt.Sprintf("for %s, %s := range %s {", stmt.Index.Value, stmt.Variable.Value, collection))
		}
	} else {
		// In stdlib/iter, all range loops are over iter.Seq which yields one value
		if g.isStdlibIter {
			g.writeLine(fmt.Sprintf("for %s := range %s {", stmt.Variable.Value, collection))
		} else {
			g.writeLine(fmt.Sprintf("for _, %s := range %s {", stmt.Variable.Value, collection))
		}
	}

	g.indent++
	g.generateBlock(stmt.Body)
	g.indent--

	g.writeLine("}")
}

func (g *Generator) generateForNumericStmt(stmt *ast.ForNumericStmt) {
	varName := stmt.Variable.Value
	start := g.exprToString(stmt.Start)
	end := g.exprToString(stmt.End)

	// for i from A to/through B  →  supports both ascending and descending
	// Generates a single loop with a step variable:
	//   _start, _end, _step := A, B, 1
	//   if _start > _end { _step = -1 }
	//   for varName := _start; varName != _end; varName += _step { ... }       // "to"
	//   for varName := _start; varName != _end+_step; varName += _step { ... } // "through"
	// Optimization: for i from 0 to N stays as range-over-int (Go 1.22+)
	if !stmt.Through && start == "0" {
		if varName == "_" {
			g.writeLine(fmt.Sprintf("for range %s {", end))
		} else {
			g.writeLine(fmt.Sprintf("for %s := range %s {", varName, end))
		}
		g.indent++
		g.generateBlock(stmt.Body)
		g.indent--
		g.writeLine("}")
	} else {
		// Use unique internal variable names to avoid collisions with varName
		startVar := "_" + varName + "Start"
		endVar := "_" + varName + "End"
		stepVar := "_" + varName + "Step"
		if varName == "_" {
			startVar = g.uniqueId("_start")
			endVar = g.uniqueId("_end")
			stepVar = g.uniqueId("_step")
		}

		loopVar := varName
		if varName == "_" {
			loopVar = g.uniqueId("_i")
		}

		// Emit a single loop with a runtime step variable to avoid duplicating
		// the loop body for ascending vs descending directions.
		g.writeLine("{")
		g.indent++
		g.writeLine(fmt.Sprintf("%s, %s, %s := %s, %s, 1", startVar, endVar, stepVar, start, end))
		g.writeLine(fmt.Sprintf("if %s > %s {", startVar, endVar))
		g.indent++
		g.writeLine(fmt.Sprintf("%s = -1", stepVar))
		g.indent--
		g.writeLine("}")

		if !stmt.Through {
			// "to" (exclusive): loop while i != end
			g.writeLine(fmt.Sprintf("for %s := %s; %s != %s; %s += %s {", loopVar, startVar, loopVar, endVar, loopVar, stepVar))
		} else {
			// "through" (inclusive): loop while i != end + step
			g.writeLine(fmt.Sprintf("for %s := %s; %s != %s+%s; %s += %s {", loopVar, startVar, loopVar, endVar, stepVar, loopVar, stepVar))
		}

		g.indent++
		g.generateBlock(stmt.Body)
		g.indent--
		g.writeLine("}")

		g.indent--
		g.writeLine("}")
	}
}

func (g *Generator) generateForConditionStmt(stmt *ast.ForConditionStmt) {
	condition := g.exprToString(stmt.Condition)
	if condition == "true" {
		g.writeLine("for {")
	} else {
		g.writeLine(fmt.Sprintf("for %s {", condition))
	}

	g.indent++
	g.generateBlock(stmt.Body)
	g.indent--

	g.writeLine("}")
}
