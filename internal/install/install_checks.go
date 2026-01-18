package install

import (
	"fmt"
	"os"

	"github.com/hegner123/modulacms/internal/bucket"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/utility"
)

type DBStatus struct {
	Driver string
	URL    string
	Err    error
}

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
	utility.DefaultLogger.Info("Config exists at", p)
	return nil
}

func CheckBucket(v *bool, c *config.Config) (string, error) {
	verbose := false
	if v != nil {
		verbose = *v
	}
	if c.Bucket_Secret_Key == "" || c.Bucket_Endpoint == "" || c.Bucket_Access_Key == "" {
		err := fmt.Errorf("bucket access key: %s\nbucket secret key: %s\nbucket endpoint: %s", c.Bucket_Access_Key, c.Bucket_Secret_Key, c.Bucket_Endpoint)
		utility.DefaultLogger.Error("Bucket fields not completed", err)
	}
	creds := bucket.S3Credentials{
		AccessKey: c.Bucket_Access_Key,
		SecretKey: c.Bucket_Secret_Key,
		URL:       c.Bucket_Endpoint,
	}
	_, err := creds.GetBucket()
	if err != nil {
		return err.Error(), err
	}
	if verbose {
		utility.DefaultLogger.Info("Bucket Connected Successfully")
	}
	return "Connected", nil
}

func CheckOauth(v *bool, c *config.Config) (string, error) {
	verbose := false
	if v != nil {
		verbose = *v
	}
	if c.Oauth_Client_Id == "" || c.Oauth_Client_Secret == "" || c.Oauth_Endpoint["oauth_auth_url"] == "" || c.Oauth_Endpoint["oauth_token_url"] == "" {
		err := fmt.Errorf("oauth fields not completed")
		if verbose {
			utility.DefaultLogger.Error("CheckOauth: ", err)
		}
		return "Oauth fields not completed", err
	}
	if verbose {
		utility.DefaultLogger.Info("Oauth, no missing fields in config")
	}
	return "Connected", nil
}

func CheckDb(v *bool, c config.Config) (DBStatus, error) {
	verbose := false
	status := DBStatus{}
	if v != nil {
		verbose = *v
	}
	dbc := db.ConfigDB(c)
	_, _, err := dbc.GetConnection()
	if err != nil {
		if verbose {
			err := fmt.Errorf("DB Not Connected")
			if verbose {
				utility.DefaultLogger.Error("DBCheck:  ", err)
			}

		}
		status.Err = err
		status.Driver = ""
		status.URL = ""
		return status, err
	}
	connected := fmt.Sprint("db connected: ", c.Db_Driver, " ", c.Db_URL)
	if verbose {
		utility.DefaultLogger.Info(connected)
	}
	status.Driver = string(c.Db_Driver)
	status.URL = c.Db_URL
	status.Err = nil

	return status, nil
}

func CheckCerts(path string) bool {
	b := true
	return b
}
