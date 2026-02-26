package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/duber000/kukicha/internal/formatter"
)

func TestCheckFile_FormattedFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "good.kuki")

	// Use content that the formatter produces (parens around return expr)
	content := "func Add(a int, b int) int\n    return (a + b)\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	opts := formatter.DefaultOptions()
	if !checkFile(path, opts) {
		t.Error("expected checkFile to return true for a formatted file")
	}
}

func TestCheckFile_UnformattedFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.kuki")

	// Go-style braces — the preprocessor will convert these, producing different output
	content := "func Add(a int, b int) int {\n    return a + b\n}\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	opts := formatter.DefaultOptions()
	if checkFile(path, opts) {
		t.Error("expected checkFile to return false for an unformatted file")
	}
}

func TestCheckFile_NonExistentFile(t *testing.T) {
	opts := formatter.DefaultOptions()
	if checkFile("/nonexistent/file.kuki", opts) {
		t.Error("expected checkFile to return false for missing file")
	}
}

func TestFormatFileInPlace_ConvertsGoStyle(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "fix.kuki")

	// Go-style braces — should be converted to Kukicha indentation style
	content := "func Add(a int, b int) int {\n    return a + b\n}\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	opts := formatter.DefaultOptions()
	if err := formatFileInPlace(path, opts); err != nil {
		t.Fatalf("formatFileInPlace error: %v", err)
	}

	result, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	// Should not contain braces after formatting
	if strings.Contains(string(result), "{") || strings.Contains(string(result), "}") {
		t.Errorf("expected braces to be removed, got: %q", string(result))
	}
}

func TestFormatFileInPlace_AlreadyFormatted(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "good.kuki")

	// Content that the formatter already produces
	content := "func Add(a int, b int) int\n    return (a + b)\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	opts := formatter.DefaultOptions()
	if err := formatFileInPlace(path, opts); err != nil {
		t.Fatalf("formatFileInPlace error: %v", err)
	}

	result, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if string(result) != content {
		t.Errorf("expected no changes for already-formatted file\ngot:  %q\nwant: %q", string(result), content)
	}
}

func TestFormatFileToStdout_WritesToStdout(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.kuki")

	content := "func Add(a int, b int) int\n    return (a + b)\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	opts := formatter.DefaultOptions()
	// formatFileToStdout writes to os.Stdout; we just verify it doesn't error
	if err := formatFileToStdout(path, opts); err != nil {
		t.Fatalf("formatFileToStdout error: %v", err)
	}
}

func TestFormatFileToStdout_NonExistentFile(t *testing.T) {
	opts := formatter.DefaultOptions()
	err := formatFileToStdout("/nonexistent/file.kuki", opts)
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestFormatFileInPlace_NonExistentFile(t *testing.T) {
	opts := formatter.DefaultOptions()
	err := formatFileInPlace("/nonexistent/file.kuki", opts)
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
