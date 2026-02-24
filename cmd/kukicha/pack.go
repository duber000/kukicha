package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/duber000/kukicha/internal/ast"
	"github.com/duber000/kukicha/internal/codegen"
	"gopkg.in/yaml.v3"
)

func packCommand(filename string, outputDir string) {
	absFile, err := filepath.Abs(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving file path: %v\n", err)
		os.Exit(1)
	}

	// 1. Parse and analyze
	program, returnCounts, err := loadAndAnalyze(absFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// 2. Validate skill declaration exists
	if program.SkillDecl == nil {
		fmt.Fprintln(os.Stderr, "Error: no skill declaration found in file")
		os.Exit(1)
	}
	skill := program.SkillDecl

	// Detect target from source (default to mcp for skills)
	t, readErr := detectTargetFromFile(absFile)
	if readErr != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not read %s for target detection: %v\n", absFile, readErr)
	}
	if t != "" {
		program.Target = t
	} else {
		program.Target = "mcp"
	}

	// 3. Determine output directory
	if outputDir == "" {
		outputDir = filepath.Join(filepath.Dir(absFile), toSnakeCase(skill.Name.Value))
	}

	// 4. Extract function schemas from AST
	functions := extractFunctionSchemas(program)

	// 5. Generate SKILL.md
	skillMD := generateSkillMD(skill, functions)

	// 6. Create output directory structure
	scriptsDir := filepath.Join(outputDir, "scripts")
	if err := os.MkdirAll(scriptsDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Write SKILL.md
	skillMDPath := filepath.Join(outputDir, "SKILL.md")
	if err := os.WriteFile(skillMDPath, []byte(skillMD), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing SKILL.md: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Generated %s\n", skillMDPath)

	// 7. Compile binary with target=mcp
	gen := codegen.New(program)
	gen.SetSourceFile(absFile)
	gen.SetExprReturnCounts(returnCounts)
	gen.SetMCPTarget(true)
	goCode, err := gen.Generate()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Code generation error: %v\n", err)
		os.Exit(1)
	}

	formatted, fmtErr := format.Source([]byte(goCode))
	if fmtErr != nil {
		formatted = []byte(goCode)
	}

	// Write Go file to temp location for building
	goFile := strings.TrimSuffix(absFile, ".kuki") + ".go"
	if err := os.WriteFile(goFile, formatted, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing Go file: %v\n", err)
		os.Exit(1)
	}

	// If the generated code imports Kukicha stdlib, extract it and configure go.mod
	projectDir := findProjectDir(absFile)
	if needsStdlib(goCode, projectDir) {
		stdlibPath, stdlibErr := ensureStdlib(projectDir)
		if stdlibErr != nil {
			fmt.Fprintf(os.Stderr, "Error extracting stdlib: %v\n", stdlibErr)
			os.Exit(1)
		}
		if modErr := ensureGoMod(projectDir, stdlibPath); modErr != nil {
			fmt.Fprintf(os.Stderr, "Error updating go.mod: %v\n", modErr)
			os.Exit(1)
		}
	}

	// Build binary into scripts/
	binaryName := toSnakeCase(skill.Name.Value)
	binaryPath := filepath.Join(scriptsDir, binaryName)
	cmd := exec.Command("go", "build", "-o", binaryPath, goFile)
	cmd.Dir = projectDir
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf
	if err := cmd.Run(); err != nil {
		if stderrBuf.Len() > 0 {
			os.Stderr.Write(rewriteGoErrors(stderrBuf.Bytes(), goFile, absFile))
		}
		fmt.Fprintf(os.Stderr, "Error building binary: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Built binary: %s\n", binaryPath)
	fmt.Printf("Skill packed successfully in %s\n", outputDir)
}

// FunctionSchema holds extracted metadata for a function declaration
type FunctionSchema struct {
	Name        string
	Description string
	Parameters  []ParameterSchema
	Returns     []string
}

// ParameterSchema holds extracted metadata for a function parameter
type ParameterSchema struct {
	Name       string
	Type       string
	Default    any
	HasDefault bool
}

func extractFunctionSchemas(program *ast.Program) []FunctionSchema {
	var schemas []FunctionSchema

	for _, decl := range program.Declarations {
		fn, ok := decl.(*ast.FunctionDecl)
		if !ok {
			continue
		}

		// Only include exported functions (starting with uppercase)
		if len(fn.Name.Value) == 0 || !unicode.IsUpper(rune(fn.Name.Value[0])) {
			continue
		}

		// Skip methods (they have receivers)
		if fn.Receiver != nil {
			continue
		}

		schema := FunctionSchema{
			Name: fn.Name.Value,
		}

		// Extract parameters
		for _, param := range fn.Parameters {
			ps := ParameterSchema{
				Name: param.Name.Value,
				Type: typeToJSONSchemaType(param.Type),
			}
			if param.DefaultValue != nil {
				if def, ok := defaultValueToYAML(param.DefaultValue); ok {
					ps.Default = def
					ps.HasDefault = true
				}
			}
			schema.Parameters = append(schema.Parameters, ps)
		}

		// Extract return types
		for _, ret := range fn.Returns {
			schema.Returns = append(schema.Returns, typeAnnotationName(ret))
		}

		schemas = append(schemas, schema)
	}

	return schemas
}

func generateSkillMD(skill *ast.SkillDecl, functions []FunctionSchema) string {
	type yamlParam struct {
		Type    string `yaml:"type"`
		Default any    `yaml:"default,omitempty"`
	}
	type yamlFunction struct {
		Name        string               `yaml:"name"`
		Description string               `yaml:"description,omitempty"`
		Parameters  map[string]yamlParam `yaml:"parameters,omitempty"`
	}
	type yamlSkill struct {
		Name        string         `yaml:"name"`
		Description string         `yaml:"description,omitempty"`
		Version     string         `yaml:"version,omitempty"`
		Functions   []yamlFunction `yaml:"functions,omitempty"`
	}

	doc := yamlSkill{
		Name:        toSnakeCase(skill.Name.Value),
		Description: skill.Description,
		Version:     skill.Version,
	}

	for _, fn := range functions {
		yfn := yamlFunction{
			Name:        fn.Name,
			Description: fn.Description,
		}
		if len(fn.Parameters) > 0 {
			yfn.Parameters = make(map[string]yamlParam, len(fn.Parameters))
			for _, p := range fn.Parameters {
				yp := yamlParam{Type: p.Type}
				if p.HasDefault {
					yp.Default = p.Default
				}
				yfn.Parameters[p.Name] = yp
			}
		}
		doc.Functions = append(doc.Functions, yfn)
	}

	out, err := yaml.Marshal(doc)
	if err != nil {
		return "---\nname: " + toSnakeCase(skill.Name.Value) + "\n---\n"
	}
	return "---\n" + string(out) + "---\n"
}

// typeToJSONSchemaType maps Kukicha/Go type annotations to JSON Schema types
func typeToJSONSchemaType(typeAnn ast.TypeAnnotation) string {
	if typeAnn == nil {
		return "object"
	}
	switch t := typeAnn.(type) {
	case *ast.PrimitiveType:
		switch t.Name {
		case "string":
			return "string"
		case "int", "int8", "int16", "int32", "int64",
			"uint", "uint8", "uint16", "uint32", "uint64":
			return "integer"
		case "float32", "float64":
			return "number"
		case "bool":
			return "boolean"
		case "byte", "rune":
			return "integer"
		}
	case *ast.ListType:
		return "array"
	case *ast.MapType:
		return "object"
	case *ast.NamedType:
		if t.Name == "error" {
			return "string"
		}
		return "object"
	}
	return "object"
}

// typeAnnotationName returns a human-readable name for a type annotation
func typeAnnotationName(typeAnn ast.TypeAnnotation) string {
	if typeAnn == nil {
		return "any"
	}
	switch t := typeAnn.(type) {
	case *ast.PrimitiveType:
		return t.Name
	case *ast.NamedType:
		return t.Name
	case *ast.ListType:
		return "list"
	case *ast.MapType:
		return "map"
	case *ast.ReferenceType:
		return "reference"
	case *ast.ChannelType:
		return "channel"
	}
	return "any"
}

// defaultValueToYAML converts a literal default value expression into a YAML value.
func defaultValueToYAML(expr ast.Expression) (any, bool) {
	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		return e.Value, true
	case *ast.FloatLiteral:
		return e.Value, true
	case *ast.StringLiteral:
		return e.Value, true
	case *ast.BooleanLiteral:
		return e.Value, true
	}
	return nil, false
}

// toSnakeCase converts PascalCase to snake_case
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				result.WriteByte('_')
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}
