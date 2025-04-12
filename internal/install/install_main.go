package install

import (
	"fmt"
	"os"

	config "github.com/hegner123/modulacms/internal/config"
	utility "github.com/hegner123/modulacms/internal/utility"
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

func CheckInstall(c *config.Config, v *bool) (ModulaInit, error) {
	Status := ModulaInit{}
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
	_, err = CheckDb(v, *c)
	if err != nil {
		Status.DBConnected = false
		return Status, err
	}
	_, err = CheckBucket(v,c)
	if err != nil {
		Status.BucketConnected = false
		return Status, err
	}
	_, err = CheckOauth(v,c)
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

func InstallMain(configPath string, c *config.Config, v *bool) error {
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
    if c == nil{
        err := fmt.Errorf("config does not exist %v", configPath)
        utility.DefaultLogger.Fatal("",err)
        return err
    }
	_, err = CheckDb(v, *c)
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
	if bucket {
		err := fmt.Errorf("bucket is not setup placeholder")
		utility.DefaultLogger.Warn("", err)
	}
	if oauth {
		err := fmt.Errorf("oAuth is not setup placeholder")
		utility.DefaultLogger.Warn("", err)
	}
	return nil
}
