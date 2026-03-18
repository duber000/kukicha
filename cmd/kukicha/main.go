package main

//go:generate sh -c "cd ../.. && go run ./cmd/genstdlibregistry"
//go:generate sh -c "cd ../.. && go run ./cmd/gengostdlib"

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/format"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/duber000/kukicha/internal/ast"
	"github.com/duber000/kukicha/internal/codegen"
	"github.com/duber000/kukicha/internal/diagnostic"
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
		skipBuild := buildFlags.Bool("skip-build", false, "Skip go build step (for test files)")
		ifChanged := buildFlags.Bool("if-changed", false, "Skip writing output if Go body (excluding generated header) is unchanged")
		vulncheck := buildFlags.Bool("vulncheck", false, "Run govulncheck after successful build")
		if err := buildFlags.Parse(args); err != nil {
			fmt.Fprintln(os.Stderr, "Usage: kukicha build [--target <target>] [--skip-build] [--if-changed] [--vulncheck] <file.kuki>")
			os.Exit(1)
		}
		buildArgs := buildFlags.Args()
		if len(buildArgs) < 1 {
			fmt.Fprintln(os.Stderr, "Usage: kukicha build [--target <target>] [--skip-build] [--if-changed] [--vulncheck] <file.kuki>")
			os.Exit(1)
		}
		buildCommand(buildArgs[0], *target, *skipBuild, *ifChanged, *vulncheck)
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
		jsonOutput := checkFlags.Bool("json", false, "Output diagnostics as JSON array")
		if err := checkFlags.Parse(args); err != nil {
			fmt.Fprintln(os.Stderr, "Usage: kukicha check [--json] [--strict-onerr] <file.kuki> [file2.kuki ...]")
			os.Exit(1)
		}
		checkArgs := checkFlags.Args()
		if len(checkArgs) < 1 {
			fmt.Fprintln(os.Stderr, "Usage: kukicha check [--json] [--strict-onerr] <file.kuki> [file2.kuki ...]")
			os.Exit(1)
		}
		checkCommand(checkArgs, *strictOnerr, *jsonOutput)
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
	case "audit":
		auditFlags := flag.NewFlagSet("audit", flag.ContinueOnError)
		auditFlags.SetOutput(os.Stderr)
		jsonFlag := auditFlags.Bool("json", false, "Output in JSON format")
		warnOnly := auditFlags.Bool("warn-only", false, "Exit 0 even if vulnerabilities are found")
		if err := auditFlags.Parse(args); err != nil {
			fmt.Fprintln(os.Stderr, "Usage: kukicha audit [--json] [--warn-only] [dir]")
			os.Exit(1)
		}
		auditCommand(auditFlags.Args(), *jsonFlag, *warnOnly)
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
	fmt.Fprintln(os.Stderr, "  kukicha build [--target t] [--vulncheck] <file.kuki>  Compile Kukicha file to Go")
	fmt.Fprintln(os.Stderr, "  kukicha run [--target t] <file.kuki>   Transpile and execute Kukicha file")
	fmt.Fprintln(os.Stderr, "  kukicha check [--json] [--strict-onerr] <file.kuki> [...]  Type check Kukicha file(s)")
	fmt.Fprintln(os.Stderr, "  kukicha audit [--json] [--warn-only] [dir]  Check dependencies for vulnerabilities")
	fmt.Fprintln(os.Stderr, "  kukicha fmt [options] <files>  Fix indentation and normalize style")
	fmt.Fprintln(os.Stderr, "    -w          Write result to file instead of stdout")
	fmt.Fprintln(os.Stderr, "    --check     Check if files are formatted (exit 1 if not)")
	fmt.Fprintln(os.Stderr, "  kukicha pack [--output dir] <skill.kuki>  Package skill for distribution")
	fmt.Fprintln(os.Stderr, "  kukicha init [module-name]  Initialize project (go mod init + extract stdlib)")
	fmt.Fprintln(os.Stderr, "  kukicha version             Show version information")
	fmt.Fprintln(os.Stderr, "  kukicha help                Show this help message")
}

func loadAndAnalyze(filename string) (*ast.Program, map[ast.Expression]int, map[ast.Expression]*semantic.TypeInfo, error) {
	source, err := os.ReadFile(filename)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Error reading file: %v", err)
	}

	p, err := parser.New(string(source), filename)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Lexer error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		var msgs []string
		for _, e := range parseErrors {
			msgs = append(msgs, fmt.Sprintf("  %v", e))
		}
		return nil, nil, nil, fmt.Errorf("Parse errors:\n%s", strings.Join(msgs, "\n"))
	}

	analyzer := semantic.NewWithFile(program, filename)
	semanticErrors := analyzer.Analyze()
	if len(semanticErrors) > 0 {
		var msgs []string
		for _, e := range semanticErrors {
			msgs = append(msgs, fmt.Sprintf("  %v", e))
		}
		return nil, nil, nil, fmt.Errorf("Semantic errors:\n%s", strings.Join(msgs, "\n"))
	}

	return program, analyzer.ReturnCounts(), analyzer.ExprTypes(), nil
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
// output with the original .kuki source path. While //line directives handle
// most source mapping, some Go compiler errors reference the physical file path
// directly (e.g., package-level errors, import failures, syntax errors in
// generated code). This function catches those residual references.
//
// The replacement is intentionally simple (strings.ReplaceAll) because Go error
// formats vary across versions and tools (go build, go vet, etc.). A regex
// approach would need to track multiple formats and could miss edge cases.
// The broad replacement is safe because goFile is a unique temp/output path
// that won't appear in error messages for any other reason.
func rewriteGoErrors(stderr []byte, goFile, kukiFile string) []byte {
	if len(stderr) == 0 {
		return stderr
	}
	result := strings.ReplaceAll(string(stderr), goFile, kukiFile)
	return []byte(result)
}

// stripFirstLine removes the first line (including its newline) from b.
// Used to compare generated Go files while ignoring the version header.
func stripFirstLine(b []byte) []byte {
	if i := bytes.IndexByte(b, '\n'); i >= 0 {
		return b[i+1:]
	}
	return b
}

func buildCommand(filename string, targetFlag string, skipBuild bool, ifChanged bool, vulncheck bool) {
	absFile, err := filepath.Abs(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving file path: %v\n", err)
		os.Exit(1)
	}
	projectDir := findProjectDir(absFile)

	program, returnCounts, exprTypes, err := loadAndAnalyze(absFile)
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
	gen.SetExprTypes(exprTypes)
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

	if ifChanged {
		if existing, readErr := os.ReadFile(outputFile); readErr == nil {
			if bytes.Equal(stripFirstLine(existing), stripFirstLine(formatted)) {
				return // body unchanged — preserve old version comment, skip write+build
			}
		}
	}

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

	// Determine the output binary name. When cross-compiling for Windows
	// (GOOS=windows), append .exe so the binary is recognised as executable.
	binaryName := strings.TrimSuffix(filepath.Base(absFile), ".kuki")
	targetOS := os.Getenv("GOOS")
	if targetOS == "" {
		targetOS = runtime.GOOS
	}
	if targetOS == "windows" {
		binaryName += ".exe"
	}
	binaryPath := filepath.Join(projectDir, binaryName)

	// Run go build on the generated file. Use -mod=mod so go.sum is updated
	// automatically when stdlib transitive dependencies are not yet listed.
	if !skipBuild {
		cmd := exec.Command("go", "build", "-mod=mod", "-o", binaryPath, outputFile)
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

		fmt.Printf("Successfully built binary: %s\n", binaryName)
	}

	if vulncheck {
		code := runAudit(AuditOptions{Dir: projectDir})
		if code != 0 {
			os.Exit(code)
		}
	}
}

func runCommand(filename string, targetFlag string, scriptArgs []string) {
	absFile, err := filepath.Abs(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving file path: %v\n", err)
		os.Exit(1)
	}
	projectDir := findProjectDir(absFile)

	program, returnCounts, exprTypes, err := loadAndAnalyze(absFile)
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
	gen.SetExprTypes(exprTypes)
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

// checkFileForDiags runs the check pipeline on a single file and returns structured
// diagnostics plus a flag indicating whether errors were found.
func checkFileForDiags(filename string) (diags []diagnostic.Diagnostic, hasErrors bool, hasWarnings bool) {
	source, err := os.ReadFile(filename)
	if err != nil {
		diags = append(diags, diagnostic.Diagnostic{
			File:     filename,
			Line:     1,
			Col:      1,
			Severity: "error",
			Message:  fmt.Sprintf("error reading file: %v", err),
		})
		return diags, true, false
	}

	p, err := parser.New(string(source), filename)
	if err != nil {
		diags = append(diags, diagnostic.Diagnostic{
			File:     filename,
			Line:     1,
			Col:      1,
			Severity: "error",
			Message:  fmt.Sprintf("lexer error: %v", err),
		})
		return diags, true, false
	}

	_, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		for _, e := range parseErrors {
			diags = append(diags, diagnostic.Diagnostic{
				File:     filename,
				Line:     1,
				Col:      1,
				Severity: "error",
				Message:  e.Error(),
			})
		}
		return diags, true, false
	}

	// Re-parse for semantic analysis (parser is consumed after Parse)
	p2, _ := parser.New(string(source), filename)
	program, _ := p2.Parse()

	analyzer := semantic.NewWithFile(program, filename)
	semanticErrors := analyzer.Analyze()

	// Collect structured diagnostics from the analyzer (security checks produce these)
	diags = append(diags, analyzer.Diagnostics()...)

	// Also collect any plain errors/warnings not already in diagnostics.
	// Security errors are already in diagnostics via errorDiag; plain errors are not.
	diagSet := make(map[string]bool)
	for _, d := range diags {
		diagSet[d.Error()] = true
	}
	for _, e := range semanticErrors {
		if !diagSet[e.Error()] {
			diags = append(diags, diagnostic.Diagnostic{
				File:     filename,
				Line:     1,
				Col:      1,
				Severity: "error",
				Message:  e.Error(),
			})
		}
	}
	for _, w := range analyzer.Warnings() {
		if !diagSet[w.Error()] {
			diags = append(diags, diagnostic.Diagnostic{
				File:     filename,
				Line:     1,
				Col:      1,
				Severity: "warning",
				Message:  w.Error(),
			})
		}
	}

	return diags, len(semanticErrors) > 0, len(analyzer.Warnings()) > 0
}

func checkCommand(files []string, strictOnerr bool, jsonOutput bool) {
	var allDiags []diagnostic.Diagnostic
	hasErrors := false
	hasWarnings := false

	for _, filename := range files {
		diags, fileHasErrors, fileHasWarnings := checkFileForDiags(filename)
		allDiags = append(allDiags, diags...)
		if fileHasErrors {
			hasErrors = true
		}
		if fileHasWarnings {
			hasWarnings = true
		}
	}

	if jsonOutput {
		// JSON mode: output structured diagnostics to stdout
		if allDiags == nil {
			allDiags = []diagnostic.Diagnostic{} // emit [] not null
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(allDiags); err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Text mode: existing output format
		for _, d := range allDiags {
			if d.Severity == "warning" {
				fmt.Fprintf(os.Stderr, "warning: %s\n", d.Error())
			} else {
				fmt.Fprintf(os.Stderr, "  %s\n", d.Error())
			}
		}
		if !hasErrors && !(strictOnerr && hasWarnings) {
			for _, filename := range files {
				fmt.Printf("✓ %s type checks successfully\n", filename)
			}
		}
	}

	if hasErrors || (strictOnerr && hasWarnings) {
		os.Exit(1)
	}
}
