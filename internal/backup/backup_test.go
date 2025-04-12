package backup

import (
	"fmt"
	"testing"

	config "github.com/hegner123/modulacms/internal/config"
	utility "github.com/hegner123/modulacms/internal/utility"
)

func makeTestBackup(output string, timestamp string) string {
	return fmt.Sprintf("%s_%s_TEST.zip", output, utility.TimestampS())
}

func TestMakeBackup(t *testing.T) {
	p := config.NewFileProvider("")
	m := config.NewManager(p)
	c, err := m.Config()
	if err != nil {
		t.Fatal(err)
	}
	err = CreateBackup("../../modula.db", "../../public/media", "../../plugins/", "../../backups/", makeTestBackup, *c)
	if err != nil {
		t.Fatal(err)
	}
}
