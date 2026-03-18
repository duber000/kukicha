package codegen

import (
	"fmt"
	"strings"
	"github.com/duber000/kukicha/internal/semantic"

	"github.com/duber000/kukicha/internal/ast"
	"github.com/duber000/kukicha/internal/version"
)

// TypeParameter represents a type parameter for stdlib special transpilation
// This is internal to codegen and separate from the removed ast.TypeParameter
type TypeParameter struct {
	Name        string // Generated name: T, U, V, etc.
	Placeholder string // Original placeholder: "any", "any2", etc.
	Constraint  string // "any", "comparable", "cmp.Ordered"
}

// Generator generates Go code from an AST.
//
// ARCHITECTURE NOTE: Kukicha uses placeholders like "any" and "any2" in stdlib
// function signatures to represent generic type parameters. When generating Go code,
// we detect these placeholders and emit proper Go generics (e.g., [T any, K comparable]).
// This allows Kukicha users to write simple code while getting type-safe Go generics.
//
// The isStdlibIter and placeholderMap fields work together:
//   - isStdlibIter is true when generating stdlib/iterator or stdlib/slice code
//   - placeholderMap maps Kukicha placeholders ("any", "any2") to Go type params ("T", "K")
//   - During type annotation generation, we substitute placeholders with type params
//
// This design keeps Kukicha's "beginner-friendly" goal: users don't write generic syntax,
// but the generated Go code is fully type-safe with proper generic constraints.
// FuncDefaults stores information about a function's default parameter values
type FuncDefaults struct {
	ParamNames    []string         // Parameter names in order
	DefaultValues []ast.Expression // Default values (nil if no default)
	HasVariadic   bool             // Whether the last parameter is variadic
}

// defaultStdlibModuleBase is the module path prefix used to rewrite "stdlib/X"
// imports to their full Go module paths. Override with Generator.SetStdlibModule
// when the kukicha module is forked or vendored under a different path.
const defaultStdlibModuleBase = "github.com/duber000/kukicha"

type Generator struct {
	program              *ast.Program
	output               strings.Builder
	indent               int
	placeholderMap       map[string]string        // Maps placeholder names to type param names (e.g., "any" -> "T", "any2" -> "K")
	autoImports          map[string]bool          // Tracks auto-imports needed (e.g., "cmp" for generic constraints)
	pkgAliases           map[string]string        // Maps original package name -> alias when collision detected (e.g., "json" -> "kukijson")
	funcDefaults         map[string]*FuncDefaults // Maps function names to their default parameter info
	isStdlibIter         bool                     // True if generating stdlib/iterator code (enables iter-specific generic transpilation)
	sourceFile           string                   // Source file path for detecting stdlib
	currentFuncName      string                   // Current function being generated (for context-aware decisions)
	currentReturnTypes   []ast.TypeAnnotation     // Return types of current function (for type coercion in returns)
	processingReturnType bool                     // Whether we are currently generating return types
	tempCounter          int                      // Counter for generating unique temporary variable names
	exprReturnCounts     map[ast.Expression]int      // Semantic return counts passed from analyzer (drives onerr multi-value split)
	// exprTypes holds per-expression type info from semantic analysis.
	// Used by isErrorOnlyReturn, inferExprReturnType, inferExprType,
	// pipedSwitchReturnType, empty keyword resolution, and zeroValueForType.
	exprTypes            map[ast.Expression]*semantic.TypeInfo
	mcpTarget            bool                        // True if targeting MCP (Model Context Protocol)
	currentOnErrVar      string                   // Render-time context: set/restored only by renderHandler in lower.go
	currentOnErrAlias    string                   // Render-time context: set/restored only by renderHandler in lower.go
	currentReturnIndex   int                      // Index of return value being generated (-1 if not in return)
	stdlibModuleBase     string                   // Base module path for rewriting "stdlib/X" imports (default: defaultStdlibModuleBase)
	reservedNames        map[string]bool          // User-declared identifiers — uniqueId skips these to avoid collisions
	releaseMode          bool                     // When true, contract checks (requires/ensures/invariant) are stripped
}

// New creates a new code generator
func New(program *ast.Program) *Generator {
	return &Generator{
		program:            program,
		indent:             0,
		autoImports:        make(map[string]bool),
		pkgAliases:         make(map[string]string),
		funcDefaults:       make(map[string]*FuncDefaults),
		stdlibModuleBase:   defaultStdlibModuleBase,
		currentReturnIndex: -1,
	}
}

// childGenerator creates a temporary Generator that shares the parent's semantic
// state (program, auto-imports, aliases, type info) but writes to a fresh output
// buffer at an adjusted indent level. This replaces manual field-by-field copies
// when generating inline code blocks (function literals, arrow lambda bodies).
//
// The child shares the parent's autoImports map by reference, so auto-imports
// discovered during child generation are visible to the parent.
func (g *Generator) childGenerator(extraIndent int) *Generator {
	return &Generator{
		program:            g.program,
		indent:             g.indent + extraIndent,
		placeholderMap:     g.placeholderMap,
		autoImports:        g.autoImports,
		pkgAliases:         g.pkgAliases,
		funcDefaults:       g.funcDefaults,
		isStdlibIter:       g.isStdlibIter,
		sourceFile:         g.sourceFile,
		exprTypes:          g.exprTypes,
		exprReturnCounts:   g.exprReturnCounts,
		currentReturnIndex: -1,
		stdlibModuleBase:   g.stdlibModuleBase,
		reservedNames:      g.reservedNames,
	}
}

// SetStdlibModule overrides the base module path used when rewriting "stdlib/X"
// imports to full Go module paths. The default is "github.com/duber000/kukicha".
// Set this when building a fork or vendoring the kukicha stdlib under a different module name.
func (g *Generator) SetStdlibModule(base string) {
	g.stdlibModuleBase = base
}

// SetSourceFile sets the source file path and detects if special transpilation is needed
func (g *Generator) SetSourceFile(path string) {
	g.sourceFile = path
	// Enable special transpilation for stdlib/iterator files
	g.isStdlibIter = strings.Contains(path, "stdlib/iterator/") || strings.Contains(path, "stdlib\\iterator\\")
	// Note: stdlib/slice uses a different approach - type parameters are detected per-function
}

// SetExprReturnCounts passes semantic analysis return counts to the generator.
func (g *Generator) SetExprReturnCounts(counts map[ast.Expression]int) {
	g.exprReturnCounts = counts
}

// SetExprTypes passes semantic analysis expression types to the generator.
// Used by isErrorOnlyReturn, inferExprReturnType, empty keyword resolution,
// piped switch return type inference, and zeroValueForType.
func (g *Generator) SetExprTypes(types map[ast.Expression]*semantic.TypeInfo) {
	g.exprTypes = types
}

// SetMCPTarget enables special codegen for MCP servers (e.g., print to stderr)
func (g *Generator) SetMCPTarget(v bool) {
	g.mcpTarget = v
}

// SetReleaseMode enables release mode, stripping all contract checks
// (requires/ensures/invariant) from generated code.
func (g *Generator) SetReleaseMode(v bool) {
	g.releaseMode = v
}

// Generate generates Go code from the AST
func (g *Generator) Generate() (string, error) {
	g.output.Reset()

	// Generate header comment
	g.writeLine(fmt.Sprintf("// Generated by Kukicha v%s (requires Go 1.26+)", version.Version))
	g.writeLine("")

	// Generate package declaration
	g.generatePackage()

	// Generate skill metadata comment if present
	g.generateSkillComment()

	// Collect user-declared identifiers so uniqueId can avoid collisions
	g.collectReservedNames()

	// Pre-scan for auto-imports (e.g. net/http for fetch wrappers)
	g.scanForAutoImports()

	// Pre-scan for function defaults (needed for named arguments and default parameter values)
	g.scanForFunctionDefaults()

	// Generate imports (including auto-imports like fmt for string interpolation, print builtin, and onerr explain)
	needsFmt := g.needsStringInterpolation() || g.needsPrintBuiltin() || g.needsExplain()
	needsErrors := g.needsErrorsPackage()
	if len(g.program.Imports) > 0 || needsFmt || needsErrors || len(g.autoImports) > 0 {
		g.writeLine("")
		g.generateImports()
	}

	// Generate declarations
	for _, decl := range g.program.Declarations {
		g.writeLine("")
		g.generateDeclaration(decl)
	}

	return g.output.String(), nil
}

func (g *Generator) generatePackage() {
	packageName := "main"
	if g.program.PetioleDecl != nil {
		packageName = g.program.PetioleDecl.Name.Value
	}
	g.writeLine(fmt.Sprintf("package %s", packageName))
}

func (g *Generator) generateSkillComment() {
	skill := g.program.SkillDecl
	if skill == nil {
		return
	}
	g.writeLine("")
	g.writeLine(fmt.Sprintf("// Skill: %s", skill.Name.Value))
	if skill.Description != "" {
		g.writeLine(fmt.Sprintf("// Description: %s", skill.Description))
	}
	if skill.Version != "" {
		g.writeLine(fmt.Sprintf("// Version: %s", skill.Version))
	}
}

func (g *Generator) generateDeclaration(decl ast.Declaration) {
	g.emitLineDirective(decl.Pos())
	switch d := decl.(type) {
	case *ast.TypeDecl:
		g.generateTypeDecl(d)
	case *ast.InterfaceDecl:
		g.generateInterfaceDecl(d)
	case *ast.FunctionDecl:
		g.generateFunctionDecl(d)
	case *ast.VarDeclStmt:
		g.generateGlobalVarDecl(d)
	case *ast.ConstDecl:
		g.generateConstDecl(d)
	}
}

func (g *Generator) write(s string) {
	g.output.WriteString(s)
}

func (g *Generator) writeLine(s string) {
	if s != "" {
		g.output.WriteString(g.indentStr() + s)
	}
	g.output.WriteString("\n")
}

func (g *Generator) indentStr() string {
	return strings.Repeat("\t", g.indent)
}

// emitLineDirective writes a //line directive that maps the generated Go code
// back to the original .kuki source file. The Go compiler and runtime honor
// these directives, so compile errors, panics, and stack traces will reference
// the .kuki file instead of the generated .go file.
func (g *Generator) emitLineDirective(pos ast.Position) {
	if pos.Line > 0 && pos.File != "" {
		g.output.WriteString(fmt.Sprintf("//line %s:%d\n", pos.File, pos.Line))
	}
}

// uniqueId generates unique identifiers to prevent variable shadowing.
// It skips names that collide with user-declared variables in reservedNames.
func (g *Generator) uniqueId(prefix string) string {
	for {
		g.tempCounter++
		name := fmt.Sprintf("%s_%d", prefix, g.tempCounter)
		if !g.reservedNames[name] {
			return name
		}
	}
}
