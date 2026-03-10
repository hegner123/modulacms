package update

import (
	"os"
	"path/filepath"
	"testing"
)

// Internal (white-box) tests for moveFile and copyFile.

func TestCopyFile_ContentAndPermissions(t *testing.T) {
	t.Parallel()

	srcDir := t.TempDir()
	dstDir := t.TempDir()

	srcPath := filepath.Join(srcDir, "src")
	dstPath := filepath.Join(dstDir, "dst")

	os.WriteFile(srcPath, []byte("binary-data"), 0755)

	if err := copyFile(srcPath, dstPath); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify content
	got, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("failed to read dst: %v", err)
	}
	if string(got) != "binary-data" {
		t.Errorf("content = %q, want %q", got, "binary-data")
	}

	// Verify permissions preserved
	info, err := os.Stat(dstPath)
	if err != nil {
		t.Fatalf("failed to stat dst: %v", err)
	}
	if info.Mode()&0111 == 0 {
		t.Error("executable permission not preserved")
	}

	// Verify source is untouched
	srcContent, err := os.ReadFile(srcPath)
	if err != nil {
		t.Fatalf("source should still exist: %v", err)
	}
	if string(srcContent) != "binary-data" {
		t.Error("source was modified")
	}
}

func TestCopyFile_SourceMissing(t *testing.T) {
	t.Parallel()

	dstPath := filepath.Join(t.TempDir(), "dst")

	err := copyFile("/nonexistent/path", dstPath)
	if err == nil {
		t.Fatal("expected error for missing source, got nil")
	}
}

func TestCopyFile_DestinationReadOnly(t *testing.T) {
	t.Parallel()

	srcDir := t.TempDir()
	dstDir := t.TempDir()

	srcPath := filepath.Join(srcDir, "src")
	dstPath := filepath.Join(dstDir, "dst")

	os.WriteFile(srcPath, []byte("data"), 0755)

	// Make destination directory read-only
	os.Chmod(dstDir, 0555)
	t.Cleanup(func() { os.Chmod(dstDir, 0755) })

	err := copyFile(srcPath, dstPath)
	if err == nil {
		t.Fatal("expected error for read-only destination, got nil")
	}
}

func TestMoveFile_SameDevice(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	srcPath := filepath.Join(dir, "src")
	dstPath := filepath.Join(dir, "dst")

	os.WriteFile(srcPath, []byte("move-me"), 0755)

	if err := moveFile(srcPath, dstPath); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Source should be gone (rename, not copy)
	if _, err := os.Stat(srcPath); !os.IsNotExist(err) {
		t.Error("source should not exist after same-device move")
	}

	got, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("failed to read destination: %v", err)
	}
	if string(got) != "move-me" {
		t.Errorf("content = %q, want %q", got, "move-me")
	}
}

func TestMoveFile_SourceMissing(t *testing.T) {
	t.Parallel()

	dstPath := filepath.Join(t.TempDir(), "dst")
	err := moveFile("/nonexistent/path", dstPath)
	if err == nil {
		t.Fatal("expected error for missing source, got nil")
	}
}
