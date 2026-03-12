package codegen

import (
	"fmt"
	"strings"

	"github.com/duber000/kukicha/internal/ast"
	"github.com/duber000/kukicha/internal/ir"
)

// Lowerer transforms high-level AST constructs (pipes, onerr) into IR nodes.
// It delegates sub-expression rendering to Generator.exprToString and uses
// Generator.inferReturnCount / isErrorOnlyReturn for semantic queries.
type Lowerer struct {
	gen *Generator
}

// newLowerer creates a Lowerer that shares the generator's semantic info.
func newLowerer(gen *Generator) *Lowerer {
	return &Lowerer{gen: gen}
}

func (l *Lowerer) uniqueId(prefix string) string {
	// Delegate to the generator's counter so variable names are identical
	// to the old direct-emission code path.
	return l.gen.uniqueId(prefix)
}

// ---------- Phase 1: pipe chain lowering (no onerr) ----------

// lowerPipeChain lowers a pipe chain into a sequence of temp variable
// assignments. Returns the IR block and the name of the final temp variable.
func (l *Lowerer) lowerPipeChain(pipe *ast.PipeExpr) (*ir.Block, string) {
	base, steps, ok := flattenPipeChain(pipe)
	if !ok || base == nil || len(steps) == 0 {
		return nil, ""
	}

	block := &ir.Block{}
	current := l.uniqueId("pipe")
	baseExpr := l.gen.exprToString(base)

	if count, ok := l.gen.inferReturnCount(base); ok && count >= 2 {
		// Multi-return base: wrap in IIFE to take first value only
		blanks := make([]string, count-1)
		for i := range blanks {
			blanks[i] = "_"
		}
		baseExpr = fmt.Sprintf("func() any { val, %s := %s; return val }()", strings.Join(blanks, ", "), baseExpr)
	}

	block.Add(&ir.Assign{Names: []string{current}, Expr: baseExpr, Walrus: true})

	for _, step := range steps {
		callExpr, ok := l.gen.generatePipedStepCall(step, current)
		if !ok {
			return nil, ""
		}

		next := l.uniqueId("pipe")
		if count, ok := l.gen.inferReturnCount(step); ok && count >= 2 {
			blanks := make([]string, count-1)
			for i := range blanks {
				blanks[i] = "_"
			}
			callExpr = fmt.Sprintf("func() any { val, %s := %s; return val }()", strings.Join(blanks, ", "), callExpr)
		}

		block.Add(&ir.Assign{Names: []string{next}, Expr: callExpr, Walrus: true})
		current = next
	}

	return block, current
}

// ---------- Phase 2: onerr on simple (non-pipe) expressions ----------

// lowerOnErr produces IR for a single expression + onerr clause.
// names are the LHS variable names (empty for statement-level).
// returnCount is the number of values returned by the expression.
func (l *Lowerer) lowerOnErr(expr string, returnCount int, clause *ast.OnErrClause, names []string, walrus bool) *ir.Block {
	block := &ir.Block{}
	errVar := l.uniqueId("err")

	if walrus {
		lhs := append(append([]string{}, names...), errVar)
		block.Add(&ir.Assign{Names: lhs, Expr: expr, Walrus: true})
	} else {
		block.Add(&ir.VarDecl{Name: errVar, Type: "error"})
		lhs := append(append([]string{}, names...), errVar)
		block.Add(&ir.Assign{Names: lhs, Expr: expr, Walrus: false})
	}

	handlerBlock := l.lowerOnErrHandler(clause, names, errVar)
	block.Add(&ir.IfErrCheck{ErrVar: errVar, Body: handlerBlock})

	return block
}

// lowerOnErrHandler produces the IR body for an if-err-check block.
func (l *Lowerer) lowerOnErrHandler(clause *ast.OnErrClause, names []string, errVar string) *ir.Block {
	body := &ir.Block{}

	if clause.ShorthandReturn {
		// onerr return (bare) — propagate error with zero values
		body.Add(&ir.RawStmt{Code: l.gen.buildShorthandReturn(errVar)})
		return body
	}

	// explain wrapping
	if clause.Explain != "" {
		l.gen.addImport("fmt")
		body.Add(&ir.RawStmt{Code: fmt.Sprintf(`%s = fmt.Errorf("%s: %%w", %s)`, errVar, clause.Explain, errVar)})
		if clause.Handler == nil {
			// Standalone explain: emit return
			body.Add(&ir.RawStmt{Code: l.gen.buildShorthandReturn(errVar)})
			return body
		}
	}

	if clause.Handler == nil {
		return body
	}

	// Render the handler using the existing codegen method via RawStmt capture.
	body.Add(&ir.RawStmt{Code: l.renderHandler(clause, names, errVar)})
	return body
}

// renderHandler captures the output of generateOnErrHandler into a string.
// It is the single point that sets currentOnErrVar and currentOnErrAlias,
// ensuring exprToString resolves "error" / alias identifiers to errVar
// during handler rendering.
func (l *Lowerer) renderHandler(clause *ast.OnErrClause, names []string, errVar string) string {
	// Save and restore generator state.
	savedOutput := l.gen.output
	l.gen.output = strings.Builder{}
	savedIndent := l.gen.indent
	l.gen.indent = 0

	prevOnErrVar := l.gen.currentOnErrVar
	l.gen.currentOnErrVar = errVar
	prevAlias := l.gen.currentOnErrAlias
	l.gen.currentOnErrAlias = clause.Alias

	idents := make([]*ast.Identifier, len(names))
	for i, n := range names {
		idents[i] = &ast.Identifier{Value: n}
	}
	l.gen.generateOnErrHandler(idents, clause.Handler, errVar)

	l.gen.currentOnErrVar = prevOnErrVar
	l.gen.currentOnErrAlias = prevAlias

	result := strings.TrimRight(l.gen.output.String(), "\n")
	l.gen.output = savedOutput
	l.gen.indent = savedIndent
	return result
}

// ---------- Phase 3: onerr pipe chains ----------

// lowerOnErrPipeChain lowers a pipe chain with onerr into IR.
// Each step gets its own error check with the handler inlined.
// Returns the IR block and the final pipe variable name.
func (l *Lowerer) lowerOnErrPipeChain(pipe *ast.PipeExpr, clause *ast.OnErrClause, names []string) (*ir.Block, string) {
	base, steps, ok := flattenPipeChain(pipe)
	if !ok || base == nil || len(steps) == 0 {
		return nil, ""
	}

	block := &ir.Block{}
	current := l.uniqueId("pipe")
	baseExpr := l.gen.exprToString(base)

	if count, ok := l.gen.inferReturnCount(base); ok && count >= 2 {
		errVar := l.uniqueId("err")
		block.Add(&ir.Assign{Names: []string{current, errVar}, Expr: baseExpr, Walrus: true})
		handlerBlock := l.lowerOnErrHandler(clause, names, errVar)
		block.Add(&ir.IfErrCheck{ErrVar: errVar, Body: handlerBlock})
	} else {
		block.Add(&ir.Assign{Names: []string{current}, Expr: baseExpr, Walrus: true})
	}

	for _, step := range steps {
		callExpr, ok := l.gen.generatePipedStepCall(step, current)
		if !ok {
			return nil, ""
		}

		next := l.uniqueId("pipe")
		if count, ok := l.gen.inferReturnCount(step); ok && count >= 2 {
			errVar := l.uniqueId("err")
			block.Add(&ir.Assign{Names: []string{next, errVar}, Expr: callExpr, Walrus: true})
			handlerBlock := l.lowerOnErrHandler(clause, names, errVar)
			block.Add(&ir.IfErrCheck{ErrVar: errVar, Body: handlerBlock})
		} else if l.gen.isErrorOnlyReturn(step) {
			errVar := l.uniqueId("err")
			block.Add(&ir.Assign{Names: []string{errVar}, Expr: callExpr, Walrus: true})
			handlerBlock := l.lowerOnErrHandler(clause, names, errVar)
			block.Add(&ir.IfErrCheck{ErrVar: errVar, Body: handlerBlock})
			continue
		} else {
			block.Add(&ir.Assign{Names: []string{next}, Expr: callExpr, Walrus: true})
		}
		current = next
	}

	return block, current
}

// lowerOnErrPipeChainWithLabels is like lowerOnErrPipeChain but uses goto
// for error handling instead of inline handlers (for piped switch support).
func (l *Lowerer) lowerOnErrPipeChainWithLabels(pipe *ast.PipeExpr, onErrLabel, endLabel string) (*ir.Block, string) {
	base, steps, ok := flattenPipeChain(pipe)
	if !ok {
		return nil, ""
	}

	block := &ir.Block{}
	current := l.uniqueId("pipe")
	baseExpr := l.gen.exprToString(base)

	if count, ok := l.gen.inferReturnCount(base); ok && count >= 2 {
		errVar := l.uniqueId("err")
		block.Add(&ir.Assign{Names: []string{current, errVar}, Expr: baseExpr, Walrus: true})
		block.Add(&ir.IfErrCheck{
			ErrVar: errVar,
			Body:   &ir.Block{Nodes: []ir.Node{&ir.Goto{Label: onErrLabel}}},
		})
	} else {
		block.Add(&ir.Assign{Names: []string{current}, Expr: baseExpr, Walrus: true})
	}

	for _, step := range steps {
		callExpr, ok := l.gen.generatePipedStepCall(step, current)
		if !ok {
			return nil, ""
		}

		next := l.uniqueId("pipe")
		if count, ok := l.gen.inferReturnCount(step); ok && count >= 2 {
			errVar := l.uniqueId("err")
			block.Add(&ir.Assign{Names: []string{next, errVar}, Expr: callExpr, Walrus: true})
			block.Add(&ir.IfErrCheck{
				ErrVar: errVar,
				Body:   &ir.Block{Nodes: []ir.Node{&ir.Goto{Label: onErrLabel}}},
			})
		} else if l.gen.isErrorOnlyReturn(step) {
			errVar := l.uniqueId("err")
			block.Add(&ir.Assign{Names: []string{errVar}, Expr: callExpr, Walrus: true})
			block.Add(&ir.IfErrCheck{
				ErrVar: errVar,
				Body:   &ir.Block{Nodes: []ir.Node{&ir.Goto{Label: onErrLabel}}},
			})
			continue
		} else {
			block.Add(&ir.Assign{Names: []string{next}, Expr: callExpr, Walrus: true})
		}
		current = next
	}

	return block, current
}

// ---------- Phase 4: piped switch with onerr ----------

// lowerPipedSwitchVarDecl produces IR for: result := A |> B() |> switch ... onerr handler
func (l *Lowerer) lowerPipedSwitchVarDecl(varName string, ps *ast.PipedSwitchExpr, clause *ast.OnErrClause, names []*ast.Identifier) *ir.Block {
	block := &ir.Block{}

	returnType := l.gen.pipedSwitchReturnType(ps)
	if returnType == "" {
		returnType = "any"
	}

	onErrLabel := l.uniqueId("onerr")
	endLabel := l.uniqueId("end")

	// Declare result variable, pre-initialized to handler default if available
	handlerDefault := l.gen.extractDefaultValue(clause, returnType)
	block.Add(&ir.VarDecl{Name: varName, Type: returnType, Value: handlerDefault})

	// Build scoped block for pipe chain + switch
	inner := &ir.Block{}

	var finalPipeVar string
	if pipe, ok := ps.Left.(*ast.PipeExpr); ok {
		pipeBlock, pipeVar := l.lowerOnErrPipeChainWithLabels(pipe, onErrLabel, endLabel)
		if pipeBlock == nil {
			return nil
		}
		inner.AddAll(pipeBlock)
		finalPipeVar = pipeVar
	} else {
		finalPipeVar = l.gen.exprToString(ps.Left)
	}

	// Render switch IIFE — temporarily bump indent to match the ScopedBlock
	// context during emission (generatePipedSwitchExpr bakes absolute
	// indentation into the IIFE string).
	savedIndent := l.gen.indent
	l.gen.indent++
	savedLeft := ps.Left
	ps.Left = &ast.Identifier{Value: finalPipeVar}
	switchStr := l.gen.generatePipedSwitchExpr(ps)
	ps.Left = savedLeft
	l.gen.indent = savedIndent

	inner.Add(&ir.Assign{Names: []string{varName}, Expr: switchStr, Walrus: false})
	inner.Add(&ir.Goto{Label: endLabel})

	block.Add(&ir.ScopedBlock{Body: inner})

	// Error handler label — no specific error variable is in scope at the
	// goto target, so we pass empty errVar.
	block.Add(&ir.Label{Name: onErrLabel})
	if handlerDefault == "" {
		handlerBlock := l.lowerOnErrHandler(clause, identNames(names), "")
		block.AddAll(handlerBlock)
	}

	block.Add(&ir.Label{Name: endLabel})
	return block
}

// lowerPipedSwitchStmt produces IR for statement-level: A |> B() |> switch ... onerr handler
func (l *Lowerer) lowerPipedSwitchStmt(ps *ast.PipedSwitchExpr, clause *ast.OnErrClause) *ir.Block {
	block := &ir.Block{}

	onErrLabel := l.uniqueId("onerr")
	endLabel := l.uniqueId("end")

	inner := &ir.Block{}

	var finalPipeVar string
	if pipe, ok := ps.Left.(*ast.PipeExpr); ok {
		pipeBlock, pipeVar := l.lowerOnErrPipeChainWithLabels(pipe, onErrLabel, endLabel)
		if pipeBlock == nil {
			return nil
		}
		inner.AddAll(pipeBlock)
		finalPipeVar = pipeVar
	} else {
		finalPipeVar = l.gen.exprToString(ps.Left)
	}

	// Render switch statement at indent=0, RawStmt multi-line handling
	// adds the emitter's indent to each line.
	switchStr := l.renderSwitchStmt(ps, finalPipeVar)
	inner.Add(&ir.RawStmt{Code: switchStr})
	inner.Add(&ir.Goto{Label: endLabel})

	block.Add(&ir.ScopedBlock{Body: inner})

	block.Add(&ir.Label{Name: onErrLabel})
	handlerBlock := l.lowerOnErrHandler(clause, []string{finalPipeVar}, "")
	block.AddAll(handlerBlock)

	block.Add(&ir.Label{Name: endLabel})
	return block
}

// renderSwitchStmt captures the output of a switch statement into a string.
func (l *Lowerer) renderSwitchStmt(ps *ast.PipedSwitchExpr, finalPipeVar string) string {
	savedOutput := l.gen.output
	l.gen.output = strings.Builder{}
	savedIndent := l.gen.indent
	l.gen.indent = 0

	switch stmt := ps.Switch.(type) {
	case *ast.SwitchStmt:
		originalExpr := stmt.Expression
		stmt.Expression = &ast.Identifier{Value: finalPipeVar}
		l.gen.generateSwitchStmt(stmt)
		stmt.Expression = originalExpr
	case *ast.TypeSwitchStmt:
		originalExpr := stmt.Expression
		stmt.Expression = &ast.Identifier{Value: finalPipeVar}
		l.gen.generateTypeSwitchStmt(stmt)
		stmt.Expression = originalExpr
	}

	result := strings.TrimRight(l.gen.output.String(), "\n")
	l.gen.output = savedOutput
	l.gen.indent = savedIndent
	return result
}

// lowerOnErrStmt produces IR for a statement-level onerr (no named targets).
// It generates blank _ assignments for non-error return values.
func (l *Lowerer) lowerOnErrStmt(exprStr string, expr ast.Expression, clause *ast.OnErrClause) *ir.Block {
	block := &ir.Block{}
	errVar := l.uniqueId("err")

	var blanks []string
	if count, ok := l.gen.inferReturnCount(expr); ok && count > 1 {
		for i := 0; i < count-1; i++ {
			blanks = append(blanks, "_")
		}
	}

	lhs := append(blanks, errVar)
	block.Add(&ir.Assign{Names: lhs, Expr: exprStr, Walrus: true})

	handlerBlock := l.lowerOnErrHandler(clause, []string{}, errVar)
	block.Add(&ir.IfErrCheck{ErrVar: errVar, Body: handlerBlock})

	return block
}

// lowerOnErrWithExplicitErr produces IR for onerr where the user provides
// the error variable as the last name (multi-return case).
func (l *Lowerer) lowerOnErrWithExplicitErr(nameStrs []string, expr string, clause *ast.OnErrClause, walrus bool) *ir.Block {
	block := &ir.Block{}
	block.Add(&ir.Assign{Names: nameStrs, Expr: expr, Walrus: walrus})
	errVar := nameStrs[len(nameStrs)-1]
	handlerNames := nameStrs[:len(nameStrs)-1]
	handlerBlock := l.lowerOnErrHandler(clause, handlerNames, errVar)
	block.Add(&ir.IfErrCheck{ErrVar: errVar, Body: handlerBlock})
	return block
}

// buildShorthandReturn renders a return statement with zero values for all
// non-error return types, plus the given error variable. Extracted from
// generateStandaloneExplainReturn for use in IR lowering.
func (g *Generator) buildShorthandReturn(errVar string) string {
	if len(g.currentReturnTypes) == 0 {
		return fmt.Sprintf("return %s", errVar)
	}

	var parts []string
	for i, ret := range g.currentReturnTypes {
		if i == len(g.currentReturnTypes)-1 {
			parts = append(parts, errVar)
		} else {
			parts = append(parts, g.zeroValueForType(ret))
		}
	}
	return fmt.Sprintf("return %s", strings.Join(parts, ", "))
}
