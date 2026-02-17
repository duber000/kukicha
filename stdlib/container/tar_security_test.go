package container

import (
	"archive/tar"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateTarFromPath_SkipsSymlinks(t *testing.T) {
	dir := t.TempDir()

	// Create a real file
	realFile := filepath.Join(dir, "real.txt")
	if err := os.WriteFile(realFile, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create a symlink pointing outside the directory
	symlink := filepath.Join(dir, "escape")
	if err := os.Symlink("/etc/passwd", symlink); err != nil {
		t.Fatal(err)
	}

	reader, err := createTarFromPath(dir)
	if err != nil {
		t.Fatalf("createTarFromPath error: %v", err)
	}

	// Read the tar and verify symlink was skipped
	tr := tar.NewReader(reader.(*bytes.Buffer))
	var names []string
	for {
		header, err := tr.Next()
		if err != nil {
			break
		}
		names = append(names, filepath.Base(header.Name))
	}

	for _, name := range names {
		if name == "escape" {
			t.Error("symlink 'escape' should have been skipped but was included in tar")
		}
	}

	found := false
	for _, name := range names {
		if name == "real.txt" {
			found = true
		}
	}
	if !found {
		t.Error("expected real.txt in tar archive")
	}
}

func TestCreateTarFromPath_RejectsSymlinkSource(t *testing.T) {
	dir := t.TempDir()

	realFile := filepath.Join(dir, "real.txt")
	if err := os.WriteFile(realFile, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	symlink := filepath.Join(dir, "link")
	if err := os.Symlink(realFile, symlink); err != nil {
		t.Fatal(err)
	}

	_, err := createTarFromPath(symlink)
	if err == nil {
		t.Fatal("expected error when source path is a symlink")
	}
	if !strings.Contains(err.Error(), "symlink") {
		t.Errorf("expected symlink error, got: %v", err)
	}
}

func TestExtractTar_RejectsSymlinkEntries(t *testing.T) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	// Write a symlink entry
	err := tw.WriteHeader(&tar.Header{
		Name:     "evil-link",
		Typeflag: tar.TypeSymlink,
		Linkname: "../../../etc/passwd",
	})
	if err != nil {
		t.Fatal(err)
	}
	tw.Close()

	dest := t.TempDir()
	err = extractTar(&buf, dest)
	if err == nil {
		t.Fatal("expected error for symlink entry in tar")
	}
	if !strings.Contains(err.Error(), "unsupported link entry") {
		t.Errorf("expected unsupported link error, got: %v", err)
	}
}

func TestExtractTar_RejectsHardlinkEntries(t *testing.T) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	err := tw.WriteHeader(&tar.Header{
		Name:     "evil-hardlink",
		Typeflag: tar.TypeLink,
		Linkname: "../../../etc/shadow",
	})
	if err != nil {
		t.Fatal(err)
	}
	tw.Close()

	dest := t.TempDir()
	err = extractTar(&buf, dest)
	if err == nil {
		t.Fatal("expected error for hardlink entry in tar")
	}
	if !strings.Contains(err.Error(), "unsupported link entry") {
		t.Errorf("expected unsupported link error, got: %v", err)
	}
}

func TestExtractTar_RejectsPathTraversal(t *testing.T) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	content := []byte("pwned")
	err := tw.WriteHeader(&tar.Header{
		Name:     "../../../tmp/escape.txt",
		Typeflag: tar.TypeReg,
		Size:     int64(len(content)),
		Mode:     0o644,
	})
	if err != nil {
		t.Fatal(err)
	}
	tw.Write(content)
	tw.Close()

	dest := t.TempDir()
	err = extractTar(&buf, dest)
	if err == nil {
		t.Fatal("expected error for path traversal in tar")
	}
	if !strings.Contains(err.Error(), "invalid archive path") {
		t.Errorf("expected invalid archive path error, got: %v", err)
	}
}

func TestExtractTar_NormalRoundTrip(t *testing.T) {
	// Create a source directory with files
	srcDir := t.TempDir()
	subDir := filepath.Join(srcDir, "sub")
	if err := os.Mkdir(subDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "a.txt"), []byte("aaa"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "b.txt"), []byte("bbb"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create tar from source
	reader, err := createTarFromPath(srcDir)
	if err != nil {
		t.Fatalf("createTarFromPath error: %v", err)
	}

	// Extract to destination
	destDir := t.TempDir()
	if err := extractTar(reader, destDir); err != nil {
		t.Fatalf("extractTar error: %v", err)
	}

	// Verify files exist in destination
	baseName := filepath.Base(srcDir)
	aContent, err := os.ReadFile(filepath.Join(destDir, baseName, "a.txt"))
	if err != nil {
		t.Fatalf("failed to read extracted a.txt: %v", err)
	}
	if string(aContent) != "aaa" {
		t.Errorf("expected 'aaa', got %q", string(aContent))
	}

	bContent, err := os.ReadFile(filepath.Join(destDir, baseName, "sub", "b.txt"))
	if err != nil {
		t.Fatalf("failed to read extracted sub/b.txt: %v", err)
	}
	if string(bContent) != "bbb" {
		t.Errorf("expected 'bbb', got %q", string(bContent))
	}
}

func TestExtractTar_RejectsDeviceFiles(t *testing.T) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	err := tw.WriteHeader(&tar.Header{
		Name:     "dev-null",
		Typeflag: tar.TypeChar,
		Mode:     0o666,
	})
	if err != nil {
		t.Fatal(err)
	}
	tw.Close()

	dest := t.TempDir()
	err = extractTar(&buf, dest)
	if err == nil {
		t.Fatal("expected error for device file entry in tar")
	}
	if !strings.Contains(err.Error(), "unsupported entry type") {
		t.Errorf("expected unsupported entry type error, got: %v", err)
	}
}
