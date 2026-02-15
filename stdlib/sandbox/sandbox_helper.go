package sandbox

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
)

// New creates a new sandboxed Root at the given directory path.
// All file operations on the returned Root are confined to this directory.
func New(path string) (Root, error) {
	r, err := os.OpenRoot(path)
	if err != nil {
		return Root{}, fmt.Errorf("sandbox open: %w", err)
	}
	return Root{root: r, path: path}, nil
}

// Close releases the resources associated with the Root.
func Close(r Root) error {
	return r.root.Close()
}

// Read reads the entire contents of a file within the sandbox.
func Read(r Root, path string) ([]byte, error) {
	data, err := r.root.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("sandbox read: %w", err)
	}
	return data, nil
}

// ReadString reads the entire contents of a file within the sandbox as a string.
func ReadString(r Root, path string) (string, error) {
	data, err := r.root.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("sandbox read: %w", err)
	}
	return string(data), nil
}

// WriteString writes a string to a file within the sandbox, creating it if needed.
func WriteString(r Root, data string, path string) error {
	err := r.root.WriteFile(path, []byte(data), 0644)
	if err != nil {
		return fmt.Errorf("sandbox write: %w", err)
	}
	return nil
}

// Write marshals data to JSON and writes it to a file within the sandbox.
func Write(r Root, data any, path string) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("sandbox write marshal: %w", err)
	}
	if err := r.root.WriteFile(path, jsonData, 0644); err != nil {
		return fmt.Errorf("sandbox write: %w", err)
	}
	return nil
}

// AppendString appends a string to a file within the sandbox, creating it if needed.
func AppendString(r Root, data string, path string) error {
	f, err := r.root.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("sandbox append: %w", err)
	}
	defer f.Close()
	if _, err := f.Write([]byte(data)); err != nil {
		return fmt.Errorf("sandbox append: %w", err)
	}
	return nil
}

// Append marshals data to JSON and appends it to a file within the sandbox.
func Append(r Root, data any, path string) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("sandbox append marshal: %w", err)
	}
	jsonData = append(jsonData, '\n')
	f, err := r.root.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("sandbox append: %w", err)
	}
	defer f.Close()
	if _, err := f.Write(jsonData); err != nil {
		return fmt.Errorf("sandbox append: %w", err)
	}
	return nil
}

// MkDir creates a directory within the sandbox.
func MkDir(r Root, path string) error {
	if err := r.root.Mkdir(path, 0755); err != nil {
		return fmt.Errorf("sandbox mkdir: %w", err)
	}
	return nil
}

// MkDirAll creates a directory and all necessary parents within the sandbox.
func MkDirAll(r Root, path string) error {
	if err := r.root.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("sandbox mkdirall: %w", err)
	}
	return nil
}

// List returns the names of files and directories within a directory in the sandbox.
func List(r Root, path string) ([]string, error) {
	f, err := r.root.Open(path)
	if err != nil {
		return nil, fmt.Errorf("sandbox list: %w", err)
	}
	defer f.Close()
	entries, err := f.ReadDir(-1)
	if err != nil {
		return nil, fmt.Errorf("sandbox list: %w", err)
	}
	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.Name()
	}
	return names, nil
}

// Exists checks if a file or directory exists within the sandbox.
func Exists(r Root, path string) bool {
	_, err := r.root.Stat(path)
	return err == nil
}

// IsDir checks if a path is a directory within the sandbox.
func IsDir(r Root, path string) bool {
	info, err := r.root.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// IsFile checks if a path is a regular file within the sandbox.
func IsFile(r Root, path string) bool {
	info, err := r.root.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// Stat returns file info for a path within the sandbox.
func Stat(r Root, path string) (os.FileInfo, error) {
	info, err := r.root.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("sandbox stat: %w", err)
	}
	return info, nil
}

// Delete removes a file or empty directory within the sandbox.
func Delete(r Root, path string) error {
	if err := r.root.Remove(path); err != nil {
		return fmt.Errorf("sandbox delete: %w", err)
	}
	return nil
}

// DeleteAll removes a file or directory tree within the sandbox.
func DeleteAll(r Root, path string) error {
	if err := r.root.RemoveAll(path); err != nil {
		return fmt.Errorf("sandbox deleteall: %w", err)
	}
	return nil
}

// Rename renames a file or directory within the sandbox.
func Rename(r Root, oldpath, newpath string) error {
	if err := r.root.Rename(oldpath, newpath); err != nil {
		return fmt.Errorf("sandbox rename: %w", err)
	}
	return nil
}

// Path returns the root directory path of the sandbox.
func Path(r Root) string {
	return r.path
}

// FS returns an fs.FS scoped to the sandbox root for use with Go stdlib.
func FS(r Root) fs.FS {
	return r.root.FS()
}
