package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func initCommand() {
	projectDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting working directory: %v\n", err)
		os.Exit(1)
	}

	// Check if go.mod exists
	goModPath := filepath.Join(projectDir, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		fmt.Fprintln(os.Stderr, "No go.mod found. Run 'go mod init <module-name>' first.")
		os.Exit(1)
	}

	// Extract stdlib
	stdlibPath, err := ensureStdlib(projectDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error extracting stdlib: %v\n", err)
		os.Exit(1)
	}

	// Update go.mod with require and replace directives
	if err := ensureGoMod(projectDir, stdlibPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error updating go.mod: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Kukicha project initialized.")
	fmt.Printf("  Stdlib extracted to: %s\n", stdlibPath)
	fmt.Println("  go.mod updated with replace directive.")
	fmt.Println()
	fmt.Println("Add .kukicha/ to your .gitignore:")
	fmt.Println("  echo '.kukicha/' >> .gitignore")
}
