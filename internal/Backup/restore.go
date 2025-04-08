package backup

import (
	"os"
	"os/exec"

	file "github.com/hegner123/modulacms/internal/file"
	utility "github.com/hegner123/modulacms/internal/utility"
)

func RestoreBackup(path string, tempDir string, appDir string) error {
	err := file.Unzip(path, "./")
	if err != nil {
		return err
	}
	// Rsync the unzipped contents to the application working directory.
	// The trailing "/" ensures rsync copies the directory contents.
	rsyncCmd := exec.Command("rsync", "-av", tempDir+"/", appDir+"/")
	rsyncCmd.Stdout = os.Stdout
	rsyncCmd.Stderr = os.Stderr

	if err := rsyncCmd.Run(); err != nil {
		utility.DefaultLogger.Error("rsync failed", err)
	}
	l := utility.NewLogger(utility.INFO)
	l.Info("Files synchronized successfully to", appDir)

	return nil
}
