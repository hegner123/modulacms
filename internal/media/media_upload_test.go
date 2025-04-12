package media

import (
	"testing"

	config "github.com/hegner123/modulacms/internal/config"
)

func TestMediaUpload(t *testing.T) {
	p := config.NewFileProvider("")
	m := config.NewManager(p)
	c, err := m.Config()
	err = HandleMediaUpload("test.png", "test.png", *c)
	if err != nil {
		t.Fatal(err)
	}
}
