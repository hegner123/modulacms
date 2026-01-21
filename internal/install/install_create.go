package install

import (
	"os"

	config "github.com/hegner123/modulacms/internal/config"
	db "github.com/hegner123/modulacms/internal/db"
)

// CreateDb creates database tables and bootstrap data with progress indicators
func CreateDb(path string, c *config.Config) error {
	d := db.ConfigDB(*c)

	progress := NewInstallProgress()

	progress.AddStep("Database tables", "Creating database tables", func() error {
		err := d.CreateAllTables()
		if err != nil {
			return ErrDBTables(err)
		}
		return nil
	})

	progress.AddStep("Bootstrap data", "Inserting bootstrap data", func() error {
		err := d.CreateBootstrapData()
		if err != nil {
			return ErrDBBootstrap(err)
		}
		return nil
	})

	progress.AddStep("Database validation", "Validating database setup", func() error {
		err := d.ValidateBootstrapData()
		if err != nil {
			return ErrDBBootstrap(err)
		}
		return nil
	})

	return progress.Run()
}

// CreateDbSimple creates database without progress indicators (for programmatic use)
func CreateDbSimple(path string, c *config.Config) error {
	d := db.ConfigDB(*c)

	err := d.CreateAllTables()
	if err != nil {
		return ErrDBTables(err)
	}

	err = d.CreateBootstrapData()
	if err != nil {
		return ErrDBBootstrap(err)
	}

	err = d.ValidateBootstrapData()
	if err != nil {
		return ErrDBBootstrap(err)
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
			return ErrConfigWrite(err, path)
		}
	} else {
		file, err = os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return ErrConfigWrite(err, path)
		}
	}
	defer file.Close()

	_, err = file.Write(c)
	if err != nil {
		return ErrConfigWrite(err, path)
	}
	return nil
}
