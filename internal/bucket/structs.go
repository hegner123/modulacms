package bucket

import (
	config "github.com/hegner123/modulacms/internal/config"
)

type S3Credintials struct {
	AccessKey string
	SecretKey string
	URL       string
}

func GetS3Creds(c *config.Config) *S3Credintials {
	var S3Creds = S3Credintials{
		AccessKey: c.Bucket_Access_Key,
		SecretKey: c.Bucket_Secret_Key,
		URL:       c.Bucket_Url,
	}
	return &S3Creds
}
