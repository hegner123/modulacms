package backup

import (
	"fmt"
	"testing"

	utility "github.com/hegner123/modulacms/internal/Utility"
)

func makeTestBackup(output string, timestamp string) string {
	return fmt.Sprintf("%s_%s_TEST.zip", output, utility.TimestampS())
}

func TestMakeBackup(t *testing.T) {
	err := createBackup("../../modula.db", "../../public/media", "../../plugins/", "../../backups/", makeTestBackup)
	if err != nil {
		t.Errorf("%s\n", err)
	}
}
