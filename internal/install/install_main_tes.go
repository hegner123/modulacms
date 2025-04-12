package install

import (
	"testing"

	"github.com/hegner123/modulacms/internal/config"
)

func TestMainInstall(t *testing.T) {
	p := config.NewFileProvider("")
	m := config.NewManager(p)
	c, err := m.Config()
	if err != nil {
		t.Fatal(err)
	}
	verbose := false
	err = InstallMain("", c, &verbose)
	if err != nil {
		t.Fatal(err)
	}
}
