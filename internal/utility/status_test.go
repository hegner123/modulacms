package utility

import (
	"os"
	"path/filepath"
	"testing"
)

// ============================================================
// FileExists
// ============================================================

func TestFileExists(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// Create a real file
	filePath := filepath.Join(dir, "exists.txt")
	if err := os.WriteFile(filePath, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Create a subdirectory
	subDir := filepath.Join(dir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}

	tests := []struct {
		name string
		path string
		want bool
	}{
		{name: "existing file", path: filePath, want: true},
		{name: "nonexistent file", path: filepath.Join(dir, "nope.txt"), want: false},
		{name: "directory is not a file", path: subDir, want: false},
		{name: "empty path", path: "", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := FileExists(tt.path)
			if got != tt.want {
				t.Errorf("FileExists(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

// ============================================================
// DirExists
// ============================================================

func TestDirExists(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// Create a real file
	filePath := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(filePath, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Create a subdirectory
	subDir := filepath.Join(dir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}

	tests := []struct {
		name string
		path string
		want bool
	}{
		{name: "existing directory", path: subDir, want: true},
		{name: "temp dir itself", path: dir, want: true},
		{name: "nonexistent directory", path: filepath.Join(dir, "nope"), want: false},
		{name: "file is not a directory", path: filePath, want: false},
		{name: "empty path", path: "", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := DirExists(tt.path)
			if got != tt.want {
				t.Errorf("DirExists(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}
