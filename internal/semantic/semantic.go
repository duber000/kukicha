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
	// Consumed by codegen for: error-only pipe step detection (isErrorOnlyReturn),
	// piped switch return type inference, empty keyword resolution, expression
	// return type inference, and typed zero-value generation (zeroValueForTypeInfo).
	exprTypes           map[ast.Expression]*TypeInfo
	sourceFile          string                 // Source file path, used to detect stdlib context
	inOnerr             bool                   // True while analyzing an onerr handler
	currentOnerrrAlias  string                 // Named alias for caught error in current onerr block (e.g., "e" for "onerr as e")
	inPipedSwitch       bool                   // True while analyzing piped switch case bodies (suppresses return-count checks)
	deprecatedFuncs     map[string]string      // Function name → deprecation message (from # kuki:deprecated directives)
	deprecatedTypes     map[string]string      // Type name → deprecation message
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
// Call after Analyze() to pass these to codegen via SetExprTypes.
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
	a.deprecatedFuncs = make(map[string]string)
	a.deprecatedTypes = make(map[string]string)

	// Check package name for collisions with Go stdlib
	a.checkPackageName()

	// Validate skill declaration if present
	a.checkSkillDecl()

	// Pre-pass: collect deprecation directives from declarations
	a.collectDeprecations()

	// First pass: Collect all type and interface declarations
	a.collectDeclarations()

	// Second pass: Analyze function bodies and validate
	a.analyzeDeclarations()

	return a.errors
}

// collectDeprecations scans all declarations for # kuki:deprecated directives
// and populates the deprecatedFuncs/deprecatedTypes maps.
func (a *Analyzer) collectDeprecations() {
	for _, decl := range a.program.Declarations {
		switch d := decl.(type) {
		case *ast.FunctionDecl:
			if msg := deprecatedMessage(d.Directives); msg != "" {
				a.deprecatedFuncs[d.Name.Value] = msg
			}
		case *ast.TypeDecl:
			if msg := deprecatedMessage(d.Directives); msg != "" {
				a.deprecatedTypes[d.Name.Value] = msg
			}
		case *ast.InterfaceDecl:
			if msg := deprecatedMessage(d.Directives); msg != "" {
				a.deprecatedTypes[d.Name.Value] = msg
			}
		}
	}
}

// deprecatedMessage returns the message from a # kuki:deprecated directive, or "".
func deprecatedMessage(dirs []ast.Directive) string {
	for _, d := range dirs {
		if d.Name == "deprecated" {
			if len(d.Args) > 0 {
				return d.Args[0]
			}
			return "deprecated"
		}
	}
	return ""
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
