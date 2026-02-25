package semantic

import (
	"fmt"
	"strings"

	"github.com/duber000/kukicha/internal/ast"
)

func (a *Analyzer) checkPackageName() {
	if a.program.PetioleDecl == nil {
		return
	}

	name := a.program.PetioleDecl.Name.Value

	// Kukicha's own stdlib packages are allowed to use Go stdlib names.
	// Detect stdlib packages by the "stdlib/" prefix in the source file path.
	if strings.Contains(a.sourceFile, "stdlib/") {
		return
	}

	// List of reserved Go standard library packages
	reservedPackages := map[string]bool{
		"bufio": true, "bytes": true, "context": true, "crypto": true,
		"database": true, "encoding": true, "errors": true, "flag": true,
		"fmt": true, "html": true, "image": true, "io": true,
		"iter": true, "log": true, "math": true, "mime": true,
		"net": true, "os": true, "path": true, "plugin": true,
		"reflect": true, "regexp": true, "runtime": true, "slices": true,
		"sort": true, "strconv": true, "strings": true, "sync": true,
		"syscall": true, "testing": true, "text": true, "time": true,
		"unicode": true, "unsafe": true,
	}

	if reservedPackages[name] {
		a.error(a.program.PetioleDecl.Pos(), fmt.Sprintf("package name '%s' conflicts with Go standard library package", name))
	}
}

func (a *Analyzer) checkSkillDecl() {
	skill := a.program.SkillDecl
	if skill == nil {
		return
	}

	// Skill name must be exported (start with uppercase)
	if skill.Name != nil && !isExported(skill.Name.Value) {
		a.error(skill.Name.Pos(), fmt.Sprintf("skill name '%s' must be exported (start with uppercase letter)", skill.Name.Value))
	}

	// Skill requires petiole (skills are packages, not main programs)
	if a.program.PetioleDecl == nil {
		a.error(skill.Pos(), "skill declaration requires a petiole declaration (skills are packages)")
	}

	// Warn if description is empty
	if skill.Description == "" {
		a.error(skill.Pos(), "skill should have a description (skills should be self-documenting)")
	}

	// Basic semver validation if version is provided
	if skill.Version != "" {
		if !isBasicSemver(skill.Version) {
			a.error(skill.Pos(), fmt.Sprintf("skill version '%s' should follow semver format (e.g., '1.0.0')", skill.Version))
		}
	}
}

// isBasicSemver performs a basic check that a version string looks semver-like
func isBasicSemver(v string) bool {
	// Allow optional leading 'v'
	s := strings.TrimPrefix(v, "v")
	parts := strings.Split(s, ".")
	if len(parts) < 2 || len(parts) > 3 {
		return false
	}
	for _, part := range parts {
		if len(part) == 0 {
			return false
		}
		for _, r := range part {
			if r < '0' || r > '9' {
				return false
			}
		}
	}
	return true
}

// collectDeclarations collects all top-level declarations
func (a *Analyzer) collectDeclarations() {
	// Collect imports
	for _, imp := range a.program.Imports {
		err := a.symbolTable.Define(&Symbol{
			Name:    a.extractPackageName(imp),
			Kind:    SymbolVariable, // Treat as variable for now
			Type:    &TypeInfo{Kind: TypeKindUnknown},
			Defined: imp.Pos(),
		})
		if err != nil {
			a.error(imp.Pos(), err.Error())
		}
	}

	for _, decl := range a.program.Declarations {
		switch d := decl.(type) {
		case *ast.TypeDecl:
			a.collectTypeDecl(d)
		case *ast.InterfaceDecl:
			a.collectInterfaceDecl(d)
		case *ast.FunctionDecl:
			a.collectFunctionDecl(d)
		}
	}
}

func (a *Analyzer) collectTypeDecl(decl *ast.TypeDecl) {
	// Check export rules: PascalCase = exported, camelCase = unexported
	if !isValidIdentifier(decl.Name.Value) {
		a.error(decl.Name.Pos(), fmt.Sprintf("invalid type name '%s'", decl.Name.Value))
		return
	}

	// Determine type kind based on alias vs struct
	typeKind := TypeKindStruct
	if decl.AliasType != nil {
		typeKind = TypeKindFunction
	}

	// Add type to symbol table
	symbol := &Symbol{
		Name:     decl.Name.Value,
		Kind:     SymbolType,
		Type:     &TypeInfo{Kind: typeKind, Name: decl.Name.Value},
		Defined:  decl.Name.Pos(),
		Exported: isExported(decl.Name.Value),
	}

	if err := a.symbolTable.Define(symbol); err != nil {
		a.error(decl.Name.Pos(), err.Error())
	}
}

func (a *Analyzer) collectInterfaceDecl(decl *ast.InterfaceDecl) {
	// Check export rules
	if !isValidIdentifier(decl.Name.Value) {
		a.error(decl.Name.Pos(), fmt.Sprintf("invalid interface name '%s'", decl.Name.Value))
		return
	}

	// Add interface to symbol table
	symbol := &Symbol{
		Name:     decl.Name.Value,
		Kind:     SymbolInterface,
		Type:     &TypeInfo{Kind: TypeKindInterface, Name: decl.Name.Value},
		Defined:  decl.Name.Pos(),
		Exported: isExported(decl.Name.Value),
	}

	if err := a.symbolTable.Define(symbol); err != nil {
		a.error(decl.Name.Pos(), err.Error())
	}
}

func (a *Analyzer) collectFunctionDecl(decl *ast.FunctionDecl) {
	// Check export rules
	if !isValidIdentifier(decl.Name.Value) {
		a.error(decl.Name.Pos(), fmt.Sprintf("invalid function name '%s'", decl.Name.Value))
		return
	}

	// Build function type
	params := make([]*TypeInfo, len(decl.Parameters))
	paramNames := make([]string, len(decl.Parameters))
	hasVariadic := false
	defaultCount := 0
	for i, param := range decl.Parameters {
		params[i] = a.typeAnnotationToTypeInfo(param.Type)
		paramNames[i] = param.Name.Value
		if param.Variadic {
			hasVariadic = true
		}
		if param.DefaultValue != nil {
			defaultCount++
		}
	}

	returns := make([]*TypeInfo, len(decl.Returns))
	for i, ret := range decl.Returns {
		returns[i] = a.typeAnnotationToTypeInfo(ret)
	}

	funcType := &TypeInfo{
		Kind:         TypeKindFunction,
		Params:       params,
		Returns:      returns,
		Variadic:     hasVariadic,
		ParamNames:   paramNames,
		DefaultCount: defaultCount,
	}

	// Add function to symbol table
	symbol := &Symbol{
		Name:     decl.Name.Value,
		Kind:     SymbolFunction,
		Type:     funcType,
		Defined:  decl.Name.Pos(),
		Exported: isExported(decl.Name.Value),
	}

	if err := a.symbolTable.Define(symbol); err != nil {
		a.error(decl.Name.Pos(), err.Error())
	}
}

// analyzeDeclarations performs deep analysis of declarations
func (a *Analyzer) analyzeDeclarations() {
	for _, decl := range a.program.Declarations {
		switch d := decl.(type) {
		case *ast.TypeDecl:
			a.analyzeTypeDecl(d)
		case *ast.InterfaceDecl:
			a.analyzeInterfaceDecl(d)
		case *ast.FunctionDecl:
			a.analyzeFunctionDecl(d)
		case *ast.VarDeclStmt:
			a.analyzeGlobalVarDecl(d)
		}
	}
}

func (a *Analyzer) analyzeTypeDecl(decl *ast.TypeDecl) {
	// Type alias: validate the alias type annotation
	if decl.AliasType != nil {
		a.validateTypeAnnotation(decl.AliasType)
		return
	}

	// Validate field types exist
	for _, field := range decl.Fields {
		if !isValidIdentifier(field.Name.Value) {
			a.error(field.Name.Pos(), fmt.Sprintf("invalid field name '%s'", field.Name.Value))
		}

		// Check that field type exists
		a.validateTypeAnnotation(field.Type)
	}
}

func (a *Analyzer) analyzeInterfaceDecl(decl *ast.InterfaceDecl) {
	// Validate method signatures
	for _, method := range decl.Methods {
		if !isValidIdentifier(method.Name.Value) {
			a.error(method.Name.Pos(), fmt.Sprintf("invalid method name '%s'", method.Name.Value))
		}

		// Validate parameter types
		for _, param := range method.Parameters {
			a.validateTypeAnnotation(param.Type)
		}

		// Validate return types
		for _, ret := range method.Returns {
			a.validateTypeAnnotation(ret)
		}
	}
}

func (a *Analyzer) analyzeGlobalVarDecl(stmt *ast.VarDeclStmt) {
	// Analyze values
	for _, val := range stmt.Values {
		a.analyzeExpression(val)
	}

	// Register each name in the global scope
	for _, name := range stmt.Names {
		if !isValidIdentifier(name.Value) {
			a.error(name.Pos(), fmt.Sprintf("invalid variable name '%s'", name.Value))
			continue
		}

		// Infer type from values or use explicit type
		var varType *TypeInfo
		if stmt.Type != nil {
			varType = a.typeAnnotationToTypeInfo(stmt.Type)
		} else if len(stmt.Values) > 0 {
			varType = a.analyzeExpression(stmt.Values[0])
		} else {
			varType = &TypeInfo{Kind: TypeKindUnknown}
		}

		symbol := &Symbol{
			Name:     name.Value,
			Kind:     SymbolVariable,
			Type:     varType,
			Defined:  name.Pos(),
			Exported: isExported(name.Value),
		}

		if err := a.symbolTable.Define(symbol); err != nil {
			a.error(name.Pos(), err.Error())
		}
	}
}

func (a *Analyzer) analyzeFunctionDecl(decl *ast.FunctionDecl) {
	// Enter new scope for function
	a.symbolTable.EnterScope()
	defer a.symbolTable.ExitScope()

	// Track current function for return checking
	a.currentFunc = decl

	// Add receiver if present (for methods)
	if decl.Receiver != nil {
		a.validateTypeAnnotation(decl.Receiver.Type)

		receiverSymbol := &Symbol{
			Name:    decl.Receiver.Name.Value,
			Kind:    SymbolParameter,
			Type:    a.typeAnnotationToTypeInfo(decl.Receiver.Type),
			Defined: decl.Receiver.Name.Pos(),
		}
		if err := a.symbolTable.Define(receiverSymbol); err != nil {
			a.error(decl.Receiver.Name.Pos(), err.Error())
		}
	}

	// Validate variadic parameters (must be last, only one)
	variadicCount := 0
	for i, param := range decl.Parameters {
		if param.Variadic {
			variadicCount++
			if variadicCount > 1 {
				a.error(param.Name.Pos(), "only one variadic parameter allowed per function")
			}
			if i != len(decl.Parameters)-1 {
				a.error(param.Name.Pos(), "variadic parameter must be the last parameter")
			}
		}
	}

	// Add parameters to scope
	for _, param := range decl.Parameters {
		if !isValidIdentifier(param.Name.Value) {
			a.error(param.Name.Pos(), fmt.Sprintf("invalid parameter name '%s'", param.Name.Value))
		}

		a.validateTypeAnnotation(param.Type)

		paramSymbol := &Symbol{
			Name:    param.Name.Value,
			Kind:    SymbolParameter,
			Type:    a.typeAnnotationToTypeInfo(param.Type),
			Defined: param.Name.Pos(),
		}
		if err := a.symbolTable.Define(paramSymbol); err != nil {
			a.error(param.Name.Pos(), err.Error())
		}
	}

	// Validate return types exist
	for _, ret := range decl.Returns {
		a.validateTypeAnnotation(ret)
	}

	// Analyze function body
	if decl.Body != nil {
		a.analyzeBlock(decl.Body)
	}

	a.currentFunc = nil
}
