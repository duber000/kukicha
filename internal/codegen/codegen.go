package codegen

import (
	"fmt"
	"maps"
	"strings"

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
	exprReturnCounts     map[ast.Expression]int   // Semantic return counts passed from analyzer
	mcpTarget            bool                     // True if targeting MCP (Model Context Protocol)
	currentOnErrVar      string                   // Current onerr error variable name (for block-style onerr {error} references)
	currentOnErrAlias    string                   // Named alias for caught error in current onerr block (e.g., "e" for "onerr as e")
	stdlibModuleBase     string                   // Base module path for rewriting "stdlib/X" imports (default: defaultStdlibModuleBase)
}

// New creates a new code generator
func New(program *ast.Program) *Generator {
	return &Generator{
		program:          program,
		indent:           0,
		autoImports:      make(map[string]bool),
		pkgAliases:       make(map[string]string),
		funcDefaults:     make(map[string]*FuncDefaults),
		stdlibModuleBase: defaultStdlibModuleBase,
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

// SetMCPTarget enables special codegen for MCP servers (e.g., print to stderr)
func (g *Generator) SetMCPTarget(v bool) {
	g.mcpTarget = v
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
	}
}

func (g *Generator) generateTypeDecl(decl *ast.TypeDecl) {
	// Type alias (e.g., type Handler func(string))
	if decl.AliasType != nil {
		g.writeLine(fmt.Sprintf("type %s %s", decl.Name.Value, g.generateTypeAnnotation(decl.AliasType)))
		return
	}

	g.write(fmt.Sprintf("type %s struct {", decl.Name.Value))
	g.writeLine("")
	g.indent++

	for _, field := range decl.Fields {
		fieldType := g.generateTypeAnnotation(field.Type)
		line := fmt.Sprintf("%s %s", field.Name.Value, fieldType)
		if field.Tag != "" {
			line += fmt.Sprintf(" `%s`", field.Tag)
		}
		g.writeLine(line)
	}

	g.indent--
	g.writeLine("}")
}

func (g *Generator) generateInterfaceDecl(decl *ast.InterfaceDecl) {
	g.write(fmt.Sprintf("type %s interface {", decl.Name.Value))
	g.writeLine("")
	g.indent++

	for _, method := range decl.Methods {
		// Generate method signature
		params := g.generateParameters(method.Parameters)
		returns := g.generateReturnTypes(method.Returns)

		if returns != "" {
			g.writeLine(fmt.Sprintf("%s(%s) %s", method.Name.Value, params, returns))
		} else {
			g.writeLine(fmt.Sprintf("%s(%s)", method.Name.Value, params))
		}
	}

	g.indent--
	g.writeLine("}")
}

func (g *Generator) generateGlobalVarDecl(stmt *ast.VarDeclStmt) {
	if len(stmt.Names) == 0 {
		return
	}

	// Build comma-separated list of names
	names := make([]string, len(stmt.Names))
	for i, n := range stmt.Names {
		names[i] = n.Value
	}
	namesStr := strings.Join(names, ", ")

	// Generate type if present
	if stmt.Type != nil {
		varType := g.generateTypeAnnotation(stmt.Type)
		if len(stmt.Values) > 0 {
			// With initializer
			values := make([]string, len(stmt.Values))
			for i, v := range stmt.Values {
				values[i] = g.exprToString(v)
			}
			valuesStr := strings.Join(values, ", ")
			g.writeLine(fmt.Sprintf("var %s %s = %s", namesStr, varType, valuesStr))
		} else {
			// Without initializer
			g.writeLine(fmt.Sprintf("var %s %s", namesStr, varType))
		}
	} else if len(stmt.Values) > 0 {
		// No explicit type, but with initializer
		values := make([]string, len(stmt.Values))
		for i, v := range stmt.Values {
			values[i] = g.exprToString(v)
		}
		valuesStr := strings.Join(values, ", ")
		g.writeLine(fmt.Sprintf("var %s = %s", namesStr, valuesStr))
	} else {
		// No type, no initializer - this is unusual for a global variable
		// but we'll generate it anyway (will be zero-valued)
		g.writeLine(fmt.Sprintf("var %s any", namesStr))
	}
}

// generateFunctionDecl generates a Go function from a Kukicha function declaration.
//
// ARCHITECTURE NOTE: For stdlib/iterator and stdlib/slice packages, this function
// performs "generic inference" - it scans the function's parameter and return types
// for placeholder types ("any", "any2") and generates proper Go type parameters.
//
// Example: A Kukicha function like:
//
//	func Filter(items list of any, predicate func(any) bool) list of any
//
// Becomes Go code like:
//
//	func Filter[T any](items []T, predicate func(T) bool) []T
//
// This happens automatically for stdlib packages. User code doesn't need this
// because users import the stdlib and call its generic functions; the Go compiler
// handles type inference for callers.
func (g *Generator) generateFunctionDecl(decl *ast.FunctionDecl) {
	// Set up placeholder mapping for this function
	g.placeholderMap = make(map[string]string)

	g.currentFuncName = decl.Name.Value

	// Check if this is a stdlib function that needs special transpilation
	// (generic type parameter inference from placeholder types)
	var typeParams []*TypeParameter
	if g.isStdlibIter {
		// Generate type parameters from function signature for iter
		typeParams = g.inferStdlibTypeParameters(decl)
		for _, tp := range typeParams {
			g.placeholderMap[tp.Placeholder] = tp.Name
		}
	} else if g.isStdlibSlice() {
		// Generate type parameters from function signature for slice
		typeParams = g.inferSliceTypeParameters(decl)
		for _, tp := range typeParams {
			g.placeholderMap[tp.Placeholder] = tp.Name
		}
	} else if g.isStdlibFetch() {
		// Generate type parameters for selected fetch helpers (e.g., Json)
		typeParams = g.inferFetchTypeParameters(decl)
		for _, tp := range typeParams {
			g.placeholderMap[tp.Placeholder] = tp.Name
		}
	} else if g.isStdlibJSON() {
		// Generate type parameters for selected json helpers (e.g., DecodeRead)
		typeParams = g.inferJSONTypeParameters(decl)
		for _, tp := range typeParams {
			g.placeholderMap[tp.Placeholder] = tp.Name
		}
	}

	// Generate function signature
	signature := "func "

	// Add receiver for methods
	if decl.Receiver != nil {
		receiverType := g.generateTypeAnnotation(decl.Receiver.Type)
		receiverName := decl.Receiver.Name.Value
		signature += fmt.Sprintf("(%s %s) ", receiverName, receiverType)
	}

	// Add function name
	signature += decl.Name.Value

	// Add type parameters if present
	if len(typeParams) > 0 {
		signature += g.generateTypeParameters(typeParams)
	}

	// Add parameters
	params := g.generateFunctionParameters(decl.Parameters)
	signature += fmt.Sprintf("(%s)", params)

	// Add return types
	g.processingReturnType = true
	returns := g.generateReturnTypes(decl.Returns)
	g.processingReturnType = false

	if returns != "" {
		signature += " " + returns
	}

	g.write(signature + " {")
	g.writeLine("")

	// Set return types for type coercion in return statements
	g.currentReturnTypes = decl.Returns

	// Generate body
	if decl.Body != nil {
		g.indent++
		g.generateBlock(decl.Body)
		g.indent--
	}

	g.writeLine("}")

	// Clear function context
	g.placeholderMap = nil
	g.currentFuncName = ""
	g.currentReturnTypes = nil
}

func (g *Generator) generateFunctionLiteral(lit *ast.FunctionLiteral) string {
	// Save current placeholder map and create new one for this literal
	oldPlaceholderMap := g.placeholderMap
	g.placeholderMap = make(map[string]string)

	// Inherit placeholders from parent scope
	maps.Copy(g.placeholderMap, oldPlaceholderMap)

	// Check if this is a stdlib/iterator function literal that needs special transpilation
	var typeParams []*TypeParameter
	if g.isStdlibIter {
		// Create a temporary function decl to reuse the inference logic
		tempDecl := &ast.FunctionDecl{
			Name:       &ast.Identifier{Value: ""}, // dummy name for inference
			Parameters: lit.Parameters,
			Returns:    lit.Returns,
		}
		typeParams = g.inferStdlibTypeParameters(tempDecl)
		for _, tp := range typeParams {
			g.placeholderMap[tp.Placeholder] = tp.Name
		}
	}

	// Generate function signature
	signature := "func"

	// Add type parameters if present
	if len(typeParams) > 0 {
		signature += g.generateTypeParameters(typeParams)
	}

	// Add parameters
	params := g.generateFunctionParameters(lit.Parameters)
	signature += fmt.Sprintf("(%s)", params)

	// Add return types
	returns := g.generateReturnTypes(lit.Returns)
	if returns != "" {
		signature += " " + returns
	}

	// Generate body inline - create temporary generator to capture output
	tempGen := &Generator{
		program:        g.program,
		output:         strings.Builder{},
		indent:         g.indent + 1,
		placeholderMap: g.placeholderMap,
		autoImports:    g.autoImports,
		isStdlibIter:   g.isStdlibIter,
		sourceFile:     g.sourceFile,
	}

	var result strings.Builder
	result.WriteString(signature + " {\n")

	if lit.Body != nil {
		// Generate body statements using temporary generator
		for _, stmt := range lit.Body.Statements {
			tempGen.generateStatement(stmt)
		}
		result.WriteString(tempGen.output.String())
	}

	// Add proper indentation for closing brace
	for i := 0; i < g.indent; i++ {
		result.WriteString("\t")
	}
	result.WriteString("}")

	// Restore placeholder mapping
	g.placeholderMap = oldPlaceholderMap

	return result.String()
}

// generateArrowLambda transpiles an arrow lambda to a Go anonymous function.
// Expression form: (r Repo) => r.Stars > 100  →  func(r Repo) bool { return r.Stars > 100 }
// Block form:      (r Repo) => BLOCK           →  func(r Repo) ReturnType { BLOCK }
func (g *Generator) generateArrowLambda(lambda *ast.ArrowLambda) string {
	// Build parameter string
	var paramParts []string
	for _, param := range lambda.Parameters {
		if param.Type != nil {
			paramParts = append(paramParts, param.Name.Value+" "+g.generateTypeAnnotation(param.Type))
		} else {
			// Untyped parameter — type must be inferred from context.
			// For now, we emit as-is; the Go compiler will catch type errors.
			// Full contextual inference is a semantic analysis extension.
			paramParts = append(paramParts, param.Name.Value)
		}
	}
	params := strings.Join(paramParts, ", ")

	if lambda.Body != nil {
		// Expression lambda: auto-return the expression
		bodyStr := g.exprToString(lambda.Body)

		// Infer return type from the expression for the Go func signature.
		// For typed params, we can determine the return type.
		// For the common case, we omit the return type and let Go infer it
		// from the context (e.g., when passed to a generic function).
		returnType := g.inferExprReturnType(lambda.Body)

		if returnType != "" {
			return fmt.Sprintf("func(%s) %s { return %s }", params, returnType, bodyStr)
		}
		return fmt.Sprintf("func(%s) { return %s }", params, bodyStr)
	}

	if lambda.Block != nil {
		// Block lambda: generate as multi-line anonymous function
		returnType := g.inferBlockReturnType(lambda.Block)

		// Create temporary generator to capture body output
		tempGen := &Generator{
			program:        g.program,
			output:         strings.Builder{},
			indent:         g.indent + 1,
			placeholderMap: g.placeholderMap,
			autoImports:    g.autoImports,
			isStdlibIter:   g.isStdlibIter,
			sourceFile:     g.sourceFile,
		}

		for _, stmt := range lambda.Block.Statements {
			tempGen.generateStatement(stmt)
		}

		var result string
		if returnType != "" {
			result = fmt.Sprintf("func(%s) %s {\n", params, returnType)
		} else {
			result = fmt.Sprintf("func(%s) {\n", params)
		}
		result += tempGen.output.String()
		for i := 0; i < g.indent; i++ {
			result += "\t"
		}
		result += "}"
		return result
	}

	// Shouldn't happen — at least one of Body or Block must be set
	return fmt.Sprintf("func(%s) {}", params)
}

// generateTypeParameters generates Go generic type parameter list
func (g *Generator) generateTypeParameters(typeParams []*TypeParameter) string {
	if len(typeParams) == 0 {
		return ""
	}

	parts := make([]string, len(typeParams))
	for i, tp := range typeParams {
		constraint := tp.Constraint
		if constraint == "cmp.Ordered" {
			g.addImport("cmp")
		}
		parts[i] = fmt.Sprintf("%s %s", tp.Name, constraint)
	}

	return "[" + strings.Join(parts, ", ") + "]"
}

func (g *Generator) generateFunctionParameters(params []*ast.Parameter) string {
	if len(params) == 0 {
		return ""
	}

	parts := make([]string, len(params))
	for i, param := range params {
		paramType := g.generateTypeAnnotation(param.Type)
		if param.Variadic {
			// Variadic parameter: use ...Type syntax
			parts[i] = fmt.Sprintf("%s ...%s", param.Name.Value, paramType)
		} else {
			parts[i] = fmt.Sprintf("%s %s", param.Name.Value, paramType)
		}
	}

	return strings.Join(parts, ", ")
}

func (g *Generator) generateParameters(params []*ast.Parameter) string {
	return g.generateFunctionParameters(params)
}

func (g *Generator) generateReturnTypes(returns []ast.TypeAnnotation) string {
	if len(returns) == 0 {
		return ""
	}

	if len(returns) == 1 {
		return g.generateTypeAnnotation(returns[0])
	}

	// Multiple return types
	parts := make([]string, len(returns))
	for i, ret := range returns {
		parts[i] = g.generateTypeAnnotation(ret)
	}

	return "(" + strings.Join(parts, ", ") + ")"
}

func (g *Generator) generateTypeAnnotation(typeAnn ast.TypeAnnotation) string {
	if typeAnn == nil {
		return ""
	}

	switch t := typeAnn.(type) {
	case *ast.PrimitiveType:
		if g.placeholderMap != nil {
			if typeParam, ok := g.placeholderMap[t.Name]; ok {
				return typeParam
			}
		}
		return t.Name
	case *ast.NamedType:
		if g.placeholderMap != nil {
			if typeParam, ok := g.placeholderMap[t.Name]; ok {
				return typeParam
			}
		}
		// Rewrite package-qualified type names if the package was auto-aliased
		if dotIdx := strings.Index(t.Name, "."); dotIdx > 0 {
			pkgPart := t.Name[:dotIdx]
			typePart := t.Name[dotIdx:]
			if alias, ok := g.pkgAliases[pkgPart]; ok {
				return alias + typePart
			}
		}
		// Special handling for iter.Seq in stdlib mode
		if g.isStdlibIter && g.placeholderMap != nil {
			if g.isIterSeqType(t) {
				// Transform iter.Seq → iter.Seq[T]
				if _, ok := g.placeholderMap["any"]; ok {
					typeParam := "T"
					// If this is a return type and U is declared, use U
					if g.processingReturnType {
						if _, hasU := g.placeholderMap["any2"]; hasU {
							typeParam = "U"
						}
					}
					return "iter.Seq[" + typeParam + "]"
				}
			}

			// iter.SeqU → iter.Seq[U]
			if t.Name == "iter.SeqU" {
				return "iter.Seq[U]"
			}

			// iter.Seq2 → iter.Seq2[T, U] or iter.Seq2[int, T] (for Enumerate) or iter.Seq2[T, T]
			if t.Name == "iter.Seq2" {
				if g.currentFuncName == "Enumerate" {
					return "iter.Seq2[int, T]"
				}
				// Only use U if it's actually declared as a type parameter
				if _, hasU := g.placeholderMap["any2"]; hasU {
					return "iter.Seq2[T, U]"
				}
				return "iter.Seq2[T, T]"
			}

			// iter.SeqSlice → iter.Seq[[]T] (for Chunk)
			if t.Name == "iter.SeqSlice" {
				return "iter.Seq[[]T]"
			}
		}
		return t.Name
	case *ast.ReferenceType:
		return "*" + g.generateTypeAnnotation(t.ElementType)
	case *ast.ListType:
		return "[]" + g.generateTypeAnnotation(t.ElementType)
	case *ast.MapType:
		keyType := g.generateTypeAnnotation(t.KeyType)
		valueType := g.generateTypeAnnotation(t.ValueType)
		return fmt.Sprintf("map[%s]%s", keyType, valueType)
		// Note: keyType and valueType already have placeholders substituted via recursion
	case *ast.ChannelType:
		return "chan " + g.generateTypeAnnotation(t.ElementType)
	case *ast.FunctionType:
		// Generate Go function type: func(params) returns
		var paramTypes []string
		for _, param := range t.Parameters {
			paramTypes = append(paramTypes, g.generateTypeAnnotation(param))
		}

		result := "func(" + strings.Join(paramTypes, ", ") + ")"

		if len(t.Returns) == 1 {
			result += " " + g.generateTypeAnnotation(t.Returns[0])
		} else if len(t.Returns) > 1 {
			var returnTypes []string
			for _, ret := range t.Returns {
				returnTypes = append(returnTypes, g.generateTypeAnnotation(ret))
			}
			result += " (" + strings.Join(returnTypes, ", ") + ")"
		}

		return result
	default:
		return "any"
	}
}

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
		valStr := g.exprToString(val)

		// Apply type coercion if we have matching return types
		// This handles cases like: return n * 1000 -> return time.Duration(n * 1000)
		if i < len(g.currentReturnTypes) {
			valStr = g.coerceReturnValue(valStr, val, g.currentReturnTypes[i])
		}

		values[i] = valStr
	}

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
		g.writeLine(fmt.Sprintf("for %s, %s := range %s {", stmt.Index.Value, stmt.Variable.Value, collection))
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

	// for i from 0 to N  →  for i := range N  (range-over-int, Go 1.22+)
	if !stmt.Through && start == "0" {
		g.writeLine(fmt.Sprintf("for %s := range %s {", varName, end))
	} else {
		var condition string
		if stmt.Through {
			condition = fmt.Sprintf("%s <= %s", varName, end)
		} else {
			condition = fmt.Sprintf("%s < %s", varName, end)
		}
		g.writeLine(fmt.Sprintf("for %s := %s; %s; %s++ {", varName, start, condition, varName))
	}

	g.indent++
	g.generateBlock(stmt.Body)
	g.indent--

	g.writeLine("}")
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

// uniqueId generates unique identifiers to prevent variable shadowing
func (g *Generator) uniqueId(prefix string) string {
	g.tempCounter++
	return fmt.Sprintf("%s_%d", prefix, g.tempCounter)
}
















