// Package bucket provides S3-compatible object storage integration for ModulaCMS,
// including client initialization, bucket operations, and file upload functionality.
package bucket

import (
	"encoding/json"
	"fmt"
	"mime"
	"os"
	"path/filepath"

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
		utility.DefaultLogger.Error("Failed to create session:", err)
		return nil, err
	}

	return s3.New(sess), nil
}

// EnsureBucket creates the named bucket if it does not already exist.
func EnsureBucket(svc *s3.S3, bucketName string) error {
	_, err := svc.HeadBucket(&s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err == nil {
		return nil
	}

	_, err = svc.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return fmt.Errorf("create bucket %q: %w", bucketName, err)
	}

	utility.DefaultLogger.Info("Created bucket:", bucketName)
	return nil
}

// bucketPolicyStatement represents a single statement in an S3 bucket policy.
type bucketPolicyStatement struct {
	Sid       string   `json:"Sid"`
	Effect    string   `json:"Effect"`
	Principal string   `json:"Principal"`
	Action    []string `json:"Action"`
	Resource  []string `json:"Resource"`
}

// bucketPolicy represents an S3 bucket policy document.
type bucketPolicy struct {
	Version   string                   `json:"Version"`
	Statement []bucketPolicyStatement  `json:"Statement"`
}

// SetPublicReadPolicy configures the bucket to allow anonymous read access.
// This is appropriate for media buckets where assets need to be publicly accessible.
func SetPublicReadPolicy(svc *s3.S3, bucketName string) error {
	policy := bucketPolicy{
		Version: "2012-10-17",
		Statement: []bucketPolicyStatement{
			{
				Sid:       "PublicRead",
				Effect:    "Allow",
				Principal: "*",
				Action:    []string{"s3:GetObject"},
				Resource:  []string{"arn:aws:s3:::" + bucketName + "/*"},
			},
		},
	}

	policyBytes, err := json.Marshal(policy)
	if err != nil {
		return fmt.Errorf("marshal public-read policy for %q: %w", bucketName, err)
	}

	policyStr := string(policyBytes)
	_, err = svc.PutBucketPolicy(&s3.PutBucketPolicyInput{
		Bucket: aws.String(bucketName),
		Policy: aws.String(policyStr),
	})
	if err != nil {
		return fmt.Errorf("set public-read policy on %q: %w", bucketName, err)
	}

	utility.DefaultLogger.Info("Set public-read policy on bucket:", bucketName)
	return nil
}

// UploadPrep prepares an S3 PutObjectInput for file upload with the specified path, bucket, file data, and ACL.
// ContentType is inferred from the file extension in uploadPath.
// ContentLength is set explicitly from the file size.
func UploadPrep(uploadPath string, bucketName string, data *os.File, acl string) (*s3.PutObjectInput, error) {
	info, err := data.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat upload file %q: %w", data.Name(), err)
	}

	size := info.Size()
	utility.DefaultLogger.Info("UploadPrep", "key", uploadPath, "file", data.Name(), "size", size)

	upload := &s3.PutObjectInput{
		Bucket:        aws.String(bucketName),
		Key:           aws.String(uploadPath),
		Body:          data,
		ACL:           aws.String(acl),
		ContentLength: aws.Int64(size),
	}

	if ct := mime.TypeByExtension(filepath.Ext(uploadPath)); ct != "" {
		upload.ContentType = aws.String(ct)
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
		utility.DefaultLogger.Info(aws.StringValue(bucket.Name), "created on", aws.TimeValue(bucket.CreationDate))
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
