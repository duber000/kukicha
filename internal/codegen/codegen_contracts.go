package codegen

import (
	"fmt"
	"strings"

	"github.com/duber000/kukicha/internal/ast"
)

// kukichaToGoExpr translates a Kukicha boolean expression string to Go syntax.
// Handles keyword substitutions: and→&&, or→||, not→!, equals→==, empty→nil.
func kukichaToGoExpr(expr string) string {
	// Replace multi-word keywords first
	result := expr
	result = strings.ReplaceAll(result, " and ", " && ")
	result = strings.ReplaceAll(result, " or ", " || ")
	result = strings.ReplaceAll(result, " equals ", " == ")

	// Replace "not " prefix and " not " infix (boolean negation)
	if strings.HasPrefix(result, "not ") {
		result = "!" + result[4:]
	}
	result = strings.ReplaceAll(result, " not ", " !")

	// Replace empty keyword with nil
	result = strings.ReplaceAll(result, "empty", "nil")

	return result
}

// generateRequiresChecks emits precondition panic checks for # kuki:requires directives.
func (g *Generator) generateRequiresChecks(decl *ast.FunctionDecl) {
	if g.releaseMode {
		return
	}
	for _, dir := range decl.Directives {
		if dir.Name == "requires" && len(dir.Args) > 0 {
			condition := kukichaToGoExpr(dir.Args[0])
			g.writeLine(fmt.Sprintf(`if !(%s) { panic("requires violated: %s") }`, condition, escapeGoString(dir.Args[0])))
		}
	}
}

// generateEnsuresChecks emits postcondition panic checks for # kuki:ensures directives.
// Called at the end of the function body (before the closing brace) as deferred checks.
func (g *Generator) generateEnsuresChecks(decl *ast.FunctionDecl) {
	if g.releaseMode {
		return
	}
	for _, dir := range decl.Directives {
		if dir.Name == "ensures" && len(dir.Args) > 0 {
			condition := kukichaToGoExpr(dir.Args[0])
			// Use named returns + defer to check postconditions on all return paths.
			// The caller (generateFunctionDecl) will have set up named return variables.
			g.writeLine(fmt.Sprintf(`defer func() { if !(%s) { panic("ensures violated: %s") } }()`, condition, escapeGoString(dir.Args[0])))
		}
	}
}

// generateInvariantMethod emits a Validate() method for types with # kuki:invariant directives.
func (g *Generator) generateInvariantMethod(decl *ast.TypeDecl) {
	if g.releaseMode {
		return
	}

	var invariants []ast.Directive
	for _, dir := range decl.Directives {
		if dir.Name == "invariant" && len(dir.Args) > 0 {
			invariants = append(invariants, dir)
		}
	}
	if len(invariants) == 0 {
		return
	}

	typeName := decl.Name.Value
	receiverName := strings.ToLower(typeName[:1])

	g.writeLine("")
	g.writeLine(fmt.Sprintf("func (%s %s) Validate() {", receiverName, typeName))
	g.indent++
	for _, dir := range invariants {
		// Replace "self." with the receiver name
		condition := kukichaToGoExpr(strings.ReplaceAll(dir.Args[0], "self.", receiverName+"."))
		original := dir.Args[0]
		g.writeLine(fmt.Sprintf(`if !(%s) { panic("invariant violated: %s") }`, condition, escapeGoString(original)))
	}
	g.indent--
	g.writeLine("}")
}

// hasEnsuresDirective returns true if the function has any # kuki:ensures directives.
func hasEnsuresDirective(decl *ast.FunctionDecl) bool {
	for _, dir := range decl.Directives {
		if dir.Name == "ensures" && len(dir.Args) > 0 {
			return true
		}
	}
	return false
}

// hasRequiresDirective returns true if the function has any # kuki:requires directives.
func hasRequiresDirective(decl *ast.FunctionDecl) bool {
	for _, dir := range decl.Directives {
		if dir.Name == "requires" && len(dir.Args) > 0 {
			return true
		}
	}
	return false
}

// hasInvariantDirective returns true if the type has any # kuki:invariant directives.
func hasInvariantDirective(decl *ast.TypeDecl) bool {
	for _, dir := range decl.Directives {
		if dir.Name == "invariant" && len(dir.Args) > 0 {
			return true
		}
	}
	return false
}

// escapeGoString escapes a string for use inside a Go string literal.
func escapeGoString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}

// ensuresReturnNames generates named return variable names for ensures directives.
// For "result" references in ensures expressions, the first return value is named "result".
// For multiple returns, they're named result, result2, result3, etc.
func ensuresReturnNames(decl *ast.FunctionDecl) []string {
	if len(decl.Returns) == 0 {
		return nil
	}
	names := make([]string, len(decl.Returns))
	names[0] = "result"
	for i := 1; i < len(decl.Returns); i++ {
		names[i] = fmt.Sprintf("result%d", i+1)
	}
	return names
}
