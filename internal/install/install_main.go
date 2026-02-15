package install

import (
	config "github.com/hegner123/modulacms/internal/config"
	utility "github.com/hegner123/modulacms/internal/utility"
)

// ModulaInit represents the installation status of various system components.
type ModulaInit struct {
	UseSSL          bool
	DbFileExists    bool
	Certificates    bool
	Key             bool
	ConfigExists    bool
	DBConnected     bool
	BucketConnected bool
	OauthConnected  bool
}

// CheckInstall validates the installation status of the ModulaCMS system by checking configuration, database, bucket, OAuth, and SSL certificates.
func CheckInstall(c *config.Config, v *bool) (ModulaInit, error) {
	Status := ModulaInit{}

	err := CheckConfigExists("")
	if err != nil {
		Status.ConfigExists = false
		Status.DBConnected = false
		Status.BucketConnected = false
		Status.OauthConnected = false
		return Status, err
	}
	Status.ConfigExists = true

	_, err = CheckDb(v, *c)
	if err != nil {
		Status.DBConnected = false
		return Status, err
	}
	Status.DBConnected = true

	// Bucket is optional -- log warning but do not fail
	bucketStatus, bucketErr := CheckBucket(v, c)
	if bucketErr != nil {
		Status.BucketConnected = false
		utility.DefaultLogger.Warn("Bucket check failed (optional): "+bucketStatus, bucketErr)
	} else {
		Status.BucketConnected = true
	}

	// OAuth is optional -- log warning but do not fail
	oauthStatus, oauthErr := CheckOauth(v, c)
	if oauthErr != nil {
		Status.OauthConnected = false
		utility.DefaultLogger.Warn("OAuth check failed (optional): "+oauthStatus, oauthErr)
	} else {
		Status.OauthConnected = true
	}

	// Check for SSL certs using config cert directory
	certDir := c.Cert_Dir
	if certDir == "" {
		certDir = "./"
	}
	certsFound := CheckCerts(certDir)
	Status.Certificates = certsFound
	Status.Key = certsFound
	if !certsFound {
		Status.UseSSL = false
	}

	return Status, nil
}
