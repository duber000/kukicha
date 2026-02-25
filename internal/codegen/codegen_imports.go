package codegen

import (
	"fmt"
	"sort"
	"strings"

	"github.com/duber000/kukicha/internal/ast"
)

// addImport adds an auto-import
func (g *Generator) addImport(path string) {
	g.autoImports[path] = true
}

func (g *Generator) generateImports() {
	// Collect all imports
	imports := make(map[string]string) // path -> alias

	for _, imp := range g.program.Imports {
		path := imp.Path.Value
		alias := ""
		if imp.Alias != nil {
			alias = imp.Alias.Value
		}

		// Rewrite stdlib imports to full module path
		path = g.rewriteStdlibImport(path)

		imports[path] = alias
	}

	// Check if we need fmt for string interpolation, print builtin, or onerr explain
	needsFmt := g.needsStringInterpolation() || g.needsPrintBuiltin() || g.needsExplain()
	if needsFmt {
		imports["fmt"] = ""
	}

	// Check if we need os for MCP target (printbuiltin uses os.Stderr)
	if g.mcpTarget && g.needsPrintBuiltin() {
		imports["os"] = ""
	}

	// Check if we need errors for error expressions
	needsErrors := g.needsErrorsPackage()
	if needsErrors {
		imports["errors"] = ""
	}

	// Add auto-imports (e.g., cmp for generic constraints)
	for path := range g.autoImports {
		if _, exists := imports[path]; !exists {
			imports[path] = ""
		}
	}

	// Detect package name collisions between Kukicha stdlib imports and Go imports.
	// If two imports resolve to the same Go package name (e.g., stdlib/json and encoding/json
	// both resolve to "json"), auto-alias the Kukicha stdlib import to prevent Go compile errors.
	pkgNameToPath := make(map[string][]string)
	for path, alias := range imports {
		effectiveName := alias
		if effectiveName == "" {
			effectiveName = extractPkgName(path)
		}
		pkgNameToPath[effectiveName] = append(pkgNameToPath[effectiveName], path)
	}
	kukichaStdlibPrefix := g.stdlibModuleBase + "/stdlib/"
	for pkgName, paths := range pkgNameToPath {
		if len(paths) <= 1 {
			continue
		}
		for _, path := range paths {
			if strings.HasPrefix(path, kukichaStdlibPrefix) && imports[path] == "" {
				aliased := "kuki" + pkgName
				imports[path] = aliased
				g.pkgAliases[pkgName] = aliased
			}
		}
	}

	// Auto-alias imports that collide with Go built-in types or keywords.
	// e.g., stdlib/string -> kukistring (because "string" is a Go built-in type)
	goBuiltins := map[string]bool{
		"string": true, "int": true, "int8": true, "int16": true, "int32": true, "int64": true,
		"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true,
		"float32": true, "float64": true, "complex64": true, "complex128": true,
		"bool": true, "byte": true, "rune": true, "error": true, "any": true,
	}
	for path, alias := range imports {
		if alias != "" {
			continue // Already aliased
		}
		pkgName := extractPkgName(path)
		if goBuiltins[pkgName] {
			aliased := "kuki" + pkgName
			imports[path] = aliased
			g.pkgAliases[pkgName] = aliased
		}
	}

	// Auto-alias imports ending with version suffixes like /v2, /v3.
	// Go uses the last path segment as the package name, so "encoding/json/v2"
	// would be imported as "v2" without an alias. We need to alias it to "json".
	for path, alias := range imports {
		if alias != "" {
			continue // Already has an explicit alias
		}
		parts := strings.Split(path, "/")
		if len(parts) < 2 {
			continue
		}
		lastSegment := parts[len(parts)-1]
		// Check if last segment is a version like v2, v3, etc.
		if len(lastSegment) >= 2 && lastSegment[0] == 'v' && lastSegment[1] >= '0' && lastSegment[1] <= '9' {
			// Use the second-to-last segment as the alias
			pkgName := parts[len(parts)-2]
			// Handle .v suffix in the package name (e.g., yaml.v3)
			if idx := strings.Index(pkgName, ".v"); idx != -1 {
				pkgName = pkgName[:idx]
			}
			imports[path] = pkgName
		}
	}

	// Generate import block
	specs := make([]importSpec, 0, len(imports))
	for path, alias := range imports {
		specs = append(specs, importSpec{path: path, alias: alias})
	}

	sort.Slice(specs, func(i, j int) bool {
		if specs[i].path == specs[j].path {
			return specs[i].alias < specs[j].alias
		}
		return specs[i].path < specs[j].path
	})

	if len(specs) == 1 {
		spec := specs[0]
		if spec.alias != "" {
			g.writeLine(fmt.Sprintf("import %s \"%s\"", spec.alias, spec.path))
		} else {
			g.writeLine(fmt.Sprintf("import \"%s\"", spec.path))
		}
		return
	}

	g.writeLine("import (")
	g.indent++
	for _, spec := range specs {
		if spec.alias != "" {
			g.writeLine(fmt.Sprintf("%s \"%s\"", spec.alias, spec.path))
		} else {
			g.writeLine(fmt.Sprintf("\"%s\"", spec.path))
		}
	}
	g.indent--
	g.writeLine(")")
}

type importSpec struct {
	path  string
	alias string
}

// extractPkgName returns the Go package name from an import path.
// e.g., "encoding/json" -> "json", "net/http" -> "http",
// "github.com/duber000/kukicha/stdlib/json" -> "json",
// "gopkg.in/yaml.v3" -> "yaml"
func extractPkgName(importPath string) string {
	parts := strings.Split(importPath, "/")
	name := parts[len(parts)-1]

	// Handle version suffixes: gopkg.in/yaml.v3 -> yaml
	if idx := strings.Index(name, ".v"); idx != -1 {
		name = name[:idx]
	}

	// Handle Go module major version directories: encoding/json/v2 -> json
	if len(parts) >= 2 && len(name) >= 2 && name[0] == 'v' && name[1] >= '0' && name[1] <= '9' {
		name = parts[len(parts)-2]
		if idx := strings.Index(name, ".v"); idx != -1 {
			name = name[:idx]
		}
	}

	return name
}

// rewriteStdlibImport rewrites stdlib/ import paths to full module paths
// e.g., "stdlib/json" → "<stdlibModuleBase>/stdlib/json"
// Note: Returns path without quotes (quotes are added by generateImports)
func (g *Generator) rewriteStdlibImport(path string) string {
	// Remove quotes to check the path
	cleanPath := strings.Trim(path, "\"")

	// Rewrite stdlib/ prefix to full module path
	if strings.HasPrefix(cleanPath, "stdlib/") {
		return g.stdlibModuleBase + "/" + cleanPath
	}

	return cleanPath
}

func (g *Generator) scanForAutoImports() {
	for _, decl := range g.program.Declarations {
		if fn, ok := decl.(*ast.FunctionDecl); ok {
			if fn.Body != nil {
				g.scanBlockForAutoImports(fn.Body)
			}
		}
	}
}

// scanForFunctionDefaults collects function parameter names and default values
// This information is used when generating function calls with named arguments
// or when arguments are omitted (relying on default values)
func (g *Generator) scanForFunctionDefaults() {
	for _, decl := range g.program.Declarations {
		if fn, ok := decl.(*ast.FunctionDecl); ok {
			defaults := &FuncDefaults{
				ParamNames:    make([]string, len(fn.Parameters)),
				DefaultValues: make([]ast.Expression, len(fn.Parameters)),
				HasVariadic:   len(fn.Parameters) > 0 && fn.Parameters[len(fn.Parameters)-1].Variadic,
			}

			for i, param := range fn.Parameters {
				defaults.ParamNames[i] = param.Name.Value
				defaults.DefaultValues[i] = param.DefaultValue // may be nil
			}

			g.funcDefaults[fn.Name.Value] = defaults
		}
	}
}

func (g *Generator) scanBlockForAutoImports(block *ast.BlockStmt) {
	for _, stmt := range block.Statements {
		g.scanStmtForAutoImports(stmt)
	}
}

func (g *Generator) scanStmtForAutoImports(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.VarDeclStmt:
		for _, val := range s.Values {
			g.scanExprForAutoImports(val)
		}
		if s.OnErr != nil {
			g.scanExprForAutoImports(s.OnErr.Handler)
		}
	case *ast.AssignStmt:
		for _, val := range s.Values {
			g.scanExprForAutoImports(val)
		}
		if s.OnErr != nil {
			g.scanExprForAutoImports(s.OnErr.Handler)
		}
	case *ast.ReturnStmt:
		for _, val := range s.Values {
			g.scanExprForAutoImports(val)
		}
	case *ast.IfStmt:
		if s.Init != nil {
			g.scanStmtForAutoImports(s.Init)
		}
		g.scanExprForAutoImports(s.Condition)
		if s.Consequence != nil {
			g.scanBlockForAutoImports(s.Consequence)
		}
		if s.Alternative != nil {
			g.scanStmtForAutoImports(s.Alternative)
		}
	case *ast.SwitchStmt:
		if s.Expression != nil {
			g.scanExprForAutoImports(s.Expression)
		}
		for _, c := range s.Cases {
			for _, v := range c.Values {
				g.scanExprForAutoImports(v)
			}
			if c.Body != nil {
				g.scanBlockForAutoImports(c.Body)
			}
		}
		if s.Otherwise != nil && s.Otherwise.Body != nil {
			g.scanBlockForAutoImports(s.Otherwise.Body)
		}
	case *ast.SelectStmt:
		for _, c := range s.Cases {
			if c.Recv != nil {
				g.scanExprForAutoImports(c.Recv.Channel)
			}
			if c.Send != nil {
				g.scanExprForAutoImports(c.Send.Value)
				g.scanExprForAutoImports(c.Send.Channel)
			}
			if c.Body != nil {
				g.scanBlockForAutoImports(c.Body)
			}
		}
		if s.Otherwise != nil && s.Otherwise.Body != nil {
			g.scanBlockForAutoImports(s.Otherwise.Body)
		}
	case *ast.TypeSwitchStmt:
		g.scanExprForAutoImports(s.Expression)
		for _, c := range s.Cases {
			if c.Body != nil {
				g.scanBlockForAutoImports(c.Body)
			}
		}
		if s.Otherwise != nil && s.Otherwise.Body != nil {
			g.scanBlockForAutoImports(s.Otherwise.Body)
		}
	case *ast.ForRangeStmt:
		g.scanExprForAutoImports(s.Collection)
		if s.Body != nil {
			g.scanBlockForAutoImports(s.Body)
		}
	case *ast.ForNumericStmt:
		g.scanExprForAutoImports(s.Start)
		g.scanExprForAutoImports(s.End)
		if s.Body != nil {
			g.scanBlockForAutoImports(s.Body)
		}
	case *ast.ForConditionStmt:
		g.scanExprForAutoImports(s.Condition)
		if s.Body != nil {
			g.scanBlockForAutoImports(s.Body)
		}
	case *ast.ExpressionStmt:
		g.scanExprForAutoImports(s.Expression)
		if s.OnErr != nil {
			g.scanExprForAutoImports(s.OnErr.Handler)
		}
	case *ast.DeferStmt:
		g.scanExprForAutoImports(s.Call)
	case *ast.GoStmt:
		if s.Call != nil {
			g.scanExprForAutoImports(s.Call)
		}
		if s.Block != nil {
			g.scanBlockForAutoImports(s.Block)
		}
	case *ast.SendStmt:
		g.scanExprForAutoImports(s.Value)
		g.scanExprForAutoImports(s.Channel)
	case *ast.ElseStmt:
		if s.Body != nil {
			g.scanBlockForAutoImports(s.Body)
		}
	}
}

func (g *Generator) scanExprForAutoImports(expr ast.Expression) {
	if expr == nil {
		return
	}

	switch e := expr.(type) {
	case *ast.BinaryExpr:
		g.scanExprForAutoImports(e.Left)
		g.scanExprForAutoImports(e.Right)
	case *ast.UnaryExpr:
		g.scanExprForAutoImports(e.Right)
	case *ast.PipeExpr:
		g.scanExprForAutoImports(e.Left)
		g.scanExprForAutoImports(e.Right)
	case *ast.CallExpr:
		g.scanExprForAutoImports(e.Function)
		for _, arg := range e.Arguments {
			g.scanExprForAutoImports(arg)
		}
	case *ast.MethodCallExpr:
		g.scanExprForAutoImports(e.Object)
		for _, arg := range e.Arguments {
			g.scanExprForAutoImports(arg)
		}
	case *ast.IndexExpr:
		g.scanExprForAutoImports(e.Left)
		g.scanExprForAutoImports(e.Index)
	case *ast.SliceExpr:
		g.scanExprForAutoImports(e.Left)
		if e.Start != nil {
			g.scanExprForAutoImports(e.Start)
		}
		if e.End != nil {
			g.scanExprForAutoImports(e.End)
		}
	case *ast.FunctionLiteral:
		if e.Body != nil {
			g.scanBlockForAutoImports(e.Body)
		}
	case *ast.ArrowLambda:
		if e.Body != nil {
			g.scanExprForAutoImports(e.Body)
		}
		if e.Block != nil {
			g.scanBlockForAutoImports(e.Block)
		}
	case *ast.StructLiteralExpr:
		for _, f := range e.Fields {
			g.scanExprForAutoImports(f.Value)
		}
	case *ast.ListLiteralExpr:
		for _, el := range e.Elements {
			g.scanExprForAutoImports(el)
		}
	case *ast.MapLiteralExpr:
		for _, p := range e.Pairs {
			g.scanExprForAutoImports(p.Key)
			g.scanExprForAutoImports(p.Value)
		}
	}
}

