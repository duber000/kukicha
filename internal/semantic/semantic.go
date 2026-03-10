package semantic

import (
	"fmt"

	"github.com/duber000/kukicha/internal/ast"
)

// Analyzer performs semantic analysis on the AST
type Analyzer struct {
	program          *ast.Program
	symbolTable      *SymbolTable
	errors           []error
	warnings         []error                // Non-fatal diagnostics (e.g. risky onerr handlers)
	currentFunc      *ast.FunctionDecl      // Track current function for return type checking
	loopDepth        int                    // Track loop nesting for break/continue
	switchDepth      int                    // Track switch nesting for break
	exprReturnCounts    map[ast.Expression]int // Inferred return counts for expressions (used by codegen for onerr multi-value split)
	// exprTypes maps each analyzed expression to its inferred TypeInfo.
	// Currently populated during analysis but not yet consumed by codegen.
	// Planned uses: contextual type inference for untyped arrow lambda parameters,
	// smarter pipe chain error handling, and typed zero-value generation.
	exprTypes           map[ast.Expression]*TypeInfo
	sourceFile          string                 // Source file path, used to detect stdlib context
	inOnerr             bool                   // True while analyzing an onerr handler
	currentOnerrrAlias  string                 // Named alias for caught error in current onerr block (e.g., "e" for "onerr as e")
}

// New creates a new semantic analyzer
func New(program *ast.Program) *Analyzer {
	return &Analyzer{
		program:     program,
		symbolTable: NewSymbolTable(),
		errors:      []error{},
	}
}

// NewWithFile creates a new semantic analyzer with the source file path.
// The file path is used to allow Kukicha stdlib packages to use Go stdlib names.
func NewWithFile(program *ast.Program, sourceFile string) *Analyzer {
	return &Analyzer{
		program:     program,
		symbolTable: NewSymbolTable(),
		errors:      []error{},
		sourceFile:  sourceFile,
	}
}

// ExprTypes returns the inferred types for expressions.
// Call after Analyze() to pass these to codegen.
// Currently populated but not yet consumed by codegen — infrastructure
// for future contextual type inference (e.g., typed lambda parameters,
// smarter zero-value generation in pipe chains).
func (a *Analyzer) ExprTypes() map[ast.Expression]*TypeInfo {
	return a.exprTypes
}

// ReturnCounts returns the inferred return counts for expressions.
// Call after Analyze() to pass these to codegen.
func (a *Analyzer) ReturnCounts() map[ast.Expression]int {
	return a.exprReturnCounts
}

// Analyze performs semantic analysis on the program
func (a *Analyzer) Analyze() []error {
	a.exprReturnCounts = make(map[ast.Expression]int)
	a.exprTypes = make(map[ast.Expression]*TypeInfo)

	// Check package name for collisions with Go stdlib
	a.checkPackageName()

	// Validate skill declaration if present
	a.checkSkillDecl()

	// First pass: Collect all type and interface declarations
	a.collectDeclarations()

	// Second pass: Analyze function bodies and validate
	a.analyzeDeclarations()

	return a.errors
}

func (a *Analyzer) recordReturnCount(expr ast.Expression, count int) {
	if expr == nil || count < 0 {
		return
	}
	a.exprReturnCounts[expr] = count
}

func (a *Analyzer) recordType(expr ast.Expression, info *TypeInfo) {
	if expr == nil || info == nil {
		return
	}
	a.exprTypes[expr] = info
}

func (a *Analyzer) error(pos ast.Position, message string) {
	err := fmt.Errorf("%s:%d:%d: %s", pos.File, pos.Line, pos.Column, message)
	a.errors = append(a.errors, err)
}

func (a *Analyzer) warn(pos ast.Position, message string) {
	w := fmt.Errorf("%s:%d:%d: %s", pos.File, pos.Line, pos.Column, message)
	a.warnings = append(a.warnings, w)
}

// Warnings returns non-fatal diagnostics collected during analysis.
// Call after Analyze(). The caller decides whether to display or promote them to errors.
func (a *Analyzer) Warnings() []error {
	return a.warnings
}
