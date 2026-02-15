package sandbox

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func newTestSandbox(t *testing.T) Root {
	t.Helper()
	dir := t.TempDir()
	r, err := New(dir)
	if err != nil {
		t.Fatalf("New(%q): %v", dir, err)
	}
	t.Cleanup(func() { Close(r) })
	return r
}

func TestWriteAndRead(t *testing.T) {
	r := newTestSandbox(t)

	if err := WriteString(r, "hello sandbox", "test.txt"); err != nil {
		t.Fatalf("WriteString: %v", err)
	}

	data, err := Read(r, "test.txt")
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if string(data) != "hello sandbox" {
		t.Errorf("got %q, want %q", string(data), "hello sandbox")
	}
}

func TestReadString(t *testing.T) {
	r := newTestSandbox(t)

	if err := WriteString(r, "read me", "str.txt"); err != nil {
		t.Fatalf("WriteString: %v", err)
	}

	s, err := ReadString(r, "str.txt")
	if err != nil {
		t.Fatalf("ReadString: %v", err)
	}
	if s != "read me" {
		t.Errorf("got %q, want %q", s, "read me")
	}
}

func TestWriteJSON(t *testing.T) {
	r := newTestSandbox(t)

	data := map[string]int{"a": 1, "b": 2}
	if err := Write(r, data, "data.json"); err != nil {
		t.Fatalf("Write: %v", err)
	}

	content, err := ReadString(r, "data.json")
	if err != nil {
		t.Fatalf("ReadString: %v", err)
	}
	if content == "" {
		t.Error("expected non-empty JSON content")
	}
}

func TestAppendString(t *testing.T) {
	r := newTestSandbox(t)

	if err := AppendString(r, "line1\n", "log.txt"); err != nil {
		t.Fatalf("AppendString: %v", err)
	}
	if err := AppendString(r, "line2\n", "log.txt"); err != nil {
		t.Fatalf("AppendString: %v", err)
	}

	s, err := ReadString(r, "log.txt")
	if err != nil {
		t.Fatalf("ReadString: %v", err)
	}
	if s != "line1\nline2\n" {
		t.Errorf("got %q, want %q", s, "line1\nline2\n")
	}
}

func TestPathTraversalRejected(t *testing.T) {
	r := newTestSandbox(t)

	_, err := Read(r, "../etc/passwd")
	if err == nil {
		t.Error("expected error for path traversal ../etc/passwd, got nil")
	}

	err = WriteString(r, "pwned", "../../tmp/evil.txt")
	if err == nil {
		t.Error("expected error for path traversal write, got nil")
	}
}

func TestSymlinkEscapeRejected(t *testing.T) {
	r := newTestSandbox(t)

	// Create a symlink inside the sandbox pointing outside
	sandboxDir := Path(r)
	linkPath := filepath.Join(sandboxDir, "escape")
	if err := os.Symlink("/tmp", linkPath); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}

	// Attempting to read through the symlink should fail
	_, err := Read(r, "escape/somefile")
	if err == nil {
		t.Error("expected error when following symlink escape, got nil")
	}
}

func TestDirectoryOperations(t *testing.T) {
	r := newTestSandbox(t)

	if err := MkDir(r, "subdir"); err != nil {
		t.Fatalf("MkDir: %v", err)
	}
	if !IsDir(r, "subdir") {
		t.Error("expected subdir to be a directory")
	}

	if err := MkDirAll(r, "a/b/c"); err != nil {
		t.Fatalf("MkDirAll: %v", err)
	}
	if !IsDir(r, "a/b/c") {
		t.Error("expected a/b/c to be a directory")
	}

	if err := WriteString(r, "in subdir", "subdir/file.txt"); err != nil {
		t.Fatalf("WriteString: %v", err)
	}

	names, err := List(r, ".")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	found := false
	for _, n := range names {
		if n == "subdir" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'subdir' in list, got %v", names)
	}
}

func TestExistsIsFileIsDir(t *testing.T) {
	r := newTestSandbox(t)

	if Exists(r, "nope.txt") {
		t.Error("expected Exists to return false for missing file")
	}

	if err := WriteString(r, "x", "exist.txt"); err != nil {
		t.Fatalf("WriteString: %v", err)
	}
	if !Exists(r, "exist.txt") {
		t.Error("expected Exists to return true")
	}
	if !IsFile(r, "exist.txt") {
		t.Error("expected IsFile to return true")
	}
	if IsDir(r, "exist.txt") {
		t.Error("expected IsDir to return false for a file")
	}
}

func TestStat(t *testing.T) {
	r := newTestSandbox(t)

	if err := WriteString(r, "hello", "stat.txt"); err != nil {
		t.Fatalf("WriteString: %v", err)
	}

	info, err := Stat(r, "stat.txt")
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if info.Size() != 5 {
		t.Errorf("expected size 5, got %d", info.Size())
	}
}

func TestDeleteAndRename(t *testing.T) {
	r := newTestSandbox(t)

	if err := WriteString(r, "del me", "del.txt"); err != nil {
		t.Fatalf("WriteString: %v", err)
	}
	if err := Delete(r, "del.txt"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if Exists(r, "del.txt") {
		t.Error("expected file to be deleted")
	}

	if err := WriteString(r, "rename me", "old.txt"); err != nil {
		t.Fatalf("WriteString: %v", err)
	}
	if err := Rename(r, "old.txt", "new.txt"); err != nil {
		t.Fatalf("Rename: %v", err)
	}
	if Exists(r, "old.txt") {
		t.Error("expected old.txt to be gone after rename")
	}
	s, err := ReadString(r, "new.txt")
	if err != nil {
		t.Fatalf("ReadString: %v", err)
	}
	if s != "rename me" {
		t.Errorf("got %q, want %q", s, "rename me")
	}
}

func TestDeleteAll(t *testing.T) {
	r := newTestSandbox(t)

	if err := MkDirAll(r, "tree/sub"); err != nil {
		t.Fatalf("MkDirAll: %v", err)
	}
	if err := WriteString(r, "deep", "tree/sub/file.txt"); err != nil {
		t.Fatalf("WriteString: %v", err)
	}
	if err := DeleteAll(r, "tree"); err != nil {
		t.Fatalf("DeleteAll: %v", err)
	}
	if Exists(r, "tree") {
		t.Error("expected tree to be deleted")
	}
}

func TestPath(t *testing.T) {
	dir := t.TempDir()
	r, err := New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer Close(r)

	if Path(r) != dir {
		t.Errorf("Path() = %q, want %q", Path(r), dir)
	}
}

func TestFS(t *testing.T) {
	r := newTestSandbox(t)

	if err := WriteString(r, "via fs", "fstest.txt"); err != nil {
		t.Fatalf("WriteString: %v", err)
	}

	fsys := FS(r)
	data, err := fs.ReadFile(fsys, "fstest.txt")
	if err != nil {
		t.Fatalf("fs.ReadFile: %v", err)
	}
	if string(data) != "via fs" {
		t.Errorf("got %q, want %q", string(data), "via fs")
	}
}

func TestCloseThenOpFails(t *testing.T) {
	dir := t.TempDir()
	r, err := New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := Close(r); err != nil {
		t.Fatalf("Close: %v", err)
	}

	_, err = Read(r, "anything.txt")
	if err == nil {
		t.Error("expected error after Close, got nil")
	}
}
