package install

import (
	"testing"

	config "github.com/hegner123/modulacms/internal/config"
)

func TestConfigPathCheck(t *testing.T) {
	err := CheckConfigExists("")
	if err != nil {
		t.Fatal(err)
	}

}

func TestDbExists(t *testing.T) {
	v := false
	c := config.Config{
		Db_Driver: "sqlite",
	}
	_, err := CheckDb(&v, c)
	if err != nil {
		t.Fatal(err)
	}

}
