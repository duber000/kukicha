package codegen

import (
	"strings"
	"testing"

	"github.com/duber000/kukicha/internal/ast"
	"github.com/duber000/kukicha/internal/ir"
)

func newTestGenerator() *Generator {
	return New(&ast.Program{})
}

func emitToString(t *testing.T, block *ir.Block) string {
	t.Helper()
	gen := newTestGenerator()
	gen.emitIR(block)
	return gen.output.String()
}

func TestEmitAssignWalrus(t *testing.T) {
	block := &ir.Block{}
	block.Add(&ir.Assign{Names: []string{"x"}, Expr: "42", Walrus: true})

	out := emitToString(t, block)
	if !strings.Contains(out, "x := 42") {
		t.Errorf("expected 'x := 42', got: %s", out)
	}
}

func TestEmitAssignPlain(t *testing.T) {
	block := &ir.Block{}
	block.Add(&ir.Assign{Names: []string{"x"}, Expr: "42", Walrus: false})

	out := emitToString(t, block)
	if !strings.Contains(out, "x = 42") {
		t.Errorf("expected 'x = 42', got: %s", out)
	}
}

func TestEmitMultiNameAssign(t *testing.T) {
	block := &ir.Block{}
	block.Add(&ir.Assign{Names: []string{"val", "err"}, Expr: "foo()", Walrus: true})

	out := emitToString(t, block)
	if !strings.Contains(out, "val, err := foo()") {
		t.Errorf("expected 'val, err := foo()', got: %s", out)
	}
}

func TestEmitVarDecl(t *testing.T) {
	block := &ir.Block{}
	block.Add(&ir.VarDecl{Name: "x", Type: "int"})

	out := emitToString(t, block)
	if !strings.Contains(out, "var x int") {
		t.Errorf("expected 'var x int', got: %s", out)
	}
}

func TestEmitVarDeclWithValue(t *testing.T) {
	block := &ir.Block{}
	block.Add(&ir.VarDecl{Name: "x", Type: "string", Value: `"hello"`})

	out := emitToString(t, block)
	if !strings.Contains(out, `var x string = "hello"`) {
		t.Errorf("expected var with init, got: %s", out)
	}
}

func TestEmitIfErrCheck(t *testing.T) {
	body := &ir.Block{}
	body.Add(&ir.RawStmt{Code: `panic("fail")`})

	block := &ir.Block{}
	block.Add(&ir.IfErrCheck{ErrVar: "err_1", Body: body})

	out := emitToString(t, block)
	if !strings.Contains(out, "if err_1 != nil {") {
		t.Errorf("expected if-err check, got: %s", out)
	}
	if !strings.Contains(out, `panic("fail")`) {
		t.Errorf("expected panic in body, got: %s", out)
	}
	if !strings.Contains(out, "}") {
		t.Errorf("expected closing brace, got: %s", out)
	}
}

func TestEmitGotoLabel(t *testing.T) {
	block := &ir.Block{}
	block.Add(&ir.Goto{Label: "onerr_1"})
	block.Add(&ir.Label{Name: "onerr_1"})

	out := emitToString(t, block)
	if !strings.Contains(out, "goto onerr_1") {
		t.Errorf("expected goto, got: %s", out)
	}
	if !strings.Contains(out, "onerr_1:") {
		t.Errorf("expected label, got: %s", out)
	}
}

func TestEmitScopedBlock(t *testing.T) {
	inner := &ir.Block{}
	inner.Add(&ir.RawStmt{Code: "x := 1"})

	block := &ir.Block{}
	block.Add(&ir.ScopedBlock{Body: inner})

	out := emitToString(t, block)
	if !strings.Contains(out, "{") {
		t.Errorf("expected opening brace, got: %s", out)
	}
	if !strings.Contains(out, "x := 1") {
		t.Errorf("expected inner stmt, got: %s", out)
	}
}

func TestEmitRawStmt(t *testing.T) {
	block := &ir.Block{}
	block.Add(&ir.RawStmt{Code: "fmt.Println(x)"})

	out := emitToString(t, block)
	if !strings.Contains(out, "fmt.Println(x)") {
		t.Errorf("expected raw stmt, got: %s", out)
	}
}

func TestEmitNilBlock(t *testing.T) {
	gen := newTestGenerator()
	gen.emitIR(nil)
	if gen.output.String() != "" {
		t.Errorf("expected empty output for nil block, got: %s", gen.output.String())
	}
}

func TestEmitIndentation(t *testing.T) {
	body := &ir.Block{}
	body.Add(&ir.RawStmt{Code: "return err_1"})

	block := &ir.Block{}
	block.Add(&ir.IfErrCheck{ErrVar: "err_1", Body: body})

	gen := newTestGenerator()
	gen.indent = 1 // Start at indent level 1
	gen.emitIR(block)
	out := gen.output.String()

	// The if-err check should be at indent 1, body at indent 2
	if !strings.Contains(out, "\tif err_1 != nil {") {
		t.Errorf("expected indented if, got: %s", out)
	}
	if !strings.Contains(out, "\t\treturn err_1") {
		t.Errorf("expected double-indented body, got: %s", out)
	}
}

func TestEmitCompositeIR(t *testing.T) {
	// Simulate a lowered pipe chain with onerr:
	// pipe_1 := getData()
	// pipe_2, err_1 := parse(pipe_1)
	// if err_1 != nil { panic("fail") }
	block := &ir.Block{}
	block.Add(&ir.Assign{Names: []string{"pipe_1"}, Expr: "getData()", Walrus: true})
	block.Add(&ir.Assign{Names: []string{"pipe_2", "err_1"}, Expr: "parse(pipe_1)", Walrus: true})

	errBody := &ir.Block{}
	errBody.Add(&ir.RawStmt{Code: `panic("fail")`})
	block.Add(&ir.IfErrCheck{ErrVar: "err_1", Body: errBody})

	out := emitToString(t, block)
	if !strings.Contains(out, "pipe_1 := getData()") {
		t.Errorf("missing pipe_1 assign, got: %s", out)
	}
	if !strings.Contains(out, "pipe_2, err_1 := parse(pipe_1)") {
		t.Errorf("missing pipe_2 assign, got: %s", out)
	}
	if !strings.Contains(out, "if err_1 != nil {") {
		t.Errorf("missing err check, got: %s", out)
	}
}
