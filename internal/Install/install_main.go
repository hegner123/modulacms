package install

import (
	"fmt"

	config "github.com/hegner123/modulacms/internal/Config"
)

func InstallMain(configPath string, v *bool) error {
	var (
		installConfig bool
		installDb     bool
		promptBucket  bool
		promptOauth   bool
	)
	err := CheckConfigExists(configPath)
	if err != nil {
        installConfig = true
	}
	c := config.LoadConfig(v, configPath)
	err = CheckDb(c)
	if err != nil {
		installDb = true
	}
	err = InstallDependencies(installConfig, installDb, promptBucket, promptOauth)
	if err != nil {
		return err
	}
	return nil
}

func InstallDependencies(config bool, db bool, bucket bool, oauth bool) error {
	if config {
		err := CreateDefaultConfig("")
		if err != nil {
			return err
		}
	}
	if db {
		err := CreateDb("")
		if err != nil {
			return err
		}
	}
    if bucket {
        fmt.Printf("Bucket is not setup placeholder\n")
    }
    if oauth {
        fmt.Printf("oAuth is not setup placeholder\n")
    }
	return nil
}
