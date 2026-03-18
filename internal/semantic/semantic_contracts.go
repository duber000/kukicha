package semantic

import (
	"fmt"
	"strings"

	"github.com/duber000/kukicha/internal/ast"
)

// validateContractDirectives checks that # kuki:requires and # kuki:ensures
// directives on a function declaration are well-formed.
func (a *Analyzer) validateContractDirectives(decl *ast.FunctionDecl) {
	for _, dir := range decl.Directives {
		switch dir.Name {
		case "requires":
			if len(dir.Args) == 0 {
				a.error(ast.Position{Line: dir.Token.Line, Column: dir.Token.Column, File: dir.Token.File},
					"# kuki:requires directive requires a condition expression")
			}
		case "ensures":
			if len(dir.Args) == 0 {
				a.error(ast.Position{Line: dir.Token.Line, Column: dir.Token.Column, File: dir.Token.File},
					"# kuki:ensures directive requires a condition expression")
				continue
			}
			if len(decl.Returns) == 0 {
				a.error(ast.Position{Line: dir.Token.Line, Column: dir.Token.Column, File: dir.Token.File},
					"# kuki:ensures directive requires the function to have return values")
			}
		}
	}
}

// validateInvariantDirectives checks that # kuki:invariant directives on a type
// declaration are well-formed.
func (a *Analyzer) validateInvariantDirectives(decl *ast.TypeDecl) {
	for _, dir := range decl.Directives {
		if dir.Name == "invariant" {
			if len(dir.Args) == 0 {
				a.error(ast.Position{Line: dir.Token.Line, Column: dir.Token.Column, File: dir.Token.File},
					"# kuki:invariant directive requires a condition expression")
				continue
			}
			// Validate that the invariant references self.field where field exists
			expr := dir.Args[0]
			if decl.AliasType != nil {
				a.error(ast.Position{Line: dir.Token.Line, Column: dir.Token.Column, File: dir.Token.File},
					"# kuki:invariant cannot be applied to type aliases")
				continue
			}
			// Check that referenced fields exist via self.fieldName pattern
			a.validateInvariantFieldRefs(decl, expr, dir)
		}
	}
}

// validateInvariantFieldRefs checks that self.X references in an invariant expression
// refer to actual fields on the struct.
func (a *Analyzer) validateInvariantFieldRefs(decl *ast.TypeDecl, expr string, dir ast.Directive) {
	// Find all self.X references
	idx := 0
	for idx < len(expr) {
		selfIdx := strings.Index(expr[idx:], "self.")
		if selfIdx == -1 {
			break
		}
		selfIdx += idx
		fieldStart := selfIdx + 5 // len("self.")
		if fieldStart >= len(expr) {
			break
		}
		// Extract field name (identifier characters)
		fieldEnd := fieldStart
		for fieldEnd < len(expr) && isIdentChar(expr[fieldEnd]) {
			fieldEnd++
		}
		if fieldEnd == fieldStart {
			idx = fieldEnd + 1
			continue
		}
		fieldName := expr[fieldStart:fieldEnd]
		// Check if field exists on the struct
		found := false
		for _, f := range decl.Fields {
			if f.Name.Value == fieldName {
				found = true
				break
			}
		}
		if !found {
			a.error(ast.Position{Line: dir.Token.Line, Column: dir.Token.Column, File: dir.Token.File},
				fmt.Sprintf("# kuki:invariant references unknown field 'self.%s' on type '%s'", fieldName, decl.Name.Value))
		}
		idx = fieldEnd
	}
}

func isIdentChar(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '_'
}
