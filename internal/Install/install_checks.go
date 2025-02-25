package install

import (
	"os"

	db "github.com/hegner123/modulacms/internal/Db"
)

func CheckConfigExists(path string) error {
	var p string
	if path != "" {
		p = path
	} else {
		p = "config.json"
	}
	_, err := os.Stat(p)
	if err != nil {
		return err
	}
	return nil
}

func CheckBucket() {}

func CheckOauth() {}

func CheckDb(path string) error {
	var p string
	if path != "" {
		p = path
	} else {
		p = "modula.db"
	}
	dbc := db.GetDb(db.Database{Src: p})
	if dbc.Err != nil {
		return dbc.Err
	}

	return nil
}
