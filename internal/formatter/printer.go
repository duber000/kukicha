package formatter

import (
	"fmt"
	"strings"

	"github.com/duber000/kukicha/internal/ast"
)

// Printer prints an AST as formatted Kukicha source code
type Printer struct {
	output      strings.Builder
	indentLevel int
	indentStr   string // 4 spaces
}

// NewPrinter creates a new printer
func NewPrinter() *Printer {
	return &Printer{
		indentStr: "    ", // 4 spaces
	}
}

// Print prints the program and returns the formatted source code
func (p *Printer) Print(program *ast.Program) string {
	p.output.Reset()
	p.indentLevel = 0

	// Print petiole declaration if present
	if program.PetioleDecl != nil {
		p.writeLine(fmt.Sprintf("petiole %s", program.PetioleDecl.Name.Value))
		p.writeLine("")
	}

	// Print imports
	for _, imp := range program.Imports {
		p.printImport(imp)
	}

	if len(program.Imports) > 0 {
		p.writeLine("")
	}

	// Print declarations with blank lines between them
	for i, decl := range program.Declarations {
		if i > 0 {
			p.writeLine("")
		}
		p.printDeclaration(decl)
	}

	return p.output.String()
}

func (p *Printer) printImport(imp *ast.ImportDecl) {
	if imp.Alias != nil {
		p.writeLine(fmt.Sprintf("import \"%s\" as %s", imp.Path.Value, imp.Alias.Value))
	} else {
		p.writeLine(fmt.Sprintf("import \"%s\"", imp.Path.Value))
	}
}

func (p *Printer) printDeclaration(decl ast.Declaration) {
	switch d := decl.(type) {
	case *ast.TypeDecl:
		p.printTypeDecl(d)
	case *ast.InterfaceDecl:
		p.printInterfaceDecl(d)
	case *ast.FunctionDecl:
		p.printFunctionDecl(d)
	}
}

func (p *Printer) printTypeDecl(decl *ast.TypeDecl) {
	p.writeLine(fmt.Sprintf("type %s", decl.Name.Value))
	p.indentLevel++

	for _, field := range decl.Fields {
		fieldType := p.typeAnnotationToString(field.Type)
		line := fmt.Sprintf("%s %s", field.Name.Value, fieldType)
		if field.Tag != "" {
			line += fmt.Sprintf(" %s", field.Tag)
		}
		p.writeLine(line)
	}

	p.indentLevel--
}

func (p *Printer) printInterfaceDecl(decl *ast.InterfaceDecl) {
	p.writeLine(fmt.Sprintf("interface %s", decl.Name.Value))
	p.indentLevel++

	for _, method := range decl.Methods {
		params := p.parametersToString(method.Parameters)
		returns := p.returnTypesToString(method.Returns)

		if returns != "" {
			p.writeLine(fmt.Sprintf("func %s(%s) %s", method.Name.Value, params, returns))
		} else {
			p.writeLine(fmt.Sprintf("func %s(%s)", method.Name.Value, params))
		}
	}

	p.indentLevel--
}

func (p *Printer) printFunctionDecl(decl *ast.FunctionDecl) {
	signature := "func "

	// Add receiver for methods
	if decl.Receiver != nil {
		receiverType := p.typeAnnotationToString(decl.Receiver.Type)
		p.writeLine(fmt.Sprintf("func %s on %s %s", decl.Name.Value, decl.Receiver.Name.Value, receiverType))
	} else {
		// Regular function
		signature += decl.Name.Value
	}

	if decl.Receiver == nil {
		// Add parameters
		params := p.parametersToString(decl.Parameters)
		signature += fmt.Sprintf("(%s)", params)

		// Add return types
		returns := p.returnTypesToString(decl.Returns)
		if returns != "" {
			signature += " " + returns
		}

		p.writeLine(signature)
	} else {
		// Method - parameters on same line after receiver
		params := p.parametersToString(decl.Parameters)
		returns := p.returnTypesToString(decl.Returns)
		line := fmt.Sprintf("func %s on %s %s(%s)", decl.Name.Value, decl.Receiver.Name.Value, p.typeAnnotationToString(decl.Receiver.Type), params)
		if returns != "" {
			line += " " + returns
		}
		p.writeLine(line)
	}

	// Print body
	if decl.Body != nil {
		p.indentLevel++
		p.printBlock(decl.Body)
		p.indentLevel--
	}
}

func (p *Printer) parametersToString(params []*ast.Parameter) string {
	if len(params) == 0 {
		return ""
	}

	parts := make([]string, len(params))
	for i, param := range params {
		paramType := p.typeAnnotationToString(param.Type)
		parts[i] = fmt.Sprintf("%s %s", param.Name.Value, paramType)
	}

	return strings.Join(parts, ", ")
}

func (p *Printer) returnTypesToString(returns []ast.TypeAnnotation) string {
	if len(returns) == 0 {
		return ""
	}

	if len(returns) == 1 {
		return p.typeAnnotationToString(returns[0])
	}

	parts := make([]string, len(returns))
	for i, ret := range returns {
		parts[i] = p.typeAnnotationToString(ret)
	}

	return "(" + strings.Join(parts, ", ") + ")"
}

func (p *Printer) typeAnnotationToString(typeAnn ast.TypeAnnotation) string {
	if typeAnn == nil {
		return ""
	}

	switch t := typeAnn.(type) {
	case *ast.PrimitiveType:
		return t.Name
	case *ast.NamedType:
		return t.Name
	case *ast.ReferenceType:
		return "reference " + p.typeAnnotationToString(t.ElementType)
	case *ast.ListType:
		return "list of " + p.typeAnnotationToString(t.ElementType)
	case *ast.MapType:
		keyType := p.typeAnnotationToString(t.KeyType)
		valueType := p.typeAnnotationToString(t.ValueType)
		return fmt.Sprintf("map of %s to %s", keyType, valueType)
	case *ast.ChannelType:
		return "channel of " + p.typeAnnotationToString(t.ElementType)
	default:
		return "any"
	}
}

func (p *Printer) printBlock(block *ast.BlockStmt) {
	for _, stmt := range block.Statements {
		p.printStatement(stmt)
	}
}

func (p *Printer) printStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.VarDeclStmt:
		p.printVarDeclStmt(s)
	case *ast.AssignStmt:
		p.printAssignStmt(s)
	case *ast.ReturnStmt:
		p.printReturnStmt(s)
	case *ast.IfStmt:
		p.printIfStmt(s)
	case *ast.SwitchStmt:
		p.printSwitchStmt(s)
	case *ast.ForRangeStmt:
		p.printForRangeStmt(s)
	case *ast.ForNumericStmt:
		p.printForNumericStmt(s)
	case *ast.ForConditionStmt:
		p.printForConditionStmt(s)
	case *ast.DeferStmt:
		p.writeLine("defer " + p.exprToString(s.Call))
	case *ast.GoStmt:
		if s.Block != nil {
			p.writeLine("go")
			p.indentLevel++
			for _, stmt := range s.Block.Statements {
				p.printStatement(stmt)
			}
			p.indentLevel--
		} else {
			p.writeLine("go " + p.exprToString(s.Call))
		}
	case *ast.SendStmt:
		channel := p.exprToString(s.Channel)
		value := p.exprToString(s.Value)
		p.writeLine(fmt.Sprintf("send %s to %s", value, channel))
	case *ast.ExpressionStmt:
		p.writeLine(p.exprToString(s.Expression) + p.onErrSuffix(s.OnErr))
	}
}

func (p *Printer) onErrSuffix(clause *ast.OnErrClause) string {
	if clause == nil {
		return ""
	}
	return " onerr " + p.exprToString(clause.Handler)
}

func (p *Printer) printVarDeclStmt(stmt *ast.VarDeclStmt) {
	names := make([]string, len(stmt.Names))
	for i, n := range stmt.Names {
		names[i] = n.Value
	}
	values := make([]string, len(stmt.Values))
	for i, v := range stmt.Values {
		values[i] = p.exprToString(v)
	}
	p.writeLine(fmt.Sprintf("%s := %s%s", strings.Join(names, ", "), strings.Join(values, ", "), p.onErrSuffix(stmt.OnErr)))
}

func (p *Printer) printAssignStmt(stmt *ast.AssignStmt) {
	targets := make([]string, len(stmt.Targets))
	for i, t := range stmt.Targets {
		targets[i] = p.exprToString(t)
	}
	values := make([]string, len(stmt.Values))
	for i, v := range stmt.Values {
		values[i] = p.exprToString(v)
	}
	p.writeLine(fmt.Sprintf("%s = %s%s", strings.Join(targets, ", "), strings.Join(values, ", "), p.onErrSuffix(stmt.OnErr)))
}

func (p *Printer) printReturnStmt(stmt *ast.ReturnStmt) {
	if len(stmt.Values) == 0 {
		p.writeLine("return")
		return
	}

	values := make([]string, len(stmt.Values))
	for i, val := range stmt.Values {
		values[i] = p.exprToString(val)
	}

	p.writeLine(fmt.Sprintf("return %s", strings.Join(values, ", ")))
}

func (p *Printer) printIfStmt(stmt *ast.IfStmt) {
	condition := p.exprToString(stmt.Condition)
	p.writeLine(fmt.Sprintf("if %s", condition))

	p.indentLevel++
	p.printBlock(stmt.Consequence)
	p.indentLevel--

	if stmt.Alternative != nil {
		switch alt := stmt.Alternative.(type) {
		case *ast.ElseStmt:
			p.writeLine("else")
			p.indentLevel++
			p.printBlock(alt.Body)
			p.indentLevel--
		case *ast.IfStmt:
			// else if - print on same conceptual level
			p.write(p.indent())
			p.output.WriteString("else ")
			// Reset to print the if without indent prefix
			condition := p.exprToString(alt.Condition)
			p.output.WriteString(fmt.Sprintf("if %s\n", condition))
			p.indentLevel++
			p.printBlock(alt.Consequence)
			p.indentLevel--
			if alt.Alternative != nil {
				p.printIfStmtAlternative(alt.Alternative)
			}
		}
	}
}

func (p *Printer) printIfStmtAlternative(alt ast.Statement) {
	switch a := alt.(type) {
	case *ast.ElseStmt:
		p.writeLine("else")
		p.indentLevel++
		p.printBlock(a.Body)
		p.indentLevel--
	case *ast.IfStmt:
		p.write(p.indent())
		p.output.WriteString("else ")
		condition := p.exprToString(a.Condition)
		p.output.WriteString(fmt.Sprintf("if %s\n", condition))
		p.indentLevel++
		p.printBlock(a.Consequence)
		p.indentLevel--
		if a.Alternative != nil {
			p.printIfStmtAlternative(a.Alternative)
		}
	}
}

func (p *Printer) printForRangeStmt(stmt *ast.ForRangeStmt) {
	collection := p.exprToString(stmt.Collection)

	if stmt.Index != nil {
		p.writeLine(fmt.Sprintf("for %s, %s in %s", stmt.Index.Value, stmt.Variable.Value, collection))
	} else {
		p.writeLine(fmt.Sprintf("for %s in %s", stmt.Variable.Value, collection))
	}

	p.indentLevel++
	p.printBlock(stmt.Body)
	p.indentLevel--
}

func (p *Printer) printForNumericStmt(stmt *ast.ForNumericStmt) {
	varName := stmt.Variable.Value
	start := p.exprToString(stmt.Start)
	end := p.exprToString(stmt.End)

	keyword := "to"
	if stmt.Through {
		keyword = "through"
	}

	p.writeLine(fmt.Sprintf("for %s from %s %s %s", varName, start, keyword, end))

	p.indentLevel++
	p.printBlock(stmt.Body)
	p.indentLevel--
}

func (p *Printer) printForConditionStmt(stmt *ast.ForConditionStmt) {
	condition := p.exprToString(stmt.Condition)
	p.writeLine(fmt.Sprintf("for %s", condition))

	p.indentLevel++
	p.printBlock(stmt.Body)
	p.indentLevel--
}

func (p *Printer) printSwitchStmt(stmt *ast.SwitchStmt) {
	if stmt.Expression != nil {
		p.writeLine(fmt.Sprintf("switch %s", p.exprToString(stmt.Expression)))
	} else {
		p.writeLine("switch")
	}

	p.indentLevel++
	for _, c := range stmt.Cases {
		values := make([]string, len(c.Values))
		for i, v := range c.Values {
			values[i] = p.exprToString(v)
		}
		p.writeLine(fmt.Sprintf("when %s", strings.Join(values, ", ")))
		p.indentLevel++
		p.printBlock(c.Body)
		p.indentLevel--
	}

	if stmt.Otherwise != nil {
		p.writeLine("otherwise")
		p.indentLevel++
		p.printBlock(stmt.Otherwise.Body)
		p.indentLevel--
	}
	p.indentLevel--
}

func (p *Printer) exprToString(expr ast.Expression) string {
	if expr == nil {
		return ""
	}

	switch e := expr.(type) {
	case *ast.Identifier:
		return e.Value
	case *ast.IntegerLiteral:
		return fmt.Sprintf("%d", e.Value)
	case *ast.FloatLiteral:
		return fmt.Sprintf("%g", e.Value)
	case *ast.StringLiteral:
		return p.stringLiteralToString(e)
	case *ast.BooleanLiteral:
		if e.Value {
			return "true"
		}
		return "false"
	case *ast.BinaryExpr:
		return p.binaryExprToString(e)
	case *ast.UnaryExpr:
		return p.unaryExprToString(e)
	case *ast.PipeExpr:
		left := p.exprToString(e.Left)
		right := p.exprToString(e.Right)
		return fmt.Sprintf("%s |> %s", left, right)
	// Note: OnErrExpr removed — onerr is now a clause on VarDeclStmt, AssignStmt, ExpressionStmt
	case *ast.CallExpr:
		return p.callExprToString(e)
	case *ast.MethodCallExpr:
		return p.methodCallExprToString(e)
	case *ast.IndexExpr:
		left := p.exprToString(e.Left)
		index := p.exprToString(e.Index)
		return fmt.Sprintf("%s[%s]", left, index)
	case *ast.SliceExpr:
		return p.sliceExprToString(e)
	case *ast.StructLiteralExpr:
		return p.structLiteralToString(e)
	case *ast.ListLiteralExpr:
		return p.listLiteralToString(e)
	case *ast.MapLiteralExpr:
		return p.mapLiteralToString(e)
	case *ast.ReceiveExpr:
		channel := p.exprToString(e.Channel)
		return fmt.Sprintf("receive %s", channel)
	case *ast.TypeCastExpr:
		targetType := p.typeAnnotationToString(e.TargetType)
		expr := p.exprToString(e.Expression)
		return fmt.Sprintf("%s(%s)", targetType, expr)
	case *ast.EmptyExpr:
		if e.Type != nil {
			targetType := p.typeAnnotationToString(e.Type)
			return fmt.Sprintf("empty %s", targetType)
		}
		return "empty"
	case *ast.DiscardExpr:
		return "discard"
	case *ast.ErrorExpr:
		message := p.exprToString(e.Message)
		return fmt.Sprintf("error %s", message)
	case *ast.MakeExpr:
		return p.makeExprToString(e)
	case *ast.CloseExpr:
		channel := p.exprToString(e.Channel)
		return fmt.Sprintf("close %s", channel)
	case *ast.PanicExpr:
		message := p.exprToString(e.Message)
		return fmt.Sprintf("panic %s", message)
	case *ast.RecoverExpr:
		return "recover"
	case *ast.ArrowLambda:
		return p.arrowLambdaToString(e)
	default:
		return ""
	}
}

func (p *Printer) stringLiteralToString(lit *ast.StringLiteral) string {
	// Preserve interpolation syntax
	return fmt.Sprintf("\"%s\"", lit.Value)
}

func (p *Printer) binaryExprToString(expr *ast.BinaryExpr) string {
	left := p.exprToString(expr.Left)
	right := p.exprToString(expr.Right)

	// Convert Go operators to Kukicha
	op := expr.Operator
	switch op {
	case "&&":
		op = "and"
	case "||":
		op = "or"
	case "==":
		op = "equals"
	case "!=":
		op = "not equals"
	}

	return fmt.Sprintf("(%s %s %s)", left, op, right)
}

func (p *Printer) unaryExprToString(expr *ast.UnaryExpr) string {
	right := p.exprToString(expr.Right)

	op := expr.Operator
	if op == "!" {
		op = "not"
	}

	if op == "not" {
		return fmt.Sprintf("not %s", right)
	}
	return fmt.Sprintf("%s%s", op, right)
}

func (p *Printer) callExprToString(expr *ast.CallExpr) string {
	funcName := p.exprToString(expr.Function)
	args := make([]string, len(expr.Arguments))
	for i, arg := range expr.Arguments {
		args[i] = p.exprToString(arg)
	}

	return fmt.Sprintf("%s(%s)", funcName, strings.Join(args, ", "))
}

func (p *Printer) methodCallExprToString(expr *ast.MethodCallExpr) string {
	object := p.exprToString(expr.Object)
	method := expr.Method.Value

	if len(expr.Arguments) == 0 {
		return fmt.Sprintf("%s.%s", object, method)
	}

	args := make([]string, len(expr.Arguments))
	for i, arg := range expr.Arguments {
		args[i] = p.exprToString(arg)
	}

	return fmt.Sprintf("%s.%s(%s)", object, method, strings.Join(args, ", "))
}

func (p *Printer) sliceExprToString(expr *ast.SliceExpr) string {
	left := p.exprToString(expr.Left)

	var start, end string
	if expr.Start != nil {
		start = p.exprToString(expr.Start)
	}
	if expr.End != nil {
		end = p.exprToString(expr.End)
	}

	return fmt.Sprintf("%s[%s:%s]", left, start, end)
}

func (p *Printer) structLiteralToString(expr *ast.StructLiteralExpr) string {
	typeName := p.typeAnnotationToString(expr.Type)

	if len(expr.Fields) == 0 {
		return fmt.Sprintf("%s{}", typeName)
	}

	fields := make([]string, len(expr.Fields))
	for i, field := range expr.Fields {
		value := p.exprToString(field.Value)
		fields[i] = fmt.Sprintf("%s: %s", field.Name.Value, value)
	}

	return fmt.Sprintf("%s{%s}", typeName, strings.Join(fields, ", "))
}

func (p *Printer) listLiteralToString(expr *ast.ListLiteralExpr) string {
	if len(expr.Elements) == 0 {
		if expr.Type != nil {
			elemType := p.typeAnnotationToString(expr.Type)
			return fmt.Sprintf("empty list of %s", elemType)
		}
		return "empty list"
	}

	elements := make([]string, len(expr.Elements))
	for i, elem := range expr.Elements {
		elements[i] = p.exprToString(elem)
	}

	return fmt.Sprintf("[%s]", strings.Join(elements, ", "))
}

func (p *Printer) mapLiteralToString(expr *ast.MapLiteralExpr) string {
	keyType := p.typeAnnotationToString(expr.KeyType)
	valType := p.typeAnnotationToString(expr.ValType)

	if len(expr.Pairs) == 0 {
		return fmt.Sprintf("empty map of %s to %s", keyType, valType)
	}

	pairs := make([]string, len(expr.Pairs))
	for i, pair := range expr.Pairs {
		key := p.exprToString(pair.Key)
		value := p.exprToString(pair.Value)
		pairs[i] = fmt.Sprintf("%s: %s", key, value)
	}

	return fmt.Sprintf("{%s}", strings.Join(pairs, ", "))
}

func (p *Printer) makeExprToString(expr *ast.MakeExpr) string {
	targetType := p.typeAnnotationToString(expr.Type)

	if len(expr.Args) == 0 {
		return fmt.Sprintf("make %s", targetType)
	}

	args := make([]string, len(expr.Args))
	for i, arg := range expr.Args {
		args[i] = p.exprToString(arg)
	}

	return fmt.Sprintf("make %s, %s", targetType, strings.Join(args, ", "))
}

func (p *Printer) arrowLambdaToString(lambda *ast.ArrowLambda) string {
	// Build parameter string
	var paramParts []string
	for _, param := range lambda.Parameters {
		if param.Type != nil {
			paramParts = append(paramParts, param.Name.Value+" "+p.typeAnnotationToString(param.Type))
		} else {
			paramParts = append(paramParts, param.Name.Value)
		}
	}

	var paramsStr string
	if len(lambda.Parameters) == 1 && lambda.Parameters[0].Type == nil {
		// Single untyped param: no parens
		paramsStr = lambda.Parameters[0].Name.Value
	} else {
		paramsStr = "(" + strings.Join(paramParts, ", ") + ")"
	}

	if lambda.Body != nil {
		return fmt.Sprintf("%s => %s", paramsStr, p.exprToString(lambda.Body))
	}

	// Block form — for now, render inline (full block rendering would need multi-line support)
	return fmt.Sprintf("%s => ...", paramsStr)
}

// Helper methods

func (p *Printer) indent() string {
	return strings.Repeat(p.indentStr, p.indentLevel)
}

func (p *Printer) write(s string) {
	p.output.WriteString(s)
}

func (p *Printer) writeLine(s string) {
	p.output.WriteString(p.indent())
	p.output.WriteString(s)
	p.output.WriteString("\n")
}
