package codegen

import (
	"strings"
	"testing"

	"github.com/duber000/kukicha/internal/ast"
	"github.com/duber000/kukicha/internal/ir"
	"github.com/duber000/kukicha/internal/semantic"
)

func TestLowererUniqueId(t *testing.T) {
	gen := New(&ast.Program{})
	l := newLowerer(gen)
	id1 := l.uniqueId("pipe")
	id2 := l.uniqueId("pipe")
	if id1 == id2 {
		t.Errorf("expected unique IDs, got %s and %s", id1, id2)
	}
	if id1 != "pipe_1" {
		t.Errorf("expected pipe_1, got %s", id1)
	}
	if id2 != "pipe_2" {
		t.Errorf("expected pipe_2, got %s", id2)
	}
}

func TestUniqueIdSkipsReservedNames(t *testing.T) {
	gen := New(&ast.Program{})
	gen.reservedNames = map[string]bool{
		"pipe_1": true,
		"err_3":  true,
	}

	// pipe_1 is reserved, so should skip to pipe_2
	id1 := gen.uniqueId("pipe")
	if id1 != "pipe_2" {
		t.Errorf("expected pipe_2 (skipping reserved pipe_1), got %s", id1)
	}

	// err_3 is reserved, so should skip to err_4
	id2 := gen.uniqueId("err")
	if id2 != "err_4" {
		t.Errorf("expected err_4 (skipping reserved err_3), got %s", id2)
	}

	// pipe_5 is not reserved, should be returned directly
	id3 := gen.uniqueId("pipe")
	if id3 != "pipe_5" {
		t.Errorf("expected pipe_5, got %s", id3)
	}
}

func TestLowerPipeChainSimple(t *testing.T) {
	// Simulate: a |> b() |> c()
	// Build AST by hand
	a := &ast.Identifier{Value: "a"}
	bCall := &ast.CallExpr{
		Function:  &ast.Identifier{Value: "b"},
		Arguments: []ast.Expression{},
	}
	cCall := &ast.CallExpr{
		Function:  &ast.Identifier{Value: "c"},
		Arguments: []ast.Expression{},
	}

	pipe1 := &ast.PipeExpr{Left: a, Right: bCall}
	pipe2 := &ast.PipeExpr{Left: pipe1, Right: cCall}

	gen := New(&ast.Program{})
	l := newLowerer(gen)

	block, finalVar := l.lowerPipeChain(pipe2)
	if block == nil {
		t.Fatal("expected non-nil block")
	}
	if finalVar == "" {
		t.Fatal("expected non-empty final var")
	}

	// Should have 3 assignments: base, step1, step2
	if len(block.Nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(block.Nodes))
	}

	// Verify the assignments
	assign0 := block.Nodes[0].(*ir.Assign)
	if assign0.Expr != "a" {
		t.Errorf("expected base expr 'a', got '%s'", assign0.Expr)
	}
	if !assign0.Walrus {
		t.Error("expected walrus assignment for base")
	}

	assign1 := block.Nodes[1].(*ir.Assign)
	if !strings.Contains(assign1.Expr, "b(") {
		t.Errorf("expected b() call in step 1, got '%s'", assign1.Expr)
	}

	assign2 := block.Nodes[2].(*ir.Assign)
	if !strings.Contains(assign2.Expr, "c(") {
		t.Errorf("expected c() call in step 2, got '%s'", assign2.Expr)
	}

	// Final var should match the last assignment target
	if assign2.Names[0] != finalVar {
		t.Errorf("final var %s doesn't match last assign target %s", finalVar, assign2.Names[0])
	}
}

func TestLowerPipeChainEmitProducesGoCode(t *testing.T) {
	// Build: a |> b()
	a := &ast.Identifier{Value: "a"}
	bCall := &ast.CallExpr{
		Function:  &ast.Identifier{Value: "b"},
		Arguments: []ast.Expression{},
	}
	pipe := &ast.PipeExpr{Left: a, Right: bCall}

	gen := New(&ast.Program{})
	l := newLowerer(gen)

	block, _ := l.lowerPipeChain(pipe)

	gen.emitIR(block)
	out := gen.output.String()

	if !strings.Contains(out, "pipe_1 := a") {
		t.Errorf("expected pipe_1 := a, got: %s", out)
	}
	if !strings.Contains(out, "pipe_2 := b(pipe_1)") {
		t.Errorf("expected pipe_2 := b(pipe_1), got: %s", out)
	}
}

func TestLowerOnErrHandlerPanic(t *testing.T) {
	gen := New(&ast.Program{})
	l := newLowerer(gen)

	clause := &ast.OnErrClause{
		Handler: &ast.PanicExpr{
			Message: &ast.StringLiteral{Value: "something failed"},
		},
	}

	block := l.lowerOnErr("foo()", clause, []string{"x"}, true)
	if block == nil {
		t.Fatal("expected non-nil block")
	}

	gen.emitIR(block)
	out := gen.output.String()

	if !strings.Contains(out, "x, err_1 := foo()") {
		t.Errorf("expected assignment with err, got: %s", out)
	}
	if !strings.Contains(out, "if err_1 != nil {") {
		t.Errorf("expected err check, got: %s", out)
	}
	if !strings.Contains(out, "panic(") {
		t.Errorf("expected panic handler, got: %s", out)
	}
}

func TestLowerOnErrPipeChainEmitsCorrectly(t *testing.T) {
	// Build: a |> b() with onerr panic
	a := &ast.Identifier{Value: "getData"}
	aCall := &ast.CallExpr{
		Function:  a,
		Arguments: []ast.Expression{},
	}
	bCall := &ast.CallExpr{
		Function:  &ast.Identifier{Value: "process"},
		Arguments: []ast.Expression{},
	}
	pipe := &ast.PipeExpr{Left: aCall, Right: bCall}

	// Set up return counts so getData returns (data, error) = 2 values
	gen := New(&ast.Program{})
	gen.exprReturnCounts = map[ast.Expression]int{
		aCall: 2,
	}

	clause := &ast.OnErrClause{
		Handler: &ast.PanicExpr{
			Message: &ast.StringLiteral{Value: "failed"},
		},
	}

	l := newLowerer(gen)
	block, finalVar := l.lowerOnErrPipeChain(pipe, clause, []string{}, "")

	if block == nil {
		t.Fatal("expected non-nil block")
	}
	if finalVar == "" {
		t.Fatal("expected non-empty final var")
	}

	gen.emitIR(block)
	out := gen.output.String()

	// Should have: pipe_1, err_2 := getData()
	if !strings.Contains(out, ", err_") {
		t.Errorf("expected err variable in base assignment, got: %s", out)
	}
	// Should have error check after base
	if !strings.Contains(out, "!= nil {") {
		t.Errorf("expected error check, got: %s", out)
	}
	// process() is non-error, so it's collapsed into the finalVar expression
	// (not emitted as IR). Verify finalVar contains the process call.
	if !strings.Contains(finalVar, "process(") {
		t.Errorf("expected finalVar to contain process call, got: %s", finalVar)
	}
}

func TestLowerOnErrPipeChainTargetName(t *testing.T) {
	// Build: result := getData() |> process() onerr panic "failed"
	// With targetName="result", the last step should assign directly to "result"
	// instead of a temp variable, eliminating the redundant final copy.
	aCall := &ast.CallExpr{
		Function:  &ast.Identifier{Value: "getData"},
		Arguments: []ast.Expression{},
	}
	bCall := &ast.CallExpr{
		Function:  &ast.Identifier{Value: "process"},
		Arguments: []ast.Expression{},
	}
	pipe := &ast.PipeExpr{Left: aCall, Right: bCall}

	gen := New(&ast.Program{})
	gen.exprReturnCounts = map[ast.Expression]int{
		aCall: 2,
		bCall: 2,
	}

	clause := &ast.OnErrClause{
		Handler: &ast.PanicExpr{
			Message: &ast.StringLiteral{Value: "failed"},
		},
	}

	l := newLowerer(gen)
	block, finalVar := l.lowerOnErrPipeChain(pipe, clause, []string{}, "result")

	if block == nil {
		t.Fatal("expected non-nil block")
	}
	if finalVar != "result" {
		t.Errorf("expected finalVar to be 'result', got '%s'", finalVar)
	}

	gen.emitIR(block)
	out := gen.output.String()

	// Last step should use target name directly
	if !strings.Contains(out, "result, err_") {
		t.Errorf("expected last step to assign to 'result', got: %s", out)
	}
	// Should NOT have a redundant "result := pipe_N" line
	if strings.Contains(out, "result := pipe_") {
		t.Errorf("expected no redundant final copy, got: %s", out)
	}
}

func TestLowerOnErrPipeChainTargetNameErrorOnlyLast(t *testing.T) {
	// Build: data |> marshalPretty() |> os.WriteFile() onerr panic "failed"
	// When the last step is error-only, targetName should apply to the
	// last value-producing step (marshalPretty), not the error-only step.
	dataIdent := &ast.Identifier{Value: "data"}
	marshalCall := &ast.CallExpr{
		Function:  &ast.Identifier{Value: "marshalPretty"},
		Arguments: []ast.Expression{},
	}
	writeCall := &ast.MethodCallExpr{
		Object:    &ast.Identifier{Value: "os"},
		Method:    &ast.Identifier{Value: "WriteFile"},
		Arguments: []ast.Expression{},
	}

	pipe1 := &ast.PipeExpr{Left: dataIdent, Right: marshalCall}
	pipe2 := &ast.PipeExpr{Left: pipe1, Right: writeCall}

	gen := New(&ast.Program{})
	gen.exprReturnCounts = map[ast.Expression]int{
		marshalCall: 2,
		writeCall:   1, // error-only returns need count=1
	}
	// Mark writeCall as error-only via exprTypes
	gen.exprTypes = map[ast.Expression]*semantic.TypeInfo{
		writeCall: {Kind: semantic.TypeKindNamed, Name: "error"},
	}

	clause := &ast.OnErrClause{
		Handler: &ast.PanicExpr{
			Message: &ast.StringLiteral{Value: "failed"},
		},
	}

	l := newLowerer(gen)
	block, finalVar := l.lowerOnErrPipeChain(pipe2, clause, []string{}, "result")

	if block == nil {
		t.Fatal("expected non-nil block")
	}
	// The last value-producing step is marshalPretty, so finalVar should be "result"
	if finalVar != "result" {
		t.Errorf("expected finalVar to be 'result', got '%s'", finalVar)
	}

	gen.emitIR(block)
	out := gen.output.String()

	// marshalPretty (last value-producing step) should assign to "result"
	if !strings.Contains(out, "result, err_") {
		t.Errorf("expected marshalPretty to assign to 'result', got: %s", out)
	}
}

func TestLowerOnErrPipeChainWithLabels(t *testing.T) {
	// Build: a |> b() with goto-based error handling
	a := &ast.Identifier{Value: "getData"}
	aCall := &ast.CallExpr{
		Function:  a,
		Arguments: []ast.Expression{},
	}
	bCall := &ast.CallExpr{
		Function:  &ast.Identifier{Value: "process"},
		Arguments: []ast.Expression{},
	}
	pipe := &ast.PipeExpr{Left: aCall, Right: bCall}

	gen := New(&ast.Program{})
	gen.exprReturnCounts = map[ast.Expression]int{
		aCall: 2,
	}

	l := newLowerer(gen)
	block, finalVar := l.lowerOnErrPipeChainWithLabels(pipe, "onerr_label")

	if block == nil {
		t.Fatal("expected non-nil block")
	}
	if finalVar == "" {
		t.Fatal("expected non-empty final var")
	}

	gen.emitIR(block)
	out := gen.output.String()

	// Should use goto instead of inline handler
	if !strings.Contains(out, "goto onerr_label") {
		t.Errorf("expected goto onerr_label, got: %s", out)
	}
}

func TestLowerOnErrShorthandReturn(t *testing.T) {
	gen := New(&ast.Program{})
	gen.currentReturnTypes = []ast.TypeAnnotation{
		&ast.PrimitiveType{Name: "string"},
		&ast.PrimitiveType{Name: "error"},
	}

	l := newLowerer(gen)

	clause := &ast.OnErrClause{
		ShorthandReturn: true,
	}

	block := l.lowerOnErr("readFile()", clause, []string{"data"}, true)
	gen.emitIR(block)
	out := gen.output.String()

	if !strings.Contains(out, "data, err_1 := readFile()") {
		t.Errorf("expected assignment, got: %s", out)
	}
	if !strings.Contains(out, `return "", err_1`) {
		t.Errorf("expected shorthand return with zero value, got: %s", out)
	}
}

func TestLowerOnErrExplain(t *testing.T) {
	gen := New(&ast.Program{})
	gen.currentReturnTypes = []ast.TypeAnnotation{
		&ast.PrimitiveType{Name: "error"},
	}

	l := newLowerer(gen)

	clause := &ast.OnErrClause{
		Explain: "failed to read config",
	}

	block := l.lowerOnErr("readConfig()", clause, []string{"cfg"}, true)
	gen.emitIR(block)
	out := gen.output.String()

	if !strings.Contains(out, "fmt.Errorf") {
		t.Errorf("expected fmt.Errorf for explain, got: %s", out)
	}
	if !strings.Contains(out, "failed to read config") {
		t.Errorf("expected explain message, got: %s", out)
	}
}

func TestLowerPipedSwitchVarDeclWithDefault(t *testing.T) {
	// Build: result := value |> Risky() |> switch as v { when string: return v; otherwise: return "other" } onerr "fallback"
	value := &ast.Identifier{Value: "value"}
	riskyCall := &ast.CallExpr{
		Function:  &ast.Identifier{Value: "Risky"},
		Arguments: []ast.Expression{},
	}
	pipe := &ast.PipeExpr{Left: value, Right: riskyCall}

	ps := &ast.PipedSwitchExpr{
		Left: pipe,
		Switch: &ast.TypeSwitchStmt{
			Expression: &ast.Identifier{Value: "_"},
			Binding:    &ast.Identifier{Value: "v"},
			Cases: []*ast.TypeCase{
				{
					Type: &ast.PrimitiveType{Name: "string"},
					Body: &ast.BlockStmt{Statements: []ast.Statement{
						&ast.ReturnStmt{Values: []ast.Expression{&ast.Identifier{Value: "v"}}},
					}},
				},
			},
			Otherwise: &ast.OtherwiseCase{Body: &ast.BlockStmt{Statements: []ast.Statement{
				&ast.ReturnStmt{Values: []ast.Expression{&ast.StringLiteral{Value: "other"}}},
			}}},
		},
	}

	clause := &ast.OnErrClause{
		Handler: &ast.StringLiteral{Value: "fallback"},
	}

	gen := New(&ast.Program{})
	gen.exprReturnCounts = map[ast.Expression]int{
		riskyCall: 2,
	}

	l := newLowerer(gen)
	block := l.lowerPipedSwitchVarDecl("result", ps, clause, []*ast.Identifier{{Value: "result"}})

	if block == nil {
		t.Fatal("expected non-nil block")
	}

	gen.emitIR(block)
	out := gen.output.String()

	if !strings.Contains(out, `var result string = "fallback"`) {
		t.Errorf("expected var decl with default, got: %s", out)
	}
	if !strings.Contains(out, "{") {
		t.Errorf("expected scoped block, got: %s", out)
	}
	if !strings.Contains(out, "goto onerr_") {
		t.Errorf("expected goto onerr label, got: %s", out)
	}
	if !strings.Contains(out, "goto end_") {
		t.Errorf("expected goto end label, got: %s", out)
	}
}

func TestLowerPipedSwitchVarDeclWithPanic(t *testing.T) {
	// Build: result := value |> Risky() |> switch ... onerr panic "failed"
	value := &ast.Identifier{Value: "value"}
	riskyCall := &ast.CallExpr{
		Function:  &ast.Identifier{Value: "Risky"},
		Arguments: []ast.Expression{},
	}
	pipe := &ast.PipeExpr{Left: value, Right: riskyCall}

	ps := &ast.PipedSwitchExpr{
		Left: pipe,
		Switch: &ast.SwitchStmt{
			Expression: &ast.Identifier{Value: "_"},
			Cases: []*ast.WhenCase{
				{
					Values: []ast.Expression{&ast.StringLiteral{Value: "a"}},
					Body:   &ast.BlockStmt{Statements: []ast.Statement{&ast.ReturnStmt{Values: []ast.Expression{&ast.IntegerLiteral{Value: 1}}}}},
				},
			},
		},
	}

	clause := &ast.OnErrClause{
		Handler: &ast.PanicExpr{
			Message: &ast.StringLiteral{Value: "failed"},
		},
	}

	gen := New(&ast.Program{})
	gen.exprReturnCounts = map[ast.Expression]int{
		riskyCall: 2,
	}

	l := newLowerer(gen)
	block := l.lowerPipedSwitchVarDecl("result", ps, clause, []*ast.Identifier{{Value: "result"}})

	if block == nil {
		t.Fatal("expected non-nil block")
	}

	gen.emitIR(block)
	out := gen.output.String()

	// No default value, so handler should be emitted
	if !strings.Contains(out, "var result int") {
		t.Errorf("expected var decl without default, got: %s", out)
	}
	if !strings.Contains(out, `panic("failed")`) {
		t.Errorf("expected panic handler, got: %s", out)
	}
}

func TestLowerPipedSwitchStmt(t *testing.T) {
	// Build: value |> Risky() |> switch ... onerr panic "failed"
	value := &ast.Identifier{Value: "value"}
	riskyCall := &ast.CallExpr{
		Function:  &ast.Identifier{Value: "Risky"},
		Arguments: []ast.Expression{},
	}
	pipe := &ast.PipeExpr{Left: value, Right: riskyCall}

	ps := &ast.PipedSwitchExpr{
		Left: pipe,
		Switch: &ast.SwitchStmt{
			Expression: &ast.Identifier{Value: "_"},
			Cases: []*ast.WhenCase{
				{
					Values: []ast.Expression{&ast.StringLiteral{Value: "a"}},
					Body:   &ast.BlockStmt{Statements: []ast.Statement{&ast.ExpressionStmt{Expression: &ast.CallExpr{Function: &ast.Identifier{Value: "doA"}, Arguments: []ast.Expression{}}}}},
				},
			},
		},
	}

	clause := &ast.OnErrClause{
		Handler: &ast.PanicExpr{
			Message: &ast.StringLiteral{Value: "failed"},
		},
	}

	gen := New(&ast.Program{})
	gen.exprReturnCounts = map[ast.Expression]int{
		riskyCall: 2,
	}

	l := newLowerer(gen)
	block := l.lowerPipedSwitchStmt(ps, clause)

	if block == nil {
		t.Fatal("expected non-nil block")
	}

	gen.emitIR(block)
	out := gen.output.String()

	if !strings.Contains(out, "{") {
		t.Errorf("expected scoped block, got: %s", out)
	}
	if !strings.Contains(out, "switch") {
		t.Errorf("expected switch stmt, got: %s", out)
	}
	if !strings.Contains(out, `panic("failed")`) {
		t.Errorf("expected panic handler, got: %s", out)
	}
	if !strings.Contains(out, "goto onerr_") {
		t.Errorf("expected goto onerr, got: %s", out)
	}
}

func TestLowerOnErrAssignNonWalrus(t *testing.T) {
	gen := New(&ast.Program{})
	l := newLowerer(gen)

	clause := &ast.OnErrClause{
		Handler: &ast.PanicExpr{
			Message: &ast.StringLiteral{Value: "fail"},
		},
	}

	block := l.lowerOnErr("fetch()", clause, []string{"result"}, false)
	gen.emitIR(block)
	out := gen.output.String()

	// Non-walrus should emit var decl + plain assignment
	if !strings.Contains(out, "var err_1 error") {
		t.Errorf("expected var decl for err, got: %s", out)
	}
	if !strings.Contains(out, "result, err_1 = fetch()") {
		t.Errorf("expected plain assignment, got: %s", out)
	}
}
