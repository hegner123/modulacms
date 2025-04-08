package install

import (
	"os"

	config "github.com/hegner123/modulacms/internal/config"
	db "github.com/hegner123/modulacms/internal/db"
)

func CreateDb(path string) error {
	c := config.Env
	d := db.ConfigDB(c)
	err := d.CreateAllTables()
	if err != nil {
		return err
	}
	return nil
}

func CreateDefaultConfig(path string) error {
	var file *os.File
	c := config.DefaultConfig().JSON()

	_, err := os.Stat(path)
	if err != nil {
		file, err = os.Create(path)
		if err != nil {
			return err
		}
	} else {

		file, err = os.Open(path)
		if err != nil {
			return err
		}
	}

	_, err = file.Write(c)
	if err != nil {
		return err
	}
	return nil
}
