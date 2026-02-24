package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNeedsStdlib_NoStdlibImport(t *testing.T) {
	goCode := `package main
import "fmt"
func main() { fmt.Println("ok") }
`
	if needsStdlib(goCode, t.TempDir()) {
		t.Fatal("expected needsStdlib to be false when no stdlib imports are present")
	}
}

func TestNeedsStdlib_RespectsProjectDirModule(t *testing.T) {
	goCode := `package main
import _ "github.com/duber000/kukicha/stdlib/json"
`

	kukichaRepoLike := t.TempDir()
	if err := os.WriteFile(filepath.Join(kukichaRepoLike, "go.mod"), []byte("module github.com/duber000/kukicha\n\ngo 1.26\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if needsStdlib(goCode, kukichaRepoLike) {
		t.Fatal("expected needsStdlib=false inside kukicha module")
	}

	userProject := t.TempDir()
	if err := os.WriteFile(filepath.Join(userProject, "go.mod"), []byte("module example.com/app\n\ngo 1.26\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if !needsStdlib(goCode, userProject) {
		t.Fatal("expected needsStdlib=true for non-kukicha projects importing stdlib")
	}
}
