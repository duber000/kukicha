package codegen

import (
	"fmt"
	"strings"

	"github.com/duber000/kukicha/internal/ir"
)

// emitIR walks an IR block and emits Go source via the generator's writeLine.
func (g *Generator) emitIR(block *ir.Block) {
	if block == nil {
		return
	}
	for _, node := range block.Nodes {
		g.emitIRNode(node)
	}
}

func (g *Generator) emitIRNode(node ir.Node) {
	switch n := node.(type) {
	case *ir.Assign:
		lhs := joinStrings(n.Names, ", ")
		op := "="
		if n.Walrus {
			op = ":="
		}
		g.writeLine(fmt.Sprintf("%s %s %s", lhs, op, n.Expr))

	case *ir.VarDecl:
		if n.Value != "" {
			g.writeLine(fmt.Sprintf("var %s %s = %s", n.Name, n.Type, n.Value))
		} else {
			g.writeLine(fmt.Sprintf("var %s %s", n.Name, n.Type))
		}

	case *ir.IfErrCheck:
		g.writeLine(fmt.Sprintf("if %s != nil {", n.ErrVar))
		g.indent++
		g.emitIR(n.Body)
		g.indent--
		g.writeLine("}")

	case *ir.Goto:
		g.writeLine(fmt.Sprintf("goto %s", n.Label))

	case *ir.Label:
		// Labels are not indented (Go convention)
		g.writeLine(fmt.Sprintf("%s:", n.Name))

	case *ir.ScopedBlock:
		g.writeLine("{")
		g.indent++
		g.emitIR(n.Body)
		g.indent--
		g.writeLine("}")

	case *ir.RawStmt:
		if strings.Contains(n.Code, "\n") {
			for _, line := range strings.Split(n.Code, "\n") {
				g.writeLine(line)
			}
		} else {
			g.writeLine(n.Code)
		}

	case *ir.Block:
		g.emitIR(n)
	}
}

func joinStrings(ss []string, sep string) string {
	return strings.Join(ss, sep)
}
