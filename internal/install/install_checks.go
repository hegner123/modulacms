package install

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/hegner123/modulacms/internal/bucket"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/utility"
)

// DBStatus contains the result of a database connection check.
type DBStatus struct {
	Driver string
	URL    string
	Err    error
}

// CheckConfigExists verifies that a configuration file exists at the specified path.
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

// CheckBucket verifies that S3-compatible bucket credentials are valid and accessible.
func CheckBucket(v *bool, c *config.Config) (string, error) {
	verbose := false
	if v != nil {
		verbose = *v
	}
	if c.Bucket_Secret_Key == "" || c.Bucket_Endpoint == "" || c.Bucket_Access_Key == "" {
		// Empty bucket credentials - this is a non-fatal condition (bucket is optional)
		if verbose {
			utility.DefaultLogger.Warn("Bucket fields not completed - S3 storage will be unavailable", nil)
		}
		return "Not configured", nil
	}
	creds := bucket.S3Credentials{
		AccessKey:      c.Bucket_Access_Key,
		SecretKey:      c.Bucket_Secret_Key,
		URL:            c.BucketEndpointURL(),
		Region:         c.Bucket_Region,
		ForcePathStyle: c.Bucket_Force_Path_Style,
	}
	_, err := creds.GetBucket()
	if err != nil {
		installErr := ErrBucketConnect(err)
		if verbose {
			utility.DefaultLogger.Error("Bucket connection failed", installErr)
		}
		return installErr.Error(), installErr
	}
	if verbose {
		utility.DefaultLogger.Info("Bucket Connected Successfully")
	}
	return "Connected", nil
}

// CheckOauth verifies that OAuth configuration fields are populated.
func CheckOauth(v *bool, c *config.Config) (string, error) {
	verbose := false
	if v != nil {
		verbose = *v
	}
	if c.Oauth_Client_Id == "" || c.Oauth_Client_Secret == "" || c.Oauth_Endpoint["oauth_auth_url"] == "" || c.Oauth_Endpoint["oauth_token_url"] == "" {
		// Empty OAuth credentials - this is a non-fatal condition (OAuth is optional)
		if verbose {
			utility.DefaultLogger.Warn("OAuth fields not completed - OAuth will be unavailable", nil)
		}
		return "Not configured", nil
	}
	if verbose {
		utility.DefaultLogger.Info("OAuth, no missing fields in config")
	}
	return "Connected", nil
}

// CheckDb verifies that a database connection can be established with the configured credentials.
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
			utility.DefaultLogger.Error("DBCheck: ", err)
		}
		installErr := ErrDBConnect(err, string(c.Db_Driver))
		status.Err = installErr
		status.Driver = ""
		status.URL = ""
		return status, installErr
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

// CheckBackupTools verifies that the required database client tool (pg_dump, mysqldump)
// is available in $PATH for the configured database driver. SQLite needs no external tool.
// Returns a warning string (empty if OK) and an error if the tool is missing.
func CheckBackupTools(driver config.DbDriver) (warning string, err error) {
	switch driver {
	case config.Psql:
		if _, err := exec.LookPath("pg_dump"); err != nil {
			return "", fmt.Errorf("pg_dump not found in $PATH (required for PostgreSQL backups). Install postgresql-client")
		}
	case config.Mysql:
		if _, err := exec.LookPath("mysqldump"); err != nil {
			return "", fmt.Errorf("mysqldump not found in $PATH (required for MySQL backups). Install default-mysql-client")
		}
	case config.Sqlite:
		// No external tool needed
	}
	return "", nil
}

// CheckCerts checks whether the required SSL certificate files exist at the specified path.
func CheckCerts(path string) bool {
	certPath := filepath.Join(path, "localhost.crt")
	keyPath := filepath.Join(path, "localhost.key")
	_, certErr := os.Stat(certPath)
	_, keyErr := os.Stat(keyPath)
	return certErr == nil && keyErr == nil
}
