package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/duber000/kukicha/internal/codegen"
	"github.com/duber000/kukicha/internal/parser"
	"github.com/duber000/kukicha/internal/semantic"
)

const version = "1.0.0"

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
		fmtCommand(os.Args[2:])
	case "version":
		fmt.Printf("kukicha version %s\n", version)
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
	fmt.Println("  kukicha fmt [options] <files>  Format Kukicha files")
	fmt.Println("    -w          Write result to file instead of stdout")
	fmt.Println("    --check     Check if files are formatted (exit 1 if not)")
	fmt.Println("  kukicha version             Show version information")
	fmt.Println("  kukicha help                Show this help message")
}

func buildCommand(filename string) {
	// Read source file
	source, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	// Parse
	p, err := parser.New(string(source), filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Lexer error: %v\n", err)
		os.Exit(1)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		fmt.Fprintf(os.Stderr, "Parse errors:\n")
		for _, err := range parseErrors {
			fmt.Fprintf(os.Stderr, "  %v\n", err)
		}
		os.Exit(1)
	}

	// Semantic analysis
	analyzer := semantic.New(program)
	semanticErrors := analyzer.Analyze()
	if len(semanticErrors) > 0 {
		fmt.Fprintf(os.Stderr, "Semantic errors:\n")
		for _, err := range semanticErrors {
			fmt.Fprintf(os.Stderr, "  %v\n", err)
		}
		os.Exit(1)
	}

	// Generate Go code
	gen := codegen.New(program)
	goCode, err := gen.Generate()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Code generation error: %v\n", err)
		os.Exit(1)
	}

	// Write Go file
	outputFile := strings.TrimSuffix(filename, ".kuki") + ".go"
	err = os.WriteFile(outputFile, []byte(goCode), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully compiled %s to %s\n", filename, outputFile)

	// Optionally run go build on the generated file
	cmd := exec.Command("go", "build", outputFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: go build failed: %v\n", err)
		fmt.Fprintf(os.Stderr, "Generated Go file is at: %s\n", outputFile)
		os.Exit(1)
	}

	// Get the binary name
	binaryName := strings.TrimSuffix(filepath.Base(filename), ".kuki")
	fmt.Printf("Successfully built binary: %s\n", binaryName)
}

func runCommand(filename string) {
	// Read source file
	source, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	// Parse
	p, err := parser.New(string(source), filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Lexer error: %v\n", err)
		os.Exit(1)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		fmt.Fprintf(os.Stderr, "Parse errors:\n")
		for _, err := range parseErrors {
			fmt.Fprintf(os.Stderr, "  %v\n", err)
		}
		os.Exit(1)
	}

	// Semantic analysis
	analyzer := semantic.New(program)
	semanticErrors := analyzer.Analyze()
	if len(semanticErrors) > 0 {
		fmt.Fprintf(os.Stderr, "Semantic errors:\n")
		for _, err := range semanticErrors {
			fmt.Fprintf(os.Stderr, "  %v\n", err)
		}
		os.Exit(1)
	}

	// Generate Go code
	gen := codegen.New(program)
	goCode, err := gen.Generate()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Code generation error: %v\n", err)
		os.Exit(1)
	}

	// Write temporary Go file
	tmpFile := filepath.Join(os.TempDir(), "kukicha_temp.go")
	err = os.WriteFile(tmpFile, []byte(goCode), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing temporary file: %v\n", err)
		os.Exit(1)
	}
	defer os.Remove(tmpFile)

	// Run with go run
	cmd := exec.Command("go", "run", tmpFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err = cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}

func checkCommand(filename string) {
	// Read source file
	source, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	// Parse
	p, err := parser.New(string(source), filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Lexer error: %v\n", err)
		os.Exit(1)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		fmt.Fprintf(os.Stderr, "Parse errors:\n")
		for _, err := range parseErrors {
			fmt.Fprintf(os.Stderr, "  %v\n", err)
		}
		os.Exit(1)
	}

	// Semantic analysis
	analyzer := semantic.New(program)
	semanticErrors := analyzer.Analyze()
	if len(semanticErrors) > 0 {
		fmt.Fprintf(os.Stderr, "Type errors:\n")
		for _, err := range semanticErrors {
			fmt.Fprintf(os.Stderr, "  %v\n", err)
		}
		os.Exit(1)
	}

	fmt.Printf("✓ %s type checks successfully\n", filename)
}
