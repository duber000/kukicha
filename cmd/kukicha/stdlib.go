package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	kukicha "github.com/duber000/kukicha"
	"golang.org/x/mod/modfile"
)

const stdlibDirName = ".kukicha/stdlib"

// stdlibGoMod is the go.mod content for the extracted stdlib module.
// This declares the stdlib as a standalone Go module so user projects can
// reference it via a replace directive.
const stdlibGoMod = `module github.com/duber000/kukicha/stdlib

go 1.25

require gopkg.in/yaml.v3 v3.0.1
`

// stdlibGoSum contains dependency checksums for the stdlib module.
const stdlibGoSum = `gopkg.in/check.v1 v0.0.0-20161208181325-20d25e280405 h1:yhCVgyC4o1eVCa2tZl7eS0r+SDo693bJlVdllGtEeKM=
gopkg.in/check.v1 v0.0.0-20161208181325-20d25e280405/go.mod h1:Co6ibVJAznAaIkqp8huTwlJQCZ016jof/cbN4VW5Yz0=
gopkg.in/yaml.v3 v3.0.1 h1:fxVm/GzAzEWqLHuvctI91KS9hhNmmWOoWu0XTYJS7CA=
gopkg.in/yaml.v3 v3.0.1/go.mod h1:K4uyk7z7BCEPqu6E+C64Yfv1cQ7kz7rIZviUmN+EgEM=
`

// ensureStdlib extracts the embedded stdlib to projectDir/.kukicha/stdlib/ if not present.
// Returns the absolute path to the extracted stdlib directory.
func ensureStdlib(projectDir string) (string, error) {
	stdlibPath := filepath.Join(projectDir, stdlibDirName)

	// Check if stdlib already exists by looking for the go.mod marker
	if _, err := os.Stat(filepath.Join(stdlibPath, "go.mod")); err == nil {
		return stdlibPath, nil
	}

	// Extract from embedded FS
	if err := extractStdlib(stdlibPath); err != nil {
		return "", fmt.Errorf("extracting stdlib: %w", err)
	}

	return stdlibPath, nil
}

// extractStdlib writes the embedded stdlib files to the target directory,
// plus a generated go.mod and go.sum for the standalone module.
func extractStdlib(targetDir string) error {
	// Walk embedded FS and extract all files under "stdlib/"
	err := fs.WalkDir(kukicha.StdlibFS, "stdlib", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Map embedded path "stdlib/json/json.go" -> targetDir + "/json/json.go"
		relPath, _ := filepath.Rel("stdlib", path)
		targetPath := filepath.Join(targetDir, relPath)

		if d.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}

		data, readErr := kukicha.StdlibFS.ReadFile(path)
		if readErr != nil {
			return readErr
		}

		return os.WriteFile(targetPath, data, 0644)
	})
	if err != nil {
		return err
	}

	// Write the generated go.mod for the extracted stdlib module
	if err := os.WriteFile(filepath.Join(targetDir, "go.mod"), []byte(stdlibGoMod), 0644); err != nil {
		return err
	}

	// Write the go.sum
	if err := os.WriteFile(filepath.Join(targetDir, "go.sum"), []byte(stdlibGoSum), 0644); err != nil {
		return err
	}

	return nil
}

// ensureGoMod checks the project's go.mod and adds the stdlib require/replace
// directives if they are not already present.
func ensureGoMod(projectDir, stdlibPath string) error {
	goModPath := filepath.Join(projectDir, "go.mod")

	data, err := os.ReadFile(goModPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("no go.mod found in %s; run 'go mod init <module>' or 'kukicha init' first", projectDir)
		}
		return err
	}

	mod, err := modfile.Parse(goModPath, data, nil)
	if err != nil {
		return fmt.Errorf("parsing go.mod: %w", err)
	}

	// Calculate relative path from project dir to stdlib
	relStdlib, err := filepath.Rel(projectDir, stdlibPath)
	if err != nil {
		relStdlib = stdlibPath
	}

	const stdlibModule = "github.com/duber000/kukicha/stdlib"
	const stdlibVersion = "v0.0.0"

	// Add require if missing
	if !hasRequire(mod, stdlibModule) {
		if err := mod.AddRequire(stdlibModule, stdlibVersion); err != nil {
			return fmt.Errorf("adding require: %w", err)
		}
	}

	// Add or update replace
	relPath := "./" + filepath.ToSlash(relStdlib)
	if err := mod.AddReplace(stdlibModule, "", relPath, ""); err != nil {
		return fmt.Errorf("adding replace: %w", err)
	}

	formatted, err := mod.Format()
	if err != nil {
		return fmt.Errorf("formatting go.mod: %w", err)
	}

	return os.WriteFile(goModPath, formatted, 0644)
}

// needsStdlib checks if the generated Go code imports any Kukicha stdlib packages.
// Returns false if we're inside the kukicha repo itself (stdlib already available).
func needsStdlib(goCode string) bool {
	if !strings.Contains(goCode, "github.com/duber000/kukicha/stdlib/") {
		return false
	}
	// Don't extract stdlib if we're inside the kukicha repo itself
	if isKukichaRepo() {
		return false
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", goCode, parser.ImportsOnly)
	if err != nil {
		// Fallback to substring check if parsing fails
		return true
	}

	for _, imp := range file.Imports {
		path := strings.Trim(imp.Path.Value, "\"")
		if strings.HasPrefix(path, "github.com/duber000/kukicha/stdlib/") {
			return true
		}
	}
	return false
}

func hasRequire(mod *modfile.File, path string) bool {
	for _, req := range mod.Require {
		if req.Mod.Path == path {
			return true
		}
	}
	return false
}

// isKukichaRepo checks if the current working directory is inside the kukicha repo.
// This is detected by checking if go.mod declares module github.com/duber000/kukicha.
func isKukichaRepo() bool {
	cwd, err := os.Getwd()
	if err != nil {
		return false
	}
	// Walk up looking for go.mod
	for d := cwd; d != filepath.Dir(d); d = filepath.Dir(d) {
		goModPath := filepath.Join(d, "go.mod")
		data, err := os.ReadFile(goModPath)
		if err != nil {
			continue
		}
		// Check if this is the kukicha repo's go.mod
		content := string(data)
		if strings.Contains(content, "module github.com/duber000/kukicha\n") ||
			strings.Contains(content, "module github.com/duber000/kukicha\r\n") {
			return true
		}
		// Found a go.mod but it's not the kukicha repo
		return false
	}
	return false
}

// findProjectDir walks up from the given filename to find the directory
// containing a go.mod file. If none is found, returns the directory of the file.
func findProjectDir(filename string) string {
	dir := filepath.Dir(filename)
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return dir
	}

	// Walk up looking for go.mod
	for d := absDir; d != filepath.Dir(d); d = filepath.Dir(d) {
		if _, err := os.Stat(filepath.Join(d, "go.mod")); err == nil {
			return d
		}
	}

	return absDir
}
