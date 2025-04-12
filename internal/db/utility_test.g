package db

import (
	"testing"

	config "github.com/hegner123/modulacms/internal/Config"
)

func TestTableSort(t *testing.T) {
	v := false
	c := config.LoadConfig(&v, "")
	d := ConfigDB(c)
	err := d.SortTables()
	if err != nil {
		t.Fatal(err)
	}

}
