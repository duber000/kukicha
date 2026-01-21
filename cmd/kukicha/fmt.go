package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/duber000/kukicha/internal/formatter"
)

func fmtCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: kukicha fmt [options] <file.kuki|directory>")
		fmt.Println("Options:")
		fmt.Println("  -w         Write result to file instead of stdout")
		fmt.Println("  --check    Check if files are formatted (exit 1 if not)")
		os.Exit(1)
	}

	var writeInPlace bool
	var checkOnly bool
	var files []string

	// Parse arguments
	for _, arg := range args {
		switch arg {
		case "-w":
			writeInPlace = true
		case "--check":
			checkOnly = true
		default:
			if strings.HasPrefix(arg, "-") {
				fmt.Fprintf(os.Stderr, "Unknown option: %s\n", arg)
				os.Exit(1)
			}
			files = append(files, arg)
		}
	}

	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "Error: no files specified")
		os.Exit(1)
	}

	// Expand directories to .kuki files
	var allFiles []string
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if info.IsDir() {
			// Find all .kuki files in directory
			err := filepath.Walk(file, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !info.IsDir() && strings.HasSuffix(path, ".kuki") {
					allFiles = append(allFiles, path)
				}
				return nil
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error walking directory: %v\n", err)
				os.Exit(1)
			}
		} else {
			allFiles = append(allFiles, file)
		}
	}

	if len(allFiles) == 0 {
		fmt.Fprintln(os.Stderr, "Error: no .kuki files found")
		os.Exit(1)
	}

	opts := formatter.DefaultOptions()
	exitCode := 0

	for _, file := range allFiles {
		if checkOnly {
			if !checkFile(file, opts) {
				exitCode = 1
			}
		} else if writeInPlace {
			if err := formatFileInPlace(file, opts); err != nil {
				fmt.Fprintf(os.Stderr, "Error formatting %s: %v\n", file, err)
				exitCode = 1
			}
		} else {
			if err := formatFileToStdout(file, opts); err != nil {
				fmt.Fprintf(os.Stderr, "Error formatting %s: %v\n", file, err)
				exitCode = 1
			}
		}
	}

	os.Exit(exitCode)
}

func checkFile(filename string, opts formatter.FormatOptions) bool {
	source, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", filename, err)
		return false
	}

	formatted, err := formatter.FormatCheck(string(source), filename, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error checking %s: %v\n", filename, err)
		return false
	}

	if !formatted {
		fmt.Printf("%s: not formatted\n", filename)
		return false
	}

	return true
}

func formatFileInPlace(filename string, opts formatter.FormatOptions) error {
	source, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	formatted, err := formatter.Format(string(source), filename, opts)
	if err != nil {
		return err
	}

	// Only write if content changed
	if string(source) != formatted {
		err = os.WriteFile(filename, []byte(formatted), 0644)
		if err != nil {
			return err
		}
		fmt.Printf("formatted %s\n", filename)
	}

	return nil
}

func formatFileToStdout(filename string, opts formatter.FormatOptions) error {
	source, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	formatted, err := formatter.Format(string(source), filename, opts)
	if err != nil {
		return err
	}

	fmt.Print(formatted)
	return nil
}
