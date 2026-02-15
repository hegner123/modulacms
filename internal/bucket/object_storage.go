// Package bucket provides S3-compatible object storage integration for ModulaCMS,
// including client initialization, bucket operations, and file upload functionality.
package bucket

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	utility "github.com/hegner123/modulacms/internal/utility"
)

// Metadata represents object metadata as key-value pairs.
type Metadata map[string]string

// GetBucket establishes and returns an S3 client session using the stored credentials.
func (cs S3Credentials) GetBucket() (*s3.S3, error) {
	region := cs.Region
	if region == "" {
		region = "us-east-1"
	}

	sess, err := session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(cs.AccessKey, cs.SecretKey, ""),
		Endpoint:         aws.String(cs.URL),
		Region:           aws.String(region),
		S3ForcePathStyle: aws.Bool(cs.ForcePathStyle),
	})
	if err != nil {
		utility.DefaultLogger.Error("Failed to create session: %v", err)
		return nil, err
	}

	return s3.New(sess), nil
}

// UploadPrep prepares an S3 PutObjectInput for file upload with the specified path, bucket, file data, and ACL.
func UploadPrep(uploadPath string, bucketName string, data *os.File, acl string) (*s3.PutObjectInput, error) {
	upload := &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(uploadPath),
		Body:   data,
		ACL:    aws.String(acl),
	}
	return upload, nil
}

// PrintBuckets lists all available S3 buckets and logs their names and creation dates.
func PrintBuckets(s3 *s3.S3) {
	// Example: List buckets
	result, err := s3.ListBuckets(nil)
	if err != nil {
		utility.DefaultLogger.Fatal("Unable to list buckets: %w", err)
	}

	utility.DefaultLogger.Info("Buckets:")
	for _, bucket := range result.Buckets {
		utility.DefaultLogger.Info("%s created on %s\n",
			aws.StringValue(bucket.Name),
			aws.TimeValue(bucket.CreationDate))
	}
}

// ObjectUpload executes an S3 object upload using the provided client and upload configuration.
// Must use output of GetBucket as s3 argument.
// Must use output of bucket.UploadPrep function as payload.
func ObjectUpload(s3 *s3.S3, payload *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
	upload, err := s3.PutObject(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to upload to S3: %w", err)
	}
	return upload, nil
}
