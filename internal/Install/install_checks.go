package install

import (
	"fmt"
	"os"

	bucket "github.com/hegner123/modulacms/internal/Bucket"
	config "github.com/hegner123/modulacms/internal/Config"
	db "github.com/hegner123/modulacms/internal/Db"
	utility "github.com/hegner123/modulacms/internal/Utility"
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
    l:=utility.NewLogger(utility.INFO)
    l.Info("Config exists at", p)
	return nil
}

func CheckBucket()error {
	v := false
	c := config.LoadConfig(&v, "")
	if c.Bucket_Secret_Key == "" || c.Bucket_Endpoint == "" || c.Bucket_Access_Key == "" {
		err := fmt.Errorf("Bucket Access Key: %s\nBucket Secret Key: %s\nBucket Endpoint: %s\n", c.Bucket_Access_Key, c.Bucket_Secret_Key, c.Bucket_Endpoint)
		utility.LogError("Bucket fields not completed", err)
	}
	creds := bucket.S3Credintials{
		AccessKey: c.Bucket_Access_Key,
		SecretKey: c.Bucket_Secret_Key,
		URL:       c.Bucket_Endpoint,
	}
	_, err := creds.GetBucket()
	if err != nil {
		return err
	}
    l:=utility.NewLogger(utility.INFO)
    l.Info("Bucket Connected Successfully")
    return nil
}

func CheckOauth()error {
    v := false
    c := config.LoadConfig(&v, "")
    if c.Oauth_Client_Id =="" || c.Oauth_Client_Secret == ""  || c.Oauth_Endpoint["oauth_auth_url"] == "" || c.Oauth_Endpoint["oauth_token_url"]== ""{
        err := fmt.Errorf("")
		utility.LogError("Oauth fields not completed", err)
        return err
    }
    l:=utility.NewLogger(utility.INFO)
    l.Info("Oauth, no missing fields in config")
    return nil
}

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
