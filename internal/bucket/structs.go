package bucket

import (
	config "github.com/hegner123/modulacms/internal/config"
)

type S3Credentials struct {
	AccessKey string
	SecretKey string
	URL       string
}

func GetS3Creds(c *config.Config) *S3Credentials {
	var S3Creds = S3Credentials{
		AccessKey: c.Bucket_Access_Key,
		SecretKey: c.Bucket_Secret_Key,
		URL:       c.Bucket_Url,
	}
	return &S3Creds
}
