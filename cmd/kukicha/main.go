package main

import (
	"bytes"
	"flag"
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
	args := os.Args[2:]

	switch command {
	case "build":
		buildFlags := flag.NewFlagSet("build", flag.ContinueOnError)
		buildFlags.SetOutput(os.Stderr)
		target := buildFlags.String("target", "", "Compile target")
		if err := buildFlags.Parse(args); err != nil {
			fmt.Fprintln(os.Stderr, "Usage: kukicha build [--target <target>] <file.kuki>")
			os.Exit(1)
		}
		buildArgs := buildFlags.Args()
		if len(buildArgs) < 1 {
			fmt.Fprintln(os.Stderr, "Usage: kukicha build [--target <target>] <file.kuki>")
			os.Exit(1)
		}
		buildCommand(buildArgs[0], *target)
	case "run":
		runFlags := flag.NewFlagSet("run", flag.ContinueOnError)
		runFlags.SetOutput(os.Stderr)
		target := runFlags.String("target", "", "Run target")
		if err := runFlags.Parse(args); err != nil {
			fmt.Fprintln(os.Stderr, "Usage: kukicha run [--target <target>] <file.kuki> [args...]")
			os.Exit(1)
		}
		runArgs := runFlags.Args()
		if len(runArgs) < 1 {
			fmt.Fprintln(os.Stderr, "Usage: kukicha run [--target <target>] <file.kuki> [args...]")
			os.Exit(1)
		}
		runCommand(runArgs[0], *target, runArgs[1:])
	case "check":
		checkFlags := flag.NewFlagSet("check", flag.ContinueOnError)
		checkFlags.SetOutput(os.Stderr)
		strictOnerr := checkFlags.Bool("strict-onerr", false, "Treat onerr lint warnings as errors")
		if err := checkFlags.Parse(args); err != nil {
			fmt.Fprintln(os.Stderr, "Usage: kukicha check [--strict-onerr] <file.kuki>")
			os.Exit(1)
		}
		checkArgs := checkFlags.Args()
		if len(checkArgs) < 1 {
			fmt.Fprintln(os.Stderr, "Usage: kukicha check [--strict-onerr] <file.kuki>")
			os.Exit(1)
		}
		checkCommand(checkArgs[0], *strictOnerr)
	case "fmt":
		if len(args) < 1 {
			fmt.Fprintln(os.Stderr, "Usage: kukicha fmt [options] <files>")
			os.Exit(1)
		}
		fmtCommand(args)
	case "pack":
		packFlags := flag.NewFlagSet("pack", flag.ContinueOnError)
		packFlags.SetOutput(os.Stderr)
		outputDir := packFlags.String("output", "", "Output directory")
		if err := packFlags.Parse(args); err != nil {
			fmt.Fprintln(os.Stderr, "Usage: kukicha pack [--output <dir>] <skill.kuki>")
			os.Exit(1)
		}
		packArgs := packFlags.Args()
		if len(packArgs) < 1 {
			fmt.Fprintln(os.Stderr, "Usage: kukicha pack [--output <dir>] <skill.kuki>")
			os.Exit(1)
		}
		packCommand(packArgs[0], *outputDir)
	case "init":
		initCommand(args)
	case "version":
		fmt.Printf("kukicha version %s\n", version.Version)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Kukicha - A transpiler that compiles Kukicha to Go")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  kukicha build [--target t] <file.kuki> Compile Kukicha file to Go")
	fmt.Fprintln(os.Stderr, "  kukicha run [--target t] <file.kuki>   Transpile and execute Kukicha file")
	fmt.Fprintln(os.Stderr, "  kukicha check <file.kuki>   Type check Kukicha file")
	fmt.Fprintln(os.Stderr, "  kukicha fmt [options] <files>  Fix indentation and normalize style")
	fmt.Fprintln(os.Stderr, "    -w          Write result to file instead of stdout")
	fmt.Fprintln(os.Stderr, "    --check     Check if files are formatted (exit 1 if not)")
	fmt.Fprintln(os.Stderr, "  kukicha pack [--output dir] <skill.kuki>  Package skill for distribution")
	fmt.Fprintln(os.Stderr, "  kukicha init [module-name]  Initialize project (go mod init + extract stdlib)")
	fmt.Fprintln(os.Stderr, "  kukicha version             Show version information")
	fmt.Fprintln(os.Stderr, "  kukicha help                Show this help message")
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

	analyzer := semantic.NewWithFile(program, filename)
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

func detectTarget(source string) string {
	lines := strings.Split(source, "\n")
	for i, line := range lines {
		if i > 10 { // Only look at first 10 lines
			break
		}
		line = strings.TrimSpace(line)
		if after, ok := strings.CutPrefix(line, "# target:"); ok {
			return strings.TrimSpace(after)
		}
	}
	return ""
}

func detectTargetFromFile(filename string) (string, error) {
	source, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return detectTarget(string(source)), nil
}

// rewriteGoErrors replaces references to the generated .go file path in stderr
// output with the original .kuki source path. This cleans up any residual file
// references that aren't covered by //line directives (e.g., temp file paths).
func rewriteGoErrors(stderr []byte, goFile, kukiFile string) []byte {
	result := strings.ReplaceAll(string(stderr), goFile, kukiFile)
	return []byte(result)
}

func buildCommand(filename string, targetFlag string) {
	absFile, err := filepath.Abs(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving file path: %v\n", err)
		os.Exit(1)
	}
	projectDir := findProjectDir(absFile)

	program, returnCounts, err := loadAndAnalyze(absFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Detect target from source if not provided by flag
	if targetFlag != "" {
		program.Target = targetFlag
	} else {
		t, readErr := detectTargetFromFile(absFile)
		if readErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not read %s for target detection: %v\n", absFile, readErr)
		}
		if t != "" {
			program.Target = t
		}
	}

	// Generate Go code
	gen := codegen.New(program)
	gen.SetSourceFile(absFile) // Enable special transpilation detection
	gen.SetExprReturnCounts(returnCounts)
	if program.Target == "mcp" {
		gen.SetMCPTarget(true)
	}
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

	// Run go build on the generated file. Use -mod=mod so go.sum is updated
	// automatically when stdlib transitive dependencies are not yet listed.
	cmd := exec.Command("go", "build", "-mod=mod", outputFile)
	cmd.Dir = projectDir
	cmd.Env = os.Environ()
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

func runCommand(filename string, targetFlag string, scriptArgs []string) {
	absFile, err := filepath.Abs(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving file path: %v\n", err)
		os.Exit(1)
	}
	projectDir := findProjectDir(absFile)

	program, returnCounts, err := loadAndAnalyze(absFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	// Detect target from source if not provided by flag
	if targetFlag != "" {
		program.Target = targetFlag
	} else {
		t, readErr := detectTargetFromFile(absFile)
		if readErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not read %s for target detection: %v\n", absFile, readErr)
		}
		if t != "" {
			program.Target = t
		}
	}

	// Generate Go code
	gen := codegen.New(program)
	gen.SetSourceFile(absFile) // Enable special transpilation detection
	gen.SetExprReturnCounts(returnCounts)
	if program.Target == "mcp" {
		gen.SetMCPTarget(true)
	}
	goCode, err := gen.Generate()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Code generation error: %v\n", err)
		os.Exit(1)
	}

	// If stdlib is needed, extract it and ensure go.mod is configured.
	// Keep temp source in project context so local replace directives resolve.
	var tmp *os.File
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
		tmp, err = os.CreateTemp(filepath.Join(projectDir, ".kukicha"), "run-*.go")
	} else {
		tmp, err = os.CreateTemp("", "kukicha-run-*.go")
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating temporary file: %v\n", err)
		os.Exit(1)
	}
	tmpFile := tmp.Name()
	defer os.Remove(tmpFile)

	formatted, fmtErr := format.Source([]byte(goCode))
	if fmtErr != nil {
		formatted = []byte(goCode)
	}

	if _, err := tmp.Write(formatted); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing temporary file: %v\n", err)
		os.Exit(1)
	}
	if err := tmp.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "Error closing temporary file: %v\n", err)
		os.Exit(1)
	}

	// Run with go run. Use -mod=mod so Go updates go.sum automatically when
	// stdlib transitive dependencies (e.g. gopkg.in/yaml.v3) are not yet listed.
	goArgs := append([]string{"run", "-mod=mod", tmpFile}, scriptArgs...)
	cmd := exec.Command("go", goArgs...)
	cmd.Dir = projectDir
	cmd.Env = os.Environ()
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

func checkCommand(filename string, strictOnerr bool) {
	source, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Errorf("Error reading file: %v", err))
		os.Exit(1)
	}

	p, err := parser.New(string(source), filename)
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Errorf("Lexer error: %v", err))
		os.Exit(1)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		var msgs []string
		for _, e := range parseErrors {
			msgs = append(msgs, fmt.Sprintf("  %v", e))
		}
		fmt.Fprintln(os.Stderr, fmt.Errorf("Parse errors:\n%s", strings.Join(msgs, "\n")))
		os.Exit(1)
	}

	analyzer := semantic.NewWithFile(program, filename)
	semanticErrors := analyzer.Analyze()
	if len(semanticErrors) > 0 {
		var msgs []string
		for _, e := range semanticErrors {
			msgs = append(msgs, fmt.Sprintf("  %v", e))
		}
		fmt.Fprintln(os.Stderr, fmt.Errorf("Semantic errors:\n%s", strings.Join(msgs, "\n")))
		os.Exit(1)
	}

	warnings := analyzer.Warnings()
	for _, w := range warnings {
		fmt.Fprintf(os.Stderr, "warning: %v\n", w)
	}
	if strictOnerr && len(warnings) > 0 {
		fmt.Fprintln(os.Stderr, "onerr warnings promoted to errors (--strict-onerr)")
		os.Exit(1)
	}

	fmt.Printf("✓ %s type checks successfully\n", filename)
}
