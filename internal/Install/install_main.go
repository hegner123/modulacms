package install

import (
	"fmt"
	"os"

	config "github.com/hegner123/modulacms/internal/Config"
	utility "github.com/hegner123/modulacms/internal/Utility"
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


func CheckInstall() (ModulaInit, error) {
	Status := ModulaInit{}
	v := false
	err := CheckConfigExists("")
	if err != nil {
		Status.ConfigExists = false
		Status.DBConnected = false
		Status.BucketConnected = false
		Status.OauthConnected = false
		return Status, err
	} else {
		Status.ConfigExists = true
	}
	c := config.LoadConfig(&v, "")
	err = CheckDb(c)
	if err != nil {
		Status.DBConnected = false
		return Status, err
	}
	err = CheckBucket()
	if err != nil {
		Status.BucketConnected = false
		return Status, err
	}
	err = CheckOauth()
	if err != nil {
		Status.OauthConnected = false
		return Status, err
	}
	//Check for ssl certs
	_, err = os.Open("localhost.crt")
	Status.Certificates = true
	if err != nil {
		Status.Certificates = false
	}
	_, err = os.Open("localhost.key")
	Status.Key = true
	if err != nil {
		Status.Key = false
	}

	if !Status.Certificates || !Status.Key {
		// HUH form
		Status.UseSSL = false
	}

	//check for content version
	_, err = os.Stat("./content.version")
	if err != nil {
		utility.DefaultLogger.Debug("", err)
		Status.ContentVersion = false

	}

	return Status, nil

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
