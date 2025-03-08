package install

import (
	"os"

	config "github.com/hegner123/modulacms/internal/Config"
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

func CheckDb(c config.Config) error {
	dbc := db.ConfigDB(c)
	_, _, err := dbc.GetConnection()
	if err != nil {
		return err
	}

	return nil
}

func CheckCerts(path string) bool {
	b := true
	return b
}
