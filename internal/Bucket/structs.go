package bucket

import (
	config "github.com/hegner123/modulacms/internal/config"
)

type S3Credintials struct {
	AccessKey string
	SecretKey string
	URL       string
}

func GetS3Creds() *S3Credintials {
	return &S3Creds
}

var S3Creds = S3Credintials{
	AccessKey: config.Env.Bucket_Access_Key,
	SecretKey: config.Env.Bucket_Secret_Key,
	URL:       config.Env.Bucket_Url,
}
