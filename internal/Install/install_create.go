package install

import (
	"os"

	config "github.com/hegner123/modulacms/internal/Config"
)

func CreateDb(path string) error {
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
