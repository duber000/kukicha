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
	source, _ := os.ReadFile(absFile)
	if t := detectTarget(string(source)); t != "" {
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
	if needsStdlib(goCode) {
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
	cmd.Env = append(os.Environ(), "GOEXPERIMENT=jsonv2")
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
	Name    string
	Type    string
	Default string
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
				ps.Default = defaultValueToString(param.DefaultValue)
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
	var b strings.Builder

	b.WriteString("---\n")
	b.WriteString(fmt.Sprintf("name: %s\n", toSnakeCase(skill.Name.Value)))
	if skill.Description != "" {
		b.WriteString(fmt.Sprintf("description: %s\n", skill.Description))
	}
	if skill.Version != "" {
		b.WriteString(fmt.Sprintf("version: \"%s\"\n", skill.Version))
	}

	if len(functions) > 0 {
		b.WriteString("functions:\n")
		for _, fn := range functions {
			b.WriteString(fmt.Sprintf("  - name: %s\n", fn.Name))
			if fn.Description != "" {
				b.WriteString(fmt.Sprintf("    description: %s\n", fn.Description))
			}
			if len(fn.Parameters) > 0 {
				b.WriteString("    parameters:\n")
				for _, p := range fn.Parameters {
					if p.Default != "" {
						b.WriteString(fmt.Sprintf("      %s: { type: %s, default: %s }\n", p.Name, p.Type, p.Default))
					} else {
						b.WriteString(fmt.Sprintf("      %s: { type: %s }\n", p.Name, p.Type))
					}
				}
			}
		}
	}

	b.WriteString("---\n")
	return b.String()
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

// defaultValueToString returns a string representation of a default value expression
func defaultValueToString(expr ast.Expression) string {
	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		return fmt.Sprintf("%d", e.Value)
	case *ast.FloatLiteral:
		return fmt.Sprintf("%g", e.Value)
	case *ast.StringLiteral:
		return fmt.Sprintf("%q", e.Value)
	case *ast.BooleanLiteral:
		if e.Value {
			return "true"
		}
		return "false"
	}
	return ""
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
