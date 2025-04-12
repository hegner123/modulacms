package cli

import (
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
)

// TODO Get Root Datatypes
func ListRootDatatypes(c config.Config) ([]string, *[]db.Datatypes, error) {
	res := []string{}
	d := db.ConfigDB(c)
	dt, err := d.ListDatatypesRoot()
	if err != nil {
		return nil, nil, err
	}
	for _, v := range *dt {
		res = append(res, v.Label)
	}
	return res, dt, nil
}

// TODO Get Fields of ROOT datatypes
func ListFieldsByDatatype(c config.Config, id int64) ([]string, *[]db.Fields, error) {
	res := []string{}
	d := db.ConfigDB(c)
	fs, err := d.ListFieldsByDatatypeID(id)
	if err != nil {
		return nil, nil, err
	}
	for _, v := range *fs {
		res = append(res, v.Label)
	}
	return res, fs, nil
}

// TODO Get Child Datatypes
// Compose Options From Results

// TODO GET Fields of child datatypes
