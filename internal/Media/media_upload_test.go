package media

import (
	"testing"

	config "github.com/hegner123/modulacms/internal/Config"
)

func TestMediaUpload(t *testing.T) {
	verbose := false
	c := config.LoadConfig(&verbose, "")
    err:=HandleMediaUpload("test.png", "test.png", c)
    if err!=nil {
        t.Fatal(err)
    }
}
