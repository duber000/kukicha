package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindProjectRoot_WithGoMod(t *testing.T) {
	dir := t.TempDir()

	// Create nested structure with go.mod at root
	subDir := filepath.Join(dir, "cmd", "app")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n\ngo 1.26.1\n"), 0644); err != nil {
		t.Fatal(err)
	}

	root, err := findProjectRoot(subDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	absDir, _ := filepath.Abs(dir)
	if root != absDir {
		t.Errorf("expected %s, got %s", absDir, root)
	}
}

func TestFindProjectRoot_NoGoMod(t *testing.T) {
	dir := t.TempDir()
	subDir := filepath.Join(dir, "src")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	_, err := findProjectRoot(subDir)
	if err == nil {
		t.Fatal("expected error when no go.mod exists")
	}
}

func TestFindProjectRoot_DirectRoot(t *testing.T) {
	dir := t.TempDir()

	// go.mod in the given dir itself
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n\ngo 1.26.1\n"), 0644); err != nil {
		t.Fatal(err)
	}

	root, err := findProjectRoot(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	absDir, _ := filepath.Abs(dir)
	if root != absDir {
		t.Errorf("expected %s, got %s", absDir, root)
	}
}

func TestRunAudit_NoGoMod(t *testing.T) {
	dir := t.TempDir()

	code := runAudit(AuditOptions{Dir: dir})
	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
}
