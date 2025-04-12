package update

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/hegner123/modulacms/internal/utility"
)

// update downloads an updated version of the current executable from the provided URL
// and replaces the current running binary with the new version.
func Fetch(url string) error {

	// Step 1: Download the new binary
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download update: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			utility.DefaultLogger.Error("", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status downloading update: %s", resp.Status)
	}

	// Step 2: Write the downloaded binary to a temporary file
	tmpFile, err := os.CreateTemp("", "update-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %v", err)
	}
	defer func() {
		if err := tmpFile.Close(); err != nil {
			utility.DefaultLogger.Error("", err)
		}
	}()

	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write update to temporary file: %v", err)
	}

	// Step 3: Set the appropriate executable permissions
	if err = os.Chmod(tmpFile.Name(), 0755); err != nil {
		return fmt.Errorf("failed to set executable permissions: %v", err)
	}

	// Step 4: Determine the path to the currently running executable
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to determine executable path: %v", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath) // resolves symlinks if needed
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %v", err)
	}

	// Optional: Backup the current executable (recommended for rollback)
	backupPath := execPath + ".bak"
	if err := os.Rename(execPath, backupPath); err != nil {
		return fmt.Errorf("failed to backup current executable: %v", err)
	}

	// Step 5: Replace the current executable with the new version
	if err := os.Rename(tmpFile.Name(), execPath); err != nil {
		// Optionally, attempt to rollback by restoring the backup
		err := os.Rename(backupPath, execPath)
		if err != nil {
			return fmt.Errorf("failed to revert executable: %v", err)
		}
		return fmt.Errorf("failed to update executable: %v", err)
	}

	// Optional: Remove the backup if update succeeded
	_ = os.Remove(backupPath)

	return nil
}
