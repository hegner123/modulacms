package model

import (
	"fmt"
	"os"
	"testing"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
)

func TestBuildTree(t *testing.T) {
    p := config.NewFileProvider("")
    m := config.NewManager(p)
    c,err := m.Config()
    if err!=nil {
        t.Fatal(err)
    }
	f, err := os.Create("log.json")
	if err != nil {
		t.Fatal(err)
	}
	d := db.ConfigDB(*c)
	cd, err := d.ListContentDataByRoute(1)
	if err != nil {
		t.Fatal(err)
	}
	dt := make([]db.Datatypes, 0)
	for _, v := range *cd {
		data, err := d.GetDatatype(v.DatatypeID)
		if err != nil {
			t.Fatal(err)
		}
		dt = append(dt, *data)
	}
	cf, err := d.ListContentFieldsByRoute(1)
	if err != nil {
		t.Fatal(err)
	}
	df := make([]db.Fields, 0)
	for _, v := range *cf {
		data, err := d.GetField(v.FieldID)
		if err != nil {
			t.Fatal(err)
		}
		df = append(df, *data)
	}

	r := BuildTree(*cd, dt, *cf, df)
	s := r.Render()
	_, err = fmt.Fprintln(f, s)
	if err != nil {
        t.Fatal(err)
	}
}
