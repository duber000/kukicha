package ir

// Node is the interface implemented by all IR nodes.
// IR nodes represent Go-level imperative operations after lowering
// high-level Kukicha constructs (pipes, onerr) into sequences of
// assignments, error checks, and control flow.
type Node interface {
	irNode()
}

// Block is an ordered sequence of IR nodes.
type Block struct {
	Nodes []Node
}

func (*Block) irNode() {}

// Add appends a node to the block.
func (b *Block) Add(n Node) {
	b.Nodes = append(b.Nodes, n)
}

// AddAll appends all nodes from another block.
func (b *Block) AddAll(other *Block) {
	if other != nil {
		b.Nodes = append(b.Nodes, other.Nodes...)
	}
}

// Assign represents an assignment: names := expr or names = expr.
type Assign struct {
	Names  []string // Left-hand side variable names
	Expr   string   // Right-hand side expression (pre-rendered)
	Walrus bool     // true for :=, false for =
}

func (*Assign) irNode() {}

// VarDecl represents a variable declaration: var name type [= value].
type VarDecl struct {
	Name  string
	Type  string
	Value string // Empty if no initializer
}

func (*VarDecl) irNode() {}

// IfErrCheck represents: if errVar != nil { body }
type IfErrCheck struct {
	ErrVar string // The error variable to check
	Body   *Block // Statements inside the if block
}

func (*IfErrCheck) irNode() {}

// Goto represents a goto statement: goto Label.
type Goto struct {
	Label string
}

func (*Goto) irNode() {}

// Label represents a label: LabelName:
type Label struct {
	Name string
}

func (*Label) irNode() {}

// ScopedBlock represents a bare { ... } block for variable scoping.
type ScopedBlock struct {
	Body *Block
}

func (*ScopedBlock) irNode() {}

// RawStmt is a pre-rendered Go statement (escape hatch).
// Used for constructs the lowerer doesn't handle yet.
type RawStmt struct {
	Code string
}

func (*RawStmt) irNode() {}
