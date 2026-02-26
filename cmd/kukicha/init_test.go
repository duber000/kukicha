package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/duber000/kukicha/internal/version"
)

func TestEnsureStdlib_ExtractsFiles(t *testing.T) {
	dir := t.TempDir()

	stdlibPath, err := ensureStdlib(dir)
	if err != nil {
		t.Fatalf("ensureStdlib error: %v", err)
	}

	// Check that the stdlib path is correct
	expected := filepath.Join(dir, stdlibDirName)
	if stdlibPath != expected {
		t.Errorf("expected stdlib path %s, got %s", expected, stdlibPath)
	}

	// Check go.mod was written
	goModPath := filepath.Join(stdlibPath, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		t.Error("expected go.mod in extracted stdlib")
	}

	// Check go.sum was written
	goSumPath := filepath.Join(stdlibPath, "go.sum")
	if _, err := os.Stat(goSumPath); os.IsNotExist(err) {
		t.Error("expected go.sum in extracted stdlib")
	}

	// Check version stamp was written
	stampPath := filepath.Join(stdlibPath, stdlibVersionFile)
	stamp, err := os.ReadFile(stampPath)
	if err != nil {
		t.Fatalf("reading version stamp: %v", err)
	}
	if strings.TrimSpace(string(stamp)) != version.Version {
		t.Errorf("expected version stamp %q, got %q", version.Version, string(stamp))
	}
}

func TestEnsureStdlib_SkipsIfVersionMatches(t *testing.T) {
	dir := t.TempDir()

	// First extraction
	stdlibPath, err := ensureStdlib(dir)
	if err != nil {
		t.Fatalf("first ensureStdlib: %v", err)
	}

	// Create a marker file to detect re-extraction
	marker := filepath.Join(stdlibPath, "test_marker.txt")
	if err := os.WriteFile(marker, []byte("exists"), 0644); err != nil {
		t.Fatal(err)
	}

	// Second call should skip extraction (version matches)
	_, err = ensureStdlib(dir)
	if err != nil {
		t.Fatalf("second ensureStdlib: %v", err)
	}

	// Marker should still exist (not re-extracted)
	if _, err := os.Stat(marker); os.IsNotExist(err) {
		t.Error("expected marker file to survive; stdlib was re-extracted when version matched")
	}
}

func TestEnsureStdlib_ReExtractsOnVersionMismatch(t *testing.T) {
	dir := t.TempDir()

	// First extraction
	stdlibPath, err := ensureStdlib(dir)
	if err != nil {
		t.Fatalf("first ensureStdlib: %v", err)
	}

	// Corrupt the version stamp to simulate an upgrade
	stampPath := filepath.Join(stdlibPath, stdlibVersionFile)
	if err := os.WriteFile(stampPath, []byte("0.0.0-old"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a marker file
	marker := filepath.Join(stdlibPath, "test_marker.txt")
	if err := os.WriteFile(marker, []byte("exists"), 0644); err != nil {
		t.Fatal(err)
	}

	// Re-run should re-extract (version mismatch)
	_, err = ensureStdlib(dir)
	if err != nil {
		t.Fatalf("second ensureStdlib: %v", err)
	}

	// Marker should be gone (directory was removed and re-extracted)
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Error("expected marker file to be removed on version mismatch re-extraction")
	}

	// Version stamp should now match
	stamp, err := os.ReadFile(stampPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(stamp)) != version.Version {
		t.Errorf("expected updated version stamp %q, got %q", version.Version, string(stamp))
	}
}

func TestEnsureGoMod_AddsRequireAndReplace(t *testing.T) {
	dir := t.TempDir()

	// Create a minimal go.mod
	goModContent := "module example.com/app\n\ngo 1.26\n"
	goModPath := filepath.Join(dir, "go.mod")
	if err := os.WriteFile(goModPath, []byte(goModContent), 0644); err != nil {
		t.Fatal(err)
	}

	stdlibPath := filepath.Join(dir, ".kukicha", "stdlib")
	if err := os.MkdirAll(stdlibPath, 0755); err != nil {
		t.Fatal(err)
	}

	if err := ensureGoMod(dir, stdlibPath); err != nil {
		t.Fatalf("ensureGoMod error: %v", err)
	}

	result, err := os.ReadFile(goModPath)
	if err != nil {
		t.Fatal(err)
	}

	content := string(result)
	if !strings.Contains(content, "github.com/duber000/kukicha/stdlib") {
		t.Error("expected go.mod to contain stdlib require")
	}
	if !strings.Contains(content, "replace") {
		t.Error("expected go.mod to contain replace directive")
	}
}

func TestEnsureGoMod_IdempotentOnRerun(t *testing.T) {
	dir := t.TempDir()

	goModContent := "module example.com/app\n\ngo 1.26\n"
	goModPath := filepath.Join(dir, "go.mod")
	if err := os.WriteFile(goModPath, []byte(goModContent), 0644); err != nil {
		t.Fatal(err)
	}

	stdlibPath := filepath.Join(dir, ".kukicha", "stdlib")
	if err := os.MkdirAll(stdlibPath, 0755); err != nil {
		t.Fatal(err)
	}

	// Run twice
	if err := ensureGoMod(dir, stdlibPath); err != nil {
		t.Fatalf("first ensureGoMod: %v", err)
	}
	firstResult, _ := os.ReadFile(goModPath)

	if err := ensureGoMod(dir, stdlibPath); err != nil {
		t.Fatalf("second ensureGoMod: %v", err)
	}
	secondResult, _ := os.ReadFile(goModPath)

	if string(firstResult) != string(secondResult) {
		t.Error("expected ensureGoMod to be idempotent")
	}
}

func TestEnsureGoMod_NoGoModErrors(t *testing.T) {
	dir := t.TempDir()
	// No go.mod exists
	err := ensureGoMod(dir, filepath.Join(dir, ".kukicha", "stdlib"))
	if err == nil {
		t.Fatal("expected error when go.mod is missing")
	}
	if !strings.Contains(err.Error(), "no go.mod found") {
		t.Errorf("expected 'no go.mod found' error, got: %v", err)
	}
}

func TestUpsertSkillSection_CreatesNewFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "AGENTS.md")

	if err := upsertSkillSection(path, "skill content"); err != nil {
		t.Fatalf("upsertSkillSection error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if !strings.Contains(content, skillStart) {
		t.Error("expected skill start marker")
	}
	if !strings.Contains(content, skillEnd) {
		t.Error("expected skill end marker")
	}
	if !strings.Contains(content, "skill content") {
		t.Error("expected skill content")
	}
}

func TestUpsertSkillSection_UpdatesExisting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "AGENTS.md")

	// First insert
	if err := upsertSkillSection(path, "old content"); err != nil {
		t.Fatal(err)
	}

	// Update
	if err := upsertSkillSection(path, "new content"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if strings.Contains(content, "old content") {
		t.Error("expected old content to be replaced")
	}
	if !strings.Contains(content, "new content") {
		t.Error("expected new content")
	}

	// Should have exactly one start/end pair
	if strings.Count(content, skillStart) != 1 {
		t.Errorf("expected exactly 1 start marker, got %d", strings.Count(content, skillStart))
	}
}

func TestUpsertSkillSection_AppendsToExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "AGENTS.md")

	// Create file with existing content
	if err := os.WriteFile(path, []byte("# My Project\n\nExisting docs.\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := upsertSkillSection(path, "skill content"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if !strings.Contains(content, "# My Project") {
		t.Error("expected existing content to be preserved")
	}
	if !strings.Contains(content, "skill content") {
		t.Error("expected skill content to be appended")
	}
}

func TestAppendIfMissing_AppendsLine(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")

	if err := os.WriteFile(path, []byte("line1\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := appendIfMissing(path, "line2"); err != nil {
		t.Fatalf("appendIfMissing error: %v", err)
	}

	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), "line2") {
		t.Error("expected line2 to be appended")
	}
}

func TestAppendIfMissing_SkipsIfPresent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")

	if err := os.WriteFile(path, []byte("line1\nline2\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := appendIfMissing(path, "line2"); err != nil {
		t.Fatalf("appendIfMissing error: %v", err)
	}

	data, _ := os.ReadFile(path)
	if strings.Count(string(data), "line2") != 1 {
		t.Errorf("expected exactly 1 occurrence of 'line2', got %d", strings.Count(string(data), "line2"))
	}
}

func TestFindProjectDir_WithGoMod(t *testing.T) {
	dir := t.TempDir()

	// Create nested structure with go.mod at root
	subDir := filepath.Join(dir, "cmd", "app")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n\ngo 1.26\n"), 0644); err != nil {
		t.Fatal(err)
	}

	result := findProjectDir(filepath.Join(subDir, "main.go"))
	absDir, _ := filepath.Abs(dir)
	if result != absDir {
		t.Errorf("expected %s, got %s", absDir, result)
	}
}

func TestFindProjectDir_NoGoMod(t *testing.T) {
	dir := t.TempDir()
	subDir := filepath.Join(dir, "src")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	result := findProjectDir(filepath.Join(subDir, "main.go"))
	absSub, _ := filepath.Abs(subDir)
	if result != absSub {
		t.Errorf("expected %s, got %s", absSub, result)
	}
}
