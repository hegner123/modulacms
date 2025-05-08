package backup

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/hegner123/modulacms/internal/file"
	"github.com/hegner123/modulacms/internal/utility"
)

type RestoreConfirmHash string

var RestoreHash RestoreConfirmHash

type RestoreCMD struct {
	Path   string             `json:"path"`
	Origin string             `json:"origin"`
	Hash   RestoreConfirmHash `json:"hash"`
}

func IssueRestoreRequest(environment string) error {
	RestoreHash = createRestoreHash()
	cmd := RestoreCMD{
		Hash: RestoreHash,
	}
	j, err := json.Marshal(cmd)
	if err != nil {
		return err
	}
	r, err := http.Post(environment, utility.AppJson, bytes.NewBuffer(j))
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			utility.DefaultLogger.Error("", err)
		}
	}()

	return nil
}

func createRestoreHash() RestoreConfirmHash {
	t := time.Now().String()
	h := base64.NewEncoding(t)
	s := h.EncodeToString([]byte(t))
	return RestoreConfirmHash(s)
}

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
	utility.DefaultLogger.Info("Files synchronized successfully to", appDir)

	return nil
}
