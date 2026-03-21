package update

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/hegner123/modulacms/internal/utility"
)

// DownloadUpdate downloads the new binary from the provided URL to a temporary file
// Returns: (tempFilePath, error)
func DownloadUpdate(url string) (string, error) {
	// Download the new binary
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download update: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			utility.DefaultLogger.Error("", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status downloading update: %s", resp.Status)
	}

	// Write the downloaded binary to a temporary file
	tmpFile, err := os.CreateTemp("", "update-*.tmp")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %v", err)
	}
	tmpPath := tmpFile.Name()

	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return "", fmt.Errorf("failed to write update to temporary file: %v", err)
	}

	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpPath)
		return "", fmt.Errorf("failed to close temporary file: %v", err)
	}

	// Set executable permissions
	if err = os.Chmod(tmpPath, 0755); err != nil {
		os.Remove(tmpPath)
		return "", fmt.Errorf("failed to set executable permissions: %v", err)
	}

	// Verify the binary
	if err := VerifyBinary(tmpPath); err != nil {
		os.Remove(tmpPath)
		return "", fmt.Errorf("binary verification failed: %v", err)
	}

	return tmpPath, nil
}

// ApplyUpdate resolves the running executable path (following symlinks)
// and replaces it with the file at tempPath.
func ApplyUpdate(tempPath string) error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to determine executable path: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}
	return ApplyUpdateTo(tempPath, execPath)
}

// ApplyUpdateTo replaces the binary at execPath with the file at tempPath.
// The execPath should be the resolved (non-symlink) path to the current binary.
// Uses moveFile internally to handle cross-device moves (e.g. /tmp to /app in Docker).
func ApplyUpdateTo(tempPath, execPath string) error {
	backupPath := execPath + ".bak"

	// Backup is always same-device (same directory), so os.Rename is safe.
	if err := os.Rename(execPath, backupPath); err != nil {
		return fmt.Errorf("failed to backup current executable: %w", err)
	}

	// The temp file may be on a different filesystem (e.g. /tmp vs /app),
	// so use moveFile which falls back to copy on cross-device errors.
	if err := moveFile(tempPath, execPath); err != nil {
		if rollbackErr := os.Rename(backupPath, execPath); rollbackErr != nil {
			return fmt.Errorf("failed to update executable and rollback failed: %v (rollback error: %v)", err, rollbackErr)
		}
		return fmt.Errorf("failed to update executable (rollback successful): %w", err)
	}

	_ = os.Remove(backupPath)
	return nil
}

// VerifyBinary performs basic sanity checks on the downloaded binary
func VerifyBinary(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat binary: %v", err)
	}

	// Check file size is reasonable (at least 1MB, less than 500MB)
	size := info.Size()
	if size < 1024*1024 {
		return fmt.Errorf("binary too small (%d bytes), likely corrupt", size)
	}
	if size > 500*1024*1024 {
		return fmt.Errorf("binary too large (%d bytes), suspiciously big", size)
	}

	// Check file is executable
	mode := info.Mode()
	if runtime.GOOS != "windows" && mode&0111 == 0 {
		return fmt.Errorf("binary is not executable")
	}

	return nil
}

// RollbackUpdate resolves the running executable path and restores its backup.
func RollbackUpdate() error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to determine executable path: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}
	return RollbackUpdateTo(execPath)
}

// RollbackUpdateTo restores the backup file (.bak) for the binary at execPath.
func RollbackUpdateTo(execPath string) error {
	backupPath := execPath + ".bak"
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("no backup file found at %s", backupPath)
	}

	if err := os.Rename(backupPath, execPath); err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
	}

	return nil
}

// moveFile attempts os.Rename, falling back to copy+remove for cross-device moves.
func moveFile(src, dst string) error {
	err := os.Rename(src, dst)
	if err == nil {
		return nil
	}

	// Cross-device rename: copy the file then remove the source.
	var linkErr *os.LinkError
	if errors.As(err, &linkErr) && errors.Is(linkErr.Err, syscall.EXDEV) {
		if copyErr := copyFile(src, dst); copyErr != nil {
			return fmt.Errorf("cross-device move failed: %w", copyErr)
		}
		os.Remove(src)
		return nil
	}

	return err
}

// copyFile copies src to dst, preserving the source file's permissions.
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source: %w", err)
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat source: %w", err)
	}

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination: %w", err)
	}

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		dstFile.Close()
		os.Remove(dst)
		return fmt.Errorf("failed to copy data: %w", err)
	}

	if err := dstFile.Close(); err != nil {
		os.Remove(dst)
		return fmt.Errorf("failed to finalize destination: %w", err)
	}

	return nil
}

// Fetch is the legacy function that combines DownloadUpdate and ApplyUpdate
// Deprecated: Use DownloadUpdate and ApplyUpdate separately for better control
func Fetch(url string) error {
	tempPath, err := DownloadUpdate(url)
	if err != nil {
		return err
	}
	defer os.Remove(tempPath) // Clean up temp file if apply fails

	return ApplyUpdate(tempPath)
}
