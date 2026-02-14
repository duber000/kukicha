package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/duber000/kukicha/internal/ast"
	"github.com/duber000/kukicha/internal/codegen"
	"github.com/duber000/kukicha/internal/parser"
	"github.com/duber000/kukicha/internal/semantic"
	"github.com/duber000/kukicha/internal/version"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "build":
		if len(os.Args) < 3 {
			fmt.Println("Usage: kukicha build <file.kuki>")
			os.Exit(1)
		}
		buildCommand(os.Args[2])
	case "run":
		if len(os.Args) < 3 {
			fmt.Println("Usage: kukicha run <file.kuki>")
			os.Exit(1)
		}
		runCommand(os.Args[2])
	case "check":
		if len(os.Args) < 3 {
			fmt.Println("Usage: kukicha check <file.kuki>")
			os.Exit(1)
		}
		checkCommand(os.Args[2])
	case "fmt":
		if len(os.Args) < 3 {
			fmt.Println("Usage: kukicha fmt [options] <files>")
			os.Exit(1)
		}
		fmtCommand(os.Args[2:])
	case "init":
		initCommand()
	case "version":
		fmt.Printf("kukicha version %s\n", version.Version)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Kukicha - A transpiler that compiles Kukicha to Go")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  kukicha build <file.kuki>   Compile Kukicha file to Go")
	fmt.Println("  kukicha run <file.kuki>     Transpile and execute Kukicha file")
	fmt.Println("  kukicha check <file.kuki>   Type check Kukicha file")
	fmt.Println("  kukicha fmt [options] <files>  Fix indentation and normalize style")
	fmt.Println("    -w          Write result to file instead of stdout")
	fmt.Println("    --check     Check if files are formatted (exit 1 if not)")
	fmt.Println("  kukicha init                Extract stdlib and configure go.mod")
	fmt.Println("  kukicha version             Show version information")
	fmt.Println("  kukicha help                Show this help message")
}

func loadAndAnalyze(filename string) (*ast.Program, map[ast.Expression]int, error) {
	source, err := os.ReadFile(filename)
	if err != nil {
		return nil, nil, fmt.Errorf("Error reading file: %v", err)
	}

	p, err := parser.New(string(source), filename)
	if err != nil {
		return nil, nil, fmt.Errorf("Lexer error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		var msgs []string
		for _, e := range parseErrors {
			msgs = append(msgs, fmt.Sprintf("  %v", e))
		}
		return nil, nil, fmt.Errorf("Parse errors:\n%s", strings.Join(msgs, "\n"))
	}

	analyzer := semantic.New(program)
	semanticErrors := analyzer.Analyze()
	if len(semanticErrors) > 0 {
		var msgs []string
		for _, e := range semanticErrors {
			msgs = append(msgs, fmt.Sprintf("  %v", e))
		}
		return nil, nil, fmt.Errorf("Semantic errors:\n%s", strings.Join(msgs, "\n"))
	}

	return program, analyzer.ReturnCounts(), nil
}

// rewriteGoErrors replaces references to the generated .go file path in stderr
// output with the original .kuki source path. This cleans up any residual file
// references that aren't covered by //line directives (e.g., temp file paths).
func rewriteGoErrors(stderr []byte, goFile, kukiFile string) []byte {
	result := strings.ReplaceAll(string(stderr), goFile, kukiFile)
	return []byte(result)
}

func buildCommand(filename string) {
	absFile, err := filepath.Abs(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving file path: %v\n", err)
		os.Exit(1)
	}

	program, returnCounts, err := loadAndAnalyze(absFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Generate Go code
	gen := codegen.New(program)
	gen.SetSourceFile(absFile) // Enable special transpilation detection
	gen.SetExprReturnCounts(returnCounts)
	goCode, err := gen.Generate()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Code generation error: %v\n", err)
		os.Exit(1)
	}

	// Format with gofmt
	formatted, err := format.Source([]byte(goCode))
	if err != nil {
		// If formatting fails, use unformatted code (shouldn't happen)
		formatted = []byte(goCode)
	}

	// Write Go file
	outputFile := strings.TrimSuffix(absFile, ".kuki") + ".go"
	err = os.WriteFile(outputFile, formatted, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully compiled %s to %s\n", absFile, outputFile)

	// If the generated code imports Kukicha stdlib, extract it and configure go.mod
	if needsStdlib(goCode) {
		projectDir := findProjectDir(absFile)
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

	// Run go build on the generated file
	projectDir := findProjectDir(absFile)
	cmd := exec.Command("go", "build", outputFile)
	cmd.Dir = projectDir
	cmd.Env = append(os.Environ(), "GOEXPERIMENT=jsonv2,greenteagc")
	cmd.Stdout = os.Stdout
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf
	err = cmd.Run()
	if stderrBuf.Len() > 0 {
		os.Stderr.Write(rewriteGoErrors(stderrBuf.Bytes(), outputFile, absFile))
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: go build failed: %v\n", err)
		os.Exit(1)
	}

	// Get the binary name
	binaryName := strings.TrimSuffix(filepath.Base(absFile), ".kuki")
	fmt.Printf("Successfully built binary: %s\n", binaryName)
}

func runCommand(filename string) {
	absFile, err := filepath.Abs(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving file path: %v\n", err)
		os.Exit(1)
	}

	program, returnCounts, err := loadAndAnalyze(absFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Generate Go code
	gen := codegen.New(program)
	gen.SetSourceFile(absFile) // Enable special transpilation detection
	gen.SetExprReturnCounts(returnCounts)
	goCode, err := gen.Generate()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Code generation error: %v\n", err)
		os.Exit(1)
	}

	// If stdlib is needed, extract it and ensure go.mod is configured
	var tmpFile string
	if needsStdlib(goCode) {
		projectDir := findProjectDir(absFile)
		stdlibPath, stdlibErr := ensureStdlib(projectDir)
		if stdlibErr != nil {
			fmt.Fprintf(os.Stderr, "Error extracting stdlib: %v\n", stdlibErr)
			os.Exit(1)
		}
		if modErr := ensureGoMod(projectDir, stdlibPath); modErr != nil {
			fmt.Fprintf(os.Stderr, "Error updating go.mod: %v\n", modErr)
			os.Exit(1)
		}
		// Write temp file into .kukicha/ (already created by ensureStdlib)
		// so go.mod resolves correctly. Must not use a dotfile name at the
		// top of projectDir: the go tool ignores files starting with ".".
		tmpFile = filepath.Join(projectDir, ".kukicha", "temp.go")
	} else {
		tmpFile = filepath.Join(os.TempDir(), "kukicha_temp.go")
	}

	formatted, fmtErr := format.Source([]byte(goCode))
	if fmtErr != nil {
		formatted = []byte(goCode)
	}

	err = os.WriteFile(tmpFile, formatted, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing temporary file: %v\n", err)
		os.Exit(1)
	}
	defer os.Remove(tmpFile)

	// Run with go run
	cmd := exec.Command("go", "run", tmpFile)
	cmd.Dir = findProjectDir(absFile)
	cmd.Env = append(os.Environ(), "GOEXPERIMENT=jsonv2,greenteagc")
	cmd.Stdout = os.Stdout
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf
	cmd.Stdin = os.Stdin
	err = cmd.Run()
	if stderrBuf.Len() > 0 {
		os.Stderr.Write(rewriteGoErrors(stderrBuf.Bytes(), tmpFile, absFile))
	}
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}

func checkCommand(filename string) {
	_, _, err := loadAndAnalyze(filename)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("✓ %s type checks successfully\n", filename)
}
