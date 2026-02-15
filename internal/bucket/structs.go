package bucket

import (
	config "github.com/hegner123/modulacms/internal/config"
)

// S3Credentials holds S3-compatible object storage configuration.
type S3Credentials struct {
	AccessKey      string
	SecretKey      string
	URL            string
	Region         string
	ForcePathStyle bool
}

// GetS3Creds extracts S3 credentials from the application configuration.
func GetS3Creds(c *config.Config) *S3Credentials {
	var S3Creds = S3Credentials{
		AccessKey:      c.Bucket_Access_Key,
		SecretKey:      c.Bucket_Secret_Key,
		URL:            c.BucketEndpointURL(),
		Region:         c.Bucket_Region,
		ForcePathStyle: c.Bucket_Force_Path_Style,
	}
	return &S3Creds
}
