package install

import (
	"fmt"

	config "github.com/hegner123/modulacms/internal/Config"
)

type ModulaInit struct {
	UseSSL          bool
	DbFileExists    bool
	ContentVersion  bool
	Certificates    bool
	Key             bool
	ConfigExists    bool
	DBConnected     bool
	BucketConnected bool
	OauthConnected  bool
}

func CheckInstall(init ModulaInit) (ModulaInit,error) {
	v := false
	err := CheckConfigExists("")
	if err != nil {
		init.ConfigExists = false
		init.DBConnected = false
		init.BucketConnected = false
		init.OauthConnected = false
        return init,err
	} else {
		init.ConfigExists = true
	}
	c := config.LoadConfig(&v, "")
	err = CheckDb(c)
	if err != nil {
		init.DBConnected = false
        return init, err
	}
	err = CheckBucket()
	if err != nil {
		init.BucketConnected = false
        return init, err
	}
	err = CheckOauth()
	if err != nil {
		init.OauthConnected = false
        return init, err
	}
	return init, nil

}

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
