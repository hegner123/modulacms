package update

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

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

// ApplyUpdate replaces the current binary with the downloaded version
// Takes the path to the downloaded temporary file
func ApplyUpdate(tempPath string) error {
	// Determine the path to the currently running executable
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to determine executable path: %v", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %v", err)
	}

	// Backup the current executable
	backupPath := execPath + ".bak"
	if err := os.Rename(execPath, backupPath); err != nil {
		return fmt.Errorf("failed to backup current executable: %v", err)
	}

	// Replace the current executable with the new version
	if err := os.Rename(tempPath, execPath); err != nil {
		// Attempt rollback
		if rollbackErr := os.Rename(backupPath, execPath); rollbackErr != nil {
			return fmt.Errorf("failed to update executable and rollback failed: %v (rollback error: %v)", err, rollbackErr)
		}
		return fmt.Errorf("failed to update executable (rollback successful): %v", err)
	}

	// Remove the backup if update succeeded
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
	if mode&0111 == 0 {
		return fmt.Errorf("binary is not executable")
	}

	return nil
}

// RollbackUpdate restores the previous version from backup
func RollbackUpdate() error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to determine executable path: %v", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %v", err)
	}

	backupPath := execPath + ".bak"
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("no backup file found at %s", backupPath)
	}

	if err := os.Rename(backupPath, execPath); err != nil {
		return fmt.Errorf("failed to restore backup: %v", err)
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
